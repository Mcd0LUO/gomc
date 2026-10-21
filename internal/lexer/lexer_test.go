package lexer

import (
	"strings"
	"testing"
	"time"
)

// tokenTexts returns the raw text of every token.
func tokenTexts(toks []Token) []string {
	out := make([]string, 0, len(toks))
	for _, t := range toks {
		out = append(out, t.Text)
	}
	return out
}

// equal compares two string slices.
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestTokenizeWords(t *testing.T) {
	toks := Tokenize("scoreboard players set Steve counter 1")
	want := []string{"scoreboard", "players", "set", "Steve", "counter", "1"}
	if got := tokenTexts(toks); !equal(got, want) {
		t.Errorf("tokens = %v, want %v", got, want)
	}
}

// TestTokenizeSplitsOnSingleSpace covers the rule the whole package is
// built on: 1.12.2 splits a command with `String.split(" ")`, so two
// spaces in a row make an **empty** argument, and the server keeps it.
//
// (Measured on a real 1.12.2 server: `say  hi` -- two spaces -- answers
// with the usage of the command, because the first argument is the empty
// one and `CommandSay.execute` refuses an empty first argument.)
func TestTokenizeSplitsOnSingleSpace(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"say hello", []string{"say", "hello"}},
		{"say  hello", []string{"say", "", "hello"}},
		{"say   hello", []string{"say", "", "", "hello"}},
		{"say hello ", []string{"say", "hello"}},
		{"say hello  ", []string{"say", "hello"}},
		{"say  ", []string{"say"}},
		{"", []string{""}},
	}
	for _, tc := range tests {
		if got := tokenTexts(Tokenize(tc.in)); !equal(got, tc.want) {
			t.Errorf("input %q: tokens = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// TestTokenizeTabIsNotASeparator guards the one separator the server
// does not know.
//
// 1.12.2 splits on a single space and on nothing else, so a tab is an
// ordinary character that belongs to the argument.
func TestTokenizeTabIsNotASeparator(t *testing.T) {
	toks := Tokenize("say\thello")
	want := []string{"say\thello"}
	if got := tokenTexts(toks); !equal(got, want) {
		t.Errorf("tokens = %v, want %v", got, want)
	}
}

// TestTokenizeKeepsQuotes covers the rule that a quote is an ordinary
// character.
//
// (Measured on a real 1.12.2 server: `say "a b"` prints `"a b"`, the
// quotes included, and `say 'c d'` prints `'c d'`.)
func TestTokenizeKeepsQuotes(t *testing.T) {
	for _, in := range []string{`say "a b"`, `say 'c d'`, `say "unterminated`} {
		toks := Tokenize(in)
		want := strings.Split(in, " ")
		if got := tokenTexts(toks); !equal(got, want) {
			t.Errorf("input %q: tokens = %v, want %v", in, got, want)
		}
	}
}

func TestTokenizeKeepsSelectorTogether(t *testing.T) {
	// A selector is a single argument, so its bracket group must not be
	// split off from the `@` prefix.
	toks := Tokenize("@e[type=armor_stand]")
	if len(toks) != 1 {
		t.Fatalf("got %d tokens (%v), want 1", len(toks), tokenTexts(toks))
	}
	if toks[0].Text != "@e[type=armor_stand]" {
		t.Errorf("token = %q, want %q", toks[0].Text, "@e[type=armor_stand]")
	}
}

// TestTokenizeKeepsBracketGroupWithTheWord covers a bracket group that
// follows a word.
//
// 1.12.2 splits the command on spaces, so `vec3[0]` is **one** argument --
// the name of a scoreboard objective, say. The group belongs to the word
// whenever there is no space in between.
func TestTokenizeKeepsBracketGroupWithTheWord(t *testing.T) {
	toks := Tokenize("scoreboard players set @s vec3[0] 10000")
	want := []string{"scoreboard", "players", "set", "@s", "vec3[0]", "10000"}
	if len(toks) != len(want) {
		t.Fatalf("got %d tokens (%v), want %d", len(toks), tokenTexts(toks), len(want))
	}
	for i, w := range want {
		if toks[i].Text != w {
			t.Errorf("token %d = %q, want %q", i, toks[i].Text, w)
		}
	}
}

// TestTokenizeSplitsNBTOnSpaces covers the compound that has a space in
// it.
//
// 1.12.2 has no compound of its own: it splits the command on spaces and
// hands the arguments to the NBT parser, which reads them back as one
// text. The compound is therefore split here as well, and the parser
// joins the arguments again when it reads the rest of the line.
func TestTokenizeSplitsNBTOnSpaces(t *testing.T) {
	toks := Tokenize("entitydata @e[type=armor_stand] {Tags:1, b:2")
	want := []string{"entitydata", "@e[type=armor_stand]", "{Tags:1,", "b:2"}
	if got := tokenTexts(toks); !equal(got, want) {
		t.Errorf("tokens = %v, want %v", got, want)
	}
}

// TestTokenizeKeepsCompoundWithoutSpaces covers the compound that has no
// space in it: it is one argument.
func TestTokenizeKeepsCompoundWithoutSpaces(t *testing.T) {
	toks := Tokenize(`summon zombie ~ ~ ~ {Health:20.0f,Tags:["a"]}`)
	want := []string{"summon", "zombie", "~", "~", "~", `{Health:20.0f,Tags:["a"]}`}
	if got := tokenTexts(toks); !equal(got, want) {
		t.Errorf("tokens = %v, want %v", got, want)
	}
}

// TestTokenizeUnbalancedGroup covers the brackets and the braces that are
// not closed.
//
// 1.12.2 does not group them either: it splits the command on spaces. A
// bracket that is never closed therefore is ordinary text -- and a
// selector that is not closed is not a selector either, because
// `PlayerSelector.isPattern` wants balanced brackets.
//
// (Measured on a real 1.12.2 server: `say @e[type=armor_stand` prints the
// text as it is, and `scoreboard players set @e[type=armor_stand t 1` takes
// the holder for the name "@e[type=armor_stand".)
func TestTokenizeUnbalancedGroup(t *testing.T) {
	for _, in := range []string{
		"scoreboard players set @e[type=armor_stand t 1",
		"summon zombie (0 4 0",
		"say (hello",
		"say hello)",
		"say [hello",
		"summon zombie {a:1",
		"summon zombie {Tags:\"abc",
	} {
		toks := Tokenize(in)
		// No character is lost: the tokens carry the whole text, the
		// bracket and the parenthesis among them.
		var b strings.Builder
		for _, tok := range toks {
			b.WriteString(tok.Text)
		}
		if got, want := b.String(), strings.ReplaceAll(in, " ", ""); got != want {
			t.Errorf("input %q: the tokens carry %q, want %q", in, got, want)
		}
	}
}

// TestTokenizeOffsetsAdvance guards against the lexer failing to make
// progress.
//
// The offsets are what the parser uses to point at the place in the file
// a mistake was made, so they have to advance with every argument.
func TestTokenizeOffsetsAdvance(t *testing.T) {
	inputs := []string{
		"",
		"say hello",
		`say "quoted"`,
		"{a:1}",
		"[a,b]",
		"(x)",
		"@e[type=zombie]",
		"{}",
		"[]",
		"()",
		"a{b}c",
		`"{"`,
		"{a:{b:1}}",
		"@e[type=zombie,name=\"a b\"]",
		"~~~",
		"{{{{",
		"}}}}",
		`\"`,
		"say \\",
		"say  hello",
	}
	for _, in := range inputs {
		toks := Tokenize(in)
		prev := -1
		for _, tok := range toks {
			if tok.Offset <= prev {
				t.Fatalf("input %q: token offsets did not advance: %v", in, toks)
			}
			prev = tok.Offset
		}
	}
}

// TestTokenizeTerminates is the hang guard.
//
// Every input is tokenized on a separate goroutine so a lexer that never
// returns is reported as a failure rather than wedging the test run.
func TestTokenizeTerminates(t *testing.T) {
	inputs := []string{
		"", "say hello", `say "unterminated`, "{unbalanced", "}}}}", "@e[",
		`"`, "{", "((((((", "@e[type=]", "a b c d e f g",
		"{unbalanced", `{Tags:"abc`, "{a:1 b:2", "{a:{b:1}", `{a:"b}`,
	}
	for _, in := range inputs {
		in := in
		done := make(chan struct{})
		go func() {
			defer close(done)
			for i := 0; i < 200; i++ {
				Tokenize(in)
			}
		}()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatalf("Tokenize did not terminate for %q", in)
		}
	}
}

func TestColumnCountsRunes(t *testing.T) {
	// The column is a rune count, not a byte count, so a multi-byte
	// character still points at the right place.
	if got := Column("say 日本語", 4); got != 5 {
		t.Errorf("Column = %d, want 5", got)
	}
}
