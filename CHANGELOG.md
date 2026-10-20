# 更新日志

本文件记录 gomc 的重要变化。格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循[语义化版本](https://semver.org/lang/zh-CN/)。

0.x.y 表示这个版本的行为**还在对齐中**：0.1.0 已经有完整的命令面，
但细节仍可能与真实服务端有出入（已知偏差见 `docs/SPEC-1.12.2.md` 的"还没对齐的"一节）。
1.0.0 会等到 1.12.2 的行为被完整复现。

## [未发布]

### 修复

一批按 1.12.2 真实行为对齐的修正（都来自反编译源码 + 真服实测）：

- **选择器**：`@a`/`@p` 忽略 `type` 参数、体积盒差一格、`r`/`rm` 是整数解析、
  负半径取消该约束、`score_x` 与 `score_x_min` 可以并存、计分板持有者用玩家名
  （选择器与命令写侧现在一致）。
- **execute**：detect 的偏移只加一次（之前加了两次）、detect 的数据值必填。
- **命令与执行模型**：`/testforblocks` 改成区域比较（之前实现成了另一个命令）、
  `/fill` 加 32768 上限与 Y 越界检查、`/clone` 加重叠检查、函数嵌套改成入队展开
  （调用者拿 0）、`maxCommandChainLength` 改成整条链共享的预算、函数内命令权限
  降到 2 级。
- **回显文案**：`players list` 的条目格式、add/remove 的回显、reset 的两种形式、
  `scoreboard teams` 整个子树的回显、`/give` 的回显、`worldborder` 的 grow/shrink
  与 center 的 0.5 偏移、`spawnpoint` 用目标玩家名、`gamerule` 无参数时列出全部规则、
  `difficulty` 用翻译名、目标解析失败区分 `Selector` / `Entity` / `Player` 三种文案。
- **语法**：`/effect`（目标在最前）、`/trigger`（模式必填）、`/particle`（速度必填，
  之前**根本解析不了**）、`/fill`（参数位置改成 1.12.2 的）。
- **其它**：只读目标（health/xp/level/…）不能写、`><` 交换要换源、
  `entitydata` 改成深合并、超 int32 的数字报错而不是静默截断、
  `gamerule` 设置不存在的规则按原版新建。
- **选择器四处**：`c` 传非整数时真服是**静默回落**（`kill @e[c=abc]` 真服照常执行，
  gomc 之前报错）；`name=` 对**被改名的玩家**比较的是用户名而非自定义名
  （`EntityHuman.getName()` 覆写了父类）；`type=` 的键必须是 1.12.2 的注册名
  （`type=pigzombie` 不认，1.12.2 已改名 `zombie_pigman`）；`m=!<未知模式>`
  命中全部玩家。实体类型表也补齐到 85 条，与真服注册名逐字对应。
- **空匹配在命令层**（真服实测，见 SPEC 3.10）：**列表**选择器匹配不到东西时
  **不是错误**——命令跑 0 次、什么都不说、命令成功。gomc 之前一律变成
  `Selector '...' found nothing`。**单实体目标**（`@p`、`c=1`、裸名字）仍然
  报错，而且措辞跟着命令的目标类型走：解析成实体的命令说 `Entity`，
  解析成玩家的说 `Player`（`gamemode creative @p` 现在说
  `Player '@p' cannot be found`）。
- **消息展开的连接符**：真服 2 个用 `" and "`、3 个用 `", … and "`
  （`say @e[...]` 三个实体是 `A, B and C`），gomc 之前一律用 `", "`。
- **`/summon` 的类型集**：从 34 种可用补到 **79 种**，与真服能 summon 的种数
  一样；类型名按 1.12.2 的 `MinecraftKey` 规则解析（`minecraft:zombie`、
  `:zombie`、`a:zombie`、`ZOMBIE` 都认，`foo:zombie` 不认）。失败文案从
  `unknown entity type "x"` 改成真服的 `Unable to summon object`。
  活体实体的初始血量按真服 `/entitydata` 实测值写入（史莱姆、 magma cube、
  马 family 是随机 roll，没写死）。
- **`scoreboard players tag list` 的标题**用计分板持有者名（玩家是用户名），
  不是实体 id；**没有 tag 的实体不打标题**（真服 `CommandScoreboard.java:735`），
  命令报的计数是 tag 的个数而不是实体数。tag 列表的连接符也跟真服一样。
- **`scoreboard players tag` 能按玩家名用**：真服 `scoreboard players tag Steve list`
  是有效的，gomc 之前什么都不打。
- **探针工具的假失败**：`tools/probe/probe.ps1` 之前把 `spawn-animals`、
  `spawn-npcs` 设成 false、难度设成 peaceful。CraftBukkit 的
  `World.addEntity` 会因此让**动物、怪物和村民在真服上 summon 不了**，
  测出来是大面积假失败。现在这三项都打开。


### 性能

- `scoreboard players list` 从 O(持有者 × 目标) 降到近 O(持有者 + 目标)：
  1000 个持有者、8 个目标时从 87.6 ms 降到 22.9 ms（`bfe7dcc`）。
- 完整的性能评估（测了什么、瓶颈在哪、哪些是伪瓶颈）见
  [`docs/PERFORMANCE.md`](docs/PERFORMANCE.md)。

### 计划中

- 方块注册表：让写错的方块名像真服一样报错。
- 实体 UUID：真服用 UUID 标识非玩家实体的计分板持有者（见 SPEC 第 12 节）。

## [0.1.0] - 2026-10-20

首个可用版本。

### 新增

- **1.12.2 命令解释器**：实现 1.12.2 专用服务器那 64 条命令中的 46 条，覆盖写 mcfunction
  最常用的那些（`execute`、`scoreboard`、`summon`、`setblock`、`fill`、`clone`、
  `function`、`gamerule`、`time`、`weather`……）。
- **目标选择器**：`@p` / `@r` / `@a` / `@e` / `@s` 与 1.12.2 的全部过滤器
  （距离、旋转、体积、分数、游戏模式、标签、队伍、名称、类型）。
- **对拍层**（`internal/difftest`）：用 YAML 风格用例描述一段场景，把执行结果与
  golden 文件比对，锁住行为；`difftest probe` 还能生成一份能丢到真实服务端上跑的
  探针函数。
- **真服探针工具**（`tools/probe`）：在隔离的 1.12.2 服务端上跑用例、抓真实输出，
  用来验证规格。隔离环境用 junction 复用依赖，用户自己的存档与进程不受影响。
- **增强报错**（默认开启）：报出文件、行、列、那一行原文、插入符、正确用法和
  `did you mean` 建议；`--plain-errors` 可关掉增强只留一行。
  错误**原文**始终是 1.12.2 的原文，增强只是附加。
- **语义化版本号**：`gomc --version` 打印版本、提交号与构建时间。

### 行为对齐

真服实测（不是查 Wiki）定下来的几条关键行为：

- `execute` 对**每个**匹配实体各执行一次，顺序是**距离升序**（不是倒序）。
- `execute` 里一个实体失败**不影响**其余实体。
- 实体的显示名是**翻译名**（`Armor Stand`），不是注册名。
- 1.12.2 **没有**独立的 `/tag` 命令，标签走 `scoreboard players tag`。
- 标签就是实体 NBT 里的 `Tags`，写 NBT 和用计分板命令是同一份数据。
- `scoreboard players test` 只给一个值时，上限是 `2147483647`（不是等于下限）。
- 函数 id 按 `MinecraftKey` 小写化，缺命名空间默认 `minecraft:`。

### 已知边界

- 只做 1.12.2；1.13+ 的 `execute as/at/run`、`scores={...}`、datapack 布局都不支持。
- 不模拟实体 AI、寻路、重力、区块生成、渲染和网络协议。
- 方块名不校验（写错方块名不会报错）。
- 实体没有建模 UUID：非玩家实体的计分板持有者用内部 id，真服用 UUID。
  UUID 不可读，实际用选择器（`@e[type=zombie]`）指定就行，那条路是通的。

[未发布]: https://github.com/Mcd0LUO/gomc/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Mcd0LUO/gomc/releases/tag/v0.1.0
