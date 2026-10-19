package v1_12_2

import "github.com/mcsim/gomc/internal/parser"

// commandUsages holds the usage line 1.12.2 prints for a command.
//
// The keys are the command path, so "scoreboard players set" is the
// usage of `/scoreboard players set`. The texts come from the language
// file that ships with the 1.12.2 client.
var commandUsages = map[string]string{
	"advancement":                      "/advancement <grant|revoke|test> <player>",
	"advancement grant":                "/advancement grant <player> <only|until|from|through|everything>",
	"advancement grant everything":     "/advancement grant <player> everything",
	"advancement grant from":           "/advancement grant <player> from <advancement>",
	"advancement grant only":           "/advancement grant <player> only <advancement> [criterion]",
	"advancement grant through":        "/advancement grant <player> through <advancement>",
	"advancement grant until":          "/advancement grant <player> until <advancement>",
	"advancement revoke":               "/advancement revoke <player> <only|until|from|through|everything>",
	"advancement revoke everything":    "/advancement revoke <player> everything",
	"advancement revoke from":          "/advancement revoke <player> from <advancement>",
	"advancement revoke only":          "/advancement revoke <player> only <advancement> [criterion]",
	"advancement revoke through":       "/advancement revoke <player> through <advancement>",
	"advancement revoke until":         "/advancement revoke <player> until <advancement>",
	"advancement test":                 "/advancement test <player> <advancement> [criterion]",
	"ban":                              "/ban <name> [reason ...]",
	"banip":                            "/ban-ip <address|name> [reason ...]",
	"banlist":                          "/banlist [ips|players]",
	"blockdata":                        "/blockdata <x> <y> <z> <dataTag>",
	"chunkinfo":                        "/chunkinfo [<x> <y> <z>]",
	"clear":                            "/clear [player] [item] [data] [maxCount] [dataTag]",
	"clone":                            "/clone <x1> <y1> <z1> <x2> <y2> <z2> <x> <y> <z> [maskMode] [cloneMode]",
	"compare":                          "/testforblocks <x1> <y1> <z1> <x2> <y2> <z2> <x> <y> <z> [mode]",
	"debug":                            "/debug <start|stop>",
	"defaultgamemode":                  "/defaultgamemode <mode>",
	"deop":                             "/deop <player>",
	"difficulty":                       "/difficulty <new difficulty>",
	"downfall":                         "/toggledownfall",
	"effect":                           "/effect <player> <effect> [seconds] [amplifier] [hideParticles] OR /effect <player> clear",
	"enchant":                          "/enchant <player> <enchantment ID> [level]",
	"entitydata":                       "/entitydata <entity> <dataTag>",
	"execute":                          "/execute <entity> <x> <y> <z> <command> OR /execute <entity> <x> <y> <z> detect <x> <y> <z> <block> <dataValue|-1|state|*> <command>",
	"fill":                             "/fill <x1> <y1> <z1> <x2> <y2> <z2> <block> [dataValue|state] [oldBlockHandling] [dataTag]",
	"function":                         "/function <name> [if <selector>|unless <selector>]",
	"gamemode":                         "/gamemode <mode> [player]",
	"gamerule":                         "/gamerule <rule name> [value]",
	"generic":                          "Usage: %s",
	"give":                             "/give <player> <item> [amount] [data] [dataTag]",
	"help":                             "/help [page|command name]",
	"kick":                             "/kick <player> [reason ...]",
	"kill":                             "/kill [player|entity]",
	"locate":                           "/locate <feature>",
	"me":                               "/me <action ...>",
	"message":                          "/tell <player> <private message ...>",
	"op":                               "/op <player>",
	"particle":                         "/particle <name> <x> <y> <z> <xd> <yd> <zd> <speed> [count] [mode] [player] [params]",
	"players":                          "/list",
	"playsound":                        "/playsound <sound> <source> <player> [x] [y] [z] [volume] [pitch] [minimumVolume]",
	"publish":                          "/publish",
	"recipe":                           "/recipe <give|take> [player] <name|*>",
	"reload":                           "/reload",
	"replaceitem":                      "/replaceitem <entity|block> ...",
	"replaceitem block":                "/replaceitem block <x> <y> <z> <slot> <item> [amount] [data] [dataTag]",
	"replaceitem entity":               "/replaceitem entity <selector> <slot> <item> [amount] [data] [dataTag]",
	"save":                             "/save-all [flush]",
	"say":                              "/say <message ...>",
	"scoreboard":                       "/scoreboard <objectives|players|teams> ...",
	"scoreboard objectives":            "/scoreboard objectives <list|add|remove|setdisplay> ...",
	"scoreboard objectives add":        "/scoreboard objectives add <name> <criteriaType> [display name ...]",
	"scoreboard objectives remove":     "/scoreboard objectives remove <name>",
	"scoreboard objectives setdisplay": "/scoreboard objectives setdisplay <slot> [objective]",
	"scoreboard players":               "/scoreboard players <set|add|remove|reset|list|enable|test|operation|tag> ...",
	"scoreboard players add":           "/scoreboard players add <player> <objective> <count> [dataTag]",
	"scoreboard players enable":        "/scoreboard players enable <player> <trigger>",
	"scoreboard players list":          "/scoreboard players list [name]",
	"scoreboard players operation":     "/scoreboard players operation <targetName> <targetObjective> <operation> <selector> <objective>",
	"scoreboard players remove":        "/scoreboard players remove <player> <objective> <count> [dataTag]",
	"scoreboard players reset":         "/scoreboard players reset <player> [objective]",
	"scoreboard players set":           "/scoreboard players set <player> <objective> <score> [dataTag]",
	"scoreboard players tag":           "/scoreboard players tag <player> <add|remove|list> <tagName> [dataTag]",
	"scoreboard players test":          "/scoreboard players test <player> <objective> <min> [max]",
	"scoreboard teams":                 "/scoreboard teams <list|add|remove|empty|join|leave|option> ...",
	"scoreboard teams add":             "/scoreboard teams add <name> [display name ...]",
	"scoreboard teams empty":           "/scoreboard teams empty <team>",
	"scoreboard teams join":            "/scoreboard teams join <team> [player]",
	"scoreboard teams leave":           "/scoreboard teams leave [player]",
	"scoreboard teams list":            "/scoreboard teams list [name]",
	"scoreboard teams option":          "/scoreboard teams option <team> <friendlyfire|color|seeFriendlyInvisibles|nametagVisibility|deathMessageVisibility|collisionRule> <value>",
	"scoreboard teams remove":          "/scoreboard teams remove <name>",
	"seed":                             "/seed",
	"setblock":                         "/setblock <x> <y> <z> <block> [dataValue|state] [oldBlockHandling] [dataTag]",
	"setidletimeout":                   "/setidletimeout <minutes>",
	"setworldspawn":                    "/setworldspawn [<x> <y> <z>]",
	"spawnpoint":                       "/spawnpoint [player] [<x> <y> <z>]",
	"spreadplayers":                    "/spreadplayers <x> <z> <spreadDistance> <maxRange> <respectTeams true|false> <player ...>",
	"stats":                            "/stats <entity|block> ...",
	"stats block":                      "/stats block <x> <y> <z> <mode> ...",
	"stats block clear":                "/stats block <x> <y> <z> clear <stat>",
	"stats block set":                  "/stats block <x> <y> <z> set <stat> <selector> <objective>",
	"stats entity":                     "/stats entity <selector> <mode> ...",
	"stats entity clear":               "/stats entity <selector> clear <stat>",
	"stats entity set":                 "/stats entity <selector> set <stat> <selector> <objective>",
	"stop":                             "/stop",
	"stopsound":                        "/stopsound <player> [source] [sound]",
	"summon":                           "/summon <entityname> [x] [y] [z] [dataTag]",
	"teleport":                         "/teleport <entity> <x> <y> <z> [<y-rot> <x-rot>]",
	"tellraw":                          "/tellraw <player> <raw json message>",
	"testfor":                          "/testfor <player> [dataTag]",
	"testforblock":                     "/testforblock <x> <y> <z> <block> [dataValue|-1|state|*] [dataTag]",
	"time":                             "/time <set|add|query> <value>",
	"title":                            "/title <player> title|subtitle|actionbar|clear|reset|times ...",
	"tp":                               "/tp [target player] <destination player> OR /tp [target player] <x> <y> <z> [<yaw> <pitch>]",
	"trigger":                          "/trigger <objective> <add|set> <value>",
	"unban":                            "/pardon <name>",
	"unbanip":                          "/pardon-ip <address>",
	"weather":                          "/weather <clear|rain|thunder> [duration in seconds]",
	"whitelist":                        "/whitelist <on|off|list|add|remove|reload>",
	"whitelist add":                    "/whitelist add <player>",
	"whitelist remove":                 "/whitelist remove <player>",
	"worldborder":                      "/worldborder <set|center|damage|warning|get|add> ...",
	"worldborder add":                  "/worldborder add <sizeInBlocks> [timeInSeconds]",
	"worldborder center":               "/worldborder center <x> <z>",
	"worldborder damage":               "/worldborder damage <buffer|amount> ...",
	"worldborder damage amount":        "/worldborder damage amount <damagePerBlock>",
	"worldborder damage buffer":        "/worldborder damage buffer <sizeInBlocks>",
	"worldborder set":                  "/worldborder set <sizeInBlocks> [timeInSeconds]",
	"worldborder warning":              "/worldborder warning <time|distance> ...",
	"worldborder warning distance":     "/worldborder warning distance <distance>",
	"worldborder warning time":         "/worldborder warning time <seconds>",
	"xp":                               "/xp <amount> [player] OR /xp <amount>L [player]",
}

// commandUsage returns the usage line 1.12.2 prints for a command path,
// or an empty string when the language file has none.
func commandUsage(key string) string {
	return commandUsages[key]
}

// usageError returns a parse error that carries the usage line of the
// command that was being parsed.
//
// It is the same error the parser would return, with the usage attached,
// so the CLI can show the correct spelling instead of just saying that
// something was wrong.
func usageError(p *parser.Parser, key, format string, args ...any) error {
	err := p.Errorf(format, args...)
	if pe, ok := err.(*parser.Error); ok {
		if usage := commandUsage(key); usage != "" {
			pe.Usage = usage
		}
	}
	return err
}

// withUsage attaches the usage line of a command to an error that already
// exists.
//
// It is used where the failure happens deeper in the parser, so the
// command path is known but the error was not created here.
func withUsage(err error, key string) error {
	if err == nil {
		return nil
	}
	if pe, ok := err.(*parser.Error); ok {
		if pe.Usage == "" {
			if usage := commandUsage(key); usage != "" {
				pe.Usage = usage
			}
		}
	}
	return err
}
