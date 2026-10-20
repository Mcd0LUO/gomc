package v1_12_2

import (
	"strconv"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// fillDataValueKind tells what the eighth argument of `/fill` held.
type fillDataValueKind int

const (
	// fillDataValueAbsent means the argument was not written at all.
	fillDataValueAbsent fillDataValueKind = iota
	// fillDataValueNumber means the argument was a data value.
	fillDataValueNumber
	// fillDataValueDefault means the argument was the word "default".
	fillDataValueDefault
	// fillDataValueTooSmall means the argument was a negative number.
	fillDataValueTooSmall
	// fillDataValueTooBig means the argument was a number above 15.
	fillDataValueTooBig
	// fillDataValueState means the argument was a block state string.
	fillDataValueState
)

// fillDataValue is the eighth argument of `/fill`: the data value of the
// block that is placed.
//
// 1.12.2 reads that slot when the command runs and not when it is
// parsed: `CommandFill.execute` hands the text to
// `CommandBase.a(Block, String)` as soon as the block name is known. The
// slot is therefore classified while the command is parsed, and the
// failure is reported when the command runs.
//
// The classification is the one of `CommandBase.a(Block, String)`:
//
//   - a decimal number is the legacy data value of the block, and it has
//     to be between 0 and 15.
//   - "default" asks for the default state of the block.
//   - everything else is a block state string: a list of `name=value`
//     pairs that the block has to know.
//
// gomc has no block states, so it refuses every state string. The wording
// is the one of the server, but the server accepts the states of the
// blocks that have them.
//
// (Measured on a real 1.12.2 server: `fill 2 4 2 3 4 3 stone abc` answers
// "'abc' is not a state for block minecraft:stone", `fill ... stone -1`
// answers "The number you have entered (-1) is too small, it must be at
// least 0" and `fill ... stone 16` answers "The number you have entered
// (16) is too big, it must be at most 15". `fill ... stone default` fills,
// and so does `fill ... stone variant=granite`, which puts the granite
// (the data value 1) there.)
type fillDataValue struct {
	// Kind tells what the slot held.
	Kind fillDataValueKind
	// Data is the data value of the slot. It is the number 1.12.2
	// complains about when Kind is fillDataValueTooSmall or
	// fillDataValueTooBig.
	Data int
	// Text is the raw text of the slot.
	Text string
}

// ArgKind implements ast.Arg.
//
// The slot belongs to the block reference of the command, so that is what
// it is reported as.
func (d *fillDataValue) ArgKind() ast.Kind { return ast.KindBlockRef }

// parseFillDataValue classifies the eighth argument of `/fill`.
//
// The text is the raw text of the argument. 1.12.2 reads it with
// `Integer.parseInt`, which takes a decimal integer with an optional sign
// and nothing else; a number that does not fit in 32 bits is not a number
// either.
func parseFillDataValue(text string) *fillDataValue {
	if n, err := strconv.ParseInt(text, 10, 32); err == nil {
		switch {
		case n < 0:
			return &fillDataValue{Kind: fillDataValueTooSmall, Data: int(n), Text: text}
		case n > 15:
			return &fillDataValue{Kind: fillDataValueTooBig, Data: int(n), Text: text}
		default:
			return &fillDataValue{Kind: fillDataValueNumber, Data: int(n), Text: text}
		}
	}
	if text == "default" {
		// The one state string every block understands: its default
		// state, which is the data value 0.
		return &fillDataValue{Kind: fillDataValueDefault, Data: 0, Text: text}
	}
	return &fillDataValue{Kind: fillDataValueState, Text: text}
}

// err returns the failure 1.12.2 answers for the slot, or nil when the
// slot was acceptable.
//
// The wording is the one of the language file:
// `commands.generic.num.tooSmall`, `commands.generic.num.tooBig` and
// `commands.generic.blockstate.invalid`.
func (d *fillDataValue) err(block string) error {
	switch d.Kind {
	case fillDataValueTooSmall:
		return exec.Errorf("The number you have entered (%d) is too small, it must be at least 0", d.Data)
	case fillDataValueTooBig:
		return exec.Errorf("The number you have entered (%d) is too big, it must be at most 15", d.Data)
	case fillDataValueState:
		return exec.Errorf("'%s' is not a state for block %s", d.Text, block)
	}
	return nil
}

// data returns the data value of the block that is placed.
func (d *fillDataValue) data() int {
	if d.Kind == fillDataValueNumber {
		return d.Data
	}
	// The slot was absent or asked for the default state of the block,
	// which is the data value 0.
	return 0
}

// fillBlockName returns the name 1.12.2 prints for a block.
//
// The server looks the block up in the block registry and prints the key
// it found there, so "stone", "STONE" and "minecraft:stone" are all
// printed as "minecraft:stone".
func fillBlockName(name string) string {
	namespace, key := world.EntityTypeKey(name)
	return namespace + ":" + key
}
