package v1_12_2

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/world"
)

// worldBorderCmd implements `/worldborder`.
//
// The 1.12.2 syntax is `worldborder <set|add|center|damage|warning|get>
// [...]`:
//
//	worldborder set <size> [<time>]      set the diameter
//	worldborder add <size> [<time>]      grow or shrink the diameter
//	worldborder center <x> <z>           move the centre
//	worldborder damage buffer <amount>   the damage buffer
//	worldborder damage amount <amount>   the damage per block
//	worldborder warning time <seconds>   the warning delay
//	worldborder warning distance <blocks> the warning distance
//	worldborder get                     report the diameter
type worldBorderCmd struct{}

// Name implements exec.Handler.
func (c *worldBorderCmd) Name() string { return "worldborder" }

// Usage implements exec.Handler.
func (c *worldBorderCmd) Usage() string {
	return "/worldborder <set|add|center|damage|warning|get> [...]"
}

// Parse implements exec.Handler.
func (c *worldBorderCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	action, err := p.String()
	if err != nil {
		return nil, err
	}
	switch action {
	case "set", "add", "center", "damage", "warning", "get":
	default:
		return nil, p.Errorf("expected set, add, center, damage, warning or get, got %q", action)
	}
	rest := p.Rest()
	return []ast.Arg{
		&ast.StringArg{Value: action},
		&ast.StringArg{Value: rest},
	}, nil
}

// Run implements exec.Handler.
func (c *worldBorderCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	action := args[0].(*ast.StringArg).Value
	rest := strings.Fields(args[1].(*ast.StringArg).Value)
	border := &ctx.World.Border

	switch action {
	case "get":
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("World border is currently %.0f blocks wide", border.Size))
		return 1, nil
	case "set", "add":
		if len(rest) < 1 {
			return 0, exec.Errorf("expected a size")
		}
		size, err := strconv.ParseFloat(rest[0], 64)
		if err != nil {
			return 0, exec.Errorf("expected a number, got %q", rest[0])
		}
		old := border.Size
		if action == "add" {
			// 1.12.2 adds the number to the current size, and the
			// change may be negative as long as the border keeps a
			// width.
			if size < -old {
				return 0, exec.Errorf(
					"The number you have entered (%s) is too small, it must be at least %s",
					vanillaDouble2(size), vanillaDouble2(-old))
			}
			size += old
		} else if size < 1 {
			return 0, exec.Errorf(
				"The number you have entered (%s) is too small, it must be at least %s",
				vanillaDouble2(size), vanillaDouble2(1))
		}
		seconds := 0
		if len(rest) > 1 {
			seconds, err = strconv.Atoi(rest[1])
			if err != nil {
				return 0, exec.Errorf("expected a number, got %q", rest[1])
			}
		}
		border.Size = size
		if seconds > 0 {
			// 1.12.2 moves the border over the given time, and it says
			// whether the border grows or shrinks: the border shrinks
			// when the new size is the smaller one.
			border.LerpSize = size
			if old > size {
				ctx.Engine.Feedback = append(ctx.Engine.Feedback,
					fmt.Sprintf("Shrinking world border to %.1f blocks wide (down from %.1f blocks) over %d seconds",
						size, old, seconds))
			} else {
				ctx.Engine.Feedback = append(ctx.Engine.Feedback,
					fmt.Sprintf("Growing world border to %.1f blocks wide (up from %.1f blocks) over %d seconds",
						size, old, seconds))
			}
			return 1, nil
		}
		// Without a time the border changes at once.
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Set world border to %.1f blocks wide (from %.1f blocks)", size, old))
		return 1, nil
	case "center":
		if len(rest) < 2 {
			return 0, exec.Errorf("expected an x and a z coordinate")
		}
		// 1.12.2 centres the border on the middle of a block, so a
		// whole coordinate is moved by half a block.
		base := ctx.Source.Pos.BlockPos()
		x, err := centerCoord(rest[0], float64(base.X)+0.5)
		if err != nil {
			return 0, err
		}
		z, err := centerCoord(rest[1], float64(base.Z)+0.5)
		if err != nil {
			return 0, err
		}
		border.Center = world.Vec{X: x, Z: z}
		// The message carries the raw double, so it looks the way
		// Java prints it.
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Set world border center to %s, %s", javaDouble(x), javaDouble(z)))
		return 1, nil
	case "damage":
		if len(rest) < 2 {
			return 0, exec.Errorf("expected buffer or amount")
		}
		value, err := strconv.ParseFloat(rest[1], 64)
		if err != nil {
			return 0, exec.Errorf("expected a number, got %q", rest[1])
		}
		switch rest[0] {
		case "buffer":
			old := border.DamageBuffer
			border.DamageBuffer = value
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Set world border damage buffer to %.1f blocks (from %.1f blocks)", value, old))
		case "amount":
			old := border.DamagePerBlock
			border.DamagePerBlock = value
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Set world border damage amount to %.2f per block (from %.2f per block)", value, old))
		default:
			return 0, exec.Errorf("expected buffer or amount, got %q", rest[0])
		}
		return 1, nil
	case "warning":
		if len(rest) < 2 {
			return 0, exec.Errorf("expected time or distance")
		}
		value, err := strconv.Atoi(rest[1])
		if err != nil {
			return 0, exec.Errorf("expected a number, got %q", rest[1])
		}
		switch rest[0] {
		case "time":
			old := border.WarningTime
			border.WarningTime = value
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Set world border warning to %d seconds away (from %d seconds)", value, old))
		case "distance":
			old := border.WarningDistance
			border.WarningDistance = float64(value)
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Set world border warning to %d blocks away (from %d blocks)", value, int(old)))
		default:
			return 0, exec.Errorf("expected time or distance, got %q", rest[0])
		}
		return 1, nil
	}
	return 0, exec.Errorf("unknown action %q", action)
}

// centerCoord parses one coordinate of `/worldborder center`.
//
// 1.12.2 uses `CommandAbstract.b(base, text, true)`: a `~` is relative to
// the position of the command, a whole number is moved to the middle of
// the block and a number with a decimal point is taken as it is.
func centerCoord(text string, base float64) (float64, error) {
	relative := strings.HasPrefix(text, "~")
	if relative {
		text = text[1:]
	}
	if relative && text == "" {
		// A bare `~` is the position of the command itself.
		return base, nil
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, exec.Errorf("expected a number, got %q", text)
	}
	if relative {
		return base + value, nil
	}
	if !strings.Contains(text, ".") {
		// A whole number names a block, and 1.12.2 centres the border
		// on the middle of it.
		value += 0.5
	}
	return value, nil
}

// vanillaDouble2 renders a double the way 1.12.2 puts a number into an
// error message.
//
// `CommandAbstract.a(String, double, double)` formats the number with
// "%.2f".
func vanillaDouble2(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

// javaDouble renders a double the way Java's Double.toString does.
//
// 1.12.2 puts the raw double into its messages, so an integral value keeps
// its ".0" and a very large or very small one switches to the "1.0E7"
// form.
func javaDouble(v float64) string {
	switch {
	case math.IsNaN(v):
		return "NaN"
	case math.IsInf(v, 1):
		return "Infinity"
	case math.IsInf(v, -1):
		return "-Infinity"
	}
	abs := math.Abs(v)
	if v != 0 && (abs < 1e-3 || abs >= 1e7) {
		s := strconv.FormatFloat(v, 'E', -1, 64)
		mant, exp, _ := strings.Cut(s, "E")
		if !strings.Contains(mant, ".") {
			mant += ".0"
		}
		exp = strings.TrimPrefix(exp, "+")
		neg := strings.HasPrefix(exp, "-")
		exp = strings.TrimLeft(strings.TrimPrefix(exp, "-"), "0")
		if exp == "" {
			exp = "0"
		}
		if neg {
			exp = "-" + exp
		}
		return mant + "E" + exp
	}
	s := strconv.FormatFloat(v, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}

// registerWorldBorder adds the world border command.
func registerWorldBorder(reg *exec.Registry) {
	reg.Register(&worldBorderCmd{})
}
