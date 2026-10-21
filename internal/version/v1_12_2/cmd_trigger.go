package v1_12_2

import (
	"fmt"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// triggerCmd implements `/trigger`.
//
// The 1.12.2 syntax is `/trigger <objective> <add|set> <value>`: the
// mode is required and the value is the third argument.
type triggerCmd struct{}

// Name implements exec.Handler.
func (c *triggerCmd) Name() string { return "trigger" }

// Usage implements exec.Handler.
func (c *triggerCmd) Usage() string { return usageLine("trigger") }

// Parse implements exec.Handler.
//
// All three arguments are required: `CommandTrigger` refuses anything
// shorter than three.
func (c *triggerCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	objective, err := p.String()
	if err != nil {
		return nil, err
	}
	mode, err := p.String()
	if err != nil {
		return nil, err
	}
	value, err := p.String()
	if err != nil {
		return nil, err
	}
	return []ast.Arg{
		&ast.StringArg{Value: objective},
		&ast.StringArg{Value: mode},
		&ast.StringArg{Value: value},
	}, nil
}

// Run implements exec.Handler.
func (c *triggerCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	objective := args[0].(*ast.StringArg).Value
	mode := args[1].(*ast.StringArg).Value
	rawValue := args[2].(*ast.StringArg).Value

	// 1.12.2 only lets a player run `/trigger`: the command looks at the
	// listener and refuses anything that is not a player. gomc has no
	// "online player" concept and every function runs as the console, so
	// the console is allowed and acts as the holder; an entity that is
	// not a player is refused the way the real server refuses it.
	holder := ctx.Source.Name
	player := false
	if ctx.Source.Entity != nil {
		if !ctx.Source.Entity.IsPlayer {
			return 0, exec.Errorf("Only players can use the /trigger command")
		}
		holder = ctx.Source.Entity.PlayerName
		if holder == "" {
			holder = ctx.Source.Entity.DisplayName()
		}
		player = true
	}

	sb := ctx.World.Scoreboard
	o := sb.Objective(objective)
	if o == nil || o.Criteria != world.CriteriaTrigger {
		return 0, exec.Errorf("Invalid trigger name %s", objective)
	}
	value, err := parseIntSafe(rawValue)
	if err != nil {
		return 0, exec.Errorf("%v", err)
	}
	// A trigger has to be enabled for the holder before it fires.
	// 1.12.2 keeps that state in the scoreboard, where a trigger score
	// starts locked and `/scoreboard players enable` unlocks it. gomc's
	// scoreboard has no lock flag and its `players enable` only reports,
	// so the tracked score is the only state that is left to look at.
	if player {
		if _, tracked := sb.Score(holder, objective); !tracked {
			return 0, exec.Errorf("Trigger %s is not enabled", objective)
		}
	}
	if mode != "add" && mode != "set" {
		return 0, exec.Errorf("Invalid trigger mode %s", mode)
	}
	current, tracked := sb.Score(holder, objective)
	if !tracked {
		current = 0
	}
	next := value
	if mode == "add" {
		next = current + value
	}
	sb.SetScore(holder, objective, next)
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Trigger %s changed with %s %s", objective, mode, rawValue))
	return 1, nil
}
