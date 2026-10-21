package v1_12_2

import (
	"testing"
)

// TestSummonReadsThePositionAndTheDataTagByArgumentCount covers the
// argument layout of `/summon`.
//
// 1.12.2's `CommandSummon.execute` reads the position only when there are
// at least three coordinates after the entity, and the dataTag only when
// there is a fourth argument. The arguments in between are ignored.
//
// So `summon <entity> {dataTag}` does **not** apply the dataTag: the entity
// appears where the listener is, and the NBT is dropped.
//
// (Measured on a real 1.12.2 server: `summon cow {NoAI:1b}` answers "Object
// successfully summoned" and `testfor @e[type=cow] {NoAI:1b}` then answers
// "Cow did not match the required data structure".)
func TestSummonReadsThePositionAndTheDataTagByArgumentCount(t *testing.T) {
	e := newTestEngine(t)

	// The full form: the position and the dataTag are both read.
	if err := runLine(e, "summon zombie 0 4 0 {NoAI:1b}"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	feedback, err := runLineFeedback(e, "testfor @e[type=zombie] {NoAI:1b}")
	if err != nil {
		t.Fatalf("testfor: %v", err)
	}
	if len(feedback) != 1 || feedback[0] != "Found Zombie" {
		t.Errorf("feedback = %v, want the dataTag to be applied", feedback)
	}

	// One argument: neither the position nor the dataTag is read, and the
	// extra argument is ignored.
	if err := runLine(e, "summon cow {NoAI:1b}"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	feedback, err = runLineFeedback(e, "testfor @e[type=cow] {NoAI:1b}")
	if err == nil {
		t.Fatalf("the dataTag was applied, but it should have been ignored: %v", feedback)
	}
	if want := "Cow did not match the required data structure"; err.Error() != want {
		t.Errorf("the error was %q, want %q", err.Error(), want)
	}

	// Two arguments: the same, and the command still succeeds.
	if err := runLine(e, "summon pig 0 4"); err != nil {
		t.Fatalf("summon: %v", err)
	}
}

// TestDataTagIsTheRestOfTheLine covers that a dataTag with spaces in it is
// one argument.
//
// 1.12.2 hands the arguments from the dataTag's place on to the NBT parser
// joined with spaces.
func TestDataTagIsTheRestOfTheLine(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, `summon armor_stand 0 4 0 {CustomName:"hello there"}`); err != nil {
		t.Fatalf("summon: %v", err)
	}
	feedback, err := runLineFeedback(e, `testfor @e[type=armor_stand] {CustomName:"hello there"}`)
	if err != nil {
		t.Fatalf("testfor: %v", err)
	}
	if len(feedback) != 1 {
		t.Errorf("feedback = %v, want the custom name to be applied", feedback)
	}
}
