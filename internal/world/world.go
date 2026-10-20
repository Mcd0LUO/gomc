// Package world holds the mutable state a command execution observes.
//
// The state is deliberately minimal: it contains exactly what Minecraft
// commands can read or write, and nothing about rendering, physics or AI.
package world

import (
	"fmt"
	"math"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/nbt"
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
//
// A coordinate that does not fit in an int is **saturated** instead of
// being converted blindly: Go's conversion of a float64 that is out of
// range gives an implementation defined value, and a block at that
// position would be garbage.
//
// 1.12.2 does not have this problem: it reads a coordinate with
// `CommandAbstract.b(base, text, min, max, flag)`, which refuses a
// coordinate outside the bounds (see the parser package). gomc has the
// same bounds now, so this only guards the values that slip through the
// relative coordinates, where the bounds are checked on the offset.
func floor(f float64) int {
	switch {
	case math.IsNaN(f):
		return 0
	case f >= math.MaxInt64:
		return math.MaxInt64
	case f <= math.MinInt64:
		return math.MinInt64
	}
	i := int(f)
	if float64(i) > f {
		i--
	}
	return i
}

// The radii of the areas the server keeps loaded.
//
// 1.12.2 refuses to touch a block whose **chunk** is not loaded, and the
// server only loads two areas:
//
//   - the area around the world spawn: the "Preparing start region" step
//     loads the 17x17 chunks around the spawn's chunk, so the radius is
//     eight chunks;
//   - the area around every player, as far as the view distance reaches.
//
// (Measured on a real 1.12.2 server: with the spawn in the chunk (-1, -1),
// `setblock 128 4 0 stone` places the block and `setblock 144 4 0 stone`
// answers "Cannot place block outside of the world".)
const (
	// spawnChunkRadius is the radius around the spawn's chunk that the
	// server loads when it starts.
	spawnChunkRadius = 8
	// playerChunkRadius is the radius around a player that the server
	// keeps loaded. 1.12.2's default view distance is ten chunks.
	playerChunkRadius = 10
)

// IsLoaded reports whether the chunk at pos is loaded.
//
// The runtime has no chunks, so this is a model of what the server has
// loaded: the area around the world spawn and the areas around the
// players.
//
// The players are few and the entities are many, so the players are
// looked up in the entity slice instead of in the whole table: this is
// called once per command.
func (w *World) IsLoaded(pos Pos) bool {
	cx, cz := pos.X>>4, pos.Z>>4
	if chunkAreaHas(cx, cz, w.Spawn.X>>4, w.Spawn.Z>>4, spawnChunkRadius) {
		return true
	}
	for _, e := range w.EntitySlice() {
		// Only the players keep chunks loaded; the mobs do not.
		if e.Type != "player" {
			continue
		}
		ex, ez := int(e.Pos.X)>>4, int(e.Pos.Z)>>4
		if chunkAreaHas(cx, cz, ex, ez, playerChunkRadius) {
			return true
		}
	}
	return false
}

// chunkAreaHas reports whether the chunk (cx, cz) is in the square of
// chunks around (ox, oz) with the given radius.
func chunkAreaHas(cx, cz, ox, oz, radius int) bool {
	dx := cx - ox
	if dx < 0 {
		dx = -dx
	}
	dz := cz - oz
	if dz < 0 {
		dz = -dz
	}
	return dx <= radius && dz <= radius
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
