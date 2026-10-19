package exec

import "strings"

// NormalizeFunctionID turns a function id into the form 1.12.2 uses.
//
// A function is identified by a `MinecraftKey`, whose constructor
// lower-cases the namespace and the path, and defaults the namespace to
// "minecraft". So "PROBE:mcns" is "probe:mcns", and a bare "mcns" means
// "minecraft:mcns".
func NormalizeFunctionID(id string) string {
	ns, name, found := strings.Cut(id, ":")
	if !found {
		return "minecraft:" + strings.ToLower(id)
	}
	return strings.ToLower(ns) + ":" + strings.ToLower(name)
}
