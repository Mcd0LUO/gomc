package exec_test

import (
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/version/v1_12_2"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// newStatsEngine returns an engine that records what it runs.
func newStatsEngine(tb testing.TB) (*exec.Engine, *exec.Stats) {
	tb.Helper()
	reg := exec.NewRegistry()
	v1_12_2.Register(reg)
	engine := exec.NewEngine(reg, world.NewWorld())
	stats := &exec.Stats{}
	engine.SetCommandRecorder(stats)
	return engine, stats
}

// statsSource returns a source that may run every command.
func statsSource() *exec.Source {
	return &exec.Source{
		Name:            "Server",
		World:           "world",
		PermissionLevel: 4,
		Feedback:        true,
	}
}

// TestStatsCountsTheCommands covers the counting.
func TestStatsCountsTheCommands(t *testing.T) {
	e, stats := newStatsEngine(t)
	for _, line := range []string{
		"say one",
		"say two",
		"scoreboard objectives add counter dummy",
	} {
		if _, err := e.ExecuteLine(line, 1, statsSource()); err != nil {
			t.Fatalf("%q: %v", line, err)
		}
	}
	if stats.Count() != 3 {
		t.Errorf("the stats count %d command(s), want 3", stats.Count())
	}
	groups := stats.ByCommand()
	if len(groups) != 2 {
		t.Fatalf("the stats have %d group(s), want 2", len(groups))
	}
	// The groups come back slowest first, and the two commands took the
	// same time, so the one with more commands comes first.
	if groups[0].Command != "say" || groups[0].Count != 2 {
		t.Errorf("the first group is %+v, want \u201csay\u201d twice", groups[0])
	}
}

// TestStatsKnowsTheFunction covers the grouping by function.
func TestStatsKnowsTheFunction(t *testing.T) {
	e, stats := newStatsEngine(t)
	fn := &exec.Function{
		ID:          "pack:main",
		Lines:       []string{"say one", "say two"},
		LineNumbers: []int{1, 2},
		Path:        "pack:main",
	}
	if _, err := e.RunFunction(fn, statsSource()); err != nil {
		t.Fatalf("the function failed: %v", err)
	}
	groups := stats.ByFunction()
	if len(groups) != 1 {
		t.Fatalf("the stats have %d group(s), want 1", len(groups))
	}
	if groups[0].Function != "pack:main" {
		t.Errorf("the group is %q, want pack:main", groups[0].Function)
	}
	if groups[0].Count != 2 {
		t.Errorf("the group counts %d command(s), want 2", groups[0].Count)
	}
}

// TestStatsForgetTheFunctionAfterTheRun covers the boundary of a run.
//
// A command that runs after the function is over does not belong to it.
func TestStatsForgetTheFunctionAfterTheRun(t *testing.T) {
	e, stats := newStatsEngine(t)
	fn := &exec.Function{
		ID:          "pack:main",
		Lines:       []string{"say inside"},
		LineNumbers: []int{1},
		Path:        "pack:main",
	}
	if _, err := e.RunFunction(fn, statsSource()); err != nil {
		t.Fatalf("the function failed: %v", err)
	}
	if _, err := e.ExecuteLine("say outside", 1, statsSource()); err != nil {
		t.Fatalf("say: %v", err)
	}

	groups := stats.ByFunction()
	if len(groups) != 2 {
		t.Fatalf("the stats have %d group(s), want 2", len(groups))
	}
	for _, g := range groups {
		if g.Function == "" && g.Count != 1 {
			t.Errorf("the group outside a function counts %d command(s), want 1", g.Count)
		}
	}
}

// TestStatsCountsTheFailures covers the failed commands.
func TestStatsCountsTheFailures(t *testing.T) {
	e, stats := newStatsEngine(t)
	// A command that does not exist is refused before it runs, so it is
	// not measured: it never took any time.
	if _, err := e.ExecuteLine("notacommand foo", 1, statsSource()); err == nil {
		t.Fatal("the command should have failed")
	}
	if stats.Count() != 0 {
		t.Errorf("the stats count %d command(s), want none", stats.Count())
	}

	// A command that runs and fails is measured.
	if _, err := e.ExecuteLine("scoreboard players set Steve nosuchobjective 1", 1, statsSource()); err == nil {
		t.Fatal("the command should have failed")
	}
	if _, err := e.ExecuteLine("say hello", 1, statsSource()); err != nil {
		t.Fatalf("say: %v", err)
	}
	groups := stats.ByCommand()
	if len(groups) != 2 {
		t.Fatalf("the stats have %d group(s), want 2", len(groups))
	}
	for _, g := range groups {
		if g.Command == "scoreboard" && g.Failed != 1 {
			t.Errorf("the failed group says %d failure(s), want 1", g.Failed)
		}
	}
}
