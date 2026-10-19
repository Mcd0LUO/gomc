package nbt

import (
	"bytes"
	"testing"
)

func TestParseSNBTTypes(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"{a:1b}", "{a:1b}"},
		{"{a:1s}", "{a:1s}"},
		{"{a:1}", "{a:1}"},
		{"{a:1L}", "{a:1L}"},
		{"{a:1.5f}", "{a:1.5f}"},
		{"{a:1.5d}", "{a:1.5d}"},
		{`{a:"text"}`, `{a:"text"}`},
		{"{a:{b:2b}}", "{a:{b:2b}}"},
		{"{a:[1b,2b]}", "{a:[1b,2b]}"},
	}
	for _, c := range cases {
		v, err := ParseSNBT(c.in)
		if err != nil {
			t.Fatalf("ParseSNBT(%q): %v", c.in, err)
		}
		if got := v.String(); got != c.want {
			t.Errorf("ParseSNBT(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseSNBTUnquotedString(t *testing.T) {
	// A bare word stays a string, matching 1.12.2's parser.
	v, err := ParseSNBT("{CustomName:Steve}")
	if err != nil {
		t.Fatalf("ParseSNBT: %v", err)
	}
	c := v.(*Compound)
	got := c.Get("CustomName")
	if got == nil {
		t.Fatal("missing CustomName")
	}
	if _, ok := got.(String); !ok {
		t.Errorf("CustomName is %T, want String", got)
	}
}

func TestParseSNBTArrayTypes(t *testing.T) {
	v, err := ParseSNBT("{a:[B;1b,2b],b:[I;1,2]}")
	if err != nil {
		t.Fatalf("ParseSNBT: %v", err)
	}
	c := v.(*Compound)
	ba, ok := c.Get("a").(ByteArray)
	if !ok {
		t.Fatalf("a is %T, want ByteArray", c.Get("a"))
	}
	if len(ba) != 2 || ba[0] != 1 || ba[1] != 2 {
		t.Errorf("a = %v, want [1 2]", ba)
	}
	ia, ok := c.Get("b").(IntArray)
	if !ok {
		t.Fatalf("b is %T, want IntArray", c.Get("b"))
	}
	if len(ia) != 2 || ia[0] != 1 || ia[1] != 2 {
		t.Errorf("b = %v, want [1 2]", ia)
	}
}

func TestMatchesSubsetSemantics(t *testing.T) {
	subject, err := ParseSNBT("{Health:20.0f,NoAI:1b,CustomName:\"Bob\"}")
	if err != nil {
		t.Fatalf("ParseSNBT: %v", err)
	}
	// A filter with fewer keys must still match: this is the subset rule
	// 1.12.2 uses for selector `nbt=` and for /testfor.
	filter, err := ParseSNBT("{NoAI:1b}")
	if err != nil {
		t.Fatalf("ParseSNBT: %v", err)
	}
	if !Matches(subject, filter) {
		t.Error("expected {NoAI:1b} to match a compound that contains it")
	}
	// A different value must not match.
	filter, _ = ParseSNBT("{NoAI:0b}")
	if Matches(subject, filter) {
		t.Error("expected {NoAI:0b} not to match")
	}
	// A missing key must not match.
	filter, _ = ParseSNBT("{Missing:1b}")
	if Matches(subject, filter) {
		t.Error("expected a missing key not to match")
	}
}

func TestMatchesNumericWidths(t *testing.T) {
	subject, _ := ParseSNBT("{a:1}")
	filter, _ := ParseSNBT("{a:1b}")
	if !Matches(subject, filter) {
		t.Error("expected numeric widths to be ignored when matching")
	}
}

func TestMatchesListLength(t *testing.T) {
	subject, _ := ParseSNBT("{a:[1b,2b]}")
	same, _ := ParseSNBT("{a:[1b,2b]}")
	if !Matches(subject, same) {
		t.Error("expected identical lists to match")
	}
	shorter, _ := ParseSNBT("{a:[1b]}")
	if Matches(subject, shorter) {
		t.Error("expected a shorter list not to match")
	}
}

func TestBinaryRoundTrip(t *testing.T) {
	src, err := ParseSNBT("{id:\"minecraft:stone\",Count:1b,tag:{x:1s,list:[1b,2b]}}")
	if err != nil {
		t.Fatalf("ParseSNBT: %v", err)
	}
	comp := src.(*Compound)
	var buf bytes.Buffer
	if err := Encode(&buf, comp); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := Decode(&buf)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got.String() != comp.String() {
		t.Errorf("round trip = %q, want %q", got.String(), comp.String())
	}
}
