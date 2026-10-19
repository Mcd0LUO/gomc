package v1_12_2

import (
	"strings"
	"testing"

	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/nbt"
	"github.com/mcsim/gomc/internal/world"
)

// runLine runs a command line and returns the error, if any.
func runLine(e *exec.Engine, line string) error {
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	_, err := e.ExecuteLine(line, 1, src)
	return err
}

// runLineFeedback runs a command line and returns the messages it printed.
func runLineFeedback(e *exec.Engine, line string) ([]string, error) {
	before := len(e.Feedback)
	if err := runLine(e, line); err != nil {
		return nil, err
	}
	return e.Feedback[before:], nil
}

// addPlayer puts a player with the given name into the world.
func addPlayer(t *testing.T, e *exec.Engine, name string) {
	t.Helper()
	p, ok := e.World.EntityTypes.NewEntity("player", "p1")
	if !ok {
		t.Fatalf("the player entity type is missing")
	}
	p.PlayerName = name
	e.World.AddEntity(p)
}

func TestWeatherCommand(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "weather rain 60"); err != nil {
		t.Fatalf("weather rain: %v", err)
	}
	if e.World.Weather.Kind != world.WeatherRain {
		t.Errorf("weather = %q, want rain", e.World.Weather.Kind)
	}
	if e.World.Weather.Duration != 60*20 {
		t.Errorf("duration = %d ticks, want %d", e.World.Weather.Duration, 60*20)
	}
	if !e.World.Weather.Storm {
		t.Error("it should be raining")
	}

	if err := runLine(e, "weather thunder"); err != nil {
		t.Fatalf("weather thunder: %v", err)
	}
	if !e.World.Weather.Thundering {
		t.Error("it should be thundering")
	}

	if err := runLine(e, "weather clear"); err != nil {
		t.Fatalf("weather clear: %v", err)
	}
	if e.World.Weather.Kind != world.WeatherClear || e.World.Weather.Storm {
		t.Errorf("weather = %+v, want clear", e.World.Weather)
	}
}

func TestWeatherRejectsUnknownKind(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "weather snow"); err == nil {
		t.Fatal("expected an error for an unknown weather")
	}
}

func TestToggleDownfallTogglesTheStorm(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "toggledownfall"); err != nil {
		t.Fatalf("toggledownfall: %v", err)
	}
	if !e.World.Weather.Storm {
		t.Error("the storm should have started")
	}
	if err := runLine(e, "toggledownfall"); err != nil {
		t.Fatalf("toggledownfall: %v", err)
	}
	if e.World.Weather.Storm {
		t.Error("the storm should have stopped")
	}
}

func TestDifficultyCommand(t *testing.T) {
	e := newTestEngine(t)
	for _, line := range []string{"difficulty hard", "difficulty h", "difficulty 3"} {
		if err := runLine(e, line); err != nil {
			t.Fatalf("%q: %v", line, err)
		}
		if e.World.Difficulty != world.DifficultyHard {
			t.Errorf("%q set the difficulty to %q, want hard", line, e.World.Difficulty)
		}
	}
	if err := runLine(e, "difficulty impossible"); err == nil {
		t.Fatal("expected an error for an unknown difficulty")
	}
}

func TestSetWorldSpawnAndSpawnPoint(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "setworldspawn 10 20 30"); err != nil {
		t.Fatalf("setworldspawn: %v", err)
	}
	want := world.Pos{X: 10, Y: 20, Z: 30}
	if e.World.Spawn != want {
		t.Errorf("spawn = %v, want %v", e.World.Spawn, want)
	}
}

func TestWorldBorderCommand(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "worldborder set 100"); err != nil {
		t.Fatalf("worldborder set: %v", err)
	}
	if e.World.Border.Size != 100 {
		t.Errorf("size = %v, want 100", e.World.Border.Size)
	}
	if err := runLine(e, "worldborder add 50"); err != nil {
		t.Fatalf("worldborder add: %v", err)
	}
	if e.World.Border.Size != 150 {
		t.Errorf("size = %v, want 150", e.World.Border.Size)
	}
	if err := runLine(e, "worldborder center 5 6"); err != nil {
		t.Fatalf("worldborder center: %v", err)
	}
	// 1.12.2 centres the border on the middle of a block, so a whole
	// coordinate is moved by half a block.
	if e.World.Border.Center.X != 5.5 || e.World.Border.Center.Z != 6.5 {
		t.Errorf("center = %+v, want (5.5, 6.5)", e.World.Border.Center)
	}
	if err := runLine(e, "worldborder warning time 30"); err != nil {
		t.Fatalf("worldborder warning time: %v", err)
	}
	if e.World.Border.WarningTime != 30 {
		t.Errorf("warning time = %d, want 30", e.World.Border.WarningTime)
	}
}

// TestTagCommandIsNotAStandaloneCommand pins down a 1.12.2 detail.
//
// The real server rejects `/tag` with "Unknown or invalid command 'tag'",
// so the tags of an entity are managed with `scoreboard players tag`.
func TestTagCommandIsNotAStandaloneCommand(t *testing.T) {
	e := newTestEngine(t)
	if _, ok := e.Registry.Get("tag"); ok {
		t.Error("/tag does not exist in 1.12.2")
	}
	if _, ok := e.Registry.Get("scoreboard"); !ok {
		t.Error("/scoreboard is missing")
	}
}

// TestExecuteRunsForEveryEntity covers the 1.12.2 expansion.
//
// `execute` runs the sub-command once per matching entity, in reverse
// order, and the target entity becomes the executor of the sub-command —
// so the output is prefixed with the entity's name.
func TestExecuteRunsForEveryEntity(t *testing.T) {
	e := newTestEngine(t)
	for _, id := range []string{"a1", "a2"} {
		s, _ := e.World.EntityTypes.NewEntity("armor_stand", id)
		s.Pos = world.Vec{X: 1, Y: 2, Z: 3}
		e.World.AddEntity(s)
	}
	if err := runLine(e, "execute @e[type=armor_stand] ~ ~ ~ say HELLO"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	var count int
	for _, msg := range e.Feedback {
		if strings.Contains(msg, "HELLO") {
			count++
			// The prefix is the executing entity's name, not "Server".
			if strings.HasPrefix(msg, "[Server]") {
				t.Errorf("the message = %q, want it to be prefixed with the entity", msg)
			}
		}
	}
	if count != 2 {
		t.Errorf("the sub-command ran %d times, want once per entity", count)
	}
}

func TestScoreboardPlayersTag(t *testing.T) {
	e := newTestEngine(t)
	z, _ := e.World.EntityTypes.NewEntity("zombie", "z1")
	e.World.AddEntity(z)
	if err := runLine(e, "scoreboard players tag @e[type=zombie] add marked"); err != nil {
		t.Fatalf("scoreboard players tag: %v", err)
	}
	if !z.HasTag("marked") {
		t.Error("the tag was not added")
	}
	if err := runLine(e, "scoreboard players tag @e[tag=marked] list"); err != nil {
		t.Fatalf("scoreboard players tag list: %v", err)
	}
	if err := runLine(e, "scoreboard players tag @e[type=zombie] remove marked"); err != nil {
		t.Fatalf("scoreboard players tag remove: %v", err)
	}
	if z.HasTag("marked") {
		t.Error("the tag was not removed")
	}
}

// TestTagsInNBTAreTheScoreboardTags covers the 1.12.2 identity between the
// NBT `Tags` list and the scoreboard tags of an entity.
//
// `Entity.getScoreboardTags` reads the `Tags` list out of the entity's NBT,
// so summoning an entity with `{Tags:["w"]}` makes it visible to
// `@e[tag=w]` right away.
func TestTagsInNBTAreTheScoreboardTags(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon armor_stand 0 0 0 {Tags:[\"w\"]}"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	entities := e.World.EntityList()
	if len(entities) != 1 {
		t.Fatalf("summoned %d entities, want 1", len(entities))
	}
	if !entities[0].HasTag("w") {
		t.Error("the NBT Tags list did not become a scoreboard tag")
	}
	// And the other way round: the scoreboard command writes the same list.
	if err := runLine(e, "scoreboard players tag @e[type=armor_stand] add other"); err != nil {
		t.Fatalf("scoreboard players tag: %v", err)
	}
	list, ok := entities[0].NBT.Get("Tags").(*nbt.List)
	if !ok {
		t.Fatal("the entity has no Tags list in its NBT")
	}
	if list.Len() != 2 {
		t.Errorf("the Tags list holds %d entries, want 2", list.Len())
	}
}

// TestScoreboardPlayersFeedback pins down the wording of the scoreboard
// messages.
//
// 1.12.2 confirms set, add and remove with the same sentence, and it
// reports the score *after* the change. A reset without an objective says
// "scores" in the plural, and the entry of `players list` shows the
// display name of the objective and its name, never its criterion.
func TestScoreboardPlayersFeedback(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "scoreboard objectives add counter dummy"); err != nil {
		t.Fatalf("scoreboard objectives add: %v", err)
	}
	got, err := runLineFeedback(e, "scoreboard players set Steve counter 10")
	if err != nil {
		t.Fatalf("scoreboard players set: %v", err)
	}
	want := "Set score of counter for player Steve to 10"
	if len(got) != 1 || got[0] != want {
		t.Errorf("set said %q, want %q", got, want)
	}
	got, err = runLineFeedback(e, "scoreboard players add Steve counter 5")
	if err != nil {
		t.Fatalf("scoreboard players add: %v", err)
	}
	want = "Set score of counter for player Steve to 15"
	if len(got) != 1 || got[0] != want {
		t.Errorf("add said %q, want %q", got, want)
	}
	got, err = runLineFeedback(e, "scoreboard players remove Steve counter 5")
	if err != nil {
		t.Fatalf("scoreboard players remove: %v", err)
	}
	want = "Set score of counter for player Steve to 10"
	if len(got) != 1 || got[0] != want {
		t.Errorf("remove said %q, want %q", got, want)
	}
	got, err = runLineFeedback(e, "scoreboard players reset Steve counter")
	if err != nil {
		t.Fatalf("scoreboard players reset: %v", err)
	}
	want = "Reset score counter of player Steve"
	if len(got) != 1 || got[0] != want {
		t.Errorf("reset said %q, want %q", got, want)
	}
	got, err = runLineFeedback(e, "scoreboard players reset Steve")
	if err != nil {
		t.Fatalf("scoreboard players reset: %v", err)
	}
	want = "Reset scores of player Steve"
	if len(got) != 1 || got[0] != want {
		t.Errorf("reset said %q, want %q", got, want)
	}
	if err := runLine(e, "scoreboard players set Steve counter 20"); err != nil {
		t.Fatalf("scoreboard players set: %v", err)
	}
	got, err = runLineFeedback(e, "scoreboard players list Steve")
	if err != nil {
		t.Fatalf("scoreboard players list: %v", err)
	}
	want = "- counter: 20 (counter)"
	if len(got) != 2 || got[1] != want {
		t.Errorf("list said %q, want %q", got, want)
	}
}

// TestScoreboardTeamsFeedback pins down the wording of the team messages.
func TestScoreboardTeamsFeedback(t *testing.T) {
	e := newTestEngine(t)
	got, err := runLineFeedback(e, "scoreboard teams add red")
	if err != nil {
		t.Fatalf("scoreboard teams add: %v", err)
	}
	want := "Added new team 'red' successfully"
	if len(got) != 1 || got[0] != want {
		t.Errorf("add said %q, want %q", got, want)
	}
	got, err = runLineFeedback(e, "scoreboard teams join red Steve")
	if err != nil {
		t.Fatalf("scoreboard teams join: %v", err)
	}
	want = "Added 1 player(s) to team red: Steve"
	if len(got) != 1 || got[0] != want {
		t.Errorf("join said %q, want %q", got, want)
	}
	got, err = runLineFeedback(e, "scoreboard teams option red friendlyFire true")
	if err != nil {
		t.Fatalf("scoreboard teams option: %v", err)
	}
	want = "Set option friendlyfire for team red to true"
	if len(got) != 1 || got[0] != want {
		t.Errorf("option said %q, want %q", got, want)
	}
	got, err = runLineFeedback(e, "scoreboard teams list")
	if err != nil {
		t.Fatalf("scoreboard teams list: %v", err)
	}
	want = "- red: 'red' has 1 players"
	if len(got) != 2 || got[1] != want {
		t.Errorf("list said %q, want %q", got, want)
	}
	// 1.12.2's `teams leave` takes only the player: it takes them out of
	// whichever team they are on.
	got, err = runLineFeedback(e, "scoreboard teams leave Steve")
	if err != nil {
		t.Fatalf("scoreboard teams leave: %v", err)
	}
	want = "Removed 1 player(s) from their teams: Steve"
	if len(got) != 1 || got[0] != want {
		t.Errorf("leave said %q, want %q", got, want)
	}
	// The team is empty now, and emptying an empty team is an error.
	_, err = runLineFeedback(e, "scoreboard teams empty red")
	if err == nil {
		t.Fatal("expected an error for an already empty team")
	}
	want = "Team red is already empty, cannot remove nonexistant players"
	if err.Error() != want {
		t.Errorf("empty said %q, want %q", err.Error(), want)
	}
	got, err = runLineFeedback(e, "scoreboard teams remove red")
	if err != nil {
		t.Fatalf("scoreboard teams remove: %v", err)
	}
	want = "Removed team red"
	if len(got) != 1 || got[0] != want {
		t.Errorf("remove said %q, want %q", got, want)
	}
}

// TestGiveFeedback pins down the wording of `/give`.
func TestGiveFeedback(t *testing.T) {
	e := newTestEngine(t)
	addPlayer(t, e, "Steve")
	got, err := runLineFeedback(e, "give Steve stone 5")
	if err != nil {
		t.Fatalf("give: %v", err)
	}
	want := "Given stone * 5 to Steve"
	if len(got) != 1 || got[0] != want {
		t.Errorf("give said %q, want %q", got, want)
	}
}

// TestSpawnPointFeedback pins down the wording of `/spawnpoint`.
//
// The message names the player whose spawn point was set, not the player
// who ran the command, and a player that is not online is an error.
func TestSpawnPointFeedback(t *testing.T) {
	e := newTestEngine(t)
	addPlayer(t, e, "Steve")
	got, err := runLineFeedback(e, "spawnpoint Steve 1 2 3")
	if err != nil {
		t.Fatalf("spawnpoint: %v", err)
	}
	want := "Set Steve's spawn point to (1, 2, 3)"
	if len(got) != 1 || got[0] != want {
		t.Errorf("spawnpoint said %q, want %q", got, want)
	}
	err = runLine(e, "spawnpoint Nobody")
	if err == nil {
		t.Fatal("expected an error for a player that is not online")
	}
	want = "Player 'Nobody' cannot be found"
	if err.Error() != want {
		t.Errorf("spawnpoint said %q, want %q", err.Error(), want)
	}
}

// TestGameruleFeedback pins down the wording of `/gamerule`.
func TestGameruleFeedback(t *testing.T) {
	e := newTestEngine(t)
	got, err := runLineFeedback(e, "gamerule")
	if err != nil {
		t.Fatalf("gamerule: %v", err)
	}
	// The rules are listed in the alphabetical order of their names, and
	// 1.12.2 joins them into a single line.
	if len(got) != 1 {
		t.Fatalf("gamerule printed %d lines, want one", len(got))
	}
	if !strings.HasPrefix(got[0], "announceAdvancements, commandBlockOutput, ") ||
		!strings.HasSuffix(got[0], "spawnRadius and spectatorsGenerateChunks") {
		t.Errorf("gamerule said %q", got[0])
	}
	got, err = runLineFeedback(e, "gamerule doFireTick")
	if err != nil {
		t.Fatalf("gamerule doFireTick: %v", err)
	}
	want := "doFireTick = true"
	if len(got) != 1 || got[0] != want {
		t.Errorf("gamerule said %q, want %q", got, want)
	}
	got, err = runLineFeedback(e, "gamerule doFireTick false")
	if err != nil {
		t.Fatalf("gamerule doFireTick false: %v", err)
	}
	want = "Game rule doFireTick has been updated to false"
	if len(got) != 1 || got[0] != want {
		t.Errorf("gamerule said %q, want %q", got, want)
	}
	err = runLine(e, "gamerule doFireTick yes")
	if err == nil {
		t.Fatal("expected an error for a value that is not a boolean")
	}
	want = "'yes' is not true or false"
	if err.Error() != want {
		t.Errorf("gamerule said %q, want %q", err.Error(), want)
	}
	// Querying a rule that does not exist is an error, but setting one
	// creates it, exactly like `GameRules.set` does.
	err = runLine(e, "gamerule noSuchRule")
	if err == nil {
		t.Fatal("expected an error for a rule that does not exist")
	}
	want = "No game rule called 'noSuchRule' is available"
	if err.Error() != want {
		t.Errorf("gamerule said %q, want %q", err.Error(), want)
	}
	err = runLine(e, "gamerule noSuchRule true")
	if err != nil {
		t.Fatalf("gamerule noSuchRule true: %v", err)
	}
}

// TestDifficultyFeedback pins down the wording of `/difficulty`.
//
// The message shows the translated name of the difficulty, so it starts
// with a capital letter.
func TestDifficultyFeedback(t *testing.T) {
	e := newTestEngine(t)
	got, err := runLineFeedback(e, "difficulty hard")
	if err != nil {
		t.Fatalf("difficulty: %v", err)
	}
	want := "Set game difficulty to Hard"
	if len(got) != 1 || got[0] != want {
		t.Errorf("difficulty said %q, want %q", got, want)
	}
}

// TestWorldBorderFeedback pins down the wording of `/worldborder`.
func TestWorldBorderFeedback(t *testing.T) {
	e := newTestEngine(t)
	got, err := runLineFeedback(e, "worldborder set 100")
	if err != nil {
		t.Fatalf("worldborder set: %v", err)
	}
	want := "Set world border to 100.0 blocks wide (from 60000000.0 blocks)"
	if len(got) != 1 || got[0] != want {
		t.Errorf("worldborder said %q, want %q", got, want)
	}
	got, err = runLineFeedback(e, "worldborder set 50 10")
	if err != nil {
		t.Fatalf("worldborder set: %v", err)
	}
	want = "Shrinking world border to 50.0 blocks wide (down from 100.0 blocks) over 10 seconds"
	if len(got) != 1 || got[0] != want {
		t.Errorf("worldborder said %q, want %q", got, want)
	}
	got, err = runLineFeedback(e, "worldborder add 20 5")
	if err != nil {
		t.Fatalf("worldborder add: %v", err)
	}
	want = "Growing world border to 70.0 blocks wide (up from 50.0 blocks) over 5 seconds"
	if len(got) != 1 || got[0] != want {
		t.Errorf("worldborder said %q, want %q", got, want)
	}
	got, err = runLineFeedback(e, "worldborder center 0 0")
	if err != nil {
		t.Fatalf("worldborder center: %v", err)
	}
	want = "Set world border center to 0.5, 0.5"
	if len(got) != 1 || got[0] != want {
		t.Errorf("worldborder said %q, want %q", got, want)
	}
}

// TestTargetResolutionWording pins down the three ways 1.12.2 reports a
// target that cannot be resolved.
func TestTargetResolutionWording(t *testing.T) {
	e := newTestEngine(t)
	err := runLine(e, "gamemode creative @a")
	if err == nil {
		t.Fatal("expected an error for a selector that matches nothing")
	}
	want := "Selector '@a' found nothing"
	if err.Error() != want {
		t.Errorf("the error was %q, want %q", err.Error(), want)
	}
	err = runLine(e, "gamemode creative @p")
	if err == nil {
		t.Fatal("expected an error for a selector that matches nothing")
	}
	want = "Entity '@p' cannot be found"
	if err.Error() != want {
		t.Errorf("the error was %q, want %q", err.Error(), want)
	}
	err = runLine(e, "gamemode creative Nobody")
	if err == nil {
		t.Fatal("expected an error for a player that is not online")
	}
	want = "Player 'Nobody' cannot be found"
	if err.Error() != want {
		t.Errorf("the error was %q, want %q", err.Error(), want)
	}
	// `/execute` resolves its target as an entity, so it uses the
	// entity wording and not the player one.
	err = runLine(e, "execute Nobody ~ ~ ~ say hi")
	if err == nil {
		t.Fatal("expected an error for a target that is not an entity")
	}
	want = "Entity 'Nobody' cannot be found"
	if err.Error() != want {
		t.Errorf("the error was %q, want %q", err.Error(), want)
	}
}
