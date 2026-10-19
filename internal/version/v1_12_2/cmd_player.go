package v1_12_2

import (
	"fmt"
	"strings"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/nbt"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/world"
)

// gamemodeCmd implements `/gamemode`.
type gamemodeCmd struct{}

// Name implements exec.Handler.
func (c *gamemodeCmd) Name() string { return "gamemode" }

// Usage implements exec.Handler.
func (c *gamemodeCmd) Usage() string { return "/gamemode <mode> [<target>]" }

// Parse implements exec.Handler.
func (c *gamemodeCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	mode, err := p.String()
	if err != nil {
		return nil, err
	}
	if !p.More() {
		return []ast.Arg{&ast.StringArg{Value: mode}}, nil
	}
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	return []ast.Arg{&ast.StringArg{Value: mode}, target}, nil
}

// Run implements exec.Handler.
func (c *gamemodeCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	mode := args[0].(*ast.StringArg).Value
	id, ok := gameModeID(mode)
	if !ok {
		return 0, exec.Errorf("unknown game mode %q", mode)
	}
	var targets []*world.Entity
	if len(args) == 1 {
		if ctx.Source.Entity == nil {
			return 0, exec.Errorf("no player to change")
		}
		targets = []*world.Entity{ctx.Source.Entity}
	} else {
		entities, err := resolveEntityTargets(ctx, args[1])
		if err != nil {
			return 0, err
		}
		targets = entities
	}
	for _, e := range targets {
		e.NBT.Set("playerGameType", nbt.Int(int32(id)))
	}
	return len(targets), nil
}

// gameModeID maps a game mode name to its 1.12.2 id.
func gameModeID(name string) (int, bool) {
	switch strings.ToLower(name) {
	case "0", "survival":
		return 0, true
	case "1", "creative":
		return 1, true
	case "2", "adventure":
		return 2, true
	case "3", "spectator":
		return 3, true
	}
	return 0, false
}

// giveCmd implements `/give`.
type giveCmd struct{}

// Name implements exec.Handler.
func (c *giveCmd) Name() string { return "give" }

// Usage implements exec.Handler.
func (c *giveCmd) Usage() string { return "/give <target> <item> [<count>] [<data>] [<nbt>]" }

// Parse implements exec.Handler.
func (c *giveCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	item, err := p.String()
	if err != nil {
		return nil, err
	}
	ref := &ast.ItemArg{Name: item, Count: -1, Data: -1}
	if p.More() {
		if t, ok := p.Peek(); ok && isDataValue(t.Text) {
			p.Next()
			n, _ := parseIntSafe(t.Text)
			ref.Count = n
		}
	}
	if p.More() {
		if t, ok := p.Peek(); ok && isDataValue(t.Text) {
			p.Next()
			n, _ := parseIntSafe(t.Text)
			ref.Data = n
		}
	}
	if p.More() {
		data, err := p.NBT()
		if err != nil {
			return nil, err
		}
		ref.NBT = data.Value
	}
	return []ast.Arg{target, ref}, nil
}

// Run implements exec.Handler.
func (c *giveCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	entities, err := resolveEntityTargets(ctx, args[0])
	if err != nil {
		return 0, err
	}
	item := args[1].(*ast.ItemArg)
	count := item.Count
	if count < 0 {
		count = 1
	}
	for _, e := range entities {
		stack := nbt.NewCompound()
		stack.Set("id", nbt.String(item.Name))
		stack.Set("Count", nbt.Int(int32(count)))
		if item.Data >= 0 {
			stack.Set("Damage", nbt.Short(int16(item.Data)))
		}
		if item.NBT != nil {
			stack.Set("tag", item.NBT.Clone())
		}
		inventory := e.NBT.Get("Inventory")
		list, ok := inventory.(*nbt.List)
		if !ok {
			list = nbt.NewList(nbt.TypeCompound)
		}
		list.Append(stack)
		e.NBT.Set("Inventory", list)
	}
	return len(entities), nil
}

// effectCmd implements `/effect`.
type effectCmd struct{}

// Name implements exec.Handler.
func (c *effectCmd) Name() string { return "effect" }

// Usage implements exec.Handler.
func (c *effectCmd) Usage() string { return "/effect <clear|give> ..." }

// Parse implements exec.Handler.
func (c *effectCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	action, err := p.String()
	if err != nil {
		return nil, err
	}
	switch action {
	case "clear":
		target, err := parseEntityTarget(p)
		if err != nil {
			return nil, err
		}
		return []ast.Arg{&ast.StringArg{Value: action}, target}, nil
	case "give":
		target, err := parseEntityTarget(p)
		if err != nil {
			return nil, err
		}
		effect, err := p.String()
		if err != nil {
			return nil, err
		}
		seconds := 30
		if p.More() {
			if n, err := p.Int(); err == nil {
				seconds = n
			}
		}
		amplifier := 0
		if p.More() {
			if n, err := p.Int(); err == nil {
				amplifier = n
			}
		}
		return []ast.Arg{
			&ast.StringArg{Value: action},
			target,
			&ast.StringArg{Value: effect},
			&ast.IntArg{Value: seconds},
			&ast.IntArg{Value: amplifier},
		}, nil
	}
	return nil, p.Errorf("expected clear or give, got %q", action)
}

// Run implements exec.Handler.
func (c *effectCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	action := args[0].(*ast.StringArg).Value
	entities, err := resolveEntityTargets(ctx, args[1])
	if err != nil {
		return 0, err
	}
	switch action {
	case "clear":
		for _, e := range entities {
			e.NBT.Remove("ActiveEffects")
		}
		return len(entities), nil
	case "give":
		name := args[2].(*ast.StringArg).Value
		seconds := args[3].(*ast.IntArg).Value
		amplifier := args[4].(*ast.IntArg).Value
		for _, e := range entities {
			effect := nbt.NewCompound()
			effect.Set("Id", nbt.String(name))
			effect.Set("Amplifier", nbt.Byte(byte(amplifier)))
			effect.Set("Duration", nbt.Int(int32(seconds*20)))
			effect.Set("Ambient", nbt.Byte(0))
			list, ok := e.NBT.Get("ActiveEffects").(*nbt.List)
			if !ok {
				list = nbt.NewList(nbt.TypeCompound)
			}
			list.Append(effect)
			e.NBT.Set("ActiveEffects", list)
		}
		return len(entities), nil
	}
	return 0, exec.Errorf("unknown action %q", action)
}

// defaultGameModeCmd implements `/defaultgamemode`.
type defaultGameModeCmd struct{}

// Name implements exec.Handler.
func (c *defaultGameModeCmd) Name() string { return "defaultgamemode" }

// Usage implements exec.Handler.
func (c *defaultGameModeCmd) Usage() string { return "/defaultgamemode <mode>" }

// Parse implements exec.Handler.
func (c *defaultGameModeCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	mode, err := p.String()
	if err != nil {
		return nil, err
	}
	return []ast.Arg{&ast.StringArg{Value: mode}}, nil
}

// Run implements exec.Handler.
func (c *defaultGameModeCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	mode := args[0].(*ast.StringArg).Value
	id, ok := gameModeID(mode)
	if !ok {
		return 0, exec.Errorf("unknown game mode %q", mode)
	}
	ctx.World.Rules.Set("defaultGameMode", fmt.Sprintf("%d", id))
	return 1, nil
}

// registerPlayer adds the player commands.
func registerPlayer(reg *exec.Registry) {
	reg.Register(&gamemodeCmd{})
	reg.Register(&defaultGameModeCmd{})
	reg.Register(&giveCmd{})
	reg.Register(&effectCmd{})
}
