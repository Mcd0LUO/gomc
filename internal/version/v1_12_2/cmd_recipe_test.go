package v1_12_2

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// recipeEngine builds an engine with the 1.12.2 command set.
//
// The recipe command is registered on its own: `Register` does not add it
// yet, because the file that does is written by another hand.
func recipeEngine(t *testing.T) *exec.Engine {
	t.Helper()
	reg := exec.NewRegistry()
	Register(reg)
	registerRecipe(reg)
	return exec.NewEngine(reg, world.NewWorld())
}

// recipeWorld builds an engine whose world holds the given recipe files.
//
// The files are the recipes of the world directory: they are written to
// `<root>/recipes/<name>` and are read by the command the first time it
// needs them. A name with a "/" in it is a file in a subdirectory.
func recipeWorld(t *testing.T, recipes map[string]string) *exec.Engine {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "recipes")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	for name, body := range recipes {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}
	reg := exec.NewRegistry()
	Register(reg)
	registerRecipe(reg)
	e := exec.NewEngine(reg, world.NewWorld())
	e.SetRoot(root)
	return e
}

// recipePlayer puts a player with the given name into the world.
func recipePlayer(t *testing.T, e *exec.Engine, name string) *world.Entity {
	t.Helper()
	p, ok := e.World.EntityTypes.NewEntity("player", "p1")
	if !ok {
		t.Fatalf("the player entity type is missing")
	}
	p.PlayerName = name
	e.World.AddEntity(p)
	return p
}

// recipeRun runs a command line and returns the messages it printed.
func recipeRun(e *exec.Engine, line string) ([]string, error) {
	before := len(e.Feedback)
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	if _, err := e.ExecuteLine(line, 1, src); err != nil {
		return nil, err
	}
	return e.Feedback[before:], nil
}

// recipeError runs a command line that has to fail and returns the message.
func recipeError(t *testing.T, e *exec.Engine, line string) string {
	t.Helper()
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	_, err := e.ExecuteLine(line, 1, src)
	if err == nil {
		t.Fatalf("%q: expected an error", line)
	}
	return err.Error()
}

// TestRecipeGiveOne checks the wording of `/recipe give <player> <name>`.
//
// (Measured on a real Paper 1.12.2.)
func TestRecipeGiveOne(t *testing.T) {
	e := recipeEngine(t)
	recipePlayer(t, e, "Steve")

	msgs, err := recipeRun(e, "recipe give Steve stone_slab")
	if err != nil {
		t.Fatalf("recipe give: %v", err)
	}
	want := "Successfully given Steve the recipe for Stone Slab"
	if len(msgs) != 1 || msgs[0] != want {
		t.Errorf("the command said %q, want %q", msgs, want)
	}
}

// TestRecipeGiveAll checks that "*" hands out every recipe.
func TestRecipeGiveAll(t *testing.T) {
	e := recipeEngine(t)
	p := recipePlayer(t, e, "Steve")

	msgs, err := recipeRun(e, "recipe give Steve *")
	if err != nil {
		t.Fatalf("recipe give *: %v", err)
	}
	want := "Successfully given all recipes to Steve"
	if len(msgs) != 1 || msgs[0] != want {
		t.Errorf("the command said %q, want %q", msgs, want)
	}
	if got := len(playerRecipes(p)); got != len(recipeRegistry) {
		t.Errorf("the player knows %d recipes, want %d", got, len(recipeRegistry))
	}
}

// TestRecipeTakeOne checks the wording of `/recipe take <player> <name>`.
func TestRecipeTakeOne(t *testing.T) {
	e := recipeEngine(t)
	p := recipePlayer(t, e, "Steve")

	if _, err := recipeRun(e, "recipe give Steve stone_slab"); err != nil {
		t.Fatalf("recipe give: %v", err)
	}
	msgs, err := recipeRun(e, "recipe take Steve stone_slab")
	if err != nil {
		t.Fatalf("recipe take: %v", err)
	}
	want := "Successfully removed the recipe for Stone Slab from Steve"
	if len(msgs) != 1 || msgs[0] != want {
		t.Errorf("the command said %q, want %q", msgs, want)
	}
	if knowsRecipe(p, "minecraft:stone_slab") {
		t.Error("the player still knows the recipe")
	}
}

// TestRecipeTakeAll checks that "*" takes every recipe away.
func TestRecipeTakeAll(t *testing.T) {
	e := recipeEngine(t)
	p := recipePlayer(t, e, "Steve")

	if _, err := recipeRun(e, "recipe give Steve *"); err != nil {
		t.Fatalf("recipe give *: %v", err)
	}
	msgs, err := recipeRun(e, "recipe take Steve *")
	if err != nil {
		t.Fatalf("recipe take *: %v", err)
	}
	want := "Successfully taken all recipes from Steve"
	if len(msgs) != 1 || msgs[0] != want {
		t.Errorf("the command said %q, want %q", msgs, want)
	}
	if got := len(playerRecipes(p)); got != 0 {
		t.Errorf("the player still knows %d recipes", got)
	}
}

// TestRecipeAlreadyHave checks the wording of a recipe the player has.
func TestRecipeAlreadyHave(t *testing.T) {
	e := recipeEngine(t)
	recipePlayer(t, e, "Steve")

	if _, err := recipeRun(e, "recipe give Steve stone_slab"); err != nil {
		t.Fatalf("recipe give: %v", err)
	}
	want := "Player Steve already has a recipe for Stone Slab"
	if got := recipeError(t, e, "recipe give Steve stone_slab"); got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestRecipeDontHave checks the wording of a recipe the player does not have.
func TestRecipeDontHave(t *testing.T) {
	e := recipeEngine(t)
	recipePlayer(t, e, "Steve")

	want := "Player Steve doesn't have the recipe for Stone Slab"
	if got := recipeError(t, e, "recipe take Steve stone_slab"); got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestRecipeUnknownRecipe checks the wording of a recipe that does not exist.
func TestRecipeUnknownRecipe(t *testing.T) {
	e := recipeEngine(t)
	recipePlayer(t, e, "Steve")

	want := "no_such_recipe is an unknown recipe"
	if got := recipeError(t, e, "recipe give Steve no_such_recipe"); got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
	if got := recipeError(t, e, "recipe take Steve no_such_recipe"); got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestRecipeUnsupportedRecipe checks the wording of a recipe the server
// refuses to hand out.
//
// The eleven recipes the server builds in code cannot be made in a crafting
// table, so it answers "unsupported" instead of "unknown".
func TestRecipeUnsupportedRecipe(t *testing.T) {
	e := recipeEngine(t)
	recipePlayer(t, e, "Steve")

	want := "armordye is an unsupported recipe"
	if got := recipeError(t, e, "recipe give Steve armordye"); got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
	want = "minecraft:armordye is an unsupported recipe"
	if got := recipeError(t, e, "recipe give Steve minecraft:armordye"); got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestRecipeNamesAreFolded checks that the name of a recipe is read the way
// the server reads it: the namespace defaults to "minecraft" and both parts
// are lower-cased.
//
// (Measured on a real Paper 1.12.2: `recipe give <player> STONE_SLAB` gives
// the stone slab recipe.)
func TestRecipeNamesAreFolded(t *testing.T) {
	e := recipeEngine(t)
	recipePlayer(t, e, "Steve")

	for _, line := range []string{
		"recipe give Steve stone_slab",
		"recipe give Steve minecraft:stone_slab",
		"recipe give Steve MINECRAFT:STONE_SLAB",
		"recipe give Steve STONE_SLAB",
	} {
		if _, err := recipeRun(e, line); err != nil {
			t.Fatalf("%q: %v", line, err)
		}
		if _, err := recipeRun(e, "recipe take Steve stone_slab"); err != nil {
			t.Fatalf("%q: %v", line, err)
		}
	}
}

// TestRecipeGivesTheRecipesThatComeWithIt checks the side effect of
// unlocking a recipe.
//
// 1.12.2 grants the advancement that is about the recipe, and that
// advancement rewards other recipes as well: the gold block recipe brings
// the gold ingot one with it.
//
// (Measured on a real Paper 1.12.2.)
func TestRecipeGivesTheRecipesThatComeWithIt(t *testing.T) {
	e := recipeEngine(t)
	p := recipePlayer(t, e, "Steve")

	if _, err := recipeRun(e, "recipe give Steve minecraft:gold_block"); err != nil {
		t.Fatalf("recipe give: %v", err)
	}
	if !knowsRecipe(p, "minecraft:gold_ingot_from_block") {
		t.Error("the player does not know the gold ingot recipe")
	}
	want := "Player Steve already has a recipe for Gold Ingot"
	if got := recipeError(t, e, "recipe give Steve minecraft:gold_ingot_from_block"); got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestRecipeUsage checks the wording of a usage mistake.
//
// (Measured on a real Paper 1.12.2.)
func TestRecipeUsage(t *testing.T) {
	e := recipeEngine(t)
	recipePlayer(t, e, "Steve")

	want := "Usage: /recipe <give|take> [player] <name|*>"
	for _, line := range []string{
		"recipe",
		"recipe give",
		"recipe foo Steve stone_slab",
	} {
		if got := recipeError(t, e, line); got != want {
			t.Errorf("%q: the error was %q, want %q", line, got, want)
		}
	}
}

// TestRecipeWithoutTheRecipeArgument checks what the server answers when
// the recipe is left out.
//
// 1.12.2 reads the third argument without checking that there is one, so
// the command falls over and the server reports a plain failure.
//
// (Measured on a real Paper 1.12.2.)
func TestRecipeWithoutTheRecipeArgument(t *testing.T) {
	e := recipeEngine(t)
	recipePlayer(t, e, "Steve")

	want := "An unknown error occurred while attempting to perform this command"
	if got := recipeError(t, e, "recipe give Steve"); got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
	// The player is looked up before the recipe, so a player who is not
	// online is reported first.
	want = "Player 'Nobody' cannot be found"
	if got := recipeError(t, e, "recipe give Nobody"); got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestRecipeUnknownPlayer checks the wording of a player who is not online.
func TestRecipeUnknownPlayer(t *testing.T) {
	e := recipeEngine(t)
	recipePlayer(t, e, "Steve")

	want := "Player 'Nobody' cannot be found"
	if got := recipeError(t, e, "recipe give Nobody stone_slab"); got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestRecipeSelectorTarget checks that a selector picks the players.
//
// (Measured on a real Paper 1.12.2: `recipe give @a <name>` gives the recipe
// to every player the selector matches.)
func TestRecipeSelectorTarget(t *testing.T) {
	e := recipeEngine(t)
	recipePlayer(t, e, "Steve")

	msgs, err := recipeRun(e, "recipe give @a stone_slab")
	if err != nil {
		t.Fatalf("recipe give @a: %v", err)
	}
	want := "Successfully given Steve the recipe for Stone Slab"
	if len(msgs) != 1 || msgs[0] != want {
		t.Errorf("the command said %q, want %q", msgs, want)
	}
}

// customRecipeBody is the body of a recipe file that produces an item.
//
// The type of a recipe of 1.12.2 is the bare name: the server compares
// it with "crafting_shaped" and "crafting_shapeless" and knows nothing
// else.
//
// (Measured on a real Paper 1.12.2: a file that says
// "minecraft:crafting_shaped" is not a recipe of the server.)
func customRecipeBody(item string, count int) string {
	return `{"type": "crafting_shaped", "pattern": ["#"], "key": {"#": {"item": "minecraft:stick"}}, "result": {"item": "` +
		item + `", "count": ` + itoa(count) + `}}`
}

// customRecipeBodyWithData is the body of a recipe file that produces an
// item of a data value.
//
// The server refuses a result whose item has more than one form when
// the file does not say which form it is, so the data value is written
// for those items.
//
// (Measured on a real Paper 1.12.2: the server logs "Missing data for
// item 'minecraft:dye'" for a file whose result is a dye and that has no
// "data".)
func customRecipeBodyWithData(item string, data, count int) string {
	return `{"type": "crafting_shaped", "pattern": ["#"], "key": {"#": {"item": "minecraft:stick"}}, "result": {"item": "` +
		item + `", "data": ` + itoa(data) + `, "count": ` + itoa(count) + `}}`
}

// itoa writes a small number, so the tests do not need the fmt package.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var out []byte
	for n > 0 {
		out = append([]byte{byte('0' + n%10)}, out...)
		n /= 10
	}
	return string(out)
}

// TestRecipeCustomRecipeNames covers the name of a custom recipe.
//
// 1.12.2 names the recipe after the item it makes, and the item name comes
// from the item and its data value: `minecraft:beef` is "Raw Beef" (an
// item, not a block), and `minecraft:diamond_block` has no forms and is
// "Block of Diamond".
//
// (Measured on a real Paper 1.12.2: both names are the ones it answers.)
func TestRecipeCustomRecipeNames(t *testing.T) {
	for _, tc := range []struct {
		item string
		want string
	}{
		{"minecraft:beef", "Raw Beef"},
		{"minecraft:diamond_block", "Block of Diamond"},
	} {
		e := recipeWorld(t, map[string]string{
			"my_recipe.json": customRecipeBody(tc.item, 1),
		})
		recipePlayer(t, e, "Steve")

		msgs, err := recipeRun(e, "recipe give Steve my_recipe")
		if err != nil {
			t.Fatalf("%s: %v", tc.item, err)
		}
		want := "Successfully given Steve the recipe for " + tc.want
		if len(msgs) != 1 || msgs[0] != want {
			t.Errorf("%s: the command said %q, want %q", tc.item, msgs, want)
		}
	}
}

// TestRecipeCustomRecipeNamesAreFolded checks that the name of a recipe
// of the world directory is read the way the server reads a name: the
// namespace defaults to "minecraft" and both parts are lower-cased.
func TestRecipeCustomRecipeNamesAreFolded(t *testing.T) {
	e := recipeWorld(t, map[string]string{
		"my_recipe.json": customRecipeBody("minecraft:diamond_block", 1),
	})
	recipePlayer(t, e, "Steve")

	for _, line := range []string{
		"recipe give Steve my_recipe",
		"recipe give Steve minecraft:my_recipe",
		"recipe give Steve MY_RECIPE",
		"recipe give Steve MineCraft:My_Recipe",
	} {
		if _, err := recipeRun(e, line); err != nil {
			t.Fatalf("%q: %v", line, err)
		}
		if _, err := recipeRun(e, "recipe take Steve my_recipe"); err != nil {
			t.Fatalf("%q: %v", line, err)
		}
	}
}

// TestRecipeCustomRecipeIsTakenAway checks the wording of taking a
// recipe of the world directory away.
func TestRecipeCustomRecipeIsTakenAway(t *testing.T) {
	e := recipeWorld(t, map[string]string{
		"my_recipe.json": customRecipeBody("minecraft:diamond_block", 1),
	})
	recipePlayer(t, e, "Steve")

	if _, err := recipeRun(e, "recipe give Steve my_recipe"); err != nil {
		t.Fatalf("recipe give: %v", err)
	}
	msgs, err := recipeRun(e, "recipe take Steve my_recipe")
	if err != nil {
		t.Fatalf("recipe take: %v", err)
	}
	want := "Successfully removed the recipe for Block of Diamond from Steve"
	if len(msgs) != 1 || msgs[0] != want {
		t.Errorf("the command said %q, want %q", msgs, want)
	}
	if got := recipeError(t, e, "recipe take Steve my_recipe"); got !=
		"Player Steve doesn't have the recipe for Block of Diamond" {
		t.Errorf("the error was %q, want the wording of a recipe the player does not have", got)
	}
}

// TestRecipeCustomRecipeOfAnUnknownType checks that a file of a type
// 1.12.2 does not know is not a recipe.
//
// (Measured on a real Paper 1.12.2: the server logs "Invalid or
// unsupported recipe type 'minecraft:crafting_shaped'" and answers
// "<name> is an unknown recipe".)
func TestRecipeCustomRecipeOfAnUnknownType(t *testing.T) {
	e := recipeWorld(t, map[string]string{
		"namespaced.json": `{"type": "minecraft:crafting_shaped", "pattern": ["#"], "key": {"#": {"item": "minecraft:stick"}}, "result": {"item": "minecraft:emerald"}}`,
		"special.json":    `{"type": "crafting_special", "result": {"item": "minecraft:firework_rocket"}}`,
	})
	recipePlayer(t, e, "Steve")

	for _, name := range []string{"namespaced", "special"} {
		want := name + " is an unknown recipe"
		if got := recipeError(t, e, "recipe give Steve "+name); got != want {
			t.Errorf("the error was %q, want %q", got, want)
		}
	}
}

// TestRecipeCustomRecipeOfABrokenFile checks that a file the server
// cannot read is not a recipe, and that it takes the recipes that come
// after it with it.
//
// The server stops reading the recipes when a file is broken, so the
// recipes after the broken file are not read either.
//
// (Measured on a real Paper 1.12.2: the server logs "Parsing error
// loading recipe minecraft:bbb_broken" and "Errors with built-in
// recipes!", the recipe before the broken file is handed out, and the
// recipe after it answers "ccc_good is an unknown recipe".)
func TestRecipeCustomRecipeOfABrokenFile(t *testing.T) {
	e := recipeWorld(t, map[string]string{
		"aaa_good.json":   customRecipeBody("minecraft:diamond_block", 1),
		"bbb_broken.json": `{"type": "crafting_shaped", "result":`,
		"ccc_good.json":   customRecipeBody("minecraft:emerald_block", 1),
	})
	recipePlayer(t, e, "Steve")

	msgs, err := recipeRun(e, "recipe give Steve aaa_good")
	if err != nil {
		t.Fatalf("recipe give: %v", err)
	}
	want := "Successfully given Steve the recipe for Block of Diamond"
	if len(msgs) != 1 || msgs[0] != want {
		t.Errorf("the command said %q, want %q", msgs, want)
	}
	if got := recipeError(t, e, "recipe give Steve bbb_broken"); got !=
		"bbb_broken is an unknown recipe" {
		t.Errorf("the error was %q, want the wording of an unknown recipe", got)
	}
	// The broken file takes the recipes that come after it with it.
	if got := recipeError(t, e, "recipe give Steve ccc_good"); got !=
		"ccc_good is an unknown recipe" {
		t.Errorf("the error was %q, want the wording of an unknown recipe", got)
	}
}

// TestABadRecipeFileMakesEveryRecipeUnknown covers a file the loader cannot
// read.
//
// 1.12.2 loads the custom recipes as one set: a file it cannot build takes
// the whole set down, so a **good** recipe that comes after the bad one is
// unknown too. The three ways a file can be bad -- an unknown type, an
// unknown item, a data value without an item that has forms -- all end the
// same way.
func TestABadRecipeFileMakesEveryRecipeUnknown(t *testing.T) {
	for _, tc := range []struct {
		why  string
		body string
	}{
		{"an unknown type", `{"type": "minecraft:crafting_shaped", "pattern": ["#"], "key": {"#": {"item": "minecraft:stick"}}, "result": {"item": "minecraft:emerald"}}`},
		{"an unknown item", customRecipeBody("minecraft:not_a_real_item", 1)},
		{"a data value without forms", `{"type": "crafting_shaped", "pattern": ["#"], "key": {"#": {"item": "minecraft:stick"}}, "result": {"item": "minecraft:dye"}}`},
	} {
		e := recipeWorld(t, map[string]string{
			"aaa_bad.json":  tc.body,
			"bbb_good.json": customRecipeBody("minecraft:diamond_block", 1),
		})
		recipePlayer(t, e, "Steve")

		for _, name := range []string{"aaa_bad", "bbb_good"} {
			want := name + " is an unknown recipe"
			if got := recipeError(t, e, "recipe give Steve "+name); got != want {
				t.Errorf("%s: %s said %q, want %q", tc.why, name, got, want)
			}
		}
	}
}

// TestRecipeCustomRecipeInASubdirectory checks that a recipe file in a
// subdirectory is a recipe, and that the name of that recipe is the path
// of the file below the recipes directory.
//
// (Measured on a real Paper 1.12.2: the server walks the recipes
// directory with `Files.walk`, so a file in a subdirectory is a recipe,
// and `recipe give <player> sub/my_recipe` hands it out.)
func TestRecipeCustomRecipeInASubdirectory(t *testing.T) {
	e := recipeWorld(t, map[string]string{
		"sub/my_recipe.json": customRecipeBody("minecraft:diamond_block", 1),
	})
	recipePlayer(t, e, "Steve")

	for _, line := range []string{
		"recipe give Steve sub/my_recipe",
		"recipe give Steve minecraft:sub/my_recipe",
	} {
		msgs, err := recipeRun(e, line)
		if err != nil {
			t.Fatalf("%q: %v", line, err)
		}
		want := "Successfully given Steve the recipe for Block of Diamond"
		if len(msgs) != 1 || msgs[0] != want {
			t.Errorf("%q: the command said %q, want %q", line, msgs, want)
		}
		if _, err := recipeRun(e, "recipe take Steve sub/my_recipe"); err != nil {
			t.Fatalf("%q: %v", line, err)
		}
	}
}

// TestRecipeCustomRecipeInADeepSubdirectory checks that a recipe file
// in a subdirectory of a subdirectory is a recipe as well.
func TestRecipeCustomRecipeInADeepSubdirectory(t *testing.T) {
	e := recipeWorld(t, map[string]string{
		"sub/deep/my_recipe.json": customRecipeBody("minecraft:emerald_block", 1),
	})
	recipePlayer(t, e, "Steve")

	msgs, err := recipeRun(e, "recipe give Steve sub/deep/my_recipe")
	if err != nil {
		t.Fatalf("recipe give: %v", err)
	}
	want := "Successfully given Steve the recipe for Block of Emerald"
	if len(msgs) != 1 || msgs[0] != want {
		t.Errorf("the command said %q, want %q", msgs, want)
	}
}

// TestRecipeCustomRecipeOfADataValue checks that the name of a recipe
// is the name of the item it produces, with the data value of that
// item.
//
// (Measured on a real Paper 1.12.2: a recipe whose result is a dye of
// the data value 3 is answered with "Cocoa Beans", one whose result is
// a banner of the data value 14 with "Orange Banner", one whose result
// is a bed of the data value 4 with "Yellow Bed" and one whose result
// is a skull of the data value 2 with "Zombie Head".)
func TestRecipeCustomRecipeOfADataValue(t *testing.T) {
	for _, tc := range []struct {
		file string
		item string
		data int
		want string
	}{
		{"dye.json", "minecraft:dye", 0, "Ink Sac"},
		{"dye3.json", "minecraft:dye", 3, "Cocoa Beans"},
		{"dye11.json", "minecraft:dye", 11, "Dandelion Yellow"},
		{"banner.json", "minecraft:banner", 0, "Black Banner"},
		{"banner14.json", "minecraft:banner", 14, "Orange Banner"},
		{"bed.json", "minecraft:bed", 4, "Yellow Bed"},
		{"skull.json", "minecraft:skull", 2, "Zombie Head"},
		{"wool.json", "minecraft:wool", 5, "Lime Wool"},
	} {
		e := recipeWorld(t, map[string]string{
			tc.file: customRecipeBodyWithData(tc.item, tc.data, 1),
		})
		recipePlayer(t, e, "Steve")

		msgs, err := recipeRun(e, "recipe give Steve "+strings.TrimSuffix(tc.file, ".json"))
		if err != nil {
			t.Fatalf("%s: %v", tc.file, err)
		}
		want := "Successfully given Steve the recipe for " + tc.want
		if len(msgs) != 1 || msgs[0] != want {
			t.Errorf("%s: the command said %q, want %q", tc.file, msgs, want)
		}
	}
}

// TestRecipeCustomRecipeOfAnUnknownItem checks that a recipe whose
// result is an item 1.12.2 does not know is not a recipe.
//
// (Measured on a real Paper 1.12.2: the server logs "Unknown item
// 'minecraft:not_a_real_item'" and the recipe is unknown.)
func TestRecipeCustomRecipeOfAnUnknownItem(t *testing.T) {
	e := recipeWorld(t, map[string]string{
		"my_recipe.json": customRecipeBody("minecraft:not_a_real_item", 1),
	})
	recipePlayer(t, e, "Steve")

	if got := recipeError(t, e, "recipe give Steve my_recipe"); got !=
		"my_recipe is an unknown recipe" {
		t.Errorf("the error was %q, want the wording of an unknown recipe", got)
	}
}

// TestRecipeCustomRecipesAreGivenByStar checks that "*" hands out the
// recipes of the world directory as well.
//
// (Measured on a real Paper 1.12.2: a custom recipe is one of the
// recipes "*" hands out.)
func TestRecipeCustomRecipesAreGivenByStar(t *testing.T) {
	e := recipeWorld(t, map[string]string{
		"my_recipe.json": customRecipeBody("minecraft:diamond_block", 1),
	})
	recipePlayer(t, e, "Steve")

	if _, err := recipeRun(e, "recipe give Steve *"); err != nil {
		t.Fatalf("recipe give *: %v", err)
	}
	if got := recipeError(t, e, "recipe give Steve my_recipe"); got !=
		"Player Steve already has a recipe for Block of Diamond" {
		t.Errorf("the error was %q, want the wording of a recipe the player has", got)
	}
}

// TestRecipeCustomRecipeIsKeptInTheRecipeBook checks that a recipe of
// the world directory is written into the recipe book of the player.
func TestRecipeCustomRecipeIsKeptInTheRecipeBook(t *testing.T) {
	e := recipeWorld(t, map[string]string{
		"my_recipe.json": customRecipeBody("minecraft:diamond_block", 1),
	})
	p := recipePlayer(t, e, "Steve")

	if _, err := recipeRun(e, "recipe give Steve my_recipe"); err != nil {
		t.Fatalf("recipe give: %v", err)
	}
	if !knowsRecipe(p, "minecraft:my_recipe") {
		t.Error("the player does not know the recipe")
	}
}

// TestRecipeCustomRecipeUnlocksNothing checks that handing out a recipe
// of the world directory does not unlock anything else.
//
// The recipes of the server come with the advancement about them, and
// that advancement rewards other recipes. A custom recipe has no
// advancement, so it brings nothing with it.
//
// (Measured on a real Paper 1.12.2: giving a custom recipe leaves the
// recipes of the server alone.)
func TestRecipeCustomRecipeUnlocksNothing(t *testing.T) {
	e := recipeWorld(t, map[string]string{
		"my_recipe.json": customRecipeBody("minecraft:diamond_block", 1),
	})
	p := recipePlayer(t, e, "Steve")

	if _, err := recipeRun(e, "recipe give Steve my_recipe"); err != nil {
		t.Fatalf("recipe give: %v", err)
	}
	if knowsRecipe(p, "minecraft:diamond") {
		t.Error("the player knows the recipe of the diamond")
	}
}

// TestRecipeCustomRecipeCountIsRead checks that the number of items a
// recipe produces is read. The server keeps the count but never prints
// it, so the check looks at the recipe that was read.
func TestRecipeCustomRecipeCountIsRead(t *testing.T) {
	e := recipeWorld(t, map[string]string{
		"my_recipe.json": customRecipeBody("minecraft:diamond_block", 3),
	})
	recipePlayer(t, e, "Steve")

	recipes := customRecipes(e.Root)
	if len(recipes) != 1 {
		t.Fatalf("the world holds %d recipes, want 1", len(recipes))
	}
	if recipes[0].count != 3 {
		t.Errorf("the recipe produces %d items, want 3", recipes[0].count)
	}
	if recipes[0].result != "diamond_block" {
		t.Errorf("the recipe produces %q, want the diamond block", recipes[0].result)
	}
}

// TestRecipeCustomRecipesAreTheSameEveryTime checks that the recipes of
// a world do not change while the runtime runs.
func TestRecipeCustomRecipesAreTheSameEveryTime(t *testing.T) {
	e := recipeWorld(t, map[string]string{
		"my_recipe.json": customRecipeBody("minecraft:diamond_block", 1),
	})
	recipePlayer(t, e, "Steve")

	if _, err := recipeRun(e, "recipe give Steve my_recipe"); err != nil {
		t.Fatalf("recipe give: %v", err)
	}
	if _, err := recipeRun(e, "recipe take Steve my_recipe"); err != nil {
		t.Fatalf("recipe take: %v", err)
	}
	msgs, err := recipeRun(e, "recipe give Steve my_recipe")
	if err != nil {
		t.Fatalf("recipe give: %v", err)
	}
	want := "Successfully given Steve the recipe for Block of Diamond"
	if len(msgs) != 1 || msgs[0] != want {
		t.Errorf("the command said %q, want %q", msgs, want)
	}
}

// TestItemDisplayName checks the name 1.12.2 shows for an item stack.
//
// (Measured on a real Paper 1.12.2: `give <player> <item> [amount]
// [data]` answers "Given [<name>] * <amount> to <player>".)
func TestItemDisplayName(t *testing.T) {
	for _, tc := range []struct {
		item string
		data int
		want string
	}{
		// An item without a namespace is the item of the "minecraft"
		// namespace.
		{"stone", 0, "Stone"},
		{"minecraft:stone", 0, "Stone"},
		{"MINECRAFT:STONE", 0, "Stone"},
		{"minecraft:beef", 0, "Raw Beef"},
		{"minecraft:map", 0, "Empty Map"},
		{"minecraft:wooden_slab", 0, "Oak Wood Slab"},
		// The names of the items that have more than one form follow
		// the data value.
		{"minecraft:dye", 0, "Ink Sac"},
		{"minecraft:dye", 3, "Cocoa Beans"},
		{"minecraft:dye", 11, "Dandelion Yellow"},
		{"minecraft:dye", 15, "Bone Meal"},
		{"minecraft:banner", 0, "Black Banner"},
		{"minecraft:banner", 14, "Orange Banner"},
		{"minecraft:bed", 0, "White Bed"},
		{"minecraft:bed", 4, "Yellow Bed"},
		{"minecraft:skull", 0, "Skeleton Skull"},
		{"minecraft:skull", 2, "Zombie Head"},
		{"minecraft:wool", 5, "Lime Wool"},
		// The name of a potion says which effect it has, and the effect
		// of a plain potion is in the data value.
		{"minecraft:potion", 0, "Water Bottle"},
		{"minecraft:potion", 1, "Potion of Regeneration"},
		{"minecraft:potion", 5, "Potion of Healing"},
		{"minecraft:potion", 14, "Potion of Invisibility"},
		// A data value outside 0..15 is read as the effect of the data
		// value and 15, and a few of them are a variant of the potion.
		{"minecraft:potion", 16, "Awkward Potion"},
		{"minecraft:potion", 17, "Potion of Regeneration"},
		{"minecraft:potion", 32, "Thick Potion"},
		{"minecraft:potion", 48, "Water Bottle"},
		{"minecraft:potion", 64, "Mundane Potion"},
		{"minecraft:potion", 100, "Potion of Poison"},
		{"minecraft:splash_potion", 0, "Splash Water Bottle"},
		{"minecraft:splash_potion", 5, "Splash Water Bottle"},
		{"minecraft:lingering_potion", 0, "Lingering Water Bottle"},
		{"minecraft:tipped_arrow", 0, "Arrow of Splashing"},
		// A data value of -1 is "not given", and the server uses the
		// data value 0 then.
		{"minecraft:dye", -1, "Ink Sac"},
		{"minecraft:wool", -1, "White Wool"},
		// A data value outside 0..15 gives the name of the data value 0.
		{"minecraft:dye", 16, "Ink Sac"},
		{"minecraft:dye", 100, "Ink Sac"},
		{"minecraft:wool", 16, "White Wool"},
		{"minecraft:bed", 16, "White Bed"},
		{"minecraft:banner", 16, "Black Banner"},
		{"minecraft:skull", 16, "Skeleton Skull"},
	} {
		got, known := itemDisplayName(tc.item, tc.data)
		if !known {
			t.Errorf("itemDisplayName(%q, %d): the item is not known", tc.item, tc.data)
			continue
		}
		if got != tc.want {
			t.Errorf("itemDisplayName(%q, %d) = %q, want %q", tc.item, tc.data, got, tc.want)
		}
	}
}

// TestItemDisplayNameOfAnUnknownItem checks that an item the server does
// not know has no name.
func TestItemDisplayNameOfAnUnknownItem(t *testing.T) {
	for _, item := range []string{
		"",
		"not_a_real_item",
		"minecraft:not_a_real_item",
		// An item of another namespace is an item the server does not
		// know: 1.12.2 knows the items of the "minecraft" namespace
		// only.
		"other:dye",
		"minecraft2:dye",
	} {
		got, known := itemDisplayName(item, 0)
		if known {
			t.Errorf("itemDisplayName(%q, 0) = %q, want the item to be unknown", item, got)
		}
		if got != "" {
			t.Errorf("itemDisplayName(%q, 0) = %q, want an empty name", item, got)
		}
	}
}

// TestRecipeItemName checks the name the server calls a recipe that
// produces an item.
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
func TestRecipeItemName(t *testing.T) {
	for _, tc := range []struct {
		item string
		data int
		want string
	}{
		{"minecraft:diamond_block", 0, "Block of Diamond"},
		{"minecraft:dye", 3, "Cocoa Beans"},
		{"minecraft:wool", 5, "Lime Wool"},
		{"minecraft:potion", 0, "Uncraftable Potion"},
		{"minecraft:potion", 5, "Uncraftable Potion"},
		{"minecraft:splash_potion", 0, "Splash Uncraftable Potion"},
		{"minecraft:lingering_potion", 0, "Lingering Uncraftable Potion"},
		{"minecraft:tipped_arrow", 0, "Uncraftable Tipped Arrow"},
	} {
		got, known := recipeItemName(tc.item, tc.data)
		if !known {
			t.Errorf("recipeItemName(%q, %d): the item is not known", tc.item, tc.data)
			continue
		}
		if got != tc.want {
			t.Errorf("recipeItemName(%q, %d) = %q, want %q", tc.item, tc.data, got, tc.want)
		}
	}
}

// TestItemFormNamesMatchTheItemNames checks that the table of the forms
// and the table of the items tell the same name for the data value 0.
//
// The two tables hold the same information in two places, so they have
// to agree.
func TestItemFormNamesMatchTheItemNames(t *testing.T) {
	for item, names := range itemFormNames {
		if len(names) != 16 {
			t.Errorf("the item %q has %d names, want 16", item, len(names))
			continue
		}
		want, known := customItemNames[item]
		if !known {
			t.Errorf("the item %q is not in the table of the items", item)
			continue
		}
		if names[0] != want {
			t.Errorf("the item %q is called %q by the table of the forms and %q by the table of the items",
				item, names[0], want)
		}
	}
	for item, names := range itemEffectNames {
		if len(names) != 16 {
			t.Errorf("the item %q has %d names, want 16", item, len(names))
			continue
		}
		want, known := customItemNames[item]
		if !known {
			t.Errorf("the item %q is not in the table of the items", item)
			continue
		}
		if names[0] != want {
			t.Errorf("the item %q is called %q by the table of the effects and %q by the table of the items",
				item, names[0], want)
		}
	}
}
