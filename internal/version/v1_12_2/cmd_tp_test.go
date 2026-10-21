package v1_12_2

import (
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// TestDetectAcceptsTheStarDataValue covers the data value of a `detect`.
//
// 1.12.2 reads it with `CommandAbstract.b(Block, String)` -- the predicate
// of a block -- so `*` and `-1` match every data value, a number is
// compared with the data value of the block, and anything else is a state
// string the block has to know.
//
// (Measured on a real 1.12.2 server: `execute @e[type=armor_stand] ~ ~ ~
// detect ~ ~ ~ air * say detected` prints "detected".)
func TestDetectAcceptsTheStarDataValue(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	feedback, err := runLineFeedback(e, "execute @e[type=armor_stand] ~ ~ ~ detect ~ ~ ~ air * say detected")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(feedback) != 1 || feedback[0] != "[Armor Stand] detected" {
		t.Errorf("feedback = %v, want the detect to succeed", feedback)
	}
}

// TestDetectComparesTheDataValue covers that a number is still compared
// with the data value of the block.
func TestDetectComparesTheDataValue(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	// The armor stand is in the air, so the block at its feet is not
	// stone and the detect has to fail.
	feedback, err := runLineFeedback(e, "execute @e[type=armor_stand] ~ ~ ~ detect ~ ~ ~ stone 0 say wrong")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(feedback) != 0 {
		t.Errorf("feedback = %v, want the detect to fail", feedback)
	}
}

// TestTpToAPlayer covers the second form of `/tp`.
//
// 1.12.2 decides the form of the command by the number of arguments: one
// or two of them mean "to an entity". The destination is the last
// argument, and the target is the first one.
//
// (Measured on a real 1.12.2 server: `tp @s @p[tag=target]` teleports to
// the player, and the message is "Teleported <a> to <b>" without the
// coordinates.)
func TestTpToAPlayer(t *testing.T) {
	e := newTestEngine(t)
	addPlayerAt(t, e, "Steve", world.Vec{X: 1, Y: 4, Z: 1})
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	feedback, err := runLineFeedback(e, "tp @e[type=armor_stand] Steve")
	if err != nil {
		t.Fatalf("tp: %v", err)
	}
	if len(feedback) != 1 || feedback[0] != "Teleported Armor Stand to Steve" {
		t.Errorf("feedback = %v, want the teleport to the player", feedback)
	}
	for _, ent := range e.World.EntityList() {
		if ent.Type == "armor_stand" && ent.Pos != (world.Vec{X: 1, Y: 4, Z: 1}) {
			t.Errorf("the armor stand is at %v, want the position of the player", ent.Pos)
		}
	}
}

// TestTpWithoutATarget covers the form of `/tp` that has no target.
//
// 1.12.2 teleports the player who runs the command when the first
// argument is not a target, which is the case when there are three
// arguments (the coordinates).
func TestTpWithoutATarget(t *testing.T) {
	e := newTestEngine(t)
	addPlayerAt(t, e, "Steve", world.Vec{X: 1, Y: 4, Z: 1})
	src := &exec.Source{
		Name: "Steve", World: "world", PermissionLevel: 4, Feedback: true,
		Entity: e.World.FindPlayer("Steve"), Pos: world.Vec{X: 1, Y: 4, Z: 1},
		ListenerPlayer: true,
	}
	if _, err := e.ExecuteLine("tp 5 6 7", 1, src); err != nil {
		t.Fatalf("tp: %v", err)
	}
	p := e.World.FindPlayer("Steve")
	if p == nil {
		t.Fatal("the player is gone")
	}
	// The X and the Z are in the "centre" mode: a whole number names
	// the middle of a block, so the player ends up at 5.5 and 7.5.
	if p.Pos != (world.Vec{X: 5.5, Y: 6, Z: 7.5}) {
		t.Errorf("the player is at %v, want (5.5, 6, 7.5)", p.Pos)
	}
}

// TestTellrawTakesTheRestOfTheLine covers the JSON argument of `/tellraw`.
//
// 1.12.2 hands the arguments after the target to the JSON parser as one
// string, so a JSON with spaces in it is one argument.
func TestTellrawTakesTheRestOfTheLine(t *testing.T) {
	e := newTestEngine(t)
	addPlayerAt(t, e, "Steve", world.Vec{X: 1, Y: 4, Z: 1})
	line := `tellraw Steve {"text":"hello there","color":"red"}`
	if err := runLine(e, line); err != nil {
		t.Fatalf("tellraw: %v", err)
	}
	feedback, err := runLineFeedback(e, line)
	if err != nil {
		t.Fatalf("tellraw: %v", err)
	}
	if len(feedback) != 1 {
		t.Fatalf("feedback = %v, want one line", feedback)
	}
}
