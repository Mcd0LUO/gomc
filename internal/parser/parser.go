// Package parser turns a tokenized command line into an AST.
//
// The package provides version-independent parsing primitives. The rules
// that differ between Minecraft versions (such as the 1.12.2 `execute`
// syntax) live in the version packages, which build on these primitives.
package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/lexer"
	"github.com/Mcd0LUO/gomc/internal/nbt"
)

// Kind classifies a parse error.
type Kind int

const (
	// KindSyntax is a malformed argument.
	KindSyntax Kind = iota
	// KindUnknown is an unknown command or subcommand.
	KindUnknown
	// KindRuntime is a failure that only appears when the command runs.
	KindRuntime
)

func (k Kind) String() string {
	switch k {
	case KindSyntax:
		return "Syntax"
	case KindUnknown:
		return "Unknown"
	case KindRuntime:
		return "Runtime"
	}
	return "Error"
}

// Error is a parse or execution error carrying a source position.
//
// The first five fields are what 1.12.2 would say. The rest is gomc's own
// addition: they carry the context needed to point at the problem in the
// file, and they are only rendered when the enhanced output is on.
type Error struct {
	// Kind classifies the error.
	Kind Kind
	// Msg explains the error.
	Msg string
	// Line is the 1-based line number, or 0 when unknown.
	Line int
	// Col is the 1-based column number, or 0 when unknown.
	Col int
	// Offset is the byte offset within the line.
	Offset int
	// Path is the file the command came from, when known.
	Path string
	// Source is the text of the offending line, when known.
	Source string
	// Usage is the usage line of the command that failed, when known.
	Usage string
	// Hint is an extra suggestion, such as "did you mean 'setblock'?".
	Hint string
}

func (e *Error) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s:%d:%d: %s", e.Kind, e.Line, e.Col, e.Msg)
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Msg)
}

// Positioned wraps an error with a source position.
func Positioned(err error, line, col, offset int) error {
	if err == nil {
		return nil
	}
	var pe *Error
	if e, ok := err.(*Error); ok {
		pe = e
	} else {
		pe = &Error{Kind: KindSyntax, Msg: err.Error()}
	}
	if pe.Line == 0 {
		pe.Line = line
		pe.Col = col
		pe.Offset = offset
	}
	return pe
}

// Parser reads arguments from a token list.
type Parser struct {
	toks []lexer.Token
	pos  int
	// line is the source line the tokens came from.
	line int
	// src is the original text of the source line, used to map byte
	// offsets to columns.
	src string
}

// New creates a parser over the given tokens.
func New(toks []lexer.Token, line int, src string) *Parser {
	return &Parser{toks: toks, line: line, src: src}
}

// Line returns the source line the tokens came from.
func (p *Parser) Line() int { return p.line }

// More reports whether unconsumed tokens remain.
func (p *Parser) More() bool { return p.pos < len(p.toks) }

// Peek returns the next token without consuming it.
func (p *Parser) Peek() (lexer.Token, bool) {
	if p.pos >= len(p.toks) {
		return lexer.Token{}, false
	}
	return p.toks[p.pos], true
}

// Next consumes and returns the next token.
func (p *Parser) Next() (lexer.Token, bool) {
	if p.pos >= len(p.toks) {
		return lexer.Token{}, false
	}
	t := p.toks[p.pos]
	p.pos++
	return t, true
}

// Pos returns the source position of the next token.
func (p *Parser) Pos() (line, col, offset int) {
	t, ok := p.Peek()
	if !ok {
		return p.line, 1, len(p.src)
	}
	return p.line, p.columnOf(t), t.Offset
}

// Errorf builds a positioned syntax error at the current position.
func (p *Parser) Errorf(format string, args ...any) error {
	line, col, offset := p.Pos()
	return &Error{Kind: KindSyntax, Msg: fmt.Sprintf(format, args...), Line: line, Col: col, Offset: offset}
}

// ErrorAt builds a positioned error at a specific token.
func (p *Parser) ErrorAt(t lexer.Token, format string, args ...any) error {
	return &Error{Kind: KindSyntax, Msg: fmt.Sprintf(format, args...), Line: p.line, Col: p.columnOf(t), Offset: t.Offset}
}

// columnOf computes the 1-based column of a token within the source line.
func (p *Parser) columnOf(t lexer.Token) int {
	return lexer.Column(p.src, t.Offset)
}

// Remaining returns the unconsumed tokens' source text, rejoined by spaces.
func (p *Parser) Remaining() string {
	parts := make([]string, 0, len(p.toks)-p.pos)
	for _, t := range p.toks[p.pos:] {
		parts = append(parts, t.Text)
	}
	return strings.Join(parts, " ")
}

// Rest consumes all remaining tokens and returns their joined text.
func (p *Parser) Rest() string {
	s := p.Remaining()
	p.pos = len(p.toks)
	return s
}

// String reads the next token as a string.
func (p *Parser) String() (string, error) {
	t, ok := p.Next()
	if !ok {
		return "", p.Errorf("expected a value but reached end of command")
	}
	return t.Unquote(), nil
}

// Literal reads the next token and requires it to equal want.
func (p *Parser) Literal(want string) error {
	t, ok := p.Next()
	if !ok {
		return p.Errorf("expected %q but reached end of command", want)
	}
	if t.Text != want {
		return p.ErrorAt(t, "expected %q, got %q", want, t.Text)
	}
	return nil
}

// OptionalLiteral consumes the next token when it equals want.
func (p *Parser) OptionalLiteral(want string) bool {
	t, ok := p.Peek()
	if !ok || t.Text != want {
		return false
	}
	p.pos++
	return true
}

// Int reads the next token as an integer.
// Int reads the next token as a number.
func (p *Parser) Int() (int, error) {
	t, ok := p.Next()
	if !ok {
		return 0, p.Errorf("expected an integer but reached end of command")
	}
	n, ok := JavaInt(strings.TrimSpace(t.Text))
	if !ok {
		return 0, p.ErrorAt(t, "'%s' is not a valid number", t.Text)
	}
	return n, nil
}

// JavaInt parses a decimal integer the way 1.12.2 does.
//
// The server parses numbers with `Integer.parseInt`, so a value outside the
// signed 32-bit range is not a number at all: it makes the command fail
// with "is not a valid number" instead of silently wrapping.
func JavaInt(s string) (int, bool) {
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 32)
	if err != nil {
		return 0, false
	}
	return int(n), true
}

// Float reads the next token as a number.
func (p *Parser) Float() (float64, error) {
	t, ok := p.Next()
	if !ok {
		return 0, p.Errorf("expected a number but reached end of command")
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(t.Text), 64)
	if err != nil {
		return 0, p.ErrorAt(t, "expected a number, got %q", t.Text)
	}
	return f, nil
}

// Coord reads a coordinate triple, resolving relative parts against base.
func (p *Parser) Coord(base ast.Coord) (*ast.Coord, error) {
	x, rx, err := p.coordComponent()
	if err != nil {
		return nil, err
	}
	y, ry, err := p.coordComponent()
	if err != nil {
		return nil, err
	}
	z, rz, err := p.coordComponent()
	if err != nil {
		return nil, err
	}
	out := &ast.Coord{RelativeX: rx, RelativeY: ry, RelativeZ: rz}
	if rx {
		out.X = base.X + x
	} else {
		out.X = x
	}
	if ry {
		out.Y = base.Y + y
	} else {
		out.Y = y
	}
	if rz {
		out.Z = base.Z + z
	} else {
		out.Z = z
	}
	return out, nil
}

// coordComponent parses one coordinate component, reporting whether it was
// relative (prefixed with '~').
func (p *Parser) coordComponent() (float64, bool, error) {
	t, ok := p.Next()
	if !ok {
		return 0, false, p.Errorf("expected a coordinate but reached end of command")
	}
	raw := t.Text
	rel := strings.HasPrefix(raw, "~")
	if rel {
		raw = raw[1:]
	}
	if raw == "" {
		return 0, rel, nil
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, rel, p.ErrorAt(t, "expected a coordinate, got %q", t.Text)
	}
	return f, rel, nil
}

// NBT reads the next token as an NBT compound.
func (p *Parser) NBT() (*ast.NBTArg, error) {
	t, ok := p.Next()
	if !ok {
		return nil, p.Errorf("expected NBT data but reached end of command")
	}
	c, err := nbt.ParseSNBT(t.Text)
	if err != nil {
		return nil, p.ErrorAt(t, "invalid NBT: %v", err)
	}
	comp, ok := c.(*nbt.Compound)
	if !ok {
		return nil, p.ErrorAt(t, "expected an NBT compound")
	}
	return &ast.NBTArg{Value: comp, Offset: t.Offset}, nil
}

// JSON reads the next token as raw JSON.
func (p *Parser) JSON() (*ast.JSONArg, error) {
	t, ok := p.Next()
	if !ok {
		return nil, p.Errorf("expected JSON but reached end of command")
	}
	return &ast.JSONArg{Value: t.Text, Offset: t.Offset}, nil
}

// Selector reads the next token as a target selector.
func (p *Parser) Selector() (*ast.Selector, error) {
	t, ok := p.Next()
	if !ok {
		return nil, p.Errorf("expected a target but reached end of command")
	}
	sel, err := ParseSelector(t.Text)
	if err != nil {
		return nil, p.ErrorAt(t, "%v", err)
	}
	sel.Offset = t.Offset
	return sel, nil
}
