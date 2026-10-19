package v1_12_2

import (
	"strings"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/world"
)

// parseEntityTarget parses a command target.
//
// 1.12.2 resolves the target of every entity command the same way, so a
// bare player name works wherever a target selector does. The commands
// that take a target are `CommandEntity` and its subclasses, and they all
// go through `CommandAbstract.a`, which accepts both forms.
func parseEntityTarget(p *parser.Parser) (ast.Arg, error) {
	t, ok := p.Peek()
	if !ok {
		return nil, p.Errorf("expected a target")
	}
	if parser.IsSelector(t.Text) {
		return p.Selector()
	}
	name, err := p.String()
	if err != nil {
		return nil, err
	}
	return &ast.StringArg{Value: name}, nil
}

// resolveEntityTargets resolves a target argument into entities.
//
// A selector goes through the selector resolver, a bare name is looked up
// among the players.
//
// An empty result is an error, and 1.12.2 words the two cases
// differently:
//
//   - a selector that matches nothing is reported by the command
//     dispatcher as "Selector '<text>' found nothing"
//   - a player name that is not online is reported by the command itself
//     as "Player '<name>' cannot be found"
func resolveEntityTargets(ctx *exec.Context, arg ast.Arg) ([]*world.Entity, error) {
	switch t := arg.(type) {
	case *ast.Selector:
		resolver := &exec.SelectorResolver{World: ctx.World}
		entities, err := resolver.Resolve(ctx.Source, t)
		if err != nil {
			return nil, exec.Errorf("%v", err)
		}
		if len(entities) == 0 {
			return nil, exec.Errorf("Selector '%s' found nothing", selectorText(t))
		}
		return entities, nil
	case *ast.StringArg:
		p := ctx.World.FindPlayer(t.Value)
		if p == nil {
			return nil, exec.Errorf("Player '%s' cannot be found", t.Value)
		}
		return []*world.Entity{p}, nil
	}
	return nil, exec.Errorf("invalid target")
}

// selectorText returns the selector as it was written.
//
// 1.12.2 quotes the original text in its error messages, so the raw
// spelling is preferred; the rebuilt form is only a fallback for
// selectors that were assembled by hand.
func selectorText(sel *ast.Selector) string {
	if sel.Raw != "" {
		return sel.Raw
	}
	var b strings.Builder
	b.WriteString("@")
	b.WriteString(sel.Base)
	if len(sel.Args) == 0 {
		return b.String()
	}
	b.WriteString("[")
	for i, arg := range sel.Args {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(arg.Key)
		b.WriteString("=")
		b.WriteString(arg.Value)
	}
	b.WriteString("]")
	return b.String()
}
