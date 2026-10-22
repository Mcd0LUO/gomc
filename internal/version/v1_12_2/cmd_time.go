package v1_12_2

import (
	"fmt"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
)

// javaMaxInt is `Integer.MAX_VALUE`, the modulus `time query gametime`
// uses in 1.12.2.
const javaMaxInt = 1<<31 - 1

// timeCmd implements `/time`.
//
// The 1.12.2 syntax is `time <set|add|query> <value>`, and the value of
// `query` is mandatory:
//
//	time set <value>        day (1000), night (13000) or a tick count
//	time add <value>        a tick count
//	time query daytime      the time of day, 0 to 23999
//	time query day          the number of days since the world was created
//	time query gametime     the total number of world ticks
//
// Note that `daytime`, `noon` and `midnight` are 1.13 additions; 1.12.2
// only knows `day` and `night`.
type timeCmd struct{}

// Name implements exec.Handler.
func (c *timeCmd) Name() string { return "time" }

// Usage implements exec.Handler.
func (c *timeCmd) Usage() string { return usageLine("time") }

// Parse implements exec.Handler.
func (c *timeCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	action, err := p.String()
	if err != nil {
		return nil, err
	}
	switch action {
	case "set", "add", "query":
	default:
		return nil, p.Errorf("expected set, add or query, got %q", action)
	}
	value, err := p.String()
	if err != nil {
		return nil, err
	}
	return []ast.Arg{&ast.StringArg{Value: action}, &ast.StringArg{Value: value}}, nil
}

// Run implements exec.Handler.
func (c *timeCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	action := args[0].(*ast.StringArg).Value
	raw := args[1].(*ast.StringArg).Value
	w := ctx.World

	switch action {
	case "set":
		value, err := parseTimeValue(raw)
		if err != nil {
			return 0, exec.Errorf("%v", err)
		}
		w.DayTime = value
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, fmt.Sprintf("Set the time to %d", value))
		return 1, nil
	case "add":
		value, err := parseTimeCount(raw)
		if err != nil {
			return 0, exec.Errorf("%v", err)
		}
		w.DayTime += value
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, fmt.Sprintf("Added %d to the time", value))
		return 1, nil
	case "query":
		var value int
		switch raw {
		case "daytime":
			value = w.DayTime % 24000
		case "day":
			value = w.DayTime / 24000
		case "gametime":
			// `CommandTime` reports `getTime() % Integer.MAX_VALUE`.
			value = w.Time % javaMaxInt
		default:
			return 0, exec.Errorf("expected daytime, day or gametime, got %q", raw)
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, fmt.Sprintf("Time is %d", value))
		// The result of the query goes into the `QueryResult` channel.
		//
		// (`CommandTime` reports
		// `EnumCommandResult.QUERY_RESULT` with the value it found.)
		ctx.QueryResult = value
		return 1, nil
	}
	return 0, exec.Errorf("unknown action %q", action)
}

// parseTimeValue parses a `/time set` argument.
//
// 1.12.2 knows `day` (1000) and `night` (13000) on top of a plain tick
// count.
func parseTimeValue(raw string) (int, error) {
	switch raw {
	case "day":
		return 1000, nil
	case "night":
		return 13000, nil
	}
	return parseTimeCount(raw)
}

// parseTimeCount parses the tick count of `/time set` and `/time add`.
//
// Both read it with `CommandTime.a(text, 0)`, which is
// `CommandAbstract.a(String, int, int)` with a minimum of 0: a number that
// is not a Java int is `'%s' is not a valid number`, and a negative one is
// refused as too small. The number in that message is the integer itself --
// the "%.2f" wording belongs to the coordinate reader, which passes doubles.
func parseTimeCount(raw string) (int, error) {
	n, ok := parser.JavaInt(raw)
	if !ok {
		return 0, fmt.Errorf("'%s' is not a valid number", strings.TrimSpace(raw))
	}
	if n < 0 {
		return 0, fmt.Errorf("The number you have entered (%d) is too small, it must be at least 0", n)
	}
	return n, nil
}
