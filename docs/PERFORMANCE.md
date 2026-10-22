# 性能评估

这份文档记录 gomc 的性能现状：**测了什么、数据是多少、瓶颈在哪、该怎么改**。
目的是让优化有据可依，而不是凭感觉改。

### pprof 说了什么

对 `RunFunction` 场景采样，平坦视角的 top（合计 2610 ms 样本）：

```
runtime.growslice                        19.16% cum   ← 切片反复增长
runtime.scanObject                      19.16% cum   ← GC 扫堆
runtime.mapaccess2_faststr               8.43% cum   ← map 查找
runtime.concatstrings                    4.21%
runtime.mallocgcSmallScanNoHeaderSC*     一堆         ← 大量小对象分配
```

一句话：**这个运行时的开销主要是"分配 + GC"，不是"算得慢"**。
`growslice` 和 `mapaccess2_faststr` 正是"每次选择器都重建实体切片"的签名；
GC 占比高说明命令线程上产生了太多本不该产生的对象。

所以优化的方向是**少分配**，而不是"换个更快的算法"：
缓存实体列表、结果切片别按候选数预分配、回显别在命令线程格式化。

## 结论先说

gomc 的命令执行本身**不慢**（每条 `say` 约 3.5 µs，一秒几十万条命令），
真正的开销集中在三处：

1. **命令被重复解析**——占每条命令约 **60%** 的时间。
2. **选择器每次都重建整个实体列表**——实体多时是主要成本。
3. **`scoreboard players list` 是 O(持有者 × 目标)**——目标一多就 quadratic。

其余的（回显字符串的分配、difftest 的快照、tick 循环）**量级上可以忽略**，
不要去优化它们。

## 怎么测的

两种手段，互相印证：

**端到端**：造一个只有函数的世界（gomc 的世界状态**不在存档里**，方块、实体、
分数全靠命令产生），然后用真实二进制计时。

```powershell
# 造一个 5000 条 setblock 的函数
.\gomc.exe run <world> demo:big
```

**微基准**：`internal/exec/bench_test.go` 里的 Go 基准，每个场景自建内存世界。

```powershell
go test ./internal/exec/... -bench . -benchmem -run '^$' -timeout 300s -p 1
```

> 所有 `go test` 都必须带 `-timeout`：这个项目的一条硬规矩，不许把机器拖死。

## 数据

机器：i9-14900HX，Windows amd64，Go 1.24。

### 端到端（`gomc.exe`，含进程启动约 8 ms）

| 场景 | 耗时 |
|---|---|
| `check`（3 个函数共 12000 行，加载 + lint） | 33 ms |
| 跑 5000 条 `setblock` | 30 ms |
| 跑 2000 次 `summon` | 22 ms |
| 跑 5000 条 `scoreboard players set` | 121 ms |
| 5000 个持有者后 `scoreboard players list` | 134 ms |
| **2000 个实体 + 2000 次带选择器的写分** | **835 ms** |
| （对照）2000 个实体 + 2000 次普通写分 | 81 ms |
| `tick 100`（每 tick 跑 50 条 `say`） | 18 ms |
| `difftest run` | 12 ms |

**对照组把范围锁死了**：同样 2000 次写分，用选择器 835 ms、不用 81 ms——
**754 ms 全在选择器上**。

### 微基准

选择器那几组是 2026-10-21 重测的（`-count=3` 取最小值，`-benchtime 300x`，
空载机器）；其余是早先测的，没有重跑。

```
BenchmarkBaselineSay/ExecuteLine       3534692 ns/op   4745987 B/op   90018 allocs/op
BenchmarkBaselineSay/RunFunction      7493582 ns/op   7400217 B/op   90038 allocs/op
BenchmarkBaselineSay/ExecuteParsed    2760157 ns/op   1908186 B/op   30000 allocs/op
BenchmarkParseOnly/Say                 4459020 ns/op   3120021 B/op   60000 allocs/op
BenchmarkSetblock/ExecuteLine        18062886 ns/op  13138929 B/op  130099 allocs/op
BenchmarkSummon/ExecuteLine          12036768 ns/op   7166032 B/op   84835 allocs/op
BenchmarkSelectorScan/Resolver      193618883 ns/op  32824149 B/op    4001 allocs/op
BenchmarkSelectorScan/Command      1153764500 ns/op 328581432 B/op  6037088 allocs/op
BenchmarkSelectorScanScaling/Entities100    3901 ns/op     2200 B/op      8 allocs/op
BenchmarkSelectorScanScaling/Entities500   13305 ns/op     9371 B/op     10 allocs/op
BenchmarkSelectorScanScaling/Entities2000  40322 ns/op    37273 B/op     12 allocs/op
BenchmarkSelectorScanScaling/Entities8000 151970 ns/op   135576 B/op     14 allocs/op
BenchmarkSelectorScanFewHits/Entities100    2867 ns/op       56 B/op      2 allocs/op
BenchmarkSelectorScanFewHits/Entities500    7209 ns/op       56 B/op      2 allocs/op
BenchmarkSelectorScanFewHits/Entities2000  24437 ns/op       56 B/op      2 allocs/op
BenchmarkSelectorScanFewHits/Entities8000 103093 ns/op       57 B/op      2 allocs/op
BenchmarkSelectorScanOneHit/Entities100     2801 ns/op       40 B/op      2 allocs/op
BenchmarkSelectorScanOneHit/Entities500     7295 ns/op       40 B/op      2 allocs/op
BenchmarkSelectorScanOneHit/Entities2000   24301 ns/op       40 B/op      2 allocs/op
BenchmarkSelectorScanOneHit/Entities8000  101571 ns/op       40 B/op      2 allocs/op
BenchmarkSelectorScanOneHitMany/Entities500  24997 ns/op     9325 B/op      9 allocs/op
BenchmarkSelectorScanOneHitMany/Entities2000 80046 ns/op    37226 B/op     11 allocs/op
BenchmarkSelectorScanOneHitMany/Entities8000 271731 ns/op  135528 B/op     13 allocs/op
BenchmarkSortThenFirst/Entities100        1075 ns/op       952 B/op      3 allocs/op
BenchmarkSortThenFirst/Entities1000      13283 ns/op      8248 B/op      3 allocs/op
BenchmarkSortThenFirst/Entities8000     161249 ns/op     65592 B/op      3 allocs/op
BenchmarkClosestOnly/Entities100           245 ns/op         0 B/op      0 allocs/op
BenchmarkClosestOnly/Entities1000         2492 ns/op         0 B/op      0 allocs/op
BenchmarkClosestOnly/Entities8000        19858 ns/op         0 B/op      0 allocs/op
BenchmarkScoreboardWrite/ExecuteLine 22724584 ns/op  12609612 B/op  170137 allocs/op
BenchmarkScoreboardList/SelectorHolders/Objectives1  16631377 ns/op   6455678 B/op  100052 allocs/op
BenchmarkScoreboardList/SelectorHolders/Objectives3  36073634 ns/op  12343544 B/op  220061 allocs/op
BenchmarkScoreboardList/SelectorHolders/Objectives8  87628067 ns/op  27706796 B/op  560072 allocs/op
BenchmarkScoreboardList/AllHolders                  11493200 ns/op   2585906 B/op      164 allocs/op
```

最后两组（`SortThenFirst` 和 `ClosestOnly`）是**同一批实体、同一个进程里**的
对比：实体只建一次，所以数字不受垃圾回收干扰。它们量的是选择器的最后一步
——"排完再取第一个"和"只找最近的"——见 2.2。

## 三个瓶颈

### 1. 命令被重复解析（约占 60%）

`Engine.runEntry` 在**运行期**把每一行重新解析一遍；而函数**加载期**只检查了
命令名，没有留下解析结果。

```
只解析 say            4.46 µs/条
只执行（AST 已解析）   2.76 µs/条
解析 + 执行           3.53 µs/条
```

另一组测量（用 `go build -overlay` 在仓库外跑的）把重复解析的代价单独量了出来：
**646 ns + 12 次分配**，而一条命令的总成本是 1063 ns——**约 60%**。

还有一处更明显的：**`RunFunction` 比 `ExecuteLine` 慢一倍以上**
（7.49 µs vs 3.53 µs）。函数路径除了重复解析，还要过一遍命令栈。

**建议**：函数**加载时**把每行解析一次并把 AST 缓存下来，运行时直接用。
注意要**保持 1.12.2 的语义**：1.12.2 加载函数时只检查命令名，参数写错要等运行时
才暴露——所以加载时解析**失败不要报错**，只是不缓存这一行。

预估：每条命令省 0.6 µs，大函数整体快 1.6~2.5 倍。

### 2. 选择器每次重建实体列表

`SelectorResolver.filter` 的第一行是 `candidates := r.World.EntityList()`——
**每次选择器都把整个实体表复制成一个新切片**（map 查找 + 分配）。
接着结果切片又按候选数**预分配**：

```
8000 个实体：一次选择器 786 µs、131 KB 分配（即使只匹配 1 个实体）
```

分配量随实体数线性增长（100 实体 1.8 KB → 8000 实体 131 KB），
说明预分配的容量就是候选数。

**建议**：
- World 缓存实体切片，实体增删时失效（选择器就只遍历不复制）
- 结果切片**不要**按候选数预分配（多数选择器只匹配一两个）
- 实体类型多的话，按类型分桶，`type=` 过滤直接在桶里做

预估：8000 实体时 786 µs → 100~150 µs。

**已做**（`d946a4d`）：

- `World` 缓存实体切片，`AddEntity` / `RemoveEntity` 增量维护；选择器只遍历
  不复制。缓存与出生顺序严格平行，顺序一字未改。
- 结果切片不按候选数预分配，改为倍增增长（起始容量 4）。
- 顺带消掉了**第二个热点**：每个实体的 `world.NormalizeEntityType`，
  它占 Resolve 的 38%（8000 实体时 161 µs），比复制实体表（32%）还大。
  现在先走可证明等价的快路径，再走纯函数记忆化。

背靠背 A/B（`go build -overlay` 把 HEAD 版映射到仓库外，同一台机交替跑）：

| 基准 | 基线 ns/op | 新版 ns/op | 倍数 | 基线 B/op | 新版 B/op |
|---|---|---|---|---|---|
| Scaling/Entities8000（全命中） | 382 574 | **150 044** | 2.55× | 131 129 | 135 576 |
| FewHits/Entities8000（只命中 1 个） | 258 783 | **56 872** | 4.55× | 131 097 | **56** |
| FewHits/Entities2000 | 91 285 | **21 259** | 4.29× | 32 796 | **59** |
| Repeated/Few（1000 次 × 8000 实体） | 253 122 025 | **41 682 221** | 6.07× | 131 096 232 | 56 005 |

**只命中 1 个时分配从 131 KB 降到 56 B**（2300 倍）——这就是"别按候选数
预分配"的价值。全命中时总分配持平，因为倍增增长约等于原来的预分配量。

**没做按 type 分桶**：`Entity.Type` 是公开可写字段，World 观测不到就地修改，
桶会静默返回错实体。可证明等价的快路径已经拿到九成收益，且零缓存。

### 2.1 排完序才发现：下一个瓶颈是**排序**

改完之后做端到端复测，发现一个更要紧的事实（2000 个实体、2000 次
`scoreboard players set @e[type=armor_stand,c=1]`，实体散在一张 64×32 的
网格上，距离各不相同）：

| | 优化前 | 优化后 |
|---|---|---|
| 总耗时 | 701 ms | **608 ms** |
| 扣掉对照组（2000 次不带选择器的写分，36 ms） | 676 ms | **572 ms** |
| 一次选择器解析 | 91 µs | 21 µs |

**一次解析快了 4.3 倍，端到端只快 15%**——因为剩下的 572 ms 的大头是
**按距离排序**：`@p`/`@a`/`@e` 要把匹配到的实体**全部**按距离排一遍，
然后才按 `c` 截断。2000 个实体排一次是 ~22000 次比较，每次比较算两次
平方距离——2000 次解析就是 8800 万次距离计算。

**这不是 gomc 的缺陷**：真服也是先排后截
（`PlayerSelector.java:536-561`，`Collections.sort` 之后才
`subList(0, min(|n|, size))`）。所以这条是"真服也这么贵"。

但它意味着：**`c=1` 这种只要一个实体的场景，光优化扫描是不够的**。
下一步该做的是**部分选择**（`c` 很小时用 quickselect 取前 k 个），
但必须保持"距离相同时按出生顺序"的稳定性语义——这正是选择器最微妙的地方，
值得单独做、单独测。

### 2.2 只要一个实体时不用把整个列表排一遍（`a301a60`）

上一步说的"部分选择"只做了一半：只做了 `c=1` 的情形。

真服**确实**是先排后截，所以这个做法**不是模仿真服，而是等价但更快**：
排序是稳定的（距离一样的保持匹配时的先后），所以"最近的那个"就是列表里
第一个不比其它远的实体——找它不用把整个列表排一遍。

同样一批实体、同一个进程里测（实体只建一次，所以数字不受垃圾回收干扰）：

| 匹配到的实体数 | 排完再取第一个 | 只找最近的 | 倍数 | 排序的分配 |
|---|---|---|---|---|
| 100 | 1 075 ns | **245 ns** | 4.4× | 952 B |
| 1 000 | 13 283 ns | **2 492 ns** | 5.3× | 8 248 B |
| 8 000 | 161 249 ns | **19 858 ns** | 8.1× | 65 592 B |

**零分配**，而且规模越大越划算：排序是 n log n 次比较（每次比较算两遍平方
距离），扫一遍是 n 次。

等价的证明在 `internal/version/v1_12_2/selector_resolve_test.go` 里：两个实现对同一批实体给出的
实体**总是同一个**，包括距离相同时取先匹配到的那个。

端到端那组基准（`BenchmarkSelectorScanOneHitMany`）每次迭代都要重建 8000 个
实体，垃圾回收的时机是随机的，数字抖得厉害（同一组基准三次测出 271~314 µs），
所以**上面这张表才是可信的**。

### 2.3 一个字段的代价：size class

给 `Entity` 加 `UUID string`（16 字节）让结构体从 136 字节涨到 152 字节，
**跨过了 Go 的分配档位**（144 → 160）。选择器要遍历整张实体表，于是：

| | 8000 个实体遍历一遍 |
|---|---|
| 加 UUID 之前 | 58 899 ns |
| 加了 `UUID string` | 72 123 ns（**+22%**） |
| 改成 `UUID *string` 并挪到结构体末尾 | **61 490 ns** |

`UUID` 现在是**指针**（8 字节），实体回到 144 字节那一档；挪到末尾是为了让
选择器要读的字段（`Type`、`Pos`）留在结构体的前面。

**教训**：在这个代码库里，给 `Entity` 加字段不是免费的——它是被成千上万次
遍历的。


### 3. `scoreboard players list` 是 O(持有者 × 目标)

真服是"列出这个持有者被跟踪的目标"。gomc 的实现对**每个持有者**都遍历一遍
**全部目标**：

| 持有者 1000 × 目标数 | 耗时 | 分配 |
|---|---|---|
| 1 个目标 | 16.6 ms | 100052 |
| 3 个目标 | 36.1 ms | 220061 |
| 8 个目标 | 87.6 ms | 560072 |

目标数 ×3，时间 ×2.1；×8，时间 ×5.3——**近似 quadratic**。
不带名字的 `players list`（列出全部持有者）反而很快：11.5 ms、164 次分配。

**建议**：先把"这个持有者跟踪了哪些目标"算出来一次，再遍历持有者；
或者反过来先扫目标表建索引。目标 → O(持有者 + 目标)。

**已做**（`bfe7dcc`）：`Scoreboard` 里维护"持有者 → 跟踪的目标"索引，列表时直接查。
索引按目标的**创建顺序**排列（用递增序号保证删掉中间的目标后顺序仍对），
分数仍然实时从分数表读，不会缓存过期值。

| 持有者 1000 × 目标数 | 优化前 | 优化后 |
|---|---|---|
| 3 个目标 | 36.1 ms | **10.5 ms** |
| 8 个目标 | 87.6 ms | **22.9 ms** |
| （全部持有者、不带名字） | 11.5 ms | 7.7 ms |

分配也从 560072 降到 180050。输出内容与顺序一个字没变。

## 端到端看：选择器的开销拆成两块

微基准说"一次选择器 786 µs"，但端到端跑出来的东西不一样。把"2000 个实体 +
2000 次带选择器的写分"拆成两种情况，差别全在**选择器匹配到几个实体**：

| 场景（2000 个实体，2000 条命令） | 耗时 |
|---|---|
| 对照：2000 次不带选择器的写分 | 25 ms |
| 2000 次选择器，每次只匹配 **1** 个实体（`c=1`） | **701 ms** |
| 2000 次选择器，每次匹配**全部 2000** 个实体 | **9 600 ms** |

后两档差了 13.7 倍，但差的**不是"扫描实体"**——扫描那部分两边一样
（都是 2000 次解析 × 2000 个实体 = 400 万次遍历）。差的是**匹配到 2000 个实体
之后要做的事**：每个实体一次写分 + 一条回显，2000 条命令就是 **400 万次**。

所以要把两件事分开看：

- **"扫描"贵**（701 ms 那档）：这才是选择器自己的开销，
  缓存实体列表、结果切片别按候选数预分配能治。
- **"展开"贵**（9 600 ms 那档）：这是**命令语义**决定的——真服也是一条实体一条
  消息地发过去，不是 gomc 的缺陷。

> 补一句：这 701 ms 里的大头**不是扫描，是排序**。扫描优化完之后端到端
> 只降到 608 ms，原因见 [2.1](#21-排完序才发现下一个瓶颈是排序)。

顺带量了内存：9 600 ms 那一档的**峰值工作集是 392 MB**。原因是 400 万条回显
先全攒在 `Engine.Feedback` 里、最后才打印；真服是**来一条发一条**，不攒。
这个差别在实体数量大时会变成内存问题，但要改就得动回显的架构
（`Engine.Feedback` 是命令层的公共出口），先记在这里。

## 看着像但其实不用管

- **计分板写入的 17 次分配/条**：主要是回显字符串（`Set score of ... to ...`）。
  真服也回显，这部分是**输出**不是计算。
- **difftest 的快照**：一个用例 12 ms，世界再大也只是线性增长。
- **tick 循环**：100 个 tick 18 ms，每 tick 才跑 50 条命令。
- **`setblock` 比 `say` 贵 3 倍**（18 µs vs 3.5 µs）：贵在坐标解析和方块存储，
  属于固有成本，不算缺陷。

## 试过但没做：把回显改成延后格式化

**想法**：命令的回显（`Engine.Feedback`）现在是在命令线程里就
`fmt.Sprintf` 好的字符串。真服那边，控制台输出走的是日志系统——消息进队列，
后台线程写。所以回显的格式化不该占命令的临界路径。

**做法**：把 `Feedback` 改成结构化消息（模板 + 参数），命令线程只入队，
真要输出时才格式化。93 个写入点用 AST 改写器批量替换。

**结果：更慢了，已回滚。**

```
                    改前          改后
allocs/op       90,018  →    110,019   （每条命令多 2 次分配）
B/op         4,745,987  →  6,402,993  （+35%）
```

**原因**是 Go 的装箱模型：`fmt.Sprintf(fmt, args...)` 和
`NewMessage(fmt, args...)` 一样，**变参 `[]any` 要分配、每个参数装箱也要分配**
（值类型 → 接口）。区别只是把"最终那个字符串"换成了"消息结构"——而消息结构
里还多了一个含指针的 `[]any` 字段，GC 扫得更重。

**所以在 Go 里，"延后格式化"实际是"延后装箱"，不省反亏。**
（这在 Java/C# 里成立，在 Go 里不成立——那边的消息本来就是对象。）

**真正省的办法**是**热路径上用字符串拼接代替 `fmt.Sprintf`**：`players list`
那条已经这么改了（`"Showing " + strconv.Itoa(n) + " tracked objective(s) for " + h + ":"`），
拼接比 `Sprintf` 少一半分配。其余命令的回显频率低，收益不大，没改。

## 优先级

| 顺序 | 做什么 | 预计收益 | 状态 |
|---|---|---|---|
| 1 | 加载时缓存 AST | 每条命令省 60% 的时间 | ✅ 已做（`b53d899`） |
| 2 | 选择器不重建实体列表 | 选择器快 5~8 倍 | ✅ 已做（`d946a4d`） |
| 3 | `players list` 去掉双重循环 | 目标多时快 4 倍 | ✅ 已做（`bfe7dcc`） |
| 4 | 选择器按 `c` 做部分选择 | `c` 很小时省掉排序 | 部分已做（`a301a60`，只做了 `c=1`，见 2.2） |

前三项都是**改动小、语义不变**的活，做完再跑一遍上面的基准就能验证。

**第 1 项的结果**：走 loader 的 `say`×10000 从 10.50 ms 降到 **3.96 ms**（分配
90038 → 30038）；走 loader 的 10000 条 setblock 从 19.52 ms 降到 **7.75 ms**
（分配 130119 → 10118）。加载期"只查命令名"的语义没变：参数写错的函数照样能
加载，只在运行期报错。

**第 2 项的结果**：一次选择器解析（8000 实体、只命中 1 个）从 258 783 ns 降到
**56 872 ns**（4.55×），分配从 131 097 B 降到 **56 B**。全命中时 382 574 →
**150 044 ns**（2.55×）。端到端（2000 实体 + 2000 次 `c=1` 的选择器写分）
701 ms → 608 ms——**只快 15%**，因为大头已经不是扫描而是**按距离排序**
（见 2.1）。

**第 4 项是新发现的**：真服也是先排后截，所以这条不是"修 bug"，
而是"同样的语义换个更省的做法"。收益上限就在 2.1 那张表里——
端到端 608 ms 里有 572 ms 花在选择器上，其中大部分是排序。

