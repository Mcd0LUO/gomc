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

// weatherCmd implements `/weather`.
//
// The 1.12.2 syntax is `weather <clear|rain|thunder> [<duration>]`, where
// the duration is a number of seconds. Without one the server picks a
// random duration between 300 and 900 seconds.
//
// The command distinguishes the "clear" time from the rain time: `/weather
// clear` schedules the clear weather and stops the rain, while `/weather
// rain` starts the rain right away.
type weatherCmd struct{}

// Name implements exec.Handler.
func (c *weatherCmd) Name() string { return "weather" }

// Usage implements exec.Handler.
func (c *weatherCmd) Usage() string { return "/weather <clear|rain|thunder> [<duration>]" }

// Parse implements exec.Handler.
func (c *weatherCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	kind, err := p.String()
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(kind) {
	case "clear", "rain", "thunder":
	default:
		return nil, p.Errorf("expected clear, rain or thunder, got %q", kind)
	}
	seconds := 0
	if p.More() {
		seconds, err = p.Int()
		if err != nil {
			return nil, err
		}
	}
	return []ast.Arg{
		&ast.StringArg{Value: strings.ToLower(kind)},
		&ast.IntArg{Value: seconds},
	}, nil
}

// Run implements exec.Handler.
func (c *weatherCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	kind := args[0].(*ast.StringArg).Value
	seconds := args[1].(*ast.IntArg).Value

	// The default duration is a random 300 to 900 seconds, counted in
	// ticks.
	ticks := (300 + exec.RandIntn(600)) * 20
	if seconds > 0 {
		ticks = seconds * 20
	}
	if seconds < 0 {
		return 0, exec.Errorf("expected a positive duration, got %d", seconds)
	}

	w := ctx.World
	switch kind {
	case "clear":
		w.Weather.ClearDuration = ticks
		w.Weather.Duration = 0
		w.Weather.Storm = false
		w.Weather.Thundering = false
		w.Weather.Kind = world.WeatherClear
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, "Changing the weather to clear")
	case "rain":
		w.Weather.ClearDuration = 0
		w.Weather.Duration = ticks
		w.Weather.Storm = true
		w.Weather.Thundering = false
		w.Weather.Kind = world.WeatherRain
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, "Changing the weather to rain")
	case "thunder":
		w.Weather.ClearDuration = 0
		w.Weather.Duration = ticks
		w.Weather.Storm = true
		w.Weather.Thundering = true
		w.Weather.Kind = world.WeatherThunder
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, "Changing the weather to thunder")
	}
	return 1, nil
}

// toggleDownfallCmd implements `/toggledownfall`.
//
// The command toggles whether it is raining, and nothing else.
type toggleDownfallCmd struct{}

// Name implements exec.Handler.
func (c *toggleDownfallCmd) Name() string { return "toggledownfall" }

// Usage implements exec.Handler.
func (c *toggleDownfallCmd) Usage() string { return "/toggledownfall" }

// Parse implements exec.Handler.
func (c *toggleDownfallCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	return nil, nil
}

// Run implements exec.Handler.
func (c *toggleDownfallCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	w := ctx.World
	w.Weather.Storm = !w.Weather.Storm
	if w.Weather.Storm {
		w.Weather.Kind = world.WeatherRain
	} else {
		w.Weather.Kind = world.WeatherClear
		w.Weather.Thundering = false
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback, "Toggled downfall")
	return 1, nil
}

// difficultyCmd implements `/difficulty`.
//
// The 1.12.2 syntax is `difficulty <difficulty>`, and the argument accepts
// the full name, the single letter abbreviation, or the numeric id.
type difficultyCmd struct{}

// Name implements exec.Handler.
func (c *difficultyCmd) Name() string { return "difficulty" }

// Usage implements exec.Handler.
func (c *difficultyCmd) Usage() string { return "/difficulty <difficulty>" }

// Parse implements exec.Handler.
func (c *difficultyCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	name, err := p.String()
	if err != nil {
		return nil, err
	}
	return []ast.Arg{&ast.StringArg{Value: name}}, nil
}

// Run implements exec.Handler.
func (c *difficultyCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	name := args[0].(*ast.StringArg).Value
	d, ok := parseDifficulty(name)
	if !ok {
		return 0, exec.Errorf("unknown difficulty %q", name)
	}
	ctx.World.Difficulty = d
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Set game difficulty to %s", d))
	return 1, nil
}

// parseDifficulty resolves a difficulty argument.
//
// 1.12.2 accepts the full name, the single letter abbreviation and the
// numeric id of a difficulty.
func parseDifficulty(name string) (world.Difficulty, bool) {
	switch strings.ToLower(name) {
	case "peaceful", "p":
		return world.DifficultyPeaceful, true
	case "easy", "e":
		return world.DifficultyEasy, true
	case "normal", "n":
		return world.DifficultyNormal, true
	case "hard", "h":
		return world.DifficultyHard, true
	}
	id, err := strconv.Atoi(name)
	if err != nil || id < 0 || id > 3 {
		return "", false
	}
	for d, v := range world.DifficultyIDs {
		if v == id {
			return d, true
		}
	}
	return "", false
}

// seedCmd implements `/seed`.
//
// The command reports the world's seed.
type seedCmd struct{}

// Name implements exec.Handler.
func (c *seedCmd) Name() string { return "seed" }

// Usage implements exec.Handler.
func (c *seedCmd) Usage() string { return "/seed" }

// Parse implements exec.Handler.
func (c *seedCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	return nil, nil
}

// Run implements exec.Handler.
func (c *seedCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Seed: %d", ctx.World.Seed))
	return 1, nil
}

// idleTimeoutCmd implements `/setidletimeout`.
//
// The argument is a number of minutes, and 0 disables the timeout.
type idleTimeoutCmd struct{}

// Name implements exec.Handler.
func (c *idleTimeoutCmd) Name() string { return "setidletimeout" }

// Usage implements exec.Handler.
func (c *idleTimeoutCmd) Usage() string { return "/setidletimeout <minutes>" }

// Parse implements exec.Handler.
func (c *idleTimeoutCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	minutes, err := p.Int()
	if err != nil {
		return nil, err
	}
	return []ast.Arg{&ast.IntArg{Value: minutes}}, nil
}

// Run implements exec.Handler.
func (c *idleTimeoutCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	minutes := args[0].(*ast.IntArg).Value
	ctx.World.IdleTimeout = minutes
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Idle timeout set to %d minutes", minutes))
	return 1, nil
}

// registerEnvironment adds the environment commands.
func registerEnvironment(reg *exec.Registry) {
	reg.Register(&weatherCmd{})
	reg.Register(&toggleDownfallCmd{})
	reg.Register(&difficultyCmd{})
	reg.Register(&seedCmd{})
	reg.Register(&idleTimeoutCmd{})
}
