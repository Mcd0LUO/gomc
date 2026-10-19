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
	case RuleString, RuleFunction:
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
func (g *GameRules) Set(name, value string) error {
	r, ok := g.rules[name]
	if !ok {
		return fmt.Errorf("unknown game rule %q", name)
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
	case RuleString, RuleFunction:
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
		Blocks:       NewBlockStore(),
		Entities:     map[string]*Entity{},
		Scoreboard:   NewScoreboard(),
		Rules:        NewGameRules(),
		EntityTypes:  NewEntityRegistry(),
		Time:         0,
		DayTime:      0,
		Weather:      WeatherState{Kind: WeatherClear},
		Difficulty:   DifficultyNormal,
		PlayerSpawns: map[string]Pos{},
		Border:       DefaultWorldBorder(),
	}
}

// AddEntity inserts an entity, assigning an id when it has none.
func (w *World) AddEntity(e *Entity) *Entity {
	if e.ID == "" {
		w.nextEntityID++
		e.ID = fmt.Sprintf("entity-%d", w.nextEntityID)
	}
	if _, exists := w.Entities[e.ID]; !exists {
		w.entityOrder = append(w.entityOrder, e.ID)
	}
	w.Entities[e.ID] = e
	return e
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
			break
		}
	}
	return true
}

// EntityList returns the entities in spawn order.
//
// Deterministic ordering matters: selectors iterate entities in a stable
// sequence, and a test harness must produce the same result every run.
func (w *World) EntityList() []*Entity {
	out := make([]*Entity, 0, len(w.entityOrder))
	for _, id := range w.entityOrder {
		if e, ok := w.Entities[id]; ok {
			out = append(out, e)
		}
	}
	return out
}

// EntityCount returns the number of entities.
func (w *World) EntityCount() int { return len(w.Entities) }

// Players returns the player entities in spawn order.
func (w *World) Players() []*Entity {
	var out []*Entity
	for _, e := range w.EntityList() {
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
