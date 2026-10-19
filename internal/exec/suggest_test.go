package exec

import "testing"

func TestClosestName(t *testing.T) {
	candidates := []string{"setblock", "fill", "summon", "scoreboard", "execute"}
	cases := []struct {
		name string
		want string
		ok   bool
	}{
		{"setblok", "setblock", true},
		{"setbloc", "setblock", true},
		{"summonn", "summon", true},
		{"scorebord", "scoreboard", true},
		// Nothing is close enough to be worth suggesting.
		{"xyzzy", "", false},
		// Too short to correct.
		{"ab", "", false},
	}
	for _, c := range cases {
		got, ok := ClosestName(c.name, candidates)
		if ok != c.ok || got != c.want {
			t.Errorf("ClosestName(%q) = %q, %v; want %q, %v", c.name, got, ok, c.want, c.ok)
		}
	}
}

func TestEditDistance(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"kitten", "sitting", 3},
		{"scoreboard", "scorebord", 1},
	}
	for _, c := range cases {
		if got := editDistance(c.a, c.b); got != c.want {
			t.Errorf("editDistance(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
