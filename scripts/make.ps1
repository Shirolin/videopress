# Videopress 构建入口。逻辑在本脚本；Windows 可直接运行：
#   pwsh -File scripts/make.ps1 test
param(
    [Parameter(Position = 0)]
    [ValidateSet('help', 'test', 'vet', 'build-go', 'dev', 'build', 'ci')]
    [string]$Target = 'help'
)

$ErrorActionPreference = 'Stop'
$Root = Resolve-Path (Join-Path $PSScriptRoot '..')
$NullDevice = if ($IsWindows) { 'NUL' } else { '/dev/null' }

function Invoke-Go {
    param(
        [Parameter(ValueFromRemainingArguments)]
        [string[]]$GoArgs
    )
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

function Assert-GuiExecutable {
    param([string]$Path)
    if (-not (Test-Path $Path)) {
        throw "产物文件不存在: $Path"
    }
    $bytes = [System.IO.File]::ReadAllBytes($Path)
    $e_lfanew = [BitConverter]::ToInt32($bytes, 0x3C)
    $subsystem = [BitConverter]::ToUInt16($bytes, $e_lfanew + 0x18 + 0x44)
    if ($subsystem -ne 2) {
        throw "产物子系统校验失败: $Path (Subsystem=$subsystem, 期望 2: IMAGE_SUBSYSTEM_WINDOWS_GUI)"
    }
    Add-Type -AssemblyName System.Drawing -ErrorAction SilentlyContinue
    $icon = [System.Drawing.Icon]::ExtractAssociatedIcon((Resolve-Path $Path))
    if ($null -eq $icon) {
        throw "产物图标校验失败: $Path (未能提取到有效应用图标)"
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
    'test' { Invoke-Go @('test', './...') }
    'vet' { Invoke-Go @('vet', './...') }
    'build-go' { Invoke-Go @('build', '-o', $NullDevice, './...') }
    'dev' { Invoke-Mygo dev }
    'build' {
        Invoke-Mygo build
        $distExe = Join-Path $Root 'dist\windows-amd64\Videopress.exe'
        if (Test-Path $distExe) {
            Assert-GuiExecutable $distExe
            $rootExe = Join-Path $Root 'videopress.exe'
            Copy-Item $distExe $rootExe -Force
            Assert-GuiExecutable $rootExe
        }
    }
    'ci' {
        Invoke-Go @('vet', './...')
        Invoke-Go @('test', './...')
        Invoke-Go @('build', '-o', $NullDevice, './...')
        $rootExe = Join-Path $Root 'videopress.exe'
        if (Test-Path $rootExe) {
            Assert-GuiExecutable $rootExe
        }
    }
}
