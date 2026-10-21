package v1_12_2

import (
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// The command result channels of 1.12.2.
//
// A channel is a way of getting the result of a command into the
// scoreboard. 1.12.2 has five of them, and they are the constants of
// `CommandObjectiveExecutor.EnumCommandResult`:
//
//	SuccessCount      AffectedBlocks   AffectedEntities
//	AffectedItems     QueryResult
//
// The names are the ones the `/stats` command takes, and the strange
// thing about them is that the `/stats` command does not take a
// **statistic** but one of these five names -- which is why the command
// fails with "Invalid parameters" for every other name.
//
// (The names come from the `EnumCommandResult` of the 1.12.2 server:
// the byte code of the class gives them one by one.)
const (
	// ResultSuccessCount counts the commands that ran.
	ResultSuccessCount = "SuccessCount"
	// ResultAffectedBlocks counts the blocks a command changed.
	ResultAffectedBlocks = "AffectedBlocks"
	// ResultAffectedEntities counts the entities a command changed.
	ResultAffectedEntities = "AffectedEntities"
	// ResultAffectedItems counts the items a command changed.
	ResultAffectedItems = "AffectedItems"
	// ResultQueryResult is the result of a query, like the time.
	ResultQueryResult = "QueryResult"
)

// statsResults are the five command results, in the order of the enum of
// the server.
var statsResults = []string{
	ResultSuccessCount,
	ResultAffectedBlocks,
	ResultAffectedEntities,
	ResultAffectedItems,
	ResultQueryResult,
}

// statsChannel is one channel of an entity.
//
// A channel says where the result of a command goes: the score of the
// holder in the objective.
type statsChannel struct {
	// holder is the name the score is stored under.
	holder string
	// objective is the objective the score is stored in.
	objective string
}

// resolveStatsHolder turns the holder of a channel into the name the
// score is stored under.
//
// 1.12.2 resolves it in three steps (`CommandAbstract.f`):
//
//  1. as a **player**, which gives the name of the player;
//  2. as an **entity**, which gives the UUID of the entity;
//  3. otherwise the text as it was written.
//
// A name that is none of those is kept as it is, so a channel can point at
// a holder that does not exist yet -- 1.12.2 does not look at the holder
// when the channel is bound either.
//
// (Measured on a real 1.12.2 server, probe case 61-stats-channel: `stats
// entity @s set QueryResult @s return` stores the result under the name
// of the player `@s` stands for, and the score of `return` changes.)
func resolveStatsHolder(ctx *exec.Context, holder string) string {
	// A player comes first: the name of the player is the name of the
	// holder.
	if e := ctx.World.FindPlayer(holder); e != nil {
		return ScoreboardHolder(e)
	}
	// Then an entity, which is tracked under its UUID.
	if e := entityByUUID(ctx, holder); e != nil {
		return ScoreboardHolder(e)
	}
	// A selector names the entities it matches, and the first one is the
	// holder.
	if IsSelector(holder) {
		if sel, err := ParseSelector(holder); err == nil {
			resolver := &SelectorResolver{World: ctx.World}
			entities, err := resolver.Resolve(ctx.Source, sel)
			if err == nil && len(entities) > 0 {
				return ScoreboardHolder(entities[0])
			}
		}
	}
	return holder
}

// statsChannels returns the channels of an entity.
//
// 1.12.2 keeps them in the `CommandStats` tag of the entity, under the
// name of the result, and the objective next to it:
//
//	CommandStats: { QueryResultName: "<holder>", QueryResultObjective: "<objective>" }
//
// (`CommandObjectiveExecutor.b` writes the tag and
// `CommandObjectiveExecutor.a` reads it back.)
func statsChannels(e *world.Entity) map[string]statsChannel {
	out := map[string]statsChannel{}
	tag, ok := e.NBTValue("CommandStats").(*nbt.Compound)
	if !ok {
		return out
	}
	for _, result := range statsResults {
		name, ok := tag.Get(result + "Name").(nbt.String)
		if !ok || string(name) == "" {
			continue
		}
		objective, ok := tag.Get(result + "Objective").(nbt.String)
		if !ok || string(objective) == "" {
			continue
		}
		out[result] = statsChannel{holder: string(name), objective: string(objective)}
	}
	return out
}

// setStatsChannel binds a channel of an entity.
//
// 1.12.2 writes the name of the holder and the name of the objective
// into the `CommandStats` tag of the entity.
func setStatsChannel(e *world.Entity, result, holder, objective string) {
	tag, ok := e.NBTValue("CommandStats").(*nbt.Compound)
	if !ok {
		tag = nbt.NewCompound()
	}
	tag.Set(result+"Name", nbt.String(holder))
	tag.Set(result+"Objective", nbt.String(objective))
	e.NBT.Set("CommandStats", tag)
}

// clearStatsChannel unbinds a channel of an entity.
//
// 1.12.2 removes both keys, and when the last channel is gone it removes
// the tag as well.
func clearStatsChannel(e *world.Entity, result string) {
	tag, ok := e.NBTValue("CommandStats").(*nbt.Compound)
	if !ok {
		return
	}
	tag.Set(result+"Name", nbt.String(""))
	tag.Set(result+"Objective", nbt.String(""))
	// The tag is only kept when a channel is still bound.
	empty := true
	for _, other := range statsResults {
		name, _ := tag.Get(other + "Name").(nbt.String)
		if string(name) != "" {
			empty = false
			break
		}
	}
	if empty {
		e.NBT.Remove("CommandStats")
		return
	}
	e.NBT.Set("CommandStats", tag)
}

// reportStatsResult writes the result of a command into the scoreboard.
//
// The result goes into the channel of the entity that ran the command, if
// it has one. 1.12.2 writes the score only when the objective exists
// **and** the holder already has a score in it: `Scoreboard.b` is asked
// first, and it says whether the entry is there.
//
// So a channel whose objective does not exist, or whose holder has no
// score in that objective yet, writes nothing.
//
// (`CommandObjectiveExecutor.a`: it looks the objective up, asks
// `Scoreboard.b(holder, objective)`, and only then sets the score.)
func reportStatsResult(ctx *exec.Context, e *world.Entity, result string, value int) {
	channel, ok := statsChannels(e)[result]
	if !ok {
		return
	}
	sb := ctx.World.Scoreboard
	if sb.Objective(channel.objective) == nil {
		return
	}
	if _, tracked := sb.Score(channel.holder, channel.objective); !tracked {
		return
	}
	sb.SetScore(channel.holder, channel.objective, value)
}

// statsReporter feeds the result of a command into the channels of the
// entity that ran it.
//
// It is the version-specific half of the reporting: the engine knows
// that a command ran and what it returned, and this turns that into the
// five channels of 1.12.2.
type statsReporter struct{}

// Report implements exec.ResultReporter.
func (r *statsReporter) Report(ctx *exec.Context, successCount, affectedBlocks, affectedEntities, affectedItems, queryResult int) {
	entity := ctx.Source.Entity
	if entity == nil {
		// The console has no channels: 1.12.2 keeps them on the entity
		// that ran the command.
		return
	}
	if successCount > 0 {
		reportStatsResult(ctx, entity, ResultSuccessCount, successCount)
	}
	if affectedBlocks >= 0 {
		reportStatsResult(ctx, entity, ResultAffectedBlocks, affectedBlocks)
	}
	if affectedEntities >= 0 {
		reportStatsResult(ctx, entity, ResultAffectedEntities, affectedEntities)
	}
	if affectedItems >= 0 {
		reportStatsResult(ctx, entity, ResultAffectedItems, affectedItems)
	}
	if queryResult >= 0 {
		reportStatsResult(ctx, entity, ResultQueryResult, queryResult)
	}
}
