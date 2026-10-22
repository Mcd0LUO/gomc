package v1_12_2

import (
	"strings"
	"testing"

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

// TestExecuteRunsForEveryEntity covers the 1.12.2 expansion.
//
// `execute` runs the sub-command once per matching entity, in reverse
// order, and the target entity becomes the executor of the sub-command —
// so the output is prefixed with the entity's name.
func TestExecuteRunsForEveryEntity(t *testing.T) {
	e := newTestEngine(t)
	for _, id := range []string{"a1", "a2"} {
		s, _ := e.World.EntityTypes.NewEntity("armor_stand", id)
		s.Pos = world.Vec{X: 1, Y: 2, Z: 3}
		e.World.AddEntity(s)
	}
	if err := runLine(e, "execute @e[type=armor_stand] ~ ~ ~ say HELLO"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	var count int
	for _, msg := range e.Feedback {
		if strings.Contains(msg, "HELLO") {
			count++
			// The prefix is the executing entity's name, not "Server".
			if strings.HasPrefix(msg, "[Server]") {
				t.Errorf("the message = %q, want it to be prefixed with the entity", msg)
			}
		}
	}
	if count != 2 {
		t.Errorf("the sub-command ran %d times, want once per entity", count)
	}
}
