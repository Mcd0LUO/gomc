package world

import "strings"

// Entity type metadata for Minecraft 1.12.2.
//
// Two properties are needed to reproduce the real server's selector
// behaviour:
//
//   - The display name, because `name=` matches the name 1.12.2 reports for
//     an entity. `Entity.getName` returns the custom name when the entity
//     has one, the player name for players, and otherwise the *translated*
//     type name: 1.12.2 looks up "entity.<class name without the Entity
//     prefix>.name" in the language file — "entity.ArmorStand.name" — so the
//     names have spaces ("Armor Stand"), not the registry spelling. A type
//     whose key is missing from the language file falls back to the key
//     itself, and that is reproduced here too.
//   - The bounding box size, because the `dx`/`dy`/`dz` volume filter tests
//     whether the entity's box *intersects* the volume, not whether its
//     position point falls inside it.
//
// The names come from `en_US.lang` in the 1.12.2 client assets; the sizes
// come from the 1.12 entity data.
//
// The keys are the names 1.12.2 registers the entity types under, which
// are the same names the `type=` argument of a selector takes. The table
// holds every one of them, so it doubles as the list of the types the
// server knows: see IsRegisteredEntityType.
var entityTypeMeta = map[string]entityMeta{
	"item":                  {"Item", 0.25, 0.25},
	"xp_orb":                {"Experience Orb", 0.5, 0.5},
	"area_effect_cloud":     {"Area Effect Cloud", 6, 0.5},
	"elder_guardian":        {"Elder Guardian", 1.9975, 1.9975},
	"wither_skeleton":       {"Wither Skeleton", 0.7, 2.4},
	"stray":                 {"Stray", 0.6, 1.99},
	"egg":                   {"entity.Egg.name", 0.25, 0.25},
	"leash_knot":            {"entity.LeashKnot.name", 0.375, 0.25},
	"painting":              {"Painting", 0.5, 0.5},
	"arrow":                 {"Arrow", 0.5, 0.5},
	"snowball":              {"Snowball", 0.25, 0.25},
	"fireball":              {"Fireball", 1, 1},
	"small_fireball":        {"Small Fireball", 0.3125, 0.3125},
	"ender_pearl":           {"entity.ThrownEnderpearl.name", 0.25, 0.25},
	"eye_of_ender_signal":   {"entity.EyeOfEnderSignal.name", 0.25, 0.25},
	"potion":                {"Potion", 0.25, 0.25},
	"xp_bottle":             {"entity.ThrownExpBottle.name", 0.25, 0.25},
	"item_frame":            {"entity.ItemFrame.name", 0.5, 0.5},
	"wither_skull":          {"entity.WitherSkull.name", 0.3125, 0.3125},
	"tnt":                   {"Block of TNT", 0.98, 0.98},
	"falling_block":         {"Falling Block", 0.98, 0.98},
	"fireworks_rocket":      {"entity.FireworksRocket.name", 0.25, 0.25},
	"husk":                  {"Husk", 0.6, 1.95},
	"spectral_arrow":        {"entity.SpectralArrow.name", 0.5, 0.5},
	"shulker_bullet":        {"entity.ShulkerBullet.name", 0.3125, 0.3125},
	"dragon_fireball":       {"Dragon Fireball", 1, 1},
	"zombie_villager":       {"Zombie Villager", 0.6, 1.95},
	"skeleton_horse":        {"Skeleton Horse", 1.396484, 1.6},
	"zombie_horse":          {"Zombie Horse", 1.396484, 1.6},
	"armor_stand":           {"Armor Stand", 0.5, 1.975},
	"donkey":                {"Donkey", 1.396484, 1.6},
	"mule":                  {"Mule", 1.396484, 1.6},
	"evocation_fangs":       {"entity.EvocationFangs.name", 0.5, 0.8},
	"evocation_illager":     {"Evoker", 0.6, 1.95},
	"vex":                   {"Vex", 0.4, 0.8},
	"vindication_illager":   {"Vindicator", 0.6, 1.95},
	"illusion_illager":      {"Illusioner", 0.6, 1.95},
	"commandblock_minecart": {"Minecart", 0.98, 0.7},
	"boat":                  {"Boat", 1.5, 0.6},
	"minecart":              {"Minecart", 0.98, 0.7},
	"chest_minecart":        {"Minecart with Chest", 0.98, 0.7},
	"furnace_minecart":      {"entity.MinecartFurnace.name", 0.98, 0.7},
	"tnt_minecart":          {"entity.MinecartTNT.name", 0.98, 0.7},
	"hopper_minecart":       {"Minecart with Hopper", 0.98, 0.7},
	"spawner_minecart":      {"entity.MinecartSpawner.name", 0.98, 0.7},
	"creeper":               {"Creeper", 0.6, 1.7},
	"skeleton":              {"Skeleton", 0.6, 1.99},
	"spider":                {"Spider", 1.4, 0.9},
	"giant":                 {"Giant", 3.6, 10.8},
	"zombie":                {"Zombie", 0.6, 1.95},
	"slime":                 {"Slime", 0.51, 0.51},
	"ghast":                 {"Ghast", 4, 4},
	"zombie_pigman":         {"Zombie Pigman", 0.6, 1.95},
	"enderman":              {"Enderman", 0.6, 2.9},
	"cave_spider":           {"Cave Spider", 0.7, 0.5},
	"silverfish":            {"Silverfish", 0.4, 0.3},
	"blaze":                 {"Blaze", 0.6, 1.8},
	"magma_cube":            {"Magma Cube", 0.51, 0.51},
	"ender_dragon":          {"Ender Dragon", 16, 8},
	"wither":                {"Wither", 0.9, 3.5},
	"bat":                   {"Bat", 0.5, 0.9},
	"witch":                 {"Witch", 0.6, 1.95},
	"endermite":             {"Endermite", 0.4, 0.3},
	"guardian":              {"Guardian", 0.85, 0.85},
	"shulker":               {"Shulker", 1, 1},
	"pig":                   {"Pig", 0.9, 0.9},
	"sheep":                 {"Sheep", 0.9, 1.3},
	"cow":                   {"Cow", 0.9, 1.4},
	"chicken":               {"Chicken", 0.4, 0.7},
	"squid":                 {"Squid", 0.8, 0.8},
	"wolf":                  {"Wolf", 0.6, 0.85},
	"mooshroom":             {"Mooshroom", 0.9, 1.4},
	"snowman":               {"Snow Golem", 0.7, 1.9},
	"ocelot":                {"Ocelot", 0.6, 0.7},
	"villager_golem":        {"Iron Golem", 1.4, 2.7},
	"horse":                 {"Horse", 1.396484, 1.6},
	"rabbit":                {"Rabbit", 0.4, 0.5},
	"polar_bear":            {"Polar Bear", 1.3, 1.4},
	"llama":                 {"Llama", 0.9, 1.87},
	"llama_spit":            {"entity.LlamaSpit.name", 0.25, 0.25},
	"parrot":                {"Parrot", 0.5, 0.9},
	"villager":              {"Villager", 0.6, 1.95},
	"ender_crystal":         {"entity.EnderCrystal.name", 2, 2},
	"lightning_bolt":        {"Lightning Bolt", 0.6, 1.95},
	"player":                {"Player", 0.6, 1.8},
}

// entityMeta holds the 1.12.2 metadata of one entity type.
type entityMeta struct {
	// DisplayName is the name 1.12.2 reports for entities of this type.
	DisplayName string
	// Width is the entity's box width.
	Width float64
	// Height is the entity's box height.
	Height float64
}

// EntityTypeMeta returns the 1.12.2 metadata for an entity type.
//
// Unknown types fall back to the generic 0.6 x 1.95 mob box and an empty
// name, which is what 1.12.2 does for types it does not know.
func EntityTypeMeta(name string) entityMeta {
	if m, ok := entityTypeMeta[NormalizeEntityType(name)]; ok {
		return m
	}
	return entityMeta{Width: 0.6, Height: 1.95}
}

// EntityTypeKey splits an entity type name into its namespace and its
// name.
//
// 1.12.2 builds a `MinecraftKey` from the name: both parts are lower-cased,
// a name without a namespace means "minecraft", and the namespace is only
// taken from the text when the colon is neither the first nor the second
// character. So "zombie" and "minecraft:zombie" are the same key, and so
// are "a:zombie" and ":zombie" — but "foo:zombie" is a key of the "foo"
// namespace.
func EntityTypeKey(name string) (namespace, key string) {
	namespace, key = "minecraft", name
	if i := strings.IndexByte(name, ':'); i >= 0 {
		key = name[i+1:]
		if i > 1 {
			namespace = name[:i]
		}
	}
	return strings.ToLower(namespace), strings.ToLower(key)
}

// IsRegisteredEntityType reports whether the name is an entity type 1.12.2
// knows.
//
// The registry holds the "minecraft" namespace only, so a name in another
// namespace is not a type the server knows: 1.12.2 reports it as invalid
// and matches nothing. The same is true for a name the registry does not
// have at all, which is why the old spellings of the renamed types —
// "pigzombie" for the `zombie_pigman` of 1.12.2 — are not types.
func IsRegisteredEntityType(name string) bool {
	namespace, key := EntityTypeKey(name)
	if namespace != "minecraft" {
		return false
	}
	_, ok := entityTypeMeta[key]
	return ok
}
