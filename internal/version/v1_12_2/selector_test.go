package v1_12_2

import (
	"testing"
)

func TestParseSelectorBase(t *testing.T) {
	for _, text := range []string{"@p", "@a", "@r", "@e", "@s"} {
		sel, err := ParseSelector(text)
		if err != nil {
			t.Fatalf("ParseSelector(%q): %v", text, err)
		}
		if sel.Base != text[1:] {
			t.Errorf("base = %q, want %q", sel.Base, text[1:])
		}
		if len(sel.Args) != 0 {
			t.Errorf("%q produced %d args, want 0", text, len(sel.Args))
		}
	}
}

func TestParseSelectorArguments(t *testing.T) {
	sel, err := ParseSelector("@e[type=zombie,r=10]")
	if err != nil {
		t.Fatalf("ParseSelector: %v", err)
	}
	if len(sel.Args) != 2 {
		t.Fatalf("got %d args, want 2", len(sel.Args))
	}
	if v, ok := SelectorString(sel, "type"); !ok || v != "zombie" {
		t.Errorf("type = %q (ok=%v), want zombie", v, ok)
	}
	if v, ok, err := SelectorInt(sel, "r"); err != nil || !ok || v != 10 {
		t.Errorf("r = %d (ok=%v, err=%v), want 10", v, ok, err)
	}
}

// TestParseSelectorRejectsThirteenSyntax locks in the 1.12.2 argument set.
//
// The 1.12.2 `PlayerSelector` only accepts a fixed list of names plus the
// `score_` prefix. `scores={...}` and `nbt={...}` are 1.13 additions, and
// a harness that accepts them would let a function pass locally that fails
// on a real 1.12.2 server.
func TestParseSelectorRejectsThirteenSyntax(t *testing.T) {
	for _, text := range []string{
		"@e[scores={kills=1}]",
		"@e[nbt={OnGround:1b}]",
		"@e[limit=1]",
		"@e[sort=random]",
		"@e[hasitem={item:stone}]",
	} {
		if _, err := ParseSelector(text); err == nil {
			t.Errorf("ParseSelector(%q) succeeded, want an error", text)
		}
	}
}

func TestParseSelectorAcceptsScorePrefix(t *testing.T) {
	sel, err := ParseSelector("@e[score_kills=5]")
	if err != nil {
		t.Fatalf("ParseSelector: %v", err)
	}
	filters := SelectorScoreFilters(sel)
	if len(filters) != 1 {
		t.Fatalf("got %d score filters, want 1", len(filters))
	}
	if filters[0].Objective != "kills" || filters[0].Value != 5 || filters[0].IsMin {
		t.Errorf("filter = %+v, want kills=5 as a maximum", filters[0])
	}
}

func TestParseSelectorAcceptsScoreMin(t *testing.T) {
	sel, err := ParseSelector("@e[score_kills_min=3]")
	if err != nil {
		t.Fatalf("ParseSelector: %v", err)
	}
	filters := SelectorScoreFilters(sel)
	if len(filters) != 1 {
		t.Fatalf("got %d score filters, want 1", len(filters))
	}
	if filters[0].Objective != "kills" || filters[0].Value != 3 || !filters[0].IsMin {
		t.Errorf("filter = %+v, want kills=3 as a minimum", filters[0])
	}
}

func TestParseSelectorScoreDefaultValue(t *testing.T) {
	// A score argument without a value counts as 1, the way
	// `MathHelper.a(String, int)` falls back to its default.
	sel, err := ParseSelector("@e[score_kills]")
	if err != nil {
		t.Fatalf("ParseSelector: %v", err)
	}
	filters := SelectorScoreFilters(sel)
	if len(filters) != 1 || filters[0].Value != 1 {
		t.Errorf("filters = %+v, want a single filter with value 1", filters)
	}
}

// TestParseSelectorLastArgumentWins matches 1.12.2's flat argument map.
//
// The server stores the arguments in a map, so a repeated name keeps the
// last value: `@e[tag=a,tag=b]` behaves like `@e[tag=b]`.
func TestParseSelectorLastArgumentWins(t *testing.T) {
	sel, err := ParseSelector("@e[tag=a,tag=b]")
	if err != nil {
		t.Fatalf("ParseSelector: %v", err)
	}
	if v, ok := SelectorString(sel, "tag"); !ok || v != "b" {
		t.Errorf("tag = %q (ok=%v), want b", v, ok)
	}
}

func TestParseSelectorRejectsUnknownArgument(t *testing.T) {
	if _, err := ParseSelector("@e[nosuchthing=1]"); err == nil {
		t.Fatal("expected an error for an unknown selector argument")
	}
}

func TestParseSelectorRejectsMalformed(t *testing.T) {
	for _, text := range []string{
		"@x",
		"@e[",
		"@e]",
		"@e[type=zombie] extra",
		"@e[type=zombie, name=x]",
		"notaselector",
		"",
	} {
		if _, err := ParseSelector(text); err == nil {
			t.Errorf("ParseSelector(%q) succeeded, want an error", text)
		}
	}
}

func TestParseSelectorEmptyArgumentList(t *testing.T) {
	sel, err := ParseSelector("@e[]")
	if err != nil {
		t.Fatalf("ParseSelector: %v", err)
	}
	if sel.Base != "e" || len(sel.Args) != 0 {
		t.Errorf("selector = %+v, want base e with no arguments", sel)
	}
}

func TestIsSelector(t *testing.T) {
	cases := map[string]bool{
		"@e":                   true,
		"@e[type=zombie]":      true,
		"@x":                   false,
		"Steve":                false,
		"@e[type=zombie] x":    false,
		"@e[type=zombie, x=1]": false,
	}
	for text, want := range cases {
		if got := IsSelector(text); got != want {
			t.Errorf("IsSelector(%q) = %v, want %v", text, got, want)
		}
	}
}

// TestSelectorIntDefaultFallsBack covers the numeric selector arguments.
//
// 1.12.2 reads them with `MathHelper.a(String, int)`, which is
// `Integer.parseInt` with the default as the fallback. A value that is not
// a Java `int` is therefore not a number at all: the argument falls back
// to its default instead of raising an error.
//
// Measured on a real 1.12.2 server: `say @e[c=abc]` expands to every
// entity, the same as `say @e` — the count is not an error.
func TestSelectorIntDefaultFallsBack(t *testing.T) {
	sel, err := ParseSelector("@e[c=abc]")
	if err != nil {
		t.Fatalf("ParseSelector: %v", err)
	}
	if got := SelectorIntDefault(sel, "c", 0); got != 0 {
		t.Errorf("c = %d, want the default 0", got)
	}
	if got := SelectorIntDefault(sel, "c", 1); got != 1 {
		t.Errorf("c = %d, want the default 1", got)
	}

	// A value that is not a number at all falls back as well.
	sel, err = ParseSelector("@e[c=1.0]")
	if err != nil {
		t.Fatalf("ParseSelector: %v", err)
	}
	if got := SelectorIntDefault(sel, "c", 0); got != 0 {
		t.Errorf("c = %d, want the default 0", got)
	}

	// A value that does not fit into a Java `int` falls back too.
	sel, err = ParseSelector("@e[c=99999999999999]")
	if err != nil {
		t.Fatalf("ParseSelector: %v", err)
	}
	if got := SelectorIntDefault(sel, "c", 0); got != 0 {
		t.Errorf("c = %d, want the default 0", got)
	}

	// A count that is an integer is used as it is.
	sel, err = ParseSelector("@e[c=2]")
	if err != nil {
		t.Fatalf("ParseSelector: %v", err)
	}
	if got := SelectorIntDefault(sel, "c", 0); got != 2 {
		t.Errorf("c = %d, want 2", got)
	}
}
