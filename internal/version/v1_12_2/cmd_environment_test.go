package v1_12_2

import (
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
	if e.World.Border.Center.X != 5 || e.World.Border.Center.Z != 6 {
		t.Errorf("center = %+v, want (5, 6)", e.World.Border.Center)
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
