# Videopress 构建入口。逻辑在本脚本；Windows 可直接运行：
#   pwsh -File scripts/make.ps1 test
param(
    [Parameter(Position = 0)]
    [ValidateSet('help', 'test', 'vet', 'build-go', 'dev', 'build', 'ci')]
    [string]$Target = 'help'
)

$ErrorActionPreference = 'Stop'
$Root = Resolve-Path (Join-Path $PSScriptRoot '..')

function Invoke-Go {
    param([string[]]$GoArgs)
    Push-Location $Root
    try {
        & go @GoArgs
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    } finally {
        Pop-Location
    }
}

function Invoke-Mygo {
    param([string[]]$MygoArgs)
    Push-Location $Root
    try {
        & go tool mygo @MygoArgs
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    } finally {
        Pop-Location
    }
}

switch ($Target) {
    'help' {
        Write-Host @"
Videopress build targets (also: make <target>)

  test         go test ./...
  vet          go vet ./...
  build-go     go build ./...
  dev          go tool mygo dev (hot-reloading dev window)
  build        go tool mygo build -> dist/windows-amd64/ (exe + installer)
  ci           vet, test and build
"@
    }
    'test' { Invoke-Go test './...' }
    'vet' { Invoke-Go vet './...' }
    'build-go' { Invoke-Go build './...' }
    'dev' { Invoke-Mygo dev }
    'build' { Invoke-Mygo build }
    'ci' {
        Invoke-Go vet './...'
        Invoke-Go test './...'
        Invoke-Go build './...'
    }
}
