package v1_12_2

import (
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
)

// TestScoreboardObjectiveAddChecksTheServerDoes covers the four checks
// 1.12.2 makes when an objective is added.
//
// `CommandScoreboard.a(...)` checks the criteria type, then whether the
// objective exists already, then the length of the name, then the length
// of the display name -- and only creates the objective when all four
// pass.
//
// (Measured on a real 1.12.2 server: adding the same objective twice
// answers "An objective with the name 't' already exists", a 17 character
// name answers "The name '...' is too long for an objective, it can be at
// most 16 characters long", and an unknown criteria type answers "Usage:
// Invalid objective criteria type 'notacriteria'" -- the last one is a
// usage error, so it carries the "Usage: " prefix.)
func TestScoreboardObjectiveAddChecksTheServerDoes(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "scoreboard objectives add t dummy"); err != nil {
		t.Fatalf("add: %v", err)
	}
	err := runLine(e, "scoreboard objectives add t dummy")
	if err == nil {
		t.Fatal("expected an error for an objective that exists already")
	}
	if want := "An objective with the name 't' already exists"; !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err, want)
	}
	err = runLine(e, "scoreboard objectives add abcdefghijklmnopq dummy")
	if err == nil {
		t.Fatal("expected an error for a name that is too long")
	}
	if want := "too long for an objective, it can be at most 16 characters long"; !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err, want)
	}
	err = runLine(e, "scoreboard objectives add t2 dummy 123456789012345678901234567890123")
	if err == nil {
		t.Fatal("expected an error for a display name that is too long")
	}
	if want := "too long for an objective, it can be at most 32 characters long"; !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err, want)
	}
	err = runLine(e, "scoreboard objectives add t3 notacriteria")
	if err == nil {
		t.Fatal("expected an error for an unknown criteria type")
	}
	if want := "Usage: Invalid objective criteria type 'notacriteria'"; !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err, want)
	}
}

// TestScoreboardTeamAddChecksTheServerDoes covers the same checks for a
// team.
func TestScoreboardTeamAddChecksTheServerDoes(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "scoreboard teams add tt"); err != nil {
		t.Fatalf("add: %v", err)
	}
	err := runLine(e, "scoreboard teams add tt")
	if err == nil {
		t.Fatal("expected an error for a team that exists already")
	}
	if want := "A team with the name 'tt' already exists"; !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err, want)
	}
	err = runLine(e, "scoreboard teams add tt2 123456789012345678901234567890123")
	if err == nil {
		t.Fatal("expected an error for a display name that is too long")
	}
	if want := "too long for a team, it can be at most 32 characters long"; !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err, want)
	}
}

// TestDetectNeedsALoadedChunk covers the precondition the `detect` of
// `/execute` has.
//
// 1.12.2 checks that the chunk of the detect position is loaded before it
// looks at the block there (`World.isLoaded`), and the failure is the
// same one as a block that does not match.
//
// (Measured on a real 1.12.2 server: `execute Steve 900 4 900 detect 900 4
// 900 stone 0 say hi` answers "Failed to execute 'detect' as Steve" -- the
// chunk at that distance is not loaded.)
func TestDetectNeedsALoadedChunk(t *testing.T) {
	e := newTestEngine(t)
	addPlayer(t, e, "Steve")
	// The failure of a detect belongs to the entity, so it is collected
	// instead of returned.
	if _, err := e.ExecuteLine("execute Steve 900 4 900 detect 900 4 900 stone 0 say hi", 1,
		&exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}); err != nil {
		t.Fatalf("ExecuteLine: %v", err)
	}
	errs := e.TakeErrors()
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1: %v", len(errs), errs)
	}
	if want := "Failed to execute 'detect' as Steve"; !strings.Contains(errs[0].Error(), want) {
		t.Errorf("error = %q, want it to contain %q", errs[0], want)
	}
}
