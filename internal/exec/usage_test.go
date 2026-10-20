package exec

import (
	"testing"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// usageCmd is a command that fails with a message that already is the usage
// line of the command, the way 1.12.2 answers a usage mistake.
type usageCmd struct{}

// Name implements Handler.
func (c *usageCmd) Name() string { return "usagecmd" }

// Usage implements Handler.
func (c *usageCmd) Usage() string { return "/usagecmd <a>" }

// Parse implements Handler.
func (c *usageCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	return nil, &Error{Msg: "Usage: /usagecmd <a>"}
}

// Run implements Handler.
func (c *usageCmd) Run(ctx *Context, args []ast.Arg) (int, error) { return 0, nil }

// bareUsageCmd is a command that fails with the bare usage line.
//
// 1.12.2 throws those as a plain `CommandException`, so the server prints the
// line as it is, without the "Usage: " prefix.
type bareUsageCmd struct{}

// Name implements Handler.
func (c *bareUsageCmd) Name() string { return "bareusage" }

// Usage implements Handler.
func (c *bareUsageCmd) Usage() string { return "/bareusage <a>" }

// Parse implements Handler.
func (c *bareUsageCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	return nil, &Error{Msg: "/bareusage <a> <b>"}
}

// Run implements Handler.
func (c *bareUsageCmd) Run(ctx *Context, args []ast.Arg) (int, error) { return 0, nil }

// describedCmd is a command that fails with a description of what is wrong,
// and leaves the usage line to the error report.
type describedCmd struct{}

// Name implements Handler.
func (c *describedCmd) Name() string { return "described" }

// Usage implements Handler.
func (c *describedCmd) Usage() string { return "/described <a>" }

// Parse implements Handler.
func (c *describedCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	return nil, &Error{Msg: "expected a value"}
}

// Run implements Handler.
func (c *describedCmd) Run(ctx *Context, args []ast.Arg) (int, error) { return 0, nil }

// TestUsageMessageIsNotRepeated pins down that a command whose error
// message already is the usage line does not get the usage line a second
// time.
//
// 1.12.2 answers a usage mistake with the usage line itself, and the error
// report of the runtime must not say it twice.
func TestUsageMessageIsNotRepeated(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  Handler
		line string
		want string
	}{
		{"prefixed", &usageCmd{}, "usagecmd", "Usage: /usagecmd <a>"},
		{"bare", &bareUsageCmd{}, "bareusage", "/bareusage <a> <b>"},
	} {
		reg := NewRegistry()
		reg.Register(tc.cmd)
		e := NewEngine(reg, world.NewWorld())
		_, err := e.ParseCommand(tc.line, 1)
		if err == nil {
			t.Errorf("%s: the command was expected to fail", tc.name)
			continue
		}
		pe, ok := err.(*parser.Error)
		if !ok {
			t.Errorf("%s: the error is a %T, want a *parser.Error", tc.name, err)
			continue
		}
		if pe.Msg != tc.want {
			t.Errorf("%s: the message was %q, want %q", tc.name, pe.Msg, tc.want)
		}
		if pe.Usage != "" {
			t.Errorf("%s: the usage line was repeated: %q", tc.name, pe.Usage)
		}
	}
}

// TestUsageOfADescribedErrorIsStillFilledIn pins down the other side: an
// error that describes what is wrong still gets the usage line of the
// command, because its message is not a usage line.
func TestUsageOfADescribedErrorIsStillFilledIn(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&describedCmd{})
	e := NewEngine(reg, world.NewWorld())
	_, err := e.ParseCommand("described", 1)
	if err == nil {
		t.Fatal("the command was expected to fail")
	}
	pe, ok := err.(*parser.Error)
	if !ok {
		t.Fatalf("the error is a %T, want a *parser.Error", err)
	}
	if pe.Msg != "expected a value" {
		t.Errorf("the message was %q, want %q", pe.Msg, "expected a value")
	}
	if pe.Usage != "/described <a>" {
		t.Errorf("the usage line was %q, want %q", pe.Usage, "/described <a>")
	}
}
