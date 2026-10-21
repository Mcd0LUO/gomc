# cmp.ps1 -- run one difftest case against BOTH worlds and diff the state
#
# The GGM build needs its generated init (objectives + constant pool) to run
# first; the original pack assumes sys/scoreboard.mcfunction already created
# the objectives.
#
# -GgmInit lists the init functions to call in the GGM world.
# -Ignore lists regular-expression patterns for scoreboard lines that are
# compiler scaffolding (ggm.c constants, t0/l0 temporaries) rather than
# author data, so the comparison is about SEMANTICS, not about scratch space.
param(
  [Parameter(Mandatory=$true)][string]$Case,
  [string[]]$GgmInit = @(),
  [string[]]$Ignore = @('ggm\.c/', '^\s+(t\d+|l\d+)(\.ggm)?/')
)
$ErrorActionPreference='Stop'
$rt='D:\code\go\GoMcRuntime'
$wA="$rt\testdata\port-orig"
$wB="$rt\testdata\port-ggm"
$casePath=$Case
if (-not [System.IO.Path]::IsPathRooted($casePath)) { $casePath="$rt\testdata\port-cases\$Case" }

function State($world, $extraSetup) {
  $tmp = $casePath
  if ($extraSetup.Count -gt 0) {
    $raw = Get-Content $casePath -Raw
    $lines = $extraSetup | ForEach-Object { "  $_" }
    $raw = $raw -replace '(?m)^setup:\s*$', ("setup:" + "`n" + ($lines -join "`n"))
    $tmp = [System.IO.Path]::GetTempFileName()
    [System.IO.File]::WriteAllText($tmp, $raw)
  }
  $o = & "$rt\gomc.exe" difftest run $world $tmp --full 2>&1 | Out-String
  if ($tmp -ne $casePath) { Remove-Item $tmp -Force -ErrorAction SilentlyContinue }
  $i = $o.IndexOf('state:')
  if ($i -lt 0) { return "<<NO STATE>>`n$o" }
  $body = $o.Substring($i)
  foreach ($pat in $Ignore) {
    $body = ($body -split "`r?`n" | Where-Object { $_ -notmatch $pat }) -join "`n"
  }
  return $body
}

$a = State $wA @()
$b = State $wB $GgmInit
if ($a -eq $b) {
  Write-Output '=== IDENTICAL ==='
  Write-Output $a
} else {
  Write-Output '=== DIFFERENT ==='
  Write-Output '--- ORIGINAL ---'
  Write-Output $a
  Write-Output '--- GGM ---'
  Write-Output $b
  Write-Output '--- DIFF (line) ---'
  $la = $a -split "`r?`n"; $lb = $b -split "`r?`n"
  Compare-Object $la $lb | ForEach-Object { "{0} {1}" -f $_.SideIndicator, $_.InputObject }
}
