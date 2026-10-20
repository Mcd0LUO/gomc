package v1_12_2

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/Mcd0LUO/gomc/internal/world"
)

// The custom advancements of a world.
//
// 1.12.2 keeps them in `<world>/advancements/<namespace>/<name>.json` --
// the layout of 1.12.2, not the `data/<namespace>/advancements/` of 1.13
// and later. The server reads them when it starts (`AdvancementDataWorld.d`)
// and the files are JSON, read with Gson.
//
// The namespace is the first segment below the directory and the name is
// the rest of the path, so `advancements/my_pack/thing.json` is the
// advancement `my_pack:thing`. Both parts are lower-cased, like every
// other `MinecraftKey`.
//
// (Read from the bytecode of `AdvancementDataWorld.d` and
// `MinecraftKey` of the 1.12.2 server jar.)
const customAdvancementDir = "advancements"

// customAdvancementFile is the shape of an advancement file of a datapack.
//
// The "criteria" object holds the conditions, the "requirements" array
// holds the groups of conditions, and the "rewards" object holds what the
// advancement hands out when it is completed.
type customAdvancementFile struct {
	// Parent is the parent advancement, e.g. "minecraft:story/root".
	Parent string `json:"parent"`
	// Criteria holds the conditions of the advancement, by name.
	Criteria map[string]customCriterionFile `json:"criteria"`
	// Requirements holds the groups of conditions. Every group has to be
	// satisfied, and a group is satisfied as soon as one of its criteria
	// is done.
	Requirements [][]string `json:"requirements"`
	// Rewards holds what the advancement hands out.
	Rewards struct {
		// Recipes are the recipes the advancement unlocks.
		Recipes []string `json:"recipes"`
	} `json:"rewards"`
}

// customCriterionFile is one condition of an advancement file.
type customCriterionFile struct {
	// Trigger is the name of the trigger, e.g. "minecraft:impossible".
	Trigger string `json:"trigger"`
	// Conditions holds the settings of the trigger.
	Conditions struct {
		// Recipe is the recipe a "minecraft:recipe_unlocked" condition
		// watches.
		Recipe string `json:"recipe"`
	} `json:"conditions"`
}

// loadCustomAdvancements reads the advancements a datapack adds to a world.
//
// The files live in `<root>/advancements`. A file that does not parse, or
// that does not describe an advancement the server would accept, is
// reported and skipped: the server logs those and keeps the other
// advancements.
//
// The advancements come back sorted by name, so the same world always
// answers the same way.
func loadCustomAdvancements(root string) ([]*advancement, []error) {
	if root == "" {
		return nil, nil
	}
	base := filepath.Join(root, customAdvancementDir)
	if info, err := os.Stat(base); err != nil || !info.IsDir() {
		// A world without a datapack has no custom advancements, and the
		// server does not say anything about it.
		return nil, nil
	}
	var out []*advancement
	var errs []error
	err := filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		rel, relErr := filepath.Rel(base, path)
		if relErr != nil {
			return nil
		}
		name, ok := customAdvancementName(filepath.ToSlash(rel))
		if !ok {
			return nil
		}
		a, loadErr := readCustomAdvancement(path, name)
		if loadErr != nil {
			errs = append(errs, loadErr)
			return nil
		}
		out = append(out, a)
		return nil
	})
	if err != nil {
		errs = append(errs, err)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out, errs
}

// customAdvancementName turns the path of an advancement file into the
// name of the advancement.
//
// The path is split at the first slash: the part before it is the
// namespace and the part after it is the name, so "my_pack/thing.json" is
// "my_pack:thing". A file right in the directory has no namespace of its
// own and is the "minecraft" one.
//
// 1.12.2 builds a `MinecraftKey` out of the two parts, and a `MinecraftKey`
// lower-cases both of them.
//
// (The server only reads the files that have a namespace of their own --
// it splits the path and skips a file that has no slash in it. gomc reads
// those files as well, as the "minecraft" ones.)
func customAdvancementName(rel string) (string, bool) {
	rel = strings.TrimSuffix(rel, ".json")
	namespace, name, found := strings.Cut(rel, "/")
	if !found {
		namespace, name = "minecraft", rel
	}
	if namespace == "" || name == "" {
		return "", false
	}
	return strings.ToLower(namespace) + ":" + strings.ToLower(name), true
}

// readCustomAdvancement reads one advancement file.
//
// The rules are the ones the server's own reader applies: the criteria
// must not be empty, every requirement has to name a criterion the
// advancement has, and every criterion has to be part of a requirement.
// A file that breaks a rule does not become an advancement at all, so it
// is reported and skipped.
func readCustomAdvancement(path, name string) (*advancement, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file customAdvancementFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parsing error loading custom advancement %s from %s: %w", name, path, err)
	}
	if len(file.Criteria) == 0 {
		return nil, fmt.Errorf("custom advancement %s from %s has no criteria", name, path)
	}
	criteria := make([]string, 0, len(file.Criteria))
	for criterion := range file.Criteria {
		criteria = append(criteria, criterion)
	}
	// The server keeps the criteria in a map, so their order is not
	// fixed. Sorting them keeps gomc deterministic.
	sort.Strings(criteria)

	groups, err := customRequirementGroups(name, file.Requirements, criteria)
	if err != nil {
		return nil, err
	}
	a := &advancement{
		name:     name,
		criteria: criteria,
		groups:   groups,
	}
	if file.Parent != "" {
		a.parent = advancementKey(file.Parent)
	}
	// The recipe the advancement watches: the condition that is triggered
	// by unlocking a recipe.
	for _, criterion := range criteria {
		cond := file.Criteria[criterion]
		if cond.Trigger != "minecraft:recipe_unlocked" || cond.Conditions.Recipe == "" {
			continue
		}
		if a.recipe == nil {
			a.recipe = &advancementRecipe{}
		}
		a.recipe.criterion = criterion
		a.recipe.watch = advancementRecipeID(cond.Conditions.Recipe)
	}
	if len(file.Rewards.Recipes) > 0 {
		if a.recipe == nil {
			a.recipe = &advancementRecipe{}
		}
		a.recipe.rewards = normalizeRecipes(file.Rewards.Recipes)
	}
	return a, nil
}

// customRequirementGroups builds the requirement groups of an advancement.
//
// A file without a "requirements" array needs every criterion of its own:
// the server makes one group per criterion, and every group has to be
// satisfied.
func customRequirementGroups(name string, requirements [][]string, criteria []string) ([][]string, error) {
	if len(requirements) == 0 {
		groups := make([][]string, 0, len(criteria))
		for _, criterion := range criteria {
			groups = append(groups, []string{criterion})
		}
		return groups, nil
	}
	known := map[string]bool{}
	for _, criterion := range criteria {
		known[criterion] = true
	}
	groups := make([][]string, 0, len(requirements))
	for _, group := range requirements {
		if len(group) == 0 {
			return nil, fmt.Errorf("custom advancement %s has an empty requirement", name)
		}
		for _, criterion := range group {
			if !known[criterion] {
				return nil, fmt.Errorf("custom advancement %s requires the unknown criterion %q", name, criterion)
			}
		}
		groups = append(groups, append([]string(nil), group...))
	}
	// Every criterion has to be part of a requirement: the server refuses
	// an advancement that has a criterion no requirement asks for.
	required := map[string]bool{}
	for _, group := range groups {
		for _, criterion := range group {
			required[criterion] = true
		}
	}
	for _, criterion := range criteria {
		if !required[criterion] {
			return nil, fmt.Errorf("custom advancement %s has the criterion %q, which no requirement asks for", name, criterion)
		}
	}
	return groups, nil
}

// normalizeRecipes turns the recipe names of a "rewards" block into recipe
// ids.
//
// A name without a namespace means "minecraft", and both parts are
// lower-cased, like every other `MinecraftKey`.
func normalizeRecipes(recipes []string) []string {
	if len(recipes) == 0 {
		return nil
	}
	out := make([]string, 0, len(recipes))
	for _, recipe := range recipes {
		out = append(out, advancementRecipeID(recipe))
	}
	return out
}

// advancementSets holds the advancement set of every world.
//
// The set of a world is read once and then kept, so the same world always
// answers the same way. The real server reads the datapack when it
// starts.
var advancementSets sync.Map

// advancementSetOf returns the advancements of a world.
//
// The advancements of the server jar are always there; the ones of the
// world's datapack are read from `<root>/advancements` the first time they
// are needed. The errors of the read are returned the first time only:
// they are the errors of the load, and the load happens once.
func advancementSetOf(w *world.World, root string) (*advancementSet, []error) {
	if v, ok := advancementSets.Load(w); ok {
		set := v.(*advancementSet)
		if set.root == root {
			return set, nil
		}
	}
	custom, errs := loadCustomAdvancements(root)
	set := newAdvancementSet(root, custom)
	advancementSets.Store(w, set)
	return set, errs
}
