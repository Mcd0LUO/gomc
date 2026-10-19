package v1_12_2

import (
	"fmt"
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
		args, err := parseObjectives(p)
		return args, withUsage(err, "scoreboard objectives")
	case "players":
		args, err := parsePlayers(p)
		return args, withUsage(err, "scoreboard players")
	case "teams":
		args, err := parseTeams(p)
		return args, withUsage(err, "scoreboard teams")
	}
	return nil, usageError(p, "scoreboard", "expected objectives, players or teams, got %q", sub)
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
//
// 1.12.2 tracks a player under the player's name and every other entity
// under its UUID, so a selector resolves to the scoreboard names of the
// entities it matched rather than to their internal ids. That is what
// makes `/scoreboard players set Steve kills 5` and the `score_kills`
// and `team=` selector arguments agree on who "Steve" is.
//
// The `scoreboard players tag` action is the exception: it adds a
// scoreboard tag to the entity itself instead of touching a scoreboard
// entry, so the entity id is returned for it.
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
		byEntity := isTagAction(ctx)
		out := make([]string, 0, len(entities))
		for _, e := range entities {
			if byEntity {
				out = append(out, e.ID)
				continue
			}
			out = append(out, exec.ScoreboardHolder(e))
		}
		return out, nil
	}
	return nil, exec.Errorf("invalid score holder")
}

// parseOptionalScoreHolder parses a score holder that may be left out.
//
// The absence is kept in the argument list as an empty name, because
// 1.12.2 reads the argument count to tell the two cases apart.
func parseOptionalScoreHolder(p *parser.Parser) (ast.Arg, error) {
	if !p.More() {
		return &ast.StringArg{Value: ""}, nil
	}
	return parseScoreHolder(p)
}

// parseScoreHolders parses the players of `scoreboard teams join` and
// `scoreboard teams leave`.
//
// 1.12.2 takes one player or more, and the executor itself when none is
// named.
func parseScoreHolders(p *parser.Parser) ([]ast.Arg, error) {
	var out []ast.Arg
	for p.More() {
		holder, err := parseScoreHolder(p)
		if err != nil {
			return nil, err
		}
		out = append(out, holder)
	}
	return out, nil
}

// resolveTeamHolders resolves the players of `scoreboard teams join` and
// `scoreboard teams leave`.
//
// 1.12.2 leaves the player out of the command when the executor is the
// player itself, but only a player may be the executor in that case: a
// console has to name somebody. The check cannot happen while parsing,
// because the executor is only known when the command runs.
func resolveTeamHolders(ctx *exec.Context, args []ast.Arg, usageKey string) ([]string, error) {
	if len(args) == 0 {
		if ctx.Source.Entity == nil || !ctx.Source.Entity.IsPlayer {
			return nil, usageErrorf(usageKey)
		}
		return []string{exec.ScoreboardHolder(ctx.Source.Entity)}, nil
	}
	var out []string
	for _, arg := range args {
		holders, err := resolveHolders(ctx, arg)
		if err != nil {
			return nil, err
		}
		out = append(out, holders...)
	}
	return out, nil
}

// usageErrorf builds the error 1.12.2 answers a usage mistake with.
//
// The server sends `commands.generic.usage` with the usage line of the
// command, so the message is "Usage: <the line>". It is used where the
// mistake can only be seen when the command runs, because it depends on
// who is running it.
func usageErrorf(key string) error {
	usage := commandUsage(key)
	if usage == "" {
		return exec.Errorf("Usage: %s", key)
	}
	return &exec.Error{Msg: fmt.Sprintf(commandUsage("generic"), usage)}
}

// isTagAction reports whether the scoreboard command being run is
// `scoreboard players tag`.
//
// That action works on the entity rather than on a scoreboard entry, so
// the holder it resolves has to be the entity id.
func isTagAction(ctx *exec.Context) bool {
	if len(ctx.Args) >= 2 {
		if sub, ok := ctx.Args[0].(*ast.StringArg); ok && sub.Value == "players" {
			if action, ok := ctx.Args[1].(*ast.StringArg); ok && action.Value == "tag" {
				return true
			}
		}
	}
	// The raw command is the fallback, so the check keeps working when
	// the parsed argument layout changes.
	if ctx.Command == nil {
		return false
	}
	fields := strings.Fields(ctx.Command.Raw)
	return len(fields) >= 3 && fields[0] == "scoreboard" &&
		fields[1] == "players" && fields[2] == "tag"
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
