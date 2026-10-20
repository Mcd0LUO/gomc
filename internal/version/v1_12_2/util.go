package v1_12_2

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// checkBlockName reports whether the server knows the block.
//
// 1.12.2 reads the name as a `MinecraftKey` and looks it up in the block
// registry. A name that is not in there is reported as
// "There is no such block with name minecraft:<name>".
//
// The check belongs to the parse step, which is when 1.12.2 does it too:
// a function with a bad block name still loads, and the failure shows up
// when the function runs.
//
// (Measured on a real 1.12.2 server.)
func checkBlockName(p *parser.Parser, name string) error {
	if world.IsKnownBlock(name) {
		return nil
	}
	namespace, key := world.EntityTypeKey(name)
	return p.Errorf("There is no such block with name %s:%s", namespace, key)
}

// parseIntSafe parses a decimal integer, reporting a friendly error.
func parseIntSafe(s string) (int, error) {
	n, ok := parser.JavaInt(s)
	if !ok {
		return 0, fmt.Errorf("'%s' is not a valid number", strings.TrimSpace(s))
	}
	return n, nil
}

// parseFloatSafe parses a decimal number, reporting a friendly error.
func parseFloatSafe(s string) (float64, error) {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, fmt.Errorf("expected a number, got %q", s)
	}
	return f, nil
}
