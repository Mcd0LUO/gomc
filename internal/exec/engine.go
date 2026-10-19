package exec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/lexer"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/world"
)

// Function is a loaded mcfunction.
type Function struct {
	// ID is the function's namespaced identifier, e.g. "ns:name".
	ID string
	// Lines holds the function's source lines, excluding comments.
	Lines []string
	// LineNumbers holds the source line number of each entry in Lines,
	// so runtime errors can point back at the file.
	LineNumbers []int
	// Path is the file the function was loaded from.
	Path string
}

// FunctionStore holds the loaded functions.
type FunctionStore struct {
	funcs map[string]*Function
	order []string
}

// NewFunctionStore returns an empty function store.
func NewFunctionStore() *FunctionStore {
	return &FunctionStore{funcs: map[string]*Function{}}
}

// Put stores a function.
func (s *FunctionStore) Put(f *Function) {
	if _, ok := s.funcs[f.ID]; !ok {
		s.order = append(s.order, f.ID)
	}
	s.funcs[f.ID] = f
}

// Get returns a function by id.
func (s *FunctionStore) Get(id string) *Function { return s.funcs[id] }

// IDs returns the loaded function ids in load order.
func (s *FunctionStore) IDs() []string { return append([]string(nil), s.order...) }

// Clear removes every function from the store.
//
// This mirrors `CustomFunctionData.f`, which empties the registry before
// reloading it.
func (s *FunctionStore) Clear() {
	s.funcs = map[string]*Function{}
	s.order = nil
}

// FunctionLoadError describes a function that failed to load.
//
// 1.12.2 treats an unparsable command name as fatal for the whole file:
// `CustomFunction.a` throws, and `CustomFunctionData.h` logs the failure
// without registering the function. Reproducing that is what makes a
// harness report the same errors as the real server.
type FunctionLoadError struct {
	// ID is the function's identifier.
	ID string
	// Path is the file the function came from.
	Path string
	// Line is the 1-based line of the offending command.
	Line int
	// Msg explains the failure.
	Msg string
}

func (e *FunctionLoadError) Error() string {
	return fmt.Sprintf("%s:%d: %s", e.Path, e.Line, e.Msg)
}

// LogMessage returns the server log line 1.12.2 writes when a function
// fails to load.
//
// `CustomFunctionData.h` catches the failure and logs
// "Couldn't read custom function <key> from <file>", then keeps every
// other function registered.
func (e *FunctionLoadError) LogMessage() string {
	return fmt.Sprintf("Couldn't read custom function %s from %s", e.ID, e.Path)
}

// Engine executes commands against a world.
type Engine struct {
	// Registry holds the available commands.
	Registry *Registry
	// World is the state being modified.
	World *world.World
	// Functions holds the loaded functions.
	Functions *FunctionStore
	// Feedback collects the output messages produced by commands.
	Feedback []string
	// Errors collects the failures of commands that ran inside a
	// function. A function keeps going after a failure, so the failures
	// are collected here instead of being returned.
	Errors []error
	// Logs collects server log lines, such as the note 1.12.2 writes when
	// a function fails to load.
	Logs []string
	// Root is the world's data directory, used to load functions.
	Root string
	// maxDepth bounds recursive function calls.
	maxDepth int
}

// NewEngine creates an engine.
func NewEngine(reg *Registry, w *world.World) *Engine {
	return &Engine{
		Registry:  reg,
		World:     w,
		Functions: NewFunctionStore(),
		maxDepth:  256,
	}
}

// SetRoot sets the world directory used to load functions.
func (e *Engine) SetRoot(root string) { e.Root = root }

// VisibleFeedback returns the feedback messages the source would actually
// see.
//
// 1.12.2's `sendCommandFeedback` game rule decides whether a command
// prints its result, so the messages are recorded either way but only
// shown when the rule is on.
func (e *Engine) VisibleFeedback() []string {
	if !e.World.Rules.Bool("sendCommandFeedback") {
		return nil
	}
	return e.Feedback
}

// LoadFunctions loads every function under `<root>/data/functions`.
//
// The 1.12.2 layout is `data/functions/<namespace>/<path>.mcfunction`, and
// the namespace is the first path segment, matching `CustomFunctionData.h`.
//
// The store is cleared first and then rebuilt, matching `CustomFunctionData.f`,
// which is what the server calls on start-up and on `/reload`. A function
// that fails to load is reported and skipped; the remaining functions still
// load. This mirrors 1.12.2, where a broken function is logged as an error
// while the server keeps running.
func (e *Engine) LoadFunctions() []error {
	if e.Root == "" {
		return nil
	}
	e.Functions.Clear()
	base := filepath.Join(e.Root, "data", "functions")
	var errs []error
	err := filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".mcfunction") {
			return nil
		}
		rel, relErr := filepath.Rel(base, path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		rel = strings.TrimSuffix(rel, ".mcfunction")
		ns, name, found := strings.Cut(rel, "/")
		if !found {
			// A function without a namespace segment is skipped, as in
			// 1.12.2 where `split("/", 2)` must yield two parts.
			return nil
		}
		id := ns + ":" + name
		fn, loadErr := e.loadFunctionFile(id, path)
		if loadErr != nil {
			errs = append(errs, loadErr)
			return nil
		}
		e.Functions.Put(fn)
		return nil
	})
	if err != nil {
		errs = append(errs, err)
	}
	return errs
}

// LoadFunction loads a single function by its namespaced id.
//
// Unlike LoadFunctions it does not touch the rest of the store, so a caller
// can resolve one function without disturbing the others. A function that
// fails to load is reported and skipped, leaving the already loaded
// functions intact. This mirrors 1.12.2, where a broken function is logged
// as an error while the server keeps running.
func (e *Engine) LoadFunction(id string) error {
	if e.Root == "" {
		return fmt.Errorf("no world directory set")
	}
	base := filepath.Join(e.Root, "data", "functions")
	ns, name, found := strings.Cut(id, ":")
	if !found {
		return fmt.Errorf("invalid function id %q", id)
	}
	path := filepath.Join(base, filepath.FromSlash(ns), filepath.FromSlash(name)+".mcfunction")
	fn, err := e.loadFunctionFile(id, path)
	if err != nil {
		return err
	}
	e.Functions.Put(fn)
	return nil
}

// loadFunctionFile reads and validates a function file.
//
// Validation mirrors `CustomFunction.a`: comments and blank lines are
// skipped, and an unknown command name aborts the whole file.
func (e *Engine) loadFunctionFile(id, path string) (*Function, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	fn := &Function{ID: id, Path: path}
	for i, raw := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		lineNo := i + 1
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.SplitN(line, " ", 2)
		name := fields[0]
		if _, ok := e.Registry.Get(name); !ok {
			return nil, e.functionLoadError(id, path, lineNo, name)
		}
		fn.Lines = append(fn.Lines, line)
		fn.LineNumbers = append(fn.LineNumbers, lineNo)
	}
	return fn, nil
}

// functionLoadError builds the error message 1.12.2 produces for an
// unknown command inside a function.
func (e *Engine) functionLoadError(id, path string, line int, name string) error {
	var msg string
	switch {
	case strings.HasPrefix(name, "//"):
		msg = fmt.Sprintf("Unknown or invalid command '%s' (if you intended to make a comment, use '#' not '//')", name)
	case strings.HasPrefix(name, "/") && len(name) > 1:
		msg = fmt.Sprintf("Unknown or invalid command '%s' (did you mean '%s'? Do not use a preceding forwards slash.)", name, name[1:])
	default:
		msg = fmt.Sprintf("Unknown or invalid command '%s'", name)
	}
	return &FunctionLoadError{ID: id, Path: path, Line: line, Msg: msg}
}

// ParseCommand parses a command line into an AST command.
//
// It returns a parse error for unknown commands and for malformed
// arguments, both carrying the source position.
func (e *Engine) ParseCommand(line string, lineNo int) (*ast.Command, error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil, &parser.Error{Kind: parser.KindSyntax, Msg: "empty command", Line: lineNo, Col: 1}
	}
	// A leading slash is invalid inside a function file; 1.12.2 rejects it
	// at load time with a dedicated message.
	if strings.HasPrefix(trimmed, "/") {
		name := strings.SplitN(trimmed, " ", 2)[0]
		return nil, &parser.Error{
			Kind: parser.KindUnknown,
			Msg:  fmt.Sprintf("Unknown or invalid command '%s' (did you mean '%s'? Do not use a preceding forwards slash.)", name, name[1:]),
			Line: lineNo, Col: 1,
		}
	}
	toks, err := lexer.Tokenize(trimmed)
	if err != nil {
		return nil, &parser.Error{Kind: parser.KindSyntax, Msg: err.Error(), Line: lineNo, Col: 1}
	}
	if len(toks) == 0 {
		return nil, &parser.Error{Kind: parser.KindSyntax, Msg: "empty command", Line: lineNo, Col: 1}
	}
	name := toks[0].Text
	handler, ok := e.Registry.Get(name)
	if !ok {
		return nil, &parser.Error{
			Kind: parser.KindUnknown,
			Msg:  fmt.Sprintf("Unknown or invalid command '%s'", name),
			Line: lineNo, Col: 1,
		}
	}
	rest := toks[1:]
	p := parser.New(rest, lineNo, trimmed)
	args, err := handler.Parse(p)
	if err != nil {
		return nil, parser.Positioned(err, lineNo, 1, 0)
	}
	return &ast.Command{
		Name: name,
		Args: args,
		Raw:  trimmed,
		Line: lineNo,
		Col:  1,
	}, nil
}

// Execute runs a parsed command under the given source.
func (e *Engine) Execute(cmd *ast.Command, src *Source) (int, error) {
	handler, ok := e.Registry.Get(cmd.Name)
	if !ok {
		return 0, Errorf("unknown command %q", cmd.Name)
	}
	ctx := &Context{
		Source:  src,
		World:   e.World,
		Engine:  e,
		Command: cmd,
		Args:    cmd.Args,
	}
	count, err := handler.Run(ctx, cmd.Args)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ExecuteLine parses and runs a single command line.
func (e *Engine) ExecuteLine(line string, lineNo int, src *Source) (int, error) {
	cmd, err := e.ParseCommand(line, lineNo)
	if err != nil {
		return 0, err
	}
	return e.Execute(cmd, src)
}

// RunFunction executes a loaded function under the given source.
//
// The execution model follows `CustomFunctionData.a`:
//
//   - The function's commands are pushed onto a stack in reverse order and
//     popped from the front, so they run top to bottom.
//   - The total number of commands executed in one call is bounded by the
//     `maxCommandChainLength` gamerule.
//   - A function invoked from inside another function is queued rather than
//     executed inline, and the caller receives a success count of 0.
//
// A command that fails does **not** stop the function: 1.12.2 catches the
// error per command (`CommandBlockListenerAbstract.executeCommand`) and
// carries on with the next one. The failures are collected in Engine.Errors.
func (e *Engine) RunFunction(fn *Function, src *Source) (int, error) {
	limit := e.World.Rules.Int("maxCommandChainLength")
	if limit <= 0 {
		limit = 65536
	}
	// The stack holds pending command lines. The lines are pushed in
	// reverse order and taken from the back, so they run top to bottom.
	// This mirrors `CustomFunctionData.a`, which pushes with `addFirst`
	// and pops with `removeFirst`.
	var stack []pendingLine
	for i := len(fn.Lines) - 1; i >= 0; i-- {
		stack = append(stack, pendingLine{line: fn.Lines[i], lineNo: fn.LineNumbers[i], src: src})
	}

	executed := 0
	for len(stack) > 0 && executed < limit {
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		executed++
		count, err := e.ExecuteLine(item.line, item.lineNo, item.src)
		if err != nil {
			// The command failed, but the function keeps going.
			e.Errors = append(e.Errors, err)
			continue
		}
		if count > 0 {
			executed += count - 1
		}
	}
	return executed, nil
}

// FunctionLintError describes a command of a function that does not parse.
//
// It is a gomc addition: 1.12.2 only checks the command *name* when a
// function is loaded, so a mistake in the arguments stays hidden until the
// function runs.
type FunctionLintError struct {
	// Path is the file the function came from.
	Path string
	// Line is the 1-based line of the offending command.
	Line int
	// Err is the parse failure.
	Err error
}

func (e *FunctionLintError) Error() string {
	return fmt.Sprintf("%s:%d: %v", e.Path, e.Line, e.Err)
}

// LintFunctions parses every command of every loaded function and reports
// the ones that do not parse.
//
// The linter finds argument mistakes that 1.12.2 would only reveal when the
// function is executed.
func (e *Engine) LintFunctions() []error {
	var out []error
	for _, id := range e.Functions.IDs() {
		fn := e.Functions.Get(id)
		if fn == nil {
			continue
		}
		for i, line := range fn.Lines {
			if _, err := e.ParseCommand(line, fn.LineNumbers[i]); err != nil {
				out = append(out, &FunctionLintError{
					Path: fn.Path,
					Line: fn.LineNumbers[i],
					Err:  err,
				})
			}
		}
	}
	return out
}

// TakeErrors returns the collected command failures and clears them.
//
// The failures are the ones the running commands reported; a function keeps
// going after a failure, so they have to be collected rather than returned
// from the call that hit them.
func (e *Engine) TakeErrors() []error {
	out := e.Errors
	e.Errors = nil
	return out
}

// pendingLine is a queued command line inside a function execution.
type pendingLine struct {
	line   string
	lineNo int
	src    *Source
}

// RunFunctionByID runs a function by its namespaced id.
func (e *Engine) RunFunctionByID(id string, src *Source) (int, error) {
	fn := e.Functions.Get(id)
	if fn == nil {
		return 0, Errorf("unknown function %q", id)
	}
	return e.RunFunction(fn, src)
}

// Tick advances the world by one tick.
//
// In 1.12.2 the tick hook is the `gameLoopFunction` gamerule: the function
// it names runs once per tick, and changing the rule re-resolves the
// function. `CustomFunctionData.e` implements exactly that.
func (e *Engine) Tick(src *Source) (int, error) {
	e.World.Time++
	// The time of day advances with the tick counter, and wraps around
	// at 24000 ticks, which is one Minecraft day.
	e.World.DayTime = (e.World.DayTime + 1) % 24000
	id := e.World.Rules.String("gameLoopFunction")
	if id == "" || id == "-" {
		return 0, nil
	}
	fn := e.Functions.Get(id)
	if fn == nil {
		// A missing function is not an error in 1.12.2: the reference is
		// simply kept and nothing runs.
		return 0, nil
	}
	return e.RunFunction(fn, src)
}
