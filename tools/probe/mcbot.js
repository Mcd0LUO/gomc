/**
 * A minimal Minecraft 1.12.2 bot, just enough to be an online player.
 *
 * Some 1.12.2 commands only work on a player that is **online**: `/advancement`
 * and `/recipe` are two of them, and a console-only probe can measure nothing
 * but their error messages. This bot logs in and stays, so those commands can
 * be measured for real.
 *
 * The probe server runs with `online-mode=false`, so the login needs no
 * encryption: a handshake and a login start are enough for the server to add
 * the player to its player list. After that the library keeps the connection
 * alive on its own -- it answers the keep alive packets, which the server
 * sends every 15 seconds and which it uses to drop a player that does not
 * answer.
 *
 * The bot is written with `minecraft-protocol` (the node-minecraft-protocol
 * package) instead of the protocol being done by hand. The package knows the
 * packet numbers of every version, the compression and the keep alive, so
 * the bot does not have to.
 *
 * Usage:
 *
 *     node mcbot.js <host> <port> <name> [seconds] [logfile]
 *
 *     node mcbot.js 127.0.0.1 25599 Steve 120
 *
 * The port has to be the one the probe server runs on, and the server has to
 * be started with `online-mode=false` (the probe does that). The bot prints
 * what it sees on the standard output; `BOT_DEBUG=1` adds a dump of every
 * packet.
 *
 * The log file is **line buffered** so that a line is on disk as soon as it
 * is printed. The reason is PowerShell: `Start-Process -RedirectStandardOutput`
 * hands the output to a pipe that is drained **after** the process ends, so a
 * bot that is killed leaves an empty log behind -- and the probe waits for
 * the line that says the bot logged in, so it never sees it.
 */
"use strict";

const fs = require("fs");
const mc = require("minecraft-protocol");

function log(line) {
  process.stdout.write(line + "\n");
}

function main() {
  const host = process.argv[2] || "127.0.0.1";
  const port = parseInt(process.argv[3] || "25599", 10);
  const name = process.argv[4] || "Steve";
  const seconds = parseFloat(process.argv[5] || "120");
  const logPath = process.argv[6] || "";

  if (logPath) {
    try {
      fs.closeSync(fs.openSync(logPath, "w"));
      const stream = fs.createWriteStream(logPath, { flags: "a" });
      const write = process.stdout.write.bind(process.stdout);
      process.stdout.write = (chunk, ...rest) => {
        stream.write(chunk);
        return write(chunk, ...rest);
      };
    } catch (exc) {
      log("BOT: cannot open the log " + logPath + ": " + exc);
      process.exit(1);
    }
  }

  log("BOT: logging in as " + name + " to " + host + ":" + port);

  const client = mc.createClient({
    host: host,
    port: port,
    username: name,
    // The probe server runs in offline mode, so the session does not go
    // through the session server.
    auth: "offline",
    version: "1.12.2",
    // The bot only has to be a player that is online; it never moves and
    // never answers a chat message.
    keepAlive: true,
    closeTimeout: 0,
  });

  client.on("login", () => {
    log("BOT: logged in as " + name);
  });

  client.on("packet", (data, meta) => {
    if (process.env.BOT_DEBUG) {
      log("BOT: packet " + meta.name + " (" + meta.state + ")");
    }
  });

  client.on("end", (reason) => {
    log("BOT: disconnected: " + reason);
  });

  client.on("error", (err) => {
    log("BOT: error: " + err);
  });

  // The bot stays online until the deadline. The probe stops it when the
  // case is over.
  setTimeout(() => {
    log("BOT: done");
    try {
      client.end();
    } catch (exc) {
      // The client is already gone; nothing to do.
    }
    process.exit(0);
  }, seconds * 1000);
}

main();
