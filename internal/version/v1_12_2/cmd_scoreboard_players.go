package v1_12_2

import (
	"fmt"
	"strings"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/world"
)

// parsePlayers parses the `scoreboard players` subtree.
func parsePlayers(p *parser.Parser) ([]ast.Arg, error) {
	action, err := p.String()
	if err != nil {
		return nil, err
	}
	args, err := parsePlayersAction(p, action)
	if err != nil {
		// The usage of the subcommand is more helpful than the usage of
		// the whole tree.
		return nil, withUsage(err, "scoreboard players "+action)
	}
	return args, nil
}

// parsePlayersAction parses the action of `scoreboard players`.
func parsePlayersAction(p *parser.Parser, action string) ([]ast.Arg, error) {
	switch action {
	case "set", "add", "remove":
		holder, err := parseScoreHolder(p)
		if err != nil {
			return nil, err
		}
		objective, err := p.String()
		if err != nil {
			return nil, err
		}
		value, err := p.Int()
		if err != nil {
			return nil, err
		}
		return []ast.Arg{
			&ast.StringArg{Value: "players"},
			&ast.StringArg{Value: action},
			holder,
			&ast.StringArg{Value: objective},
			&ast.IntArg{Value: value},
		}, nil
	case "reset":
		holder, err := parseScoreHolder(p)
		if err != nil {
			return nil, err
		}
		objective := ""
		if p.More() {
			objective, _ = p.String()
		}
		return []ast.Arg{
			&ast.StringArg{Value: "players"},
			&ast.StringArg{Value: action},
			holder,
			&ast.StringArg{Value: objective},
		}, nil
	case "operation":
		target, err := parseScoreHolder(p)
		if err != nil {
			return nil, err
		}
		targetObj, err := p.String()
		if err != nil {
			return nil, err
		}
		op, err := p.String()
		if err != nil {
			return nil, err
		}
		source, err := parseScoreHolder(p)
		if err != nil {
			return nil, err
		}
		sourceObj, err := p.String()
		if err != nil {
			return nil, err
		}
		return []ast.Arg{
			&ast.StringArg{Value: "players"},
			&ast.StringArg{Value: action},
			target,
			&ast.StringArg{Value: targetObj},
			&ast.StringArg{Value: op},
			source,
			&ast.StringArg{Value: sourceObj},
		}, nil
	case "enable":
		holder, err := parseScoreHolder(p)
		if err != nil {
			return nil, err
		}
		trigger, err := p.String()
		if err != nil {
			return nil, err
		}
		return []ast.Arg{
			&ast.StringArg{Value: "players"},
			&ast.StringArg{Value: action},
			holder,
			&ast.StringArg{Value: trigger},
		}, nil
	case "list":
		holder, err := parseScoreHolder(p)
		if err != nil {
			return nil, err
		}
		return []ast.Arg{
			&ast.StringArg{Value: "players"},
			&ast.StringArg{Value: action},
			holder,
		}, nil
	case "test":
		holder, err := parseScoreHolder(p)
		if err != nil {
			return nil, err
		}
		objective, err := p.String()
		if err != nil {
			return nil, err
		}
		// The bounds are kept as text, because 1.12.2 accepts a `*`
		// wildcard there.
		low, err := p.String()
		if err != nil {
			return nil, err
		}
		high := ""
		if p.More() {
			high, err = p.String()
			if err != nil {
				return nil, err
			}
		}
		return []ast.Arg{
			&ast.StringArg{Value: "players"},
			&ast.StringArg{Value: action},
			holder,
			&ast.StringArg{Value: objective},
			&ast.StringArg{Value: low},
			&ast.StringArg{Value: high},
		}, nil
	case "tag":
		// `scoreboard players tag <target> <add|remove|list> [tag]` is
		// how 1.12.2 manages the tags of an entity.
		holder, err := parseScoreHolder(p)
		if err != nil {
			return nil, err
		}
		sub, err := p.String()
		if err != nil {
			return nil, err
		}
		tag := ""
		switch sub {
		case "add", "remove":
			tag, err = p.String()
			if err != nil {
				return nil, err
			}
		case "list":
		default:
			return nil, usageError(p, "scoreboard players tag", "expected add, remove or list, got %q", sub)
		}
		return []ast.Arg{
			&ast.StringArg{Value: "players"},
			&ast.StringArg{Value: "tag"},
			holder,
			&ast.StringArg{Value: sub},
			&ast.StringArg{Value: tag},
		}, nil
	}
	return nil, usageError(p, "scoreboard players", "expected set, add, remove, reset, operation, enable, test, tag or list, got %q", action)
}

// parseScoreHolder parses a score holder, which is a player name or a
// selector.
func parseScoreHolder(p *parser.Parser) (ast.Arg, error) {
	t, ok := p.Peek()
	if !ok {
		return nil, usageError(p, "scoreboard players set", "expected a player or target")
	}
	if parser.IsSelector(t.Text) {
		sel, err := p.Selector()
		if err != nil {
			return nil, err
		}
		return sel, nil
	}
	name, err := p.String()
	if err != nil {
		return nil, err
	}
	return &ast.StringArg{Value: name}, nil
}

// runPlayers executes the `scoreboard players` subtree.
func runPlayers(ctx *exec.Context, args []ast.Arg) (int, error) {
	sb := ctx.World.Scoreboard
	action := args[1].(*ast.StringArg).Value
	switch action {
	case "set", "add", "remove":
		holders, err := resolveHolders(ctx, args[2])
		if err != nil {
			return 0, err
		}
		objective := args[3].(*ast.StringArg).Value
		o := sb.Objective(objective)
		if o == nil {
			return 0, exec.Errorf("No objective was found by the name '%s'", objective)
		}
		if o.Criteria.ReadOnly() {
			// 1.12.2 refuses to write the criteria the server keeps
			// for itself: health, xp, level, food, air and armor.
			return 0, exec.Errorf("The objective '%s' is read-only and cannot be set", objective)
		}
		value := args[4].(*ast.IntArg).Value
		for _, h := range holders {
			current, tracked := sb.Score(h, objective)
			if !tracked {
				current = 0
			}
			var next int
			switch action {
			case "set":
				next = value
			case "add":
				next = current + value
			case "remove":
				next = current - value
			}
			sb.SetScore(h, objective, next)
			if action == "set" {
				// 1.12.2 confirms every holder separately.
				ctx.Engine.Feedback = append(ctx.Engine.Feedback,
					fmt.Sprintf("Set score of %s for player %s to %d", objective, h, next))
			}
		}
		return len(holders), nil
	case "reset":
		holders, err := resolveHolders(ctx, args[2])
		if err != nil {
			return 0, err
		}
		objective := args[3].(*ast.StringArg).Value
		for _, h := range holders {
			sb.ResetScore(h, objective)
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Reset score %s of player %s", objective, h))
		}
		return len(holders), nil
	case "operation":
		targets, err := resolveHolders(ctx, args[2])
		if err != nil {
			return 0, err
		}
		targetObj := args[3].(*ast.StringArg).Value
		op := world.Operation(args[4].(*ast.StringArg).Value)
		if !op.Valid() {
			return 0, exec.Errorf("unknown operation %q", op)
		}
		sources, err := resolveHolders(ctx, args[5])
		if err != nil {
			return 0, err
		}
		sourceObj := args[6].(*ast.StringArg).Value
		if sb.Objective(targetObj) == nil {
			return 0, exec.Errorf("No objective was found by the name '%s'", targetObj)
		}
		if sb.Objective(sourceObj) == nil {
			return 0, exec.Errorf("No objective was found by the name '%s'", sourceObj)
		}
		if o := sb.Objective(targetObj); o.Criteria.ReadOnly() {
			return 0, exec.Errorf("The objective '%s' is read-only and cannot be set", targetObj)
		}
		for _, t := range targets {
			for _, s := range sources {
				targetVal, _ := sb.Score(t, targetObj)
				sourceVal, ok := sb.Score(s, sourceObj)
				if !ok {
					return 0, exec.Errorf("No %s score for %s found", sourceObj, s)
				}
				if op == world.OpSwap {
					// The swap exchanges the two scores; it does not just
					// copy the source over the target.
					sb.SetScore(t, targetObj, sourceVal)
					sb.SetScore(s, sourceObj, targetVal)
					continue
				}
				next, err := world.Apply(op, targetVal, sourceVal)
				if err != nil {
					return 0, exec.Errorf("%v", err)
				}
				sb.SetScore(t, targetObj, next)
			}
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, "Operation applied successfully")
		return len(targets), nil
	case "enable":
		holders, err := resolveHolders(ctx, args[2])
		if err != nil {
			return 0, err
		}
		trigger := args[3].(*ast.StringArg).Value
		o := sb.Objective(trigger)
		if o == nil {
			return 0, exec.Errorf("No objective was found by the name '%s'", trigger)
		}
		if o.Criteria != world.CriteriaTrigger {
			return 0, exec.Errorf("Objective %s is not a trigger", trigger)
		}
		for _, h := range holders {
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Enabled trigger %s for %s", trigger, h))
		}
		return len(holders), nil
	case "list":
		holders, err := resolveHolders(ctx, args[2])
		if err != nil {
			return 0, err
		}
		for _, h := range holders {
			// 1.12.2 prints a heading and then one line per tracked
			// objective, as "- <objective>: <score> (<criterion>)".
			tracked := 0
			for _, o := range sb.Objectives() {
				if _, ok := sb.Score(h, o.Name); ok {
					tracked++
				}
			}
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Showing %d tracked objective(s) for %s:", tracked, h))
			for _, o := range sb.Objectives() {
				if v, ok := sb.Score(h, o.Name); ok {
					ctx.Engine.Feedback = append(ctx.Engine.Feedback,
						fmt.Sprintf("- %s: %d (%s)", o.Name, v, o.Criteria))
				}
			}
		}
		return len(holders), nil
	case "test":
		holders, err := resolveHolders(ctx, args[2])
		if err != nil {
			return 0, err
		}
		objective := args[3].(*ast.StringArg).Value
		if sb.Objective(objective) == nil {
			return 0, exec.Errorf("No objective was found by the name '%s'", objective)
		}
		low, high, err := parseTestRange(args[4].(*ast.StringArg).Value, args[5].(*ast.StringArg).Value)
		if err != nil {
			return 0, err
		}
		for _, h := range holders {
			score, tracked := sb.Score(h, objective)
			if !tracked {
				return 0, exec.Errorf("No %s score for %s found", objective, h)
			}
			if score < low || score > high {
				return 0, exec.Errorf("Score %d is NOT in range %d to %d", score, low, high)
			}
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Score %d is in range %d to %d", score, low, high))
		}
		return len(holders), nil
	case "tag":
		// 1.12.2 has no standalone `/tag` command: the tags of an entity
		// are managed through `scoreboard players tag`.
		holders, err := resolveHolders(ctx, args[2])
		if err != nil {
			return 0, err
		}
		action := args[3].(*ast.StringArg).Value
		switch action {
		case "add", "remove":
			tag := args[4].(*ast.StringArg).Value
			for _, h := range holders {
				e := ctx.World.FindEntity(h)
				if e == nil {
					continue
				}
				if action == "remove" && !e.HasTag(tag) {
					return 0, exec.Errorf("Entity didn't have the %s tag", tag)
				}
				e.SetTag(tag, action == "add")
				if action == "add" {
					ctx.Engine.Feedback = append(ctx.Engine.Feedback,
						fmt.Sprintf("Tag %s added", tag))
				} else {
					ctx.Engine.Feedback = append(ctx.Engine.Feedback,
						fmt.Sprintf("Tag %s removed", tag))
				}
			}
			return len(holders), nil
		case "list":
			for _, h := range holders {
				e := ctx.World.FindEntity(h)
				if e == nil {
					continue
				}
				// 1.12.2 prints a heading and then the tags joined
				// with " and " on the next line.
				tags := e.TagList()
				ctx.Engine.Feedback = append(ctx.Engine.Feedback,
					fmt.Sprintf("Tags on entity %s are:", h))
				ctx.Engine.Feedback = append(ctx.Engine.Feedback,
					strings.Join(tags, " and "))
			}
			return len(holders), nil
		}
		return 0, exec.Errorf("expected add, remove or list, got %q", action)
	}
	return 0, exec.Errorf("unknown action %q", action)
}

// parseTestRange parses the bounds of `scoreboard players test`.
//
// 1.12.2 accepts a `*` wildcard for either bound: a `*` lower bound means
// "no lower limit" and a `*` upper bound means "no upper limit".
//
// When only one bound is given, the other one is **not** a mirror of it:
// the missing upper bound is Integer.MAX_VALUE, so `players test Steve t 6`
// means "at least 6", not "exactly 6". (Measured: the server answers
// "Score 5 is NOT in range 6 to 2147483647".)
func parseTestRange(low, high string) (int, int, error) {
	lo := -2147483648
	if low != "*" {
		n, err := parseIntSafe(low)
		if err != nil {
			return 0, 0, exec.Errorf("expected a number, got %q", low)
		}
		lo = n
	}
	hi := 2147483647
	switch {
	case high == "":
		// 1.12.2 leaves the upper bound at Integer.MAX_VALUE.
	case high == "*":
		// No upper limit.
	default:
		n, err := parseIntSafe(high)
		if err != nil {
			return 0, 0, exec.Errorf("expected a number, got %q", high)
		}
		hi = n
	}
	return lo, hi, nil
}
