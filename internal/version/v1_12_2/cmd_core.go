package v1_12_2

import (
	"fmt"
	"strings"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/parser"
)

// sayCmd implements `/say`.
type sayCmd struct{}

// Name implements exec.Handler.
func (c *sayCmd) Name() string { return "say" }

// Usage implements exec.Handler.
func (c *sayCmd) Usage() string { return "/say <message>" }

// Parse implements exec.Handler.
func (c *sayCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	msg := p.Rest()
	if strings.TrimSpace(msg) == "" {
		return nil, p.Errorf("expected a message")
	}
	return []ast.Arg{&ast.StringArg{Value: msg}}, nil
}

// Run implements exec.Handler.
func (c *sayCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	// The prefix is the name of whoever ran the command. When the command
	// runs inside an /execute, that is the target entity, not the server.
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		"["+ctx.Source.Name+"] "+args[0].(*ast.StringArg).Value)
	return 1, nil
}

// tellCmd implements `/tell` and its `/w` alias.
type tellCmd struct {
	// alias is the alternative command name, empty for the primary one.
	alias string
}

// Name implements exec.Handler.
func (c *tellCmd) Name() string {
	if c.alias != "" {
		return c.alias
	}
	return "tell"
}

// Usage implements exec.Handler.
func (c *tellCmd) Usage() string { return "/tell <player> <message>" }

// Parse implements exec.Handler.
func (c *tellCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target, err := p.String()
	if err != nil {
		return nil, err
	}
	msg := p.Rest()
	if strings.TrimSpace(msg) == "" {
		return nil, p.Errorf("expected a message")
	}
	return []ast.Arg{&ast.StringArg{Value: target}, &ast.StringArg{Value: msg}}, nil
}

// Run implements exec.Handler.
func (c *tellCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	target := args[0].(*ast.StringArg).Value
	msg := args[1].(*ast.StringArg).Value
	player := ctx.World.FindPlayer(target)
	if player == nil {
		return 0, exec.Errorf("no player named %q is online", target)
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback, fmt.Sprintf("[%s -> %s] %s",
		ctx.Source.Name, player.PlayerName, msg))
	return 1, nil
}

// tellRawCmd implements `/tellraw`.
type tellRawCmd struct{}

// Name implements exec.Handler.
func (c *tellRawCmd) Name() string { return "tellraw" }

// Usage implements exec.Handler.
func (c *tellRawCmd) Usage() string { return "/tellraw <target> <raw json message>" }

// Parse implements exec.Handler.
func (c *tellRawCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	msg, err := p.JSON()
	if err != nil {
		return nil, err
	}
	return []ast.Arg{target, msg}, nil
}

// Run implements exec.Handler.
func (c *tellRawCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	entities, err := resolveEntityTargets(ctx, args[0])
	if err != nil {
		return 0, err
	}
	raw := args[1].(*ast.JSONArg).Value
	for _, e := range entities {
		name := e.PlayerName
		if name == "" {
			name = e.ID
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, fmt.Sprintf("[%s -> %s] %s",
			ctx.Source.Name, name, raw))
	}
	return len(entities), nil
}

// gameruleCmd implements `/gamerule`.
type gameruleCmd struct{}

// Name implements exec.Handler.
func (c *gameruleCmd) Name() string { return "gamerule" }

// Usage implements exec.Handler.
func (c *gameruleCmd) Usage() string { return "/gamerule <rule> [<value>]" }

// Parse implements exec.Handler.
func (c *gameruleCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	rule, err := p.String()
	if err != nil {
		return nil, err
	}
	if !p.More() {
		return []ast.Arg{&ast.StringArg{Value: rule}}, nil
	}
	value := p.Rest()
	return []ast.Arg{&ast.StringArg{Value: rule}, &ast.StringArg{Value: value}}, nil
}

// Run implements exec.Handler.
func (c *gameruleCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	rules := ctx.World.Rules
	rule := args[0].(*ast.StringArg).Value
	if len(args) == 1 {
		r := rules.Get(rule)
		if r == nil {
			return 0, exec.Errorf("unknown game rule %q", rule)
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, fmt.Sprintf("%s = %s", rule, r.Display()))
		return 1, nil
	}
	value := args[1].(*ast.StringArg).Value
	if err := rules.Set(rule, value); err != nil {
		return 0, exec.Errorf("%v", err)
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback, fmt.Sprintf("Game rule %s has been updated to %s", rule, value))
	return 1, nil
}
