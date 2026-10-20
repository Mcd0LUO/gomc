package v1_12_2

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// testSource is the source the commands of the tests run as.
func testSource() *exec.Source {
	return &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
}

// TestMessageExpansionJoinsNames pins down the connector 1.12.2 puts
// between the names of the entities a selector matched.
//
// `PlayerSelector.getPlayerNames` collects the display names of the
// entities and hands them to `CommandAbstract.a(List<IChatBaseComponent>)`
// (CommandAbstract.java:510), which puts " and " in front of the last
// name and ", " between the others.
//
// (Measured on a real 1.12.2 server: `say @e[type=armor_stand]` with
// armour stands named A, B and C prints "A, B and C".)
func TestMessageExpansionJoinsNames(t *testing.T) {
	for _, tc := range []struct {
		count int
		want  string
	}{
		{1, "[Server] A"},
		{2, "[Server] A and B"},
		{3, "[Server] A, B and C"},
		{4, "[Server] A, B, C and D"},
	} {
		e := newTestEngine(t)
		for i := 0; i < tc.count; i++ {
			name := string(rune('A' + i))
			s, ok := e.World.EntityTypes.NewEntity("armor_stand", fmt.Sprintf("a%d", i))
			if !ok {
				t.Fatalf("the armour stand entity type is missing")
			}
			// The entities are lined up along the x axis, so the
			// selector sorts them in that order.
			s.Pos = world.Vec{X: float64(i), Y: 0, Z: 0}
			s.NBT.Set("CustomName", nbt.String(name))
			e.World.AddEntity(s)
		}
		got, err := runLineFeedback(e, "say @e[type=armor_stand]")
		if err != nil {
			t.Fatalf("say: %v", err)
		}
		if len(got) != 1 || got[0] != tc.want {
			t.Errorf("say said %v, want %q", got, tc.want)
		}
	}
}

// TestMessageExpansionOfAnEmptySelector pins down that a selector that
// matches nothing is an error in a message.
//
// 1.12.2 expands the message in `CommandAbstract.b(listener, args, i,
// true)`, and a selector that matches nothing is reported as
// "Selector '<text>' found nothing" (measured: `say @e[type=zombie]`
// on a server without zombies).
//
// This is the one place where an empty match *is* an error, and it is
// not the place where the command dispatcher expands a target list.
func TestMessageExpansionOfAnEmptySelector(t *testing.T) {
	e := newTestEngine(t)
	err := runLine(e, "say @e[type=zombie]")
	if err == nil {
		t.Fatal("expected an error for a selector that matches nothing")
	}
	want := "Selector '@e[type=zombie]' found nothing"
	if err.Error() != want {
		t.Errorf("the error was %q, want %q", err.Error(), want)
	}
}

// TestTagListNamesTheEntityTheWayTheScoreboardDoes pins down the heading
// of `scoreboard players tag <target> list`.
//
// 1.12.2 names the entity in the heading the way the scoreboard does:
// `CommandAbstract.f` answers a player with its user name and every
// other entity with its UUID. The runtime does not model UUIDs, so a
// non-player entity is named by its internal id.
//
// (Measured on a real 1.12.2 server: the heading of a non-player entity
// is its UUID, e.g. "Tags on entity 44f2a5d4-... are:".)
func TestTagListNamesTheEntityTheWayTheScoreboardDoes(t *testing.T) {
	e := newTestEngine(t)
	p, ok := e.World.EntityTypes.NewEntity("player", "p1")
	if !ok {
		t.Fatalf("the player entity type is missing")
	}
	p.PlayerName = "Steve"
	e.World.AddEntity(p)
	if err := runLine(e, "scoreboard players tag @a add marked"); err != nil {
		t.Fatalf("scoreboard players tag add: %v", err)
	}
	got, err := runLineFeedback(e, "scoreboard players tag @a list")
	if err != nil {
		t.Fatalf("scoreboard players tag list: %v", err)
	}
	// The player is named by its user name, not by its internal id.
	want := []string{"Tags on entity Steve are:", "marked"}
	if len(got) != len(want) {
		t.Fatalf("the listing was %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("the listing said %q, want %q", got[i], want[i])
		}
	}
}

// TestEmptyTargetListIsSilent pins down what an empty target *list* means
// for the commands that resolve their target as an entity.
//
// 1.12.2 expands a target list in the command dispatcher and then runs
// the command once per matched entity. A list that matches nothing
// therefore makes the command run zero times, and the server says
// nothing at all.
//
// (Measured on a real 1.12.2 server: every one of the commands below
// prints no message for `@e[type=zombie]` on a server without zombies,
// and an enclosing `execute` reports "All invocations failed" rather
// than the selector being reported as empty.)
func TestEmptyTargetListIsSilent(t *testing.T) {
	for _, line := range []string{
		"kill @e[type=zombie]",
		"tp @e[type=zombie] 0 4 0",
		"teleport @e[type=zombie] 0 4 0",
		"testfor @e[type=zombie]",
		"effect @e[type=zombie] minecraft:speed 1 1",
		"entitydata @e[type=zombie] {Health:10.0}",
		"execute @e[type=zombie] ~ ~ ~ say hi",
	} {
		e := newTestEngine(t)
		got, err := runLineFeedback(e, line)
		if err != nil {
			t.Errorf("%q: %v", line, err)
			continue
		}
		if len(got) != 0 {
			t.Errorf("%q said %v, want nothing", line, got)
		}
	}
}

// TestEmptyTargetListStillRunsTheCommand checks that an empty target list
// is not an error but still counts as a command that affected nothing.
//
// 1.12.2 counts the commands a function runs, and an empty list makes the
// command run zero times, so the command is not a failure -- the success
// count of the command is zero, and an enclosing `execute` reports "All
// invocations failed" because of that count.
func TestEmptyTargetListStillRunsTheCommand(t *testing.T) {
	e := newTestEngine(t)
	count, err := e.ExecuteLine("kill @e[type=zombie]", 1, testSource())
	if err != nil {
		t.Fatalf("kill: %v", err)
	}
	if count != 0 {
		t.Errorf("the command counted %d, want 0", count)
	}
}

// TestEmptyTargetListInsideAnExecute pins down the wording 1.12.2 uses
// when an empty target list makes the command affect nothing.
//
// (Measured on a real 1.12.2 server: `execute @e[type=armor_stand] ~ ~
// ~ kill @e[type=zombie]` answers "All invocations failed: 'kill
// @e[type=zombie]'", while the same command with a name that is not an
// entity answers "Entity '...' cannot be found".)
func TestEmptyTargetListInsideAnExecute(t *testing.T) {
	e := newTestEngine(t)
	s, _ := e.World.EntityTypes.NewEntity("armor_stand", "a1")
	e.World.AddEntity(s)
	if _, err := e.ExecuteLine("execute @e[type=armor_stand] ~ ~ ~ kill @e[type=zombie]", 1, testSource()); err != nil {
		t.Fatalf("execute: %v", err)
	}
	errs := e.TakeErrors()
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1: %v", len(errs), errs)
	}
	if want := "All invocations failed: 'kill @e[type=zombie]'"; !strings.Contains(errs[0].Error(), want) {
		t.Errorf("the error was %q, want it to contain %q", errs[0], want)
	}
}

// TestEmptySingleTargetIsReported pins down the cases in which 1.12.2
// does report a target that cannot be resolved.
//
// A selector that is limited to a single entity is not a list, so the
// command resolves it itself and reports the missing entity; a name that
// is not an entity is reported the same way.
//
// (Measured on a real 1.12.2 server: `kill @p` and `kill NoSuchPlayer`
// both answer "Entity '...' cannot be found".)
func TestEmptySingleTargetIsReported(t *testing.T) {
	e := newTestEngine(t)
	for _, tc := range []struct{ line, want string }{
		{"kill @p", "Entity '@p' cannot be found"},
		{"kill NoSuchPlayer", "Entity 'NoSuchPlayer' cannot be found"},
		{"kill @e[c=1]", "Entity '@e[c=1]' cannot be found"},
		{"execute NoSuchPlayer ~ ~ ~ say hi", "Entity 'NoSuchPlayer' cannot be found"},
	} {
		err := runLine(e, tc.line)
		if err == nil {
			t.Errorf("%q: expected an error", tc.line)
			continue
		}
		if err.Error() != tc.want {
			t.Errorf("%q: the error was %q, want %q", tc.line, err.Error(), tc.want)
		}
	}
}

// TestSelfWithoutAnExecutor pins down what 1.12.2 answers when @s has
// no executor.
//
// A console, or a function that is not run through /execute, has no
// executor, and then @s matches nothing. The wording comes from the
// command that used the selector.
//
// (Measured on a real 1.12.2 server: `kill @s` answers "Entity '@s'
// cannot be found" and a message that expands @s answers "Selector
// '@s' found nothing".)
func TestSelfWithoutAnExecutor(t *testing.T) {
	e := newTestEngine(t)
	err := runLine(e, "kill @s")
	if err == nil {
		t.Fatal("kill @s: expected an error")
	}
	if want := "Entity '@s' cannot be found"; err.Error() != want {
		t.Errorf("kill @s: the error was %q, want %q", err.Error(), want)
	}
	err = runLine(e, "say @s")
	if err == nil {
		t.Fatal("say @s: expected an error")
	}
	if want := "Selector '@s' found nothing"; err.Error() != want {
		t.Errorf("say @s: the error was %q, want %q", err.Error(), want)
	}
}

// TestInvalidEntityTypeIsNoted pins down the note 1.12.2 prints when a
// selector names an entity type the server does not know.
//
// The type makes the whole selector match nothing, and the command
// still succeeds.
//
// (Measured on a real 1.12.2 server: `kill @e[type=pigzombie]` prints
// "Entity type 'minecraft:pigzombie' is invalid" and succeeds. The
// negated spelling `type=!pigzombie` gets the same note, without the
// `!`.)
func TestInvalidEntityTypeIsNoted(t *testing.T) {
	e := newTestEngine(t)
	got, err := runLineFeedback(e, "kill @e[type=pigzombie]")
	if err != nil {
		t.Fatalf("kill: %v", err)
	}
	want := "Entity type 'minecraft:pigzombie' is invalid"
	if len(got) != 1 || got[0] != want {
		t.Errorf("kill said %v, want [%q]", got, want)
	}
	got, err = runLineFeedback(e, "kill @e[type=!pigzombie]")
	if err != nil {
		t.Fatalf("kill with the negated type: %v", err)
	}
	if len(got) != 1 || got[0] != want {
		t.Errorf("kill with the negated type said %v, want [%q]", got, want)
	}
	// A message that expands the selector fails, and the note comes
	// first.
	e2 := newTestEngine(t)
	before := len(e2.Feedback)
	err = runLine(e2, "say @e[type=pigzombie]")
	notes := e2.Feedback[before:]
	if err == nil {
		t.Fatal("say: expected an error")
	}
	if wantErr := "Selector '@e[type=pigzombie]' found nothing"; err.Error() != wantErr {
		t.Errorf("say: the error was %q, want %q", err.Error(), wantErr)
	}
	if len(notes) != 1 || notes[0] != want {
		t.Errorf("say said %v, want [%q]", notes, want)
	}
}
