package v1_12_2

import (
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/world"
)

// TestFillDataValueIsStrict pins down what 1.12.2 does with the eighth
// argument of `/fill`, the data value.
//
// The slot is read strictly: as soon as there is an argument after the
// block, that argument is the data value, and only a number is a data
// value. A word like "replace" is therefore a broken data value and not
// a mode, and the mode stays the ninth argument.
//
// The wording is the one of the language file:
// `commands.generic.blockstate.invalid` for a text that is not a number,
// `commands.generic.num.tooSmall` and `commands.generic.num.tooBig` for
// a number outside 0..15.
//
// (Measured on a real 1.12.2 server.)
func TestFillDataValueIsStrict(t *testing.T) {
	e := newTestEngine(t)
	for _, tc := range []struct {
		line string
		want string
	}{
		{"fill 0 4 0 1 4 1 stone {Tags:", "'{Tags:' is not a state for block minecraft:stone"},
		{"fill 0 4 0 1 4 1 stone abc", "'abc' is not a state for block minecraft:stone"},
		{"fill 0 4 0 1 4 1 stone 1.5", "'1.5' is not a state for block minecraft:stone"},
		{"fill 0 4 0 1 4 1 stone replace", "'replace' is not a state for block minecraft:stone"},
		{"fill 0 4 0 1 4 1 stone keep", "'keep' is not a state for block minecraft:stone"},
		{"fill 0 4 0 1 4 1 stone -1", "The number you have entered (-1) is too small, it must be at least 0"},
		{"fill 0 4 0 1 4 1 stone 16", "The number you have entered (16) is too big, it must be at most 15"},
	} {
		err := runLine(e, tc.line)
		if err == nil {
			t.Errorf("%q: expected an error", tc.line)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%q: the error was %q, want it to contain %q", tc.line, err.Error(), tc.want)
		}
	}
	// The data value of the block that is placed is the eighth argument,
	// so the mode is the ninth one. Every line fills its own region,
	// because the keep mode only fills what is empty.
	for _, line := range []string{
		"fill 0 4 0 1 4 1 stone",
		"fill 2 4 2 3 4 3 stone 0",
		"fill 4 4 4 5 4 5 stone 0 keep",
		"fill 6 4 6 7 4 7 stone 15 outline",
		"fill 8 4 8 9 4 9 stone 007",
		"fill 10 4 10 11 4 11 stone +1",
		"fill 12 4 12 13 4 13 stone default",
	} {
		if err := runLine(e, line); err != nil {
			t.Errorf("%q: %v", line, err)
		}
	}
}

// TestFillDataValueSetsTheBlock pins down that the data value of the
// eighth argument is the data value of the block that is placed.
//
// (Measured on a real 1.12.2 server: a fill with the data value 1 puts a
// stone with the data value 1 there, which a following
// `fill ... stone 0 replace stone 1` replaces, and `fill ... stone
// default` puts the default state of the block there.)
func TestFillDataValueSetsTheBlock(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "fill 0 4 0 1 4 1 stone 3"); err != nil {
		t.Fatalf("fill: %v", err)
	}
	got := e.World.Blocks.Get(world.Pos{X: 0, Y: 4, Z: 0})
	if got.Name != "stone" || got.Data != 3 {
		t.Errorf("the block is %s with the data value %d, want stone with 3", got.Name, got.Data)
	}
	// Without a data value the block gets the default one.
	if err := runLine(e, "fill 2 4 2 3 4 3 stone"); err != nil {
		t.Fatalf("fill: %v", err)
	}
	got = e.World.Blocks.Get(world.Pos{X: 2, Y: 4, Z: 2})
	if got.Name != "stone" || got.Data != 0 {
		t.Errorf("the block is %s with the data value %d, want stone with 0", got.Name, got.Data)
	}
}

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
