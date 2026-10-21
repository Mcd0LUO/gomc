package v1_12_2

import (
	"fmt"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// parseObjectives parses the `scoreboard objectives` subtree.
func parseObjectives(p *parser.Parser) ([]ast.Arg, error) {
	action, err := p.String()
	if err != nil {
		return nil, err
	}
	// 1.12.2 counts the arguments of the subcommand before it looks at
	// any of them, and every subcommand has its own usage line.
	n := p.Count() + 2
	switch action {
	case "add":
		if n < 4 {
			return nil, objectivesUsageError(p, "scoreboard objectives add")
		}
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		criteria, err := p.String()
		if err != nil {
			return nil, err
		}
		display := ""
		if p.More() {
			display = p.Rest()
		}
		return []ast.Arg{
			&ast.StringArg{Value: "objectives"},
			&ast.StringArg{Value: action},
			&ast.StringArg{Value: name},
			&ast.StringArg{Value: criteria},
			&ast.StringArg{Value: display},
		}, nil
	case "remove":
		if n != 3 {
			return nil, objectivesUsageError(p, "scoreboard objectives remove")
		}
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		return []ast.Arg{
			&ast.StringArg{Value: "objectives"},
			&ast.StringArg{Value: action},
			&ast.StringArg{Value: name},
		}, nil
	case "list":
		return []ast.Arg{
			&ast.StringArg{Value: "objectives"},
			&ast.StringArg{Value: action},
		}, nil
	case "setdisplay":
		if n != 3 && n != 4 {
			return nil, objectivesUsageError(p, "scoreboard objectives setdisplay")
		}
		slot, err := p.String()
		if err != nil {
			return nil, err
		}
		name := ""
		if p.More() {
			name, _ = p.String()
		}
		return []ast.Arg{
			&ast.StringArg{Value: "objectives"},
			&ast.StringArg{Value: action},
			&ast.StringArg{Value: slot},
			&ast.StringArg{Value: name},
		}, nil
	}
	return nil, objectivesUsageError(p, "scoreboard objectives")
}

// objectivesUsageError reports a usage error of a `scoreboard objectives`
// subcommand.
//
// 1.12.2 counts the arguments of the subcommand before it looks at any of
// them, and every subcommand has its own usage line in the language file.
func objectivesUsageError(p *parser.Parser, key string) error {
	return usageError(p, key, "Usage: %s", usageLine(key))
}

// objectiveDisplayName returns the name 1.12.2 shows for an objective.
//
// `ScoreboardObjective.getDisplayName()` falls back to the name of the
// objective when no display name was set.
func objectiveDisplayName(o *world.Objective) string {
	if o.DisplayName == "" {
		return o.Name
	}
	return o.DisplayName
}

// runObjectives executes the `scoreboard objectives` subtree.
func runObjectives(ctx *exec.Context, args []ast.Arg) (int, error) {
	sb := ctx.World.Scoreboard
	action := args[1].(*ast.StringArg).Value
	switch action {
	case "add":
		name := args[2].(*ast.StringArg).Value
		criteria := world.Criteria(args[3].(*ast.StringArg).Value)
		if !criteria.Valid() {
			return 0, exec.Errorf("unknown scoreboard criteria %q", criteria)
		}
		display := args[4].(*ast.StringArg).Value
		sb.AddObjective(name, criteria, display)
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Added new objective '%s' successfully", name))
		return 1, nil
	case "remove":
		name := args[2].(*ast.StringArg).Value
		if !sb.RemoveObjective(name) {
			return 0, exec.Errorf("No objective was found by the name '%s'", name)
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Removed objective '%s' successfully", name))
		return 1, nil
	case "list":
		objectives := sb.Objectives()
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Showing %d objective(s) on scoreboard:", len(objectives)))
		for _, o := range objectives {
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("- %s: displays as '%s' and is type '%s'",
					o.Name, objectiveDisplayName(o), o.Criteria))
		}
		return 1, nil
	case "setdisplay":
		slot := args[2].(*ast.StringArg).Value
		name := args[3].(*ast.StringArg).Value
		if name == "" {
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Cleared objective display slot '%s'", slot))
			return 1, nil
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Set the display objective in slot '%s' to '%s'", slot, name))
		return 1, nil
	}
	return 0, exec.Errorf("unknown action %q", action)
}
