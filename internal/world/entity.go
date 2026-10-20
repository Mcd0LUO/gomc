package world

import (
	"sort"
	"strconv"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/nbt"
)

// Entity is a world entity.
//
// Entities are represented generically: a type name plus an NBT payload.
// Minecraft 1.12.2 stores entity state in NBT anyway, so mirroring that
// keeps selectors, `/entitydata` and `/testfor` consistent with the real
// server without one Go type per mob.
type Entity struct {
	// ID is the entity's unique identifier.
	ID string
	// Type is the entity type, e.g. "zombie".
	Type string
	// Pos is the entity's position.
	Pos Vec
	// Rot is the entity's rotation in degrees.
	Rot Rotation
	// World is the dimension the entity belongs to.
	World string
	// NBT holds the entity's data tags.
	NBT *nbt.Compound
	// IsPlayer marks player entities.
	IsPlayer bool
	// PlayerName is set for player entities.
	PlayerName string
	// Width is the entity's box width, in blocks.
	Width float64
	// Height is the entity's box height, in blocks.
	Height float64
}

// BoundingBox returns the entity's collision box.
//
// The box matches 1.12.2's `Entity.getBoundingBox`: it is centred on the
// entity's X and Z, and extends from the entity's feet upwards.
func (e *Entity) BoundingBox() AABB {
	w := e.Width
	h := e.Height
	if w <= 0 {
		w = 0.6
	}
	if h <= 0 {
		h = 1.95
	}
	half := w / 2
	return AABB{
		MinX: e.Pos.X - half, MinY: e.Pos.Y, MinZ: e.Pos.Z - half,
		MaxX: e.Pos.X + half, MaxY: e.Pos.Y + h, MaxZ: e.Pos.Z + half,
	}
}

// AABB is an axis-aligned bounding box.
type AABB struct {
	// MinX, MinY, MinZ are the lower corner.
	MinX, MinY, MinZ float64
	// MaxX, MaxY, MaxZ are the upper corner.
	MaxX, MaxY, MaxZ float64
}

// Intersects reports whether two boxes overlap.
//
// This mirrors 1.12.2's `AxisAlignedBB.c`, which the volume filter uses to
// decide whether an entity is inside a `dx`/`dy`/`dz` volume.
func (a AABB) Intersects(b AABB) bool {
	return a.MinX < b.MaxX && a.MaxX > b.MinX &&
		a.MinY < b.MaxY && a.MaxY > b.MinY &&
		a.MinZ < b.MaxZ && a.MaxZ > b.MinZ
}

// DisplayName returns the name 1.12.2 reports for the entity.
//
// `Entity.getName` prefers the custom name, then the type's translated
// name. `EntityHuman` overrides it and returns the name of the game
// profile instead, so a player is always known by its user name and never
// by its custom name.
//
// The custom name is returned as it is stored: 1.12.2 keeps it in the data
// watcher as a plain string and hands it out unchanged, so the quotes of a
// name like `"Bob"` are part of the name.
func (e *Entity) DisplayName() string {
	if e.IsPlayer && e.PlayerName != "" {
		return e.PlayerName
	}
	if v, ok := e.NBTValue("CustomName").(nbt.String); ok && string(v) != "" {
		return string(v)
	}
	// An item entity is named after the item it carries. 1.12.2 builds
	// the name from the stack, and the server's language file has no key
	// for it, so the name comes out as the raw key.
	if e.Type == "item" {
		if name, ok := ItemDisplayName(e.itemID()); ok {
			return name
		}
	}
	return EntityTypeMeta(e.Type).DisplayName
}

// itemID returns the registry name of the item the entity carries.
func (e *Entity) itemID() string {
	item, ok := e.NBTValue("Item").(*nbt.Compound)
	if !ok {
		return ""
	}
	id, _ := item.Get("id").(nbt.String)
	return string(id)
}

// Rotation is an entity's facing, in degrees.
type Rotation struct {
	// Pitch is the vertical angle.
	Pitch float64
	// Yaw is the horizontal angle.
	Yaw float64
	// Roll is unused in 1.12.2 but kept for completeness.
	Roll float64
}

// Clone returns a deep copy of the entity.
func (e *Entity) Clone() *Entity {
	cp := *e
	cp.NBT = e.NBT.Clone()
	return &cp
}

// HasTag reports whether the entity carries tag.
func (e *Entity) HasTag(tag string) bool {
	for _, t := range e.TagList() {
		if t == tag {
			return true
		}
	}
	return false
}

// TagList returns the entity's scoreboard tags, sorted.
//
// 1.12.2 keeps the tags in the entity's NBT under the `Tags` key
// (`Entity.getScoreboardTags` reads that list), so they are not a separate
// piece of state: writing `Tags` into the NBT and using `/scoreboard
// players tag` are the same thing.
func (e *Entity) TagList() []string {
	list, ok := e.NBTValue("Tags").(*nbt.List)
	if !ok {
		return nil
	}
	var out []string
	for _, v := range list.Items() {
		if s, ok := v.(nbt.String); ok {
			out = append(out, string(s))
		}
	}
	sort.Strings(out)
	return out
}

// NBTValue returns an NBT value from the entity's data.
//
// It tolerates an entity without any NBT, which the selector filters have
// to cope with.
func (e *Entity) NBTValue(key string) nbt.Value {
	if e.NBT == nil {
		return nil
	}
	return e.NBT.Get(key)
}

// SetTag adds or removes a tag.
//
// The tags live in the entity's NBT under the `Tags` key, which is where
// 1.12.2 keeps them, so writing them through the NBT and through the
// scoreboard command are the same operation.
func (e *Entity) SetTag(tag string, present bool) {
	if e.NBT == nil {
		return
	}
	list, _ := e.NBT.Get("Tags").(*nbt.List)
	if list == nil {
		if !present {
			return
		}
		list = nbt.NewList(nbt.TypeString)
	}
	items := list.Items()
	if present {
		for _, v := range items {
			if s, ok := v.(nbt.String); ok && string(s) == tag {
				return
			}
		}
		list.Append(nbt.String(tag))
		// list 可能是刚建的，必须写回 NBT。
		e.NBT.Set("Tags", list)
	} else {
		kept := nbt.NewList(nbt.TypeString)
		for _, v := range items {
			if s, ok := v.(nbt.String); ok && string(s) == tag {
				continue
			}
			kept.Append(v)
		}
		e.NBT.Set("Tags", kept)
	}
}

// EntityType describes one entity type.
type EntityType struct {
	// Name is the type identifier.
	Name string
	// DefaultNBT holds the NBT a freshly spawned entity of this type has.
	DefaultNBT *nbt.Compound
	// IsPlayer marks player-only types.
	IsPlayer bool
	// NeedsItemStack marks a type the 1.12.2 entity factory only accepts
	// with a non-empty item stack.
	//
	// An item entity without a stack is empty and dies right after it is
	// created, so the server refuses to bring it into the world: `/summon
	// item` answers "Unable to summon object".
	NeedsItemStack bool
	// DisplayName is the name 1.12.2 reports for entities of this type.
	DisplayName string
	// Width is the entity's box width, in blocks.
	Width float64
	// Height is the entity's box height, in blocks.
	Height float64
}

// EntityRegistry holds the known entity types.
//
// The table is intentionally small and data-driven: only the types that
// functions realistically summon or filter on need an entry.
type EntityRegistry struct {
	types map[string]*EntityType
	order []string
}

// NewEntityRegistry returns a registry seeded with the common 1.12.2 types.
func NewEntityRegistry() *EntityRegistry {
	r := &EntityRegistry{types: map[string]*EntityType{}}
	for _, t := range defaultEntityTypes {
		r.Add(t)
	}
	return r
}

// Add registers an entity type.
//
// The 1.12.2 display name and box size are filled in from the built-in
// metadata table, so a type only needs to declare its NBT defaults.
func (r *EntityRegistry) Add(t *EntityType) {
	meta := EntityTypeMeta(t.Name)
	if t.DisplayName == "" {
		t.DisplayName = meta.DisplayName
	}
	if t.Width <= 0 {
		t.Width = meta.Width
	}
	if t.Height <= 0 {
		t.Height = meta.Height
	}
	if _, ok := r.types[t.Name]; !ok {
		r.order = append(r.order, t.Name)
	}
	r.types[t.Name] = t
}

// Get returns the type definition, or nil when unknown.
func (r *EntityRegistry) Get(name string) *EntityType { return r.types[name] }

// Has reports whether the type is known.
func (r *EntityRegistry) Has(name string) bool { _, ok := r.types[name]; return ok }

// Names returns the registered type names in sorted order.
func (r *EntityRegistry) Names() []string {
	out := append([]string(nil), r.order...)
	sort.Strings(out)
	return out
}

// NewEntity creates an entity of the given type with its default NBT.
func (r *EntityRegistry) NewEntity(name string, id string) (*Entity, bool) {
	t, ok := r.types[name]
	if !ok {
		return nil, false
	}
	e := &Entity{
		ID:     id,
		Type:   name,
		World:  "world",
		NBT:    t.DefaultNBT.Clone(),
		Rot:    Rotation{},
		Width:  t.Width,
		Height: t.Height,
	}
	if t.IsPlayer {
		e.IsPlayer = true
		e.PlayerName = name
	}
	return e, true
}

// defaultEntityTypes lists the entity types this runtime understands.
//
// The list is the entity registry of 1.12.2 (`EntityTypes.c()`) without
// the types the server cannot bring into the world:
//
//   - the four fireballs (`fireball`, `small_fireball`, `wither_skull` and
//     `dragon_fireball`) and the thrown potion: 1.12.2 only builds them
//     as the shot of something else, and `/summon` fails for them;
//   - the player: 1.12.2 keeps the players out of the entity registry and
//     handles them separately.
//
// The player is listed here anyway, because the runtime has to be able to
// create player entities; `/summon` refuses them.
//
// The NBT defaults mirror the values 1.12.2 gives a freshly spawned
// entity, because selectors and `/testfor` compare against them. The
// health values are the ones the real server reports for a freshly
// summoned entity.
//
// The slimes and the horse family have no health here: 1.12.2 rolls their
// health per entity (a slime comes with 1 or 16, a horse with 16 to 29),
// so there is no single value to write down.
var defaultEntityTypes = []*EntityType{
	{Name: "item", DefaultNBT: compoundWith("Item", nbt.NewCompound()), NeedsItemStack: true},
	{Name: "xp_orb", DefaultNBT: nbt.NewCompound()},
	{Name: "area_effect_cloud", DefaultNBT: nbt.NewCompound()},
	{Name: "elder_guardian", DefaultNBT: compoundWith("Health", nbt.Float(80))},
	{Name: "wither_skeleton", DefaultNBT: compoundWith("Health", nbt.Float(20))},
	{Name: "stray", DefaultNBT: compoundWith("Health", nbt.Float(20))},
	{Name: "egg", DefaultNBT: nbt.NewCompound()},
	{Name: "leash_knot", DefaultNBT: nbt.NewCompound()},
	{Name: "painting", DefaultNBT: nbt.NewCompound()},
	{Name: "arrow", DefaultNBT: nbt.NewCompound()},
	{Name: "snowball", DefaultNBT: nbt.NewCompound()},
	{Name: "ender_pearl", DefaultNBT: nbt.NewCompound()},
	{Name: "eye_of_ender_signal", DefaultNBT: nbt.NewCompound()},
	{Name: "xp_bottle", DefaultNBT: nbt.NewCompound()},
	{Name: "item_frame", DefaultNBT: nbt.NewCompound()},
	{Name: "tnt", DefaultNBT: compoundWith("Fuse", nbt.Int(80))},
	{Name: "falling_block", DefaultNBT: nbt.NewCompound()},
	{Name: "fireworks_rocket", DefaultNBT: nbt.NewCompound()},
	{Name: "husk", DefaultNBT: compoundWith("Health", nbt.Float(20))},
	{Name: "spectral_arrow", DefaultNBT: nbt.NewCompound()},
	{Name: "shulker_bullet", DefaultNBT: nbt.NewCompound()},
	{Name: "zombie_villager", DefaultNBT: compoundWith("Health", nbt.Float(20))},
	{Name: "skeleton_horse", DefaultNBT: compoundWith("Health", nbt.Float(15))},
	{Name: "zombie_horse", DefaultNBT: compoundWith("Health", nbt.Float(15))},
	{Name: "armor_stand", DefaultNBT: compoundWith("Health", nbt.Float(20))},
	{Name: "donkey", DefaultNBT: nbt.NewCompound()},
	{Name: "mule", DefaultNBT: nbt.NewCompound()},
	{Name: "evocation_fangs", DefaultNBT: nbt.NewCompound()},
	{Name: "evocation_illager", DefaultNBT: compoundWith("Health", nbt.Float(24))},
	{Name: "vex", DefaultNBT: compoundWith("Health", nbt.Float(14))},
	{Name: "vindication_illager", DefaultNBT: compoundWith("Health", nbt.Float(24))},
	{Name: "illusion_illager", DefaultNBT: compoundWith("Health", nbt.Float(32))},
	{Name: "commandblock_minecart", DefaultNBT: nbt.NewCompound()},
	{Name: "boat", DefaultNBT: nbt.NewCompound()},
	{Name: "minecart", DefaultNBT: nbt.NewCompound()},
	{Name: "chest_minecart", DefaultNBT: nbt.NewCompound()},
	{Name: "furnace_minecart", DefaultNBT: nbt.NewCompound()},
	{Name: "tnt_minecart", DefaultNBT: nbt.NewCompound()},
	{Name: "hopper_minecart", DefaultNBT: nbt.NewCompound()},
	{Name: "spawner_minecart", DefaultNBT: nbt.NewCompound()},
	{Name: "creeper", DefaultNBT: compoundWith("Health", nbt.Float(20))},
	{Name: "skeleton", DefaultNBT: compoundWith("Health", nbt.Float(20))},
	{Name: "spider", DefaultNBT: compoundWith("Health", nbt.Float(16))},
	{Name: "giant", DefaultNBT: compoundWith("Health", nbt.Float(100))},
	{Name: "zombie", DefaultNBT: compoundWith("Health", nbt.Float(20))},
	{Name: "slime", DefaultNBT: nbt.NewCompound()},
	{Name: "ghast", DefaultNBT: compoundWith("Health", nbt.Float(10))},
	{Name: "zombie_pigman", DefaultNBT: compoundWith("Health", nbt.Float(20))},
	{Name: "enderman", DefaultNBT: compoundWith("Health", nbt.Float(40))},
	{Name: "cave_spider", DefaultNBT: compoundWith("Health", nbt.Float(12))},
	{Name: "silverfish", DefaultNBT: compoundWith("Health", nbt.Float(8))},
	{Name: "blaze", DefaultNBT: compoundWith("Health", nbt.Float(20))},
	{Name: "magma_cube", DefaultNBT: nbt.NewCompound()},
	{Name: "ender_dragon", DefaultNBT: compoundWith("Health", nbt.Float(200))},
	{Name: "wither", DefaultNBT: compoundWith("Health", nbt.Float(300))},
	{Name: "bat", DefaultNBT: compoundWith("Health", nbt.Float(6))},
	{Name: "witch", DefaultNBT: compoundWith("Health", nbt.Float(26))},
	{Name: "endermite", DefaultNBT: compoundWith("Health", nbt.Float(8))},
	{Name: "guardian", DefaultNBT: compoundWith("Health", nbt.Float(30))},
	{Name: "shulker", DefaultNBT: compoundWith("Health", nbt.Float(30))},
	{Name: "pig", DefaultNBT: compoundWith("Health", nbt.Float(10))},
	{Name: "sheep", DefaultNBT: compoundWith("Health", nbt.Float(8))},
	{Name: "cow", DefaultNBT: compoundWith("Health", nbt.Float(10))},
	{Name: "chicken", DefaultNBT: compoundWith("Health", nbt.Float(4))},
	{Name: "squid", DefaultNBT: compoundWith("Health", nbt.Float(10))},
	{Name: "wolf", DefaultNBT: compoundWith("Health", nbt.Float(8))},
	{Name: "mooshroom", DefaultNBT: compoundWith("Health", nbt.Float(10))},
	{Name: "snowman", DefaultNBT: compoundWith("Health", nbt.Float(4))},
	{Name: "ocelot", DefaultNBT: compoundWith("Health", nbt.Float(10))},
	{Name: "villager_golem", DefaultNBT: compoundWith("Health", nbt.Float(100))},
	{Name: "horse", DefaultNBT: nbt.NewCompound()},
	{Name: "rabbit", DefaultNBT: compoundWith("Health", nbt.Float(3))},
	{Name: "polar_bear", DefaultNBT: compoundWith("Health", nbt.Float(30))},
	{Name: "llama", DefaultNBT: nbt.NewCompound()},
	{Name: "llama_spit", DefaultNBT: nbt.NewCompound()},
	{Name: "parrot", DefaultNBT: compoundWith("Health", nbt.Float(6))},
	{Name: "villager", DefaultNBT: compoundWith("Health", nbt.Float(20))},
	{Name: "ender_crystal", DefaultNBT: nbt.NewCompound()},
	{Name: "lightning_bolt", DefaultNBT: nbt.NewCompound()},
	{Name: "player", DefaultNBT: compoundWith("Health", nbt.Float(20)), IsPlayer: true},
}

// compoundWith builds a compound holding a single entry.
func compoundWith(key string, v nbt.Value) *nbt.Compound {
	c := nbt.NewCompound()
	c.Set(key, v)
	return c
}

// NormalizeEntityType maps a 1.12.2 entity type name to its canonical form.
//
// 1.12.2 registers the entity types under a lower-case key, so the name is
// lower-cased and the namespace is dropped.
//
// The spellings of the older versions are deliberately *not* mapped: 1.12.2
// renamed them — a "pig zombie" is a `zombie_pigman` there, and "pigzombie"
// is not a type the server knows at all. Mapping the old spellings would
// make a type the server rejects look like a valid one.
func NormalizeEntityType(name string) string {
	name = strings.ToLower(name)
	if i := strings.IndexByte(name, ':'); i >= 0 {
		name = name[i+1:]
	}
	return name
}

// ParseEntityTypeName resolves a possibly namespaced entity type.
//
// 1.12.2 builds a `MinecraftKey` from the text: the namespace is only
// read from the name when the colon is neither the first nor the second
// character, and the registry holds the "minecraft" namespace only. So
// "zombie", "minecraft:zombie", ":zombie" and "a:zombie" are the same
// key, while "foo:zombie" is a key of the "foo" namespace and names no
// type the server knows.
//
// The rule is the one of `IsRegisteredEntityType`, which the `type=`
// argument of a selector uses.
func (r *EntityRegistry) ParseEntityTypeName(name string) (string, bool) {
	namespace, key := EntityTypeKey(name)
	if namespace != "minecraft" {
		return "", false
	}
	if r.Has(key) {
		return key, true
	}
	return "", false
}

// FormatEntityID renders an entity id for messages.
func FormatEntityID(id string) string {
	if id == "" {
		return "entity"
	}
	return id
}

// ParseIntDefault parses s as an int, returning def when it fails.
func ParseIntDefault(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
