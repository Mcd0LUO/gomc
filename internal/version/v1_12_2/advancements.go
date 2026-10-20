package v1_12_2

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// advancementDef is one entry of the 1.12.2 advancement registry.
//
// 1.12.2 keeps the advancements in JSON files, and the server reads the
// built-in ones from its own jar: `AdvancementDataWorld` walks
// `assets/minecraft/advancements` and turns every JSON file into an
// advancement whose name is the path of the file below that directory,
// in the "minecraft" namespace.
//
// A datapack may add more, under `<world>/advancements/<namespace>/<path>.json`
// (`AdvancementDataWorld.d`). Those are read by advancements_custom.go and
// are presented to the command as the same kind of advancement.
type advancementDef struct {
	// name is the namespaced id, e.g. "minecraft:story/root".
	name string
	// parent is the parent's namespaced id. It is empty for a root
	// advancement, one that has no parent.
	parent string
	// criteria are the criterion names, in the order the JSON declares
	// them.
	criteria []string
	// kind tells which of the criteria have to be done for the advancement
	// to be complete.
	kind requirementKind
}

// requirementKind tells which criteria an advancement needs to be complete.
type requirementKind int

const (
	// requireAll needs every criterion of the advancement to be done.
	//
	// It is what 1.12.2 builds for an advancement JSON without a
	// "requirements" array: the server makes one group per criterion, and
	// every group has to be satisfied.
	requireAll requirementKind = iota
	// requireAny needs only one of the criteria to be done.
	//
	// It is what 1.12.2 builds for an advancement JSON whose
	// "requirements" array holds a single group with every criterion in it:
	// such a group is satisfied as soon as one of its criteria is done.
	requireAny
)

// The advancements of the 1.12.2 server.
//
// The data is the content of the JSON files of the server jar
// (`assets/minecraft/advancements/<path>.json`), 488 of them. The parent
// is the "parent" field of the file and the criteria are the keys of its
// "criteria" object.
//
// The kind tells what the "requirements" array of the file means: 438 of
// the files declare a single group holding every criterion, so one done
// criterion is enough (requireAny), and the other 50 declare no
// "requirements" at all, so the server builds one group per criterion and
// every criterion has to be done (requireAll).
var advancementRegistry = []advancementDef{

	{"minecraft:adventure/adventuring_time", "minecraft:adventure/sleep_in_bed", []string{
		"birch_forest_hills",
		"river",
		"swampland",
		"desert",
		"forest_hills",
		"redwood_taiga_hills",
		"taiga_cold",
		"mesa",
		"forest",
		"stone_beach",
		"ice_flats",
		"taiga_hills",
		"ice_mountains",
		"mesa_rock",
		"savanna",
		"plains",
		"frozen_river",
		"redwood_taiga",
		"cold_beach",
		"deep_ocean",
		"jungle_hills",
		"jungle_edge",
		"ocean",
		"mushroom_island_shore",
		"extreme_hills",
		"desert_hills",
		"jungle",
		"beaches",
		"savanna_rock",
		"taiga_cold_hills",
		"mesa_clear_rock",
		"roofed_forest",
		"taiga",
		"birch_forest",
		"mushroom_island",
		"extreme_hills_with_trees",
	}, requireAll},
	{"minecraft:adventure/kill_a_mob", "minecraft:adventure/root", []string{
		"cave_spider",
		"spider",
		"zombie_pigman",
		"enderman",
		"polar_bear",
		"blaze",
		"creeper",
		"evoker",
		"ghast",
		"guardian",
		"husk",
		"magma_cube",
		"shulker",
		"silverfish",
		"skeleton",
		"slime",
		"stray",
		"vindicator",
		"witch",
		"wither_skeleton",
		"zombie",
		"zombie_villager",
	}, requireAny},
	{"minecraft:adventure/kill_all_mobs", "minecraft:adventure/kill_a_mob", []string{
		"cave_spider",
		"spider",
		"zombie_pigman",
		"enderman",
		"polar_bear",
		"blaze",
		"creeper",
		"evoker",
		"ghast",
		"guardian",
		"husk",
		"magma_cube",
		"shulker",
		"silverfish",
		"skeleton",
		"slime",
		"stray",
		"vindicator",
		"witch",
		"wither_skeleton",
		"zombie",
		"zombie_villager",
	}, requireAll},
	{"minecraft:adventure/root", "", []string{"killed_something", "killed_by_something"}, requireAny},
	{"minecraft:adventure/shoot_arrow", "minecraft:adventure/kill_a_mob", []string{"shot_arrow"}, requireAll},
	{"minecraft:adventure/sleep_in_bed", "minecraft:adventure/root", []string{"slept_in_bed"}, requireAll},
	{"minecraft:adventure/sniper_duel", "minecraft:adventure/shoot_arrow", []string{
		"killed_skeleton",
	}, requireAll},
	{"minecraft:adventure/summon_iron_golem", "minecraft:adventure/trade", []string{
		"summoned_golem",
	}, requireAll},
	{"minecraft:adventure/totem_of_undying", "minecraft:adventure/kill_a_mob", []string{
		"used_totem",
	}, requireAll},
	{"minecraft:adventure/trade", "minecraft:adventure/root", []string{"traded"}, requireAll},
	{"minecraft:end/dragon_breath", "minecraft:end/kill_dragon", []string{"dragon_breath"}, requireAll},
	{"minecraft:end/dragon_egg", "minecraft:end/kill_dragon", []string{"dragon_egg"}, requireAll},
	{"minecraft:end/elytra", "minecraft:end/find_end_city", []string{"elytra"}, requireAll},
	{"minecraft:end/enter_end_gateway", "minecraft:end/kill_dragon", []string{
		"entered_end_gateway",
	}, requireAll},
	{"minecraft:end/find_end_city", "minecraft:end/enter_end_gateway", []string{"in_city"}, requireAll},
	{"minecraft:end/kill_dragon", "minecraft:end/root", []string{"killed_dragon"}, requireAll},
	{"minecraft:end/levitate", "minecraft:end/find_end_city", []string{"levitated"}, requireAll},
	{"minecraft:end/respawn_dragon", "minecraft:end/kill_dragon", []string{"summoned_dragon"}, requireAll},
	{"minecraft:end/root", "", []string{"entered_end"}, requireAll},
	{"minecraft:husbandry/balanced_diet", "minecraft:husbandry/plant_seed", []string{
		"apple",
		"mushroom_stew",
		"bread",
		"porkchop",
		"cooked_porkchop",
		"golden_apple",
		"enchanted_golden_apple",
		"raw_cod",
		"raw_salmon",
		"clownfish",
		"pufferfish",
		"cooked_cod",
		"cooked_salmon",
		"cookie",
		"melon",
		"beef",
		"cooked_beef",
		"chicken",
		"cooked_chicken",
		"rotten_flesh",
		"spider_eye",
		"carrot",
		"potato",
		"baked_potato",
		"poisonous_potato",
		"golden_carrot",
		"pumpkin_pie",
		"rabbit",
		"cooked_rabbit",
		"rabbit_stew",
		"mutton",
		"cooked_mutton",
		"chorus_fruit",
		"beetroot",
		"beetroot_soup",
	}, requireAll},
	{"minecraft:husbandry/break_diamond_hoe", "minecraft:husbandry/plant_seed", []string{
		"broke_hoe",
	}, requireAll},
	{"minecraft:husbandry/bred_all_animals", "minecraft:husbandry/breed_an_animal", []string{
		"bred_horse",
		"bred_sheep",
		"bred_cow",
		"bred_mooshroom",
		"bred_pig",
		"bred_chicken",
		"bred_wolf",
		"bred_ocelot",
		"bred_rabbit",
		"bred_llama",
	}, requireAll},
	{"minecraft:husbandry/breed_an_animal", "minecraft:husbandry/root", []string{
		"bred_horse",
		"bred_sheep",
		"bred_cow",
		"bred_mooshroom",
		"bred_pig",
		"bred_chicken",
		"bred_wolf",
		"bred_ocelot",
		"bred_rabbit",
		"bred_llama",
	}, requireAny},
	{"minecraft:husbandry/plant_seed", "minecraft:husbandry/root", []string{
		"wheat",
		"pumpkin_stem",
		"melon_stem",
		"beetroots",
		"nether_wart",
	}, requireAny},
	{"minecraft:husbandry/root", "", []string{"consumed_item"}, requireAll},
	{"minecraft:husbandry/tame_an_animal", "minecraft:husbandry/root", []string{"tamed_animal"}, requireAll},
	{"minecraft:nether/all_effects", "minecraft:nether/all_potions", []string{"all_effects"}, requireAll},
	{"minecraft:nether/all_potions", "minecraft:nether/brew_potion", []string{"all_effects"}, requireAll},
	{"minecraft:nether/brew_potion", "minecraft:nether/obtain_blaze_rod", []string{"potion"}, requireAll},
	{"minecraft:nether/create_beacon", "minecraft:nether/summon_wither", []string{"beacon"}, requireAll},
	{"minecraft:nether/create_full_beacon", "minecraft:nether/create_beacon", []string{"beacon"}, requireAll},
	{"minecraft:nether/fast_travel", "minecraft:nether/root", []string{"travelled"}, requireAll},
	{"minecraft:nether/find_fortress", "minecraft:nether/root", []string{"fortress"}, requireAll},
	{"minecraft:nether/get_wither_skull", "minecraft:nether/find_fortress", []string{
		"wither_skull",
	}, requireAll},
	{"minecraft:nether/obtain_blaze_rod", "minecraft:nether/find_fortress", []string{"blaze_rod"}, requireAll},
	{"minecraft:nether/return_to_sender", "minecraft:nether/root", []string{"killed_ghast"}, requireAll},
	{"minecraft:nether/root", "", []string{"entered_nether"}, requireAll},
	{"minecraft:nether/summon_wither", "minecraft:nether/get_wither_skull", []string{"summoned"}, requireAll},
	{"minecraft:nether/uneasy_alliance", "minecraft:nether/return_to_sender", []string{
		"killed_ghast",
	}, requireAll},
	{"minecraft:recipes/brewing/blaze_powder", "minecraft:recipes/root", []string{
		"has_blaze_rod",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/brewing/brewing_stand", "minecraft:recipes/root", []string{
		"has_blaze_rod",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/brewing/cauldron", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_water_bucket",
	}, requireAny},
	{"minecraft:recipes/brewing/fermented_spider_eye", "minecraft:recipes/root", []string{
		"has_spider_eye",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/brewing/glass_bottle", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/brewing/golden_carrot", "minecraft:recipes/root", []string{
		"has_gold_nugget",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/brewing/magma_cream", "minecraft:recipes/root", []string{
		"has_blaze_powder",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/brewing/speckled_melon", "minecraft:recipes/root", []string{
		"has_melon",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/acacia_planks", "minecraft:recipes/root", []string{
		"has_log2",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/acacia_stairs", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/building_blocks/acacia_wooden_slab", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/building_blocks/andesite", "minecraft:recipes/root", []string{
		"has_stone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/birch_planks", "minecraft:recipes/root", []string{
		"has_log",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/birch_stairs", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/building_blocks/birch_wooden_slab", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/building_blocks/black_concrete_powder", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/black_stained_glass", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/black_stained_hardened_clay", "minecraft:recipes/root", []string{
		"has_hardened_clay",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/black_wool", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/blue_concrete_powder", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/blue_stained_glass", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/blue_stained_hardened_clay", "minecraft:recipes/root", []string{
		"has_hardened_clay",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/blue_wool", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/bone_block", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_at_least_9_bonemeal",
	}, requireAny},
	{"minecraft:recipes/building_blocks/bookshelf", "minecraft:recipes/root", []string{
		"has_book",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/brick_block", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_brick",
	}, requireAny},
	{"minecraft:recipes/building_blocks/brick_slab", "minecraft:recipes/root", []string{
		"has_brick_block",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/brick_stairs", "minecraft:recipes/root", []string{
		"has_brick_block",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/brown_concrete_powder", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/brown_stained_glass", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/brown_stained_hardened_clay", "minecraft:recipes/root", []string{
		"has_hardened_clay",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/brown_wool", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/chiseled_quartz_block", "minecraft:recipes/root", []string{
		"has_quartz_block",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/chiseled_red_sandstone", "minecraft:recipes/root", []string{
		"has_red_sandstone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/chiseled_sandstone", "minecraft:recipes/root", []string{
		"has_stone_slab",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/chiseled_stonebrick", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_stonebrick",
	}, requireAny},
	{"minecraft:recipes/building_blocks/clay", "minecraft:recipes/root", []string{
		"has_clay_ball",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/coal_block", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_at_least_9_coal",
	}, requireAny},
	{"minecraft:recipes/building_blocks/coarse_dirt", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/cobblestone_slab", "minecraft:recipes/root", []string{
		"has_cobblestone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/cyan_concrete_powder", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/cyan_stained_glass", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/cyan_stained_hardened_clay", "minecraft:recipes/root", []string{
		"has_hardened_clay",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/cyan_wool", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/dark_oak_planks", "minecraft:recipes/root", []string{
		"has_log2",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/dark_oak_stairs", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/building_blocks/dark_oak_wooden_slab", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/building_blocks/dark_prismarine", "minecraft:recipes/root", []string{
		"has_prismarine_shard",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/diamond_block", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_at_least_9_diamond",
	}, requireAny},
	{"minecraft:recipes/building_blocks/diorite", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_quartz",
	}, requireAny},
	{"minecraft:recipes/building_blocks/emerald_block", "minecraft:recipes/root", []string{
		"has_at_least_9_emerald",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/end_bricks", "minecraft:recipes/root", []string{
		"has_end_stone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/glowstone", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_glowstone_dust",
	}, requireAny},
	{"minecraft:recipes/building_blocks/gold_block", "minecraft:recipes/root", []string{
		"has_at_least_9_gold_ingot",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/granite", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_quartz",
	}, requireAny},
	{"minecraft:recipes/building_blocks/gray_concrete_powder", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/gray_stained_glass", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/gray_stained_hardened_clay", "minecraft:recipes/root", []string{
		"has_hardened_clay",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/gray_wool", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/green_concrete_powder", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/green_stained_glass", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/green_stained_hardened_clay", "minecraft:recipes/root", []string{
		"has_hardened_clay",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/green_wool", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/hay_block", "minecraft:recipes/root", []string{
		"has_at_least_9_wheat",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/iron_block", "minecraft:recipes/root", []string{
		"has_at_least_9_iron_ingot",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/jungle_planks", "minecraft:recipes/root", []string{
		"has_log",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/jungle_stairs", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/building_blocks/jungle_wooden_slab", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/building_blocks/lapis_block", "minecraft:recipes/root", []string{
		"has_at_least_9_lapis",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/light_blue_concrete_powder", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/light_blue_stained_glass", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/light_blue_stained_hardened_clay", "minecraft:recipes/root", []string{
		"has_hardened_clay",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/light_blue_wool", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/light_gray_concrete_powder", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/light_gray_stained_glass", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/light_gray_stained_hardened_clay", "minecraft:recipes/root", []string{
		"has_hardened_clay",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/light_gray_wool", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/lime_concrete_powder", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/lime_stained_glass", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/lime_stained_hardened_clay", "minecraft:recipes/root", []string{
		"has_hardened_clay",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/lime_wool", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/lit_pumpkin", "minecraft:recipes/root", []string{
		"has_pumpkin",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/magenta_concrete_powder", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/magenta_stained_glass", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/magenta_stained_hardened_clay", "minecraft:recipes/root", []string{
		"has_hardened_clay",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/magenta_wool", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/magma", "minecraft:recipes/root", []string{
		"has_magma_cream",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/melon_block", "minecraft:recipes/root", []string{
		"has_melon",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/mossy_cobblestone", "minecraft:recipes/root", []string{
		"has_vine",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/mossy_stonebrick", "minecraft:recipes/root", []string{
		"has_mossy_cobblestone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/nether_brick", "minecraft:recipes/root", []string{
		"has_netherbrick",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/nether_brick_slab", "minecraft:recipes/root", []string{
		"has_nether_brick",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/nether_brick_stairs", "minecraft:recipes/root", []string{
		"has_nether_brick",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/nether_wart_block", "minecraft:recipes/root", []string{
		"has_nether_wart",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/oak_planks", "minecraft:recipes/root", []string{
		"has_log",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/oak_stairs", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/building_blocks/oak_wooden_slab", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/building_blocks/orange_concrete_powder", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/orange_stained_glass", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/orange_stained_hardened_clay", "minecraft:recipes/root", []string{
		"has_hardened_clay",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/orange_wool", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/pillar_quartz_block", "minecraft:recipes/root", []string{
		"has_quartz_block",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/pink_concrete_powder", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/pink_stained_glass", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/pink_stained_hardened_clay", "minecraft:recipes/root", []string{
		"has_hardened_clay",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/pink_wool", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/polished_andesite", "minecraft:recipes/root", []string{
		"has_stone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/polished_diorite", "minecraft:recipes/root", []string{
		"has_stone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/polished_granite", "minecraft:recipes/root", []string{
		"has_stone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/prismarine", "minecraft:recipes/root", []string{
		"has_prismarine_shard",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/prismarine_bricks", "minecraft:recipes/root", []string{
		"has_prismarine_shard",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/purple_concrete_powder", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/purple_stained_glass", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/purple_stained_hardened_clay", "minecraft:recipes/root", []string{
		"has_hardened_clay",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/purple_wool", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/purpur_block", "minecraft:recipes/root", []string{
		"has_chorus_fruit_popped",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/purpur_pillar", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_purpur_block",
	}, requireAny},
	{"minecraft:recipes/building_blocks/purpur_slab", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_purpur_block",
	}, requireAny},
	{"minecraft:recipes/building_blocks/purpur_stairs", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_purpur_block",
	}, requireAny},
	{"minecraft:recipes/building_blocks/quartz_block", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_quartz",
	}, requireAny},
	{"minecraft:recipes/building_blocks/quartz_slab", "minecraft:recipes/root", []string{
		"has_quartz_block",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/quartz_stairs", "minecraft:recipes/root", []string{
		"has_quartz_block",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/red_concrete_powder", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/red_nether_brick", "minecraft:recipes/root", []string{
		"has_nether_wart",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/red_sandstone", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
	}, requireAny},
	{"minecraft:recipes/building_blocks/red_sandstone_slab", "minecraft:recipes/root", []string{
		"has_red_sandstone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/red_sandstone_stairs", "minecraft:recipes/root", []string{
		"has_red_sandstone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/red_stained_glass", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/red_stained_hardened_clay", "minecraft:recipes/root", []string{
		"has_hardened_clay",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/red_wool", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/sandstone", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
	}, requireAny},
	{"minecraft:recipes/building_blocks/sandstone_slab", "minecraft:recipes/root", []string{
		"has_sandstone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/sandstone_stairs", "minecraft:recipes/root", []string{
		"has_sandstone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/sea_lantern", "minecraft:recipes/root", []string{
		"has_prismarine_crystals",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/smooth_red_sandstone", "minecraft:recipes/root", []string{
		"has_red_sandstone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/smooth_sandstone", "minecraft:recipes/root", []string{
		"has_sandstone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/snow", "minecraft:recipes/root", []string{
		"has_snowball",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/spruce_planks", "minecraft:recipes/root", []string{
		"has_log",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/spruce_stairs", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/building_blocks/spruce_wooden_slab", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/building_blocks/stone_brick_slab", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_stonebrick",
	}, requireAny},
	{"minecraft:recipes/building_blocks/stone_brick_stairs", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_stonebrick",
	}, requireAny},
	{"minecraft:recipes/building_blocks/stone_slab", "minecraft:recipes/root", []string{
		"has_stone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/stone_stairs", "minecraft:recipes/root", []string{
		"has_cobblestone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/stonebrick", "minecraft:recipes/root", []string{
		"has_stone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/string_to_wool", "minecraft:recipes/root", []string{
		"has_string",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/white_concrete_powder", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/white_stained_glass", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/white_stained_hardened_clay", "minecraft:recipes/root", []string{
		"has_hardened_clay",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/yellow_concrete_powder", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_sand",
		"has_gravel",
	}, requireAny},
	{"minecraft:recipes/building_blocks/yellow_stained_glass", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/yellow_stained_hardened_clay", "minecraft:recipes/root", []string{
		"has_hardened_clay",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/building_blocks/yellow_wool", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/combat/arrow", "minecraft:recipes/root", []string{
		"has_feather",
		"has_flint",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/combat/bow", "minecraft:recipes/root", []string{
		"has_string",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/combat/diamond_boots", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_diamond",
	}, requireAny},
	{"minecraft:recipes/combat/diamond_chestplate", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_diamond",
	}, requireAny},
	{"minecraft:recipes/combat/diamond_helmet", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_diamond",
	}, requireAny},
	{"minecraft:recipes/combat/diamond_leggings", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_diamond",
	}, requireAny},
	{"minecraft:recipes/combat/diamond_sword", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_diamond",
	}, requireAny},
	{"minecraft:recipes/combat/golden_boots", "minecraft:recipes/root", []string{
		"has_gold_ingot",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/combat/golden_chestplate", "minecraft:recipes/root", []string{
		"has_gold_ingot",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/combat/golden_helmet", "minecraft:recipes/root", []string{
		"has_gold_ingot",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/combat/golden_leggings", "minecraft:recipes/root", []string{
		"has_gold_ingot",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/combat/golden_sword", "minecraft:recipes/root", []string{
		"has_gold_ingot",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/combat/iron_boots", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/combat/iron_chestplate", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/combat/iron_helmet", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/combat/iron_leggings", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/combat/iron_sword", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/combat/leather_boots", "minecraft:recipes/root", []string{
		"has_leather",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/combat/leather_chestplate", "minecraft:recipes/root", []string{
		"has_leather",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/combat/leather_helmet", "minecraft:recipes/root", []string{
		"has_leather",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/combat/leather_leggings", "minecraft:recipes/root", []string{
		"has_leather",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/combat/shield", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/combat/spectral_arrow", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_glowstone_dust",
	}, requireAny},
	{"minecraft:recipes/combat/stone_sword", "minecraft:recipes/root", []string{
		"has_cobblestone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/combat/wooden_sword", "minecraft:recipes/root", []string{
		"has_stick",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/acacia_fence", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/decorations/anvil", "minecraft:recipes/root", []string{
		"has_iron_block",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/armor_stand", "minecraft:recipes/root", []string{
		"has_stone_slab",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/birch_fence", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/decorations/black_banner", "minecraft:recipes/root", []string{
		"has_black_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/black_bed", "minecraft:recipes/root", []string{
		"has_black_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/black_bed_from_white_bed", "minecraft:recipes/root", []string{
		"has_bed",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/black_carpet", "minecraft:recipes/root", []string{
		"has_black_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/black_stained_glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/blue_banner", "minecraft:recipes/root", []string{
		"has_blue_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/blue_bed", "minecraft:recipes/root", []string{
		"has_blue_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/blue_bed_from_white_bed", "minecraft:recipes/root", []string{
		"has_bed",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/blue_carpet", "minecraft:recipes/root", []string{
		"has_blue_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/blue_stained_glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/brown_banner", "minecraft:recipes/root", []string{
		"has_brown_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/brown_bed", "minecraft:recipes/root", []string{
		"has_brown_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/brown_bed_from_white_bed", "minecraft:recipes/root", []string{
		"has_bed",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/brown_carpet", "minecraft:recipes/root", []string{
		"has_brown_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/brown_stained_glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/chest", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_lots_of_items",
	}, requireAny},
	{"minecraft:recipes/decorations/cobblestone_wall", "minecraft:recipes/root", []string{
		"has_cobblestone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/crafting_table", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/decorations/cyan_banner", "minecraft:recipes/root", []string{
		"has_cyan_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/cyan_bed", "minecraft:recipes/root", []string{
		"has_cyan_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/cyan_bed_from_white_bed", "minecraft:recipes/root", []string{
		"has_bed",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/cyan_carpet", "minecraft:recipes/root", []string{
		"has_cyan_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/cyan_stained_glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/dark_oak_fence", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/decorations/enchanting_table", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_obsidian",
	}, requireAny},
	{"minecraft:recipes/decorations/end_crystal", "minecraft:recipes/root", []string{
		"has_ender_eye",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/end_rod", "minecraft:recipes/root", []string{
		"has_chorus_fruit_popped",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/ender_chest", "minecraft:recipes/root", []string{
		"has_ender_eye",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/fence", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/decorations/flower_pot", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_brick",
	}, requireAny},
	{"minecraft:recipes/decorations/furnace", "minecraft:recipes/root", []string{
		"has_cobblestone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/gray_banner", "minecraft:recipes/root", []string{
		"has_gray_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/gray_bed", "minecraft:recipes/root", []string{
		"has_gray_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/gray_bed_from_white_bed", "minecraft:recipes/root", []string{
		"has_bed",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/gray_carpet", "minecraft:recipes/root", []string{
		"has_gray_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/gray_stained_glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/green_banner", "minecraft:recipes/root", []string{
		"has_green_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/green_bed", "minecraft:recipes/root", []string{
		"has_green_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/green_bed_from_white_bed", "minecraft:recipes/root", []string{
		"has_bed",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/green_carpet", "minecraft:recipes/root", []string{
		"has_green_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/green_stained_glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/iron_bars", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/decorations/item_frame", "minecraft:recipes/root", []string{
		"has_leather",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/jukebox", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_diamond",
	}, requireAny},
	{"minecraft:recipes/decorations/jungle_fence", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/decorations/ladder", "minecraft:recipes/root", []string{
		"has_stick",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/light_blue_banner", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_light_blue_wool",
	}, requireAny},
	{"minecraft:recipes/decorations/light_blue_bed", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_light_blue_wool",
	}, requireAny},
	{"minecraft:recipes/decorations/light_blue_bed_from_white_bed", "minecraft:recipes/root", []string{
		"has_bed",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/light_blue_carpet", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_light_blue_wool",
	}, requireAny},
	{"minecraft:recipes/decorations/light_blue_stained_glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/light_gray_banner", "minecraft:recipes/root", []string{
		"has_light_gray_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/light_gray_bed", "minecraft:recipes/root", []string{
		"has_light_gray_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/light_gray_bed_from_white_bed", "minecraft:recipes/root", []string{
		"has_bed",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/light_gray_carpet", "minecraft:recipes/root", []string{
		"has_light_gray_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/light_gray_stained_glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/lime_banner", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_lime_wool",
	}, requireAny},
	{"minecraft:recipes/decorations/lime_bed", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_lime_wool",
	}, requireAny},
	{"minecraft:recipes/decorations/lime_bed_from_white_bed", "minecraft:recipes/root", []string{
		"has_bed",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/lime_carpet", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_lime_wool",
	}, requireAny},
	{"minecraft:recipes/decorations/lime_stained_glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/magenta_banner", "minecraft:recipes/root", []string{
		"has_magenta_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/magenta_bed", "minecraft:recipes/root", []string{
		"has_magenta_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/magenta_bed_from_white_bed", "minecraft:recipes/root", []string{
		"has_bed",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/magenta_carpet", "minecraft:recipes/root", []string{
		"has_magenta_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/magenta_stained_glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/mossy_cobblestone_wall", "minecraft:recipes/root", []string{
		"has_mossy_cobblestone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/nether_brick_fence", "minecraft:recipes/root", []string{
		"has_nether_brick",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/orange_banner", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_orange_wool",
	}, requireAny},
	{"minecraft:recipes/decorations/orange_bed", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_orange_wool",
	}, requireAny},
	{"minecraft:recipes/decorations/orange_bed_from_white_bed", "minecraft:recipes/root", []string{
		"has_bed",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/orange_carpet", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_orange_wool",
	}, requireAny},
	{"minecraft:recipes/decorations/orange_stained_glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/painting", "minecraft:recipes/root", []string{
		"has_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/pink_banner", "minecraft:recipes/root", []string{
		"has_pink_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/pink_bed", "minecraft:recipes/root", []string{
		"has_pink_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/pink_bed_from_white_bed", "minecraft:recipes/root", []string{
		"has_bed",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/pink_carpet", "minecraft:recipes/root", []string{
		"has_pink_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/pink_stained_glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/purple_banner", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_purple_wool",
	}, requireAny},
	{"minecraft:recipes/decorations/purple_bed", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_purple_wool",
	}, requireAny},
	{"minecraft:recipes/decorations/purple_bed_from_white_bed", "minecraft:recipes/root", []string{
		"has_bed",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/purple_carpet", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_purple_wool",
	}, requireAny},
	{"minecraft:recipes/decorations/purple_shulker_box", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_shulker_shell",
	}, requireAny},
	{"minecraft:recipes/decorations/purple_stained_glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/red_banner", "minecraft:recipes/root", []string{
		"has_red_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/red_bed", "minecraft:recipes/root", []string{
		"has_red_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/red_bed_from_white_bed", "minecraft:recipes/root", []string{
		"has_bed",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/red_carpet", "minecraft:recipes/root", []string{
		"has_red_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/red_stained_glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/sign", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/decorations/slime", "minecraft:recipes/root", []string{
		"has_at_least_9_slime_ball",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/snow_layer", "minecraft:recipes/root", []string{
		"has_snowball",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/spruce_fence", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/decorations/torch", "minecraft:recipes/root", []string{
		"has_stone_pickaxe",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/white_banner", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/white_bed", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/white_carpet", "minecraft:recipes/root", []string{
		"has_white_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/white_stained_glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/yellow_banner", "minecraft:recipes/root", []string{
		"has_yellow_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/yellow_bed", "minecraft:recipes/root", []string{
		"has_yellow_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/yellow_bed_from_white_bed", "minecraft:recipes/root", []string{
		"has_bed",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/yellow_carpet", "minecraft:recipes/root", []string{
		"has_yellow_wool",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/decorations/yellow_stained_glass_pane", "minecraft:recipes/root", []string{
		"has_glass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/food/beetroot_soup", "minecraft:recipes/root", []string{
		"has_beetroot",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/food/bread", "minecraft:recipes/root", []string{
		"has_wheat",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/food/cake", "minecraft:recipes/root", []string{"has_the_recipe", "has_egg"}, requireAny},
	{"minecraft:recipes/food/cookie", "minecraft:recipes/root", []string{
		"has_cocoa",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/food/golden_apple", "minecraft:recipes/root", []string{
		"has_gold_ingot",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/food/mushroom_stew", "minecraft:recipes/root", []string{
		"has_bowl",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/food/pumpkin_pie", "minecraft:recipes/root", []string{
		"has_pumpkin",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/food/rabbit_stew_from_brown_mushroom", "minecraft:recipes/root", []string{
		"has_cooked_rabbit",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/food/rabbit_stew_from_red_mushroom", "minecraft:recipes/root", []string{
		"has_cooked_rabbit",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/beacon", "minecraft:recipes/root", []string{
		"has_nether_star",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/bone_meal_from_block", "minecraft:recipes/root", []string{
		"has_bone_block",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/bone_meal_from_bone", "minecraft:recipes/root", []string{
		"has_bone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/book", "minecraft:recipes/root", []string{
		"has_paper",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/bowl", "minecraft:recipes/root", []string{
		"has_brown_mushroom",
		"has_red_mushroom",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/bucket", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/misc/coal", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_coal_block",
	}, requireAny},
	{"minecraft:recipes/misc/cyan_dye", "minecraft:recipes/root", []string{
		"has_green_dye",
		"has_lapis",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/diamond", "minecraft:recipes/root", []string{
		"has_diamond_block",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/emerald", "minecraft:recipes/root", []string{
		"has_emerald_block",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/ender_eye", "minecraft:recipes/root", []string{
		"has_blaze_powder",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/fire_charge", "minecraft:recipes/root", []string{
		"has_blaze_powder",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/gold_ingot_from_block", "minecraft:recipes/root", []string{
		"has_gold_block",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/gold_ingot_from_nuggets", "minecraft:recipes/root", []string{
		"has_at_least_9_gold_nugget",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/gold_nugget", "minecraft:recipes/root", []string{
		"has_gold_ingot",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/gray_dye", "minecraft:recipes/root", []string{
		"has_bonemeal",
		"has_the_recipe",
		"has_black_dye",
	}, requireAny},
	{"minecraft:recipes/misc/iron_ingot_from_block", "minecraft:recipes/root", []string{
		"has_iron_block",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/iron_ingot_from_nuggets", "minecraft:recipes/root", []string{
		"has_at_least_9_iron_nugget",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/iron_nugget", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/misc/lapis_lazuli", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_lapis_block",
	}, requireAny},
	{"minecraft:recipes/misc/leather", "minecraft:recipes/root", []string{
		"has_rabbit_hide",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/light_blue_dye_from_blue_orchid", "minecraft:recipes/root", []string{
		"has_red_flower",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/light_blue_dye_from_lapis_bonemeal", "minecraft:recipes/root", []string{
		"has_lapis",
		"has_bonemeal",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/light_gray_dye_from_azure_bluet", "minecraft:recipes/root", []string{
		"has_red_flower",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/light_gray_dye_from_gray_bonemeal", "minecraft:recipes/root", []string{
		"has_gray_dye",
		"has_bonemeal",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/light_gray_dye_from_ink_bonemeal", "minecraft:recipes/root", []string{
		"has_bonemeal",
		"has_the_recipe",
		"has_black_dye",
	}, requireAny},
	{"minecraft:recipes/misc/light_gray_dye_from_oxeye_daisy", "minecraft:recipes/root", []string{
		"has_red_flower",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/light_gray_dye_from_white_tulip", "minecraft:recipes/root", []string{
		"has_red_flower",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/lime_dye", "minecraft:recipes/root", []string{
		"has_green_dye",
		"has_bonemeal",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/magenta_dye_from_allium", "minecraft:recipes/root", []string{
		"has_red_flower",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/magenta_dye_from_lapis_ink_bonemeal", "minecraft:recipes/root", []string{
		"has_lapis",
		"has_bonemeal",
		"has_the_recipe",
		"has_black_dye",
	}, requireAny},
	{"minecraft:recipes/misc/magenta_dye_from_lapis_red_pink", "minecraft:recipes/root", []string{
		"has_pink_dye",
		"has_lapis",
		"has_red_dye",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/magenta_dye_from_lilac", "minecraft:recipes/root", []string{
		"has_double_plant",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/magenta_dye_from_purple_and_pink", "minecraft:recipes/root", []string{
		"has_pink_dye",
		"has_purple_dye",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/map", "minecraft:recipes/root", []string{
		"has_compass",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/melon_seeds", "minecraft:recipes/root", []string{
		"has_melon",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/orange_dye_from_orange_tulip", "minecraft:recipes/root", []string{
		"has_red_flower",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/orange_dye_from_red_yellow", "minecraft:recipes/root", []string{
		"has_red_dye",
		"has_the_recipe",
		"has_yellow_dye",
	}, requireAny},
	{"minecraft:recipes/misc/paper", "minecraft:recipes/root", []string{
		"has_reeds",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/pink_dye_from_peony", "minecraft:recipes/root", []string{
		"has_double_plant",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/pink_dye_from_pink_tulip", "minecraft:recipes/root", []string{
		"has_red_flower",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/pink_dye_from_red_bonemeal", "minecraft:recipes/root", []string{
		"has_bonemeal",
		"has_red_dye",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/pumpkin_seeds", "minecraft:recipes/root", []string{
		"has_pumpkin",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/purple_dye", "minecraft:recipes/root", []string{
		"has_lapis",
		"has_red_dye",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/red_dye_from_beetroot", "minecraft:recipes/root", []string{
		"has_beetroot",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/red_dye_from_poppy", "minecraft:recipes/root", []string{
		"has_red_flower",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/red_dye_from_rose_bush", "minecraft:recipes/root", []string{
		"has_double_plant",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/red_dye_from_tulip", "minecraft:recipes/root", []string{
		"has_red_flower",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/slime_ball", "minecraft:recipes/root", []string{
		"has_slime",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/stick", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/misc/sugar", "minecraft:recipes/root", []string{
		"has_reeds",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/wheat", "minecraft:recipes/root", []string{
		"has_hay_block",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/writable_book", "minecraft:recipes/root", []string{
		"has_book",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/yellow_dye_from_dandelion", "minecraft:recipes/root", []string{
		"has_yellow_flower",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/misc/yellow_dye_from_sunflower", "minecraft:recipes/root", []string{
		"has_double_plant",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/redstone/acacia_door", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/redstone/acacia_fence_gate", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/redstone/birch_door", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/redstone/birch_fence_gate", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/redstone/comparator", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_quartz",
	}, requireAny},
	{"minecraft:recipes/redstone/dark_oak_door", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/redstone/dark_oak_fence_gate", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/redstone/daylight_detector", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_quartz",
	}, requireAny},
	{"minecraft:recipes/redstone/dispenser", "minecraft:recipes/root", []string{
		"has_bow",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/redstone/dropper", "minecraft:recipes/root", []string{
		"has_redstone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/redstone/fence_gate", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/redstone/heavy_weighted_pressure_plate", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/redstone/hopper", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/redstone/iron_door", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/redstone/iron_trapdoor", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/redstone/jungle_door", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/redstone/jungle_fence_gate", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/redstone/lever", "minecraft:recipes/root", []string{
		"has_cobblestone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/redstone/light_weighted_pressure_plate", "minecraft:recipes/root", []string{
		"has_gold_ingot",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/redstone/noteblock", "minecraft:recipes/root", []string{
		"has_redstone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/redstone/observer", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_quartz",
	}, requireAny},
	{"minecraft:recipes/redstone/piston", "minecraft:recipes/root", []string{
		"has_redstone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/redstone/redstone", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_redstone_block",
	}, requireAny},
	{"minecraft:recipes/redstone/redstone_block", "minecraft:recipes/root", []string{
		"has_at_least_9_redstone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/redstone/redstone_lamp", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_glowstone",
	}, requireAny},
	{"minecraft:recipes/redstone/redstone_torch", "minecraft:recipes/root", []string{
		"has_redstone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/redstone/repeater", "minecraft:recipes/root", []string{
		"has_redstone_torch",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/redstone/spruce_door", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/redstone/spruce_fence_gate", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/redstone/sticky_piston", "minecraft:recipes/root", []string{
		"has_slime_ball",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/redstone/stone_button", "minecraft:recipes/root", []string{
		"has_stone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/redstone/stone_pressure_plate", "minecraft:recipes/root", []string{
		"has_stone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/redstone/tnt", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_gunpowder",
	}, requireAny},
	{"minecraft:recipes/redstone/trapdoor", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/redstone/trapped_chest", "minecraft:recipes/root", []string{
		"has_tripwire_hook",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/redstone/tripwire_hook", "minecraft:recipes/root", []string{
		"has_string",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/redstone/wooden_button", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/redstone/wooden_door", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/redstone/wooden_pressure_plate", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_planks",
	}, requireAny},
	{"minecraft:recipes/root", "", []string{"impossible"}, requireAll},
	{"minecraft:recipes/tools/clock", "minecraft:recipes/root", []string{
		"has_redstone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/tools/compass", "minecraft:recipes/root", []string{
		"has_redstone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/tools/diamond_axe", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_diamond",
	}, requireAny},
	{"minecraft:recipes/tools/diamond_hoe", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_diamond",
	}, requireAny},
	{"minecraft:recipes/tools/diamond_pickaxe", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_diamond",
	}, requireAny},
	{"minecraft:recipes/tools/diamond_shovel", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_diamond",
	}, requireAny},
	{"minecraft:recipes/tools/fishing_rod", "minecraft:recipes/root", []string{
		"has_string",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/tools/flint_and_steel", "minecraft:recipes/root", []string{
		"has_flint",
		"has_the_recipe",
		"has_obsidian",
	}, requireAny},
	{"minecraft:recipes/tools/golden_axe", "minecraft:recipes/root", []string{
		"has_gold_ingot",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/tools/golden_hoe", "minecraft:recipes/root", []string{
		"has_gold_ingot",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/tools/golden_pickaxe", "minecraft:recipes/root", []string{
		"has_gold_ingot",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/tools/golden_shovel", "minecraft:recipes/root", []string{
		"has_gold_ingot",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/tools/iron_axe", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/tools/iron_hoe", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/tools/iron_pickaxe", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/tools/iron_shovel", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/tools/lead", "minecraft:recipes/root", []string{
		"has_slime_ball",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/tools/shears", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/tools/stone_axe", "minecraft:recipes/root", []string{
		"has_cobblestone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/tools/stone_hoe", "minecraft:recipes/root", []string{
		"has_cobblestone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/tools/stone_pickaxe", "minecraft:recipes/root", []string{
		"has_cobblestone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/tools/stone_shovel", "minecraft:recipes/root", []string{
		"has_cobblestone",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/tools/wooden_axe", "minecraft:recipes/root", []string{
		"has_stick",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/tools/wooden_hoe", "minecraft:recipes/root", []string{
		"has_stick",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/tools/wooden_pickaxe", "minecraft:recipes/root", []string{
		"has_stick",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/tools/wooden_shovel", "minecraft:recipes/root", []string{
		"has_stick",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/transportation/acacia_boat", "minecraft:recipes/root", []string{
		"entered_water",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/transportation/activator_rail", "minecraft:recipes/root", []string{
		"has_rail",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/transportation/birch_boat", "minecraft:recipes/root", []string{
		"entered_water",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/transportation/boat", "minecraft:recipes/root", []string{
		"entered_water",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/transportation/carrot_on_a_stick", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_carrot",
	}, requireAny},
	{"minecraft:recipes/transportation/chest_minecart", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_minecart",
	}, requireAny},
	{"minecraft:recipes/transportation/dark_oak_boat", "minecraft:recipes/root", []string{
		"entered_water",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/transportation/detector_rail", "minecraft:recipes/root", []string{
		"has_rail",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/transportation/furnace_minecart", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_minecart",
	}, requireAny},
	{"minecraft:recipes/transportation/golden_rail", "minecraft:recipes/root", []string{
		"has_rail",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/transportation/hopper_minecart", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_minecart",
	}, requireAny},
	{"minecraft:recipes/transportation/jungle_boat", "minecraft:recipes/root", []string{
		"entered_water",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/transportation/minecart", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_iron_ingot",
	}, requireAny},
	{"minecraft:recipes/transportation/rail", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_minecart",
	}, requireAny},
	{"minecraft:recipes/transportation/spruce_boat", "minecraft:recipes/root", []string{
		"entered_water",
		"has_the_recipe",
	}, requireAny},
	{"minecraft:recipes/transportation/tnt_minecart", "minecraft:recipes/root", []string{
		"has_the_recipe",
		"has_minecart",
	}, requireAny},
	{"minecraft:story/cure_zombie_villager", "minecraft:story/enter_the_nether", []string{
		"cured_zombie",
	}, requireAll},
	{"minecraft:story/deflect_arrow", "minecraft:story/obtain_armor", []string{
		"deflected_projectile",
	}, requireAll},
	{"minecraft:story/enchant_item", "minecraft:story/mine_diamond", []string{"enchanted_item"}, requireAll},
	{"minecraft:story/enter_the_end", "minecraft:story/follow_ender_eye", []string{"entered_end"}, requireAll},
	{"minecraft:story/enter_the_nether", "minecraft:story/form_obsidian", []string{
		"entered_nether",
	}, requireAll},
	{"minecraft:story/follow_ender_eye", "minecraft:story/enter_the_nether", []string{
		"in_stronghold",
	}, requireAll},
	{"minecraft:story/form_obsidian", "minecraft:story/lava_bucket", []string{"obsidian"}, requireAll},
	{"minecraft:story/iron_tools", "minecraft:story/smelt_iron", []string{"iron_pickaxe"}, requireAll},
	{"minecraft:story/lava_bucket", "minecraft:story/smelt_iron", []string{"lava_bucket"}, requireAll},
	{"minecraft:story/mine_diamond", "minecraft:story/iron_tools", []string{"diamond"}, requireAll},
	{"minecraft:story/mine_stone", "minecraft:story/root", []string{"get_stone"}, requireAll},
	{"minecraft:story/obtain_armor", "minecraft:story/smelt_iron", []string{
		"iron_helmet",
		"iron_chestplate",
		"iron_leggings",
		"iron_boots",
	}, requireAny},
	{"minecraft:story/root", "", []string{"crafting_table"}, requireAll},
	{"minecraft:story/shiny_gear", "minecraft:story/mine_diamond", []string{
		"diamond_helmet",
		"diamond_chestplate",
		"diamond_leggings",
		"diamond_boots",
	}, requireAny},
	{"minecraft:story/smelt_iron", "minecraft:story/upgrade_tools", []string{"iron"}, requireAll},
	{"minecraft:story/upgrade_tools", "minecraft:story/mine_stone", []string{"stone_pickaxe"}, requireAll},
}

// advancement is what the command needs to know about one advancement.
//
// The built-in advancements of the server jar are the entries of
// advancementRegistry; a datapack adds more, and those are read from the
// world directory. Both are presented as an advancement, so the command
// does not have to care where one came from.
type advancement struct {
	// name is the namespaced id, e.g. "minecraft:story/root".
	name string
	// parent is the parent's namespaced id. It is empty for a root
	// advancement, one that has no parent.
	parent string
	// criteria are the criterion names.
	criteria []string
	// groups are the requirement groups: every group has to be satisfied,
	// and a group is satisfied as soon as one of its criteria is done.
	groups [][]string
	// recipe is the recipe the advancement watches, if it watches one.
	recipe *advancementRecipe
	// experience is the experience the advancement hands out when it is
	// completed. It is zero for the advancements without such a reward.
	experience int
}

// advancementRecipe is the recipe side of an advancement.
//
// The recipe advancements of 1.12.2 watch a recipe: the criterion named
// "has_the_recipe" is completed by the "minecraft:recipe_unlocked" trigger
// when the player learns that recipe. Completing an advancement hands out
// the recipes of its "rewards" block, and learning a recipe completes the
// advancement that watches it.
type advancementRecipe struct {
	// criterion is the criterion that watches the recipe.
	criterion string
	// watch is the recipe the criterion watches, e.g. "minecraft:acacia_planks".
	watch string
	// rewards are the recipes the advancement hands out when it is done.
	rewards []string
}

// advancementSet is the set of advancements of one world.
//
// It holds the advancements of the server jar and the ones the world's
// datapack adds.
type advancementSet struct {
	// root is the world directory the set was read from.
	root string
	// all holds every advancement of the set, in the order the command
	// walks them.
	all []*advancement
	// index holds the advancements by name.
	index map[string]*advancement
	// children holds the children of every advancement, by the name of the
	// parent.
	children map[string][]*advancement
	// watchers holds the advancements that watch a recipe, by the recipe.
	watchers map[string][]*advancement
}

// newAdvancementSet builds the set of a world.
//
// The advancements of the server jar and the ones of the datapack are put
// into the set in the order 1.12.2 walks them, which is the order of the
// map the server keeps them in.
func newAdvancementSet(root string, custom []*advancement) *advancementSet {
	builtIn := make([]*advancement, 0, len(advancementRegistry))
	for i := range advancementRegistry {
		def := &advancementRegistry[i]
		builtIn = append(builtIn, &advancement{
			name:       def.name,
			parent:     def.parent,
			criteria:   def.criteria,
			groups:     def.requirementGroups(),
			recipe:     builtInAdvancementRecipe(def.name),
			experience: builtInAdvancementExperience[def.name],
		})
	}
	s := &advancementSet{
		root:     root,
		all:      advancementOrder(builtIn, custom),
		index:    make(map[string]*advancement, len(builtIn)+len(custom)),
		children: map[string][]*advancement{},
		watchers: map[string][]*advancement{},
	}
	for _, a := range s.all {
		s.index[a.name] = a
		s.children[a.parent] = append(s.children[a.parent], a)
		if a.recipe != nil && a.recipe.watch != "" {
			s.watchers[a.recipe.watch] = append(s.watchers[a.recipe.watch], a)
		}
	}
	for parent := range s.children {
		list := s.children[parent]
		sort.Slice(list, func(i, j int) bool { return list[i].name < list[j].name })
	}
	return s
}

// lookup finds the advancement a command named.
//
// 1.12.2 reads the name as a `MinecraftKey`, so the namespace defaults to
// "minecraft" and the whole text is lower-cased: "story/root" and
// "minecraft:story/root" are the same advancement.
func (s *advancementSet) lookup(name string) (*advancement, bool) {
	a, ok := s.index[advancementKey(name)]
	return a, ok
}

// advancementKey turns the text of an advancement name into the key the
// registry is indexed by.
//
// `MinecraftKey` splits the text at the first colon and only reads a
// namespace when the colon is neither the first nor the second character,
// so ":story/root" and "a:story/root" are both "minecraft:story/root".
func advancementKey(name string) string {
	namespace, key := world.EntityTypeKey(name)
	return namespace + ":" + key
}

// childrenOf returns the advancements whose parent is the given one.
//
// The result is sorted by name, so the order in which the descendants of
// an advancement are granted does not depend on the order of the registry.
func (s *advancementSet) childrenOf(parent string) []*advancement {
	return s.children[parent]
}

// ancestorsOf returns the parents of an advancement, closest one first.
//
// 1.12.2 walks the parents of an advancement in the same order, and the
// order matters only for the order in which the advancements are granted.
func (s *advancementSet) ancestorsOf(a *advancement) []*advancement {
	var out []*advancement
	for parent := a.parent; parent != ""; {
		p, ok := s.index[parent]
		if !ok {
			break
		}
		out = append(out, p)
		parent = p.parent
	}
	return out
}

// requirementGroups returns the requirement groups of a built-in
// advancement.
//
// The built-in advancements only ever need one group or none: 438 of them
// declare a single group holding every criterion, so one done criterion is
// enough (requireAny), and the other 50 declare no "requirements" at all,
// so the server makes one group per criterion and every criterion has to
// be done (requireAll).
func (d *advancementDef) requirementGroups() [][]string {
	switch d.kind {
	case requireAny:
		return [][]string{d.criteria}
	default:
		out := make([][]string, 0, len(d.criteria))
		for _, criterion := range d.criteria {
			out = append(out, []string{criterion})
		}
		return out
	}
}

// done reports whether the player has completed the advancement.
//
// 1.12.2 reads the "requirements" array of the advancement JSON: every
// group has to be satisfied, and a group is satisfied as soon as one of
// its criteria is done. An advancement without any criterion is never
// done.
func (a *advancement) done(obtained map[string]bool) bool {
	if len(a.groups) == 0 {
		return false
	}
	for _, group := range a.groups {
		satisfied := false
		for _, criterion := range group {
			if obtained[criterion] {
				satisfied = true
				break
			}
		}
		if !satisfied {
			return false
		}
	}
	return true
}

// hasCriterion reports whether the advancement has a criterion of that
// name.
func (a *advancement) hasCriterion(name string) bool {
	for _, c := range a.criteria {
		if c == name {
			return true
		}
	}
	return false
}

// advancementStore holds the advancement progress of the players of one
// world.
//
// 1.12.2 keeps the progress of a player in `advancements/<uuid>.json`, in
// the directory of the world. The world state of gomc lives in package
// world, which has no room for it, so the command layer keeps it here.
type advancementStore struct {
	mu sync.Mutex
	// root is the world directory the progress is read from and written
	// to. It is empty when the runtime has no world directory.
	root string
	// players holds the criteria every player has completed, keyed by the
	// player name and then by the advancement name.
	players map[string]map[string]map[string]bool
	// loaded holds the progress that was read from the world directory,
	// keyed by the uuid of the file it came from.
	loaded map[string]map[string]map[string]bool
	// attached holds the players whose progress was already taken out of
	// "loaded".
	attached map[string]bool
	// dirty holds the players whose progress changed since the last save.
	dirty map[string]bool
	// loadErrs holds the problems of the read of the world directory.
	loadErrs []error
	// reported tells whether the problems were already put in the log.
	reported bool
}

// newAdvancementStore returns the store of a world.
//
// The progress the world directory holds is read right away: the server
// reads the progress of a player when the player joins, and gomc reads it
// the first time the advancements of the world are needed.
func newAdvancementStore(root string) *advancementStore {
	loaded, errs := readAdvancementProgress(root)
	return &advancementStore{
		root:     root,
		players:  map[string]map[string]map[string]bool{},
		loaded:   loaded,
		attached: map[string]bool{},
		dirty:    map[string]bool{},
		loadErrs: errs,
	}
}

// attachLocked takes the progress of a player out of the files of the
// world.
//
// The files are named after the uuid of a player, and gomc derives that
// uuid from the name of the player, so the file of a player is found the
// first time the player is asked about.
//
// The caller holds the lock.
func (s *advancementStore) attachLocked(player string) {
	if s.attached[player] {
		return
	}
	s.attached[player] = true
	file, ok := s.loaded[world.DeriveEntityUUID(player)]
	if !ok {
		return
	}
	for advancement, criteria := range file {
		byAdvancement := s.players[player]
		if byAdvancement == nil {
			byAdvancement = map[string]map[string]bool{}
			s.players[player] = byAdvancement
		}
		if byAdvancement[advancement] == nil {
			byAdvancement[advancement] = map[string]bool{}
		}
		for criterion, done := range criteria {
			byAdvancement[advancement][criterion] = done
		}
	}
}

// obtained returns the criteria the player has completed in the
// advancement.
func (s *advancementStore) obtained(player, advancement string) map[string]bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attachLocked(player)
	out := map[string]bool{}
	for criterion, done := range s.players[player][advancement] {
		if done {
			out[criterion] = true
		}
	}
	return out
}

// grant completes a criterion and reports whether it was not completed
// before.
//
// 1.12.2 refuses to complete a criterion twice: `AdvancementProgress.a`
// only marks a criterion that is not obtained yet.
func (s *advancementStore) grant(player, advancement, criterion string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attachLocked(player)
	byAdvancement := s.players[player]
	if byAdvancement == nil {
		byAdvancement = map[string]map[string]bool{}
		s.players[player] = byAdvancement
	}
	criteria := byAdvancement[advancement]
	if criteria == nil {
		criteria = map[string]bool{}
		byAdvancement[advancement] = criteria
	}
	if criteria[criterion] {
		return false
	}
	criteria[criterion] = true
	s.dirty[player] = true
	return true
}

// revoke undoes a criterion and reports whether it was completed before.
//
// 1.12.2 only takes a criterion away that was obtained
// (`AdvancementProgress.b`).
func (s *advancementStore) revoke(player, advancement, criterion string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attachLocked(player)
	criteria := s.players[player][advancement]
	if criteria == nil || !criteria[criterion] {
		return false
	}
	delete(criteria, criterion)
	s.dirty[player] = true
	return true
}

// save writes the progress of a player to the world directory.
//
// The file is written when the progress changed. A file that cannot be
// written is reported and the progress stays in memory: the server logs
// the problem and keeps running as well.
func (s *advancementStore) save(set *advancementSet, player string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.dirty[player] {
		return nil
	}
	s.dirty[player] = false
	if s.root == "" {
		return nil
	}
	path := advancementProgressPath(s.root, player)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("couldn't save the advancements of %s to %s: %w", player, path, err)
	}
	if err := writeAdvancementProgress(path, set, s.players[player]); err != nil {
		return fmt.Errorf("couldn't save the advancements of %s to %s: %w", player, path, err)
	}
	return nil
}

// takeLoadErrors returns the problems of the read of the world directory.
//
// They are returned once: the server reports them when it reads the
// progress of a player, and gomc reports them the first time the
// advancements of the world are needed.
func (s *advancementStore) takeLoadErrors() []error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.reported {
		return nil
	}
	s.reported = true
	return s.loadErrs
}

// advancementStores holds the advancement state of every world.
//
// The state cannot live in world.World, so the command layer keeps it
// here. The stores are keyed by the world they belong to, so two worlds
// never share the progress of a player.
var advancementStores sync.Map

// advancementStoreOf returns the advancement state of a world.
//
// The progress of the players is read from the world directory the first
// time the state is needed.
func advancementStoreOf(w *world.World, root string) *advancementStore {
	if v, ok := advancementStores.Load(w); ok {
		store := v.(*advancementStore)
		if store.root == root {
			return store
		}
	}
	store := newAdvancementStore(root)
	advancementStores.Store(w, store)
	return store
}

// advancementUsageError builds the error 1.12.2 answers a syntax mistake
// it recognises as a usage mistake with.
//
// 1.12.2 throws an `ExceptionUsage` for those, and the server renders them
// with the "Usage: " prefix of `commands.generic.usage`.
func advancementUsageError(usage string) error {
	return &exec.Error{Msg: fmt.Sprintf(commandUsage("generic"), usage)}
}
