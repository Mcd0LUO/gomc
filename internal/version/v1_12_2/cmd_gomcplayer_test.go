package v1_12_2

import (
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// TestGomcPlayerBringsAPlayerIntoTheWorld pins down that the command puts
// a player into the world that the world can find.
//
// 1.12.2 cannot do this itself: `/summon player` answers "Unable to
// summon object", so a function that needs a player has no way to get one
// on a real server.
func TestGomcPlayerBringsAPlayerIntoTheWorld(t *testing.T) {
	e := newTestEngine(t)

	msgs, err := runLineFeedback(e, "gomcplayer Steve")
	if err != nil {
		t.Fatalf("gomcplayer: %v", err)
	}
	if len(msgs) != 1 || msgs[0] != "Successfully created the player Steve" {
		t.Errorf("the command printed %q, want the created message", msgs)
	}

	p := e.World.FindPlayer("Steve")
	if p == nil {
		t.Fatal("the player is not in the world")
	}
	if !p.IsPlayer {
		t.Error("the entity is not marked as a player")
	}
	// The scoreboard tracks a player under its name, not under a UUID.
	if got := exec.ScoreboardHolder(p); got != "Steve" {
		t.Errorf("the scoreboard holder is %q, want %q", got, "Steve")
	}
}

// TestGomcPlayerIsPickedUpByThePlayerSelectors pins down that the player
// is an ordinary player for the selectors: `@a` and `@p` find it, and it
// is the executor of a command run for it.
func TestGomcPlayerIsPickedUpByThePlayerSelectors(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "gomcplayer Steve"); err != nil {
		t.Fatalf("gomcplayer: %v", err)
	}

	// 1.12.2's execute takes the position of the executor, so it is
	// written out even when the command does not use it.
	msgs, err := runLineFeedback(e, "execute @a ~ ~ ~ say hi")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(msgs) != 1 || msgs[0] != "[Steve] hi" {
		t.Errorf("the command printed %q, want the message of the player", msgs)
	}
}

// TestGomcPlayerLetsThePlayerCommandsRun pins down the point of the
// command: the commands that only work on a player can be run against it.
//
// 1.12.2 needs a player that is online, and without this command the only
// thing that can be tested of `/advancement` and `/recipe` is their error
// messages.
func TestGomcPlayerLetsThePlayerCommandsRun(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "gomcplayer Steve"); err != nil {
		t.Fatalf("gomcplayer: %v", err)
	}

	if err := runLine(e, "advancement grant Steve everything"); err != nil {
		t.Errorf("advancement could not find the player: %v", err)
	}
	// The recipe may well be one the player already has -- 1.12.2 hands a
	// player a few recipes when it logs in -- but it must not be reported
	// as a missing player.
	if err := runLine(e, "recipe give Steve stone_slab"); err != nil {
		if strings.Contains(err.Error(), "cannot be found") {
			t.Errorf("recipe could not find the player: %v", err)
		}
	}
	if err := runLine(e, "gamemode creative Steve"); err != nil {
		t.Errorf("gamemode could not find the player: %v", err)
	}
}

// TestGomcPlayerRefusesASecondPlayerWithTheSameName pins down that two
// players cannot share a name: the world finds a player by its name, so
// the name has to be unique.
func TestGomcPlayerRefusesASecondPlayerWithTheSameName(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "gomcplayer Steve"); err != nil {
		t.Fatalf("gomcplayer: %v", err)
	}
	err := runLine(e, "gomcplayer Steve")
	if err == nil {
		t.Fatal("the second player was expected to be refused")
	}
	if got := err.Error(); got != "A player called 'Steve' already exists" {
		t.Errorf("the error was %q, want the already-exists message", got)
	}
}

// TestGomcPlayerPutsThePlayerWhereItIsTold pins down the position: the
// one that is given, or the one of the command.
func TestGomcPlayerPutsThePlayerWhereItIsTold(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "gomcplayer Steve 10 20 30"); err != nil {
		t.Fatalf("gomcplayer: %v", err)
	}
	p := e.World.FindPlayer("Steve")
	if p == nil {
		t.Fatal("the player is not in the world")
	}
	if p.Pos != (world.Vec{X: 10, Y: 20, Z: 30}) {
		t.Errorf("the player is at %v, want 10 20 30", p.Pos)
	}

	// Without a position the player is put where the command runs.
	e2 := newTestEngine(t)
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4,
		Pos: world.Vec{X: 5, Y: 6, Z: 7}}
	if _, err := e2.ExecuteLine("gomcplayer Alex", 1, src); err != nil {
		t.Fatalf("gomcplayer: %v", err)
	}
	a := e2.World.FindPlayer("Alex")
	if a == nil {
		t.Fatal("the player is not in the world")
	}
	if a.Pos != src.Pos {
		t.Errorf("the player is at %v, want the position of the command %v", a.Pos, src.Pos)
	}
}

// TestGomcPlayerIsAnExtensionCommand pins down that the command is marked
// as one gomc adds on top of the 1.12.2 set.
//
// 1.12.2 refuses a function whose command it does not know, so the check
// has to be able to tell the user that the function would not load on a
// real server.
func TestGomcPlayerIsAnExtensionCommand(t *testing.T) {
	reg := exec.NewRegistry()
	Register(reg)

	note, ok := reg.Extension("gomcplayer")
	if !ok {
		t.Fatal("gomcplayer is not marked as an extension")
	}
	if note == "" {
		t.Error("the extension has no note")
	}
	// A command of the 1.12.2 set is not an extension.
	if _, ok := reg.Extension("summon"); ok {
		t.Error("summon is marked as an extension")
	}
}
