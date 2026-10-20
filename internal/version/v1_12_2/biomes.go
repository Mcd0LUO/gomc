package v1_12_2

import "strings"

// Biome is one entry of the 1.12.2 biome registry.
type Biome struct {
	// ID is the numeric id of the biome.
	//
	// 1.12.2 keeps the biomes in an array, so every biome has a number.
	// The numbers are the order the registry registers the biomes in.
	ID int
	// Name is the registry name of the biome, e.g. "plains".
	//
	// The name is the key the biome is registered under, so the full
	// name of a biome is "minecraft:<name>".
	Name string
	// DisplayName is the name the game shows for the biome, e.g.
	// "Plains".
	//
	// The name of a biome is not in the language file of 1.12.2: the
	// server keeps it in the biome itself, so the table below is the
	// only place the names are written down.
	DisplayName string
}

// Biomes lists the biome registry of 1.12.2 in id order.
//
// The table comes from the server's `BiomeBase` class, which registers
// every biome under its name and gives it the display name it shows.
// There are 62 of them: the last 21 are the mutated biomes, the variants
// the world generator makes of the biomes before them.
//
// (Measured by reading the `BiomeBase` class of the 1.12.2 server.)
var Biomes = []Biome{
	{0, "ocean", "Ocean"},
	{1, "plains", "Plains"},
	{2, "desert", "Desert"},
	{3, "extreme_hills", "Extreme Hills"},
	{4, "forest", "Forest"},
	{5, "taiga", "Taiga"},
	{6, "swampland", "Swampland"},
	{7, "river", "River"},
	{8, "hell", "Hell"},
	{9, "sky", "The End"},
	{10, "frozen_ocean", "FrozenOcean"},
	{11, "frozen_river", "FrozenRiver"},
	{12, "ice_flats", "Ice Plains"},
	{13, "ice_mountains", "Ice Mountains"},
	{14, "mushroom_island", "MushroomIsland"},
	{15, "mushroom_island_shore", "MushroomIslandShore"},
	{16, "beaches", "Beach"},
	{17, "desert_hills", "DesertHills"},
	{18, "forest_hills", "ForestHills"},
	{19, "taiga_hills", "TaigaHills"},
	{20, "smaller_extreme_hills", "Extreme Hills Edge"},
	{21, "jungle", "Jungle"},
	{22, "jungle_hills", "JungleHills"},
	{23, "jungle_edge", "JungleEdge"},
	{24, "deep_ocean", "Deep Ocean"},
	{25, "stone_beach", "Stone Beach"},
	{26, "cold_beach", "Cold Beach"},
	{27, "birch_forest", "Birch Forest"},
	{28, "birch_forest_hills", "Birch Forest Hills"},
	{29, "roofed_forest", "Roofed Forest"},
	{30, "taiga_cold", "Cold Taiga"},
	{31, "taiga_cold_hills", "Cold Taiga Hills"},
	{32, "redwood_taiga", "Mega Taiga"},
	{33, "redwood_taiga_hills", "Mega Taiga Hills"},
	{34, "extreme_hills_with_trees", "Extreme Hills+"},
	{35, "savanna", "Savanna"},
	{36, "savanna_rock", "Savanna Plateau"},
	{37, "mesa", "Mesa"},
	{38, "mesa_rock", "Mesa Plateau F"},
	{39, "mesa_clear_rock", "Mesa Plateau"},
	{40, "void", "The Void"},
	{41, "mutated_plains", "Sunflower Plains"},
	{42, "mutated_desert", "Desert M"},
	{43, "mutated_extreme_hills", "Extreme Hills M"},
	{44, "mutated_forest", "Flower Forest"},
	{45, "mutated_taiga", "Taiga M"},
	{46, "mutated_swampland", "Swampland M"},
	{47, "mutated_ice_flats", "Ice Plains Spikes"},
	{48, "mutated_jungle", "Jungle M"},
	{49, "mutated_jungle_edge", "JungleEdge M"},
	{50, "mutated_birch_forest", "Birch Forest M"},
	{51, "mutated_birch_forest_hills", "Birch Forest Hills M"},
	{52, "mutated_roofed_forest", "Roofed Forest M"},
	{53, "mutated_taiga_cold", "Cold Taiga M"},
	{54, "mutated_redwood_taiga", "Mega Spruce Taiga"},
	{55, "mutated_redwood_taiga_hills", "Redwood Taiga Hills M"},
	{56, "mutated_extreme_hills_with_trees", "Extreme Hills+ M"},
	{57, "mutated_savanna", "Savanna M"},
	{58, "mutated_savanna_rock", "Savanna Plateau M"},
	{59, "mutated_mesa", "Mesa (Bryce)"},
	{60, "mutated_mesa_rock", "Mesa Plateau F M"},
	{61, "mutated_mesa_clear_rock", "Mesa Plateau M"},
}

// BiomeByName returns the biome with the given registry name.
//
// The name may be given with its namespace, so both "plains" and
// "minecraft:plains" find the biome.
func BiomeByName(name string) (Biome, bool) {
	name = strings.TrimPrefix(name, "minecraft:")
	for _, b := range Biomes {
		if b.Name == name {
			return b, true
		}
	}
	return Biome{}, false
}

// BiomeByID returns the biome with the given id.
func BiomeByID(id int) (Biome, bool) {
	if id < 0 || id >= len(Biomes) {
		return Biome{}, false
	}
	b := Biomes[id]
	if b.ID != id {
		return Biome{}, false
	}
	return b, true
}
