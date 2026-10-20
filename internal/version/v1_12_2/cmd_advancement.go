package v1_12_2

import (
	"fmt"
	"strconv"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// advancementCmd implements `/advancement`.
//
// The 1.12.2 command has three subtrees:
//
//	advancement grant  <player> <only|until|from|through|everything>
//	advancement revoke <player> <only|until|from|through|everything>
//	advancement test   <player> <advancement> [criterion]
//
// The two change subtrees work on the whole advancement or on a single
// criterion of it, and the filter decides how far the change reaches:
// "only" touches the advancement itself, "until" also touches its
// ancestors, "from" also touches its descendants and "through" touches
// both.
type advancementCmd struct{}

// Name implements exec.Handler.
func (c *advancementCmd) Name() string { return "advancement" }

// Usage implements exec.Handler.
func (c *advancementCmd) Usage() string { return "/advancement <grant|revoke|test> <player>" }

// Parse implements exec.Handler.
//
// Only the shape of the command is checked here. The player and the
// advancement are looked up when the command runs, because 1.12.2 does the
// same: it resolves the player before it looks at the filter, so
// `/advancement grant <unknown player> bogus` complains about the player
// and not about the filter.
func (c *advancementCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	action, err := p.String()
	if err != nil {
		return nil, advancementUsageError(commandUsage("advancement"))
	}
	switch action {
	case "grant", "revoke":
		return parseAdvancementChange(p, action)
	case "test":
		return parseAdvancementTest(p)
	}
	return nil, advancementUsageError(commandUsage("advancement"))
}

// parseAdvancementChange parses `grant` and `revoke`.
//
// Both take a player, a filter and, unless the filter is "everything", the
// name of an advancement. The "only" filter takes a criterion as well.
func parseAdvancementChange(p *parser.Parser, action string) ([]ast.Arg, error) {
	player, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	filter, err := p.String()
	if err != nil {
		// 1.12.2 answers a missing filter with the usage of the
		// subcommand, without the "Usage: " prefix.
		return nil, advancementPlainUsageError(commandUsage("advancement " + action))
	}
	args := []ast.Arg{&ast.StringArg{Value: action}, player, &ast.StringArg{Value: filter}}
	if !p.More() {
		return args, nil
	}
	advancement, err := p.String()
	if err != nil {
		return nil, err
	}
	args = append(args, &ast.StringArg{Value: advancement})
	if !p.More() {
		return args, nil
	}
	criterion, err := p.String()
	if err != nil {
		return nil, err
	}
	return append(args, &ast.StringArg{Value: criterion}), nil
}

// parseAdvancementTest parses `test`.
//
// It takes a player, the name of an advancement and, optionally, the name
// of one of its criteria.
func parseAdvancementTest(p *parser.Parser) ([]ast.Arg, error) {
	player, err := parseEntityTarget(p)
	if err != nil {
		return nil, err
	}
	advancement, err := p.String()
	if err != nil {
		return nil, advancementUsageError(commandUsage("advancement test"))
	}
	args := []ast.Arg{&ast.StringArg{Value: "test"}, player, &ast.StringArg{Value: advancement}}
	if !p.More() {
		return args, nil
	}
	criterion, err := p.String()
	if err != nil {
		return nil, err
	}
	return append(args, &ast.StringArg{Value: criterion}), nil
}

// Run implements exec.Handler.
func (c *advancementCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	action := args[0].(*ast.StringArg).Value
	if action == "test" {
		return runAdvancementTest(ctx, args)
	}
	return runAdvancementChange(ctx, args, action)
}

// runAdvancementTest runs `/advancement test`.
func runAdvancementTest(ctx *exec.Context, args []ast.Arg) (int, error) {
	player, err := resolveAdvancementPlayer(ctx, args[1])
	if err != nil {
		return 0, err
	}
	name := args[2].(*ast.StringArg).Value
	advancement, ok := lookupAdvancement(name)
	if !ok {
		return 0, exec.Errorf("No advancement was found by the name '%s'", name)
	}
	store := advancementStoreOf(ctx.World)
	obtained := store.obtained(advancementPlayerName(player), advancement.name)
	if len(args) < 4 {
		// The whole advancement is tested.
		if !advancement.done(obtained) {
			return 0, exec.Errorf("Player %s has not completed advancement '%s'",
				advancementPlayerName(player), advancement.name)
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Player %s has completed advancement '%s'",
				advancementPlayerName(player), advancement.name))
		return 1, nil
	}
	criterion := args[3].(*ast.StringArg).Value
	if !advancement.hasCriterion(criterion) {
		return 0, exec.Errorf("The advancement '%s' does not contain the criterion '%s'",
			advancement.name, criterion)
	}
	if !obtained[criterion] {
		return 0, exec.Errorf("Player %s has not completed criterion '%s' of advancement '%s'",
			advancementPlayerName(player), criterion, advancement.name)
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Player %s has completed criterion '%s' of advancement '%s'",
			advancementPlayerName(player), criterion, advancement.name))
	return 1, nil
}

// runAdvancementChange runs `/advancement grant` and `/advancement revoke`.
func runAdvancementChange(ctx *exec.Context, args []ast.Arg, action string) (int, error) {
	player, err := resolveAdvancementPlayer(ctx, args[1])
	if err != nil {
		return 0, err
	}
	filter := args[2].(*ast.StringArg).Value
	if !isAdvancementFilter(filter) {
		return 0, advancementPlainUsageError(commandUsage("advancement " + action))
	}
	playerName := advancementPlayerName(player)
	store := advancementStoreOf(ctx.World)

	if filter == "everything" {
		// "everything" takes no advancement, so any further argument is a
		// mistake.
		if len(args) != 3 {
			return 0, advancementPlainUsageError(commandUsage("advancement " + action + " everything"))
		}
		count := 0
		for i := range advancementRegistry {
			if changeAdvancement(store, playerName, action, &advancementRegistry[i]) {
				count++
			}
		}
		if count == 0 {
			return 0, exec.Errorf("%s", advancementEverythingFailed(action, playerName))
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			advancementEverythingSucceeded(action, count, playerName))
		return count, nil
	}

	// Every other filter needs the name of an advancement.
	if len(args) < 4 {
		return 0, advancementPlainUsageError(commandUsage("advancement " + action + " " + filter))
	}
	name := args[3].(*ast.StringArg).Value
	advancement, ok := lookupAdvancement(name)
	if !ok {
		return 0, exec.Errorf("No advancement was found by the name '%s'", name)
	}

	// The "only" filter may name a single criterion.
	if filter == "only" && len(args) == 5 {
		criterion := args[4].(*ast.StringArg).Value
		if !advancement.hasCriterion(criterion) {
			return 0, exec.Errorf("The advancement '%s' does not contain the criterion '%s'",
				advancement.name, criterion)
		}
		if !changeCriterion(store, playerName, action, advancement.name, criterion) {
			return 0, exec.Errorf("%s", advancementCriterionFailed(action, advancement.name, playerName, criterion))
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			advancementCriterionSucceeded(action, advancement.name, playerName, criterion))
		return 1, nil
	}
	if len(args) != 4 {
		return 0, advancementPlainUsageError(commandUsage("advancement " + action + " " + filter))
	}

	count := 0
	for _, a := range advancementFilterTargets(advancement, filter) {
		if changeAdvancement(store, playerName, action, a) {
			count++
		}
	}
	if count == 0 {
		return 0, exec.Errorf("%s", advancementFilterFailed(action, filter, advancement.name, playerName))
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		advancementFilterSucceeded(action, filter, advancement.name, count, playerName))
	return count, nil
}

// advancementFilterTargets returns the advancements a filter reaches.
//
// 1.12.2 walks the parents of the advancement first, then the advancement
// itself, then its descendants. The order matters only for the order in
// which the advancements are granted.
func advancementFilterTargets(advancement *advancementDef, filter string) []*advancementDef {
	var out []*advancementDef
	switch filter {
	case "until", "through":
		out = append(out, ancestors(advancement)...)
	}
	out = append(out, advancement)
	switch filter {
	case "from", "through":
		out = append(out, advancementDescendants(advancement)...)
	}
	return out
}

// advancementDescendants returns the descendants of an advancement.
//
// 1.12.2 walks the tree depth first, and the children of an advancement
// are the ones that name it as their parent.
func advancementDescendants(advancement *advancementDef) []*advancementDef {
	var out []*advancementDef
	for _, child := range children(advancement.name) {
		out = append(out, child)
		out = append(out, advancementDescendants(child)...)
	}
	return out
}

// changeAdvancement changes a whole advancement and reports whether it
// changed.
//
// 1.12.2 grants every criterion that is not done yet, unless the
// advancement is already complete, and it takes away every criterion that
// is done, unless the player never started the advancement.
func changeAdvancement(store *advancementStore, player, action string, advancement *advancementDef) bool {
	obtained := store.obtained(player, advancement.name)
	if action == "grant" {
		if advancement.done(obtained) {
			return false
		}
		for _, criterion := range advancement.criteria {
			store.grant(player, advancement.name, criterion)
		}
		return true
	}
	started := false
	for _, criterion := range advancement.criteria {
		if obtained[criterion] {
			started = true
			break
		}
	}
	if !started {
		return false
	}
	for _, criterion := range advancement.criteria {
		if obtained[criterion] {
			store.revoke(player, advancement.name, criterion)
		}
	}
	return true
}

// changeCriterion changes a single criterion and reports whether it
// changed.
func changeCriterion(store *advancementStore, player, action, advancement, criterion string) bool {
	if action == "grant" {
		return store.grant(player, advancement, criterion)
	}
	return store.revoke(player, advancement, criterion)
}

// isAdvancementFilter reports whether the text names one of the filters
// 1.12.2 knows.
func isAdvancementFilter(filter string) bool {
	switch filter {
	case "only", "until", "from", "through", "everything":
		return true
	}
	return false
}

// resolveAdvancementPlayer resolves the player an advancement command
// targets.
//
// 1.12.2 looks the name up in the player list and reports "Player
// '<name>' cannot be found" when nobody has that name. A target selector
// is resolved as well, and the first player it matches is used.
func resolveAdvancementPlayer(ctx *exec.Context, arg ast.Arg) (*world.Entity, error) {
	switch t := arg.(type) {
	case *ast.StringArg:
		p := ctx.World.FindPlayer(t.Value)
		if p == nil {
			return nil, exec.Errorf("Player '%s' cannot be found", t.Value)
		}
		return p, nil
	case *ast.Selector:
		resolver := &exec.SelectorResolver{World: ctx.World}
		entities, err := resolver.Resolve(ctx.Source, t)
		selectorWarnings(ctx, resolver)
		if err != nil {
			return nil, exec.Errorf("%v", err)
		}
		if len(entities) == 0 {
			return nil, exec.Errorf("Player '%s' cannot be found", selectorText(t))
		}
		return entities[0], nil
	}
	return nil, exec.Errorf("Player '%s' cannot be found", "")
}

// advancementPlayerName returns the name 1.12.2 reports for a player.
//
// A player is known by its player name.
func advancementPlayerName(player *world.Entity) string {
	if player.PlayerName != "" {
		return player.PlayerName
	}
	return player.DisplayName()
}

// advancementPlainUsageError builds the error 1.12.2 answers a syntax
// mistake with when it names a subcommand.
//
// Those are thrown as a plain `CommandException`, so the server prints the
// usage line as it is, without the "Usage: " prefix.
func advancementPlainUsageError(usage string) error {
	return exec.Errorf("%s", usage)
}

// advancementEverythingSucceeded renders the message of a successful
// "everything".
func advancementEverythingSucceeded(action string, count int, player string) string {
	if action == "grant" {
		return fmt.Sprintf("Granted every advancement (%s total granted) to %s",
			strconv.Itoa(count), player)
	}
	return fmt.Sprintf("Revoked every advancement (%s total revoked) from %s",
		strconv.Itoa(count), player)
}

// advancementEverythingFailed renders the message of an "everything" that
// changed nothing.
func advancementEverythingFailed(action, player string) string {
	if action == "grant" {
		return fmt.Sprintf("Couldn't grant any advancements to %s because they already have them all", player)
	}
	return fmt.Sprintf("Couldn't revoke any advancements to %s because they haven't started any", player)
}

// advancementCriterionSucceeded renders the message of a criterion that
// changed.
func advancementCriterionSucceeded(action, advancement, player, criterion string) string {
	if action == "grant" {
		return fmt.Sprintf("Granted the criterion '%s' of advancement '%s' to %s",
			criterion, advancement, player)
	}
	return fmt.Sprintf("Revoked the criterion '%s' of advancement '%s' from %s",
		criterion, advancement, player)
}

// advancementCriterionFailed renders the message of a criterion that did
// not change.
func advancementCriterionFailed(action, advancement, player, criterion string) string {
	if action == "grant" {
		return fmt.Sprintf("Couldn't grant the criterion '%s' of advancement '%s' to %s because they already have it",
			criterion, advancement, player)
	}
	return fmt.Sprintf("Couldn't revoke the criterion '%s' of advancement '%s' from %s because they haven't started it",
		criterion, advancement, player)
}

// advancementFilterSucceeded renders the message of a filter that changed
// something.
func advancementFilterSucceeded(action, filter, advancement string, count int, player string) string {
	switch filter {
	case "only":
		if action == "grant" {
			return fmt.Sprintf("Granted the entire advancement '%s' to %s", advancement, player)
		}
		return fmt.Sprintf("Revoked the entire advancement '%s' from %s", advancement, player)
	case "until":
		if action == "grant" {
			return fmt.Sprintf("Granted '%s' and all ancestors (%s total granted) to %s",
				advancement, strconv.Itoa(count), player)
		}
		return fmt.Sprintf("Revoked '%s' and all ancestors (%s total revoked) from %s",
			advancement, strconv.Itoa(count), player)
	case "from":
		if action == "grant" {
			return fmt.Sprintf("Granted '%s' and all descendants (%s total granted) to %s",
				advancement, strconv.Itoa(count), player)
		}
		return fmt.Sprintf("Revoked '%s' and all descendants (%s total revoked) from %s",
			advancement, strconv.Itoa(count), player)
	default:
		if action == "grant" {
			return fmt.Sprintf("Granted '%s', all ancestors and all descendants (%s total granted) to %s",
				advancement, strconv.Itoa(count), player)
		}
		return fmt.Sprintf("Revoked '%s', all ancestors and all descendants (%s total revoked) from %s",
			advancement, strconv.Itoa(count), player)
	}
}

// advancementFilterFailed renders the message of a filter that changed
// nothing.
func advancementFilterFailed(action, filter, advancement, player string) string {
	switch filter {
	case "only":
		if action == "grant" {
			return fmt.Sprintf("Couldn't grant the advancement '%s' to %s because they already have it",
				advancement, player)
		}
		return fmt.Sprintf("Couldn't revoke the advancement '%s' from %s because they haven't started it",
			advancement, player)
	case "until":
		if action == "grant" {
			return fmt.Sprintf("Couldn't grant the advancement '%s' or its ancestors to %s because they already have them all",
				advancement, player)
		}
		return fmt.Sprintf("Couldn't revoke the advancement '%s' or its ancestors from %s because they haven't started any",
			advancement, player)
	case "from":
		if action == "grant" {
			return fmt.Sprintf("Couldn't grant the advancement '%s' or its descendants to %s because they already have them all",
				advancement, player)
		}
		return fmt.Sprintf("Couldn't revoke the advancement '%s' or its descendants from %s because they haven't started any",
			advancement, player)
	default:
		if action == "grant" {
			return fmt.Sprintf("Couldn't grant the advancement '%s', its ancestors or its descendants to %s because they already have them all",
				advancement, player)
		}
		return fmt.Sprintf("Couldn't revoke the advancement '%s', its ancestors or its descendants from %s because they haven't started any",
			advancement, player)
	}
}

// registerAdvancement adds the advancement command.
func registerAdvancement(reg *exec.Registry) {
	reg.Register(&advancementCmd{})
}
