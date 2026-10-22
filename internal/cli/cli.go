// Package cli implements the gomc command line interface.
//
// The tool has three subcommands:
//
//	gomc check <world>   validate every function in a world directory
//	gomc run   <world> <function> [ticks]
//	gomc tick  <world> [ticks]
//
// The world directory is a Minecraft 1.12.2 save folder. Functions are read
// from `<world>/data/functions/<namespace>/<name>.mcfunction`, which is the
// 1.12.2 layout.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/buildinfo"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/version/v1_12_2"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// Usage prints the top level help.
func Usage(w io.Writer) {
	fmt.Fprint(w, `gomc - a Minecraft 1.12.2 mcfunction test runner

Usage:
  gomc check <world>                    validate the functions in a world
  gomc run   <world> <function>         run a function (default: 1 tick)
  gomc tick  <world> [ticks]            advance the world by N ticks
  gomc commands                         list the supported commands
  gomc bench <world>                    measure a project: what it runs and how long it takes
  gomc version                          print the version

  gomc difftest run    <world> <case>   print what a case does
  gomc difftest check  <world> <case>   compare a case with its golden file
  gomc difftest update <world> <case>   rewrite the golden file of a case
  gomc difftest probe  <case>           write the probe function for a real server
  gomc difftest list   [dir]            list the cases and their golden files

The world argument is a Minecraft 1.12.2 save directory. Functions are
loaded from <world>/data/functions/<namespace>/<name>.mcfunction.

Switches (accepted anywhere):
  --plain-errors        one line per error, for scripts
  --json                one JSON object per error, for scripts
  --as <player>         run the function as that player instead of the console

Every subcommand explains itself with -h.
`)
}

// usageCheck prints the help of the check subcommand.
func usageCheck(w io.Writer) {
	fmt.Fprint(w, `gomc check - validate the functions in a world

Usage:
  gomc check <world>

It reports two kinds of problem:

  - a command name 1.12.2 does not know. The server refuses the whole
    function when it loads it, so the function cannot run at all.
  - a mistake in the arguments. 1.12.2 only reveals it when the function
    runs; gomc checks it while loading, so it is found before the first
    run.

The second kind is gomc's own addition. The message of every problem is
the one 1.12.2 would say; the position, the usage line and the
suggestion are added by gomc and are not part of it.

Exit code: 0 when nothing is wrong, 1 when there is at least one problem.
`)
}

// usageRun prints the help of the run subcommand.
func usageRun(w io.Writer) {
	fmt.Fprint(w, `gomc run - run a function in a world

Usage:
  gomc run <world> <function> [--as <player>]

The function is identified the way 1.12.2 identifies it: "namespace:name",
both parts lower-cased, and a missing namespace means "minecraft".

--as <player> runs the function as that player instead of as the console.
The player appears at the world spawn, and it is the listener of the
commands: 1.12.2 measures a relative coordinate against the listener's
block position, and a player's is half a block higher than the block it
stands in. The player has operator permissions, because 1.12.2 refuses
"/function" for a player without them.

A broken sibling function does not stop the function under test: 1.12.2
logs it and keeps every other function registered.

Exit code: 0 when the function ran, 1 when a command inside it failed.
`)
}

// usageTick prints the help of the tick subcommand.
func usageTick(w io.Writer) {
	fmt.Fprint(w, `gomc tick - advance the world by N ticks

Usage:
  gomc tick <world> [ticks]

Every tick runs the function named by the "gameLoopFunction" game rule,
the way 1.12.2 does. Without the rule, or with a function that does not
exist, the tick does nothing and says nothing.

Exit code: 0 unless a tick failed.
`)
}

// usageBench prints the help of the bench subcommand.
func usageBench(w io.Writer) {
	fmt.Fprint(w, `gomc bench - measure a project

Usage:
  gomc bench <world> [switches]

It loads the functions of the world the way the server loads them, runs
them, and reports what happened: how many commands there were, which
commands were used, how long each function took, and what failed.

The numbers are the numbers of gomc, so they say how the project behaves
under this runtime.

What is run:

  (default)             the gameLoopFunction game rule, for ten ticks.
                        This is the function a server runs on its own, so
                        it is the one that shows the real behaviour.

  --function <id>       one function, by its id ("namespace:name").
  --ticks <n>           n ticks of the game loop function.
  --all                 every function of the world, one after the other.
                        The functions are not called from each other, so
                        this measures each one **on its own** -- and a
                        function that needs what another one set up
                        fails. A high failure count here usually means
                        the project is built out of functions that call
                        each other, not that the functions are broken.

Switches:

  --as <player>         run as that player instead of the console.
  --top <n>             how many functions the detailed report shows
                        (20 by default).
  --json                print the report as JSON, for scripts.

Exit code: 0 when the measurement ran, 1 when it could not.
`)
}

// usageCommands prints the help of the commands subcommand.
func usageCommands(w io.Writer) {
	fmt.Fprint(w, `gomc commands - list the supported commands

Usage:
  gomc commands

The list is grouped by domain, and every command comes with the usage
line 1.12.2 prints for it, so it is also the syntax of the command.
`)
}

// usageDifftest prints the help of the difftest subcommand.
func usageDifftest(w io.Writer) {
	fmt.Fprint(w, `gomc difftest - the differential testing layer

Usage:
  gomc difftest run    <world> <case>   print what a case does
  gomc difftest check  <world> <case>   compare a case with its golden file
  gomc difftest update <world> <case>   rewrite the golden file of a case
  gomc difftest probe  <case>           write the probe function for a real server
  gomc difftest list   [dir]            list the cases and their golden files

A case is a small scenario: a "setup" section that prepares the world and
a "run" section with the commands under test. Running it produces the
canonical observation of what happened; the golden file holds that
observation, and "check" compares the two.

"probe" writes the same case as a function for a real 1.12.2 server, so
the two can be compared command by command.

"check" and "run" take --full, which adds the state of the world after
the run to the observation.
`)
}

// options holds the global switches.
type options struct {
	// plainErrors turns off the enhanced error output, leaving the
	// compact one-liner that scripts can parse.
	plainErrors bool
	// json prints the errors as one JSON object per line.
	json bool
	// as is the player a function runs as, empty for the console.
	as string
	// err is the mistake in the switches themselves.
	err error
}

// parseOptions takes the global switches out of the argument list.
//
// The switches are accepted anywhere, so both "gomc --plain-errors check
// world" and "gomc check world --plain-errors" work.
func parseOptions(args []string) (options, []string) {
	var opts options
	rest := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--plain-errors":
			opts.plainErrors = true
		case "--json":
			opts.json = true
		case "--as":
			if i+1 >= len(args) {
				if opts.err == nil {
					opts.err = fmt.Errorf("--as needs the name of a player")
				}
				continue
			}
			i++
			opts.as = args[i]
		default:
			rest = append(rest, args[i])
		}
	}
	return opts, rest
}

// wantsHelp reports whether the arguments ask for the help of a
// subcommand.
func wantsHelp(args []string) bool {
	for _, a := range args {
		if a == "-h" || a == "--help" {
			return true
		}
	}
	return false
}

// Run dispatches a command line and returns the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		Usage(stdout)
		return 2
	}
	opts, args := parseOptions(args)
	if opts.err != nil {
		fmt.Fprintf(stderr, "error: %v\n", opts.err)
		return 2
	}
	if len(args) == 0 {
		Usage(stdout)
		return 2
	}
	switch args[0] {
	case "check":
		if wantsHelp(args[1:]) {
			usageCheck(stdout)
			return 0
		}
		return runCheck(args[1:], stdout, stderr, opts)
	case "run":
		if wantsHelp(args[1:]) {
			usageRun(stdout)
			return 0
		}
		return runRun(args[1:], stdout, stderr, opts)
	case "tick":
		if wantsHelp(args[1:]) {
			usageTick(stdout)
			return 0
		}
		return runTick(args[1:], stdout, stderr, opts)
	case "commands":
		if wantsHelp(args[1:]) {
			usageCommands(stdout)
			return 0
		}
		return runCommands(stdout)
	case "bench":
		if wantsHelp(args[1:]) {
			usageBench(stdout)
			return 0
		}
		return runBench(args[1:], stdout, stderr, opts)
	case "difftest":
		return runDifftest(args[1:], stdout, stderr, opts)
	case "version", "--version", "-v":
		fmt.Fprintln(stdout, buildinfo.String())
		return 0
	case "help", "-h", "--help":
		Usage(stdout)
		return 0
	}
	fmt.Fprintf(stderr, "unknown subcommand %q\n\n", args[0])
	Usage(stderr)
	return 2
}

// runCheck validates every function in a world.
//
// It reports two kinds of problem: functions that fail to load (an unknown
// command name, which 1.12.2 also rejects) and commands that do not parse
// (a mistake in the arguments, which 1.12.2 only reveals when the function
// runs).
func runCheck(args []string, stdout, stderr io.Writer, opts options) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "check requires a world directory")
		return 2
	}
	engine, err := newEngine(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	errs := engine.LoadFunctions()
	lint := engine.LintFunctions()
	uses := engine.ExtensionUses
	if len(errs) == 0 && len(lint) == 0 && len(uses) == 0 {
		fmt.Fprintf(stdout, "ok: %d function(s) loaded\n", len(engine.Functions.IDs()))
		return 0
	}
	for _, e := range errs {
		reportError(stderr, e, opts)
	}
	for _, e := range lint {
		reportError(stderr, e, opts)
	}
	for _, u := range uses {
		reportExtensionUse(stderr, u, opts)
	}
	if len(errs)+len(lint) > 0 {
		if opts.json {
			printJSON(stderr, jsonError{
				Level: "error",
				Kind:  "Summary",
				Msg:   fmt.Sprintf("%d error(s)", len(errs)+len(lint)),
			})
		} else {
			fmt.Fprintf(stderr, "%d error(s)\n", len(errs)+len(lint))
		}
		return 1
	}
	return 0
}

// reportExtensionUse reports a command of a function that gomc knows but
// 1.12.2 does not.
//
// The command works on gomc, but 1.12.2 refuses the whole function when it
// loads it, so the difference has to be visible: a function that only runs
// on gomc is not a function that can be tested on a real server.
func reportExtensionUse(w io.Writer, u exec.ExtensionUse, opts options) {
	if opts.json {
		printJSON(w, jsonError{
			Level: "warning",
			Kind:  "Extension",
			Path:  u.Path,
			Line:  u.Line,
			Msg:   fmt.Sprintf("%q is a gomc-only command: %s", u.Command, u.Note),
		})
		return
	}
	if opts.plainErrors {
		fmt.Fprintf(w, "warning: %s:%d: gomc-only command %q\n", u.Path, u.Line, u.Command)
		return
	}
	fmt.Fprintf(w, "warning: %s:%d\n", u.Path, u.Line)
	fmt.Fprintf(w, "  %q is a gomc-only command: %s\n", u.Command, u.Note)
	fmt.Fprintf(w, "  a real 1.12.2 server does not know it and refuses the function\n")
}

// jsonError is the machine readable form of a problem.
//
// The fields are the ones the other formats carry, so a script can read
// where the problem is without parsing the text.
type jsonError struct {
	// Level is "error" or "warning".
	Level string `json:"level"`
	// Kind is the class of the problem, e.g. "Syntax" or "Unknown".
	Kind string `json:"kind"`
	// Path is the file the command came from.
	Path string `json:"path,omitempty"`
	// Line is the 1-based line number.
	Line int `json:"line,omitempty"`
	// Col is the 1-based column number.
	Col int `json:"col,omitempty"`
	// Msg is what went wrong.
	Msg string `json:"msg"`
	// Usage is the usage line of the command that failed.
	Usage string `json:"usage,omitempty"`
	// Hint is an extra suggestion.
	Hint string `json:"hint,omitempty"`
}

// reportJSON prints a problem as one JSON object per line.
func reportJSON(w io.Writer, err error, level string) {
	d, ok := detailOf(err)
	if !ok {
		printJSON(w, jsonError{Level: level, Kind: "Error", Msg: err.Error()})
		return
	}
	printJSON(w, jsonError{
		Level: level,
		Kind:  d.kind,
		Path:  d.path,
		Line:  d.line,
		Col:   d.col,
		Msg:   d.msg,
		Usage: d.usage,
		Hint:  d.hint,
	})
}

// printJSON writes one JSON object.
func printJSON(w io.Writer, e jsonError) {
	data, err := json.Marshal(e)
	if err != nil {
		// The struct has nothing that cannot be marshalled, so this
		// cannot happen; the message is still better than a panic.
		fmt.Fprintf(w, "{\"level\":%q,\"kind\":%q,\"msg\":%q}\n", e.Level, e.Kind, e.Msg)
		return
	}
	fmt.Fprintln(w, string(data))
}

// runRun executes a function in a world.
func runRun(args []string, stdout, stderr io.Writer, opts options) int {
	if len(args) < 2 {
		fmt.Fprintln(stderr, "run requires a world directory and a function id")
		return 2
	}
	engine, err := newEngine(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	id := args[1]
	// 1.12.2 identifies a function with a `MinecraftKey`: both parts are
	// lower-cased and a missing namespace means "minecraft".
	id = exec.NormalizeFunctionID(id)
	// The world is loaded as a whole, and a function that fails to load is
	// only skipped. That is what 1.12.2 does, so a broken sibling must not
	// stop the function under test from running.
	loadErrs := engine.LoadFunctions()
	if engine.Functions.Get(id) == nil {
		// The target itself did not load. Re-resolve it alone so the
		// failure names the real cause instead of "unknown function".
		if err := engine.LoadFunction(id); err != nil {
			reportError(stderr, err, opts)
			return 1
		}
		fmt.Fprintf(stderr, "error: Unknown function '%s'\n", id)
		return 1
	}
	for _, e := range loadErrs {
		reportWarning(stderr, e, opts)
	}
	src := &exec.Source{
		Name:            "Server",
		Pos:             world.Vec{},
		World:           "world",
		PermissionLevel: 4,
		Feedback:        true,
	}
	if opts.as != "" {
		// The function runs as a player, so the player is the listener
		// of the commands.
		src, err = playerSource(engine, opts.as)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
	}
	count, err := engine.RunFunctionByID(id, src)
	if err != nil {
		reportError(stderr, err, opts)
		return 1
	}
	// A command that fails inside a function does not stop the function,
	// so the failures are collected and reported here.
	failed := false
	for _, e := range engine.TakeErrors() {
		reportError(stderr, e, opts)
		failed = true
	}
	for _, msg := range engine.Logs {
		fmt.Fprintln(stdout, msg)
	}
	for _, msg := range engine.VisibleFeedback() {
		fmt.Fprintln(stdout, msg)
	}
	fmt.Fprintf(stdout, "ran %d command(s) from %s\n", count, id)
	if failed {
		return 1
	}
	return 0
}

// runTick advances the world by N ticks.
func runTick(args []string, stdout, stderr io.Writer, opts options) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "tick requires a world directory")
		return 2
	}
	ticks := 1
	if len(args) > 1 {
		n, err := parseInt(args[1])
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 2
		}
		ticks = n
	}
	engine, err := newEngine(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	// A broken function is skipped rather than fatal, so the tick loop keeps
	// going exactly as it does on a real server.
	for _, e := range engine.LoadFunctions() {
		reportWarning(stderr, e, opts)
	}
	src := &exec.Source{
		Name:            "Server",
		World:           "world",
		PermissionLevel: 4,
		Feedback:        true,
	}
	for i := 0; i < ticks; i++ {
		if _, err := engine.Tick(src); err != nil {
			reportError(stderr, err, opts)
			return 1
		}
	}
	for _, msg := range engine.Logs {
		fmt.Fprintln(stdout, msg)
	}
	for _, msg := range engine.VisibleFeedback() {
		fmt.Fprintln(stdout, msg)
	}
	fmt.Fprintf(stdout, "advanced %d tick(s)\n", ticks)
	return 0
}

// runCommands lists the supported commands.
func runCommands(stdout io.Writer) int {
	reg := exec.NewRegistry()
	v1_12_2.Register(reg)
	groups := v1_12_2.CommandGroups()
	total := 0
	for _, g := range groups {
		fmt.Fprintf(stdout, "%s:\n", g.Name)
		for _, name := range g.Commands {
			h, ok := reg.Get(name)
			if !ok {
				continue
			}
			fmt.Fprintf(stdout, "  %-16s %s\n", name, h.Usage())
			total++
		}
	}
	fmt.Fprintf(stdout, "\n%d command(s)\n", total)
	return 0
}

// playerSource returns a source that runs as the given player.
//
// The player is put at the world spawn, where a player who joins a server
// appears, and it is the **listener** of the commands: 1.12.2 measures a
// relative coordinate against the listener's block position, and a
// player's is half a block higher than the block it stands in
// (`EntityPlayer.getChunkCoordinates`).
//
// The level is the operator's: 1.12.2 refuses "/function" for a player
// without the permission, so the function could not be run at all.
func playerSource(engine *exec.Engine, name string) (*exec.Source, error) {
	player, ok := engine.World.EntityTypes.NewEntity("player", "")
	if !ok {
		return nil, fmt.Errorf("the player entity type is missing")
	}
	player.PlayerName = name
	spawn := engine.World.Spawn
	// A player who joins appears in the middle of the spawn block.
	player.Pos = world.Vec{
		X: float64(spawn.X) + 0.5,
		Y: float64(spawn.Y),
		Z: float64(spawn.Z) + 0.5,
	}
	engine.World.AddEntity(player)
	return &exec.Source{
		Name:            name,
		Entity:          player,
		Pos:             player.Pos,
		ListenerPlayer:  true,
		World:           "world",
		PermissionLevel: 4,
		Feedback:        true,
	}, nil
}

// newEngine builds an engine for a world directory.
func newEngine(root string) (*exec.Engine, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("cannot open world %q: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", root)
	}
	reg := exec.NewRegistry()
	v1_12_2.Register(reg)
	engine := exec.NewEngine(reg, world.NewWorld())
	engine.SetRoot(root)
	return engine, nil
}

// reportError prints an error.
//
// The enhanced form is the one a compiler would print: the path and
// position, the offending line with a caret under the column, what went
// wrong, and then the usage of the command and a suggestion when there is
// one. The plain form is a single compact line, for scripts that parse
// the output.
//
// The message itself is never changed: it stays what 1.12.2 would say.
func reportError(w io.Writer, err error, opts options) {
	if opts.json {
		reportJSON(w, err, "error")
		return
	}
	if !opts.plainErrors && reportDetailed(w, err) {
		return
	}
	switch e := err.(type) {
	case *parser.Error:
		fmt.Fprintf(w, "error[%s]: %d:%d: %s\n", e.Kind, e.Line, e.Col, e.Msg)
	case *exec.FunctionLoadError:
		fmt.Fprintf(w, "error[Unknown]: %s:%d: %s\n", e.Path, e.Line, e.Msg)
	case *exec.FunctionLintError:
		fmt.Fprintf(w, "error[Syntax]: %s\n", e.Error())
	case *exec.Error:
		fmt.Fprintf(w, "error[Runtime]: %s\n", e.Msg)
	case *os.PathError:
		fmt.Fprintf(w, "error[IO]: %s: %s\n", e.Path, e.Err)
	default:
		fmt.Fprintf(w, "error: %v\n", err)
	}
}

// errorDetail is the context an error can carry.
type errorDetail struct {
	// kind is the error class shown in brackets.
	kind string
	// path is the file the command came from.
	path string
	// line is the 1-based line number.
	line int
	// col is the 1-based column number.
	col int
	// source is the text of the offending line.
	source string
	// msg is what went wrong.
	msg string
	// usage is the usage line of the command that failed.
	usage string
	// hint is an extra suggestion.
	hint string
}

// detailOf extracts the context from an error.
//
// It reports false for the errors that carry no context worth showing,
// such as a file that could not be read.
func detailOf(err error) (errorDetail, bool) {
	switch e := err.(type) {
	case *parser.Error:
		return errorDetail{
			kind: e.Kind.String(), path: e.Path, line: e.Line, col: e.Col,
			source: e.Source, msg: e.Msg, usage: e.Usage, hint: e.Hint,
		}, true
	case *exec.Error:
		return errorDetail{
			kind: "Runtime", path: e.Path, line: e.Line, col: e.Col,
			source: e.Source, msg: e.Msg, usage: e.Usage, hint: e.Hint,
		}, true
	case *exec.FunctionLoadError:
		return errorDetail{
			kind: "Unknown", path: e.Path, line: e.Line, col: 1,
			source: e.Source, msg: e.Msg, hint: e.Hint,
		}, true
	case *exec.FunctionLintError:
		d, ok := detailOf(e.Err)
		if !ok {
			return errorDetail{}, false
		}
		d.kind = "Syntax"
		if d.path == "" {
			d.path = e.Path
		}
		if d.line == 0 {
			d.line = e.Line
		}
		return d, true
	}
	return errorDetail{}, false
}

// reportDetailed prints the enhanced form of an error.
//
// It reports false when the error carries nothing to show, so the caller
// can fall back to the compact form.
func reportDetailed(w io.Writer, err error) bool {
	return reportDetailedKind(w, err, "error")
}

// reportDetailedKind prints the enhanced form of an error with the given
// prefix, "error" or "warning".
func reportDetailedKind(w io.Writer, err error, prefix string) bool {
	d, ok := detailOf(err)
	if !ok || d.msg == "" {
		return false
	}
	loc := ""
	switch {
	case d.path != "" && d.line > 0 && d.col > 0:
		loc = fmt.Sprintf("%s:%d:%d", d.path, d.line, d.col)
	case d.path != "" && d.line > 0:
		loc = fmt.Sprintf("%s:%d", d.path, d.line)
	case d.path != "":
		loc = d.path
	case d.line > 0 && d.col > 0:
		loc = fmt.Sprintf("%d:%d", d.line, d.col)
	}
	if loc == "" {
		fmt.Fprintf(w, "%s[%s]: %s\n", prefix, d.kind, d.msg)
	} else {
		fmt.Fprintf(w, "%s[%s]: %s\n", prefix, d.kind, loc)
	}
	if d.source != "" {
		fmt.Fprintf(w, "  %s\n", d.source)
		if d.col > 0 {
			// The caret sits under the offending column.
			caret := strings.Repeat(" ", d.col-1)
			if d.col-1 > len(d.source) {
				caret = strings.Repeat(" ", len(d.source))
			}
			fmt.Fprintf(w, "  %s^\n", caret)
		}
	}
	fmt.Fprintf(w, "  %s\n", d.msg)
	if d.usage != "" {
		fmt.Fprintf(w, "  usage: %s\n", d.usage)
	}
	if d.hint != "" {
		fmt.Fprintf(w, "  hint: %s\n", d.hint)
	}
	return true
}

// reportWarning prints a non-fatal problem.
//
// 1.12.2 logs a function that fails to load and keeps every other function
// registered, so a broken file is a warning for everything except `check`.
func reportWarning(w io.Writer, err error, opts options) {
	if opts.json {
		reportJSON(w, err, "warning")
		return
	}
	if !opts.plainErrors && reportDetailedKind(w, err, "warning") {
		return
	}
	switch e := err.(type) {
	case *exec.FunctionLoadError:
		fmt.Fprintf(w, "warning[Unknown]: %s:%d: %s\n", e.Path, e.Line, e.Msg)
	case *os.PathError:
		fmt.Fprintf(w, "warning[IO]: %s: %s\n", e.Path, e.Err)
	default:
		fmt.Fprintf(w, "warning: %v\n", err)
	}
}

// parseInt parses a decimal integer.
func parseInt(s string) (int, error) {
	var n int
	if _, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &n); err != nil {
		return 0, fmt.Errorf("expected an integer, got %q", s)
	}
	return n, nil
}

// WorldFunctions returns the function ids found in a world directory.
func WorldFunctions(root string) ([]string, error) {
	base := filepath.Join(root, "data", "functions")
	var out []string
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
		rel = strings.TrimSuffix(filepath.ToSlash(rel), ".mcfunction")
		if ns, name, ok := strings.Cut(rel, "/"); ok {
			out = append(out, ns+":"+name)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}
