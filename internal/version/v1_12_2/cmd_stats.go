package v1_12_2

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/ast"
	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/parser"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// statsCmd implements `/stats`.
//
// 1.12.2 的 `/stats` 实际上是**坏的**：它把统计项的名字当成了"命令结果
// 类型"（`CommandObjectiveExecutor$EnumCommandResult`）去映射，映射不到
// 就抛 `commands.stats.failed`，也就是 "Invalid parameters"。所以这条命令
// 在真服上几乎总是失败——只有统计项的名字恰好是那五个命令结果名之一时
// 才过得去。
//
// 除此之外，1.12.2 检查参数的顺序也和平时的直觉不同，而且 entity 形式
// 和 block 形式还不一样（2026-10-21 在真实 1.12.2 服务端上逐条实测）：
//
//  1. **entity 形式先看目标。** 目标是列表选择器（@e / @a）却一个都没匹配
//     到时，命令**什么都不说**就结束了——列表展开成零次执行。目标是单个
//     选择器（@p / @s）或一个玩家名却解析不出来时，报的是 entity 的用法，
//     而不是 "Entity '@p' cannot be found"。
//  2. **block 形式先数参数。** 少于 5 个报 block 的用法；set 形式少于 8 个
//     报 set 的用法；clear 形式少于 6 个报 clear 的用法。
//  3. **参数齐了才看统计项。** 统计项映射不到就是 "Invalid parameters"；
//     映射到了才会去看位置对不对、方块能不能统计。
//  4. **方块一律存不了统计。** 箱子、漏斗、投掷器、熔炉、酿造台、信标都
//     是同一句 "Block at X, Y, Z can not track stats"——方块的统计是 1.13
//     才有的东西，1.12.2 没有。
//  5. **持有者和目标只是回显。** 1.12.2 不去解析它们：把一个不存在的
//     玩家名当持有者，它照样说 "Storing ... on Steve"。
//
// 用例 tools/probe/cases/08-stats-params.case、09-stats-entity.case、
// 10-stats-target.case、11-stats-block.case 和 12-stats-names.case 记着
// 这些实测。
type statsCmd struct{}

// Name implements exec.Handler.
func (c *statsCmd) Name() string { return "stats" }

// Usage implements exec.Handler.
func (c *statsCmd) Usage() string { return usageLine("stats") }

// statsArgs is the parsed shape of a `/stats` command.
//
// 1.12.2 looks at the arguments in an order the language file does not
// suggest, and the entity form and the block form even differ, so
// everything is kept in the order it was written and the rules are
// applied when the command runs.
type statsArgs struct {
	// kind is "entity", "block", or "" when the command had no kind at
	// all.
	kind string
	// target is the entity form's target: a selector or a name.
	target ast.Arg
	// pos holds the three parts of the block form's position, as they
	// were written. They are parsed only when the command runs, because
	// 1.12.2 refuses a position only after it has accepted everything
	// else.
	pos []string
	// rest holds the tokens that follow the target or the position.
	rest []string
}

// ArgKind implements ast.Arg.
//
// The whole command is kept in one argument, because 1.12.2 looks at the
// arguments in an order that does not fit into a flat list.
func (a *statsArgs) ArgKind() ast.Kind { return ast.KindList }

// Parse implements exec.Handler.
//
// Only the shape of the command is read here. Everything that 1.12.2
// decides later -- whether the target exists, whether the arguments are
// enough, whether the statistic name maps -- is decided in Run, because
// the answers depend on the world and on the order 1.12.2 looks at the
// things.
func (c *statsCmd) Parse(p *parser.Parser) ([]ast.Arg, error) {
	kind, err := p.String()
	if err != nil {
		// 1.12.2 answers a command without a kind with the usage of the
		// command, so the error is raised when the command runs.
		return []ast.Arg{&statsArgs{}}, nil
	}
	switch kind {
	case "entity":
		spec := &statsArgs{kind: kind}
		t, ok := p.Peek()
		if !ok {
			return []ast.Arg{spec}, nil
		}
		if IsSelector(t.Text) {
			sel, err := parseSelector(p)
			if err != nil {
				// A selector that does not parse is reported as such,
				// before anything else is looked at.
				return nil, err
			}
			spec.target = sel
		} else {
			name, err := p.String()
			if err != nil {
				return []ast.Arg{spec}, nil
			}
			spec.target = &ast.StringArg{Value: name}
		}
		spec.rest = statsRest(p)
		return []ast.Arg{spec}, nil
	case "block":
		spec := &statsArgs{kind: kind}
		for i := 0; i < 3; i++ {
			t, ok := p.Next()
			if !ok {
				return []ast.Arg{spec}, nil
			}
			spec.pos = append(spec.pos, t.Text)
		}
		spec.rest = statsRest(p)
		return []ast.Arg{spec}, nil
	}
	return []ast.Arg{&statsArgs{kind: kind}}, nil
}

// statsRest reads the tokens that follow the target or the position.
func statsRest(p *parser.Parser) []string {
	var out []string
	for p.More() {
		t, ok := p.Next()
		if !ok {
			break
		}
		out = append(out, t.Text)
	}
	return out
}

// Run implements exec.Handler.
func (c *statsCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) {
	spec, ok := args[0].(*statsArgs)
	if !ok {
		return 0, usageErrorf("stats")
	}
	switch spec.kind {
	case "entity":
		return runStatsEntity(ctx, spec)
	case "block":
		return runStatsBlock(ctx, spec)
	}
	return 0, usageErrorf("stats")
}

// runStatsEntity runs the entity form of `/stats`.
func runStatsEntity(ctx *exec.Context, spec *statsArgs) (int, error) {
	if spec.target == nil {
		return 0, usageErrorf("stats entity")
	}
	// 1.12.2 looks at the target before it counts the arguments, so a
	// target that resolves to nothing decides the outcome on its own.
	entities, err := resolveStatsTarget(ctx, spec.target)
	if err != nil {
		return 0, err
	}
	if len(entities) == 0 {
		// A target list that matches nothing: the command runs zero
		// times and the server says nothing at all.
		return 0, nil
	}
	mode, err := statsMode(spec.rest, "stats entity")
	if err != nil {
		return 0, err
	}
	switch mode {
	case "set":
		if len(spec.rest) < 4 {
			return 0, usageErrorf("stats entity set")
		}
	case "clear":
		if len(spec.rest) < 2 {
			return 0, usageErrorf("stats entity clear")
		}
	}
	// The statistic name is read only once the arguments are complete.
	stat := spec.rest[1]
	if !statsIsCommandResult(stat) {
		return 0, exec.Errorf("Invalid parameters")
	}
	// The statistic name was one of the command result names, so 1.12.2
	// goes on and stores the statistic.
	//
	// The holder and the objective are **not** looked at: the server
	// writes them into the message and into the channel of the entity,
	// and it does not check that the holder is a player or that the
	// objective exists.
	//
	// (Measured on a real 1.12.2 server: `stats entity @e[type=armor_stand]
	// set QueryResult Steve obj` says "Storing QueryResult stats in obj on
	// Steve", even though there is no player called Steve.)
	//
	// The channel belongs to the **target** of the command, and it is
	// what makes the result of the commands the entity runs end up in
	// the scoreboard.
	if mode == "set" {
		// The holder is kept as it was written: 1.12.2 stores the text
		// of the argument, and the message repeats it as it is.
		//
		// (Measured on a real 1.12.2 server, probe case 61-stats-channel:
		// `stats entity @e[type=armor_stand] set QueryResult
		// @e[type=armor_stand] return` says "Storing QueryResult stats
		// in return on @e[type=armor_stand]", with the selector as it
		// was written.)
		holder := spec.rest[2]
		objective := spec.rest[3]
		for _, e := range entities {
			setStatsChannel(e, stat, resolveStatsHolder(ctx, holder), objective)
		}
		ctx.Engine.Feedback = append(ctx.Engine.Feedback,
			fmt.Sprintf("Storing %s stats in %s on %s", stat, objective, holder))
		return 1, nil
	}
	for _, e := range entities {
		clearStatsChannel(e, stat)
	}
	ctx.Engine.Feedback = append(ctx.Engine.Feedback,
		fmt.Sprintf("Cleared %s stats", stat))
	return 1, nil
}

// runStatsBlock runs the block form of `/stats`.
func runStatsBlock(ctx *exec.Context, spec *statsArgs) (int, error) {
	// 1.12.2 counts the arguments before it looks at any of them: a
	// block needs its three coordinates and a mode.
	if len(spec.pos) < 3 || len(spec.rest) < 1 {
		return 0, usageErrorf("stats block")
	}
	mode, err := statsMode(spec.rest, "stats block")
	if err != nil {
		return 0, err
	}
	switch mode {
	case "set":
		if len(spec.rest) < 4 {
			return 0, usageErrorf("stats block set")
		}
	case "clear":
		if len(spec.rest) < 2 {
			return 0, usageErrorf("stats block clear")
		}
	}
	// The statistic name is read only once the arguments are complete.
	if !statsIsCommandResult(spec.rest[1]) {
		return 0, exec.Errorf("Invalid parameters")
	}
	// The position is parsed only after that.
	pos, err := statsCoord(spec.pos, ctx.Source.RelativeBase())
	if err != nil {
		return 0, err
	}
	// 1.12.2 asks the block whether it can hold statistics at all, and in
	// 1.12.2 no block can: the statistics of a block are a 1.13 feature.
	// Every block is refused, containers included.
	//
	// (Measured on a real 1.12.2 server: a chest, a hopper, a dropper, a
	// furnace, a brewing stand and a beacon all get the same answer.)
	return 0, exec.Errorf("Block at %d, %d, %d can not track stats", pos.BlockPos().X,
		pos.BlockPos().Y, pos.BlockPos().Z)
}

// statsMode reads the mode of a `/stats` command.
//
// The mode is the first token after the target or the position. 1.12.2
// knows two of them, and a third one is reported as the usage of the
// command as a whole.
func statsMode(rest []string, usageKey string) (string, error) {
	if len(rest) < 1 {
		return "", usageErrorf(usageKey)
	}
	switch rest[0] {
	case "set", "clear":
		return rest[0], nil
	}
	return "", usageErrorf("stats")
}

// statsIsCommandResult reports whether the statistic name happens to be one
// of the command result types.
//
// This is the bug 1.12.2 has: the statistic name is mapped onto the
// `EnumCommandResult` of the command, so a real statistic name never maps
// and the command always fails with "Invalid parameters". Only these five
// names get past it.
//
// (Measured on a real 1.12.2 server: `stats block 0 4 0 set QueryResult
// @e[type=zombie] obj` is the one form that does not say "Invalid
// parameters" -- it goes on to complain about the block instead.)
func statsIsCommandResult(name string) bool {
	switch name {
	case "SuccessCount", "AffectedBlocks", "AffectedEntities", "AffectedItems", "QueryResult":
		return true
	}
	return false
}

// resolveStatsTarget resolves the target of the entity form.
//
// 1.12.2 is not as talkative here as it is elsewhere: a target it cannot
// resolve is reported as the usage of the entity form, not as "Entity
// '@p' cannot be found" or "Player 'Steve' cannot be found".
//
// (Measured on a real 1.12.2 server: `stats entity @p` says "Usage:
// /stats entity <selector> <mode> ...", and so does `stats entity Steve`.)
func resolveStatsTarget(ctx *exec.Context, arg ast.Arg) ([]*world.Entity, error) {
	switch t := arg.(type) {
	case *ast.Selector:
		resolver := &SelectorResolver{World: ctx.World}
		entities, err := resolver.Resolve(ctx.Source, t)
		selectorWarnings(ctx, resolver)
		if err != nil {
			return nil, exec.Errorf("%v", err)
		}
		if len(entities) == 0 && !selectorIsList(t) {
			// A selector that is limited to a single entity and finds
			// nothing: 1.12.2 reports the usage of the entity form.
			return nil, usageErrorf("stats entity")
		}
		return entities, nil
	case *ast.StringArg:
		// 1.12.2 resolves the target as an **entity**
		// (`CommandStats.c` -> `CommandAbstract.c`), so a name that
		// is not an entity is reported as one that is not found -- and
		// not as the usage of the entity form.
		//
		// (Measured on a real 1.12.2 server, probe case
		// 61-stats-channel: `stats entity Steve` says "Entity 'Steve'
		// cannot be found", even though Steve is a player that is
		// online.)
		if e := entityByUUID(ctx, t.Value); e != nil {
			return []*world.Entity{e}, nil
		}
		return nil, entityNotFoundError(t.Value)
	}
	return nil, usageErrorf("stats entity")
}

// statsCoord parses the position of the block form.
//
// 1.12.2 parses the position only after it has accepted everything else,
// and it takes whole numbers only: a coordinate with a fraction is
// refused with the key `commands.generic.pos.invalid`, which the language
// file of the server does not have, so the raw key is what shows up.
func statsCoord(parts []string, base world.Vec) (world.Vec, error) {
	var out world.Vec
	var vals [3]float64
	var rel [3]bool
	for i, text := range parts {
		if i > 2 {
			break
		}
		r := strings.HasPrefix(text, "~")
		raw := strings.TrimPrefix(text, "~")
		if raw != "" {
			f, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				return out, exec.Errorf("'%s' is not a valid number", text)
			}
			if f != math.Trunc(f) {
				return out, exec.Errorf("commands.generic.pos.invalid")
			}
			vals[i] = f
		}
		rel[i] = r
	}
	out.X, out.Y, out.Z = vals[0], vals[1], vals[2]
	if rel[0] {
		out.X += base.X
	}
	if rel[1] {
		out.Y += base.Y
	}
	if rel[2] {
		out.Z += base.Z
	}
	return out, nil
}

// registerStats adds the `/stats` command to the registry.
func registerStats(reg *exec.Registry) {
	reg.Register(&statsCmd{})
}

// registerStatsReporter adds the reporter that feeds the result of a
// command into the channels of 1.12.2.
func registerStatsReporter(reg *exec.Registry) {
	reg.SetResultReporter(&statsReporter{})
}
