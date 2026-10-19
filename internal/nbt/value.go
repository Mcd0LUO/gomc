// Package nbt provides the NBT value model used by Minecraft 1.12.2.
//
// The model mirrors the 1.12.2 tag set exactly: the numeric types keep their
// distinct widths (Byte/Short/Int/Long/Float/Double) because NBT is a
// self-describing binary format and widening a value changes its encoded
// size and equality semantics.
package nbt

import (
	"fmt"
	"strconv"
)

// Type identifies a NBT tag type.
type Type uint8

// The NBT tag types, in their canonical numeric order.
const (
	TypeEnd Type = iota
	TypeByte
	TypeShort
	TypeInt
	TypeLong
	TypeFloat
	TypeDouble
	TypeByteArray
	TypeString
	TypeList
	TypeCompound
	TypeIntArray
)

var typeNames = map[Type]string{
	TypeEnd:       "TAG_End",
	TypeByte:      "TAG_Byte",
	TypeShort:     "TAG_Short",
	TypeInt:       "TAG_Int",
	TypeLong:      "TAG_Long",
	TypeFloat:     "TAG_Float",
	TypeDouble:    "TAG_Double",
	TypeByteArray: "TAG_Byte_Array",
	TypeString:    "TAG_String",
	TypeList:      "TAG_List",
	TypeCompound:  "TAG_Compound",
	TypeIntArray:  "TAG_Int_Array",
}

// String returns the canonical name of the tag type.
func (t Type) String() string {
	if n, ok := typeNames[t]; ok {
		return n
	}
	return "TAG_Unknown(" + strconv.Itoa(int(t)) + ")"
}

// Value is a single NBT value. Implementations are immutable in practice;
// mutation happens through Compound and List helpers.
type Value interface {
	// Type returns the tag type.
	Type() Type
	// String renders the value in SNBT form.
	String() string
}

// Byte is a signed 8-bit integer.
type Byte int8

// Type implements Value.
func (b Byte) Type() Type { return TypeByte }

// String implements Value.
func (b Byte) String() string { return strconv.FormatInt(int64(b), 10) + "b" }

// Short is a signed 16-bit integer.
type Short int16

// Type implements Value.
func (s Short) Type() Type { return TypeShort }

// String implements Value.
func (s Short) String() string { return strconv.FormatInt(int64(s), 10) + "s" }

// Int is a signed 32-bit integer.
type Int int32

// Type implements Value.
func (i Int) Type() Type { return TypeInt }

// String implements Value.
func (i Int) String() string { return strconv.FormatInt(int64(i), 10) }

// Long is a signed 64-bit integer.
type Long int64

// Type implements Value.
func (l Long) Type() Type { return TypeLong }

// String implements Value.
func (l Long) String() string { return strconv.FormatInt(int64(l), 10) + "L" }

// Float is a 32-bit floating point value.
type Float float32

// Type implements Value.
func (f Float) Type() Type { return TypeFloat }

// String implements Value.
func (f Float) String() string { return strconv.FormatFloat(float64(f), 'g', -1, 32) + "f" }

// Double is a 64-bit floating point value.
type Double float64

// Type implements Value.
func (d Double) Type() Type { return TypeDouble }

// String implements Value.
func (d Double) String() string { return strconv.FormatFloat(float64(d), 'g', -1, 64) + "d" }

// String is a UTF-8 text value.
type String string

// Type implements Value.
func (s String) Type() Type { return TypeString }

// String implements Value.
func (s String) String() string { return quote(string(s)) }

// ByteArray is a length-prefixed array of bytes.
type ByteArray []byte

// Type implements Value.
func (b ByteArray) Type() Type { return TypeByteArray }

// String implements Value.
func (b ByteArray) String() string {
	return fmt.Sprintf("[B;%s]", joinBytes(b))
}

// IntArray is a length-prefixed array of 32-bit integers.
type IntArray []int32

// Type implements Value.
func (i IntArray) Type() Type { return TypeIntArray }

// String implements Value.
func (i IntArray) String() string {
	parts := make([]string, len(i))
	for idx, v := range i {
		parts[idx] = strconv.FormatInt(int64(v), 10)
	}
	return "[" + joinStrings(parts) + "I;]"
}

// List is an ordered sequence of values of a single type.
type List struct {
	elemType Type
	items    []Value
}

// NewList returns an empty list of the given element type.
func NewList(elemType Type) *List {
	return &List{elemType: elemType, items: []Value{}}
}

// ElemType returns the declared element type of the list.
func (l *List) ElemType() Type { return l.elemType }

// Len returns the number of elements.
func (l *List) Len() int { return len(l.items) }

// At returns the element at index i, or nil when out of range.
func (l *List) At(i int) Value {
	if i < 0 || i >= len(l.items) {
		return nil
	}
	return l.items[i]
}

// Items returns the backing slice. The caller must not mutate it.
func (l *List) Items() []Value { return l.items }

// Append adds a value to the list, converting it to the list's element type
// when the types differ but are numerically compatible.
func (l *List) Append(v Value) {
	if v == nil {
		return
	}
	l.items = append(l.items, coerce(v, l.elemType))
}

// Type implements Value.
func (l *List) Type() Type { return TypeList }

// String implements Value.
func (l *List) String() string {
	parts := make([]string, len(l.items))
	for i, v := range l.items {
		parts[i] = v.String()
	}
	return "[" + joinStrings(parts) + "]"
}

// Compound is an ordered map of named values.
type Compound struct {
	keys   []string
	values map[string]Value
}

// NewCompound returns an empty compound.
func NewCompound() *Compound {
	return &Compound{values: map[string]Value{}}
}

// Type implements Value.
func (c *Compound) Type() Type { return TypeCompound }

// Len returns the number of entries.
func (c *Compound) Len() int { return len(c.keys) }

// Keys returns the entry names in insertion order.
func (c *Compound) Keys() []string { return c.keys }

// Get returns the value stored under key, or nil when absent.
func (c *Compound) Get(key string) Value {
	if c.values == nil {
		return nil
	}
	return c.values[key]
}

// Has reports whether the compound contains key.
func (c *Compound) Has(key string) bool {
	if c.values == nil {
		return false
	}
	_, ok := c.values[key]
	return ok
}

// Set stores a value under key, preserving first-insertion ordering.
func (c *Compound) Set(key string, v Value) {
	if c.values == nil {
		c.values = map[string]Value{}
	}
	if _, ok := c.values[key]; !ok {
		c.keys = append(c.keys, key)
	}
	c.values[key] = v
}

// Remove deletes key from the compound.
func (c *Compound) Remove(key string) {
	if c.values == nil {
		return
	}
	if _, ok := c.values[key]; !ok {
		return
	}
	delete(c.values, key)
	for i, k := range c.keys {
		if k == key {
			c.keys = append(c.keys[:i], c.keys[i+1:]...)
			break
		}
	}
}

// Clone returns a deep copy of the compound.
func (c *Compound) Clone() *Compound {
	out := NewCompound()
	for _, k := range c.keys {
		out.Set(k, CloneValue(c.values[k]))
	}
	return out
}

// String implements Value.
func (c *Compound) String() string {
	parts := make([]string, 0, len(c.keys))
	for _, k := range c.keys {
		parts = append(parts, quoteKey(k)+":"+c.values[k].String())
	}
	return "{" + joinStrings(parts) + "}"
}

// quoteKey renders a compound key.
//
// 1.12.2's SNBT writer emits keys unquoted unless they contain characters
// that would break the syntax, so that output round-trips through its own
// parser unchanged.
func quoteKey(s string) string {
	if s == "" {
		return `""`
	}
	for _, r := range s {
		switch r {
		case ' ', '\t', '\n', '\r', ':', ',', '{', '}', '[', ']', '"', '\'':
			return quote(s)
		}
	}
	return s
}

// CloneValue returns a deep copy of v. Scalars are returned as-is.
func CloneValue(v Value) Value {
	switch t := v.(type) {
	case *Compound:
		return t.Clone()
	case *List:
		out := NewList(t.elemType)
		for _, item := range t.items {
			out.items = append(out.items, CloneValue(item))
		}
		return out
	case ByteArray:
		cp := make(ByteArray, len(t))
		copy(cp, t)
		return cp
	case IntArray:
		cp := make(IntArray, len(t))
		copy(cp, t)
		return cp
	default:
		return v
	}
}

// coerce converts v to the target numeric type where that is lossless in
// spirit (Java's NBT numeric widening). Non-numeric values are returned
// unchanged so callers can detect the mismatch.
func coerce(v Value, target Type) Value {
	if v == nil || v.Type() == target {
		return v
	}
	switch target {
	case TypeByte:
		if n, ok := asInt64(v); ok {
			return Byte(int8(n))
		}
	case TypeShort:
		if n, ok := asInt64(v); ok {
			return Short(int16(n))
		}
	case TypeInt:
		if n, ok := asInt64(v); ok {
			return Int(int32(n))
		}
	case TypeLong:
		if n, ok := asInt64(v); ok {
			return Long(n)
		}
	case TypeFloat:
		if f, ok := asFloat64(v); ok {
			return Float(float32(f))
		}
	case TypeDouble:
		if f, ok := asFloat64(v); ok {
			return Double(f)
		}
	}
	return v
}

func asInt64(v Value) (int64, bool) {
	switch t := v.(type) {
	case Byte:
		return int64(t), true
	case Short:
		return int64(t), true
	case Int:
		return int64(t), true
	case Long:
		return int64(t), true
	}
	return 0, false
}

func asFloat64(v Value) (float64, bool) {
	switch t := v.(type) {
	case Float:
		return float64(t), true
	case Double:
		return float64(t), true
	case Byte:
		return float64(t), true
	case Short:
		return float64(t), true
	case Int:
		return float64(t), true
	case Long:
		return float64(t), true
	}
	return 0, false
}

func joinStrings(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ","
		}
		out += p
	}
	return out
}

func joinBytes(b []byte) string {
	out := ""
	for i, v := range b {
		if i > 0 {
			out += "B;"
		}
		out += strconv.Itoa(int(v))
	}
	return out
}
