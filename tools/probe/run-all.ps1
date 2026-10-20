<#
.SYNOPSIS
	跑完 tools\probe\cases 下的全部用例，逐个和 .expected 比对。

.DESCRIPTION
	这是"实测基准"的回归套件：每个用例都在真实 1.12.2 服务端上跑一遍，
	和录好的 .expected 比对。服务端行为变了（换了服务端版本、或者
	期望文件录错了）就会红。

	每个用例都从超平坦世界模板复制一份世界，然后起一次服务端（约 40 秒）。
	第一次跑时会先生成模板（多花一次起服）。

.PARAMETER Case
	只跑指定的用例（文件名或路径），默认全部。

.PARAMETER Work
	隔离环境的落地目录。

.PARAMETER Template
	超平坦世界模板的落地目录。

.PARAMETER Filter
	只跑文件名包含这个串的用例。

.PARAMETER Record
	把抓到的输出写成 .expected（重建基准用），录完**必须人工核对**。

.PARAMETER Bot
	默认的机器人名字。用例自己写了 `# bot: <名字>` 的以用例里的为准。

.EXAMPLE
	.\run-all.ps1
	.\run-all.ps1 -Filter execute
	.\run-all.ps1 -Record
#>
[CmdletBinding()]
param(
	[string]$Case,
	[string]$Work = (Join-Path $env:TEMP 'mc122probe'),
	[string]$Template = (Join-Path $env:TEMP 'mc122probe-template'),
	[string]$Filter = '',
	[string]$Bot = '',
	[switch]$Record
)

$ErrorActionPreference = 'Stop'
$here = Split-Path -Parent $MyInvocation.MyCommand.Path
$probe = Join-Path $here 'probe.ps1'

if ($Case) {
	$cases = @(Get-Item $Case)
}
else {
	$cases = @(Get-ChildItem (Join-Path $here 'cases') -Filter '*.case' | Sort-Object Name)
}
if ($Filter) {
	$cases = @($cases | Where-Object { $_.Name -like "*$Filter*" })
}
if ($cases.Count -eq 0) {
	Write-Host "没有匹配的用例" -ForegroundColor Yellow
	exit 0
}

$recordArg = @()
$compareArg = @()
if ($Record) { $recordArg = @('-Record') } else { $compareArg = @('-Compare') }

$failed = @()
$skipped = @()
foreach ($c in $cases) {
	Write-Host ""
	Write-Host "######## $($c.Name) ########" -ForegroundColor Magenta

	# 有些命令只对**在线玩家**有用（/advancement 和 /recipe），控制台上没有
	# 玩家，所以那些用例要靠一个协议机器人登录进去。用例自己声明要谁：
	# 在注释里写一行 `# bot: Steve`。
	$botArg = @()
	if ($Bot) { $botArg = @('-Bot', $Bot) }
	$head = Get-Content $c.FullName -TotalCount 30 -ErrorAction SilentlyContinue
	$botLine = $head | Where-Object { $_ -match '^\s*#\s*bot:\s*(\S+)' } | Select-Object -First 1
	if ($botLine -and $botLine -match '^\s*#\s*bot:\s*(\S+)') {
		$botArg = @('-Bot', $matches[1])
	}

	# 没有基线的用例（还没录、或者录了会自己跟自己对不上的）只跑一遍，
	# 不算失败——但要说出来，免得被当成"通过了"。
	$expectedPath = [IO.Path]::ChangeExtension($c.FullName, '.expected')
	$hasBaseline = Test-Path $expectedPath
	if (-not $Record -and -not $hasBaseline) {
		$skipped += $c.Name
		# 没有基线就没有可比的东西，所以不比对——探针的 -Compare 会
		# 因为找不到期望文件而退出 1，那样会被记成失败。
		& pwsh -NoProfile -ExecutionPolicy Bypass -File $probe -Case $c.FullName -Work $Work -Template $Template @botArg
		if ($LASTEXITCODE -ne 0) { $failed += $c.Name }
		continue
	}

	& pwsh -NoProfile -ExecutionPolicy Bypass -File $probe -Case $c.FullName -Work $Work -Template $Template @recordArg @compareArg @botArg
	if ($LASTEXITCODE -ne 0) { $failed += $c.Name }
}

Write-Host ""
if ($skipped.Count -gt 0) {
	Write-Host "跳过比对 $($skipped.Count) 个（没有基线）：" -ForegroundColor Yellow
	$skipped | ForEach-Object { Write-Host "  $_" -ForegroundColor Yellow }
}
if ($failed.Count -gt 0) {
	Write-Host "失败 $($failed.Count)/$($cases.Count) 个：" -ForegroundColor Red
	$failed | ForEach-Object { Write-Host "  $_" -ForegroundColor Red }
	exit 1
}
if ($Record) {
	Write-Host "已重录 $($cases.Count) 个用例的期望 —— 请人工核对一遍" -ForegroundColor Green
	exit 0
}
Write-Host "全部 $($cases.Count) 个用例与期望一致" -ForegroundColor Green
exit 0
