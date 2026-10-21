<#
.EXAMPLE
	.\compare-with-gomc.ps1 -Case .\cases\34-blockstates.case `
	    -Golden ..\..\..\testdata\cases\blockstates.difftest.golden
#>
<#
	把真服的探针输出和 gomc 的 difftest golden **逐条**对比。

	这是「对拍」的机器版本：探针在真实 1.12.2 服务端上跑一遍用例，把输出录成
	`.expected`；gomc 那边有同一个用例的 `.difftest`（命令一字不差）和它的
	`.golden`（gomc 自己的输出）。这个脚本把两边的输出按命令对齐，一条一条比。

	对齐规则：用例里 `say Sxx` 标记在命令**之前**，标记后面那段输出属于它
	后面那条命令；第一个标记之前的那几行属于最前面的那几条命令。

	退出码 0 表示每一条都一致，1 表示有出入。
#>
param(
	[string]$Case = '',
	[string]$Golden = '',
	[int]$Setup = 0
)

$ErrorActionPreference = 'Stop'

if (-not $Case -or -not $Golden) {
	Write-Host "用法：.\compare-with-gomc.ps1 -Case <用例.case> -Golden <用例.difftest.golden> [-Setup <setup 步数>]" -ForegroundColor Yellow
	exit 2
}

# 1. 用例里的控制台命令（跳过 say 标记），以及第一个标记前有几条命令。
#
# 标记的形状是 `say <一个大写字母><数字>`，比如 `say M01`。别的 `say`
# 是**要对比的命令**，不是标记：真服对 `say "a b"` 的回答是
# `[Server] "a b"`，引号和空格都在里面，和标记长得不一样。
$caseLines = Get-Content $Case
$inConsole = $false
$cmds = @()
$leading = 0
$seenMarker = $false
foreach ($l in $caseLines) {
	if ($l -match '^--- console ---') { $inConsole = $true; continue }
	if (-not $inConsole) { continue }
	$t = $l.Trim()
	if ($t -eq '' -or $t.StartsWith('#')) { continue }
	if ($t -match '^say\s+[A-Z]\d+$') { $seenMarker = $true; continue }
	$cmds += $t
	if (-not $seenMarker) { $leading++ }
}

# 2. 真服的输出：按 [Server] M01 这样的标记切段。
$expectedPath = [IO.Path]::ChangeExtension($Case, '.expected')
$serverLines = Get-Content $expectedPath
$segments = @()
$pre = @()
$cur = -1
foreach ($l in $serverLines) {
	if ($l -match '^\[Server\] [A-Z]\d+$') {
		$segments += ,@()
		$cur = $segments.Count - 1
		continue
	}
	if ($cur -lt 0) { $pre += $l } else { $segments[$cur] += $l }
}

# 第一个标记之前的命令有几条，那几行输出就属于它们。
$server = @()
for ($i = 0; $i -lt $leading; $i++) { $server += ,@($pre[$i]) }
for ($i = 0; $i -lt $segments.Count; $i++) {
	# 末尾可能有一段是空的（最后一个标记后面没有命令），丢掉。
	if ($i -eq $segments.Count - 1 -and $segments[$i].Count -eq 0) { continue }
	$server += ,$segments[$i]
}

# 3. gomc 的输出：每个 step 的 error 或 feedback。
$goldenLines = Get-Content $Golden
$gomc = @()
foreach ($l in $goldenLines) {
	if ($l -match '^step \d+: ') { $gomc += ,@(); continue }
	if ($l -match '^  error: (.*)$') { $gomc[-1] += $Matches[1]; continue }
	if ($l -match '^  feedback: (.*)$') { $gomc[-1] += $Matches[1]; continue }
}

# difftest 的 setup 步在真服那边没有对应的命令（那是 gomc 自己搭场景用的），
# 所以比的时候跳过前面那几步。
if ($Setup -gt 0) {
	$gomc = $gomc[$Setup..($gomc.Count - 1)]
}

Write-Host "用例命令 $($cmds.Count) 条，真服输出 $($server.Count) 段，gomc $($gomc.Count) 步" -ForegroundColor Cyan
$bad = 0
for ($i = 0; $i -lt $cmds.Count; $i++) {
	$s = ($server[$i] -join ' | ')
	$g = ($gomc[$i] -join ' | ')
	$ok = $s -eq $g
	if (-not $ok) { $bad++ }
	$mark = if ($ok) { 'OK  ' } else { 'DIFF' }
	Write-Host "$mark [$i] $($cmds[$i])" -ForegroundColor $(if ($ok) { 'DarkGray' } else { 'Yellow' })
	if (-not $ok) {
		Write-Host "       真服: $s" -ForegroundColor Yellow
		Write-Host "       gomc: $g" -ForegroundColor Yellow
	}
}
Write-Host ""
if ($bad -eq 0) {
	Write-Host "全部 $($cmds.Count) 条一致" -ForegroundColor Green
	exit 0
}
Write-Host "$bad/$($cmds.Count) 条不一致" -ForegroundColor Red
exit 1
