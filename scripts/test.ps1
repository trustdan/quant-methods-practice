# PowerShell verification script for Quant Methods Practice
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$rootDir = Split-Path -Parent $scriptDir

Write-Host "==> Checking documentation and scaffold..." -ForegroundColor Cyan
python "$rootDir\scripts\validate_scaffold.py"

Write-Host "==> Running Go unit tests..." -ForegroundColor Cyan
Push-Location "$rootDir"
try {
    go test -v ./internal/... ./cmd/...
    Write-Host "==> Running Go vet..." -ForegroundColor Cyan
    go vet ./internal/... ./cmd/...
} finally {
    Pop-Location
}

Write-Host "==> Running frontend typecheck & tests..." -ForegroundColor Cyan
Push-Location "$rootDir\web"
try {
    npm run typecheck
    npm run test
    npm run test:e2e
} finally {
    Pop-Location
}

Write-Host "==> All verification checks PASSED!" -ForegroundColor Green
