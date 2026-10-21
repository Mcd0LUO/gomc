package v1_12_2

import (
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/world"
)

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
