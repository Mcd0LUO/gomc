<#
.SYNOPSIS
	在隔离的 1.12.2 服务端上跑一个探针用例，把服务端的真实输出抓出来。

.DESCRIPTION
	gomc 的行为基准来自两处：反编译服务端 jar（"读"）和跑真实服务端（"验"）。
	这个脚本负责后者。

	它做四件事：
	  1. 准备一个**超平坦**世界模板（一次性生成：出生点、gamerule 都固化进去）。
	  2. 每个用例从模板复制一份世界，再装上用例里的 .mcfunction。
	  3. 用 JDK 8 起一次 nogui，把控制台命令喂进去，把日志抓下来。
	  4. 打印出和命令输出相关的那些行（[Server] ... / [Server: ... / [实体名] ...）。

	**用户的存档和原服务端目录不会被写入或删除。**

	为什么要有世界模板：1.12.2 只在服务端**启动时**加载世界出生点所在的那个区块，
	运行中执行 /setworldspawn 只是记录新位置，并不会把新区块加载进来。于是
	"先设出生点、再跑用例"要起两次服。模板把出生点直接固化进 level.dat，
	复制出来就能用——每个用例只起一次服，快一倍。

	为什么用超平坦：地形完全确定，任何坐标的方块和实体位置都可预期，
	录出来的期望文件不会因为随机地形而漂移。
	超平坦结构 = 基岩 1 层 + 泥土 2 层 + 草地 1 层，站在 y=4 的地面上。

	用例文件格式（.case）：

	    # 描述这一轮要验证什么
	    # 期望：...
	    --- function probe/demo.mcfunction ---
	    scoreboard players tag @e[type=zombie] add marked
	    --- console ---
	    function probe:demo

	`--- function <路径> ---` 之后到 `--- console ---` 之前是函数文件内容，
	`--- console ---` 之后每行是一条控制台命令。

.PARAMETER ServerHome
	真实服务端所在的目录（要能从里面读到 paper jar、libraries、cache）。

.PARAMETER Work
	隔离环境的落地目录。默认用系统临时目录下的 mc122probe。

.PARAMETER Template
	超平坦世界模板的落地目录。默认用系统临时目录下的 mc122probe-template。
	不存在时会自动生成（要起一次服，约 40 秒）。

.PARAMETER Case
	用例文件（.case）。

.PARAMETER Java
	JDK 8 的 java.exe 路径。1.12.2 在 JDK 21 上起不来，必须用 8。

.PARAMETER Port
	测试服端口。默认 25599。

.PARAMETER Compare
	把抓到的输出和同名 .expected 文件比对，不同则退出码为 1。

.PARAMETER Record
	把抓到的输出写成 .expected 文件（用于给一个新用例建立基准）。
	录完**必须人工核对**一遍：真实服务端才是权威。

.PARAMETER Keep
	保留隔离环境（默认跑完就清掉）。

.PARAMETER RebuildTemplate
	强制重建世界模板。

.EXAMPLE
	.\probe.ps1 -Case .\cases\01-tag-command.case
	.\probe.ps1 -Case .\cases\02-execute-multi.case -Compare
	.\probe.ps1 -Case .\cases\03-new.case -Record
#>
[CmdletBinding()]
param(
	[string]$ServerHome = 'D:\game\McServer\1.12.2',
	[string]$Work = (Join-Path $env:TEMP 'mc122probe'),
	[string]$Template = (Join-Path $env:TEMP 'mc122probe-template'),
	[string]$Case,
	[string]$Java = 'D:\tools\jdk\JDK_1.8\bin\java.exe',
	[int]$Port = 25599,
	[switch]$Compare,
	[switch]$Record,
	[switch]$Keep,
	[switch]$RebuildTemplate
)

$ErrorActionPreference = 'Stop'

# 记录这个脚本启动的服务端进程，脚本被打断时据此清理。
$script:StartedProcs = @()

# 要用 junction 指回原目录的大目录。
$Linked = @('libraries', 'cache', 'assets', 'lang')

# 超平坦世界的地形：基岩 1 层 + 泥土 2 层 + 草地 1 层。
#
# 1.12.2 的 generator-settings 是三段式：<层数>;<方块层>;<生物群系 id>。
# 写成两段（<方块层>;<生物群系>）时解析会退化成"单层石头"，整个世界变成
# 石头底 + 虚空——实测发现的，别再改回去。
# 草地层在 y=3，站立面是 y=4。
$FlatLayers = '3;minecraft:bedrock,2*minecraft:dirt,minecraft:grass;1'

function Fail($msg) {
	Write-Host "错误：$msg" -ForegroundColor Red
	exit 1
}

function Write-Step($msg) {
	Write-Host "==> $msg" -ForegroundColor Cyan
}

# ---------------------------------------------------------------- 服务端骨架

# 搭一份能跑起来的服务端：jar + 配置文件 + 指回原目录的 junction。
# **不包含世界**——世界要么新建（生成模板），要么从模板复制。
function New-ServerSkeleton {
	param([string]$Dir)

	if (-not (Test-Path $ServerHome)) { Fail "找不到服务端目录 $ServerHome" }
	$jar = Get-ChildItem $ServerHome -Filter 'paper-*.jar' -File | Select-Object -First 1
	if (-not $jar) { Fail "$ServerHome 里没有 paper-*.jar" }
	if (-not (Test-Path $Java)) { Fail "找不到 JDK 8：$Java" }

	New-Item -ItemType Directory -Force -Path $Dir | Out-Null
	Copy-Item $jar.FullName $Dir -Force
	foreach ($f in @('eula.txt', 'bukkit.yml', 'spigot.yml', 'paper.yml',
			'commands.yml', 'permissions.yml', 'ops.json', 'whitelist.json',
			'banned-ips.json', 'banned-players.json')) {
		$p = Join-Path $ServerHome $f
		if (Test-Path $p) { Copy-Item $p $Dir -Force }
	}

	# 用 junction 指回原目录：既省空间，也保证用的是同一份依赖。
	foreach ($d in $Linked) {
		$target = Join-Path $ServerHome $d
		if (Test-Path $target) {
			New-Item -ItemType Junction -Path (Join-Path $Dir $d) -Target $target | Out-Null
		}
	}

	$props = @(
		'level-name=probe'
		'level-type=FLAT'
		"generator-settings=$FlatLayers"
		'online-mode=false'
		"server-port=$Port"
		'max-players=5'
		'view-distance=4'
		# 保持 16（vanilla 默认值）：这样出生点区块会被服务端加载，
		# /setblock 和 /summon 才不会因为"区块没加载"而失败。
		'spawn-protection=16'
		# 这三条不能关。
		#
		# CraftBukkit 的 `World.addEntity` 会拦：动物、怪物和村民只要
		# 对应的开关是关的，就直接 `entity.dead = true` 然后返回 false ——
		# 于是 `/summon zombie` 在真服上会失败，**不是**因为 1.12.2 的
		# 实体工厂造不出僵尸。同理 `difficulty=peaceful` 也会让怪物
		# summon 不了。
		#
		# 代价是世界里可能自然刷怪，但探针是纯控制台驱动、没有玩家，
		# 而 1.12.2 的刷怪是围着玩家发生的，所以实际上不会刷出来。
		'difficulty=normal'
		'spawn-animals=true'
		'spawn-monsters=true'
		'spawn-npcs=true'
		'allow-nether=false'
		'generate-structures=false'
		'enable-command-block=false'
		'motd=gomc-probe'
	)
	Set-Content -Path (Join-Path $Dir 'server.properties') -Value $props -Encoding ascii
}

# 清理时必须先摘掉 junction，否则会顺着链接删到真实服务端。
function Remove-ProbeServer {
	param([string]$Dir)
	if (-not (Test-Path $Dir)) { return }
	foreach ($d in $Linked) {
		$link = Join-Path $Dir $d
		if (Test-Path $link) {
			$item = Get-Item $link -Force
			if ($item.LinkType -eq 'Junction') {
				Write-Step "摘掉 junction $d"
				cmd /c rmdir "$link"
			}
		}
	}
	if (Test-Path $Dir) { Remove-Item -Recurse -Force $Dir }
}

# ---------------------------------------------------------------- 世界模板

# 一次性生成超平坦世界模板：起一次空服，把出生点和 gamerule 定好，
# 关服时服务端会把它们写进 level.dat。之后每个用例复制这份 level.dat 就行。
function Initialize-Template {
	param([string]$TemplateDir)

	Write-Step "生成超平坦世界模板：$TemplateDir"
	$staging = Join-Path $env:TEMP 'mc122probe-init'
	Remove-ProbeServer -Dir $staging
	New-ServerSkeleton -Dir $staging

	$server = Start-Server -Dir $staging -ConsoleFile (Join-Path $staging 'console-init.txt')
	# setworldspawn 会写进 level.dat；关服时一起落盘。
	& $server.Send 'setworldspawn 0 4 0'
	& $server.Send 'gamerule doMobSpawning false'
	& $server.Send 'difficulty normal'
	& $server.Send 'save-all'
	$null = & $server.Finish

	if (-not (Test-Path (Join-Path $staging 'probe\level.dat'))) {
		Fail "模板生成失败：$staging\probe\level.dat 不存在"
	}

	# 只把世界搬进模板目录，骨架（jar、配置、junction）不留。
	if (Test-Path $TemplateDir) { Remove-Item -Recurse -Force $TemplateDir }
	New-Item -ItemType Directory -Force -Path $TemplateDir | Out-Null
	Move-Item (Join-Path $staging 'probe') (Join-Path $TemplateDir 'world')
	Remove-ProbeServer -Dir $staging
	Write-Step "模板就绪：$TemplateDir\world" -ForegroundColor Green
}

function Ensure-Template {
	param([string]$TemplateDir)
	$levelDat = Join-Path $TemplateDir 'world\level.dat'
	if ($RebuildTemplate -or -not (Test-Path $levelDat)) {
		Initialize-Template -TemplateDir $TemplateDir
	}
	else {
		Write-Step "复用世界模板：$TemplateDir" -ForegroundColor DarkGray
	}
}

# ---------------------------------------------------------------- 解析用例

function Read-Case {
	param([string]$Path)
	if (-not (Test-Path $Path)) { Fail "找不到用例文件 $Path" }
	$lines = Get-Content $Path
	$funcs = @{}
	$cmds = @()
	$mode = 'head'
	$current = $null
	foreach ($l in $lines) {
		if ($l -match '^--- function (.+) ---$') {
			$current = $Matches[1]
			$funcs[$current] = @()
			continue
		}
		if ($l -eq '--- console ---') { $mode = 'console'; $current = $null; continue }
		if ($mode -eq 'console') {
			if ($l.Trim() -ne '') { $cmds += $l }
			continue
		}
		if ($current) { $funcs[$current] += $l }
	}
	return @{ Funcs = $funcs; Commands = $cmds }
}

function Install-Functions {
	param([string]$Dir, $Funcs)
	$base = Join-Path $Dir 'probe\data\functions'
	foreach ($path in $Funcs.Keys) {
		$full = Join-Path $base $path
		New-Item -ItemType Directory -Force -Path (Split-Path $full) | Out-Null
		Set-Content -Path $full -Value $Funcs[$path] -Encoding ascii
	}
}

# ---------------------------------------------------------------- 起服

# 起一次服务端，等它 ready，返回一个带 Send/Finish 的对象。
#
# 输出用 CopyToAsync 直接写进文件，不用事件处理器——事件处理器在另一个
# runspace 里写共享对象，和主线程读它会打架（症状是永远等不到
# "Done ("）。就绪判断直接看服务端自己写的 logs/latest.log。
function Start-Server {
	param([string]$Dir, [string]$ConsoleFile, [int]$TimeoutSec = 180)

	# 清掉上一轮的日志，免得把旧的 "Done (" 当成这次的。
	$logDir = Join-Path $Dir 'logs'
	if (Test-Path $logDir) {
		Get-ChildItem $logDir -Filter 'latest.log*' -File | Remove-Item -Force
	}
	$readyLog = Join-Path $logDir 'latest.log'

	$psi = New-Object System.Diagnostics.ProcessStartInfo
	$psi.FileName = $Java
	$psi.Arguments = "-Xms512M -Xmx1200M -jar paper-1.12.2-1620.jar nogui"
	$psi.WorkingDirectory = $Dir
	$psi.UseShellExecute = $false
	$psi.RedirectStandardInput = $true
	$psi.RedirectStandardOutput = $true
	$psi.RedirectStandardError = $true

	$p = New-Object System.Diagnostics.Process
	$p.StartInfo = $psi
	$p.Start() | Out-Null
	# 记下来，脚本被打断时才能只收自己的进程。
	$script:StartedProcs += $p

	# 控制台输出直接落文件，避免和脚本自己的写入打架。
	# StandardOutput 是 StreamReader，异步拷贝要用它的 BaseStream。
	$outStream = [System.IO.File]::Create($ConsoleFile)
	$errFile = "$ConsoleFile.err"
	$errStream = [System.IO.File]::Create($errFile)
	$null = $p.StandardOutput.BaseStream.CopyToAsync($outStream)
	$null = $p.StandardError.BaseStream.CopyToAsync($errStream)

	$sw = [Diagnostics.Stopwatch]::StartNew()
	$ready = $false
	while ($sw.Elapsed.TotalSeconds -lt $TimeoutSec) {
		if (Test-Path $readyLog) {
			$txt = Get-Content $readyLog -Raw -ErrorAction SilentlyContinue
			if ($txt -and $txt -match 'Done \(') { $ready = $true; break }
		}
		if ($p.HasExited) { break }
		Start-Sleep -Milliseconds 500
	}
	if (-not $ready) {
		$p.Kill()
		Fail "服务端没能在 $TimeoutSec 秒内启动（日志 $readyLog）"
	}

	$server = @{
		Proc = $p
		# GetNewClosure() 是必须的：PowerShell 7 里，返回出去的 scriptblock
		# 不会自动捕获函数里的局部变量，不加这个 $p 就是 null。
		Send = {
			param([string]$c)
			Write-Host ">>> $c"
			$p.StandardInput.WriteLine($c)
			$p.StandardInput.Flush()
			# 控制台命令是同步执行的，等一小会儿让日志落盘即可。
			Start-Sleep -Milliseconds 1500
		}.GetNewClosure()
		# 收尾：关服，把输出文件合在一起。
		Finish = {
			$p.StandardInput.WriteLine('stop')
			$p.StandardInput.Flush()
			if (-not $p.WaitForExit(60000)) { $p.Kill() }
			Start-Sleep -Seconds 1
			$outStream.Flush(); $outStream.Dispose()
			$errStream.Flush(); $errStream.Dispose()
			$text = (Get-Content $ConsoleFile -Raw -ErrorAction SilentlyContinue)
			if (Test-Path $errFile) {
				$e = (Get-Content $errFile -Raw -ErrorAction SilentlyContinue)
				if ($e) { $text += "`n" + $e }
			}
			return $text
		}.GetNewClosure()
	}
	return $server
}

# ---------------------------------------------------------------- 提取观察

# 服务端的日志行格式是 `[HH:mm:ss LEVEL]: <消息>`。命令输出有三种形态，
# 方括号的位置不一样，所以这里原样保留文本，不做二次解析：
#
#	[Server] PROBE_SAY                              ← say，括号只包住来源名
#	[Server: Added new objective 'probe' ...]       ← 命令结果，括号包住整条
#	[Armor Stand] ORDER_TEST                        ← execute 子命令，来源是目标实体
#
# 有些命令的输出不是聊天格式，而是多行的普通 INFO（比如
# `scoreboard players list` 的输出），所以不按前缀筛。
#
# 取值范围：以服务端的 "Done (" 为界（之前是启动噪声），
# 之后去掉关服噪声。ERROR 行一律保留——函数加载失败发生在启动阶段。
$Noise = '^(Executing|Executed) \d+ command|^Stopping|^Saving|^Closing|^Flushing|^Unloading|^ThreadedAnvilChunkStorage|^Time elapsed|^Timings Reset|^Preparing spawn|^Calculating'

function Select-Observation {
	param([string]$ConsoleFile)
	$out = @()
	$started = $false
	foreach ($l in Get-Content $ConsoleFile -ErrorAction SilentlyContinue) {
		if ($l -match '^\[\d\d:\d\d:\d\d ERROR\]: (.+)$') {
			$out += ("[ERROR] " + $Matches[1])
			continue
		}
		if (-not $started) {
			if ($l -match 'Done \(') { $started = $true }
			continue
		}
		if ($l -match '^\[\d\d:\d\d:\d\d INFO\]: (.+)$') {
			$t = $Matches[1]
			if ($t -match $Noise) { continue }
			$out += $t
		}
	}
	return $out
}

# ---------------------------------------------------------------- 主流程

if (-not $Case) { Fail "必须给 -Case" }
$casePath = (Resolve-Path $Case).Path
$name = [IO.Path]::GetFileNameWithoutExtension($casePath)
$expectedPath = [IO.Path]::ChangeExtension($casePath, '.expected')

# 脚本被打断时也要把服务端进程收掉，否则它会一直占着端口。
try {
	Write-Step "用例 $name"

	Ensure-Template -TemplateDir $Template

	# 每个用例从模板复制一份世界，保证互相不干扰。
	Remove-ProbeServer -Dir $Work
	New-ServerSkeleton -Dir $Work
	Copy-Item (Join-Path $Template 'world') -Destination (Join-Path $Work 'probe') -Recurse -Force

	$spec = Read-Case -Path $casePath
	Install-Functions -Dir $Work -Funcs $spec.Funcs

	Write-Step "跑用例"
	$server = Start-Server -Dir $Work -ConsoleFile (Join-Path $Work 'console-case.txt')
	foreach ($c in $spec.Commands) { & $server.Send $c }
	$null = & $server.Finish

	$observed = Select-Observation -ConsoleFile (Join-Path $Work 'console-case.txt')

	Write-Host ""
	Write-Host "---- 服务端观察 ----" -ForegroundColor Yellow
	$observed | ForEach-Object { Write-Host $_ }
	Write-Host "---- 完整日志: $(Join-Path $Work 'console-case.txt') ----" -ForegroundColor DarkGray

	$code = 0
	if ($Record) {
		Set-Content -Path $expectedPath -Value $observed -Encoding utf8
		Write-Host ""
		Write-Host "已写入 $expectedPath —— 请人工核对一遍" -ForegroundColor Green
	}
	if ($Compare) {
		if (-not (Test-Path $expectedPath)) { Fail "找不到期望文件 $expectedPath" }
		$expected = Get-Content $expectedPath
		$diff = Compare-Object -ReferenceObject $expected -DifferenceObject $observed
		if ($diff) {
			Write-Host ""
			Write-Host "---- 与期望不一致 ----" -ForegroundColor Red
			$diff | ForEach-Object {
				$sign = if ($_.SideIndicator -eq '<=') { '-' } else { '+' }
				Write-Host "$sign $($_.InputObject)"
			}
			$code = 1
		}
		else {
			Write-Host ""
			Write-Host "与期望一致" -ForegroundColor Green
		}
	}
}
finally {
	# 兜底：只把**这个脚本启动的**服务端进程收掉。
	# 绝不能按"最近启动的 java 进程"来杀——用户自己可能开着 Minecraft 服务端。
	foreach ($proc in $script:StartedProcs) {
		if ($proc -and -not $proc.HasExited) {
			try { $proc.Kill() } catch { }
		}
	}
}

if (-not $Keep) {
	Write-Step "清理隔离环境"
	Remove-ProbeServer -Dir $Work
}
else {
	Write-Host "保留了隔离环境：$Work" -ForegroundColor DarkGray
}

exit $code
