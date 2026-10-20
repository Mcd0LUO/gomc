package v1_12_2

import (
	"fmt"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// parseTeams parses the `scoreboard teams` subtree.
//
// The layouts follow `CommandScoreboard`:
//
//	teams add <name> [display name ...]
//	teams remove <name>
//	teams empty <team>
//	teams join <team> [player]
//	teams leave [player]
//	teams list [name]
//	teams option <team> <option> [value]
//
// The argument list of a command is the text after the command name, so
// `teams leave` is two arguments long and `teams option <team> <option>`
// is three: 1.12.2 counts them the same way.
func parseTeams(p *parser.Parser) ([]ast.Arg, error) {
	action, err := p.String()
	if err != nil {
		return nil, err
	}
	args, err := parseTeamsAction(p, action)
	if err != nil {
		// The usage of the action is more helpful than the usage of
		// the whole tree.
		return nil, withUsage(err, "scoreboard teams "+action)
	}
	return args, nil
}

// parseTeamsAction parses the action of `scoreboard teams`.
func parseTeamsAction(p *parser.Parser, action string) ([]ast.Arg, error) {
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
	case "join":
		// The player is the executor when it is left out.
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		holders, err := parseScoreHolders(p)
		if err != nil {
			return nil, err
		}
		args := []ast.Arg{
			&ast.StringArg{Value: "teams"},
			&ast.StringArg{Value: action},
			&ast.StringArg{Value: name},
		}
		return append(args, holders...), nil
	case "leave":
		// 1.12.2 has no team argument here: the player leaves every
		// team it is on, so only the player can be named.
		holders, err := parseScoreHolders(p)
		if err != nil {
			return nil, err
		}
		args := []ast.Arg{
			&ast.StringArg{Value: "teams"},
			&ast.StringArg{Value: action},
		}
		return append(args, holders...), nil
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
		name := ""
		if p.More() {
			var err error
			name, err = p.String()
			if err != nil {
				return nil, err
			}
			if p.More() {
				return nil, usageError(p, "scoreboard teams list", "expected at most one team name")
			}
		}
		return []ast.Arg{
			&ast.StringArg{Value: "teams"},
			&ast.StringArg{Value: action},
			&ast.StringArg{Value: name},
		}, nil
	case "option":
		// The value may be left out; 1.12.2 then answers with the
		// values the option takes.
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		option, err := p.String()
		if err != nil {
			return nil, err
		}
		value := ""
		if p.More() {
			value, err = p.String()
			if err != nil {
				return nil, err
			}
			if p.More() {
				return nil, usageError(p, "scoreboard teams option", "expected at most one value")
			}
		}
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
		if _, created := sb.AddTeam(name); !created {
			// 1.12.2 refuses to create a team that exists already.
			return 0, exec.Errorf("A team with the name '%s' already exists", name)
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Added new team '%s' successfully", name))
		return 1, nil
	case "remove":
		name := args[2].(*ast.StringArg).Value
		if !sb.RemoveTeam(name) {
			return 0, exec.Errorf("No team was found by the name '%s'", name)
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Removed team %s", name))
		return 1, nil
	case "empty":
		name := args[2].(*ast.StringArg).Value
		t := sb.Team(name)
		if t == nil {
			return 0, exec.Errorf("No team was found by the name '%s'", name)
		}
		members := t.MemberList()
		if len(members) == 0 {
			return 0, exec.Errorf("Team %s is already empty, cannot remove nonexistant players", name)
		}
		for _, m := range members {
			sb.RemoveHolderFromTeam(m)
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Removed all %d player(s) from team %s", len(members), name))
		return len(members), nil
	case "join":
		name := args[2].(*ast.StringArg).Value
		holders, err := resolveTeamHolders(ctx, args[3:], "scoreboard teams join")
		if err != nil {
			return 0, err
		}
		// 1.12.2 adds a player that is on another team to the new
		// team and leaves the old one, so only a team that does not
		// exist can fail here.
		var added, failed []string
		for _, h := range holders {
			if sb.AddTeamMember(name, h) {
				added = append(added, h)
				continue
			}
			failed = append(failed, h)
		}
		if len(added) > 0 {
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Added %d player(s) to team %s: %s",
					len(added), name, joinVanillaList(added)))
		}
		if len(failed) > 0 {
			return 0, exec.Errorf("Could not add %d player(s) to team %s: %s",
				len(failed), name, joinVanillaList(failed))
		}
		return len(added), nil
	case "leave":
		// The player leaves every team it is on, so the message does
		// not name a team.
		holders, err := resolveTeamHolders(ctx, args[2:], "scoreboard teams leave")
		if err != nil {
			return 0, err
		}
		var removed, failed []string
		for _, h := range holders {
			if sb.RemoveHolderFromTeam(h) {
				removed = append(removed, h)
				continue
			}
			failed = append(failed, h)
		}
		if len(removed) > 0 {
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Removed %d player(s) from their teams: %s",
					len(removed), joinVanillaList(removed)))
		}
		if len(failed) > 0 {
			return 0, exec.Errorf("Could not remove %d player(s) from their teams: %s",
				len(failed), joinVanillaList(failed))
		}
		return len(removed), nil
	case "list":
		name := args[2].(*ast.StringArg).Value
		if name == "" {
			teams := sb.Teams()
			if len(teams) == 0 {
				return 0, exec.Errorf("There are no teams registered on the scoreboard")
			}
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Showing %d teams on the scoreboard:", len(teams)))
			for _, t := range teams {
				ctx.Engine.Feedback = append(ctx.Engine.Feedback,
					fmt.Sprintf("- %s: '%s' has %d players",
						t.Name, teamDisplayName(t), len(t.Members)))
			}
			return 1, nil
		}
		t := sb.Team(name)
		if t == nil {
			return 0, exec.Errorf("No team was found by the name '%s'", name)
		}
		members := t.MemberList()
		if len(members) == 0 {
			return 0, exec.Errorf("Team %s has no players", name)
		}
		// 1.12.2 prints the members as one line of names, not as one
		// line per member.
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Showing %d player(s) in team %s:", len(members), name))
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, joinVanillaList(members))
		return 1, nil
	case "option":
		name := args[2].(*ast.StringArg).Value
		t := sb.Team(name)
		if t == nil {
			return 0, exec.Errorf("No team was found by the name '%s'", name)
		}
		// 1.12.2 lowercases the option name before it uses it, both as
		// the key of the option and in the message.
		option := strings.ToLower(args[3].(*ast.StringArg).Value)
		value := args[4].(*ast.StringArg).Value
		if !validTeamOption(option) {
			// 1.12.2 answers an unknown option with the usage of the
			// command, not with the list of the valid values.
			return 0, usageErrorf("scoreboard teams option")
		}
		if value == "" || !validTeamOptionValue(option, value) {
			return 0, exec.Errorf("Valid values for option %s are: %s",
				option, teamOptionValues(option))
		}
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

// teamColorValues are the colours 1.12.2 accepts for the `color` option.
//
// The list is `EnumChatFormat.a(true, false)`: the sixteen colours and
// `reset`. The formatting codes (bold, italic, ...) are not in it,
// because `EnumChatFormat.isFormat()` rules them out.
var teamColorValues = []string{
	"black", "dark_blue", "dark_green", "dark_aqua", "dark_red",
	"dark_purple", "gold", "gray", "dark_gray", "blue", "green",
	"aqua", "red", "light_purple", "yellow", "white", "reset",
}

// teamNameTagVisibility are the values of the `nametagVisibility` and the
// `deathMessageVisibility` option.
//
// They are the names of `ScoreboardTeamBase.EnumNameTagVisibility`.
var teamNameTagVisibility = []string{"always", "hideForOtherTeams", "hideForOwnTeam"}

// teamCollisionRules are the values of the `collisionRule` option.
//
// They are the names of `ScoreboardTeamBase.EnumTeamPush`.
var teamCollisionRules = []string{"always", "never", "pushOtherTeams", "pushOwnTeam"}

// validTeamOption reports whether 1.12.2 knows the team option.
//
// The names are the ones `CommandScoreboard` compares, lower-cased.
func validTeamOption(option string) bool {
	switch option {
	case "color", "friendlyfire", "seefriendlyinvisibles",
		"nametagvisibility", "deathmessagevisibility", "collisionrule":
		return true
	}
	return false
}

// validTeamOptionValue reports whether the value is one the option takes.
func validTeamOptionValue(option, value string) bool {
	switch option {
	case "color":
		return validChatFormatColor(value)
	case "friendlyfire", "seefriendlyinvisibles":
		return strings.EqualFold(value, "true") || strings.EqualFold(value, "false")
	case "nametagvisibility", "deathmessagevisibility":
		return containsString(teamNameTagVisibility, value)
	case "collisionrule":
		return containsString(teamCollisionRules, value)
	}
	return false
}

// teamOptionValues returns the values the option takes, in the text 1.12.2
// answers with.
//
// The lists are joined the way `CommandAbstract.a(Collection)` joins them.
func teamOptionValues(option string) string {
	switch option {
	case "color":
		return joinVanillaList(teamColorValues)
	case "friendlyfire", "seefriendlyinvisibles":
		return joinVanillaList([]string{"true", "false"})
	case "nametagvisibility", "deathmessagevisibility":
		return joinVanillaList(teamNameTagVisibility)
	case "collisionrule":
		return joinVanillaList(teamCollisionRules)
	}
	return ""
}

// validChatFormatColor reports whether the text is a colour 1.12.2 accepts.
//
// `EnumChatFormat.b(String)` lower-cases the text and drops everything
// that is not a letter before it looks the colour up, so `DARK_BLUE` and
// `dark blue` both name the same colour.
func validChatFormatColor(value string) bool {
	key := chatFormatKey(value)
	for _, c := range teamColorValues {
		if chatFormatKey(c) == key {
			return true
		}
	}
	return false
}

// chatFormatKey is the key `EnumChatFormat` looks a colour up with.
func chatFormatKey(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// containsString reports whether the list holds the value.
func containsString(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
