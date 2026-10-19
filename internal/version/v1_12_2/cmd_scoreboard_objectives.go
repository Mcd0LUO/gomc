package v1_12_2

import (
	"fmt"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/world"
)

// parseObjectives parses the `scoreboard objectives` subtree.
func parseObjectives(p *parser.Parser) ([]ast.Arg, error) {
	action, err := p.String()
	if err != nil {
		return nil, err
	}
	switch action {
	case "add":
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
	return nil, p.Errorf("expected add, remove, list or setdisplay, got %q", action)
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
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, fmt.Sprintf("Added objective %s", name))
		return 1, nil
	case "remove":
		name := args[2].(*ast.StringArg).Value
		if !sb.RemoveObjective(name) {
			return 0, exec.Errorf("unknown objective %q", name)
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, fmt.Sprintf("Removed objective %s", name))
		return 1, nil
	case "list":
		for _, o := range sb.Objectives() {
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("%s (%s)", o.Name, o.Criteria))
		}
		return 1, nil
	case "setdisplay":
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, "Objective display slot updated")
		return 1, nil
	}
	return 0, exec.Errorf("unknown action %q", action)
}
