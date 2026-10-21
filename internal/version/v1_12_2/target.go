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
	if IsSelector(t.Text) {
		return parseSelector(p)
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

// resolvePlayerTargets resolves a target the way the commands that only work
// on a player do.
//
// 1.12.2 asks the selector for a *player* and keeps only the players it
// finds, so a target that matches entities but no player is not a player
// and the command refuses it. The refusal names the entity that was found
// by its UUID.
//
// A target that matches nothing at all is not an error: the command runs
// zero times and the server says nothing.
//
// (Measured on a real 1.12.2 server: `give @e[type=armor_stand] stone` on
// a server with an armour stand answers "Player '<the uuid of the armour
// stand>' cannot be found", while `give @e[type=zombie] stone` on a server
// without zombies prints nothing at all.)
func resolvePlayerTargets(ctx *exec.Context, arg ast.Arg) ([]*world.Entity, error) {
	switch t := arg.(type) {
	case *ast.Selector:
		resolver := &SelectorResolver{World: ctx.World}
		entities, err := resolver.Resolve(ctx.Source, t)
		selectorWarnings(ctx, resolver)
		if err != nil {
			return nil, exec.Errorf("%v", err)
		}
		players := make([]*world.Entity, 0, len(entities))
		for _, e := range entities {
			if e.IsPlayer {
				players = append(players, e)
			}
		}
		if len(players) == 0 {
			if len(entities) > 0 {
				// The target matched entities, but none of them is a
				// player. 1.12.2 names the first of them by its UUID.
				return nil, exec.Errorf("Player '%s' cannot be found", entities[0].EnsureUUID())
			}
			if !selectorIsList(t) {
				// A target that is limited to a single entity and finds
				// nothing is reported by the command itself.
				return nil, exec.Errorf("Player '%s' cannot be found", selectorText(t))
			}
			// A target list that matches nothing makes the command run
			// zero times, and the server says nothing at all.
			return nil, nil
		}
		return players, nil
	case *ast.StringArg:
		if p := ctx.World.FindPlayer(t.Value); p != nil {
			return []*world.Entity{p}, nil
		}
		return nil, exec.Errorf("Player '%s' cannot be found", t.Value)
	}
	return nil, exec.Errorf("invalid target")
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
func selectorWarnings(ctx *exec.Context, r *SelectorResolver) {
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
		resolver := &SelectorResolver{World: ctx.World}
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
				// 1.12.2 resolves an entity target by its UUID as
				// well: `CommandAbstract.c` looks the name up as a
				// player first and then parses it as a UUID.
				if e := entityByUUID(ctx, t.Value); e != nil {
					return []*world.Entity{e}, nil
				}
				return nil, entityNotFoundError(t.Value)
			}
			return nil, exec.Errorf("Player '%s' cannot be found", t.Value)
		}
		return []*world.Entity{p}, nil
	}
	return nil, exec.Errorf("invalid target")
}

// entityByUUID returns the entity with the given UUID.
//
// 1.12.2 looks an entity up by its UUID in `CommandAbstract.c`, which is
// how `scoreboard players tag <uuid> add x` and `kill <uuid>` find their
// target. The runtime's entities carry a UUID derived from their internal
// id, so the lookup compares against it.
func entityByUUID(ctx *exec.Context, uuid string) *world.Entity {
	if uuid == "" {
		return nil
	}
	for _, e := range ctx.World.EntitySlice() {
		if e.EnsureUUID() == uuid {
			return e
		}
	}
	return nil
}

// entityNotFoundError builds the error 1.12.2 answers an entity target it
// cannot resolve with.
//
// A name that is written like a UUID but is not one is reported as a
// malformed UUID: `CommandAbstract.c` parses the name with
// `UUID.fromString`, and only a name that fails that parse *and* has the
// five groups a UUID has is called out as malformed.
//
// (Measured: `scoreboard players set a-b-c-d-e ...` says "Entity
// 'a-b-c-d-e' cannot be found", because every group of it is a hex number
// and so it parses as a UUID -- one that no entity has.)
func entityNotFoundError(name string) error {
	if malformedUUID(name) {
		return exec.Errorf("The entity UUID '%s' is in an invalid format", name)
	}
	return exec.Errorf("Entity '%s' cannot be found", name)
}

// malformedUUID reports whether a name is written like a UUID but cannot
// be parsed as one.
//
// `java.util.UUID.fromString` is lenient: it wants five groups of hex
// digits separated by dashes, and it does not look at how long they are.
func malformedUUID(name string) bool {
	groups := strings.Split(name, "-")
	if len(groups) != 5 {
		return false
	}
	for _, g := range groups {
		if g == "" {
			return true
		}
		for i := 0; i < len(g); i++ {
			c := g[i]
			switch {
			case c >= '0' && c <= '9':
			case c >= 'a' && c <= 'f':
			case c >= 'A' && c <= 'F':
			default:
				return true
			}
		}
	}
	return false
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
	return SelectorIntDefault(sel, "c", def) != 1
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
