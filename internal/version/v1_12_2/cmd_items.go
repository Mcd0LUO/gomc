package v1_12_2

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// clearCmd implements `/clear`.
//
// The 1.12.2 syntax is `clear [<target>] [<item> [<data> [<maxCount>]]
// [<nbt>]]`. Without an item the whole inventory is cleared, and without a
// target the player running the command is cleared.
type clearCmd struct{}

// Name implements exec.Handler.
func (c *clearCmd) Name() string { return "clear" }

// Usage implements exec.Handler.
func (c *clearCmd) Usage() string { return usageLine("clear") }

// Parse implements exec.Handler.
func (c *clearCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target := ast.Arg(&ast.StringArg{Value: ""})
	if p.More() {
		t, ok := p.Peek()
		if !ok {
			return nil, p.Errorf("expected a target")
		}
		if IsSelector(t.Text) {
			sel, err := parseSelector(p)
			if err != nil {
				return nil, err
			}
			target = sel
		} else {
			name, err := p.String()
			if err != nil {
				return nil, err
			}
			target = &ast.StringArg{Value: name}
		}
	}
	item := ""
	if p.More() {
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		item = name
	}
	data := -1
	if p.More() {
		n, err := p.Int()
		if err != nil {
			return nil, err
		}
		data = n
	}
	maxCount := -1
	if p.More() {
		n, err := p.Int()
		if err != nil {
			return nil, err
		}
		maxCount = n
	}
	var tag *nbt.Compound
	if p.More() {
		arg, err := p.NBT()
		if err != nil {
			return nil, err
		}
		tag = arg.Value
	}
	return []ast.Arg{
		target,
		&ast.StringArg{Value: item},
		&ast.IntArg{Value: data},
		&ast.IntArg{Value: maxCount},
		&ast.NBTArg{Value: tag},
	}, nil
}

// Run implements exec.Handler.
func (c *clearCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	// 1.12.2 clears the inventory of a *player*: the target is resolved
	// with `CommandAbstract.b`, which keeps only the players it finds.
	holders, err := resolvePlayerTargets(ctx, args[0])
	if err != nil {
		return 0, err
	}
	item := args[1].(*ast.StringArg).Value
	data := args[2].(*ast.IntArg).Value
	maxCount := args[3].(*ast.IntArg).Value
	tag := exec.NBTOf(args, 4)

	// An empty target list makes the command run zero times on the
	// server, so it says nothing at all.
	if len(holders) == 0 {
		return 0, nil
	}

	removed := 0
	for _, e := range holders {
		removed += clearInventory(e, item, data, maxCount, tag)
	}
	if removed == 0 {
		return 0, exec.Errorf("Could not clear the inventory of %s, no items to remove",
			holders[0].DisplayName())
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Cleared the inventory of %s, removing %d items",
			holders[0].DisplayName(), removed))
	// The items go into the `AffectedItems` channel.
	//
	// (`CommandClear` reports
	// `EnumCommandResult.AFFECTED_ITEMS`.)
	ctx.AffectedItems = removed
	return removed, nil
}

// clearInventory removes the matching items from an entity's inventory.
//
// It returns the number of removed stacks.
func clearInventory(e *world.Entity, item string, data, maxCount int, tag *nbt.Compound) int {
	list, ok := e.NBTValue("Inventory").(*nbt.List)
	if !ok {
		return 0
	}
	kept := nbt.NewList(nbt.TypeCompound)
	removed := 0
	for _, v := range list.Items() {
		stack, ok := v.(*nbt.Compound)
		if !ok {
			continue
		}
		if item != "" && !itemMatches(stack, item, data, tag) {
			kept.Append(stack)
			continue
		}
		removed++
		if maxCount == 0 {
			// A maximum of zero clears the slot without dropping
			// anything, so the stack stays.
			kept.Append(stack)
		}
	}
	e.NBT.Set("Inventory", kept)
	return removed
}

// itemMatches reports whether an item stack matches the given filter.
func itemMatches(stack *nbt.Compound, item string, data int, tag *nbt.Compound) bool {
	id, _ := stack.Get("id").(nbt.String)
	if item != "" && !strings.EqualFold(string(id), item) {
		return false
	}
	if data >= 0 {
		damage, _ := stack.Get("Damage").(nbt.Short)
		if int(damage) != data {
			return false
		}
	}
	if tag != nil {
		stackTag, _ := stack.Get("tag").(*nbt.Compound)
		if stackTag == nil || !nbt.Matches(stackTag, tag) {
			return false
		}
	}
	return true
}

// replaceItemCmd implements `/replaceitem`.
//
// The 1.12.2 syntax is `replaceitem <entity|block> ...`. For an entity it
// is `replaceitem entity <target> <slot> <item> [<count>]`; for a block it
// is `replaceitem block <x> <y> <z> <slot> <item> [<count>]`.
type replaceItemCmd struct{}

// Name implements exec.Handler.
func (c *replaceItemCmd) Name() string { return "replaceitem" }

// Usage implements exec.Handler.
func (c *replaceItemCmd) Usage() string { return usageLine("replaceitem") }

// Parse implements exec.Handler.
func (c *replaceItemCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	mode, err := p.String()
	if err != nil {
		return nil, err
	}
	switch mode {
	case "entity", "block":
	default:
		return nil, p.Errorf("expected entity or block, got %q", mode)
	}
	if mode == "block" {
		pos, err := p.CoordWith(coordBlockPos)
		if err != nil {
			return nil, err
		}
		slot, err := p.String()
		if err != nil {
			return nil, err
		}
		item, count, err := parseItemAndCount(p)
		if err != nil {
			return nil, err
		}
		return []ast.Arg{
			&ast.StringArg{Value: mode},
			&ast.Coord{X: pos.X, Y: pos.Y, Z: pos.Z,
				RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ},
			&ast.StringArg{Value: slot},
			&ast.ItemArg{Name: item, Count: count, Data: -1},
		}, nil
	}
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	slot, err := p.String()
	if err != nil {
		return nil, err
	}
	item, count, err := parseItemAndCount(p)
	if err != nil {
		return nil, err
	}
	return []ast.Arg{
		&ast.StringArg{Value: mode},
		target,
		&ast.StringArg{Value: slot},
		&ast.ItemArg{Name: item, Count: count, Data: -1},
	}, nil
}

// parseItemAndCount parses an item name and an optional count.
func parseItemAndCount(p *parser.Parser) (string, int, error) {
	item, err := p.String()
	if err != nil {
		return "", 0, err
	}
	count := 1
	if p.More() {
		n, err := p.Int()
		if err != nil {
			return "", 0, err
		}
		count = n
	}
	return item, count, nil
}

// Run implements exec.Handler.
func (c *replaceItemCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	mode := args[0].(*ast.StringArg).Value
	if mode == "block" {
		// A block's container is not modelled, so there is nothing to
		// replace.
		return 0, exec.Errorf("the block at the given position has no replaceable slot")
	}
	target := args[1]
	slot := args[2].(*ast.StringArg).Value
	item := args[3].(*ast.ItemArg)

	holders, err := resolveEntityTargets(ctx, target)
	if err != nil {
		return 0, err
	}
	// An empty target list makes the command run zero times on the
	// server, so it says nothing at all.
	if len(holders) == 0 {
		return 0, nil
	}
	index, ok := lookupReplaceItemSlot(slot)
	if !ok {
		// The name is not in the table of 1.12.2, so the command never
		// gets as far as the entity.
		//
		// (Measured on a real 1.12.2 server, probe case
		// 58-replaceitem-player: "replaceitem entity Steve
		// slot.inventory.27" says "'slot.inventory.27' is not a valid
		// parameter".)
		return 0, exec.Errorf("'%s' is not a valid parameter", slot)
	}
	stack := astItem{name: item.Name, count: item.Count, data: item.Data}
	display := replaceItemDisplayName(stack)
	for _, e := range holders {
		// The slot has to be one the entity has. 1.12.2 asks the
		// entity, and the entity refuses the slot when it has no such
		// container.
		//
		// (Measured on a real 1.12.2 server, probe cases 56 and 57: an
		// armour stand refuses slot.armor.chest and slot.inventory.18,
		// a zombie refuses slot.inventory.0, and a player takes them
		// all.)
		if !containerOf(e).has(index) {
			return 0, replaceItemError(slot, stack, display)
		}
		replaceItemInSlot(e, index, &stack)
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Replaced slot %s with %d * [%s]", slot, stack.count, display))
	return len(holders), nil
}

// enchantCmd implements `/enchant`.
//
// The 1.12.2 syntax is `enchant <target> <enchantment> [<level>]`, and the
// target has to be a living entity holding an item.
type enchantCmd struct{}

// Name implements exec.Handler.
func (c *enchantCmd) Name() string { return "enchant" }

// Usage implements exec.Handler.
func (c *enchantCmd) Usage() string { return usageLine("enchant") }

// Parse implements exec.Handler.
func (c *enchantCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	enchantment, err := p.String()
	if err != nil {
		return nil, err
	}
	level := 1
	if p.More() {
		level, err = p.Int()
		if err != nil {
			return nil, err
		}
	}
	return []ast.Arg{
		target,
		&ast.StringArg{Value: enchantment},
		&ast.IntArg{Value: level},
	}, nil
}

// Run implements exec.Handler.
func (c *enchantCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	holders, err := resolveEntityTargets(ctx, args[0])
	if err != nil {
		return 0, err
	}
	enchantment := args[1].(*ast.StringArg).Value
	level := args[2].(*ast.IntArg).Value
	if level < 1 {
		level = 1
	}
	// An empty target list makes the command run zero times on the
	// server, so it says nothing at all.
	if len(holders) == 0 {
		return 0, nil
	}
	for _, e := range holders {
		stack := mainHand(e)
		if stack == nil {
			return 0, exec.Errorf("The target doesn't hold an item")
		}
		list, _ := stack.Get("Enchantments").(*nbt.List)
		if list == nil {
			list = nbt.NewList(nbt.TypeCompound)
		}
		entry := nbt.NewCompound()
		entry.Set("id", nbt.String(enchantment))
		entry.Set("lvl", nbt.Short(int16(level)))
		list.Append(entry)
		stack.Set("Enchantments", list)
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Enchanting succeeded"))
	return len(holders), nil
}

// mainHand returns the item stack in the entity's main hand.
func mainHand(e *world.Entity) *nbt.Compound {
	list, ok := e.NBTValue("Inventory").(*nbt.List)
	if !ok {
		return nil
	}
	values := list.Items()
	if len(values) == 0 {
		return nil
	}
	stack, _ := values[0].(*nbt.Compound)
	if stack == nil || stack.Len() == 0 {
		return nil
	}
	return stack
}

// xpCmd implements `/xp`.
//
// The 1.12.2 syntax is `xp <amount> [<player>]`, where a trailing `l` on
// the amount counts experience levels instead of points, and a negative
// amount takes experience away.
type xpCmd struct{}

// Name implements exec.Handler.
func (c *xpCmd) Name() string { return "xp" }

// Usage implements exec.Handler.
func (c *xpCmd) Usage() string { return usageLine("xp") }

// Parse implements exec.Handler.
func (c *xpCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	amount, err := p.String()
	if err != nil {
		return nil, err
	}
	target := ast.Arg(&ast.StringArg{Value: ""})
	if p.More() {
		t, ok := p.Peek()
		if !ok {
			return nil, p.Errorf("expected a player")
		}
		if IsSelector(t.Text) {
			sel, err := parseSelector(p)
			if err != nil {
				return nil, err
			}
			target = sel
		} else {
			name, err := p.String()
			if err != nil {
				return nil, err
			}
			target = &ast.StringArg{Value: name}
		}
	}
	return []ast.Arg{
		&ast.StringArg{Value: amount},
		target,
	}, nil
}

// Run implements exec.Handler.
func (c *xpCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	raw := args[0].(*ast.StringArg).Value
	levels := strings.HasSuffix(raw, "l") || strings.HasSuffix(raw, "L")
	if levels {
		raw = raw[:len(raw)-1]
	}
	amount, err := strconv.Atoi(raw)
	if err != nil {
		return 0, exec.Errorf("expected a number, got %q", raw)
	}
	holders, err := resolvePlayerTargets(ctx, args[1])
	if err != nil {
		return 0, err
	}
	// An empty target list makes the command run zero times on the
	// server, so it says nothing at all.
	if len(holders) == 0 {
		return 0, nil
	}
	for _, e := range holders {
		if levels {
			level := 0
			if v, ok := e.NBTValue("XpLevel").(nbt.Int); ok {
				level = int(v)
			}
			level += amount
			if level < 0 {
				level = 0
			}
			e.NBT.Set("XpLevel", nbt.Int(int32(level)))
		} else {
			if amount < 0 {
				return 0, exec.Errorf("cannot take experience points away")
			}
			total := 0
			if v, ok := e.NBTValue("XpTotal").(nbt.Int); ok {
				total = int(v)
			}
			e.NBT.Set("XpTotal", nbt.Int(int32(total+amount)))
		}
	}
	if levels {
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Given %d levels to %s", amount, holders[0].DisplayName()))
	} else {
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Given %d experience to %s", amount, holders[0].DisplayName()))
	}
	return len(holders), nil
}

// registerItems adds the inventory commands.
func registerItems(reg *exec.Registry) {
	reg.Register(&clearCmd{})
	reg.Register(&replaceItemCmd{})
	reg.Register(&enchantCmd{})
	reg.Register(&xpCmd{})
}
