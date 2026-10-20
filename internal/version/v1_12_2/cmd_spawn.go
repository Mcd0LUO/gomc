package v1_12_2

import (
	"fmt"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// setWorldSpawnCmd implements `/setworldspawn`.
//
// The 1.12.2 syntax is `setworldspawn [<x> <y> <z>]`. Without coordinates
// the spawn point becomes the block the command runs at.
type setWorldSpawnCmd struct{}

// Name implements exec.Handler.
func (c *setWorldSpawnCmd) Name() string { return "setworldspawn" }

// Usage implements exec.Handler.
func (c *setWorldSpawnCmd) Usage() string { return "/setworldspawn [<x> <y> <z>]" }

// Parse implements exec.Handler.
func (c *setWorldSpawnCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	if !p.More() {
		return []ast.Arg{&ast.Coord{}}, nil
	}
	pos, err := p.Coord(ast.Coord{})
	if err != nil {
		return nil, err
	}
	return []ast.Arg{&ast.Coord{X: pos.X, Y: pos.Y, Z: pos.Z,
		RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ}}, nil
}

// Run implements exec.Handler.
func (c *setWorldSpawnCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	pos := resolveRelative(*args[0].(*ast.Coord), ctx.Source.Pos).BlockPos()
	ctx.World.Spawn = pos
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Set the world spawn point to (%d, %d, %d)", pos.X, pos.Y, pos.Z))
	return 1, nil
}

// spawnPointCmd implements `/spawnpoint`.
//
// The 1.12.2 syntax is `spawnpoint [<target>] [<x> <y> <z>]`, and the target
// defaults to the player running the command.
type spawnPointCmd struct{}

// Name implements exec.Handler.
func (c *spawnPointCmd) Name() string { return "spawnpoint" }

// Usage implements exec.Handler.
func (c *spawnPointCmd) Usage() string { return "/spawnpoint [<target>] [<x> <y> <z>]" }

// Parse implements exec.Handler.
func (c *spawnPointCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target := ast.Arg(&ast.StringArg{Value: ""})
	if p.More() {
		t, ok := p.Peek()
		if !ok {
			return nil, p.Errorf("expected a target")
		}
		if parser.IsSelector(t.Text) || t.Text == "" {
			sel, err := p.Selector()
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
	pos := ast.Coord{}
	if p.More() {
		parsed, err := p.Coord(ast.Coord{})
		if err != nil {
			return nil, err
		}
		pos = *parsed
	}
	return []ast.Arg{
		target,
		&ast.Coord{X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ},
	}, nil
}

// Run implements exec.Handler.
func (c *spawnPointCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	target := args[0]
	if s, ok := target.(*ast.StringArg); ok && s.Value == "" {
		// Without a target 1.12.2 sets the spawn point of the player
		// running the command.
		if ctx.Source.Entity == nil || !ctx.Source.Entity.IsPlayer {
			return 0, exec.Errorf("You must specify which player you wish to perform this action on.")
		}
		target = &ast.StringArg{Value: ctx.Source.Entity.PlayerName}
	}
	holders, err := resolveHolders(ctx, target)
	if err != nil {
		return 0, err
	}
	posArg := args[1].(*ast.Coord)
	pos := resolveRelative(*posArg, ctx.Source.Pos).BlockPos()
	for _, h := range holders {
		// The target of the command is a player, and 1.12.2 refuses a
		// name that is not online.
		if ctx.World.FindPlayer(h) == nil && ctx.World.FindEntity(h) == nil {
			return 0, exec.Errorf("Player '%s' cannot be found", h)
		}
		ctx.World.PlayerSpawns[h] = pos
		// The message names the player whose spawn point was set, not
		// the player who ran the command.
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Set %s's spawn point to (%d, %d, %d)", h, pos.X, pos.Y, pos.Z))
	}
	return len(holders), nil
}

// registerSpawn adds the spawn point commands.
func registerSpawn(reg *exec.Registry) {
	reg.Register(&setWorldSpawnCmd{})
	reg.Register(&spawnPointCmd{})
}

// spawnPointOf returns the spawn point of a holder, falling back to the
// world spawn.
func spawnPointOf(w *world.World, holder string) world.Pos {
	if p, ok := w.PlayerSpawns[holder]; ok {
		return p
	}
	return w.Spawn
}
