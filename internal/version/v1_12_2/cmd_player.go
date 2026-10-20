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
		// 1.12.2 changes the game mode of a *player*: the target is
		// resolved with `CommandAbstract.b`, which keeps only the players
		// it finds.
		entities, err := resolvePlayerTargets(ctx, args[1])
		if err != nil {
			return 0, err
		}
		targets = entities
	}
	for _, e := range targets {
		e.NBT.Set("playerGameType", nbt.Int(int32(id)))
		if e == ctx.Source.Entity {
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Set own game mode to %s", gameModeName(id)))
		} else {
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Set %s's game mode to %s", e.DisplayName(), gameModeName(id)))
		}
	}
	return len(targets), nil
}

// gameModeName returns the name 1.12.2 shows for a game mode.
//
// It comes from the language file: `gameMode.survival` is "Survival Mode",
// not just "Survival".
func gameModeName(id int) string {
	switch id {
	case 0:
		return "Survival Mode"
	case 1:
		return "Creative Mode"
	case 2:
		return "Adventure Mode"
	case 3:
		return "Spectator Mode"
	}
	return "Survival Mode"
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
	// 1.12.2 gives items to a *player*: the target is resolved with
	// `CommandAbstract.b`, which keeps only the players it finds.
	entities, err := resolvePlayerTargets(ctx, args[0])
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
		// 1.12.2 confirms the gift with the item, the amount and the
		// player who got it.
		name := e.PlayerName
		if name == "" {
			name = e.DisplayName()
		}
		// The item is named by its **display name**, and the name is
		// wrapped in square brackets. The display name follows the data
		// value: `minecraft:dye` with the data value 3 is "Cocoa Beans".
		//
		// (Measured on a real 1.12.2 server: `give Steve stone` answers
		// "Given [Stone] * 1 to Steve" and `give Steve minecraft:dye 1 3`
		// answers "Given [Cocoa Beans] * 1 to Steve".)
		shown := item.Name
		if display, known := itemDisplayName(item.Name, item.Data); known {
			shown = "[" + display + "]"
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Given %s * %d to %s", shown, count, name))
	}
	return len(entities), nil
}

// effectCmd implements `/effect`.
//
// The 1.12.2 syntax is `/effect <player> <effect> [seconds] [amplifier]
// [hideParticles] OR /effect <player> clear`: the target comes first and
// `clear` is the *second* argument, not the first.
type effectCmd struct{}

// Name implements exec.Handler.
func (c *effectCmd) Name() string { return "effect" }

// Usage implements exec.Handler.
func (c *effectCmd) Usage() string {
	return "/effect <player> <effect> [seconds] [amplifier] [hideParticles] OR /effect <player> clear"
}

// The defaults and limits of the `/effect` arguments.
//
// `CommandEffect` uses 30 seconds and amplifier 0 when they are not
// given, and it accepts 0..1000000 seconds and 0..255 amplifier levels.
const (
	effectDefaultSeconds = 30
	effectMaxSeconds     = 1000000
	effectMaxAmplifier   = 255
)

// Parse implements exec.Handler.
//
// The arguments are `<player> <effect> [seconds] [amplifier]
// [hideParticles]`. Only the first two are required, and `clear` is
// accepted as the effect name.
func (c *effectCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	effect, err := p.String()
	if err != nil {
		return nil, err
	}
	seconds := effectDefaultSeconds
	if p.More() {
		n, err := p.Int()
		if err != nil {
			return nil, err
		}
		seconds = n
	}
	amplifier := 0
	if p.More() {
		n, err := p.Int()
		if err != nil {
			return nil, err
		}
		amplifier = n
	}
	hide := false
	if p.More() {
		raw, err := p.String()
		if err != nil {
			return nil, err
		}
		// 1.12.2 only looks for "true" here; every other value keeps the
		// particles visible.
		hide = strings.EqualFold(raw, "true")
	}
	return []ast.Arg{
		target,
		&ast.StringArg{Value: effect},
		&ast.IntArg{Value: seconds},
		&ast.IntArg{Value: amplifier},
		&ast.StringArg{Value: strconv.FormatBool(hide)},
	}, nil
}

// Run implements exec.Handler.
func (c *effectCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	name := args[1].(*ast.StringArg).Value
	seconds := args[2].(*ast.IntArg).Value
	amplifier := args[3].(*ast.IntArg).Value
	hide := args[4].(*ast.StringArg).Value == "true"

	entities, err := resolveEntityTargets(ctx, args[0])
	if err != nil {
		return 0, err
	}
	// `clear` is the second argument, so it is checked before the effect
	// is looked up.
	if name == "clear" {
		for _, e := range entities {
			if !hasAnyEffect(e) {
				return 0, exec.Errorf("Couldn't take any effects from %s as they do not have any",
					effectTargetName(e))
			}
		}
		for _, e := range entities {
			if e.NBT != nil {
				e.NBT.Remove("ActiveEffects")
			}
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Took all effects from %s", effectTargetName(e)))
		}
		return len(entities), nil
	}
	eff, ok := lookupMobEffect(name)
	if !ok {
		return 0, exec.Errorf("There is no such mob effect with ID %s", name)
	}
	if seconds < 0 {
		return 0, exec.Errorf("The number you have entered (%d) is too small, it must be at least 0", seconds)
	}
	if seconds > effectMaxSeconds {
		return 0, exec.Errorf("The number you have entered (%d) is too big, it must be at most %d",
			seconds, effectMaxSeconds)
	}
	if amplifier < 0 {
		return 0, exec.Errorf("The number you have entered (%d) is too small, it must be at least 0", amplifier)
	}
	if amplifier > effectMaxAmplifier {
		return 0, exec.Errorf("The number you have entered (%d) is too big, it must be at most %d",
			amplifier, effectMaxAmplifier)
	}
	for _, e := range entities {
		target := effectTargetName(e)
		if seconds == 0 {
			// A duration of zero takes the effect away again.
			if !removeMobEffect(e, eff.id) {
				return 0, exec.Errorf("Couldn't take %s from %s as they do not have the effect",
					eff.display, target)
			}
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Took %s from %s", eff.display, target))
			continue
		}
		addMobEffect(e, eff, seconds*20, amplifier, !hide)
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Given %s (ID %d) * %d to %s for %d seconds",
				eff.display, eff.id, amplifier, target, seconds))
	}
	return len(entities), nil
}

// mobEffect is one of the mob effects 1.12.2 knows.
//
// The table comes from the 1.12.2 server's `MobEffectList` registry: the
// id is the registry id (the one `/effect` reports and the one the
// entity NBT stores) and the name is the registry key, which is what
// `/effect` accepts as the effect name.
type mobEffect struct {
	id      int
	name    string
	display string
}

// mobEffects lists the 1.12.2 mob effects in registry order.
//
// There are 27 of them: `luck` and `unluck` are the last two, and
// 1.12.2 has no conduit power, slow falling, digger or repair effect.
var mobEffects = []mobEffect{
	{1, "speed", "Speed"},
	{2, "slowness", "Slowness"},
	{3, "haste", "Haste"},
	{4, "mining_fatigue", "Mining Fatigue"},
	{5, "strength", "Strength"},
	{6, "instant_health", "Instant Health"},
	{7, "instant_damage", "Instant Damage"},
	{8, "jump_boost", "Jump Boost"},
	{9, "nausea", "Nausea"},
	{10, "regeneration", "Regeneration"},
	{11, "resistance", "Resistance"},
	{12, "fire_resistance", "Fire Resistance"},
	{13, "water_breathing", "Water Breathing"},
	{14, "invisibility", "Invisibility"},
	{15, "blindness", "Blindness"},
	{16, "night_vision", "Night Vision"},
	{17, "hunger", "Hunger"},
	{18, "weakness", "Weakness"},
	{19, "poison", "Poison"},
	{20, "wither", "Wither"},
	{21, "health_boost", "Health Boost"},
	{22, "absorption", "Absorption"},
	{23, "saturation", "Saturation"},
	{24, "glowing", "Glowing"},
	{25, "levitation", "Levitation"},
	{26, "luck", "Luck"},
	{27, "unluck", "Bad Luck"},
}

// lookupMobEffect resolves the `<effect>` argument.
//
// 1.12.2 tries the registry id first and falls back to the name, which is
// looked up as a resource location, so the namespace may be given
// ("minecraft:speed").
func lookupMobEffect(text string) (mobEffect, bool) {
	if n, err := strconv.Atoi(text); err == nil {
		for _, e := range mobEffects {
			if e.id == n {
				return e, true
			}
		}
		return mobEffect{}, false
	}
	name := text
	if i := strings.IndexByte(name, ':'); i >= 0 {
		name = name[i+1:]
	}
	for _, e := range mobEffects {
		if e.name == name {
			return e, true
		}
	}
	return mobEffect{}, false
}

// effectTargetName returns the name 1.12.2 reports for an effect's
// target.
//
// A player is known by its player name, any other entity by its
// display name.
func effectTargetName(e *world.Entity) string {
	if e.IsPlayer && e.PlayerName != "" {
		return e.PlayerName
	}
	return e.DisplayName()
}

// hasAnyEffect reports whether the entity has any active effect.
func hasAnyEffect(e *world.Entity) bool {
	if e.NBT == nil {
		return false
	}
	list, ok := e.NBT.Get("ActiveEffects").(*nbt.List)
	return ok && list.Len() > 0
}

// addMobEffect adds an effect to an entity.
//
// 1.12.2 keeps the effects in the `ActiveEffects` list of the entity
// NBT: one compound per effect with the registry id, the amplifier, the
// duration in ticks and the two flags. `MobEffect.a` merges a new effect
// into an existing one of the same kind: the stronger amplifier wins
// and, at the same amplifier, the longer duration wins.
func addMobEffect(e *world.Entity, eff mobEffect, duration, amplifier int, showParticles bool) {
	if e.NBT == nil {
		return
	}
	list, ok := e.NBT.Get("ActiveEffects").(*nbt.List)
	if !ok {
		list = nbt.NewList(nbt.TypeCompound)
		e.NBT.Set("ActiveEffects", list)
	}
	for _, item := range list.Items() {
		c, ok := item.(*nbt.Compound)
		if !ok || effectID(c) != eff.id {
			continue
		}
		old := effectInt(c, "Amplifier")
		if amplifier > old {
			c.Set("Amplifier", nbt.Byte(byte(amplifier)))
			c.Set("Duration", nbt.Int(int32(duration)))
		} else if amplifier == old && duration > effectInt(c, "Duration") {
			c.Set("Duration", nbt.Int(int32(duration)))
		}
		c.Set("ShowParticles", nbt.Byte(boolByte(showParticles)))
		return
	}
	entry := nbt.NewCompound()
	entry.Set("Id", nbt.Byte(byte(eff.id)))
	entry.Set("Amplifier", nbt.Byte(byte(amplifier)))
	entry.Set("Duration", nbt.Int(int32(duration)))
	entry.Set("Ambient", nbt.Byte(0))
	entry.Set("ShowParticles", nbt.Byte(boolByte(showParticles)))
	list.Append(entry)
}

// removeMobEffect takes an effect away from an entity.
//
// It reports whether the entity had the effect at all.
func removeMobEffect(e *world.Entity, id int) bool {
	if e.NBT == nil {
		return false
	}
	list, ok := e.NBT.Get("ActiveEffects").(*nbt.List)
	if !ok {
		return false
	}
	kept := nbt.NewList(nbt.TypeCompound)
	found := false
	for _, item := range list.Items() {
		if c, ok := item.(*nbt.Compound); ok && effectID(c) == id {
			found = true
			continue
		}
		kept.Append(item)
	}
	if found {
		e.NBT.Set("ActiveEffects", kept)
	}
	return found
}

// effectID reads the registry id of an effect compound.
func effectID(c *nbt.Compound) int { return effectInt(c, "Id") }

// effectInt reads a numeric field of an effect compound.
func effectInt(c *nbt.Compound, key string) int {
	switch v := c.Get(key).(type) {
	case nbt.Byte:
		return int(v)
	case nbt.Short:
		return int(v)
	case nbt.Int:
		return int(v)
	}
	return 0
}

// boolByte renders a boolean the way 1.12.2 stores it: as a byte.
func boolByte(b bool) byte {
	if b {
		return 1
	}
	return 0
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
