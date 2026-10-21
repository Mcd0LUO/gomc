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
//   - A brace group that never closes is not a lexical error. 1.12.2 has no
//     brace group at all: it splits the command on spaces and hands every
//     remaining argument to its NBT parser, which reports a broken compound
//     in the wording of the server. Such a group therefore runs to the end
//     of the line and is passed on as it is.
//   - A bracket or a parenthesis that never closes is **not** a group
//     either: 1.12.2 splits the command on spaces, so such a character is
//     ordinary text and belongs to the argument it is in.
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
// It returns an error only for an unterminated quoted string, which is a
// genuine authoring mistake worth reporting. A group that does not close is
// not one of them: see scanGroup.
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
			if next == start {
				// The character opens no group: 1.12.2 does not group
				// brackets and parentheses, it splits the command on
				// spaces. So the character and the text after it are
				// one argument.
				j := i + 1
				for j < len(src) && !isSpace(src[j]) {
					j++
				}
				toks = append(toks, Token{Kind: KindWord, Text: src[start:j], Offset: start})
				i = j
				continue
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
				if isSpace(ch) {
					break
				}
				if ch == '{' || ch == '[' || ch == '(' || ch == '"' || ch == '\'' {
					if ch == '[' || ch == '(' {
						// A bracket or a parenthesis that is never closed
						// is ordinary text: 1.12.2 does not group them.
						// So the character belongs to this word.
						next, err := scanGroup(src, i)
						if err != nil {
							return nil, err
						}
						if next == i {
							i++
							continue
						}
						// A group that follows a word without a space
						// belongs to that word: 1.12.2 splits the command
						// on spaces, so `vec3[0]` is **one** argument --
						// the name of a scoreboard objective, say -- and a
						// selector such as `@e[type=zombie]` is one too.
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

// isSpace reports whether c separates two arguments.
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

// scanGroup returns the offset just past the group that starts at start.
//
// A brace group that never closes is not an error. 1.12.2 does not group
// braces at all: it splits the command on spaces and hands every remaining
// argument to its NBT parser, which reports a broken compound in the
// wording of the server. The group therefore runs to the end of the line
// and is handed over as it is, and the NBT parser is the one that
// complains.
//
// (Measured on a real 1.12.2 server: `scoreboard players set @e[type=armor_stand] t 1 {Tags`
// answers "Could not parse dataTag, reason: Expected ':' but got '<EOF>' at: {Tags<--[HERE]",
// and the same command with `{Tags:"abc` answers
// "Could not parse dataTag, reason: Missing termination quote at: {Tags:\"abc<--[HERE]".)
//
// A quote inside such a group belongs to the NBT, so a quote that is never
// closed is reported by the NBT parser as well and does not fail here.
//
// A bracket or a parenthesis that never closes is **not** a group: 1.12.2 does
// not group those either, it splits the command on spaces. A selector that is
// not closed is not a selector either -- `PlayerSelector.isPattern` wants
// balanced brackets -- so the text is a plain name.
//
// (Measured on a real 1.12.2 server: `say @e[type=armor_stand` prints the
// text as it is, `scoreboard players set @e[type=armor_stand t 1` takes the
// holder for the name "@e[type=armor_stand", and `say (hello` prints
// "(hello".)
//
// The caller gets the start back and treats the character as ordinary text.
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
	// A brace group belongs to the NBT, so it may run to the end of the
	// line instead of failing.
	nbtContext := open == '{'
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
				// The quote is never closed. Inside a brace group the NBT
				// parser reports it, everywhere else it is a mistake of
				// the author.
				if nbtContext {
					return len(src), nil
				}
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
	if nbtContext {
		// The brace is never closed. 1.12.2 hands the rest of the command
		// to its NBT parser, so the group ends with the line.
		return len(src), nil
	}
	// The bracket or the parenthesis is never closed: it is ordinary text.
	return start, nil
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
