package v1_12_2

import (
	"strings"

	"github.com/mcsim/gomc/internal/exec"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/world"
)

// expandMessage replaces the selectors in a message with the names of the
// entities they match.
//
// 1.12.2 splits the message into tokens and tries to expand every one of
// them (`CommandAbstract.b`): a token that looks like a selector is
// replaced by the names of the entities it matches, and a selector that
// matches nothing is an error.
//
// The tokens are joined with single spaces, so the message is rebuilt
// rather than patched.
func expandMessage(ctx *exec.Context, msg string) (string, error) {
	tokens := strings.Fields(msg)
	parts := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		if !parser.IsSelector(tok) {
			parts = append(parts, tok)
			continue
		}
		sel, err := parser.ParseSelector(tok)
		if err != nil {
			return "", exec.Errorf("%v", err)
		}
		resolver := &exec.SelectorResolver{World: ctx.World}
		entities, err := resolver.Resolve(ctx.Source, sel)
		selectorWarnings(ctx, resolver)
		if err != nil {
			return "", exec.Errorf("%v", err)
		}
		if len(entities) == 0 {
			return "", exec.Errorf("Selector '%s' found nothing", tok)
		}
		names := make([]string, 0, len(entities))
		for _, e := range entities {
			names = append(names, scoreboardDisplayName(e))
		}
		// 1.12.2 hands the names to `CommandAbstract.a(List<IChatBaseComponent>)`
		// (CommandAbstract.java:510), which puts " and " in front of the
		// last name and ", " between the others. So two names read
		// "A and B", three read "A, B and C" and four read
		// "A, B, C and D". That is the same rule `joinVanillaList`
		// implements.
		//
		// (Measured: `say @e[type=armor_stand]` with three armour
		// stands named A, B and C prints "A, B and C".)
		parts = append(parts, joinVanillaList(names))
	}
	return strings.Join(parts, " "), nil
}

// scoreboardDisplayName returns the name the scoreboard shows for an
// entity.
//
// 1.12.2 uses `Entity.getScoreboardDisplayName`, which is the custom name
// when there is one and the type's translated name otherwise.
func scoreboardDisplayName(e *world.Entity) string {
	return e.DisplayName()
}
