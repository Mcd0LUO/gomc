package v1_12_2

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/nbt"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/world"
)

// setBlockCmd implements `/setblock`.
type setBlockCmd struct{}

// Name implements exec.Handler.
func (c *setBlockCmd) Name() string { return "setblock" }

// Usage implements exec.Handler.
func (c *setBlockCmd) Usage() string { return "/setblock <x> <y> <z> <block> [<data>] [<replaceMode>]" }

// Parse implements exec.Handler.
func (c *setBlockCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	pos, err := p.Coord(ast.Coord{})
	if err != nil {
		return nil, err
	}
	block, err := p.String()
	if err != nil {
		return nil, err
	}
	ref := &ast.BlockRef{Name: block, Data: -1}
	if p.More() {
		// The next argument is either a data value or a replace mode.
		if t, ok := p.Peek(); ok && isDataValue(t.Text) {
			p.Next()
			n, _ := strconv.Atoi(t.Text)
			ref.Data = n
			ref.HasData = true
		}
	}
	mode := "replace"
	if p.More() {
		// The mode is a bare word, the NBT is a brace group.
		if t, ok := p.Peek(); ok && !strings.HasPrefix(t.Text, "{") {
			p.Next()
			mode = t.Text
		}
	}
	switch mode {
	case "replace", "destroy", "keep":
	default:
		return nil, p.Errorf("expected replace, destroy or keep, got %q", mode)
	}
	var data *nbt.Compound
	if p.More() {
		arg, err := p.NBT()
		if err != nil {
			return nil, err
		}
		data = arg.Value
	}
	return []ast.Arg{
		&ast.Coord{X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ},
		ref,
		&ast.StringArg{Value: mode},
		&ast.NBTArg{Value: data},
	}, nil
}

// isDataValue reports whether text is a numeric block data value.
func isDataValue(text string) bool {
	if text == "" {
		return false
	}
	for i := 0; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
	}
	return true
}

// Run implements exec.Handler.
func (c *setBlockCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	posArg := args[0].(*ast.Coord)
	pos := resolveRelative(*posArg, ctx.Source.Pos).BlockPos()
	ref := args[1].(*ast.BlockRef)
	mode := args[2].(*ast.StringArg).Value
	data := exec.NBTOf(args, 3)

	switch mode {
	case "destroy":
		// The old block is dropped, and placing air over air changes
		// nothing.
		ctx.World.Blocks.Set(pos, world.BlockState{Name: "air"})
		if strings.EqualFold(ref.Name, "air") {
			return 1, nil
		}
	case "keep":
		if !ctx.World.Blocks.Get(pos).IsAir() {
			return 0, exec.Errorf("The block couldn't be placed")
		}
	}

	blockData := ref.Data
	if blockData < 0 {
		blockData = 0
	}
	ctx.World.Blocks.Set(pos, world.BlockState{Name: ref.Name, Data: blockData})
	if data != nil {
		existing := ctx.World.Blocks.TileEntity(pos)
		if existing == nil {
			existing = nbt.NewCompound()
		}
		mergeNBT(existing, data)
		ctx.World.Blocks.SetTileEntity(pos, existing)
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback, "Block placed")
	return 1, nil
}

// fillCmd implements `/fill`.
type fillCmd struct{}

// Name implements exec.Handler.
func (c *fillCmd) Name() string { return "fill" }

// Usage implements exec.Handler.
func (c *fillCmd) Usage() string {
	return "/fill <x1> <y1> <z1> <x2> <y2> <z2> <block> [<data>] [<fillMode>] [<replaceBlock>] [<replaceData>]"
}

// Parse implements exec.Handler.
func (c *fillCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
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
			n, _ := strconv.Atoi(t.Text)
			ref.Data = n
			ref.HasData = true
		}
	}
	mode := "replace"
	if p.More() {
		mode, err = p.String()
		if err != nil {
			return nil, err
		}
	}
	switch mode {
	case "replace", "destroy", "keep", "hollow", "outline":
	default:
		return nil, p.Errorf("expected replace, destroy, keep, hollow or outline, got %q", mode)
	}
	replaceBlock := ""
	if p.More() {
		replaceBlock, err = p.String()
		if err != nil {
			return nil, err
		}
	}
	replaceData := -1
	if p.More() {
		replaceData, err = p.Int()
		if err != nil {
			return nil, err
		}
	}
	return []ast.Arg{
		&ast.Coord{X: from.X, Y: from.Y, Z: from.Z,
			RelativeX: from.RelativeX, RelativeY: from.RelativeY, RelativeZ: from.RelativeZ},
		&ast.Coord{X: to.X, Y: to.Y, Z: to.Z,
			RelativeX: to.RelativeX, RelativeY: to.RelativeY, RelativeZ: to.RelativeZ},
		ref,
		&ast.StringArg{Value: mode},
		&ast.StringArg{Value: replaceBlock},
		&ast.IntArg{Value: replaceData},
	}, nil
}

// Run implements exec.Handler.
func (c *fillCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	from := resolveRelative(*args[0].(*ast.Coord), ctx.Source.Pos).BlockPos()
	to := resolveRelative(*args[1].(*ast.Coord), ctx.Source.Pos).BlockPos()
	ref := args[2].(*ast.BlockRef)
	mode := args[3].(*ast.StringArg).Value
	replaceBlock := args[4].(*ast.StringArg).Value
	replaceData := args[5].(*ast.IntArg).Value

	data := ref.Data
	if data < 0 {
		data = 0
	}
	min, max := orderCorners(from, to)
	count := 0
	for x := min.X; x <= max.X; x++ {
		for y := min.Y; y <= max.Y; y++ {
			for z := min.Z; z <= max.Z; z++ {
				pos := world.Pos{X: x, Y: y, Z: z}
				existing := ctx.World.Blocks.Get(pos)
				onShell := x == min.X || x == max.X ||
					y == min.Y || y == max.Y ||
					z == min.Z || z == max.Z
				switch mode {
				case "outline":
					if !onShell {
						continue
					}
				case "hollow":
					if !onShell {
						// The inside of a hollow fill becomes air.
						ctx.World.Blocks.Set(pos, world.BlockState{Name: "air"})
						continue
					}
				case "destroy":
					ctx.World.Blocks.Set(pos, world.BlockState{Name: "air"})
				case "keep":
					if !existing.IsAir() {
						continue
					}
				case "replace":
					// The replacement filter only applies when a
					// replacement block was given.
					if replaceBlock != "" {
						if !strings.EqualFold(existing.Name, replaceBlock) {
							continue
						}
						if replaceData >= 0 && existing.Data != replaceData {
							continue
						}
					}
				}
				ctx.World.Blocks.Set(pos, world.BlockState{Name: ref.Name, Data: data})
				count++
			}
		}
	}
	if count == 0 {
		return 0, exec.Errorf("No blocks filled")
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("%d blocks filled", count))
	return count, nil
}

// orderCorners returns the two corners in ascending order.
func orderCorners(a, b world.Pos) (world.Pos, world.Pos) {
	min := world.Pos{X: minInt(a.X, b.X), Y: minInt(a.Y, b.Y), Z: minInt(a.Z, b.Z)}
	max := world.Pos{X: maxInt(a.X, b.X), Y: maxInt(a.Y, b.Y), Z: maxInt(a.Z, b.Z)}
	return min, max
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// blockDataCmd implements `/blockdata`.
type blockDataCmd struct{}

// Name implements exec.Handler.
func (c *blockDataCmd) Name() string { return "blockdata" }

// Usage implements exec.Handler.
func (c *blockDataCmd) Usage() string { return "/blockdata <x> <y> <z> <nbt>" }

// Parse implements exec.Handler.
func (c *blockDataCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	pos, err := p.Coord(ast.Coord{})
	if err != nil {
		return nil, err
	}
	data, err := p.NBT()
	if err != nil {
		return nil, err
	}
	return []ast.Arg{
		&ast.Coord{X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ},
		data,
	}, nil
}

// Run implements exec.Handler.
func (c *blockDataCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	pos := resolveRelative(*args[0].(*ast.Coord), ctx.Source.Pos).BlockPos()
	existing := ctx.World.Blocks.TileEntity(pos)
	if existing == nil {
		existing = nbt.NewCompound()
	}
	mergeNBT(existing, exec.NBTOf(args, 1))
	ctx.World.Blocks.SetTileEntity(pos, existing)
	return 1, nil
}

// testForBlockCmd implements `/testforblock`.
type testForBlockCmd struct{}

// Name implements exec.Handler.
func (c *testForBlockCmd) Name() string { return "testforblock" }

// Usage implements exec.Handler.
func (c *testForBlockCmd) Usage() string { return "/testforblock <x> <y> <z> <block> [<data>]" }

// Parse implements exec.Handler.
func (c *testForBlockCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	pos, err := p.Coord(ast.Coord{})
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
			n, _ := strconv.Atoi(t.Text)
			ref.Data = n
			ref.HasData = true
		}
	}
	return []ast.Arg{
		&ast.Coord{X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ},
		ref,
	}, nil
}

// Run implements exec.Handler.
func (c *testForBlockCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	pos := resolveRelative(*args[0].(*ast.Coord), ctx.Source.Pos).BlockPos()
	ref := args[1].(*ast.BlockRef)
	block := ctx.World.Blocks.Get(pos)
	if !strings.EqualFold(block.Name, ref.Name) {
		return 0, exec.Errorf("The block at %s is %s (expected: %s)",
			blockPos(pos), describeBlocks(block), describeBlocks(world.BlockState{Name: ref.Name, Data: ref.Data}))
	}
	if ref.HasData && block.Data != ref.Data {
		return 0, exec.Errorf("The block at %s had the data value of %d (expected: %d)",
			blockPos(pos), block.Data, ref.Data)
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Successfully found the block at %s", blockPos(pos)))
	return 1, nil
}

// registerWorld adds the world commands.
func registerWorld(reg *exec.Registry) {
	reg.Register(&setBlockCmd{})
	reg.Register(&fillCmd{})
	reg.Register(&blockDataCmd{})
	reg.Register(&testForBlockCmd{})
}

// describeBlocks renders a block for messages.
//
// 1.12.2 shows the translated name ("Stone"), not the registry id
// ("stone"), and it does not include the data value.
func describeBlocks(s world.BlockState) string {
	return world.BlockDisplayName(s.Name)
}

// blockPos renders a block position the way 1.12.2 does in its messages:
// "x, y, z".
func blockPos(p world.Pos) string {
	return fmt.Sprintf("%d, %d, %d", p.X, p.Y, p.Z)
}
