package v1_12_2

import (
	"strings"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
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
// 1.12.2 says three different things about a target it cannot resolve:
//
//   - a target *list* selector that matches nothing is **not** an error:
//     the dispatcher expands the list and the command runs zero times, so
//     the server says nothing at all
//   - a selector that is limited to a single entity is resolved by the
//     command itself, which reports "Entity '<text>' cannot be found" or
//     "Player '<text>' cannot be found", depending on the kind of target
//     the command takes
//   - a player name that is not online is reported by the command itself
//     as "Player '<name>' cannot be found"
func resolveEntityTargets(ctx *exec.Context, arg ast.Arg) ([]*world.Entity, error) {
	return resolveTarget(ctx, arg, entityTargetCommand(ctx))
}

// resolveEntityTargetsEntity resolves a target argument into entities the
// way the commands that take an *entity* do.
//
// The difference is the wording for a name that is not a known entity:
// `CommandAbstract.a(server, listener, s, Entity.class)` throws
// `ExceptionEntityNotFound`, so the message is "Entity '<name>' cannot be
// found" and not the player wording of `resolveEntityTargets`.
func resolveEntityTargetsEntity(ctx *exec.Context, arg ast.Arg) ([]*world.Entity, error) {
	return resolveTarget(ctx, arg, true)
}

// entityTargetCommands are the commands whose target 1.12.2 resolves as
// an entity.
//
// They call `CommandAbstract.c(server, listener, s)`, which ends in
// `CommandAbstract.a(server, listener, s, Entity.class)` and throws
// `ExceptionEntityNotFound`, so a name that is not an entity is reported
// as "Entity '<name>' cannot be found". The other commands resolve a
// *player* with `CommandAbstract.b(server, listener, s)`, which reports
// "Player '<name>' cannot be found".
//
// The table also decides what an empty target *list* means. 1.12.2
// expands such a list in the command dispatcher, and the command then
// runs once per matched entity. Measured on a real server, a list that
// matches nothing makes the command run zero times and say nothing at
// all -- but only the commands that resolve an entity do that; the ones
// that resolve a player still report the selector as empty.
var entityTargetCommands = map[string]bool{
	// CommandExecute: CommandAbstract.a(..., Entity.class)
	"execute": true,
	// CommandEntityData: CommandAbstract.c(...)
	"entitydata": true,
	// CommandKill: CommandAbstract.c(...)
	"kill": true,
	// CommandTestFor: CommandAbstract.c(...)
	"testfor": true,
	// CommandTp and CommandTeleport: CommandAbstract.c(...)
	"tp":       true,
	"teleport": true,
	// CommandEffect: CommandAbstract.a(..., EntityLiving.class)
	"effect": true,
	// CommandEnchant: CommandAbstract.a(..., EntityLiving.class)
	"enchant": true,
}

// entityTargetCommand reports whether the command being run resolves its
// target as an entity rather than as a player.
func entityTargetCommand(ctx *exec.Context) bool {
	if ctx == nil || ctx.Command == nil {
		return false
	}
	return entityTargetCommands[ctx.Command.Name]
}

// selectorWarnings copies the notes a selector produced into the output of
// the command.
//
// 1.12.2 sends them next to the command's own output: `kill
// @e[type=pigzombie]` answers "Entity type 'minecraft:pigzombie' is
// invalid" and then succeeds.
func selectorWarnings(ctx *exec.Context, r *exec.SelectorResolver) {
	for _, w := range r.Warnings {
		ctx.Engine.Feedback = append(ctx.Engine.Feedback, w)
	}
}

// resolveTarget resolves a target argument into entities.
//
// entityVariant selects the wording of the failure message for a bare
// name: the entity variant for the commands that resolve an entity, the
// player variant for the commands that resolve a player.
func resolveTarget(ctx *exec.Context, arg ast.Arg, entityVariant bool) ([]*world.Entity, error) {
	switch t := arg.(type) {
	case *ast.Selector:
		resolver := &exec.SelectorResolver{World: ctx.World}
		entities, err := resolver.Resolve(ctx.Source, t)
		selectorWarnings(ctx, resolver)
		if err != nil {
			return nil, exec.Errorf("%v", err)
		}
		if len(entities) == 0 {
			if selectorIsList(t) {
				// A target *list* is expanded by the command dispatcher:
				// the command runs once per matched entity. When the
				// list is empty the command runs zero times, and the
				// server says nothing at all.
				//
				// (Measured: `kill @e[type=zombie]` on a server without
				// zombies prints no message and counts as a command
				// that affected no entity; an enclosing `execute`
				// reports "All invocations failed: 'kill @e[type=zombie]'"
				// instead of the selector being reported as empty.)
				//
				// That holds for every command, not only for the ones
				// that resolve an entity: `gamemode creative @a` on a
				// server without players is just as silent.
				return nil, nil
			}
			// A selector that is limited to one entity is resolved by
			// the command, which reports the missing target. The wording
			// follows the kind of target the command takes: the ones
			// that resolve an entity say "Entity", the ones that
			// resolve a player say "Player".
			//
			// (Measured: `kill @p` says "Entity '@p' cannot be found"
			// and `gamemode creative @p` says "Player '@p' cannot be
			// found".)
			if entityVariant {
				return nil, exec.Errorf("Entity '%s' cannot be found", selectorText(t))
			}
			return nil, exec.Errorf("Player '%s' cannot be found", selectorText(t))
		}
		return entities, nil
	case *ast.StringArg:
		p := ctx.World.FindPlayer(t.Value)
		if p == nil {
			if entityVariant {
				return nil, exec.Errorf("Entity '%s' cannot be found", t.Value)
			}
			return nil, exec.Errorf("Player '%s' cannot be found", t.Value)
		}
		return []*world.Entity{p}, nil
	}
	return nil, exec.Errorf("invalid target")
}

// selectorIsList reports whether a selector is a target list.
//
// 1.12.2 decides this in `PlayerSelector.isList`: the `c` argument counts
// the entities, and everything that is not limited to a single entity is
// a list. The default count is 0 (no limit) for `@a` and `@e` and 1 for
// `@p`, `@r` and `@s`, so `@e[c=1]` is *not* a list while `@a[c=1]` is
// not one either.
func selectorIsList(sel *ast.Selector) bool {
	def := 1
	if sel.Base == "a" || sel.Base == "e" {
		def = 0
	}
	return parser.SelectorIntDefault(sel, "c", def) != 1
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
