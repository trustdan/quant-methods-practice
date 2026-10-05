# PowerShell build script for Quant Methods Practice
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$rootDir = Split-Path -Parent $scriptDir

Write-Host "==> Building frontend assets..." -ForegroundColor Cyan
Push-Location "$rootDir\web"
try {
    npm run build
} finally {
    Pop-Location
}

Write-Host "==> Compiling Go binary..." -ForegroundColor Cyan
if (-not (Test-Path "$rootDir\bin")) {
    New-Item -ItemType Directory -Path "$rootDir\bin" | Out-Null
}

Push-Location "$rootDir"
try {
    go build -o "$rootDir\bin\quant-practice.exe" ./cmd/quant-practice
} finally {
    Pop-Location
}

Write-Host "==> Build complete: bin\quant-practice.exe" -ForegroundColor Green
