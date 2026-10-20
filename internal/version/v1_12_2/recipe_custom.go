package v1_12_2

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// The custom recipes of a world.
//
// 1.12.2 keeps its recipes in JSON files and reads them from its own
// assets root: `CraftingManager.b` looks up "/assets/.mcassetsroot" and
// walks "/assets/minecraft/recipes" below it. A world folder holds no
// recipes, so a real server never looks at `<world>/recipes`.
//
// (Measured on a real Paper 1.12.2: a recipe file in the world folder is
// not a recipe of the server -- `recipe give <player> <name>` answers
// "<name> is an unknown recipe". The custom recipes of the measurement
// were put into the assets root of the server instead, which is where
// the server reads them.)
//
// gomc has no assets root, so the world directory is where a custom
// recipe lives: `<world>/recipes/<name>.json`. The rules below are the
// ones the server follows for the recipes it does read, so that the text
// of `/recipe` is the same for a custom recipe.
//
// A custom recipe has no advancement of its own, so handing it out does
// not unlock anything else: the chain of unlocks of 1.12.2 starts at the
// advancement about a recipe, and a custom recipe has none.
//
// (Measured on a real Paper 1.12.2: a custom recipe hands out only
// itself.)
type customRecipe struct {
	// key is the id of the recipe, e.g. "minecraft:my_recipe".
	key string
	// name is what the server calls the recipe: the display name of the
	// item the recipe produces.
	name string
	// result is the item the recipe produces, e.g. "minecraft:diamond_block".
	result string
	// count is how many items the recipe produces. The server keeps the
	// count but never prints it.
	count int
}

// customRecipeFile is the part of a recipe file the server reads.
//
// The recipes of 1.12.2 are JSON, not SNBT. Only the type of the recipe
// and its result matter here: `/recipe` hands out a recipe, it never
// crafts anything.
type customRecipeFile struct {
	// Type is the kind of the recipe.
	Type string `json:"type"`
	// Result is what the recipe produces.
	Result struct {
		// Item is the item the recipe produces.
		Item string `json:"item"`
		// Count is how many of them it produces. It defaults to one.
		Count int `json:"count"`
	} `json:"result"`
}

// customRecipeTypes are the recipe types 1.12.2 reads.
//
// The server compares the type of the file with these two names, so a
// file that says "minecraft:crafting_shaped" -- the way 1.13 writes it --
// is not a recipe of a 1.12.2 server, and neither is a file that says
// "crafting_special", the type 1.13 gave to the recipes the server
// builds in code.
//
// (Measured on a real Paper 1.12.2: the server logs "Invalid or
// unsupported recipe type 'minecraft:crafting_shaped'" for such a file
// and answers "<name> is an unknown recipe".)
var customRecipeTypes = map[string]bool{
	"crafting_shaped":    true,
	"crafting_shapeless": true,
}

// loadCustomRecipes reads the custom recipes of a world directory.
//
// The name of a recipe is the name of its file, in the "minecraft"
// namespace and lower-cased, the way 1.12.2 reads a name: the file
// "my_recipe.json" is the recipe "minecraft:my_recipe".
//
// A file the server would not read is not a recipe: a file that is not
// JSON, a file of another type and a file whose result is an item
// 1.12.2 does not know are all left out. The real server stops loading
// the recipes that come after such a file; gomc keeps them, because the
// recipes of the world directory are not the recipes of the server and
// one broken file should not take the others with it.
//
// (Measured on a real Paper 1.12.2: the server answers "<name> is an
// unknown recipe" for a file of an unknown type and for a file whose
// result is an item it does not know.)
func loadCustomRecipes(root string) []customRecipe {
	if root == "" {
		return nil
	}
	dir := filepath.Join(root, "recipes")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []customRecipe
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		var file customRecipeFile
		if err := json.Unmarshal(data, &file); err != nil {
			continue
		}
		if !customRecipeTypes[file.Type] {
			continue
		}
		result := customResultItem(file.Result.Item)
		if result == "" {
			continue
		}
		name, known := customItemNames[result]
		if !known {
			// The result is an item 1.12.2 does not know, so the server
			// cannot read the file either.
			continue
		}
		count := file.Result.Count
		if count < 1 {
			count = 1
		}
		out = append(out, customRecipe{
			key:    recipeKey(strings.TrimSuffix(entry.Name(), ".json")),
			name:   name,
			result: result,
			count:  count,
		})
	}
	return out
}

// customResultItem turns the result of a recipe file into the id of the
// item the recipe produces.
//
// 1.12.2 knows the items of the "minecraft" namespace only, so a result
// of another namespace is an item the server does not know.
func customResultItem(item string) string {
	namespace, name := "minecraft", item
	if i := strings.Index(item, ":"); i >= 0 {
		namespace, name = item[:i], item[i+1:]
	}
	if namespace != "minecraft" {
		return ""
	}
	return strings.ToLower(name)
}

// customRecipeCache holds the custom recipes of the world directories
// that have been used already.
//
// The real server reads its recipes once, when it starts. gomc has no
// start-up hook for the world directory, so the recipes are read the
// first time a command needs them and are kept for the rest of the run.
// What is kept only depends on the files of the world, so the same world
// always gives the same recipes.
var (
	customRecipeMu    sync.Mutex
	customRecipeCache = map[string][]customRecipe{}
)

// customRecipes returns the custom recipes of a world directory.
func customRecipes(root string) []customRecipe {
	customRecipeMu.Lock()
	defer customRecipeMu.Unlock()
	if recipes, ok := customRecipeCache[root]; ok {
		return recipes
	}
	recipes := loadCustomRecipes(root)
	customRecipeCache[root] = recipes
	return recipes
}

// customRecipeByKey finds the custom recipe of an id.
func customRecipeByKey(root, key string) (customRecipe, bool) {
	for _, recipe := range customRecipes(root) {
		if recipe.key == key {
			return recipe, true
		}
	}
	return customRecipe{}, false
}
