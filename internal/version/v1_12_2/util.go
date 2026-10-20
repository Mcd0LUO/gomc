package v1_12_2

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/parser"
)

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
