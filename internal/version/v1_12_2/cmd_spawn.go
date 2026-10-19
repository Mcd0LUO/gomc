package v1_12_2

import (
	"fmt"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/world"
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
	holders, err := resolveHolders(ctx, args[0])
	if err != nil {
		return 0, err
	}
	if len(holders) == 0 {
		return 0, exec.Errorf("no matching entities")
	}
	posArg := args[1].(*ast.Coord)
	pos := resolveRelative(*posArg, ctx.Source.Pos).BlockPos()
	for _, h := range holders {
		ctx.World.PlayerSpawns[h] = pos
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Set the spawn point to (%d, %d, %d)", pos.X, pos.Y, pos.Z))
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
