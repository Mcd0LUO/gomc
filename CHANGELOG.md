# 更新日志

本文件记录 gomc 的重要变化。格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循[语义化版本](https://semver.org/lang/zh-CN/)。

0.x.y 表示这个版本的行为**还在对齐中**：0.1.0 已经有完整的命令面，
但细节仍可能与真实服务端有出入（已知偏差见 `docs/SPEC-1.12.2.md` 的"还没对齐的"一节）。
1.0.0 会等到 1.12.2 的行为被完整复现。

## [未发布]

### 新增

- **`/gomcplayer`——把玩家带进世界的 gomc 扩展命令**（1.12.2 没有这条命令）：
  ```
  /gomcplayer <名字> [x] [y] [z]
  ```
  1.12.2 **造不出玩家**（`/summon player` 报 `Unable to summon object`），所以一个
  需要玩家才能测的函数在真服上根本没法测。而 `/advancement` 和 `/recipe` 都要
  在线玩家——没有这条命令，它们能测的就只有错误消息了。
  造出来的玩家是**世界里的普通玩家**：按名字能找到、`@a`/`@p` 能选中它、
  计分板用它的**玩家名**当持有者（不是 UUID）、`/advancement` 和 `/recipe`
  的成功路径都能对它跑。
  它被标记成**扩展命令**（`exec.Registry.SetExtension`），所以 `gomc check`
  会明确说：这个函数在真服上**加载就会失败**（1.12.2 拒绝含未知命令的函数），
  而不是让这个差异悄悄过去。

- **四条命令**（命令数 46 → 50）：`/advancement`、`/recipe`、`/spreadplayers`、`/stats`。
  每一条都对着真实 1.12.2 服务端逐条实测过，规格记在
  [`docs/SPEC-1.12.2.md`](docs/SPEC-1.12.2.md) 第 13 节。
  - **`/advancement`**：进度注册表取自服务端 jar 里的
    `assets/minecraft/advancements/**`，**488 条**（`minecraft:<路径>`）。每个进度记下
    父进度、条件名和条件的组合方式：438 个是"任一条件完成即可"，50 个是"全部条件
    都要完成"（`adventure/kill_a_mob` 杀一种怪就行，`adventure/kill_all_mobs` 要杀
    22 种）。五种过滤器（`only`/`until`/`from`/`through`/`everything`）都实现了。
    1.12.2 的用法行有**两种**形态：顶层和 `test` 带 `Usage: ` 前缀，各个子过滤器是
    **裸行**，gomc 照做。
  - **`/recipe`**：**432 个**数据驱动配方（jar 里的 `assets/minecraft/recipes/*.json`）
    加 **11 个**代码里建的配方（那 11 个在服务端看来都是"特殊"的，给不出去）。回显
    里用的是配方的**结果物品**的显示名（`minecraft:stone_slab` → "Stone Slab"）。
    用法行写的是 `[player]`，但玩家其实**不能省**——服务端总是把第二个参数当玩家，
    所以 `/recipe give <player>` 会读越界，报一句普通的话（真服日志里是
    `ArrayIndexOutOfBoundsException`）。那条也照抄了。
  - **`/spreadplayers`**：算法照搬 1.12.2 的**松弛迭代**（每轮把互相靠太近的
    玩家推开，收敛或到迭代上限为止）。随机数用**自实现的 splitmix64**（不依赖
    标准库，免得 Go 的版本漂移改结果），种子从命令参数派生，所以同一条命令永远得到
    同一个位置。"能站"的规则是第一个非空气方块，**排除**液体和火。
  - **`/stats`**：**1.12.2 上这条命令是坏的**，而且坏法很特别——它把**统计项的名字**
    当成了"命令结果类型"去映射，映射不到就抛 `Invalid parameters`。gomc 复现的就是
    这个坏法：entity 形式先看目标（列表选择器匹配 0 个就**什么都不说**，单个选择器
    解析不出就报 entity 的用法），block 形式先数参数，参数齐了才看统计项，而且
    **方块一律存不了统计**（箱子、漏斗、熔炉、信标都是同一句 `can not track stats`）。
- **计分板持有者的 dataTag 过滤**：`scoreboard players set|add|remove <持有者> <目标>
  <分数> [dataTag]`。dataTag 要和实体的 NBT 对得上才生效，对不上报
  `The dataTag does not match for <持有者>`；解析失败报
  `Could not parse dataTag, reason: <Mojangson 的原话>`。**add 和 remove 也校验**
  ——语言文件里只有 set 的键，但 1.12.2 三个动作共用一个 handler。`scoreboard players
  tag` 也吃 dataTag（它的报错措辞不同：`Players tag command failed, reason: …`）。
  解析出来的 dataTag 放在**运行期**才解析，这样才能复现真服"先查目标、再查目标、
  最后才看 dataTag"的顺序。
- **实体 UUID**：非玩家实体的计分板持有者现在用 UUID（v4 形状），`scoreboard players
  set <uuid> …` 这类写法也认得出。UUID 从实体的内部 id **确定性派生**（`md5` + v4 的
  版本位和 variant 位）——真服是 `UUID.randomUUID()` 随机生成的，而 gomc 是确定性的
  命令解释器，所以值不可能一样，格式和稳定性才是对齐的。
- **`scoreboard players tag` 的持有者总是按实体解析**：解析不出就报
  `Entity '…' cannot be found`（`add`/`remove`/`list` 三个分支都一样），而不是像
  之前那样静默跳过。

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
- **`/summon item` 需要一个非空、且是已注册物品的物品堆**：真服的实体工厂只接受
  带着非空物品堆的物品实体，空堆的实体建出来就死，所以 summon 直接失败。决定
  成败的是 `Count`——1.12.2 用 `getByte` 读它，**任何数字类型都截断成有符号
  字节**（`Count:1` 算数、`Count:256` 截成 0 不算、`Count:-1` 是负数不算，
  缺省就是 0）；`Slot` 无关。物品名也必须是已注册的物品。gomc 之前一律放行。
- **物品注册表**（新，`internal/world/itemnames.go`）：从服务端 jar 的字节码里
  挖出 1.12.2 的 **409 个物品名**（`Item` 类注册纯物品、`Block` 类注册方块，
  每个方块自带一个同名物品），顺带得到每个物品实体的**显示名**——真服是按
  物品命名实体的，而服务端语言文件里没有那个键，所以真服直接打印键本身：
  钻石剑是 `item.item.swordDiamond`、石头是 `item.tile.stone.stone`、铁砧是
  `item.tile.anvil.intact`。gomc 之前只显示实体类型的名字（`Item`）。
- **方块注册表**（新，`internal/world/blockregistry.go`）：1.12.2 的 **254 个
  方块名**（`Block` 类注册旧命名、`Blocks` 类注册 1.12 命名，两套都认）。
  名字写错现在会报真服那句 `There is no such block with name minecraft:<名字>`
  （命名空间和大小写按 `MinecraftKey` 规范化）。变种（`granite`、`white_wool`）
  是数据值不是方块名，仍会报错。校验放在解析期——和真服一样，函数照常加载，
  错误在运行到那条命令时才出现。
- **方块的显示名**也一并对齐：1.12.2 显示的是**注册名**对应的名字，而 1.12
  **复用了 1.11 的几个名字**（`stonebrick` 在 1.11 是圆石、在 1.12.2 是石砖），
  所以查表时新表优先、旧表其次。254 个方块的显示名是用真服逐个量出来的
  （`setblock` 之后用 `testforblock` 读回来）。服务端的语言文件是精简版，所以
  有些方块的名字在真服上就是**未翻译的键**（`tile.bed.name`）——那也是真服的
  真实输出。改之前 254 个方块里有 208 个的名字和真服对不上。
- **`scoreboard players tag list` 的标题**用计分板持有者名（玩家是用户名），
  不是实体 id；**没有 tag 的实体不打标题**（真服 `CommandScoreboard.java:735`），
  命令报的计数是 tag 的个数而不是实体数。tag 列表的连接符也跟真服一样。
- **`scoreboard players tag` 能按玩家名用**：真服 `scoreboard players tag Steve list`
  是有效的，gomc 之前什么都不打。
- **探针工具的假失败**：`tools/probe/probe.ps1` 之前把 `spawn-animals`、
  `spawn-npcs` 设成 false、难度设成 peaceful。CraftBukkit 的
  `World.addEntity` 会因此让**动物、怪物和村民在真服上 summon 不了**，
  测出来是大面积假失败。现在这三项都打开。
- **NBT 匹配要求类型一致**：1.12.2 用 `NBTBase.equals` 比较 NBT，**类型是值的一部分**，
  所以存着的 `20.0f` 匹配不上过滤器里的 `20`。gomc 之前是跨宽度比较，比真服宽松。
  真服实测：`scoreboard players set <实体> t 7 {Health:20.0f}` 成功，
  `{Health:20}` 报 `The dataTag does not match`。
- **用法消息不再被打印两遍**：1.12.2 回答用法错误时给的就是用法行本身，所以
  错误报告里不该再补一遍 `usage:`。命令的用法行仍然照旧附在描述性错误后面。


### 性能

- `scoreboard players list` 从 O(持有者 × 目标) 降到近 O(持有者 + 目标)：
  1000 个持有者、8 个目标时从 87.6 ms 降到 22.9 ms（`bfe7dcc`）。
- **选择器解析快了 2.5~4.5 倍**（`d946a4d`）：世界缓存实体列表，选择器不再
  每次复制整张实体表；结果切片不按候选数预分配；顺带消掉了每个实体的
  类型归一化（它占 Resolve 的 38%）。8000 个实体、只命中 1 个时：
  258 783 → **56 872 ns**，分配 131 097 B → **56 B**。
- **只要一个实体时不用把整个列表排一遍**（`a301a60`）：`@e[c=1]` 这类选择器
  真服是**排完再取第一个**（`Collections.sort` 之后才 `subList`），而排序是
  稳定的，所以"最近的那个"就是列表里第一个不比其它远的实体——找它不用排序。
  这是**等价但更快**的做法，同样一批实体、同一个进程里测：100 个实体
  1 075 → **245 ns**，1000 个 13 283 → **2 492 ns**，8000 个
  161 249 → **19 858 ns**（**4.4~8.1 倍**），而且零分配。
- **给 `Entity` 加字段不是免费的**：加 `UUID string` 让结构体从 136 字节涨到
  152 字节，**跨过了 Go 的分配档位**，选择器遍历整张实体表因此慢了 22%。
  `UUID` 现在是指针（8 字节）且挪到结构体末尾，实体回到 144 字节那一档。
- 完整的性能评估（测了什么、瓶颈在哪、哪些是伪瓶颈）见
  [`docs/PERFORMANCE.md`](docs/PERFORMANCE.md)。

### 计划中

- **数据包的进度与配方**：1.12.2 的自定义进度在 `<世界>/advancements/<命名空间>/<路径>.json`，
  自定义配方在 `<世界>/recipes/`。gomc 没有 JSON 解析器，所以只认服务端 jar 里自带的那一套。
- **`/locate`**：需要结构注册表。
- **条件的触发机制**：`/advancement` 只能通过命令设置条件，没有触发器（`minecraft:recipe_unlocked`
  之类）。这也是 `grant everything` 的数字和真服差 7 个的原因。
- **词法器不配平 NBT 的花括号**：真服先切参数再解析 NBT，报
  `Could not parse dataTag, reason: …`；gomc 的词法器读命令时配平花括号，不配平就整行报错。
  改它对 `testfor`/`summon`/`entitydata` 一并有益。

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
