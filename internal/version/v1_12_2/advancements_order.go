package v1_12_2

import (
	"sort"
	"strings"
	"unicode/utf16"
)

// The order the command walks the advancements in.
//
// 1.12.2 keeps the advancements in a `java.util.HashMap`, so the order of
// the "everything" filter is the hash order of that map:
//
//   - the server reads the files of the datapack and of its own jar into
//     one map, the advancements of the datapack first and the ones of the
//     jar after them, and
//   - the registry the command walks is filled from that map, in the order
//     the map hands its entries out.
//
// A `HashMap` hands its entries out bucket by bucket, and the buckets are
// indexed by the hash code of the key with the high bits folded into the
// low ones. The entries of one bucket keep the order they were put in.
//
// The order matters: `grant everything` walks the advancements and grants
// the ones that are not complete yet, so an advancement the recipe
// trigger completed before the command reached it is skipped -- and the
// number the command reports depends on the order.
//
// (Measured on a real 1.12.2 server: the order this reproduces is the one
// that makes `grant everything` answer "Granted every advancement (481
// total granted) to <player>" for a player that has just logged in, while
// an alphabetical order would skip eleven advancements instead of seven.)
func advancementOrder(builtIn, custom []*advancement) []*advancement {
	// The map the server reads the files into holds the advancements of
	// the datapack first and the ones of the server jar after them.
	inserted := make([]*advancement, 0, len(custom)+len(builtIn))
	inserted = append(inserted, custom...)
	inserted = append(inserted, builtIn...)

	// The map the server reads the files into.
	first := append([]*advancement(nil), inserted...)
	firstSize := javaHashMapSize(len(first))
	sort.SliceStable(first, func(i, j int) bool {
		return javaHashBucket(first[i].name, firstSize) < javaHashBucket(first[j].name, firstSize)
	})

	// The registry the command walks.
	second := append([]*advancement(nil), first...)
	secondSize := javaHashMapSize(len(second))
	sort.SliceStable(second, func(i, j int) bool {
		return javaHashBucket(second[i].name, secondSize) < javaHashBucket(second[j].name, secondSize)
	})
	return second
}

// javaHashMapSize returns the size of the table a `java.util.HashMap`
// ends up with.
//
// The table starts with sixteen buckets and doubles when it holds more
// than three quarters of them.
func javaHashMapSize(entries int) int {
	size := 16
	for entries > size*3/4 {
		size *= 2
	}
	return size
}

// javaHashBucket returns the bucket a key falls into.
//
// `MinecraftKey.hashCode` is `31 * namespace.hashCode() + name.hashCode()`,
// and the table is indexed by that hash code with the high bits folded
// into the low ones, so that the keys are spread over the table.
func javaHashBucket(name string, tableSize int) uint32 {
	namespace, path := name, ""
	if i := strings.IndexByte(name, ':'); i >= 0 {
		namespace, path = name[:i], name[i+1:]
	}
	hash := 31*javaStringHash(namespace) + javaStringHash(path)
	return (hash ^ (hash >> 16)) & uint32(tableSize-1)
}

// javaStringHash returns the hash code `String.hashCode` computes.
//
// Java hashes the UTF-16 code units of the text, so a character outside
// the basic plane counts as two.
func javaStringHash(s string) uint32 {
	var hash uint32
	for _, unit := range utf16.Encode([]rune(s)) {
		hash = 31*hash + uint32(unit)
	}
	return hash
}
