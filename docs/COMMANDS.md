# 命令支持矩阵

1.12.2 的完整命令表共 **65 条**（权威来源：服务端 jar 里的
`CommandDispatcher` 构造函数——1.12.2 的原版命令和服务器专用命令都在这里注册），
gomc 目前实现了 **46 条**。

矩阵里的"1.12.2 语法"一栏是 1.12.2 **真实**的语法，不是 1.13+ 的。
1.13 重写了命令系统，`execute`、`scoreboard`、`time`、`fill` 等命令的参数
都变了，直接照搬现代 Wiki 会写出在 1.12.2 上根本跑不出来的函数。

## 已实现（46 条）

| 命令 | 1.12.2 语法 | 备注 |
|---|---|---|
| `blockdata` | `/blockdata <x> <y> <z> <nbt>` | 合并进方块实体的 NBT |
| `clear` | `/clear [<目标>] [<物品> [<数据值> [<最大数量>]] [<nbt>]]` | 目标缺省为执行者 |
| `clone` | `/clone <x1> <y1> <z1> <x2> <y2> <z2> <x3> <y3> <z3> [<掩码>] [<克隆模式>]` | 掩码在第 10 位、克隆模式在第 11 位 |
| `debug` | `/debug <start\|stop>` | 只报告状态，不产出调试报告 |
| `defaultgamemode` | `/defaultgamemode <模式>` | |
| `difficulty` | `/difficulty <难度>` | 全名、单字母缩写、数字 id 都接受 |
| `effect` | `/effect <clear\|give> ...` | |
| `enchant` | `/enchant <目标> <附魔> [<等级>]` | 等级缺省为 1 |
| `entitydata` | `/entitydata <目标> <nbt>` | |
| `execute` | `/execute <实体> <x> <y> <z> [detect <x2> <y2> <z2> <方块> <数据值>] <命令>` | **1.12.2 旧语法**，没有 `as`/`at`/`run` |
| `fill` | `/fill <x1> <y1> <z1> <x2> <y2> <z2> <方块> [<数据值>] [<模式> [<替换方块> [<替换数据值>]]]` | 模式在第 9 位 |
| `function` | `/function <id> [if\|unless <目标>]` | 1.12.2 特有的条件形式 |
| `gamemode` | `/gamemode <模式> [<目标>]` | |
| `gamerule` | `/gamerule <规则> [<值>]` | 24 条规则，见 SPEC |
| `give` | `/give <目标> <物品> [<数量>] [<数据值>] [<nbt>]` | |
| `help` | `/help [<命令>]` | |
| `kill` | `/kill [<目标>]` | |
| `list` | `/list` | |
| `me` | `/me <动作>` | 消息里的选择器会被解析 |
| `particle` | `/particle <名称> <x> <y> <z> <dx> <dy> <dz> [<速度>] [<数量>] [<force\|normal>] [<观察者>]` | |
| `playsound` | `/playsound <音效> <声音源> <目标> [<x> <y> <z>] [<音量>] [<音调>] [<最小音量>]` | **目标在声音源之后** |
| `reload` | `/reload` | 清空注册表后重载，坏函数只记日志 |
| `replaceitem` | `/replaceitem <entity\|block> ...` | 实体形式：`entity <目标> <槽位> <物品> [<数量>]` |
| `say` | `/say <消息>` | 消息里的选择器会被解析 |
| `scoreboard` | `/scoreboard <objectives\|players\|teams> ...` | 三棵子树 |
| `seed` | `/seed` | |
| `setblock` | `/setblock <x> <y> <z> <方块> [<数据值>] [<模式>] [<nbt>]` | 模式在第 6 位 |
| `setidletimeout` | `/setidletimeout <分钟>` | |
| `setworldspawn` | `/setworldspawn [<x> <y> <z>]` | |
| `spawnpoint` | `/spawnpoint [<目标>] [<x> <y> <z>]` | |
| `stopsound` | `/stopsound <目标> [<声音源>] [<音效>]` | **目标在最前** |
| `summon` | `/summon <实体> [<x> <y> <z>] [<nbt>]` | |
| `teleport` | `/teleport <目标> <x> <y> <z>` | |
| `tell` | `/tell <目标> <消息>` | |
| `tellraw` | `/tellraw <目标> <原始 JSON>` | |
| `testfor` | `/testfor <目标> [<nbt>]` | NBT 走子集匹配 |
| `testforblock` | `/testforblock <x> <y> <z> <方块> [<数据值>]` | |
| `testforblocks` | `/testforblocks <x1> <y1> <z1> <x2> <y2> <z2> <方块> [<数据值>]` | |
| `time` | `/time <set\|add\|query> <值>` | `set` 认 `day`(1000)/`night`(13000)，`query` 认 `daytime`/`day`/`gametime` |
| `title` | `/title <目标> <title\|subtitle\|actionbar\|clear\|reset\|times> [<参数>]` | |
| `toggledownfall` | `/toggledownfall` | |
| `tp` | `/tp <目标> <x> <y> <z>` | |
| `trigger` | `/trigger <目标> [<值>]` | |
| `weather` | `/weather <clear\|rain\|thunder> [<时长>]` | 时长单位是秒 |
| `worldborder` | `/worldborder <set\|add\|center\|damage\|warning\|get> [...]` | |
| `xp` | `/xp <数量> [<玩家>]` | 数量带 `L`/`l` 后缀表示经验等级 |

## 未实现（19 条）

分两类：**服务端管理类**对测试 mcfunction 没有意义，**内容类**需要本项目没有建模的
注册表（配方、进度、数据结构）。

| 命令 | 为什么不实现 |
|---|---|
| `advancement` | 需要进度注册表，本项目不建模进度 |
| `ban` / `ban-ip` / `banlist` / `pardon` / `pardon-ip` | 服务端管理类 |
| `deop` / `op` | 服务端管理类 |
| `kick` | 服务端管理类 |
| `locate` | 需要生物群系与结构注册表 |
| `publish` | 服务端管理类 |
| `recipe` | 需要配方注册表 |
| `save-all` / `save-off` / `save-on` | 服务端管理类 |
| `spreadplayers` | 需要生物群系数据 |
| `stats` | 需要统计注册表 |
| `stop` | 服务端管理类 |
| `whitelist` | 服务端管理类 |

## 与 1.12.2 的已知偏差

除了"未实现"之外，还有几处 gomc **有意**偏离 1.12.2 的地方：

| 地方 | 1.12.2 | gomc | 为什么 |
|---|---|---|---|
| 实体标识 | 非玩家实体用 UUID 当计分板持有者 | 用 `entity-N` 这类内部 id | 本项目不建模 UUID |
| `/scoreboard players list` 的输出 | 发给玩家，不打到控制台 | 打到控制台 | 便于对拍 |
| 函数内命令失败 | 有的报有的不报 | 统一报在 `error[Runtime]` | 便于定位 |
| `sendCommandFeedback` gamerule | 控制命令回显 | 控制 CLI 的回显输出 | 便于对拍 |
| 实体 AI / 寻路 / 重力 | 模拟 | 不模拟 | 明确不在范围内 |

对拍时这些差异会体现在 golden 里，属于**已知且可解释**的偏差。

## 权威清单在哪

```
.ref\src\CommandDispatcher.java     # 1.12.2 全部 65 条命令的注册顺序
```

反编译一份新的服务端 jar 后，这个文件就是最权威的命令清单。
