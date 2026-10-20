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
.\probe.ps1 -Case .\cases\07-new.case -Record

# 跑完全部用例并比对 / 重录全部基准
.\run-all.ps1
.\run-all.ps1 -Record
```

常用参数：

| 参数 | 默认 | 说明 |
|---|---|---|
| `-ServerHome` | `D:\game\McServer\1.12.2` | 真实服务端目录（只读） |
| `-Work` | `%TEMP%\mc122probe` | 隔离环境的落地位置 |
| `-Template` | `%TEMP%\mc122probe-template` | 超平坦世界模板的落地位置 |
| `-Case` | 必填 | 用例文件 |
| `-Java` | `D:\tools\jdk\JDK_1.8\bin\java.exe` | **必须是 JDK 8** |
| `-Port` | 25599 | 测试服端口 |
| `-Compare` / `-Record` / `-Keep` | | 比对 / 录制 / 保留环境 |
| `-RebuildTemplate` | | 强制重建世界模板 |
| `-Fast` | | 跳过每条命令前的 1.5 秒等待 |
| `-Bot <名字>` | | 让一个协议机器人先登录进去（见下） |

### 需要在线玩家的命令

有些命令**只对在线玩家**有用：`/advancement` 和 `/recipe` 就是两条——进度和
配方书都是**玩家**的状态。控制台上没有玩家，所以只测得到它们的错误消息。

`mcbot.py` 是一个最小的 1.12.2 协议机器人：握手 + 离线登录 + 回 Keep Alive，
进去之后就一直挂着。探针的 `-Bot` 开关会先让它登录，再喂命令：

```powershell
.\probe.ps1 -Case .\cases\24-advancement-bot.case -Port 25599 -Bot Steve
```

服务端必须用 `online-mode=false`（探针已经是这样），所以登录不需要加密。
机器人需要 `python` 在 PATH 上。

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

1. **准备世界模板**（只需一次）：起一次空服，把出生点定在 `(0,4,0)`、
   难度设成 peaceful、gamerule 调好，关服。服务端会把这些写进 `level.dat`，
   模板就是这个 `level.dat` 加上超平坦地形。
2. **搭隔离环境**：只复制 jar 和配置文件；`libraries` / `cache` / `assets` /
   `lang` 用 **junction** 指回原目录。用户的存档和原服务端目录**不会被写入或删除**。
3. **复制世界**：每个用例从模板复制一份，超平坦地形 + 固化的出生点，
   互不干扰。
4. **跑用例**：起一次服，把用例命令喂给标准输入，把日志抓下来。
5. **提取观察**：按命令分段，只留用例阶段的输出。

### 为什么要用世界模板

1.12.2 **只在服务端启动时**加载世界出生点所在的那个区块。运行中执行
`/setworldspawn` 只是记录了新位置，并不会把区块加载进来——于是紧接着的
`/setblock` 和 `/summon` 都会报 "Cannot place block outside of the world"。

第一版脚本没有模板，只能"先起一次服定出生点、关掉、再起第二次跑用例"，
于是每个用例要起**两次**服。把出生点固化进模板的 `level.dat` 之后，
复制出来就能直接用——**每个用例只起一次服**，时间减半。

> 这个坑是踩出来的：第一版脚本没有预热，实体相关的用例全部录成空输出，
> 而且**看起来像是命令本身有问题**。

### 为什么用超平坦

地形完全确定，任何坐标的方块和实体位置都可预期，录出来的期望文件不会
因为随机地形而漂移。超平坦结构 = 基岩 1 层 + 泥土 2 层 + 草地 1 层，
草地层在 `y=3`，地面在 `y=4`。

`generator-settings` **必须是三段式**：`<层数>;<方块层>;<生物群系 id>`，即

```
3;minecraft:bedrock,2*minecraft:dirt,minecraft:grass;1
```

写成两段（`<方块层>;<生物群系>`）会解析退化成"单层石头 + 虚空"——整个世界
只剩 y=0 一层石头，上面全是空气。这个坑踩过：当时以为地面在 y=4，
结果盔甲架一路掉到 y=1，`detect` 的行为全对不上。

> 顺带一条 1.12.2 的事实：**实体有重力**。`summon` 出来的盔甲架会落到地面上
> （超平坦世界里是 y=4），所以写用例时要按落地后的坐标算相对位置。
> `tp` 则不会触发下落。

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
- 脚本被打断时只清理**自己启动的**服务端进程（`$script:StartedProcs`）。
  绝不能按"最近启动的 java 进程"来杀——你可能自己开着 Minecraft 服务端。
- 每次跑都用同一个 `-Work` 目录，跑之前先清掉旧的。
- 模板目录里**只有世界**，没有 jar、配置和 junction，所以模板可以随便复制、
  随便删。
- 真实服务端的原存档完全没被碰过。

## 两个 PowerShell 的坑（写这个脚本时踩的）

**1. scriptblock 不会自动捕获局部变量。**

```powershell
function Make {
    $p = "VALUE"
    return @{ Send = { Write-Host $p }.GetNewClosure() }   # 不加就是 null
}
```

PowerShell 7 里，返回出去的 scriptblock 在别处调用时**看不到**定义它的函数里的
局部变量。必须显式调 `.GetNewClosure()`。

**2. `CopyToAsync` 在 StreamReader 上没有。**

`$p.StandardOutput` 是 `StreamReader`，要拿到原始字节流得用
`$p.StandardOutput.BaseStream.CopyToAsync($fileStream)`。

另外，事件处理器（`Register-ObjectEvent -Action`）在另一个 runspace 里写共享对象，
和主线程读它会打架——症状是永远等不到服务端就绪。所以脚本改成**读服务端自己写的
`logs/latest.log`** 判断就绪，用 `CopyToAsync` 把控制台输出落文件。

---

## 已有用例

| 用例 | 验证什么 |
|---|---|
| `01-tag-command` | 1.12.2 **没有**独立的 `/tag` 命令，函数会整个加载失败 |
| `02-execute-multi` | `execute` 对每个匹配实体各执行一次；输出前缀是实体名；按距离升序 |
| `03-say-selector` | `say` 会解析消息里的选择器，无匹配时报错 |
| `04-detect-fail` | 一个实体 `detect` 失败**不影响其余实体**；失败的报错文本 |
| `05-command-feedback` | 函数内 `say` 是 `[Server]`，命令结果是 `[Server:]`；`time`/`gamerule`/`difficulty` 的回显 |
| `06-scoreboard-list` | `scoreboard players list` 的输出格式 |

跑完全部用例并比对：

```powershell
.\run-all.ps1
```

`run-all.ps1` 是**回归套件**：它逐个跑用例，和同名的 `.expected` 比对，不同就
报出来并以退出码 1 结束。没有 `.expected` 的用例**只跑一遍**，不算失败，但会
在结尾列出来（免得被当成"通过了"）。

需要**在线玩家**的用例（`/advancement`、`/recipe`）在用例文件里写一行注释
声明要谁：

```
# bot: Steve
```

`run-all.ps1` 读到就会给探针加上 `-Bot Steve`。

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

> 录基准时**给实体带上 `CustomName`**，否则几个实体输出长得一模一样，
> 顺序之类的信息根本看不出来。`04-detect-fail` 就是靠这个才把
> "最近的那个先执行"看清的。

---

## 注意

- 每个用例要起一次服务端，**约 40 秒**；第一次跑还要先生成模板（多一次起服）。
- 用例里别用需要玩家在线的命令（`@a` 之类），控制台没有玩家。
- 引号在多层传递里容易被吃掉，命令多的用例建议把命令写进文件再由脚本读。
