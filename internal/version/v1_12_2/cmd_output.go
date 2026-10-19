package v1_12_2

import (
	"fmt"
	"strings"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/nbt"
	"github.com/mcsim/gomc/internal/parser"
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
	holders, err := resolveHolders(ctx, args[0])
	if err != nil {
		return 0, err
	}
	action := args[1].(*ast.StringArg).Value

	for _, h := range holders {
		e := ctx.World.FindEntity(h)
		if e == nil {
			continue
		}
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
	return len(holders), nil
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
	holders, err := resolveHolders(ctx, args[2])
	if err != nil {
		return 0, err
	}
	for _, h := range holders {
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Played sound '%s' to %s", sound, h))
	}
	return len(holders), nil
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
	holders, err := resolveHolders(ctx, args[0])
	if err != nil {
		return 0, err
	}
	source := args[1].(*ast.StringArg).Value
	if source != "" && !validSoundCategory(source) {
		return 0, exec.Errorf("unknown sound source %q", source)
	}
	sound := args[2].(*ast.StringArg).Value
	for _, h := range holders {
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Stopped sound '%s' with source '%s' for %s", sound, source, h))
	}
	return len(holders), nil
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
// The 1.12.2 syntax is `particle <name> <x> <y> <z> <dx> <dy> <dz>
// [<speed>] [<count>] [<force|normal>] [<viewers>]`.
type particleCmd struct{}

// Name implements exec.Handler.
func (c *particleCmd) Name() string { return "particle" }

// Usage implements exec.Handler.
func (c *particleCmd) Usage() string {
	return "/particle <name> <x> <y> <z> <dx> <dy> <dz> [<speed>] [<count>] [<force|normal>] [<viewers>]"
}

// Parse implements exec.Handler.
func (c *particleCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	name, err := p.String()
	if err != nil {
		return nil, err
	}
	coords := make([]*ast.Coord, 0, 3)
	for i := 0; i < 3; i++ {
		pos, err := p.Coord(ast.Coord{})
		if err != nil {
			return nil, err
		}
		coords = append(coords, &ast.Coord{
			X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ,
		})
	}
	spread := make([]*ast.FloatArg, 0, 3)
	for i := 0; i < 3; i++ {
		v, err := p.Float()
		if err != nil {
			return nil, err
		}
		spread = append(spread, &ast.FloatArg{Value: v})
	}
	speed := 0.0
	if p.More() {
		v, err := p.Float()
		if err != nil {
			return nil, err
		}
		speed = v
	}
	count := 1
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
		mode = strings.ToLower(m)
	}
	switch mode {
	case "normal", "force":
	default:
		return nil, p.Errorf("expected normal or force, got %q", mode)
	}
	return []ast.Arg{
		&ast.StringArg{Value: name},
		coords[0], coords[1], coords[2],
		spread[0], spread[1], spread[2],
		&ast.FloatArg{Value: speed},
		&ast.IntArg{Value: count},
		&ast.StringArg{Value: mode},
	}, nil
}

// Run implements exec.Handler.
func (c *particleCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	name := args[0].(*ast.StringArg).Value
	count := args[9].(*ast.IntArg).Value
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Playing effect %s for %d times", name, count))
	return count, nil
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
