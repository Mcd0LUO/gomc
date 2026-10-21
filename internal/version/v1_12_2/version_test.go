package v1_12_2

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// newTestEngine builds an engine with the 1.12.2 command set.
func newTestEngine(t *testing.T) *exec.Engine {
	t.Helper()
	reg := exec.NewRegistry()
	Register(reg)
	return exec.NewEngine(reg, world.NewWorld())
}

// writeFunction writes a function file into a temporary world layout.
func writeFunction(t *testing.T, root, ns, name, body string) {
	t.Helper()
	dir := filepath.Join(root, "data", "functions", ns)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(dir, name+".mcfunction")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// parseTestSelector parses a selector for the tests.
func parseTestSelector(text string) (*ast.Selector, error) {
	return ParseSelector(text)
}

func TestFunctionLoadsFromWorldLayout(t *testing.T) {
	root := t.TempDir()
	writeFunction(t, root, "demo", "hello", "# a comment\nsay hello world\n")
	e := newTestEngine(t)
	e.SetRoot(root)
	if errs := e.LoadFunctions(); len(errs) > 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	fn := e.Functions.Get("demo:hello")
	if fn == nil {
		t.Fatal("demo:hello was not loaded")
	}
	if len(fn.Lines) != 1 {
		t.Errorf("loaded %d lines, want 1 (the comment must be skipped)", len(fn.Lines))
	}
	if fn.LineNumbers[0] != 2 {
		t.Errorf("line number = %d, want 2", fn.LineNumbers[0])
	}
}

// TestCommandGroupsCoverEveryCommand keeps the listing of the commands in
// step with the registry.
//
// The groups are only for the reader, but a command that is in no group
// would simply not be listed.
func TestCommandGroupsCoverEveryCommand(t *testing.T) {
	reg := exec.NewRegistry()
	Register(reg)
	seen := map[string]bool{}
	for _, g := range CommandGroups() {
		for _, name := range g.Commands {
			if seen[name] {
				t.Errorf("the command %q is in two groups", name)
			}
			seen[name] = true
			if _, ok := reg.Get(name); !ok {
				t.Errorf("the group %q lists %q, which is not a command", g.Name, name)
			}
		}
	}
	for _, name := range reg.Names() {
		if !seen[name] {
			t.Errorf("the command %q is in no group", name)
		}
	}
}

func TestUnknownCommandFailsWholeFunction(t *testing.T) {
	// 1.12.2 throws when a function references an unknown command, and the
	// whole file fails to load rather than skipping the offending line.
	root := t.TempDir()
	writeFunction(t, root, "demo", "bad", "say ok\nnotacommand arg\n")
	e := newTestEngine(t)
	e.SetRoot(root)
	errs := e.LoadFunctions()
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
	if e.Functions.Get("demo:bad") != nil {
		t.Error("the function must not be registered when a command is unknown")
	}
	loadErr, ok := errs[0].(*exec.FunctionLoadError)
	if !ok {
		t.Fatalf("error type = %T, want *exec.FunctionLoadError", errs[0])
	}
	if loadErr.Line != 2 {
		t.Errorf("error line = %d, want 2", loadErr.Line)
	}
}

// TestLintChecksTheLinesOfABrokenFunction covers that one unknown command
// name does not hide the other mistakes in the same file.
//
// 1.12.2 refuses the whole function, so the file never runs — but the
// lines around the mistake are still worth reporting, otherwise one typo
// hides everything else the author wrote wrong.
func TestLintChecksTheLinesOfABrokenFunction(t *testing.T) {
	root := t.TempDir()
	writeFunction(t, root, "demo", "bad",
		"say ok\nnotacommand arg\nscoreboard players add Steve counter notanumber\n")
	e := newTestEngine(t)
	e.SetRoot(root)
	e.LoadFunctions()
	lint := e.LintFunctions()
	if len(lint) != 1 {
		t.Fatalf("got %d lint errors, want 1: %v", len(lint), lint)
	}
	lintErr, ok := lint[0].(*exec.FunctionLintError)
	if !ok {
		t.Fatalf("error type = %T, want *exec.FunctionLintError", lint[0])
	}
	if lintErr.Line != 3 {
		t.Errorf("error line = %d, want 3", lintErr.Line)
	}
}

func TestLeadingSlashIsRejected(t *testing.T) {
	root := t.TempDir()
	writeFunction(t, root, "demo", "slash", "/say hello\n")
	e := newTestEngine(t)
	e.SetRoot(root)
	errs := e.LoadFunctions()
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
	loadErr := errs[0].(*exec.FunctionLoadError)
	want := "Unknown or invalid command '/say' (did you mean 'say'? Do not use a preceding forwards slash.)"
	if loadErr.Msg != want {
		t.Errorf("message = %q, want %q", loadErr.Msg, want)
	}
}

func TestFunctionRunsCommandsInOrder(t *testing.T) {
	root := t.TempDir()
	writeFunction(t, root, "demo", "seq", "scoreboard objectives add a dummy\nscoreboard players set s a 1\nscoreboard players add s a 2\n")
	e := newTestEngine(t)
	e.SetRoot(root)
	if errs := e.LoadFunctions(); len(errs) > 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	if _, err := e.RunFunctionByID("demo:seq", src); err != nil {
		t.Fatalf("RunFunctionByID: %v", err)
	}
	score, ok := e.World.Scoreboard.Score("s", "a")
	if !ok {
		t.Fatal("holder s has no score for objective a")
	}
	if score != 3 {
		t.Errorf("score = %d, want 3", score)
	}
}

func TestFunctionRecursion(t *testing.T) {
	root := t.TempDir()
	writeFunction(t, root, "demo", "rec", "scoreboard objectives add a dummy\nscoreboard players add s a 1\n")
	e := newTestEngine(t)
	e.SetRoot(root)
	if errs := e.LoadFunctions(); len(errs) > 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	if _, err := e.RunFunctionByID("demo:rec", src); err != nil {
		t.Fatalf("RunFunctionByID: %v", err)
	}
	if score, _ := e.World.Scoreboard.Score("s", "a"); score != 1 {
		t.Errorf("score = %d, want 1", score)
	}
}

func TestExecuteUsesOldSyntax(t *testing.T) {
	// The 1.12.2 execute takes an entity and explicit coordinates, and the
	// coordinates are relative to the target entity.
	e := newTestEngine(t)
	e.World.AddEntity(&world.Entity{ID: "e1", Type: "armor_stand", Pos: world.Vec{X: 10, Y: 20, Z: 30}})
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	line := "execute @e[type=armor_stand] ~ ~1 ~ setblock ~ ~ ~ stone"
	if _, err := e.ExecuteLine(line, 1, src); err != nil {
		t.Fatalf("ExecuteLine: %v", err)
	}
	// The block lands one block above the entity.
	want := world.Pos{X: 10, Y: 21, Z: 30}
	got := e.World.Blocks.Get(want)
	if got.Name != "stone" {
		t.Errorf("block at %v = %s, want stone", want, got.Name)
	}
}

func TestExecuteDetectPrecondition(t *testing.T) {
	e := newTestEngine(t)
	e.World.AddEntity(&world.Entity{ID: "e1", Type: "armor_stand", Pos: world.Vec{X: 0, Y: 0, Z: 0}})
	e.World.Blocks.Set(world.Pos{X: 0, Y: -1, Z: 0}, world.BlockState{Name: "sand", Data: 0})
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	// The detect succeeds, so the command runs.
	line := "execute @e[type=armor_stand] ~ ~ ~ detect ~ ~-1 ~ sand 0 setblock ~ ~1 ~ stone"
	if _, err := e.ExecuteLine(line, 1, src); err != nil {
		t.Fatalf("ExecuteLine: %v", err)
	}
	if got := e.World.Blocks.Get(world.Pos{X: 0, Y: 1, Z: 0}); got.Name != "stone" {
		t.Errorf("block = %s, want stone", got.Name)
	}
}

func TestExecuteDetectFailureBlocksCommand(t *testing.T) {
	e := newTestEngine(t)
	e.World.AddEntity(&world.Entity{ID: "e1", Type: "armor_stand", Pos: world.Vec{X: 0, Y: 0, Z: 0}})
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	// There is no sand below, so the command must not run. 1.12.2 reports
	// the failure of that one entity and carries on.
	line := "execute @e[type=armor_stand] ~ ~ ~ detect ~ ~-1 ~ sand 0 setblock ~ ~1 ~ stone"
	if _, err := e.ExecuteLine(line, 1, src); err != nil {
		t.Fatalf("ExecuteLine: %v", err)
	}
	if got := e.World.Blocks.Get(world.Pos{X: 0, Y: 1, Z: 0}); got.Name != "air" {
		t.Errorf("block = %s, want air", got.Name)
	}
	errs := e.TakeErrors()
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1: %v", len(errs), errs)
	}
	if want := "Failed to execute 'detect' as Armor Stand"; !strings.Contains(errs[0].Error(), want) {
		t.Errorf("error = %q, want it to contain %q", errs[0], want)
	}
}

// 1.12.2 expands a multi-entity target in the command dispatcher, which
// visits the entities in the order the selector returned them. `PlayerSelector`
// sorts @e by distance, so the nearest entity is visited first.
func TestExecuteVisitsEntitiesInDistanceOrder(t *testing.T) {
	e := newTestEngine(t)
	e.World.AddEntity(&world.Entity{ID: "far", Type: "zombie", Pos: world.Vec{X: 8, Y: 0, Z: 0}})
	e.World.AddEntity(&world.Entity{ID: "near", Type: "armor_stand", Pos: world.Vec{X: 1, Y: 0, Z: 0}})
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	if _, err := e.ExecuteLine("execute @e ~ ~ ~ say ORDER", 1, src); err != nil {
		t.Fatalf("ExecuteLine: %v", err)
	}
	want := []string{"[Armor Stand] ORDER", "[Zombie] ORDER"}
	if len(e.Feedback) != len(want) {
		t.Fatalf("feedback = %v, want %v", e.Feedback, want)
	}
	for i := range want {
		if e.Feedback[i] != want[i] {
			t.Errorf("feedback[%d] = %q, want %q", i, e.Feedback[i], want[i])
		}
	}
}

// A failure belongs to the entity that caused it: the dispatcher reports
// it and still runs the command for the remaining entities.
func TestExecuteFailureDoesNotStopTheOtherEntities(t *testing.T) {
	e := newTestEngine(t)
	// The near entity has no sand below, so its detect fails. The far one
	// has sand, so it must still run.
	e.World.AddEntity(&world.Entity{ID: "near", Type: "zombie", Pos: world.Vec{X: 1, Y: 0, Z: 0}})
	e.World.AddEntity(&world.Entity{ID: "far", Type: "armor_stand", Pos: world.Vec{X: 8, Y: 0, Z: 0}})
	e.World.Blocks.Set(world.Pos{X: 8, Y: -1, Z: 0}, world.BlockState{Name: "sand", Data: 0})
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	line := "execute @e ~ ~ ~ detect ~ ~-1 ~ sand 0 say RAN"
	if _, err := e.ExecuteLine(line, 1, src); err != nil {
		t.Fatalf("ExecuteLine: %v", err)
	}
	if len(e.Feedback) != 1 || e.Feedback[0] != "[Armor Stand] RAN" {
		t.Fatalf("feedback = %v, want [Armor Stand] RAN", e.Feedback)
	}
	errs := e.TakeErrors()
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1: %v", len(errs), errs)
	}
	if want := "Failed to execute 'detect' as Zombie"; !strings.Contains(errs[0].Error(), want) {
		t.Errorf("error = %q, want it to contain %q", errs[0], want)
	}
}

func TestGameLoopFunctionRunsEveryTick(t *testing.T) {
	root := t.TempDir()
	writeFunction(t, root, "demo", "loop", "scoreboard objectives add tick dummy\nscoreboard players add s tick 1\n")
	e := newTestEngine(t)
	e.SetRoot(root)
	if errs := e.LoadFunctions(); len(errs) > 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	if err := e.World.Rules.Set("gameLoopFunction", "demo:loop"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	for i := 0; i < 3; i++ {
		if _, err := e.Tick(src); err != nil {
			t.Fatalf("Tick: %v", err)
		}
	}
	score, ok := e.World.Scoreboard.Score("s", "tick")
	if !ok {
		t.Fatal("no score recorded")
	}
	if score != 3 {
		t.Errorf("score after 3 ticks = %d, want 3", score)
	}
	if e.World.Time != 3 {
		t.Errorf("world time = %d, want 3", e.World.Time)
	}
}

func TestSelectorTypeAndTagFilters(t *testing.T) {
	e := newTestEngine(t)
	z, _ := e.World.EntityTypes.NewEntity("zombie", "z1")
	e.World.AddEntity(z)
	s, _ := e.World.EntityTypes.NewEntity("armor_stand", "a1")
	s.SetTag("marked", true)
	e.World.AddEntity(s)

	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	resolver := &SelectorResolver{World: e.World}

	sel, err := parseTestSelector("@e[type=zombie]")
	if err != nil {
		t.Fatalf("selector: %v", err)
	}
	entities, err := resolver.Resolve(src, sel)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(entities) != 1 || entities[0].ID != "z1" {
		t.Errorf("resolved %v, want only z1", entities)
	}

	sel, _ = parseTestSelector("@e[tag=marked]")
	entities, _ = resolver.Resolve(src, sel)
	if len(entities) != 1 || entities[0].ID != "a1" {
		t.Errorf("resolved %v, want only a1", entities)
	}

	sel, _ = parseTestSelector("@e[tag=!marked]")
	entities, _ = resolver.Resolve(src, sel)
	if len(entities) != 1 || entities[0].ID != "z1" {
		t.Errorf("resolved %v, want only z1", entities)
	}
}

func TestScoreboardOperationSemantics(t *testing.T) {
	e := newTestEngine(t)
	sb := e.World.Scoreboard
	sb.AddObjective("a", world.CriteriaDummy, "")
	sb.AddObjective("b", world.CriteriaDummy, "")
	sb.SetScore("x", "b", 7)

	// Integer division truncates toward zero, as in Java.
	got, err := world.Apply(world.OpDiv, 7, 2)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got != 3 {
		t.Errorf("7/2 = %d, want 3", got)
	}
	// `<` keeps the smaller value.
	got, _ = world.Apply(world.OpKeepMin, 5, 9)
	if got != 5 {
		t.Errorf("min(5,9) = %d, want 5", got)
	}
	// `><` swaps the values.
	got, _ = world.Apply(world.OpSwap, 5, 9)
	if got != 9 {
		t.Errorf("swap(5,9) = %d, want 9", got)
	}
}
