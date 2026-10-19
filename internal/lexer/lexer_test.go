package lexer

import (
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
	toks, err := Tokenize("scoreboard players set Steve counter 1")
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	want := []string{"scoreboard", "players", "set", "Steve", "counter", "1"}
	if got := tokenTexts(toks); !equal(got, want) {
		t.Errorf("tokens = %v, want %v", got, want)
	}
}

func TestTokenizeCollapsesWhitespace(t *testing.T) {
	toks, err := Tokenize("  say \t hello \r\n world  ")
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	want := []string{"say", "hello", "world"}
	if got := tokenTexts(toks); !equal(got, want) {
		t.Errorf("tokens = %v, want %v", got, want)
	}
}

func TestTokenizeQuotedString(t *testing.T) {
	toks, err := Tokenize(`say "hello world"`)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	if len(toks) != 2 {
		t.Fatalf("got %d tokens, want 2", len(toks))
	}
	if toks[1].Kind != KindQuoted {
		t.Errorf("kind = %v, want KindQuoted", toks[1].Kind)
	}
	if got := toks[1].Unquote(); got != "hello world" {
		t.Errorf("unquoted = %q, want %q", got, "hello world")
	}
}

func TestTokenizeKeepsSelectorTogether(t *testing.T) {
	// A selector is a single argument, so its bracket group must not be
	// split off from the `@` prefix.
	toks, err := Tokenize("@e[type=armor_stand]")
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	if len(toks) != 1 {
		t.Fatalf("got %d tokens (%v), want 1", len(toks), tokenTexts(toks))
	}
	if toks[0].Text != "@e[type=armor_stand]" {
		t.Errorf("token = %q, want %q", toks[0].Text, "@e[type=armor_stand]")
	}
}

func TestTokenizeKeepsNBTGroupTogether(t *testing.T) {
	toks, err := Tokenize("summon zombie ~ ~ ~ {Health:20.0f,Tags:[\"a\"]}")
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	want := []string{"summon", "zombie", "~", "~", "~", `{Health:20.0f,Tags:["a"]}`}
	if got := tokenTexts(toks); !equal(got, want) {
		t.Errorf("tokens = %v, want %v", got, want)
	}
}

func TestTokenizeUnterminatedQuote(t *testing.T) {
	if _, err := Tokenize(`say "unterminated`); err == nil {
		t.Fatal("expected an error for an unterminated quote")
	}
}

func TestTokenizeUnbalancedGroup(t *testing.T) {
	if _, err := Tokenize("summon zombie {a:1"); err == nil {
		t.Fatal("expected an error for an unbalanced group")
	}
}

// TestTokenizeOffsetsAdvance guards against the lexer failing to make
// progress.
//
// An earlier version shadowed the loop index while scanning a group, which
// produced the same token forever and hung the process. The structural
// check below turns that into a failure instead of a hang, and the
// watchdog covers the case where the loop never returns at all.
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
	}
	for _, in := range inputs {
		toks, err := Tokenize(in)
		if err != nil {
			continue
		}
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
	}
	for _, in := range inputs {
		in := in
		done := make(chan struct{})
		go func() {
			defer close(done)
			for i := 0; i < 200; i++ {
				_, _ = Tokenize(in)
			}
		}()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatalf("Tokenize did not terminate for %q", in)
		}
	}
}

func TestUnquoteResolvesEscapes(t *testing.T) {
	tok := Token{Kind: KindQuoted, Text: `"a\nb\"c\\d"`}
	want := "a\nb\"c\\d"
	if got := tok.Unquote(); got != want {
		t.Errorf("Unquote = %q, want %q", got, want)
	}
}

func TestColumnCountsRunes(t *testing.T) {
	// The column is a rune count, not a byte count, so a multi-byte
	// character still points at the right place.
	if got := Column("say 日本語", 4); got != 5 {
		t.Errorf("Column = %d, want 5", got)
	}
}
