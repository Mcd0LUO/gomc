package v1_12_2

import (
	"fmt"
	"strconv"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// The slots of `/replaceitem entity`.
//
// 1.12.2 maps a slot name to a number with a table that is built once, when
// the class is loaded (`CommandReplaceItem`'s static block). The numbers
// are **indices into the container of the entity**, and every kind of
// entity has a container of its own: a player has 36 slots of inventory,
// four of armour, one in the off hand and 27 in the ender chest, while an
// armour stand has two in its hands and four of armour, and a horse has
// its saddle and its own chest.
//
// A slot the entity does not have cannot be replaced.
//
// (Measured on a real 1.12.2 server, probe cases 56-replaceitem-slots and
// 57-replaceitem-containers: an armour stand takes slot.weapon.mainhand,
// slot.weapon.offhand and slot.armor.head but refuses slot.armor.chest; a
// zombie takes slot.weapon.mainhand and slot.armor.head; a player takes
// all of them.)
const (
	// slotMainhand is the number of the slot in the main hand.
	slotMainhand = 98
	// slotOffhand is the number of the slot in the off hand.
	slotOffhand = 99
	// slotArmorFeet is the number of the slot of the boots.
	slotArmorFeet = 100
	// slotArmorLegs is the number of the slot of the leggings.
	slotArmorLegs = 101
	// slotArmorChest is the number of the slot of the chestplate.
	slotArmorChest = 102
	// slotArmorHead is the number of the slot of the helmet.
	slotArmorHead = 103
	// slotEnderChest is the number of the first slot of the ender chest.
	slotEnderChest = 200
	// slotVillager is the number of the first slot of a villager.
	slotVillager = 300
	// slotHorseSaddle is the number of the slot of the saddle.
	slotHorseSaddle = 400
	// slotHorseArmor is the number of the slot of the horse armour.
	slotHorseArmor = 401
	// slotHorseChest is the number of the slot of the horse chest.
	slotHorseChest = 499
	// slotHorse is the number of the first slot of a horse.
	slotHorse = 500
)

// replaceItemSlots is the table 1.12.2 builds in the static block of
// `CommandReplaceItem`: it maps every slot name to the number the
// command uses.
//
// The armour slots are `100 + EnumItemSlot.b()`, and `EnumItemSlot.b()`
// is the place of the constant in the enum: FEET is 0, LEGS is 1, CHEST is
// 2 and HEAD is 3.
var replaceItemSlots = buildReplaceItemSlots()

// buildReplaceItemSlots builds the table of the slot names.
//
// The table is built instead of written out, because the ranges are long
// and the numbers of the armour slots come from the enum.
func buildReplaceItemSlots() map[string]int {
	slots := map[string]int{}
	// The container of a block: 54 slots.
	for i := 0; i < 54; i++ {
		slots[fmt.Sprintf("slot.container.%d", i)] = i
	}
	// The hot bar: 9 slots.
	for i := 0; i < 9; i++ {
		slots[fmt.Sprintf("slot.hotbar.%d", i)] = i
	}
	// The inventory: 27 slots, behind the hot bar.
	for i := 0; i < 27; i++ {
		slots[fmt.Sprintf("slot.inventory.%d", i)] = 9 + i
	}
	// The ender chest: 27 slots.
	for i := 0; i < 27; i++ {
		slots[fmt.Sprintf("slot.enderchest.%d", i)] = slotEnderChest + i
	}
	// The trade slots of a villager: 8 slots.
	for i := 0; i < 8; i++ {
		slots[fmt.Sprintf("slot.villager.%d", i)] = slotVillager + i
	}
	// The slots of a horse: 15 slots.
	for i := 0; i < 15; i++ {
		slots[fmt.Sprintf("slot.horse.%d", i)] = slotHorse + i
	}
	slots["slot.weapon"] = slotMainhand
	slots["slot.weapon.mainhand"] = slotMainhand
	slots["slot.weapon.offhand"] = slotOffhand
	slots["slot.armor.feet"] = slotArmorFeet
	slots["slot.armor.legs"] = slotArmorLegs
	slots["slot.armor.chest"] = slotArmorChest
	slots["slot.armor.head"] = slotArmorHead
	slots["slot.horse.saddle"] = slotHorseSaddle
	slots["slot.horse.armor"] = slotHorseArmor
	slots["slot.horse.chest"] = slotHorseChest
	return slots
}

// lookupReplaceItemSlot maps a slot name to the number 1.12.2 uses.
//
// A name that is not in the table is refused with
// `commands.generic.parameter.invalid`, which is the language key
// "commands.generic.parameter.invalid" -- the message reads
// "'<name>' is not a valid parameter".
//
// (Measured on a real 1.12.2 server, probe case 58-replaceitem-player:
// `replaceitem entity Steve slot.inventory.27` says "'slot.inventory.27'
// is not a valid parameter", and so do slot.enderchest.27,
// slot.container.54 and slot.armor.hand.)
func lookupReplaceItemSlot(name string) (int, bool) {
	if n, ok := replaceItemSlots[name]; ok {
		return n, true
	}
	// A plain number is a slot of the container as well, and 1.12.2
	// takes it as it is.
	if n, err := strconv.Atoi(name); err == nil && n >= 0 {
		return n, true
	}
	return 0, false
}

// The parts of a container 1.12.2 gives an entity.
//
// A container is a set of ranges of slot numbers. A slot that is not in
// any of the ranges cannot be replaced on that entity.
type containerRanges struct {
	// main is the range of the slots of the inventory, and the range of
	// the hot bar is its first part.
	main []slotRange
	// armor holds the slots of the armour.
	armor []slotRange
	// offhand is the slot of the off hand, or -1 when the entity has
	// none.
	offhand int
	// extra holds the slots that are neither the inventory nor the
	// armour: the ender chest of a player, the saddle of a horse.
	extra []slotRange
}

// slotRange is a range of slot numbers, both ends included.
type slotRange struct {
	// from is the first slot of the range.
	from int
	// to is the last slot of the range.
	to int
}

// contains reports whether the range holds the slot.
func (r slotRange) contains(slot int) bool {
	return slot >= r.from && slot <= r.to
}

// has reports whether the container has the slot.
func (c containerRanges) has(slot int) bool {
	if c.offhand >= 0 && slot == c.offhand {
		return true
	}
	for _, r := range c.main {
		if r.contains(slot) {
			return true
		}
	}
	for _, r := range c.armor {
		if r.contains(slot) {
			return true
		}
	}
	for _, r := range c.extra {
		if r.contains(slot) {
			return true
		}
	}
	return false
}

// The containers of the kinds of entity 1.12.2 knows.
//
// A player has a full inventory, four slots of armour, a slot in the off
// hand and the ender chest. A living entity that is not a player has a
// small inventory, a slot in the main hand and the four slots of the
// armour, but only the helmet of the four is one it can wear. An armour
// stand has two slots in its hands and four of armour, of which the
// helmet and the leggings and the boots are the ones it wears.
//
// (Measured on a real 1.12.2 server, probe cases 56 and 57.)
var (
	// playerContainer is the container of a player.
	playerContainer = containerRanges{
		main:    []slotRange{{0, 35}},
		armor:   []slotRange{{slotArmorFeet, slotArmorHead}},
		offhand: slotOffhand,
		extra:   []slotRange{{slotEnderChest, slotEnderChest + 26}},
	}
	// livingContainer is the container of a living entity that is not
	// a player: the slot in its main hand and the four slots of the
	// armour, of which it can wear the helmet.
	livingContainer = containerRanges{
		main:    []slotRange{{slotMainhand, slotMainhand}},
		armor:   []slotRange{{slotArmorHead, slotArmorHead}},
		offhand: -1,
	}
	// armorStandContainer is the container of an armour stand.
	armorStandContainer = containerRanges{
		main:    []slotRange{{slotMainhand, slotOffhand}},
		armor:   []slotRange{{slotArmorFeet, slotArmorLegs}, {slotArmorHead, slotArmorHead}},
		offhand: slotOffhand,
	}
	// horseContainer is the container of a horse: the saddle, the
	// armour and the chest.
	horseContainer = containerRanges{
		main:    []slotRange{{slotHorseSaddle, slotHorseArmor}, {slotHorseChest, slotHorseChest}},
		armor:   []slotRange{},
		offhand: -1,
		extra:   []slotRange{{slotHorse, slotHorse + 14}},
	}
	// villagerContainer is the container of a villager.
	villagerContainer = containerRanges{
		main:    []slotRange{{0, 8}},
		armor:   []slotRange{},
		offhand: -1,
		extra:   []slotRange{{slotVillager, slotVillager + 7}},
	}
	// noContainer is the container of an entity that has none.
	noContainer = containerRanges{offhand: -1}
)

// containerOf returns the container of an entity.
//
// 1.12.2 asks the entity for its container, and every kind of entity has
// its own.
func containerOf(e *world.Entity) containerRanges {
	switch {
	case e.IsPlayer:
		return playerContainer
	case e.Type == "armor_stand":
		return armorStandContainer
	case e.Type == "horse":
		return horseContainer
	case e.Type == "villager":
		return villagerContainer
	case e.Type == "zombie" || e.Type == "skeleton" || e.Type == "creeper" ||
		e.Type == "spider" || e.Type == "pig" || e.Type == "cow" ||
		e.Type == "sheep" || e.Type == "chicken" || e.Type == "enderman" ||
		e.Type == "endermite" || e.Type == "slime" || e.Type == "witch" ||
		e.Type == "blaze" || e.Type == "ghast" || e.Type == "magma_cube":
		return livingContainer
	}
	return noContainer
}

// replaceItemInSlot puts an item into a slot of an entity.
//
// The item goes into the list of the inventory of the entity, at the
// place the slot number names. The armour and the other containers of an
// entity are not modelled separately, so they share the list of the
// inventory.
func replaceItemInSlot(e *world.Entity, slot int, item *astItem) {
	list, _ := e.NBTValue("Inventory").(*nbt.List)
	if list == nil {
		list = nbt.NewList(nbt.TypeCompound)
	}
	values := list.Items()
	for len(values) <= slot {
		list.Append(nbt.NewCompound())
		values = list.Items()
	}
	stack := nbt.NewCompound()
	stack.Set("id", nbt.String(item.name))
	stack.Set("Count", nbt.Byte(byte(item.count)))
	values[slot].(*nbt.Compound).Set("id", nbt.String(item.name))
	values[slot].(*nbt.Compound).Set("Count", nbt.Byte(byte(item.count)))
	e.NBT.Set("Inventory", list)
}

// astItem is the item a `/replaceitem` puts into a slot.
type astItem struct {
	// name is the registry name of the item.
	name string
	// count is how many of them are put into the slot.
	count int
	// data is the data value of the stack.
	data int
}

// replaceItemError builds the error 1.12.2 answers a slot it cannot
// replace with.
//
// The message is `commands.replaceitem.failed`, which reads "Could not
// replace slot %s with %s * [%s]".
//
// (Measured on a real 1.12.2 server, probe case 56-replaceitem-slots:
// `replaceitem entity @e[type=armor_stand] slot.armor.chest stone` says
// "Could not replace slot slot.armor.chest with 1 * [Stone]".)
func replaceItemError(slot string, item astItem, display string) error {
	return exec.Errorf("Could not replace slot %s with %d * [%s]", slot, item.count, display)
}

// replaceItemDisplayName returns the name 1.12.2 shows an item by.
//
// The name is the one of the item, and the server wraps it in brackets.
func replaceItemDisplayName(item astItem) string {
	if name, ok := itemDisplayName(item.name, item.data); ok {
		return name
	}
	return item.name
}
