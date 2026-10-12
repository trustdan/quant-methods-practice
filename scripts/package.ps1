# Windows PowerShell Packaging Script for Quant Methods Practice
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$rootDir = Split-Path -Parent $scriptDir
$version = "0.1.0"
$distDir = Join-Path $rootDir "dist\release"
$stagingBase = Join-Path $rootDir "dist\staging"

Write-Host "=========================================================="
Write-Host "  Quant Methods Practice -- Multi-Platform Release Packager"
Write-Host "  Version: $version"
Write-Host "=========================================================="

# 1. Build frontend
Write-Host "`n==> [1/4] Building production frontend assets..."
Push-Location (Join-Path $rootDir "web")
npm run build
Pop-Location

# 2. Prepare directories
if (Test-Path $distDir) { Remove-Item -Recurse -Force $distDir }
if (Test-Path $stagingBase) { Remove-Item -Recurse -Force $stagingBase }
New-Item -ItemType Directory -Path $distDir | Out-Null
New-Item -ItemType Directory -Path $stagingBase | Out-Null

# 3. Target builds
$targets = @(
    @{ OS="windows"; Arch="amd64"; Bin="quant-practice.exe"; Launcher="start.bat"; Ext="zip" },
    @{ OS="darwin"; Arch="arm64"; Bin="quant-practice"; Launcher="start.sh"; Ext="tar.gz" },
    @{ OS="darwin"; Arch="amd64"; Bin="quant-practice"; Launcher="start.sh"; Ext="tar.gz" },
    @{ OS="linux"; Arch="amd64"; Bin="quant-practice"; Launcher="start.sh"; Ext="tar.gz" },
    @{ OS="linux"; Arch="arm64"; Bin="quant-practice"; Launcher="start.sh"; Ext="tar.gz" }
)

Push-Location $rootDir
$env:CGO_ENABLED = "0"

foreach ($t in $targets) {
    $pkgName = "quant-practice-v$version-$($t.OS)-$($t.Arch)"
    $pkgDir = Join-Path $stagingBase $pkgName
    New-Item -ItemType Directory -Path $pkgDir | Out-Null

    Write-Host "  --> Compiling $($t.OS)/$($t.Arch)..."
    $env:GOOS = $t.OS
    $env:GOARCH = $t.Arch
    go build -ldflags "-s -w" -o (Join-Path $pkgDir $t.Bin) ./cmd/quant-practice

    if ($t.Launcher -eq "start.bat") {
        Copy-Item (Join-Path $rootDir "scripts\release-launchers\start.bat") (Join-Path $pkgDir "start.bat")
    } else {
        Copy-Item (Join-Path $rootDir "scripts\release-launchers\start.sh") (Join-Path $pkgDir "start.sh")
    }

    Copy-Item (Join-Path $rootDir "scripts\release-launchers\README.txt") (Join-Path $pkgDir "README.txt")
    Copy-Item (Join-Path $rootDir "LICENSE") (Join-Path $pkgDir "LICENSE")
    Copy-Item (Join-Path $rootDir "NOTICE") (Join-Path $pkgDir "NOTICE")

    $archiveFile = Join-Path $distDir "$pkgName.$($t.Ext)"
    Write-Host "  --> Packaging $archiveFile..."
    if ($t.Ext -eq "zip") {
        Compress-Archive -Path "$pkgDir\*" -DestinationPath $archiveFile -Force
    } else {
        tar -czf $archiveFile -C $stagingBase $pkgName
    }
}
Pop-Location

# 4. Checksums
Write-Host "`n==> [4/4] Generating SHA-256 Checksums..."
Push-Location $distDir
$checksums = Get-ChildItem "quant-practice-*" | ForEach-Object {
    $hash = (Get-FileHash -Algorithm SHA256 $_.FullName).Hash.ToLower()
    "$hash  $($_.Name)"
}
$checksums | Out-File -Encoding ascii "SHA256SUMS.txt"
Get-Content "SHA256SUMS.txt"
Pop-Location
Write-Host ""
Write-Host "Release packaging complete."
