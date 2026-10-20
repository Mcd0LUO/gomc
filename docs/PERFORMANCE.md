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

```
BenchmarkBaselineSay/ExecuteLine       3534692 ns/op   4745987 B/op   90018 allocs/op
BenchmarkBaselineSay/RunFunction      7493582 ns/op   7400217 B/op   90038 allocs/op
BenchmarkBaselineSay/ExecuteParsed    2760157 ns/op   1908186 B/op   30000 allocs/op
BenchmarkParseOnly/Say                 4459020 ns/op   3120021 B/op   60000 allocs/op
BenchmarkSetblock/ExecuteLine        18062886 ns/op  13138929 B/op  130099 allocs/op
BenchmarkSummon/ExecuteLine          12036768 ns/op   7166032 B/op   84835 allocs/op
BenchmarkSelectorScan/Resolver      193618883 ns/op  32824149 B/op    4001 allocs/op
BenchmarkSelectorScan/Command      1153764500 ns/op 328581432 B/op  6037088 allocs/op
BenchmarkSelectorScanScaling/Entities100    11698 ns/op     1848 B/op      4 allocs/op
BenchmarkSelectorScanScaling/Entities500    33152 ns/op     8249 B/op      4 allocs/op
BenchmarkSelectorScanScaling/Entities2000  123397 ns/op    32824 B/op      4 allocs/op
BenchmarkSelectorScanScaling/Entities8000  785841 ns/op   131143 B/op      4 allocs/op
BenchmarkScoreboardWrite/ExecuteLine 22724584 ns/op  12609612 B/op  170137 allocs/op
BenchmarkScoreboardList/SelectorHolders/Objectives1  16631377 ns/op   6455678 B/op  100052 allocs/op
BenchmarkScoreboardList/SelectorHolders/Objectives3  36073634 ns/op  12343544 B/op  220061 allocs/op
BenchmarkScoreboardList/SelectorHolders/Objectives8  87628067 ns/op  27706796 B/op  560072 allocs/op
BenchmarkScoreboardList/AllHolders                  11493200 ns/op   2585906 B/op      164 allocs/op
```

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
| 2 | 选择器不重建实体列表 | 选择器快 5~8 倍 | 未做 |
| 3 | `players list` 去掉双重循环 | 目标多时快 4 倍 | ✅ 已做（`bfe7dcc`） |

三个都是**改动小、语义不变**的活，做完再跑一遍上面的基准就能验证。

**第 1 项的结果**：走 loader 的 `say`×10000 从 10.50 ms 降到 **3.96 ms**（分配
90038 → 30038）；走 loader 的 10000 条 setblock 从 19.52 ms 降到 **7.75 ms**
（分配 130119 → 10118）。加载期"只查命令名"的语义没变：参数写错的函数照样能
加载，只在运行期报错。

**第 2 项还没做**。它要改 `internal/exec/selector.go` 和 `internal/world/state.go`，
而这两处刚做完一轮语义复查，所以先等复查结论稳定下来再动。
