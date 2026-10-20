<#
	一次跑完"坏文件中止加载"的四个场景。

	每个场景的 assets 根里都有三个文件：
	  aaa_good.json   —— 好的配方（在坏文件之前）
	  bbb_bad.json    —— 坏文件
	  ccc_good.json   —— 好的配方（在坏文件之后）

	期望：aaa_good 是已知的配方，bbb_bad 和 ccc_good 都是 unknown。
	场景：
	  broken     —— 文件不是 JSON
	  badtype    —— type 不是 1.12.2 认识的那两个
	  unknownitem—— 结果的物品 1.12.2 不认识
	  nodata     —— 结果的物品有多个形态，却没写 data
#>
param(
	[int]$Port = 25880,
	[string]$Java = 'D:\tools\jdk\JDK_1.8\bin\java.exe'
)

$ErrorActionPreference = 'Stop'
$probe = Join-Path $PSScriptRoot 'probe-assets.ps1'
$base = Join-Path $env:TEMP "gomc-assets-bad-$Port"

$scenarios = @{
	broken      = '{"type": "crafting_shaped", "result":'
	badtype     = '{"type": "minecraft:crafting_shaped", "pattern": ["#"], "key": {"#": {"item": "minecraft:stick"}}, "result": {"item": "minecraft:emerald"}}'
	unknownitem = '{"type": "crafting_shaped", "pattern": ["#"], "key": {"#": {"item": "minecraft:stick"}}, "result": {"item": "minecraft:not_a_real_item"}}'
	nodata      = '{"type": "crafting_shaped", "pattern": ["#"], "key": {"#": {"item": "minecraft:stick"}}, "result": {"item": "minecraft:dye"}}'
}
$good = '{"type": "crafting_shaped", "pattern": ["#"], "key": {"#": {"item": "minecraft:stick"}}, "result": {"item": "minecraft:diamond_block"}}'

$cmdFile = Join-Path $env:TEMP "gomc-bad-$Port.txt"
@(
	'say B1'
	'recipe give Steve aaa_good'
	'say B2'
	'recipe give Steve bbb_bad'
	'say B3'
	'recipe give Steve ccc_good'
	'say B4'
) | Set-Content -Path $cmdFile -Encoding ascii

$p = $Port
foreach ($name in @('broken', 'badtype', 'unknownitem', 'nodata')) {
	$root = Join-Path $base $name
	if (Test-Path $root) { Remove-Item -Recurse -Force $root }
	New-Item -ItemType Directory -Force -Path (Join-Path $root 'assets\minecraft\recipes') | Out-Null
	New-Item -ItemType Directory -Force -Path (Join-Path $root 'assets\minecraft\advancements') | Out-Null
	Set-Content -Path (Join-Path $root 'assets\.mcassetsroot') -Value $root -Encoding ascii
	Set-Content -Path (Join-Path $root 'assets\minecraft\recipes\aaa_good.json') -Value $good -Encoding ascii
	Set-Content -Path (Join-Path $root 'assets\minecraft\recipes\bbb_bad.json') -Value $scenarios[$name] -Encoding ascii
	Set-Content -Path (Join-Path $root 'assets\minecraft\recipes\ccc_good.json') -Value $good -Encoding ascii

	Write-Host ""
	Write-Host "######## 场景 $name ########" -ForegroundColor Magenta
	& pwsh -NoProfile -ExecutionPolicy Bypass -File $probe -Assets $root -Commands $cmdFile -Port $p -Bot Steve 2>&1 |
		Where-Object { $_ -match '^\[Server\]|^B[0-9]$|unknown recipe|recipe for|Parsing error|Errors with built-in|Unknown item|Missing data|Invalid or unsupported' } |
		ForEach-Object { Write-Host $_ }
	$p++
}
