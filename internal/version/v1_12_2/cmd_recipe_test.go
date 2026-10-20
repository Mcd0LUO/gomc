package v1_12_2

import (
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
