package v1_12_2

import (
	"errors"
	"fmt"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/parser"
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
func (c *functionCmd) Usage() string { return "/function <id> [if|unless <target>]" }

// Parse implements exec.Handler.
func (c *functionCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	id, err := p.String()
	if err != nil {
		return nil, err
	}
	a := &functionArgs{id: id}
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
			// The condition was not met; the function is skipped.
			return 0, nil
		}
	}
	fn := ctx.Engine.Functions.Get(a.id)
	if fn == nil {
		return 0, exec.Errorf("Unknown function '%s'", a.id)
	}
	// The function inherits the caller's execution context.
	count, err := ctx.Engine.RunFunction(fn, ctx.Source)
	if err != nil {
		return 0, err
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Executed %d command(s) from function '%s'", count, a.id))
	return count, nil
}

// reloadCmd implements `/reload`.
type reloadCmd struct{}

// Name implements exec.Handler.
func (c *reloadCmd) Name() string { return "reload" }

// Usage implements exec.Handler.
func (c *reloadCmd) Usage() string { return "/reload" }

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
}
