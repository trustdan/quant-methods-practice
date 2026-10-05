# PowerShell launch script for Quant Methods Practice
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$rootDir = Split-Path -Parent $scriptDir
$binPath = "$rootDir\bin\quant-practice.exe"

if (-not (Test-Path $binPath)) {
    Write-Host "Binary not found at $binPath. Running build first..." -ForegroundColor Yellow
    & "$scriptDir\build.ps1"
}

Write-Host "==> Launching Quant Methods Practice..." -ForegroundColor Cyan
& $binPath @args
