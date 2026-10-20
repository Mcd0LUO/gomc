package v1_12_2

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// newAdvancementEngine builds an engine with the advancement command.
//
// The command is registered here instead of in Register, because the
// version package wires the commands together in one place.
func newAdvancementEngine(t *testing.T) *exec.Engine {
	t.Helper()
	reg := exec.NewRegistry()
	Register(reg)
	registerAdvancement(reg)
	return exec.NewEngine(reg, world.NewWorld())
}

// advancementFeedback runs a command and returns the message it printed.
func advancementFeedback(t *testing.T, e *exec.Engine, line string) string {
	t.Helper()
	msgs, err := runLineFeedback(e, line)
	if err != nil {
		t.Fatalf("%q: %v", line, err)
	}
	if len(msgs) != 1 {
		t.Fatalf("%q: the command printed %d messages, want one", line, len(msgs))
	}
	return msgs[0]
}

// advancementError runs a command and returns the error it failed with.
func advancementError(t *testing.T, e *exec.Engine, line string) string {
	t.Helper()
	err := runLine(e, line)
	if err == nil {
		t.Fatalf("%q: expected an error", line)
	}
	return err.Error()
}

// TestAdvancementGrantModes pins down the five filters of
// `/advancement grant`.
//
// 1.12.2 reaches the ancestors with "until", the descendants with "from"
// and both with "through"; "only" touches the advancement itself and
// "everything" touches all of them.
//
// (Measured on a real 1.12.2 server.)
func TestAdvancementGrantModes(t *testing.T) {
	e := newAdvancementEngine(t)
	addPlayer(t, e, "Steve")

	// "only" touches the advancement itself.
	if got, want := advancementFeedback(t, e, "advancement grant Steve only minecraft:story/root"),
		"Granted the entire advancement 'minecraft:story/root' to Steve"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	// The story tree has 15 descendants below its root, so "from" reaches
	// 16 advancements.
	e1 := newAdvancementEngine(t)
	addPlayer(t, e1, "Steve")
	if got, want := advancementFeedback(t, e1, "advancement grant Steve from minecraft:story/root"),
		"Granted 'minecraft:story/root' and all descendants (16 total granted) to Steve"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	// "until" reaches the advancement and its parents.
	e2 := newAdvancementEngine(t)
	addPlayer(t, e2, "Steve")
	if got, want := advancementFeedback(t, e2, "advancement grant Steve until minecraft:story/mine_stone"),
		"Granted 'minecraft:story/mine_stone' and all ancestors (2 total granted) to Steve"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	// "through" reaches the parents and the descendants.
	e3 := newAdvancementEngine(t)
	addPlayer(t, e3, "Steve")
	if got, want := advancementFeedback(t, e3, "advancement grant Steve through minecraft:story/mine_stone"),
		"Granted 'minecraft:story/mine_stone', all ancestors and all descendants (16 total granted) to Steve"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	// "everything" reaches every advancement of the registry.
	//
	// The number is smaller than the registry holds: an advancement the
	// recipe trigger completed before the command reached it is skipped,
	// and the command does not count it.
	//
	// (Measured on a real 1.12.2 server: 481.)
	e4 := newAdvancementEngine(t)
	addPlayer(t, e4, "Steve")
	if got, want := advancementFeedback(t, e4, "advancement grant Steve everything"),
		"Granted every advancement (481 total granted) to Steve"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	// Granting everything twice changes nothing, and the server says so.
	if got, want := advancementError(t, e4, "advancement grant Steve everything"),
		"Couldn't grant any advancements to Steve because they already have them all"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestAdvancementGrantFailsWhenAlreadyDone pins down what 1.12.2 answers
// when the player already has the advancement.
//
// (Measured on a real 1.12.2 server.)
func TestAdvancementGrantFailsWhenAlreadyDone(t *testing.T) {
	e := newAdvancementEngine(t)
	addPlayer(t, e, "Steve")
	advancementFeedback(t, e, "advancement grant Steve only minecraft:story/root")
	if got, want := advancementError(t, e, "advancement grant Steve only minecraft:story/root"),
		"Couldn't grant the advancement 'minecraft:story/root' to Steve because they already have it"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestAdvancementCriterionGrantAndRevoke pins down the criterion level of
// the command.
//
// 1.12.2 completes a single criterion and reports it by name, and it
// refuses a criterion the advancement does not have.
//
// (Measured on a real 1.12.2 server.)
func TestAdvancementCriterionGrantAndRevoke(t *testing.T) {
	e := newAdvancementEngine(t)
	addPlayer(t, e, "Steve")

	if got, want := advancementFeedback(t, e, "advancement grant Steve only minecraft:story/mine_stone get_stone"),
		"Granted the criterion 'get_stone' of advancement 'minecraft:story/mine_stone' to Steve"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	// The same criterion twice changes nothing.
	if got, want := advancementError(t, e, "advancement grant Steve only minecraft:story/mine_stone get_stone"),
		"Couldn't grant the criterion 'get_stone' of advancement 'minecraft:story/mine_stone' to Steve because they already have it"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
	// A criterion the advancement does not have is refused.
	if got, want := advancementError(t, e, "advancement grant Steve only minecraft:story/root nosuchcriterion"),
		"The advancement 'minecraft:story/root' does not contain the criterion 'nosuchcriterion'"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
	// Revoking takes the criterion away again.
	if got, want := advancementFeedback(t, e, "advancement revoke Steve only minecraft:story/mine_stone get_stone"),
		"Revoked the criterion 'get_stone' of advancement 'minecraft:story/mine_stone' from Steve"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	// Revoking it again changes nothing.
	if got, want := advancementError(t, e, "advancement revoke Steve only minecraft:story/mine_stone get_stone"),
		"Couldn't revoke the criterion 'get_stone' of advancement 'minecraft:story/mine_stone' from Steve because they haven't started it"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestAdvancementRevoke pins down `/advancement revoke`.
//
// (Measured on a real 1.12.2 server.)
func TestAdvancementRevoke(t *testing.T) {
	e := newAdvancementEngine(t)
	addPlayer(t, e, "Steve")
	advancementFeedback(t, e, "advancement grant Steve everything")

	if got, want := advancementFeedback(t, e, "advancement revoke Steve only minecraft:story/root"),
		"Revoked the entire advancement 'minecraft:story/root' from Steve"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	// The advancement is not started any more, so revoking it again
	// changes nothing.
	if got, want := advancementError(t, e, "advancement revoke Steve only minecraft:story/root"),
		"Couldn't revoke the advancement 'minecraft:story/root' from Steve because they haven't started it"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
	// "until" takes the ancestors away as well.
	if got, want := advancementFeedback(t, e, "advancement revoke Steve until minecraft:story/mine_stone"),
		"Revoked 'minecraft:story/mine_stone' and all ancestors (1 total revoked) from Steve"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
}

// TestAdvancementTestCommand pins down `/advancement test`.
//
// (Measured on a real 1.12.2 server.)
func TestAdvancementTestCommand(t *testing.T) {
	e := newAdvancementEngine(t)
	addPlayer(t, e, "Steve")

	// An advancement the player does not have.
	if got, want := advancementError(t, e, "advancement test Steve minecraft:story/root"),
		"Player Steve has not completed advancement 'minecraft:story/root'"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
	advancementFeedback(t, e, "advancement grant Steve only minecraft:story/root")

	if got, want := advancementFeedback(t, e, "advancement test Steve minecraft:story/root"),
		"Player Steve has completed advancement 'minecraft:story/root'"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	if got, want := advancementFeedback(t, e, "advancement test Steve minecraft:story/root crafting_table"),
		"Player Steve has completed criterion 'crafting_table' of advancement 'minecraft:story/root'"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	// A criterion the player does not have.
	if got, want := advancementError(t, e, "advancement test Steve minecraft:story/mine_stone get_stone"),
		"Player Steve has not completed criterion 'get_stone' of advancement 'minecraft:story/mine_stone'"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
	// A criterion the advancement does not have at all.
	if got, want := advancementError(t, e, "advancement test Steve minecraft:story/root nosuchcriterion"),
		"The advancement 'minecraft:story/root' does not contain the criterion 'nosuchcriterion'"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestAdvancementNotFound pins down the error for an unknown advancement.
//
// 1.12.2 reports the name as it was written.
//
// (Measured on a real 1.12.2 server.)
func TestAdvancementNotFound(t *testing.T) {
	e := newAdvancementEngine(t)
	addPlayer(t, e, "Steve")
	if got, want := advancementError(t, e, "advancement test Steve minecraft:nosuchadv"),
		"No advancement was found by the name 'minecraft:nosuchadv'"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
	if got, want := advancementError(t, e, "advancement grant Steve only minecraft:nosuchadv"),
		"No advancement was found by the name 'minecraft:nosuchadv'"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestAdvancementPlayerNotFound pins down the error for a player nobody
// has.
//
// (Measured on a real 1.12.2 server.)
func TestAdvancementPlayerNotFound(t *testing.T) {
	e := newAdvancementEngine(t)
	if got, want := advancementError(t, e, "advancement grant NoSuchPlayer everything"),
		"Player 'NoSuchPlayer' cannot be found"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
	if got, want := advancementError(t, e, "advancement test NoSuchPlayer minecraft:story/root"),
		"Player 'NoSuchPlayer' cannot be found"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestAdvancementUsageErrors pins down the usage lines.
//
// 1.12.2 throws an `ExceptionUsage` for the mistakes it recognises as a
// usage mistake, and the server prints those with the "Usage: " prefix.
// The messages that name a subcommand are thrown as a plain
// `CommandException` and come out without the prefix.
//
// (Measured on a real 1.12.2 server.)
func TestAdvancementUsageErrors(t *testing.T) {
	e := newAdvancementEngine(t)
	addPlayer(t, e, "Steve")
	for _, c := range []struct {
		line string
		want string
	}{
		{"advancement", "Usage: /advancement <grant|revoke|test> <player>"},
		{"advancement bogus", "Usage: /advancement <grant|revoke|test> <player>"},
		{"advancement grant Steve", "/advancement grant <player> <only|until|from|through|everything>"},
		{"advancement revoke Steve", "/advancement revoke <player> <only|until|from|through|everything>"},
		{"advancement grant Steve bogus minecraft:story/root", "/advancement grant <player> <only|until|from|through|everything>"},
		{"advancement grant Steve only", "/advancement grant <player> only <advancement> [criterion]"},
		{"advancement grant Steve until", "/advancement grant <player> until <advancement>"},
		{"advancement grant Steve from", "/advancement grant <player> from <advancement>"},
		{"advancement grant Steve through", "/advancement grant <player> through <advancement>"},
		{"advancement grant Steve everything extra", "/advancement grant <player> everything"},
		{"advancement test Steve", "Usage: /advancement test <player> <advancement> [criterion]"},
	} {
		got := advancementError(t, e, c.line)
		if !strings.Contains(got, c.want) {
			t.Errorf("%q: the error was %q, want it to contain %q", c.line, got, c.want)
		}
	}
}

// TestAdvancementNames pins down the names the command accepts.
//
// 1.12.2 reads the name as a `MinecraftKey`, so the namespace defaults to
// "minecraft" and the text is lower-cased. The messages name the
// advancement by its full name.
//
// (Measured on a real 1.12.2 server.)
func TestAdvancementNames(t *testing.T) {
	e := newAdvancementEngine(t)
	addPlayer(t, e, "Steve")
	if got, want := advancementFeedback(t, e, "advancement grant Steve only story/root"),
		"Granted the entire advancement 'minecraft:story/root' to Steve"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	// The namespace of another pack is not known.
	if got, want := advancementError(t, e, "advancement grant Steve only other:story/root"),
		"No advancement was found by the name 'other:story/root'"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestAdvancementRequirements pins down what makes an advancement
// complete.
//
// 1.12.2 reads the "requirements" array of the advancement JSON: every
// group has to be satisfied and a group is satisfied as soon as one of
// its criteria is done. An advancement without a "requirements" array
// needs all of its criteria, one that declares a single group with every
// criterion in it needs only one.
//
// "minecraft:story/root" declares no requirements and has one criterion,
// "minecraft:adventure/kill_a_mob" declares one group with all of its 22
// criteria.
//
// (Measured on a real 1.12.2 server.)
func TestAdvancementRequirements(t *testing.T) {
	e := newAdvancementEngine(t)
	addPlayer(t, e, "Steve")

	// One criterion of the 22 is enough for the advancement.
	advancementFeedback(t, e, "advancement grant Steve only minecraft:adventure/kill_a_mob blaze")
	if got, want := advancementFeedback(t, e, "advancement test Steve minecraft:adventure/kill_a_mob"),
		"Player Steve has completed advancement 'minecraft:adventure/kill_a_mob'"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	// The other criteria are not done, because the command completed a
	// single criterion.
	if got, want := advancementError(t, e, "advancement test Steve minecraft:adventure/kill_a_mob zombie"),
		"Player Steve has not completed criterion 'zombie' of advancement 'minecraft:adventure/kill_a_mob'"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
	// An advancement that needs all of its criteria is not complete after
	// one of them.
	e2 := newAdvancementEngine(t)
	addPlayer(t, e2, "Steve")
	advancementFeedback(t, e2, "advancement grant Steve only minecraft:adventure/kill_all_mobs blaze")
	if got, want := advancementError(t, e2, "advancement test Steve minecraft:adventure/kill_all_mobs"),
		"Player Steve has not completed advancement 'minecraft:adventure/kill_all_mobs'"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestAdvancementStateIsPerWorld pins down that the progress of a player
// belongs to the world it was made in.
func TestAdvancementStateIsPerWorld(t *testing.T) {
	e := newAdvancementEngine(t)
	addPlayer(t, e, "Steve")
	advancementFeedback(t, e, "advancement grant Steve only minecraft:story/root")

	other := newAdvancementEngine(t)
	addPlayer(t, other, "Steve")
	if got, want := advancementError(t, other, "advancement test Steve minecraft:story/root"),
		"Player Steve has not completed advancement 'minecraft:story/root'"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestAdvancementRegistry pins down the registry itself.
//
// The data comes from the JSON files of the 1.12.2 server jar, so the
// numbers below are the ones of that jar.
func TestAdvancementRegistry(t *testing.T) {
	if len(advancementRegistry) != 488 {
		t.Errorf("the registry holds %d advancements, want 488", len(advancementRegistry))
	}
	roots := 0
	for i := range advancementRegistry {
		if advancementRegistry[i].parent == "" {
			roots++
		}
	}
	if roots != 6 {
		t.Errorf("the registry holds %d roots, want 6", roots)
	}
	set := newAdvancementSet("", nil)
	for _, c := range []struct {
		name     string
		parent   string
		criteria int
	}{
		{"minecraft:story/root", "", 1},
		{"minecraft:story/mine_stone", "minecraft:story/root", 1},
		{"minecraft:adventure/kill_a_mob", "minecraft:adventure/root", 22},
		{"minecraft:adventure/kill_all_mobs", "minecraft:adventure/kill_a_mob", 22},
		{"minecraft:recipes/root", "", 1},
	} {
		a, ok := set.lookup(c.name)
		if !ok {
			t.Errorf("the registry does not hold %q", c.name)
			continue
		}
		if a.parent != c.parent {
			t.Errorf("%q has the parent %q, want %q", c.name, a.parent, c.parent)
		}
		if len(a.criteria) != c.criteria {
			t.Errorf("%q has %d criteria, want %d", c.name, len(a.criteria), c.criteria)
		}
	}
	// Every parent has to be an advancement of the registry, or the tree
	// would break in the middle.
	for i := range advancementRegistry {
		a := &advancementRegistry[i]
		if a.parent == "" {
			continue
		}
		if _, ok := set.lookup(a.parent); !ok {
			t.Errorf("%q has the parent %q, which is not in the registry", a.name, a.parent)
		}
	}
}

// writeAdvancementFile writes an advancement file of a datapack.
//
// The path is relative to the "advancements" directory of the world.
func writeAdvancementFile(t *testing.T, root, name, body string) {
	t.Helper()
	path := filepath.Join(root, "advancements", filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

// TestAdvancementCustomDatapack pins down the custom advancements of a
// world.
//
// 1.12.2 keeps them in `<world>/advancements/<namespace>/<name>.json` and
// reads them when the server starts. The namespace is the first segment
// of the path and the name is the rest, and both are lower-cased like
// every other `MinecraftKey`.
func TestAdvancementCustomDatapack(t *testing.T) {
	root := t.TempDir()
	writeAdvancementFile(t, root, "my_pack/thing.json", `{
		"parent": "minecraft:story/root",
		"criteria": {"done": {"trigger": "minecraft:impossible"}},
		"requirements": [["done"]]
	}`)
	e := newAdvancementEngine(t)
	e.SetRoot(root)
	addPlayer(t, e, "Steve")

	if got, want := advancementFeedback(t, e, "advancement grant Steve only my_pack:thing"),
		"Granted the entire advancement 'my_pack:thing' to Steve"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	if got, want := advancementFeedback(t, e, "advancement test Steve my_pack:thing"),
		"Player Steve has completed advancement 'my_pack:thing'"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	// The name of the advancement is read as a `MinecraftKey`, so the
	// namespace of the pack is needed.
	if got, want := advancementError(t, e, "advancement test Steve thing"),
		"No advancement was found by the name 'thing'"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
	// The advancements of the server jar are still there.
	if got, want := advancementError(t, e, "advancement test Steve minecraft:story/root"),
		"Player Steve has not completed advancement 'minecraft:story/root'"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestAdvancementCustomRequirements pins down the "requirements" array of
// a custom advancement.
//
// 1.12.2 reads the array: every group has to be satisfied, and a group is
// satisfied as soon as one of its criteria is done. An advancement
// without a "requirements" array needs all of its criteria.
func TestAdvancementCustomRequirements(t *testing.T) {
	root := t.TempDir()
	writeAdvancementFile(t, root, "my_pack/grouped.json", `{
		"criteria": {"a": {"trigger": "minecraft:impossible"}, "b": {"trigger": "minecraft:impossible"}},
		"requirements": [["a", "b"]]
	}`)
	writeAdvancementFile(t, root, "my_pack/all.json", `{
		"criteria": {"a": {"trigger": "minecraft:impossible"}, "b": {"trigger": "minecraft:impossible"}}
	}`)
	e := newAdvancementEngine(t)
	e.SetRoot(root)
	addPlayer(t, e, "Steve")

	// One criterion of the group is enough.
	advancementFeedback(t, e, "advancement grant Steve only my_pack:grouped a")
	if got, want := advancementFeedback(t, e, "advancement test Steve my_pack:grouped"),
		"Player Steve has completed advancement 'my_pack:grouped'"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	// An advancement without a "requirements" array needs all of them.
	advancementFeedback(t, e, "advancement grant Steve only my_pack:all a")
	if got, want := advancementError(t, e, "advancement test Steve my_pack:all"),
		"Player Steve has not completed advancement 'my_pack:all'"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestAdvancementCustomBrokenFile pins down what happens to a file the
// server would not accept.
//
// 1.12.2 refuses an advancement without criteria, one that requires a
// criterion it does not have, and one that has a criterion no requirement
// asks for. Such a file does not become an advancement at all.
func TestAdvancementCustomBrokenFile(t *testing.T) {
	root := t.TempDir()
	writeAdvancementFile(t, root, "my_pack/no_criteria.json", `{
		"criteria": {}
	}`)
	writeAdvancementFile(t, root, "my_pack/unknown_criterion.json", `{
		"criteria": {"a": {"trigger": "minecraft:impossible"}},
		"requirements": [["b"]]
	}`)
	writeAdvancementFile(t, root, "my_pack/unused_criterion.json", `{
		"criteria": {"a": {"trigger": "minecraft:impossible"}, "b": {"trigger": "minecraft:impossible"}},
		"requirements": [["a"]]
	}`)
	e := newAdvancementEngine(t)
	e.SetRoot(root)
	addPlayer(t, e, "Steve")

	for _, name := range []string{"my_pack:no_criteria", "my_pack:unknown_criterion", "my_pack:unused_criterion"} {
		if got, want := advancementError(t, e, "advancement test Steve "+name),
			"No advancement was found by the name '"+name+"'"; got != want {
			t.Errorf("the error was %q, want %q", got, want)
		}
	}
	// Every broken file is reported in the log of the server.
	if len(e.Logs) != 3 {
		t.Errorf("the server logged %d problems, want 3", len(e.Logs))
	}
}

// TestAdvancementRecipeUnlock pins down the recipe trigger.
//
// 1.12.2 hands out the recipes of the "rewards" block of an advancement
// when the advancement is completed, and learning a recipe completes the
// advancement that is about that recipe. The trigger only fires for a
// recipe the player does not know yet.
//
// Twelve of the recipe advancements hand out a recipe that is about
// another advancement, so the trigger completes that other advancement
// before the command reaches it -- and the command skips it.
//
// (Measured on a real 1.12.2 server: after `advancement grant <player>
// everything`, seven of them are only half done, and the command answered
// "Granted every advancement (481 total granted) to <player>" both times.)
func TestAdvancementRecipeUnlock(t *testing.T) {
	e := newAdvancementEngine(t)
	addPlayer(t, e, "Steve")
	advancementFeedback(t, e, "advancement grant Steve everything")

	// The advancements the trigger completed before the command reached
	// them, and the criterion of each that the command did not grant.
	skipped := map[string]string{
		"minecraft:recipes/misc/coal":                  "has_coal_block",
		"minecraft:recipes/misc/diamond":               "has_diamond_block",
		"minecraft:recipes/misc/emerald":               "has_emerald_block",
		"minecraft:recipes/misc/gold_ingot_from_block": "has_gold_block",
		"minecraft:recipes/misc/lapis_lazuli":          "has_lapis_block",
		"minecraft:recipes/misc/slime_ball":            "has_slime",
		"minecraft:recipes/redstone/redstone":          "has_redstone_block",
	}
	// The ones the command reached first, so it granted all of them.
	granted := map[string]string{
		"minecraft:recipes/misc/bone_meal_from_block":  "has_bone_block",
		"minecraft:recipes/misc/wheat":                 "has_hay_block",
		"minecraft:recipes/misc/gold_nugget":           "has_gold_ingot",
		"minecraft:recipes/misc/iron_nugget":           "has_iron_ingot",
		"minecraft:recipes/misc/iron_ingot_from_block": "has_iron_block",
	}
	for advancement, criterion := range skipped {
		if got, want := advancementFeedback(t, e, "advancement test Steve "+advancement+" has_the_recipe"),
			"Player Steve has completed criterion 'has_the_recipe' of advancement '"+advancement+"'"; got != want {
			t.Errorf("the message was %q, want %q", got, want)
		}
		if got, want := advancementError(t, e, "advancement test Steve "+advancement+" "+criterion),
			"Player Steve has not completed criterion '"+criterion+"' of advancement '"+advancement+"'"; got != want {
			t.Errorf("the error was %q, want %q", got, want)
		}
	}
	for advancement, criterion := range granted {
		if got, want := advancementFeedback(t, e, "advancement test Steve "+advancement+" "+criterion),
			"Player Steve has completed criterion '"+criterion+"' of advancement '"+advancement+"'"; got != want {
			t.Errorf("the message was %q, want %q", got, want)
		}
	}
	// The player knows the recipes the advancements handed out.
	p := e.World.FindPlayer("Steve")
	known := knownRecipes(p)
	if !known["minecraft:gold_ingot_from_block"] {
		t.Errorf("the player does not know the gold ingot recipe")
	}
	if !known["minecraft:acacia_planks"] {
		t.Errorf("the player does not know the acacia planks recipe")
	}
}

// advancementExperience checks the experience of a player.
func advancementExperience(t *testing.T, p *world.Entity, total, level int) {
	t.Helper()
	if got, _ := p.NBT.Get("XpTotal").(nbt.Int); int(got) != total {
		t.Errorf("the total experience is %d, want %d", got, total)
	}
	if got, _ := p.NBT.Get("XpLevel").(nbt.Int); int(got) != level {
		t.Errorf("the experience level is %d, want %d", got, level)
	}
}

// TestAdvancementExperienceReward pins down the experience reward of an
// advancement.
//
// 1.12.2 hands out the experience of the "rewards" block of an advancement
// when the advancement is completed. The experience is *points*: they go
// into the total experience of the player and move the level up along the
// usual curve.
//
// (Measured on a real 1.12.2 server: granting "minecraft:adventure/
// kill_all_mobs" -- 100 experience -- leaves the player with 100 total
// experience and level 7, granting "minecraft:adventure/sniper_duel" --
// 50 experience -- with 150 and level 9, and granting "minecraft:nether/
// all_effects" -- 1000 experience -- with 1150 and level 27.)
func TestAdvancementExperienceReward(t *testing.T) {
	e := newAdvancementEngine(t)
	addPlayer(t, e, "Steve")
	p := e.World.FindPlayer("Steve")

	advancementFeedback(t, e, "advancement grant Steve only minecraft:adventure/kill_all_mobs")
	advancementExperience(t, p, 100, 7)

	advancementFeedback(t, e, "advancement grant Steve only minecraft:adventure/sniper_duel")
	advancementExperience(t, p, 150, 9)

	advancementFeedback(t, e, "advancement grant Steve only minecraft:nether/all_effects")
	advancementExperience(t, p, 1150, 27)

	// The progress towards the next level is the one of the server: 61
	// of the 97 points of level 28.
	if got, _ := p.NBT.Get("XpP").(nbt.Float); float64(got) < 0.6288 || float64(got) > 0.6289 {
		t.Errorf("the progress towards the next level is %v, want about 0.62887", got)
	}
}

// TestAdvancementExperienceOfACustomAdvancement pins down the experience
// reward of a custom advancement, and what the server does with a reward
// it does not know.
//
// 1.12.2 reads four keys of the "rewards" block: "experience", "loot",
// "recipes" and "function". A key it does not know -- "items" is the one
// of 1.13 -- is ignored: the advancement still loads and the other
// rewards still work.
//
// (Measured on a real 1.12.2 server: an advancement with an "items"
// reward and an "experience" of 7 loads, hands out the 7 points and says
// nothing about the items.)
func TestAdvancementExperienceOfACustomAdvancement(t *testing.T) {
	root := t.TempDir()
	writeAdvancementFile(t, root, "my_pack/thing.json", `{
		"criteria": {"done": {"trigger": "minecraft:impossible"}},
		"requirements": [["done"]],
		"rewards": {"items": [{"item": "minecraft:diamond", "count": 5}], "experience": 7}
	}`)
	e := newAdvancementEngine(t)
	e.SetRoot(root)
	addPlayer(t, e, "Steve")
	p := e.World.FindPlayer("Steve")

	// The advancement exists: the key the server does not know does not
	// keep it from loading.
	if got, want := advancementFeedback(t, e, "advancement grant Steve only my_pack:thing"),
		"Granted the entire advancement 'my_pack:thing' to Steve"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	// The experience of the block is handed out, the items are not: gomc
	// has no items to hand out, and the server does not know the key
	// either.
	advancementExperience(t, p, 7, 1)
}

// writeAdvancementProgressFile writes the progress file of a player.
//
// The file is the one 1.12.2 keeps in `<world>/advancements/<uuid>.json`.
func writeAdvancementProgressFile(t *testing.T, root, uuid, body string) {
	t.Helper()
	path := filepath.Join(root, "advancements", uuid+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

// TestAdvancementProgressIsSaved pins down the progress file of a player.
//
// 1.12.2 keeps the progress of a player in `advancements/<uuid>.json`, in
// the directory of the world, and writes it when the progress changes.
//
// The file holds the advancements the player has started, with the moment
// every criterion was completed and the "done" flag of the advancement.
//
// (Measured on a real 1.12.2 server: the file of a player holds the
// advancements the player has started, with the moment every criterion
// was completed and the "done" flag of the advancement.)
func TestAdvancementProgressIsSaved(t *testing.T) {
	root := t.TempDir()
	e := newAdvancementEngine(t)
	e.SetRoot(root)
	addPlayer(t, e, "Steve")

	advancementFeedback(t, e, "advancement grant Steve only minecraft:story/root")

	path := filepath.Join(root, "advancements", world.DeriveEntityUUID("Steve")+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	want := `{
  "minecraft:story/root": {
    "criteria": {
      "crafting_table": "1970-01-01 00:00:00 +0000"
    },
    "done": true
  }
}`
	if string(data) != want {
		t.Errorf("the file holds\n%s\nwant\n%s", data, want)
	}
}

// TestAdvancementProgressIsRead pins down that the progress of a player is
// read back.
//
// 1.12.2 reads the file when the player joins: the criteria that are in it
// are the ones the player has completed. The "done" flag of the file is not
// read -- the server works it out from the requirements of the
// advancement, and so does gomc.
func TestAdvancementProgressIsRead(t *testing.T) {
	root := t.TempDir()
	writeAdvancementProgressFile(t, root, world.DeriveEntityUUID("Steve"), `{
  "minecraft:story/root": {
    "criteria": {
      "crafting_table": "2026-10-20 21:44:38 +0800"
    },
    "done": true
  }
}`)
	e := newAdvancementEngine(t)
	e.SetRoot(root)
	addPlayer(t, e, "Steve")

	if got, want := advancementFeedback(t, e, "advancement test Steve minecraft:story/root"),
		"Player Steve has completed advancement 'minecraft:story/root'"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
	// The player cannot be given the advancement again: it is already
	// completed.
	if got, want := advancementError(t, e, "advancement grant Steve only minecraft:story/root"),
		"Couldn't grant the advancement 'minecraft:story/root' to Steve because they already have it"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
}

// TestAdvancementProgressIsNotACustomAdvancement pins down that the
// progress file of a player is not read as a custom advancement.
//
// The two live in the same directory: the progress of a player is a file
// right in it (`<world>/advancements/<uuid>.json`) and a custom
// advancement is in a directory named after its namespace
// (`<world>/advancements/<namespace>/<path>.json`). The server skips
// every file that is not in such a directory.
func TestAdvancementProgressIsNotACustomAdvancement(t *testing.T) {
	root := t.TempDir()
	uuid := world.DeriveEntityUUID("Steve")
	writeAdvancementProgressFile(t, root, uuid, `{
  "minecraft:story/root": {
    "criteria": {
      "crafting_table": "2026-10-20 21:44:38 +0800"
    },
    "done": true
  }
}`)
	e := newAdvancementEngine(t)
	e.SetRoot(root)
	addPlayer(t, e, "Steve")

	if got, want := advancementError(t, e, "advancement test Steve minecraft:"+uuid),
		"No advancement was found by the name 'minecraft:"+uuid+"'"; got != want {
		t.Errorf("the error was %q, want %q", got, want)
	}
	// The file is still the progress of the player.
	if got, want := advancementFeedback(t, e, "advancement test Steve minecraft:story/root"),
		"Player Steve has completed advancement 'minecraft:story/root'"; got != want {
		t.Errorf("the message was %q, want %q", got, want)
	}
}
