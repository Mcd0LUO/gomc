package world

// Entity type metadata for Minecraft 1.12.2.
//
// Two properties are needed to reproduce the real server's selector
// behaviour:
//
//   - The display name, because `name=` matches the name 1.12.2 reports for
//     an entity. `Entity.getName` returns the custom name when the entity
//     has one, the player name for players, and otherwise the type's name
//     as registered in `EntityTypes` ("Zombie", "ArmorStand", ...).
//   - The bounding box size, because the `dx`/`dy`/`dz` volume filter tests
//     whether the entity's box *intersects* the volume, not whether its
//     position point falls inside it.
//
// The names come from `EntityTypes.a(int, String, Class, String)` in the
// 1.12.2 NMS; the sizes come from the 1.12 entity data.
var entityTypeMeta = map[string]entityMeta{
	"item":                  {"Item", 0.25, 0.25},
	"xp_orb":                {"XPOrb", 0.5, 0.5},
	"area_effect_cloud":     {"AreaEffectCloud", 0.5, 0.5},
	"elder_guardian":        {"ElderGuardian", 1.9975, 1.9975},
	"wither_skeleton":       {"WitherSkeleton", 0.7, 2.4},
	"stray":                 {"Stray", 0.6, 1.99},
	"egg":                   {"ThrownEgg", 0.25, 0.25},
	"leash_knot":            {"LeashKnot", 0.375, 0.25},
	"painting":              {"Painting", 0.5, 0.5},
	"arrow":                 {"Arrow", 0.5, 0.5},
	"snowball":              {"Snowball", 0.25, 0.25},
	"fireball":              {"Fireball", 1, 1},
	"small_fireball":        {"SmallFireball", 0.3125, 0.3125},
	"ender_pearl":           {"ThrownEnderpearl", 0.25, 0.25},
	"eye_of_ender_signal":   {"EyeOfEnderSignal", 0.25, 0.25},
	"potion":                {"ThrownPotion", 0.25, 0.25},
	"xp_bottle":             {"ThrownExpBottle", 0.25, 0.25},
	"item_frame":            {"ItemFrame", 0.5, 0.5},
	"wither_skull":          {"WitherSkull", 0.3125, 0.3125},
	"tnt":                   {"PrimedTnt", 0.98, 0.98},
	"falling_block":         {"FallingSand", 0.98, 0.98},
	"fireworks_rocket":      {"FireworksRocketEntity", 0.25, 0.25},
	"husk":                  {"Husk", 0.6, 1.95},
	"spectral_arrow":        {"SpectralArrow", 0.5, 0.5},
	"shulker_bullet":        {"ShulkerBullet", 0.3125, 0.3125},
	"dragon_fireball":       {"DragonFireball", 1, 1},
	"zombie_villager":       {"ZombieVillager", 0.6, 1.95},
	"skeleton_horse":        {"SkeletonHorse", 1.396484, 1.6},
	"zombie_horse":          {"ZombieHorse", 1.396484, 1.6},
	"armor_stand":           {"ArmorStand", 0.5, 1.975},
	"donkey":                {"Donkey", 1.396484, 1.6},
	"mule":                  {"Mule", 1.396484, 1.6},
	"evocation_fangs":       {"EvocationFangs", 0.5, 0.8},
	"evocation_illager":     {"EvocationIllager", 0.6, 1.95},
	"vex":                   {"Vex", 0.4, 0.8},
	"vindication_illager":   {"VindicationIllager", 0.6, 1.95},
	"illusion_illager":      {"IllusionIllager", 0.6, 1.95},
	"commandblock_minecart": {"Minecart", 0.98, 0.7},
	"boat":                  {"Boat", 1.5, 0.6},
	"minecart":              {"Minecart", 0.98, 0.7},
	"chest_minecart":        {"Chest Minecart", 0.98, 0.7},
	"furnace_minecart":      {"Furnace Minecart", 0.98, 0.7},
	"tnt_minecart":          {"TNT Minecart", 0.98, 0.7},
	"hopper_minecart":       {"Hopper Minecart", 0.98, 0.7},
	"spawner_minecart":      {"Spawner Minecart", 0.98, 0.7},
	"creeper":               {"Creeper", 0.6, 1.7},
	"skeleton":              {"Skeleton", 0.6, 1.99},
	"spider":                {"Spider", 1.4, 0.9},
	"giant":                 {"Giant", 3.6, 10.8},
	"zombie":                {"Zombie", 0.6, 1.95},
	"slime":                 {"Slime", 0.51, 0.51},
	"ghast":                 {"Ghast", 4, 4},
	"zombie_pigman":         {"PigZombie", 0.6, 1.95},
	"enderman":              {"Enderman", 0.6, 2.9},
	"cave_spider":           {"CaveSpider", 0.7, 0.5},
	"silverfish":            {"Silverfish", 0.4, 0.3},
	"blaze":                 {"Blaze", 0.6, 1.8},
	"magma_cube":            {"LavaSlime", 0.51, 0.51},
	"ender_dragon":          {"EnderDragon", 16, 8},
	"wither":                {"WitherBoss", 0.9, 3.5},
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
	"mooshroom":             {"MushroomCow", 0.9, 1.4},
	"snowman":               {"SnowMan", 0.7, 1.9},
	"ocelot":                {"Ozelot", 0.6, 0.7},
	"villager_golem":        {"VillagerGolem", 1.4, 2.7},
	"horse":                 {"Horse", 1.396484, 1.6},
	"rabbit":                {"Rabbit", 0.4, 0.5},
	"polar_bear":            {"PolarBear", 1.3, 1.4},
	"llama":                 {"Llama", 0.9, 1.87},
	"llama_spit":            {"LlamaSpit", 0.25, 0.25},
	"parrot":                {"Parrot", 0.5, 0.9},
	"villager":              {"Villager", 0.6, 1.95},
	"ender_crystal":         {"EnderCrystal", 2, 2},
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
