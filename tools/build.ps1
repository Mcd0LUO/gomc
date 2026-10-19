<#
.SYNOPSIS
	构建 gomc，并把版本号、提交号、构建时间写进二进制。

.DESCRIPTION
	版本号遵循语义化版本（semver）：
	  - 有 git tag 时用 tag（如 v0.1.0）
	  - 没有 tag 时用 internal/buildinfo 里的默认值
	提交号取 HEAD 的短 hash；工作区有未提交改动时加 "-dirty" 后缀，
	这样拿到一个二进制就知道它是不是干净的构建。

.EXAMPLE
	.\build.ps1
	.\build.ps1 -Version v0.2.0
	.\build.ps1 -Output ..\gomc.exe
#>
[CmdletBinding()]
param(
	[string]$Output = 'gomc.exe',
	[string]$Version = ''
)

$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
Push-Location $root
try {
	if (-not $Version) {
		$tag = git describe --tags --abbrev=0 2>$null
		if ($tag) { $Version = $tag } else { $Version = '0.1.0' }
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
