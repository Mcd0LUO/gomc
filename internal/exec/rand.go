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
