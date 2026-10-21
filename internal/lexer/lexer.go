// Package lexer splits a raw command line into tokens.
//
// 1.12.2 splits a command with `String.split(" ")`, so the arguments are
// separated by **single spaces** and by nothing else:
//
//   - A quote is an ordinary character. `say "a b"` has the arguments
//     `"a` and `b"`, so the message keeps its quotes.
//   - A brace or a bracket is an ordinary character too. `setblock 1 2 3
//     {a:1}` has the argument `{a:1}`.
//   - Two spaces in a row make an **empty** argument, and the server keeps
//     it: `say  hi` says " hi" -- and, because the first argument is the
//     empty one, the server refuses the command with its usage.
//   - A tab is not a separator. Only the space is.
//
// What the server does with the arguments afterwards is a separate matter.
// The commands that take a compound or a JSON take **the rest of the
// line**: the arguments from that place on are joined with a single space
// and handed to the NBT or JSON parser (`CommandAbstract.a(String[], int)`).
// That is how a compound with spaces in it survives the split, and it is
// why the join has to use one space and not the original text.
//
// (Measured on a real 1.12.2 server: `say "a b"` prints `"a b"`, the
// quotes included, and `setblock 77 17 -202 {metadata:"",mirror:"NONE"}`
// answers "There is no such block with name {metadata:"",mirror:"none"}",
// with the name in lower case.)
package lexer

import (
	"strings"
	"unicode/utf8"
)

// Token is a single argument of a command line.
type Token struct {
	// Text is the argument's text, exactly as it was written.
	Text string
	// Offset is the byte offset of the argument within the source line.
	Offset int
}

// Tokenize splits src into the arguments of a command line.
//
// The line is split on **single spaces**. An empty argument between two
// spaces is kept, and the empty arguments at the end of the line are
// dropped -- that is what Java's `String.split(" ")` does.
func Tokenize(src string) []Token {
	// An empty line has no match, and Java keeps the line itself as the
	// one argument.
	if src == "" {
		return []Token{{Text: "", Offset: 0}}
	}
	parts := strings.Split(src, " ")
	// Java's split drops the empty arguments at the end of the line.
	end := len(parts)
	for end > 0 && parts[end-1] == "" {
		end--
	}
	parts = parts[:end]
	toks := make([]Token, 0, len(parts))
	offset := 0
	for _, part := range parts {
		toks = append(toks, Token{Text: part, Offset: offset})
		offset += len(part) + 1
	}
	return toks
}

// Column returns the 1-based column of the offset within line.
func Column(line string, offset int) int {
	if offset > len(line) {
		offset = len(line)
	}
	col := 1
	for i := 0; i < offset; {
		r, size := utf8.DecodeRuneInString(line[i:])
		if r == '\t' {
			col++
		} else {
			col++
		}
		i += size
	}
	return col
}
