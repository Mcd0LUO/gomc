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

// cloneCmd implements `/clone`.
//
// The 1.12.2 syntax is `clone <x1> <y1> <z1> <x2> <y2> <z2> <x> <y> <z>
// [<maskMode> [<cloneMode>]]`, where the mask mode is `normal`, `masked` or
// `filtered` and the clone mode is `normal`, `force` or `move`.
//
// The source region is the box between the first two corners, and the
// destination starts at the third corner, so the two regions have to be the
// same size.
//
// The order of the two modes is the one of `CommandClone`: the mask mode
// is the tenth argument and the clone mode the eleventh. A `filtered` mask
// takes its block from the twelfth argument, so the clone mode comes
// between them.
//
// A mode that 1.12.2 does not know is not an error: an unknown mask mode
// behaves like `normal` and an unknown clone mode like `normal` as well.
type cloneCmd struct{}

// Name implements exec.Handler.
func (c *cloneCmd) Name() string { return "clone" }

// Usage implements exec.Handler.
func (c *cloneCmd) Usage() string { return usageLine("clone") }

// Parse implements exec.Handler.
func (c *cloneCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	corners := make([]*ast.Coord, 0, 3)
	for i := 0; i < 3; i++ {
		pos, err := p.CoordWith(ast.Coord{}, coordBlockPos)
		if err != nil {
			return nil, err
		}
		corners = append(corners, &ast.Coord{
			X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ,
		})
	}
	mask := "normal"
	if p.More() {
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		mask = name
	}
	// The clone mode is empty when it was not given, because 1.12.2 only
	// allows an overlapping clone when the mode was given and is `force`
	// or `move`.
	clone := ""
	if p.More() {
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		clone = name
	}
	// A `filtered` mask needs the block to filter on. 1.12.2 reads it
	// from the twelfth argument and refuses the command when it is
	// missing.
	filter := &ast.BlockRef{Name: "", Data: -1}
	filterData := -1
	if mask == "filtered" {
		if !p.More() {
			return nil, usageError(p, "clone", "expected a block to filter on")
		}
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		if err := checkBlockName(p, name); err != nil {
			return nil, err
		}
		filter = &ast.BlockRef{Name: name, Data: -1}
		if p.More() {
			data, err := p.String()
			if err != nil {
				return nil, err
			}
			// `-1` and `*` match every data value, exactly like the
			// predicate of `CommandAbstract.b`.
			if data != "-1" && data != "*" {
				n, err := parseIntSafe(data)
				if err != nil {
					return nil, p.Errorf("expected a data value, got %q", data)
				}
				filterData = n
				filter.Data = n
				filter.HasData = true
			}
		}
	}
	return []ast.Arg{
		corners[0], corners[1], corners[2],
		&ast.StringArg{Value: mask},
		&ast.StringArg{Value: clone},
		filter,
		&ast.IntArg{Value: filterData},
	}, nil
}

// Run implements exec.Handler.
func (c *cloneCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	srcMin, srcMax := orderCorners(
		resolveRelative(*args[0].(*ast.Coord), ctx.Source.Pos).BlockPos(),
		resolveRelative(*args[1].(*ast.Coord), ctx.Source.Pos).BlockPos())
	dst := resolveRelative(*args[2].(*ast.Coord), ctx.Source.Pos).BlockPos()
	mask := args[3].(*ast.StringArg).Value
	clone := args[4].(*ast.StringArg).Value
	filter := args[5].(*ast.BlockRef)
	filterData := args[6].(*ast.IntArg).Value

	sizeX, sizeY, sizeZ := regionSizes(srcMin, srcMax)

	volume := regionVolume(srcMin, srcMax)
	if volume > 32768 {
		return 0, exec.Errorf("Too many blocks in the specified area (%d > 32768)", volume)
	}
	// The destination region has the size of the source region.
	dstMax := world.Pos{X: dst.X + sizeX - 1, Y: dst.Y + sizeY - 1, Z: dst.Z + sizeZ - 1}
	// 1.12.2 refuses to clone a region onto itself, unless the clone mode
	// is `force` or `move`.
	if clone != "force" && clone != "move" && boxesOverlap(srcMin, srcMax, dst, dstMax) {
		return 0, exec.Errorf("Source and destination can not overlap")
	}
	if srcMin.Y < 0 || srcMax.Y >= 256 || dst.Y < 0 || dstMax.Y >= 256 {
		return 0, exec.Errorf("Cannot access blocks outside of the world")
	}

	// The source is read before anything is written, so a region that
	// overlaps its destination is copied the way 1.12.2 copies it: the
	// blocks are collected first and written afterwards.
	type cloneEntry struct {
		dst   world.Pos
		state world.BlockState
		nbt   *nbt.Compound
	}
	var copies []cloneEntry
	var sources []world.Pos
	for x := 0; x < sizeX; x++ {
		for y := 0; y < sizeY; y++ {
			for z := 0; z < sizeZ; z++ {
				from := world.Pos{X: srcMin.X + x, Y: srcMin.Y + y, Z: srcMin.Z + z}
				block := ctx.World.Blocks.Get(from)
				if mask == "masked" && block.IsAir() {
					continue
				}
				if filter.Name != "" {
					if !strings.EqualFold(block.Name, filter.Name) {
						continue
					}
					if filterData >= 0 && block.Data != filterData {
						continue
					}
				}
				copies = append(copies, cloneEntry{
					dst:   world.Pos{X: dst.X + x, Y: dst.Y + y, Z: dst.Z + z},
					state: block,
					nbt:   ctx.World.Blocks.TileEntity(from),
				})
				sources = append(sources, from)
			}
		}
	}
	if clone == "move" {
		// The source is emptied before the destination is written, so a
		// move onto an overlapping region keeps the copied blocks.
		for _, from := range sources {
			ctx.World.Blocks.Set(from, world.BlockState{Name: "air"})
			ctx.World.Blocks.SetTileEntity(from, nil)
		}
	}
	for _, c := range copies {
		ctx.World.Blocks.Set(c.dst, c.state)
		if c.nbt != nil {
			ctx.World.Blocks.SetTileEntity(c.dst, c.nbt.Clone())
		}
	}
	if len(copies) == 0 {
		return 0, exec.Errorf("No blocks cloned")
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("%d blocks cloned", len(copies)))
	return len(copies), nil
}

// boxesOverlap reports whether two regions share a block.
//
// It is `StructureBoundingBox.a(StructureBoundingBox)`, which 1.12.2 uses
// to decide whether a clone would overwrite its own source.
func boxesOverlap(aMin, aMax, bMin, bMax world.Pos) bool {
	return aMin.X <= bMax.X && bMin.X <= aMax.X &&
		aMin.Y <= bMax.Y && bMin.Y <= aMax.Y &&
		aMin.Z <= bMax.Z && bMin.Z <= aMax.Z
}

// testForBlocksCmd implements `/testforblocks`.
//
// The 1.12.2 syntax is `testforblocks <x1> <y1> <z1> <x2> <y2> <z2>
// <x> <y> <z> [mode]`: it compares two regions of the same size and
// reports whether they are identical.
//
// The first two corners are the source region, the third corner is the
// origin of the destination region, which has the size of the source
// region. The optional mode is `masked`, which ignores the air blocks of
// the source region, or `all`, which compares everything.
type testForBlocksCmd struct{}

// Name implements exec.Handler.
func (c *testForBlocksCmd) Name() string { return "testforblocks" }

// Usage implements exec.Handler.
func (c *testForBlocksCmd) Usage() string { return usageLine("compare") }

// Parse implements exec.Handler.
func (c *testForBlocksCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	corners := make([]*ast.Coord, 0, 3)
	for i := 0; i < 3; i++ {
		pos, err := p.CoordWith(ast.Coord{}, coordBlockPos)
		if err != nil {
			return nil, err
		}
		corners = append(corners, &ast.Coord{
			X: pos.X, Y: pos.Y, Z: pos.Z,
			RelativeX: pos.RelativeX, RelativeY: pos.RelativeY, RelativeZ: pos.RelativeZ,
		})
	}
	// 1.12.2 only knows `masked`; every other word, and no word at all,
	// compares everything.
	mode := "all"
	if p.More() {
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		mode = name
	}
	return []ast.Arg{
		corners[0], corners[1], corners[2],
		&ast.StringArg{Value: mode},
	}, nil
}

// Run implements exec.Handler.
func (c *testForBlocksCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	srcMin, srcMax := orderCorners(
		resolveRelative(*args[0].(*ast.Coord), ctx.Source.Pos).BlockPos(),
		resolveRelative(*args[1].(*ast.Coord), ctx.Source.Pos).BlockPos())
	dst := resolveRelative(*args[2].(*ast.Coord), ctx.Source.Pos).BlockPos()
	mode := args[3].(*ast.StringArg).Value

	sizeX, sizeY, sizeZ := regionSizes(srcMin, srcMax)

	volume := regionVolume(srcMin, srcMax)
	// The limit of the comparison is much higher than the one of a fill.
	if volume > 524288 {
		return 0, exec.Errorf("Too many blocks in the specified area (%d > 524288)", volume)
	}
	dstMax := world.Pos{X: dst.X + sizeX - 1, Y: dst.Y + sizeY - 1, Z: dst.Z + sizeZ - 1}
	if srcMin.Y < 0 || srcMax.Y >= 256 || dst.Y < 0 || dstMax.Y >= 256 {
		return 0, exec.Errorf("Cannot access blocks outside of the world")
	}
	masked := mode == "masked"
	compared := 0
	for x := 0; x < sizeX; x++ {
		for y := 0; y < sizeY; y++ {
			for z := 0; z < sizeZ; z++ {
				from := world.Pos{X: srcMin.X + x, Y: srcMin.Y + y, Z: srcMin.Z + z}
				to := world.Pos{X: dst.X + x, Y: dst.Y + y, Z: dst.Z + z}
				src := ctx.World.Blocks.Get(from)
				if masked && src.IsAir() {
					// The air blocks of the source region are skipped
					// and do not count as compared.
					continue
				}
				if !sameBlockState(src, ctx.World.Blocks.Get(to)) ||
					!sameTileEntity(ctx.World, from, to) {
					return 0, exec.Errorf("Source and destination are not identical")
				}
				compared++
			}
		}
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("%d blocks compared", compared))
	return compared, nil
}

// sameBlockState reports whether two blocks are the same.
//
// 1.12.2 compares the block data, which is the block and its properties.
func sameBlockState(a, b world.BlockState) bool {
	return strings.EqualFold(a.Name, b.Name) && a.Data == b.Data
}

// sameTileEntity reports whether the tile entities of two positions are
// the same.
//
// 1.12.2 saves the tile entity, drops the position from the saved data
// and compares the rest. A tile entity that only the destination has is
// not a difference, because 1.12.2 only looks at the source.
func sameTileEntity(w *world.World, from, to world.Pos) bool {
	src := w.Blocks.TileEntity(from)
	dst := w.Blocks.TileEntity(to)
	if src == nil {
		return true
	}
	if dst == nil {
		return false
	}
	return sameNBT(tileEntityData(src), tileEntityData(dst))
}

// tileEntityData returns a copy of a tile entity without its position.
//
// 1.12.2 removes the `x`, `y` and `z` entries before it compares the
// saved data, so two chests at different places still count as the same.
func tileEntityData(c *nbt.Compound) *nbt.Compound {
	cp := c.Clone()
	cp.Remove("x")
	cp.Remove("y")
	cp.Remove("z")
	return cp
}

// sameNBT compares two NBT compounds.
//
// The comparison ignores the order of the keys, because 1.12.2 compares
// the saved data as a map.
func sameNBT(a, b *nbt.Compound) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Len() != b.Len() {
		return false
	}
	for _, k := range a.Keys() {
		av, bv := a.Get(k), b.Get(k)
		if av == nil || bv == nil {
			if av != bv {
				return false
			}
			continue
		}
		if av.Type() != bv.Type() {
			return false
		}
		switch at := av.(type) {
		case *nbt.Compound:
			bt, ok := bv.(*nbt.Compound)
			if !ok || !sameNBT(at, bt) {
				return false
			}
		case *nbt.List:
			bt, ok := bv.(*nbt.List)
			if !ok || at.Len() != bt.Len() {
				return false
			}
			for i := 0; i < at.Len(); i++ {
				if !sameNBTValue(at.At(i), bt.At(i)) {
					return false
				}
			}
		default:
			if av.String() != bv.String() {
				return false
			}
		}
	}
	return true
}

// sameNBTValue compares two NBT values.
func sameNBTValue(a, b nbt.Value) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Type() != b.Type() {
		return false
	}
	switch at := a.(type) {
	case *nbt.Compound:
		bt, ok := b.(*nbt.Compound)
		return ok && sameNBT(at, bt)
	case *nbt.List:
		bt, ok := b.(*nbt.List)
		if !ok || at.Len() != bt.Len() {
			return false
		}
		for i := 0; i < at.Len(); i++ {
			if !sameNBTValue(at.At(i), bt.At(i)) {
				return false
			}
		}
		return true
	default:
		return a.String() == b.String()
	}
}

// registerBlocksExtra adds the block region commands.
func registerBlocksExtra(reg *exec.Registry) {
	reg.Register(&cloneCmd{})
	reg.Register(&testForBlocksCmd{})
}
