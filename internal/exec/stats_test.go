package exec_test

import (
	"testing"
	"time"

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

// TestStatsCountsTheCommands covers the counting and the grouping.
//
// It does **not** look at the order of the groups: that order is by the
// time the commands took, and two commands that really ran take different
// times from run to run. The order is covered by
// TestStatsOrdersTheGroups, which hands the collector the durations.
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
	counts := map[string]int{}
	for _, g := range groups {
		counts[g.Command] = g.Count
	}
	if counts["say"] != 2 || counts["scoreboard"] != 1 {
		t.Errorf("the counts are %v, want say twice and scoreboard once", counts)
	}
}

// TestStatsOrdersTheGroups covers the order of the report.
//
// The groups come back slowest first; a tie goes to the group with more
// commands, and a tie there goes by name. The durations are given to the
// collector, so the test does not depend on how fast the machine is.
func TestStatsOrdersTheGroups(t *testing.T) {
	stats := &exec.Stats{}
	stats.Record(exec.CommandStat{Command: "say", Duration: time.Millisecond})
	stats.Record(exec.CommandStat{Command: "say", Duration: time.Millisecond})
	stats.Record(exec.CommandStat{Command: "alpha", Duration: 2 * time.Millisecond})
	stats.Record(exec.CommandStat{Command: "setblock", Duration: 2 * time.Millisecond})
	stats.Record(exec.CommandStat{Command: "kill", Duration: 5 * time.Millisecond})

	groups := stats.ByCommand()
	if len(groups) != 4 {
		t.Fatalf("the stats have %d group(s), want 4", len(groups))
	}
	// kill is the slowest. say and setblock and alpha all took 2 ms:
	// say has more commands than the other two, and setblock sorts
	// before alpha.
	want := []string{"kill", "say", "alpha", "setblock"}
	for i, g := range groups {
		if g.Command != want[i] {
			t.Errorf("group %d is %q, want %q", i, g.Command, want[i])
		}
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
