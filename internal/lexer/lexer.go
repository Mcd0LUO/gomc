// Package lexer splits a raw command line into tokens.
//
// 1.12.2 itself splits a command with `split(" ", 2)`, keeping everything
// after the first space intact. That is far too coarse for a parser that has
// to understand NBT and JSON arguments, so this package provides a finer
// tokenizer that still reproduces 1.12.2's observable behaviour where it
// matters:
//
//   - Whitespace separates tokens, and runs of whitespace collapse.
//   - Double-quoted and single-quoted spans are preserved as one token,
//     including the quotes, so the parser can tell a quoted string from a
//     bare one.
//   - Braces, brackets and parentheses group, so `{a:1, b:2}` stays one
//     token even when it contains spaces.
//   - The raw source of every token is recorded with its byte offset so the
//     parser can report precise file:line:column positions.
package lexer

import (
	"strings"
	"unicode/utf8"
)

// Kind classifies a token.
type Kind int

const (
	// KindWord is a bare word such as `scoreboard` or `@a`.
	KindWord Kind = iota
	// KindQuoted is a quoted string, quotes included.
	KindQuoted
	// KindBrace is a `{...}` group.
	KindBrace
	// KindBracket is a `[...]` group.
	KindBracket
)

// Token is a single lexical unit.
type Token struct {
	// Kind classifies the token.
	Kind Kind
	// Text is the token's raw source text.
	Text string
	// Offset is the byte offset of the token start within the source line.
	Offset int
	// Quoted reports whether a quoted token should be unquoted when read.
	Quoted bool
}

// Unquote returns the token's textual value with surrounding quotes removed
// and backslash escapes resolved.
func (t Token) Unquote() string {
	if t.Kind != KindQuoted || len(t.Text) < 2 {
		return t.Text
	}
	body := t.Text[1 : len(t.Text)-1]
	var sb strings.Builder
	sb.Grow(len(body))
	for i := 0; i < len(body); i++ {
		c := body[i]
		if c != '\\' || i+1 >= len(body) {
			sb.WriteByte(c)
			continue
		}
		i++
		switch body[i] {
		case 'n':
			sb.WriteByte('\n')
		case 'r':
			sb.WriteByte('\r')
		case 't':
			sb.WriteByte('\t')
		case '\\':
			sb.WriteByte('\\')
		case '"':
			sb.WriteByte('"')
		case '\'':
			sb.WriteByte('\'')
		default:
			sb.WriteByte('\\')
			sb.WriteByte(body[i])
		}
	}
	return sb.String()
}

// Tokenize splits src into tokens.
//
// It returns an error only for unterminated quotes or unbalanced groups,
// which are genuine authoring mistakes worth reporting.
func Tokenize(src string) ([]Token, error) {
	var toks []Token
	i := 0
	for i < len(src) {
		c := src[i]
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			i++
			continue
		}
		switch c {
		case '"', '\'':
			start := i
			i++
			terminated := false
			for i < len(src) {
				if src[i] == '\\' {
					i += 2
					continue
				}
				if src[i] == c {
					i++
					terminated = true
					break
				}
				i++
			}
			if !terminated {
				return nil, &Error{Offset: start, Msg: "unterminated quoted string"}
			}
			toks = append(toks, Token{Kind: KindQuoted, Text: src[start:i], Offset: start, Quoted: true})
		case '{', '[', '(':
			start := i
			// The result must be assigned back to i, not shadowed by a
			// short variable declaration, or the loop never advances.
			next, err := scanGroup(src, i)
			if err != nil {
				return nil, err
			}
			i = next
			kind := KindBrace
			if src[start] == '[' {
				kind = KindBracket
			}
			toks = append(toks, Token{Kind: kind, Text: src[start:i], Offset: start})
		default:
			start := i
			for i < len(src) {
				ch := src[i]
				if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
					break
				}
				if ch == '{' || ch == '[' || ch == '(' || ch == '"' || ch == '\'' {
					// A selector such as `@e[type=zombie]` is a single
					// argument, so a bracket group directly following
					// an `@`-prefixed word belongs to that word.
					if ch == '[' && len(src[start:i]) > 0 && src[start] == '@' {
						next, err := scanGroup(src, i)
						if err != nil {
							return nil, err
						}
						i = next
						continue
					}
					break
				}
				i++
			}
			toks = append(toks, Token{Kind: KindWord, Text: src[start:i], Offset: start})
		}
	}
	return toks, nil
}

// scanGroup returns the offset just past the group that starts at start.
func scanGroup(src string, start int) (int, error) {
	open := src[start]
	var close byte
	switch open {
	case '{':
		close = '}'
	case '[':
		close = ']'
	case '(':
		close = ')'
	default:
		return start, nil
	}
	depth := 0
	i := start
	for i < len(src) {
		switch src[i] {
		case '"', '\'':
			quote := src[i]
			i++
			for i < len(src) {
				if src[i] == '\\' {
					i += 2
					continue
				}
				if src[i] == quote {
					break
				}
				i++
			}
			if i >= len(src) {
				return 0, &Error{Offset: start, Msg: "unterminated quoted string"}
			}
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return i + 1, nil
			}
		}
		i++
	}
	return 0, &Error{Offset: start, Msg: "unbalanced " + string(open)}
}

// Error reports a lexical failure.
type Error struct {
	// Offset is the byte offset of the problem.
	Offset int
	// Msg explains the problem.
	Msg string
}

func (e *Error) Error() string { return e.Msg }

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
