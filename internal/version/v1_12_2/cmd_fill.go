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
	//
	// 1.12.2 then places the default state of the block, which is not
	// always the data value 0: the default state of a chest is the data
	// value 2 and the default state of a torch the data value 5.
	fillDataValueAbsent fillDataValueKind = iota
	// fillDataValueNumber means the argument was a data value.
	fillDataValueNumber
	// fillDataValueDefault means the argument was the word "default".
	fillDataValueDefault
	// fillDataValueTooSmall means the argument was a negative number.
	fillDataValueTooSmall
	// fillDataValueTooBig means the argument was a number above 15.
	fillDataValueTooBig
	// fillDataValueState means the argument was a block state string
	// that the block knows.
	fillDataValueState
	// fillDataValueBadState means the argument was a block state string
	// that the block does not know.
	fillDataValueBadState
)

// fillDataValue is the eighth argument of `/fill` and the fifth argument of
// `/setblock`: the data value of the block that is placed.
//
// 1.12.2 reads that slot when the command runs and not when it is
// parsed: `CommandFill.execute` hands the text to
// `CommandAbstract.a(Block, String)` as soon as the block name is known. The
// slot is therefore classified while the command is parsed, and the
// failure is reported when the command runs.
//
// The classification is the one of `CommandAbstract.a(Block, String)`:
//
//   - a decimal number is the legacy data value of the block, and it has
//     to be between 0 and 15.
//   - "default" asks for the default state of the block.
//   - everything else is a block state string: a list of `name=value`
//     pairs that the block has to know.
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
	// fillDataValueTooBig, and the data value of the block that is
	// placed otherwise.
	//
	// For a number that is not the data value of the block as it is: the
	// number goes through `fromLegacyData` first, see blockLegacyData.
	Data int
	// Text is the raw text of the slot.
	Text string
}

// ArgKind implements ast.Arg.
//
// The slot belongs to the block reference of the command, so that is what
// it is reported as.
func (d *fillDataValue) ArgKind() ast.Kind { return ast.KindBlockRef }

// parseFillDataValue classifies the data value slot of `/fill` and
// `/setblock`.
//
// The block is the name of the block in the registry ("minecraft:stone") and
// the text is the raw text of the argument. 1.12.2 reads the text with
// `Integer.parseInt`, which takes a decimal integer with an optional sign
// and nothing else; a number that does not fit in 32 bits is not a number
// either.
//
// A text that is not a number is a block state string, and the block has to
// know it. The state string is read here, while the command is parsed, so
// that the command does not have to read text when it runs.
func parseFillDataValue(block, text string) *fillDataValue {
	if n, err := strconv.ParseInt(text, 10, 32); err == nil {
		switch {
		case n < 0:
			return &fillDataValue{Kind: fillDataValueTooSmall, Data: int(n), Text: text}
		case n > 15:
			return &fillDataValue{Kind: fillDataValueTooBig, Data: int(n), Text: text}
		}
		// The number is a data value of the block, but the block does
		// not necessarily have that data value: 1.12.2 hands the
		// number to `fromLegacyData` and the block that is placed has
		// the data value of the state that comes back. A number the
		// block has no state for makes `fromLegacyData` fail, and
		// then 1.12.2 falls through to the state string and refuses
		// that too.
		if data, ok := blockLegacyData(block, int(n)); ok {
			return &fillDataValue{Kind: fillDataValueNumber, Data: data, Text: text}
		}
		return &fillDataValue{Kind: fillDataValueBadState, Text: text}
	}
	if text == "default" {
		// The one state string every block understands: its default
		// state, whose data value is not always 0.
		return &fillDataValue{Kind: fillDataValueDefault, Data: blockDefaultData(block), Text: text}
	}
	if data, ok := blockStateData(block, text); ok {
		return &fillDataValue{Kind: fillDataValueState, Data: data, Text: text}
	}
	return &fillDataValue{Kind: fillDataValueBadState, Text: text}
}

// absentFillDataValue returns the slot of a command that has no data value.
//
// 1.12.2 places the default state of the block then, and the data value of
// that state is not always 0.
func absentFillDataValue(block string) *fillDataValue {
	return &fillDataValue{Kind: fillDataValueAbsent, Data: blockDefaultData(block)}
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
	case fillDataValueBadState:
		return exec.Errorf("'%s' is not a state for block %s", d.Text, block)
	}
	return nil
}

// data returns the data value of the block that is placed.
func (d *fillDataValue) data() int {
	return d.Data
}

// replaceBlockData reads the data value of the block that is replaced.
//
// 1.12.2 reads it with `CommandAbstract.b(Block, String)`, which is the
// predicate of a block rather than its data value: a number is compared
// with the data value of the block, and anything else is a state string
// that the block has to know.
//
// The state string is resolved to the data value of that state, so a block
// with more than one property is compared by the part of its state that the
// data value can hold.
//
// The block is the name of the block in the registry ("minecraft:stone").
func replaceBlockData(block, text string) (int, error) {
	if n, err := strconv.ParseInt(text, 10, 32); err == nil {
		// The number is compared with the data value of the block as it
		// is, so a number outside 0..15 matches nothing.
		return int(n), nil
	}
	if data, ok := blockStateData(block, text); ok {
		return data, nil
	}
	return 0, exec.Errorf("'%s' is not a state for block %s", text, block)
}

// blockDataArg is the data value of a command that asks about a block.
//
// 1.12.2 reads that argument with `CommandAbstract.b(Block, String)`, which
// is the predicate of a block:
//
//   - "*" and "-1" match every data value;
//   - a number is compared with the data value of the block;
//   - anything else is a state string that the block has to know.
//
// The argument is read while the command is parsed, but the failure is only
// reported when the command runs: 1.12.2 looks at the block first and only
// then at its data value.
type blockDataArg struct {
	// Text is the raw text of the argument.
	Text string
	// Data is the data value to compare with.
	Data int
	// Any is true when the argument matches every data value.
	Any bool
	// IsState is true when the argument was a state string.
	IsState bool
	// Err is the failure of 1.12.2, if the argument was refused.
	Err error
}

// ArgKind implements ast.Arg.
//
// The argument belongs to the block reference of the command, so that is
// what it is reported as.
func (d *blockDataArg) ArgKind() ast.Kind { return ast.KindBlockRef }

// parseBlockDataArg reads the data value argument of a command that asks
// about a block.
//
// The block is the name of the block in the registry ("minecraft:stone").
func parseBlockDataArg(block, text string) *blockDataArg {
	arg := &blockDataArg{Text: text}
	if text == "*" || text == "-1" {
		arg.Any = true
		return arg
	}
	if n, err := strconv.ParseInt(text, 10, 32); err == nil {
		// The number is compared with the data value of the block as it
		// is, so a number outside 0..15 matches nothing.
		arg.Data = int(n)
		return arg
	}
	data, ok := blockStateData(block, text)
	if !ok {
		arg.Err = exec.Errorf("'%s' is not a state for block %s", text, block)
		return arg
	}
	arg.Data = data
	arg.IsState = true
	return arg
}

// blockStateText renders the state of a block the way 1.12.2 does.
//
// It is the text of `IBlockData.toString()`: the name of the block in the
// registry, and the properties of the state in brackets.
//
// A block with more than one property can have several states with the
// same data value, and gomc only keeps the data value. The first state of
// the table with that data value is the one that is shown.
func blockStateText(block string, data int) string {
	table := blockStateTables[block]
	if table == nil || len(table.props) == 0 {
		return block
	}
	for offset := range table.metas {
		if int(table.metas[offset]) == data {
			return block + table.stateText(table.stateIndexes(int32(offset)))
		}
	}
	return block
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
