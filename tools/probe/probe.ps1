<#
.SYNOPSIS
	在隔离的 1.12.2 服务端上跑一个探针用例，把服务端的真实输出抓出来。

.DESCRIPTION
	gomc 的行为基准来自两处：反编译服务端 jar（"读"）和跑真实服务端（"验"）。
	这个脚本负责后者。

	它做四件事：
	  1. 在临时目录里搭一份隔离的服务端：只复制 jar 和配置文件，
	     libraries / cache / assets / lang 用 junction 指回原目录。
	     **用户的存档和原服务端目录不会被写入或删除。**
	  2. 把用例里的 .mcfunction 放进测试世界。
	  3. 用 JDK 8 起 nogui，把控制台命令喂进去，把日志抓下来。
	  4. 打印出和命令输出相关的那些行（[Server] ... / [Server: ... / [实体名] ...）。

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

.EXAMPLE
	.\probe.ps1 -Case .\cases\01-tag-command.case
	.\probe.ps1 -Case .\cases\02-execute-multi.case -Compare
	.\probe.ps1 -Case .\cases\03-new.case -Record
#>
[CmdletBinding()]
param(
	[string]$ServerHome = 'D:\game\McServer\1.12.2',
	[string]$Work = (Join-Path $env:TEMP 'mc122probe'),
	[string]$Case,
	[string]$Java = 'D:\tools\jdk\JDK_1.8\bin\java.exe',
	[int]$Port = 25599,
	[switch]$Compare,
	[switch]$Record,
	[switch]$Keep
)

$ErrorActionPreference = 'Stop'

# 要用 junction 指回原目录的大目录。
$Linked = @('libraries', 'cache', 'assets', 'lang')
# 每次起服都会先发的公共命令。
#
# doMobSpawning 必须打开：server.properties 里的 spawn-monsters=false
# 会连带把 doMobSpawning gamerule 关掉，那样 /summon 直接失败。
$Prelude = @(
	'gamerule doMobSpawning true'
	'gamerule doFireTick false'
	'setworldspawn 0 4 0'
)

function Fail($msg) {
	Write-Host "错误：$msg" -ForegroundColor Red
	exit 1
}

function Write-Step($msg) {
	Write-Host "==> $msg" -ForegroundColor Cyan
}

# ---------------------------------------------------------------- 搭环境

function New-ProbeServer {
	param([string]$ServerDir, [string]$Dir)

	if (-not (Test-Path $ServerDir)) { Fail "找不到服务端目录 $ServerDir" }
	$jar = Get-ChildItem $ServerDir -Filter 'paper-*.jar' -File | Select-Object -First 1
	if (-not $jar) { Fail "$ServerDir 里没有 paper-*.jar" }
	if (-not (Test-Path $Java)) { Fail "找不到 JDK 8：$Java" }

	if (Test-Path $Dir) {
		Write-Step "清理旧的隔离环境 $Dir"
		Remove-ProbeServer -Dir $Dir
	}
	New-Item -ItemType Directory -Force -Path $Dir | Out-Null

	Copy-Item $jar.FullName $Dir
	foreach ($f in @('eula.txt', 'bukkit.yml', 'spigot.yml', 'paper.yml',
			'commands.yml', 'permissions.yml', 'ops.json', 'whitelist.json',
			'banned-ips.json', 'banned-players.json')) {
		$p = Join-Path $ServerDir $f
		if (Test-Path $p) { Copy-Item $p $Dir }
	}

	# 用 junction 指回原目录：既省空间，也保证用的是同一份依赖。
	foreach ($d in $Linked) {
		$target = Join-Path $ServerDir $d
		if (Test-Path $target) {
			New-Item -ItemType Junction -Path (Join-Path $Dir $d) -Target $target | Out-Null
		}
	}

	$props = @(
		'level-name=probe'
		'level-type=FLAT'
		'generator-settings=3;minecraft:bedrock,2*minecraft:dirt,minecraft:grass;1'
		'online-mode=false'
		"server-port=$Port"
		'max-players=5'
		'view-distance=4'
		# 保持 16（vanilla 默认值）：这样出生点区块会被服务端加载，
		# /setblock 和 /summon 才不会因为"区块没加载"而失败。
		'spawn-protection=16'
		'difficulty=peaceful'
		'spawn-animals=false'
		'spawn-npcs=false'
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
	Remove-Item -Recurse -Force $Dir
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

# ---------------------------------------------------------------- 跑

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
		Send = {
			param([string]$c)
			Write-Host ">>> $c"
			$p.StandardInput.WriteLine($c)
			$p.StandardInput.Flush()
			Start-Sleep -Seconds 3
		}
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
		}
	}
	return $server
}

function Invoke-Probe {
	param([string]$Dir, [string[]]$Commands)

	$log = Join-Path $Dir 'probe.log'

	# 第一阶段：预热。
	#
	# 1.12.2 只在**服务端启动时**加载世界出生点所在的那个区块，
	# 运行中改 /setworldspawn 不会立刻把区块加载进来。所以先起一次服、
	# 把出生点定好、关掉，再起第二次跑用例——这时区块已经是加载状态。
	#
	# 预热做的 gamerule 和出生点会存进 level.dat，第二次启动时自动生效。
	Write-Step "预热：设置 gamerule 与世界出生点"
	$warm = Start-Server -Dir $Dir -ConsoleFile (Join-Path $Dir 'console-warmup.txt')
	foreach ($c in $Prelude) { & $warm.Send $c }
	$warmText = & $warm.Finish

	# 第二阶段：跑用例。
	Write-Step "跑用例"
	$run = Start-Server -Dir $Dir -ConsoleFile (Join-Path $Dir 'console-case.txt')
	foreach ($c in $Commands) { & $run.Send $c }
	$runText = & $run.Finish

	($warmText + "`n### SESSION: case`n" + $runText) | Out-File -FilePath $log -Encoding utf8
	return $log
}

# 提取命令输出。
#
# 服务端的日志行格式是 `[HH:mm:ss LEVEL]: <消息>`。命令输出有三种形态，
# 方括号的位置不一样，所以这里原样保留文本，不做二次解析：
#
#	[Server] PROBE_SAY                              ← say，括号只包住来源名
#	[Server: Added new objective 'probe' ...]       ← 命令结果，括号包住整条
#	[Armor Stand] ORDER_TEST                        ← execute 子命令，来源是目标实体
#
# 有些命令的输出不是聊天格式，而是多行的普通 INFO（比如
# `scoreboard players list`），所以这里不按前缀筛，而是**按命令分段**：
# 只保留 `>>> <命令>` 之后的输出，启动阶段的噪声自然被排除。
# 例外是 ERROR 行——函数加载失败发生在启动阶段，也要留下。
# 提取命令输出。
#
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
$Noise = '^(Executing|Executed) \d+ command|^Stopping|^Saving|^Closing|^Flushing|^Unloading|^ThreadedAnvilChunkStorage|^Time elapsed'

function Select-Observation {
	param([string]$Log)
	$out = @()
	$started = $false
	foreach ($l in Get-Content $Log) {
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

Write-Step "用例 $name"
New-ProbeServer -ServerDir $ServerHome -Dir $Work
$spec = Read-Case -Path $casePath
Install-Functions -Dir $Work -Funcs $spec.Funcs
$log = Invoke-Probe -Dir $Work -Commands $spec.Commands
$observed = Select-Observation -Log $log

Write-Host ""
Write-Host "---- 服务端观察 ----" -ForegroundColor Yellow
$observed | ForEach-Object { Write-Host $_ }
Write-Host "---- 完整日志: $log ----" -ForegroundColor DarkGray

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

if (-not $Keep) {
	Write-Step "清理隔离环境"
	Remove-ProbeServer -Dir $Work
}
else {
	Write-Host "保留了隔离环境：$Work" -ForegroundColor DarkGray
}

exit $code
