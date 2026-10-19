# 1.12.2 行为规格

这份文档是 gomc 的**行为基准**。每一条都标注了来源：

- **[读]** —— 从服务端 jar 反编译 / `javap` 读字节码得到
- **[验]** —— 启动真实 Paper 1.12.2 服务端实测得到
- **[推]** —— 由前两者推导

`[读]` 和 `[验]` 冲突时**以 `[验]` 为准**。已经发生过一次：`/tag` 命令
从字节码里"读"出来是有的，实测发现 1.12.2 根本没有这条命令。

架构层面的不变量见 [`ARCHITECTURE.md`](ARCHITECTURE.md)，命令清单见
[`COMMANDS.md`](COMMANDS.md)。

---

## 1. 函数

### 1.1 文件布局 [读]

```
<world>/data/functions/<namespace>/<name>.mcfunction
```

namespace 是路径的第一段（`split("/", 2)` 的第一段）。不是 1.13+ 的
datapack 布局。

### 1.2 加载 [读][验]

- 每一行 trim 后按**第一个空格**切成两段（`split(" ", 2)`）。
- `#` 开头的行跳过，空行跳过。
- `//` 开头 → 报错 `Unknown or invalid command '//' (if you intended to make a comment, use '#' not '//')`。
- 以 `/` 开头 → 报错 `Unknown or invalid command '/xxx' (did you mean 'xxx'? Do not use a preceding forwards slash.)`。
- 命令名未注册 → 报错 `Unknown or invalid command 'xxx'`，**整个文件加载失败**。

实测（服务端日志）：

```
[ERROR]: Couldn't read custom function probe:fn3 from .\testworld\data\functions\probe\fn3.mcfunction
java.lang.IllegalArgumentException: Unknown or invalid command 'tag'
```

### 1.3 执行 [读]

- 命令逆序压栈、正序弹出（`CustomFunctionData.a` 的 `for j = len-1; j >= 0; --j`）。
- `maxCommandChainLength` gamerule 限制单次调用的命令总数，默认 65536。
- 函数重入时命令入队而不是内联执行，调用者拿到 0。

### 1.4 坏函数不阻断其他函数 [读]

`CustomFunctionData.h` 对每个文件单独 try/catch，坏文件只记日志、不注册，
其余函数照常加载。`/reload`（`CustomFunctionData.f`）会**先清空注册表再重载**，
所以重载后加载失败的函数会从注册表里消失。

> gomc 的 `check` 子命令把加载失败当**错误**（退出码 1），
> `run`/`tick` 把它当**警告**并继续跑——这与 1.12.2 的容错语义一致。

### 1.5 函数内的输出前缀 [验]

真实服务端里，同一个函数内不同来源的输出前缀不同：

```
[Server] PROBE_FN4_START              ← say 的输出
[Server: Added new objective 'p2' successfully]   ← 命令结果
[Armor Stand] ORDER_TEST              ← execute 子命令的输出（前缀是目标实体名）
[Server: Failed to execute 'detect' as Armor Stand]  ← detect 前置条件失败
```

---

## 2. execute（1.12.2 旧语法）

### 2.1 语法 [读]

```
execute <实体> <x> <y> <z> [detect <x2> <y2> <z2> <方块> <数据值>] <命令>
```

- 至少 5 段，否则报用法错误。
- 坐标**相对于目标实体**，不是相对于执行者。
- 没有 1.13 的 `as` / `at` / `run`。
- `detect` 后面固定 5 段（三个坐标 + 方块 + 数据值），数据值可以是 `-1` 表示任意。

### 2.2 多实体 [读][验]

`execute` 对**每个匹配实体各执行一次**。

实测：召唤两个盔甲架后，`execute @e[type=armor_stand] ~ ~ ~ say ORDER_TEST`
执行了两次。

执行**顺序为倒序**（MC-121728，1.13 才修），依据是 1.13 官方变更日志
"The order, in which execute runs the command for each entity, is in reverse"。
（次数是实测的，顺序是查证来的——实测时无法让两个实体输出可区分的名字。）

### 2.3 失败传播 [读]

多实体时若某条失败，**其余实体的执行被取消**（MC-121727）。

### 2.4 执行者 [验]

被选中的实体成为子命令的执行者，因此子命令的输出前缀是**目标实体的名字**，
而不是 `[Server]`。`@s` 因此指向目标实体。

---

## 3. 目标选择器

### 3.1 形状 [读]

正则 `^@([pares])(?:\[([^ ]*)\])?$`：

- base 只能是 `p` `a` `r` `e` `s`。
- 参数表里**不允许出现空格**。
- 整个字符串必须匹配，所以 `@e[x=1] 后面有东西` 会被当成普通名字而不是选择器。

### 3.2 参数表是平坦 map [读]

`PlayerSelector.d` 的解析方式：

- 用 `Splitter.on(',')` 切分参数表（忽略空段）。
- 每一段用 `Splitter.on('=').limit(2)` 切分，得到 key 和 value。
- **重复的参数名后者覆盖前者**：`@e[tag=a,tag=b]` 等价于 `@e[tag=b]`。
- key 不在白名单里 → 报错 `Invalid selector argument '<整段>'`。

### 3.3 参数清单 [读]

| 参数 | 含义 | 方向 |
|---|---|---|
| `r` / `rm` | 距离上限 / 下限 | `r` 大、`rm` 小 |
| `l` / `lm` | 经验等级上限 / 下限 | `l` 大、`lm` 小 |
| `x` `y` `z` | 参考位置 | |
| `dx` `dy` `dz` | 体积尺寸 | |
| `rx` / `rxm` | 俯仰角**上限** / 下限 | `rx` 大 |
| `ry` / `rym` | 偏航角**上限** / 下限 | `ry` 大 |
| `c` | 结果数量 | |
| `m` | 游戏模式 | |
| `team` | 队伍名 | |
| `name` | 实体名 | |
| `type` | 实体类型 | |
| `tag` | 计分板标签 | |
| `score_<目标>` | 分数**上限** | |
| `score_<目标>_min` | 分数**下限** | |

### 3.4 1.12.2 **没有**的东西 [读]

- **`scores={...}`** —— 1.13 才有的写法。1.12.2 用顶层参数 `score_<目标>`。
- **`nbt={...}`** —— 1.13 才加的。Minecraft Wiki 变更日志原文：
  "1.13 17w45a 在目标选择器中加入了 nbt"。1.12.2 写这个会直接报
  `Invalid selector argument`。
- **隐式位置参数**（`@e[167,28,454,5]`）—— 1.11 起移除了。

### 3.5 各过滤器的精确语义 [读]

全部来自 `PlayerSelector` 的匿名内部类（字节码）：

| 过滤器 | 内部类 | 语义 |
|---|---|---|
| 距离 | `$2` | 比较**平方距离**。`r` 是上限、`rm` 是下限，两者都给时都要满足 |
| 俯仰角 | `$4` | 角度取 `floor` 后折到 `[-180,180)`。`rx` 是**上限**、`rxm` 是下限 |
| 偏航角 | `$3` | 同上，`ry` 上限、`rym` 下限 |
| 体积 | `$5` | 选**包围盒与体积相交**的实体，不是位置点落入体积的实体 |
| 排序 | `$6` | 按到参考位置的**平方距离**升序 |
| 类型 | `$7` | `EntityTypes.a(entity, key)`，支持 `!` 取反 |
| 等级 | `$9` | 只对玩家生效。`l` 是上限、`lm` 是下限 |
| 游戏模式 | `$10` | 只对玩家生效。支持 `!` 取反 |
| 分数 | `$12` | 目标必须存在，且实体必须在该目标下有分数 |
| 名称 | `$13` | `Entity.getName()`：自定义名 → 玩家名 → 类型注册名 |
| 标签 | `$14` | 空标签名表示"没有标签" |

**角度区间支持环绕**：下界大于上界时，区间跨北展开。
例如 `rym=90,ry=-90` 匹配 `[-180,-90] ∪ [90,180]`。

只有 `rx`/`rxm` 之一存在时才启用俯仰角过滤，只有 `ry`/`rym` 之一存在时
才启用偏航角过滤。缺省的上界是 359（折到 -1），下界是 0。

### 3.6 base 的差异 [读]

- `@a` / `@p` —— 加"必须是玩家"的过滤。
- `@r` —— 同样加"必须是玩家"，**但**如果给了 `type=`，类型过滤会**取代**它，
  于是 `@r[type=zombie]` 可以选到僵尸。
- `@e` —— 不加玩家过滤，选所有实体。
- `@s` —— 只指执行者自己。

### 3.7 `c` 的语义 [读]

- `@a` / `@e` 默认 **0**（不限），`@p` / `@r` / `@s` 默认 **1**。
- 负数会**先反转整个列表**再截断。所以 `@e[c=-3]` 不是"最后三个"，
  而是"反转后的前三个"。

### 3.8 参考位置的 0.5 偏移 [读]

`x` / `y` / `z` 覆盖参考位置时，x 和 z **加 0.5**（块中心），y 不加。
所以 `@e[x=10]` 的参考点是 x=10.5。

### 3.9 名称匹配 [读]

`Entity.getName()` 的顺序是：自定义名（`CustomName`）→ 玩家名 → 类型注册名
（`EntityTypes` 里注册的显示名，比如 "Zombie"、"ArmorStand"、"PigZombie"）。

---

## 4. scoreboard

### 4.1 分数是 Java int [读]

所有分数按**有符号 32 位**存储，算术结果回绕，除法向零取整，
`%` 的符号跟被除数走。

### 4.2 子命令 [读]

```
scoreboard objectives <add|remove|list|setdisplay>
scoreboard players    <set|add|remove|reset|list|enable|test|operation|tag>
scoreboard teams      <add|remove|join|leave|empty|list|option>
```

### 4.3 `players test` [读]

```
scoreboard players test <持有者> <目标> <最小值> [<最大值>]
```

- 最小值可以是 `*`（无下限），最大值也可以是 `*`（无上限）。
- 只给一个值时，上下限相等，即相等判断。
- 目标不存在、或持有者在该目标下没有分数 → 报错。
- 分数不在区间内 → 报错（是**失败**不是静默）。

### 4.4 `players tag` [验]

**1.12.2 没有独立的 `/tag` 命令**（实测确认）。实体标签用：

```
scoreboard players tag <目标> <add|remove|list> [标签]
```

### 4.5 `list` 的输出 [验]

真实服务端**从控制台**执行时的输出：

```
Showing 3 tracked objective(s) for Steve:
- p2: 5 (p2)
- q: 5 (q)
- probe: 5 (probe)
```

两个要注意的地方：

1. 这不是聊天格式（不以 `[` 开头），而是普通的多行 INFO 输出。
2. **在函数里看不到**——同样一条命令放在 `.mcfunction` 里，控制台不会有输出。
   实测两处都确认过。

所以自动化对拍时，如果要校验 `scoreboard players list`，得从控制台跑，
不能放在函数里。

---

## 5. gamerule

1.12.2 有 **24 条** gamerule（`GameRules` 的静态初始化块）：

```
announceAdvancements      true      disableElytraMovementCheck  false
commandBlockOutput        true      doDaylightCycle            true
doEntityDrops            true      doFireTick                 true
doLimitedCrafting         false     doMobLoot                 true
doMobSpawning             true      doTileDrops                true
doWeatherCycle            true      gameLoopFunction           "-"
keepInventory             false     logAdminCommands           true
maxCommandChainLength     65536     maxEntityCramming         24
mobGriefing               true      naturalRegeneration        true
randomTickSpeed           3         reducedDebugInfo           false
sendCommandFeedback       true      showDeathMessages          true
spectatorsGenerateChunks  true      spawnRadius                10
```

**没有** `tntExplodes`（1.13 才加）。

`sendCommandFeedback` 决定命令结果是否回显，gomc 用它控制 CLI 的输出。

---

## 6. 方块

### 6.1 `/setblock` 的模式 [读]

```
setblock <x> <y> <z> <方块> [<数据值>] [<模式>] [<nbt>]
```

模式在**第 6 位**（参数索引 5）：

- `replace`（默认）—— 直接覆盖。
- `destroy` —— 先清空位置再放，方块是空气时直接成功返回。
- `keep` —— 位置非空时报错 `The block couldn't be placed`。

### 6.2 `/fill` 的模式 [读]

```
fill <x1> <y1> <z1> <x2> <y2> <z2> <方块> [<数据值>] [<模式> [<替换方块> [<替换数据值>]]]
```

模式在**第 9 位**（参数索引 8）：

- `replace`（默认）—— 给了替换方块时只替换匹配的方块。
- `destroy` —— 先清空。
- `keep` —— 跳过非空的位置。
- `hollow` —— 只填壳，内部清空成空气。
- `outline` —— 只填壳。

---

## 7. 时间

### 7.1 `/time` [读][验]

```
time set <值>      值可以是 day(1000) / night(13000) / 数字
time add <值>      数字
time query <daytime|day|gametime>
```

- `daytime` = 一天内的时刻（0..23999）。
- `day` = 创建世界以来经过的天数。
- `gametime` = 世界总 tick 数。

**`daytime` / `noon` / `midnight` 作为 `set` 的值是 1.13 才有的**，
1.12.2 只认 `day` 和 `night`。实测 `time set 18000` 的回显是
`Set the time to 18000`。

---

## 8. 消息里的选择器 [验]

`/say` 和 `/me` 会解析消息里的选择器：

```
say @a          → 展开成匹配到的实体名
say @a（无匹配） → Selector '@a' found nothing
```

---

## 9. 还没搞清楚的

| 问题 | 现状 |
|---|---|
| execute 多实体的**顺序** | 依据 1.13 变更日志判定为倒序，未能实测（无法让两个实体输出可区分的名字） |
| 实体 UUID 与计分板持有者的对应 | 未建模，gomc 用内部 id 代替 |
| 函数内命令失败何时报错 | 真实服务端有的报有的不报，gomc 统一报 |
