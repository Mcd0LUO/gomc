// Package exec_test holds the benchmarks and the checks that guard the
// performance of the target selector.
//
// The file lives in the external test package on purpose: package
// v1_12_2 imports package exec, so a test file inside package exec could
// not import the real command set back without an import cycle.
//
// The selector walks the entities of the world without copying them, so
// the world keeps the list of its entities cached. The checks here make
// sure that the cache follows the entities that come and go, and the
// benchmarks make sure that a selection does not get more expensive when
// the world grows.
package exec_test

import (
	"fmt"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// perfEntities is the number of entities the selector benchmarks walk.
const perfEntities = 8000

// addEntitiesOfType puts n entities of the given type into the world.
//
// The entities are spread along the x axis, so that a selector that
// measures distances has something to sort.
func addEntitiesOfType(w *world.World, typ string, n int) {
	for i := 0; i < n; i++ {
		e, ok := w.EntityTypes.NewEntity(typ, "")
		if !ok {
			return
		}
		e.Pos = world.Vec{X: float64(i), Y: 64, Z: 0}
		w.AddEntity(e)
	}
}

// resolveIds resolves a selector and returns the ids of the matched
// entities, in the order the selector returned them.
func resolveIds(tb testing.TB, w *world.World, text string) []string {
	tb.Helper()
	sel, err := parser.ParseSelector(text)
	if err != nil {
		tb.Fatalf("ParseSelector(%q): %v", text, err)
	}
	r := &exec.SelectorResolver{World: w}
	src := &exec.Source{Pos: world.Vec{X: 0, Y: 64, Z: 0}}
	entities, err := r.Resolve(src, sel)
	if err != nil {
		tb.Fatalf("Resolve(%q): %v", text, err)
	}
	out := make([]string, 0, len(entities))
	for _, e := range entities {
		out = append(out, e.ID)
	}
	return out
}

// BenchmarkSelectorScanFewHits measures a selector that walks a big world
// and matches almost nothing.
//
// This is the case the cached entity list is for. The world is walked
// once per selector, so the price of a selection must not grow with the
// number of entities of the world, and the result must not be sized after
// the number of candidates either — hardly any of them match.
func BenchmarkSelectorScanFewHits(b *testing.B) {
	sel, err := parser.ParseSelector("@e[type=zombie]")
	if err != nil {
		b.Fatalf("ParseSelector: %v", err)
	}

	for _, n := range []int{100, 500, 2000, perfEntities} {
		b.Run(fmt.Sprintf("Entities%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				eng := newBenchEngine(b)
				src := benchSource()
				addEntitiesOfType(eng.World, "armor_stand", n)
				// One entity of the wanted type, at the far end of
				// the world.
				zombie, ok := eng.World.EntityTypes.NewEntity("zombie", "")
				if !ok {
					b.Fatal("zombie is not a registered entity type")
				}
				zombie.Pos = world.Vec{X: float64(n), Y: 64, Z: 0}
				eng.World.AddEntity(zombie)
				resolver := &exec.SelectorResolver{World: eng.World}
				b.StartTimer()
				matched, err := resolver.Resolve(src, sel)
				if err != nil {
					b.Fatalf("Resolve: %v", err)
				}
				if len(matched) != 1 {
					b.Fatalf("matched %d entities, want 1", len(matched))
				}
			}
		})
	}
}

// BenchmarkSelectorScanRepeated measures many selections over a world
// that does not change.
//
// This is the shape of a real command function: the world stands still
// while the selectors run over and over again.
func BenchmarkSelectorScanRepeated(b *testing.B) {
	for _, tc := range []struct {
		name string
		sel  string
		hits int
	}{
		{"All", "@e[type=armor_stand]", perfEntities},
		{"Few", "@e[type=zombie]", 1},
	} {
		b.Run(tc.name, func(b *testing.B) {
			sel, err := parser.ParseSelector(tc.sel)
			if err != nil {
				b.Fatalf("ParseSelector: %v", err)
			}
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				eng := newBenchEngine(b)
				src := benchSource()
				addEntitiesOfType(eng.World, "armor_stand", perfEntities)
				if tc.hits == 1 {
					zombie, ok := eng.World.EntityTypes.NewEntity("zombie", "")
					if !ok {
						b.Fatal("zombie is not a registered entity type")
					}
					zombie.Pos = world.Vec{X: perfEntities, Y: 64, Z: 0}
					eng.World.AddEntity(zombie)
				}
				resolver := &exec.SelectorResolver{World: eng.World}
				b.StartTimer()
				for j := 0; j < benchSelectorRuns; j++ {
					matched, err := resolver.Resolve(src, sel)
					if err != nil {
						b.Fatalf("Resolve: %v", err)
					}
					if len(matched) != tc.hits {
						b.Fatalf("matched %d entities, want %d", len(matched), tc.hits)
					}
				}
			}
		})
	}
}

// TestSelectorSeesEntityChanges makes sure that the cached entity list
// of the world follows the entities that come and go.
//
// The selector walks the entities of the world without copying them, so
// the world has to keep its cached list in step with the entities it
// holds.
func TestSelectorSeesEntityChanges(t *testing.T) {
	w := world.NewWorld()

	stand, ok := w.EntityTypes.NewEntity("armor_stand", "")
	if !ok {
		t.Fatal("armor_stand is not a registered entity type")
	}
	stand.ID = "stand-1"
	w.AddEntity(stand)

	if got := resolveIds(t, w, "@e[type=armor_stand]"); len(got) != 1 || got[0] != "stand-1" {
		t.Fatalf("after the first add: got %v, want [stand-1]", got)
	}

	// A second entity of the same type has to show up.
	stand2, _ := w.EntityTypes.NewEntity("armor_stand", "")
	stand2.ID = "stand-2"
	w.AddEntity(stand2)
	if got := resolveIds(t, w, "@e[type=armor_stand]"); len(got) != 2 {
		t.Fatalf("after the second add: got %v, want two entities", got)
	}

	// An entity that is removed has to disappear.
	if !w.RemoveEntity("stand-1") {
		t.Fatal("RemoveEntity(stand-1) reported no removal")
	}
	if got := resolveIds(t, w, "@e[type=armor_stand]"); len(got) != 1 || got[0] != "stand-2" {
		t.Fatalf("after the removal: got %v, want [stand-2]", got)
	}

	// An entity that replaces another one under the same id has to be
	// the one the selector finds.
	replacement, _ := w.EntityTypes.NewEntity("zombie", "")
	replacement.ID = "stand-2"
	w.AddEntity(replacement)
	if got := resolveIds(t, w, "@e[type=armor_stand]"); len(got) != 0 {
		t.Fatalf("after the replacement: got %v, want no entity", got)
	}
	if got := resolveIds(t, w, "@e[type=zombie]"); len(got) != 1 || got[0] != "stand-2" {
		t.Fatalf("after the replacement: got %v, want [stand-2]", got)
	}
}

// TestSelectorKeepsSpawnOrder makes sure that entities at the same
// distance keep the order they were spawned in.
//
// The selector sorts the matched entities by their distance, and the
// order of the entities of the world decides the order of the entities
// that are equally far away.
func TestSelectorKeepsSpawnOrder(t *testing.T) {
	w := world.NewWorld()
	for _, id := range []string{"a", "b", "c"} {
		e, ok := w.EntityTypes.NewEntity("armor_stand", "")
		if !ok {
			t.Fatal("armor_stand is not a registered entity type")
		}
		e.ID = id
		// All three sit at the same place, so all three are equally
		// far away from the reference.
		e.Pos = world.Vec{X: 10, Y: 64, Z: 0}
		w.AddEntity(e)
	}

	got := resolveIds(t, w, "@e[type=armor_stand]")
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

// TestWorldEntityListOrder makes sure that the cached entity list of the
// world holds the entities in spawn order, and that it follows the
// entities that come and go.
func TestWorldEntityListOrder(t *testing.T) {
	w := world.NewWorld()
	for _, id := range []string{"a", "b", "c"} {
		e, _ := w.EntityTypes.NewEntity("armor_stand", "")
		e.ID = id
		w.AddEntity(e)
	}

	ids := func(entities []*world.Entity) []string {
		out := make([]string, 0, len(entities))
		for _, e := range entities {
			out = append(out, e.ID)
		}
		return out
	}

	want := []string{"a", "b", "c"}
	got := ids(w.EntityList())
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("EntityList: got %v, want %v", got, want)
		}
	}

	w.RemoveEntity("b")
	want = []string{"a", "c"}
	got = ids(w.EntityList())
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("EntityList after the removal: got %v, want %v", got, want)
		}
	}

	// The cached list is the same list, and it is not a copy: the world
	// hands out the same entities in the same order.
	cached := w.EntitySlice()
	if len(cached) != 2 || cached[0].ID != "a" || cached[1].ID != "c" {
		t.Fatalf("EntitySlice: got %v, want [a c]", ids(cached))
	}
}
