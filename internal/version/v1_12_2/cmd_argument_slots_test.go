package v1_12_2

import (
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
)

// TestQuotesStayInTheMessage covers the way 1.12.2 cuts a command line.
//
// The server splits the command with `String.split(" ")`, so a quote is an
// ordinary character that belongs to the argument.
//
// (Measured on a real 1.12.2 server: `say "a b"` prints `"a b"`, the quotes
// included.)
func TestQuotesStayInTheMessage(t *testing.T) {
	e := newTestEngine(t)
	msgs, err := runLineFeedback(e, `say "a b"`)
	if err != nil {
		t.Fatalf("say: %v", err)
	}
	if len(msgs) != 1 || msgs[0] != `[Server] "a b"` {
		t.Errorf("messages = %v, want [Server] \"a b\"", msgs)
	}
}

// TestEmptyFirstArgumentIsAUsageError covers the argument that is empty
// because the line has two spaces in a row.
//
// 1.12.2 refuses the command when the **first** argument is empty
// (`CommandSay.execute`: `stringArray.length <= 0 || stringArray[0].length()
// <= 0`).
func TestEmptyFirstArgumentIsAUsageError(t *testing.T) {
	e := newTestEngine(t)
	err := runLine(e, "say  hi")
	if err == nil {
		t.Fatal("expected an error for a command with an empty first argument")
	}
	if !strings.Contains(err.Error(), "Usage: /say <message ...>") {
		t.Errorf("error = %q, want the usage of the command", err)
	}
}

// TestSetBlockAcceptsAnyMode covers the mode of `/setblock`.
//
// 1.12.2 only gives a meaning to `destroy` and `keep`; every other word
// behaves like `replace` and is **not** a mistake.
//
// (Measured on a real 1.12.2 server: `setblock 4 4 4 stone 0 0` places the
// block.)
func TestSetBlockAcceptsAnyMode(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "setblock 4 4 4 stone 0 0"); err != nil {
		t.Fatalf("setblock: %v", err)
	}
}

// TestTitleTreatsAnUnknownActionAsATitle covers the action of `/title`.
//
// 1.12.2 maps the action to the enum `PacketPlayOutTitle.EnumTitleAction`,
// whose `a(String)` **falls back to `title`** for an action it does not
// know. So a misspelled action is a title, and the JSON is the rest of the
// line.
func TestTitleTreatsAnUnknownActionAsATitle(t *testing.T) {
	e := newTestEngine(t)
	addPlayer(t, e, "Steve")
	// The action "time" is not one of the enum's names, so the command is
	// a title -- and its JSON is "0 30 0", which is not JSON.
	err := runLine(e, "title Steve time 0 30 0")
	if err == nil {
		t.Fatal("expected an error for a title whose JSON is broken")
	}
	if !strings.Contains(err.Error(), "Invalid json: ") {
		t.Errorf("error = %q, want an invalid json", err)
	}
	// A real title works.
	if err := runLine(e, `title Steve title {"text":"hi"}`); err != nil {
		t.Fatalf("title: %v", err)
	}
}

// TestScoreboardPlayersCountsItsArguments covers the argument count of the
// `scoreboard players` subcommands.
//
// 1.12.2 counts the arguments of the subcommand before it looks at any of
// them, and every subcommand has its own usage line.
func TestScoreboardPlayersCountsItsArguments(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"scoreboard players add Steve", "Usage: /scoreboard players add <player> <objective> <count> [dataTag]"},
		{"scoreboard players operation @s a *= b", "Usage: /scoreboard players operation <targetName> <targetObjective> <operation> <selector> <objective>"},
		{"scoreboard players redset @s obj", "Usage: /scoreboard players <set|add|remove|reset|list|enable|test|operation|tag> ..."},
		{"scoreboard players tag @s reset foo", "Usage: /scoreboard players tag <player> <add|remove|list> <tagName> [dataTag]"},
		{"scoreboard objectives add math.", "Usage: /scoreboard objectives add <name> <criteriaType> [display name ...]"},
	}
	for _, tc := range tests {
		e := newTestEngine(t)
		err := runLine(e, tc.line)
		if err == nil {
			t.Errorf("%q: expected an error", tc.line)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%q: error = %q, want %q", tc.line, err, tc.want)
		}
	}
}

// TestScoreboardPlayersOperationNeedsSevenArguments covers the one
// subcommand whose count is exact.
//
// 1.12.2 wants **exactly** seven arguments for an operation, so a command
// with one argument too many is a usage mistake.
func TestScoreboardPlayersOperationNeedsSevenArguments(t *testing.T) {
	e := newTestEngine(t)
	err := runLine(e, "scoreboard players operation @s a *= b c d")
	if err == nil {
		t.Fatal("expected an error for an operation with one argument too many")
	}
	if !strings.Contains(err.Error(), "Usage: /scoreboard players operation") {
		t.Errorf("error = %q, want the usage of the subcommand", err)
	}
}

// TestExecuteDetectNeedsEnoughArguments covers the detect of `/execute`.
//
// 1.12.2 only takes the detect branch when the command has enough arguments
// for it (`"detect".equals(astring[4]) && astring.length > 10`). A line
// whose detect is not complete is not a detect at all: the rest of the
// line, "detect" included, is the command to run -- and "detect" is not a
// command.
//
// (Measured on a real 1.12.2 server: `execute Steve ~ ~ ~ detect ~ ~ ~`
// answers "All invocations failed: 'detect ~ ~ ~'".)
func TestExecuteDetectNeedsEnoughArguments(t *testing.T) {
	e := newTestEngine(t)
	addPlayer(t, e, "Steve")
	err := runLine(e, "execute Steve ~ ~ ~ detect ~ ~ ~")
	if err == nil {
		t.Fatal("expected an error for a detect without arguments")
	}
	if !strings.Contains(err.Error(), "All invocations failed: 'detect ~ ~ ~'") {
		t.Errorf("error = %q, want the summary of the execute", err)
	}
	// With a command after the detect, the detect is a detect -- and it
	// fails, because there is no stone where it looks. The failure belongs
	// to the entity, so it is collected instead of returned.
	if _, err := e.ExecuteLine("execute Steve ~ ~ ~ detect ~ ~ ~ stone 0 say hi", 1,
		&exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}); err != nil {
		t.Fatalf("ExecuteLine: %v", err)
	}
	errs := e.TakeErrors()
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Error(), "Failed to execute 'detect' as Steve") {
		t.Errorf("error = %q, want the failure of the detect", errs[0])
	}
}

// TestLintMeasuresRelativeCoordinatesFromTheGround covers the base the
// linter uses for a relative coordinate.
//
// 1.12.2 measures a relative coordinate against the position of whoever
// runs the function, and the linter does not know that. It measures from
// the ground instead: `~-1` is one block below the ground (y=4 → 3) and not
// below the world, so the linter does not report it.
func TestLintMeasuresRelativeCoordinatesFromTheGround(t *testing.T) {
	root := t.TempDir()
	writeFunction(t, root, "demo", "rel", "blockdata ~ ~-1 ~ {id:\"minecraft:mob_spawner\"}\n")
	e := newTestEngine(t)
	e.SetRoot(root)
	if errs := e.LoadFunctions(); len(errs) > 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	if errs := e.LintFunctions(); len(errs) > 0 {
		t.Errorf("the linter reported %v, want nothing", errs)
	}
}

// TestLintReportsARelativeHeightThatIsOutOfRange covers the other side of
// the same rule: a relative height that is out of range even from the
// ground is a mistake.
func TestLintReportsARelativeHeightThatIsOutOfRange(t *testing.T) {
	root := t.TempDir()
	writeFunction(t, root, "demo", "rel", "blockdata ~ ~-300 ~ {id:\"minecraft:mob_spawner\"}\n")
	e := newTestEngine(t)
	e.SetRoot(root)
	if errs := e.LoadFunctions(); len(errs) > 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	errs := e.LintFunctions()
	if len(errs) == 0 {
		t.Fatal("expected the linter to report the height")
	}
	if !strings.Contains(errs[0].Error(), "too small") {
		t.Errorf("error = %q, want a height that is too small", errs[0])
	}
}

// TestPlaySoundChecksTheSource covers the source of `/playsound`.
//
// 1.12.2 looks the source up in a map of the sound categories
// (`SoundCategory.a`), so only the ten names are accepted -- and the
// lookup is **case sensitive**.
//
// (Measured on a real 1.12.2 server: `playsound minecraft:entity.wolf.growl
// mast` answers "Source mast is unknown", and "MASTER" answers "Source
// MASTER is unknown".)
func TestPlaySoundChecksTheSource(t *testing.T) {
	e := newTestEngine(t)
	addPlayer(t, e, "Steve")
	err := runLine(e, "playsound minecraft:entity.wolf.growl mast Steve")
	if err == nil {
		t.Fatal("expected an error for an unknown source")
	}
	if !strings.Contains(err.Error(), "Source mast is unknown") {
		t.Errorf("error = %q, want an unknown source", err)
	}
	// The names are matched exactly, so "MASTER" is not "master".
	err = runLine(e, "playsound minecraft:entity.wolf.growl MASTER Steve")
	if err == nil {
		t.Fatal("expected an error for a source in the wrong case")
	}
	if !strings.Contains(err.Error(), "Source MASTER is unknown") {
		t.Errorf("error = %q, want an unknown source", err)
	}
	// A known source works.
	if err := runLine(e, "playsound minecraft:entity.wolf.growl master Steve"); err != nil {
		t.Fatalf("playsound: %v", err)
	}
}
