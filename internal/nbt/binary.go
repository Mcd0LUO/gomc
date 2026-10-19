package nbt

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Encode writes a compound in the 1.12.2 binary NBT format.
//
// The layout is: a TAG_Compound root with an empty name, the payload, then
// TAG_End. This is the form used by entity NBT files and by `/summon` NBT.
func Encode(w io.Writer, c *Compound) error {
	if c == nil {
		c = NewCompound()
	}
	if err := writeHeader(w, TypeCompound, ""); err != nil {
		return err
	}
	// writePayload terminates the compound's payload with TAG_End.
	return writePayload(w, c)
}

// Decode reads a compound from the 1.12.2 binary NBT format.
func Decode(r io.Reader) (*Compound, error) {
	typ, name, err := readHeader(r)
	if err != nil {
		return nil, err
	}
	if typ != TypeCompound {
		return nil, fmt.Errorf("nbt: root tag must be a compound, got %s", typ)
	}
	_ = name
	// readCompoundPayload consumes the payload and its terminating TAG_End.
	return readCompoundPayload(r)
}

func writeHeader(w io.Writer, typ Type, name string) error {
	if err := writeByte(w, byte(typ)); err != nil {
		return err
	}
	return writeString(w, name)
}

func writePayload(w io.Writer, v Value) error {
	switch t := v.(type) {
	case Byte:
		return writeByte(w, byte(t))
	case Short:
		return writeShort(w, int16(t))
	case Int:
		return writeInt(w, int32(t))
	case Long:
		return writeLong(w, int64(t))
	case Float:
		return writeInt(w, int32(math.Float32bits(float32(t))))
	case Double:
		return writeLong(w, int64(math.Float64bits(float64(t))))
	case String:
		return writeString(w, string(t))
	case ByteArray:
		if err := writeInt(w, int32(len(t))); err != nil {
			return err
		}
		_, err := w.Write(t)
		return err
	case IntArray:
		if err := writeInt(w, int32(len(t))); err != nil {
			return err
		}
		for _, n := range t {
			if err := writeInt(w, n); err != nil {
				return err
			}
		}
		return nil
	case *List:
		if err := writeByte(w, byte(t.elemType)); err != nil {
			return err
		}
		if err := writeInt(w, int32(len(t.items))); err != nil {
			return err
		}
		for _, item := range t.items {
			if err := writePayload(w, item); err != nil {
				return err
			}
		}
		return nil
	case *Compound:
		for _, k := range t.keys {
			if err := writeHeader(w, t.values[k].Type(), k); err != nil {
				return err
			}
			if err := writePayload(w, t.values[k]); err != nil {
				return err
			}
		}
		// A compound's payload is terminated by TAG_End, nested ones
		// included.
		return writeByte(w, byte(TypeEnd))
	}
	return fmt.Errorf("nbt: cannot encode %T", v)
}

func readHeader(r io.Reader) (Type, string, error) {
	b, err := readByte(r)
	if err != nil {
		return TypeEnd, "", err
	}
	name, err := readString(r)
	if err != nil {
		return TypeEnd, "", err
	}
	return Type(b), name, nil
}

func readCompoundPayload(r io.Reader) (*Compound, error) {
	c := NewCompound()
	for {
		// The tag type is read on its own: TAG_End carries no name, so
		// reading the name before checking the type would desynchronise
		// the stream.
		raw, err := readByte(r)
		if err != nil {
			return nil, err
		}
		typ := Type(raw)
		if typ == TypeEnd {
			return c, nil
		}
		name, err := readString(r)
		if err != nil {
			return nil, err
		}
		v, err := readPayload(r, typ)
		if err != nil {
			return nil, err
		}
		c.Set(name, v)
	}
}

func readPayload(r io.Reader, typ Type) (Value, error) {
	switch typ {
	case TypeByte:
		b, err := readByte(r)
		return Byte(int8(b)), err
	case TypeShort:
		n, err := readShort(r)
		return Short(n), err
	case TypeInt:
		n, err := readInt(r)
		return Int(n), err
	case TypeLong:
		n, err := readLong(r)
		return Long(n), err
	case TypeFloat:
		n, err := readInt(r)
		return Float(math.Float32frombits(uint32(n))), err
	case TypeDouble:
		n, err := readLong(r)
		return Double(math.Float64frombits(uint64(n))), err
	case TypeString:
		s, err := readString(r)
		return String(s), err
	case TypeByteArray:
		n, err := readInt(r)
		if err != nil {
			return nil, err
		}
		if n < 0 {
			return nil, fmt.Errorf("nbt: negative byte array length %d", n)
		}
		buf := make([]byte, n)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		return ByteArray(buf), nil
	case TypeIntArray:
		n, err := readInt(r)
		if err != nil {
			return nil, err
		}
		if n < 0 {
			return nil, fmt.Errorf("nbt: negative int array length %d", n)
		}
		out := make(IntArray, n)
		for i := range out {
			v, err := readInt(r)
			if err != nil {
				return nil, err
			}
			out[i] = v
		}
		return out, nil
	case TypeList:
		elemType, err := readByte(r)
		if err != nil {
			return nil, err
		}
		n, err := readInt(r)
		if err != nil {
			return nil, err
		}
		if n < 0 {
			return nil, fmt.Errorf("nbt: negative list length %d", n)
		}
		list := NewList(Type(elemType))
		for i := int32(0); i < n; i++ {
			v, err := readPayload(r, Type(elemType))
			if err != nil {
				return nil, err
			}
			list.items = append(list.items, v)
		}
		return list, nil
	case TypeCompound:
		return readCompoundPayload(r)
	}
	return nil, fmt.Errorf("nbt: unknown tag type %d", typ)
}

func writeByte(w io.Writer, b byte) error {
	_, err := w.Write([]byte{b})
	return err
}

func writeShort(w io.Writer, n int16) error {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], uint16(n))
	_, err := w.Write(buf[:])
	return err
}

func writeInt(w io.Writer, n int32) error {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(n))
	_, err := w.Write(buf[:])
	return err
}

func writeLong(w io.Writer, n int64) error {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(n))
	_, err := w.Write(buf[:])
	return err
}

func writeString(w io.Writer, s string) error {
	if err := writeShort(w, int16(len(s))); err != nil {
		return err
	}
	_, err := io.WriteString(w, s)
	return err
}

func readByte(r io.Reader) (byte, error) {
	var buf [1]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func readShort(r io.Reader) (int16, error) {
	var buf [2]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return int16(binary.BigEndian.Uint16(buf[:])), nil
}

func readInt(r io.Reader) (int32, error) {
	var buf [4]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(buf[:])), nil
}

func readLong(r io.Reader) (int64, error) {
	var buf [8]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(buf[:])), nil
}

func readString(r io.Reader) (string, error) {
	n, err := readShort(r)
	if err != nil {
		return "", err
	}
	if n < 0 {
		return "", fmt.Errorf("nbt: negative string length %d", n)
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}
