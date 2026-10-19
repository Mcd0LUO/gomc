// Package world holds the mutable state a command execution observes.
//
// The state is deliberately minimal: it contains exactly what Minecraft
// commands can read or write, and nothing about rendering, physics or AI.
package world

import (
	"fmt"
	"strings"

	"github.com/mcsim/gomc/internal/nbt"
)

// Pos is a block position.
type Pos struct {
	// X, Y, Z are the block coordinates.
	X, Y, Z int
}

// String renders the position in "x y z" form.
func (p Pos) String() string { return fmt.Sprintf("%d %d %d", p.X, p.Y, p.Z) }

// Vec is a floating point position.
type Vec struct {
	// X, Y, Z are the coordinates.
	X, Y, Z float64
}

// BlockPos floors a vector to block coordinates.
func (v Vec) BlockPos() Pos {
	return Pos{
		X: floor(v.X),
		Y: floor(v.Y),
		Z: floor(v.Z),
	}
}

// floor returns the floor of f as an int, matching Minecraft's block
// coordinate rounding (which truncates toward negative infinity).
func floor(f float64) int {
	i := int(f)
	if float64(i) > f {
		i--
	}
	return i
}

// BlockState is a block type plus its data value.
type BlockState struct {
	// Name is the block identifier, e.g. "stone".
	Name string
	// Data is the block's data value, 0-15.
	Data int
}

// String renders the block as "name" or "name data".
func (b BlockState) String() string {
	if b.Data == 0 {
		return b.Name
	}
	return fmt.Sprintf("%s %d", b.Name, b.Data)
}

// IsAir reports whether the block is air.
//
// 1.12.2's `World.isEmpty` treats air and its empty variants as empty,
// which is what the `keep` fill and setblock modes check.
func (b BlockState) IsAir() bool {
	return b.Name == "" || strings.EqualFold(b.Name, "air")
}

// BlockStore holds the world's block data.
//
// Only modified blocks are stored; everything else is air. This mirrors how
// a test harness needs to behave and keeps memory proportional to what the
// functions actually touched.
type BlockStore struct {
	blocks map[Pos]BlockState
	// entities holds tile entity NBT keyed by position.
	entities map[Pos]*nbt.Compound
}

// NewBlockStore returns an empty block store.
func NewBlockStore() *BlockStore {
	return &BlockStore{
		blocks:   map[Pos]BlockState{},
		entities: map[Pos]*nbt.Compound{},
	}
}

// Get returns the block at pos. Unset positions are air.
func (b *BlockStore) Get(pos Pos) BlockState {
	if st, ok := b.blocks[pos]; ok {
		return st
	}
	return BlockState{Name: "air", Data: 0}
}

// Set stores the block at pos.
func (b *BlockStore) Set(pos Pos, st BlockState) {
	if st.Name == "air" && st.Data == 0 {
		delete(b.blocks, pos)
		return
	}
	b.blocks[pos] = st
}

// Len returns the number of non-air blocks.
func (b *BlockStore) Len() int { return len(b.blocks) }

// Positions returns every position holding a non-air block.
func (b *BlockStore) Positions() []Pos {
	out := make([]Pos, 0, len(b.blocks))
	for p := range b.blocks {
		out = append(out, p)
	}
	return out
}

// TileEntity returns the tile entity NBT at pos, or nil.
func (b *BlockStore) TileEntity(pos Pos) *nbt.Compound {
	return b.entities[pos]
}

// SetTileEntity stores tile entity NBT at pos.
func (b *BlockStore) SetTileEntity(pos Pos, c *nbt.Compound) {
	if c == nil {
		delete(b.entities, pos)
		return
	}
	b.entities[pos] = c
}
