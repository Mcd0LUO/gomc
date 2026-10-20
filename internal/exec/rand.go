package exec

import (
	"math/rand"
	"sync"
)

// randSource is the engine's random number generator.
//
// Determinism matters for a test harness: the same script must produce the
// same result on every run. The source is seeded explicitly and can be
// locked to a fixed seed for reproducible tests.
//
// The source is shared by everything that is randomised — the `@r`
// selector and the default duration of `/weather` — which is what 1.12.2
// does as well: both take from the random source of the world. The order
// of the commands therefore matters, exactly as it does on a server.
//
// A new engine starts the source over, so that a run does not depend on
// how many random numbers the runs before it took.
var (
	randMu   sync.Mutex
	randSrc  = rand.New(rand.NewSource(1))
	randSeed = int64(1)
)

// SetSeed reseeds the random source used by `@r` selectors and other
// randomised behaviour.
func SetSeed(seed int64) {
	randMu.Lock()
	defer randMu.Unlock()
	randSeed = seed
	randSrc = rand.New(rand.NewSource(seed))
}

// Seed returns the current random seed.
func Seed() int64 {
	randMu.Lock()
	defer randMu.Unlock()
	return randSeed
}

// randIntn returns a random int in [0,n).
func randIntn(n int) int {
	if n <= 0 {
		return 0
	}
	randMu.Lock()
	defer randMu.Unlock()
	return randSrc.Intn(n)
}

// RandIntn returns a random int in [0,n).
//
// The commands use it for the parts of 1.12.2 that are randomised, such as
// the default duration of `/weather`, so that a seeded run stays
// reproducible.
func RandIntn(n int) int { return randIntn(n) }
