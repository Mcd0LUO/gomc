package v1_12_2

import (
	"strconv"
	"strings"
)

// The block states of 1.12.2.
//
// 1.12.2 gives every block a set of properties, and a command can ask for a
// state of the block instead of a data value:
//
//	/fill 2 4 2 3 4 3 stone variant=granite
//	/setblock 8 4 8 wool color=red
//
// The state string is read by `CommandAbstract.c(Block, String)`: it is
// split on ',' and every part is split on '=' into a name and a value. The
// name has to be a property of the block and the value has to be one of the
// values of that property. The properties that are not named keep the value
// they have in the default state of the block, and the data value of the
// result is the data value of that state.
//
// The table in blockstates_table.go is dug out of the real 1.12.2 server
// jar: it holds, for every block that has properties, the properties it has,
// the values they take and the data value of every combination of them.
//
// (Measured on a real 1.12.2 server: `fill 2 4 2 3 4 3 stone variant=granite`
// fills and puts the granite there, which is the data value 1;
// `fill ... stone variant=bogus` answers "'variant=bogus' is not a state for
// block minecraft:stone"; `fill ... stone DEFAULT` answers the same, because
// the names are matched exactly; `fill ... stone default` fills with the
// default state of the stone.)
type blockStateKind int

const (
	// stateKindName means the value of the property is one of its names,
	// spelled exactly like the server spells it.
	stateKindName blockStateKind = iota
	// stateKindInt means the value of the property is a number between
	// the smallest and the biggest value it takes.
	stateKindInt
)

// blockStateProp is one property of a block.
type blockStateProp struct {
	// name is the name of the property, like "variant" or "color".
	name string
	// kind tells how the values of the property are written.
	kind blockStateKind
	// values are the values of the property, in the order the server
	// gives them. The index of a value in this list is what the data
	// values of the block are indexed by.
	values []string
	// min is the smallest value of an integer property. The values of
	// an integer property are the numbers from min to min+len(values)-1.
	min int32
}

// blockStateTable is the state table of one block.
type blockStateTable struct {
	// props are the properties of the block, in the order the server
	// gives them.
	props []blockStateProp
	// defaults holds, for every property, the index of the value it has
	// in the default state of the block.
	defaults []int32
	// metas holds the data value of every combination of the values of
	// the properties. The first property is the slowest one, so the
	// combination of the indexes i0, i1, ... is at
	// ((i0*len(values1) + i1)*len(values2) + i2)...
	metas []uint8
	// legacy holds the translation of a data value: the data value of
	// the state that `fromLegacyData(n)` gives. A -1 means the block has
	// no such state, so the data value is refused.
	legacy [16]int8
}

// blockStateData returns the data value of a state of a block.
//
// The block is the name of the block in the registry ("minecraft:stone") and
// the text is the state string of the command. The second return value is
// false when 1.12.2 refuses the string, which it reports as
// "'%s' is not a state for block %s".
func blockStateData(block, text string) (int, bool) {
	table := blockStateTables[block]
	if table == nil {
		// A block without properties only knows the word "default",
		// which is its own default state.
		if text == "default" {
			return 0, true
		}
		return 0, false
	}
	if text == "default" {
		return table.defaultData(), true
	}
	// The properties that the string does not name keep the value they
	// have in the default state of the block.
	indexes := make([]int32, len(table.props))
	copy(indexes, table.defaults)
	for _, part := range strings.Split(text, ",") {
		name, value, ok := splitStatePair(part)
		if !ok {
			return 0, false
		}
		prop := table.propertyIndex(name)
		if prop < 0 {
			return 0, false
		}
		index, ok := table.props[prop].valueIndex(value)
		if !ok {
			return 0, false
		}
		indexes[prop] = int32(index)
	}
	return table.data(indexes), true
}

// blockDefaultData returns the data value of the default state of a block.
//
// 1.12.2 places the default state of the block when the command has no data
// value, so the default state of a chest is the data value 2 and the
// default state of a torch the data value 5.
func blockDefaultData(block string) int {
	table := blockStateTables[block]
	if table == nil {
		return 0
	}
	return table.defaultData()
}

// blockLegacyData returns the data value a command puts down when it asks
// for a data value.
//
// 1.12.2 keeps the whole state of a block and only the command speaks in
// data values, so the two are not the same thing: `CommandAbstract.a(Block,
// String)` hands the data value to `Block.fromLegacyData` and the block
// that is placed has the data value of that state.
//
// For most blocks the two are the same, but not for all of them:
//   - a chest that is asked for the data value 0 is a chest that looks
//     north, and that is the data value 2;
//   - a block without properties is always the data value 0, whatever the
//     command asks for;
//   - a wheat only has the data values 0 to 7, so the data value 9 is
//     refused.
//
// The second return value is false when the block has no such data value.
// 1.12.2 then falls through to the state string and refuses that too.
func blockLegacyData(block string, n int) (int, bool) {
	table := blockStateTables[block]
	if table == nil {
		// A block the table does not know keeps the number as it is.
		return n, true
	}
	meta := table.legacy[n]
	if meta < 0 {
		return 0, false
	}
	return int(meta), true
}

// splitStatePair splits a part of a state string into a name and a value.
//
// 1.12.2 splits the part on '=' and keeps at most two pieces, so the value
// of "variant=granite=1" is "granite=1" and not a name of its own.
func splitStatePair(part string) (string, string, bool) {
	i := strings.IndexByte(part, '=')
	if i < 0 {
		return "", "", false
	}
	return part[:i], part[i+1:], true
}

// propertyIndex returns the index of the property with that name, or -1.
//
// The names are matched exactly: 1.12.2 looks them up in a map.
func (t *blockStateTable) propertyIndex(name string) int {
	for i := range t.props {
		if t.props[i].name == name {
			return i
		}
	}
	return -1
}

// valueIndex returns the index of the value with that text, or false.
//
// A name has to be spelled exactly like the server spells it. A number is
// read the way Java reads it, so "007" is the number 7.
func (p *blockStateProp) valueIndex(text string) (int, bool) {
	if p.kind == stateKindInt {
		n, err := strconv.ParseInt(text, 10, 32)
		if err != nil {
			return 0, false
		}
		index := n - int64(p.min)
		if index < 0 || index >= int64(len(p.values)) {
			return 0, false
		}
		return int(index), true
	}
	for i, value := range p.values {
		if value == text {
			return i, true
		}
	}
	return 0, false
}

// data returns the data value of the state that has those values.
func (t *blockStateTable) data(indexes []int32) int {
	offset := int32(0)
	for k, index := range indexes {
		offset = offset*int32(len(t.props[k].values)) + index
	}
	return int(t.metas[offset])
}

// defaultData returns the data value of the default state of the block.
func (t *blockStateTable) defaultData() int {
	return t.data(t.defaults)
}

// stateIndexes returns the value indexes of the state at that place of the
// table.
func (t *blockStateTable) stateIndexes(offset int32) []int32 {
	indexes := make([]int32, len(t.props))
	for k := len(t.props) - 1; k >= 0; k-- {
		n := int32(len(t.props[k].values))
		indexes[k] = offset % n
		offset /= n
	}
	return indexes
}

// stateText renders the properties of a state the way 1.12.2 does.
func (t *blockStateTable) stateText(indexes []int32) string {
	var sb strings.Builder
	sb.WriteString("[")
	for k, prop := range t.props {
		if k > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(prop.name)
		sb.WriteString("=")
		sb.WriteString(prop.values[indexes[k]])
	}
	sb.WriteString("]")
	return sb.String()
}
