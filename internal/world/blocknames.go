package world

import "strings"

// blockDisplayNames maps a 1.12.2 block id to the name the server shows.
//
// The names come from the language file that ships with the 1.12.2 client
// assets: the server looks them up as "tile.<id>.name". Blocks with
// variants keep their name under a two level key, so they are listed
// explicitly.
var blockDisplayNames = map[string]string{
	"acaciaFence":               "Acacia Fence",
	"acaciaFenceGate":           "Acacia Fence Gate",
	"activatorRail":             "Activator Rail",
	"air":                       "Air",
	"andesite":                  "Andesite",
	"anvil":                     "Anvil",
	"barrier":                   "Barrier",
	"beacon":                    "Beacon",
	"bedrock":                   "Bedrock",
	"beetroots":                 "Beetroots",
	"birchFence":                "Birch Fence",
	"birchFenceGate":            "Birch Fence Gate",
	"blockCoal":                 "Block of Coal",
	"blockDiamond":              "Block of Diamond",
	"blockEmerald":              "Block of Emerald",
	"blockGold":                 "Block of Gold",
	"blockIron":                 "Block of Iron",
	"blockLapis":                "Lapis Lazuli Block",
	"blockRedstone":             "Block of Redstone",
	"boneBlock":                 "Bone Block",
	"bookshelf":                 "Bookshelf",
	"brick":                     "Bricks",
	"button":                    "Button",
	"cactus":                    "Cactus",
	"cake":                      "Cake",
	"carrots":                   "Carrots",
	"cauldron":                  "Cauldron",
	"chainCommandBlock":         "Chain Command Block",
	"chest":                     "Chest",
	"chestTrap":                 "Trapped Chest",
	"chiseled_sandstone":        "Chiseled Sandstone",
	"chorusFlower":              "Chorus Flower",
	"chorusPlant":               "Chorus Plant",
	"clay":                      "Clay",
	"clayHardened":              "Terracotta",
	"clayHardenedStained":       "Stained Terracotta",
	"cloth":                     "Wool",
	"cocoa":                     "Cocoa",
	"commandBlock":              "Command Block",
	"crops":                     "Crops",
	"darkOakFence":              "Dark Oak Fence",
	"darkOakFenceGate":          "Dark Oak Fence Gate",
	"daylightDetector":          "Daylight Sensor",
	"deadbush":                  "Dead Bush",
	"detectorRail":              "Detector Rail",
	"diorite":                   "Diorite",
	"dirt":                      "Dirt",
	"dispenser":                 "Dispenser",
	"doorIron":                  "Iron Door",
	"doorWood":                  "Wooden Door",
	"doublePlant":               "Plant",
	"dragonEgg":                 "Dragon Egg",
	"dropper":                   "Dropper",
	"enchantmentTable":          "Enchantment Table",
	"endBricks":                 "End Stone Bricks",
	"enderChest":                "Ender Chest",
	"endPortalFrame":            "End Portal",
	"endRod":                    "End Rod",
	"farmland":                  "Farmland",
	"fence":                     "Oak Fence",
	"fenceGate":                 "Oak Fence Gate",
	"fenceIron":                 "Iron Bars",
	"fire":                      "Fire",
	"flower1":                   "Flower",
	"flower2":                   "Flower",
	"frostedIce":                "Frosted Ice",
	"furnace":                   "Furnace",
	"glass":                     "Glass",
	"glazedTerracottaBlack":     "Black Glazed Terracotta",
	"glazedTerracottaBlue":      "Blue Glazed Terracotta",
	"glazedTerracottaBrown":     "Brown Glazed Terracotta",
	"glazedTerracottaCyan":      "Cyan Glazed Terracotta",
	"glazedTerracottaGray":      "Gray Glazed Terracotta",
	"glazedTerracottaGreen":     "Green Glazed Terracotta",
	"glazedTerracottaLightBlue": "Light Blue Glazed Terracotta",
	"glazedTerracottaLime":      "Lime Glazed Terracotta",
	"glazedTerracottaMagenta":   "Magenta Glazed Terracotta",
	"glazedTerracottaOrange":    "Orange Glazed Terracotta",
	"glazedTerracottaPink":      "Pink Glazed Terracotta",
	"glazedTerracottaPurple":    "Purple Glazed Terracotta",
	"glazedTerracottaRed":       "Red Glazed Terracotta",
	"glazedTerracottaSilver":    "Light Gray Glazed Terracotta",
	"glazedTerracottaWhite":     "White Glazed Terracotta",
	"glazedTerracottaYellow":    "Yellow Glazed Terracotta",
	"goldenRail":                "Powered Rail",
	"granite":                   "Granite",
	"grass":                     "Grass Block",
	"grassPath":                 "Grass Path",
	"gravel":                    "Gravel",
	"hayBlock":                  "Hay Bale",
	"hellrock":                  "Netherrack",
	"hellsand":                  "Soul Sand",
	"hopper":                    "Hopper",
	"ice":                       "Ice",
	"icePacked":                 "Packed Ice",
	"ironTrapdoor":              "Iron Trapdoor",
	"jukebox":                   "Jukebox",
	"jungleFence":               "Jungle Fence",
	"jungleFenceGate":           "Jungle Fence Gate",
	"ladder":                    "Ladder",
	"lava":                      "Lava",
	"leaves":                    "Leaves",
	"lever":                     "Lever",
	"lightgem":                  "Glowstone",
	"litpumpkin":                "Jack o'Lantern",
	"lockedchest":               "Locked chest",
	"log":                       "Wood",
	"magma":                     "Magma Block",
	"melon":                     "Melon",
	"mobSpawner":                "Monster Spawner",
	"monsterStoneEgg":           "Stone Monster Egg",
	"mushroom":                  "Mushroom",
	"musicBlock":                "Note Block",
	"mycel":                     "Mycelium",
	"netherBrick":               "Nether Brick",
	"netherFence":               "Nether Brick Fence",
	"netherquartz":              "Nether Quartz Ore",
	"netherStalk":               "Nether Wart",
	"netherWartBlock":           "Nether Wart Block",
	"notGate":                   "Redstone Torch",
	"observer":                  "Observer",
	"obsidian":                  "Obsidian",
	"oreCoal":                   "Coal Ore",
	"oreDiamond":                "Diamond Ore",
	"oreEmerald":                "Emerald Ore",
	"oreGold":                   "Gold Ore",
	"oreIron":                   "Iron Ore",
	"oreLapis":                  "Lapis Lazuli Ore",
	"oreRedstone":               "Redstone Ore",
	"pistonBase":                "Piston",
	"pistonStickyBase":          "Sticky Piston",
	"polished_andesite":         "Polished Andesite",
	"polished_diorite":          "Polished Diorite",
	"polished_granite":          "Polished Granite",
	"portal":                    "Portal",
	"potatoes":                  "Potatoes",
	"pressurePlateStone":        "Stone Pressure Plate",
	"pressurePlateWood":         "Wooden Pressure Plate",
	"pumpkin":                   "Pumpkin",
	"pumpkinStem":               "Pumpkin Stem",
	"purpurBlock":               "Purpur Block",
	"purpurPillar":              "Purpur Pillar",
	"purpurSlab":                "Purpur Slab",
	"quartzBlock":               "Block of Quartz",
	"rail":                      "Rail",
	"red_sandstone":             "Red Sandstone",
	"redNetherBrick":            "Red Nether Brick",
	"redSandStone":              "Red Sandstone",
	"redstoneDust":              "Redstone Dust",
	"redstoneLight":             "Redstone Lamp",
	"reeds":                     "Sugar cane",
	"repeatingCommandBlock":     "Repeating Command Block",
	"sand":                      "Sand",
	"sandstone":                 "Sandstone",
	"seaLantern":                "Sea Lantern",
	"shulkerBoxBlack":           "Black Shulker Box",
	"shulkerBoxBlue":            "Blue Shulker Box",
	"shulkerBoxBrown":           "Brown Shulker Box",
	"shulkerBoxCyan":            "Cyan Shulker Box",
	"shulkerBoxGray":            "Gray Shulker Box",
	"shulkerBoxGreen":           "Green Shulker Box",
	"shulkerBoxLightBlue":       "Light Blue Shulker Box",
	"shulkerBoxLime":            "Lime Shulker Box",
	"shulkerBoxMagenta":         "Magenta Shulker Box",
	"shulkerBoxOrange":          "Orange Shulker Box",
	"shulkerBoxPink":            "Pink Shulker Box",
	"shulkerBoxPurple":          "Purple Shulker Box",
	"shulkerBoxRed":             "Red Shulker Box",
	"shulkerBoxSilver":          "Light Gray Shulker Box",
	"shulkerBoxWhite":           "White Shulker Box",
	"shulkerBoxYellow":          "Yellow Shulker Box",
	"sign":                      "Sign",
	"slime":                     "Slime Block",
	"smooth_sandstone":          "Smooth Sandstone",
	"snow":                      "Snow",
	"spruceFence":               "Spruce Fence",
	"spruceFenceGate":           "Spruce Fence Gate",
	"stainedGlass":              "Stained Glass",
	"stairsBrick":               "Brick Stairs",
	"stairsNetherBrick":         "Nether Brick Stairs",
	"stairsPurpur":              "Purpur Stairs",
	"stairsQuartz":              "Quartz Stairs",
	"stairsRedSandStone":        "Red Sandstone Stairs",
	"stairsSandStone":           "Sandstone Stairs",
	"stairsStone":               "Cobblestone Stairs",
	"stairsStoneBrickSmooth":    "Stone Brick Stairs",
	"stairsWood":                "Oak Wood Stairs",
	"stairsWoodAcacia":          "Acacia Wood Stairs",
	"stairsWoodBirch":           "Birch Wood Stairs",
	"stairsWoodDarkOak":         "Dark Oak Wood Stairs",
	"stairsWoodJungle":          "Jungle Wood Stairs",
	"stairsWoodSpruce":          "Spruce Wood Stairs",
	"stone":                     "Stone",
	"stonebrick":                "Cobblestone",
	"stonebricksmooth":          "Stone Bricks",
	"stoneMoss":                 "Moss Stone",
	"stoneSlab":                 "Stone Slab",
	"structureBlock":            "Structure Block",
	"structureVoid":             "Structure Void",
	"tallgrass":                 "Grass",
	"thinGlass":                 "Glass Pane",
	"thinStainedGlass":          "Stained Glass Pane",
	"tnt":                       "TNT",
	"torch":                     "Torch",
	"trapdoor":                  "Wooden Trapdoor",
	"tripWire":                  "Tripwire",
	"tripWireSource":            "Tripwire Hook",
	"vine":                      "Vines",
	"water":                     "Water",
	"waterlily":                 "Lily Pad",
	"web":                       "Cobweb",
	"weightedPlate_heavy":       "Weighted Pressure Plate (Heavy)",
	"weightedPlate_light":       "Weighted Pressure Plate (Light)",
	"whiteStone":                "End Stone",
	"wood":                      "Wooden Planks",
	"woodSlab":                  "Wood Slab",
	"woolCarpet":                "Carpet",
	"workbench":                 "Crafting Table",
}

// BlockDisplayName returns the name 1.12.2 shows for a block.
//
// The name of the block is looked up in the table of the registry names
// first, because 1.12.2 reuses some of the names of 1.11 for a
// different block: "stonebrick" is the cobblestone in 1.11 and the stone
// bricks in 1.12.2. The table of the older names comes second, since
// 1.12.2 still takes those.
//
// An unknown id is turned into something readable rather than being
// dropped, so a message still says which block was meant.
func BlockDisplayName(name string) string {
	if name == "" {
		return "Air"
	}
	key := name
	if strings.HasPrefix(key, "minecraft:") {
		key = strings.TrimPrefix(key, "minecraft:")
	}
	if display, ok := blockRegistryDisplayNames[key]; ok {
		return display
	}
	if display, ok := blockDisplayNames[key]; ok {
		return display
	}
	return readableBlockName(key)
}

// readableBlockName turns an id into a readable name.
//
// It is the fallback for the blocks whose name lives under a two level
// key ("tile.stone.stone.name" and friends). The result looks like the
// names in the language file: every word starts with a capital.
func readableBlockName(key string) string {
	out := make([]rune, 0, len(key))
	for i, r := range key {
		switch {
		case r == '_':
			r = ' '
		case r >= 'A' && r <= 'Z' && i > 0:
			r = r + ('a' - 'A')
		}
		out = append(out, r)
	}
	capitalize := true
	for i, r := range out {
		if capitalize && r >= 'a' && r <= 'z' {
			out[i] = r - ('a' - 'A')
		}
		capitalize = r == ' '
	}
	return string(out)
}
