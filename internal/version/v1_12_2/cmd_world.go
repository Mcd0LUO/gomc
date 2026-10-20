package v1_12_2

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
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
//
// The 1.12.2 syntax is `/fill <x1> <y1> <z1> <x2> <y2> <z2> <block>
// [dataValue|state] [oldBlockHandling] [dataTag]`: the data value is the
// eighth argument, the mode the ninth and the tenth is the data tag of
// the block.
//
// The tenth argument is also where the block to replace goes in `replace`
// mode, and the eleventh is that block's data value. 1.12.2 tells the two
// apart by the block that is placed: a block with a tile entity takes a
// data tag, any other block takes the block to replace.
type fillCmd struct{}

// Name implements exec.Handler.
func (c *fillCmd) Name() string { return "fill" }

// Usage implements exec.Handler.
func (c *fillCmd) Usage() string {
	return "/fill <x1> <y1> <z1> <x2> <y2> <z2> <block> [dataValue|state] [oldBlockHandling] [dataTag]"
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
		// The data value is the eighth argument. 1.12.2 reads that slot
		// strictly, but the mode is accepted there as well, which is the
		// form most functions use.
		if t, ok := p.Peek(); ok && isFillDataValue(t.Text) {
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
	// The tenth argument is the block to replace, or the data tag of the
	// block that is placed. gomc has no list of blocks with a tile
	// entity, so the shape of the argument decides.
	replaceBlock := ""
	replaceData := ""
	var data *nbt.Compound
	if p.More() {
		if t, ok := p.Peek(); ok && strings.HasPrefix(t.Text, "{") {
			arg, err := p.NBT()
			if err != nil {
				return nil, p.Errorf("Data tag parsing failed: %s", nbtErrorMessage(err))
			}
			data = arg.Value
		} else {
			replaceBlock, err = p.String()
			if err != nil {
				return nil, err
			}
		}
	}
	if p.More() {
		replaceData, err = p.String()
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
		&ast.StringArg{Value: replaceData},
		&ast.NBTArg{Value: data},
	}, nil
}

// isFillDataValue reports whether text is a block data value.
//
// A data value is a number, and 1.12.2 refuses the negative ones with
// "The number you have entered (-1) is too small, it must be at least 0",
// so they are read as a data value here as well.
func isFillDataValue(text string) bool {
	if text == "" {
		return false
	}
	i := 0
	if text[0] == '-' {
		i = 1
		if len(text) == 1 {
			return false
		}
	}
	for ; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
	}
	return true
}

// nbtErrorMessage returns the message of a parse error without the
// position it carries.
func nbtErrorMessage(err error) string {
	if pe, ok := err.(*parser.Error); ok {
		return pe.Msg
	}
	return err.Error()
}

// Run implements exec.Handler.
func (c *fillCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	from := resolveRelative(*args[0].(*ast.Coord), ctx.Source.Pos).BlockPos()
	to := resolveRelative(*args[1].(*ast.Coord), ctx.Source.Pos).BlockPos()
	ref := args[2].(*ast.BlockRef)
	mode := args[3].(*ast.StringArg).Value
	replaceBlock := args[4].(*ast.StringArg).Value
	replaceData := args[5].(*ast.StringArg).Value
	data := exec.NBTOf(args, 6)

	// 1.12.2 does not check the mode: every word it does not know
	// behaves like "replace".
	switch mode {
	case "replace", "destroy", "keep", "hollow", "outline":
	default:
		mode = "replace"
	}

	blockData := ref.Data
	if ref.HasData && blockData < 0 {
		return 0, exec.Errorf("The number you have entered (%d) is too small, it must be at least 0", blockData)
	}
	if blockData < 0 {
		blockData = 0
	}
	if blockData > 15 {
		return 0, exec.Errorf("The number you have entered (%d) is too big, it must be at most 15", blockData)
	}
	// The data value of the block to replace may be "-1" or "*", which
	// mean that any data value matches.
	replaceValue := -1
	if replaceData != "" && replaceData != "-1" && replaceData != "*" {
		n, err := strconv.Atoi(replaceData)
		if err != nil {
			return 0, exec.Errorf("'%s' is not a state for block %s", replaceData, ref.Name)
		}
		replaceValue = n
	}

	min, max := orderCorners(from, to)
	// 1.12.2 refuses a fill that is too big and a fill that reaches out
	// of the world, before it changes anything.
	sizeX := max.X - min.X + 1
	sizeY := max.Y - min.Y + 1
	sizeZ := max.Z - min.Z + 1
	volume := sizeX * sizeY * sizeZ
	if volume > 32768 {
		return 0, exec.Errorf("Too many blocks in the specified area (%d > 32768)", volume)
	}
	if min.Y < 0 || max.Y >= 256 {
		return 0, exec.Errorf("Cannot place blocks outside of the world")
	}
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
						if replaceValue >= 0 && existing.Data != replaceValue {
							continue
						}
					}
				}
				ctx.World.Blocks.Set(pos, world.BlockState{Name: ref.Name, Data: blockData})
				if data != nil {
					tag := data.Clone()
					tag.Set("x", nbt.Int(int32(x)))
					tag.Set("y", nbt.Int(int32(y)))
					tag.Set("z", nbt.Int(int32(z)))
					ctx.World.Blocks.SetTileEntity(pos, tag)
				}
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
