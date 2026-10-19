package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mcsim/gomc/internal/ast"
)

// Selector argument names supported by 1.12.2.
//
// The set comes from `PlayerSelector`'s static fields in the 1.12.2 NMS:
//
//	r, rm       distance bounds
//	l, lm       experience level bounds
//	x, y, z     reference position
//	dx, dy, dz  volume size
//	rx, rxm     pitch bounds
//	ry, rym     yaw bounds
//	c           result count
//	m           game mode
//	team        team name
//	name        entity name
//	type        entity type
//	tag         scoreboard tag
//
// In addition, any argument whose name starts with `score_` is a
// scoreboard condition: `score_<objective>` is the maximum score and
// `score_<objective>_min` the minimum.
//
// Note what is *not* here: `scores={...}` and `nbt={...}` are 1.13 syntax.
// The Minecraft Wiki changelog records "1.13 17w45a - added nbt to target
// selectors", and the 1.12.2 `PlayerSelector` rejects both names outright.
var selectorOptions = map[string]bool{
	"r": true, "rm": true, "l": true, "lm": true,
	"x": true, "y": true, "z": true,
	"dx": true, "dy": true, "dz": true,
	"rx": true, "rxm": true, "ry": true, "rym": true,
	"c": true, "m": true, "team": true, "name": true,
	"type": true, "tag": true,
}

// scorePrefix marks a selector argument that filters on a scoreboard
// objective, as in `score_kills=5`.
const scorePrefix = "score_"

// ParseSelector parses a target selector such as `@a[r=10,score_kills=5]`.
//
// The accepted shape is 1.12.2's `^@([pares])(?:\[([^ ]*)\])?$`: a base of
// p/a/r/e/s, optionally followed by a bracketed argument list that may not
// contain spaces.
//
// The arguments form a flat map, exactly as in 1.12.2: the body is split on
// commas, each part on its first `=`, and a repeated name keeps the last
// value. `@e[tag=a,tag=b]` therefore behaves like `@e[tag=b]`.
func ParseSelector(text string) (*ast.Selector, error) {
	base, body, ok := splitSelector(text)
	if !ok {
		return nil, fmt.Errorf("malformed selector %q", text)
	}
	sel := &ast.Selector{Base: base}
	if body == "" {
		return sel, nil
	}
	for _, part := range strings.Split(body, ",") {
		if part == "" {
			// 1.12.2 uses `Splitter.on(',').omitEmptyStrings()`.
			continue
		}
		key, value, _ := strings.Cut(part, "=")
		key = strings.TrimSpace(key)
		if !isSelectorOption(key) {
			return nil, fmt.Errorf("invalid selector argument %q", part)
		}
		sel.Args = append(sel.Args, ast.SelectorArg{Key: key, Value: strings.TrimSpace(value)})
	}
	return sel, nil
}

// splitSelector splits a selector into its base and argument body.
//
// It implements 1.12.2's pattern `^@([pares])(?:\[([^ ]*)\])?$`, so the
// whole text has to match: trailing text or a space inside the brackets
// makes the string a plain name rather than a selector.
func splitSelector(text string) (base, body string, ok bool) {
	if len(text) < 2 || text[0] != '@' {
		return "", "", false
	}
	switch text[1] {
	case 'p', 'a', 'r', 'e', 's':
	default:
		return "", "", false
	}
	rest := text[2:]
	if rest == "" {
		return string(text[1]), "", true
	}
	if !strings.HasPrefix(rest, "[") || !strings.HasSuffix(rest, "]") {
		return "", "", false
	}
	inner := rest[1 : len(rest)-1]
	if strings.ContainsAny(inner, " \t") {
		return "", "", false
	}
	return string(text[1]), inner, true
}

// isSelectorOption reports whether a selector argument name is accepted.
func isSelectorOption(key string) bool {
	if selectorOptions[key] {
		return true
	}
	return strings.HasPrefix(key, scorePrefix) && len(key) > len(scorePrefix)
}

// IsSelector reports whether text looks like a target selector.
//
// This is the cheap prefix test used to tell a selector apart from a plain
// player name. ParseSelector applies the full 1.12.2 pattern.
func IsSelector(text string) bool {
	_, _, ok := splitSelector(text)
	return ok
}

// SelectorInt returns the integer value of a selector argument.
//
// A repeated argument keeps its last value, matching the flat map 1.12.2
// builds.
func SelectorInt(sel *ast.Selector, key string) (int, bool, error) {
	v, ok := SelectorString(sel, key)
	if !ok {
		return 0, false, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, false, fmt.Errorf("selector argument %q expects an integer, got %q", key, v)
	}
	return n, true, nil
}

// SelectorFloat returns the float value of a selector argument.
func SelectorFloat(sel *ast.Selector, key string) (float64, bool, error) {
	v, ok := SelectorString(sel, key)
	if !ok {
		return 0, false, nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, false, fmt.Errorf("selector argument %q expects a number, got %q", key, v)
	}
	return f, true, nil
}

// SelectorString returns the string value of a selector argument.
//
// A repeated argument keeps its last value, matching the flat map 1.12.2
// builds: `@e[tag=a,tag=b]` behaves like `@e[tag=b]`.
func SelectorString(sel *ast.Selector, key string) (string, bool) {
	if sel == nil {
		return "", false
	}
	val, found := "", false
	for _, a := range sel.Args {
		if a.Key == key {
			val, found = a.Value, true
		}
	}
	return val, found
}

// SelectorStrings returns the list value of a selector argument.
//
// 1.12.2 has no list-valued selector argument, so this only splits a
// single value on commas. It is kept for the callers that read `tag`.
func SelectorStrings(sel *ast.Selector, key string) ([]string, bool) {
	v, ok := SelectorString(sel, key)
	if !ok {
		return nil, false
	}
	var out []string
	for _, part := range strings.Split(v, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out, true
}

// ScoreFilter is one scoreboard condition of a selector.
//
// 1.12.2 spells these as separate arguments rather than as a nested map:
// `score_kills=5` is a maximum and `score_kills_min=3` a minimum.
type ScoreFilter struct {
	// Objective is the scoreboard objective name.
	Objective string
	// Value is the bound to compare against.
	Value int
	// IsMin reports whether the bound is a minimum rather than a maximum.
	IsMin bool
}

// SelectorScoreFilters returns the scoreboard conditions of a selector.
//
// The value of a `score_` argument is an integer, defaulting to 1 when it
// is missing or unparsable, exactly as `MathHelper.a(String, int)` does.
func SelectorScoreFilters(sel *ast.Selector) []ScoreFilter {
	if sel == nil {
		return nil
	}
	// Collect in a stable order: the arguments in source order.
	seen := map[string]int{}
	var out []ScoreFilter
	for _, a := range sel.Args {
		if !strings.HasPrefix(a.Key, scorePrefix) || len(a.Key) <= len(scorePrefix) {
			continue
		}
		name := a.Key[len(scorePrefix):]
		isMin := false
		if strings.HasSuffix(name, "_min") && len(name) > len("_min") {
			isMin = true
			name = strings.TrimSuffix(name, "_min")
		}
		n, err := strconv.Atoi(a.Value)
		if err != nil {
			n = 1
		}
		f := ScoreFilter{Objective: name, Value: n, IsMin: isMin}
		if i, ok := seen[name]; ok {
			out[i] = f
			continue
		}
		seen[name] = len(out)
		out = append(out, f)
	}
	return out
}
