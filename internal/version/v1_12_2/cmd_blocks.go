package v1_12_2

import (
	"fmt"
	"strings"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/world"
)

// cloneCmd implements `/clone`.
//
// The 1.12.2 syntax is `clone <x1> <y1> <z1> <x2> <y2> <z2> <x3> <y3> <z3>
// [<maskMode> [<cloneMode>]]`, where the mask mode is `normal`, `masked` or
// `filtered` and the clone mode is `normal`, `force` or `move`.
//
// The source region is the box between the first two corners, and the
// destination starts at the third corner, so the two regions have to be the
// same size.
type cloneCmd struct{}

// Name implements exec.Handler.
func (c *cloneCmd) Name() string { return "clone" }

// Usage implements exec.Handler.
func (c *cloneCmd) Usage() string {
	return "/clone <x1> <y1> <z1> <x2> <y2> <z2> <x3> <y3> <z3> [<maskMode> [<cloneMode>]]"
}

// Parse implements exec.Handler.
func (c *cloneCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	corners := make([]*ast.Coord, 0, 3)
	for i := 0; i < 3; i++ {
		pos, err := p.Coord(ast.Coord{})
		if err != nil {
			return nil, err
		}
		corners = append(corners, &ast.Coord{
			X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ,
		})
	}
	mask := "normal"
	clone := "normal"
	if p.More() {
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		mask = strings.ToLower(name)
	}
	if p.More() {
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		clone = strings.ToLower(name)
	}
	switch mask {
	case "normal", "masked", "filtered":
	default:
		return nil, p.Errorf("expected normal, masked or filtered, got %q", mask)
	}
	switch clone {
	case "normal", "force", "move":
	default:
		return nil, p.Errorf("expected normal, force or move, got %q", clone)
	}
	return []ast.Arg{
		corners[0], corners[1], corners[2],
		&ast.StringArg{Value: mask},
		&ast.StringArg{Value: clone},
	}, nil
}

// Run implements exec.Handler.
func (c *cloneCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	srcMin, srcMax := orderCorners(
		resolveRelative(*args[0].(*ast.Coord), ctx.Source.Pos).BlockPos(),
		resolveRelative(*args[1].(*ast.Coord), ctx.Source.Pos).BlockPos())
	src := struct{ Min, Max world.Pos }{Min: srcMin, Max: srcMax}
	dst := resolveRelative(*args[2].(*ast.Coord), ctx.Source.Pos).BlockPos()
	mask := args[3].(*ast.StringArg).Value
	clone := args[4].(*ast.StringArg).Value

	sizeX := src.Max.X - src.Min.X + 1
	sizeY := src.Max.Y - src.Min.Y + 1
	sizeZ := src.Max.Z - src.Min.Z + 1
	if sizeX*sizeY*sizeZ > 32768 {
		return 0, exec.Errorf("the clone is too big: %d blocks, the limit is 32768", sizeX*sizeY*sizeZ)
	}

	copied := 0
	for x := 0; x < sizeX; x++ {
		for y := 0; y < sizeY; y++ {
			for z := 0; z < sizeZ; z++ {
				from := world.Pos{X: src.Min.X + x, Y: src.Min.Y + y, Z: src.Min.Z + z}
				to := world.Pos{X: dst.X + x, Y: dst.Y + y, Z: dst.Z + z}
				block := ctx.World.Blocks.Get(from)
				if mask != "normal" && block.IsAir() {
					continue
				}
				ctx.World.Blocks.Set(to, block)
				copied++
			}
		}
	}
	if clone == "move" {
		for x := 0; x < sizeX; x++ {
			for y := 0; y < sizeY; y++ {
				for z := 0; z < sizeZ; z++ {
					from := world.Pos{X: src.Min.X + x, Y: src.Min.Y + y, Z: src.Min.Z + z}
					ctx.World.Blocks.Set(from, world.BlockState{Name: "air"})
				}
			}
		}
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Cloned %d block(s)", copied))
	return copied, nil
}

// testForBlocksCmd implements `/testforblocks`.
//
// The 1.12.2 syntax is `testforblocks <x1> <y1> <z1> <x2> <y2> <z2>
// <block> [<data>]`, and the whole region has to consist of the given
// block.
type testForBlocksCmd struct{}

// Name implements exec.Handler.
func (c *testForBlocksCmd) Name() string { return "testforblocks" }

// Usage implements exec.Handler.
func (c *testForBlocksCmd) Usage() string {
	return "/testforblocks <x1> <y1> <z1> <x2> <y2> <z2> <block> [<data>]"
}

// Parse implements exec.Handler.
func (c *testForBlocksCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	from, err := p.Coord(ast.Coord{})
	if err != nil {
		return nil, err
	}
	to, err := p.Coord(ast.Coord{})
	if err != nil {
		return nil, err
	}
	block, err := p.String()
	if err != nil {
		return nil, err
	}
	ref := &ast.BlockRef{Name: block, Data: -1}
	if p.More() {
		if t, ok := p.Peek(); ok && isDataValue(t.Text) {
			p.Next()
			n, _ := parseIntSafe(t.Text)
			ref.Data = n
			ref.HasData = true
		}
	}
	return []ast.Arg{
		&ast.Coord{X: from.X, Y: from.Y, Z: from.Z,
			RelativeX: from.RelativeX, RelativeY: from.RelativeY, RelativeZ: from.RelativeZ},
		&ast.Coord{X: to.X, Y: to.Y, Z: to.Z,
			RelativeX: to.RelativeX, RelativeY: to.RelativeY, RelativeZ: to.RelativeZ},
		ref,
	}, nil
}

// Run implements exec.Handler.
func (c *testForBlocksCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	min, max := orderCorners(
		resolveRelative(*args[0].(*ast.Coord), ctx.Source.Pos).BlockPos(),
		resolveRelative(*args[1].(*ast.Coord), ctx.Source.Pos).BlockPos())
	ref := args[2].(*ast.BlockRef)

	count := 0
	for x := min.X; x <= max.X; x++ {
		for y := min.Y; y <= max.Y; y++ {
			for z := min.Z; z <= max.Z; z++ {
				pos := world.Pos{X: x, Y: y, Z: z}
				block := ctx.World.Blocks.Get(pos)
				if !strings.EqualFold(block.Name, ref.Name) {
					return 0, exec.Errorf("the block at %s is %s, not %s", pos, block.Name, ref.Name)
				}
				if ref.HasData && block.Data != ref.Data {
					return 0, exec.Errorf("the block at %s has data value %d, not %d",
						pos, block.Data, ref.Data)
				}
				count++
			}
		}
	}
	return count, nil
}

// registerBlocksExtra adds the block region commands.
func registerBlocksExtra(reg *exec.Registry) {
	reg.Register(&cloneCmd{})
	reg.Register(&testForBlocksCmd{})
}
