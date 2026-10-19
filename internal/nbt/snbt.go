package nbt

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// SNBTError describes a failure to parse SNBT text.
type SNBTError struct {
	// Offset is the byte offset within the input where parsing failed.
	Offset int
	// Msg explains the failure.
	Msg string
}

func (e *SNBTError) Error() string {
	return fmt.Sprintf("invalid NBT at offset %d: %s", e.Offset, e.Msg)
}

// ParseSNBT parses a single NBT value in 1.12.2 SNBT form.
//
// The accepted grammar follows the 1.12.2 `JsonToNBT` implementation:
// unquoted keys and string values, numeric suffixes (b/s/L/f/d) selecting the
// tag type, `[B;..]` byte arrays, `[I;..]` int arrays, and nested compounds.
func ParseSNBT(input string) (Value, error) {
	p := &snbtParser{src: input}
	p.skipSpace()
	v, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	p.skipSpace()
	if p.pos != len(p.src) {
		return nil, &SNBTError{Offset: p.pos, Msg: "trailing characters after value"}
	}
	return v, nil
}

type snbtParser struct {
	src string
	pos int
}

func (p *snbtParser) skipSpace() {
	for p.pos < len(p.src) {
		switch p.src[p.pos] {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}

func (p *snbtParser) parseValue() (Value, error) {
	if p.pos >= len(p.src) {
		return nil, &SNBTError{Offset: p.pos, Msg: "unexpected end of input"}
	}
	switch c := p.src[p.pos]; {
	case c == '{':
		return p.parseCompound()
	case c == '[':
		return p.parseList()
	case c == '"' || c == '\'':
		return p.parseQuotedString()
	default:
		return p.parseUnquoted()
	}
}

func (p *snbtParser) parseCompound() (Value, error) {
	p.pos++ // '{'
	out := NewCompound()
	for {
		p.skipSpace()
		if p.pos >= len(p.src) {
			return nil, &SNBTError{Offset: p.pos, Msg: "unterminated compound"}
		}
		if p.src[p.pos] == '}' {
			p.pos++
			return out, nil
		}
		key, err := p.parseKey()
		if err != nil {
			return nil, err
		}
		p.skipSpace()
		if p.pos >= len(p.src) || p.src[p.pos] != ':' {
			return nil, &SNBTError{Offset: p.pos, Msg: "expected ':' after key " + strconv.Quote(key)}
		}
		p.pos++
		p.skipSpace()
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		out.Set(key, val)
		p.skipSpace()
		if p.pos < len(p.src) && p.src[p.pos] == ',' {
			p.pos++
			continue
		}
	}
}

func (p *snbtParser) parseKey() (string, error) {
	if p.pos < len(p.src) && (p.src[p.pos] == '"' || p.src[p.pos] == '\'') {
		v, err := p.parseQuotedString()
		if err != nil {
			return "", err
		}
		return string(v.(String)), nil
	}
	start := p.pos
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if c == ':' || c == ' ' || c == '\t' || c == ',' || c == '}' {
			break
		}
		p.pos++
	}
	if p.pos == start {
		return "", &SNBTError{Offset: p.pos, Msg: "expected key"}
	}
	return p.src[start:p.pos], nil
}

func (p *snbtParser) parseList() (Value, error) {
	p.pos++ // '['
	p.skipSpace()
	if p.pos < len(p.src) && p.src[p.pos] == 'B' && p.pos+1 < len(p.src) && p.src[p.pos+1] == ';' {
		p.pos += 2
		return p.parseByteArray()
	}
	if p.pos < len(p.src) && p.src[p.pos] == 'I' && p.pos+1 < len(p.src) && p.src[p.pos+1] == ';' {
		p.pos += 2
		return p.parseIntArray()
	}
	// Collect elements first so the list's element type is known before
	// construction, matching how 1.12.2 infers TAG_List's element type.
	var elems []Value
	elemType := TypeEnd
	for {
		p.skipSpace()
		if p.pos >= len(p.src) {
			return nil, &SNBTError{Offset: p.pos, Msg: "unterminated list"}
		}
		if p.src[p.pos] == ']' {
			p.pos++
			break
		}
		v, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		if elemType == TypeEnd {
			elemType = v.Type()
		}
		elems = append(elems, v)
		p.skipSpace()
		if p.pos < len(p.src) && p.src[p.pos] == ',' {
			p.pos++
		}
	}
	if elemType == TypeEnd {
		elemType = TypeEnd
	}
	out := NewList(elemType)
	for _, v := range elems {
		out.items = append(out.items, v)
	}
	return out, nil
}

func (p *snbtParser) parseByteArray() (Value, error) {
	var out ByteArray
	for {
		p.skipSpace()
		if p.pos >= len(p.src) {
			return nil, &SNBTError{Offset: p.pos, Msg: "unterminated byte array"}
		}
		if p.src[p.pos] == ']' {
			p.pos++
			return out, nil
		}
		start := p.pos
		for p.pos < len(p.src) && p.src[p.pos] != ',' && p.src[p.pos] != ']' {
			p.pos++
		}
		n, err := parseIntLiteral(strings.TrimSpace(p.src[start:p.pos]))
		if err != nil {
			return nil, &SNBTError{Offset: start, Msg: "invalid byte value"}
		}
		out = append(out, byte(n))
		if p.pos < len(p.src) && p.src[p.pos] == ',' {
			p.pos++
		}
	}
}

func (p *snbtParser) parseIntArray() (Value, error) {
	var out IntArray
	for {
		p.skipSpace()
		if p.pos >= len(p.src) {
			return nil, &SNBTError{Offset: p.pos, Msg: "unterminated int array"}
		}
		if p.src[p.pos] == ']' {
			p.pos++
			return out, nil
		}
		start := p.pos
		for p.pos < len(p.src) && p.src[p.pos] != ',' && p.src[p.pos] != ']' {
			p.pos++
		}
		n, err := parseIntLiteral(strings.TrimSpace(p.src[start:p.pos]))
		if err != nil {
			return nil, &SNBTError{Offset: start, Msg: "invalid int value"}
		}
		out = append(out, int32(n))
		if p.pos < len(p.src) && p.src[p.pos] == ',' {
			p.pos++
		}
	}
}

func (p *snbtParser) parseQuotedString() (Value, error) {
	quote := p.src[p.pos]
	p.pos++
	var sb strings.Builder
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if c == '\\' && p.pos+1 < len(p.src) {
			p.pos++
			switch p.src[p.pos] {
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
				sb.WriteByte(p.src[p.pos])
			}
			p.pos++
			continue
		}
		if c == quote {
			p.pos++
			return String(sb.String()), nil
		}
		sb.WriteByte(c)
		p.pos++
	}
	return nil, &SNBTError{Offset: p.pos, Msg: "unterminated string"}
}

// parseUnquoted handles bare words: numbers with optional type suffix, or
// plain strings (which 1.12.2 treats as TAG_String).
func (p *snbtParser) parseUnquoted() (Value, error) {
	start := p.pos
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if c == ':' || c == ',' || c == '}' || c == ']' || c == ' ' || c == '\t' {
			break
		}
		p.pos++
	}
	raw := p.src[start:p.pos]
	if raw == "" {
		return nil, &SNBTError{Offset: start, Msg: "expected value"}
	}
	return parseBare(raw, start)
}

// parseBare interprets a bare token, using the suffix to pick the tag type.
func parseBare(raw string, offset int) (Value, error) {
	body, suffix := raw, byte(0)
	if n := len(raw); n > 0 {
		last := raw[n-1]
		switch last {
		case 'b', 'B', 's', 'S', 'l', 'L', 'f', 'F', 'd', 'D':
			// Only treat it as a suffix when the remainder is numeric;
			// plain words like "dead" must stay strings.
			if _, err := strconv.ParseFloat(raw[:n-1], 64); err == nil {
				body = raw[:n-1]
				suffix = last
			}
		}
	}
	switch suffix {
	case 'b', 'B':
		n, err := strconv.ParseInt(body, 10, 8)
		if err != nil {
			return nil, &SNBTError{Offset: offset, Msg: "invalid byte value " + strconv.Quote(raw)}
		}
		return Byte(int8(n)), nil
	case 's', 'S':
		n, err := strconv.ParseInt(body, 10, 16)
		if err != nil {
			return nil, &SNBTError{Offset: offset, Msg: "invalid short value " + strconv.Quote(raw)}
		}
		return Short(int16(n)), nil
	case 'l', 'L':
		n, err := strconv.ParseInt(body, 10, 64)
		if err != nil {
			return nil, &SNBTError{Offset: offset, Msg: "invalid long value " + strconv.Quote(raw)}
		}
		return Long(n), nil
	case 'f', 'F':
		n, err := strconv.ParseFloat(body, 32)
		if err != nil {
			return nil, &SNBTError{Offset: offset, Msg: "invalid float value " + strconv.Quote(raw)}
		}
		return Float(float32(n)), nil
	case 'd', 'D':
		n, err := strconv.ParseFloat(body, 64)
		if err != nil {
			return nil, &SNBTError{Offset: offset, Msg: "invalid double value " + strconv.Quote(raw)}
		}
		return Double(n), nil
	}
	// No suffix: an integer literal becomes TAG_Int, a float TAG_Double,
	// anything else TAG_String.
	if n, err := strconv.ParseInt(body, 10, 32); err == nil {
		return Int(int32(n)), nil
	}
	if f, err := strconv.ParseFloat(body, 64); err == nil && !isAllDigits(body) {
		return Double(f), nil
	}
	return String(raw), nil
}

// parseIntLiteral parses an integer that may carry a type suffix such as
// `1b`. Array elements accept the same suffixes as scalar values.
func parseIntLiteral(raw string) (int64, error) {
	body := raw
	if n := len(body); n > 0 {
		switch body[n-1] {
		case 'b', 'B', 's', 'S', 'l', 'L':
			if _, err := strconv.ParseInt(body[:n-1], 10, 64); err == nil {
				body = body[:n-1]
			}
		}
	}
	return strconv.ParseInt(body, 10, 64)
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) && r != '-' && r != '+' {
			return false
		}
	}
	return true
}

// quote renders a string as a quoted, escaped SNBT literal.
func quote(s string) string {
	var sb strings.Builder
	sb.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			sb.WriteString("\\\"")
		case '\\':
			sb.WriteString("\\\\")
		case '\n':
			sb.WriteString("\\n")
		case '\r':
			sb.WriteString("\\r")
		case '\t':
			sb.WriteString("\\t")
		default:
			sb.WriteRune(r)
		}
	}
	sb.WriteByte('"')
	return sb.String()
}
