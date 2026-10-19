package v1_12_2

import (
	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/parser"
)

// parseTeams parses the `scoreboard teams` subtree.
func parseTeams(p *parser.Parser) ([]ast.Arg, error) {
	action, err := p.String()
	if err != nil {
		return nil, err
	}
	switch action {
	case "add", "remove":
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		return []ast.Arg{
			&ast.StringArg{Value: "teams"},
			&ast.StringArg{Value: action},
			&ast.StringArg{Value: name},
		}, nil
	case "join", "leave":
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		holder, err := parseScoreHolder(p)
		if err != nil {
			return nil, err
		}
		return []ast.Arg{
			&ast.StringArg{Value: "teams"},
			&ast.StringArg{Value: action},
			&ast.StringArg{Value: name},
			holder,
		}, nil
	case "empty":
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		return []ast.Arg{
			&ast.StringArg{Value: "teams"},
			&ast.StringArg{Value: action},
			&ast.StringArg{Value: name},
		}, nil
	case "list":
		return []ast.Arg{
			&ast.StringArg{Value: "teams"},
			&ast.StringArg{Value: action},
		}, nil
	case "option":
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		option, err := p.String()
		if err != nil {
			return nil, err
		}
		value := p.Rest()
		return []ast.Arg{
			&ast.StringArg{Value: "teams"},
			&ast.StringArg{Value: action},
			&ast.StringArg{Value: name},
			&ast.StringArg{Value: option},
			&ast.StringArg{Value: value},
		}, nil
	}
	return nil, usageError(p, "scoreboard teams", "expected add, remove, join, leave, empty, list or option, got %q", action)
}

// runTeams executes the `scoreboard teams` subtree.
func runTeams(ctx *exec.Context, args []ast.Arg) (int, error) {
	sb := ctx.World.Scoreboard
	action := args[1].(*ast.StringArg).Value
	switch action {
	case "add":
		name := args[2].(*ast.StringArg).Value
		sb.AddTeam(name)
		return 1, nil
	case "remove":
		name := args[2].(*ast.StringArg).Value
		if !sb.RemoveTeam(name) {
			return 0, exec.Errorf("unknown team %q", name)
		}
		return 1, nil
	case "join":
		name := args[2].(*ast.StringArg).Value
		t := sb.Team(name)
		if t == nil {
			return 0, exec.Errorf("unknown team %q", name)
		}
		holders, err := resolveHolders(ctx, args[3])
		if err != nil {
			return 0, err
		}
		for _, h := range holders {
			sb.AddTeamMember(name, h)
		}
		return len(holders), nil
	case "leave":
		name := args[2].(*ast.StringArg).Value
		t := sb.Team(name)
		if t == nil {
			return 0, exec.Errorf("unknown team %q", name)
		}
		holders, err := resolveHolders(ctx, args[3])
		if err != nil {
			return 0, err
		}
		for _, h := range holders {
			sb.RemoveTeamMember(name, h)
		}
		return len(holders), nil
	case "empty":
		name := args[2].(*ast.StringArg).Value
		t := sb.Team(name)
		if t == nil {
			return 0, exec.Errorf("unknown team %q", name)
		}
		t.Members = map[string]bool{}
		return 1, nil
	case "list":
		for _, t := range sb.Teams() {
			ctx.Engine.Feedback = append(ctx.Engine.Feedback, t.Name)
		}
		return 1, nil
	case "option":
		name := args[2].(*ast.StringArg).Value
		t := sb.Team(name)
		if t == nil {
			return 0, exec.Errorf("unknown team %q", name)
		}
		option := args[3].(*ast.StringArg).Value
		value := args[4].(*ast.StringArg).Value
		t.Options[option] = value
		return 1, nil
	}
	return 0, exec.Errorf("unknown action %q", action)
}
