package v1_12_2

import "strings"

// The display names of the items of 1.12.2.
//
// A name is the name of the language file of the server: an item is
// named after the key "<unlocalized name>.name". The unlocalized name
// of an item that has more than one form carries the form, so the dye
// of the data value 3 is "item.dyePowder.brown.name" -- "Cocoa Beans",
// and the banner of the data value 0 is "item.banner.black.name" --
// "Black Banner".
//
// (Measured on a real Paper 1.12.2: `give <player> minecraft:dye 1 3`
// answers "Given [Cocoa Beans] * 1 to <player>" and `give <player>
// minecraft:banner 1 0` answers "Given [Black Banner] * 1 to
// <player>".)

// itemDisplayName returns the name 1.12.2 shows for an item stack, and
// whether the server knows the item at all.
//
// The item is the id of the item, with or without a namespace: "dye" and
// "minecraft:dye" are the same item, and an item of another namespace
// is an item the server does not know.
//
// The data value is the data value of the stack. A data value of -1
// means "not given", and the server uses the data value 0 then. A data
// value outside 0..15 gives the name of the data value 0.
//
// (Measured on a real Paper 1.12.2: `give <player> minecraft:dye 1` and
// `give <player> minecraft:dye 1 0` both answer "Given [Ink Sac] * 1 to
// <player>", and `give <player> minecraft:dye 1 16` and `give <player>
// minecraft:dye 1 100` both answer "Given [Ink Sac] * 1 to <player>"
// too -- the name of the data value 0.)
func itemDisplayName(item string, data int) (string, bool) {
	key, ok := minecraftItemKey(item)
	if !ok {
		return "", false
	}
	if names, ok := itemFormNames[key]; ok {
		if data < 0 || data > 15 {
			data = 0
		}
		return names[data], true
	}
	if names, ok := itemEffectNames[key]; ok {
		return itemEffectName(names, data), true
	}
	name, ok := customItemNames[key]
	return name, ok
}

// itemEffectName returns the name of an item whose name says which
// effect it has.
//
// The effect of a plain potion is in the data value of the stack: the
// data value 0 is the water one, the data value 1 the regeneration one
// and so on. A data value outside 0..15 is read as the effect of the
// data value and 15, and the data values 16, 32, 64, 80 and 96 are the
// "awkward", the "thick" and the "mundane" potion.
//
// (Measured on a real Paper 1.12.2: `give <player> minecraft:potion 1 1`
// answers "Given [Potion of Regeneration] * 1 to <player>", `give
// <player> minecraft:potion 1 16` answers "Given [Awkward Potion] * 1 to
// <player>" and `give <player> minecraft:potion 1 100` answers "Given
// [Potion of Poison] * 1 to <player>" -- the name of the data value 4.)
func itemEffectName(names []string, data int) string {
	switch data {
	case 16, 80:
		return "Awkward Potion"
	case 32, 96:
		return "Thick Potion"
	case 64:
		return "Mundane Potion"
	}
	return names[data&15]
}

// itemEffectNames are the items whose name says which effect they have.
//
// The name of a potion says which effect it has, and the effect of a
// plain potion is in the data value of the stack: the data value 0 is
// the water one, the data value 1 the regeneration one and so on. The
// splash potion, the lingering potion and the tipped arrow have only
// the water effect in 1.12.2, so the name of those does not follow the
// data value.
//
// (Measured on a real Paper 1.12.2: `give <player> minecraft:potion 1 1`
// answers "Given [Potion of Regeneration] * 1 to <player>" and `give
// <player> minecraft:splash_potion 1 1` answers "Given [Splash Water
// Bottle] * 1 to <player>".)
var itemEffectNames = map[string][]string{
	"potion": {"Water Bottle", "Potion of Regeneration", "Potion of Swiftness",
		"Potion of Fire Resistance", "Potion of Poison", "Potion of Healing",
		"Potion of Night Vision", "Water Bottle", "Potion of Weakness",
		"Potion of Strength", "Potion of Slowness", "Potion of Leaping",
		"Potion of Harming", "Potion of Water Breathing", "Potion of Invisibility",
		"Water Bottle"},
}

// recipeItemName returns the name the server calls a recipe that
// produces the item, and whether the server knows the item at all.
//
// It is the name of the item, except for the potions: the name of a
// potion says which effect it has, and the effect is in the NBT of the
// stack. The server builds the stack of a recipe before it starts, so
// the stack has no NBT then and the potion is the uncraftable one.
//
// (Measured on a real Paper 1.12.2: a recipe whose result is
// "minecraft:potion" is answered with "the recipe for Uncraftable
// Potion", while `give <player> minecraft:potion 1` answers "Given
// [Water Bottle] * 1 to <player>".)
func recipeItemName(item string, data int) (string, bool) {
	key, ok := minecraftItemKey(item)
	if !ok {
		return "", false
	}
	if name, ok := itemEmptyNames[key]; ok {
		return name, true
	}
	return itemDisplayName(item, data)
}

// itemEmptyNames are the names of the potions that have no effect.
//
// A potion is named after its effect, and the effect is in the NBT of
// the stack. A stack without that NBT is the uncraftable one.
//
// (Measured on a real Paper 1.12.2: the server answers "the recipe for
// Uncraftable Potion" for a recipe whose result is a potion, and "the
// recipe for Splash Uncraftable Potion" for one whose result is a
// splash potion.)
var itemEmptyNames = map[string]string{
	"potion":           "Uncraftable Potion",
	"splash_potion":    "Splash Uncraftable Potion",
	"lingering_potion": "Lingering Uncraftable Potion",
	"tipped_arrow":     "Uncraftable Tipped Arrow",
}

// minecraftItemKey turns the id of an item into the key of the item
// registry of the server.
//
// 1.12.2 knows the items of the "minecraft" namespace only, so an item
// of another namespace is an item the server does not know. The name of
// the item is lower-cased, the way a `MinecraftKey` is read.
func minecraftItemKey(item string) (string, bool) {
	if item == "" {
		return "", false
	}
	namespace, name := "minecraft", item
	if i := strings.Index(item, ":"); i >= 0 {
		namespace, name = item[:i], item[i+1:]
	}
	if strings.ToLower(namespace) != "minecraft" || name == "" {
		return "", false
	}
	return strings.ToLower(name), true
}

// itemHasForms reports whether 1.12.2 knows more than one form of the
// item.
//
// The server refuses a recipe whose result is such an item when the
// file does not say which form it is, so the data value of the result
// is then missing.
//
// (Measured on a real Paper 1.12.2: the server logs "Missing data for
// item 'minecraft:dye'" for a recipe file whose result is a dye and
// that has no "data", and the recipe is not a recipe of the server.)
func itemHasForms(item string) bool {
	_, ok := itemFormNames[item]
	return ok
}

// itemFormNames are the items 1.12.2 knows more than one form of: the
// name of such an item depends on the data value of the stack.
//
// The value holds the names of the forms, in the order of the data
// values 0 to 15. The name of the data value 0 is the one of the table
// of the items as well.
//
// The names come from the language file of the server. The order of the
// data values is not the same for every item: the dye and the banner
// are numbered by the dye data value, so the data value 0 of the dye is
// the black one -- "Ink Sac" -- and the data value 0 of the banner is
// the black one as well -- "Black Banner"; the bed and the wool are
// numbered by the color, so their data value 0 is the white one.
//
// (Measured on a real Paper 1.12.2: `recipe give <player> <name>` of a
// recipe whose result is a dye of the data value 3 answers "the recipe
// for Cocoa Beans", one whose result is a banner of the data value 14
// answers "the recipe for Orange Banner", one whose result is a bed of
// the data value 4 answers "the recipe for Yellow Bed" and one whose
// result is a skull of the data value 2 answers "the recipe for Zombie
// Head".)
var itemFormNames = map[string][]string{
	"anvil": {"Anvil", "Slightly Damaged Anvil", "Very Damaged Anvil", "Anvil",
		"Anvil", "Anvil", "Anvil", "Anvil",
		"Anvil", "Anvil", "Anvil", "Anvil",
		"Anvil", "Anvil", "Anvil", "Anvil"},
	"banner": {"Black Banner", "Red Banner", "Green Banner", "Brown Banner",
		"Blue Banner", "Purple Banner", "Cyan Banner", "Light Gray Banner",
		"Gray Banner", "Pink Banner", "Lime Banner", "Yellow Banner",
		"Light Blue Banner", "Magenta Banner", "Orange Banner", "White Banner"},
	"bed": {"White Bed", "Orange Bed", "Magenta Bed", "Light Blue Bed",
		"Yellow Bed", "Lime Bed", "Pink Bed", "Gray Bed",
		"Light Gray Bed", "Cyan Bed", "Purple Bed", "Blue Bed",
		"Brown Bed", "Green Bed", "Red Bed", "Black Bed"},
	"carpet": {"White Carpet", "Orange Carpet", "Magenta Carpet", "Light Blue Carpet",
		"Yellow Carpet", "Lime Carpet", "Pink Carpet", "Gray Carpet",
		"Light Gray Carpet", "Cyan Carpet", "Purple Carpet", "Blue Carpet",
		"Brown Carpet", "Green Carpet", "Red Carpet", "Black Carpet"},
	"coal": {"Coal", "Charcoal", "Coal", "Coal",
		"Coal", "Coal", "Coal", "Coal",
		"Coal", "Coal", "Coal", "Coal",
		"Coal", "Coal", "Coal", "Coal"},
	"concrete": {"White Concrete", "Orange Concrete", "Magenta Concrete", "Light Blue Concrete",
		"Yellow Concrete", "Lime Concrete", "Pink Concrete", "Gray Concrete",
		"Light Gray Concrete", "Cyan Concrete", "Purple Concrete", "Blue Concrete",
		"Brown Concrete", "Green Concrete", "Red Concrete", "Black Concrete"},
	"concrete_powder": {"White Concrete Powder", "Orange Concrete Powder", "Magenta Concrete Powder", "Light Blue Concrete Powder",
		"Yellow Concrete Powder", "Lime Concrete Powder", "Pink Concrete Powder", "Gray Concrete Powder",
		"Light Gray Concrete Powder", "Cyan Concrete Powder", "Purple Concrete Powder", "Blue Concrete Powder",
		"Brown Concrete Powder", "Green Concrete Powder", "Red Concrete Powder", "Black Concrete Powder"},
	"cooked_fish": {"Cooked Fish", "Cooked Salmon", "Clownfish", "Pufferfish",
		"Cooked Fish", "Cooked Fish", "Cooked Fish", "Cooked Fish",
		"Cooked Fish", "Cooked Fish", "Cooked Fish", "Cooked Fish",
		"Cooked Fish", "Cooked Fish", "Cooked Fish", "Cooked Fish"},
	"cobblestone_wall": {"Cobblestone Wall", "Mossy Cobblestone Wall", "Cobblestone Wall", "Cobblestone Wall",
		"Cobblestone Wall", "Cobblestone Wall", "Cobblestone Wall", "Cobblestone Wall",
		"Cobblestone Wall", "Cobblestone Wall", "Cobblestone Wall", "Cobblestone Wall",
		"Cobblestone Wall", "Cobblestone Wall", "Cobblestone Wall", "Cobblestone Wall"},
	"dirt": {"Dirt", "Coarse Dirt", "Podzol", "Dirt",
		"Dirt", "Dirt", "Dirt", "Dirt",
		"Dirt", "Dirt", "Dirt", "Dirt",
		"Dirt", "Dirt", "Dirt", "Dirt"},
	"double_plant": {"Sunflower", "Lilac", "Double Tallgrass", "Large Fern",
		"Rose Bush", "Peony", "Sunflower", "Sunflower",
		"Sunflower", "Sunflower", "Sunflower", "Sunflower",
		"Sunflower", "Sunflower", "Sunflower", "Sunflower"},
	"dye": {"Ink Sac", "Rose Red", "Cactus Green", "Cocoa Beans",
		"Lapis Lazuli", "Purple Dye", "Cyan Dye", "Light Gray Dye",
		"Gray Dye", "Pink Dye", "Lime Dye", "Dandelion Yellow",
		"Light Blue Dye", "Magenta Dye", "Orange Dye", "Bone Meal"},
	"fish": {"Raw Fish", "Raw Salmon", "Clownfish", "Pufferfish",
		"Raw Fish", "Raw Fish", "Raw Fish", "Raw Fish",
		"Raw Fish", "Raw Fish", "Raw Fish", "Raw Fish",
		"Raw Fish", "Raw Fish", "Raw Fish", "Raw Fish"},
	"filled_map": {"Map", "Map", "Map", "Map",
		"Map", "Map", "Map", "Map",
		"Map", "Map", "Map", "Map",
		"Map", "Map", "Map", "Map"},
	"golden_apple": {"Golden Apple", "Golden Apple", "Golden Apple", "Golden Apple",
		"Golden Apple", "Golden Apple", "Golden Apple", "Golden Apple",
		"Golden Apple", "Golden Apple", "Golden Apple", "Golden Apple",
		"Golden Apple", "Golden Apple", "Golden Apple", "Golden Apple"},
	"leaves": {"Oak Leaves", "Spruce Leaves", "Birch Leaves", "Jungle Leaves",
		"Oak Leaves", "Spruce Leaves", "Birch Leaves", "Jungle Leaves",
		"Oak Leaves", "Spruce Leaves", "Birch Leaves", "Jungle Leaves",
		"Oak Leaves", "Spruce Leaves", "Birch Leaves", "Jungle Leaves"},
	"leaves2": {"Acacia Leaves", "Dark Oak Leaves", "Oak Leaves", "Oak Leaves",
		"Acacia Leaves", "Dark Oak Leaves", "Oak Leaves", "Oak Leaves",
		"Acacia Leaves", "Dark Oak Leaves", "Oak Leaves", "Oak Leaves",
		"Acacia Leaves", "Dark Oak Leaves", "Oak Leaves", "Oak Leaves"},
	"log": {"Oak Wood", "Spruce Wood", "Birch Wood", "Jungle Wood",
		"Acacia Wood", "Dark Oak Wood", "Oak Wood", "Oak Wood",
		"Oak Wood", "Oak Wood", "Oak Wood", "Oak Wood",
		"Oak Wood", "Oak Wood", "Oak Wood", "Oak Wood"},
	"log2": {"Acacia Wood", "Dark Oak Wood", "Oak Wood", "Oak Wood",
		"Oak Wood", "Oak Wood", "Oak Wood", "Oak Wood",
		"Oak Wood", "Oak Wood", "Oak Wood", "Oak Wood",
		"Oak Wood", "Oak Wood", "Oak Wood", "Oak Wood"},
	"monster_egg": {"Stone Monster Egg", "Cobblestone Monster Egg", "Stone Brick Monster Egg", "Mossy Stone Brick Monster Egg",
		"Cracked Stone Brick Monster Egg", "Chiseled Stone Brick Monster Egg", "Stone Monster Egg", "Stone Monster Egg",
		"Stone Monster Egg", "Stone Monster Egg", "Stone Monster Egg", "Stone Monster Egg",
		"Stone Monster Egg", "Stone Monster Egg", "Stone Monster Egg", "Stone Monster Egg"},
	"planks": {"Oak Wood Planks", "Spruce Wood Planks", "Birch Wood Planks", "Jungle Wood Planks",
		"Acacia Wood Planks", "Dark Oak Wood Planks", "Oak Wood Planks", "Oak Wood Planks",
		"Oak Wood Planks", "Oak Wood Planks", "Oak Wood Planks", "Oak Wood Planks",
		"Oak Wood Planks", "Oak Wood Planks", "Oak Wood Planks", "Oak Wood Planks"},
	"prismarine": {"Prismarine", "Prismarine Bricks", "Dark Prismarine", "Prismarine",
		"Prismarine", "Prismarine", "Prismarine", "Prismarine",
		"Prismarine", "Prismarine", "Prismarine", "Prismarine",
		"Prismarine", "Prismarine", "Prismarine", "Prismarine"},
	"purpur_slab": {"Purpur Slab", "Purpur Slab", "Purpur Slab", "Purpur Slab",
		"Purpur Slab", "Purpur Slab", "Purpur Slab", "Purpur Slab",
		"Purpur Slab", "Purpur Slab", "Purpur Slab", "Purpur Slab",
		"Purpur Slab", "Purpur Slab", "Purpur Slab", "Purpur Slab"},
	"quartz_block": {"Block of Quartz", "Chiseled Quartz Block", "Pillar Quartz Block", "Block of Quartz",
		"Block of Quartz", "Block of Quartz", "Block of Quartz", "Block of Quartz",
		"Block of Quartz", "Block of Quartz", "Block of Quartz", "Block of Quartz",
		"Block of Quartz", "Block of Quartz", "Block of Quartz", "Block of Quartz"},
	"red_flower": {"Poppy", "Blue Orchid", "Allium", "Azure Bluet",
		"Red Tulip", "Orange Tulip", "White Tulip", "Pink Tulip",
		"Oxeye Daisy", "Poppy", "Poppy", "Poppy",
		"Poppy", "Poppy", "Poppy", "Poppy"},
	"red_sandstone": {"Red Sandstone", "Chiseled Red Sandstone", "Smooth Red Sandstone", "Red Sandstone",
		"Red Sandstone", "Red Sandstone", "Red Sandstone", "Red Sandstone",
		"Red Sandstone", "Red Sandstone", "Red Sandstone", "Red Sandstone",
		"Red Sandstone", "Red Sandstone", "Red Sandstone", "Red Sandstone"},
	"sand": {"Sand", "Red Sand", "Sand", "Sand",
		"Sand", "Sand", "Sand", "Sand",
		"Sand", "Sand", "Sand", "Sand",
		"Sand", "Sand", "Sand", "Sand"},
	"sandstone": {"Sandstone", "Chiseled Sandstone", "Smooth Sandstone", "Sandstone",
		"Sandstone", "Sandstone", "Sandstone", "Sandstone",
		"Sandstone", "Sandstone", "Sandstone", "Sandstone",
		"Sandstone", "Sandstone", "Sandstone", "Sandstone"},
	"sapling": {"Oak Sapling", "Spruce Sapling", "Birch Sapling", "Jungle Sapling",
		"Acacia Sapling", "Dark Oak Sapling", "Oak Sapling", "Oak Sapling",
		"Oak Sapling", "Oak Sapling", "Oak Sapling", "Oak Sapling",
		"Oak Sapling", "Oak Sapling", "Oak Sapling", "Oak Sapling"},
	"skull": {"Skeleton Skull", "Wither Skeleton Skull", "Zombie Head", "Head",
		"Creeper Head", "Dragon Head", "Skeleton Skull", "Skeleton Skull",
		"Skeleton Skull", "Skeleton Skull", "Skeleton Skull", "Skeleton Skull",
		"Skeleton Skull", "Skeleton Skull", "Skeleton Skull", "Skeleton Skull"},
	"sponge": {"Sponge", "Wet Sponge", "Sponge", "Wet Sponge",
		"Sponge", "Wet Sponge", "Sponge", "Wet Sponge",
		"Sponge", "Wet Sponge", "Sponge", "Wet Sponge",
		"Sponge", "Wet Sponge", "Sponge", "Wet Sponge"},
	"stained_glass": {"White Stained Glass", "Orange Stained Glass", "Magenta Stained Glass", "Light Blue Stained Glass",
		"Yellow Stained Glass", "Lime Stained Glass", "Pink Stained Glass", "Gray Stained Glass",
		"Light Gray Stained Glass", "Cyan Stained Glass", "Purple Stained Glass", "Blue Stained Glass",
		"Brown Stained Glass", "Green Stained Glass", "Red Stained Glass", "Black Stained Glass"},
	"stained_glass_pane": {"White Stained Glass Pane", "Orange Stained Glass Pane", "Magenta Stained Glass Pane", "Light Blue Stained Glass Pane",
		"Yellow Stained Glass Pane", "Lime Stained Glass Pane", "Pink Stained Glass Pane", "Gray Stained Glass Pane",
		"Light Gray Stained Glass Pane", "Cyan Stained Glass Pane", "Purple Stained Glass Pane", "Blue Stained Glass Pane",
		"Brown Stained Glass Pane", "Green Stained Glass Pane", "Red Stained Glass Pane", "Black Stained Glass Pane"},
	"stained_hardened_clay": {"White Terracotta", "Orange Terracotta", "Magenta Terracotta", "Light Blue Terracotta",
		"Yellow Terracotta", "Lime Terracotta", "Pink Terracotta", "Gray Terracotta",
		"Light Gray Terracotta", "Cyan Terracotta", "Purple Terracotta", "Blue Terracotta",
		"Brown Terracotta", "Green Terracotta", "Red Terracotta", "Black Terracotta"},
	"stone": {"Stone", "Granite", "Polished Granite", "Diorite",
		"Polished Diorite", "Andesite", "Polished Andesite", "Stone",
		"Stone", "Stone", "Stone", "Stone",
		"Stone", "Stone", "Stone", "Stone"},
	"stone_slab": {"Stone Slab", "Sandstone Slab", "Wooden Slab", "Cobblestone Slab",
		"Bricks Slab", "Stone Bricks Slab", "Nether Brick Slab", "Quartz Slab",
		"Stone Slab", "Stone Slab", "Stone Slab", "Stone Slab",
		"Stone Slab", "Stone Slab", "Stone Slab", "Stone Slab"},
	"stone_slab2": {"Red Sandstone Slab", "Red Sandstone Slab", "Red Sandstone Slab", "Red Sandstone Slab",
		"Red Sandstone Slab", "Red Sandstone Slab", "Red Sandstone Slab", "Red Sandstone Slab",
		"Red Sandstone Slab", "Red Sandstone Slab", "Red Sandstone Slab", "Red Sandstone Slab",
		"Red Sandstone Slab", "Red Sandstone Slab", "Red Sandstone Slab", "Red Sandstone Slab"},
	"stonebrick": {"Stone Bricks", "Mossy Stone Bricks", "Cracked Stone Bricks", "Chiseled Stone Bricks",
		"Stone Bricks", "Stone Bricks", "Stone Bricks", "Stone Bricks",
		"Stone Bricks", "Stone Bricks", "Stone Bricks", "Stone Bricks",
		"Stone Bricks", "Stone Bricks", "Stone Bricks", "Stone Bricks"},
	"tallgrass": {"Shrub", "Grass", "Fern", "Grass",
		"Grass", "Grass", "Grass", "Grass",
		"Grass", "Grass", "Grass", "Grass",
		"Grass", "Grass", "Grass", "Grass"},
	"wooden_slab": {"Oak Wood Slab", "Spruce Wood Slab", "Birch Wood Slab", "Jungle Wood Slab",
		"Acacia Wood Slab", "Dark Oak Wood Slab", "Oak Wood Slab", "Oak Wood Slab",
		"Oak Wood Slab", "Oak Wood Slab", "Oak Wood Slab", "Oak Wood Slab",
		"Oak Wood Slab", "Oak Wood Slab", "Oak Wood Slab", "Oak Wood Slab"},
	"wool": {"White Wool", "Orange Wool", "Magenta Wool", "Light Blue Wool",
		"Yellow Wool", "Lime Wool", "Pink Wool", "Gray Wool",
		"Light Gray Wool", "Cyan Wool", "Purple Wool", "Blue Wool",
		"Brown Wool", "Green Wool", "Red Wool", "Black Wool"},
	"yellow_flower": {"Dandelion", "Dandelion", "Dandelion", "Dandelion",
		"Dandelion", "Dandelion", "Dandelion", "Dandelion",
		"Dandelion", "Dandelion", "Dandelion", "Dandelion",
		"Dandelion", "Dandelion", "Dandelion", "Dandelion"},
}
