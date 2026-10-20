package v1_12_2

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// sayCmd implements `/say`.
type sayCmd struct{}

// Name implements exec.Handler.
func (c *sayCmd) Name() string { return "say" }

// Usage implements exec.Handler.
func (c *sayCmd) Usage() string { return "/say <message>" }

// Parse implements exec.Handler.
func (c *sayCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	msg := p.Rest()
	if strings.TrimSpace(msg) == "" {
		return nil, p.Errorf("expected a message")
	}
	return []ast.Arg{&ast.StringArg{Value: msg}}, nil
}

// Run implements exec.Handler.
func (c *sayCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	// 1.12.2 expands the selectors in the message, and a selector that
	// matches nothing is an error.
	msg, err := expandMessage(ctx, args[0].(*ast.StringArg).Value)
	if err != nil {
		return 0, err
	}
	// The prefix is the name of whoever ran the command. When the command
	// runs inside an /execute, that is the target entity, not the server.
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		"["+ctx.Source.Name+"] "+msg)
	return 1, nil
}

// tellCmd implements `/tell` and its `/w` alias.
type tellCmd struct {
	// alias is the alternative command name, empty for the primary one.
	alias string
}

// Name implements exec.Handler.
func (c *tellCmd) Name() string {
	if c.alias != "" {
		return c.alias
	}
	return "tell"
}

// Usage implements exec.Handler.
func (c *tellCmd) Usage() string { return "/tell <player> <message>" }

// Parse implements exec.Handler.
func (c *tellCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target, err := p.String()
	if err != nil {
		return nil, err
	}
	msg := p.Rest()
	if strings.TrimSpace(msg) == "" {
		return nil, p.Errorf("expected a message")
	}
	return []ast.Arg{&ast.StringArg{Value: target}, &ast.StringArg{Value: msg}}, nil
}

// Run implements exec.Handler.
func (c *tellCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	target := args[0].(*ast.StringArg).Value
	msg := args[1].(*ast.StringArg).Value
	player := ctx.World.FindPlayer(target)
	if player == nil {
		return 0, exec.Errorf("no player named %q is online", target)
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback, fmt.Sprintf("You whisper to %s: %s",
		player.PlayerName, msg))
	return 1, nil
}

// tellRawCmd implements `/tellraw`.
type tellRawCmd struct{}

// Name implements exec.Handler.
func (c *tellRawCmd) Name() string { return "tellraw" }

// Usage implements exec.Handler.
func (c *tellRawCmd) Usage() string { return "/tellraw <target> <raw json message>" }

// Parse implements exec.Handler.
func (c *tellRawCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	msg, err := p.JSON()
	if err != nil {
		return nil, err
	}
	return []ast.Arg{target, msg}, nil
}

// Run implements exec.Handler.
func (c *tellRawCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	// 1.12.2 sends the message to a *player*: the target is resolved with
	// `CommandAbstract.b`, which keeps only the players it finds.
	entities, err := resolvePlayerTargets(ctx, args[0])
	if err != nil {
		return 0, err
	}
	raw := args[1].(*ast.JSONArg).Value
	for _, e := range entities {
		name := e.PlayerName
		if name == "" {
			name = e.ID
		}
		// 1.12.2 sends the raw message to the target; the console shows
		// the sender's name in front of it.
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, fmt.Sprintf("[%s] %s",
			ctx.Source.Name, raw))
	}
	return len(entities), nil
}

// gameruleCmd implements `/gamerule`.
type gameruleCmd struct{}

// Name implements exec.Handler.
func (c *gameruleCmd) Name() string { return "gamerule" }

// Usage implements exec.Handler.
func (c *gameruleCmd) Usage() string { return "/gamerule <rule> [<value>]" }

// Parse implements exec.Handler.
func (c *gameruleCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	// 1.12.2 lists every rule when the command has no argument.
	if !p.More() {
		return nil, nil
	}
	rule, err := p.String()
	if err != nil {
		return nil, err
	}
	if !p.More() {
		return []ast.Arg{&ast.StringArg{Value: rule}}, nil
	}
	value := p.Rest()
	return []ast.Arg{&ast.StringArg{Value: rule}, &ast.StringArg{Value: value}}, nil
}

// Run implements exec.Handler.
func (c *gameruleCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	rules := ctx.World.Rules
	if len(args) == 0 {
		// 1.12.2 keeps the rules in a sorted map, so the list is in
		// the alphabetical order of the names, and it is joined into a
		// single line.
		names := rules.Names()
		sort.Strings(names)
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, joinVanillaList(names))
		return 1, nil
	}
	rule := args[0].(*ast.StringArg).Value
	r := rules.Get(rule)
	if r == nil && len(args) == 1 {
		// 1.12.2 only reports an unknown rule when it is *queried*.
		// Setting a rule that does not exist creates it.
		return 0, exec.Errorf("No game rule called '%s' is available", rule)
	}
	if len(args) == 1 {
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, fmt.Sprintf("%s = %s", rule, r.Display()))
		return 1, nil
	}
	value := args[1].(*ast.StringArg).Value
	// 1.12.2 only accepts "true" and "false" for a boolean rule, and it
	// compares them case sensitively.
	if r != nil && r.Type == world.RuleBoolean && value != "true" && value != "false" {
		return 0, exec.Errorf("'%s' is not true or false", value)
	}
	if err := rules.Set(rule, value); err != nil {
		return 0, exec.Errorf("%v", err)
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback, fmt.Sprintf("Game rule %s has been updated to %s", rule, value))
	return 1, nil
}

// joinVanillaList joins names the way 1.12.2 does.
//
// `CommandAbstract.a(Object[])` separates the entries with ", " and puts
// " and " in front of the last one.
func joinVanillaList(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	}
	return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
}
