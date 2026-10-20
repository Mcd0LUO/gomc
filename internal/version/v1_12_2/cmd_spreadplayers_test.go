package v1_12_2

import (
	"math"
	"strings"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// newSpreadEngine builds an engine that knows the spread players command.
//
// The command is registered on its own, so the tests do not depend on the
// whole command set being there.
func newSpreadEngine(t *testing.T) *exec.Engine {
	t.Helper()
	reg := exec.NewRegistry()
	registerSpreadPlayers(reg)
	registerScoreboard(reg)
	return exec.NewEngine(reg, world.NewWorld())
}

// addSpreadPlayer puts a player with the given name into the world.
func addSpreadPlayer(t *testing.T, e *exec.Engine, name string) *world.Entity {
	t.Helper()
	p, ok := e.World.EntityTypes.NewEntity("player", "p-"+name)
	if !ok {
		t.Fatalf("the player entity type is missing")
	}
	p.PlayerName = name
	p.Pos = world.Vec{X: 0, Y: 64, Z: 0}
	e.World.AddEntity(p)
	return p
}

// TestSpreadPlayersMovesThePlayers pins down what 1.12.2 does with the
// players it is given.
//
// (Measured on a real 1.12.2 server: the command answers with three
// lines -- what it is about to do, that it did it, and how far apart the
// players ended up.)
func TestSpreadPlayersMovesThePlayers(t *testing.T) {
	e := newSpreadEngine(t)
	steve := addSpreadPlayer(t, e, "Steve")
	alex := addSpreadPlayer(t, e, "Alex")

	msgs, err := runLineFeedback(e, "spreadplayers 0 0 5 10 false Steve Alex")
	if err != nil {
		t.Fatalf("spreadplayers: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("the command printed %d lines, want 3: %q", len(msgs), msgs)
	}
	if want := "Spreading 2 player(s) 10.0 blocks around 0.5, 0.5 (min 5.0 blocks apart)"; msgs[0] != want {
		t.Errorf("the first line was %q, want %q", msgs[0], want)
	}
	if want := "Successfully spread 2 player(s) around 0.5, 0.5"; msgs[1] != want {
		t.Errorf("the second line was %q, want %q", msgs[1], want)
	}
	if !strings.HasPrefix(msgs[2], "(Average distance between players is ") ||
		!strings.HasSuffix(msgs[2], " iterations)") {
		t.Errorf("the third line was %q, want the average distance", msgs[2])
	}
	// The players are put somewhere else, and the somewhere else is inside
	// the box around the centre.
	for _, p := range []*world.Entity{steve, alex} {
		if p.Pos == (world.Vec{X: 0, Y: 64, Z: 0}) {
			t.Errorf("%s was not moved", p.PlayerName)
		}
		if p.Pos.X < -10.5 || p.Pos.X > 10.5 || p.Pos.Z < -10.5 || p.Pos.Z > 10.5 {
			t.Errorf("%s ended up at %v, which is outside the box", p.PlayerName, p.Pos)
		}
	}
	// The middle of a block is where a player is put.
	if math.Floor(steve.Pos.X) != steve.Pos.X-0.5 || math.Floor(steve.Pos.Z) != steve.Pos.Z-0.5 {
		t.Errorf("%s ended up at %v, want the middle of a block", steve.PlayerName, steve.Pos)
	}
	// The two players do not end up on the same spot.
	if steve.Pos == alex.Pos {
		t.Errorf("both players ended up at %v", steve.Pos)
	}
}

// TestSpreadPlayersRespectsTeams pins down what the team flag does.
//
// With respectTeams the players of one team are spread together: they all
// get the same spot, and the number of spots is the number of teams.
//
// (Measured on a real 1.12.2 server: `spreadplayers ... true` answers
// "Successfully spread 1 team(s)" when the three players are not on a
// team, because a player without a team is a team of its own.)
func TestSpreadPlayersRespectsTeams(t *testing.T) {
	e := newSpreadEngine(t)
	steve := addSpreadPlayer(t, e, "Steve")
	alex := addSpreadPlayer(t, e, "Alex")
	bob := addSpreadPlayer(t, e, "Bob")
	if err := runLine(e, "scoreboard teams add red"); err != nil {
		t.Fatalf("scoreboard teams add: %v", err)
	}
	if err := runLine(e, "scoreboard teams join red Steve"); err != nil {
		t.Fatalf("scoreboard teams join: %v", err)
	}
	if err := runLine(e, "scoreboard teams join red Bob"); err != nil {
		t.Fatalf("scoreboard teams join: %v", err)
	}

	msgs, err := runLineFeedback(e, "spreadplayers 0 0 5 10 true Steve Alex Bob")
	if err != nil {
		t.Fatalf("spreadplayers: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("the command printed %d lines, want 3: %q", len(msgs), msgs)
	}
	if want := "Spreading 3 team(s) 10.0 blocks around 0.5, 0.5 (min 5.0 blocks apart)"; msgs[0] != want {
		t.Errorf("the first line was %q, want %q", msgs[0], want)
	}
	// Two teams: "red" and the team without a name that Alex is on.
	if want := "Successfully spread 2 team(s) around 0.5, 0.5"; msgs[1] != want {
		t.Errorf("the second line was %q, want %q", msgs[1], want)
	}
	if !strings.Contains(msgs[2], "between teams is") {
		t.Errorf("the third line was %q, want the distance between the teams", msgs[2])
	}
	// The players of one team share their spot, the other player does not.
	if steve.Pos != bob.Pos {
		t.Errorf("the players of the team are at %v and %v, want the same spot", steve.Pos, bob.Pos)
	}
	if alex.Pos == steve.Pos {
		t.Errorf("the player of the other team is at the same spot %v", alex.Pos)
	}
}

// TestSpreadPlayersWithoutTeams checks the other side of the team flag.
//
// Without respectTeams every player gets a spot of its own.
func TestSpreadPlayersWithoutTeams(t *testing.T) {
	e := newSpreadEngine(t)
	steve := addSpreadPlayer(t, e, "Steve")
	bob := addSpreadPlayer(t, e, "Bob")
	if err := runLine(e, "scoreboard teams add red"); err != nil {
		t.Fatalf("scoreboard teams add: %v", err)
	}
	if err := runLine(e, "scoreboard teams join red Steve"); err != nil {
		t.Fatalf("scoreboard teams join: %v", err)
	}
	if err := runLine(e, "scoreboard teams join red Bob"); err != nil {
		t.Fatalf("scoreboard teams join: %v", err)
	}
	msgs, err := runLineFeedback(e, "spreadplayers 0 0 5 10 false Steve Bob")
	if err != nil {
		t.Fatalf("spreadplayers: %v", err)
	}
	if want := "Successfully spread 2 player(s) around 0.5, 0.5"; msgs[1] != want {
		t.Errorf("the second line was %q, want %q", msgs[1], want)
	}
	if steve.Pos == bob.Pos {
		t.Errorf("both players ended up at %v", steve.Pos)
	}
}

// TestSpreadPlayersReportsUnknownPlayers pins down what 1.12.2 says when
// there is nobody to spread.
//
// (Measured on a real 1.12.2 server.)
func TestSpreadPlayersReportsUnknownPlayers(t *testing.T) {
	e := newSpreadEngine(t)
	err := runLine(e, "spreadplayers 0 0 5 10 false NoSuchPlayer")
	if err == nil {
		t.Fatal("expected an error for a player who is not there")
	}
	if want := "Player 'NoSuchPlayer' cannot be found"; !strings.Contains(err.Error(), want) {
		t.Errorf("the error was %q, want it to contain %q", err.Error(), want)
	}
	err = runLine(e, "spreadplayers 0 0 5 10 false @e[type=sheep]")
	if err == nil {
		t.Fatal("expected an error for a selector that matches nothing")
	}
	if want := "Selector '@e[type=sheep]' found nothing"; !strings.Contains(err.Error(), want) {
		t.Errorf("the error was %q, want it to contain %q", err.Error(), want)
	}
}

// TestSpreadPlayersReportsMistakes pins down the errors of the arguments.
//
// (Measured on a real 1.12.2 server.)
func TestSpreadPlayersReportsMistakes(t *testing.T) {
	e := newSpreadEngine(t)
	addSpreadPlayer(t, e, "Steve")
	for _, tc := range []struct {
		line string
		want string
	}{
		{
			"spreadplayers 0 0 5 10 false",
			"Usage: /spreadplayers <x> <z> <spreadDistance> <maxRange> <respectTeams true|false> <player ...>",
		},
		{"spreadplayers 0 0 5 10 maybe Steve", "'maybe' is not true or false"},
		{"spreadplayers 0 0 -1 10 false Steve", "The number you have entered (-1.00) is too small, it must be at least 0.00"},
		{"spreadplayers 0 0 5 4 false Steve", "The number you have entered (4.00) is too small, it must be at least 6.00"},
		{"spreadplayers 0 0 abc 10 false Steve", "'abc' is not a valid number"},
		{"spreadplayers abc 0 5 10 false Steve", "'abc' is not a valid number"},
	} {
		err := runLine(e, tc.line)
		if err == nil {
			t.Errorf("%q: expected an error", tc.line)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%q: the error was %q, want it to contain %q", tc.line, err.Error(), tc.want)
		}
	}
}

// TestSpreadPlayersCentreIsTheMiddleOfABlock pins down how the centre is
// read.
//
// 1.12.2 reads the centre with `CommandAbstract.b(base, text, true)`: a
// whole number names a block, and the middle of that block is the centre.
//
// (Measured on a real 1.12.2 server: `spreadplayers 100 200 ...` answers
// "around 100.5, 200.5", and `spreadplayers 100.25 200.75 ...` answers
// "around 100.25, 200.75".)
func TestSpreadPlayersCentreIsTheMiddleOfABlock(t *testing.T) {
	e := newSpreadEngine(t)
	addSpreadPlayer(t, e, "Steve")
	msgs, err := runLineFeedback(e, "spreadplayers 100 200 5 10 false Steve")
	if err != nil {
		t.Fatalf("spreadplayers: %v", err)
	}
	if want := "Spreading 1 player(s) 10.0 blocks around 100.5, 200.5 (min 5.0 blocks apart)"; msgs[0] != want {
		t.Errorf("the first line was %q, want %q", msgs[0], want)
	}
	msgs, err = runLineFeedback(e, "spreadplayers 100.25 200.75 5 10 false Steve")
	if err != nil {
		t.Fatalf("spreadplayers: %v", err)
	}
	if want := "Spreading 1 player(s) 10.0 blocks around 100.25, 200.75 (min 5.0 blocks apart)"; msgs[0] != want {
		t.Errorf("the first line was %q, want %q", msgs[0], want)
	}
}

// TestSpreadPlayersPutsThemOnTheGround pins down the height a player is
// put at.
//
// 1.12.2 looks for the first block under the spot that is not air and
// stands the player on top of it.
//
// (Measured on a real 1.12.2 server: on a flat world the block under the
// player is the grass at y=3, so the player stands at y=4.)
func TestSpreadPlayersPutsThemOnTheGround(t *testing.T) {
	e := newSpreadEngine(t)
	steve := addSpreadPlayer(t, e, "Steve")
	// A floor of stone at y=3, so the player has to stand at y=4.
	for x := -20; x <= 20; x++ {
		for z := -20; z <= 20; z++ {
			e.World.Blocks.Set(world.Pos{X: x, Y: 3, Z: z}, world.BlockState{Name: "stone"})
		}
	}
	if err := runLine(e, "spreadplayers 0 0 5 10 false Steve"); err != nil {
		t.Fatalf("spreadplayers: %v", err)
	}
	if steve.Pos.Y != 4 {
		t.Errorf("the player stands at y=%v, want y=4", steve.Pos.Y)
	}
}

// TestSpreadPlayersIsReproducible pins down that the same command always
// spreads the players the same way.
//
// 1.12.2 uses `new Random()`, which is seeded from the clock, so the
// players land somewhere different every time. gomc has to be reproducible,
// so the random source is seeded from the arguments of the command.
func TestSpreadPlayersIsReproducible(t *testing.T) {
	run := func() world.Vec {
		e := newSpreadEngine(t)
		p := addSpreadPlayer(t, e, "Steve")
		if err := runLine(e, "spreadplayers 0 0 5 10 false Steve"); err != nil {
			t.Fatalf("spreadplayers: %v", err)
		}
		return p.Pos
	}
	first := run()
	second := run()
	if first != second {
		t.Errorf("the two runs put the player at %v and %v, want the same spot", first, second)
	}
}

// TestSpreadPlayersFailsWhenThereIsNoRoom pins down what 1.12.2 says when
// the players do not fit.
//
// (Measured on a real 1.12.2 server: "Could not spread 12 player(s)
// around 0.5, 0.5 (too many players for space - try using spread of at
// most 2.16)".)
func TestSpreadPlayersFailsWhenThereIsNoRoom(t *testing.T) {
	e := newSpreadEngine(t)
	for i := 0; i < 12; i++ {
		addSpreadPlayer(t, e, "Steve"+strings.Repeat("x", i))
	}
	err := runLine(e, "spreadplayers 0 0 100 101 false "+strings.Join(playerNames(12), " "))
	if err == nil {
		t.Fatal("expected an error when the players do not fit")
	}
	want := "Could not spread 12 player(s) around 0.5, 0.5 (too many players for space - try using spread of at most "
	if !strings.Contains(err.Error(), want) {
		t.Errorf("the error was %q, want it to contain %q", err.Error(), want)
	}
}

// playerNames returns the names of the players of the failure test.
func playerNames(n int) []string {
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, "Steve"+strings.Repeat("x", i))
	}
	return out
}

// TestSpreadPlayersRelativeCentre pins down how a relative centre is read.
//
// 1.12.2 reads the centre with `CommandAbstract.b(base, text, true)`: a
// `~` is the position of the command and a `~` with a number is that
// number of blocks away from it. A relative centre is never moved to the
// middle of a block.
//
// (Measured on a real 1.12.2 server: `spreadplayers ~5 ~5 ...` answers
// "around 5.0, 5.0", `spreadplayers ~ ~5.5 ...` answers "around 0.0, 5.5"
// and `spreadplayers ~ ~ ...` answers "around 0.0, 0.0".)
func TestSpreadPlayersRelativeCentre(t *testing.T) {
	e := newSpreadEngine(t)
	addSpreadPlayer(t, e, "Steve")
	for _, tc := range []struct {
		line string
		want string
	}{
		{
			"spreadplayers ~5 ~5 5 10 false Steve",
			"Spreading 1 player(s) 10.0 blocks around 5.0, 5.0 (min 5.0 blocks apart)",
		},
		{
			"spreadplayers ~ ~5.5 5 10 false Steve",
			"Spreading 1 player(s) 10.0 blocks around 0.0, 5.5 (min 5.0 blocks apart)",
		},
		{
			"spreadplayers ~ ~ 5 10 false Steve",
			"Spreading 1 player(s) 10.0 blocks around 0.0, 0.0 (min 5.0 blocks apart)",
		},
		{
			"spreadplayers 0 0 0 1 false Steve",
			"Spreading 1 player(s) 1.0 blocks around 0.5, 0.5 (min 0.0 blocks apart)",
		},
	} {
		msgs, err := runLineFeedback(e, tc.line)
		if err != nil {
			t.Errorf("%q: %v", tc.line, err)
			continue
		}
		if msgs[0] != tc.want {
			t.Errorf("%q: the first line was %q, want %q", tc.line, msgs[0], tc.want)
		}
	}
}

// TestSpreadPlayersEndToEnd runs a function that spreads the players, the
// way `gomc run <world> <function>` does.
//
// The command is registered next to the whole 1.12.2 command set, which is
// what the command set of the server looks like once the command is part
// of it.
//
// gomc's world has no players, so the players of the function are the ones
// `/summon` puts into the world. 1.12.2 takes any entity a selector
// matches, not only the players.
func TestSpreadPlayersEndToEnd(t *testing.T) {
	root := t.TempDir()
	body := strings.Join([]string{
		"summon zombie 0 4 0",
		"summon zombie 6 4 6",
		"summon zombie -6 4 -6",
		"spreadplayers 0 0 5 10 false @e[type=zombie]",
		"spreadplayers 0 0 5 10 true @e[type=zombie]",
	}, "\n") + "\n"
	writeFunction(t, root, "demo", "spread", body)

	reg := exec.NewRegistry()
	Register(reg)
	registerSpreadPlayers(reg)
	e := exec.NewEngine(reg, world.NewWorld())
	e.SetRoot(root)
	if errs := e.LoadFunctions(); len(errs) > 0 {
		t.Fatalf("LoadFunctions: %v", errs)
	}
	src := &exec.Source{Name: "Server", World: "world", PermissionLevel: 4, Feedback: true}
	count, err := e.RunFunctionByID("demo:spread", src)
	if err != nil {
		t.Fatalf("RunFunctionByID: %v", err)
	}
	for _, msg := range e.VisibleFeedback() {
		t.Log(msg)
	}
	t.Logf("ran %d command(s) from demo:spread", count)
	for _, p := range e.World.Players() {
		t.Logf("player %s is at %v", p.PlayerName, p.Pos)
	}
	for _, p := range e.World.EntitySlice() {
		if p.IsPlayer {
			continue
		}
		t.Logf("%s is at %v", p.Type, p.Pos)
	}
}

// TestBiomeRegistry pins down the biome registry of 1.12.2.
//
// The table comes from the server's `BiomeBase` class, which registers 62
// biomes.
func TestBiomeRegistry(t *testing.T) {
	if len(Biomes) != 62 {
		t.Fatalf("the registry has %d biomes, want 62", len(Biomes))
	}
	for i, b := range Biomes {
		if b.ID != i {
			t.Errorf("the biome at %d has the id %d", i, b.ID)
		}
		if b.Name == "" || b.DisplayName == "" {
			t.Errorf("the biome at %d is %q / %q", i, b.Name, b.DisplayName)
		}
	}
	for _, tc := range []struct {
		name    string
		display string
	}{
		{"ocean", "Ocean"},
		{"plains", "Plains"},
		{"ice_flats", "Ice Plains"},
		{"sky", "The End"},
		{"void", "The Void"},
		{"mutated_forest", "Flower Forest"},
		{"mutated_mesa", "Mesa (Bryce)"},
	} {
		b, ok := BiomeByName(tc.name)
		if !ok {
			t.Errorf("the biome %q is not in the registry", tc.name)
			continue
		}
		if b.DisplayName != tc.display {
			t.Errorf("the biome %q is shown as %q, want %q", tc.name, b.DisplayName, tc.display)
		}
	}
	// The name may be given with its namespace.
	if b, ok := BiomeByName("minecraft:plains"); !ok || b.DisplayName != "Plains" {
		t.Errorf("the biome \"minecraft:plains\" is not in the registry")
	}
	if _, ok := BiomeByName("notab Biome"); ok {
		t.Errorf("a biome that does not exist was found")
	}
	if b, ok := BiomeByID(1); !ok || b.Name != "plains" {
		t.Errorf("the biome with the id 1 is %q", b.Name)
	}
	if _, ok := BiomeByID(62); ok {
		t.Errorf("a biome with the id 62 was found")
	}
}
