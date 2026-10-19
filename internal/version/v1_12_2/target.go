package v1_12_2

import (
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
// among the players. 1.12.2 reports "no matching entities" when nothing
// is found.
func resolveEntityTargets(ctx *exec.Context, arg ast.Arg) ([]*world.Entity, error) {
	switch t := arg.(type) {
	case *ast.Selector:
		resolver := &exec.SelectorResolver{World: ctx.World}
		entities, err := resolver.Resolve(ctx.Source, t)
		if err != nil {
			return nil, exec.Errorf("%v", err)
		}
		return entities, nil
	case *ast.StringArg:
		p := ctx.World.FindPlayer(t.Value)
		if p == nil {
			return nil, exec.Errorf("no matching entities")
		}
		return []*world.Entity{p}, nil
	}
	return nil, exec.Errorf("invalid target")
}
