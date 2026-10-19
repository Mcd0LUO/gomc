package exec

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/mcsim/gomc/internal/ast"
	"github.com/mcsim/gomc/internal/nbt"
	"github.com/mcsim/gomc/internal/parser"
	"github.com/mcsim/gomc/internal/world"
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
//     `@s` and to no limit for `@a` and `@e`. A negative value reverses the
//     order before truncating.
//   - `@p` and `@a` are ordered by distance to the reference position, and
//     when exactly one entity is asked for the executing entity wins.
type SelectorResolver struct {
	// World is the state to resolve against.
	World *world.World
}

// Resolve returns the entities matched by the selector.
func (r *SelectorResolver) Resolve(src *Source, sel *ast.Selector) ([]*world.Entity, error) {
	if sel == nil {
		return nil, nil
	}
	if sel.Base == "s" {
		// @s is the executing entity. A console, or a function that is
		// not run through /execute, has no executor, and saying so is
		// more helpful than a bare "no matching entities".
		if src == nil || src.Entity == nil {
			return nil, fmt.Errorf("@s has no executor in this context: the command was not run through /execute")
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

	limit, hasLimit, err := parser.SelectorInt(sel, "c")
	if err != nil {
		return nil, err
	}
	if !hasLimit {
		// @a and @e default to no limit, the others to a single entity.
		if sel.Base == "a" || sel.Base == "e" {
			limit = 0
		} else {
			limit = 1
		}
	}

	switch sel.Base {
	case "p", "a", "e":
		ref := r.referenceVec(src, sel)
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
	out := make([]*world.Entity, 0, len(entities))
	for _, e := range entities {
		if e.IsPlayer {
			out = append(out, e)
		}
	}
	return out
}

// filter applies the selector's filters to the world's entities.
func (r *SelectorResolver) filter(src *Source, sel *ast.Selector) []*world.Entity {
	candidates := r.World.EntityList()

	// A `type` filter restricts to one entity type. The negation form
	// `type=!zombie` excludes that type.
	typeFilter := ""
	typeNegated := false
	hasType := false
	if v, ok := parser.SelectorString(sel, "type"); ok {
		hasType = true
		typeNegated = strings.HasPrefix(v, "!")
		typeFilter = world.NormalizeEntityType(strings.TrimPrefix(v, "!"))
	}

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
	radius, hasRadius := r.parseRadius(sel, ref)
	pitch, hasPitch := r.parseRotation(sel, "rx", "rxm")
	yaw, hasYaw := r.parseRotation(sel, "ry", "rym")

	out := make([]*world.Entity, 0, len(candidates))
	for _, e := range candidates {
		if hasType {
			isType := world.NormalizeEntityType(e.Type) == typeFilter
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
			name := r.World.Scoreboard.TeamNameOf(e.ID)
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
		out = append(out, e)
	}
	return out
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
// The volume is a block range: a negative size extends the volume backwards
// from the anchor, and the box is one block larger on the far side, which is
// what 1.12.2's `AxisAlignedBB` construction does.
func (r *SelectorResolver) parseVolume(sel *ast.Selector, src *Source) (world.AABB, bool) {
	dx, hasDX, _ := parser.SelectorInt(sel, "dx")
	dy, hasDY, _ := parser.SelectorInt(sel, "dy")
	dz, hasDZ, _ := parser.SelectorInt(sel, "dz")
	if !hasDX && !hasDY && !hasDZ {
		return world.AABB{}, false
	}
	anchor := r.referenceBlock(src, sel)
	minX, maxX := anchor.X, anchor.X
	minY, maxY := anchor.Y, anchor.Y
	minZ, maxZ := anchor.Z, anchor.Z
	if hasDX {
		if dx < 0 {
			minX = anchor.X + float64(dx)
		} else {
			maxX = anchor.X + float64(dx) + 1
		}
	}
	if hasDY {
		if dy < 0 {
			minY = anchor.Y + float64(dy)
		} else {
			maxY = anchor.Y + float64(dy) + 1
		}
	}
	if hasDZ {
		if dz < 0 {
			minZ = anchor.Z + float64(dz)
		} else {
			maxZ = anchor.Z + float64(dz) + 1
		}
	}
	return world.AABB{
		MinX: minX, MinY: minY, MinZ: minZ,
		MaxX: maxX, MaxY: maxY, MaxZ: maxZ,
	}, true
}

// radiusRange holds the `r`/`rm` distance bounds.
//
// In 1.12.2 `r` is the maximum distance and `rm` the minimum, so an entity
// must satisfy both when both are given.
type radiusRange struct {
	min    float64
	max    float64
	hasMin bool
	hasMax bool
}

// parseRadius computes the distance bounds from the `r`/`rm` arguments.
func (r *SelectorResolver) parseRadius(sel *ast.Selector, ref world.Vec) (radiusRange, bool) {
	maxR, hasMax, _ := parser.SelectorFloat(sel, "r")
	minR, hasMin, _ := parser.SelectorFloat(sel, "rm")
	if !hasMax && !hasMin {
		return radiusRange{}, false
	}
	return radiusRange{min: minR, max: maxR, hasMin: hasMin, hasMax: hasMax}, true
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

// matchScores applies the scoreboard conditions of a selector.
//
// 1.12.2 spells these as `score_<objective>` and `score_<objective>_min`.
// The objective has to exist and the entity has to have a score in it.
func matchScores(w *world.World, e *world.Entity, filters []parser.ScoreFilter) bool {
	for _, f := range filters {
		score, ok := w.Scoreboard.Score(e.ID, f.Objective)
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
func matchGamemode(e *world.Entity, mode string) bool {
	if !e.IsPlayer {
		return false
	}
	negated := strings.HasPrefix(mode, "!")
	value := strings.TrimPrefix(mode, "!")
	want, known := normalizeGameMode(value)
	if !known {
		// 1.12.2 resolves an unknown mode to "not set", which never
		// equals a player's mode.
		return false
	}
	current := int32(0)
	if v, ok := e.NBTValue("playerGameType").(nbt.Int); ok {
		current = int32(v)
	}
	name, known := gameModeName(int(current))
	if !known {
		return false
	}
	if negated {
		return name != want
	}
	return name == want
}

// normalizeGameMode resolves a game mode argument to its canonical name.
//
// 1.12.2 accepts the numeric id, the full name and the one-letter
// abbreviation, and the match is case sensitive.
func normalizeGameMode(v string) (string, bool) {
	switch v {
	case "survival", "s", "0":
		return "survival", true
	case "creative", "c", "1":
		return "creative", true
	case "adventure", "a", "2":
		return "adventure", true
	case "spectator", "sp", "3":
		return "spectator", true
	}
	return "", false
}

// gameModeName maps a 1.12.2 game mode id to its name.
//
// The second result reports whether the id is a known game mode.
func gameModeName(id int) (string, bool) {
	switch id {
	case 0:
		return "survival", true
	case 1:
		return "creative", true
	case 2:
		return "adventure", true
	case 3:
		return "spectator", true
	}
	return "", false
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
