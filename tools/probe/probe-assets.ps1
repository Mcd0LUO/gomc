<#
.SYNOPSIS
	在**自定义 assets 根**的真实 1.12.2 服务端上跑命令，把服务端的真实输出抓出来。

.DESCRIPTION
	1.12.2 的配方不是从世界目录读的，而是从**服务端的 assets 根**读的：
	`CraftingManager` 找 classpath 上的 "/assets/.mcassetsroot"，然后走
	"/assets/minecraft/recipes" 下的所有 .json。

	所以要测"自定义配方"（子目录、坏文件、数据值），得改 classpath 的
	assets 根：这个脚本把 -Assets 目录放在 classpath 的**最前面**，于是
	`getResource("/assets/minecraft/recipes")` 找到的是它，而不是 jar 里的。

	脚本做的事和 probe.ps1 一样：搭隔离环境、复制世界模板、起服、喂命令、
	抓输出。区别只有两点：
	  * classpath 前面加了 -Assets；
	  * 命令来自一个纯文本文件（每行一条），不走 .case 的函数文件。

.PARAMETER Assets
	自己的 assets 根。里面要有 assets/.mcassetsroot 和
	assets/minecraft/recipes/。

.PARAMETER Commands
	命令文件，每行一条命令。空行和 # 开头的行会被跳过。

.PARAMETER Bot
	非空时先让协议机器人登录进去——/recipe 只对在线玩家有用。

.EXAMPLE
	.\probe-assets.ps1 -Assets ..\..\tmp\assets -Commands .\cases-assets\01-subdir.txt -Port 25861 -Bot Steve
#>
[CmdletBinding()]
param(
	[string]$ServerHome = 'D:\game\McServer\1.12.2',
	[string]$Work = '',
	[string]$Template = (Join-Path $env:TEMP 'mc122probe-template'),
	[string]$Assets = '',
	[string]$Commands = '',
	[string]$Java = 'D:\tools\jdk\JDK_1.8\bin\java.exe',
	[int]$Port = 25599,
	[string]$Bot = '',
	[switch]$Keep
)

$ErrorActionPreference = 'Stop'

if (-not $Work) { $Work = Join-Path $env:TEMP "mc122probe-assets-$Port" }
if (-not $Assets) { Fail '必须给 -Assets' }
if (-not $Commands) { Fail '必须给 -Commands' }
if (-not (Test-Path $Assets)) { Fail "找不到 assets 目录 $Assets" }

$script:StartedProcs = @()
$Linked = @('libraries', 'cache', 'assets', 'lang')
$FlatLayers = '3;minecraft:bedrock,2*minecraft:dirt,minecraft:grass;1'

function Fail($msg) {
	Write-Host "错误：$msg" -ForegroundColor Red
	exit 1
}

function Write-Step($msg) {
	Write-Host "==> $msg" -ForegroundColor Cyan
}

function New-ServerSkeleton {
	param([string]$Dir)

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
		'spawn-protection=16'
		'difficulty=normal'
		'spawn-animals=true'
		'spawn-monsters=true'
		'spawn-npcs=true'
		'allow-nether=false'
		'generate-structures=false'
		'enable-command-block=false'
		'motd=gomc-probe-assets'
	)
	Set-Content -Path (Join-Path $Dir 'server.properties') -Value $props -Encoding ascii
}

function Remove-ProbeServer {
	param([string]$Dir)
	if (-not (Test-Path $Dir)) { return }
	foreach ($d in $Linked) {
		$link = Join-Path $Dir $d
		if (Test-Path $link) {
			$item = Get-Item $link -Force
			if ($item.LinkType -eq 'Junction') { cmd /c rmdir "$link" }
		}
	}
	if (Test-Path $Dir) { Remove-Item -Recurse -Force $Dir }
}

function Start-Server {
	param([string]$Dir, [string]$ConsoleFile, [int]$TimeoutSec = 180)

	$logDir = Join-Path $Dir 'logs'
	if (Test-Path $logDir) {
		Get-ChildItem $logDir -Filter 'latest.log*' -File | Remove-Item -Force
	}
	$readyLog = Join-Path $logDir 'latest.log'

	$psi = New-Object System.Diagnostics.ProcessStartInfo
	$psi.FileName = $Java
	# 自己的 assets 根放在 classpath 的**最前面**：getResource 找的是
	# classpath 顺序里的第一个命中，所以 jar 里的配方就被它盖住了。
	$paper = (Get-ChildItem $Dir -Filter 'paper-*.jar' -File | Select-Object -First 1).FullName
	$psi.Arguments = "-Xms512M -Xmx1200M -cp `"$Assets;$paper`" io.papermc.paperclip.Paperclip nogui"
	$psi.WorkingDirectory = $Dir
	$psi.UseShellExecute = $false
	$psi.RedirectStandardInput = $true
	$psi.RedirectStandardOutput = $true
	$psi.RedirectStandardError = $true

	$p = New-Object System.Diagnostics.Process
	$p.StartInfo = $psi
	$p.Start() | Out-Null
	$script:StartedProcs += $p

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

	return @{
		Proc = $p
		Send = {
			param([string]$c)
			Write-Host ">>> $c"
			$p.StandardInput.WriteLine($c)
			$p.StandardInput.Flush()
			Start-Sleep -Milliseconds 1200
		}.GetNewClosure()
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
}

try {
	Write-Step "assets 根：$Assets"
	Write-Step "命令文件：$Commands"

	if (-not (Test-Path (Join-Path $Template 'world\level.dat'))) {
		Fail "找不到世界模板 $Template，先跑一次 probe.ps1 生成"
	}

	Remove-ProbeServer -Dir $Work
	New-ServerSkeleton -Dir $Work
	Copy-Item (Join-Path $Template 'world') -Destination (Join-Path $Work 'probe') -Recurse -Force

	$cmds = Get-Content $Commands | Where-Object { $_.Trim() -ne '' -and -not $_.Trim().StartsWith('#') }

	$server = Start-Server -Dir $Work -ConsoleFile (Join-Path $Work 'console-case.txt')

	$botProc = $null
	if ($Bot) {
		$BotScript = Join-Path $PSScriptRoot 'mcbot.py'
		if (-not (Test-Path $BotScript)) { Fail "找不到 $BotScript" }
		Write-Step "让机器人 $Bot 登录"
		$botLog = Join-Path $Work 'bot.log'
		$botProc = Start-Process -FilePath 'python' `
			-ArgumentList @($BotScript, '127.0.0.1', $Port, $Bot, '300') `
			-NoNewWindow -PassThru `
			-RedirectStandardOutput $botLog -RedirectStandardError "$botLog.err"
		$deadline = (Get-Date).AddSeconds(30)
		while ((Get-Date) -lt $deadline) {
			if (Test-Path $botLog) {
				$seen = Get-Content $botLog -ErrorAction SilentlyContinue
				if ($seen -match 'logged in as') { break }
				if ($seen -match 'login refused|never logged in') {
					Fail "机器人登录被拒绝：$($seen -join ' ')"
				}
			}
			Start-Sleep -Milliseconds 300
		}
		Write-Step "机器人已登录"
	}

	foreach ($c in $cmds) { & $server.Send $c }
	$null = & $server.Finish
	if ($botProc) {
		try { $botProc | Stop-Process -Force -ErrorAction SilentlyContinue } catch { }
	}

	Write-Host ""
	Write-Host "---- 服务端观察 ----" -ForegroundColor Yellow
	# 取值范围：以服务端的 "Done (" 为界（之前是启动噪声），ERROR 行一律保留。
	$console = Join-Path $Work 'console-case.txt'
	$started = $false
	foreach ($l in Get-Content $console -ErrorAction SilentlyContinue) {
		if ($l -match '^\[\d\d:\d\d:\d\d ERROR\]: (.+)$') {
			Write-Host "[ERROR] $($Matches[1])"
			continue
		}
		if (-not $started) {
			if ($l -match 'Done \(') { $started = $true }
			continue
		}
		if ($l -match '^\[\d\d:\d\d:\d\d INFO\]: (.+)$') {
			$t = $Matches[1]
			if ($t -match '^(Executing|Executed) \d+ command|^Stopping|^Saving|^Closing|^Flushing|^Unloading|^ThreadedAnvilChunkStorage|^Time elapsed|^Timings Reset|^Preparing spawn|^Calculating') { continue }
			Write-Host $t
		}
	}
	Write-Host "---- 完整日志: $console ----" -ForegroundColor DarkGray
}
finally {
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
