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
			return nil, p.Errorf("expected add, remove or list, got %q", sub)
		}
		return []ast.Arg{
			&ast.StringArg{Value: "players"},
			&ast.StringArg{Value: "tag"},
			holder,
			&ast.StringArg{Value: sub},
			&ast.StringArg{Value: tag},
		}, nil
	}
	return nil, p.Errorf("expected set, add, remove, reset, operation, enable, test, tag or list, got %q", action)
}

// parseScoreHolder parses a score holder, which is a player name or a
// selector.
func parseScoreHolder(p *parser.Parser) (ast.Arg, error) {
	t, ok := p.Peek()
	if !ok {
		return nil, p.Errorf("expected a player or target")
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
		if sb.Objective(objective) == nil {
			return 0, exec.Errorf("unknown objective %q", objective)
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
			return 0, exec.Errorf("unknown objective %q", targetObj)
		}
		if sb.Objective(sourceObj) == nil {
			return 0, exec.Errorf("unknown objective %q", sourceObj)
		}
		for _, t := range targets {
			for _, s := range sources {
				targetVal, _ := sb.Score(t, targetObj)
				sourceVal, ok := sb.Score(s, sourceObj)
				if !ok {
					return 0, exec.Errorf("holder %q has no score for %q", s, sourceObj)
				}
				next, err := world.Apply(op, targetVal, sourceVal)
				if err != nil {
					return 0, exec.Errorf("%v", err)
				}
				sb.SetScore(t, targetObj, next)
			}
		}
		return len(targets), nil
	case "enable":
		holders, err := resolveHolders(ctx, args[2])
		if err != nil {
			return 0, err
		}
		trigger := args[3].(*ast.StringArg).Value
		o := sb.Objective(trigger)
		if o == nil {
			return 0, exec.Errorf("unknown objective %q", trigger)
		}
		if o.Criteria != world.CriteriaTrigger {
			return 0, exec.Errorf("objective %q is not a trigger", trigger)
		}
		return len(holders), nil
	case "list":
		holders, err := resolveHolders(ctx, args[2])
		if err != nil {
			return 0, err
		}
		for _, h := range holders {
			for _, o := range sb.Objectives() {
				if v, ok := sb.Score(h, o.Name); ok {
					ctx.Engine.Feedback = append(ctx.Engine.Feedback,
						fmt.Sprintf("%s: %d (%s)", h, v, o.Name))
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
			return 0, exec.Errorf("unknown objective %q", objective)
		}
		low, high, err := parseTestRange(args[4].(*ast.StringArg).Value, args[5].(*ast.StringArg).Value)
		if err != nil {
			return 0, err
		}
		for _, h := range holders {
			score, tracked := sb.Score(h, objective)
			if !tracked {
				return 0, exec.Errorf("%q has no score for the objective %q", h, objective)
			}
			if score < low || score > high {
				return 0, exec.Errorf("the score %d of %q is not in the range %d..%d", score, h, low, high)
			}
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Test passed: %d is in the range %d..%d", score, low, high))
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
				e.SetTag(tag, action == "add")
			}
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Successfully added tag %s", tag))
			return len(holders), nil
		case "list":
			for _, h := range holders {
				e := ctx.World.FindEntity(h)
				if e == nil {
					continue
				}
				tags := e.TagList()
				ctx.Engine.Feedback = append(ctx.Engine.Feedback,
					fmt.Sprintf("Tags for %s: %s", h, strings.Join(tags, ", ")))
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
// "no lower limit" and a `*` upper bound means "no upper limit". When only
// one bound is given the other one mirrors it, so the test is an equality
// check.
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
		hi = lo
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
