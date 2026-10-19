// Package v1_12_2 implements the Minecraft 1.12.2 command set.
//
// The package owns every version-specific decision: which commands exist,
// how their arguments are shaped, and how they change the world. The core
// engine in package exec stays version-agnostic.
//
// The 1.12.2 dialect differs from modern Minecraft in ways that matter:
//
//   - `/execute` uses the pre-1.13 syntax `execute <entity> <x> <y> <z>
//     [detect ...] <command>` and has no `as`/`at`/`run` subcommands.
//   - `/data` does not exist; `/entitydata` and `/blockdata` do.
//   - `/testfor`, `/testforblock` and `/testforblocks` exist.
//   - Functions live in `data/functions/<namespace>/<name>.mcfunction`
//     rather than in a data pack, and the per-tick hook is the
//     `gameLoopFunction` game rule.
package v1_12_2

import "github.com/mcsim/gomc/internal/exec"

// Register adds every 1.12.2 command to the registry.
func Register(reg *exec.Registry) {
	registerCore(reg)
	registerTime(reg)
	registerEnvironment(reg)
	registerSpawn(reg)
	registerWorldBorder(reg)
	registerFunction(reg)
	registerScoreboard(reg)
	registerEntity(reg)
	registerWorld(reg)
	registerBlocksExtra(reg)
	registerItems(reg)
	registerOutput(reg)
	registerMisc(reg)
	registerPlayer(reg)
}
