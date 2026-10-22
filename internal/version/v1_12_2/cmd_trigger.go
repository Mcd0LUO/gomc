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
	var source *world.Entity
	if ctx.Source.Entity != nil {
		if !ctx.Source.Entity.IsPlayer {
			return 0, exec.Errorf("Only players can use the /trigger command")
		}
		source = ctx.Source.Entity
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
	// `CommandTrigger` reads the value before it looks at the score, so a
	// value that is not a number is reported even when the trigger is
	// disabled.
	value, err := parseIntSafe(rawValue)
	if err != nil {
		return 0, exec.Errorf("%v", err)
	}
	// The score has to exist. 1.12.2 reports that with the same message it
	// uses for an objective that is not a trigger
	// (`commands.trigger.invalidObjective`), not with a message of its
	// own.
	if _, tracked := sb.Score(holder, objective); !tracked {
		return 0, exec.Errorf("Invalid trigger name %s", objective)
	}
	// A trigger locks when it fires: the score stays until
	// `/scoreboard players enable` unlocks it again.
	if sb.ScoreLocked(holder, objective) {
		return 0, exec.Errorf("Trigger %s is not enabled", objective)
	}
	// The mode is checked **after** the lock, so a disabled trigger
	// reports that it is disabled rather than a bad mode.
	if mode != "add" && mode != "set" {
		return 0, exec.Errorf("Invalid trigger mode %s", mode)
	}
	current, _ := sb.Score(holder, objective)
	next := value
	if mode == "add" {
		next = current + value
	}
	sb.SetScore(holder, objective, next)
	sb.SetScoreLocked(holder, objective, true)
	// 1.12.2 only tells the player the trigger changed when it is in
	// creative mode (`playerInteractManager.isCreative`). The console is
	// not a player at all -- gomc lets it act as the holder, so it always
	// gets the line.
	if !player || playerGameMode(source) == gameModeCreative {
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Trigger %s changed with %s %s", objective, mode, rawValue))
	}
	return 1, nil
}
