package v1_12_2

import (
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/nbt"
)

// TestReplaceItemSlotTable pins down the table of the slot names.
//
// 1.12.2 builds the table in the static block of `CommandReplaceItem`, and
// it is the table that decides whether a name is a slot at all.
//
// (Measured on a real 1.12.2 server, probe case 58-replaceitem-player:
// `replaceitem entity Steve slot.inventory.27` says "'slot.inventory.27'
// is not a valid parameter", and so do slot.enderchest.27,
// slot.container.54 and slot.armor.hand.)
func TestReplaceItemSlotTable(t *testing.T) {
	for _, tc := range []struct {
		name string
		want int
	}{
		{"slot.container.0", 0},
		{"slot.container.53", 53},
		{"slot.hotbar.0", 0},
		{"slot.hotbar.8", 8},
		{"slot.inventory.0", 9},
		{"slot.inventory.26", 35},
		{"slot.enderchest.0", 200},
		{"slot.enderchest.26", 226},
		{"slot.villager.0", 300},
		{"slot.villager.7", 307},
		{"slot.horse.0", 500},
		{"slot.horse.14", 514},
		{"slot.weapon", 98},
		{"slot.weapon.mainhand", 98},
		{"slot.weapon.offhand", 99},
		{"slot.armor.feet", 100},
		{"slot.armor.legs", 101},
		{"slot.armor.chest", 102},
		{"slot.armor.head", 103},
		{"slot.horse.saddle", 400},
		{"slot.horse.armor", 401},
		{"slot.horse.chest", 499},
	} {
		got, ok := lookupReplaceItemSlot(tc.name)
		if !ok {
			t.Errorf("%q is not a slot", tc.name)
			continue
		}
		if got != tc.want {
			t.Errorf("%q is %d, want %d", tc.name, got, tc.want)
		}
	}
	for _, name := range []string{
		"slot.container.54",
		"slot.inventory.27",
		"slot.enderchest.27",
		"slot.villager.8",
		"slot.horse.15",
		"slot.armor.hand",
		"slot.weapon.bow",
		"notaslot",
	} {
		if _, ok := lookupReplaceItemSlot(name); ok {
			t.Errorf("%q is a slot, but 1.12.2 does not know it", name)
		}
	}
}

// TestReplaceItemUnknownSlotName covers the name that is not in the table.
func TestReplaceItemUnknownSlotName(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	err := runLine(e, "replaceitem entity @e[type=armor_stand] slot.inventory.27 stone")
	if err == nil {
		t.Fatal("the command should have failed")
	}
	want := "'slot.inventory.27' is not a valid parameter"
	if err.Error() != want {
		t.Errorf("the command said %q, want %q", err.Error(), want)
	}
}

// TestReplaceItemSlotOfTheEntity pins down that the slot has to be one the
// entity has.
//
// (Measured on a real 1.12.2 server, probe cases 56 and 57: an armour
// stand takes slot.weapon.mainhand, slot.weapon.offhand and
// slot.armor.head, but refuses slot.armor.chest, slot.inventory.18 and
// slot.container.40 with "Could not replace slot <slot> with 1 * [Stone]".)
func TestReplaceItemSlotOfTheEntity(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}

	for _, slot := range []string{"slot.weapon.mainhand", "slot.weapon.offhand", "slot.armor.head"} {
		if err := runLine(e, "replaceitem entity @e[type=armor_stand] "+slot+" stone"); err != nil {
			t.Errorf("%s: %v", slot, err)
		}
	}
	for _, slot := range []string{"slot.armor.chest", "slot.inventory.18", "slot.container.40"} {
		err := runLine(e, "replaceitem entity @e[type=armor_stand] "+slot+" stone")
		if err == nil {
			t.Errorf("%s should have been refused", slot)
			continue
		}
		want := "Could not replace slot " + slot + " with 1 * [Stone]"
		if err.Error() != want {
			t.Errorf("%s said %q, want %q", slot, err.Error(), want)
		}
	}
}

// TestReplaceItemWritesTheItem checks that the item lands in the slot.
func TestReplaceItemWritesTheItem(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	if err := runLine(e, "replaceitem entity @e[type=armor_stand] slot.weapon.mainhand stone 5"); err != nil {
		t.Fatalf("replaceitem: %v", err)
	}
	list, ok := e.World.EntitySlice()[0].NBTValue("Inventory").(*nbt.List)
	if !ok {
		t.Fatal("the entity has no inventory")
	}
	items := list.Items()
	// The main hand is the slot 98, so the list reaches that far.
	if len(items) != 99 {
		t.Fatalf("the inventory has %d items, want 99", len(items))
	}
	stack := items[98].(*nbt.Compound)
	id, _ := stack.Get("id").(nbt.String)
	if string(id) != "stone" {
		t.Errorf("the item is %q, want stone", string(id))
	}
	count, _ := stack.Get("Count").(nbt.Byte)
	if byte(count) != 5 {
		t.Errorf("the count is %d, want 5", byte(count))
	}
}

// TestReplaceItemPlayerSlots covers the player, whose container is the
// largest one.
//
// (Measured on a real 1.12.2 server, probe case 56-replaceitem-slots: a
// player takes slot.inventory.18, slot.inventory.26, slot.armor.head,
// slot.armor.chest, slot.weapon.offhand and slot.enderchest.0, all of
// which an armour stand refuses.)
func TestReplaceItemPlayerSlots(t *testing.T) {
	e := newTestEngine(t)
	addPlayer(t, e, "Steve")

	for _, slot := range []string{
		"slot.inventory.18",
		"slot.inventory.26",
		"slot.armor.head",
		"slot.armor.chest",
		"slot.weapon.offhand",
		"slot.enderchest.0",
	} {
		if err := runLine(e, "replaceitem entity Steve "+slot+" stone"); err != nil {
			t.Errorf("%s: %v", slot, err)
		}
	}
}

// TestReplaceItemLivingEntitySlots covers a living entity, which has a
// smaller container than a player.
//
// (Measured on a real 1.12.2 server, probe case 57-replaceitem-containers:
// a zombie takes slot.weapon.mainhand and slot.armor.head, and refuses
// slot.weapon.offhand and slot.armor.chest.)
func TestReplaceItemLivingEntitySlots(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon zombie 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}

	for _, slot := range []string{"slot.weapon.mainhand", "slot.armor.head"} {
		if err := runLine(e, "replaceitem entity @e[type=zombie] "+slot+" stone"); err != nil {
			t.Errorf("%s: %v", slot, err)
		}
	}
	for _, slot := range []string{"slot.weapon.offhand", "slot.armor.chest"} {
		err := runLine(e, "replaceitem entity @e[type=zombie] "+slot+" stone")
		if err == nil {
			t.Errorf("%s should have been refused", slot)
			continue
		}
		if !strings.HasPrefix(err.Error(), "Could not replace slot "+slot) {
			t.Errorf("%s said %q", slot, err.Error())
		}
	}
}
