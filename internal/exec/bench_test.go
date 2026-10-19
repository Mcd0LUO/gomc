// Package exec_test holds the performance benchmarks of the command
// execution engine.
//
// The file lives in the external test package on purpose: package
// v1_12_2 imports package exec, so a test file inside package exec could
// not import the real command set back without an import cycle.
//
// Every scenario builds its own synthetic world in memory; nothing is
// read from or written to a save file. The only files that are written are
// the function files the loader needs, and they go into the test's own
// temporary directory.
//
// One "op" of a benchmark is one full pass of the scenario, so the
// reported numbers are per scenario and not per command. The scenario
// sizes are the constants below.
package exec_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/version/v1_12_2"
	"github.com/mcsim/gomc/internal/world"
)

// The sizes of the benchmark scenarios.
const (
	// benchSayLines is the number of `say` commands of the baseline.
	benchSayLines = 10000
	// benchSetBlockLines is the number of `setblock` commands.
	benchSetBlockLines = 10000
	// benchSummonLines is the number of `summon` commands.
	benchSummonLines = 5000
	// benchEntities is the number of entities the selector scan sees.
	benchEntities = 2000
	// benchSelectorRuns is the number of selector resolutions.
	benchSelectorRuns = 1000
	// benchScoreWrites is the number of scoreboard writes.
	benchScoreWrites = 10000
	// benchHolders is the number of tracked scoreboard holders.
	benchHolders = 20000
	// benchFunctionLines is the number of lines of the benchmark function.
	benchFunctionLines = 10000
	// benchTicks is the number of ticks of the tick loop.
	benchTicks = 100
	// benchTickLines is the number of lines of the game loop function.
	benchTickLines = 100
	// benchObjective is the name of the objective the scoreboard
	// benchmarks use.
	benchObjective = "bench"
)

// newBenchEngine returns an engine that knows the real 1.12.2 command set
// and owns an empty world.
func newBenchEngine(tb testing.TB) *exec.Engine {
	tb.Helper()
	reg := exec.NewRegistry()
	v1_12_2.Register(reg)
	return exec.NewEngine(reg, world.NewWorld())
}

// benchSource returns a console-like source that is allowed to run every
// command of the benchmark set.
func benchSource() *exec.Source {
	return &exec.Source{
		Name:            "Server",
		World:           "world",
		PermissionLevel: 4,
		Feedback:        true,
		Pos:             world.Vec{X: 0, Y: 64, Z: 0},
	}
}

// sayLines builds the baseline command lines.
func sayLines(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = "say hello"
	}
	return out
}

// setBlockLines builds the `setblock` command lines, one block per line.
func setBlockLines(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("setblock %d 64 %d stone", i, i)
	}
	return out
}

// summonLines builds the `summon` command lines, one entity per line.
func summonLines(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("summon armor_stand %d 64 %d", i, i)
	}
	return out
}

// scoreSetLines builds the `scoreboard players set` command lines, one
// holder per line.
func scoreSetLines(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("scoreboard players set holder%d %s %d", i, benchObjective, i%100)
	}
	return out
}

// newBenchFunction builds a function out of raw command lines.
//
// The runtime keeps the lines as text and parses them again on every run,
// so the function is built the same way the loader builds one.
func newBenchFunction(id string, lines []string) *exec.Function {
	fn := &exec.Function{
		ID:          id,
		Lines:       lines,
		LineNumbers: make([]int, len(lines)),
		Path:        "bench:" + id,
	}
	for i := range fn.LineNumbers {
		fn.LineNumbers[i] = i + 1
	}
	return fn
}

// writeBenchFunction writes a function file into the test's own temporary
// world directory and returns the world root.
func writeBenchFunction(tb testing.TB, ns, name string, lines []string) string {
	tb.Helper()
	root := tb.TempDir()
	dir := filepath.Join(root, "data", "functions", ns)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		tb.Fatalf("MkdirAll: %v", err)
	}
	body := strings.Join(lines, "\n") + "\n"
	path := filepath.Join(dir, name+".mcfunction")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		tb.Fatalf("WriteFile: %v", err)
	}
	return root
}

// loadBenchFunction loads a function through the real loader and returns
// it.
func loadBenchFunction(tb testing.TB, eng *exec.Engine, root, id string) *exec.Function {
	tb.Helper()
	eng.SetRoot(root)
	if errs := eng.LoadFunctions(); len(errs) > 0 {
		tb.Fatalf("LoadFunctions: %v", errs)
	}
	fn := eng.Functions.Get(id)
	if fn == nil {
		tb.Fatalf("function %q was not loaded", id)
	}
	return fn
}

// runLines runs every line through Engine.ExecuteLine.
func runLines(tb testing.TB, eng *exec.Engine, lines []string, src *exec.Source) {
	tb.Helper()
	for i, l := range lines {
		if _, err := eng.ExecuteLine(l, i+1, src); err != nil {
			tb.Fatalf("ExecuteLine(%q): %v", l, err)
		}
	}
}

// addEntities puts n armor stands into the world, spread along the x axis.
func addEntities(w *world.World, n int) {
	for i := 0; i < n; i++ {
		e, ok := w.EntityTypes.NewEntity("armor_stand", "")
		if !ok {
			return
		}
		e.Pos = world.Vec{X: float64(i), Y: 64, Z: 0}
		w.AddEntity(e)
	}
}

// BenchmarkBaselineSay measures the cost of a trivial command.
//
// The two sub-benchmarks differ only in the way the command reaches the
// engine: `ExecuteLine` parses the line and then runs it, while
// `ExecuteParsed` runs a command that was parsed before the timer
// started. The difference between the two is the price of parsing.
func BenchmarkBaselineSay(b *testing.B) {
	lines := sayLines(benchSayLines)

	b.Run("ExecuteLine", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			eng := newBenchEngine(b)
			src := benchSource()
			b.StartTimer()
			runLines(b, eng, lines, src)
		}
	})

	b.Run("RunFunction", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			eng := newBenchEngine(b)
			src := benchSource()
			fn := newBenchFunction("bench:say", lines)
			b.StartTimer()
			if _, err := eng.RunFunction(fn, src); err != nil {
				b.Fatalf("RunFunction: %v", err)
			}
		}
	})

	b.Run("ExecuteParsed", func(b *testing.B) {
		b.ReportAllocs()
		// The commands are parsed before the timer starts, so only the
		// run itself is measured.
		eng := newBenchEngine(b)
		src := benchSource()
		cmds := make([]*ast.Command, 0, len(lines))
		for i, l := range lines {
			cmd, err := eng.ParseCommand(l, i+1)
			if err != nil {
				b.Fatalf("ParseCommand(%q): %v", l, err)
			}
			cmds = append(cmds, cmd)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, cmd := range cmds {
				if _, err := eng.Execute(cmd, src); err != nil {
					b.Fatalf("Execute: %v", err)
				}
			}
		}
	})
}

// BenchmarkParseOnly measures the parsing of the command lines alone.
//
// This is the part of a command that is repeated on every single
// execution of a function, because the runtime keeps the lines as text.
func BenchmarkParseOnly(b *testing.B) {
	eng := newBenchEngine(b)
	for _, tc := range []struct {
		name  string
		lines []string
	}{
		{"Say", sayLines(benchSayLines)},
		{"Setblock", setBlockLines(benchSetBlockLines)},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				for j, l := range tc.lines {
					if _, err := eng.ParseCommand(l, j+1); err != nil {
						b.Fatalf("ParseCommand(%q): %v", l, err)
					}
				}
			}
		})
	}
}

// BenchmarkSetblock measures many `setblock` commands.
func BenchmarkSetblock(b *testing.B) {
	lines := setBlockLines(benchSetBlockLines)

	b.Run("ExecuteLine", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			eng := newBenchEngine(b)
			src := benchSource()
			b.StartTimer()
			runLines(b, eng, lines, src)
		}
	})

	b.Run("RunFunction", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			eng := newBenchEngine(b)
			src := benchSource()
			fn := newBenchFunction("bench:setblock", lines)
			b.StartTimer()
			if _, err := eng.RunFunction(fn, src); err != nil {
				b.Fatalf("RunFunction: %v", err)
			}
		}
	})
}

// BenchmarkSummon measures many `summon` commands.
func BenchmarkSummon(b *testing.B) {
	lines := summonLines(benchSummonLines)

	b.Run("ExecuteLine", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			eng := newBenchEngine(b)
			src := benchSource()
			b.StartTimer()
			runLines(b, eng, lines, src)
		}
	})

	b.Run("RunFunction", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			eng := newBenchEngine(b)
			src := benchSource()
			fn := newBenchFunction("bench:summon", lines)
			b.StartTimer()
			if _, err := eng.RunFunction(fn, src); err != nil {
				b.Fatalf("RunFunction: %v", err)
			}
		}
	})
}

// BenchmarkSelectorScan measures the cost of resolving a target selector
// when the world holds many entities.
//
// The `Resolver` sub-benchmark calls the selector resolver directly, so
// the number is the price of one selection. The `Command` sub-benchmark
// runs the whole command path, which adds the parsing of the command and
// the work that follows the selection.
func BenchmarkSelectorScan(b *testing.B) {
	sel, err := parser.ParseSelector("@e[type=armor_stand]")
	if err != nil {
		b.Fatalf("ParseSelector: %v", err)
	}

	b.Run("Resolver", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			eng := newBenchEngine(b)
			src := benchSource()
			addEntities(eng.World, benchEntities)
			resolver := &exec.SelectorResolver{World: eng.World}
			b.StartTimer()
			for j := 0; j < benchSelectorRuns; j++ {
				matched, err := resolver.Resolve(src, sel)
				if err != nil {
					b.Fatalf("Resolve: %v", err)
				}
				if len(matched) != benchEntities {
					b.Fatalf("matched %d entities, want %d", len(matched), benchEntities)
				}
			}
		}
	})

	b.Run("Command", func(b *testing.B) {
		b.ReportAllocs()
		line := "scoreboard players set @e[type=armor_stand] " + benchObjective + " 1"
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			eng := newBenchEngine(b)
			src := benchSource()
			eng.World.Scoreboard.AddObjective(benchObjective, world.CriteriaDummy, "")
			addEntities(eng.World, benchEntities)
			b.StartTimer()
			for j := 0; j < benchSelectorRuns; j++ {
				if _, err := eng.ExecuteLine(line, 1, src); err != nil {
					b.Fatalf("ExecuteLine: %v", err)
				}
				// The command answers once per matched entity. The
				// answer is not part of the selection, so it is dropped
				// outside the measured window.
				b.StopTimer()
				eng.Feedback = nil
				b.StartTimer()
			}
		}
	})
}

// BenchmarkSelectorScanScaling measures one selector resolution for
// different entity counts, which shows how the price of a selection grows
// with the size of the world.
func BenchmarkSelectorScanScaling(b *testing.B) {
	sel, err := parser.ParseSelector("@e[type=armor_stand]")
	if err != nil {
		b.Fatalf("ParseSelector: %v", err)
	}
	for _, n := range []int{100, 500, 2000, 8000} {
		b.Run(fmt.Sprintf("Entities%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				eng := newBenchEngine(b)
				src := benchSource()
				addEntities(eng.World, n)
				resolver := &exec.SelectorResolver{World: eng.World}
				b.StartTimer()
				matched, err := resolver.Resolve(src, sel)
				if err != nil {
					b.Fatalf("Resolve: %v", err)
				}
				if len(matched) != n {
					b.Fatalf("matched %d entities, want %d", len(matched), n)
				}
			}
		})
	}
}

// BenchmarkScoreboardWrite measures many `scoreboard players set`
// commands.
func BenchmarkScoreboardWrite(b *testing.B) {
	lines := scoreSetLines(benchScoreWrites)

	b.Run("ExecuteLine", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			eng := newBenchEngine(b)
			src := benchSource()
			eng.World.Scoreboard.AddObjective(benchObjective, world.CriteriaDummy, "")
			b.StartTimer()
			runLines(b, eng, lines, src)
		}
	})

	b.Run("RunFunction", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			eng := newBenchEngine(b)
			src := benchSource()
			eng.World.Scoreboard.AddObjective(benchObjective, world.CriteriaDummy, "")
			fn := newBenchFunction("bench:scores", lines)
			b.StartTimer()
			if _, err := eng.RunFunction(fn, src); err != nil {
				b.Fatalf("RunFunction: %v", err)
			}
		}
	})
}

// BenchmarkScoreboardList measures `scoreboard players list` with many
// tracked holders.
//
// The three forms of the command answer different questions:
//
//   - `AllHolders` lists every tracked holder, which walks the scores of
//     every objective.
//   - `SingleHolder` names one holder, which only walks the objectives.
//   - `SelectorHolders` names many holders through a selector, which
//     walks the objectives once per matched holder.
func BenchmarkScoreboardList(b *testing.B) {
	for _, objectives := range []int{1, 3, 8} {
		b.Run(fmt.Sprintf("SelectorHolders/Objectives%d", objectives), func(b *testing.B) {
			b.ReportAllocs()
			line := "scoreboard players list @e[type=armor_stand]"
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				eng := newBenchEngine(b)
				src := benchSource()
				addEntities(eng.World, benchHolders)
				for o := 0; o < objectives; o++ {
					name := fmt.Sprintf("%s%d", benchObjective, o)
					eng.World.Scoreboard.AddObjective(name, world.CriteriaDummy, "")
					for id := range eng.World.Entities {
						eng.World.Scoreboard.SetScore(id, name, o)
					}
				}
				b.StartTimer()
				if _, err := eng.ExecuteLine(line, 1, src); err != nil {
					b.Fatalf("ExecuteLine: %v", err)
				}
			}
		})
	}

	b.Run("AllHolders", func(b *testing.B) {
		b.ReportAllocs()
		line := "scoreboard players list"
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			eng := newBenchEngine(b)
			src := benchSource()
			for o := 0; o < 3; o++ {
				name := fmt.Sprintf("%s%d", benchObjective, o)
				eng.World.Scoreboard.AddObjective(name, world.CriteriaDummy, "")
				for h := 0; h < benchHolders; h++ {
					eng.World.Scoreboard.SetScore(fmt.Sprintf("holder%d", h), name, o)
				}
			}
			b.StartTimer()
			if _, err := eng.ExecuteLine(line, 1, src); err != nil {
				b.Fatalf("ExecuteLine: %v", err)
			}
		}
	})

	b.Run("SingleHolder", func(b *testing.B) {
		b.ReportAllocs()
		line := "scoreboard players list holder0"
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			eng := newBenchEngine(b)
			src := benchSource()
			for o := 0; o < 3; o++ {
				name := fmt.Sprintf("%s%d", benchObjective, o)
				eng.World.Scoreboard.AddObjective(name, world.CriteriaDummy, "")
				for h := 0; h < benchHolders; h++ {
					eng.World.Scoreboard.SetScore(fmt.Sprintf("holder%d", h), name, o)
				}
			}
			b.StartTimer()
			if _, err := eng.ExecuteLine(line, 1, src); err != nil {
				b.Fatalf("ExecuteLine: %v", err)
			}
		}
	})
}

// BenchmarkFunctionRun measures a function of ten thousand lines, run
// through Engine.RunFunction.
func BenchmarkFunctionRun(b *testing.B) {
	lines := setBlockLines(benchFunctionLines)
	root := writeBenchFunction(b, "bench", "big", lines)

	b.Run("RunFunction", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			eng := newBenchEngine(b)
			src := benchSource()
			fn := loadBenchFunction(b, eng, root, "bench:big")
			b.StartTimer()
			if _, err := eng.RunFunction(fn, src); err != nil {
				b.Fatalf("RunFunction: %v", err)
			}
		}
	})

	b.Run("LoadFunctions", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			eng := newBenchEngine(b)
			b.StartTimer()
			if errs := eng.LoadFunctions(); len(errs) > 0 {
				b.Fatalf("LoadFunctions: %v", errs)
			}
		}
	})
}

// BenchmarkTickLoop measures the per-tick hook.
//
// The `gameLoopFunction` game rule names the function that runs once per
// tick, so the benchmark runs the tick a hundred times over a function of
// a hundred lines.
func BenchmarkTickLoop(b *testing.B) {
	lines := setBlockLines(benchTickLines)
	root := writeBenchFunction(b, "bench", "loop", lines)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		eng := newBenchEngine(b)
		src := benchSource()
		loadBenchFunction(b, eng, root, "bench:loop")
		if err := eng.World.Rules.Set("gameLoopFunction", "bench:loop"); err != nil {
			b.Fatalf("Rules.Set: %v", err)
		}
		b.StartTimer()
		for t := 0; t < benchTicks; t++ {
			if _, err := eng.Tick(src); err != nil {
				b.Fatalf("Tick: %v", err)
			}
		}
	}
}
