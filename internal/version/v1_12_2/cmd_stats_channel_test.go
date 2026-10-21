package v1_12_2

import "testing"

// TestStatsChannelQueryResult covers the channel of a query.
//
// 1.12.2's `/stats entity set` does not only print a message: it binds
// a channel, and the result of the commands the entity runs afterwards
// goes into the scoreboard.
//
// The channel belongs to the entity that runs the command, and the score
// is only written when the objective exists **and** the holder already
// has a score in it.
//
// (Measured on a real 1.12.2 server, probe case
// 62-stats-channel-write: after `stats entity @e[type=armor_stand] set
// QueryResult @e[type=armor_stand] return`, the command `execute
// @e[type=armor_stand] ~ ~ ~ time query daytime` writes the time into
// the objective `return` of the armour stand.)
func TestStatsChannelQueryResult(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "scoreboard objectives add return dummy"); err != nil {
		t.Fatalf("objectives add: %v", err)
	}
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	if err := runLine(e, "scoreboard players set @e[type=armor_stand] return 0"); err != nil {
		t.Fatalf("players set: %v", err)
	}
	if err := runLine(e, "stats entity @e[type=armor_stand] set QueryResult @e[type=armor_stand] return"); err != nil {
		t.Fatalf("stats: %v", err)
	}
	if err := runLine(e, "time set 5000"); err != nil {
		t.Fatalf("time set: %v", err)
	}
	if err := runLine(e, "execute @e[type=armor_stand] ~ ~ ~ time query daytime"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	holder := e.World.EntitySlice()[0].EnsureUUID()
	score, tracked := e.World.Scoreboard.Score(holder, "return")
	if !tracked {
		t.Fatal("the holder has no score in the objective")
	}
	if score != 5000 {
		t.Errorf("the score is %d, want 5000", score)
	}
}

// TestStatsChannelNeedsTheEntry covers the second rule of the write: the
// holder must already have a score in the objective.
//
// (`CommandObjectiveExecutor.a` asks `Scoreboard.b(holder, objective)`
// first, and it says whether the entry is there.)
func TestStatsChannelNeedsTheEntry(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "scoreboard objectives add return dummy"); err != nil {
		t.Fatalf("objectives add: %v", err)
	}
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	if err := runLine(e, "stats entity @e[type=armor_stand] set QueryResult @e[type=armor_stand] return"); err != nil {
		t.Fatalf("stats: %v", err)
	}
	if err := runLine(e, "time set 5000"); err != nil {
		t.Fatalf("time set: %v", err)
	}
	if err := runLine(e, "execute @e[type=armor_stand] ~ ~ ~ time query daytime"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	holder := e.World.EntitySlice()[0].EnsureUUID()
	if score, tracked := e.World.Scoreboard.Score(holder, "return"); tracked {
		t.Errorf("the score is %d, want no score at all", score)
	}
}

// TestStatsChannelNeedsTheObjective covers the first rule: the objective
// has to exist.
func TestStatsChannelNeedsTheObjective(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	if err := runLine(e, "stats entity @e[type=armor_stand] set QueryResult @e[type=armor_stand] return"); err != nil {
		t.Fatalf("stats: %v", err)
	}
	if err := runLine(e, "time set 5000"); err != nil {
		t.Fatalf("time set: %v", err)
	}
	if err := runLine(e, "execute @e[type=armor_stand] ~ ~ ~ time query daytime"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	holder := e.World.EntitySlice()[0].EnsureUUID()
	if _, tracked := e.World.Scoreboard.Score(holder, "return"); tracked {
		t.Error("the score was written, but the objective does not exist")
	}
}

// TestStatsChannelClear covers that a cleared channel writes nothing
// anymore.
//
// (Measured on a real 1.12.2 server, probe case
// 62-stats-channel-write: after `stats entity @e[type=armor_stand] clear
// QueryResult` the score of `return` stays where it was.)
func TestStatsChannelClear(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "scoreboard objectives add return dummy"); err != nil {
		t.Fatalf("objectives add: %v", err)
	}
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	if err := runLine(e, "scoreboard players set @e[type=armor_stand] return 0"); err != nil {
		t.Fatalf("players set: %v", err)
	}
	if err := runLine(e, "stats entity @e[type=armor_stand] set QueryResult @e[type=armor_stand] return"); err != nil {
		t.Fatalf("stats: %v", err)
	}
	if err := runLine(e, "time set 5000"); err != nil {
		t.Fatalf("time set: %v", err)
	}
	if err := runLine(e, "execute @e[type=armor_stand] ~ ~ ~ time query daytime"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	holder := e.World.EntitySlice()[0].EnsureUUID()
	if score, _ := e.World.Scoreboard.Score(holder, "return"); score != 5000 {
		t.Fatalf("the score is %d, want 5000", score)
	}

	if err := runLine(e, "stats entity @e[type=armor_stand] clear QueryResult"); err != nil {
		t.Fatalf("stats clear: %v", err)
	}
	if err := runLine(e, "time set 6000"); err != nil {
		t.Fatalf("time set: %v", err)
	}
	if err := runLine(e, "execute @e[type=armor_stand] ~ ~ ~ time query daytime"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if score, _ := e.World.Scoreboard.Score(holder, "return"); score != 5000 {
		t.Errorf("the score is %d, want 5000 -- the channel was cleared", score)
	}
}

// TestGameruleReportsItsValue covers the `QueryResult` of a game rule.
//
// 1.12.2 reports the value of the rule that was asked for, and a boolean
// rule is 0 or 1.
//
// (`CommandGamerule` reports `EnumCommandResult.QUERY_RESULT` with
// `GameRules.c(rule)`.)
func TestGameruleReportsItsValue(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "scoreboard objectives add return dummy"); err != nil {
		t.Fatalf("objectives add: %v", err)
	}
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	if err := runLine(e, "scoreboard players set @e[type=armor_stand] return 0"); err != nil {
		t.Fatalf("players set: %v", err)
	}
	if err := runLine(e, "stats entity @e[type=armor_stand] set QueryResult @e[type=armor_stand] return"); err != nil {
		t.Fatalf("stats: %v", err)
	}
	if err := runLine(e, "execute @e[type=armor_stand] ~ ~ ~ gamerule randomTickSpeed"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	holder := e.World.EntitySlice()[0].EnsureUUID()
	score, tracked := e.World.Scoreboard.Score(holder, "return")
	if !tracked {
		t.Fatal("the holder has no score in the objective")
	}
	// The rule `randomTickSpeed` is 3 by default.
	if score != 3 {
		t.Errorf("the score is %d, want 3", score)
	}
}

// TestStatsChannelSuccessCount covers the count of the commands that ran.
//
// (Measured on a real 1.12.2 server, probe case
// 62-stats-channel-write: the command `execute @e[type=armor_stand] ~ ~ ~
// time query daytime` writes 1 into the `SuccessCount` channel.)
func TestStatsChannelSuccessCount(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "scoreboard objectives add cnt dummy"); err != nil {
		t.Fatalf("objectives add: %v", err)
	}
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	if err := runLine(e, "scoreboard players set @e[type=armor_stand] cnt 0"); err != nil {
		t.Fatalf("players set: %v", err)
	}
	if err := runLine(e, "stats entity @e[type=armor_stand] set SuccessCount @e[type=armor_stand] cnt"); err != nil {
		t.Fatalf("stats: %v", err)
	}
	if err := runLine(e, "execute @e[type=armor_stand] ~ ~ ~ time query daytime"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	holder := e.World.EntitySlice()[0].EnsureUUID()
	score, tracked := e.World.Scoreboard.Score(holder, "cnt")
	if !tracked {
		t.Fatal("the holder has no score in the objective")
	}
	if score != 1 {
		t.Errorf("the score is %d, want 1", score)
	}
}

// TestStatsChannelIsKeptOnTheEntity covers that the channel belongs to the
// entity and travels with it.
func TestStatsChannelIsKeptOnTheEntity(t *testing.T) {
	e := newTestEngine(t)
	if err := runLine(e, "summon armor_stand 0 4 0"); err != nil {
		t.Fatalf("summon: %v", err)
	}
	if err := runLine(e, "stats entity @e[type=armor_stand] set QueryResult @e[type=armor_stand] return"); err != nil {
		t.Fatalf("stats: %v", err)
	}
	stand := e.World.EntitySlice()[0]
	channels := statsChannels(stand)
	channel, ok := channels[ResultQueryResult]
	if !ok {
		t.Fatalf("the entity has no QueryResult channel, it has %v", channels)
	}
	if channel.objective != "return" {
		t.Errorf("the objective is %q, want return", channel.objective)
	}
	if channel.holder != stand.EnsureUUID() {
		t.Errorf("the holder is %q, want %q", channel.holder, stand.EnsureUUID())
	}
}
