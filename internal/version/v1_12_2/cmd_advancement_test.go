package v1_12_2

import (
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
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
	e4 := newAdvancementEngine(t)
	addPlayer(t, e4, "Steve")
	if got, want := advancementFeedback(t, e4, "advancement grant Steve everything"),
		"Granted every advancement (488 total granted) to Steve"; got != want {
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
		a, ok := lookupAdvancement(c.name)
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
		if _, ok := lookupAdvancement(a.parent); !ok {
			t.Errorf("%q has the parent %q, which is not in the registry", a.name, a.parent)
		}
	}
}
