package v1_12_2

import (
	"fmt"
	"strconv"
	"strings"
)

// parseIntSafe parses a decimal integer, reporting a friendly error.
func parseIntSafe(s string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, fmt.Errorf("expected an integer, got %q", s)
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
