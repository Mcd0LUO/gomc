package v1_12_2

import (
	"fmt"
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
func (c *setBlockCmd) Usage() string { return usageLine("setblock") }

// Parse implements exec.Handler.
func (c *setBlockCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	pos, err := p.CoordWith(coordBlockPos)
	if err != nil {
		return nil, err
	}
	block, err := p.String()
	if err != nil {
		return nil, err
	}
	if err := checkBlockName(p, block); err != nil {
		return nil, err
	}
	ref := &ast.BlockRef{Name: block, Data: -1}
	// The fifth argument is the data value. 1.12.2 reads it when the
	// command runs and reads it **strictly**: the slot only takes a
	// number, a state of the block or the word "default", so a word
	// like "keep" is a broken data value there and not a mode.
	//
	// (Measured on a real 1.12.2 server: `setblock 2 4 2 stone keep` answers
	// "'keep' is not a state for block minecraft:stone", while
	// `setblock 4 4 4 stone 0 keep` places the block.)
	registryName := fillBlockName(block)
	dataValue := absentFillDataValue(registryName)
	if p.More() {
		t, _ := p.Next()
		dataValue = parseFillDataValue(registryName, t.Text)
	}
	// The sixth argument is the oldBlockHandling. 1.12.2 reads it by its
	// **position**, and only "destroy" and "keep" have a meaning: any
	// other word behaves like "replace" and is not a mistake.
	//
	// (Measured on a real 1.12.2 server: `setblock 4 4 4 stone 0 0` places
	// the block.)
	mode := "replace"
	if p.More() {
		t, _ := p.Next()
		mode = t.Text
	}
	// The seventh argument is the dataTag: the rest of the line.
	text := ""
	if p.More() {
		// The dataTag is parsed when the command runs: 1.12.2 places the
		// block first and only then looks at the dataTag.
		text, err = dataTagText(p)
		if err != nil {
			return nil, err
		}
	}
	return []ast.Arg{
		&ast.Coord{X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ},
		ref,
		dataValue,
		&ast.StringArg{Value: mode},
		&ast.StringArg{Value: text},
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
	pos := resolveRelative(*posArg, ctx.Source.RelativeBase()).BlockPos()
	ref := args[1].(*ast.BlockRef)
	mode := args[3].(*ast.StringArg).Value
	var data *nbt.Compound
	if text := args[4].(*ast.StringArg).Value; text != "" {
		parsed, err := parseDataTagArg(text)
		if err != nil {
			return 0, err
		}
		data = parsed
	}

	// 1.12.2 reads the data value before it places the block, and before it
	// looks at the dataTag, so a broken data value is reported first.
	dataValue := args[2].(*fillDataValue)
	if err := dataValue.err(fillBlockName(ref.Name)); err != nil {
		return 0, err
	}

	// 1.12.2 refuses a position whose chunk is not loaded.
	//
	// (Measured on a real 1.12.2 server: `setblock 144 4 0 stone` answers
	// "Cannot place block outside of the world".)
	if !ctx.World.IsLoaded(pos) {
		return 0, exec.Errorf("Cannot place block outside of the world")
	}

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

	// 1.12.2 puts the block with `World.setTypeAndData`, which refuses a
	// position outside the height of the world: the command then says the
	// block couldn't be placed.
	//
	// (Measured on a real 1.12.2 server: `setblock 0 256 0 stone` answers
	// "The block couldn't be placed".)
	if pos.Y < 0 || pos.Y >= 256 {
		return 0, exec.Errorf("The block couldn't be placed")
	}

	blockData := dataValue.data()
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
// The eighth argument is read strictly: as soon as there is an argument
// after the block, that argument is the data value. It is not a mode, so
// `fill ... stone replace` does not fill in `replace` mode, it fails on
// the data value.
//
// The tenth argument is also where the block to replace goes in `replace`
// mode, and the eleventh is that block's data value. 1.12.2 tells the two
// apart by the block that is placed: a block with a tile entity takes a
// data tag, any other block takes the block to replace.
type fillCmd struct{}

// Name implements exec.Handler.
func (c *fillCmd) Name() string { return "fill" }

// Usage implements exec.Handler.
func (c *fillCmd) Usage() string { return usageLine("fill") }

// Parse implements exec.Handler.
func (c *fillCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	from, err := p.CoordWith(coordBlockPos)
	if err != nil {
		return nil, err
	}
	to, err := p.CoordWith(coordBlockPos)
	if err != nil {
		return nil, err
	}
	block, err := p.String()
	if err != nil {
		return nil, err
	}
	if err := checkBlockName(p, block); err != nil {
		return nil, err
	}
	ref := &ast.BlockRef{Name: block, Data: -1}
	// The eighth argument is the data value of the block. 1.12.2 reads it
	// when the command runs and reads it strictly: the slot only takes a
	// number, a state of the block or the word "default", so a word like
	// "replace" is a broken data value there and not a mode. The slot is
	// classified here and the failure is reported when the command runs.
	registryName := fillBlockName(block)
	dataValue := absentFillDataValue(registryName)
	if p.More() {
		t, _ := p.Next()
		dataValue = parseFillDataValue(registryName, t.Text)
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
	text := ""
	if p.More() {
		if t, ok := p.Peek(); ok && strings.HasPrefix(t.Text, "{") {
			// The dataTag is parsed when the command runs: 1.12.2 fills
			// the region first and only then looks at the dataTag.
			text, err = dataTagText(p)
			if err != nil {
				return nil, err
			}
		} else {
			replaceBlock, err = p.String()
			if err != nil {
				return nil, err
			}
			if err := checkBlockName(p, replaceBlock); err != nil {
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
		dataValue,
		&ast.StringArg{Value: mode},
		&ast.StringArg{Value: replaceBlock},
		&ast.StringArg{Value: replaceData},
		&ast.StringArg{Value: text},
	}, nil
}

// regionSizes returns the number of blocks of a region along its three
// axes.
//
// The difference of two coordinates can be bigger than an int, so it is
// computed with the unsigned arithmetic and cut off.
func regionSizes(min, max world.Pos) (int, int, int) {
	return int(regionSize(min.X, max.X)),
		int(regionSize(min.Y, max.Y)),
		int(regionSize(min.Z, max.Z))
}

// regionSize returns the number of blocks of a region along one axis.
func regionSize(a, b int) int64 {
	if a > b {
		a, b = b, a
	}
	// The difference is computed with the unsigned arithmetic, because
	// the signed one would overflow.
	d := uint64(b) - uint64(a)
	if d >= uint64(1)<<31 {
		return int64(1) << 31
	}
	return int64(d) + 1
}

// regionVolume returns the number of blocks in a region.
//
// 1.12.2 counts the blocks in a 32 bit int, so its number is only right
// while it fits in an int. gomc's coordinates have no such limit, so the
// runtime counts without overflowing and cuts the result off at the
// biggest number 1.12.2 could have printed.
func regionVolume(min, max world.Pos) int64 {
	const maxCount = int64(1) << 31
	sx := regionSize(min.X, max.X)
	sy := regionSize(min.Y, max.Y)
	sz := regionSize(min.Z, max.Z)
	if sx == 0 || sy == 0 || sz == 0 {
		return 0
	}
	if sy > maxCount/sx {
		return maxCount
	}
	p := sx * sy
	if sz > maxCount/p {
		return maxCount
	}
	return p * sz
}

// regionLoaded reports whether every chunk of the region is loaded.
//
// 1.12.2 walks the region in steps of 16 blocks and checks the chunk of
// every step, so the chunks of the whole region have to be loaded.
func regionLoaded(w *world.World, min, max world.Pos) bool {
	for x := min.X; x <= max.X; x += 16 {
		for z := min.Z; z <= max.Z; z += 16 {
			if !w.IsLoaded(world.Pos{X: x, Y: min.Y, Z: z}) {
				return false
			}
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
	from := resolveRelative(*args[0].(*ast.Coord), ctx.Source.RelativeBase()).BlockPos()
	to := resolveRelative(*args[1].(*ast.Coord), ctx.Source.RelativeBase()).BlockPos()
	ref := args[2].(*ast.BlockRef)
	dataValue := args[3].(*fillDataValue)
	mode := args[4].(*ast.StringArg).Value
	replaceBlock := args[5].(*ast.StringArg).Value
	replaceData := args[6].(*ast.StringArg).Value
	// 1.12.2 reads the data value before it fills the region, and before
	// it looks at the dataTag, so a broken data value is reported first.
	if err := dataValue.err(fillBlockName(ref.Name)); err != nil {
		return 0, err
	}
	blockData := dataValue.data()
	var data *nbt.Compound
	if text := args[7].(*ast.StringArg).Value; text != "" {
		parsed, err := parseDataTagArg(text)
		if err != nil {
			return 0, err
		}
		data = parsed
	}

	// 1.12.2 does not check the mode: every word it does not know
	// behaves like "replace".
	switch mode {
	case "replace", "destroy", "keep", "hollow", "outline":
	default:
		mode = "replace"
	}

	// The data value of the block to replace may be "-1" or "*", which
	// mean that any data value matches.
	//
	// 1.12.2 reads it **inside** the loop, and only for a block that is
	// the block to replace: so a data value the block does not know is
	// only reported when the region holds such a block.
	replaceValue := 0
	hasReplaceValue := false
	replaceValueRead := false

	min, max := orderCorners(from, to)
	// 1.12.2 refuses a fill that is too big and a fill that reaches out
	// of the world, before it changes anything.
	volume := regionVolume(min, max)
	if volume > 32768 {
		return 0, exec.Errorf("Too many blocks in the specified area (%d > 32768)", volume)
	}
	if min.Y < 0 || max.Y >= 256 {
		// 1.12.2 puts the blocks with `World.setTypeAndData`, which
		// refuses a position outside the height of the world. The blocks
		// that cannot be placed are not counted, so the command says
		// that nothing was filled.
		//
		// (Measured on a real 1.12.2 server: `fill 0 256 0 1 256 1
		// stone` answers "No blocks filled".)
	}
	if min.Y < 0 || max.Y >= 256 {
		// 1.12.2 refuses a region that reaches out of the height of the
		// world, before it changes anything.
		//
		// (Measured on a real 1.12.2 server: `fill 0 256 0 1 256 1
		// stone` answers "Cannot place blocks outside of the world".)
		return 0, exec.Errorf("Cannot place blocks outside of the world")
	}
	// 1.12.2 checks that every chunk of the region is loaded.
	if !regionLoaded(ctx.World, min, max) {
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
						if !replaceValueRead {
							replaceValueRead = true
							if replaceData != "" && replaceData != "-1" && replaceData != "*" {
								v, err := replaceBlockData(fillBlockName(replaceBlock), replaceData)
								if err != nil {
									return 0, err
								}
								replaceValue = v
								hasReplaceValue = true
							}
						}
						if hasReplaceValue && existing.Data != replaceValue {
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
func (c *blockDataCmd) Usage() string { return usageLine("blockdata") }

// Parse implements exec.Handler.
func (c *blockDataCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	pos, err := p.CoordWith(coordBlockPos)
	if err != nil {
		return nil, err
	}
	// The dataTag is parsed when the command runs: 1.12.2 looks at the
	// block first and only then at the dataTag.
	text, err := dataTagText(p)
	if err != nil {
		return nil, err
	}
	return []ast.Arg{
		&ast.Coord{X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ},
		&ast.StringArg{Value: text},
	}, nil
}

// Run implements exec.Handler.
func (c *blockDataCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	pos := resolveRelative(*args[0].(*ast.Coord), ctx.Source.RelativeBase()).BlockPos()
	// 1.12.2 refuses a position whose chunk is not loaded.
	if !ctx.World.IsLoaded(pos) {
		return 0, exec.Errorf("Cannot change block outside of the world")
	}
	data, err := parseDataTagArg(args[1].(*ast.StringArg).Value)
	if err != nil {
		return 0, err
	}
	existing := ctx.World.Blocks.TileEntity(pos)
	if existing == nil {
		existing = nbt.NewCompound()
	}
	mergeNBT(existing, data)
	ctx.World.Blocks.SetTileEntity(pos, existing)
	return 1, nil
}

// testForBlockCmd implements `/testforblock`.
type testForBlockCmd struct{}

// Name implements exec.Handler.
func (c *testForBlockCmd) Name() string { return "testforblock" }

// Usage implements exec.Handler.
func (c *testForBlockCmd) Usage() string { return usageLine("testforblock") }

// Parse implements exec.Handler.
func (c *testForBlockCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	pos, err := p.CoordWith(coordBlockPos)
	if err != nil {
		return nil, err
	}
	block, err := p.String()
	if err != nil {
		return nil, err
	}
	if err := checkBlockName(p, block); err != nil {
		return nil, err
	}
	ref := &ast.BlockRef{Name: block, Data: -1}
	// The fifth argument is the data value of the block. 1.12.2 reads it
	// with `CommandAbstract.b(Block, String)`, which is the predicate of a
	// block: a number is compared with the data value of the block, and
	// anything else is a state string the block has to know.
	//
	// The argument is read here, but the failure is reported when the
	// command runs: 1.12.2 looks at the block first.
	data := &blockDataArg{Text: "", Any: true}
	if p.More() {
		t, _ := p.Next()
		data = parseBlockDataArg(fillBlockName(block), t.Text)
	}
	return []ast.Arg{
		&ast.Coord{X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ},
		ref,
		data,
	}, nil
}

// Run implements exec.Handler.
func (c *testForBlockCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	pos := resolveRelative(*args[0].(*ast.Coord), ctx.Source.RelativeBase()).BlockPos()
	ref := args[1].(*ast.BlockRef)
	data := args[2].(*blockDataArg)
	// 1.12.2 refuses a position whose chunk is not loaded, before it
	// looks at the block.
	if !ctx.World.IsLoaded(pos) {
		return 0, exec.Errorf("Cannot test for block outside of the world")
	}
	block := ctx.World.Blocks.Get(pos)
	if !strings.EqualFold(block.Name, ref.Name) {
		return 0, exec.Errorf("The block at %s is %s (expected: %s)",
			blockPos(pos), describeBlocks(block), describeBlocks(world.BlockState{Name: ref.Name, Data: ref.Data}))
	}
	// 1.12.2 looks at the block first and only then at its data value, so
	// a data value the block does not know is reported after the block
	// turned out to be the right one.
	if data.Err != nil {
		return 0, data.Err
	}
	if !data.Any && block.Data != data.Data {
		if data.IsState {
			// The server shows the state of the block and the state
			// that was asked for.
			return 0, exec.Errorf("The block at %s had the data value of %s (expected: %s)",
				blockPos(pos), blockStateText(fillBlockName(block.Name), block.Data), data.Text)
		}
		return 0, exec.Errorf("The block at %s had the data value of %d (expected: %d)",
			blockPos(pos), block.Data, data.Data)
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
