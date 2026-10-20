package v1_12_2

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// blockStateEngine builds an engine with the 1.12.2 command set.
func blockStateEngine(t *testing.T) *exec.Engine {
	t.Helper()
	reg := exec.NewRegistry()
	Register(reg)
	return exec.NewEngine(reg, world.NewWorld())
}

// TestBlockStateData checks the state strings of the data value slot.
//
// (Measured on a real 1.12.2 server: `fill 2 4 2 3 4 3 stone variant=granite`
// puts the granite there, which is the data value 1, and
// `fill ... wool color=red` puts the red wool there, which is the data
// value 14.)
func TestBlockStateData(t *testing.T) {
	cases := []struct {
		block string
		text  string
		want  int
		ok    bool
	}{
		// A single property.
		{"minecraft:stone", "variant=granite", 1, true},
		{"minecraft:stone", "variant=stone", 0, true},
		{"minecraft:stone", "variant=smooth_andesite", 6, true},
		{"minecraft:wool", "color=red", 14, true},
		{"minecraft:wool", "color=light_blue", 3, true},
		// The default state of the block.
		{"minecraft:stone", "default", 0, true},
		{"minecraft:chest", "default", 2, true},
		{"minecraft:torch", "default", 5, true},
		// A property that is not named keeps the value of the default
		// state. The door of 1.12.2 is registered as "wooden_door",
		// and its data value holds the direction it looks at: 0 is
		// east, 1 south, 2 west and 3 north.
		{"minecraft:wooden_door", "facing=east", 0, true},
		{"minecraft:wooden_door", "facing=north", 3, true},
		// The last pair wins.
		{"minecraft:stone", "variant=granite,variant=diorite", 3, true},
		// A number is read the way Java reads it.
		{"minecraft:wheat", "age=3", 3, true},
		{"minecraft:wheat", "age=07", 7, true},
		// A block without properties only knows "default".
		{"minecraft:cobblestone", "default", 0, true},
		{"minecraft:cobblestone", "variant=granite", 0, false},
		// A name that is not a property of the block.
		{"minecraft:stone", "color=red", 0, false},
		// The names are matched exactly.
		{"minecraft:stone", "variant=GRANITE", 0, false},
		{"minecraft:stone", "VARIANT=granite", 0, false},
		{"minecraft:stone", "DEFAULT", 0, false},
		// The value keeps the rest of the pair, so "granite=1" is not
		// a value of the variant.
		{"minecraft:stone", "variant=granite=1", 0, false},
		// "default" is only a state of the whole string.
		{"minecraft:stone", "default,default", 0, false},
		// A pair without a value.
		{"minecraft:stone", "variant", 0, false},
		// A number that is not a value of the property.
		{"minecraft:wheat", "age=9", 0, false},
		{"minecraft:wheat", "age=x", 0, false},
	}
	for _, c := range cases {
		got, ok := blockStateData(c.block, c.text)
		if ok != c.ok {
			t.Errorf("%s %q: ok = %v, want %v", c.block, c.text, ok, c.ok)
			continue
		}
		if ok && got != c.want {
			t.Errorf("%s %q: data = %d, want %d", c.block, c.text, got, c.want)
		}
	}
}

// TestBlockDefaultData checks the data value of the default state.
//
// 1.12.2 places the default state of the block when the command has no data
// value, and the data value of that state is not always 0.
//
// (Measured on a real 1.12.2 server: `setblock 10 4 10 chest` puts a chest
// with the data value 2 and `setblock 14 4 14 torch` a torch with the data
// value 5.)
func TestBlockDefaultData(t *testing.T) {
	cases := map[string]int{
		"minecraft:stone":              0,
		"minecraft:wool":               0,
		"minecraft:chest":              2,
		"minecraft:trapped_chest":      2,
		"minecraft:torch":              5,
		"minecraft:oak_stairs":         3,
		"minecraft:stone_brick_stairs": 3,
		"minecraft:cobblestone":        0,
	}
	for block, want := range cases {
		if got := blockDefaultData(block); got != want {
			t.Errorf("%s: data = %d, want %d", block, got, want)
		}
	}
}

// TestFillWithAState checks that /fill reads a state string.
//
// (Measured on a real 1.12.2 server.)
func TestFillWithAState(t *testing.T) {
	e := blockStateEngine(t)
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}

	if _, err := e.ExecuteLine("fill 2 4 2 3 4 3 stone variant=granite", 1, src); err != nil {
		t.Fatalf("fill: %v", err)
	}
	got := e.World.Blocks.Get(world.Pos{X: 2, Y: 4, Z: 2})
	if got.Name != "stone" || got.Data != 1 {
		t.Errorf("the block is %s %d, want stone 1", got.Name, got.Data)
	}
}

// TestSetBlockWithAState checks that /setblock reads a state string.
func TestSetBlockWithAState(t *testing.T) {
	e := blockStateEngine(t)
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}

	if _, err := e.ExecuteLine("setblock 8 4 8 wool color=red", 1, src); err != nil {
		t.Fatalf("setblock: %v", err)
	}
	got := e.World.Blocks.Get(world.Pos{X: 8, Y: 4, Z: 8})
	if got.Name != "wool" || got.Data != 14 {
		t.Errorf("the block is %s %d, want wool 14", got.Name, got.Data)
	}
}

// TestSetBlockWithoutADataValue checks that the default state is placed.
//
// 1.12.2 places the default state of the block when the command has no data
// value, and that state is not always the data value 0.
func TestSetBlockWithoutADataValue(t *testing.T) {
	e := blockStateEngine(t)
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}

	if _, err := e.ExecuteLine("setblock 10 4 10 chest", 1, src); err != nil {
		t.Fatalf("setblock: %v", err)
	}
	got := e.World.Blocks.Get(world.Pos{X: 10, Y: 4, Z: 10})
	if got.Name != "chest" || got.Data != 2 {
		t.Errorf("the block is %s %d, want chest 2", got.Name, got.Data)
	}
}

// TestBlockStateError checks the wording of a broken state.
//
// (Measured on a real 1.12.2 server: `fill 2 4 2 3 4 3 stone abc` answers
// "'abc' is not a state for block minecraft:stone".)
func TestBlockStateError(t *testing.T) {
	e := blockStateEngine(t)
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}

	cases := map[string]string{
		"fill 2 4 2 3 4 3 stone abc":       "'abc' is not a state for block minecraft:stone",
		"fill 2 4 2 3 4 3 stone variant=x": "'variant=x' is not a state for block minecraft:stone",
		"setblock 2 4 2 stone keep":        "'keep' is not a state for block minecraft:stone",
		"fill 2 4 2 3 4 3 stone -1":        "The number you have entered (-1) is too small, it must be at least 0",
		"fill 2 4 2 3 4 3 stone 16":        "The number you have entered (16) is too big, it must be at most 15",
	}
	for line, want := range cases {
		_, err := e.ExecuteLine(line, 1, src)
		if err == nil {
			t.Errorf("%q: expected an error", line)
			continue
		}
		if err.Error() != want {
			t.Errorf("%q: the error is %q, want %q", line, err.Error(), want)
		}
	}
}

// TestBlockStateText checks the text of a state.
//
// (Measured on a real 1.12.2 server: `testforblock 2 4 2 stone variant=diorite`
// on a granite answers "The block at 2, 4, 2 had the data value of
// minecraft:stone[variant=granite] (expected: variant=diorite)".)
func TestBlockStateText(t *testing.T) {
	cases := []struct {
		block string
		data  int
		want  string
	}{
		{"minecraft:stone", 1, "minecraft:stone[variant=granite]"},
		{"minecraft:wool", 14, "minecraft:wool[color=red]"},
		{"minecraft:cobblestone", 0, "minecraft:cobblestone"},
	}
	for _, c := range cases {
		if got := blockStateText(c.block, c.data); got != c.want {
			t.Errorf("blockStateText(%q, %d) = %q, want %q", c.block, c.data, got, c.want)
		}
	}
}

// TestReplaceBlockData checks the data value of the block to replace.
//
// (Measured on a real 1.12.2 server: `fill 4 4 4 5 4 5 stone 0 replace wool abc`
// answers "No blocks filled" -- the bad data value is only read for a
// position that holds the block to replace, and the region holds air.)
func TestReplaceBlockData(t *testing.T) {
	cases := []struct {
		block string
		text  string
		want  int
		fails bool
	}{
		{"minecraft:wool", "0", 0, false},
		{"minecraft:wool", "14", 14, false},
		{"minecraft:wool", "16", 16, false},
		{"minecraft:wool", "abc", 0, true},
		{"minecraft:stone", "variant=granite", 1, false},
		{"minecraft:cobblestone", "variant=granite", 0, true},
	}
	for _, c := range cases {
		got, err := replaceBlockData(c.block, c.text)
		if c.fails {
			if err == nil {
				t.Errorf("replaceBlockData(%q, %q): expected an error", c.block, c.text)
			}
			continue
		}
		if err != nil {
			t.Errorf("replaceBlockData(%q, %q): %v", c.block, c.text, err)
			continue
		}
		if got != c.want {
			t.Errorf("replaceBlockData(%q, %q) = %d, want %d", c.block, c.text, got, c.want)
		}
	}
}

// TestFillReplaceDataIsLazy checks that the data value of the block to
// replace is only read when the region holds that block.
//
// 1.12.2 reads it inside the loop, so a bad data value is not reported
// when nothing matches.
func TestFillReplaceDataIsLazy(t *testing.T) {
	e := blockStateEngine(t)
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}

	// The region holds air, so the bad data value is never read: the
	// command fails because nothing was filled, not because of the data
	// value.
	_, err := e.ExecuteLine("fill 4 4 4 5 4 5 stone 0 replace wool abc", 1, src)
	if err == nil || err.Error() != "No blocks filled" {
		t.Fatalf("fill: %v", err)
	}
	// The region holds a wool, so the data value is read and refused.
	e = blockStateEngine(t)
	e.World.Blocks.Set(world.Pos{X: 4, Y: 4, Z: 4}, world.BlockState{Name: "wool"})
	_, err = e.ExecuteLine("fill 4 4 4 5 4 5 stone 0 replace wool abc", 1, src)
	if err == nil {
		t.Fatal("fill: expected an error")
	}
	want := "'abc' is not a state for block minecraft:wool"
	if err.Error() != want {
		t.Errorf("the error is %q, want %q", err.Error(), want)
	}
}

// TestBlockStateTableMatchesTheDump checks the table against the dump it
// was made of.
//
// The dump is the output of the block registry of the real 1.12.2 server
// jar; see tools/blockstates. The check makes sure that the table in the
// package is the one the dump describes.
func TestBlockStateTableMatchesTheDump(t *testing.T) {
	path := filepath.Join("..", "..", "..", "tools", "blockstates", "states.jsonl")
	file, err := os.Open(path)
	if err != nil {
		t.Skipf("the dump is not there: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	seen := 0
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 || line[0] != '{' {
			continue
		}
		var dumped struct {
			Block string `json:"block"`
			Props []struct {
				Name   string   `json:"name"`
				Kind   string   `json:"kind"`
				Values []string `json:"values"`
			} `json:"props"`
			Default     []int32 `json:"default"`
			DefaultMeta int     `json:"defaultMeta"`
			Metas       []int   `json:"metas"`
		}
		if err := json.Unmarshal(line, &dumped); err != nil {
			t.Fatalf("the dump is broken: %v", err)
		}
		seen++
		table, ok := blockStateTables[dumped.Block]
		if !ok {
			t.Errorf("%s: the table does not have it", dumped.Block)
			continue
		}
		if len(table.props) != len(dumped.Props) {
			t.Errorf("%s: %d properties, want %d", dumped.Block, len(table.props), len(dumped.Props))
			continue
		}
		for i, prop := range dumped.Props {
			got := table.props[i]
			if got.name != prop.Name {
				t.Errorf("%s: property %d is %q, want %q", dumped.Block, i, got.name, prop.Name)
			}
			if len(got.values) != len(prop.Values) {
				t.Errorf("%s: property %q has %d values, want %d",
					dumped.Block, prop.Name, len(got.values), len(prop.Values))
				continue
			}
			for j, value := range prop.Values {
				if got.values[j] != value {
					t.Errorf("%s: property %q has the value %q at %d, want %q",
						dumped.Block, prop.Name, got.values[j], j, value)
				}
			}
		}
		if len(table.metas) != len(dumped.Metas) {
			t.Errorf("%s: %d data values, want %d", dumped.Block, len(table.metas), len(dumped.Metas))
			continue
		}
		for j, meta := range dumped.Metas {
			if int(table.metas[j]) != meta {
				t.Errorf("%s: the data value at %d is %d, want %d",
					dumped.Block, j, table.metas[j], meta)
			}
		}
		if got := table.defaultData(); got != dumped.DefaultMeta {
			t.Errorf("%s: the default state has the data value %d, want %d",
				dumped.Block, got, dumped.DefaultMeta)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("the dump is broken: %v", err)
	}
	if seen != len(blockStateTables) {
		t.Errorf("the dump has %d blocks, the table has %d", seen, len(blockStateTables))
	}
}
