package exec

import "strings"

// ClosestName returns the candidate that is closest to name, when one is
// close enough to be worth suggesting.
//
// The threshold is the usual one for "did you mean" hints: two edits, or a
// third of the name's length, whichever is larger. A name shorter than
// three characters is never corrected, because the suggestions are noise
// at that length.
func ClosestName(name string, candidates []string) (string, bool) {
	if len(name) < 3 {
		return "", false
	}
	limit := len(name) / 3
	if limit < 2 {
		limit = 2
	}
	lower := strings.ToLower(name)
	best := ""
	bestDist := limit + 1
	for _, c := range candidates {
		d := editDistance(lower, strings.ToLower(c))
		if d < bestDist {
			bestDist, best = d, c
		}
	}
	if best == "" || bestDist > limit {
		return "", false
	}
	return best, true
}

// editDistance returns the Levenshtein distance between a and b.
func editDistance(a, b string) int {
	ar := []rune(a)
	br := []rune(b)
	prev := make([]int, len(br)+1)
	curr := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		curr[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			curr[j] = minInt(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(br)]
}

func minInt(values ...int) int {
	best := values[0]
	for _, v := range values[1:] {
		if v < best {
			best = v
		}
	}
	return best
}
