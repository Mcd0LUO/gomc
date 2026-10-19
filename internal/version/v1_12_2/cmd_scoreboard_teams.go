package v1_12_2

import (
	"fmt"
	"strings"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/world"
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
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Added new team '%s' successfully", name))
		return 1, nil
	case "remove":
		name := args[2].(*ast.StringArg).Value
		if !sb.RemoveTeam(name) {
			return 0, exec.Errorf("unknown team %q", name)
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Removed team %s", name))
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
		// The message names the team and the players that joined it.
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Added %d player(s) to team %s: %s",
				len(holders), name, joinVanillaList(holders)))
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
		// The players leave whatever team they are on, so the message
		// does not name a team.
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Removed %d player(s) from their teams: %s",
				len(holders), joinVanillaList(holders)))
		return len(holders), nil
	case "empty":
		name := args[2].(*ast.StringArg).Value
		t := sb.Team(name)
		if t == nil {
			return 0, exec.Errorf("unknown team %q", name)
		}
		count := len(t.Members)
		t.Members = map[string]bool{}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Removed all %d player(s) from team %s", count, name))
		return 1, nil
	case "list":
		teams := sb.Teams()
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Showing %d teams on the scoreboard:", len(teams)))
		for _, t := range teams {
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("- %s: '%s' has %d players",
					t.Name, teamDisplayName(t), len(t.Members)))
		}
		return 1, nil
	case "option":
		name := args[2].(*ast.StringArg).Value
		t := sb.Team(name)
		if t == nil {
			return 0, exec.Errorf("unknown team %q", name)
		}
		// 1.12.2 lowercases the option name before it uses it, both as
		// the key of the option and in the message.
		option := strings.ToLower(args[3].(*ast.StringArg).Value)
		value := args[4].(*ast.StringArg).Value
		t.Options[option] = value
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Set option %s for team %s to %s", option, name, value))
		return 1, nil
	}
	return 0, exec.Errorf("unknown action %q", action)
}

// teamDisplayName returns the name 1.12.2 shows for a team.
//
// `ScoreboardTeam.getDisplayName()` falls back to the name of the team
// when no display name was set.
func teamDisplayName(t *world.Team) string {
	if t.DisplayName == "" {
		return t.Name
	}
	return t.DisplayName
}
