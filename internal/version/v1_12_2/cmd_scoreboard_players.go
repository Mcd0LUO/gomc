package v1_12_2

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
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
	// 1.12.2 counts the arguments of the subcommand **before** it looks at
	// any of them (`CommandScoreboard.execute`), and every subcommand has
	// its own usage line in the language file.
	//
	// The count is the one of the server: the word "players" and the
	// action are part of it.
	n := p.Count() + 2
	switch action {
	case "set", "add", "remove":
		if n < 5 {
			return nil, playersUsageError(p, "scoreboard players "+action)
		}
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
		// The dataTag is the last argument of the three actions and is
		// optional. It is kept as text and only parsed when the command
		// runs, because 1.12.2 parses it *after* it has looked the
		// objective up: a wrong objective is reported before a wrong
		// dataTag.
		dataTag := ""
		if p.More() {
			dataTag = p.Rest()
		}
		return []ast.Arg{
			&ast.StringArg{Value: "players"},
			&ast.StringArg{Value: action},
			holder,
			&ast.StringArg{Value: objective},
			&ast.IntArg{Value: value},
			&ast.StringArg{Value: dataTag},
		}, nil
	case "reset":
		if n != 3 && n != 4 {
			return nil, playersUsageError(p, "scoreboard players reset")
		}
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
		if n != 7 {
			return nil, playersUsageError(p, "scoreboard players operation")
		}
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
		if n != 4 {
			return nil, playersUsageError(p, "scoreboard players enable")
		}
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
		// `scoreboard players list [name]`: without a name 1.12.2
		// lists every player the scoreboard tracks.
		holder, err := parseOptionalScoreHolder(p)
		if err != nil {
			return nil, err
		}
		if p.More() {
			return nil, playersUsageError(p, "scoreboard players list")
		}
		return []ast.Arg{
			&ast.StringArg{Value: "players"},
			&ast.StringArg{Value: action},
			holder,
		}, nil
	case "test":
		if n != 5 && n != 6 {
			return nil, playersUsageError(p, "scoreboard players test")
		}
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
		// `scoreboard players tag <target> <add|remove|list> [tag] [dataTag]`
		// is how 1.12.2 manages the tags of an entity.
		if n < 4 {
			return nil, playersUsageError(p, "scoreboard players tag")
		}
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
			// 1.12.2 resolves the entity first and only then looks at the
			// mode, so the usage it reports is the one of the whole
			// subcommand.
			return nil, playersUsageError(p, "scoreboard players tag")
		}
		// The dataTag of a tag action is only read for `add` and
		// `remove`: 1.12.2 hands the `list` branch the tags of the
		// entity and never looks at the rest of the command.
		dataTag := ""
		if sub != "list" && p.More() {
			dataTag = p.Rest()
		}
		return []ast.Arg{
			&ast.StringArg{Value: "players"},
			&ast.StringArg{Value: "tag"},
			holder,
			&ast.StringArg{Value: sub},
			&ast.StringArg{Value: tag},
			&ast.StringArg{Value: dataTag},
		}, nil
	}
	return nil, playersUsageError(p, "scoreboard players")
}

// playersUsageError reports a usage error of a `scoreboard players`
// subcommand.
//
// 1.12.2 counts the arguments of the subcommand before it looks at any of
// them, and every subcommand has its own usage line in the language file.
func playersUsageError(p *parser.Parser, key string) error {
	return usageError(p, key, "Usage: %s", usageLine(key))
}

// parseScoreHolder parses a score holder, which is a player name or a
// selector.
func parseScoreHolder(p *parser.Parser) (ast.Arg, error) {
	t, ok := p.Peek()
	if !ok {
		return nil, usageError(p, "scoreboard players set", "expected a player or target")
	}
	if IsSelector(t.Text) {
		sel, err := parseSelector(p)
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

// tagEntity finds the entity a tag action applies to.
//
// The holder of a tag action is an entity id when it comes from a
// selector, and a player name when it was written by hand: 1.12.2
// resolves the name with `CommandAbstract.c`, which finds the player
// first and the entity by its UUID second.
func tagEntity(ctx *exec.Context, holder string) *world.Entity {
	if e := ctx.World.FindEntity(holder); e != nil {
		return e
	}
	if p := ctx.World.FindPlayer(holder); p != nil {
		return p
	}
	return entityByUUID(ctx, holder)
}

// scoreHolderEntity returns the entity a score holder names.
//
// 1.12.2 resolves the holder of a scoreboard command with
// `CommandAbstract.c`: the player with that name first, and then the
// entity with that UUID. The runtime's own entity ids are looked up as
// well, because they are how a function names an entity.
//
// A holder that names no entity at all is reported by the caller.
func scoreHolderEntity(ctx *exec.Context, holder string) *world.Entity {
	if p := ctx.World.FindPlayer(holder); p != nil {
		return p
	}
	if e := ctx.World.FindEntity(holder); e != nil {
		return e
	}
	return entityByUUID(ctx, holder)
}

// checkDataTag checks the dataTag of a scoreboard command against the
// entity a holder names.
//
// 1.12.2 compares the dataTag with the NBT of the entity the holder
// names, with the same subset match the `nbt=` argument of a selector
// uses: every key of the dataTag has to be in the entity's NBT with a
// matching value. A dataTag that does not match is reported with the name
// the scoreboard knows the entity by, which is the UUID of an entity that
// is not a player.
//
// parsePrefix is the wording the command uses for a dataTag that could
// not be parsed: the two commands that take a dataTag word it
// differently.
func checkDataTag(ctx *exec.Context, holder, text, parsePrefix string) error {
	e := scoreHolderEntity(ctx, holder)
	if e == nil {
		return entityNotFoundError(holder)
	}
	filter, err := parseDataTag(text)
	if err != nil {
		return exec.Errorf("%s%s", parsePrefix, err)
	}
	if !nbt.Matches(e.NBT, filter) {
		return exec.Errorf("The dataTag does not match for %s", ScoreboardHolder(e))
	}
	return nil
}

// parseDataTag parses the dataTag of a scoreboard command.
//
// 1.12.2 parses it with `MojangsonParser`, which only accepts a compound.
// The error is the *reason* of the failure, without the wording of the
// command that reports it.
func parseDataTag(text string) (*nbt.Compound, error) {
	if text == "" {
		return nil, errors.New("Expected compound tag, got nothing at: <--[HERE]")
	}
	// The parser looks at the first character before it reads anything
	// else, so a dataTag that does not start with a brace is reported
	// that way and not as a broken compound.
	if text[0] != '{' {
		return nil, errors.New(mojangsonNotCompound(text))
	}
	v, err := nbt.ParseSNBT(text)
	if err != nil {
		return nil, errors.New(mojangsonReason(text, err))
	}
	c, ok := v.(*nbt.Compound)
	if !ok {
		return nil, errors.New(mojangsonNotCompound(text))
	}
	return c, nil
}

// mojangsonReason renders the reason 1.12.2 reports a dataTag it could
// not parse.
//
// `MojangsonParseException` builds its message out of the reason and the
// text up to the place the parser stopped at:
//
//	<reason> at: <text><--[HERE]
//
// The runtime's own NBT parser reports the same kind of problem with its
// own wording, so the reason is translated here. Every wording below was
// measured on a real 1.12.2 server.
func mojangsonReason(text string, err error) string {
	e, ok := err.(*nbt.SNBTError)
	if !ok {
		return err.Error()
	}
	reason, pos := mojangsonReasonAt(text, e.Msg, e.Offset)
	if pos > len(text) {
		pos = len(text)
	}
	if pos < 0 {
		pos = 0
	}
	return reason + " at: " + mojangsonText(text, pos) + "<--[HERE]"
}

// mojangsonText renders the text up to the place the parser stopped at.
//
// `MojangsonParseException` shows at most the last 35 characters of it, and
// a leading "..." when there are more.
//
// (Measured on a real 1.12.2 server: a dataTag of more than 35 characters
// is shown as "...:1,Key2:2,Key3:3,Key4:4,Key5:5,Key6<--[HERE]".)
func mojangsonText(text string, pos int) string {
	const maxChars = 35
	if pos <= maxChars {
		return text[:pos]
	}
	return "..." + text[pos-maxChars:pos]
}

// mojangsonReasonAt translates the failure of the runtime's NBT parser
// into the reason 1.12.2 reports, and returns the place in the text the
// parser stopped at.
func mojangsonReasonAt(text, msg string, off int) (string, int) {
	switch {
	case strings.HasPrefix(msg, "expected ':' after key"):
		if off < len(text) {
			return fmt.Sprintf("Expected ':' but got '%c'", text[off]), off + 1
		}
		return "Expected ':' but got '<EOF>'", off
	case msg == "expected value":
		// A value that is missing inside a list is a missing ']';
		// everywhere else it is a missing value.
		if off < len(text) && (text[off] == '}' || text[off] == ']') &&
			innermostOpen(text, off) == '[' {
			return fmt.Sprintf("Expected ']' but got '%c'", text[off]), off + 1
		}
		return "Expected value", off
	case msg == "unterminated string":
		return "Missing termination quote", off
	case msg == "unterminated compound":
		if strings.HasSuffix(strings.TrimSpace(text), ",") {
			// 1.12.2 expects a key after the comma.
			return "Expected key", off
		}
		return "Expected '}' but got '<EOF>'", off
	case msg == "unterminated list", msg == "unterminated byte array",
		msg == "unterminated int array":
		return "Expected ']' but got '<EOF>'", off
	case msg == "expected key":
		return "Expected non-empty key", off
	case msg == "trailing characters after value":
		return "Trailing data found", off + 1
	case strings.HasPrefix(msg, "invalid "):
		// A number the runtime's parser refuses. 1.12.2 reads it as a
		// string instead, so the dataTag is compared and does not
		// match; the runtime reports it as a broken value.
		return "Expected value", off
	}
	return "Expected value", off
}

// mojangsonNotCompound renders the reason 1.12.2 reports a dataTag that
// is not a compound.
//
// Only a compound is a dataTag, and the parser stops at the first
// character and names it.
func mojangsonNotCompound(text string) string {
	if text == "" {
		return "Expected compound tag, got nothing at: <--[HERE]"
	}
	return fmt.Sprintf("Expected '{' but got '%c' at: %c<--[HERE]", text[0], text[0])
}

// innermostOpen returns the bracket that is open at the given place of
// the text, or 0 when there is none.
//
// It tells a value that is missing inside a list from one that is missing
// inside a compound: 1.12.2 words the two differently.
func innermostOpen(text string, off int) byte {
	var stack []byte
	quote := byte(0)
	for i := 0; i < off && i < len(text); i++ {
		c := text[i]
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		switch c {
		case '"', '\'':
			quote = c
		case '{', '[':
			stack = append(stack, c)
		case '}', ']':
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	if len(stack) == 0 {
		return 0
	}
	return stack[len(stack)-1]
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
		dataTag := args[5].(*ast.StringArg).Value
		// 1.12.2 expands a target list in the command dispatcher and
		// runs the command once per matched entity, so a holder whose
		// dataTag does not match does not stop the holders after it.
		var dataTagErr error
		for _, h := range holders {
			// 1.12.2 checks the dataTag before it writes the score,
			// and it checks it against the NBT of the entity the
			// holder names. A dataTag that does not match leaves the
			// score alone.
			if dataTag != "" {
				if err := checkDataTag(ctx, h, dataTag, "Could not parse dataTag, reason: "); err != nil {
					if dataTagErr == nil {
						dataTagErr = err
					}
					continue
				}
			}
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
			// 1.12.2 confirms every holder separately, and the score
			// it reports is the one *after* the change. The message is
			// the same for set, add and remove.
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Set score of %s for player %s to %d", objective, h, next))
		}
		if dataTagErr != nil {
			return 0, dataTagErr
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
			if objective == "" {
				// Without an objective every score of the holder is
				// reset, and 1.12.2 says so in the plural.
				ctx.Engine.Feedback = append(ctx.Engine.Feedback,
					fmt.Sprintf("Reset scores of player %s", h))
				continue
			}
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
			// `enable` unlocks the trigger, and it creates the score
			// when the holder has none yet: 1.12.2 calls
			// `getPlayerScoreForObjective`, which makes the entry with
			// a score of 0, and then unlocks it. An entry that is
			// already there keeps its score.
			if _, tracked := sb.Score(h, trigger); !tracked {
				sb.SetScore(h, trigger, 0)
			}
			sb.SetScoreLocked(h, trigger, false)
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Enabled trigger %s for %s", trigger, h))
		}
		return len(holders), nil
	case "list":
		if s, ok := args[2].(*ast.StringArg); ok && s.Value == "" {
			// Without a name 1.12.2 lists every player the
			// scoreboard tracks, as one line of names.
			players := sb.TrackedPlayers()
			if len(players) == 0 {
				return 0, exec.Errorf("There are no tracked players on the scoreboard")
			}
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				fmt.Sprintf("Showing %d tracked players on the scoreboard:", len(players)))
			ctx.Engine.Feedback = append(ctx.Engine.Feedback, joinVanillaList(players))
			return 1, nil
		}
		holders, err := resolveHolders(ctx, args[2])
		if err != nil {
			return 0, err
		}
		for _, h := range holders {
			// 1.12.2 prints a heading and then one line per tracked
			// objective. The line shows the *display name* of the
			// objective, the score and the name of the objective; the
			// criterion is not part of it.
			//
			// The tracked objectives are read from the index of the
			// scoreboard instead of from a scan of every objective:
			// the index already holds them in the order the
			// objectives were created.
			tracked := sb.TrackedObjectives(h)
			if len(tracked) == 0 {
				// A player without a score has no listing at all.
				return 0, exec.Errorf("Player %s has no scores recorded", h)
			}
			ctx.Engine.Feedback = append(ctx.Engine.Feedback,
				"Showing "+strconv.Itoa(len(tracked))+" tracked objective(s) for "+h+":")
			for _, t := range tracked {
				ctx.Engine.Feedback = append(ctx.Engine.Feedback,
					"- "+objectiveDisplayName(t.Objective)+": "+strconv.Itoa(t.Score(h))+
						" ("+t.Objective.Name+")")
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
		dataTag := args[5].(*ast.StringArg).Value
		switch action {
		case "add", "remove":
			tag := args[4].(*ast.StringArg).Value
			// 1.12.2 expands a target list in the command dispatcher
			// and runs the command once per matched entity, so a
			// holder whose dataTag does not match does not stop the
			// holders after it.
			var dataTagErr error
			for _, h := range holders {
				e := tagEntity(ctx, h)
				if e == nil {
					// 1.12.2 resolves the holder of a tag action as an
					// entity, and it does that before it looks at the
					// dataTag -- so a holder it cannot find is reported
					// whether there is a dataTag or not.
					//
					// (Measured on a real 1.12.2 server: `scoreboard
					// players tag Steve add x` says "Entity 'Steve' cannot
					// be found".)
					return 0, entityNotFoundError(h)
				}
				if dataTag != "" {
					if err := checkDataTag(ctx, h, dataTag, "Players tag command failed, reason: "); err != nil {
						if dataTagErr == nil {
							dataTagErr = err
						}
						continue
					}
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
			if dataTagErr != nil {
				return 0, dataTagErr
			}
			return len(holders), nil
		case "list":
			// 1.12.2 prints a heading and then the tags joined
			// with " and " on the next line -- but only when the
			// entity carries tags at all. An entity without tags
			// makes the command report nothing and the number of
			// tags it has.
			//
			// (CommandScoreboard.java:735-741: the heading and the
			// tag list are sent only when the set of tags is not
			// empty, and the count the command reports is the size
			// of that set.)
			tags := 0
			for _, h := range holders {
				e := tagEntity(ctx, h)
				if e == nil {
					// The holder of a tag action is resolved as an entity,
					// and the `list` branch does that as well.
					//
					// (Measured on a real 1.12.2 server: `scoreboard players
					// tag Steve list` answers "Entity 'Steve' cannot be
					// found".)
					return 0, entityNotFoundError(h)
				}
				// The heading names the entity the way the scoreboard
				// does: `CommandAbstract.f` answers a player with its
				// user name and every other entity with its UUID.
				list := e.TagList()
				if len(list) == 0 {
					continue
				}
				ctx.Engine.Feedback = append(ctx.Engine.Feedback,
					fmt.Sprintf("Tags on entity %s are:", ScoreboardHolder(e)))
				ctx.Engine.Feedback = append(ctx.Engine.Feedback,
					joinVanillaList(list))
				tags += len(list)
			}
			return tags, nil
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
