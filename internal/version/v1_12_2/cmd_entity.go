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
func (c *summonCmd) Usage() string { return usageLine("summon") }

// Parse implements exec.Handler.
func (c *summonCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	entityType, err := p.String()
	if err != nil {
		return nil, err
	}
	pos := ast.Coord{}
	if p.More() {
		parsed, err := p.CoordWith(coordSummon)
		if err != nil {
			return nil, err
		}
		pos = *parsed
	}
	text := ""
	if p.More() {
		// The dataTag is parsed when the command runs: 1.12.2 builds the
		// entity first and only then looks at the dataTag.
		text, err = dataTagText(p)
		if err != nil {
			return nil, err
		}
	}
	return []ast.Arg{
		&ast.EntityTypeArg{Name: entityType},
		&ast.Coord{X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ},
		&ast.StringArg{Value: text},
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
	// 1.12.2 refuses a position whose chunk is not loaded.
	if !ctx.World.IsLoaded(world.Pos{X: int(pos.X), Y: int(pos.Y), Z: int(pos.Z)}) {
		return 0, exec.Errorf("Cannot summon the object out of the world")
	}
	e.Pos = pos
	if text := args[2].(*ast.StringArg).Value; text != "" {
		data, err := parseDataTagArg(text)
		if err != nil {
			return 0, err
		}
		mergeNBT(e.NBT, data)
	}
	if def := ctx.World.EntityTypes.Get(canonical); def != nil && def.NeedsItemStack {
		// 1.12.2's entity factory only accepts an item entity that
		// carries a non-empty item stack of an item the server knows.
		// An empty stack makes the entity die right after it is
		// created, and a stack of a name that is not an item is empty
		// too, so the summon fails.
		//
		// (Measured on a real 1.12.2 server: `summon item` and
		// `summon item {Item:{}}` both answer "Unable to summon
		// object", and so does `summon item {Item:{id:"stone"}}` —
		// the count is what decides, see hasItemStack. `summon item
		// {Item:{id:"stone",Count:1b}}` works, and
		// `{Item:{id:"notanitem",Count:1b}}` does not.)
		if !hasItemStack(e.NBT) {
			return 0, exec.Errorf("Unable to summon object")
		}
	}
	ctx.World.AddEntity(e)
	// 1.12.2 does not say where the entity was put; it only confirms the
	// summon.
	ctx.Engine.Feedback = append(ctx.Engine.Feedback, "Object successfully summoned")
	return 1, nil
}

// hasItemStack reports whether the entity carries a non-empty item stack.
//
// 1.12.2 reads the stack with `ItemStack.a(NBTTagCompound)`: the count
// comes from `getByte("Count")`, so a missing count counts as zero, and
// a stack whose id names no item is empty too. An empty stack makes the
// entity die right after it is created, so the summon fails.
func hasItemStack(n *nbt.Compound) bool {
	item, ok := n.Get("Item").(*nbt.Compound)
	if !ok {
		return false
	}
	id, _ := item.Get("id").(nbt.String)
	if strings.TrimSpace(string(id)) == "" {
		return false
	}
	if !world.IsKnownItem(string(id)) {
		// The stack names something that is not an item, so it is empty.
		return false
	}
	return itemStackCount(item.Get("Count")) > 0
}

// itemStackCount reads the count of an item stack the way 1.12.2 does.
//
// `getByte` truncates any number to a **signed** byte, so a count of 256
// becomes 0 and a count of -1 stays -1. Only a positive byte is a stack
// with something in it.
//
// (Measured on a real 1.12.2 server: `Count:1` as an int, a short, a
// long and a float all summon, `Count:256` and `Count:-1` do not.)
func itemStackCount(v nbt.Value) int {
	switch n := v.(type) {
	case nbt.Byte:
		return int(n)
	case nbt.Short:
		return int(int8(n))
	case nbt.Int:
		return int(int8(n))
	case nbt.Long:
		return int(int8(n))
	case nbt.Float:
		return int(int8(int32(n)))
	case nbt.Double:
		return int(int8(int32(n)))
	}
	return 0
}

// killCmd implements `/kill`.
type killCmd struct{}

// Name implements exec.Handler.
func (c *killCmd) Name() string { return "kill" }

// Usage implements exec.Handler.
func (c *killCmd) Usage() string { return usageLine("kill") }

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
func (c *entityDataCmd) Usage() string { return usageLine("entitydata") }

// Parse implements exec.Handler.
func (c *entityDataCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	// The dataTag is parsed when the command runs: 1.12.2 resolves the
	// target first and only then looks at the dataTag.
	text, err := dataTagText(p)
	if err != nil {
		return nil, err
	}
	return []ast.Arg{target, &ast.StringArg{Value: text}}, nil
}

// Run implements exec.Handler.
func (c *entityDataCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	entities, err := resolveEntityTargets(ctx, args[0])
	if err != nil {
		return 0, err
	}
	data, err := parseDataTagArg(args[1].(*ast.StringArg).Value)
	if err != nil {
		return 0, err
	}
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
func (c *testForCmd) Usage() string { return usageLine("testfor") }

// Parse implements exec.Handler.
func (c *testForCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	// The dataTag is parsed when the command runs: 1.12.2 resolves the
	// target first and only then looks at the dataTag.
	text := ""
	if p.More() {
		text, err = dataTagText(p)
		if err != nil {
			return nil, err
		}
	}
	return []ast.Arg{target, &ast.StringArg{Value: text}}, nil
}

// Run implements exec.Handler.
func (c *testForCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	entities, err := resolveEntityTargets(ctx, args[0])
	if err != nil {
		return 0, err
	}
	text := args[1].(*ast.StringArg).Value
	var data *nbt.Compound
	if text != "" {
		data, err = parseDataTagArg(text)
		if err != nil {
			return 0, err
		}
	}
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
func (c *tpCmd) Usage() string { return usageLine(c.name) }

// Parse implements exec.Handler.
func (c *tpCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	pos, err := p.CoordWith(coordTp)
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
