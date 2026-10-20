package v1_12_2

import (
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/world"
)

func TestGiveAndClear(t *testing.T) {
	e := newTestEngine(t)
	p, _ := e.World.EntityTypes.NewEntity("player", "p1")
	p.PlayerName = "Steve"
	e.World.AddEntity(p)

	if err := runLine(e, "give Steve stone 3"); err != nil {
		t.Fatalf("give: %v", err)
	}
	list, ok := p.NBT.Get("Inventory").(*nbt.List)
	if !ok || list.Len() != 1 {
		t.Fatalf("the inventory holds %v, want one stack", list)
	}
	stack, _ := list.At(0).(*nbt.Compound)
	if id, _ := stack.Get("id").(nbt.String); id != "stone" {
		t.Errorf("the item is %q, want stone", id)
	}

	if err := runLine(e, "clear Steve"); err != nil {
		t.Fatalf("clear: %v", err)
	}
	list, _ = p.NBT.Get("Inventory").(*nbt.List)
	if list.Len() != 0 {
		t.Errorf("the inventory holds %d stacks, want none", list.Len())
	}
}

func TestEnchantAddsAnEnchantment(t *testing.T) {
	e := newTestEngine(t)
	p, _ := e.World.EntityTypes.NewEntity("player", "p1")
	p.PlayerName = "Steve"
	e.World.AddEntity(p)
	if err := runLine(e, "give Steve diamond_sword"); err != nil {
		t.Fatalf("give: %v", err)
	}
	if err := runLine(e, "enchant Steve sharpness 3"); err != nil {
		t.Fatalf("enchant: %v", err)
	}
	list, _ := p.NBT.Get("Inventory").(*nbt.List)
	stack, _ := list.At(0).(*nbt.Compound)
	ench, ok := stack.Get("Enchantments").(*nbt.List)
	if !ok || ench.Len() != 1 {
		t.Fatalf("the item holds %v enchantments, want one", ench)
	}
	entry, _ := ench.At(0).(*nbt.Compound)
	if lvl, _ := entry.Get("lvl").(nbt.Short); lvl != 3 {
		t.Errorf("the level is %d, want 3", lvl)
	}
}

func TestXpCommand(t *testing.T) {
	e := newTestEngine(t)
	p, _ := e.World.EntityTypes.NewEntity("player", "p1")
	p.PlayerName = "Steve"
	e.World.AddEntity(p)
	if err := runLine(e, "xp 10 Steve"); err != nil {
		t.Fatalf("xp: %v", err)
	}
	if total, _ := p.NBT.Get("XpTotal").(nbt.Int); total != 10 {
		t.Errorf("the total experience is %d, want 10", total)
	}
	if err := runLine(e, "xp 2L Steve"); err != nil {
		t.Fatalf("xp levels: %v", err)
	}
	if level, _ := p.NBT.Get("XpLevel").(nbt.Int); level != 2 {
		t.Errorf("the level is %d, want 2", level)
	}
}

func TestCloneCopiesTheRegion(t *testing.T) {
	e := newTestEngine(t)
	e.World.Blocks.Set(world.Pos{X: 0, Y: 0, Z: 0}, world.BlockState{Name: "stone"})
	if err := runLine(e, "clone 0 0 0 0 0 0 10 0 0"); err != nil {
		t.Fatalf("clone: %v", err)
	}
	if got := e.World.Blocks.Get(world.Pos{X: 10, Y: 0, Z: 0}); got.Name != "stone" {
		t.Errorf("the block at the destination is %s, want stone", got.Name)
	}
	// The source is left alone unless the move mode is used.
	if got := e.World.Blocks.Get(world.Pos{X: 0, Y: 0, Z: 0}); got.Name != "stone" {
		t.Errorf("the source block is %s, want stone", got.Name)
	}
}

func TestCloneMoveClearsTheSource(t *testing.T) {
	e := newTestEngine(t)
	e.World.Blocks.Set(world.Pos{X: 0, Y: 0, Z: 0}, world.BlockState{Name: "stone"})
	if err := runLine(e, "clone 0 0 0 0 0 0 10 0 0 normal move"); err != nil {
		t.Fatalf("clone: %v", err)
	}
	if got := e.World.Blocks.Get(world.Pos{X: 0, Y: 0, Z: 0}); !got.IsAir() {
		t.Errorf("the source block is %s, want air", got.Name)
	}
}

func TestTestForBlocks(t *testing.T) {
	e := newTestEngine(t)
	e.World.Blocks.Set(world.Pos{X: 0, Y: 0, Z: 0}, world.BlockState{Name: "stone"})
	e.World.Blocks.Set(world.Pos{X: 1, Y: 0, Z: 0}, world.BlockState{Name: "stone"})
	e.World.Blocks.Set(world.Pos{X: 10, Y: 0, Z: 0}, world.BlockState{Name: "stone"})
	e.World.Blocks.Set(world.Pos{X: 11, Y: 0, Z: 0}, world.BlockState{Name: "stone"})
	// 1.12.2 compares two regions of the same size, so the last three
	// arguments are the origin of the destination region.
	if err := runLine(e, "testforblocks 0 0 0 1 0 0 10 0 0"); err != nil {
		t.Fatalf("testforblocks: %v", err)
	}
	e.World.Blocks.Set(world.Pos{X: 11, Y: 0, Z: 0}, world.BlockState{Name: "dirt"})
	if err := runLine(e, "testforblocks 0 0 0 1 0 0 10 0 0"); err == nil {
		t.Fatal("expected the test to fail")
	}
}

func TestSetBlockModes(t *testing.T) {
	e := newTestEngine(t)
	e.World.Blocks.Set(world.Pos{X: 0, Y: 0, Z: 0}, world.BlockState{Name: "stone"})
	// The mode comes **after** the data value: 1.12.2 reads the fifth
	// argument strictly as the data value, so a mode written there is a
	// broken data value.
	//
	// (Measured on a real 1.12.2 server: `setblock 2 4 2 stone keep`
	// answers "'keep' is not a state for block minecraft:stone", while
	// `setblock 4 4 4 stone 0 keep` places the block.)
	if err := runLine(e, "setblock 0 0 0 dirt 0 keep"); err == nil {
		t.Error("expected the keep mode to refuse")
	}
	// The destroy mode clears the position first.
	if err := runLine(e, "setblock 0 0 0 dirt 0 destroy"); err != nil {
		t.Fatalf("setblock destroy: %v", err)
	}
	if got := e.World.Blocks.Get(world.Pos{X: 0, Y: 0, Z: 0}); got.Name != "dirt" {
		t.Errorf("the block is %s, want dirt", got.Name)
	}
}

// TestSetBlockDataValueIsStrict pins down that the fifth argument of
// `/setblock` is the data value and nothing else.
//
// (Measured on a real 1.12.2 server: `setblock 2 4 2 stone keep` answers
// "'keep' is not a state for block minecraft:stone".)
func TestSetBlockDataValueIsStrict(t *testing.T) {
	e := newTestEngine(t)
	for _, line := range []string{
		"setblock 0 4 0 stone keep",
		"setblock 0 4 0 stone destroy",
		"setblock 0 4 0 stone {Tags:[]}",
		"setblock 0 4 0 stone abc",
	} {
		err := runLine(e, line)
		if err == nil {
			t.Errorf("%q: expected an error", line)
			continue
		}
		if want := "is not a state for block minecraft:stone"; !strings.Contains(err.Error(), want) {
			t.Errorf("%q: the error was %q, want it to contain %q", line, err.Error(), want)
		}
	}
	// The data value itself is fine, and so is the mode after it.
	if err := runLine(e, "setblock 0 4 0 stone 0"); err != nil {
		t.Errorf("setblock with a data value: %v", err)
	}
	if err := runLine(e, "setblock 0 4 0 stone 0 replace"); err != nil {
		t.Errorf("setblock with a data value and a mode: %v", err)
	}
}

func TestFillModes(t *testing.T) {
	e := newTestEngine(t)
	// The outline mode only fills the shell. The data value is the
	// eighth argument and 1.12.2 reads it strictly, so the mode cannot
	// take its place.
	if err := runLine(e, "fill 0 0 0 2 2 2 stone 0 outline"); err != nil {
		t.Fatalf("fill outline: %v", err)
	}
	if got := e.World.Blocks.Get(world.Pos{X: 1, Y: 1, Z: 1}); !got.IsAir() {
		t.Errorf("the inside is %s, want air", got.Name)
	}
	if got := e.World.Blocks.Get(world.Pos{X: 0, Y: 0, Z: 0}); got.Name != "stone" {
		t.Errorf("the shell is %s, want stone", got.Name)
	}
}

func TestPlaySoundArgumentOrder(t *testing.T) {
	e := newTestEngine(t)
	p, _ := e.World.EntityTypes.NewEntity("player", "p1")
	p.PlayerName = "Steve"
	e.World.AddEntity(p)
	// 1.12.2 puts the target after the sound source, and the target of
	// the command is a *player*: it is looked up by its name, not by the
	// internal id the runtime gives it.
	if err := runLine(e, "playsound entity.horse.saddle master Steve"); err != nil {
		t.Fatalf("playsound: %v", err)
	}
	if err := runLine(e, "playsound entity.horse.saddle nosuchsource Steve"); err == nil {
		t.Error("expected an error for an unknown sound source")
	}
}

func TestStopSoundArgumentOrder(t *testing.T) {
	e := newTestEngine(t)
	p, _ := e.World.EntityTypes.NewEntity("player", "p1")
	p.PlayerName = "Steve"
	e.World.AddEntity(p)
	// 1.12.2 puts the target first, and the target of the command is a
	// *player*: it is looked up by its name, not by the internal id the
	// runtime gives it.
	if err := runLine(e, "stopsound Steve master entity.horse.saddle"); err != nil {
		t.Fatalf("stopsound: %v", err)
	}
}

func TestTitleCommand(t *testing.T) {
	e := newTestEngine(t)
	p, _ := e.World.EntityTypes.NewEntity("player", "p1")
	p.PlayerName = "Steve"
	e.World.AddEntity(p)
	// The target of the command is a *player*: it is looked up by its
	// name, not by the internal id the runtime gives it.
	if err := runLine(e, `title Steve title {"text":"hello"}`); err != nil {
		t.Fatalf("title: %v", err)
	}
	state, ok := p.NBT.Get("Title").(*nbt.Compound)
	if !ok {
		t.Fatal("the title state is missing")
	}
	if msg, _ := state.Get("title").(nbt.String); msg != `{"text":"hello"}` {
		t.Errorf("the title is %q, want the JSON message", msg)
	}
	if err := runLine(e, "title Steve clear"); err != nil {
		t.Fatalf("title clear: %v", err)
	}
}

func TestHelpAndList(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "help"); err != nil {
		t.Fatalf("help: %v", err)
	}
	if len(e.Feedback) == 0 {
		t.Error("help produced no output")
	}
	if err := runLine(e, "help say"); err != nil {
		t.Fatalf("help say: %v", err)
	}
	if err := runLine(e, "help nosuchcommand"); err == nil {
		t.Error("expected an error for an unknown command")
	}
	if err := runLine(e, "list"); err != nil {
		t.Fatalf("list: %v", err)
	}
}
