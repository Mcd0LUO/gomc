package v1_12_2

import (
	"strings"
	"testing"
)

// TestUnknownBlockNameIsReported pins down what 1.12.2 answers when a
// command names a block the server does not know.
//
// The name is read as a `MinecraftKey` and looked up in the block
// registry, and a name that is not in there is reported as
// "There is no such block with name minecraft:<name>".
//
// (Measured on a real 1.12.2 server.)
func TestUnknownBlockNameIsReported(t *testing.T) {
	e := newTestEngine(t)
	const want = "There is no such block with name minecraft:notablock"
	for _, line := range []string{
		"setblock 0 4 0 notablock",
		"fill 0 4 0 1 4 1 notablock",
		"testforblock 0 4 0 notablock",
		"clone 0 4 0 1 4 1 2 4 2 filtered force notablock",
		"execute @e ~ ~ ~ detect ~ ~ ~ notablock 0 say hi",
	} {
		err := runLine(e, line)
		if err == nil {
			t.Errorf("%q: expected an error", line)
			continue
		}
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%q: the error was %q, want it to contain %q", line, err.Error(), want)
		}
	}
	// A name of another namespace is reported the same way, with its own
	// namespace.
	err := runLine(e, "setblock 0 4 0 foo:stone")
	if err == nil {
		t.Fatal("expected an error for a block of another namespace")
	}
	if want := "There is no such block with name foo:stone"; !strings.Contains(err.Error(), want) {
		t.Errorf("the error was %q, want it to contain %q", err.Error(), want)
	}
}

// TestKnownBlockNamesAreAccepted pins down the names 1.12.2 knows.
//
// 1.12.2 keeps the naming of the older versions next to the one of 1.12,
// so "stonebrick" (the old name of the cobblestone) works as well as
// "cobblestone". The variants of a block are data values of it, not
// blocks of their own, so "granite" is not a block.
//
// (Measured on a real 1.12.2 server.)
func TestKnownBlockNamesAreAccepted(t *testing.T) {
	e := newTestEngine(t)
	for _, line := range []string{
		"setblock 0 4 0 stone",
		"setblock 0 4 0 minecraft:stone",
		"setblock 0 4 0 STONE",
		"setblock 0 4 0 stonebrick",
		"setblock 0 4 0 cobblestone",
		"setblock 0 4 0 :stone",
		"fill 0 4 0 1 4 1 stone",
		"testforblock 0 4 0 stone",
	} {
		if err := runLine(e, line); err != nil {
			t.Errorf("%q: %v", line, err)
		}
	}
	for _, line := range []string{
		// A variant is a data value, not a block.
		"setblock 0 4 0 granite",
		"setblock 0 4 0 white_wool",
		"setblock 0 4 0 polished_granite",
		// The naming of 1.13 does not exist in 1.12.2.
		"setblock 0 4 0 oak_planks",
		"setblock 0 4 0 grass_block",
	} {
		if err := runLine(e, line); err == nil {
			t.Errorf("%q: expected an error", line)
		}
	}
}
