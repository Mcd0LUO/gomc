# GoMcRuntime 架构

用 Go 实现的 Minecraft 1.12.2 服务端运行时，用于直接测试 mcfunction。

**相关文档**：[`SPEC-1.12.2.md`](SPEC-1.12.2.md)（行为基准）·
[`COMMANDS.md`](COMMANDS.md)（命令矩阵）·[`DIFFTEST.md`](DIFFTEST.md)（对拍层）·
[`DEV.md`](DEV.md)（开发指南）

## 目标与非目标

**目标**：在无 JVM 的情况下，执行 1.12.2 的 mcfunction，并给出与真实服务端一致的
可观察行为（命令成功/失败、scoreboard、实体、方块、函数递归、tick 循环）。

**非目标**：不模拟 AI、寻路、重力物理、区块生成、渲染、网络协议。

## 依赖方向（严格单向，禁止反向 import）

```
                        cmd/gomc            CLI 入口
                            |
                            v
              version/v1_12_2              1.12.2 方言：命令表、execute 语法、坐标规则、selector
                            |
                            v
                    exec                 执行引擎：上下文、函数栈、tick
                    |    \
                    v     v
             world      ast/lexer/parser   世界状态 / 词法 / 语法 / AST
               |              |
               v              v
              nbt                          NBT 值模型 + SNBT + 二进制
```

规则：

- `nbt`、`world` 不得 import `exec`。
- `exec` 不得 import `version`；版本差异通过接口/注册表注入。
- `version/v1_12_2` 依赖 `exec` 的通用能力，但**不复用**其版本相关实现。
- 任何包不得 import `cmd/gomc`。

## 包职责

| 包 | 职责 | 关键类型 |
|---|---|---|
| `internal/nbt` | NBT 值模型、SNBT 解析、二进制编解码、子集匹配 | `Value`, `Compound`, `ParseSNBT` |
| `internal/lexer` | 把一行命令切成 token（1.12.2 的规则：**按空格切**，引号和花括号都不是分组符） | `Token`, `Tokenize` |
| `internal/ast` | 命令的抽象语法树节点 | `Node` 家族 |
| `internal/parser` | token -> AST，带列号错误（**版本无关的原语**） | `Parser`, `Error` |
| `internal/world` | 世界状态：方块、实体、玩家、scoreboard、tag、gamerule | `World`, `Scoreboard` |
| `internal/exec` | 执行引擎：上下文栈、函数调用栈、tick 循环 | `Engine`, `Context` |
| `internal/version/v1_12_2` | 1.12.2 语义：命令实现、execute 旧语法、坐标规则、selector 语法与语义 | `Register` |
| `internal/difftest` | 对拍层：用例、观察记录、golden 比对 | `Case`, `Observation`, `Diff` |
| `internal/cli` | CLI 子命令分发 | `Run` |
| `cmd/gomc` | CLI | `main` |

## 核心不变量（反编译得到，并经真实服务端实测校验）

这些是本项目与"通用 MC 模拟器"的分歧点，必须逐条对齐：

1. **命令切分**：`split(" ", 2)` —— 只按第一个空格切，剩余部分整体保留。
2. **函数存储**：`<world>/data/functions/<ns>/<path>.mcfunction`，非 1.13+ 的 datapack 布局。
3. **函数加载**：某行命令名未注册会导致**整个函数加载失败**（抛 `IllegalArgumentException`），
   而非仅跳过该行。
4. **函数注释**：`#` 开头跳过；`//` 开头报错（明确提示应用 `#`）。
5. **函数执行**：命令逆序压栈、正序弹出（`CustomFunctionData.a` 的 `for j = len-1; j >= 0; --j`）。
6. **链长上限**：`maxCommandChainLength` gamerule，默认 65536，超出即中断。
7. **tick 驱动**：`gameLoopFunction` gamerule 指向的函数每 tick 执行一次；
   该值变化时重新解析函数引用。
8. **execute 旧语法**：`execute <entity> <x> <y> <z> [detect <x2> <y2> <z2> <block> <dataValue>] <cmd>`。
   坐标相对于**目标实体**而非执行者；`detect` 的相对坐标以**执行位置**
   （实体位置 + 偏移）为基准，偏移**只加一次**。无 `as`/`at`/`run`。
9. **execute 多实体**：`execute <selector> ...` 对**每个匹配实体各执行一次**。
   展开发生在**命令分发层**（`CommandHandler`），不在 execute 内部。
   顺序 = 选择器返回的顺序 = **距离升序**（`PlayerSelector` 按距离平方稳定排序），
   **不是倒序**。真实服务端实测：三个具名盔甲架（距离 5.10 / 6.40 / 9.43）→
   输出依次是 `[NEAR]` `[MID]` `[FAR]`。
10. **execute 失败传播**：一个实体失败**不影响其余实体**——分发层对每个实体单独
    try/catch，捕获后 `continue` 到下一个。失败文本是
    `Failed to execute 'detect' as <实体名>`。真实服务端实测：最近的实体 detect 失败，
    后面两个照样执行。
11. **execute 的执行者**：被选中的实体成为子命令的执行者，因此输出前缀是
    **目标实体的名字**（`[Armor Stand] hello`），而不是 `[Server]`。
    名字取法与 `Entity.getName` 一致：自定义名 → 玩家名 → 类型的**翻译名**
    （语言文件里的 `entity.<类名>.name`，所以是 "Armor Stand" 这种带空格的写法）。
12. **function 的 if/unless**：`function <id> if|unless <selector>` 存在于 1.12.2，
    语义是"该 selector 匹配到实体则执行"；名字找不到时条件为**假**而不是报错。
13. **selector 选项集**：`r rm l lm x y z dx dy dz rx rxm ry rym c m team name type tag`。
    1.12.2 **没有** `scores={...}`，分数过滤写作顶层参数 `score_<objective>`（上限）
    与 `score_<objective>_min`（下限）；也**没有** `nbt={...}`（Minecraft Wiki 变更日志：
    "1.13 17w45a 在目标选择器中加入了 nbt"）。`1.11` 起也移除了隐式位置参数。
14. **selector 参数是平坦 map**：`PlayerSelector.d` 用 `Splitter.on(',')` + `Splitter.on('=').limit(2)`
    解析，重复的参数名**后者覆盖前者**（`@e[tag=a,tag=b]` 等价于 `@e[tag=b]`）。
    参数列表内不允许出现空格（正则 `\[([^ ]*)\]`）。
15. **旋转上下限方向相反**：`rx`/`ry` 是**上限**，`rxm`/`rym` 是**下限**；
    角度取整并折到 `[-180,180)`；下界大于上界时区间**跨北环绕**。
16. **距离边界**：`r` 是上限、`rm` 是下限，与直觉一致；比较用平方距离。
17. **体积过滤用包围盒**：`dx`/`dy`/`dz` 选的是**包围盒与体积相交**的实体，
    不是位置点落在体积内的实体。体积是方块区间，远端多出 1 格。
18. **m= 只对玩家生效**：`m=` 读的是玩家的 interact manager 游戏模式，
    其他实体一律不匹配；支持 `!` 取反、全名、缩写（s/c/a/sp）与数字 id。
19. **@a/@p/@r 限玩家**：1.12.2 给 `@a`/`@p`/`@r` 加"必须是玩家"的过滤；
    `@r` 若带 `type=` 则由类型过滤**取代**该玩家过滤。`@e` 不加。
20. **c 的语义**：`@a`/`@e` 默认不限（0），`@p`/`@r`/`@s` 默认 1；
    负数会**先反转**整个列表再截断。
21. **分数是 Java int**：所有分数按 32 位有符号整数回绕，除法向零取整。
22. **NBT 匹配**：`/testfor` 与 `/entitydata` 用的是**子集匹配**语义。
23. **坏函数不阻断其他函数**：`CustomFunctionData.h` 对每个文件单独 try/catch，
    坏文件只记日志、不注册，其余函数照常加载。`/reload`（`CustomFunctionData.f`）
    会先清空注册表再重载。
24. **1.12.2 没有独立的 `/tag` 命令**：真实服务端对 `tag` 报
    "Unknown or invalid command 'tag'"，整个函数加载失败。实体标签用
    `scoreboard players tag <target> <add|remove|list> [tag]` 管理。
25. **`say`/`me` 会解析消息里的选择器**：`say @a` 在 1.12.2 里会展开成实体名，
    没有匹配时报 "Selector '@a' found nothing"。
26. **函数内的输出前缀不同**：`say` 是 `[Server] <消息>`，而命令结果
    （如 `scoreboard objectives add`）是 `[Server: <消息>]`。
27. **gamerule 清单是 24 条**（`GameRules` 的静态初始化块），
    含 `sendCommandFeedback` / `spectatorsGenerateChunks` / `announceAdvancements`，
    **不含** 1.13 才有的 `tntExplodes`。
28. **`/time` 的参数**：`set` 接受 `day`(1000) / `night`(13000) / 数字，
    `query` 接受 `daytime` / `day` / `gametime`；`daytime`/`noon`/`midnight`
    是 1.13 才有的写法。
29. **`/setblock` 与 `/fill` 的模式参数**：`setblock` 的模式在参数 6 位
    （`replace`/`destroy`/`keep`），`fill` 的模式在参数 9 位
    （`replace`/`destroy`/`keep`/`hollow`/`outline`），1.12.2 两者都有。
30. **实体的显示名是翻译名**：`Entity.getName` 走语言文件里的
    `entity.<去掉 Entity 前缀的类名>.name`，所以是 "Armor Stand" 这种带空格的
    写法，不是注册名 `armor_stand`；语言文件缺键时原样返回键本身。
31. **实体有重力**：`summon` 出来的实体会落到地面上（`tp` 不会）。
    gomc 明确不模拟重力，所以 `/summon` 的实体留在给定坐标上——
    依赖落地位置的相对坐标，两边会差一格。

## 权威规格的来源

行为规格不是靠记忆或 Wiki 猜的，而是从本机 Paper 1.12.2 的服务端 jar 里挖出来的：

```
D:\game\McServer\1.12.2\cache\patched_1.12.2.jar   # 含 2347 个 v1_12_R1 NMS 类的真 jar
.ref\tools\cfr.jar                                # CFR 0.152 反编译器
.ref\src\                                          # 178 个反编译 .java
.ref\mcdata\                                        # minecraft-data 的 1.12 数据
```

选择器的细节尤其关键：1.12.2 的 `PlayerSelector` 把每个过滤器都写成了匿名内部类，
CFR 无法还原，于是用 `javap -p -c` 直接读字节码才拿到精确语义
（`PlayerSelector$2` 距离、`$3`/`$4` 旋转、`$5` 体积、`$6` 排序、`$9` 等级、
`$10` 游戏模式、`$12` 分数、`$13` 名称、`$14` 标签）。

### 实测校验

反编译给出的是"读出来的"规格，**跑一遍真实服务端**给出的是"验出来的"规格。
两者不一致时以实测为准。做法是复制一份隔离的服务端目录（`libraries`/`cache`
用 junction 指回原目录，不动用户的存档），用 JDK 1.8 启动 `nogui`，
把命令喂给控制台并读日志。

实测直接推翻了几个反编译读出来的推断：

| 结论 | 来源 | 对应用例 |
|---|---|---|
| **1.12.2 没有 `/tag` 命令**，写 `tag` 的函数整个加载失败 | 服务端日志 `Unknown or invalid command 'tag'` | `01-tag-command` |
| `execute <selector>` 对**每个匹配实体各执行一次**，且按**距离升序**执行 | 服务端日志 | `02-execute-multi` |
| `execute` 的子命令输出前缀是**目标实体名**（`[Armor Stand] ...`） | 服务端日志 | `02-execute-multi` |
| `say @a` 会展开选择器，无匹配时报 `Selector '@a' found nothing` | 服务端日志 | `03-say-selector` |
| 一个实体 `detect` 失败**不影响其余实体**，失败文本是 `Failed to execute 'detect' as <实体>` | 服务端日志 | `04-detect-fail` |
| 函数内 `say` 是 `[Server] ...`，命令结果是 `[Server: ...]` | 服务端日志 | `05-command-feedback` |
| `time` / `gamerule` / `difficulty` 的回显文本 | 服务端日志 | `05-command-feedback` |
| `scoreboard players list` 的真实输出格式 | 服务端日志 | `06-scoreboard-list` |

这些用例在 [`tools/probe/`](../tools/probe/README.md) 里，可以随时重跑复核。

实测用的服务端是**隔离副本**：`libraries`/`cache` 用 junction 指回原目录，
存档另开一个**超平坦**世界，用户自己的存档与服务端进程都没有被碰到。

### 怎么复现一次实测

仓库里带了这个工具：[`tools/probe/`](../tools/probe/README.md)。

```powershell
cd tools\probe
.\probe.ps1 -Case .\cases\02-execute-multi.case          # 跑一个用例
.\probe.ps1 -Case .\cases\02-execute-multi.case -Compare # 和期望文件比对
.\run-all.ps1                                            # 跑完全部用例并比对
```

它替你做完了下面这些事：

1. 复制一份服务端目录到临时路径，只复制 jar 与配置文件，`libraries`/`cache`/
   `assets`/`lang` 用 junction 指回原目录（不要复制，更不要删 junction）。
2. 写一份 `server.properties`：`level-type=FLAT`、另开一个世界、
   `online-mode=false`、`port` 换一个、`spawn-protection=16`。
3. **从超平坦世界模板复制一份世界**。模板里已经固化了出生点 `(0,4,0)` 和 gamerule，
   所以**每个用例只起一次服**——1.12.2 只在启动时加载出生点区块，运行中改
   `/setworldspawn` 不会把新区块加载进来。
4. 在 `<world>/data/functions/<ns>/` 下放要测的 `.mcfunction`。
5. 用 **JDK 8**（`D:\tools\jdk\JDK_1.8`）启动：1.12.2 在 JDK 21 上跑不起来。
   把命令喂给控制台的标准输入并读标准输出即可。
6. 观察服务端日志里的 `[Server] ...` / `[Server: ...]` / `[实体名] ...` 三种前缀，
   这就是命令输出在真实服务端上的样子。

> 删临时目录前**先删掉那 4 个 junction**，否则可能误伤真实服务端。
> 脚本的清理函数已经按这个顺序做了。

## 错误模型

错误分两种类型：

- **解析期**（`internal/parser`）用 `*parser.Error`，携带：
  - `Line` / `Col`：1-based，指向 mcfunction 文件内位置
  - `Msg`：**1.12.2 会怎么说**（原文照抄，不改一个字）
  - `Kind`：`Syntax` | `Unknown` | `Runtime`
  - `Path` / `Source`：文件路径与出错的那一行
  - `Usage`：该命令的用法行（来自语言文件）
  - `Hint`：额外建议，比如 `did you mean 'setblock'?`
- **执行期**（`internal/exec`）用 `*exec.Error`，只带 `Msg`（和 `Usage`）：
  命令已经解析过了，失败发生在跑的时候，位置没有意义。

**消息本身永远是 1.12.2 的原文**，增强的部分都是附加的：位置、用法、建议。
这样既保持了"对着真服复现"，又能多指出真服不会指出的东西。

CLI 默认输出编译器风格（rustc/TS）：

```
error[Syntax]: data/functions/ns/f.mcfunction:12:5
  scoreboard players set @a obj = 5
      ^
  expected an integer, got "="
  usage: /scoreboard players set <player> <objective> <score> [dataTag]
  hint: did you mean 'setblock'?
```

`--plain-errors` 关掉增强，只留一行，给要解析输出的脚本用。
**但两类错误的行格式不一样**，脚本不能按一种格式吃全部错误：

| 错误种类 | `--plain-errors` 的输出 |
|---|---|
| 解析错误（`parser.Error`） | `error[Kind]: line:col: msg` —— 有行号**和列号**，**不带路径** |
| 函数加载错误（`exec.FunctionLoadError`） | `error[Unknown]: <路径>:<行>: <消息>` —— 带**路径**，**没有列号** |
| lint 错误（`exec.FunctionLintError`） | `error[Syntax]: <路径>:<行>: <消息>` |
| 运行期错误（`exec.Error`） | `error[Runtime]: <消息>` |

也就是说：解析错误看 `line:col`，函数加载错误看 `path:line`。
开关可以放在子命令前后的任何位置。

增强只发生在 CLI 的渲染层，**不进对拍 golden**：difftest 记录的错误仍然是
1.12.2 的原文。

## 扩展指引

- **加一条命令**：在 `version/v1_12_2` 下按域新建文件（如 `cmd_scoreboard.go`），
  实现 `exec.Handler` 接口（`Name` / `Usage` / `Parse` / `Run`），在该版本的注册表里登记。
不要新建独立顶层包。
- **加一个版本**：新建 `version/v1_13` 等，实现同样的注册入口；
  `exec` 层只暴露版本无关的原语。
- **加一类 selector 过滤**：在 `version/v1_12_2/selector_resolve.go` 增加 filter 构造。
  选择器的**规则**（选项名、每个选项的语义、结果的排序）都是版本特有的，
  所以它们住在版本包里；`exec` 只提供执行引擎的通用能力。
