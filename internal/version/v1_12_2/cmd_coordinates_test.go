package v1_12_2

import (
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
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

// runLineAt runs a command line under a source at the given position.
func runLineAt(e *exec.Engine, line string, pos world.Vec) error {
	src := &exec.Source{
		Name: "Server", World: "world", PermissionLevel: 4, Feedback: true,
		Pos: pos,
	}
	_, err := e.ExecuteLine(line, 1, src)
	return err
}

// TestRelativeCoordinateBaseIsTheBlockPosition covers the base a relative
// coordinate is measured from.
//
// 1.12.2 reads a relative coordinate against
// `ICommandListener.getChunkCoordinates()`, and that is a **block
// position**: the console is at (0, 0, 0) and the command of an
// `/execute` at the block the execute position points at.
//
// So a fractional offset lands on a different block than it would if the
// base were the exact position: at x=7.5 the offset 2.6 gives 10.1 (the
// tenth block) with the exact base, and 7+2.6 = 9.6 (the ninth block)
// with the block position.
//
// (Measured on a real 1.12.2 server: the dirt lands at 9, 4, 4 — see the
// probe case 44-relative-base.)
func TestRelativeCoordinateBaseIsTheBlockPosition(t *testing.T) {
	e := newTestEngine(t)
	src := world.Vec{X: 7.5, Y: 4, Z: 4.5}
	if err := runLineAt(e, "setblock ~2.6 ~ ~ dirt", src); err != nil {
		t.Fatalf("setblock: %v", err)
	}
	if got := e.World.Blocks.Get(world.Pos{X: 9, Y: 4, Z: 4}); got.Name != "dirt" {
		t.Errorf("the block at 9, 4, 4 is %q, want dirt", got.Name)
	}
	if got := e.World.Blocks.Get(world.Pos{X: 10, Y: 4, Z: 4}); got.Name != "air" {
		t.Errorf("the block at 10, 4, 4 is %q, want air", got.Name)
	}
}

// TestRelativeCoordinateBoundsUseTheBase covers the bounds of a relative
// coordinate.
//
// 1.12.2 checks the bounds on the **sum** of the base and the offset, so
// the message of a coordinate that is out of bounds shows the sum. The
// base is the block position, so at the height 4.6 it is 4 and the sum is
// 304 — not 304.6.
//
// (Measured on a real 1.12.2 server: `setblock ~ ~300 ~ stone` at the
// height 4 answers "The number you have entered (304.00) is too big, it
// must be at most 256" — see the probe case 45-relative-bounds.)
func TestRelativeCoordinateBoundsUseTheBase(t *testing.T) {
	e := newTestEngine(t)
	err := runLineAt(e, "setblock ~ ~300 ~ stone", world.Vec{X: 0, Y: 4.6, Z: 0})
	if err == nil {
		t.Fatal("no error, want the coordinate to be refused")
	}
	want := "The number you have entered (304.00) is too big, it must be at most 256"
	if got := errMessage(err); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestExecuteSubCommandUsesTheBlockPosition covers the base inside an
// `/execute`.
//
// The command of an `/execute` runs at the block the execute position
// points at: 1.12.2 wraps the listener in a `CommandListenerWrapper`
// whose block position is the one of the execute position, and a relative
// coordinate of the sub-command is measured against it.
func TestExecuteSubCommandUsesTheBlockPosition(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon armor_stand 7.5 4.6 4.5"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	if err := runLine(e, "execute @e[type=armor_stand] ~ ~ ~ setblock ~2.6 ~ ~ dirt"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got := e.World.Blocks.Get(world.Pos{X: 9, Y: 4, Z: 4}); got.Name != "dirt" {
		t.Errorf("the block at 9, 4, 4 is %q, want dirt", got.Name)
	}
	if got := e.World.Blocks.Get(world.Pos{X: 10, Y: 4, Z: 4}); got.Name != "air" {
		t.Errorf("the block at 10, 4, 4 is %q, want air", got.Name)
	}
}

// TestAbsoluteCoordinatesDoNotUseTheBase covers that an absolute
// coordinate is not measured from anything.
//
// 1.12.2 starts the sum at zero for a coordinate that is not relative, so
// the base plays no part in it.
func TestAbsoluteCoordinatesDoNotUseTheBase(t *testing.T) {
	e := newTestEngine(t)
	if err := runLineAt(e, "setblock 9 4 4 dirt", world.Vec{X: 7.5, Y: 4, Z: 4.5}); err != nil {
		t.Fatalf("setblock: %v", err)
	}
	if got := e.World.Blocks.Get(world.Pos{X: 9, Y: 4, Z: 4}); got.Name != "dirt" {
		t.Errorf("the block at 9, 4, 4 is %q, want dirt", got.Name)
	}
}

// TestPlayerBaseIsHalfABlockUp covers the base of a player.
//
// 1.12.2 measures a relative coordinate against the "chunk coordinates"
// of the listener, and a player's are
// `new BlockPosition(locX, locY + 0.5, locZ)` — the Y is half a block up,
// because that is the block the body of the player is in.
//
// So at the height 4.6 the base is 5 and the offset 0.6 lands on the sixth
// block; the plain block position 4 would land on the fifth.
func TestPlayerBaseIsHalfABlockUp(t *testing.T) {
	e := newTestEngine(t)
	p, ok := e.World.EntityTypes.NewEntity("player", "p1")
	if !ok {
		t.Fatal("the player entity type is missing")
	}
	p.PlayerName = "Steve"
	p.Pos = world.Vec{X: 7.5, Y: 4.6, Z: 4.5}
	e.World.AddEntity(p)

	src := &exec.Source{
		Name: "Steve", World: "world", PermissionLevel: 4, Feedback: true,
		Entity: p, Pos: p.Pos, ListenerPlayer: true,
	}
	if _, err := e.ExecuteLine("setblock ~ ~0.6 ~ stone", 1, src); err != nil {
		t.Fatalf("setblock: %v", err)
	}
	if got := e.World.Blocks.Get(world.Pos{X: 7, Y: 5, Z: 4}); got.Name != "stone" {
		t.Errorf("the block at 7, 5, 4 is %q, want stone", got.Name)
	}
	if got := e.World.Blocks.Get(world.Pos{X: 7, Y: 4, Z: 4}); got.Name != "air" {
		t.Errorf("the block at 7, 4, 4 is %q, want air", got.Name)
	}
}

// TestExecuteNeedsItsCoordinates covers the argument count of `/execute`.
//
// 1.12.2 counts the arguments before it looks at any of them:
// `CommandExecute.execute` starts with `if (astring.length < 5)`, so the
// entity, the three coordinates and the command are all required.
func TestExecuteNeedsItsCoordinates(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	err := runLine(e, "execute @e[type=armor_stand] spawnpoint")
	if err == nil {
		t.Fatal("no error, want the usage error")
	}
	want := "Usage: /execute <entity> <x> <y> <z> <command> OR /execute <entity> <x> <y> <z> detect <x> <y> <z> <block> <dataValue|-1|state|*> <command>"
	if got := errMessage(err); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestExecuteSubCommandHasNoPlayerListener covers the commands of an
// `/execute` that need a player.
//
// The command of an `/execute` runs *as* the target entity, but the
// listener is the wrapper of the execute — and that is not a player. So
// `/spawnpoint` and `/setworldspawn` without a position are refused, the
// same way they are from the console.
func TestExecuteSubCommandHasNoPlayerListener(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	for _, line := range []string{
		"execute @e[type=armor_stand] ~ ~ ~ spawnpoint",
		"execute @e[type=armor_stand] ~ ~ ~ setworldspawn",
	} {
		// The failure of a sub-command belongs to the entity, so the
		// execute itself succeeds and the failure is collected.
		if err := runLine(e, line); err != nil {
			t.Errorf("%q: %v", line, err)
		}
		errs := e.TakeErrors()
		if len(errs) == 0 {
			t.Errorf("%q: no error, want the refusal of the player", line)
			continue
		}
		want := "You must specify which player you wish to perform this action on."
		if got := errs[0].Error(); got != want {
			t.Errorf("%q: %q, want %q", line, got, want)
		}
	}
}

// TestSummonWithoutCoordinatesUsesTheListener covers the position of a
// `/summon` without coordinates.
//
// 1.12.2 takes it from the chunk coordinates of the listener, so the entity
// appears where the command runs — not at the origin.
func TestSummonWithoutCoordinatesUsesTheListener(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon armor_stand 7.5 4.6 4.5"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	if err := runLine(e, "execute @e[type=armor_stand] ~ ~ ~ summon zombie"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	var zombie *world.Entity
	for _, ent := range e.World.EntityList() {
		if ent.Type == "zombie" {
			zombie = ent
		}
	}
	if zombie == nil {
		t.Fatal("the zombie was not summoned")
	}
	want := world.Vec{X: 7, Y: 4, Z: 4}
	if zombie.Pos != want {
		t.Errorf("the zombie is at %v, want %v", zombie.Pos, want)
	}
}
