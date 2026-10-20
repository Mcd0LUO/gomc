package v1_12_2

import (
	"fmt"
	"sort"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
)

// listCmd implements `/list`.
//
// The command reports the online players and how many of them there are.
type listCmd struct{}

// Name implements exec.Handler.
func (c *listCmd) Name() string { return "list" }

// Usage implements exec.Handler.
func (c *listCmd) Usage() string { return "/list" }

// Parse implements exec.Handler.
func (c *listCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	return nil, nil
}

// Run implements exec.Handler.
func (c *listCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	players := ctx.World.Players()
	names := make([]string, 0, len(players))
	for _, p := range players {
		names = append(names, p.PlayerName)
	}
	sort.Strings(names)
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("There are %d/%d players online:", len(names), len(names)))
	for _, n := range names {
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, n)
	}
	return 1, nil
}

// helpCmd implements `/help`.
//
// Without an argument the command lists the registered commands, and with
// one it reports the usage of that command.
type helpCmd struct{}

// Name implements exec.Handler.
func (c *helpCmd) Name() string { return "help" }

// Usage implements exec.Handler.
func (c *helpCmd) Usage() string { return "/help [<command>]" }

// Parse implements exec.Handler.
func (c *helpCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	if !p.More() {
		return []ast.Arg{&ast.StringArg{Value: ""}}, nil
	}
	name, err := p.String()
	if err != nil {
		return nil, err
	}
	return []ast.Arg{&ast.StringArg{Value: name}}, nil
}

// Run implements exec.Handler.
func (c *helpCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	name := args[0].(*ast.StringArg).Value
	if name == "" {
		names := ctx.Engine.Registry.Names()
		sort.Strings(names)
		for _, n := range names {
			ctx.Engine.Feedback = append(ctx.Engine.Feedback, n)
		}
		return 1, nil
	}
	handler, ok := ctx.Engine.Registry.Get(name)
	if !ok {
		return 0, exec.Errorf("unknown command %q", name)
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback, handler.Usage())
	return 1, nil
}

// debugCmd implements `/debug`.
//
// The command starts and stops the server's debug report, which the runtime
// does not model, so it only reports the state.
type debugCmd struct{}

// Name implements exec.Handler.
func (c *debugCmd) Name() string { return "debug" }

// Usage implements exec.Handler.
func (c *debugCmd) Usage() string { return "/debug <start|stop>" }

// Parse implements exec.Handler.
func (c *debugCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	action, err := p.String()
	if err != nil {
		return nil, err
	}
	switch action {
	case "start", "stop":
	default:
		return nil, p.Errorf("expected start or stop, got %q", action)
	}
	return []ast.Arg{&ast.StringArg{Value: action}}, nil
}

// Run implements exec.Handler.
func (c *debugCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	action := args[0].(*ast.StringArg).Value
	if action == "start" {
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, "Started debug profiling")
	} else {
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, "Stopped debug profiling after 0 seconds (0 ticks)")
	}
	return 1, nil
}

// registerMisc adds the informational commands.
func registerMisc(reg *exec.Registry) {
	reg.Register(&listCmd{})
	reg.Register(&helpCmd{})
	reg.Register(&debugCmd{})
}
