package v1_12_2

import "strings"

// The statistic registry of 1.12.2.
//
// 1.12.2 keeps its statistics in `net.minecraft.server.v1_12_R1.Statistic`,
// and `StatisticList` builds them. Every statistic has an id and a display
// name, and **the id is the key of the language file**: `StatisticList`
// builds a statistic with
//
//	new CounterStatistic("stat.jump", new ChatMessage("stat.jump"))
//
// so the id of a statistic and the key its name is translated with are the
// same string. The id is *not* the "minecraft:jump" spelling of the later
// versions, and 1.12.2 has no "minecraft:custom" either.
//
// The id is also the name of the scoreboard criterion the statistic
// registers itself as, which is where the ids are reachable: a statistic
// can be the criterion of a scoreboard objective.
//
// (Measured on a real 1.12.2 server: `scoreboard objectives add x stat.jump`
// answers "Added new objective 'x' successfully" and the objective is of
// type 'stat.jump', while `scoreboard objectives add x minecraft:jump`
// answers "Usage: Invalid objective criteria type 'minecraft:jump'".)

// statKind tells how the value of a statistic is counted.
type statKind int

const (
	// statInt is a whole number: a count of events, or a number of
	// points of damage.
	statInt statKind = iota
	// statFloat is a number the client shows with two decimals: the
	// distances and the durations.
	//
	// The server keeps every statistic as an integer, so the difference
	// is only in the way the value is shown: the client divides the
	// distances by a hundred and prints the result with the
	// "########0.00" format `Statistic` holds for it.
	statFloat
)

// statSpec describes one statistic of the registry.
type statSpec struct {
	// ID is the identifier of the statistic, e.g. "stat.jump". It is the
	// key of the language file its name is translated with.
	ID string
	// Display is the English name of the statistic.
	Display string
	// Kind tells how the value of the statistic is counted.
	Kind statKind
}

// statRegistry holds the statistics 1.12.2 builds for itself.
//
// The list was read out of the static initializer of `StatisticList`, which
// holds 50 of them, and the names are the English translations of the
// language file that ships with the 1.12.2 client.
//
// Two of the statistics share a name: the statistic "stat.craftingTableInteraction"
// is shown as "Interactions with Crafting Table", because the language file
// has no key of its own for it and 1.12.2 gives it the key of
// "stat.workbenchInteraction".
var statRegistry = []statSpec{
	{ID: "stat.leaveGame", Display: "Games Quit", Kind: statInt},
	{ID: "stat.playOneMinute", Display: "Time Played", Kind: statFloat},
	{ID: "stat.timeSinceDeath", Display: "Since Last Death", Kind: statFloat},
	{ID: "stat.sneakTime", Display: "Sneak Time", Kind: statFloat},
	{ID: "stat.walkOneCm", Display: "Distance Walked", Kind: statFloat},
	{ID: "stat.crouchOneCm", Display: "Distance Crouched", Kind: statFloat},
	{ID: "stat.sprintOneCm", Display: "Distance Sprinted", Kind: statFloat},
	{ID: "stat.swimOneCm", Display: "Distance Swum", Kind: statFloat},
	{ID: "stat.fallOneCm", Display: "Distance Fallen", Kind: statFloat},
	{ID: "stat.climbOneCm", Display: "Distance Climbed", Kind: statFloat},
	{ID: "stat.flyOneCm", Display: "Distance Flown", Kind: statFloat},
	{ID: "stat.diveOneCm", Display: "Distance Dove", Kind: statFloat},
	{ID: "stat.minecartOneCm", Display: "Distance by Minecart", Kind: statFloat},
	{ID: "stat.boatOneCm", Display: "Distance by Boat", Kind: statFloat},
	{ID: "stat.pigOneCm", Display: "Distance by Pig", Kind: statFloat},
	{ID: "stat.horseOneCm", Display: "Distance by Horse", Kind: statFloat},
	{ID: "stat.aviateOneCm", Display: "Distance by Elytra", Kind: statFloat},
	{ID: "stat.jump", Display: "Jumps", Kind: statInt},
	{ID: "stat.drop", Display: "Items Dropped", Kind: statInt},
	{ID: "stat.damageDealt", Display: "Damage Dealt", Kind: statInt},
	{ID: "stat.damageTaken", Display: "Damage Taken", Kind: statInt},
	{ID: "stat.deaths", Display: "Number of Deaths", Kind: statInt},
	{ID: "stat.mobKills", Display: "Mob Kills", Kind: statInt},
	{ID: "stat.animalsBred", Display: "Animals Bred", Kind: statInt},
	{ID: "stat.playerKills", Display: "Player Kills", Kind: statInt},
	{ID: "stat.fishCaught", Display: "Fish Caught", Kind: statInt},
	{ID: "stat.talkedToVillager", Display: "Talked to Villagers", Kind: statInt},
	{ID: "stat.tradedWithVillager", Display: "Traded with Villagers", Kind: statInt},
	{ID: "stat.cakeSlicesEaten", Display: "Cake Slices Eaten", Kind: statInt},
	{ID: "stat.cauldronFilled", Display: "Cauldrons Filled", Kind: statInt},
	{ID: "stat.cauldronUsed", Display: "Water Taken from Cauldron", Kind: statInt},
	{ID: "stat.armorCleaned", Display: "Armor Pieces Cleaned", Kind: statInt},
	{ID: "stat.bannerCleaned", Display: "Banners Cleaned", Kind: statInt},
	{ID: "stat.brewingstandInteraction", Display: "Interactions with Brewing Stand", Kind: statInt},
	{ID: "stat.beaconInteraction", Display: "Interactions with Beacon", Kind: statInt},
	{ID: "stat.dropperInspected", Display: "Droppers Searched", Kind: statInt},
	{ID: "stat.hopperInspected", Display: "Hoppers Searched", Kind: statInt},
	{ID: "stat.dispenserInspected", Display: "Dispensers Searched", Kind: statInt},
	{ID: "stat.noteblockPlayed", Display: "Note Blocks Played", Kind: statInt},
	{ID: "stat.noteblockTuned", Display: "Note Blocks Tuned", Kind: statInt},
	{ID: "stat.flowerPotted", Display: "Plants Potted", Kind: statInt},
	{ID: "stat.trappedChestTriggered", Display: "Trapped Chests Triggered", Kind: statInt},
	{ID: "stat.enderchestOpened", Display: "Ender Chests Opened", Kind: statInt},
	{ID: "stat.itemEnchanted", Display: "Items Enchanted", Kind: statInt},
	{ID: "stat.recordPlayed", Display: "Records Played", Kind: statInt},
	{ID: "stat.furnaceInteraction", Display: "Interactions with Furnace", Kind: statInt},
	{ID: "stat.craftingTableInteraction", Display: "Interactions with Crafting Table", Kind: statInt},
	{ID: "stat.chestOpened", Display: "Chests Opened", Kind: statInt},
	{ID: "stat.sleepInBed", Display: "Times Slept in a Bed", Kind: statInt},
	{ID: "stat.shulkerBoxOpened", Display: "Shulker Boxes Opened", Kind: statInt},
}

// statFamily is a group of statistics that 1.12.2 builds from a registry
// instead of writing them out one by one.
type statFamily struct {
	// Prefix is the beginning of the id of every statistic of the family.
	Prefix string
	// Display is the key of the language file the name is translated with.
	Display string
	// Kind tells how the value of the statistics is counted.
	Kind statKind
}

// statItemFamilies are the statistics 1.12.2 builds from the item registry.
//
// The suffix of the id is the name of the item with the namespace
// separated by a dot, so the statistic of the diamond pickaxe is
// "stat.useItem.minecraft.diamond_pickaxe".
//
// Only the items that qualify are in the family: a stick is used but not
// broken, and a block that is not minable has no "stat.mineBlock" at all.
//
// (Measured on a real 1.12.2 server: `scoreboard objectives add x
// stat.useItem.minecraft.diamond_pickaxe` is accepted, while
// `scoreboard objectives add x stat.useItem.diamond_pickaxe` is answered
// with "Usage: Invalid objective criteria type
// 'stat.useItem.diamond_pickaxe'".)
var statItemFamilies = []statFamily{
	{Prefix: "stat.craftItem.", Display: "stat.craftItem", Kind: statInt},
	{Prefix: "stat.mineBlock.", Display: "stat.mineBlock", Kind: statInt},
	{Prefix: "stat.useItem.", Display: "stat.useItem", Kind: statInt},
	{Prefix: "stat.breakItem.", Display: "stat.breakItem", Kind: statInt},
	{Prefix: "stat.pickup.", Display: "stat.pickup", Kind: statInt},
	{Prefix: "stat.drop.", Display: "stat.drop", Kind: statInt},
}

// statEntityFamilies are the statistics 1.12.2 builds from the entity
// registry.
//
// The suffix of the id is the name of the entity, so the statistic of the
// zombie is "stat.killEntity.zombie".
//
// The server does not accept them as scoreboard criteria, so they are only
// reachable through the statistics themselves.
//
// (Measured on a real 1.12.2 server: `scoreboard objectives add x
// stat.killEntity.zombie` is answered with "Usage: Invalid objective
// criteria type 'stat.killEntity.zombie'".)
var statEntityFamilies = []statFamily{
	{Prefix: "stat.killEntity.", Display: "stat.entityKill", Kind: statInt},
	{Prefix: "stat.entityKilledBy.", Display: "stat.entityKilledBy", Kind: statInt},
}

// lookupStat returns the statistic with the given id.
//
// The id of a statistic of a family is the prefix of the family followed by
// the name of the item or the entity, so a family matches every id that
// starts with its prefix.
//
// The name of a statistic of a family is the name of the family with the
// name of the item or the entity filled in, so it is not spelled out here.
func lookupStat(id string) (statSpec, bool) {
	for _, s := range statRegistry {
		if s.ID == id {
			return s, true
		}
	}
	for _, f := range statItemFamilies {
		if strings.HasPrefix(id, f.Prefix) && len(id) > len(f.Prefix) {
			return statSpec{ID: id, Display: f.Display, Kind: f.Kind}, true
		}
	}
	for _, f := range statEntityFamilies {
		if strings.HasPrefix(id, f.Prefix) && len(id) > len(f.Prefix) {
			return statSpec{ID: id, Display: f.Display, Kind: f.Kind}, true
		}
	}
	return statSpec{}, false
}
