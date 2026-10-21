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
func (c *setWorldSpawnCmd) Usage() string { return usageLine("setworldspawn") }

// Parse implements exec.Handler.
//
// 1.12.2 counts the arguments before it looks at any of them: no argument
// puts the spawn point where the command runs, three arguments give the
// position, and one, two or more than three are a usage mistake.
func (c *setWorldSpawnCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	switch p.Count() {
	case 0:
		// The spawn point becomes the block the command runs at. The
		// position is not known until the command runs, so it is left
		// empty here.
		return []ast.Arg{nil}, nil
	case 3:
		pos, err := p.CoordWith(ast.Coord{}, coordBlockPosCentered)
		if err != nil {
			return nil, err
		}
		return []ast.Arg{&ast.Coord{X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ}}, nil
	default:
		return nil, p.Errorf("Usage: %s", usageLine("setworldspawn"))
	}
}

// Run implements exec.Handler.
func (c *setWorldSpawnCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	posArg, _ := args[0].(*ast.Coord)
	var pos world.Pos
	if posArg == nil {
		// Without a position the spawn point becomes the block the
		// command runs at. 1.12.2 takes it from the player who runs the
		// command (`CommandAbstract.a(listener)`), so a command that has
		// no player behind it is refused.
		if ctx.Source.Entity == nil || !ctx.Source.Entity.IsPlayer {
			return 0, exec.Errorf("You must specify which player you wish to perform this action on.")
		}
		pos = spawnChunkPos(ctx.Source.Entity)
	} else {
		pos = resolveRelative(*posArg, ctx.Source.Pos).BlockPos()
	}
	ctx.World.Spawn = pos
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Set the world spawn point to (%d, %d, %d)", pos.X, pos.Y, pos.Z))
	return 1, nil
}

// spawnPointCmd implements `/spawnpoint`.
//
// The 1.12.2 syntax is `spawnpoint [player] [<x> <y> <z>]`, and the target
// defaults to the player running the command.
type spawnPointCmd struct{}

// Name implements exec.Handler.
func (c *spawnPointCmd) Name() string { return "spawnpoint" }

// Usage implements exec.Handler.
func (c *spawnPointCmd) Usage() string { return usageLine("spawnpoint") }

// Parse implements exec.Handler.
//
// 1.12.2 counts the arguments before it looks at any of them:
// `CommandSpawnpoint.execute` starts with
// `if (args.length > 1 && args.length < 4)`, so two or three arguments are
// a usage mistake -- the position cannot be given without the player. The
// position is only read when there are four arguments, and without one the
// spawn point goes where the player stands.
func (c *spawnPointCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	switch n := p.Count(); {
	case n == 0:
		// The player running the command, at its own position.
		return []ast.Arg{&ast.StringArg{Value: ""}, nil}, nil
	case n == 1:
		target, err := parseSpawnTarget(p)
		if err != nil {
			return nil, err
		}
		return []ast.Arg{target, nil}, nil
	case n < 4:
		return nil, p.Errorf("Usage: %s", usageLine("spawnpoint"))
	default:
		target, err := parseSpawnTarget(p)
		if err != nil {
			return nil, err
		}
		pos, err := p.CoordWith(ast.Coord{}, coordBlockPosCentered)
		if err != nil {
			return nil, err
		}
		return []ast.Arg{
			target,
			&ast.Coord{X: pos.X, Y: pos.Y, Z: pos.Z,
				RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ},
		}, nil
	}
}

// parseSpawnTarget parses the player of a `/spawnpoint`.
//
// It is the target of a command that only works on a player, so a bare
// name and a selector are both accepted.
func parseSpawnTarget(p *parser.Parser) (ast.Arg, error) {
	t, ok := p.Peek()
	if !ok {
		return nil, p.Errorf("expected a target")
	}
	if IsSelector(t.Text) || t.Text == "" {
		return parseSelector(p)
	}
	name, err := p.String()
	if err != nil {
		return nil, err
	}
	return &ast.StringArg{Value: name}, nil
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
	// 1.12.2 sets the spawn point of a *player*: the target is resolved
	// with `CommandAbstract.b`, which ends in `PlayerSelector.getPlayer`
	// -- and that hands back **one** player, the first of the list the
	// selector matched, not all of them.
	players, err := resolvePlayerTargets(ctx, target)
	if err != nil {
		return 0, err
	}
	if len(players) == 0 {
		return 0, nil
	}
	p := players[0]
	posArg, _ := args[1].(*ast.Coord)
	var pos world.Pos
	if posArg == nil {
		// Without a position the spawn point goes where the player
		// stands.
		pos = spawnChunkPos(p)
	} else {
		pos = resolveRelative(*posArg, ctx.Source.Pos).BlockPos()
	}
	ctx.World.PlayerSpawns[p.PlayerName] = pos
	// The message names the player whose spawn point was set, not the
	// player who ran the command.
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Set %s's spawn point to (%d, %d, %d)", p.PlayerName, pos.X, pos.Y, pos.Z))
	return 1, nil
}

// spawnChunkPos returns the block position 1.12.2 calls the "chunk
// coordinates" of an entity.
//
// `EntityPlayer.getChunkCoordinates` is
// `new BlockPosition(locX, locY + 0.5, locZ)`, and a block position floors
// all three, so the Y of the player is the block it stands in and the X and
// Z are the block it is over.
func spawnChunkPos(e *world.Entity) world.Pos {
	return world.Vec{X: e.Pos.X, Y: e.Pos.Y + 0.5, Z: e.Pos.Z}.BlockPos()
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
