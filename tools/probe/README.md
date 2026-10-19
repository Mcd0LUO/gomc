# 探针：在真实 1.12.2 服务端上验证行为

gomc 的行为基准来自两处：**反编译服务端 jar**（"读"）和**跑真实服务端**（"验"）。
这个目录负责后者。

它解决的问题是：写 mcfunction 的人需要知道"这条命令在 1.12.2 上到底什么效果"，
而现代 Wiki 讲的是 1.13+ 的命令系统，很多写法在 1.12.2 上直接报错。

---

## 用法

```powershell
# 跑一个用例，把服务端的真实输出打出来
.\probe.ps1 -Case .\cases\02-execute-multi.case

# 和录好的期望文件比对（不同则退出码 1）
.\probe.ps1 -Case .\cases\02-execute-multi.case -Compare

# 给一个新用例录基准（录完必须人工核对）
.\probe.ps1 -Case .\cases\09-new.case -Record
```

常用参数：

| 参数 | 默认 | 说明 |
|---|---|---|
| `-ServerHome` | `D:\game\McServer\1.12.2` | 真实服务端目录（只读） |
| `-Work` | `%TEMP%\mc122probe` | 隔离环境的落地位置 |
| `-Case` | 必填 | 用例文件 |
| `-Java` | `D:\tools\jdk\JDK_1.8\bin\java.exe` | **必须是 JDK 8** |
| `-Port` | 25599 | 测试服端口 |
| `-Compare` / `-Record` / `-Keep` | | 比对 / 录制 / 保留环境 |

---

## 用例格式

```
# 描述这一轮要验证什么
# 期望：...
--- function probe/demo.mcfunction ---
<函数内容，一行一条命令>

--- console ---
<控制台命令，一行一条>
```

`--- function <路径> ---` 之后到 `--- console ---` 之前是函数文件内容，
路径相对于 `<world>/data/functions/`。`--- console ---` 之后每行是一条
控制台命令，按顺序喂给服务端。

---

## 脚本做的五件事

1. **搭隔离环境**：只复制 jar 和配置文件；`libraries` / `cache` / `assets` /
   `lang` 用 **junction** 指回原目录。用户的存档和原服务端目录**不会被写入或删除**。
2. **装函数**：把用例里的 `.mcfunction` 放进测试世界（FLAT 世界，另开一个）。
3. **预热**：起一次服，把 gamerule 和世界出生点定好，关掉。
4. **跑用例**：再起一次服，把用例命令喂给标准输入，把日志抓下来。
5. **提取观察**：按命令分段，只留用例阶段的输出。

### 为什么要有预热这一步

1.12.2 **只在服务端启动时加载世界出生点所在的那个区块**。运行中执行
`/setworldspawn` 只是记录了新位置，并不会把区块加载进来——于是紧接着的
`/setblock` 和 `/summon` 都会报 "Cannot place block outside of the world"。

所以脚本先起一次服把出生点定在 `(0, 4, 0)`，关掉，再起第二次——
这时该区块已经是加载状态。预热设置的 gamerule 和出生点会存进 `level.dat`，
第二次启动自动生效。

> 这个坑是踩出来的：第一版脚本没有预热，实体相关的用例全部录成空输出，
> 而且**看起来像是命令本身有问题**。

输出有三种前缀，对应真实服务端的三种来源：

```
[Server] PROBE_SAY                                    ← say 的输出
[Server: Added new objective 'probe' successfully]      ← 命令结果
[Armor Stand] ORDER_TEST                              ← execute 子命令，前缀是目标实体名
```

注意方括号的位置**不一样**：`say` 的方括号只包住来源名，命令结果则包住整条。
脚本原样保留文本，不做二次解析。

还有一类输出既不是聊天格式、也不是错误，而是普通的多行 INFO
（比如 `scoreboard players list` 的输出）。所以提取方式是**按命令分段**：
只保留 `>>> <命令>` 之后的输出，启动阶段的噪声自然被排除；
函数加载失败发生在启动阶段，所以 ERROR 行一律保留。

每条用例的期望输出录在同名 `.expected` 文件里，`-Compare` 就是拿它比对。

---

## 隔离与安全

- 脚本**先摘 junction 再删目录**。直接 `Remove-Item -Recurse` 会顺着 junction
  删进真实服务端，这是脚本里最需要小心的地方。
- 每次跑都用同一个 `-Work` 目录，跑之前先清掉旧的。
- `spawn-monsters=false`（脚本写死的）会连带关掉 `doMobSpawning` gamerule，
  `/summon` 会失败，所以脚本每次先发 `gamerule doMobSpawning true`。
- 真实服务端的原存档完全没被碰过。

---

## 已有用例

| 用例 | 验证什么 |
|---|---|
| `01-tag-command` | 1.12.2 **没有**独立的 `/tag` 命令，函数会整个加载失败 |
| `02-execute-multi` | `execute` 对每个匹配实体各执行一次 |
| `03-execute-prefix` | `execute` 的子命令输出前缀是目标实体的名字 |
| `04-say-selector` | `say` 会解析消息里的选择器，无匹配时报错 |
| `05-function-prefix` | 函数内 `say` 是 `[Server]`，命令结果是 `[Server:]` |
| `06-detect-fail` | `detect` 前置条件失败时的报错文本 |
| `07-time-gamerule` | `time` / `gamerule` / `difficulty` 的回显文本 |
| `08-scoreboard-list` | `scoreboard players list` 的输出格式 |

跑完全部用例并比对：

```powershell
.\run-all.ps1
```

这些用例验证的结论都写进了 [`../../docs/SPEC-1.12.2.md`](../../docs/SPEC-1.12.2.md)，
每条标了 **[验]**。

---

## 加一个新用例

1. 想清楚要验证什么，先写进 SPEC 的草稿里。
2. 写 `.case` 文件。
3. `.\probe.ps1 -Case ... -Record` 录基准。
4. **人工核对**录出来的观察，和你预期的对不对。
   - 对得上 → 存进 SPEC，标上 `[验]`。
   - 对不上 → 真实服务端赢，改 SPEC（可能也说明 gomc 实现错了）。
5. 对着 gomc 也跑一遍同样的命令，确认两边一致。

---

## 注意

- 每次跑要起一次服务端，**约 40 秒**。
- 用例里别用需要玩家在线的命令（`@a` 之类），控制台没有玩家。
- 引号在多层传递里容易被吃掉，命令多的用例建议把命令写进文件再由脚本读。
