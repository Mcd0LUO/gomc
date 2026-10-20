package v1_12_2

import (
	"strings"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
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
// How 1.12.2 expands a multi-entity target is not obvious from the
// command itself, and getting it wrong is easy:
//
//   - The expansion happens in the *command dispatcher* (`CommandHandler`),
//     not inside `execute`. The dispatcher finds the first argument that
//     is a target list, resolves it, and then runs the whole command once
//     per matching entity, with that entity substituted for the
//     argument.
//   - The entities are visited in the order the selector returned them.
//     `PlayerSelector` sorts `@a`, `@e` and `@p` by distance (a stable
//     sort, so equally distant entities keep their world order), so the
//     nearest entity is visited first. There is no reversal.
//   - A failure belongs to the entity that caused it. The dispatcher
//     catches it, reports it, and carries on with the next entity.
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
		// The detect coordinates are relative to the *execution* position,
		// which is only known at run time (it depends on the target
		// entity). So they are parsed against zero here and resolved
		// against the execution position later.
		//
		// Parsing them against the execute coordinates would apply the
		// execution offset twice.
		dp, err := p.Coord(ast.Coord{})
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
		// The data value is required in 1.12.2: the detect block is five
		// arguments, the last one the data value. A value of -1 means
		// "any data value".
		n, err := p.Int()
		if err != nil {
			return nil, err
		}
		a.detectData = n
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
	// The entities are visited in the order the selector returned them,
	// which is the order the dispatcher would use.
	//
	// The target entity becomes the *executor* of the sub-command, so the
	// sub-command's output is prefixed with the entity's name, and `@s`
	// inside it refers to that entity.
	total := 0
	for _, e := range entities {
		pos := resolveRelative(a.Value.pos, e.Pos)
		if a.Value.detect {
			ok, err := checkDetect(ctx.World, a.Value, pos)
			if err != nil {
				return 0, err
			}
			if !ok {
				// 1.12.2 reports the entity it could not run the
				// command as, and then moves on to the next entity.
				ctx.Engine.Errors = append(ctx.Engine.Errors,
					exec.Errorf("Failed to execute 'detect' as %s", e.DisplayName()))
				continue
			}
		}
		sub := ctx.Source.WithEntity(e).WithPos(pos)
		sub.Name = e.DisplayName()
		count, err := ctx.Engine.ExecuteLine(a.Value.sub, ctx.Command.Line, sub)
		if err != nil {
			// The failure belongs to this entity only; the remaining
			// entities still run.
			ctx.Engine.Errors = append(ctx.Engine.Errors, err)
			continue
		}
		if count < 1 {
			// 1.12.2 treats a sub-command that did nothing as a
			// failure of the whole invocation.
			ctx.Engine.Errors = append(ctx.Engine.Errors,
				exec.Errorf("All invocations failed: '%s'", a.Value.sub))
			continue
		}
		total += count
	}
	return total, nil
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
