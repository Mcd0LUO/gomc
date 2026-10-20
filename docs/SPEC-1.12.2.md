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

**函数 id 的大小写**（实测）：1.12.2 用 `MinecraftKey` 标识函数，它的构造器把
**命名空间和路径都小写化**，并且没有命名空间时默认补 `minecraft:`。所以：

| 写法 | 实际指向 |
|---|---|
| `function PROBE:mcns` | `probe:mcns` |
| `function mcns` | `minecraft:mcns` |
| `function probe:MCNS` | `probe:mcns` |

找不到时回显的是**规范化之后**的 id：`Unknown function 'minecraft:mcns'`。

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
- 函数重入时命令**入队**而不是内联执行，调用者拿到 0（回显是
  `Executed 0 command(s) from function '...'`）。被调函数的命令被
  `addFirst` 塞进共享队列，所以在**当前这条命令之后立刻**执行，排在调用者
  剩下的命令**前面**（`CustomFunctionData.a` 的 `this.h` 分支 +
  `CustomFunction$d.a`）。
- 函数里的命令以**权限等级 2** 执行：`CommandFunction.execute` 用
  `CommandListenerWrapper.a(2)` 压住执行者的等级，而
  `CommandListenerWrapper.a(level, name)` 取两者里**较低**的那个。
  所以需要更高权限的命令在函数里会被拒——`/reload` 需要等级 3
  （`CommandReload.a()`），原版 1.12.2 会报
  `You do not have permission to use this command`
  （`commands.generic.permission`），而不是真的重载。
  > Paper 把这条改成了**静默忽略**，见第 10 节。
  > 原版 1.12.2 里，同样的道理适用于 `op` / `deop` / `stop` / `save-*` /
  > `ban*` / `kick` / `whitelist` / `setidletimeout` 这些等级 3 的命令。

### 1.3.1 一条命令失败不会中止函数 [读]

1.12.2 是**逐条命令**捕获异常的
（`CommandBlockListenerAbstract.executeCommand` 里每个命令一个 try/catch），
所以一条命令失败之后，函数**继续执行下一条**。

实测：函数里放一条会失败的 `execute ... detect ...`，服务端报
`[Server: Failed to execute 'detect' as Armor Stand]`，然后函数继续跑后面的命令。

> gomc 早期版本是"第一条命令失败就中止整个函数"，这会让被测函数的后半段
> 悄悄不执行——对拍时表现为"结果对不上但看不出为什么"。现在失败会被收集起来
> 统一报告，执行不再中断。

### 1.3.2 实体标签就是 NBT 里的 `Tags` [读]

1.12.2 的 `Entity.getScoreboardTags()` 直接读实体 NBT 里的 `Tags` 列表，
所以这两件事是**同一份数据**：

- `summon armor_stand 0 0 0 {Tags:["w"]}` → `@e[tag=w]` 立刻能选中它。
- `scoreboard players tag @e[...] add w` → 实体的 NBT 里也会出现 `Tags`。

gomc 把 `Tags` 做成 NBT 的虚拟字段（`Entity.TagList` / `Entity.SetTag`），
不做第二份拷贝。

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
- `detect` 后面固定 5 段（三个坐标 + 方块 + **数据值必填**），数据值可以是 `-1` 表示任意。
  数据值缺失或不是数字都会让命令失败——1.12.2 的 `detect` 只有在参数数够时才进入
  检查分支（`astring.length > 10`），而方块状态（`state`）在 1.12.2 会被当成方块名解析。

**坐标基准（实测 + 源码）**：`execute` 设定的执行位置 = 目标实体的位置 + 坐标偏移；
`detect` 的相对坐标以**执行位置**为基准（不是目标实体的位置），绝对坐标按世界原点算。

源码里的顺序是固定的（`CommandExecute.execute`）：

```
d0 = b(entity.locX, astring[1], false)   // 执行位置 X
d1 = b(entity.locY, astring[2], false)   // 执行位置 Y
d2 = b(entity.locZ, astring[3], false)   // 执行位置 Z
...
d3 = b(d0, astring[5], false)            // detect 位置 X
d4 = b(d1, astring[6], false)            // detect 位置 Y
d5 = b(d2, astring[7], false)            // detect 位置 Z
```

`b(base, s, false)`（`CommandAbstract.b`）的语义是：参数以 `~` 开头 = `base + 数字`，
否则就是**绝对坐标**；第三个参数 `false` 表示不加块中心的 `0.5`。

验证方法：实体在 (0, 10, 0)，只有 (0, 11, 0) 有石头。

| 命令 | 检查的位置 | 结果 |
|---|---|---|
| `execute @e ~ ~1 ~ detect ~ ~ ~ stone 0` | (0, 11, 0) | 通过 |
| `execute @e ~ ~1 ~ detect ~ ~-1 ~ stone 0` | (0, 10, 0) | 失败 |
| `execute @e ~ ~ ~ detect ~ ~1 ~ stone 0` | (0, 11, 0) | 通过 |

> **偏移只能加一次。** detect 的相对坐标是在**执行位置**上再折算一遍的，
> 不是在实体的原始位置上折算。如果 detect 的坐标在解析期就按执行坐标折算一遍、
> 运行期再折算一遍，偏移就会被加两次（`~ ~1 ~` 会变成 +2）。

### 2.2 多实体 [读][验]

`execute` 对**每个匹配实体各执行一次**。

**展开发生在命令分发层，不在 `execute` 内部。** 1.12.2 的命令分发入口
（`CommandHandler.a(ICommandListener, String)`）会找出命令的第一个"目标列表"参数，
把选择器解析出来，然后**对每个匹配实体把该参数替换掉、单独执行一次整条命令**。
`execute` 的第 0 个参数正是这样的参数（`CommandExecute.isListStart` 只对 `i == 0`
返回 true）。

**顺序就是选择器返回的顺序，不是倒序。** `PlayerSelector` 对 `@a` `@e` `@p` 按距离平方
**升序稳定排序**（`Collections.sort`，比较 `entity.d(x,y,z)`），所以离执行点最近的实体
先执行；距离相同时保持世界遍历顺序。`@r` 先打乱。`c` 为负数时先 `Collections.reverse`
再截断。

实测：三个带 `CustomName` 的盔甲架，到执行点的距离分别是 5.10 / 6.40 / 9.43，
输出依次是 `[NEAR]` `[MID]` `[FAR]`——按距离升序。

> 这里此前写的是"倒序（MC-121728，依据 1.13 官方变更日志）"，那是**查证**来的、没有实测。
> 用可区分名字的实体一测就推翻了。MC-121728 的实质是"顺序不确定"，不是"确定倒序"。

### 2.3 失败传播 [读][验]

**一个实体失败不影响其余实体。** 分发层对每个实体单独 try/catch：捕获异常、发消息、
`continue` 到下一个实体（`CommandHandler.a(...)` 里的 `continue`，不是 `break`）。

实测：距离最近的实体 `detect` 失败（报 `Failed to execute 'detect' as NEAR`），
后面两个实体照样执行。

> 这里此前写的是"其余实体的执行被取消（MC-121727）"，同样是查证来的、没有实测，
> 源码里也不是 `break`。

### 2.4 执行者 [验]

被选中的实体成为子命令的执行者，因此子命令的输出前缀是**目标实体的名字**，
而不是 `[Server]`。`@s` 因此指向目标实体。

名字的取法与 `Entity.getName` 一致：自定义名（`CustomName`）→ 玩家名 → 类型的**翻译名**。
翻译名来自语言文件里的 `entity.<去掉 Entity 前缀的类名>.name`，所以是 "Armor Stand"
这样带空格的写法，不是注册名 `armor_stand`。语言文件里没有的键（比如扔出去的蛋
`entity.Egg.name`）会原样返回键本身，gomc 也照抄。

### 2.5 目标实体有重力 [验]

**目标实体会落到地面上**，所以"实体的位置"通常是它落地后的位置，而不是
`/summon` 时给的那个。

实测（超平坦世界，草地层在 y=3）：

| 操作 | 实体的实际位置 |
|---|---|
| `summon armor_stand 1 5 0` | y=4（落到地面上） |
| `tp @e[type=armor_stand] 3 9 0` | y=9（tp 不触发下落） |

这条很容易把人带沟里：早先的探针用例以为"detect 失败"能复现，结果三个实体
全部检测成功——因为那时候的世界（`generator-settings` 写错了）地面在 y=0，
盔甲架一路掉到 y=1，脚下正好是石头。

**gomc 不模拟重力**（明确不做），`/summon` 的实体就留在给定的坐标上。
这是 gomc 与真实服务端的一处已知差异：凡是依赖"落地后位置"的相对坐标，
两边会差一格。

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
- key 不在白名单里 → 报错 `Invalid selector argument: '<逗号分隔的那一段>'`
  （注意有冒号；引号里是整段，比如 `nbt={...}`，不是 key 本身）。

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
| 类型 | `$7` | `EntityTypes.a(entity, key)`，支持 `!` 取反。**键必须是 1.12.2 的注册名**（大小写和命名空间都不敏感：`type=Zombie`、`type=MINECRAFT:ZOMBIE` 都认；`type=pigzombie` **不认**，1.12.2 已改名 `zombie_pigman`） |
| 等级 | `$9` | 只对玩家生效。`l` 是上限、`lm` 是下限 |
| 游戏模式 | `$10` | 只对玩家生效。支持 `!` 取反。**未知模式算 `NOT_SET`**：非取反时恒假，取反时恒真，所以 `@a[m=!abc]` 命中**全部**玩家 |
| 分数 | `$12` | 目标必须存在，且实体必须在该目标下有分数 |
| 名称 | `$13` | `Entity.getName()`：自定义名 → 玩家名 → 类型的**翻译名**（见 3.9）。**玩家走的是 `EntityHuman.getName()`，只返回用户名、忽略自定义名** |
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

### 3.7 `c` 的语义 [读][验]

- `@a` / `@e` 默认 **0**（不限），`@p` / `@r` / `@s` 默认 **1**。
- 负数会**先反转整个列表**再截断。所以 `@e[c=-3]` 不是"最后三个"，
  而是"反转后的前三个"。
- **`c` 不是整数时不报错，而是回落到默认值**。1.12.2 用
  `MathHelper.a(String, int)` 解析它，那是 `Integer.parseInt` 失败就用默认值。
  所以 `c=abc`、`c=1.0`、`c=~` 都不是错误——真服上 `kill @e[c=abc]`
  照常执行。

### 3.8 参考位置的 0.5 偏移 [读]

`x` / `y` / `z` 覆盖参考位置时，x 和 z **加 0.5**（块中心），y 不加。
所以 `@e[x=10]` 的参考点是 x=10.5。

### 3.9 名称匹配 [读][验]

`Entity.getName()` 的顺序是：自定义名（`CustomName`）→ 玩家名 → 类型的**翻译名**。

**但玩家是例外**：`EntityHuman.getName()` 覆写了父类，**只返回 GameProfile 的
用户名，完全忽略 CustomName**。所以对玩家来说，`name=` 比较的是**用户名**。

自定义名**原样返回**（真服 `getCustomName()` 不去引号）——名字里的引号
是名字的一部分，所以 `@e[name="Bob"]` 和 `@e[name=Bob]` 是两个不同的东西。

翻译名来自语言文件里的 `entity.<去掉 Entity 前缀的类名>.name`，
所以是 "Zombie"、"Armor Stand"、"Zombie Pigman" 这样**带空格**的写法，
不是注册名 `armor_stand`。语言文件里没有的键会原样返回键本身
（比如扔出去的蛋是 "entity.Egg.name"）。

实测：真服里盔甲架的输出前缀是 `[Armor Stand]`，不是 `[ArmorStand]`。

### 3.10 空匹配在命令层的处理 [验]

**列表**选择器（`@e` / `@a`，也就是 `c` 不为 1 的那种）匹配不到东西时
**不是错误**：真服上命令跑 0 次，什么都不说，命令成功。

实测（Paper 1.12.2，用 `say MARKER-x` 把命令隔开，定位错误落在哪一条）：

```
scoreboard players set @e[type=zombie] obj 1   → 静默
kill @e[type=zombie]                            → 静默
gamemode creative @a                            → 静默
scoreboard players tag @e[type=zombie] list     → 静默
execute @e[type=zombie] ~ ~ ~ say HELLO         → 静默
say @e[type=zombie]                             → Selector '@e[type=zombie]' found nothing
kill @p                                         → Entity '@p' cannot be found
gamemode creative @p                            → Player '@p' cannot be found
kill NoSuchPlayer                               → Entity 'NoSuchPlayer' cannot be found
```

三条要点：

- **只有消息展开**（把选择器写进消息里的那些命令）才报
  "Selector '...' found nothing"——那是命令自己的检查
  （`CommandAbstract.b` → `PlayerSelector.getPlayerNames` 返回 null）。
- **单实体目标**（`@p`、`c=1`、裸名字）匹配不到时**要报错**，而且措辞
  跟着命令的目标类型走：解析成实体的命令说 "Entity"，解析成玩家的说
  "Player"。
- `execute` 包着的子命令如果影响了 0 个实体，`execute` 报
  "All invocations failed: '<子命令>'"。

> ⚠️ **源码和真服对不上**：vanilla 的分发器（`CommandHandler.java:59-61`）
> 在列表参数为空时是会抛 `ExceptionPlayerNotFound("commands.generic.selector.notFound")`
> 的，字节码里也确实有这段。但参照服务端（Paper 1.12.2）跑出来是**静默**。
> 本项目以**真服实测**为准。

### 3.11 列表连接符 [验]

真服把一组名字拼成一行时，最后一个前面用 `" and "`，其余之间用 `", "`：

```
say @e[type=armor_stand]        → A, B and C
say @e[type=armor_stand,c=2]    → A and B
say @e[type=armor_stand,c=1]    → A
```

这是 `CommandAbstract.a(List<IChatBaseComponent>)`（`CommandAbstract.java:510`）
的规则，和 `CommandAbstract.a(Object[])` 一样。

### 3.12 还没对齐的选择器细节 [验]

- **`scoreboard players tag list` 的标题名**：真服用解析后的用户名 / UUID，
  gomc 对**玩家**已经改成用户名了，但非玩家实体仍是内部 id（gomc 不建模 UUID，
  见第 12 节）。

### 3.13 已经对齐的几条细节 [验]

下面这几条原来是"没对齐"，现在按真服实测改掉了：

- **空匹配在命令层**——见 3.10。
- **消息展开的连接符**——见 3.11。
- **`Entity type 'x' is invalid` 那条提示**。真服在 `type=` 指向未注册类型时
  会发一条红字，然后**匹配为空、命令照样成功**。名字是**带命名空间的完整键**：
  `Entity type 'minecraft:pigzombie' is invalid`；否定写法 `type=!pigzombie`
  得到的是**同一条**（不带 `!`）。gomc 现在也会打这条提示——它不属于命令的
  错误，而是**选择器解析时产出的附注**，所以走的是命令的输出通道。
- **`@s` 没有执行者时的措辞**。真服上 `kill @s`（控制台）说
  `Entity '@s' cannot be found`，而把 `@s` 写进消息里说
  `Selector '@s' found nothing`——**措辞归命令管，不归选择器管**。
- **tag list 对没有 tag 的实体不打标题**。真服的
  `CommandScoreboard.java:735-741` 只在 tag 集合非空时才发标题和列表，
  命令报的计数是**tag 的个数**。
- **`/summon item` 需要一个非空的物品堆**。1.12.2 的实体工厂只接受带着
  非空物品堆的物品实体，空堆的实体建出来就死，所以 summon 直接失败：

  ```
  summon item                                        → Unable to summon object
  summon item {Item:{}}                              → Unable to summon object
  summon item {Item:{id:"stone"}}                    → Unable to summon object
  summon item {Item:{id:"stone",Count:0b,Slot:0b}}  → Unable to summon object
  summon item {Item:{id:"stone",Count:256}}          → Unable to summon object
  summon item {Item:{id:"stone",Count:-1}}           → Unable to summon object
  summon item {Item:{id:"stone",Count:1b}}           → Object successfully summoned
  summon item {Item:{id:"stone",Count:1}}            → Object successfully summoned
  ```

  决定成败的是 **Count**：1.12.2 用 `getByte("Count")` 读它，**任何数字类型都
  会被截断成一个有符号字节**——所以 `Count:1`（Int）算数，`Count:256` 截断成 0
  不算，`Count:-1` 是负数也不算；缺省的 Count 就是 0。`Slot` 缺不缺无所谓
  （默认 0）。物品名也必须是**已注册的物品**——`{Item:{id:"notanitem",Count:1b}}`
  真服也会失败（那个堆是空的）。
- **物品实体的显示名**。真服上物品实体是**按它带的那件物品命名**的，而服务端
  的语言文件里没有那个键，所以真服直接把键打出来：

  ```
  summon item {Item:{id:"diamond_sword",Count:1b}}  → item.item.swordDiamond
  summon item {Item:{id:"stone",Count:1b}}           → item.tile.stone.stone
  summon item {Item:{id:"anvil",Count:1b}}           → item.tile.anvil.intact
  summon item {Item:{id:"acacia_boat",Count:1b}}    → item.item.boat.acacia
  ```

  普通物品是 `item.item.<旧名>`（`swordDiamond` 是 1.12.2 的**旧**名字，注册名
  是 `diamond_sword`），方块物品是 `item.tile.<方块旧名>[.<变种>]`。
  gomc 现在按同一张表还原（那张表是从服务端 jar 的字节码里挖出来的 409 个
  物品名，见 `internal/world/itemnames.go`）。

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
- **只给一个值时，上限是 `2147483647`（Integer.MAX_VALUE），不是等于下限**：
  `players test Steve t 6` 的意思是"至少 6"，不是"恰好 6"。
  实测：真服回 `Score 5 is NOT in range 6 to 2147483647`。
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

### 4.6 输出文本 [验][读]

计分板各子命令的回显文字全部来自语言文件，逐条对齐（`<目标>` 一律指
**目标名**，`<显示名>` 指目标的显示名）：

| 命令 | 输出 |
|---|---|
| `objectives add` | `Added new objective '<名字>' successfully` |
| `objectives remove` | `Removed objective '<名字>' successfully` |
| `objectives list` | `Showing <n> objective(s) on scoreboard:`，每行 `- <名字>: displays as '<显示名>' and is type '<类型>'` |
| `objectives setdisplay <槽> <名字>` | `Set the display objective in slot '<槽>' to '<名字>'` |
| `objectives setdisplay <槽>` | `Cleared objective display slot '<槽>'` |
| `players set` | `Set score of <目标> for player <持有者> to <分数>` |
| `players add` / `players remove` | **有回显**，和 `set` 共用同一条：`Set score of <目标> for player <持有者> to <分数>`（分数是**操作之后**的值） |
| `players reset <持有者> <目标>` | `Reset score <目标> of player <持有者>` |
| `players reset <持有者>` | `Reset scores of player <持有者>`（注意是复数 `scores`） |
| `players operation` | `Operation applied successfully` |
| `players enable` | `Enabled trigger <目标> for <持有者>` |
| `players test`（通过） | `Score <分数> is in range <下限> to <上限>` |
| `players test`（不通过） | `Score <分数> is NOT in range <下限> to <上限>` |
| `players list <持有者>` | `Showing <n> tracked objective(s) for <持有者>:`，每行 `- <显示名>: <分数> (<目标名>)` |
| `players list`（不带持有者） | `Showing <n> tracked players on the scoreboard:` |
| `players tag ... add` | `Tag <标签> added` |
| `players tag ... remove` | `Tag <标签> removed` |
| `players tag ... list` | `Tags on entity <实体> are:`，下一行是标签，多个用 ` and ` 连接 |
| `teams add` | `Added new team '<名字>' successfully` |
| `teams remove` | `Removed team <名字>` |
| `teams join` | `Added <n> player(s) to team <队伍>: <玩家列表>` |
| `teams leave` | `Removed <n> player(s) from their teams: <玩家列表>` |
| `teams empty` | `Removed all <n> player(s) from team <队伍>` |
| `teams list` | `Showing <n> teams on the scoreboard:`，每行 `- <队伍名>: '<显示名>' has <n> players` |
| `teams option` | `Set option <选项> for team <队伍> to <值>` |

**`players list` 每行的三段分别是什么**（`CommandScoreboard` 传的参数是
`{分数, 目标的显示名, 目标名}`，语言键 `commands.scoreboard.players.list.player.entry`
的格式是 `- %2$s: %1$s (%3$s)`）：

| 位置 | 含义 |
|---|---|
| `- <显示名>` | 目标的**显示名**（`ScoreboardObjective.getDisplayName()`） |
| `<分数>` | 该持有者在这个目标下的分数 |
| `(<目标名>)` | 目标的**名字**（`getName()`），**不是**计分标准类型 |

所以目标 `counter`（计分标准 `dummy`）在真服里是 `- counter: 20 (counter)`，
不是 `- counter: 20 (dummy)`——`(dummy)` 是 `objectives list` 那行才有的东西
（`- <名字>: displays as '<显示名>' and is type '<类型>'`）。

注意 `players set` 那条用的是**目标名**（源码传的是 `getName()`），
而 `players list` 的第一段用的是**显示名**——两者不一样，不要混。

错误文本：

| 场景 | 文本 | 语言键 |
|---|---|---|
| 目标不存在 | `No objective was found by the name '<名字>'` | `commands.scoreboard.objectiveNotFound` |
| 持有者在该目标下没有分数 | `No <目标> score for <持有者> found` | `commands.scoreboard.players.test.notFound` |
| 要删的标签不存在 | `Entity didn't have the <标签> tag` | `commands.scoreboard.players.tag.notFound` |

`players tag ... list` 里的实体标识符，真服用的是 UUID；gomc 用内部 id
（已知简化，见"还没搞清楚的"）。

### 4.6.1 只读目标 [验][读]

有些目标是**只读**的，命令改不了：`health`、`xp`、`level`、`food`、`air`、`armor`。
往这些目标里写分数会报：

```
The objective '<名字>' is read-only and cannot be set
```

（语言键 `commands.scoreboard.objectiveReadOnly`。）

`dummy`、`trigger`、`deathCount`、`playerKillCount`、`totalKillCount` 可以写。

### 4.7 `teams` 子树 [读]

1.12.2 的 `scoreboard teams` 用法（语言键 `commands.scoreboard.teams.*.usage`）：

```
scoreboard teams add <名字> [显示名 ...]
scoreboard teams remove <名字>
scoreboard teams join <队伍> [玩家 ...]
scoreboard teams leave [玩家 ...]
scoreboard teams empty <队伍>
scoreboard teams list [队伍]
scoreboard teams option <队伍> <选项> <值>
```

几条容易搞错的：

- **`leave` 没有队伍参数**。它把人从**所在的所有队伍**里移出去
  （`Scoreboard.removePlayerFromTeam`）；不在任何队伍时报
  `Could not remove %s player(s) from their teams: %s`。
- **`join` 不会因为人已经在别的队伍而报错**。命令层没有这个检查，
  `Scoreboard.addPlayerToTeam` 是"先从旧队移除、再加入新队"，也就是**移动**，
  回显是成功那条。唯一的失败是**队伍不存在**：
  `Could not add %s player(s) to team %s: %s`。
- **`empty` 对空队会报错**：`Team %s is already empty, cannot remove nonexistant players`
  （原版就带着 "nonexistant" 这个拼写错误）。
- **`option` 的选项名不合法是用法错误**，不是"取值不对"：
  `noValue`（`Valid values for option %s are: %s`）是给**取值**不对用的。
  合法选项：`friendlyfire` / `color` / `seeFriendlyInvisibles` /
  `nametagVisibility` / `deathMessageVisibility` / `collisionRule`。
- **队伍名找不到**用 `No team was found by the name '%s'`，
  但 `join` 不用它（见上）。
- **`list` 的两种形式的条目格式不一样**：
  - `teams list`（全部队伍）：`Showing %s teams on the scoreboard:` +
    每行 `- <队伍名>: '<显示名>' has <人数> players`
  - `teams list <队伍>`：`Showing %s player(s) in team %s:` +
    **一行纯文本**的成员名列表（不是条目格式）
  - 同理，`players list`（不带持有者）也是 `Showing %s tracked players on
    the scoreboard:` + **一行纯文本**；只有 `players list <持有者>` 才是
    条目格式 `- <显示名>: <分数> (<目标名>)`。

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

### 6.0 方块的注册名 [验]

1.12.2 的方块注册名有 **254** 个（**不是** 390——390 是 `blockstates.json` 的
条目数，把变种也算进去了）。名字写错会报：

```
There is no such block with name minecraft:<名字>
```

名字按 `MinecraftKey` 解析：转小写，冒号只有在**不是第 1 位也不是第 2 位**时
才是命名空间分隔符，而注册表里只有 `minecraft` 这一个命名空间。所以

- `stone`、`minecraft:stone`、`:stone`、`a:stone` 都是同一个方块 ✓
- `foo:stone` 是别的命名空间，报 `There is no such block with name foo:stone`
- `stone:1` 是命名空间 `stone` 的键 `1`，报 `There is no such block with name stone:1`
- 数字形式（`setblock 0 4 0 1`）不行，报 `There is no such block with name minecraft:1`

**1.12.2 同时保留两套命名**：旧版的（`stonebrick`、`planks`、`log`）和 1.12 的
（`cobblestone`、`stone`），两套都认。

**变种不是独立的方块**：`granite`、`polished_granite`、`white_wool` 这些是方块的
**数据值**，不是方块注册名，写进去会报"没有这个方块"。

校验放在**解析期**——那正是 1.12.2 做这件事的时机，所以函数里的方块名写错时，
函数**照样加载**，错误在**运行到那条命令时**才出现。

表是从服务端 jar 的字节码里挖出来的（`Block` 类注册旧命名、`Blocks` 类注册
1.12 命名），见 `internal/world/blockregistry.go`；再用真服把 254 个名字逐个
放一遍确认过。

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

**只有这五个词特殊，其它词一律当 `replace`**——真服不对模式位报错。
（真服的位置是严格按位数的：`stone replace` 在真服上会被当成"数据值"，
而数据值只接受 0..15，所以报 "'replace' is not a state for block minecraft:stone"。
gomc 更宽松，两种写法都收。）

**第 10 位有两种身份**（真服 `CommandFill`）：放置的方块**有方块实体**时它是
**dataTag**（NBT），否则它是**要替换的方块**。gomc 没有方块实体注册表，
用形状判断：`{...}` 开头当 NBT，否则当方块名。

**`-1` / `*`（"任意数据值"）只对第 11 位的替换数据值有效**，
放置方块自己的数据值（第 8 位）只接受 0..15。

### 6.3 `/effect`、`/trigger`、`/particle` 的参数 [读]

```
effect <玩家> <效果> [<秒数>] [<等级>] [<隐藏粒子>]  OR  effect <玩家> clear
trigger <目标> <add|set> <值>
particle <名称> <x> <y> <z> <dx> <dy> <dz> <速度> [<数量>] [<模式>] [<观察者>]
```

- **`/effect` 的 `<效果>` 接受 id 或注册名**（`speed` / `1` 都行），
  错误文案仍然说 "ID"（那是语言文件里的原话）。
  秒数默认 30、范围 0..1000000；等级默认 0、范围 0..255；
  `隐藏粒子` 只有字面量 `true`（不分大小写）才算。
- **1.12.2 没有"瞬间效果"的特殊分支**。`MobEffectList.isInstant()` 在 1.12.2 的
  字节码里恒返回 false，所以时长永远是 `秒数 * 20` tick。
  > 反编译出来的 `CommandEffect.java` 里**看着**有个 `isInstant()` 分支，
  > 但在 1.12.2 里那是**死代码**。又一次"读"出来的印象不如字节码。
- **`/trigger` 的 `set` 是赋值不是累加**（gomc 之前一律当加法）。
  真服只有**创造模式**才发成功消息，gomc 一律发。
- **`/particle` 的模式从不校验**：只有字面量 `force` 特殊，其它词当 normal。

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

## 8. 消息里的选择器 [验][读]

`/say` 和 `/me` 会解析消息里的选择器：

```
say @a          → 展开成匹配到的实体名
say @a（无匹配） → Selector '@a' found nothing
```

细节（`CommandAbstract.b`）：

- 消息**按空格拆成 token**，逐个尝试展开，然后**用单个空格重新拼回去**
  ——所以消息里的多余空格会被规范化掉。
- 展开成的是**所有**匹配实体的名字，用 `, ` 连接（不是只取第一个）。
- 名字取的是计分板显示名（自定义名 → 玩家名 → 类型的翻译名）。
- `/say` **总是**展开；`/me` **只在执行者不是玩家时**展开（玩家发 `/me` 不会
  把选择器展开成名字列表）。

---

## 9. 错误文本

1.12.2 的错误消息来自语言文件，gomc 照抄原文：

| 场景 | 文本 | 语言键 |
|---|---|---|
| 选择器没匹配到 | `Selector '@e[type=zombie]' found nothing` | `commands.generic.selector.notFound` |
| 玩家名不在 | `Player 'Steve' cannot be found` | `commands.generic.player.notFound` |
| 实体目标找不到 | `Entity 'Steve' cannot be found` | `commands.generic.entity.notFound` |
| 选择器参数不认识 | `Invalid selector argument: 'nbt={...}'` | `commands.generic.selector_argument` |
| 数字参数不合法 | `'abc' is not a valid number` | `commands.generic.num.invalid` |
| 布尔参数不合法 | `'yes' is not true or false` | `commands.generic.boolean.invalid` |
| 没指定玩家 | `You must specify which player you wish to perform this action on.` | `commands.generic.player.unspecified` |
| 目标只读 | `The objective 'health' is read-only and cannot be set` | `commands.scoreboard.objectiveReadOnly` |
| 目标不存在 | `No objective was found by the name 'counter'` | `commands.scoreboard.objectiveNotFound` |
| 持有者在该目标下没有分数 | `No counter score for Steve found` | `commands.scoreboard.players.test.notFound` |
| 实体没有那个标签 | `Entity didn't have the w tag` | `commands.scoreboard.players.tag.notFound` |
| gamerule 不存在 | `No game rule called 'foo' is available` | `commands.gamerule.norule` |
| 函数不存在 | `Unknown function 'minecraft:mcns'` | `commands.function.unknown` |
| 函数的 if/unless 条件不满足 | `Skipped execution of function 'probe:mcns'` | `commands.function.skipped` |
| `execute` 的 detect 没过 | `Failed to execute 'detect' as <实体名>` | `commands.execute.failed` |
| 子命令一次都没成功 | `All invocations failed: '<命令>'` | `commands.execute.allInvocationsFailed` |

选择器没匹配到这条是**分发层**抛的，所以它不只出现在 `execute`：凡是
`isListStart` 为真的命令（execute、kill、give、effect、tp、title、tell、……）
表现都一样。玩家名那条则是命令自己抛的。

引号里包的是**用户写的原文**：

- `Unknown function '...'` 里是**规范化之后**的 id——`MinecraftKey` 会把命名空间
  和路径都小写化，没写命名空间时补 `minecraft:`。所以 `function mcns` 找不到时
  报的是 `Unknown function 'minecraft:mcns'`。
- `Invalid selector argument: '...'` 里是**逗号分隔的那一段**
  （`@e[nbt={...}]` 报的是 `'nbt={...}'`），不是 key 本身，也不是整个 `@e[...]`。
- `Skipped execution of function '...'` 里也是规范化之后的 id。

`No game rule called '%s' is available` 在 1.12.2 里**只在查询分支**抛
（`gamerule <规则>` 不带值、而规则不存在时）。设置分支
（`gamerule <规则> <值>`）不查存在性——`GameRules.set` 遇到未知规则会**把它加进去**
（类型 `ANY_VALUE`）。gomc 的设置分支目前也用这条报错，见第 10 节。

> 消息里引号包的是**用户写的原文**，所以 `@e[tag=a,tag=b]` 报的是
> `Selector '@e[tag=a,tag=b]' found nothing`（保留重复参数，不折叠）。

---

## 10. 回显文案 [验][读]

命令的回显文字全部来自语言文件 `en_US.lang`（在服务端 jar 的
`assets/minecraft/lang/` 下），gomc 逐条对齐：

| 命令 | 回显 |
|---|---|
| `summon` | `Object successfully summoned` |
| `kill` | `Killed <实体名>` |
| `give` | `Given <物品> * <数量> to <玩家名>` |
| `setblock` | `Block placed` |
| `fill` | `<n> blocks filled` |
| `clone` | `<n> blocks cloned` |
| `testforblock` | `Successfully found the block at <x>, <y>, <z>` |
| `testfor` | `Found <实体名>` |
| `entitydata` | `Entity data updated to: <NBT>` |
| `tp` / `teleport` | `Teleported <实体名> to <x>, <y>, <z>` |
| `time set` / `add` / `query` | `Set the time to <n>` / `Added <n> to the time` / `Time is <n>` |
| `weather` | `Changing to clear weather` / `Changing to rainy weather` / `Changing to rain and thunder` |
| `difficulty` | `Set game difficulty to <难度>` |
| `gamerule` | `Game rule <名> has been updated to <值>`；查询是 `<名> = <值>` |
| `setworldspawn` | `Set the world spawn point to (<x>, <y>, <z>)` |
| `worldborder` | `World border is currently <n> blocks wide`、`Set world border to <n> blocks wide (from <m> blocks)` 等 6 条 |
| `particle` | `Playing effect <名> for <n> times` |
| `list` | `There are <n>/<max> players online:` |
| `function` | `Executed <n> command(s) from function '<规范化 id>'` |
| `debug` | `Started debug profiling` / `Stopped debug profiling after <n> seconds (<n> ticks)` |
| `title` | `Title command successfully executed` |
| `playsound` | `Played sound '<名>' to <目标>` |
| `stopsound` | `Stopped sound '<名>' with source '<源>' for <目标>` |
| `gamemode` | `Set own game mode to <模式>` / `Set <谁>'s game mode to <模式>`（模式名带 " Mode" 后缀） |

消息里的**实体名**一律是**翻译名**（"Armor Stand"、"Zombie Pigman"），
不是注册名（`armor_stand`）——见 3.9。`function` 的名字是**规范化之后**的 id
（命名空间和路径都小写化，见 1.1）。

**还没对齐的**（下一批）：

- **实体重力**：真服的实体会落到地面上（见 2.5），所以"实体的位置"通常是落地后的
  位置，而不是 `/summon` 时给的那个。gomc 明确**不模拟**重力（见第 11 节），
  `/summon` 的实体留在给定坐标上——凡是依赖"落地后位置"的相对坐标，两边会差一格。
- **实体 UUID**：真服用 UUID 指定非玩家实体，计分板持有者也是 UUID
  （`scoreboard players tag <UUID> ...`）。gomc 用内部 id（`entity-N`）代替，
  这类写法还不支持（见第 12 节）。
- **需要玩家在线的命令**（give / effect / title / tp / tell / xp / clear / enchant /
  trigger / gamemode / spawnpoint）：真服在控制台执行时的报错文本与 gomc 不同。
- **玩家专属命令的目标**：1.12.2 里下面这些命令的目标用
  `CommandAbstract.b`（**EntityPlayer** 变体）解析，**只接受玩家**：
  `give` / `gamemode` / `title` / `spawnpoint` / `tellraw` / `playsound` /
  `stopsound` / `clear` / `trigger` / `tp`（`/teleport` 才是实体版）。
  gomc 用的是通用的目标解析，**实体也能当目标**——方向是 gomc 更宽松：
  真服会报 `Player '...' cannot be found`，gomc 则照做。
  > 这条没改是有原因的：gomc 的世界里**没有"在线玩家"这个概念**，
  > 改成"只接受玩家"会让这些命令在 gomc 里几乎永远报错，
  > 而它们本来就没什么可操作的对象。等 gomc 建模了玩家再说。
- **Paper 私有扩展：函数里的管理命令被静默忽略**。Paper 给
  `CommandBlockListenerAbstract.executeCommand` 打了补丁：从命令方块或
  **函数**里执行命令时走的是
  `CustomFunction$b` → `CommandBlockListenerAbstract.executeSafely` →
  `executeCommand`，而这条路径上，下面的命令**直接 `return 0`**——不执行、
  也不给任何提示：
  `stop` / `kick` / `op` / `deop` / `ban` / `ban-ip` / `pardon` /
  `pardon-ip` / `reload`。
  > 所以真服（Paper 1.12.2）上，函数里的 `reload` 是**被静默忽略**的：
  > 日志里什么也没有，函数也不会重载。
  > 原版 1.12.2 没有这个补丁，走的是权限检查，函数里的 `reload` 会报
  > `You do not have permission to use this command`（见 1.3）。
  > gomc 现在按**原版**处理：权限不足就报错。

---

## 11. 明确不做的 [决策]

- **不模拟重力**。真服的实体有重力，`summon` 之后会落到地面上，所以"实体的位置"
  通常是落地后的位置。gomc 的 `/summon` 把实体留在给定坐标上，依赖落地位置的
  相对坐标两边会差一格。
  这是有意为之：重力属于实体模拟，而 gomc 的定位是**确定性的命令解释器**，
  不模拟实体行为。

---

## 12. 还没搞清楚的

| 问题 | 现状 |
|---|---|
| 实体 UUID 与计分板持有者的对应 | 玩家用玩家名 ✓；非玩家实体用内部 id（`entity-N`），真服用 UUID |
| 实体落地后的精确位置 | 真服有重力，gomc 不模拟（见 2.5） |
| 函数内命令失败何时报错 | 真实服务端有的报有的不报，gomc 统一报 |

### 为什么不建模实体 UUID

1.12.2 里，计分板持有者的标识是：玩家用**玩家名**，非玩家实体用**实体的 UUID**
（`Entity.bn()` 返回 `uniqueID.toString()`）。gomc 现在玩家用玩家名 ✓，
非玩家实体用内部 id（`entity-N`）。

所以 `scoreboard players set <uuid> obj 5` 这种**硬编码 UUID** 的写法在 gomc 上
认不出来。**但实际影响很小**：UUID 不可读，mcfunction 里没人这么写，作者写的是
选择器（`@e[type=zombie]`），那条路 gomc 是通的。

补齐的代价也不低：1.12.2 的 UUID 是 `UUID.randomUUID()` **随机生成**的，
照抄会让 difftest 的 golden 每次运行都漂；改成确定性 UUID（从实体 id 派生）
虽然可复现，但和真服不是一回事——真服的 UUID 本来就不可预测，
"对齐"出来的只是另一个假 UUID。

**决定**：保持现状，记在这里。要用 UUID 指定非玩家实体时用选择器。
