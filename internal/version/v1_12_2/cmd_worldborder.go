package v1_12_2

import (
	"fmt"
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
			fmt.Sprintf("The world border is currently %.0f blocks wide", border.Size))
		return 1, nil
	case "set", "add":
		if len(rest) < 1 {
			return 0, exec.Errorf("expected a size")
		}
		size, err := strconv.ParseFloat(rest[0], 64)
		if err != nil {
			return 0, exec.Errorf("expected a number, got %q", rest[0])
		}
		if size < 1 {
			return 0, exec.Errorf("the world border has to be at least 1 block wide")
		}
		if action == "add" {
			size += border.Size
		}
		old := border.Size
		border.Size = size
		if len(rest) > 1 {
			seconds, err := strconv.Atoi(rest[1])
			if err != nil {
				return 0, exec.Errorf("expected a number, got %q", rest[1])
			}
			border.LerpSize = size
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("The world border is now %.1f blocks wide, shrinking from %.1f in %d seconds",
					size, old, seconds))
			return 1, nil
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("The world border is now %.1f blocks wide", size))
		return 1, nil
	case "center":
		if len(rest) < 2 {
			return 0, exec.Errorf("expected an x and a z coordinate")
		}
		x, err := strconv.ParseFloat(rest[0], 64)
		if err != nil {
			return 0, exec.Errorf("expected a number, got %q", rest[0])
		}
		z, err := strconv.ParseFloat(rest[1], 64)
		if err != nil {
			return 0, exec.Errorf("expected a number, got %q", rest[1])
		}
		border.Center = world.Vec{X: x, Z: z}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("The world border centre is now %.1f, %.1f", x, z))
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
			old := border.DamagePerBlock
			border.DamagePerBlock = value
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("The world border damage buffer has been set to %.1f from %.1f", value, old))
		case "amount":
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("The world border damage amount has been set to %.2f", value))
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
				fmt.Sprintf("The world border warning time has been set to %d from %d", value, old))
		case "distance":
			old := border.WarningDistance
			border.WarningDistance = float64(value)
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("The world border warning distance has been set to %d from %d", value, int(old)))
		default:
			return 0, exec.Errorf("expected time or distance, got %q", rest[0])
		}
		return 1, nil
	}
	return 0, exec.Errorf("unknown action %q", action)
}

// registerWorldBorder adds the world border command.
func registerWorldBorder(reg *exec.Registry) {
	reg.Register(&worldBorderCmd{})
}
