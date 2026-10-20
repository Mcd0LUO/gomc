package v1_12_2

import (
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/world"
)

// TestCoordinateBounds pins down the bounds 1.12.2 puts on a coordinate.
//
// The coordinate is read with `CommandAbstract.b(base, text, min, max,
// flag)`, and a coordinate outside the bounds is refused with
// `commands.generic.num.tooBig` / `tooSmall`. The number in the message is
// formatted with "%.2f".
//
// The bounds depend on what the coordinate is for:
//
//   - a block position: X and Z are limited to ±30000000, Y to [0, 256];
//   - a plain number (`/execute`, `/particle`, `/playsound`): all three are
//     limited to ±30000000;
//   - `/tp`: X and Z are limited to ±30000000, Y to ±4096.
//
// (Measured on a real 1.12.2 server.)
func TestCoordinateBounds(t *testing.T) {
	e := newTestEngine(t)
	for _, tc := range []struct {
		line string
		want string
	}{
		{"setblock 30000001 4 0 stone", "The number you have entered (30000001.00) is too big, it must be at most 30000000"},
		{"setblock -30000001 4 0 stone", "The number you have entered (-30000001.00) is too small, it must be at least -30000000"},
		{"setblock 0 257 0 stone", "The number you have entered (257.00) is too big, it must be at most 256"},
		{"setblock 0 -1 0 stone", "The number you have entered (-1.00) is too small, it must be at least 0"},
		{"setblock 1e20 4 0 stone", "The number you have entered (100000000000000000000.00) is too big, it must be at most 30000000"},
		// `/execute` reads a plain number, so its Y is not limited to
		// [0, 256].
		{"execute @e[type=armor_stand] ~ ~-1 ~ say down", ""},
	} {
		err := runLine(e, tc.line)
		if tc.want == "" {
			if err != nil {
				t.Errorf("%q: %v", tc.line, err)
			}
			continue
		}
		if err == nil {
			t.Errorf("%q: expected an error", tc.line)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%q: the error was %q, want it to contain %q", tc.line, err.Error(), tc.want)
		}
	}
}

// TestCenteredCoordinate pins down the "centre" mode.
//
// When the reader is in the centre mode, a whole number names the middle
// of a block, so 1.12.2 adds 0.5 to it -- and the message of a coordinate
// that is out of bounds shows the number with the 0.5.
//
// (Measured on a real 1.12.2 server: `summon zombie 30000001 4 0` answers
// "The number you have entered (30000001.50) is too big, it must be at most
// 30000000".)
func TestCenteredCoordinate(t *testing.T) {
	e := newTestEngine(t)
	err := runLine(e, "summon zombie 30000001 4 0")
	if err == nil {
		t.Fatal("expected an error")
	}
	want := "The number you have entered (30000001.50) is too big, it must be at most 30000000"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("the error was %q, want it to contain %q", err.Error(), want)
	}
}

// TestOutOfWorld pins down the "outside of the world" wording.
//
// 1.12.2 refuses a position whose chunk is not loaded. The server only
// loads the area around the world spawn and the areas around the players,
// so the wording depends on the command.
//
// (Measured on a real 1.12.2 server.)
func TestOutOfWorld(t *testing.T) {
	e := newTestEngine(t)
	for _, tc := range []struct {
		line string
		want string
	}{
		{"setblock 30000 4 0 stone", "Cannot place block outside of the world"},
		{"fill 30000 4 0 30001 4 1 stone", "Cannot place blocks outside of the world"},
		{"testforblock 30000 4 0 stone", "Cannot test for block outside of the world"},
		{"blockdata 30000 4 0 {a:1b}", "Cannot change block outside of the world"},
		{"summon zombie 30000 4 0", "Cannot summon the object out of the world"},
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
}

// TestSummonPlacesTheEntityInTheMiddle pins down where a summoned entity
// lands.
//
// 1.12.2 reads the position in the centre mode, so the entity lands in the
// middle of the block: the X and Z grow by 0.5, the Y does not.
//
// (`CommandSummon.b(x, text, true)` for the X and Z and
// `CommandSummon.b(y, text, false)` for the Y.)
func TestSummonPlacesTheEntityInTheMiddle(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon zombie 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	for _, ent := range e.World.EntityList() {
		if ent.Type != "zombie" {
			continue
		}
		if ent.Pos.X != 0.5 || ent.Pos.Y != 4 || ent.Pos.Z != 0.5 {
			t.Errorf("the entity is at %v, want (0.5, 4, 0.5)", ent.Pos)
		}
		return
	}
	t.Error("the zombie is not in the world")
}

// TestWorldHeight pins down what happens outside the height of the world.
//
// `/fill` refuses the whole region before it changes anything, while
// `/setblock` only fails when it places the block.
//
// (Measured on a real 1.12.2 server.)
func TestWorldHeight(t *testing.T) {
	e := newTestEngine(t)
	for _, tc := range []struct {
		line string
		want string
	}{
		{"setblock 0 256 0 stone", "The block couldn't be placed"},
		{"fill 0 256 0 1 256 1 stone", "Cannot place blocks outside of the world"},
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
	// A block inside the world is placed.
	if err := runLine(e, "fill 0 255 0 1 255 1 stone"); err != nil {
		t.Errorf("fill: %v", err)
	}
	block := e.World.Blocks.Get(world.Pos{X: 0, Y: 255, Z: 0})
	if block.Name != "stone" {
		t.Errorf("the block is %q, want stone", block.Name)
	}
}
