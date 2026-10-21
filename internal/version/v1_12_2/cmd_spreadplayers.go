package v1_12_2

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/lexer"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// spreadPlayersCmd implements `/spreadplayers`.
//
// The 1.12.2 syntax is `/spreadplayers <x> <z> <spreadDistance> <maxRange>
// <respectTeams true|false> <player ...>`. There is no y: the players are
// spread over the x/z plane around the given point.
//
// 1.12.2's algorithm (`CommandSpreadPlayers`) drops every player on a random
// spot of the box around the centre and then pushes the spots apart until
// they are far enough from each other. The random part is why gomc cannot
// copy the algorithm as it is: gomc has to be reproducible, so the random
// source is seeded from the arguments of the command instead of the clock.
// Everything else -- the order of the messages, the wording of the errors
// and the way the spots are relaxed -- is the server's.
type spreadPlayersCmd struct{}

// Name implements exec.Handler.
func (c *spreadPlayersCmd) Name() string { return "spreadplayers" }

// Usage implements exec.Handler.
func (c *spreadPlayersCmd) Usage() string { return usageLine("spreadplayers") }

// The limits 1.12.2 puts on the centre of the spread.
//
// `CommandAbstract.b(base, text, -30000000, 30000000, true)` reads the
// centre of the command with those bounds.
const (
	spreadCentreMin = -30000000.0
	spreadCentreMax = 30000000.0
)

// spreadMaxIterations is how often 1.12.2 relaxes the spots before it
// gives up.
//
// `CommandSpreadPlayers.a(Location2D[], double, World, Random, ...)` runs
// its loop at most 10000 times; when the loop runs out the command fails
// with "Could not spread ...".
const spreadMaxIterations = 10000

// Parse implements exec.Handler.
//
// 1.12.2 counts the arguments before it looks at any of them:
// `CommandSpreadPlayers.execute` starts with `if (astring.length < 6)` and
// throws the usage of the command. The arguments are therefore collected
// first, and the numbers are read afterwards, in the order the server reads
// them: the centre, the distance between the players, the range around the
// centre and the team flag.
func (c *spreadPlayersCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	toks := make([]lexer.Token, 0, 8)
	for p.More() {
		t, ok := p.Next()
		if !ok {
			break
		}
		toks = append(toks, t)
	}
	if len(toks) < 6 {
		if len(toks) == 0 {
			return nil, p.Errorf("Usage: %s", c.Usage())
		}
		return nil, p.ErrorAt(toks[0], "Usage: %s", c.Usage())
	}
	// The centre is read with `CommandAbstract.b(base, text, true)`: a `~`
	// is relative to the position of the command, a whole number names the
	// middle of a block and a number with a decimal point is taken as it is.
	for _, t := range toks[:2] {
		if err := checkSpreadCentre(t.Unquote()); err != nil {
			return nil, p.ErrorAt(t, "%v", err)
		}
	}
	// The two distances are read with `CommandAbstract.a(text, min)`: the
	// spread distance is at least 0 and the range is at least one block
	// more than the spread distance.
	spread, err := parseSpreadDistance(toks[2], 0)
	if err != nil {
		return nil, p.ErrorAt(toks[2], "%v", err)
	}
	maxRange, err := parseSpreadDistance(toks[3], spread+1)
	if err != nil {
		return nil, p.ErrorAt(toks[3], "%v", err)
	}
	// The team flag is read with `CommandAbstract.d(text)`.
	teams, err := parseSpreadTeams(toks[4])
	if err != nil {
		return nil, p.ErrorAt(toks[4], "%v", err)
	}
	// The rest names the players. 1.12.2 tells a selector from a name with
	// `PlayerSelector.isPattern`, which looks at the first character.
	args := []ast.Arg{
		&ast.StringArg{Value: toks[0].Unquote()},
		&ast.StringArg{Value: toks[1].Unquote()},
		&ast.FloatArg{Value: spread},
		&ast.FloatArg{Value: maxRange},
		&ast.StringArg{Value: strconv.FormatBool(teams)},
	}
	for _, t := range toks[5:] {
		if IsSelector(t.Text) {
			sel, err := ParseSelector(t.Text)
			if err != nil {
				return nil, p.ErrorAt(t, "%v", err)
			}
			args = append(args, sel)
			continue
		}
		args = append(args, &ast.StringArg{Value: t.Unquote()})
	}
	return args, nil
}

// Run implements exec.Handler.
func (c *spreadPlayersCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	// The centre is relative to the position of the command, which is only
	// known when the command runs.
	base := ctx.Source.RelativeBase().BlockPos()
	centreX, err := resolveSpreadCentre(args[0].(*ast.StringArg).Value, float64(base.X))
	if err != nil {
		return 0, err
	}
	centreZ, err := resolveSpreadCentre(args[1].(*ast.StringArg).Value, float64(base.Z))
	if err != nil {
		return 0, err
	}
	spread := args[2].(*ast.FloatArg).Value
	maxRange := args[3].(*ast.FloatArg).Value
	teams := args[4].(*ast.StringArg).Value == "true"

	var entities []*world.Entity
	for _, arg := range args[5:] {
		found, err := resolveSpreadTarget(ctx, arg)
		if err != nil {
			return 0, err
		}
		entities = append(entities, found...)
	}
	if len(entities) == 0 {
		// 1.12.2 cannot reach this: the argument count guarantees a
		// player and a player that is not there is an error. The server
		// keeps the branch, so gomc keeps it too.
		return 0, exec.Errorf("No players found to spread")
	}

	// The word of the messages depends on the team flag: 1.12.2 says
	// "team(s)" when the teams are respected and "player(s)" otherwise.
	word := "player(s)"
	if teams {
		word = "team(s)"
	}

	// The players of one team are spread together, so they share a spot.
	// The number of spots is the number of teams, or the number of
	// players when the teams are not respected.
	groupOf, groupCount := spreadGroups(ctx, entities, teams)
	spots := make([]spreadSpot, groupCount)
	rnd := newSpreadRand(centreX, centreZ, spread, maxRange, groupCount)
	minX, maxX := centreX-maxRange, centreX+maxRange
	minZ, maxZ := centreZ-maxRange, centreZ+maxRange
	for i := range spots {
		spots[i] = spreadSpot{
			X: rnd.between(minX, maxX),
			Z: rnd.between(minZ, maxZ),
		}
	}

	// 1.12.2 says what it is about to do before it does it.
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Spreading %d %s %s blocks around %s, %s (min %s blocks apart)",
			len(entities), word, javaDouble(maxRange),
			javaDouble(centreX), javaDouble(centreZ), javaDouble(spread)))

	iterations, minDistance, ok := spreadRelax(&rnd, spots, spread, ctx, minX, minZ, maxX, maxZ)
	if !ok {
		return 0, exec.Errorf(
			"Could not spread %d %s around %s, %s (too many players for space - try using spread of at most %s)",
			len(spots), word, javaDouble(centreX), javaDouble(centreZ),
			fmt.Sprintf("%.2f", minDistance))
	}

	average := spreadPlace(ctx, entities, groupOf, spots)

	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Successfully spread %d %s around %s, %s",
			len(spots), word, javaDouble(centreX), javaDouble(centreZ)))
	// 1.12.2 only reports the average when there is more than one spot:
	// with a single spot there is no distance to average.
	if len(spots) > 1 {
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("(Average distance between %s is %s blocks apart after %d iterations)",
				spreadWord(word), fmt.Sprintf("%.2f", average), iterations))
	}
	return len(entities), nil
}

// spreadWord returns the word the info message uses.
//
// 1.12.2 has one message for the teams and one for the players
// ("commands.spreadplayers.info.teams" and "commands.spreadplayers.info.players"),
// and the difference is the word in the middle of the sentence.
func spreadWord(word string) string {
	if word == "team(s)" {
		return "teams"
	}
	return "players"
}

// spreadGroups returns the group of every entity and the number of groups.
//
// With respectTeams the players of one team are spread together: they all
// get the same spot. 1.12.2 keys the groups by the team object, so a player
// without a team and every entity that is not a player share one group.
//
// The groups are numbered in the order the entities are named in, which is
// the order 1.12.2 hands the spots out in.
func spreadGroups(ctx *exec.Context, entities []*world.Entity, teams bool) ([]int, int) {
	if !teams {
		// Every entity is on its own.
		out := make([]int, len(entities))
		for i := range out {
			out[i] = i
		}
		return out, len(entities)
	}
	index := map[string]int{}
	out := make([]int, 0, len(entities))
	for _, e := range entities {
		key := ""
		if e.IsPlayer {
			key = ctx.World.Scoreboard.TeamNameOf(e.PlayerName)
		}
		g, ok := index[key]
		if !ok {
			g = len(index)
			index[key] = g
		}
		out = append(out, g)
	}
	return out, len(index)
}

// spreadPlace moves the entities onto their spots.
//
// It returns the average distance between a player and the nearest other
// spot, which is what 1.12.2 reports.
func spreadPlace(ctx *exec.Context, entities []*world.Entity, groupOf []int, spots []spreadSpot) float64 {
	total := 0.0
	for i, e := range entities {
		own := groupOf[i]
		spot := &spots[own]
		e.Pos = world.Vec{
			X: math.Floor(spot.X) + 0.5,
			Y: float64(spreadStandY(ctx, spot)),
			Z: math.Floor(spot.Z) + 0.5,
		}
		// The distance to the nearest other spot.
		nearest := math.MaxFloat64
		for j := range spots {
			if j == own {
				continue
			}
			if d := spreadDistance(spot, &spots[j]); d < nearest {
				nearest = d
			}
		}
		if nearest == math.MaxFloat64 {
			// There is only one spot, so there is no distance. 1.12.2
			// would add the largest double there is, but it never shows
			// the number.
			continue
		}
		total += nearest
	}
	if len(spots) < 2 {
		return 0
	}
	return total / float64(len(entities))
}

// spreadStandY returns the height a player stands at.
//
// 1.12.2 looks for the first block under the spot that is not air and puts
// the player on top of it (`CommandSpreadPlayers.Location2D.a(World)`).
func spreadStandY(ctx *exec.Context, spot *spreadSpot) int {
	pos := world.Vec{X: spot.X, Z: spot.Z}.BlockPos()
	for y := 255; y >= 0; y-- {
		st := ctx.World.Blocks.Get(world.Pos{X: pos.X, Y: y, Z: pos.Z})
		if st.IsAir() {
			continue
		}
		return y + 1
	}
	// gomc's world has no terrain of its own, so a column without blocks
	// has nothing to stand on and the player stays at the bottom of the
	// world.
	return 0
}

// spreadStandable reports whether a player can stand on a block.
//
// 1.12.2 looks at the material of the block:
// `CommandSpreadPlayers.Location2D.b(World)` refuses a liquid and fire,
// because a player would drown or burn there.
//
// The materials of 1.12.2 that are liquid are the ones of the water and
// the lava blocks; `MaterialLiquid` is the only subclass of `Material` that
// answers `true` to `isLiquid`.
func spreadStandable(name string) bool {
	switch name {
	case "water", "flowing_water", "lava", "flowing_lava":
		return false
	case "fire":
		return false
	}
	return true
}

// spreadRelax pushes the spots apart until they are far enough from each
// other.
//
// It is the loop of `CommandSpreadPlayers.a(Location2D[], double, World,
// Random, ...)`: every round every spot that is too close to another one
// moves one block away from the average of the spots that are too close,
// and the spots that left the box are put back into it. When a round
// changes nothing, the spots that have no ground under them are moved
// somewhere else.
//
// It reports the number of rounds it took, the smallest distance between
// two spots it saw, and whether the spots could be spread at all.
func spreadRelax(rnd *spreadRand, spots []spreadSpot, spread float64, ctx *exec.Context,
	minX, minZ, maxX, maxZ float64) (int, float64, bool) {
	moved := true
	minDistance := 0.0
	rounds := 0
	for rounds = 0; rounds < spreadMaxIterations && moved; rounds++ {
		moved = false
		minDistance = math.MaxFloat64
		for i := range spots {
			spot := &spots[i]
			tooClose := 0
			var awayX, awayZ float64
			for j := range spots {
				if i == j {
					continue
				}
				d := spreadDistance(spot, &spots[j])
				if d < minDistance {
					minDistance = d
				}
				if d >= spread {
					continue
				}
				tooClose++
				awayX += spots[j].X - spot.X
				awayZ += spots[j].Z - spot.Z
			}
			if tooClose > 0 {
				awayX /= float64(tooClose)
				awayZ /= float64(tooClose)
				if length := math.Sqrt(awayX*awayX + awayZ*awayZ); length > 0 {
					// Move one block away from the average.
					spot.X -= awayX / length
					spot.Z -= awayZ / length
				} else {
					// The spot is exactly in the middle of the ones
					// that are too close, so there is no direction to
					// move to.
					spot.X = rnd.between(minX, maxX)
					spot.Z = rnd.between(minZ, maxZ)
				}
				moved = true
			}
			if spreadClamp(spot, minX, minZ, maxX, maxZ) {
				moved = true
			}
		}
		if moved {
			continue
		}
		// Nothing moved, so the spots are far enough apart. What is left
		// is to check that there is ground under them.
		for i := range spots {
			if spreadHasGround(ctx, &spots[i]) {
				continue
			}
			spots[i].X = rnd.between(minX, maxX)
			spots[i].Z = rnd.between(minZ, maxZ)
			moved = true
		}
	}
	if rounds >= spreadMaxIterations {
		return rounds, minDistance, false
	}
	return rounds, minDistance, true
}

// spreadHasGround reports whether there is ground under a spot.
//
// 1.12.2 looks for the first block under the spot that is not air and
// checks that it is neither a liquid nor fire.
func spreadHasGround(ctx *exec.Context, spot *spreadSpot) bool {
	pos := world.Vec{X: spot.X, Z: spot.Z}.BlockPos()
	for y := 255; y >= 0; y-- {
		st := ctx.World.Blocks.Get(world.Pos{X: pos.X, Y: y, Z: pos.Z})
		if st.IsAir() {
			continue
		}
		return spreadStandable(st.Name)
	}
	// An empty column has no ground in 1.12.2 either, but gomc's world
	// has no terrain of its own, so an empty column is the floor.
	return true
}

// spreadClamp puts a spot back into the box when it left it.
//
// It reports whether the spot had to be moved.
func spreadClamp(spot *spreadSpot, minX, minZ, maxX, maxZ float64) bool {
	moved := false
	if spot.X < minX {
		spot.X = minX
		moved = true
	} else if spot.X > maxX {
		spot.X = maxX
		moved = true
	}
	if spot.Z < minZ {
		spot.Z = minZ
		moved = true
	} else if spot.Z > maxZ {
		spot.Z = maxZ
		moved = true
	}
	return moved
}

// spreadDistance returns the distance between two spots.
func spreadDistance(a, b *spreadSpot) float64 {
	dx := a.X - b.X
	dz := a.Z - b.Z
	return math.Sqrt(dx*dx + dz*dz)
}

// spreadSpot is one place a player is spread to.
//
// 1.12.2 calls it `Location2D`: it has an x and a z, and the y is looked
// up in the world when the player is put there.
type spreadSpot struct {
	X float64
	Z float64
}

// resolveSpreadTarget resolves one of the `<player ...>` arguments.
//
// 1.12.2 reads them with `PlayerSelector.isPattern`: an argument that
// starts with `@` is a selector and everything else is the name of a
// player. A selector that matches nothing is an error, and so is the name
// of a player who is not online.
func resolveSpreadTarget(ctx *exec.Context, arg ast.Arg) ([]*world.Entity, error) {
	sel, ok := arg.(*ast.Selector)
	if !ok {
		name := arg.(*ast.StringArg).Value
		p := ctx.World.FindPlayer(name)
		if p == nil {
			return nil, exec.Errorf("Player '%s' cannot be found", name)
		}
		return []*world.Entity{p}, nil
	}
	resolver := &SelectorResolver{World: ctx.World}
	entities, err := resolver.Resolve(ctx.Source, sel)
	selectorWarnings(ctx, resolver)
	if err != nil {
		return nil, exec.Errorf("%v", err)
	}
	if len(entities) == 0 {
		return nil, exec.Errorf("Selector '%s' found nothing", selectorText(sel))
	}
	return entities, nil
}

// checkSpreadCentre checks that the centre of the spread is a number.
//
// 1.12.2 reads the centre with `CommandAbstract.b`, which reports a number
// it cannot read as "'<text>' is not a valid number". The text of the
// error is the number without the `~`, because the `~` is taken off
// before the number is read.
func checkSpreadCentre(text string) error {
	body := text
	if strings.HasPrefix(body, "~") {
		body = body[1:]
		if body == "" {
			// A bare `~` is the position of the command itself.
			return nil
		}
	}
	if _, ok := parseJavaDouble(body); !ok {
		return exec.Errorf("'%s' is not a valid number", body)
	}
	return nil
}

// resolveSpreadCentre returns the centre of the spread.
//
// It is `CommandAbstract.b(base, text, true)`: a `~` is relative to the
// position of the command, a whole number names the middle of a block and
// a number with a decimal point is taken as it is.
func resolveSpreadCentre(text string, base float64) (float64, error) {
	relative := strings.HasPrefix(text, "~")
	body := text
	if relative {
		body = text[1:]
	}
	value := 0.0
	if !relative || body != "" {
		v, ok := parseJavaDouble(body)
		if !ok {
			return 0, exec.Errorf("'%s' is not a valid number", body)
		}
		value = v
		if !relative && !strings.Contains(body, ".") {
			// A whole number names a block, and the middle of a
			// block is where the player is put.
			value += 0.5
		}
	}
	if relative {
		value += base
	}
	if value < spreadCentreMin {
		return 0, exec.Errorf("The number you have entered (%s) is too small, it must be at least %s",
			vanillaDouble2(value), vanillaDouble2(spreadCentreMin))
	}
	if value > spreadCentreMax {
		return 0, exec.Errorf("The number you have entered (%s) is too big, it must be at most %s",
			vanillaDouble2(value), vanillaDouble2(spreadCentreMax))
	}
	return value, nil
}

// parseSpreadDistance reads one of the two distances of the command.
//
// 1.12.2 reads them with `CommandAbstract.a(text, min)`: the number has to
// be at least `min`, and the server says how much too small it is.
func parseSpreadDistance(t lexer.Token, min float64) (float64, error) {
	v, ok := parseJavaDouble(t.Text)
	if !ok {
		return 0, exec.Errorf("'%s' is not a valid number", t.Text)
	}
	if v < min {
		return 0, exec.Errorf("The number you have entered (%s) is too small, it must be at least %s",
			vanillaDouble2(v), vanillaDouble2(min))
	}
	return v, nil
}

// parseSpreadTeams reads the `<respectTeams>` argument.
//
// 1.12.2 reads it with `CommandAbstract.d(text)`, which takes "true" and
// "1" as true and "false" and "0" as false.
func parseSpreadTeams(t lexer.Token) (bool, error) {
	switch t.Text {
	case "true", "1":
		return true, nil
	case "false", "0":
		return false, nil
	}
	return false, exec.Errorf("'%s' is not true or false", t.Text)
}

// parseJavaDouble reads a number the way Java's `Double.parseDouble` does.
//
// The parser package holds the one implementation, so that a coordinate
// and a distance are read the same way.
func parseJavaDouble(text string) (float64, bool) {
	return parser.JavaDouble(text)
}

// spreadRand is the random source the spreading uses.
//
// 1.12.2 uses `new Random()`, which is seeded from the clock, so the same
// command puts the players somewhere different every time. gomc has to be
// reproducible, so the seed is derived from the arguments of the command
// instead: the same command always spreads the same way, and two different
// commands do not all land on the same pattern.
//
// The generator is a splitmix64, so the sequence does not depend on the
// version of Go.
type spreadRand struct {
	state uint64
}

// newSpreadRand returns the random source of a spread.
func newSpreadRand(centreX, centreZ, spread, maxRange float64, count int) spreadRand {
	// The seed is a hash of the arguments, so that the same command always
	// gives the same spots and two different commands do not.
	h := uint64(14695981039346656037)
	mix := func(v uint64) {
		for i := 0; i < 8; i++ {
			h ^= (v >> (8 * i)) & 0xff
			h *= 1099511628211
		}
	}
	mix(math.Float64bits(centreX))
	mix(math.Float64bits(centreZ))
	mix(math.Float64bits(spread))
	mix(math.Float64bits(maxRange))
	mix(uint64(count))
	return spreadRand{state: h}
}

// next returns the next value of the generator.
func (r *spreadRand) next() uint64 {
	r.state += 0x9E3779B97F4A7C15
	z := r.state
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}

// between returns a value in [lo,hi).
//
// It is `MathHelper.a(random, lo, hi)`, the way 1.12.2 picks a spot in the
// box around the centre.
func (r *spreadRand) between(lo, hi float64) float64 {
	v := float64(r.next()>>11) / (1 << 53)
	return lo + (hi-lo)*v
}

// registerSpreadPlayers adds the spread players command.
func registerSpreadPlayers(reg *exec.Registry) {
	reg.Register(&spreadPlayersCmd{})
}
