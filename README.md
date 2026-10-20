# gomc

用 Go 实现的 Minecraft **1.12.2** mcfunction 运行时。

它不是"通用 Minecraft 模拟器"，而是一个**确定性的命令解释器**：给一份 1.12.2 存档目录，
它能直接执行其中的 `.mcfunction`，并且**尽量复现真实 1.12.2 服务端的行为**——
包括那些只有把服务端 jar 反编译才看得见的细节。

```
gomc check  <world>                    # 校验世界里的所有函数（含参数 lint）
gomc run    <world> <ns:function>      # 执行一个函数
gomc tick   <world> [ticks]            # 推进 N 个 tick
gomc commands                          # 列出支持的命令
gomc difftest run    <world> <case>    # 打印一个用例的执行观察
gomc difftest check  <world> <case>    # 与 golden 文件比对
gomc difftest update <world> <case>    # 重写 golden 文件
gomc difftest probe  <case>            # 生成给真实服务端用的探针函数

gomc --plain-errors check <world>      # 只输出一行错误（给脚本用）
```

`check` 做两件事：一是查**命令名**未注册（1.12.2 加载函数时就会拒绝的），
二是 lint **参数**（1.12.2 要等函数真跑了才会暴露的问题，比如
`scoreboard players add s obj notanumber`）。第二项是 gomc 自己加的。

## 错误输出

报错**默认是增强的**：指出文件、行、列、那一行原文、正确的用法，以及可能的建议。

```
error[Unknown]: data/functions/demo/f.mcfunction:2:1
  setblok 1 2 3 stone
  ^
  Unknown or invalid command 'setblok'
  hint: did you mean 'setblock'?

error[Syntax]: data/functions/demo/f.mcfunction:1:31
  scoreboard players set @a obj = 5
                                ^
  expected an integer, got "="
  usage: /scoreboard players set <player> <objective> <score> [dataTag]
```

关键在于**消息本身一个字没改**：那仍然是 1.12.2 会说的话（真服也这么说），
位置、用法、建议都是 gomc **附加**的。增强只发生在显示层，不进对拍 golden。

不想要这些附加信息时用 `--plain-errors`，只留一行：

```
error[Syntax]: 1:31: expected an integer, got "="
```

上面是**解析错误**的格式（`line:col`）。**函数加载错误**的行格式不一样——
它带路径、没有列号：`error[Unknown]: <路径>:<行>: <消息>`。脚本不能按一种格式
吃全部错误。

开关放在子命令前后的任何位置都行。

## 文档

| 文档 | 内容 |
|---|---|
| [`docs/SPEC-1.12.2.md`](docs/SPEC-1.12.2.md) | **行为基准**：每条语义都标注了是"读"（反编译）、"验"（真服实测）还是"推"出来的 |
| [`docs/COMMANDS.md`](docs/COMMANDS.md) | 命令支持矩阵：1.12.2 专用服务器 64 条（`CommandDispatcher` 注册 65 次，三个分支互斥），已实现 51 条（外加 1 条 gomc 扩展命令），含 1.12.2 真实语法与已知偏差 |
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | 架构、依赖方向、31 条核心不变量、权威规格的来源与实测流程 |
| [`docs/DIFFTEST.md`](docs/DIFFTEST.md) | 对拍层使用手册：写用例、跑 golden、和真服对齐 |
| [`docs/PERFORMANCE.md`](docs/PERFORMANCE.md) | 性能评估：测了什么、瓶颈在哪、该怎么改 |
| [`docs/DEV.md`](docs/DEV.md) | 开发指南：目录结构、怎么加一条命令、测试纪律、怎么确认 1.12.2 行为 |

## 为什么不是"模拟器"

一个自称模拟器的项目很容易变成命令数量的竞赛。gomc 的验收标准只有一个：

> **对着真实 1.12.2 服务端复现行为。**

所以每一条语义都来自权威来源，而不是来自记忆或 Wiki 猜测。

### 规格是怎么来的

两条腿走路：**读出来**（反编译服务端 jar）+ **验出来**（把真实服务端跑起来）。

#### 读：从服务端 jar 里挖

本机装着一套 Paper 1.12.2 服务端，里面有一个真正的服务端 jar：

```
D:\game\McServer\1.12.2\cache\patched_1.12.2.jar
```

它含有 2347 个 `v1_12_R1` NMS 类。用 CFR 反编译出 178 个类，选择器的每个过滤器
都藏在匿名内部类里（`PlayerSelector$2` 距离、`$3`/`$4` 旋转、`$5` 体积、`$10` 游戏模式、
`$12` 分数……），CFR 还原不出来，就用 `javap -p -c` 直接读字节码。

#### 验：把真实服务端跑起来

反编译给出的是"读出来的"规格，跑一遍真实服务端给出的是"验出来的"规格，
**两者冲突时以实测为准**。做法是复制一份隔离的服务端目录（`libraries`/`cache`
用 junction 指回原目录，不动你的存档），用 JDK 1.8 启动 `nogui`，
把命令喂给控制台并读日志。

这一步直接推翻了好几条从字节码推断出来的结论：

| 结论 | 读出来的 | 实测 |
|---|---|---|
| **`/tag` 命令** | 我以为 1.12.2 有，还实现了 | **没有**，服务端报 `Unknown or invalid command 'tag'`，整个函数加载失败 |
| 实体标签 | `/tag <实体> add` | `scoreboard players tag <实体> <add\|remove\|list> [标签]` |
| `execute` 输出前缀 | `[Server]` | 是**目标实体的名字**（`[Armor Stand] ...`） |
| `say @a` | 只是字面量 | **会解析选择器**，无匹配时报 `Selector '@a' found nothing` |
| 函数内输出 | 统一 | `say` 是 `[Server] ...`，命令结果是 `[Server: ...]` |
| `/time set day` | 24000 | **1000**（`daytime`/`noon`/`midnight` 是 1.13 才有的） |

#### 两条腿挖出来的语义

这套流程挖出并修正了**十几个会让人写出"看起来对但服务器上会错"的函数**的语义：

| 特性 | 常见误解 | 1.12.2 的真实行为 |
|---|---|---|
| 分数过滤 | `@e[scores={kills=1}]` | `@e[score_kills=1]`，且是**上限**；下限写作 `score_kills_min` |
| NBT 过滤 | `@e[nbt={...}]` | 1.12.2 **没有**这个语法（1.13 才加），写了真服直接报错 |
| 旋转过滤 | `ry=90` 表示"正好 90 度" | `ry` 是**上限**、`rym` 是**下限**；下界大于上界时区间跨北环绕 |
| 游戏模式 | `@e[m=!creative]` 取反 | 取反有效，但 `m=` **只对玩家生效**，非玩家一律不匹配 |
| 随机选择 | `@r` 从所有实体里随机 | `@r` 默认**只选玩家**；带 `type=` 时才由类型过滤取代 |
| 体积过滤 | 位置点落在体积内 | 实体**包围盒**与体积相交即可；体积远端多出 1 格 |
| 计数 | `c=3` 取前三个 | 负数会**先反转整个列表**再截断 |
| 重复参数 | 后者覆盖前者 | 正确；但 1.12.2 允许**最后一个**覆盖，语义是"等价于最后一个" |
| 分数溢出 | 64 位无限增长 | Java `int`，**32 位回绕** |

完整的规格清单见 [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)。

## 快速开始

```
go install github.com/Mcd0LUO/gomc/cmd/gomc@latest
```

或者从源码构建：

```powershell
go build -o gomc.exe ./cmd/gomc

# 校验
.\gomc.exe check testdata\demo-world

# 执行一个函数
.\gomc.exe run testdata\demo-world demo:hello

# 推进 tick（每 tick 执行 gameLoopFunction 指向的函数）
.\gomc.exe tick testdata\demo-world 20
```

世界的目录结构就是 1.12.2 的存档结构：

```
<world>/
  data/
    functions/
      <namespace>/
        <name>.mcfunction
```

## 对拍层

`internal/difftest` 是这个项目的验收装置。

一个**用例**描述一段场景：

```
# testdata/cases/scoreboard.difftest
name: scoreboard basics

setup:
  scoreboard objectives add counter dummy
  scoreboard players set Steve counter 1

run:
  scoreboard players add Steve counter 2
  scoreboard players test Steve counter 3
```

跑一遍会得到一份**观察记录**——每条命令的返回码、错误文本、feedback，
外加最终的 scoreboard、实体和方块快照：

```
$ gomc difftest run testdata/demo-world testdata/cases/scoreboard.difftest
# case: scoreboard basics
step 1: scoreboard objectives add counter dummy
  count: 1
  feedback: Added new objective 'counter' successfully
step 2: scoreboard players set Steve counter 1
  count: 1
  feedback: Set score of counter for player Steve to 1
step 3: scoreboard players add Steve counter 2
  count: 1
  feedback: Set score of counter for player Steve to 3
step 4: scoreboard players set Steve counter 10
  count: 1
  feedback: Set score of counter for player Steve to 10
step 5: scoreboard players operation Steve counter += Steve counter
  count: 1
  feedback: Operation applied successfully
step 6: scoreboard players test Steve counter 20
  count: 1
  feedback: Score 20 is in range 20 to 2147483647
step 7: scoreboard players list Steve
  count: 1
  feedback: Showing 1 tracked objective(s) for Steve:
  feedback: - counter: 20 (counter)
```

回显文字是 1.12.2 的**原文**（语言文件里逐字抄的），比如
`Added new objective 'counter' successfully`、`Score 20 is in range 20 to 2147483647`
（只给一个值时上限是 `Integer.MAX_VALUE`），`players list` 的每行是
`- <显示名>: <分数> (<目标名>)`。

观察记录分两个 scope：

- **server scope**（默认）只包含**真实服务端也会打印的东西**：命令的 feedback 与错误文本。
  这部分可以直接和真服的日志对比。
- **full scope** 额外包含世界状态。1.12.2 服务端没法用命令导出 scoreboard / 实体 / 方块，
  所以这一层是用来防 gomc 自己回归的。

要把用例搬到真服上跑，用 `probe` 生成探针函数：

```
$ gomc difftest probe testdata/cases/scoreboard.difftest
# Probe for the differential case "scoreboard basics".
# Run this on a 1.12.2 server and compare the log with
# `gomc difftest run <world> <case>`.
say ## setup: scoreboard objectives add counter dummy
scoreboard objectives add counter dummy
say ## setup: scoreboard players set Steve counter 1
scoreboard players set Steve counter 1
say ## step 1: scoreboard players add Steve counter 2
scoreboard players add Steve counter 2
...
```

把它丢进真服的 `data/functions/` 里跑一遍，再把服务端日志和
`gomc difftest run` 的输出对齐，就完成了一次真正的对拍。

注意两边的编号**不是同一个 N**：观察记录里的 `step N` 把 `setup:` 的命令也数进去
（所以 setup 的两条是 step 1、step 2），而探针里 setup 段的标记是 `## setup:`
（不带编号），只有 `run` 段才是 `## step 1..N`。对齐要按**命令原文**对，
不能按 N 对。

golden 文件则用来锁住行为：

```
$ gomc difftest update testdata/demo-world testdata/cases/scoreboard.difftest
$ gomc difftest check  testdata/demo-world testdata/cases/scoreboard.difftest
ok: testdata/cases/scoreboard.difftest matches scoreboard.difftest.golden
```

## 架构

依赖严格单向，反向 import 一律禁止：

```
              cmd/gomc  ->  internal/cli
                                 |
                                 v
                    internal/version/v1_12_2      1.12.2 方言
                                 |
                                 v
                          internal/exec            执行引擎
                          |         |
                          v         v
                   internal/world  internal/{ast,lexer,parser}
                          |
                          v
                       internal/nbt
```

- `nbt` / `world` 不知道 `exec` 的存在。
- `exec` 不知道任何具体版本，版本差异通过 `Registry` 注入。
- 版本特有的命令实现在 `internal/version/v1_12_2` 下**按域拆文件**
  （`cmd_scoreboard.go`、`cmd_entity.go`、`cmd_world.go`……），不塞进上帝文件。
- 所有命令在**解析期**就产出结构化参数，执行期不再二次解析文本。

详见 [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)。

## tick

1.12.2 没有 1.13 的 `#minecraft:tick` 标签，它的 tick 钩子是 `gameLoopFunction` gamerule：

```
scoreboard objectives add tick dummy
gamerule gameLoopFunction mynamespace:myfunction
```

`gomc tick <world> 20` 会推进 20 个 tick，每个 tick 执行一次该函数。

## 测试

**测试一律带硬超时**，避免把机器拖死：

```powershell
go test ./internal/... -timeout 20s -p 2
```

lexer 里有一个专门防死循环的回归测试（`TestTokenizeOffsetsAdvance` /
`TestTokenizeTerminates`）——早期版本里一个变量遮蔽导致游标不前进，
整个进程被卡死；现在这种情况会直接失败而不是挂起。

## 已知边界

- 只做 1.12.2。1.13+ 的 `execute as/at/run`、`scores={...}`、datapack 布局都不支持，
  因为它们和 1.12.2 的语义是**冲突**的，不是加个开关就能兼容的。
- 不模拟实体 AI、寻路、重力、区块生成、渲染和网络协议。
- 1.12.2 专用服务器的命令表是 **64 条**（`CommandDispatcher` 里注册了 65 次，但
  专用服分支 15 条、客户端分支 1 条是互斥的），当前实现了 **51 条**，
  优先覆盖写函数时最常用的那些。
  `docs/COMMANDS.md` 记录了完整矩阵（`CommandDispatcher.java`）。
- 另有一条 **gomc 自己的扩展命令** `/gomcplayer`：1.12.2 造不出玩家（`/summon
  player` 报 `Unable to summon object`），而 `/advancement` 和 `/recipe` 都要在线
  玩家。它被标记成扩展，`gomc check` 会说清用了它的函数在真服上加载就会失败。
- 服务端管理类命令（`op`/`ban`/`whitelist`/`save-all` 等）没有实现，它们对测试
  mcfunction 没有意义。
- 1.12.2 **没有独立的 `/tag` 命令**（这一点是跑真实服务端验证出来的），
  实体标签用 `scoreboard players tag <target> <add|remove|list> [tag]` 管理。

## 许可证

[MIT](LICENSE)。本仓库只包含本项目自己的代码；反编译出来的 Mojang 代码和服务端
jar 都在 `.ref/` 下，**不入库**（见 `.gitignore`）。
