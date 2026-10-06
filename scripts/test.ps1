# PowerShell verification script for Quant Methods Practice
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$rootDir = Split-Path -Parent $scriptDir

# $ErrorActionPreference does not apply to native executables in Windows
# PowerShell 5.1, so every native step must check its exit code explicitly.
function Invoke-Checked {
    param([string]$Label, [scriptblock]$Command)
    & $Command
    if ($LASTEXITCODE -ne 0) {
        Write-Host "==> FAILED: $Label (exit code $LASTEXITCODE)" -ForegroundColor Red
        exit $LASTEXITCODE
    }
}

Write-Host "==> Checking documentation and scaffold..." -ForegroundColor Cyan
Invoke-Checked "scaffold validation" { python "$rootDir\scripts\validate_scaffold.py" }

Write-Host "==> Running frontend typecheck, lint & unit tests..." -ForegroundColor Cyan
Push-Location "$rootDir\web"
try {
    Invoke-Checked "frontend typecheck" { npm run typecheck }
    Invoke-Checked "frontend lint" { npm run lint }
    Invoke-Checked "frontend unit tests" { npm run test }
} finally {
    Pop-Location
}

# Browser tests drive bin\quant-practice.exe with embedded assets, so build
# first; otherwise they silently exercise a stale UI.
Write-Host "==> Building frontend assets and Go binary..." -ForegroundColor Cyan
Invoke-Checked "build" { powershell -ExecutionPolicy Bypass -File "$rootDir\scripts\build.ps1" }

Write-Host "==> Running browser end-to-end tests..." -ForegroundColor Cyan
Push-Location "$rootDir\web"
try {
    Invoke-Checked "browser end-to-end tests" { npm run test:e2e }
} finally {
    Pop-Location
}

Write-Host "==> Running Go formatting check, unit tests and vet..." -ForegroundColor Cyan
Push-Location "$rootDir"
try {
    $unformatted = gofmt -s -l cmd internal
    if ($unformatted) {
        Write-Host "==> FAILED: gofmt reports unformatted files:`n$unformatted" -ForegroundColor Red
        exit 1
    }
    Invoke-Checked "go test" { go test -count=1 ./internal/... ./cmd/... }
    Invoke-Checked "go vet" { go vet ./internal/... ./cmd/... }
} finally {
    Pop-Location
}

Write-Host "==> All verification checks PASSED!" -ForegroundColor Green
