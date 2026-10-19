package v1_12_2

import (
	"strings"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/parser"
)

// scoreboardCmd implements `/scoreboard`.
//
// The 1.12.2 command has three subtrees, handled by separate handlers so
// each stays small:
//
//	scoreboard objectives <add|remove|list|setdisplay>
//	scoreboard players   <set|add|remove|reset|operation|enable>
//	scoreboard teams     <add|remove|join|leave|list|option|empty>
type scoreboardCmd struct{}

// Name implements exec.Handler.
func (c *scoreboardCmd) Name() string { return "scoreboard" }

// Usage implements exec.Handler.
func (c *scoreboardCmd) Usage() string {
	return "/scoreboard (objectives|players|teams) ..."
}

// Parse implements exec.Handler.
func (c *scoreboardCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	sub, err := p.String()
	if err != nil {
		return nil, err
	}
	switch sub {
	case "objectives":
		return parseObjectives(p)
	case "players":
		return parsePlayers(p)
	case "teams":
		return parseTeams(p)
	}
	return nil, p.Errorf("expected objectives, players or teams, got %q", sub)
}

// Run implements exec.Handler.
func (c *scoreboardCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	sub := args[0].(*ast.StringArg).Value
	switch sub {
	case "objectives":
		return runObjectives(ctx, args)
	case "players":
		return runPlayers(ctx, args)
	case "teams":
		return runTeams(ctx, args)
	}
	return 0, exec.Errorf("unknown scoreboard subcommand %q", sub)
}

// resolveHolders turns a score holder argument into concrete holder ids.
func resolveHolders(ctx *exec.Context, arg ast.Arg) ([]string, error) {
	switch t := arg.(type) {
	case *ast.StringArg:
		return []string{t.Value}, nil
	case *ast.Selector:
		resolver := &exec.SelectorResolver{World: ctx.World}
		entities, err := resolver.Resolve(ctx.Source, t)
		if err != nil {
			return nil, exec.Errorf("%v", err)
		}
		if len(entities) == 0 {
			// The score holder is a target list, so the dispatcher
			// expands it; an empty list is an error there.
			return nil, exec.Errorf("Selector '%s' found nothing", selectorText(t))
		}
		out := make([]string, 0, len(entities))
		for _, e := range entities {
			out = append(out, e.ID)
		}
		return out, nil
	}
	return nil, exec.Errorf("invalid score holder")
}

// registerScoreboard adds the scoreboard commands.
func registerScoreboard(reg *exec.Registry) {
	reg.Register(&scoreboardCmd{})
	reg.Register(&triggerCmd{})
}

// scoreHolderName renders a holder for messages.
func scoreHolderName(arg ast.Arg) string {
	switch t := arg.(type) {
	case *ast.StringArg:
		return t.Value
	case *ast.Selector:
		return "@" + t.Base
	}
	return strings.TrimSpace("?")
}
