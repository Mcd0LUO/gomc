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
	// data is the data value of the item the recipe produces. The name
	// of the recipe depends on it: a recipe that produces a dye of the
	// data value 3 is the recipe for "Cocoa Beans".
	data int
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
		// Data is the data value of the item the recipe produces. It is
		// nil when the file does not say, which the server refuses for
		// an item that has more than one form.
		Data *int `json:"data"`
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
// "my_recipe.json" is the recipe "minecraft:my_recipe". A file in a
// subdirectory is a recipe as well, and the name of that recipe is the
// path of the file below the recipes directory: the file
// "sub/my_recipe.json" is the recipe "minecraft:sub/my_recipe".
//
// (Measured on a real Paper 1.12.2: the server walks the recipes
// directory with `Files.walk`, so a file in a subdirectory is a recipe,
// and `recipe give <player> sub/my_recipe` hands it out.)
//
// A file the server would not read is not a recipe, and it takes the
// recipes that come after it with it: the server stops reading the
// recipes when a file is broken, so the recipes after the broken file
// are not read either.
//
// (Measured on a real Paper 1.12.2: the server logs "Parsing error
// loading recipe minecraft:<name>" and "Errors with built-in recipes!",
// and a recipe that comes after the broken file answers "<name> is an
// unknown recipe".)
//
// The files are read in the order of their names, so that the same world
// always gives the same recipes. The server reads them in the order of
// the file system, which is not the same from one machine to the next.
func loadCustomRecipes(root string) []customRecipe {
	if root == "" {
		return nil
	}
	dir := filepath.Join(root, "recipes")
	var out []customRecipe
	for _, rel := range recipeFiles(dir, "") {
		recipe, ok := readCustomRecipe(filepath.Join(dir, filepath.FromSlash(rel)), rel)
		if !ok {
			// The server stops reading the recipes here, so the
			// recipes that come after the broken file are not read
			// either.
			break
		}
		out = append(out, recipe)
	}
	return out
}

// recipeFiles returns the recipe files of a directory, as the paths below
// that directory.
//
// The files are the ones whose name ends in ".json": the server walks
// the recipes directory and reads the files with that ending, in the
// directories below it as well. A directory whose name ends in ".json"
// is a file for the server as well -- it tries to read it and fails,
// which stops the reading of the recipes.
//
// The paths are separated by "/", the way the server writes them, and
// the order is the order of the names, so that the same world always
// gives the same recipes.
func recipeFiles(dir, prefix string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() && !strings.HasSuffix(name, ".json") {
			out = append(out, recipeFiles(filepath.Join(dir, name), prefix+name+"/")...)
			continue
		}
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		out = append(out, prefix+name)
	}
	return out
}

// readCustomRecipe reads one recipe file of the world directory.
//
// The second value is false when the server would not read the file: a
// file that is not JSON, a file of a type the server does not know, a
// file whose result the server cannot read.
//
// (Measured on a real Paper 1.12.2: the server logs "Invalid or
// unsupported recipe type 'minecraft:crafting_shaped'" for a file of an
// unknown type, "Unknown item 'minecraft:not_a_real_item'" for a file
// whose result is an item it does not know, and "Missing data for item
// 'minecraft:dye'" for a file whose result is an item that has more than
// one form and that does not say which form it is.)
//
// The path is the path of the file below the recipes directory, and it
// is the name of the recipe.
func readCustomRecipe(path, rel string) (customRecipe, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return customRecipe{}, false
	}
	var file customRecipeFile
	if err := json.Unmarshal(data, &file); err != nil {
		return customRecipe{}, false
	}
	if !customRecipeTypes[file.Type] {
		return customRecipe{}, false
	}
	result, ok := minecraftItemKey(file.Result.Item)
	if !ok {
		return customRecipe{}, false
	}
	if _, known := customItemNames[result]; !known {
		// The result is an item 1.12.2 does not know, so the server
		// cannot read the file either.
		return customRecipe{}, false
	}
	dataValue := 0
	if file.Result.Data != nil {
		dataValue = *file.Result.Data
	} else if itemHasForms(result) {
		// The result is an item with more than one form and the file
		// does not say which form it is.
		return customRecipe{}, false
	}
	name, _ := recipeItemName(result, dataValue)
	count := file.Result.Count
	if count < 1 {
		count = 1
	}
	return customRecipe{
		key:    recipeKey(strings.TrimSuffix(rel, ".json")),
		name:   name,
		result: result,
		data:   dataValue,
		count:  count,
	}, true
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
