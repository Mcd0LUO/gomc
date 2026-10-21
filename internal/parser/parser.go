// Package parser turns a tokenized command line into an AST.
//
// The package provides version-independent parsing primitives. The rules
// that differ between Minecraft versions (such as the 1.12.2 `execute`
// syntax) live in the version packages, which build on these primitives.
package parser

import (
	"fmt"
	"math"
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
// The message is what the server would say. The rest is gomc's own
// addition: it carries the context needed to point at the problem in the
// file, and it is only rendered when the enhanced output is on.
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
	// base is the position a relative coordinate is measured from.
	//
	// 1.12.2 reads a relative coordinate against the **block position**
	// the command runs at (`ICommandListener.getChunkCoordinates()`),
	// and the bounds of the coordinate are checked on the sum of the
	// base and the offset.
	base CoordBase
}

// CoordBase is the position a relative coordinate is measured from.
//
// The three numbers are the block position the command runs at: 1.12.2
// measures a relative coordinate against the block position of the
// listener, so the base is a whole number.
type CoordBase struct {
	// X, Y and Z are the coordinates of the base.
	X, Y, Z float64
}

// New creates a parser over the given tokens.
//
// A relative coordinate is measured from the origin.
func New(toks []lexer.Token, line int, src string) *Parser {
	return &Parser{toks: toks, line: line, src: src}
}

// NewAt creates a parser over the given tokens, with the position a
// relative coordinate is measured from.
//
// The base is the block position the command runs at.
func NewAt(toks []lexer.Token, line int, src string, base CoordBase) *Parser {
	return &Parser{toks: toks, line: line, src: src, base: base}
}

// Line returns the source line the tokens came from.
func (p *Parser) Line() int { return p.line }

// More reports whether unconsumed tokens remain.
func (p *Parser) More() bool { return p.pos < len(p.toks) }

// Count returns the number of tokens that are left.
//
// It is how a command counts its arguments, the way the server counts the
// strings of a command line: some commands look at the number before they
// look at the arguments themselves.
func (p *Parser) Count() int { return len(p.toks) - p.pos }

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
//
// The text is taken as it is: 1.12.2 has no quoting of its own, so a
// quote is an ordinary character that belongs to the argument.
func (p *Parser) String() (string, error) {
	t, ok := p.Next()
	if !ok {
		return "", p.Errorf("expected a value but reached end of command")
	}
	return t.Text, nil
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

// JavaInt parses a decimal integer the way Java's `Integer.parseInt` does.
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

// JavaDouble reads a number the way Java's `Double.parseDouble` does.
//
// Minecraft reads every number of a command with it, so the accepted
// shapes are Java's: a trailing `d`, `D`, `f` or `F` is allowed, and the
// result has to be finite.
//
// It reports whether the text is a number at all.
func JavaDouble(text string) (float64, bool) {
	s := text
	if len(s) > 0 {
		switch s[len(s)-1] {
		case 'd', 'D', 'f', 'F':
			// Java allows the suffix, but only after a number.
			if _, err := strconv.ParseFloat(s[:len(s)-1], 64); err == nil {
				s = s[:len(s)-1]
			}
		}
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return 0, false
	}
	return v, true
}

// CoordAxis says how one axis of a coordinate is read.
type CoordAxis struct {
	// Min and Max are the bounds of the axis.
	Min, Max float64
	// Center says that a whole number names the middle of a block.
	//
	// The number then gets 0.5 added to it, and the message of a
	// coordinate that is out of bounds shows the number with the 0.5.
	//
	// Which bounds an axis gets, and whether it is centred, is up to
	// the version: the specs live in the version packages.
	Center bool
}

// CoordSpec says how the three axes of a coordinate are read.
//
// A spec is the version's answer to "how does this command read a
// coordinate": the bounds of each axis, and whether a whole number names
// the middle of a block. Every command that reads a coordinate picks one
// of the specs its version defines.
type CoordSpec struct {
	// X, Y and Z are the specs of the three axes.
	X, Y, Z CoordAxis
}

// CoordWith reads a coordinate triple with the given spec.
//
// A relative part stays an **offset** in the result: the base it is
// measured from is only known when the command runs, so the caller
// resolves it then.
func (p *Parser) CoordWith(spec CoordSpec) (*ast.Coord, error) {
	x, rx, err := p.coordComponent(spec.X, p.base.X)
	if err != nil {
		return nil, err
	}
	y, ry, err := p.coordComponent(spec.Y, p.base.Y)
	if err != nil {
		return nil, err
	}
	z, rz, err := p.coordComponent(spec.Z, p.base.Z)
	if err != nil {
		return nil, err
	}
	return &ast.Coord{X: x, Y: y, Z: z, RelativeX: rx, RelativeY: ry, RelativeZ: rz}, nil
}

// coordComponent parses one coordinate component, reporting whether it was
// relative (prefixed with '~') and whether it is within the bounds the
// spec puts on that axis.
//
// base is the position a relative component is measured from.
func (p *Parser) coordComponent(axis CoordAxis, base float64) (float64, bool, error) {
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
	f, ok := JavaDouble(raw)
	if !ok {
		// A coordinate is read with `parseDouble`, which reports
		// `commands.generic.num.invalid` when the text is not a number.
		//
		// (Measured on a real 1.12.2 server: `setblock a 4 0 stone`
		// answers "'a' is not a valid number".)
		//
		// A coordinate with a fraction is **not** refused: the server
		// takes the whole number of it, and so does the runtime.
		return 0, rel, p.ErrorAt(t, "'%s' is not a valid number", t.Text)
	}
	// A whole number in the "centre" mode names the middle of a block,
	// so 0.5 is added to it -- and the message of a coordinate that is
	// out of bounds shows the number with the 0.5.
	if axis.Center && !rel && !strings.Contains(raw, ".") {
		f += 0.5
	}
	// A relative coordinate is measured from the position the command
	// runs at, and 1.12.2 checks the bounds of the **sum**: the message
	// of a coordinate that is out of bounds shows the sum, not the
	// offset.
	value := f
	if rel {
		value += base
	}
	if value < axis.Min {
		return 0, rel, p.ErrorAt(t, "The number you have entered (%s) is too small, it must be at least %s",
			fmt.Sprintf("%.2f", value), formatCoordBound(axis.Min))
	}
	if value > axis.Max {
		return 0, rel, p.ErrorAt(t, "The number you have entered (%s) is too big, it must be at most %s",
			fmt.Sprintf("%.2f", value), formatCoordBound(axis.Max))
	}
	return f, rel, nil
}

// formatCoordBound renders a bound the way the server puts it into its
// message: the bound is an int there.
func formatCoordBound(v float64) string {
	return strconv.FormatInt(int64(v), 10)
}

// NBT reads the rest of the line as an NBT compound.
//
// 1.12.2 hands the arguments from the compound's place on to the NBT
// parser as **one** string: they are joined with a single space
// (`CommandAbstract.a(String[], int)`). A compound with spaces in it is
// therefore one argument, even though the command line was split on
// spaces in the first place.
func (p *Parser) NBT() (*ast.NBTArg, error) {
	first, ok := p.Peek()
	if !ok {
		return nil, p.Errorf("expected NBT data but reached end of command")
	}
	text := p.Rest()
	c, err := nbt.ParseSNBT(text)
	if err != nil {
		return nil, p.ErrorAt(first, "invalid NBT: %v", err)
	}
	comp, ok := c.(*nbt.Compound)
	if !ok {
		return nil, p.ErrorAt(first, "expected an NBT compound")
	}
	return &ast.NBTArg{Value: comp, Offset: first.Offset}, nil
}

// JSON reads the rest of the line as raw JSON.
//
// Like a compound, a JSON is the rest of the line: the arguments from
// that place on are joined with a single space and handed to the JSON
// parser.
func (p *Parser) JSON() (*ast.JSONArg, error) {
	first, ok := p.Peek()
	if !ok {
		return nil, p.Errorf("expected JSON but reached end of command")
	}
	return &ast.JSONArg{Value: p.Rest(), Offset: first.Offset}, nil
}
