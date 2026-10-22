package v1_12_2

import (
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// addNamedPlayer puts a player with its own internal id into the world.
//
// `addPlayerAt` always uses the id "p1", so it can only make one player.
func addNamedPlayer(t *testing.T, e *exec.Engine, id, name string, pos world.Vec) {
	t.Helper()
	p, ok := e.World.EntityTypes.NewEntity("player", id)
	if !ok {
		t.Fatalf("the player entity type is missing")
	}
	p.PlayerName = name
	p.Pos = pos
	e.World.AddEntity(p)
}

// TestTellrawRendersTheText covers the two shapes a message can take.
//
// An object is a component with its `extra` siblings after it, and an array
// is a component whose first element is the root and whose other elements
// are the siblings.
func TestTellrawRendersTheText(t *testing.T) {
	for _, tc := range []struct {
		name string
		json string
		want string
	}{
		{"text and extra", `{"text":"a","extra":[{"text":"b"},{"text":"c"}]}`, "abc"},
		{"an array", `["a",{"text":"b"}]`, "ab"},
		{"a JSON with spaces", `{"text":"hello there","color":"red"}`, "hello there"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newTestEngine(t)
			addPlayerAt(t, e, "Steve", world.Vec{X: 1, Y: 4, Z: 1})
			feedback, err := runLineFeedback(e, "tellraw Steve "+tc.json)
			if err != nil {
				t.Fatalf("tellraw: %v", err)
			}
			if want := "[Server] " + tc.want; len(feedback) != 1 || feedback[0] != want {
				t.Fatalf("feedback = %v, want %q", feedback, want)
			}
		})
	}
}

// TestTellrawRendersAScore covers the "score" component.
//
// The name is looked up as a holder of the scoreboard; a holder without an
// entry renders as nothing at all.
func TestTellrawRendersAScore(t *testing.T) {
	e := newTestEngine(t)
	addPlayerAt(t, e, "Steve", world.Vec{X: 1, Y: 4, Z: 1})
	if err := runLine(e, "scoreboard objectives add t dummy"); err != nil {
		t.Fatalf("objectives add: %v", err)
	}
	if err := runLine(e, "scoreboard players set Steve t 42"); err != nil {
		t.Fatalf("players set: %v", err)
	}
	feedback, err := runLineFeedback(e, `tellraw Steve {"text":"score ","extra":[{"score":{"name":"Steve","objective":"t"}}]}`)
	if err != nil {
		t.Fatalf("tellraw: %v", err)
	}
	if len(feedback) != 1 || feedback[0] != "[Server] score 42" {
		t.Fatalf("feedback = %v, want [Server] score 42", feedback)
	}
	// A holder with no entry renders as the empty string.
	feedback, err = runLineFeedback(e, `tellraw Steve {"score":{"name":"Nobody","objective":"t"}}`)
	if err != nil {
		t.Fatalf("tellraw: %v", err)
	}
	if len(feedback) != 1 || feedback[0] != "[Server] " {
		t.Fatalf("feedback = %v, want an empty message", feedback)
	}
}

// TestTellrawScoreStarIsTheRecipient pins down what `name:"*"` means.
//
// `ChatComponentUtils.filterForDisplay` replaces a name of "*" with the name
// of the player the message goes to. It is **not** a sum over the objective
// (that is a 1.13 feature), and 1.12.2 does not read the objective's
// display name either.
func TestTellrawScoreStarIsTheRecipient(t *testing.T) {
	e := newTestEngine(t)
	addNamedPlayer(t, e, "p1", "Steve", world.Vec{X: 1, Y: 4, Z: 1})
	addNamedPlayer(t, e, "p2", "Alex", world.Vec{X: 3, Y: 4, Z: 1})
	if err := runLine(e, "scoreboard objectives add t dummy"); err != nil {
		t.Fatalf("objectives add: %v", err)
	}
	if err := runLine(e, "scoreboard players set Steve t 5"); err != nil {
		t.Fatalf("players set: %v", err)
	}
	if err := runLine(e, "scoreboard players set Alex t 9"); err != nil {
		t.Fatalf("players set: %v", err)
	}
	feedback, err := runLineFeedback(e, `tellraw Steve {"score":{"name":"*","objective":"t"}}`)
	if err != nil {
		t.Fatalf("tellraw: %v", err)
	}
	if len(feedback) != 1 || feedback[0] != "[Server] 5" {
		t.Fatalf("feedback = %v, want Steve's 5", feedback)
	}
	feedback, err = runLineFeedback(e, `tellraw Alex {"score":{"name":"*","objective":"t"}}`)
	if err != nil {
		t.Fatalf("tellraw: %v", err)
	}
	if len(feedback) != 1 || feedback[0] != "[Server] 9" {
		t.Fatalf("feedback = %v, want Alex's 9", feedback)
	}
}

// TestTellrawScoreNameSelector covers a selector as the score holder.
//
// `filterForDisplay` resolves it and requires exactly one match; a player
// gives its name, any other entity its UUID.
func TestTellrawScoreNameSelector(t *testing.T) {
	e := newTestEngine(t)
	addPlayerAt(t, e, "Steve", world.Vec{X: 1, Y: 4, Z: 1})
	if err := runLine(e, "scoreboard objectives add t dummy"); err != nil {
		t.Fatalf("objectives add: %v", err)
	}
	if err := runLine(e, "scoreboard players set Steve t 7"); err != nil {
		t.Fatalf("players set: %v", err)
	}
	feedback, err := runLineFeedback(e, `tellraw Steve {"score":{"name":"@p","objective":"t"}}`)
	if err != nil {
		t.Fatalf("tellraw: %v", err)
	}
	if len(feedback) != 1 || feedback[0] != "[Server] 7" {
		t.Fatalf("feedback = %v, want 7", feedback)
	}
}

// TestTellrawSelectorComponent covers the "selector" component.
//
// The selector is expanded to the display names of the entities it matches,
// joined the way 1.12.2 joins a list of names.
func TestTellrawSelectorComponent(t *testing.T) {
	e := newTestEngine(t)
	addNamedPlayer(t, e, "p1", "Steve", world.Vec{X: 1, Y: 4, Z: 1})
	addNamedPlayer(t, e, "p2", "Alex", world.Vec{X: 3, Y: 4, Z: 1})
	feedback, err := runLineFeedback(e, `tellraw Steve {"selector":"@a"}`)
	if err != nil {
		t.Fatalf("tellraw: %v", err)
	}
	if len(feedback) != 1 || feedback[0] != "[Server] Steve and Alex" {
		t.Fatalf("feedback = %v, want both names", feedback)
	}
}

// TestTellrawTranslate covers the "translate" component.
//
// The key is looked up in the language table and the arguments are formatted
// into it; a key the table does not know renders as the key itself.
func TestTellrawTranslate(t *testing.T) {
	e := newTestEngine(t)
	addPlayerAt(t, e, "Steve", world.Vec{X: 1, Y: 4, Z: 1})
	feedback, err := runLineFeedback(e, `tellraw Steve {"translate":"chat.type.text","with":[{"text":"Steve"},{"text":"hi"}]}`)
	if err != nil {
		t.Fatalf("tellraw: %v", err)
	}
	if len(feedback) != 1 || feedback[0] != "[Server] <Steve> hi" {
		t.Fatalf("feedback = %v, want <Steve> hi", feedback)
	}
	feedback, err = runLineFeedback(e, `tellraw Steve {"translate":"no.such.key"}`)
	if err != nil {
		t.Fatalf("tellraw: %v", err)
	}
	if len(feedback) != 1 || feedback[0] != "[Server] no.such.key" {
		t.Fatalf("feedback = %v, want the key itself", feedback)
	}
}

// TestTellrawClickAndHoverEvents covers the events.
//
// 1.12.2 builds a `ChatClickable`/`ChatHoverable` out of them while it
// deserializes the message, and it **does not check the action**: a name the
// server does not know is accepted and the message goes out. (Measured on a
// real 1.12.2 server: `{"action":"bogus"}` succeeds, and so does a `title`
// with an unknown hover action.) gomc has no client to click on, so the
// events only have to be accepted, and a wrong shape has to be rejected.
func TestTellrawClickAndHoverEvents(t *testing.T) {
	e := newTestEngine(t)
	addPlayerAt(t, e, "Steve", world.Vec{X: 1, Y: 4, Z: 1})
	ok := `tellraw Steve {"text":"menu","clickEvent":{"action":"run_command","value":"/help"},"hoverEvent":{"action":"show_text","value":{"text":"tip"}}}`
	if err := runLine(e, ok); err != nil {
		t.Fatalf("a valid click and hover event failed: %v", err)
	}
	unknown := `tellraw Steve {"text":"menu","clickEvent":{"action":"explode","value":"x"},"hoverEvent":{"action":"show_nothing","value":{"text":"x"}}}`
	if err := runLine(e, unknown); err != nil {
		t.Fatalf("an unknown action must be accepted: %v", err)
	}
	err := runLine(e, `tellraw Steve {"text":"menu","clickEvent":"not an object"}`)
	if err == nil || !strings.Contains(err.Error(), "Invalid json") {
		t.Fatalf("a clickEvent that is not an object = %v, want a JSON error", err)
	}
}

// TestTitleChecksTheComponent covers the component of `/title`.
//
// 1.12.2 deserializes the title message into a component, so a message it
// cannot build fails the command. An action it does not know is **not** a
// failure: it is accepted (measured on a real server).
func TestTitleChecksTheComponent(t *testing.T) {
	e := newTestEngine(t)
	addPlayerAt(t, e, "Steve", world.Vec{X: 1, Y: 4, Z: 1})
	if err := runLine(e, `title Steve title {"text":"hi"}`); err != nil {
		t.Fatalf("a valid title failed: %v", err)
	}
	if err := runLine(e, `title Steve title {"text":"hi","hoverEvent":{"action":"bogus","value":"x"}}`); err != nil {
		t.Fatalf("an unknown hover action must be accepted: %v", err)
	}
	err := runLine(e, `title Steve title {"text":"hi","hoverEvent":"not an object"}`)
	if err == nil || !strings.Contains(err.Error(), "Invalid json") {
		t.Fatalf("a hoverEvent that is not an object = %v, want a JSON error", err)
	}
}
