package exec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/world"
)

// recorder is a test command that records the lines it was given.
type recorder struct {
	lines *[]string
}

// Name implements Handler.
func (r *recorder) Name() string { return "record" }

// Usage implements Handler.
func (r *recorder) Usage() string { return "/record <text>" }

// Parse implements Handler.
func (r *recorder) Parse(p *parser.Parser) ([]ast.Arg, error) {
	return []ast.Arg{&ast.StringArg{Value: p.Rest()}}, nil
}

// Run implements Handler.
func (r *recorder) Run(ctx *Context, args []ast.Arg) (int, error) {
	*r.lines = append(*r.lines, args[0].(*ast.StringArg).Value)
	return 1, nil
}

// failCmd is a command that loads fine but always fails when it runs.
type failCmd struct{}

// Name implements Handler.
func (c *failCmd) Name() string { return "fail" }

// Usage implements Handler.
func (c *failCmd) Usage() string { return "/fail" }

// Parse implements Handler.
func (c *failCmd) Parse(p *parser.Parser) ([]ast.Arg, error) { return nil, nil }

// Run implements Handler.
func (c *failCmd) Run(ctx *Context, args []ast.Arg) (int, error) {
	return 0, Errorf("this command always fails")
}

// twoArgsCmd is a command that needs exactly two arguments.
type twoArgsCmd struct{}

// Name implements Handler.
func (c *twoArgsCmd) Name() string { return "twoargs" }

// Usage implements Handler.
func (c *twoArgsCmd) Usage() string { return "/twoargs <a> <b>" }

// Parse implements Handler.
func (c *twoArgsCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	a, err := p.String()
	if err != nil {
		return nil, err
	}
	b, err := p.String()
	if err != nil {
		return nil, err
	}
	return []ast.Arg{&ast.StringArg{Value: a}, &ast.StringArg{Value: b}}, nil
}

// Run implements Handler.
func (c *twoArgsCmd) Run(ctx *Context, args []ast.Arg) (int, error) { return 1, nil }

// newRecordingEngine builds an engine whose only command records its input.
func newRecordingEngine(t *testing.T) (*Engine, *[]string) {
	t.Helper()
	var lines []string
	reg := NewRegistry()
	reg.Register(&recorder{lines: &lines})
	reg.Register(&failCmd{})
	reg.Register(&twoArgsCmd{})
	return NewEngine(reg, world.NewWorld()), &lines
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

func TestFunctionRunsCommandsTopToBottom(t *testing.T) {
	root := t.TempDir()
	writeFunction(t, root, "demo", "seq", "record one\nrecord two\nrecord three\n")
	e, lines := newRecordingEngine(t)
	e.SetRoot(root)
	if errs := e.LoadFunctions(); len(errs) > 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}
	if _, err := e.RunFunctionByID("demo:seq", src); err != nil {
		t.Fatalf("RunFunctionByID: %v", err)
	}
	want := []string{"one", "two", "three"}
	if len(*lines) != len(want) {
		t.Fatalf("recorded %v, want %v", *lines, want)
	}
	for i := range want {
		if (*lines)[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, (*lines)[i], want[i])
		}
	}
}

// TestBrokenFunctionDoesNotBlockItsSiblings covers 1.12.2's per-file
// error handling.
//
// `CustomFunctionData.h` catches the failure of one file, logs it and
// keeps every other function registered.
func TestBrokenFunctionDoesNotBlockItsSiblings(t *testing.T) {
	root := t.TempDir()
	writeFunction(t, root, "demo", "good", "record fine\n")
	writeFunction(t, root, "demo", "bad", "record fine\nnosuchcommand arg\n")
	e, _ := newRecordingEngine(t)
	e.SetRoot(root)
	errs := e.LoadFunctions()
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
	if e.Functions.Get("demo:good") == nil {
		t.Error("the healthy function was not registered")
	}
	if e.Functions.Get("demo:bad") != nil {
		t.Error("the broken function must not be registered")
	}
}

// TestLoadFunctionsClearsTheStore covers the reload semantics.
//
// `CustomFunctionData.f` empties the registry before reloading it, so a
// function that no longer exists disappears after a reload.
func TestLoadFunctionsClearsTheStore(t *testing.T) {
	root := t.TempDir()
	writeFunction(t, root, "demo", "gone", "record x\n")
	e, _ := newRecordingEngine(t)
	e.SetRoot(root)
	if errs := e.LoadFunctions(); len(errs) > 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	if e.Functions.Get("demo:gone") == nil {
		t.Fatal("the function was not loaded")
	}
	if err := os.Remove(filepath.Join(root, "data", "functions", "demo", "gone.mcfunction")); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if errs := e.LoadFunctions(); len(errs) > 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	if e.Functions.Get("demo:gone") != nil {
		t.Error("the removed function survived the reload")
	}
}

// TestFunctionContinuesAfterAFailure covers the 1.12.2 error handling.
//
// A command that fails inside a function does not stop the function:
// `CommandBlockListenerAbstract.executeCommand` catches the failure per
// command and carries on with the next one.
func TestFunctionContinuesAfterAFailure(t *testing.T) {
	root := t.TempDir()
	writeFunction(t, root, "demo", "keepgoing",
		"fail\nrecord after\n")
	e, lines := newRecordingEngine(t)
	e.SetRoot(root)
	if errs := e.LoadFunctions(); len(errs) > 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	if _, err := e.RunFunctionByID("demo:keepgoing", src); err != nil {
		t.Fatalf("RunFunctionByID: %v", err)
	}
	if len(*lines) != 1 || (*lines)[0] != "after" {
		t.Errorf("recorded %v, want the command after the failure to have run", *lines)
	}
	errs := e.TakeErrors()
	if len(errs) != 1 {
		t.Errorf("collected %d errors, want 1", len(errs))
	}
}

// TestLinterFindsArgumentMistakes covers the extra check `check` does.
//
// 1.12.2 only validates the command name when a function is loaded, so an
// argument mistake stays hidden until the function runs.
func TestLinterFindsArgumentMistakes(t *testing.T) {
	root := t.TempDir()
	writeFunction(t, root, "demo", "badargs", "record one\ntwoargs only-one\n")
	e, _ := newRecordingEngine(t)
	e.SetRoot(root)
	if errs := e.LoadFunctions(); len(errs) > 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	// The function loads, because 1.12.2 only checks the command name.
	if e.Functions.Get("demo:badargs") == nil {
		t.Fatal("the function should have loaded")
	}
	lint := e.LintFunctions()
	if len(lint) != 1 {
		t.Fatalf("the linter found %d problems, want 1: %v", len(lint), lint)
	}
	if !strings.Contains(lint[0].Error(), ":2:") {
		t.Errorf("the report = %q, want it to point at line 2", lint[0].Error())
	}
}

func TestUnknownCommandIsRejected(t *testing.T) {
	e, _ := newRecordingEngine(t)
	_, err := e.ParseCommand("nosuchcommand foo", 3)
	if err == nil {
		t.Fatal("expected an error for an unknown command")
	}
	perr, ok := err.(*parser.Error)
	if !ok {
		t.Fatalf("error type = %T, want *parser.Error", err)
	}
	if perr.Line != 3 {
		t.Errorf("line = %d, want 3", perr.Line)
	}
}

func TestLeadingSlashIsRejected(t *testing.T) {
	e, _ := newRecordingEngine(t)
	_, err := e.ParseCommand("/record x", 1)
	if err == nil {
		t.Fatal("expected an error for a leading slash")
	}
	if !strings.Contains(err.Error(), "Do not use a preceding forwards slash") {
		t.Errorf("error = %q, want the forward slash hint", err.Error())
	}
}

func TestCommentsAndBlankLinesAreSkipped(t *testing.T) {
	root := t.TempDir()
	writeFunction(t, root, "demo", "comments", "# a comment\n\nrecord one\n   \n#another\nrecord two\n")
	e, lines := newRecordingEngine(t)
	e.SetRoot(root)
	if errs := e.LoadFunctions(); len(errs) > 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}
	if _, err := e.RunFunctionByID("demo:comments", src); err != nil {
		t.Fatalf("RunFunctionByID: %v", err)
	}
	if len(*lines) != 2 {
		t.Errorf("recorded %v, want two entries", *lines)
	}
}

// TestTickRunsTheGameLoopFunction covers the 1.12.2 tick hook.
//
// The `gameLoopFunction` gamerule names the function that runs once per
// tick, and `CustomFunctionData.e` re-resolves it when the rule changes.
func TestTickRunsTheGameLoopFunction(t *testing.T) {
	root := t.TempDir()
	writeFunction(t, root, "demo", "loop", "record tick\n")
	e, lines := newRecordingEngine(t)
	e.SetRoot(root)
	if errs := e.LoadFunctions(); len(errs) > 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}

	// Without the rule nothing runs.
	for i := 0; i < 3; i++ {
		if _, err := e.Tick(src); err != nil {
			t.Fatalf("Tick: %v", err)
		}
	}
	if len(*lines) != 0 {
		t.Errorf("recorded %v, want nothing", *lines)
	}

	if err := e.World.Rules.Set("gameLoopFunction", "demo:loop"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	for i := 0; i < 3; i++ {
		if _, err := e.Tick(src); err != nil {
			t.Fatalf("Tick: %v", err)
		}
	}
	if len(*lines) != 3 {
		t.Errorf("recorded %v, want three entries", *lines)
	}
	if e.World.Time != 6 {
		t.Errorf("world time = %d, want 6", e.World.Time)
	}
}

// TestTickWithUnknownFunctionIsSilent covers a game loop rule that points
// at a function that does not exist.
func TestTickWithUnknownFunctionIsSilent(t *testing.T) {
	e, _ := newRecordingEngine(t)
	if err := e.World.Rules.Set("gameLoopFunction", "demo:missing"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}
	if _, err := e.Tick(src); err != nil {
		t.Fatalf("Tick: %v", err)
	}
}
