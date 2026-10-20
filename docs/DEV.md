# 开发指南

## 目录结构

```
cmd/gomc/                  CLI 入口
internal/
  nbt/                     NBT 值模型、SNBT 解析、二进制编解码、子集匹配
  lexer/                   一行命令 -> token
  ast/                     命令的抽象语法树
  parser/                  token -> AST，带列号错误
  world/                   世界状态：方块、实体、scoreboard、gamerule、环境
  exec/                    执行引擎：上下文、选择器解析、函数栈、tick
  version/v1_12_2/         1.12.2 方言：命令实现
  difftest/                对拍层：用例、观察记录、golden 比对
  cli/                     CLI 子命令分发
docs/                      本目录
testdata/
  demo-world/              演示用的 1.12.2 存档
  cases/                   对拍用例 + golden
tools/
  probe/                   对真实 1.12.2 服务端做差分测试（PowerShell + Python）
  blockstates/             从服务端 jar 抽方块状态表（Java + Go 生成器）
.ref/                      反编译产物与参考资料（不进版本库）
```

依赖严格单向，**禁止反向 import**：

```
cmd/gomc -> cli -> version/v1_12_2 -> exec -> world -> nbt
                                  -> ast/lexer/parser
```

- `nbt` / `world` 不知道 `exec` 的存在。
- `exec` 不知道任何具体版本，版本差异通过 `Registry` 注入。
- 任何包不得 import `cmd/gomc`。

---

## 加一条命令

1. 在 `internal/version/v1_12_2/` 下**按域新建文件**
   （`cmd_scoreboard.go`、`cmd_entity.go`、`cmd_world.go`……），
   不要往已有的上帝文件里塞。
2. 实现 `exec.Handler`：

   ```go
   type xxxCmd struct{}

   func (c *xxxCmd) Name() string { return "xxx" }        // 命令名
   func (c *xxxCmd) Usage() string { return "/xxx <a> <b>" } // 用法
   func (c *xxxCmd) Parse(p *parser.Parser) ([]ast.Arg, error) { ... }
   func (c *xxxCmd) Run(ctx *exec.Context, args []ast.Arg) (int, error) { ... }
   ```

3. **在解析期把参数变成结构化值**，`Run` 里不要再解析文本。
   需要自定义参数类型时，在自己的文件里定义结构体 + 一个实现 `ast.Arg`
   的包装（参考 `cmd_execute.go` 的 `executeArgs` / `executeArg`）。
4. 在本域的 `registerXxx(reg)` 里 `reg.Register(&xxxCmd{})`，
   再在 `version.go` 的 `Register` 里调用它。
5. 目标参数用 `parseEntityTarget(p)` + `resolveEntityTargets(ctx, arg)`
   （1.12.2 的实体命令都接受选择器**或**玩家名）。
6. 错误用 `exec.Errorf("...")`，成功消息 append 到 `ctx.Engine.Feedback`。
7. 写测试，放在同目录的 `cmd_<域>_test.go`。
8. 把 1.12.2 的语法和怪癖写进 `docs/SPEC-1.12.2.md`，
   清单写进 `docs/COMMANDS.md`。

**文件不要超过 ~400 行。** 装不下就按子命令再拆。

---

## 测试纪律

```powershell
go test ./internal/... -timeout 20s -p 2
```

- **必须带 `-timeout`**。这个项目被一个挂死的测试拖垮过机器，
  挂死的测试比失败的测试糟糕得多。
- **不要跑没有超时的 `go test ./...`**。
- 不要在后台跑没有超时的测试。
- lexer 里有两个专门的回归测试：
  `TestTokenizeOffsetsAdvance`（游标必须推进）和 `TestTokenizeTerminates`
  （看门狗 goroutine）。它们的存在是因为早期版本有个变量遮蔽导致死循环。

---

## 怎么确认 1.12.2 的行为

**不要靠记忆，也不要直接抄现代 Wiki**——1.13 重写了命令系统，
现代 Wiki 的语法在 1.12.2 上很多是跑不出来的。

按可信度从高到低：

1. **实测**：起一份隔离的真实服务端跑一遍。仓库里带了这个工具：
   [`tools/probe/`](../tools/probe/README.md)——`probe.ps1` 会搭好隔离环境、
   喂命令、抓日志，`cases/` 里是已经验证过的用例。
2. **读字节码**：反编译服务端 jar。`PlayerSelector` 这类把逻辑藏在匿名
   内部类里的，CFR 还原不出来，用 `javap -p -c` 读字节码。
3. **查 Wiki 变更日志**：Minecraft Wiki 的 Java 版更新记录里明确写了每个版本
   加了什么、修了什么。**"1.13 修了这个 bug" 就意味着 1.12.2 有这个 bug**。

> 只用第 2、3 条是不够的：反编译会读出**推断**，而推断可能是错的。
> 已经发生过一次——从字节码里"读"出 1.12.2 有 `/tag` 命令，实测发现没有。
>
> 这条纪律后来又救了三次场：`execute` 的多实体顺序、失败传播、实体显示名，
> 三条都是"读"出来的结论，被实测推翻了。**凡是能测的，都要测。**
>
> 还有一条捷径值得记住：1.12.2 的**提示文字全在语言文件里**
> （`assets/minecraft/lang/en_US.lang`，在服务端 jar 里）。要文案不必起服，
> 见下面的脚本。

## 对齐一条命令的语法：语言文件给用法，源码给顺序

语言文件里有两个东西很容易搞混：

- `commands.<命令>.usage` —— **用法**（参数叫什么、哪些必填）
- `commands.<命令>.<动作>.success` —— **回显文案模板**（带 `%s`）

**但 `%s` 的顺序不在语言文件里，在源码里。** 模板写
`Set score of %s for player %s to %s`，源码传的却是
`{目标名, 玩家名, 分数}`——顺序不对就全错，而且**错得很隐蔽**：
显示名和目标名相同时（`- p2: 5 (p2)`）看不出来，
用显示名不同时（`- MyCounter: 20 (counter)`）才暴露。

> 这一条栽过跟头：曾经只对着语言文件猜 `%s` 顺序，把
> `players list` 的条目写成 `- <目标>: <分数> (<类型>)`，
> 真服其实是 `- <显示名>: <分数> (<目标名>)`；
> 更糟的是把 `players add` / `teams` 全部子命令 / `give`
> 写成了"没有回显"，而真服**都有**——然后这些错误被录进了
> SPEC 和 golden，把错误**固化成了基准**。
>
> **规矩**：每条回显文案，都要能在源码里找到那个 `sendMessage`
> 或 `new ChatMessage(...)`，把传参顺序抄下来再写。

读语言文件的脚本：

```powershell
Add-Type -AssemblyName System.IO.Compression.FileSystem
$zip = [System.IO.Compression.ZipFile]::OpenRead('D:\game\McServer\1.12.2\cache\patched_1.12.2.jar')
$e = $zip.Entries | Where-Object { $_.FullName -eq 'assets/minecraft/lang/en_US.lang' }
$sr = New-Object System.IO.StreamReader($e.Open()); $txt = $sr.ReadToEnd(); $sr.Close(); $zip.Dispose()
($txt -split "`r?`n") | Where-Object { $_ -match '^commands\.scoreboard\.' }
```

要确认**参数顺序**，就看 `.ref/src/Command<名字>.java` 里那个
`new ChatMessage("commands.xxx", 参数1, 参数2, ...)`。

参考资源在本机：

```
D:\game\McServer\1.12.2\cache\patched_1.12.2.jar   # 含 NMS 类的真 jar
.ref\tools\cfr.jar                                # CFR 0.152
.ref\src\                                          # 反编译产物
.ref\mcdata\                                       # minecraft-data 的 1.12 数据
```

反编译一个新 jar 的流程：

```powershell
# 1) 从 jar 里取出目标类
# 2) 打成 jar（CFR 不接受目录输入）
# 3) 反编译
java -jar .ref\tools\cfr.jar <jar> --outputdir .ref\src --comments false
```

---

## 方块状态表是怎么来的

`/fill` 和 `/setblock` 的数据值那一格能写状态字符串（`stone variant=granite`），
背后有两张表：**属性 → 取值 → 数据值**，和**数据值 → 数据值**（命令里写的数据值
和 1.12.2 存下来的数据值不是一回事，见 SPEC 6.4）。

这两张表**不是读字节码读出来的**，是**跑出来的**：

```powershell
# 1) 从真服务端 jar 里 dump（Paper 的 patched jar 里的 NMS 类是 SRG 重映射过的，
#    名字可读，不用啃混淆名）
$srv = 'D:\game\McServer\1.12.2'
$libs = (Get-ChildItem "$srv\libraries" -Recurse -Filter *.jar -File).FullName -join ';'
$cp = "$srv\cache\patched_1.12.2.jar;$libs"
javac -encoding UTF-8 -cp $cp -d tools\blockstates\out tools\blockstates\DumpStates.java
java -cp "tools\blockstates\out;$cp" DumpStates > tools\blockstates\states.jsonl

# 2) 生成 Go 表
go run ./tools/blockstates -in tools/blockstates/states.jsonl `
    -out internal/version/v1_12_2/blockstates_table.go
```

`DumpStates.java` 会**加载服务端的方块注册表**，然后对每个方块问三件事：
它有哪些属性、每个属性有哪些取值、每个取值组合的数据值是多少；外加
`fromLegacyData(n)` 之后数据值是多少。`DispenserRegistry.c()` 是 1.12.2 里
那个 bootstrap 的入口，不先调它，`Blocks` 的静态块会抛
"Accessed Blocks before Bootstrap!"。

**为什么不静态读字节码**：属性是运行期构造的
（`BlockStateEnum.of(name, EnumClass)` 从枚举常量生成名字到取值的映射），
静态读容易漏；而 `fromLegacyData` 的翻译关系是各个方块自己覆写的，
只有真跑一遍才准。

## 热路径上的小技巧

- **命令的回显用字符串拼接，别用 `fmt.Sprintf`**。`Sprintf` 要装箱参数、
  要走反射，拼接只要一次分配。`scoreboard players list` 那条已经这么写了
  （见 `cmd_scoreboard_players.go`）。
- **别想着把格式化推到"输出阶段"**。听起来像真服的做法（那边消息是对象，
  进队列后台写），但 Go 的装箱模型会让它**更慢**：变参 `[]any` 照样要分配、
  每个参数照样要装箱。实测数据见 [`docs/PERFORMANCE.md`](PERFORMANCE.md)
  的"试过但没做"一节。
- **别用正则改 Go 源码**。要批量改写就写个小工具走 `go/ast`：
  字符串里的括号会让正则的括号配对切错位置。

---

## 改公共逻辑时
- `internal/parser` 或 `internal/exec` 改了 → 跑全部 golden，
  `gomc difftest check` 看 diff 合不合理。
- 选择器语义改了 → 同步更新 `docs/SPEC-1.12.2.md` 第 3 节，
  以及 `internal/parser/selector_test.go` 和 `internal/exec/selector_test.go`。
- gamerule 清单改了 → 同步 `internal/world/state.go` 和 SPEC 第 5 节。
- **命令的回显文字改了** → 同步 SPEC 里对应的小节。
  真服的文案是语言文件里的原文，gomc 必须一字不差地对齐；
  对齐后重录 golden：

  ```powershell
  $env:GOMC_UPDATE_GOLDEN='1'
  go test ./internal/difftest/... -run TestGoldenCases
  $env:GOMC_UPDATE_GOLDEN=''
  ```

  **重录后一定要看 diff**，确认变化只来自预期的行为改动。

---

## 提交前

```powershell
gofmt -l .                              # 必须没有输出
go build ./...
go vet ./...
go test ./... -timeout 30s -p 2
gomc.exe difftest check testdata/demo-world testdata/cases/scoreboard.difftest
```
