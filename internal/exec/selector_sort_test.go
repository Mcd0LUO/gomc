package exec

import (
	"sort"
	"testing"

	"github.com/Mcd0LUO/gomc/internal/world"
)

// benchEntities builds n entities spread over a line, the way a world full
// of mobs is spread over the ground.
func benchEntities(n int) []*world.Entity {
	out := make([]*world.Entity, 0, n)
	for i := 0; i < n; i++ {
		e := &world.Entity{ID: "e" + string(rune('a'+i%26)) + string(rune('a'+i/26))}
		e.Pos = world.Vec{X: float64(i) * 0.37, Y: 64, Z: float64(i % 7)}
		out = append(out, e)
	}
	return out
}

// BenchmarkSortThenFirst measures what 1.12.2 does: it sorts the matching
// entities by their distance and then takes the first one.
//
// This is the cost the shortcut for `c=1` avoids.
func BenchmarkSortThenFirst(b *testing.B) {
	for _, n := range []int{100, 1000, 8000} {
		b.Run("Entities"+itoa(n), func(b *testing.B) {
			entities := benchEntities(n)
			ref := world.Vec{X: 0, Y: 64, Z: 0}
			r := &SelectorResolver{}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				list := make([]*world.Entity, len(entities))
				copy(list, entities)
				sort.SliceStable(list, func(a, c int) bool {
					return r.squaredDistance(ref, list[a]) < r.squaredDistance(ref, list[c])
				})
				_ = list[0]
			}
		})
	}
}

// BenchmarkClosestOnly measures what the shortcut for `c=1` does: it walks
// the list once and keeps the closest entity.
//
// The entities are built once, so the numbers are not blurred by the
// garbage collector the way the end-to-end benchmarks are.
func BenchmarkClosestOnly(b *testing.B) {
	for _, n := range []int{100, 1000, 8000} {
		b.Run("Entities"+itoa(n), func(b *testing.B) {
			entities := benchEntities(n)
			ref := world.Vec{X: 0, Y: 64, Z: 0}
			r := &SelectorResolver{}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = r.closestEntity(entities, ref)
			}
		})
	}
}

// itoa turns a number into its decimal form.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [8]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
