package v1_12_2

import (
	"strings"
	"testing"
)

// TestTriggerLocksAfterItFires covers the trigger lock.
//
// 1.12.2 keeps a lock flag on the score (`ScoreboardScore`): `/trigger`
// requires it to be clear and sets it when the trigger fires, so a trigger
// works once and then stays locked until `/scoreboard players enable`
// unlocks it. A score that was just created is unlocked -- the constructor
// only sets the "needs an update" flag -- so the first `/trigger` works
// without an `enable`.
func TestTriggerLocksAfterItFires(t *testing.T) {
	e := newTestEngine(t)
	run := func(line string) {
		t.Helper()
		if err := runLine(e, line); err != nil {
			t.Fatalf("%s: %v", line, err)
		}
	}
	run("scoreboard objectives add t trigger")
	run("scoreboard players set Server t 0")
	run("trigger t add 3")
	if v, ok := e.World.Scoreboard.Score("Server", "t"); !ok || v != 3 {
		t.Fatalf("score = %v (%v), want 3", v, ok)
	}
	// The trigger fired once: it is locked now.
	err := runLine(e, "trigger t add 1")
	if err == nil || !strings.Contains(err.Error(), "Trigger t is not enabled") {
		t.Fatalf("the second trigger = %v, want the disabled message", err)
	}
	if v, _ := e.World.Scoreboard.Score("Server", "t"); v != 3 {
		t.Errorf("score = %v, want 3: a locked trigger does not change the score", v)
	}
	// `enable` unlocks it and keeps the score.
	run("scoreboard players enable Server t")
	if v, _ := e.World.Scoreboard.Score("Server", "t"); v != 3 {
		t.Errorf("score = %v, want 3: enable must not reset it", v)
	}
	run("trigger t set 7")
	if v, _ := e.World.Scoreboard.Score("Server", "t"); v != 7 {
		t.Errorf("score = %v, want 7", v)
	}
}

// TestPlayersEnableCreatesTheScore covers the entry `enable` makes.
//
// 1.12.2 calls `getPlayerScoreForObjective`, which creates the score with 0
// when the holder has none, and then unlocks it. So `enable` is enough to
// arm a trigger for a holder that has no score yet.
func TestPlayersEnableCreatesTheScore(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "scoreboard objectives add t trigger"); err != nil {
		t.Fatalf("objectives add: %v", err)
	}
	if err := runLine(e, "scoreboard players enable Server t"); err != nil {
		t.Fatalf("enable: %v", err)
	}
	if v, ok := e.World.Scoreboard.Score("Server", "t"); !ok || v != 0 {
		t.Fatalf("score = %v (%v), want 0 and tracked", v, ok)
	}
	if err := runLine(e, "trigger t add 1"); err != nil {
		t.Fatalf("trigger after enable: %v", err)
	}
}

// TestTriggerNeedsTheRightObjective covers the two ways the objective is
// wrong.
//
// A holder without a score and an objective that is not a trigger are
// reported with the **same** message (`commands.trigger.invalidObjective`):
// 1.12.2 has no separate wording for a missing score.
func TestTriggerNeedsTheRightObjective(t *testing.T) {
	for _, tc := range []struct {
		name   string
		setup  []string
		line   string
		expect string
	}{
		{
			"a trigger with no score",
			[]string{"scoreboard objectives add t trigger"},
			"trigger t add 1",
			"Invalid trigger name t",
		},
		{
			"a dummy objective",
			[]string{"scoreboard objectives add d dummy", "scoreboard players set Server d 0"},
			"trigger d add 1",
			"Invalid trigger name d",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newTestEngine(t)
			for _, line := range tc.setup {
				if err := runLine(e, line); err != nil {
					t.Fatalf("%q: %v", line, err)
				}
			}
			err := runLine(e, tc.line)
			if err == nil || !strings.Contains(err.Error(), tc.expect) {
				t.Fatalf("%q = %v, want %q", tc.line, err, tc.expect)
			}
		})
	}
}

// TestTriggerChecksTheArgumentsInOrder covers the order of the checks.
//
// `CommandTrigger` reads the value first, then the score and the lock, and
// only then the mode. So a value that is not a number is reported even when
// the trigger is locked, while a bad mode on a locked trigger reports the
// lock.
func TestTriggerChecksTheArgumentsInOrder(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "scoreboard objectives add t trigger"); err != nil {
		t.Fatalf("objectives add: %v", err)
	}
	if err := runLine(e, "scoreboard players set Server t 0"); err != nil {
		t.Fatalf("players set: %v", err)
	}
	if err := runLine(e, "trigger t add 1"); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	// Locked now. The number is read before the lock.
	err := runLine(e, "trigger t add nope")
	if err == nil || !strings.Contains(err.Error(), "is not a valid number") {
		t.Fatalf("bad number = %v, want the number message", err)
	}
	// The mode is read after the lock.
	err = runLine(e, "trigger t multiply 1")
	if err == nil || !strings.Contains(err.Error(), "Trigger t is not enabled") {
		t.Fatalf("bad mode on a locked trigger = %v, want the disabled message", err)
	}
}

// TestResetClearsTheTriggerLock covers `players reset`.
//
// The lock lives on the score, so removing the score removes the lock too:
// a trigger that fired, was reset and was set again works without an
// `enable`.
func TestResetClearsTheTriggerLock(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "scoreboard objectives add t trigger"); err != nil {
		t.Fatalf("objectives add: %v", err)
	}
	if err := runLine(e, "scoreboard players set Server t 0"); err != nil {
		t.Fatalf("players set: %v", err)
	}
	if err := runLine(e, "trigger t add 1"); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if err := runLine(e, "scoreboard players reset Server t"); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if err := runLine(e, "scoreboard players set Server t 0"); err != nil {
		t.Fatalf("players set: %v", err)
	}
	if err := runLine(e, "trigger t add 1"); err != nil {
		t.Fatalf("trigger after a reset: %v", err)
	}
}
