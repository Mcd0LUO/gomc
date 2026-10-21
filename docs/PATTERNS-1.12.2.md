# 1.12.2 的工程手法

这份文档记录**真实工程包**里那些在 1.12.2 的限制下想出来的做法，以及
gomc 为了能跑它们还缺什么。

素材来自四个真实的世界存档（作者都是国内的 mcfunction 作者），全部用
`gomc check` 扫过：

| 工程 | 函数数 | 行数 | `gomc check` |
|---|---|---|---|
| `PFS-噬魂劫` | 14047 | 181521 | 89 条 |
| `悟空梦魂旅` | 10140 | 302375 | 19 条 |
| `末法` | 1963 | 11723 | 19 条 |
| `旧末法` | 519 | 6486 | 17 条 |
| `ZX2-3 8.23` | 275 | 5708 | **0 条** |

合计 **26944 个函数、507813 行**。这五个体量差两个数量级的包，是目前
gomc 唯一能拿到的「真实 1.12.2 工程语料」。

文档里每条手法都带**包 + 文件 + 行号**，可以自己核对。

---

## 1. 计分板就是那台计算机

1.12.2 没有 datapack、没有宏、没有 `/data`、没有 `execute as/at/run`、
没有 `execute if`。于是这些包把计分板用成了**变量、寄存器、条件分支和
函数参数**：

| 1.13+ 想要的 | 1.12.2 的替代 | 出处 |
|---|---|---|
| 全局变量 | **假玩家名**（`world_公用_100`、`#math.1000`、`compute`） | 悟空 `DreamRPG/compute.mcfunction:3-15`；末法 `luo/math/ini.mcfunction:9-16`；PFS `SYSTEM/math/创建计分板.mcfunction:20-52` |
| 数字字面量运算 | **假玩家常量表**（1…10000 预先塞成假玩家，`*= 100 compute`） | 悟空 `DreamRPG/compute.mcfunction:3-15` |
| `min` / `max` | **`<` / `>`**（1.12.2 的 keep-min / keep-max） | 悟空 `DreamRPG/system/loadAttribute.mcfunction:84`；PFS `API/取等级.mcfunction:2-4` |
| `clamp` | **选择器门控**：`scoreboard players set @s[score_x=-1] x 0` | 末法 `rpg/player/hurt/main.mcfunction:5-7` |
| `matches a..b` | **`score_x_min=a` + `score_x=b`** 同一选择器 AND | PFS `SYSTEM/help/main.mcfunction:11,14` |
| 条件分支 | **`function <id> if <selector>`**（1.12.2 独有，1.13 删除） | 末法 1851 处、悟空 9178 处 |
| 函数参数 | **`arg1` / `arg2` / `return`** 三个约定 objective | 末法 `luo/math/ini.mcfunction:3-5`（新旧两版一致，是个 ABI） |
| 事件总线 | **`trigger` objective + 数字事件码** + 每 tick `enable` 重新上膛 | 末法 `main/timer/cycle.mcfunction:3-8`；ZX2-3 `system/menu/cd.mcfunction:24-25` |
| 条件原语 | **`detect`**（方块判定） | 悟空 `DreamRPG/boss/食美鱼/被动.mcfunction:4`；PFS 1093 处 |
| 布尔量 | **tag**（`tag=!x` 就是「不在某状态」） | PFS `GIVE/items/附魔书/剧毒1.mcfunction:2` |
| 定时器 | **盔甲架 + 计分板**分频链（0.1s/0.2s/0.3s/1s/3s/5s） | PFS `MAIN/0.1s.mcfunction:3-9`；ZX2-3 `main/main.mcfunction:1-3` |
| 随机数 | **纯 32 位 LCG**（`%` 模运算 + Java int 回绕） | PFS `SYSTEM/math/随机数/随机数.mcfunction:5-21`；ZX2-3 `LightRPG/system/random/main.mcfunction:16-47` |
| 三角函数 | **360 个单行函数**（`tp_y_45` = `~-0.35355 ~0 ~0.35355`） | PFS `SYSTEM/RPG/枪/指向/y/tp_y_45.mcfunction:1` + `SYSTEM/math/视角计算/取指向ID.mcfunction:1-361` |
| 读玩家坐标 | **世界边界当比较器** + 26 步减半二分 | ZX2-3 `sethome/pos.mcfunction:3-6,83-87`；末法 `luo/easyCoord/LogCoord.mcfunction:10-13` |

> **`execute <实体> ~ ~ ~` 是 1.13 的 `as` + `at` 合体**：一条命令同时换
> 执行者和坐标基准，悟空的 208790 行 execute **100% 都带相对坐标**。

---

## 2. 没有 `/data` 时，拿什么当存储

| 1.13+ 想要的 | 1.12.2 的替代 | 出处 |
|---|---|---|
| `nbt={}` 选择器 | **`scoreboard players tag/set/add/remove` 的 dataTag 过滤** | PFS `MAIN/1s.mcfunction:140`（`{OnGround:1b}` 落地检测）、`MAIN/MAIN.mcfunction:31`（`{SelectedItem:…}` 手持物检测） |
| `/data get/set` | **`entitydata` 深层合并** | 末法 `rpg/api/Bind/uid/8192.mcfunction:5-13` |
| `/item` | **`replaceitem` + `slot.container.N`** | ZX2-3 `chestmenu/GUI/shop_L1.mcfunction:4` |
| 物品组件 | **`display.Name/Lore`（人读）+ `Tags[]`（机器读）** | PFS `GIVE/items/附魔书/剧毒1.mcfunction:1`；ZX2-3 `chestmenu/GUI/shop_L1.mcfunction:4` |
| 存数字到物品 | **NBT 位图**（1 bit = 1 个键，`%=2` 取位、`/=2` 右移） | 悟空 `DreamRPG/system/math.mcfunction:8`；末法 `rpg/api/Bind/uid/8192.mcfunction:5-13` |
| `/data` 存全局 | **假人（armor stand）当寄存器** | 悟空 `DreamRPG/particle/13_翅膀/h3/gen.mcfunction:6,28,41` |
| 文本显示屏 | **10 个盔甲架当数码管**（改 `CustomName`） | 旧末法 `sys/api/数显/asg_1.mcfunction:1-13` |

> **最惊艳的一条**：`scoreboard players set @s return 1 {SelectedItem:{Count:1b}}`
> 连写 65 行，**只有与玩家当前选中物品堆叠数吻合的那一条会写成功**——
> 「最后一次成功写入即答案」。1.13 删掉了 dataTag 参数，1.13.2 才有
> `execute if items`。（末法 `rpg/api/get/GetSelectItemCount.mcfunction:1-65`）

---

## 3. 没有 `/stats` 就没有「命令返回值 → 记分板」

1.12.2 的 `/stats entity <目标> set <结果名> <holder> <objective>` **不是
回显**，它是**注册一条通道**：统计量变化时，值被写进那个 objective。

真服语义（反编译 `CommandObjectiveExecutor`）：

```
stats entity @s set QueryResult @s return   ← 绑定：QueryResult → "@s" 的 return 分
scoreboard players set @s return 0          ← 条目必须已存在，否则放弃写入
time query gametime                         ← 产生 QueryResult
stats entity @s clear QueryResult @s return ← 解绑
```

（末法 `luo/getter/get_server_time.mcfunction:1-4`）

同理 `AffectedItems` 通道用来数 `/clear` 实际清掉了几个（末法
`rpg/api/item/stat_start.mcfunction:1` + `stat_end.mcfunction:1` 成对使用；
ZX2-3 的**整个箱子菜单点击检测**也押在这条链上）。

> 1.13+ 只能靠 datapack 的 `store`。

---

## 4. 表现层

| 手法 | 出处 |
|---|---|
| **`{"score":{…}}` 活组件**：HUD 自动跟随数据，不用每 tick 拼字符串 | 末法 `rpg/user/TitleActionBar.mcfunction:2`；ZX2-3 `main/10tick.mcfunction:10` |
| **100 个队伍名当血条**（队伍名显示在玩家名前缀） | ZX2-3 `init/team_hpBar.mcfunction:2-101` |
| **11 档预写好的 `▌` 段落**做渐变血条 | PFS `SYSTEM/通用/状态栏/法师/法师_ok.mcfunction:6-17` |
| **离线烘焙的轨迹表**（24~210 个点写死） | 末法 `rpg/particle/tp/elliptic_ns@reddust.mcfunction:1-3` |
| **聊天栏当画布**（连续 `tellraw` 顶位置） | 悟空 `DreamRPG/jj/meun/go.mcfunction:25-29` |
| **血条格数累计余量**（按帧节流的表现更新） | 末法 `rpg/monster/api/UpdateBar.mcfunction:6-17` |

---

## 5. 工程规范（ZX2-3 为什么能零错误通过检查）

ZX2-3 是唯一**静态检查零问题**的包。它的做法值得抄：

1. **只写真·1.12.2 语法**：`scores={}` / `nbt={}` / `execute if score` /
   `as-at-run` 各 0 处。这一条就消灭了 1.12.2 工程最常见的失败模式
   （照搬 1.13 Wiki）。
2. **参数顺序全部符合 1.12.2**：`effect` 目标在前、`clear` 的 `clear` 在
   第二段、`particle` 速度必填、`trigger` 模式必填、`detect` 数据值必填。
3. **命名空间即子系统，门面统一叫 `main.mcfunction`**，父级只调门面
   → 调用图是一棵树，没有环。
4. **机械重复被模板化**：15 个同构的 API 文件、100 行同构的队伍表、
   27 行同构的菜单占位——复制粘贴的错误率被「同构」消化。
5. **参数化到极致**：所有魔数抽成 objective，计分板成为唯一的状态中枢。

> ⚠️ 零静态错误 ≠ 语义正确：ZX2-3 的 `stats` 链整体未经验证，而
> `API/getAttribute/*.mcfunction` 的 15 个文件是机械复制的——**位序写错
> 静态检查发现不了**。作者的更新日志里就写着「修正掉率异常」。

---

## 6. gomc 的缺口（按价值排序）

三个子代理**独立地**指向了同一个缺口（互证）：

### ① `/stats` 的结果捕获语义 ★★★★（三个代理都点了）

gomc 的 `/stats entity set` 只回显，**从不写 objective**，`world.Entity`
也没有统计量字段。后果：上面第 3 节那整类「命令返回值 → 记分板」的函数
在 gomc 里**全是死的**（末法的 `get_server_time`、旧末法 `LogCoord` 的
Y 轴、ZX2-3 的整个箱子菜单点击检测、PFS 的 `stats` 通道）。

要补的东西：
1. 每个实体挂 5 个「命令结果通道」；
2. `stats entity <t> set <结果> <holder> <objective>` 建立绑定、`clear` 解除；
3. `/clear`、`/fill`、`/testforblocks`、`/time query`、`/execute` 等命令
   上报 AffectedItems / AffectedBlocks / QueryResult / SuccessCount；
4. 写入前检查计分板条目已存在（真服 `Scoreboard.b` 的行为）。

### ② `/tp` 的 `[<yaw> <pitch>]` ★★★

用法串承诺了，实现只解析 x/y/z，**多出来的参数被丢弃**，传送时也不改朝向。
1.12.2 的 `/tp` 有这个参数（1.13+ 反而没有），末法用它做「转向送回」
（`luo/detect/isTpSafe.mcfunction:5`）。而悟空有 74526 行 `ry/rym` 选择器
依赖实体的旋转——**旋转不落到实体，这些选择器永远选不中**。

### ③ `replaceitem` 的槽位模型不全 ★★★

只认 `slot.weapon.*`、`slot.hotbar.0-8`、`slot.inventory.0-8`、裸数字，
其余报 unknown slot。真实工程用的是 `slot.inventory.18/19/20`、
`slot.enderchest.0..26`、`slot.armor.*`、`slot.container.0..26`。

### ④ `summon` / `entitydata` 的 `Pose` / `Rotation` 不落到实体旋转 ★★

走路动画、朝向全丢。

### ⑤ `/particle` 的尾部 int 参数不校验 ★★

1.12.2 的尾部整数是**粒子自带的参数**（`blockcrack` 的方块数据值），
个数由粒子决定，多一个就报 `Invalid parameters`。悟空有 **224 行**在真服上
必然失败，gomc 静默放过了。

### ⑥ trigger 没有锁标志 ★★

真服的 trigger objective 触发一次就锁死，必须 `enable` 才能再触发。
gomc 用「分数是否被跟踪」近似。

### ⑦ tellraw/title 的 JSON 组件没有建模 ★

`score`（含 `name:"*"` 聚合）、`clickEvent`、`hoverEvent`、`translate`+`with`、
`extra` 都没有解析与渲染。菜单的 `clickEvent` 完全不可验证。

### ⑧ `detect` 的「区块已加载」前置条件 ★（**已修**）

1.12.2 先 `World.isLoaded(detectPos)`，不过就报
`Failed to execute 'detect' as <实体>`（和方块对不上是同一条）。PFS 有 1093
条 detect 依赖它。**已在这一轮修好**（真服实测，探针用例
`54-scoreboard-limits`）。

### ⑨ 计分板目标/队伍的重名与长度检查 ★（**已修**）

真服：重复 `objectives add` 报 `An objective with the name 'X' already
exists`；名字 > 16 字符、显示名 > 32 字符各有各的错；目标类型不认识报
`Usage: Invalid objective criteria type 'X'`。PFS 的 964 个目标名里有大量
重复 add，**真服上第二次 add 就报错**。**已在这一轮修好**。

---

## 7. 用这些包做回归用例（还没做的）

这几条是现成的、**自带规格**的测试素材，优先级很高：

| 素材 | 断言什么 |
|---|---|
| 末法 `rpg/api/get/GetSelectItemCount.mcfunction` | dataTag 子集匹配的**黄金用例**：写进去的位域要能原样读出来 |
| 末法 `luo/getter/get_server_time.mcfunction` | `/stats` 通道（**要等缺口 ① 补完**） |
| 末法 `luo/math/sort/_low2high.mcfunction` | `<` 的 keep-min 语义 + 选择器当 operation 源 |
| ZX2-3 `sethome/pos.mcfunction` | 坐标边界的端到端：减半二分要收敛到 `playerX` |
| ZX2-3 `main/main.mcfunction` | **时序规格**：分频比 = 5/10/20 tick、1/3/5 s（gomc 现在最缺的测试素材类型） |
| 悟空 `DreamRPG/system/playerHurt.mcfunction` | 影子分数 + 尾部统一 reset 的「无状态管线」 |
| 悟空 `one/function/facing/facing.mcfunction` | 8 个临时实体 + 三层 execute + 位运算的朝向计算 |

---

## 8. 两条「作者自制语法」

不是 1.12.2 的命令，但值得记：

- 末法 `luo/init.mcfunction:8` 的 `$easyCoord.getRatioByScale()`——作者自制的
  **预处理器调用语法**，真服同样跑不了这一行。
- 末法 `luo/vec3/IsLookingAt.mcfunction:10,21,24,38` 的
  `#@macro start: $vec3.dotAB(@s,#out, return)` / `#@macro end:`——作者自制的
  **宏标记约定**，用 `#` 注释藏起来，靠外部工具展开。

> 「语法缺失就自己造一个」——这是 1.12.2 工程最典型的心态。
