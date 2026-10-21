package v1_12_2

import (
	"fmt"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
)

// gomcPlayerCmd implements `/gomcplayer`.
//
// This is **not** a 1.12.2 command. 1.12.2 cannot bring a player into the
// world -- `/summon player` answers "Unable to summon object" -- and a
// function that needs a player to test against has no way to get one.
// This command makes one.
//
// It exists so that the commands that only work on a player can be
// exercised end to end: `/advancement` and `/recipe` need an online player,
// and without this command the only thing that can be tested of them is
// their error messages.
//
// It is registered as an extension command, so `gomc check` says that a
// function using it would not load on a real server -- 1.12.2 refuses a
// function whose command it does not know.
type gomcPlayerCmd struct{}

// Name implements exec.Handler.
func (c *gomcPlayerCmd) Name() string { return "gomcplayer" }

// Usage implements exec.Handler.
func (c *gomcPlayerCmd) Usage() string { return usageLine("gomcplayer") }

// Parse implements exec.Handler.
//
// The name is required, the position is not: without one the player is put
// where the command runs.
func (c *gomcPlayerCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	name, err := p.String()
	if err != nil {
		return nil, usageError(p, "gomcplayer", "expected a player name")
	}
	args := []ast.Arg{&ast.StringArg{Value: name}}
	if p.More() {
		pos, err := p.CoordWith(coordBlockPos)
		if err != nil {
			return nil, err
		}
		args = append(args, &ast.Coord{X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ})
	}
	return args, nil
}

// Run implements exec.Handler.
//
// The player is an ordinary entity of the world: it is found by its name,
// it is picked up by `@a` and by a selector, and the scoreboard tracks it
// under its name, exactly like a player that logged in.
func (c *gomcPlayerCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	name := args[0].(*ast.StringArg).Value
	if name == "" {
		return 0, exec.Errorf("A player needs a name")
	}
	if ctx.World.FindPlayer(name) != nil {
		return 0, exec.Errorf("A player called '%s' already exists", name)
	}

	// The position is the one of the command unless it is given.
	pos := ctx.Source.Pos
	if len(args) > 1 {
		pos = resolveRelative(*args[1].(*ast.Coord), ctx.Source.RelativeBase())
	}

	player, ok := ctx.World.EntityTypes.NewEntity("player", "")
	if !ok {
		return 0, exec.Errorf("The player entity type is missing")
	}
	player.PlayerName = name
	player.Pos = pos
	player.Rot = ctx.Source.Rot
	player.World = ctx.Source.World
	ctx.World.AddEntity(player)

	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Successfully created the player %s", name))
	return 1, nil
}

// registerGomcPlayer adds the command that brings a player into the world.
//
// The command is not part of the 1.12.2 command set, so it is marked as an
// extension: `gomc check` then says that a function using it would not
// load on a real server.
func registerGomcPlayer(reg *exec.Registry) {
	reg.Register(&gomcPlayerCmd{})
	// The level of `/summon`: the command creates an entity, so it is a
	// thing an operator may do.
	reg.SetRequiredPermission("gomcplayer", 2)
	reg.SetExtension("gomcplayer",
		"it brings a player into the world, which 1.12.2 itself cannot do")
}
