package v1_12_2

import (
	"fmt"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// summonCmd implements `/summon`.
type summonCmd struct{}

// Name implements exec.Handler.
func (c *summonCmd) Name() string { return "summon" }

// Usage implements exec.Handler.
func (c *summonCmd) Usage() string { return "/summon <entityname> [x] [y] [z] [dataTag]" }

// Parse implements exec.Handler.
func (c *summonCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	entityType, err := p.String()
	if err != nil {
		return nil, err
	}
	pos := ast.Coord{}
	if p.More() {
		parsed, err := p.Coord(ast.Coord{})
		if err != nil {
			return nil, err
		}
		pos = *parsed
	}
	var data *nbt.Compound
	if p.More() {
		arg, err := p.NBT()
		if err != nil {
			return nil, err
		}
		data = arg.Value
	}
	return []ast.Arg{
		&ast.EntityTypeArg{Name: entityType},
		&ast.Coord{X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ},
		&ast.NBTArg{Value: data},
	}, nil
}

// Run implements exec.Handler.
func (c *summonCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	name := args[0].(*ast.EntityTypeArg).Name
	canonical, ok := ctx.World.EntityTypes.ParseEntityTypeName(name)
	if !ok {
		// 1.12.2 does not check the type name at all: the command puts
		// the raw text into the NBT "id" and lets the entity factory
		// build the entity. Every name the factory cannot build is
		// reported the same way.
		return 0, exec.Errorf("Unable to summon object")
	}
	if def := ctx.World.EntityTypes.Get(canonical); def == nil || def.IsPlayer {
		// 1.12.2 keeps the players out of the entity registry, so a
		// player is not something `/summon` can bring into the world.
		return 0, exec.Errorf("Unable to summon object")
	}
	posArg := args[1].(*ast.Coord)
	pos := world.Vec{X: posArg.X, Y: posArg.Y, Z: posArg.Z}
	if posArg.RelativeX {
		pos.X = ctx.Source.Pos.X + pos.X
	}
	if posArg.RelativeY {
		pos.Y = ctx.Source.Pos.Y + pos.Y
	}
	if posArg.RelativeZ {
		pos.Z = ctx.Source.Pos.Z + pos.Z
	}
	e, ok := ctx.World.EntityTypes.NewEntity(canonical, "")
	if !ok {
		return 0, exec.Errorf("Unable to summon object")
	}
	e.Pos = pos
	if data := exec.NBTOf(args, 2); data != nil {
		mergeNBT(e.NBT, data)
	}
	ctx.World.AddEntity(e)
	// 1.12.2 does not say where the entity was put; it only confirms the
	// summon.
	ctx.Engine.Feedback = append(ctx.Engine.Feedback, "Object successfully summoned")
	return 1, nil
}

// killCmd implements `/kill`.
type killCmd struct{}

// Name implements exec.Handler.
func (c *killCmd) Name() string { return "kill" }

// Usage implements exec.Handler.
func (c *killCmd) Usage() string { return "/kill [<target>]" }

// Parse implements exec.Handler.
func (c *killCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	if !p.More() {
		return nil, nil
	}
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	return []ast.Arg{target}, nil
}

// Run implements exec.Handler.
func (c *killCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	if len(args) == 0 {
		if ctx.Source.Entity == nil {
			return 0, exec.Errorf("no entity to kill")
		}
		ctx.World.RemoveEntity(ctx.Source.Entity.ID)
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			"Killed "+ctx.Source.Entity.DisplayName())
		return 1, nil
	}
	entities, err := resolveEntityTargets(ctx, args[0])
	if err != nil {
		return 0, err
	}
	for _, e := range entities {
		ctx.World.RemoveEntity(e.ID)
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, "Killed "+e.DisplayName())
	}
	return len(entities), nil
}

// entityDataCmd implements `/entitydata`.
type entityDataCmd struct{}

// Name implements exec.Handler.
func (c *entityDataCmd) Name() string { return "entitydata" }

// Usage implements exec.Handler.
func (c *entityDataCmd) Usage() string { return "/entitydata <target> <nbt>" }

// Parse implements exec.Handler.
func (c *entityDataCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	data, err := p.NBT()
	if err != nil {
		return nil, err
	}
	return []ast.Arg{target, data}, nil
}

// Run implements exec.Handler.
func (c *entityDataCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	entities, err := resolveEntityTargets(ctx, args[0])
	if err != nil {
		return 0, err
	}
	data := exec.NBTOf(args, 1)
	for _, e := range entities {
		mergeNBT(e.NBT, data)
		// 1.12.2 echoes the whole entity data back.
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			"Entity data updated to: "+e.NBT.String())
	}
	return len(entities), nil
}

// testForCmd implements `/testfor`.
type testForCmd struct{}

// Name implements exec.Handler.
func (c *testForCmd) Name() string { return "testfor" }

// Usage implements exec.Handler.
func (c *testForCmd) Usage() string { return "/testfor <target> [<nbt>]" }

// Parse implements exec.Handler.
func (c *testForCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	var data *nbt.Compound
	if p.More() {
		arg, err := p.NBT()
		if err != nil {
			return nil, err
		}
		data = arg.Value
	}
	return []ast.Arg{target, &ast.NBTArg{Value: data}}, nil
}

// Run implements exec.Handler.
func (c *testForCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	entities, err := resolveEntityTargets(ctx, args[0])
	if err != nil {
		return 0, err
	}
	data := exec.NBTOf(args, 1)
	if data != nil {
		for _, e := range entities {
			if !nbt.Matches(e.NBT, data) {
				return 0, exec.Errorf("%s did not match the required data structure", e.DisplayName())
			}
		}
	}
	for _, e := range entities {
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, "Found "+e.DisplayName())
	}
	return len(entities), nil
}

// tpCmd implements `/tp` and `/teleport`.
type tpCmd struct {
	// name is the command name.
	name string
}

// Name implements exec.Handler.
func (c *tpCmd) Name() string { return c.name }

// Usage implements exec.Handler.
func (c *tpCmd) Usage() string { return "/" + c.name + " <target> <x> <y> <z>" }

// Parse implements exec.Handler.
func (c *tpCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	pos, err := p.Coord(ast.Coord{})
	if err != nil {
		return nil, err
	}
	return []ast.Arg{
		target,
		&ast.Coord{X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ},
	}, nil
}

// Run implements exec.Handler.
func (c *tpCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	entities, err := resolveEntityTargets(ctx, args[0])
	if err != nil {
		return 0, err
	}
	posArg := args[1].(*ast.Coord)
	for _, e := range entities {
		pos := world.Vec{X: posArg.X, Y: posArg.Y, Z: posArg.Z}
		if posArg.RelativeX {
			pos.X = e.Pos.X + pos.X
		}
		if posArg.RelativeY {
			pos.Y = e.Pos.Y + pos.Y
		}
		if posArg.RelativeZ {
			pos.Z = e.Pos.Z + pos.Z
		}
		e.Pos = pos
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Teleported %s to %.1f, %.1f, %.1f",
				e.DisplayName(), pos.X, pos.Y, pos.Z))
	}
	return len(entities), nil
}

// mergeNBT merges src into dst, overwriting existing keys.
//
// This is the behaviour `/entitydata` and `/blockdata` have in 1.12.2: the
// given compound is merged into the target's NBT. A compound value is
// merged recursively, so writing {Item:{Count:64}} keeps the other fields
// of the item; every other type replaces the old value.
func mergeNBT(dst, src *nbt.Compound) {
	if dst == nil || src == nil {
		return
	}
	for _, k := range src.Keys() {
		if sub, ok := src.Get(k).(*nbt.Compound); ok {
			if old, ok := dst.Get(k).(*nbt.Compound); ok {
				mergeNBT(old, sub)
				continue
			}
		}
		dst.Set(k, nbt.CloneValue(src.Get(k)))
	}
}

// registerEntity adds the entity commands.
func registerEntity(reg *exec.Registry) {
	reg.Register(&summonCmd{})
	reg.Register(&killCmd{})
	reg.Register(&entityDataCmd{})
	reg.Register(&testForCmd{})
	reg.Register(&tpCmd{name: "tp"})
	reg.Register(&tpCmd{name: "teleport"})
}

// entityTypeName renders an entity type for messages.
func entityTypeName(arg ast.Arg) string {
	if t, ok := arg.(*ast.EntityTypeArg); ok {
		return t.Name
	}
	return strings.TrimSpace("?")
}
