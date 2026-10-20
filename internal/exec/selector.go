package exec

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// SelectorResolver resolves target selectors against the world.
//
// The semantics follow 1.12.2's `PlayerSelector`, which was recovered from
// the server jar rather than guessed:
//
//   - `@a`, `@p` and `@r` select players. For `@r` a `type` argument replaces
//     the player filter, so `@r[type=zombie]` can pick a mob.
//   - `@e` selects every entity, `@s` only the executing entity.
//   - The arguments form a flat map, so a repeated name keeps the last value.
//   - `r`/`rm` bound the distance, `rx`/`rxm` the pitch and `ry`/`rym` the
//     yaw. For the rotation bounds the unsuffixed name is the *upper* bound
//     and the `_m` suffixed one the lower bound.
//   - `dx`/`dy`/`dz` select entities whose bounding box intersects the
//     volume, not entities whose position falls inside it.
//   - `c` limits the result. It defaults to one entity for `@p`, `@r` and
//     `@s` and to no limit for `@a` and `@e`. A count that is not an
//     integer falls back to that default instead of being an error, and a
//     negative value reverses the order before truncating.
//   - A `type` argument naming a type the server does not know makes the
//     whole selector match nothing.
//   - `name=` compares the name 1.12.2 reports for the entity, which is
//     the user name for a player and the custom name for anything else.
//   - `@p` and `@a` are ordered by distance to the reference position, and
//     when exactly one entity is asked for the executing entity wins.
type SelectorResolver struct {
	// World is the state to resolve against.
	World *world.World
	// Warnings collects the notes the server prints while it resolves a
	// selector.
	//
	// 1.12.2 sends them next to the command's own output — a red
	// message about an entity type it does not know, for example — so
	// the command layer has to pass them on.
	Warnings []string
}

// Resolve returns the entities matched by the selector.
func (r *SelectorResolver) Resolve(src *Source, sel *ast.Selector) ([]*world.Entity, error) {
	if sel == nil {
		return nil, nil
	}
	if sel.Base == "s" {
		// @s is the executing entity. A console, or a function that is
		// not run through /execute, has no executor, and then @s
		// matches nothing.
		//
		// 1.12.2 says so in the wording of the command that used it:
		// `kill @s` from the console answers "Entity '@s' cannot be
		// found", and a message that expands @s answers "Selector
		// '@s' found nothing". The command layer owns that wording, so
		// the resolver just reports an empty result.
		if src == nil || src.Entity == nil {
			return nil, nil
		}
		matched := r.filter(src, sel)
		for _, e := range matched {
			if e.ID == src.Entity.ID {
				return []*world.Entity{e}, nil
			}
		}
		return nil, nil
	}
	matched := r.filter(src, sel)

	switch sel.Base {
	case "a", "p":
		matched = onlyPlayers(matched)
	case "r":
		// A type argument replaces the player filter for @r.
		if _, hasType := parser.SelectorString(sel, "type"); !hasType {
			matched = onlyPlayers(matched)
		}
	}

	// 1.12.2 reads the count with `MathHelper.a(String, int)`, which is
	// `Integer.parseInt` with the default as the fallback. A count that is
	// not an integer is therefore not an error: it falls back to the
	// default of the base.
	limit := parser.SelectorIntDefault(sel, "c", defaultCount(sel.Base))

	switch sel.Base {
	case "p", "a", "e":
		ref := r.referenceVec(src, sel)
		if limit == 1 {
			// Only one entity is asked for, so 1.12.2's sort by
			// distance only has to find the closest one. The sort is
			// stable, so entities at the same distance keep the order
			// they were matched in, and the closest one is the first
			// of the list that is not farther away than any of the
			// others. Finding that one does not need the whole list
			// sorted, and the executor wins over it either way.
			if src.Entity != nil {
				for _, e := range matched {
					if e.ID == src.Entity.ID {
						return []*world.Entity{e}, nil
					}
				}
			}
			closest := r.closestEntity(matched, ref)
			if closest == nil {
				return nil, nil
			}
			return []*world.Entity{closest}, nil
		}
		sort.SliceStable(matched, func(i, j int) bool {
			return r.squaredDistance(ref, matched[i]) < r.squaredDistance(ref, matched[j])
		})
	case "r":
		// 1.12.2 shuffles for @r so the pick is genuinely random.
		shuffle(matched)
	}

	if sel.Base == "s" {
		// @s can only ever be the executing entity.
		if src.Entity == nil {
			return nil, nil
		}
		for _, e := range matched {
			if e.ID == src.Entity.ID {
				matched = []*world.Entity{e}
				break
			}
		}
		return matched, nil
	}

	// When exactly one entity is asked for, 1.12.2 prefers the executor.
	if limit == 1 && sel.Base != "r" && src.Entity != nil {
		for _, e := range matched {
			if e.ID == src.Entity.ID {
				return []*world.Entity{e}, nil
			}
		}
	}

	if limit != 0 {
		if limit < 0 {
			// A negative count reverses the order before truncating.
			for i, j := 0, len(matched)-1; i < j; i, j = i+1, j-1 {
				matched[i], matched[j] = matched[j], matched[i]
			}
			limit = -limit
		}
		if limit < len(matched) {
			matched = matched[:limit]
		}
	}
	return matched, nil
}

// onlyPlayers keeps the player entities of a list.
func onlyPlayers(entities []*world.Entity) []*world.Entity {
	var out []*world.Entity
	for _, e := range entities {
		if e.IsPlayer {
			out = appendEntity(out, e)
		}
	}
	return out
}

// appendEntity adds an entity to the result of a selector.
//
// A selector matches a handful of entities at most, so the result starts
// out small instead of being sized after the number of candidates.
//
// The growth is by doubling, which is what `append` does too, but the
// first step matters: `append` on an empty slice gives a capacity of one,
// so matching two or three entities would allocate two or three times.
// Starting at four covers the usual case in one allocation.
func appendEntity(out []*world.Entity, e *world.Entity) []*world.Entity {
	if len(out) == cap(out) {
		grown := make([]*world.Entity, len(out), max(2*cap(out), 4))
		copy(grown, out)
		out = grown
	}
	return append(out, e)
}

// defaultCount returns the count a selector falls back to.
//
// 1.12.2 counts no limit for @a and @e and a single entity for @p, @r and
// @s. The count is also the fallback for a `c` argument that is not an
// integer.
func defaultCount(base string) int {
	if base == "a" || base == "e" {
		return 0
	}
	return 1
}

// typeFilterApplies reports whether the `type` argument filters the
// entities of a selector base.
//
// 1.12.2 adds the type predicate for @e, @r and @s only. For @a and @p
// it adds the "must be a player" predicate instead, so the type
// argument is silently ignored.
func typeFilterApplies(base string) bool {
	return base == "e" || base == "r" || base == "s"
}

// typeNorms remembers the canonical form of the entity types it has seen.
//
// A world holds few entity types, so the canonical form of a type is worth
// remembering: computing it again for every entity of the world is the
// price of a type filter.
type typeNorms struct {
	raws  [4]string
	norms [4]string
	count int
}

// of returns the canonical form of an entity type.
//
// It is the very same value `world.NormalizeEntityType` returns; the
// remembered ones are only reused, never computed differently.
func (t *typeNorms) of(raw string) string {
	for i := 0; i < t.count; i++ {
		if t.raws[i] == raw {
			return t.norms[i]
		}
	}
	norm := world.NormalizeEntityType(raw)
	if t.count < len(t.raws) {
		t.raws[t.count] = raw
		t.norms[t.count] = norm
		t.count++
	}
	return norm
}

// filter applies the selector's filters to the world's entities.
func (r *SelectorResolver) filter(src *Source, sel *ast.Selector) []*world.Entity {
	// 1.12.2 checks the type argument before anything else: a type the
	// entity registry does not know makes the whole selector match
	// nothing, whatever the base is. The server also says the type is
	// invalid, and then matches nothing without failing.
	//
	// (Measured: `kill @e[type=pigzombie]` answers "Entity type
	// 'minecraft:pigzombie' is invalid" and succeeds. The negated
	// spelling `type=!pigzombie` gets the same note, without the `!`.)
	if v, ok := parser.SelectorString(sel, "type"); ok {
		name := strings.TrimPrefix(v, "!")
		if !world.IsRegisteredEntityType(name) {
			namespace, key := world.EntityTypeKey(name)
			r.Warnings = append(r.Warnings,
				fmt.Sprintf("Entity type '%s:%s' is invalid", namespace, key))
			return nil
		}
	}

	// The candidates are the entities of the world in spawn order. The
	// world hands out its own slice for them, which is only read here:
	// the result of the selector is a slice of its own, so the world
	// keeps the order of its entities whatever the selector does.
	candidates := r.World.EntitySlice()

	// A `type` filter restricts to one entity type. The negation form
	// `type=!zombie` excludes that type.
	//
	// 1.12.2 only applies the argument to @e, @r and @s: for @a and @p
	// the type is ignored and the "must be a player" filter is used
	// instead, so `@a[type=zombie]` still selects every player.
	typeFilter := ""
	typeNegated := false
	hasType := false
	if v, ok := parser.SelectorString(sel, "type"); ok && typeFilterApplies(sel.Base) {
		hasType = true
		typeNegated = strings.HasPrefix(v, "!")
		typeFilter = world.NormalizeEntityType(strings.TrimPrefix(v, "!"))
	}

	// The `name` filter compares the name 1.12.2 reports for the entity,
	// which is `Entity.getName`: the user name for a player — a player
	// with a custom name is still known by its user name — and the custom
	// name for anything else.
	nameFilter, hasName := parser.SelectorString(sel, "name")
	nameNegated := strings.HasPrefix(nameFilter, "!")
	nameFilter = strings.TrimPrefix(nameFilter, "!")

	teamFilter, hasTeam := parser.SelectorString(sel, "team")
	teamNegated := strings.HasPrefix(teamFilter, "!")
	teamFilter = strings.TrimPrefix(teamFilter, "!")

	tagFilter, hasTag := parser.SelectorString(sel, "tag")
	tagNegated := strings.HasPrefix(tagFilter, "!")
	tagFilter = strings.TrimPrefix(tagFilter, "!")

	scoreFilters := parser.SelectorScoreFilters(sel)

	gamemode, hasGamemode := parseGamemode(sel)

	levelMax, hasLevelMax, _ := parser.SelectorInt(sel, "l")
	levelMin, hasLevelMin, _ := parser.SelectorInt(sel, "lm")

	ref := r.referenceVec(src, sel)
	volume, hasVolume := r.parseVolume(sel, src)
	radius, hasRadius := r.parseRadius(sel)
	pitch, hasPitch := r.parseRotation(sel, "rx", "rxm")
	yaw, hasYaw := r.parseRotation(sel, "ry", "rym")

	var out []*world.Entity
	var types typeNorms
	for _, e := range candidates {
		if hasType {
			// The type of an entity is compared in the canonical form
			// 1.12.2 uses, and the filter is canonical itself: the
			// check above only lets a type through that the server
			// knows, and such a type carries neither a namespace nor
			// an upper-case letter. An entity whose type is the filter
			// is therefore of that type, and comparing the two strings
			// is much cheaper than normalising the type of every
			// entity of the world. Every other entity still has to be
			// normalised, because it may spell the same type in
			// another way.
			isType := e.Type == typeFilter
			if !isType {
				isType = types.of(e.Type) == typeFilter
			}
			if isType == typeNegated {
				continue
			}
		}
		if hasName {
			if (e.DisplayName() == nameFilter) == nameNegated {
				continue
			}
		}
		if hasTeam {
			// An entity that is on no team reports an empty team name,
			// so `team=!red` still matches it.
			name := r.World.Scoreboard.TeamNameOf(ScoreboardHolder(e))
			if (name == teamFilter) == teamNegated {
				continue
			}
		}
		if hasTag {
			if !matchTag(e, tagFilter, tagNegated) {
				continue
			}
		}
		if len(scoreFilters) > 0 && !matchScores(r.World, e, scoreFilters) {
			continue
		}
		if hasGamemode && !matchGamemode(e, gamemode) {
			continue
		}
		if (hasLevelMax || hasLevelMin) && !matchLevel(e, levelMin, hasLevelMin, levelMax, hasLevelMax) {
			continue
		}
		if hasVolume && !e.BoundingBox().Intersects(volume) {
			continue
		}
		if hasRadius {
			d := r.squaredDistance(ref, e)
			if radius.hasMax && d > radius.max*radius.max {
				continue
			}
			if radius.hasMin && d < radius.min*radius.min {
				continue
			}
		}
		if hasPitch && !matchRotation(e.Rot.Pitch, pitch) {
			continue
		}
		if hasYaw && !matchRotation(e.Rot.Yaw, yaw) {
			continue
		}
		out = appendEntity(out, e)
	}
	return out
}

// closestEntity returns the entity of the list that is closest to the
// reference point, or nil when the list is empty.
//
// 1.12.2 sorts the matching entities by their distance and then takes the
// first one. The sort is stable, so entities at the same distance keep the
// order they were matched in, and the closest one is therefore the first
// of the list that is not farther away than any of the others. That is
// what this finds, without sorting the whole list.
//
// (Measured on a real 1.12.2 server and against the golden files: the
// entity a `@e[c=1]` picks is the same one the sorted list would give.)
func (r *SelectorResolver) closestEntity(entities []*world.Entity, ref world.Vec) *world.Entity {
	var best *world.Entity
	bestDist := 0.0
	for _, e := range entities {
		d := r.squaredDistance(ref, e)
		if best == nil || d < bestDist {
			best = e
			bestDist = d
		}
	}
	return best
}

// squaredDistance returns the squared distance between a point and an entity.
func (r *SelectorResolver) squaredDistance(ref world.Vec, e *world.Entity) float64 {
	dx := e.Pos.X - ref.X
	dy := e.Pos.Y - ref.Y
	dz := e.Pos.Z - ref.Z
	return dx*dx + dy*dy + dz*dz
}

// referenceVec computes the position selectors measure from.
//
// It honours the `x`/`y`/`z` arguments, which override the source position.
// 1.12.2 parses the override as an integer and then offsets X and Z by half
// a block, because the reference is a block position.
func (r *SelectorResolver) referenceVec(src *Source, sel *ast.Selector) world.Vec {
	pos := src.Pos
	if sel == nil {
		return pos
	}
	if v, ok, err := parser.SelectorInt(sel, "x"); err == nil && ok {
		pos.X = float64(v) + 0.5
	}
	if v, ok, err := parser.SelectorInt(sel, "y"); err == nil && ok {
		pos.Y = float64(v)
	}
	if v, ok, err := parser.SelectorInt(sel, "z"); err == nil && ok {
		pos.Z = float64(v) + 0.5
	}
	return pos
}

// referenceBlock computes the block position the volume is anchored at.
//
// 1.12.2 uses the source's block position, with the `x`/`y`/`z` arguments
// overriding it as plain integers.
func (r *SelectorResolver) referenceBlock(src *Source, sel *ast.Selector) world.Vec {
	pos := world.Vec{
		X: math.Floor(src.Pos.X),
		Y: math.Floor(src.Pos.Y),
		Z: math.Floor(src.Pos.Z),
	}
	if sel == nil {
		return pos
	}
	if v, ok, err := parser.SelectorInt(sel, "x"); err == nil && ok {
		pos.X = float64(v)
	}
	if v, ok, err := parser.SelectorInt(sel, "y"); err == nil && ok {
		pos.Y = float64(v)
	}
	if v, ok, err := parser.SelectorInt(sel, "z"); err == nil && ok {
		pos.Z = float64(v)
	}
	return pos
}

// volume is an axis-aligned selection volume.
type volume struct {
	min, max world.Vec
}

// parseVolume computes the volume from the `dx`/`dy`/`dz` arguments.
//
// The volume is a block range anchored at the reference block. A negative
// size reaches backwards from the anchor, a positive one forwards, and
// the far end is always one block past the last block of the range. An
// axis without a size argument counts as a size of zero, so the box
// still covers the anchor block.
func (r *SelectorResolver) parseVolume(sel *ast.Selector, src *Source) (world.AABB, bool) {
	// 1.12.2 checks whether the arguments are present, not whether
	// they parse, so an unparsable size still builds a volume.
	_, hasDX := parser.SelectorString(sel, "dx")
	_, hasDY := parser.SelectorString(sel, "dy")
	_, hasDZ := parser.SelectorString(sel, "dz")
	if !hasDX && !hasDY && !hasDZ {
		return world.AABB{}, false
	}
	dx := float64(parser.SelectorIntDefault(sel, "dx", 0))
	dy := float64(parser.SelectorIntDefault(sel, "dy", 0))
	dz := float64(parser.SelectorIntDefault(sel, "dz", 0))
	anchor := r.referenceBlock(src, sel)
	minX, maxX := volumeAxis(anchor.X, dx)
	minY, maxY := volumeAxis(anchor.Y, dy)
	minZ, maxZ := volumeAxis(anchor.Z, dz)
	return world.AABB{
		MinX: minX, MinY: minY, MinZ: minZ,
		MaxX: maxX, MaxY: maxY, MaxZ: maxZ,
	}, true
}

// volumeAxis computes the two ends of one axis of the selection volume.
//
// This mirrors `PlayerSelector.a(BlockPosition, int, int, int)`: a
// negative size moves the near end backwards and leaves the far end one
// block past the anchor, a positive size moves the far end out and
// leaves the near end at the anchor.
func volumeAxis(anchor, size float64) (float64, float64) {
	if size < 0 {
		return anchor + size, anchor + 1
	}
	return anchor, anchor + size + 1
}

// radiusRange holds the `r`/`rm` distance bounds.
//
// In 1.12.2 `r` is the maximum distance and `rm` the minimum, so an entity
// must satisfy both when both are given. A bound that does not apply is
// left out instead of being set to a meaningless value.
type radiusRange struct {
	min    float64
	max    float64
	hasMin bool
	hasMax bool
}

// parseRadius computes the distance bounds from the `r`/`rm` arguments.
//
// 1.12.2 parses both as integers, so a value such as "1.5" is not a
// number at all: the argument falls back to the default of -1. A bound
// below -0.5 is then dropped rather than applied, which means that
// `@e[rm=-1]` does not ask for "at least one block away" but for no
// lower bound whatsoever.
func (r *SelectorResolver) parseRadius(sel *ast.Selector) (radiusRange, bool) {
	maxR := parser.SelectorIntDefault(sel, "r", -1)
	minR := parser.SelectorIntDefault(sel, "rm", -1)
	out := radiusRange{}
	// The bounds below -0.5 do not take part in the test.
	if maxR >= 0 {
		out.max = math.Max(float64(maxR), 1e-4)
		out.hasMax = true
	}
	if minR >= 0 {
		out.min = math.Max(float64(minR), 1e-4)
		out.hasMin = true
	}
	if !out.hasMax && !out.hasMin {
		return radiusRange{}, false
	}
	return out, true
}

// angleRange is a rotation window.
//
// 1.12.2 stores the bounds as integers wrapped into [-180, 180). When the
// lower bound is above the upper bound the range wraps around north.
type angleRange struct {
	min int
	max int
}

// parseRotation reads a rotation window from the `rx`/`rxm` or `ry`/`rym`
// arguments.
//
// The unsuffixed argument is the upper bound and the `_m` suffixed one the
// lower bound, which is the opposite of the naming intuition but matches
// the 1.12.2 server.
func (r *SelectorResolver) parseRotation(sel *ast.Selector, upper, lower string) (angleRange, bool) {
	hi, hasHi, err := parser.SelectorInt(sel, upper)
	if err != nil {
		hasHi = false
	}
	lo, hasLo, err := parser.SelectorInt(sel, lower)
	if err != nil {
		hasLo = false
	}
	if !hasHi && !hasLo {
		return angleRange{}, false
	}
	if !hasHi {
		// The default upper bound is 359, which wraps to -1.
		hi = -1
	}
	if !hasLo {
		lo = 0
	}
	return angleRange{min: wrapAngle(lo), max: wrapAngle(hi)}, true
}

// wrapAngle folds an angle into the range [-180, 180).
//
// This mirrors `MathHelper.b`, which 1.12.2 uses for the rotation bounds.
func wrapAngle(a int) int {
	a %= 360
	if a >= 180 {
		a -= 360
	}
	if a < -180 {
		a += 360
	}
	return a
}

// matchRotation reports whether an angle falls within the window.
//
// The entity's angle is floored and wrapped, and a window whose lower bound
// is above its upper bound wraps around north.
func matchRotation(angle float64, r angleRange) bool {
	a := wrapAngle(int(math.Floor(angle)))
	if r.min <= r.max {
		return a >= r.min && a <= r.max
	}
	return a >= r.min || a <= r.max
}

// matchTag applies the tag filter.
//
// An empty tag name selects the entities that carry no tag at all.
func matchTag(e *world.Entity, tag string, negated bool) bool {
	if tag == "" {
		return (len(e.TagList()) == 0) != negated
	}
	return e.HasTag(tag) != negated
}

// ScoreboardHolder returns the name the scoreboard tracks an entity under.
//
// 1.12.2 keeps a player's scoreboard entry under the player's name and
// every other entity's under its UUID (`Entity.bn()`), which is what
// `scoreboard players set <name> ...` and the `team=` and `score_...`
// selector arguments refer to.
//
// The runtime does not model UUIDs, so a non-player entity is tracked
// under its internal id instead. That is a known gap: a score written
// for such an entity is only reachable through the same id, never
// through a UUID.
func ScoreboardHolder(e *world.Entity) string {
	if e == nil {
		return ""
	}
	if e.IsPlayer && e.PlayerName != "" {
		return e.PlayerName
	}
	return e.ID
}

// matchScores applies the scoreboard conditions of a selector.
//
// 1.12.2 spells these as `score_<objective>` and `score_<objective>_min`.
// The objective has to exist and the entity has to have a score in it.
func matchScores(w *world.World, e *world.Entity, filters []parser.ScoreFilter) bool {
	holder := ScoreboardHolder(e)
	for _, f := range filters {
		score, ok := w.Scoreboard.Score(holder, f.Objective)
		if !ok {
			return false
		}
		if f.IsMin {
			if score < f.Value {
				return false
			}
		} else if score > f.Value {
			return false
		}
	}
	return true
}

// gameMode is a 1.12.2 game mode.
//
// The "not set" mode is the one 1.12.2 falls back to when a game mode
// argument names no mode at all. No player is in it, so it is what makes
// an unknown mode match nobody — and everybody, when the filter is
// negated.
type gameMode int

const (
	gameModeNotSet    gameMode = -1
	gameModeSurvival  gameMode = 0
	gameModeCreative  gameMode = 1
	gameModeAdventure gameMode = 2
	gameModeSpectator gameMode = 3
)

// parseGamemode parses the `m=` game mode filter.
//
// 1.12.2 accepts the numeric id or the name, plus the one-letter
// abbreviations, and a leading `!` negates the filter.
func parseGamemode(sel *ast.Selector) (string, bool) {
	v, ok := parser.SelectorString(sel, "m")
	if !ok {
		return "", false
	}
	return v, true
}

// matchGamemode applies the game mode filter.
//
// The filter only ever matches players: 1.12.2 reads the game mode from
// the player's interact manager, and any other entity fails the test.
//
// A mode the server does not know resolves to "not set", which no player
// is in. The filter therefore matches nobody, and the negated filter
// matches everybody.
func matchGamemode(e *world.Entity, mode string) bool {
	if !e.IsPlayer {
		return false
	}
	negated := strings.HasPrefix(mode, "!")
	want := resolveGameMode(strings.TrimPrefix(mode, "!"))
	current := playerGameMode(e)
	if negated {
		return current != want
	}
	return current == want
}

// resolveGameMode resolves a game mode argument to the mode it selects.
//
// 1.12.2 parses the argument as a number first and falls back to the name,
// and the names are matched case sensitively. An argument that names no
// mode at all resolves to "not set".
func resolveGameMode(v string) gameMode {
	if n, err := strconv.ParseInt(v, 10, 32); err == nil {
		return gameModeByID(int(n))
	}
	switch v {
	case "survival", "s":
		return gameModeSurvival
	case "creative", "c":
		return gameModeCreative
	case "adventure", "a":
		return gameModeAdventure
	case "spectator", "sp":
		return gameModeSpectator
	}
	return gameModeNotSet
}

// playerGameMode returns the game mode a player is in.
//
// 1.12.2 reads the mode from the player's data, and a mode the server
// does not know falls back to survival.
func playerGameMode(e *world.Entity) gameMode {
	id := 0
	if v, ok := e.NBTValue("playerGameType").(nbt.Int); ok {
		id = int(v)
	}
	return gameModeByID(id)
}

// gameModeByID maps a 1.12.2 game mode id to the mode.
func gameModeByID(id int) gameMode {
	switch id {
	case -1:
		return gameModeNotSet
	case 0:
		return gameModeSurvival
	case 1:
		return gameModeCreative
	case 2:
		return gameModeAdventure
	case 3:
		return gameModeSpectator
	}
	return gameModeSurvival
}

// matchLevel applies the `l`/`lm` level filter.
//
// `l` is the maximum experience level and `lm` the minimum, and only
// players have one.
func matchLevel(e *world.Entity, min int, hasMin bool, max int, hasMax bool) bool {
	if !e.IsPlayer {
		return false
	}
	level := 0
	if v, ok := e.NBTValue("XpLevel").(nbt.Int); ok {
		level = int(v)
	}
	if hasMin && level < min {
		return false
	}
	if hasMax && level > max {
		return false
	}
	return true
}

// shuffle permutes the slice using the engine's random source.
//
// 1.12.2 uses `Collections.shuffle` for `@r`, so the pick is uniform rather
// than "the first match".
func shuffle(items []*world.Entity) {
	if len(items) < 2 {
		return
	}
	// A deterministic Fisher-Yates driven by the process random source
	// keeps `@r` unpredictable while remaining reproducible when a seed
	// is configured.
	for i := len(items) - 1; i > 0; i-- {
		j := randIntn(i + 1)
		items[i], items[j] = items[j], items[i]
	}
}
