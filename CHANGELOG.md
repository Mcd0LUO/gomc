# 更新日志

本文件记录 gomc 的重要变化。格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循[语义化版本](https://semver.org/lang/zh-CN/)。

0.x.y 表示这个版本的行为**还在对齐中**：0.1.0 已经有完整的命令面，
但细节仍可能与真实服务端有出入（已知偏差见 `docs/SPEC-1.12.2.md` 的"还没对齐的"一节）。
1.0.0 会等到 1.12.2 的行为被完整复现。

## [未发布]

### 新增

- **1.12.2 的方块状态表**——`/fill` 和 `/setblock` 的数据值那一格现在认
  **状态字符串**了：
  ```
  /fill 2 4 2 3 4 3 stone variant=granite
  /setblock 8 4 8 wool color=red
  ```
  1.12.2 的方块有属性（`variant`、`color`、`age`、`facing`…），命令可以要
  某个状态而不是数据值。表是从**真服务端 jar 的方块注册表**里跑出来的
  （`tools/blockstates`）：254 个方块，每个方块有哪些属性、每个属性有哪些
  取值、每个取值组合对应哪个数据值，外加每个方块的数据值翻译表。
  解析规则照抄 `CommandAbstract.c(Block, String)`：按 `,` 拆，每段按 `=`
  拆成**最多两段**（所以 `variant=granite=1` 的值是 `granite=1`）、名字和
  取值都**大小写敏感**、`default` 只能整个参数就是它、没写到的属性保持
  默认状态里的值、同一个属性写多次**后面的赢**。
- **`/locate`**：1.12.2 的结构注册表和生物群系注册表都实现了，`/locate` 认
  `biome` 和 `structure` 两类，规格记在
  [`docs/SPEC-1.12.2.md`](docs/SPEC-1.12.2.md) 第 17 节。找不到时真服说
  `Unable to locate any <x> feature`。
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
- **`/advancement` 认数据包里自定义的进度**：1.12.2 把自定义进度放在
  **`<世界>/advancements/<命名空间>/<路径>.json`**（不是 1.13+ 的
  `data/<命名空间>/advancements/`），gomc 现在也读它。命名规则和别的
  `MinecraftKey` 一样：命名空间可省（默认 `minecraft`），两段都小写。
  进度文件的 `requirements` 也按 1.12.2 的形状读：一个组里任一条件成立即可；
  **没有** `requirements` 时每个条件自成一组、全部要成立。
- **`grant everything` 的数字对齐真服**：真服报 **481**，gomc 之前报 488。
  差别**不是**登录时给了玩家一批配方（实测：刚登录时 `recipe give` 说
  "Successfully given"，玩家本来什么配方都没有），而是**命令执行过程中**
  的连锁反应：进度完成 → 发 `rewards.recipes` 里的配方 →
  `minecraft:recipe_unlocked` 触发器把"盯着这个配方"的进度条件点亮 →
  命令走到那个进度时它已经完成 → 跳过、不计数。
  gomc 建模了这条机制，所以数字对上了（`revoke everything` 两边都是 488；
  **revoke 之后再 grant** 也是 488，因为触发器只对玩家还不认识的配方生效）。
  **没有写死"跳过这 7 个"的列表**——模型是"玩家已解锁的配方会完成盯着它的
  进度"，数据来自服务端 jar 的 `rewards.recipes`，遍历顺序也复刻了服务端的
  `java.util.HashMap`（所以跳过的是**同样那 7 个**）。
- **`/recipe` 认 `<世界>/recipes/<名字>.json`**——**这一条是 gomc 的扩展，
  真服没有**：1.12.2 只从自己的 assets 根读配方（jar 或服务端目录的
  `assets/`），三种世界目录布局（`recipes/`、`data/recipes/`、
  `data/minecraft/recipes/`）在真服上实测全是 `is an unknown recipe`。
  gomc 读世界目录，是给写函数的人一个能用的地方，差别记在 SPEC 里。

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

- **随机源每个引擎重置一次**。`@r` 选择器和 `/weather` 的默认时长共用世界那条
  随机流（真服也是这样 ✓），但**引擎从不重置**它，于是同一个用例在不同的测试
  顺序下会选中不同的实体，golden 变成顺序相关的 flaky 测试。现在每个引擎从
  固定种子重新开始，并补了一条走随机路径的确定性测试。
- **`/give` 的数量有上下界**。真服用
  `CommandAbstract.a(text, 1, item.getMaxStackSize())` 读数量，所以 0 和 65 都被
  拒绝：`The number you have entered (65) is too big, it must be at most 64`。
  gomc 之前把解析错误**丢掉**（`n, _ := parseIntSafe(...)`），于是
  `give Steve stone 99999999999999` 会给出 "Given [Stone] * 0"。
- **`gamerule` 对未知规则再设一次值，现在真的会改**。1.12.2 把未知规则存成
  ANY_VALUE 规则、值是字符串，gomc 的 `GameRules.Set` 漏了 `RuleAny` 那个分支，
  于是第二次设置被静默丢掉，而回显还说"已更新"。
- **`/spawnpoint` 的 Y 不加 0.5**。中心模式只作用于 X/Z，1.12.2 读方块位置的
  Y 时传的是 `false`（`b(base.getY(), text, 0, 256, false)`），所以
  `setworldspawn 0 256 0` 是合法的（真服：Set the world spawn point to
  (0, 256, 0)）。
- **`/give` 的数据值是 0 而不是「没给」**。真服没写数据值时用 0，物品堆的
  `Damage` 只在不为 0 时才写进 NBT。
- **探针的比较多归一化了几处每次都变的内容**：实体 NBT 里的 `Motion`、
  `Rotation`、`FallDistance`（每个刻都在变，读到的是第几个刻取决于命令的快慢）。
  另外 `run-all.ps1` 不再把**没有基线**的用例记成失败。
- **自带的两个基准又能跑了**：它们把方块放在 0..9999，而那落在"世界之外"
  （新加的判定）里，基准于是直接失败。坐标改到加载区之内。
- **命令里的数据值和存下来的数据值不是一回事**。1.12.2 存的是方块的**完整
  状态**，命令里写的却是老的数据值：数字要交给 `fromLegacyData`，放下去的
  那个状态的**数据值**是 `toLegacyData(fromLegacyData(n))`。对多数方块两者
  相同，但不是全部——`setblock x y z chest 0` 放下去的是朝北的箱子，数据值
  是 **2**；`setblock x y z cobblestone 5` 放下去的数据值是 **0**（没有属性
  的方块只有数据值 0）。gomc 之前**原样存用户写的数字**，254 个方块里有
  **214 个**会因此存错。现在按真服的表翻译。
- **数据值被拒绝时的措辞**：不是"数字太大/太小"，而是状态那句。`fromLegacyData`
  抛异常时（小麦只有数据值 0..7）1.12.2 会落到状态串分支，报
  `'9' is not a state for block minecraft:wheat`。
- **省略数据值时放的是方块的默认状态**，而默认状态的数据值**不是**总是 0：
  `chest` 是 2、`torch` 是 5、`oak_stairs` 是 3、`dispenser` 是 2。
  gomc 之前一律放数据值 0。
- **坐标有边界，越界就报错**。1.12.2 读坐标用
  `CommandAbstract.b(base, text, min, max, flag)`，越界报
  `commands.generic.num.tooBig` / `tooSmall`，数字按 `"%.2f"` 印。边界分三种：
  - **方块位置**（`/setblock` `/fill` `/clone` `/testforblock(s)` `/spawnpoint`…）：
    X/Z 是 ±30000000，**Y 是 [0, 256]**；
  - **普通数字**（`/execute` `/particle` `/playsound` `/worldborder`）：三个都是
    ±30000000；
  - **`/tp`**：X/Z 是 ±30000000，**Y 是 ±4096**。

  gomc 之前**一个坐标都不查**，`setblock 1e20 4 0 stone` 会把坐标算成垃圾值。
  现在和真服逐字一致：
  ```
  setblock 30000001 4 0 stone
  → The number you have entered (30000001.00) is too big, it must be at most 30000000
  setblock 0 -1 0 stone
  → The number you have entered (-1.00) is too small, it must be at least 0
  ```
  顺带把**「中心模式」**也搬过来了：`flag` 为真时整数坐标指方块的**中心**，
  要加 0.5（消息里印的就是加过 0.5 的数）。`/tp` `/particle` `/playsound`
  `/spawnpoint` 和 `/summon` 的 X、Z 都是这个模式（`/summon` 的 Y 不是）。
- **`/summon` 把实体放在方块中心**，不是角上。`CommandSummon.b(x, text, true)`
  用的是中心模式，所以 `summon zombie 0 4 0` 落在 **(0.5, 4, 0.5)**（Y 不加
  0.5）。gomc 之前放在 (0, 4, 0)。
- **世界高度之外**：真服的 `/setblock 0 256 0 stone` 说
  `The block couldn't be placed`（放的时候才失败），
  `/fill 0 256 0 1 256 1 stone` 说 `Cannot place blocks outside of the world`
  （动手之前就拒绝整个区域）。gomc 之前只有后者。
- **「世界之外」也搬过来了**。真服除了坐标的数值边界，还有一道
  `World.isLoaded(pos)`：那个位置的**区块**得是加载着的。真服只加载出生点
  所在区块 **±8 个区块**（启动时 "Preparing start region" 那一步）和每个
  玩家周围（视距，默认 10 个区块）。
  gomc 的世界是稀疏表、没有区块，所以**照着真服的样子建模**：出生点区块
  ±8、玩家周围 ±10，之外就报 "outside of the world"（各命令的措辞不一样，
  见 SPEC 6.6）。之前 gomc 坐标多远都放得下。
- **问「方块是不是这一个」的命令走的是谓词，不是数据值**。`/testforblock`
  的第 5 个参数和 `/fill` 的 `replace` 模式里「要替换的方块」后面那个数据值，
  1.12.2 用的是 `CommandBase.b(Block, String)`：数字**不翻译**（直接和数据值
  比），其它是方块的状态串。gomc 之前把 `/testforblock` 的状态串**默默丢掉**了
  ——`testforblock x y z stone variant=bogus` 在 gomc 上说"找到了"，真服报
  `'variant=bogus' is not a state for block minecraft:stone`。
  状态串对不上时，真服的消息里给的是方块的**状态**
  （`minecraft:stone[variant=granite]`），不是数字。
  `/fill` 的替换数据值还是**惰性**的：1.12.2 在循环里读它，而且只对
  「是该方块」的位置读，所以区域里没有匹配方块时，坏的数据值根本不报错。
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
- **玩家专属命令的目标真的只接受玩家**：1.12.2 里 give / gamemode / title /
  tellraw / playsound / stopsound / trigger / xp / clear / spawnpoint 的目标是
  用 `CommandAbstract.b` 解析的，那个变体**只找玩家**；gomc 用的是通用解析，
  任何实体都收。实测（2026-10-21）：
  - 目标匹配到**非玩家**时，报的是那个实体的 **UUID**
    （`give @e[type=armor_stand] stone` → `Player '8de4ef1b-...' cannot be found`）
  - 目标匹配 **0 个**时**什么都不说**（列表选择器展开成零次执行）
  - 名字找不到玩家 → `Player '<名字>' cannot be found`

  顺带修掉一个由此暴露的老偏差：gomc 之前能用**内部 id** 当玩家目标
  （`title p1 title ...` 能跑），真服只认玩家名。
  `effect` / `enchant` / `kill` / `testfor` / `replaceitem` / `particle` 的
  目标**本来就是**实体变体（逐条实测确认），一个字没动。
- **坏 dataTag 的措辞对齐真服**：词法器不再因为 NBT 的花括号不配平而拒绝
  整条命令，坏 dataTag 交给 NBT 解析器报错，措辞是
  `Data tag parsing failed: <reason>`（scoreboard 那边是
  `Could not parse dataTag, reason: <reason>`）。原因用 MojangsonParser 的
  原话，只显示**最后 35 个字符**（长了前面加 `...`）。dataTag 的解析推到
  运行期，所以**目标先于 dataTag** 被报出来——真服上
  `testfor @e[type=armor_stand {Tags:` 先报实体找不到。


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

### 修复

- **`/fill` 的数据值位置严格了**：方块之后**只要还有参数**，那一个就一定是数据值
  ——读不成数据值就报错，**不会**被当成模式。1.12.2 的
  `CommandBase.a(Block, String)` 有三种写法：十进制数字（0..15）、`default`
  （默认状态）、以及**方块状态字符串**。真服实测（2026-10-21）：
  `fill 2 4 2 3 4 3 stone {Tags:` 说
  `'{Tags:' is not a state for block minecraft:stone`，
  `fill … stone -1` 说 `The number you have entered (-1) is too small, it must be
  at least 0`，`stone 16` 说 `… too big, it must be at most 15`。
  > 这句 "is not a state for block" 是**原版 1.12.2** 的措辞，不是 Paper 加的：
  > vanilla 的 `en_US.lang` 里有 `commands.generic.blockstate.invalid`，vanilla 的
  > `CommandBase`（混淆名 `bi`）里也握着这个字符串。
- **`/setblock` 的数据值位置也一样严格了**（它有**同一个**缺陷）：第 5 个参数**一定**
  是数据值，所以**模式必须写在数据值之后**——`setblock x y z dirt 0 keep` 才是
  "放泥土、保留模式"，而 `setblock x y z dirt keep` 在真服上是**错的**（`keep` 落在
  数据值那一格，报 `'keep' is not a state for block minecraft:dirt`）。
  gomc 之前把第 5 个参数"是数字就当数据值、否则当模式"，所以那种写法在 gomc 上
  能跑、在真服上跑不了。
- **进度会存盘**：1.12.2 把玩家的进度放在 **`<世界>/advancements/<UUID>.json`**
  （**不是** `playerdata/<uuid>.advancements.json`，那是 1.13+ 的），gomc 现在也写
  ——**这是 gomc 第一次写用户的存档目录**。文件名里的 UUID 是 gomc **派生的**
  （`world.DeriveEntityUUID`，玩家用玩家名派生），所以**不会覆盖用户真服存档里的
  那一个**（真 UUID 是随机的，名字不同）。文件格式与真服一致（条件名 → 完成时间），
  但**时间是固定的**——gomc 是确定性的命令解释器，同一个世界跑两次必须得到一模一样的
  文件。读回的时候只看"条件在不在"，不看时间。
- **`/advancement` 的奖励会发经验**：1.12.2 的进度奖励有两种，`recipes`（已建模）
  和 `experience`（这次补的，12 个进度有）。
- **`/recipe` 的自定义配方支持子目录**：`<世界>/recipes/sub/x.json` 是配方
  `minecraft:sub/x`（真服的 assets 布局就是这么定的）。
- **坏的自定义配方文件会中止后面的加载**：真服解析一个配方文件失败时**后面的配方
  全都不认识**（不是跳过那一个），gomc 现在照做。
- **物品的显示名跟着数据值走**：染料、旗帜、**床、头骷髅、毛**这些物品的显示名
  跟着数据值变（`minecraft:dye` 的数据值 3 是 "Cocoa Beans"），不只是染料和旗帜。
  数据值没给、或者超出 0..15，都用数据值 0 的名字。
- **`/give` 的回显改用显示名**：真服打的是 `Given [Stone] * 1 to Steve`——
  **显示名外面套方括号**，命名空间被换掉，而且**显示名跟着数据值走**
  （`minecraft:dye` 的数据值 3 是 "Cocoa Beans"）。gomc 之前打的是**注册名**
  （`Given stone * 1 to Steve`）。
  跟着数据值变的不只是染料和旗帜：**床、头骷髅、毛**也算。
  数据值没给、或者超出 0..15，都用数据值 0 的名字。
- **不配平的方括号和圆括号是普通文字**了。1.12.2 根本不做分组，它按空格切参数——
  所以 `say @e[type=armor_stand` 打印的是原文，`scoreboard players set
  @e[type=armor_stand t 1` 把持有者当成名字 `@e[type=armor_stand`（选择器的正则要求
  括号配平，不配平就不是选择器），`say (hello` 打印 `(hello`。
  gomc 之前在词法阶段就报 `unbalanced [`。
- **坐标的报错措辞对齐真服**了：不是数字时真服说 `'a' is not a valid number`
  （`commands.generic.num.invalid`），gomc 之前说 `expected a coordinate, got "a"`。
  而**小数坐标真服是接受的**（`0.5` 不报错，它取整数部分），gomc 也一样。
- **探针的协议机器人修好了一个 bug**：`tools/probe/mcbot.py` 原来把 clientbound 的
  `0x00` 当成断线，但 1.12.2 的 `0x00` 是**生成实体**，真正的断线是 `0x1A`、
  keep alive 是 `0x1F`（回 `0x0B`）。所以机器人登录后会**自己走掉**，之后所有命令
  都变成 `Player 'Steve' cannot be found`——测出来的东西是错的。
- **`run-all.ps1` 现在真的做比对**：它以前只跑一遍、从不和 `.expected` 比。现在默认
  `-Compare`，没有基线的用例只跑一遍（但会在结尾列出来，免得被当成"通过了"），
  需要在线玩家的用例在文件里写一行 `# bot: Steve` 就会自动带上机器人。
- **比较时把 UUID 归一化**：真服每次开服给的实体 UUID 都不一样，所以期望文件里
  留着真的 UUID（好看），比的是占位符——不然含 UUID 的用例自己跟自己对不上。

### 修复

- **49 条命令的用法行改成查语言文件那张表**：以前每条命令的用法行是 gomc
  自己写的，有 **25 条**和 1.12.2 客户端语言文件里的原文对不上
  （`/blockdata` 的 `<nbt>` 该是 `<dataTag>`、`/say <message>` 该是
  `/say <message ...>`、`/execute` 该是那句带 `OR` 的两段式…）。
  真服用法错误的时候回的就是那行原文，写错了等于告诉读者一个错的写法。
  现在 `Usage()` 统一查 `commandUsages` 那张表，**一处改动、不会漂**。
- **`/spawnpoint` 和 `/setworldspawn` 的参数布局对齐真服**。1.12.2 是
  **先数参数再看参数**：
  - **`/spawnpoint` 给 2~3 个参数是用法错误**——位置不能不给玩家就单独给；
    以前 gomc 会把它当成「玩家 + 坐标」，报一个看不懂的错。
  - **不给位置时，出生点落在玩家自己站的地方**
    （`getChunkCoordinates` 是 locX、locY+0.5、locZ 再取整），
    以前 gomc 落在 (0, 0, 0)。
  - **0 个参数时控制台会被拒绝**：
    `You must specify which player you wish to perform this action on.`
    （`/setworldspawn` 也一样）。
  - **`/spawnpoint` 只改第一个匹配到的玩家**（真服走的是
    `PlayerSelector.getPlayer`，给回一个玩家而不是整个列表）。
  - **`/setworldspawn` 也是数参数**：0 个用执行命令的玩家，3 个给位置，
    其它都是用法错误。

  探针用例 `43-spawnpoint-args` 在真服上录了基线，对拍 **9/9 逐字一致**。

### 内部结构

- **1.12.2 特有的规则搬回版本包**。`internal/parser` 和 `internal/exec` 是
  **版本无关的原语层**，但有几处 1.12.2 的规则住错了地方：

  | 搬走的东西 | 从 | 到 |
  |---|---|---|
  | 坐标的上下界和六套坐标规格 | `internal/parser/parser.go` | `internal/version/v1_12_2/coords.go` |
  | 选择器的语法（选项名、参数形状） | `internal/parser/selector.go` | `internal/version/v1_12_2/selector.go` |
  | 选择器的语义（每个过滤器、结果的排序） | `internal/exec/selector.go` | `internal/version/v1_12_2/selector_resolve.go` |
  | 计分板用什么名字跟踪实体 | `internal/exec/selector.go` | `internal/version/v1_12_2/selector_resolve.go` |

  搬完之后 `internal/parser` 里**没有一条 1.12.2 的规则**了，`internal/exec`
  也不再知道选择器长什么样——它只提供执行引擎的通用能力。

  两个「猜着来」的方法删掉了：`parser.Parser.Coord`（它默认用方块位置
  的上下界）和 `parser.Parser.Selector`（它默认用 1.12.2 的选择器语法）。
  现在分别用 `CoordWith(base, spec)` 和版本包自己的 `parseSelector(p)`——
  **坐标和选择器都是版本规则，版本包说了算**。

  **行为没变**：这是纯粹的搬家，测试、golden 和探针基线一行没改。

### 计划中

- **版本抽象**：`internal/world` 里那 1500 行 1.12.2 专属数据（方块名、物品名、
  实体类型、gamerule 默认值）挡在"加第二个版本"的路上，得挪进
  `internal/version/v1_12_2`。（选择器和坐标的规则已经搬完了，见上面的
  「内部结构」。）
- **`ast.Kind` 这套接口是死的**：22 处实现、0 处消费，真正分派靠类型断言；留着
  只会让人以为可以按 kind 写分派。
- **命令的用法行**已经全部改成查语言文件那张表（见上面的「修复」）。

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
