package v1_12_2

// CommandGroup is a set of commands that belong together.
//
// The groups are only for the listing of the command set: they say nothing
// about how a command behaves.
type CommandGroup struct {
	// Name is the name of the group, e.g. "world".
	Name string
	// Commands are the names of the commands in the group.
	Commands []string
}

// CommandGroups returns the command set, grouped by domain.
//
// Every command gomc knows belongs to exactly one group, so the listing
// covers all of them (the test checks that).
//
// The grouping is for the reader: 52 commands in one list are hard to
// scan, and the domain is what a person looks for ("which command sets a
// block?").
func CommandGroups() []CommandGroup {
	return []CommandGroup{
		{
			Name: "world",
			Commands: []string{
				"setblock", "fill", "clone", "blockdata",
				"testforblock", "testforblocks",
				"setworldspawn", "spawnpoint", "worldborder", "locate",
			},
		},
		{
			Name: "entity",
			Commands: []string{
				"summon", "kill", "tp", "teleport", "effect", "enchant",
				"gamemode", "defaultgamemode", "clear", "give", "xp", "trigger",
				"title", "testfor", "entitydata", "replaceitem", "stats",
				"spreadplayers",
			},
		},
		{
			Name: "player",
			Commands: []string{
				"advancement", "recipe",
			},
		},
		{
			Name: "scoreboard",
			Commands: []string{
				"scoreboard",
			},
		},
		{
			Name: "chat",
			Commands: []string{
				"say", "me", "tell", "tellraw", "particle", "playsound", "stopsound",
			},
		},
		{
			Name: "time",
			Commands: []string{
				"time", "weather", "difficulty", "toggledownfall", "seed",
				"setidletimeout", "gamerule",
			},
		},
		{
			Name: "function",
			Commands: []string{
				"function", "execute",
			},
		},
		{
			Name: "misc",
			Commands: []string{
				"help", "list", "debug", "reload",
			},
		},
		{
			Name: "gomc",
			Commands: []string{
				"gomcplayer",
			},
		},
	}
}
