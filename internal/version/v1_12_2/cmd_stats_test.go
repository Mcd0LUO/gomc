package v1_12_2

import (
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// newStatsEngine builds an engine that knows the stats command.
//
// The command is registered on its own, so the tests do not depend on the
// whole command set being there.
func newStatsEngine(t *testing.T) *exec.Engine {
	t.Helper()
	reg := exec.NewRegistry()
	registerStats(reg)
	return exec.NewEngine(reg, world.NewWorld())
}

// addStatsEntity puts an entity of the given type into the world.
func addStatsEntity(t *testing.T, e *exec.Engine, id, typ string) *world.Entity {
	t.Helper()
	ent, ok := e.World.EntityTypes.NewEntity(typ, id)
	if !ok {
		t.Fatalf("the %s entity type is missing", typ)
	}
	e.World.AddEntity(ent)
	return ent
}

// statsError runs a command and returns the message of the error it failed
// with.
//
// The message is taken from the error itself, so that the position the
// parser adds to its own errors does not get in the way.
func statsError(t *testing.T, e *exec.Engine, line string) string {
	t.Helper()
	err := runLine(e, line)
	if err == nil {
		t.Fatalf("%q was expected to fail", line)
	}
	switch e := err.(type) {
	case *parser.Error:
		return e.Msg
	case *exec.Error:
		return e.Msg
	}
	return err.Error()
}

// TestStatsUsageLines pins down the usage line of every shape of the
// command.
//
// (Measured on a real 1.12.2 server: each of these says exactly the usage
// line, and the wording of the line is the one of the language file.)
func TestStatsUsageLines(t *testing.T) {
	e := newStatsEngine(t)
	cases := []struct {
		line string
		want string
	}{
		{"stats", "Usage: /stats <entity|block> ..."},
		{"stats bogus", "Usage: /stats <entity|block> ..."},
		{"stats frobnicate @e[type=zombie]", "Usage: /stats <entity|block> ..."},
		{"stats entity", "Usage: /stats entity <selector> <mode> ..."},
		{"stats block", "Usage: /stats block <x> <y> <z> <mode> ..."},
		{"stats block 0 4 0", "Usage: /stats block <x> <y> <z> <mode> ..."},
		{"stats block 0 4 0 set", "Usage: /stats block <x> <y> <z> set <stat> <selector> <objective>"},
		{"stats block 0 4 0 set stat.jump", "Usage: /stats block <x> <y> <z> set <stat> <selector> <objective>"},
		{"stats block 0 4 0 set stat.jump @e[type=zombie]", "Usage: /stats block <x> <y> <z> set <stat> <selector> <objective>"},
		{"stats block 0 4 0 clear", "Usage: /stats block <x> <y> <z> clear <stat>"},
		{"stats block 0 4 0 bogus", "Usage: /stats <entity|block> ..."},
	}
	for _, c := range cases {
		if got := statsError(t, e, c.line); got != c.want {
			t.Errorf("%q said %q, want %q", c.line, got, c.want)
		}
	}
}

// TestStatsEntityFormNeedsATargetThatResolves pins down that the entity
// form looks at the target before anything else.
//
// A target list that matches nothing makes the command run zero times, and
// the server says nothing at all. A single target that cannot be resolved
// is reported as the usage of the entity form.
//
// (Measured on a real 1.12.2 server.)
func TestStatsEntityFormNeedsATargetThatResolves(t *testing.T) {
	e := newStatsEngine(t)

	// A list that matches nothing: the command says nothing.
	msgs, err := runLineFeedback(e, "stats entity @e[type=zombie]")
	if err != nil {
		t.Fatalf("the command failed: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("the command printed %q, want nothing", msgs)
	}

	// A single target that cannot be resolved is the usage of the form.
	for _, line := range []string{
		"stats entity @p",
		"stats entity @s",
		"stats entity Steve",
	} {
		want := "Usage: /stats entity <selector> <mode> ..."
		if got := statsError(t, e, line); got != want {
			t.Errorf("%q said %q, want %q", line, got, want)
		}
	}
}

// TestStatsEntityFormCountsTheArguments pins down the argument counts of
// the entity form.
//
// The counts are only looked at once the target has been resolved.
//
// (Measured on a real 1.12.2 server.)
func TestStatsEntityFormCountsTheArguments(t *testing.T) {
	e := newStatsEngine(t)
	addStatsEntity(t, e, "as-1", "armor_stand")

	cases := []struct {
		line string
		want string
	}{
		{"stats entity @e[type=armor_stand]", "Usage: /stats entity <selector> <mode> ..."},
		{"stats entity @e[type=armor_stand] bogus", "Usage: /stats <entity|block> ..."},
		{"stats entity @e[type=armor_stand] set", "Usage: /stats entity <selector> set <stat> <selector> <objective>"},
		{"stats entity @e[type=armor_stand] set stat.jump", "Usage: /stats entity <selector> set <stat> <selector> <objective>"},
		{"stats entity @e[type=armor_stand] set stat.jump @e[type=armor_stand]", "Usage: /stats entity <selector> set <stat> <selector> <objective>"},
		{"stats entity @e[type=armor_stand] clear", "Usage: /stats entity <selector> clear <stat>"},
	}
	for _, c := range cases {
		if got := statsError(t, e, c.line); got != c.want {
			t.Errorf("%q said %q, want %q", c.line, got, c.want)
		}
	}
}

// TestStatsStatisticNameIsMappedAsACommandResult pins down the bug 1.12.2
// has: the statistic name is mapped onto the command result types, so a
// real statistic name never maps.
//
// (Measured on a real 1.12.2 server.)
func TestStatsStatisticNameIsMappedAsACommandResult(t *testing.T) {
	e := newStatsEngine(t)
	addStatsEntity(t, e, "as-1", "armor_stand")

	for _, line := range []string{
		"stats entity @e[type=armor_stand] set stat.jump @e[type=armor_stand] obj",
		"stats entity @e[type=armor_stand] clear stat.jump",
		"stats block 0 4 0 set stat.jump @e[type=zombie] obj",
		"stats block 0 4 0 clear stat.jump",
	} {
		if got := statsError(t, e, line); got != "Invalid parameters" {
			t.Errorf("%q said %q, want \"Invalid parameters\"", line, got)
		}
	}
}

// TestStatsStoresTheStatistic pins down the one form that works: the
// statistic name is one of the command result names.
//
// The holder and the objective are only written into the message -- 1.12.2
// does not look at either of them.
//
// (Measured on a real 1.12.2 server: `stats entity @e[type=armor_stand] set
// QueryResult Steve obj` says "Storing QueryResult stats in obj on Steve",
// even though there is no player called Steve.)
func TestStatsStoresTheStatistic(t *testing.T) {
	e := newStatsEngine(t)
	addStatsEntity(t, e, "as-1", "armor_stand")

	msgs, err := runLineFeedback(e, "stats entity @e[type=armor_stand] set QueryResult @e[type=armor_stand] obj")
	if err != nil {
		t.Fatalf("the command failed: %v", err)
	}
	if len(msgs) != 1 || msgs[0] != "Storing QueryResult stats in obj on @e[type=armor_stand]" {
		t.Errorf("the command printed %q, want the storing message", msgs)
	}

	msgs, err = runLineFeedback(e, "stats entity @e[type=armor_stand] set QueryResult Steve obj")
	if err != nil {
		t.Fatalf("the command failed: %v", err)
	}
	if len(msgs) != 1 || msgs[0] != "Storing QueryResult stats in obj on Steve" {
		t.Errorf("the command printed %q, want the holder to be used as it was written", msgs)
	}

	msgs, err = runLineFeedback(e, "stats entity @e[type=armor_stand] clear QueryResult")
	if err != nil {
		t.Fatalf("the command failed: %v", err)
	}
	if len(msgs) != 1 || msgs[0] != "Cleared QueryResult stats" {
		t.Errorf("the command printed %q, want the clearing message", msgs)
	}
}

// TestStatsBlockFormRefusesEveryBlock pins down that 1.12.2 cannot store
// the statistics of a block at all.
//
// (Measured on a real 1.12.2 server: a chest, a hopper, a dropper, a
// furnace, a brewing stand and a beacon all get the same answer.)
func TestStatsBlockFormRefusesEveryBlock(t *testing.T) {
	e := newStatsEngine(t)
	for _, line := range []string{
		"stats block 0 4 0 set QueryResult @e[type=zombie] obj",
		"stats block 0 4 0 clear QueryResult",
	} {
		want := "Block at 0, 4, 0 can not track stats"
		if got := statsError(t, e, line); got != want {
			t.Errorf("%q said %q, want %q", line, got, want)
		}
	}
}

// TestStatsPositionIsParsedAfterTheStatistic pins down the order 1.12.2
// looks at the arguments in: the statistic name comes first, the position
// second.
//
// (Measured on a real 1.12.2 server: `stats block a b c set stat.jump
// @e[type=zombie] obj` says "Invalid parameters" and not "'a' is not a valid
// number".)
func TestStatsPositionIsParsedAfterTheStatistic(t *testing.T) {
	e := newStatsEngine(t)

	line := "stats block a b c set stat.jump @e[type=zombie] obj"
	if got := statsError(t, e, line); got != "Invalid parameters" {
		t.Errorf("%q said %q, want \"Invalid parameters\"", line, got)
	}

	line = "stats block a b c set QueryResult @e[type=zombie] obj"
	if got := statsError(t, e, line); got != "'a' is not a valid number" {
		t.Errorf("%q said %q, want the number error", line, got)
	}
}

// TestStatsRejectsASelectorThatDoesNotParse pins down that a selector that
// does not parse is reported before anything else is looked at.
//
// (Measured on a real 1.12.2 server: `stats entity @e[bad]` says
// "Invalid selector argument: 'bad'".)
func TestStatsRejectsASelectorThatDoesNotParse(t *testing.T) {
	e := newStatsEngine(t)
	line := "stats entity @e[bad]"
	if got := statsError(t, e, line); got != "Invalid selector argument: 'bad'" {
		t.Errorf("%q said %q, want the selector error", line, got)
	}
}
