package v1_12_2

import (
	"sort"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// The recipe side of the advancements of the server jar.
//
// 1.12.2's recipe advancements watch a recipe: the criterion named
// "has_the_recipe" is completed by the "minecraft:recipe_unlocked" trigger
// when the player learns that recipe. Completing an advancement hands out
// the recipes of its "rewards" block, and learning a recipe completes the
// advancement that watches it.
//
// The data comes from the JSON files of the 1.12.2 server jar
// (`assets/minecraft/advancements/recipes/**.json`).

// builtInAdvancementRecipe returns the recipe a built-in advancement is
// about, and the recipes it hands out.
//
// Every advancement below "minecraft:recipes/" but the root watches a
// recipe with the criterion "has_the_recipe", and the recipe is the one
// its name ends with: "minecraft:recipes/misc/bone_meal_from_block" is
// about "minecraft:bone_meal_from_block".
//
// 420 of the 432 recipe advancements hand out exactly the recipe they are
// about; the twelve of advancementRecipeExtras hand out one more.
func builtInAdvancementRecipe(name string) *advancementRecipe {
	const prefix = "minecraft:recipes/"
	if !strings.HasPrefix(name, prefix) || name == "minecraft:recipes/root" {
		return nil
	}
	path := strings.TrimPrefix(name, prefix)
	recipe := "minecraft:" + path[strings.LastIndexByte(path, '/')+1:]
	return &advancementRecipe{
		criterion: "has_the_recipe",
		watch:     recipe,
		rewards:   append([]string{recipe}, advancementRecipeExtras[name]...),
	}
}

// advancementRecipeExtras are the recipes the built-in recipe advancements
// hand out besides the one they are about.
//
// The "rewards" block of the advancement file names them, e.g. the one of
// `minecraft:recipes/building_blocks/gold_block` is the gold block recipe
// and `minecraft:gold_ingot_from_block`. A player who is given the gold
// block recipe therefore knows the gold ingot one as well, and the
// advancement that is about the gold ingot recipe is completed by the
// trigger without the command having a say in it.
var advancementRecipeExtras = map[string][]string{
	"minecraft:recipes/building_blocks/bone_block":    {"minecraft:bone_meal_from_block"},
	"minecraft:recipes/building_blocks/coal_block":    {"minecraft:coal"},
	"minecraft:recipes/building_blocks/diamond_block": {"minecraft:diamond"},
	"minecraft:recipes/building_blocks/emerald_block": {"minecraft:emerald"},
	"minecraft:recipes/building_blocks/gold_block":    {"minecraft:gold_ingot_from_block"},
	"minecraft:recipes/building_blocks/hay_block":     {"minecraft:wheat"},
	"minecraft:recipes/building_blocks/iron_block":    {"minecraft:iron_ingot_from_block"},
	"minecraft:recipes/building_blocks/lapis_block":   {"minecraft:lapis_lazuli"},
	"minecraft:recipes/decorations/slime":             {"minecraft:slime_ball"},
	"minecraft:recipes/misc/gold_ingot_from_nuggets":  {"minecraft:gold_nugget"},
	"minecraft:recipes/misc/iron_ingot_from_nuggets":  {"minecraft:iron_nugget"},
	"minecraft:recipes/redstone/redstone_block":       {"minecraft:redstone"},
}

// advancementRecipeID turns the name of a recipe into the id the recipe
// book holds.
//
// A name without a namespace means "minecraft", and both parts are
// lower-cased, like every other `MinecraftKey`.
func advancementRecipeID(name string) string {
	namespace, path := "minecraft", name
	if i := strings.IndexByte(name, ':'); i >= 0 {
		namespace, path = name[:i], name[i+1:]
	}
	return strings.ToLower(namespace) + ":" + strings.ToLower(path)
}

// knownRecipes returns the recipes the player knows.
//
// 1.12.2 keeps the recipe book of a player in its NBT, under "recipeBook",
// with the ids of the known recipes in a "recipes" list.
func knownRecipes(e *world.Entity) map[string]bool {
	out := map[string]bool{}
	if e == nil || e.NBT == nil {
		return out
	}
	book, ok := e.NBT.Get("recipeBook").(*nbt.Compound)
	if !ok {
		return out
	}
	list, ok := book.Get("recipes").(*nbt.List)
	if !ok {
		return out
	}
	for _, v := range list.Items() {
		if s, ok := v.(nbt.String); ok {
			out[string(s)] = true
		}
	}
	return out
}

// storeKnownRecipes stores the recipes the player knows.
//
// The recipes are written in the order of their ids, so the data of a
// player does not depend on the order the map happens to have.
func storeKnownRecipes(e *world.Entity, known map[string]bool) {
	if e == nil || e.NBT == nil {
		return
	}
	recipes := make([]string, 0, len(known))
	for recipe := range known {
		recipes = append(recipes, recipe)
	}
	sort.Strings(recipes)
	list := nbt.NewList(nbt.TypeString)
	for _, recipe := range recipes {
		list.Append(nbt.String(recipe))
	}
	book := nbt.NewCompound()
	book.Set("recipes", list)
	e.NBT.Set("recipeBook", book)
}
