package world

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// GameruleType classifies a gamerule's value type.
type GameruleType int

const (
	// RuleBoolean is a boolean gamerule.
	RuleBoolean GameruleType = iota
	// RuleInt is an integer gamerule.
	RuleInt
	// RuleString is a string gamerule.
	RuleString
	// RuleFunction is a gamerule holding a function reference.
	RuleFunction
	// RuleAny is a gamerule that holds any value as text.
	//
	// 1.12.2 creates one of these when a rule that does not exist yet
	// is set: `GameRules.set` falls back to `ANY_VALUE`.
	RuleAny
)

// Gamerule is one game rule.
type Gamerule struct {
	// Name is the rule's identifier.
	Name string
	// Type is the rule's value type.
	Type GameruleType
	// Bool holds the value for boolean rules.
	Bool bool
	// Int holds the value for integer rules.
	Int int
	// String holds the value for string and function rules.
	String string
}

// Display renders the rule's value as a string.
func (g *Gamerule) Display() string {
	switch g.Type {
	case RuleBoolean:
		if g.Bool {
			return "true"
		}
		return "false"
	case RuleInt:
		return strconv.Itoa(g.Int)
	case RuleString, RuleFunction, RuleAny:
		return g.String
	}
	return ""
}

// GameRules holds the world's game rules.
type GameRules struct {
	rules map[string]*Gamerule
	order []string
}

// NewGameRules returns the 1.12.2 default game rules.
func NewGameRules() *GameRules {
	g := &GameRules{rules: map[string]*Gamerule{}}
	for _, r := range defaultGamerules {
		cp := *r
		g.rules[r.Name] = &cp
		g.order = append(g.order, r.Name)
	}
	return g
}

// Get returns a gamerule by name.
func (g *GameRules) Get(name string) *Gamerule { return g.rules[name] }

// Has reports whether the rule exists.
func (g *GameRules) Has(name string) bool { _, ok := g.rules[name]; return ok }

// Set stores a gamerule value from its textual form.
//
// A rule that does not exist yet is created: 1.12.2's `GameRules.set` falls
// back to `ANY_VALUE` and adds the rule, so `gamerule <new> <value>`
// succeeds on a real server.
func (g *GameRules) Set(name, value string) error {
	r, ok := g.rules[name]
	if !ok {
		g.rules[name] = &Gamerule{Name: name, Type: RuleAny, String: value}
		g.order = append(g.order, name)
		return nil
	}
	switch r.Type {
	case RuleBoolean:
		switch strings.ToLower(value) {
		case "true":
			r.Bool = true
		case "false":
			r.Bool = false
		default:
			return fmt.Errorf("game rule %q expects true or false, got %q", name, value)
		}
	case RuleInt:
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("game rule %q expects a number, got %q", name, value)
		}
		r.Int = n
	case RuleString, RuleFunction, RuleAny:
		// A rule that was created on the fly (1.12.2's `ANY_VALUE`) keeps
		// its value as a string, and setting it again replaces the value.
		r.String = value
	}
	return nil
}

// Bool returns a boolean rule's value.
func (g *GameRules) Bool(name string) bool {
	if r, ok := g.rules[name]; ok && r.Type == RuleBoolean {
		return r.Bool
	}
	return false
}

// Int returns an integer rule's value.
func (g *GameRules) Int(name string) int {
	if r, ok := g.rules[name]; ok && r.Type == RuleInt {
		return r.Int
	}
	return 0
}

// String returns a string rule's value.
func (g *GameRules) String(name string) string {
	if r, ok := g.rules[name]; ok {
		return r.String
	}
	return ""
}

// Names returns the rule names in definition order.
func (g *GameRules) Names() []string { return append([]string(nil), g.order...) }

// defaultGamerules are the 1.12.2 game rules and their defaults.
//
// The list comes from `GameRules`' static initialiser in the 1.12.2 NMS.
// Note what is not here: `tntExplodes` is a 1.13 addition, and 1.12.2 has
// no way to turn the explosion off.
var defaultGamerules = []*Gamerule{
	{Name: "announceAdvancements", Type: RuleBoolean, Bool: true},
	{Name: "commandBlockOutput", Type: RuleBoolean, Bool: true},
	{Name: "disableElytraMovementCheck", Type: RuleBoolean, Bool: false},
	{Name: "doDaylightCycle", Type: RuleBoolean, Bool: true},
	{Name: "doEntityDrops", Type: RuleBoolean, Bool: true},
	{Name: "doFireTick", Type: RuleBoolean, Bool: true},
	{Name: "doLimitedCrafting", Type: RuleBoolean, Bool: false},
	{Name: "doMobLoot", Type: RuleBoolean, Bool: true},
	{Name: "doMobSpawning", Type: RuleBoolean, Bool: true},
	{Name: "doTileDrops", Type: RuleBoolean, Bool: true},
	{Name: "doWeatherCycle", Type: RuleBoolean, Bool: true},
	{Name: "gameLoopFunction", Type: RuleFunction, String: "-"},
	{Name: "keepInventory", Type: RuleBoolean, Bool: false},
	{Name: "logAdminCommands", Type: RuleBoolean, Bool: true},
	{Name: "maxCommandChainLength", Type: RuleInt, Int: 65536},
	{Name: "maxEntityCramming", Type: RuleInt, Int: 24},
	{Name: "mobGriefing", Type: RuleBoolean, Bool: true},
	{Name: "naturalRegeneration", Type: RuleBoolean, Bool: true},
	{Name: "randomTickSpeed", Type: RuleInt, Int: 3},
	{Name: "reducedDebugInfo", Type: RuleBoolean, Bool: false},
	{Name: "sendCommandFeedback", Type: RuleBoolean, Bool: true},
	{Name: "showDeathMessages", Type: RuleBoolean, Bool: true},
	{Name: "spectatorsGenerateChunks", Type: RuleBoolean, Bool: true},
	{Name: "spawnRadius", Type: RuleInt, Int: 10},
}

// World is the aggregate state commands operate on.
type World struct {
	// Blocks holds the block data.
	Blocks *BlockStore
	// Entities holds the world's entities, keyed by id.
	Entities map[string]*Entity
	// entityOrder preserves spawn order for deterministic iteration.
	entityOrder []string
	// entityCache holds the entities of entityOrder, in the same order.
	//
	// The cache exists so that the hot paths — a selector walking the
	// entities, above all — do not have to copy the whole table every
	// time. It is written by the same two calls that write entityOrder,
	// so the two cannot fall out of step.
	entityCache []*Entity
	// entityCacheOK tells whether the cache is known to hold exactly the
	// entities of the map, in spawn order.
	//
	// It is false only when a change could not be applied to the cache
	// exactly, in which case the cache is rebuilt on the next read.
	entityCacheOK bool
	// Scoreboard holds objectives, scores and teams.
	Scoreboard *Scoreboard
	// Rules holds the game rules.
	Rules *GameRules
	// EntityTypes is the entity type registry.
	EntityTypes *EntityRegistry
	// Time is the absolute tick count of the world.
	Time int
	// DayTime is the time of day in ticks, 0 to 23999.
	//
	// 1.12.2 keeps this separate from the tick counter: `/time set`
	// changes the time of day, while the server keeps counting ticks.
	DayTime int
	// Weather is the world's weather.
	Weather WeatherState
	// Difficulty is the world's difficulty.
	Difficulty Difficulty
	// Spawn is the world spawn point.
	Spawn Pos
	// PlayerSpawns holds the per-player spawn points.
	PlayerSpawns map[string]Pos
	// Border is the world border.
	Border WorldBorder
	// Seed is the world seed.
	Seed int64
	// IdleTimeout is the idle timeout in minutes, 0 to disable it.
	IdleTimeout int
	// nextEntityID seeds generated entity ids.
	nextEntityID int
}

// NewWorld returns a world with default state.
func NewWorld() *World {
	return &World{
		Blocks:        NewBlockStore(),
		Entities:      map[string]*Entity{},
		Scoreboard:    NewScoreboard(),
		Rules:         NewGameRules(),
		EntityTypes:   NewEntityRegistry(),
		Time:          0,
		DayTime:       0,
		Weather:       WeatherState{Kind: WeatherClear},
		Difficulty:    DifficultyNormal,
		PlayerSpawns:  map[string]Pos{},
		Border:        DefaultWorldBorder(),
		entityCacheOK: true,
	}
}

// AddEntity inserts an entity, assigning an id when it has none.
//
// An entity whose id is already taken replaces the old one and keeps the
// place the old one had in the spawn order.
func (w *World) AddEntity(e *Entity) *Entity {
	if e.ID == "" {
		w.nextEntityID++
		e.ID = fmt.Sprintf("entity-%d", w.nextEntityID)
	}
	if old, exists := w.Entities[e.ID]; exists {
		// The id is taken: the order stays, only the entity changes.
		if old != e {
			w.replaceEntity(e)
		}
	} else {
		w.entityOrder = append(w.entityOrder, e.ID)
		w.entityCache = append(w.entityCache, e)
	}
	w.Entities[e.ID] = e
	return e
}

// replaceEntity puts an entity in the place of the one it replaces.
//
// The entity keeps its id, so it keeps its place in the spawn order.
func (w *World) replaceEntity(e *Entity) {
	for i, cached := range w.entityCache {
		if cached.ID == e.ID {
			w.entityCache[i] = e
			return
		}
	}
	// The old entity is not in the cache, so the cache cannot be patched
	// and has to be rebuilt on the next read.
	w.entityCacheOK = false
}

// RemoveEntity deletes an entity by id.
func (w *World) RemoveEntity(id string) bool {
	if _, ok := w.Entities[id]; !ok {
		return false
	}
	delete(w.Entities, id)
	for i, eid := range w.entityOrder {
		if eid == id {
			w.entityOrder = append(w.entityOrder[:i], w.entityOrder[i+1:]...)
			// The cache runs parallel to the order, so the entity to
			// drop sits at the same index.
			if i < len(w.entityCache) {
				n := len(w.entityCache)
				w.entityCache = append(w.entityCache[:i], w.entityCache[i+1:]...)
				// The slot the last entity moved out of still holds
				// it, so it is cleared when there is room for it.
				if n <= cap(w.entityCache) {
					spare := w.entityCache[:n]
					spare[n-1] = nil
				}
			} else {
				w.entityCacheOK = false
			}
			break
		}
	}
	return true
}

// EntityList returns the entities in spawn order.
//
// The caller gets a copy of its own and may keep and modify it.
//
// Deterministic ordering matters: selectors iterate entities in a stable
// sequence, and a test harness must produce the same result every run.
func (w *World) EntityList() []*Entity {
	cache := w.EntitySlice()
	out := make([]*Entity, len(cache))
	copy(out, cache)
	return out
}

// EntitySlice returns the entities in spawn order, without copying them.
//
// The slice is the world's own cache: the caller must not modify it, not
// even the order of its elements. It is meant for the read-only hot
// paths, and it stays valid until the next AddEntity or RemoveEntity.
//
// The world is the only writer of the entity table, so the slice always
// agrees with the Entities map. Should the map be changed behind the
// world's back, the slice notices the change of the entity count and
// rebuilds itself.
func (w *World) EntitySlice() []*Entity {
	if !w.entityCacheOK || len(w.Entities) != len(w.entityCache) {
		w.rebuildEntityCache()
	}
	return w.entityCache
}

// rebuildEntityCache fills the cache from the spawn order and the map.
//
// It is the safety net for a table that was changed behind the world's
// back; the world's own calls keep the cache up to date instead.
func (w *World) rebuildEntityCache() {
	cache := make([]*Entity, 0, len(w.entityOrder))
	for _, id := range w.entityOrder {
		if e, ok := w.Entities[id]; ok {
			cache = append(cache, e)
		}
	}
	w.entityCache = cache
	w.entityCacheOK = true
}

// EntityCount returns the number of entities.
func (w *World) EntityCount() int { return len(w.Entities) }

// Players returns the player entities in spawn order.
func (w *World) Players() []*Entity {
	var out []*Entity
	for _, e := range w.EntitySlice() {
		if e.IsPlayer {
			out = append(out, e)
		}
	}
	return out
}

// FindPlayer returns the player entity with the given name.
func (w *World) FindPlayer(name string) *Entity {
	for _, e := range w.Players() {
		if e.PlayerName == name {
			return e
		}
	}
	return nil
}

// FindEntity returns the entity with the given id.
func (w *World) FindEntity(id string) *Entity {
	return w.Entities[id]
}

// ObjectiveNames returns the scoreboard objective names, sorted.
func (w *World) ObjectiveNames() []string {
	objs := w.Scoreboard.Objectives()
	out := make([]string, len(objs))
	for i, o := range objs {
		out[i] = o.Name
	}
	sort.Strings(out)
	return out
}
