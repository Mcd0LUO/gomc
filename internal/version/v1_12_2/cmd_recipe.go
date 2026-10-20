package v1_12_2

import (
	"fmt"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// recipeCmd implements `/recipe`.
//
// The 1.12.2 syntax is `/recipe <give|take> <player> <name|*>`. The usage
// line calls the player optional, but the server always reads the second
// argument as the player -- only the third one is the recipe -- so a
// `/recipe give <player>` without a recipe runs off the end of the argument
// list and the server answers "An unknown error occurred while attempting
// to perform this command".
//
// (Measured on a real Paper 1.12.2: `/recipe give <player>` answers exactly
// that, and the server log shows an ArrayIndexOutOfBoundsException in
// `CommandRecipe.execute`.)
type recipeCmd struct{}

// Name implements exec.Handler.
func (c *recipeCmd) Name() string { return "recipe" }

// Usage implements exec.Handler.
func (c *recipeCmd) Usage() string { return recipeUsageLine() }

// Parse implements exec.Handler.
//
// The arguments are read one by one and kept as they came, because the
// server counts them: `/recipe give <player>` has two arguments and the
// server runs into the end of the list when it looks for the recipe. The
// list therefore holds only the arguments that were really written.
func (c *recipeCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	var args []ast.Arg
	if p.More() {
		mode, err := p.String()
		if err != nil {
			return nil, err
		}
		args = append(args, &ast.StringArg{Value: mode})
	}
	if p.More() {
		target, err := parseEntityTarget(p)
		if err != nil {
			return nil, err
		}
		args = append(args, target)
	}
	if p.More() {
		name, err := p.String()
		if err != nil {
			return nil, err
		}
		args = append(args, &ast.StringArg{Value: name})
	}
	return args, nil
}

// Run implements exec.Handler.
func (c *recipeCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	// 1.12.2 refuses the command when it has fewer than two arguments.
	if len(args) < 2 {
		return 0, recipeUsageError()
	}
	mode := args[0].(*ast.StringArg).Value
	give := strings.EqualFold(mode, "give")
	take := strings.EqualFold(mode, "take")
	if !give && !take {
		return 0, recipeUsageError()
	}

	players, err := resolveRecipePlayers(ctx, args[1])
	if err != nil {
		return 0, err
	}

	// The server reads the third argument without checking that there is
	// one, so `/recipe give <player>` falls over with a plain failure.
	if len(args) < 3 {
		return 0, exec.Errorf("An unknown error occurred while attempting to perform this command")
	}
	raw := args[2].(*ast.StringArg).Value

	// "*" is every recipe the server can hand out.
	if raw == "*" {
		for _, p := range players {
			if give {
				learnAllRecipes(p)
				ctx.Engine.Feedback = append(ctx.Engine.Feedback,
					fmt.Sprintf("Successfully given all recipes to %s", p.DisplayName()))
			} else {
				forgetAllRecipes(p)
				ctx.Engine.Feedback = append(ctx.Engine.Feedback,
					fmt.Sprintf("Successfully taken all recipes from %s", p.DisplayName()))
			}
		}
		return len(players), nil
	}

	key := recipeKey(raw)
	name, known := recipeNames[key]
	if !known {
		// The eleven recipes the server builds in code cannot be made in
		// a crafting table, so it refuses to hand them out.
		if specialRecipes[key] {
			return 0, exec.Errorf("%s is an unsupported recipe", raw)
		}
		return 0, exec.Errorf("%s is an unknown recipe", raw)
	}

	for _, p := range players {
		// The server refuses a recipe the player already has, and one it
		// does not have when the recipe is to be taken away.
		if give == knowsRecipe(p, key) {
			if give {
				return 0, exec.Errorf("Player %s already has a recipe for %s",
					p.DisplayName(), name)
			}
			return 0, exec.Errorf("Player %s doesn't have the recipe for %s",
				p.DisplayName(), name)
		}
		if give {
			learnRecipe(p, key)
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Successfully given %s the recipe for %s", p.DisplayName(), name))
			continue
		}
		forgetRecipe(p, key)
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Successfully removed the recipe for %s from %s", name, p.DisplayName()))
	}
	return len(players), nil
}

// resolveRecipePlayers resolves the player argument of `/recipe`.
//
// 1.12.2 reads it with `CommandAbstract.a(server, listener, name)`: a
// selector is resolved first, and a name that no player has is reported as
// "Player '<name>' cannot be found". Only players count, because the recipe
// book belongs to a player.
func resolveRecipePlayers(ctx *exec.Context, arg ast.Arg) ([]*world.Entity, error) {
	switch t := arg.(type) {
	case *ast.StringArg:
		p := ctx.World.FindPlayer(t.Value)
		if p == nil {
			return nil, exec.Errorf("Player '%s' cannot be found", t.Value)
		}
		return []*world.Entity{p}, nil
	case *ast.Selector:
		resolver := &exec.SelectorResolver{World: ctx.World}
		entities, err := resolver.Resolve(ctx.Source, t)
		selectorWarnings(ctx, resolver)
		if err != nil {
			return nil, exec.Errorf("%v", err)
		}
		var players []*world.Entity
		for _, e := range entities {
			if e.IsPlayer {
				players = append(players, e)
			}
		}
		if len(players) == 0 {
			// The server falls back to looking the text up as a player
			// name when the selector matched nobody.
			return nil, exec.Errorf("Player '%s' cannot be found", selectorText(t))
		}
		return players, nil
	}
	return nil, exec.Errorf("invalid target")
}

// recipeKey turns the name of a recipe into the id the server looks it up
// by.
//
// 1.12.2 reads the name as a `MinecraftKey`: the namespace defaults to
// "minecraft" and both parts are lower-cased, so "STONE_SLAB" is the same
// recipe as "minecraft:stone_slab".
func recipeKey(name string) string {
	ns, path := "minecraft", name
	if i := strings.Index(name, ":"); i >= 0 {
		ns, path = name[:i], name[i+1:]
	}
	return strings.ToLower(ns) + ":" + strings.ToLower(path)
}

// recipeNames maps a recipe id to the name the server calls the recipe.
var recipeNames = buildRecipeNames()

// specialRecipes are the recipes the server refuses to hand out.
var specialRecipes = buildSpecialRecipes()

func buildRecipeNames() map[string]string {
	out := make(map[string]string, len(recipeRegistry))
	for _, r := range recipeRegistry {
		out[r.key] = r.name
	}
	return out
}

func buildSpecialRecipes() map[string]bool {
	out := make(map[string]bool, len(specialRecipeKeys))
	for _, k := range specialRecipeKeys {
		out[k] = true
	}
	return out
}

// knowsRecipe reports whether the player knows the recipe.
func knowsRecipe(e *world.Entity, key string) bool {
	return playerRecipes(e)[key]
}

// playerRecipes returns the recipes the player knows.
//
// 1.12.2 keeps the recipe book of a player in its NBT, under "recipeBook",
// with the ids of the known recipes in a "recipes" list.
func playerRecipes(e *world.Entity) map[string]bool {
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

// setPlayerRecipes stores the recipes the player knows.
func setPlayerRecipes(e *world.Entity, known map[string]bool) {
	if e == nil || e.NBT == nil {
		return
	}
	list := nbt.NewList(nbt.TypeString)
	for _, r := range recipeRegistry {
		if known[r.key] {
			list.Append(nbt.String(r.key))
		}
	}
	book := nbt.NewCompound()
	book.Set("recipes", list)
	e.NBT.Set("recipeBook", book)
}

// learnRecipe gives the player the recipe.
//
// Unlocking a recipe grants the advancement that is about it, and that
// advancement rewards other recipes as well, so the player learns those
// too.
func learnRecipe(e *world.Entity, key string) {
	known := playerRecipes(e)
	seen := map[string]bool{}
	pending := []string{key}
	for len(pending) > 0 {
		cur := pending[0]
		pending = pending[1:]
		if seen[cur] {
			continue
		}
		seen[cur] = true
		known[cur] = true
		pending = append(pending, recipeUnlocks[cur]...)
	}
	setPlayerRecipes(e, known)
}

// forgetRecipe takes the recipe away from the player.
//
// 1.12.2 removes exactly the recipe that was named: the recipes that came
// with it stay.
func forgetRecipe(e *world.Entity, key string) {
	known := playerRecipes(e)
	delete(known, key)
	setPlayerRecipes(e, known)
}

// learnAllRecipes gives the player every recipe the server can hand out.
func learnAllRecipes(e *world.Entity) {
	known := playerRecipes(e)
	for _, r := range recipeRegistry {
		known[r.key] = true
	}
	setPlayerRecipes(e, known)
}

// forgetAllRecipes takes every recipe away from the player.
func forgetAllRecipes(e *world.Entity) {
	setPlayerRecipes(e, map[string]bool{})
}

// recipeUsageLine is the usage line 1.12.2 prints for `/recipe`.
func recipeUsageLine() string {
	if usage := commandUsage("recipe"); usage != "" {
		return usage
	}
	return "/recipe <give|take> [player] <name|*>"
}

// recipeUsageMessage is the message 1.12.2 answers a usage mistake with.
//
// The server sends `commands.generic.usage` with the usage line of the
// command, so the message is "Usage: <the line>".
func recipeUsageMessage() string {
	prefix := commandUsage("generic")
	if prefix == "" {
		prefix = "Usage: %s"
	}
	return fmt.Sprintf(prefix, recipeUsageLine())
}

// recipeUsageError is the error 1.12.2 answers a usage mistake with.
func recipeUsageError() error {
	return exec.Errorf("%s", recipeUsageMessage())
}

// registerRecipe adds the recipe command.
//
// The permission level of the command is in the table of
// `commandPermissions`: 1.12.2 asks the command which level it needs
// (`CommandRecipe.a`), and it answers 2.
func registerRecipe(reg *exec.Registry) {
	reg.Register(&recipeCmd{})
}
