"""A minimal Minecraft 1.12.2 bot, just enough to be an online player.

Some 1.12.2 commands only work on a player that is **online**: `/advancement`
and `/recipe` are two of them, and a console-only probe can measure nothing
but their error messages. This bot logs in and stays, so those commands can
be measured for real.

The probe server runs with `online-mode=false`, so the login needs no
encryption: a handshake and a login start are enough for the server to add
the player to its player list. After that the bot only has to answer the
keep-alive packets so the connection is not dropped.

Usage:

    python mcbot.py <host> <port> <name> [seconds]

    python mcbot.py 127.0.0.1 25599 Steve 120

The port has to be the one the probe server runs on, and the server has to
be started with `online-mode=false` (the probe does that). The bot prints
what it sees on the standard output; `BOT_DEBUG=1` adds a hex dump of every
packet.

The packet ids are the ones of **Paper** 1.12.2, whose registration order
differs from the vanilla one: the clientbound keep alive is 0x1F and the
serverbound one is 0x0B.
"""
import socket
import struct
import sys
import time
import zlib
import os


def varint(n):
    out = b""
    while True:
        b = n & 0x7F
        n >>= 7
        if n:
            out += bytes([b | 0x80])
        else:
            return out + bytes([b])


def mcstring(s):
    raw = s.encode("utf-8")
    return varint(len(raw)) + raw


def split_varint(data):
    """Return (value, rest) for the varint at the start of data."""
    num = 0
    shift = 0
    for i, b in enumerate(data):
        num |= (b & 0x7F) << shift
        if not b & 0x80:
            return num, data[i + 1:]
        shift += 7
    raise ValueError("truncated varint")


def varint_size(n):
    size = 1
    while n >= 0x80:
        n >>= 7
        size += 1
    return size


class Conn:
    def __init__(self, sock):
        self.sock = sock
        self.buf = b""
        self.threshold = -1

    def _read(self, n):
        while len(self.buf) < n:
            chunk = self.sock.recv(65536)
            if not chunk:
                raise EOFError("connection closed")
            self.buf += chunk
        out, self.buf = self.buf[:n], self.buf[n:]
        return out

    def read_varint(self):
        num = 0
        shift = 0
        while True:
            b = self._read(1)[0]
            num |= (b & 0x7F) << shift
            if not b & 0x80:
                return num
            shift += 7

    def read_packet(self):
        length = self.read_varint()
        if self.threshold >= 0:
            data_len = self.read_varint()
            # The dataLength varint is part of the packet length.
            data = self._read(length - varint_size(data_len))
            if data_len != 0:
                data = zlib.decompress(data)
            return data
        return self._read(length)

    def write_packet(self, pid, payload=b""):
        body = varint(pid) + payload
        if self.threshold >= 0:
            if len(body) >= self.threshold:
                raw = zlib.compress(body)
                framed = varint(varint_size(len(body)) + len(raw)) + varint(len(body)) + raw
            else:
                framed = varint(varint_size(0) + len(body)) + varint(0) + body
        else:
            framed = varint(len(body)) + body
        self.sock.sendall(framed)


def main():
    host = sys.argv[1] if len(sys.argv) > 1 else "127.0.0.1"
    port = int(sys.argv[2]) if len(sys.argv) > 2 else 25599
    name = sys.argv[3] if len(sys.argv) > 3 else "Steve"
    seconds = float(sys.argv[4]) if len(sys.argv) > 4 else 120.0

    deadline = time.time() + seconds
    sock = None
    while time.time() < deadline:
        try:
            sock = socket.create_connection((host, port), timeout=5)
            break
        except OSError:
            time.sleep(0.3)
    if sock is None:
        print("BOT: could not connect", flush=True)
        return 1
    sock.settimeout(10)
    conn = Conn(sock)

    handshake = varint(0) + varint(340) + mcstring(host) + struct.pack(">H", port) + varint(2)
    conn.sock.sendall(varint(len(handshake)) + handshake)
    login = varint(0) + mcstring(name)
    conn.sock.sendall(varint(len(login)) + login)

    state = "login"
    logged_in = False
    while time.time() < deadline:
        try:
            data = conn.read_packet()
        except (EOFError, socket.timeout):
            break
        pid, rest = split_varint(data)
        if os.environ.get("BOT_DEBUG"):
            print("BOT: raw packet %s" % data[:80].hex(), flush=True)
        if pid == 0x03:  # set compression (login or play state)
            conn.threshold, _ = split_varint(rest)
            print("BOT: compression threshold %d" % conn.threshold, flush=True)
            continue
        if state == "login":
            if pid == 0x02:  # login success
                logged_in = True
                print("BOT: logged in as %s" % name, flush=True)
                state = "play"
            elif pid == 0x00:  # disconnect
                print("BOT: login refused: %r" % rest, flush=True)
                break
            continue
        if pid == 0x1F:  # keep alive (play)
            # The serverbound keep alive is packet 0x0B in 1.12.2.
            conn.write_packet(0x0B, rest)
        elif pid == 0x00:  # disconnect
            print("BOT: disconnected by server", flush=True)
            break
    if not logged_in:
        print("BOT: never logged in", flush=True)
        return 1
    print("BOT: done", flush=True)
    return 0


if __name__ == "__main__":
    sys.exit(main())
