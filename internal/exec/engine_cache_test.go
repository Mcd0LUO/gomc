package exec_test

import (
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

// The tests in this file cover the cache of parsed function commands.
//
// The loader parses every line of a function once and keeps the result,
// so a function does not parse its commands again on every run. The tests
// pin down the two properties that matter: the cache does not change what
// a command does, and it does not turn 1.12.2's "the arguments are only
// checked when the function runs" into a load-time check.

// cacheCountCmd is a command that counts how often it is parsed and how
// often it runs.
type cacheCountCmd struct {
	// parses counts the calls to Parse.
	parses *int
	// runs counts the calls to Run.
	runs *int
}

// Name implements exec.Handler.
func (c *cacheCountCmd) Name() string { return "cachecount" }

// Usage implements exec.Handler.
func (c *cacheCountCmd) Usage() string { return "/cachecount <text>" }

// Parse implements exec.Handler.
func (c *cacheCountCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	*c.parses++
	return []ast.Arg{&ast.StringArg{Value: p.Rest()}}, nil
}

// Run implements exec.Handler.
func (c *cacheCountCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	*c.runs++
	return 1, nil
}

// writeCacheFunction writes a function file into a temporary world layout.
func writeCacheFunction(t *testing.T, root, ns, name, body string) {
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

// newCacheEngine returns an engine that knows the real 1.12.2 command set.
func newCacheEngine(t *testing.T) *exec.Engine {
	t.Helper()
	reg := exec.NewRegistry()
	v1_12_2.Register(reg)
	return exec.NewEngine(reg, world.NewWorld())
}

// TestFunctionWithBrokenArgumentsLoadsAnyway covers the 1.12.2 rule that a
// function is only checked for its command names when it is loaded.
//
// A mistake in the arguments stays hidden until the function runs, so the
// cache must not turn it into a load error.
func TestFunctionWithBrokenArgumentsLoadsAnyway(t *testing.T) {
	root := t.TempDir()
	// `setblock 1 2 3` misses the block name, so the line does not parse.
	writeCacheFunction(t, root, "demo", "broken", "say hello\nsetblock 1 2 3\nsay bye\n")
	eng := newCacheEngine(t)
	eng.SetRoot(root)

	if errs := eng.LoadFunctions(); len(errs) != 0 {
		t.Fatalf("LoadFunctions reported %d errors, want none: %v", len(errs), errs)
	}
	fn := eng.Functions.Get("demo:broken")
	if fn == nil {
		t.Fatal("the function was not registered")
	}
	// The good lines are cached, the broken one is not.
	if len(fn.Commands) != 3 {
		t.Fatalf("the function cached %d commands, want 3", len(fn.Commands))
	}
	if fn.Commands[0] == nil || fn.Commands[2] == nil {
		t.Error("the good lines should be cached")
	}
	if fn.Commands[1] != nil {
		t.Error("the broken line should not be cached")
	}

	// Running the function reports the mistake, and only then.
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	if _, err := eng.RunFunctionByID("demo:broken", src); err != nil {
		t.Fatalf("RunFunctionByID: %v", err)
	}
	errs := eng.TakeErrors()
	if len(errs) != 1 {
		t.Fatalf("the run collected %d errors, want 1: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Error(), "expected a value") {
		t.Errorf("the error = %q, want the parse failure of the second line", errs[0].Error())
	}
	if !strings.Contains(errs[0].Error(), ":2:") {
		t.Errorf("the error = %q, want it to point at line 2", errs[0].Error())
	}
	// The good commands still ran.
	want := []string{"[Server] hello", "[Server] bye"}
	if len(eng.Feedback) != len(want) {
		t.Fatalf("the feedback is %v, want %v", eng.Feedback, want)
	}
	for i := range want {
		if eng.Feedback[i] != want[i] {
			t.Errorf("the feedback %d = %q, want %q", i, eng.Feedback[i], want[i])
		}
	}
}

// TestFunctionIsParsedOnce covers the cache itself.
//
// The loader parses the lines of a function once, and running the function
// does not parse them again.
func TestFunctionIsParsedOnce(t *testing.T) {
	root := t.TempDir()
	writeCacheFunction(t, root, "demo", "counted", "cachecount a\ncachecount b\n")

	parses, runs := 0, 0
	reg := exec.NewRegistry()
	reg.Register(&cacheCountCmd{parses: &parses, runs: &runs})
	eng := exec.NewEngine(reg, world.NewWorld())
	eng.SetRoot(root)
	if errs := eng.LoadFunctions(); len(errs) != 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	if parses != 2 {
		t.Fatalf("the loader parsed %d lines, want 2", parses)
	}

	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4}
	for i := 0; i < 3; i++ {
		if _, err := eng.RunFunctionByID("demo:counted", src); err != nil {
			t.Fatalf("RunFunctionByID: %v", err)
		}
	}
	if parses != 2 {
		t.Errorf("the runs parsed %d extra lines, want none", parses-2)
	}
	if runs != 6 {
		t.Errorf("the command ran %d times, want 6", runs)
	}
}

// TestReloadRebuildsTheCache covers the invalidation.
//
// `/reload` empties the function registry and loads the functions again,
// so the cached commands are dropped together with the functions they
// belong to, and the new functions are parsed again.
func TestReloadRebuildsTheCache(t *testing.T) {
	root := t.TempDir()
	writeCacheFunction(t, root, "demo", "counted", "cachecount a\ncachecount b\n")

	parses, runs := 0, 0
	reg := exec.NewRegistry()
	reg.Register(&cacheCountCmd{parses: &parses, runs: &runs})
	eng := exec.NewEngine(reg, world.NewWorld())
	eng.SetRoot(root)
	if errs := eng.LoadFunctions(); len(errs) != 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	if parses != 2 {
		t.Fatalf("the loader parsed %d lines, want 2", parses)
	}
	first := eng.Functions.Get("demo:counted")

	// A reload empties the registry and loads the functions again.
	if errs := eng.LoadFunctions(); len(errs) != 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	if parses != 4 {
		t.Errorf("the reload parsed %d lines in total, want 4", parses)
	}
	second := eng.Functions.Get("demo:counted")
	if second == nil {
		t.Fatal("the function was not registered again")
	}
	if first == second {
		t.Error("the reload should have built a new function")
	}

	// The reloaded function still runs.
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4}
	if _, err := eng.RunFunctionByID("demo:counted", src); err != nil {
		t.Fatalf("RunFunctionByID: %v", err)
	}
	if runs != 2 {
		t.Errorf("the command ran %d times, want 2", runs)
	}
	if parses != 4 {
		t.Errorf("the run parsed %d extra lines, want none", parses-4)
	}
}

// TestCachedFunctionResolvesRelativeCoordinatesAtRunTime covers the
// coordinates.
//
// A `~` is parsed against zero and resolved against the position the
// command runs at, so the cached form has to keep resolving it against
// the position of the current run.
func TestCachedFunctionResolvesRelativeCoordinatesAtRunTime(t *testing.T) {
	root := t.TempDir()
	writeCacheFunction(t, root, "demo", "spawn", "setworldspawn ~ ~ ~\n")
	eng := newCacheEngine(t)
	eng.SetRoot(root)
	if errs := eng.LoadFunctions(); len(errs) != 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	fn := eng.Functions.Get("demo:spawn")
	if fn == nil {
		t.Fatal("the function was not registered")
	}
	if len(fn.Commands) != 1 || fn.Commands[0] == nil {
		t.Fatal("the line should have been cached")
	}

	for _, pos := range []world.Vec{{X: 1, Y: 2, Z: 3}, {X: 10, Y: 20, Z: 30}} {
		src := &exec.Source{
			Name: "Server", World: "world", PermissionLevel: 4,
			Pos: pos,
		}
		if _, err := eng.RunFunctionByID("demo:spawn", src); err != nil {
			t.Fatalf("RunFunctionByID: %v", err)
		}
		want := pos.BlockPos()
		if eng.World.Spawn != want {
			t.Errorf("the world spawn is %v, want %v", eng.World.Spawn, want)
		}
	}
}

// TestCachedFunctionBehavesLikeTheUncachedOne compares the two paths.
//
// The commands of a function are parsed once and then reused, so they have
// to do exactly what they did when every run parsed them again.
func TestCachedFunctionBehavesLikeTheUncachedOne(t *testing.T) {
	root := t.TempDir()
	body := "say hello\nsetblock 1 2 3 stone\nsetblock 1 2 3\nlist\n"
	writeCacheFunction(t, root, "demo", "same", body)
	eng := newCacheEngine(t)
	eng.SetRoot(root)
	if errs := eng.LoadFunctions(); len(errs) != 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}

	// The same lines, run one by one.
	direct := newCacheEngine(t)
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	var directErrs []error
	for i, line := range strings.Split(strings.TrimSpace(body), "\n") {
		if _, err := direct.ExecuteLine(line, i+1, src); err != nil {
			directErrs = append(directErrs, err)
		}
	}

	// The same lines, run as a function.
	if _, err := eng.RunFunctionByID("demo:same", src); err != nil {
		t.Fatalf("RunFunctionByID: %v", err)
	}

	if len(eng.Feedback) != len(direct.Feedback) {
		t.Fatalf("the function produced %v, want %v", eng.Feedback, direct.Feedback)
	}
	for i := range direct.Feedback {
		if eng.Feedback[i] != direct.Feedback[i] {
			t.Errorf("the feedback %d = %q, want %q", i, eng.Feedback[i], direct.Feedback[i])
		}
	}
	fnErrs := eng.TakeErrors()
	if len(fnErrs) != len(directErrs) {
		t.Fatalf("the function collected %v, want %v", fnErrs, directErrs)
	}
	for i := range directErrs {
		if fnErrs[i].Error() != directErrs[i].Error() {
			t.Errorf("the error %d = %q, want %q", i, fnErrs[i], directErrs[i])
		}
	}
	// The block was placed in both runs.
	if eng.World.Blocks.Get(world.Pos{X: 1, Y: 2, Z: 3}).Name != "stone" {
		t.Error("the function did not place the block")
	}
	if direct.World.Blocks.Get(world.Pos{X: 1, Y: 2, Z: 3}).Name != "stone" {
		t.Error("the direct run did not place the block")
	}
}
