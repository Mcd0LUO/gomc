package v1_12_2

import (
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
)

// tickSource is the listener a tick runs under.
func tickSource() *exec.Source {
	return &exec.Source{Name: "Server", World: "world", PermissionLevel: 4}
}

// TestTimeAdvancesWithTheTick covers the clock.
//
// 1.12.2 adds one to the tick counter on every tick, and one to the day
// time as well. The day time is **not** wrapped at 24000: it keeps
// counting, and only `time query daytime` takes it modulo 24000.
func TestTimeAdvancesWithTheTick(t *testing.T) {
	e := newTestEngine(t)
	for i := 0; i < 3; i++ {
		if _, err := e.Tick(tickSource()); err != nil {
			t.Fatalf("Tick: %v", err)
		}
	}
	if e.World.Time != 3 || e.World.DayTime != 3 {
		t.Fatalf("time = %d, day time = %d, want 3 and 3", e.World.Time, e.World.DayTime)
	}
}

// TestDaylightCycleFreezesTheTimeOfDay covers `doDaylightCycle`.
//
// 1.12.2's `WorldServer.doTick` always adds one to the tick counter, and
// adds one to the day time only while the rule is on. So a world with the
// rule off keeps counting ticks while the sun stands still.
func TestDaylightCycleFreezesTheTimeOfDay(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "gamerule doDaylightCycle false"); err != nil {
		t.Fatalf("gamerule: %v", err)
	}
	for i := 0; i < 2; i++ {
		if _, err := e.Tick(tickSource()); err != nil {
			t.Fatalf("Tick: %v", err)
		}
	}
	if e.World.Time != 2 {
		t.Errorf("tick counter = %d, want 2: it advances either way", e.World.Time)
	}
	if e.World.DayTime != 0 {
		t.Errorf("day time = %d, want 0: the rule freezes it", e.World.DayTime)
	}

	// Turning the rule back on lets it run again.
	if err := runLine(e, "gamerule doDaylightCycle true"); err != nil {
		t.Fatalf("gamerule: %v", err)
	}
	if _, err := e.Tick(tickSource()); err != nil {
		t.Fatalf("Tick: %v", err)
	}
	if e.World.DayTime != 1 {
		t.Errorf("day time = %d, want 1", e.World.DayTime)
	}
}

// TestTimeQueryCountsDays covers the three queries.
//
// `daytime` is the day time modulo 24000, `day` is the day time divided by
// 24000, and `gametime` is the tick counter. A day time that crossed the
// 24000 mark has to make `day` grow: 1.12.2 does not wrap the day time
// itself.
func TestTimeQueryCountsDays(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "time set 23999"); err != nil {
		t.Fatalf("time set: %v", err)
	}
	if _, err := e.Tick(tickSource()); err != nil {
		t.Fatalf("Tick: %v", err)
	}
	for _, tc := range []struct{ query, want string }{
		{"daytime", "Time is 0"},
		{"day", "Time is 1"},
		{"gametime", "Time is 1"},
	} {
		got, err := runLineFeedback(e, "time query "+tc.query)
		if err != nil {
			t.Fatalf("time query %s: %v", tc.query, err)
		}
		if len(got) != 1 || got[0] != tc.want {
			t.Errorf("time query %s said %q, want %q", tc.query, got, tc.want)
		}
	}
}

// TestTimeSetDoesNotMoveTheGametime covers the two counters.
//
// `/time set` moves the time of day and leaves the tick counter alone
// (`CommandTime.a(server, n)` calls `setDayTime`).
func TestTimeSetDoesNotMoveTheGametime(t *testing.T) {
	e := newTestEngine(t)
	if _, err := e.Tick(tickSource()); err != nil {
		t.Fatalf("Tick: %v", err)
	}
	if err := runLine(e, "time set 5000"); err != nil {
		t.Fatalf("time set: %v", err)
	}
	if e.World.Time != 1 {
		t.Errorf("tick counter = %d, want 1: /time set must not touch it", e.World.Time)
	}
	if e.World.DayTime != 5000 {
		t.Errorf("day time = %d, want 5000", e.World.DayTime)
	}
}

// TestTimeRejectsBadCounts covers the argument of `/time set` and `/time
// add`.
//
// Both read the value with `CommandTime.a(text, 0)`, so a negative number
// is refused as too small and a number that is not a Java int is not a
// valid number. The number in the "too small" message is the integer
// itself, not a two-decimal one.
func TestTimeRejectsBadCounts(t *testing.T) {
	e := newTestEngine(t)
	for _, tc := range []struct{ line, want string }{
		{"time set -5", "The number you have entered (-5) is too small, it must be at least 0"},
		{"time add -1", "The number you have entered (-1) is too small, it must be at least 0"},
		{"time set nope", "'nope' is not a valid number"},
		{"time set 3000000000", "'3000000000' is not a valid number"},
	} {
		err := runLine(e, tc.line)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%q = %v, want %q", tc.line, err, tc.want)
		}
	}
}
