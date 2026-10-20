package v1_12_2

import (
	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// The experience side of the advancements.
//
// 1.12.2 hands out the experience of the "rewards" block of an advancement
// when the advancement is completed. The block holds an "experience" key
// with the number of experience *points*, not levels: the server calls
// `EntityHuman.giveExp(int)`, which adds the points to the total experience
// and moves the level up along the usual curve.
//
// Twelve of the 488 advancements of the server jar have such a reward.
//
// (Measured on a real 1.12.2 server: granting "minecraft:adventure/
// kill_all_mobs" -- 100 experience -- leaves the player with 100 total
// experience and level 7; the server says nothing about the experience in
// its answer to the command.)

// builtInAdvancementExperience is the experience the built-in advancements
// hand out.
//
// The data comes from the JSON files of the 1.12.2 server jar
// (`assets/minecraft/advancements/**.json`): the "experience" of the
// "rewards" block of each of the twelve files that has one.
var builtInAdvancementExperience = map[string]int{
	"minecraft:adventure/adventuring_time":  500,
	"minecraft:adventure/kill_all_mobs":     100,
	"minecraft:adventure/sniper_duel":       50,
	"minecraft:end/levitate":                50,
	"minecraft:husbandry/balanced_diet":     100,
	"minecraft:husbandry/break_diamond_hoe": 100,
	"minecraft:husbandry/bred_all_animals":  100,
	"minecraft:nether/all_effects":          1000,
	"minecraft:nether/all_potions":          100,
	"minecraft:nether/fast_travel":          100,
	"minecraft:nether/return_to_sender":     50,
	"minecraft:nether/uneasy_alliance":      100,
}

// giveAdvancementExperience hands the player the experience of an
// advancement.
//
// 1.12.2 gives experience *points*: `EntityHuman.giveExp(int)` adds them to
// the total experience of the player and moves the level up while the
// progress towards the next level is full. The three numbers live in the
// NBT of the player under "XpTotal", "XpLevel" and "XpP".
//
// The level curve is the one of `EntityHuman.getExpToLevel`: a level below
// 15 costs 7 + 2 * level points, a level below 30 costs 37 + 5 * (level - 15)
// points and every level above that costs 112 + 9 * (level - 30) points.
//
// (Measured on a real 1.12.2 server: 100 points put a fresh player on level
// 7 with 9 of the 21 points of the next level, 1150 points put the player on
// level 28.)
func giveAdvancementExperience(player *world.Entity, amount int) {
	if player == nil || player.NBT == nil || amount <= 0 {
		return
	}
	level := int(nbtIntValue(player.NBT, "XpLevel"))
	progress := float32(nbtFloatValue(player.NBT, "XpP"))
	total := int(nbtIntValue(player.NBT, "XpTotal"))

	// The same steps the server takes: the points go into the bar of the
	// current level first, and every full bar moves the player one level
	// up and leaves the rest of the points in the bar of the new level.
	progress += float32(amount) / float32(experienceToLevel(level))
	total += amount
	for progress >= 1 {
		progress = (progress - 1) * float32(experienceToLevel(level))
		level++
		progress /= float32(experienceToLevel(level))
	}

	player.NBT.Set("XpLevel", nbt.Int(int32(level)))
	player.NBT.Set("XpTotal", nbt.Int(int32(total)))
	player.NBT.Set("XpP", nbt.Float(progress))
}

// experienceToLevel returns the experience a player needs to go from the
// given level to the next one.
//
// This is `EntityHuman.getExpToLevel` of the 1.12.2 server.
func experienceToLevel(level int) int {
	switch {
	case level >= 30:
		return 112 + (level-30)*9
	case level >= 15:
		return 37 + (level-15)*5
	default:
		return 7 + level*2
	}
}

// nbtIntValue reads an integer out of a compound.
//
// A tag that is missing or holds another type counts as zero, the way the
// server reads the tags of a player.
func nbtIntValue(c *nbt.Compound, key string) int32 {
	if c == nil {
		return 0
	}
	if v, ok := c.Get(key).(nbt.Int); ok {
		return int32(v)
	}
	return 0
}

// nbtFloatValue reads a float out of a compound.
//
// A tag that is missing or holds another type counts as zero, the way the
// server reads the tags of a player.
func nbtFloatValue(c *nbt.Compound, key string) float32 {
	if c == nil {
		return 0
	}
	if v, ok := c.Get(key).(nbt.Float); ok {
		return float32(v)
	}
	return 0
}
