package v1_12_2

import (
	"testing"

	"github.com/Mcd0LUO/gomc/internal/nbt"
)

// TestDataTagMatchesAnInventoryWithSeveralItems covers the dataTag of the
// scoreboard against an entity that holds more than one item.
//
// 1.12.2 compares the dataTag with the NBT of the entity with
// `GameProfileSerializer.a`, and its rule for a **list** is: every element
// of the filter has to match *some* element of the subject. The two lists
// need not have the same length.
//
// So a dataTag that names one item matches an inventory of several, as
// long as one of them is the item the dataTag names.
func TestDataTagMatchesAnInventoryWithSeveralItems(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "scoreboard objectives add kills dummy"); err != nil {
		t.Fatalf("scoreboard objectives add: %v", err)
	}
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	stand := e.World.EntityList()[0]
	list := nbt.NewList(nbt.TypeCompound)
	first := nbt.NewCompound()
	first.Set("id", nbt.String("minecraft:stone"))
	first.Set("Count", nbt.Byte(1))
	list.Append(first)
	second := nbt.NewCompound()
	second.Set("id", nbt.String("minecraft:dirt"))
	second.Set("Count", nbt.Byte(2))
	list.Append(second)
	stand.NBT.Set("Inventory", list)

	// The dataTag names one of them: it matches.
	if err := runLine(e, `scoreboard players set @e[type=armor_stand] kills 1 {Inventory:[{id:"minecraft:stone",Count:1b}]}`); err != nil {
		t.Fatalf("scoreboard players set: %v", err)
	}
	// A dataTag that names an item the inventory does not hold does not
	// match.
	err := runLine(e, `scoreboard players set @e[type=armor_stand] kills 2 {Inventory:[{id:"minecraft:diamond",Count:1b}]}`)
	if err == nil {
		t.Fatal("expected an error for a dataTag that does not match")
	}
}

// TestDataTagWithSeveralItemsMatchesAllOfThem covers the other direction: a
// dataTag that names several items matches only when the entity holds all
// of them.
func TestDataTagWithSeveralItemsMatchesAllOfThem(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "scoreboard objectives add kills dummy"); err != nil {
		t.Fatalf("scoreboard objectives add: %v", err)
	}
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	stand := e.World.EntityList()[0]
	list := nbt.NewList(nbt.TypeCompound)
	first := nbt.NewCompound()
	first.Set("id", nbt.String("minecraft:stone"))
	first.Set("Count", nbt.Byte(1))
	list.Append(first)
	second := nbt.NewCompound()
	second.Set("id", nbt.String("minecraft:dirt"))
	second.Set("Count", nbt.Byte(2))
	list.Append(second)
	stand.NBT.Set("Inventory", list)

	line := `scoreboard players set @e[type=armor_stand] kills 1 ` +
		`{Inventory:[{id:"minecraft:stone",Count:1b},{id:"minecraft:dirt",Count:2b}]}`
	if err := runLine(e, line); err != nil {
		t.Fatalf("scoreboard players set: %v", err)
	}
	// One of the two is missing from the inventory, so it does not match.
	line = `scoreboard players set @e[type=armor_stand] kills 2 ` +
		`{Inventory:[{id:"minecraft:stone",Count:1b},{id:"minecraft:gold",Count:1b}]}`
	if err := runLine(e, line); err == nil {
		t.Fatal("expected an error for a dataTag that names an item the inventory does not hold")
	}
}

// TestDataTagOfAnEmptyListMatchesAnEmptyInventory covers the rule for an
// empty filter: it matches only an empty list.
func TestDataTagOfAnEmptyListMatchesAnEmptyInventory(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "scoreboard objectives add kills dummy"); err != nil {
		t.Fatalf("scoreboard objectives add: %v", err)
	}
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	stand := e.World.EntityList()[0]
	stand.NBT.Set("Inventory", nbt.NewList(nbt.TypeCompound))

	// The inventory is empty, so an empty filter matches it.
	if err := runLine(e, `scoreboard players set @e[type=armor_stand] kills 1 {Inventory:[]}`); err != nil {
		t.Fatalf("scoreboard players set: %v", err)
	}
	// A filter with an item does not match an empty inventory.
	err := runLine(e, `scoreboard players set @e[type=armor_stand] kills 2 {Inventory:[{id:"minecraft:stone",Count:1b}]}`)
	if err == nil {
		t.Fatal("expected an error for a dataTag that does not match")
	}
}
