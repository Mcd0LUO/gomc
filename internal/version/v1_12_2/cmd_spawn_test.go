package v1_12_2

import (
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// errMessage returns the message of an error.
//
// A parse error carries the position it happened at, and that position is
// gomc's own addition: the message is what the server would say.
func errMessage(err error) string {
	if pe, ok := err.(*parser.Error); ok {
		return pe.Msg
	}
	return err.Error()
}

// addPlayerAt puts a player with the given name at the given position.
func addPlayerAt(t *testing.T, e *exec.Engine, name string, pos world.Vec) {
	t.Helper()
	p, ok := e.World.EntityTypes.NewEntity("player", "p1")
	if !ok {
		t.Fatalf("the player entity type is missing")
	}
	p.PlayerName = name
	p.Pos = pos
	e.World.AddEntity(p)
}

// TestSpawnPointArgumentLayout covers the argument layout of `/spawnpoint`.
//
// 1.12.2's `CommandSpawnpoint.execute` counts the arguments before it looks
// at any of them:
//
//	if (args.length > 1 && args.length < 4) throw new ExceptionUsage(...);
//	player = args.length > 0 ? getPlayer(args[0]) : the executing player;
//	pos    = args.length > 3 ? parseBlockPos(args, 1, true) : player.getChunkCoordinates();
//
// So two or three arguments are a usage mistake -- the position cannot be
// given without the player -- and without a position the spawn point goes
// where the player stands.
//
// (Measured on a real 1.12.2 server: `spawnpoint` and `spawnpoint 1 2 3`
// answer "Usage: /spawnpoint [player] [<x> <y> <z>]", `spawnpoint Steve`
// answers "Set Steve's spawn point to (7, 4, 4)" for a player standing at
// (7.5, 4.0, 4.5), and `spawnpoint Steve 1 2 3 4` answers "Set Steve's
// spawn point to (1, 2, 3)" -- the fourth argument is ignored.)
func TestSpawnPointArgumentLayout(t *testing.T) {
	e := newTestEngine(t)
	addPlayerAt(t, e, "Steve", world.Vec{X: 7.5, Y: 4, Z: 4.5})

	// Two or three arguments are a usage mistake.
	for _, line := range []string{"spawnpoint Steve 1 2", "spawnpoint 1 2 3"} {
		err := runLine(e, line)
		if err == nil {
			t.Errorf("%q: no error, want a usage error", line)
			continue
		}
		if want := "Usage: /spawnpoint [player] [<x> <y> <z>]"; errMessage(err) != want {
			t.Errorf("%q: %v, want %q", line, err, want)
		}
	}

	// Without a position the spawn point goes where the player stands.
	if err := runLine(e, "spawnpoint Steve"); err != nil {
		t.Fatalf("spawnpoint Steve: %v", err)
	}
	if got, want := e.World.PlayerSpawns["Steve"], (world.Pos{X: 7, Y: 4, Z: 4}); got != want {
		t.Errorf("the spawn point of Steve is %v, want %v", got, want)
	}

	// With a position it goes there, and the extra argument is ignored.
	if err := runLine(e, "spawnpoint Steve 1 2 3 4"); err != nil {
		t.Fatalf("spawnpoint Steve 1 2 3 4: %v", err)
	}
	if got, want := e.World.PlayerSpawns["Steve"], (world.Pos{X: 1, Y: 2, Z: 3}); got != want {
		t.Errorf("the spawn point of Steve is %v, want %v", got, want)
	}
}

// TestSpawnPointNeedsAPlayer covers the two commands that need a player.
//
// 1.12.2 takes the position of the player who runs the command
// (`CommandAbstract.a(listener)`), and that helper refuses a source that is
// not a player with "You must specify which player you wish to perform this
// action on."
func TestSpawnPointNeedsAPlayer(t *testing.T) {
	e := newTestEngine(t)
	for _, line := range []string{"spawnpoint", "setworldspawn"} {
		err := runLine(e, line)
		if err == nil {
			t.Errorf("%q: no error, want the refusal of the player", line)
			continue
		}
		if want := "You must specify which player you wish to perform this action on."; errMessage(err) != want {
			t.Errorf("%q: %v, want %q", line, err, want)
		}
	}
}

// TestSetWorldSpawnArgumentLayout covers the argument layout of
// `/setworldspawn`.
//
// 1.12.2's `CommandSetWorldSpawn.execute` takes no argument at all (the
// spawn point becomes the block the command runs at), three arguments (the
// position), and refuses anything else with a usage error.
func TestSetWorldSpawnArgumentLayout(t *testing.T) {
	e := newTestEngine(t)

	// One or two arguments are a usage mistake.
	for _, line := range []string{"setworldspawn 1", "setworldspawn 1 2"} {
		err := runLine(e, line)
		if err == nil {
			t.Errorf("%q: no error, want a usage error", line)
			continue
		}
		if want := "Usage: /setworldspawn [<x> <y> <z>]"; errMessage(err) != want {
			t.Errorf("%q: %v, want %q", line, err, want)
		}
	}

	// Three arguments give the position.
	if err := runLine(e, "setworldspawn 1 2 3"); err != nil {
		t.Fatalf("setworldspawn 1 2 3: %v", err)
	}
	if got, want := e.World.Spawn, (world.Pos{X: 1, Y: 2, Z: 3}); got != want {
		t.Errorf("the world spawn is %v, want %v", got, want)
	}
}

// TestSpawnPointNamesThePlayer covers the message of `/spawnpoint`.
//
// The message names the player whose spawn point was set, not the player
// who ran the command.
func TestSpawnPointNamesThePlayer(t *testing.T) {
	e := newTestEngine(t)
	addPlayerAt(t, e, "Steve", world.Vec{X: 7.5, Y: 4, Z: 4.5})
	feedback, err := runLineFeedback(e, "spawnpoint Steve 1 2 3")
	if err != nil {
		t.Fatalf("spawnpoint Steve 1 2 3: %v", err)
	}
	if len(feedback) != 1 {
		t.Fatalf("the command printed %v, want one line", feedback)
	}
	if want := "Set Steve's spawn point to (1, 2, 3)"; feedback[0] != want {
		t.Errorf("the command printed %q, want %q", feedback[0], want)
	}
}

// TestUsageLinesComeFromTheLanguageFile checks that the usage line of every
// command is one the language file of the 1.12.2 client has.
//
// The server answers a usage mistake with that line, so a command that
// shows a different one tells the reader the wrong spelling.
func TestUsageLinesComeFromTheLanguageFile(t *testing.T) {
	known := map[string]bool{}
	for _, u := range commandUsages {
		known[u] = true
	}
	e := newTestEngine(t)
	for _, name := range e.Registry.Names() {
		h, ok := e.Registry.Get(name)
		if !ok {
			t.Fatalf("the registry does not have %q", name)
		}
		usage := h.Usage()
		if !known[usage] {
			t.Errorf("the usage line of %q is %q, want one of the language file", name, usage)
		}
	}
}
