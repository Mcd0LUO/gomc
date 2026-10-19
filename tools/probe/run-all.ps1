<#
.SYNOPSIS
	跑完 tools\probe\cases 下的全部用例，逐个和 .expected 比对。

.DESCRIPTION
	这是"实测基准"的回归套件：每个用例都在真实 1.12.2 服务端上跑一遍，
	和录好的 .expected 比对。服务端行为变了（换了服务端版本、或者
	期望文件录错了）就会红。

	每个用例都要起一次服务端，约 40 秒一个。

.PARAMETER Case
	只跑指定的用例（文件名或路径），默认全部。

.PARAMETER Work
	隔离环境的落地目录。

.PARAMETER Filter
	只跑文件名包含这个串的用例。

.EXAMPLE
	.\run-all.ps1
	.\run-all.ps1 -Filter execute
#>
[CmdletBinding()]
param(
	[string]$Case,
	[string]$Work = (Join-Path $env:TEMP 'mc122probe'),
	[string]$Filter = ''
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

$failed = @()
foreach ($c in $cases) {
	Write-Host ""
	Write-Host "######## $($c.Name) ########" -ForegroundColor Magenta
	& pwsh -NoProfile -ExecutionPolicy Bypass -File $probe -Case $c.FullName -Work $Work -Compare
	if ($LASTEXITCODE -ne 0) { $failed += $c.Name }
}

Write-Host ""
if ($failed.Count -gt 0) {
	Write-Host "失败 $($failed.Count)/$($cases.Count) 个：" -ForegroundColor Red
	$failed | ForEach-Object { Write-Host "  $_" -ForegroundColor Red }
	exit 1
}
Write-Host "全部 $($cases.Count) 个用例与期望一致" -ForegroundColor Green
exit 0
