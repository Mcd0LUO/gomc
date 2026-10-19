# 对拍层使用手册

`internal/difftest` 是这个项目的验收装置。它解决的问题是：

> gomc 的行为，和真实 1.12.2 服务端一致吗？

**相关文档**：[`SPEC-1.12.2.md`](SPEC-1.12.2.md)（行为基准）·
[`COMMANDS.md`](COMMANDS.md)（命令矩阵）·[`ARCHITECTURE.md`](ARCHITECTURE.md)（架构）

---

## 1. 两个 scope

观察记录（observation）分两个范围，因为**真实服务端只能报告一部分状态**。

| scope | 内容 | 能否和真服对比 |
|---|---|---|
| `server`（默认） | 每条命令的返回码、错误文本、feedback | **能** —— 这些正是服务端会打印的东西 |
| `full` | 上面加上 scoreboard 表、实体表、方块表 | 不能 —— 1.12.2 没有命令能导出这些 |

所以：

- 拿 `server` scope 去和真服日志对比 → 这是真正的**对拍**。
- 拿 `full` scope 做回归测试 → 这防的是 gomc 自己改坏。

---

## 2. 写一个用例

用例是一个 `.difftest` 文本文件：

```
# 注释
name: scoreboard basics

setup:
  scoreboard objectives add counter dummy
  scoreboard players set Steve counter 1

run:
  scoreboard players add Steve counter 2
  scoreboard players test Steve counter 3
```

- `name:` 可选，缺省是 `unnamed`。
- `setup:` 里的命令先跑，用来准备世界。
- `run:` 里的命令是**被测对象**。
- 两段都必需，命令不能写在段落之外。

放在 `testdata/cases/` 下。

---

## 3. 跑

```powershell
# 打印观察记录
gomc difftest run testdata/demo-world testdata/cases/scoreboard.difftest

# 和 golden 比对
gomc difftest check testdata/demo-world testdata/cases/scoreboard.difftest

# 重写 golden
gomc difftest update testdata/demo-world testdata/cases/scoreboard.difftest

# 加上世界状态（server scope 之外的部分）
gomc difftest run testdata/demo-world testdata/cases/scoreboard.difftest --full
```

输出长这样：

```
# case: scoreboard basics
step 1: scoreboard objectives add counter dummy
  count: 1
  feedback: Added objective counter
step 2: scoreboard players set Steve counter 1
  count: 1
step 3: scoreboard players add Steve counter 2
  count: 1
```

每一步记录：命令原文、返回码（失败时是错误文本）、以及 feedback。

### 顺序是稳定的

`--full` 里的 scoreboard 表是 `objective/holder` 的**字典序**，
实体表按生成顺序，方块表按 `x,y,z` 排序。

**这一点不是可有可无的**：golden 文件靠它才能稳定比对。要是你自己写脚本解析
`--full` 的输出，可以依赖这个顺序——但它属于 gomc 的实现约定，不是 1.12.2 的语义。

### 函数内部的失败

1.12.2 里一条命令失败**不会**中止函数（逐条命令 try/catch）。
所以：

- 失败会被记在**对应的那一步**上，后面的步骤照常执行。
- 整个用例跑完，失败集中体现在各自的步骤里。
- `gomc run` 则是在函数跑完后统一报告这些失败，退出码 1。

golden 文件就是这份文本。`testdata/cases/*.difftest.golden` 会被
`go test ./internal/difftest/` 自动校验，所以行为一变测试就会红。

---

## 4. 和真实服务端对拍

### 4.1 生成探针

```powershell
gomc difftest probe testdata/cases/scoreboard.difftest
```

输出是一份 `.mcfunction`：

```
# Probe for the differential case "scoreboard basics".
say ## setup: scoreboard objectives add counter dummy
scoreboard objectives add counter dummy
say ## step 1: scoreboard players add Steve counter 2
scoreboard players add Steve counter 2
...
```

`say ## ...` 是**分隔标记**，用来把服务端的日志切段，好和 gomc 的步骤对齐。

### 4.2 在真服上跑

1. 把探针放进真服的 `data/functions/<ns>/<name>.mcfunction`。
2. `/reload`。
3. 从控制台执行 `/function <ns>:<name>`。
4. 从服务端日志里取出 `[Server] ...` / `[Server: ...]` 的行。

### 4.3 对齐两边

两边要归一化掉这些差异：

| 差异 | 处理 |
|---|---|
| gomc 的 `feedback: ` 前缀 | 去掉 |
| 真服的 `[22:01:23 INFO]: ` 时间戳 | 去掉 |
| 真服多出 `Executed N command(s) from function '...'` | 去掉 |
| gomc 的 `step N:` 标题 | 对应真服的 `## step N:` 标记 |
| 错误文本 | 1.12.2 用本地化消息，gomc 用英文原文，需要人工对应 |

> `server` scope 只能覆盖**命令的输出**。命令**改了什么状态**（方块、实体、
> 分数）在真服上无法用命令导出，所以那部分只能靠 `full` scope 自证。

---

## 5. 跑一次真实服务端

这套流程在 `docs/ARCHITECTURE.md` 的"怎么复现一次实测"里有完整步骤。
要点：

- 必须用 **JDK 8**（1.12.2 在 JDK 21 上起不来）。
- `libraries` / `cache` 用 junction 指回原目录，**不要复制**。
- 删临时目录前**先删掉 junction**，否则可能误伤真实服务端。
- `spawn-monsters=false` 会连带关掉 `doMobSpawning`，`/summon` 会失败，
  实测时用 `gamerule doMobSpawning true` 打开。

---

## 6. 什么时候该用对拍

| 场景 | 做法 |
|---|---|
| 加了一条新命令 | 写用例 + 生成 golden（挡回归） |
| 不确定某条命令的 1.12.2 行为 | 先反编译，再实测，最后写进 SPEC |
| 怀疑选择器语义 | 读 `PlayerSelector` 的字节码 + 实测 |
| 改了公共逻辑（解析器、引擎） | 跑全部 golden，diff 看变化合不合理 |
