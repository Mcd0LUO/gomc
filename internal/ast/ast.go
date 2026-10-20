// Package ast defines the abstract syntax tree for Minecraft commands.
//
// The tree is intentionally data-only: it describes what a command means,
// never how it is executed. Execution lives in the exec package, and the
// version-specific parsing rules live in the version packages.
package ast

import "github.com/Mcd0LUO/gomc/internal/nbt"

// Node is any AST node.
type Node interface {
	// Command returns the command name this node belongs to.
	Command() string
}

// Command is the root of a parsed command.
type Command struct {
	// Name is the command's literal name, e.g. "scoreboard".
	Name string
	// Args holds the command's arguments in declaration order.
	Args []Arg
	// Raw is the original source text of the whole command.
	Raw string
	// Line is the 1-based line the command started on.
	Line int
	// Col is the 1-based column the command started on.
	Col int
}

// Command implements Node.
func (c *Command) Command() string { return c.Name }

// Arg is a single argument of a command.
type Arg interface {
	// ArgKind identifies the argument's type.
	ArgKind() Kind
}

// Kind identifies an argument type.
type Kind int

const (
	// KindString is a plain string argument.
	KindString Kind = iota
	// KindInt is an integer argument.
	KindInt
	// KindFloat is a floating point argument.
	KindFloat
	// KindCoord is a coordinate triple.
	KindCoord
	// KindSelector is a target selector such as @a[r=10].
	KindSelector
	// KindBlockRef is a block reference with optional data value.
	KindBlockRef
	// KindNBT is an NBT compound.
	KindNBT
	// KindJSON is a raw JSON text component.
	KindJSON
	// KindList is a list of arguments.
	KindList
	// KindLiteral is a fixed keyword.
	KindLiteral
	// KindScore is a scoreboard holder/objective pair.
	KindScore
	// KindEntityType is an entity type identifier.
	KindEntityType
	// KindItem is an item identifier with optional count and data.
	KindItem
)

// StringArg is a plain string.
type StringArg struct {
	// Value is the unquoted text.
	Value string
	// Offset is the byte offset within the line.
	Offset int
}

// ArgKind implements Arg.
func (a *StringArg) ArgKind() Kind { return KindString }

// IntArg is an integer.
type IntArg struct {
	// Value is the parsed number.
	Value int
	// Offset is the byte offset within the line.
	Offset int
}

// ArgKind implements Arg.
func (a *IntArg) ArgKind() Kind { return KindInt }

// FloatArg is a floating point number.
type FloatArg struct {
	// Value is the parsed number.
	Value float64
	// Offset is the byte offset within the line.
	Offset int
}

// ArgKind implements Arg.
func (a *FloatArg) ArgKind() Kind { return KindFloat }

// Coord is a coordinate triple.
type Coord struct {
	// X, Y, Z are the resolved absolute coordinates.
	X, Y, Z float64
	// RelativeX, RelativeY, RelativeZ mark which components were relative
	// to the execution position.
	RelativeX, RelativeY, RelativeZ bool
	// Offset is the byte offset within the line.
	Offset int
}

// ArgKind implements Arg.
func (a *Coord) ArgKind() Kind { return KindCoord }

// Selector is a target selector.
type Selector struct {
	// Base is the selector base: 'p', 'r', 'a', 'e' or 's'.
	Base string
	// Args holds the raw selector arguments in source order.
	Args []SelectorArg
	// Raw is the selector exactly as it was written, e.g. "@e[type=zombie]".
	// 1.12.2 quotes the original text in its error messages, so the raw
	// spelling has to survive parsing.
	Raw string
	// Offset is the byte offset within the line.
	Offset int
}

// ArgKind implements Arg.
func (a *Selector) ArgKind() Kind { return KindSelector }

// SelectorArg is one `key=value` pair inside a selector.
type SelectorArg struct {
	// Key is the argument name.
	Key string
	// Value is the raw value text.
	Value string
	// Offset is the byte offset within the line.
	Offset int
}

// BlockRef references a block type with an optional data value.
type BlockRef struct {
	// Name is the block identifier, e.g. "stone" or "wool".
	Name string
	// Data is the data value, or -1 when unspecified.
	Data int
	// HasData reports whether a data value was given.
	HasData bool
	// Offset is the byte offset within the line.
	Offset int
}

// ArgKind implements Arg.
func (a *BlockRef) ArgKind() Kind { return KindBlockRef }

// NBTArg is an NBT compound.
type NBTArg struct {
	// Value is the parsed compound.
	Value *nbt.Compound
	// Offset is the byte offset within the line.
	Offset int
}

// ArgKind implements Arg.
func (a *NBTArg) ArgKind() Kind { return KindNBT }

// JSONArg is a raw JSON text component.
type JSONArg struct {
	// Value is the raw JSON text.
	Value string
	// Offset is the byte offset within the line.
	Offset int
}

// ArgKind implements Arg.
func (a *JSONArg) ArgKind() Kind { return KindJSON }

// ListArg is a list of arguments.
type ListArg struct {
	// Items holds the list elements.
	Items []Arg
	// Offset is the byte offset within the line.
	Offset int
}

// ArgKind implements Arg.
func (a *ListArg) ArgKind() Kind { return KindList }

// LiteralArg is a fixed keyword.
type LiteralArg struct {
	// Value is the keyword.
	Value string
	// Offset is the byte offset within the line.
	Offset int
}

// ArgKind implements Arg.
func (a *LiteralArg) ArgKind() Kind { return KindLiteral }

// ScoreArg names a scoreboard holder and objective.
type ScoreArg struct {
	// Holder is the target: a player name or a selector.
	Holder string
	// HolderSelector is set when Holder is a selector.
	HolderSelector *Selector
	// Objective is the scoreboard objective name.
	Objective string
	// Offset is the byte offset within the line.
	Offset int
}

// ArgKind implements Arg.
func (a *ScoreArg) ArgKind() Kind { return KindScore }

// EntityTypeArg names an entity type.
type EntityTypeArg struct {
	// Name is the entity type identifier, e.g. "zombie".
	Name string
	// Offset is the byte offset within the line.
	Offset int
}

// ArgKind implements Arg.
func (a *EntityTypeArg) ArgKind() Kind { return KindEntityType }

// ItemArg names an item with optional count and data value.
type ItemArg struct {
	// Name is the item identifier.
	Name string
	// Count is the stack size, or -1 when unspecified.
	Count int
	// Data is the data value, or -1 when unspecified.
	Data int
	// NBT holds optional item NBT.
	NBT *nbt.Compound
	// Offset is the byte offset within the line.
	Offset int
}

// ArgKind implements Arg.
func (a *ItemArg) ArgKind() Kind { return KindItem }
