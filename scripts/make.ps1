# Videopress 构建入口：保证 frontend/dist 在 Go 编译前已生成（frontend/embed.go 依赖）。
param(
    [Parameter(Position = 0)]
    [ValidateSet('help', 'frontend-deps', 'frontend-dist', 'wails-bindings', 'frontend-check', 'test', 'vet', 'build-go', 'build', 'ci')]
    [string]$Target = 'help'
)

$ErrorActionPreference = 'Stop'
$Root = Resolve-Path (Join-Path $PSScriptRoot '..')

function Invoke-FrontendDeps {
    Push-Location (Join-Path $Root 'frontend')
    try {
        npm ci
    } finally {
        Pop-Location
    }
}

function Invoke-FrontendBuild {
    Push-Location (Join-Path $Root 'frontend')
    try {
        if (-not (Test-Path 'node_modules')) {
            npm ci
        }
        npm run build
    } finally {
        Pop-Location
    }
}

function Invoke-WailsBindings {
    Push-Location $Root
    try {
        wails generate module
    } finally {
        Pop-Location
    }
}

function Invoke-FrontendCheck {
    Push-Location (Join-Path $Root 'frontend')
    try {
        npm run check
        npm run build
    } finally {
        Pop-Location
    }
}

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

switch ($Target) {
    'help' {
        Write-Host @"
Videopress build targets (also: make <target>)

  frontend-deps    npm ci in frontend/
  frontend-dist    build frontend/dist (required before any go command)
  wails-bindings   wails generate module (needs frontend-dist)
  frontend-check   svelte-check + rebuild after bindings
  test             frontend-dist + go test ./...
  vet              frontend-dist + go vet ./...
  build-go         frontend-dist + go build ./...
  build            wails build -> build/bin/videopress.exe
  ci               full CI pipeline (deps, dist, bindings, check, go)
"@
    }
    'frontend-deps' { Invoke-FrontendDeps }
    'frontend-dist' { Invoke-FrontendBuild }
    'wails-bindings' {
        Invoke-FrontendBuild
        Invoke-WailsBindings
    }
    'frontend-check' {
        Invoke-FrontendBuild
        Invoke-WailsBindings
        Invoke-FrontendCheck
    }
    'test' {
        Invoke-FrontendBuild
        Invoke-Go test './...'
    }
    'vet' {
        Invoke-FrontendBuild
        Invoke-Go vet './...'
    }
    'build-go' {
        Invoke-FrontendBuild
        Invoke-Go build './...'
    }
    'build' {
        Push-Location $Root
        try {
            wails build
            if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
        } finally {
            Pop-Location
        }
    }
    'ci' {
        Invoke-FrontendDeps
        Invoke-FrontendBuild
        Invoke-WailsBindings
        Invoke-FrontendCheck
        Invoke-Go build './...'
        Invoke-Go vet './...'
        Invoke-Go test './...'
    }
}
