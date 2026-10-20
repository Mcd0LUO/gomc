package v1_12_2

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/parser"
)

// titleCmd implements `/title`.
//
// The 1.12.2 syntax is `title <target> <title|subtitle|actionbar|clear|
// reset|times> [<arguments>]`:
//
//   - `title`, `subtitle` and `actionbar` take a JSON message.
//   - `clear` and `reset` take nothing.
//   - `times` takes the fade in, stay and fade out durations.
type titleCmd struct{}

// Name implements exec.Handler.
func (c *titleCmd) Name() string { return "title" }

// Usage implements exec.Handler.
func (c *titleCmd) Usage() string {
	return "/title <target> <title|subtitle|actionbar|clear|reset|times> [<arguments>]"
}

// Parse implements exec.Handler.
func (c *titleCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	action, err := p.String()
	if err != nil {
		return nil, err
	}
	switch action {
	case "title", "subtitle", "actionbar":
		msg, err := p.JSON()
		if err != nil {
			return nil, err
		}
		return []ast.Arg{target, &ast.StringArg{Value: action}, msg}, nil
	case "clear", "reset":
		return []ast.Arg{target, &ast.StringArg{Value: action}}, nil
	case "times":
		fadeIn, err := p.Int()
		if err != nil {
			return nil, err
		}
		stay, err := p.Int()
		if err != nil {
			return nil, err
		}
		fadeOut, err := p.Int()
		if err != nil {
			return nil, err
		}
		return []ast.Arg{
			target,
			&ast.StringArg{Value: action},
			&ast.IntArg{Value: fadeIn},
			&ast.IntArg{Value: stay},
			&ast.IntArg{Value: fadeOut},
		}, nil
	}
	return nil, p.Errorf("expected title, subtitle, actionbar, clear, reset or times, got %q", action)
}

// Run implements exec.Handler.
func (c *titleCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	// 1.12.2 shows the title to a *player*: the target is resolved with
	// `CommandAbstract.b`, which keeps only the players it finds.
	players, err := resolvePlayerTargets(ctx, args[0])
	if err != nil {
		return 0, err
	}
	action := args[1].(*ast.StringArg).Value

	for _, e := range players {
		title := e.NBTValue("Title")
		state, _ := title.(*nbt.Compound)
		if state == nil {
			state = nbt.NewCompound()
			e.NBT.Set("Title", state)
		}
		switch action {
		case "title", "subtitle", "actionbar":
			msg := ""
			if len(args) > 2 {
				if j, ok := args[2].(*ast.JSONArg); ok {
					msg = j.Value
				}
			}
			state.Set(action, nbt.String(msg))
		case "clear":
			state.Set("clear", nbt.Byte(1))
		case "reset":
			state.Set("reset", nbt.Byte(1))
		case "times":
			times := nbt.NewCompound()
			times.Set("fadeIn", nbt.Int(args[2].(*ast.IntArg).Value))
			times.Set("stay", nbt.Int(args[3].(*ast.IntArg).Value))
			times.Set("fadeOut", nbt.Int(args[4].(*ast.IntArg).Value))
			state.Set("times", times)
		}
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Title command successfully executed"))
	return len(players), nil
}

// playSoundCmd implements `/playsound`.
//
// The 1.12.2 syntax is `playsound <sound> <source> <target> [<x> <y> <z>]
// [<volume>] [<pitch>] [<minimumVolume>]`. Note that the target comes
// after the sound source.
type playSoundCmd struct{}

// Name implements exec.Handler.
func (c *playSoundCmd) Name() string { return "playsound" }

// Usage implements exec.Handler.
func (c *playSoundCmd) Usage() string {
	return "/playsound <sound> <source> <target> [<x> <y> <z>] [<volume>] [<pitch>] [<minimumVolume>]"
}

// Parse implements exec.Handler.
func (c *playSoundCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	sound, err := p.String()
	if err != nil {
		return nil, err
	}
	source, err := p.String()
	if err != nil {
		return nil, err
	}
	target, err := parseEntityTarget(p)
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
	volume := 1.0
	if p.More() {
		v, err := p.Float()
		if err != nil {
			return nil, err
		}
		volume = v
	}
	pitch := 1.0
	if p.More() {
		v, err := p.Float()
		if err != nil {
			return nil, err
		}
		pitch = v
	}
	minVolume := 0.0
	if p.More() {
		v, err := p.Float()
		if err != nil {
			return nil, err
		}
		minVolume = v
	}
	return []ast.Arg{
		&ast.StringArg{Value: sound},
		&ast.StringArg{Value: source},
		target,
		&ast.Coord{X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ},
		&ast.FloatArg{Value: volume},
		&ast.FloatArg{Value: pitch},
		&ast.FloatArg{Value: minVolume},
	}, nil
}

// Run implements exec.Handler.
func (c *playSoundCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	sound := args[0].(*ast.StringArg).Value
	source := args[1].(*ast.StringArg).Value
	if !validSoundCategory(source) {
		return 0, exec.Errorf("Source %s is unknown", source)
	}
	// 1.12.2 plays the sound for a *player*: the target is resolved with
	// `CommandAbstract.b`, which keeps only the players it finds.
	players, err := resolvePlayerTargets(ctx, args[2])
	if err != nil {
		return 0, err
	}
	for _, p := range players {
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Played sound '%s' to %s", sound, p.PlayerName))
	}
	return len(players), nil
}

// stopSoundCmd implements `/stopsound`.
//
// The 1.12.2 syntax is `stopsound <target> [<source>] [<sound>]`, so the
// target comes first here.
type stopSoundCmd struct{}

// Name implements exec.Handler.
func (c *stopSoundCmd) Name() string { return "stopsound" }

// Usage implements exec.Handler.
func (c *stopSoundCmd) Usage() string { return "/stopsound <target> [<source>] [<sound>]" }

// Parse implements exec.Handler.
func (c *stopSoundCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	source := ""
	if p.More() {
		source, err = p.String()
		if err != nil {
			return nil, err
		}
	}
	sound := ""
	if p.More() {
		sound, err = p.String()
		if err != nil {
			return nil, err
		}
	}
	return []ast.Arg{
		target,
		&ast.StringArg{Value: source},
		&ast.StringArg{Value: sound},
	}, nil
}

// Run implements exec.Handler.
func (c *stopSoundCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	// 1.12.2 stops the sound for a *player*: the target is resolved with
	// `CommandAbstract.b`, which keeps only the players it finds.
	players, err := resolvePlayerTargets(ctx, args[0])
	if err != nil {
		return 0, err
	}
	source := args[1].(*ast.StringArg).Value
	if source != "" && !validSoundCategory(source) {
		return 0, exec.Errorf("unknown sound source %q", source)
	}
	sound := args[2].(*ast.StringArg).Value
	for _, p := range players {
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Stopped sound '%s' with source '%s' for %s", sound, source, p.PlayerName))
	}
	return len(players), nil
}

// validSoundCategory reports whether a name is a 1.12.2 sound category.
func validSoundCategory(name string) bool {
	switch name {
	case "master", "music", "record", "weather", "hostile", "neutral", "player", "ambient", "voice":
		return true
	}
	return false
}

// particleCmd implements `/particle`.
//
// The 1.12.2 syntax is `/particle <name> <x> <y> <z> <xd> <yd> <zd>
// <speed> [count] [mode] [player] [params]`: the speed is required, it
// is the eighth argument.
type particleCmd struct{}

// Name implements exec.Handler.
func (c *particleCmd) Name() string { return "particle" }

// Usage implements exec.Handler.
func (c *particleCmd) Usage() string {
	return "/particle <name> <x> <y> <z> <xd> <yd> <zd> <speed> [count] [mode] [player] [params]"
}

// Parse implements exec.Handler.
func (c *particleCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	name, err := p.String()
	if err != nil {
		return nil, err
	}
	// The position is one triple of coordinates.
	pos, err := p.Coord(ast.Coord{})
	if err != nil {
		return nil, err
	}
	coords := &ast.Coord{
		X: pos.X, Y: pos.Y, Z: pos.Z,
		RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ,
	}
	spread := make([]*ast.FloatArg, 0, 3)
	for i := 0; i < 3; i++ {
		v, err := p.Float()
		if err != nil {
			return nil, err
		}
		spread = append(spread, &ast.FloatArg{Value: v})
	}
	// The speed is the eighth argument and 1.12.2 insists on it.
	speed, err := p.Float()
	if err != nil {
		return nil, err
	}
	count := 0
	if p.More() {
		n, err := p.Int()
		if err != nil {
			return nil, err
		}
		count = n
	}
	mode := "normal"
	if p.More() {
		m, err := p.String()
		if err != nil {
			return nil, err
		}
		if strings.EqualFold(m, "force") {
			mode = "force"
		}
	}
	// The viewer is optional.
	viewer := ast.Arg(&ast.StringArg{Value: ""})
	if p.More() {
		viewer, err = parseEntityTarget(p)
		if err != nil {
			return nil, err
		}
	}
	// The particle parameters are read at run time, because how many
	// there are depends on the particle itself.
	params := &ast.ListArg{}
	for p.More() {
		t, _ := p.Next()
		params.Items = append(params.Items, &ast.StringArg{Value: t.Text})
	}
	return []ast.Arg{
		&ast.StringArg{Value: name},
		coords,
		spread[0], spread[1], spread[2],
		&ast.FloatArg{Value: speed},
		&ast.IntArg{Value: count},
		&ast.StringArg{Value: mode},
		viewer,
		params,
	}, nil
}

// Run implements exec.Handler.
func (c *particleCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	name := args[0].(*ast.StringArg).Value
	count := args[6].(*ast.IntArg).Value
	params, ok := lookupParticle(name)
	if !ok {
		return 0, exec.Errorf("Unknown effect name (%s)", name)
	}
	if count < 0 {
		return 0, exec.Errorf("The number you have entered (%d) is too small, it must be at least 0", count)
	}
	// The viewer, when there is one, has to exist.
	if viewer := particleViewer(args[8]); viewer != nil {
		if _, err := resolveEntityTargets(ctx, viewer); err != nil {
			return 0, err
		}
	}
	// Only as many parameters as the particle takes are read, and they
	// have to be numbers.
	list := args[9].(*ast.ListArg)
	for i := 0; i < params && i < len(list.Items); i++ {
		raw := list.Items[i].(*ast.StringArg).Value
		if _, err := strconv.Atoi(raw); err != nil {
			return 0, exec.Errorf("Invalid param given (%s)", raw)
		}
	}
	// 1.12.2 reports at least one effect, even when no count was given.
	times := count
	if times < 1 {
		times = 1
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Playing effect %s for %d times", name, times))
	return times, nil
}

// particleViewer returns the viewer argument of `/particle`, or nil when
// the command did not name one.
func particleViewer(arg ast.Arg) ast.Arg {
	switch v := arg.(type) {
	case *ast.StringArg:
		if v.Value == "" {
			return nil
		}
	case *ast.Selector:
		if v.Raw == "" {
			return nil
		}
	}
	return arg
}

// particles holds the 1.12.2 particle names and the number of parameters
// each of them takes.
//
// The names come from the 1.12.2 server's `EnumParticle` registry. The
// parameter count is the number of extra numbers the particle needs: the
// block id for the block particles and the block id plus the damage for
// the item particles.
var particles = map[string]int{
	"explode":          0,
	"largeexplode":     0,
	"hugeexplosion":    0,
	"fireworksSpark":   0,
	"bubble":           0,
	"splash":           0,
	"wake":             0,
	"suspended":        0,
	"depthsuspend":     0,
	"crit":             0,
	"magicCrit":        0,
	"smoke":            0,
	"largesmoke":       0,
	"spell":            0,
	"instantSpell":     0,
	"mobSpell":         0,
	"mobSpellAmbient":  0,
	"witchMagic":       0,
	"dripWater":        0,
	"dripLava":         0,
	"angryVillager":    0,
	"happyVillager":    0,
	"townaura":         0,
	"note":             0,
	"portal":           0,
	"enchantmenttable": 0,
	"flame":            0,
	"lava":             0,
	"footstep":         0,
	"cloud":            0,
	"reddust":          0,
	"snowballpoof":     0,
	"snowshovel":       0,
	"slime":            0,
	"heart":            0,
	"barrier":          0,
	"iconcrack":        2,
	"blockcrack":       1,
	"blockdust":        1,
	"droplet":          0,
	"take":             0,
	"mobappearance":    0,
	"dragonbreath":     0,
	"endRod":           0,
	"damageIndicator":  0,
	"sweepAttack":      0,
	"fallingdust":      1,
	"totem":            0,
	"spit":             0,
}

// lookupParticle returns how many parameters a particle takes.
func lookupParticle(name string) (int, bool) {
	n, ok := particles[name]
	return n, ok
}

// meCmd implements `/me`.
//
// The command shows the message as an emote, and 1.12.2 resolves a
// selector in the message the same way `/say` does.
type meCmd struct{}

// Name implements exec.Handler.
func (c *meCmd) Name() string { return "me" }

// Usage implements exec.Handler.
func (c *meCmd) Usage() string { return "/me <action>" }

// Parse implements exec.Handler.
func (c *meCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	text := p.Rest()
	if strings.TrimSpace(text) == "" {
		return nil, p.Errorf("expected a message")
	}
	return []ast.Arg{&ast.StringArg{Value: text}}, nil
}

// Run implements exec.Handler.
func (c *meCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	msg := args[0].(*ast.StringArg).Value
	// 1.12.2 only expands the selectors when the sender is not a player:
	// a player cannot make the server announce a name list.
	if ctx.Source.Entity == nil {
		expanded, err := expandMessage(ctx, msg)
		if err != nil {
			return 0, err
		}
		msg = expanded
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("* %s %s", ctx.Source.Name, msg))
	return 1, nil
}

// registerOutput adds the commands that produce chat output.
func registerOutput(reg *exec.Registry) {
	reg.Register(&titleCmd{})
	reg.Register(&playSoundCmd{})
	reg.Register(&stopSoundCmd{})
	reg.Register(&particleCmd{})
	reg.Register(&meCmd{})
}
