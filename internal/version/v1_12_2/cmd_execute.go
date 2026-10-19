package v1_12_2

import (
	"strings"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/world"
)

// executeArgs is the parsed form of a 1.12.2 `/execute`.
//
// It is produced during parsing and reused at run time, so the command is
// never parsed twice.
type executeArgs struct {
	// target is the entity the command runs as.
	target ast.Arg
	// pos is the position the command runs at, relative to the target.
	pos ast.Coord
	// detect enables the block precondition.
	detect bool
	// detectPos is the position checked by the precondition.
	detectPos ast.Coord
	// detectBlock is the expected block name.
	detectBlock string
	// detectData is the expected data value, or -1 for any.
	detectData int
	// sub is the command to run.
	sub string
}

// executeArg is the AST wrapper carrying the parsed execute arguments.
type executeArg struct {
	// Value holds the parsed arguments.
	Value *executeArgs
}

// ArgKind implements ast.Arg.
func (a *executeArg) ArgKind() ast.Kind { return ast.KindList }

// executeCmd implements the 1.12.2 `/execute`.
//
// The syntax is the pre-1.13 one:
//
//	execute <entity> <x> <y> <z> [detect <x2> <y2> <z2> <block> <dataValue>] <command>
//
// Two behaviours of 1.12.2 are preserved deliberately:
//
//   - When the target selector matches several entities, the command runs
//     once per entity in reverse order (MC-121728, fixed in 1.13).
//   - If the command fails for one entity, the remaining entities are
//     skipped (MC-121727, fixed in 1.13).
type executeCmd struct{}

// Name implements exec.Handler.
func (c *executeCmd) Name() string { return "execute" }

// Usage implements exec.Handler.
func (c *executeCmd) Usage() string {
	return "/execute <entity> <x> <y> <z> [detect <x2> <y2> <z2> <block> <dataValue>] <command>"
}

// Parse implements exec.Handler.
func (c *executeCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	target, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	// The coordinates are relative to the target entity, not the source.
	pos, err := p.Coord(ast.Coord{})
	if err != nil {
		return nil, err
	}
	a := &executeArgs{target: target, pos: *pos, detectData: -1}
	if p.OptionalLiteral("detect") {
		dp, err := p.Coord(*pos)
		if err != nil {
			return nil, err
		}
		a.detect = true
		a.detectPos = *dp
		block, err := p.String()
		if err != nil {
			return nil, err
		}
		a.detectBlock = block
		// The data value is optional; -1 means "any data value".
		if p.More() {
			if n, err := p.Int(); err == nil {
				a.detectData = n
			}
		}
	}
	sub := p.Rest()
	if strings.TrimSpace(sub) == "" {
		return nil, p.Errorf("expected a command to execute")
	}
	a.sub = sub
	return []ast.Arg{&executeArg{Value: a}}, nil
}

// Run implements exec.Handler.
func (c *executeCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	a, ok := args[0].(*executeArg)
	if !ok {
		return 0, exec.Errorf("invalid execute arguments")
	}
	entities, err := resolveEntityTargets(ctx, a.Value.target)
	if err != nil {
		return 0, err
	}
	if len(entities) == 0 {
		return 0, exec.Errorf("no matching entities")
	}
	// 1.12.2 expands a multi-entity selector into one command per entity,
	// and those commands run in reverse order.
	for i := len(entities) - 1; i >= 0; i-- {
		e := entities[i]
		pos := resolveRelative(a.Value.pos, e.Pos)
		if a.Value.detect {
			ok, err := checkDetect(ctx.World, a.Value, pos)
			if err != nil {
				return 0, err
			}
			if !ok {
				return 0, exec.Errorf("the detect precondition was not met")
			}
		}
		sub := ctx.Source.WithEntity(e).WithPos(pos)
		count, err := ctx.Engine.ExecuteLine(a.Value.sub, ctx.Command.Line, sub)
		if err != nil {
			// MC-121727: a failure cancels the remaining entities.
			return 0, err
		}
		if count > 0 {
			return count, nil
		}
	}
	return 0, nil
}

// resolveRelative applies relative coordinates to a base position.
func resolveRelative(c ast.Coord, base world.Vec) world.Vec {
	out := world.Vec{X: c.X, Y: c.Y, Z: c.Z}
	if c.RelativeX {
		out.X = base.X + c.X
	}
	if c.RelativeY {
		out.Y = base.Y + c.Y
	}
	if c.RelativeZ {
		out.Z = base.Z + c.Z
	}
	return out
}

// checkDetect evaluates the `detect` precondition.
func checkDetect(w *world.World, a *executeArgs, pos world.Vec) (bool, error) {
	dp := resolveRelative(a.detectPos, pos)
	block := w.Blocks.Get(dp.BlockPos())
	if !strings.EqualFold(block.Name, a.detectBlock) {
		return false, nil
	}
	if a.detectData >= 0 && block.Data != a.detectData {
		return false, nil
	}
	return true, nil
}

// registerCore adds the core commands.
func registerCore(reg *exec.Registry) {
	reg.Register(&executeCmd{})
	reg.Register(&sayCmd{})
	reg.Register(&tellCmd{})
	reg.Register(&tellRawCmd{})
	reg.Register(&gameruleCmd{})
}

// registerTime adds the time commands.
func registerTime(reg *exec.Registry) {
	reg.Register(&timeCmd{})
}
