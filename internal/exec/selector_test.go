package exec

import (
	"testing"

	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// testWorld builds a world with a handful of entities to filter against.
func testWorld(t *testing.T) *world.World {
	t.Helper()
	w := world.NewWorld()

	player := &world.Entity{
		ID: "p1", Type: "player", IsPlayer: true, PlayerName: "Steve",
		Pos: world.Vec{X: 0, Y: 0, Z: 0},
		NBT: nbt.NewCompound(), Width: 0.6, Height: 1.8,
	}
	w.AddEntity(player)

	far := &world.Entity{
		ID: "p2", Type: "player", IsPlayer: true, PlayerName: "Alex",
		Pos: world.Vec{X: 100, Y: 0, Z: 0},
		NBT: nbt.NewCompound(), Width: 0.6, Height: 1.8,
	}
	w.AddEntity(far)

	zombie := &world.Entity{
		ID: "z1", Type: "zombie",
		Pos: world.Vec{X: 2, Y: 0, Z: 0},
		NBT: nbt.NewCompound(), Width: 0.6, Height: 1.95,
	}
	w.AddEntity(zombie)

	stand := &world.Entity{
		ID: "a1", Type: "armor_stand",
		Pos: world.Vec{X: 3, Y: 0, Z: 0},
		NBT: nbt.NewCompound(), Width: 0.5, Height: 1.975,
	}
	w.AddEntity(stand)

	return w
}

// resolve runs a selector against a world and returns the matched ids.
func resolve(t *testing.T, w *world.World, src *Source, text string) []string {
	t.Helper()
	sel, err := parser.ParseSelector(text)
	if err != nil {
		t.Fatalf("ParseSelector(%q): %v", text, err)
	}
	r := &SelectorResolver{World: w}
	entities, err := r.Resolve(src, sel)
	if err != nil {
		t.Fatalf("Resolve(%q): %v", text, err)
	}
	out := make([]string, 0, len(entities))
	for _, e := range entities {
		out = append(out, e.ID)
	}
	return out
}

// sameSet compares two id lists ignoring order.
func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := map[string]int{}
	for _, x := range a {
		seen[x]++
	}
	for _, x := range b {
		seen[x]--
	}
	for _, n := range seen {
		if n != 0 {
			return false
		}
	}
	return true
}

func TestSelectorBases(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}

	// @e selects every entity.
	got := resolve(t, w, src, "@e")
	if !sameSet(got, []string{"p1", "p2", "z1", "a1"}) {
		t.Errorf("@e = %v, want all four entities", got)
	}

	// @a selects players only.
	got = resolve(t, w, src, "@a")
	if !sameSet(got, []string{"p1", "p2"}) {
		t.Errorf("@a = %v, want the two players", got)
	}

	// @p selects the nearest player.
	got = resolve(t, w, src, "@p")
	if len(got) != 1 || got[0] != "p1" {
		t.Errorf("@p = %v, want p1", got)
	}
}

// TestSelectorRandomIsPlayerOnly covers the @r player filter.
//
// 1.12.2 adds a "must be a player" filter to @r, so @r never picks a mob
// unless a type argument replaces that filter.
func TestSelectorRandomIsPlayerOnly(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}

	// @r defaults to a single entity.
	got := resolve(t, w, src, "@r")
	if len(got) != 1 {
		t.Fatalf("@r = %v, want a single entity", got)
	}
	if got[0] != "p1" && got[0] != "p2" {
		t.Errorf("@r = %v, want a player", got)
	}

	// Asking for more shows that only players are eligible.
	got = resolve(t, w, src, "@r[c=2]")
	if !sameSet(got, []string{"p1", "p2"}) {
		t.Errorf("@r[c=2] = %v, want the two players", got)
	}

	// With a type argument the player filter is replaced, so mobs are
	// eligible too.
	got = resolve(t, w, src, "@r[type=zombie]")
	if !sameSet(got, []string{"z1"}) {
		t.Errorf("@r[type=zombie] = %v, want z1", got)
	}
}

func TestSelectorTypeFilter(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}

	got := resolve(t, w, src, "@e[type=zombie]")
	if !sameSet(got, []string{"z1"}) {
		t.Errorf("@e[type=zombie] = %v, want z1", got)
	}

	// The negated form excludes the type.
	got = resolve(t, w, src, "@e[type=!zombie]")
	if !sameSet(got, []string{"p1", "p2", "a1"}) {
		t.Errorf("@e[type=!zombie] = %v, want everything but z1", got)
	}
}

func TestSelectorNameFilter(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}

	// A player is matched by the player name.
	got := resolve(t, w, src, "@e[name=Steve]")
	if !sameSet(got, []string{"p1"}) {
		t.Errorf("@e[name=Steve] = %v, want p1", got)
	}

	// A mob is matched by the type's registered name.
	got = resolve(t, w, src, "@e[name=Zombie]")
	if !sameSet(got, []string{"z1"}) {
		t.Errorf("@e[name=Zombie] = %v, want z1", got)
	}

	// An entity with a custom name is matched by that name.
	w.Entities["z1"].NBT.Set("CustomName", nbt.String("Bob"))
	got = resolve(t, w, src, "@e[name=Bob]")
	if !sameSet(got, []string{"z1"}) {
		t.Errorf("@e[name=Bob] = %v, want z1", got)
	}

	// The negated form excludes the name.
	got = resolve(t, w, src, "@e[name=!Bob]")
	if !sameSet(got, []string{"p1", "p2", "a1"}) {
		t.Errorf("@e[name=!Bob] = %v, want everything but z1", got)
	}
}

func TestSelectorTeamFilter(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}
	w.Scoreboard.AddTeam("red")
	w.Scoreboard.AddTeamMember("red", ScoreboardHolder(w.Entities["z1"]))

	got := resolve(t, w, src, "@e[team=red]")
	if !sameSet(got, []string{"z1"}) {
		t.Errorf("@e[team=red] = %v, want z1", got)
	}

	// An entity on no team reports an empty team name, so the negated
	// form still matches it.
	got = resolve(t, w, src, "@e[team=!red]")
	if !sameSet(got, []string{"p1", "p2", "a1"}) {
		t.Errorf("@e[team=!red] = %v, want the three entities on no team", got)
	}
}

func TestSelectorTagFilter(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}
	w.Entities["z1"].SetTag("marked", true)

	got := resolve(t, w, src, "@e[tag=marked]")
	if !sameSet(got, []string{"z1"}) {
		t.Errorf("@e[tag=marked] = %v, want z1", got)
	}

	got = resolve(t, w, src, "@e[tag=!marked]")
	if !sameSet(got, []string{"p1", "p2", "a1"}) {
		t.Errorf("@e[tag=!marked] = %v, want the untagged entities", got)
	}

	// An empty tag name selects the entities that carry no tag at all.
	got = resolve(t, w, src, "@e[tag=]")
	if !sameSet(got, []string{"p1", "p2", "a1"}) {
		t.Errorf("@e[tag=] = %v, want the untagged entities", got)
	}
}

// TestSelectorScoreFilter covers the 1.12.2 score syntax.
//
// 1.12.2 spells the conditions `score_<objective>` (a maximum) and
// `score_<objective>_min` (a minimum). The objective has to exist and the
// entity has to have a score in it.
func TestSelectorScoreFilter(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}
	w.Scoreboard.AddObjective("kills", world.CriteriaDummy, "")
	w.Scoreboard.SetScore(ScoreboardHolder(w.Entities["z1"]), "kills", 5)

	got := resolve(t, w, src, "@e[score_kills=5]")
	if !sameSet(got, []string{"z1"}) {
		t.Errorf("@e[score_kills=5] = %v, want z1", got)
	}

	// The unsuffixed name is the maximum, so a lower score still matches.
	got = resolve(t, w, src, "@e[score_kills=9]")
	if !sameSet(got, []string{"z1"}) {
		t.Errorf("@e[score_kills=9] = %v, want z1", got)
	}

	// A score above the maximum does not match.
	got = resolve(t, w, src, "@e[score_kills=4]")
	if len(got) != 0 {
		t.Errorf("@e[score_kills=4] = %v, want no match", got)
	}

	// The _min form is the minimum.
	got = resolve(t, w, src, "@e[score_kills_min=5]")
	if !sameSet(got, []string{"z1"}) {
		t.Errorf("@e[score_kills_min=5] = %v, want z1", got)
	}
	got = resolve(t, w, src, "@e[score_kills_min=6]")
	if len(got) != 0 {
		t.Errorf("@e[score_kills_min=6] = %v, want no match", got)
	}

	// The unsuffixed name is a maximum, so a score below it still matches.
	got = resolve(t, w, src, "@e[score_kills=100]")
	if !sameSet(got, []string{"z1"}) {
		t.Errorf("@e[score_kills=100] = %v, want z1", got)
	}

	// An objective that does not exist never matches.
	got = resolve(t, w, src, "@e[score_nosuch=1]")
	if len(got) != 0 {
		t.Errorf("@e[score_nosuch=1] = %v, want no match", got)
	}
}

// TestSelectorGameModeFilter covers the m= filter.
//
// 1.12.2 reads the game mode from the player's interact manager, so the
// filter only ever matches players, and a leading ! negates it.
func TestSelectorGameModeFilter(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}
	w.Entities["p1"].NBT.Set("playerGameType", nbt.Int(1))
	w.Entities["p2"].NBT.Set("playerGameType", nbt.Int(0))

	got := resolve(t, w, src, "@e[m=creative]")
	if !sameSet(got, []string{"p1"}) {
		t.Errorf("@e[m=creative] = %v, want p1", got)
	}

	// The numeric id works too.
	got = resolve(t, w, src, "@e[m=1]")
	if !sameSet(got, []string{"p1"}) {
		t.Errorf("@e[m=1] = %v, want p1", got)
	}

	// The negation must not be swallowed.
	got = resolve(t, w, src, "@e[m=!creative]")
	if !sameSet(got, []string{"p2"}) {
		t.Errorf("@e[m=!creative] = %v, want p2", got)
	}

	// A mob is never matched by the game mode filter.
	got = resolve(t, w, src, "@e[m=survival]")
	if !sameSet(got, []string{"p2"}) {
		t.Errorf("@e[m=survival] = %v, want p2", got)
	}
}

// TestSelectorRotationFilter covers the rx/ry bounds.
//
// The unsuffixed argument is the upper bound and the _m suffixed one the
// lower bound, and a window whose lower bound is above its upper bound
// wraps around north.
func TestSelectorRotationFilter(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}
	w.Entities["p1"].Rot = world.Rotation{Pitch: 0, Yaw: 0}
	w.Entities["p2"].Rot = world.Rotation{Pitch: 90, Yaw: 90}
	w.Entities["z1"].Rot = world.Rotation{Pitch: 0, Yaw: 0}
	w.Entities["a1"].Rot = world.Rotation{Pitch: 0, Yaw: -90}

	// ry is the maximum, so only the entities at 0 degrees are inside.
	got := resolve(t, w, src, "@e[ry=45]")
	if !sameSet(got, []string{"p1", "z1"}) {
		t.Errorf("@e[ry=45] = %v, want p1 and z1", got)
	}

	// rym is the minimum. The default upper bound is 359, which wraps to
	// -1, so a window with only a lower bound wraps around north and
	// covers everything from 45 degrees up as well as everything at or
	// below -1 degrees.
	got = resolve(t, w, src, "@e[rym=45]")
	if !sameSet(got, []string{"p2", "a1"}) {
		t.Errorf("@e[rym=45] = %v, want p2 and a1", got)
	}

	// A window from 45 to 90 contains the 90 degree entity.
	got = resolve(t, w, src, "@e[rym=45,ry=90]")
	if !sameSet(got, []string{"p2"}) {
		t.Errorf("@e[rym=45,ry=90] = %v, want p2", got)
	}

	// A wrapping window covers both sides of north: the entity at 90
	// degrees and the one at -90 degrees.
	got = resolve(t, w, src, "@e[rym=90,ry=-90]")
	if !sameSet(got, []string{"p2", "a1"}) {
		t.Errorf("@e[rym=90,ry=-90] = %v, want p2 and a1", got)
	}

	// The pitch bounds work the same way.
	got = resolve(t, w, src, "@e[rx=45]")
	if !sameSet(got, []string{"p1", "z1", "a1"}) {
		t.Errorf("@e[rx=45] = %v, want p1, z1 and a1", got)
	}
	got = resolve(t, w, src, "@e[rxm=45]")
	if !sameSet(got, []string{"p2"}) {
		t.Errorf("@e[rxm=45] = %v, want p2", got)
	}
}

func TestSelectorDistanceFilter(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}

	// r is the maximum distance.
	got := resolve(t, w, src, "@e[r=5]")
	if !sameSet(got, []string{"p1", "z1", "a1"}) {
		t.Errorf("@e[r=5] = %v, want p1, z1 and a1", got)
	}

	// rm is the minimum distance.
	got = resolve(t, w, src, "@e[rm=50]")
	if !sameSet(got, []string{"p2"}) {
		t.Errorf("@e[rm=50] = %v, want p2", got)
	}

	// Both together bound the distance from both sides.
	got = resolve(t, w, src, "@e[rm=1,r=5]")
	if !sameSet(got, []string{"z1", "a1"}) {
		t.Errorf("@e[rm=1,r=5] = %v, want z1 and a1", got)
	}
}

// TestSelectorVolumeFilter covers the dx/dy/dz volume.
//
// 1.12.2 selects the entities whose bounding box intersects the volume,
// not the entities whose position point falls inside it. The box is one
// block larger on the far side, so a dx of 1 covers two blocks.
func TestSelectorVolumeFilter(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}

	// A volume of one block at the origin also reaches the zombie at x=2,
	// because the box ends at x=2 and the zombie is 0.6 wide.
	got := resolve(t, w, src, "@e[dx=1,dy=1,dz=1]")
	if !sameSet(got, []string{"p1", "z1"}) {
		t.Errorf("@e[dx=1,dy=1,dz=1] = %v, want p1 and z1", got)
	}

	// The armor stand at x=3 is outside a volume anchored at the origin.
	got = resolve(t, w, src, "@e[dx=1,dy=1,dz=1,x=3]")
	if !sameSet(got, []string{"a1"}) {
		t.Errorf("@e[dx=1,dy=1,dz=1,x=3] = %v, want a1", got)
	}

	// A tall volume reaches the player standing on the ground.
	got = resolve(t, w, src, "@e[dx=0,dy=0,dz=0]")
	if !sameSet(got, []string{"p1"}) {
		t.Errorf("@e[dx=0,dy=0,dz=0] = %v, want p1", got)
	}
}

func TestSelectorCountLimit(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}

	// @e defaults to no limit.
	got := resolve(t, w, src, "@e")
	if len(got) != 4 {
		t.Errorf("@e matched %v, want all four entities", got)
	}

	// c caps the result.
	got = resolve(t, w, src, "@e[c=2]")
	if len(got) != 2 {
		t.Errorf("@e[c=2] = %v, want two entities", got)
	}

	// A negative count reverses the whole list before truncating, so it
	// is not simply the reverse of the positive result: the far away
	// entity comes first.
	forward := resolve(t, w, src, "@e[c=3]")
	backward := resolve(t, w, src, "@e[c=-3]")
	wantForward := []string{"p1", "z1", "a1"}
	wantBackward := []string{"p2", "a1", "z1"}
	if !sameSet(forward, wantForward) {
		t.Errorf("@e[c=3] = %v, want %v", forward, wantForward)
	}
	if !sameSet(backward, wantBackward) {
		t.Errorf("@e[c=-3] = %v, want %v", backward, wantBackward)
	}
}

// TestSelectorSelf covers the @s base.
//
// @s is the executing entity. Without one — a console, or a function that
// is not run through /execute — the selector reports that there is no
// executor rather than silently matching nothing.
func TestSelectorSelf(t *testing.T) {
	w := testWorld(t)
	player := w.Entities["p1"]
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4, Entity: player}

	got := resolve(t, w, src, "@s")
	if len(got) != 1 || got[0] != "p1" {
		t.Errorf("@s = %v, want p1", got)
	}

	// A console has no entity, so @s has no executor and matches
	// nothing. The wording of the failure belongs to the command that
	// used the selector: 1.12.2 answers "Entity '@s' cannot be found"
	// for `kill @s` and "Selector '@s' found nothing" for a message that
	// expands @s.
	plain := &Source{Name: "Server", World: "world", PermissionLevel: 4}
	sel, err := parser.ParseSelector("@s")
	if err != nil {
		t.Fatalf("ParseSelector: %v", err)
	}
	r := &SelectorResolver{World: w}
	selfResolved, resolveErr := r.Resolve(plain, sel)
	if resolveErr != nil {
		t.Errorf("@s without an executor: %v, want no error", resolveErr)
	}
	if len(selfResolved) != 0 {
		t.Errorf("@s without an executor = %v, want nothing", selfResolved)
	}
}

// TestSelectorCountFallsBackToDefault covers a `c` argument that is not an
// integer.
//
// 1.12.2 reads the count with `MathHelper.a(String, int)`, which is
// `Integer.parseInt` with the default as the fallback. A count that is not
// an integer is therefore not an error, it falls back to the default of
// the base: no limit for @a and @e, a single entity for the others.
//
// Measured on a real 1.12.2 server: `say @e[c=abc]` expands to every
// entity, the same as `say @e`.
func TestSelectorCountFallsBackToDefault(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}

	all := []string{"p1", "p2", "z1", "a1"}
	for _, text := range []string{"@e[c=abc]", "@e[c=1.0]", "@e[c=~]", "@e[c=99999999999999]"} {
		got := resolve(t, w, src, text)
		if !sameSet(got, all) {
			t.Errorf("%s = %v, want every entity", text, got)
		}
	}

	// A count that *is* an integer still limits the result.
	got := resolve(t, w, src, "@e[c=2]")
	if len(got) != 2 {
		t.Errorf("@e[c=2] = %v, want two entities", got)
	}
}

// TestSelectorNameUsesUserName covers the `name` filter of a player.
//
// 1.12.2's `EntityHuman.getName` overrides `Entity.getName` and returns
// the name of the game profile, so a player is matched by its user name
// and never by its custom name.
func TestSelectorNameUsesUserName(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}
	// A selector argument may not contain a space, so the custom name of
	// the test player has to be a single word.
	w.Entities["p1"].NBT.Set("CustomName", nbt.String("SteveTheGreat"))

	// The user name matches.
	got := resolve(t, w, src, "@e[name=Steve]")
	if !sameSet(got, []string{"p1"}) {
		t.Errorf("@e[name=Steve] = %v, want p1", got)
	}

	// The custom name of a player does not.
	got = resolve(t, w, src, "@e[name=SteveTheGreat]")
	if len(got) != 0 {
		t.Errorf("@e[name=SteveTheGreat] = %v, want no match", got)
	}
}

// TestSelectorCustomNameIsRaw covers the custom name of a non-player.
//
// 1.12.2 keeps the custom name in the data watcher as a plain string and
// hands it out unchanged, so the quotes of a name like `"Bob"` are part of
// the name. Measured on a real 1.12.2 server: an entity with the custom
// name `"Bob"` is printed as `"Bob"`, and it is matched by
// `@e[name="Bob"]` and not by `@e[name=Bob]`.
func TestSelectorCustomNameIsRaw(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}
	w.Entities["a1"].NBT.Set("CustomName", nbt.String(`"Bob"`))

	// The name with the quotes matches.
	got := resolve(t, w, src, `@e[name="Bob"]`)
	if !sameSet(got, []string{"a1"}) {
		t.Errorf(`@e[name="Bob"] = %v, want a1`, got)
	}

	// The name without the quotes does not.
	got = resolve(t, w, src, `@e[name=Bob]`)
	if len(got) != 0 {
		t.Errorf(`@e[name=Bob] = %v, want no match`, got)
	}
}

// TestSelectorUnknownTypeMatchesNothing covers a `type` argument the
// server does not know.
//
// 1.12.2 checks the type against the entity registry before it looks at
// anything else. A type the registry does not have — the old spelling of a
// renamed type, say — is reported as invalid and the selector matches
// nothing. The command itself does not fail.
//
// Measured on a real 1.12.2 server: `kill @e[type=pigzombie]` prints
// "Entity type 'minecraft:pigzombie' is invalid" and kills nothing.
func TestSelectorUnknownTypeMatchesNothing(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}
	// The type 1.12.2 registers the zombie pigman under. The old
	// spelling of it, "pigzombie", is not a type the server knows.
	w.AddEntity(&world.Entity{
		ID: "z2", Type: "zombie_pigman",
		Pos: world.Vec{X: 4, Y: 0, Z: 0},
		NBT: nbt.NewCompound(), Width: 0.6, Height: 1.95,
	})

	for _, text := range []string{
		"@e[type=pigzombie]",
		"@e[type=!pigzombie]",
		"@e[type=nosuchtype]",
		"@e[type=foo:zombie]",
	} {
		got := resolve(t, w, src, text)
		if len(got) != 0 {
			t.Errorf("%s = %v, want no match", text, got)
		}
	}

	// The check applies to every base: 1.12.2 rejects the type before it
	// decides what the base selects.
	got := resolve(t, w, src, "@a[type=pigzombie]")
	if len(got) != 0 {
		t.Errorf("@a[type=pigzombie] = %v, want no match", got)
	}

	// The registered name still selects the entity.
	got = resolve(t, w, src, "@e[type=zombie_pigman]")
	if !sameSet(got, []string{"z2"}) {
		t.Errorf("@e[type=zombie_pigman] = %v, want z2", got)
	}
}

// TestSelectorKnownTypeSpellings covers the spellings 1.12.2 accepts.
//
// The registry keys are lower-case and the namespace is "minecraft", but
// 1.12.2 lower-cases both parts, so the case does not matter. Measured on
// a real 1.12.2 server: `type=Zombie` and `type=MINECRAFT:ZOMBIE` both
// select the zombie.
func TestSelectorKnownTypeSpellings(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}

	for _, text := range []string{
		"@e[type=zombie]",
		"@e[type=Zombie]",
		"@e[type=ZOMBIE]",
		"@e[type=minecraft:zombie]",
		"@e[type=MINECRAFT:ZOMBIE]",
	} {
		got := resolve(t, w, src, text)
		if !sameSet(got, []string{"z1"}) {
			t.Errorf("%s = %v, want z1", text, got)
		}
	}
}

// TestSelectorGameModeNotSet covers a game mode the server does not know.
//
// 1.12.2 resolves an unknown game mode to "not set", and no player is in
// that mode. The filter therefore matches nobody, and the negated filter
// matches everybody.
func TestSelectorGameModeNotSet(t *testing.T) {
	w := testWorld(t)
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}
	w.Entities["p1"].NBT.Set("playerGameType", nbt.Int(1))
	w.Entities["p2"].NBT.Set("playerGameType", nbt.Int(0))

	// An unknown mode matches no player.
	got := resolve(t, w, src, "@e[m=abc]")
	if len(got) != 0 {
		t.Errorf("@e[m=abc] = %v, want no match", got)
	}

	// The negated form of an unknown mode matches every player.
	got = resolve(t, w, src, "@e[m=!abc]")
	if !sameSet(got, []string{"p1", "p2"}) {
		t.Errorf("@e[m=!abc] = %v, want the two players", got)
	}

	// A mob is still never matched by the game mode filter.
	got = resolve(t, w, src, "@e[m=!abc,type=zombie]")
	if len(got) != 0 {
		t.Errorf("@e[m=!abc,type=zombie] = %v, want no match", got)
	}
}

// TestCountOnePicksTheSameEntityAsTheSortedList pins down the shortcut
// that is used when exactly one entity is asked for.
//
// 1.12.2 sorts the matching entities by their distance and then takes
// the first one, and the sort is stable: entities at the same distance
// keep the order they were matched in. The shortcut finds that one
// without sorting, so it has to give the same answer -- including when
// two entities are equally far away.
func TestCountOnePicksTheSameEntityAsTheSortedList(t *testing.T) {
	w := world.NewWorld()
	// Two entities at the same distance, then one that is farther, then
	// one that is closer.
	places := []struct {
		id  string
		pos world.Vec
	}{
		{"same-a", world.Vec{X: 5, Y: 0, Z: 0}},
		{"same-b", world.Vec{X: 0, Y: 0, Z: 5}},
		{"far", world.Vec{X: 50, Y: 0, Z: 0}},
		{"near", world.Vec{X: 1, Y: 0, Z: 0}},
	}
	for _, p := range places {
		e := &world.Entity{
			ID: p.id, Type: "zombie",
			Pos: p.pos,
			NBT: nbt.NewCompound(), Width: 0.6, Height: 1.95,
		}
		w.AddEntity(e)
	}
	src := &Source{Name: "Server", World: "world", PermissionLevel: 4}

	// The closest one wins.
	got := resolve(t, w, src, "@e[c=1]")
	if len(got) != 1 || got[0] != "near" {
		t.Errorf("@e[c=1] = %v, want [near]", got)
	}

	// With the two closest ones left out, the one that was matched
	// first wins the tie.
	got = resolve(t, w, src, "@e[c=1,type=!zombie]")
	if len(got) != 0 {
		t.Errorf("@e[c=1,type=!zombie] = %v, want no match", got)
	}
	sel, err := parser.ParseSelector("@e[c=1]")
	if err != nil {
		t.Fatalf("ParseSelector: %v", err)
	}
	r := &SelectorResolver{World: w}
	entities, err := r.Resolve(src, sel)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(entities) != 1 || entities[0].ID != "near" {
		t.Fatalf("@e[c=1] = %v, want [near]", entities)
	}

	// The executor wins over the closest one, exactly like it does when
	// the whole list is sorted.
	executor := w.Entities["far"]
	srcWithEntity := &Source{Name: "Server", World: "world", PermissionLevel: 4, Entity: executor}
	entities, err = r.Resolve(srcWithEntity, sel)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(entities) != 1 || entities[0].ID != "far" {
		t.Errorf("@e[c=1] with the executor = %v, want [far]", entities)
	}
}
