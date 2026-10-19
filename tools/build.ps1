<#
.SYNOPSIS
	构建 gomc，并把版本号、提交号、构建时间写进二进制。

.DESCRIPTION
	版本号遵循语义化版本（semver），由 git 历史推导：
	  - 正好在最新的 tag 上：就是那个 tag（如 0.1.0）
	  - 在 tag 之后：下一个版本的预发布号（如 0.1.1-dev.3.gabc1234）
	  - 还没有任何 tag：0.1.0
	提交号取 HEAD 的短 hash；工作区有未提交改动时加 "-dirty" 后缀，
	这样拿到一个二进制就知道它是不是干净的构建。

.EXAMPLE
	.\build.ps1
	.\build.ps1 -Version 0.2.0
	.\build.ps1 -Output ..\gomc.exe
#>
[CmdletBinding()]
param(
	[string]$Output = 'gomc.exe',
	[string]$Version = ''
)

$ErrorActionPreference = 'Stop'

# Get-Version derives the version from the git history.
#
# Using the latest tag unconditionally would report the same version for
# every build, which says nothing about what the binary actually is.
function Get-Version {
	$tag = git describe --tags --abbrev=0 2>$null
	if (-not $tag) { return '0.1.0' }
	$base = $tag -replace '^v', ''
	$ahead = 0
	$count = git rev-list --count "$tag..HEAD" 2>$null
	if ($count) { $ahead = [int]$count }
	if ($ahead -eq 0) { return $base }
	$parts = $base -split '\.'
	$next = "$($parts[0]).$($parts[1]).$([int]$parts[2] + 1)"
	$short = git rev-parse --short HEAD
	return "$next-dev.$ahead.g$short"
}

$root = Split-Path -Parent $PSScriptRoot
Push-Location $root
try {
	if (-not $Version) {
		$Version = Get-Version
	}
	$commit = git rev-parse --short HEAD
	if (git status --porcelain) { $commit = "$commit-dirty" }
	$date = (Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')

	$ldflags = @(
		"-X github.com/mcsim/gomc/internal/buildinfo.Version=$Version",
		"-X github.com/mcsim/gomc/internal/buildinfo.Commit=$commit",
		"-X github.com/mcsim/gomc/internal/buildinfo.Date=$date"
	) -join ' '

	go build -ldflags $ldflags -o $Output ./cmd/gomc
	if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
	Write-Host "built $Output ($Version, $commit, $date)" -ForegroundColor Green
}
finally {
	Pop-Location
}
