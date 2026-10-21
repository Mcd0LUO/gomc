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
