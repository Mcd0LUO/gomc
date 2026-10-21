package exec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/lexer"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// isUsageMessage reports whether an error message is already the usage line
// of a command.
//
// 1.12.2 answers a usage mistake with the usage line itself: sometimes
// with the "Usage: " prefix of `commands.generic.usage`, sometimes as a
// bare line. The error report of the runtime must not say it a second
// time as its own "usage:" line.
//
// The bare form is recognised by its shape: a usage line starts with a
// slash and spells out its arguments in angle brackets. No other message
// of the runtime starts with a slash.
func isUsageMessage(msg string) bool {
	if strings.HasPrefix(msg, "Usage: ") {
		return true
	}
	return strings.HasPrefix(msg, "/") && strings.Contains(msg, "<")
}

// ExtensionUse records a command of a function that gomc knows but 1.12.2
// does not.
type ExtensionUse struct {
	// Function is the id of the function that uses the command.
	Function string
	// Path is the file the function was loaded from.
	Path string
	// Line is the line the command is on.
	Line int
	// Command is the name of the command.
	Command string
	// Note says what the command is for.
	Note string
}

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
	// Commands holds the parsed form of each entry in Lines, in the same
	// order.
	//
	// The loader parses every line once and keeps the result here, so a
	// command is not parsed again on every run of the function. A nil
	// entry means the line was not cached: either it does not parse, or
	// the function was not built by the loader. Such a line is parsed
	// when it runs, which is where 1.12.2 reports the mistake as well.
	//
	// The cached commands are shared by every run of the function, so a
	// handler must not modify the arguments it is given.
	Commands []*ast.Command
}

// cachedCommand returns the parsed form of the line at index i.
//
// It returns nil when the line was not cached, in which case the caller
// has to parse it.
func (f *Function) cachedCommand(i int) *ast.Command {
	if i < 0 || i >= len(f.Commands) {
		return nil
	}
	return f.Commands[i]
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
//
// The cached commands of a function belong to the function, so they are
// dropped with it: nothing keeps pointing at a function that is gone.
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
	// Source is the text of the offending line.
	Source string
	// Hint is an extra suggestion, such as "did you mean 'setblock'?".
	Hint string
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
	// broken holds the functions that failed to load, with the lines that
	// could be read.
	//
	// 1.12.2 refuses such a function as a whole, so it is not registered
	// and cannot run. The lines are kept anyway, so the linter can still
	// check them: one unknown command name should not hide every other
	// mistake in the file.
	broken []*Function
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
	// ExtensionUses holds the commands of the loaded functions that gomc
	// knows but 1.12.2 does not.
	//
	// 1.12.2 refuses a function whose command it does not know, so a
	// function that uses one of these loads on gomc but not on a real
	// server. The check reports them, so the difference is never silent.
	ExtensionUses []ExtensionUse
	// queue holds the pending entries of the running function call chain.
	//
	// 1.12.2 keeps a single deque in `CustomFunctionData.g` and shares it
	// with every function of the chain, so a nested call is queued in the
	// same deque instead of getting a stack of its own.
	queue []pendingEntry
	// running reports whether a function call chain is active.
	//
	// It is `CustomFunctionData.h`: while it is set, a call to
	// `CustomFunctionData.a` queues the function and reports 0.
	running bool
}

// NewEngine creates an engine.
func NewEngine(reg *Registry, w *world.World) *Engine {
	// The random source starts from the seed the runtime was given, so
	// that a run does not depend on how many random numbers the runs
	// before it took.
	//
	// The source is shared by everything that is randomised — the `@r`
	// selector and the default duration of `/weather` — which is what
	// 1.12.2 does as well: both take from the random source of the
	// world.
	SetSeed(Seed())
	return &Engine{
		Registry:  reg,
		World:     w,
		Functions: NewFunctionStore(),
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
	e.ExtensionUses = nil
	e.broken = nil
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
		id := NormalizeFunctionID(ns + ":" + name)
		fn, loadErr := e.loadFunctionFile(id, path)
		if loadErr != nil {
			errs = append(errs, loadErr)
			// The function is not registered — 1.12.2 refuses it as a
			// whole — but its lines are kept, so the linter still
			// checks them.
			if fn != nil {
				e.broken = append(e.broken, fn)
			}
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
//
// The rest of the file is read anyway, so the linter can check the other
// lines: one unknown command name should not hide every other mistake in
// the file. Only the **first** unknown command name is reported, because
// that is the one 1.12.2 throws on.
//
// Every other line is parsed as well, and the parsed form is kept, so
// the function does not have to parse its commands again on every run.
// A line that does not parse is *not* a load error: 1.12.2 only checks
// the command name when a function is loaded, so a mistake in the
// arguments stays hidden until the function runs. Such a line is simply
// left out of the cache and parsed when it runs.
func (e *Engine) loadFunctionFile(id, path string) (*Function, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	fn := &Function{ID: id, Path: path}
	var loadErr error
	for i, raw := range splitLines(string(data)) {
		lineNo := i + 1
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.SplitN(line, " ", 2)
		name := fields[0]
		if _, ok := e.Registry.Get(name); !ok {
			if loadErr == nil {
				loadErr = e.functionLoadError(id, path, lineNo, name, line)
			}
			// The line is left out, but the file is read to the end so
			// the linter sees the lines around it.
			continue
		}
		if note, ok := e.Registry.Extension(name); ok {
			// The command is one gomc adds on top of the 1.12.2 set.
			// 1.12.2 would refuse the whole function, so the use is
			// recorded and reported by the check.
			e.ExtensionUses = append(e.ExtensionUses, ExtensionUse{
				Function: id, Path: path, Line: lineNo, Command: name, Note: note,
			})
		}
		fn.Lines = append(fn.Lines, line)
		fn.LineNumbers = append(fn.LineNumbers, lineNo)
		cmd, parseErr := e.ParseCommand(line, lineNo)
		if parseErr != nil {
			// The line is kept, but not its parsed form: the failure
			// is reported when the function runs.
			fn.Commands = append(fn.Commands, nil)
			continue
		}
		fn.Commands = append(fn.Commands, cmd)
	}
	return fn, loadErr
}

// functionLoadError builds the error message 1.12.2 produces for an
// unknown command inside a function.
func (e *Engine) functionLoadError(id, path string, line int, name, source string) error {
	var msg string
	switch {
	case strings.HasPrefix(name, "//"):
		msg = fmt.Sprintf("Unknown or invalid command '%s' (if you intended to make a comment, use '#' not '//')", name)
	case strings.HasPrefix(name, "/") && len(name) > 1:
		msg = fmt.Sprintf("Unknown or invalid command '%s' (did you mean '%s'? Do not use a preceding forwards slash.)", name, name[1:])
	default:
		msg = fmt.Sprintf("Unknown or invalid command '%s'", name)
	}
	loadErr := &FunctionLoadError{ID: id, Path: path, Line: line, Msg: msg, Source: source}
	if suggestion, found := ClosestName(name, e.Registry.Names()); found {
		loadErr.Hint = fmt.Sprintf("did you mean '%s'?", suggestion)
	}
	return loadErr
}

// ParseCommand parses a command line into an AST command.
//
// It returns a parse error for unknown commands and for malformed
// arguments, both carrying the source position.
//
// A relative coordinate is measured from the origin; use ParseCommandAt
// for a command that runs somewhere else.
func (e *Engine) ParseCommand(line string, lineNo int) (*ast.Command, error) {
	return e.ParseCommandAt(line, lineNo, parser.CoordBase{})
}

// lintBase is the position a relative coordinate in a function file is
// measured from when the file is **checked**.
//
// 1.12.2 measures it against the position of **whoever runs the function**,
// and the checker does not know that. The ground level is the best guess: a
// function is usually called by a player or a command block standing on the
// ground, and the difference decides the answer -- `~-1` is one block below
// the ground (y=4 → 3) but below the world when it is measured from the
// origin (y=0 → -1).
//
// (Measured: a real project's `blockdata ~ ~-1 ~ {...}` is refused by the
// origin and accepted by the ground, and the server accepts it when the
// function is called from the ground.)
var lintBase = parser.CoordBase{X: 0, Y: 4, Z: 0}

// ParseCommandAt parses a command line that runs at the given base into
// an AST command.
//
// The base is the block position the command runs at: 1.12.2 measures a
// relative coordinate against it, and the bounds of the coordinate are
// checked on the sum of the base and the offset.
func (e *Engine) ParseCommandAt(line string, lineNo int, base parser.CoordBase) (*ast.Command, error) {
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
	toks := lexer.Tokenize(trimmed)
	if len(toks) == 0 {
		return nil, &parser.Error{Kind: parser.KindSyntax, Msg: "empty command", Line: lineNo, Col: 1}
	}
	name := toks[0].Text
	handler, ok := e.Registry.Get(name)
	if !ok {
		err := &parser.Error{
			Kind:   parser.KindUnknown,
			Msg:    fmt.Sprintf("Unknown or invalid command '%s'", name),
			Line:   lineNo,
			Col:    1,
			Source: trimmed,
		}
		// A typo in the command name is the most common mistake in a
		// function file, so point at the command that was probably meant.
		if suggestion, found := ClosestName(name, e.Registry.Names()); found {
			err.Hint = fmt.Sprintf("did you mean '%s'?", suggestion)
		}
		return nil, err
	}
	rest := toks[1:]
	p := parser.NewAt(rest, lineNo, trimmed, base)
	args, err := handler.Parse(p)
	if err != nil {
		perr := parser.Positioned(err, lineNo, 1, 0)
		if pe, ok := perr.(*parser.Error); ok {
			if pe.Source == "" {
				pe.Source = trimmed
			}
			if pe.Usage == "" && !isUsageMessage(pe.Msg) {
				pe.Usage = handler.Usage()
			}
		}
		return nil, perr
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
	// 1.12.2 asks the command whether the source may run it before it
	// runs it (`ICommand.canUse`, which is `listener.a(level, name)`).
	// A source that is not high enough is refused with
	// `commands.generic.permission`.
	if level := e.Registry.RequiredPermission(cmd.Name); level > src.PermissionLevel {
		return 0, Errorf("You do not have permission to use this command")
	}
	ctx := &Context{
		Source:  src,
		World:   e.World,
		Engine:  e,
		Command: cmd,
		Args:    cmd.Args,
		// A command that does not report a count reports nothing, and
		// -1 is what says so.
		AffectedBlocks:   -1,
		AffectedEntities: -1,
		AffectedItems:    -1,
		QueryResult:      -1,
	}
	count, err := handler.Run(ctx, cmd.Args)
	if err != nil {
		// The command failed while running; the usage of the command is
		// still worth showing.
		if ee, ok := err.(*Error); ok && ee.Usage == "" && !isUsageMessage(ee.Msg) {
			ee.Usage = handler.Usage()
		}
		return 0, err
	}
	// The outcome of the command is reported, so that a version that has
	// a way of putting the result of a command somewhere can do it.
	//
	// 1.12.2 counts the commands that ran in the `SuccessCount` channel
	// of `/stats`, and the dispatcher reports the count of the
	// entities a command affected there as well.
	if reporter := e.Registry.ResultReporter(); reporter != nil {
		reporter.Report(ctx, count, ctx.AffectedBlocks, ctx.AffectedEntities, ctx.AffectedItems, ctx.QueryResult)
	}
	return count, nil
}

// ExecuteLine parses and runs a single command line.
func (e *Engine) ExecuteLine(line string, lineNo int, src *Source) (int, error) {
	cmd, err := e.ParseCommandAt(line, lineNo, coordBaseOf(src))
	if err != nil {
		return 0, err
	}
	return e.Execute(cmd, src)
}

// coordBaseOf returns the position a relative coordinate is measured from
// under the given source.
//
// It is the block position of the source, which is what 1.12.2 measures a
// relative coordinate against.
func coordBaseOf(src *Source) parser.CoordBase {
	if src == nil {
		return parser.CoordBase{}
	}
	b := src.RelativeBase()
	return parser.CoordBase{X: b.X, Y: b.Y, Z: b.Z}
}

// RunFunction executes a loaded function under the given source.
//
// The execution model follows `CustomFunctionData.a`:
//
//   - The function's commands are pushed onto a queue in reverse order and
//     popped from the front, so they run top to bottom.
//   - A function called from inside a running function is queued in that
//     same queue instead of being executed inline, and the caller is told
//     that 0 commands ran.
//   - The `maxCommandChainLength` game rule is the budget of the whole
//     call chain, not of a single call: the loop runs at most that many
//     times, and a queued function only contributes the commands that
//     still fit in the budget. The rest are dropped.
//   - The commands of a function run at permission level 2, because
//     `CommandFunction.execute` caps the listener with `.a(2)` and the
//     game loop function's listener answers `i <= 2`.
//
// The budget is spent by the loop counter, and the loop of
// `CustomFunctionData.a` is a do-while, so a limit of 0 still runs one
// command.
//
// A command that fails does **not** stop the function: 1.12.2 catches the
// error per command (`CommandBlockListenerAbstract.executeSafely`) and
// carries on with the next one. The failures are collected in
// Engine.Errors.
//
// The commands of the function are parsed when it is loaded, so a run does
// not parse them again. A line that does not parse is parsed when it runs,
// which is where 1.12.2 reports the mistake as well.
func (e *Engine) RunFunction(fn *Function, src *Source) (int, error) {
	limit := e.World.Rules.Int("maxCommandChainLength")
	// The commands of a function run at permission level 2. The cap is
	// the lower of the two levels, so a source that is already lower
	// keeps its own level.
	inner := src.WithPermissionLevel(2)
	if e.running {
		// A nested call: 1.12.2 queues the function when the queue is
		// not full yet and reports 0 commands to the caller.
		if len(e.queue) < limit {
			e.queue = append(e.queue, pendingEntry{fn: fn, src: inner})
		}
		return 0, nil
	}
	e.running = true
	defer func() {
		// The queue belongs to the chain, so it is dropped when the chain
		// ends. `CustomFunctionData.a` clears the deque in a finally
		// block.
		e.queue = nil
		e.running = false
	}()
	e.pushFunction(fn, inner)
	executed := 0
	for {
		if len(e.queue) == 0 {
			return executed, nil
		}
		item := e.queue[len(e.queue)-1]
		e.queue = e.queue[:len(e.queue)-1]
		if item.fn != nil {
			e.expandFunction(item, limit)
		} else {
			e.runEntry(item)
		}
		executed++
		// The counter is the number of entries that were popped, which is
		// what `CustomFunctionData.a` reports.
		if executed >= limit {
			return executed, nil
		}
	}
}

// pushFunction queues the commands of a function.
//
// The commands are pushed in reverse order so that the first one is
// popped first: `CustomFunctionData.a` pushes with `addFirst` and pops
// with `removeFirst`.
func (e *Engine) pushFunction(fn *Function, src *Source) {
	for i := len(fn.Lines) - 1; i >= 0; i-- {
		e.queue = append(e.queue, pendingEntry{
			line:   fn.Lines[i],
			lineNo: fn.LineNumbers[i],
			path:   fn.Path,
			src:    src,
			cmd:    fn.cachedCommand(i),
		})
	}
}

// expandFunction queues the commands of a function that was popped from
// the queue.
//
// `CustomFunction.d` only contributes the commands that still fit in the
// budget of the chain: the rest are dropped.
func (e *Engine) expandFunction(item pendingEntry, limit int) {
	remaining := limit - len(e.queue)
	n := len(item.fn.Lines)
	if remaining < n {
		n = remaining
	}
	for i := n - 1; i >= 0; i-- {
		e.queue = append(e.queue, pendingEntry{
			line:   item.fn.Lines[i],
			lineNo: item.fn.LineNumbers[i],
			path:   item.fn.Path,
			src:    item.src,
			cmd:    item.fn.cachedCommand(i),
		})
	}
}

// runEntry runs one queued command line.
//
// A command that fails does not stop the function: the failure is
// collected and the next entry runs.
//
// The line is parsed again only when it was not cached, which is the
// case for a line that does not parse and for a function that was not
// built by the loader.
func (e *Engine) runEntry(item pendingEntry) {
	cmd := item.cmd
	if cmd == nil {
		parsed, err := e.ParseCommand(item.line, item.lineNo)
		if err != nil {
			e.Errors = append(e.Errors, e.decorate(err, item.lineNo, item.line, item.path))
			return
		}
		cmd = parsed
	}
	if _, err := e.Execute(cmd, item.src); err != nil {
		e.Errors = append(e.Errors, e.decorate(err, item.lineNo, item.line, item.path))
	}
}

// decorate adds the source context to an error.
//
// The engine knows which file and which line a command came from, so a
// failure inside a function can point at it. Errors that already carry a
// position keep it.
func (e *Engine) decorate(err error, lineNo int, source, path string) error {
	switch t := err.(type) {
	case *parser.Error:
		if t.Line == 0 {
			t.Line = lineNo
		}
		if t.Source == "" {
			t.Source = source
		}
		if t.Path == "" {
			t.Path = path
		}
	case *Error:
		if t.Line == 0 {
			t.Line = lineNo
		}
		if t.Source == "" {
			t.Source = source
		}
		if t.Path == "" {
			t.Path = path
		}
	}
	return err
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
		out = append(out, e.lintFunction(fn)...)
	}
	// The functions that failed to load are linted too: 1.12.2 refuses
	// them as a whole, but the mistakes in the lines around the unknown
	// command name are worth reporting all the same.
	for _, fn := range e.broken {
		out = append(out, e.lintFunction(fn)...)
	}
	return out
}

// splitLines splits a function file into lines the way Java's
// `BufferedReader.readLine` does.
//
// A line ends at a line feed, a carriage return, or the two of them
// together -- and a file may mix them. That is not a rare curiosity: a
// real project has function files whose lines end with a lone carriage
// return (the old Mac convention) next to lines that end with CRLF.
//
// The server reads the file with Guava's `Files.readLines`, which uses
// `BufferedReader.readLine`, so it sees the same lines gomc does. If the
// carriage return were not a line ending, two commands would be read as
// one, and the second one would be parsed as an argument of the first.
//
// (Measured on a real 1.12.2 server: a function file with a lone CR
// between two commands loads as two commands, and the second one runs.)
func splitLines(data string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case '\n':
			lines = append(lines, data[start:i])
			start = i + 1
		case '\r':
			lines = append(lines, data[start:i])
			// A CRLF is one line ending, not two.
			if i+1 < len(data) && data[i+1] == '\n' {
				i++
			}
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}

// lintFunction checks the lines of one function.
func (e *Engine) lintFunction(fn *Function) []error {
	var out []error
	for i, line := range fn.Lines {
		// The linter does not know who will run the function, so it
		// measures a relative coordinate from the ground.
		cmd, err := e.ParseCommandAt(line, fn.LineNumbers[i], lintBase)
		if err != nil {
			out = append(out, &FunctionLintError{
				Path: fn.Path,
				Line: fn.LineNumbers[i],
				Err:  err,
			})
			continue
		}
		// A JSON message that is not JSON is a mistake 1.12.2 only
		// reveals when the command runs, and it reports it after the
		// target of the command. So it is not a failure of the parsing,
		// but it is worth reporting all the same.
		for _, arg := range cmd.Args {
			j, ok := arg.(*ast.JSONArg)
			if !ok || j.Err == nil {
				continue
			}
			out = append(out, &FunctionLintError{
				Path: fn.Path,
				Line: fn.LineNumbers[i],
				Err:  j.Err,
			})
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

// pendingEntry is a queued item of a function call chain.
//
// It is either a command line or a function that is expanded when it is
// popped, which is what `CustomFunctionData.a` keeps in its deque.
type pendingEntry struct {
	// line is the command line to run. It is empty for a function entry.
	line string
	// lineNo is the source line of the command line.
	lineNo int
	// path is the file the command line came from.
	path string
	// fn is the function to expand. It is nil for a command entry.
	fn *Function
	// src is the source the entry runs under.
	src *Source
	// cmd is the parsed form of the command line, when the line was
	// cached. It is nil for a function entry and for a line that has to
	// be parsed when it runs.
	cmd *ast.Command
}

// RunFunctionByID runs a function by its namespaced id.
func (e *Engine) RunFunctionByID(id string, src *Source) (int, error) {
	fn := e.Functions.Get(NormalizeFunctionID(id))
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
	fn := e.Functions.Get(NormalizeFunctionID(id))
	if fn == nil {
		// A missing function is not an error in 1.12.2: the reference is
		// simply kept and nothing runs.
		return 0, nil
	}
	return e.RunFunction(fn, src)
}
