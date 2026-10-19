package v1_12_2

import (
	"strconv"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/world"
)

// triggerCmd implements `/trigger`.
type triggerCmd struct{}

// Name implements exec.Handler.
func (c *triggerCmd) Name() string { return "trigger" }

// Usage implements exec.Handler.
func (c *triggerCmd) Usage() string { return "/trigger <objective> [<value>]" }

// Parse implements exec.Handler.
func (c *triggerCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	objective, err := p.String()
	if err != nil {
		return nil, err
	}
	value := 0
	if p.More() {
		raw, err := p.String()
		if err != nil {
			return nil, err
		}
		if n, err := strconv.Atoi(raw); err == nil {
			value = n
		}
	}
	return []ast.Arg{
		&ast.StringArg{Value: objective},
		&ast.IntArg{Value: value},
	}, nil
}

// Run implements exec.Handler.
func (c *triggerCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	sb := ctx.World.Scoreboard
	objective := args[0].(*ast.StringArg).Value
	o := sb.Objective(objective)
	if o == nil {
		return 0, exec.Errorf("No objective was found by the name '%s'", objective)
	}
	if o.Criteria != world.CriteriaTrigger {
		return 0, exec.Errorf("objective %q is not a trigger", objective)
	}
	holder := ctx.Source.Name
	if ctx.Source.Entity != nil {
		holder = ctx.Source.Entity.ID
	}
	value := args[1].(*ast.IntArg).Value
	current, tracked := sb.Score(holder, objective)
	if !tracked {
		current = 0
	}
	sb.SetScore(holder, objective, current+value)
	return 1, nil
}
