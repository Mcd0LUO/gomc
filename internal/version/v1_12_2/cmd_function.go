package v1_12_2

import (
	"errors"
	"fmt"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
)

// functionArgs is the parsed form of `/function`.
type functionArgs struct {
	// id is the function's namespaced identifier.
	id string
	// condition is "if" or "unless" when a condition was given.
	condition string
	// conditionTarget is the selector the condition applies to.
	conditionTarget ast.Arg
}

// functionArg is the AST wrapper for the parsed function arguments.
type functionArg struct {
	// Value holds the parsed arguments.
	Value *functionArgs
}

// ArgKind implements ast.Arg.
func (a *functionArg) ArgKind() ast.Kind { return ast.KindList }

// functionCmd implements the 1.12.2 `/function`.
//
// The 1.12.2 syntax is `function <id> [if|unless <target>]`, where the
// optional condition decides whether the function runs based on whether the
// target selector matches anything. This predates the 1.13 rewrite and is
// not documented on most modern pages, but it is present in the 1.12.2
// server code (`CommandFunction.execute`).
type functionCmd struct{}

// Name implements exec.Handler.
func (c *functionCmd) Name() string { return "function" }

// Usage implements exec.Handler.
func (c *functionCmd) Usage() string { return usageLine("function") }

// Parse implements exec.Handler.
func (c *functionCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	id, err := p.String()
	if err != nil {
		return nil, err
	}
	// 1.12.2 identifies a function with a `MinecraftKey`: both parts are
	// lower-cased and a missing namespace means "minecraft".
	a := &functionArgs{id: exec.NormalizeFunctionID(id)}
	if p.More() {
		cond, err := p.String()
		if err != nil {
			return nil, err
		}
		if cond != "if" && cond != "unless" {
			return nil, p.Errorf("expected if or unless, got %q", cond)
		}
		target, err := parseEntityTarget(p)
		if err != nil {
			return nil, err
		}
		a.condition = cond
		a.conditionTarget = target
	}
	return []ast.Arg{&functionArg{Value: a}}, nil
}

// Run implements exec.Handler.
func (c *functionCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	a := args[0].(*functionArg).Value
	if a.condition != "" {
		// 1.12.2 catches the "entity not found" error and treats it as
		// "the condition did not match", so a name that resolves to
		// nobody simply makes the condition false.
		hasMatch := false
		if entities, err := resolveEntityTargets(ctx, a.conditionTarget); err == nil {
			hasMatch = len(entities) > 0
		}
		if (a.condition == "if") != hasMatch {
			// The condition was not met. 1.12.2 reports that the
			// function was skipped instead of running it silently.
			return 0, exec.Errorf("Skipped execution of function '%s'", a.id)
		}
	}
	fn := ctx.Engine.Functions.Get(a.id)
	if fn == nil {
		return 0, exec.Errorf("Unknown function '%s'", a.id)
	}
	// The function is queued rather than run inline when it is called
	// from inside another function, and `CustomFunctionData.a` then
	// reports 0 commands to the caller. The engine runs the function's
	// commands at permission level 2, exactly like the listener
	// `CommandFunction.execute` builds.
	count, err := ctx.Engine.RunFunction(fn, ctx.Source)
	if err != nil {
		return 0, err
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Executed %d command(s) from function '%s'", count, a.id))
	// The command itself did not fail, so it counts as one success. The
	// number of commands the function ran is only part of the message:
	// 1.12.2 reports it through `commands.function.success` and leaves
	// the success count of the command at one.
	return 1, nil
}

// reloadCmd implements `/reload`.
type reloadCmd struct{}

// Name implements exec.Handler.
func (c *reloadCmd) Name() string { return "reload" }

// Usage implements exec.Handler.
func (c *reloadCmd) Usage() string { return usageLine("reload") }

// Parse implements exec.Handler.
func (c *reloadCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	return nil, nil
}

// Run implements exec.Handler.
func (c *reloadCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	// `MinecraftServer.reload` calls `CustomFunctionData.f`, which clears
	// the registry and reloads it file by file. A file that fails to load
	// is logged and skipped, so the reload itself still succeeds.
	for _, err := range ctx.Engine.LoadFunctions() {
		var loadErr *exec.FunctionLoadError
		if errors.As(err, &loadErr) {
			ctx.Engine.Logs = append(ctx.Engine.Logs, loadErr.LogMessage())
			continue
		}
		ctx.Engine.Logs = append(ctx.Engine.Logs, err.Error())
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback, "Successfully reloaded loot tables, advancements and functions")
	return 1, nil
}

// registerFunction adds the function-related commands.
func registerFunction(reg *exec.Registry) {
	reg.Register(&functionCmd{})
	reg.Register(&reloadCmd{})
	registerPermissions(reg)
}

// commandPermissions is the permission level every 1.12.2 command needs.
//
// 1.12.2 asks the command whether the source may run it
// (`ICommand.canUse`, which is `listener.a(level, name)`), and the command
// answers with the level it declares. `CommandAbstract` declares 4 and the
// commands that need less override it, so the table below holds the
// declarations of the 1.12.2 commands gomc implements.
//
// The levels matter for functions: a function runs its commands at level
// 2, so a level 3 command such as `/reload` is refused inside a function,
// exactly as it is on a real server.
var commandPermissions = map[string]int{
	// Level 0: everybody may run them.
	"help":    0,
	"list":    0,
	"me":      0,
	"trigger": 0,
	// Level 1: `say` only has to be able to talk.
	"say": 1,
	// Level 2: the commands an operator may use.
	"advancement":     2,
	"blockdata":       2,
	"clear":           2,
	"clone":           2,
	"defaultgamemode": 2,
	"difficulty":      2,
	"effect":          2,
	"enchant":         2,
	"entitydata":      2,
	"execute":         2,
	"fill":            2,
	"function":        2,
	"gamemode":        2,
	"gamerule":        2,
	"give":            2,
	"kill":            2,
	"locate":          2,
	"particle":        2,
	"playsound":       2,
	"replaceitem":     2,
	"recipe":          2,
	"scoreboard":      2,
	"seed":            2,
	"setblock":        2,
	"setworldspawn":   2,
	"spawnpoint":      2,
	"spreadplayers":   2,
	"stats":           2,
	"stopsound":       2,
	"summon":          2,
	"tellraw":         2,
	"testfor":         2,
	"testforblock":    2,
	"testforblocks":   2,
	"time":            2,
	"title":           2,
	"toggledownfall":  2,
	"weather":         2,
	"worldborder":     2,
	"xp":              2,
	// Level 3: the administrative commands.
	"debug":          3,
	"reload":         3,
	"setidletimeout": 3,
}

// registerPermissions declares the permission level of every command.
//
// The table is registered from here because the permission model belongs
// with the function model: a function runs its commands at level 2, so
// the level a command needs decides whether it may run inside a
// function.
func registerPermissions(reg *exec.Registry) {
	for name, level := range commandPermissions {
		reg.SetRequiredPermission(name, level)
	}
}
