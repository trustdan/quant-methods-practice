# PowerShell Release Package Verification Script for Quant Methods Practice
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$rootDir = Split-Path -Parent $scriptDir
$distDir = Join-Path $rootDir "dist\release"
$version = "0.1.0"
$pkgName = "quant-practice-v$version-windows-amd64"
$archivePath = Join-Path $distDir "$pkgName.zip"

Write-Host "=========================================================="
Write-Host "  Quant Methods Practice -- Windows Release Verification  "
Write-Host "=========================================================="

if (-not (Test-Path $archivePath)) {
    Write-Host "Error: Release archive not found at $archivePath" -ForegroundColor Red
    Write-Host "Run scripts\package.ps1 first." -ForegroundColor Yellow
    exit 1
}

Write-Host "Testing native release package: $pkgName.zip"

# 1. Checksum validation
Write-Host "--> [1/6] Verifying SHA-256 checksum..."
$sumsFile = Join-Path $distDir "SHA256SUMS.txt"
if (-not (Test-Path $sumsFile)) {
    Write-Host "Error: SHA256SUMS.txt not found at $sumsFile" -ForegroundColor Red
    exit 1
}
$sumLines = Get-Content $sumsFile
$expectedLine = $sumLines | Where-Object { $_ -match "$pkgName\.zip" }
if (-not $expectedLine) {
    Write-Host "Error: Checksum entry not found in SHA256SUMS.txt for $pkgName.zip" -ForegroundColor Red
    exit 1
}
$expectedHash = ($expectedLine -split "\s+")[0].ToLower()
$actualHash = (Get-FileHash -Algorithm SHA256 $archivePath).Hash.ToLower()

if ($expectedHash -ne $actualHash) {
    Write-Host "ERROR: Checksum mismatch! Expected $expectedHash, got $actualHash" -ForegroundColor Red
    exit 1
}
Write-Host "    Checksum OK: $actualHash" -ForegroundColor Green

# 2. Isolated extraction into temporary sandbox
$tempBase = [System.IO.Path]::GetTempPath()
$sandbox = Join-Path $tempBase "quant-verify-$([System.Guid]::NewGuid().ToString('N').Substring(0, 8))"
New-Item -ItemType Directory -Path $sandbox | Out-Null

try {
    Write-Host "--> [2/6] Extracting into isolated sandbox ($sandbox)..."
    Expand-Archive -Path $archivePath -DestinationPath $sandbox -Force

    # Check required files
    $requiredFiles = @("quant-practice.exe", "start.bat", "README.txt", "LICENSE", "NOTICE")
    foreach ($rf in $requiredFiles) {
        $fullPath = Join-Path $sandbox $rf
        if (-not (Test-Path $fullPath)) {
            Write-Host "ERROR: Missing required file in archive: $rf" -ForegroundColor Red
            exit 1
        }
    }

    # Check hygiene (no source code, test files, sqlite databases, or env files)
    $forbidden = Get-ChildItem -Path $sandbox -Recurse -File | Where-Object {
        $_.Extension -in @(".go", ".ts", ".tsx", ".db", ".env")
    }
    if ($forbidden) {
        Write-Host "ERROR: Archive contains source or private files: $($forbidden.Name -join ', ')" -ForegroundColor Red
        exit 1
    }
    Write-Host "    Package structure and hygiene verified." -ForegroundColor Green

    # 3. Test CLI offline commands in sandbox
    Write-Host "--> [3/6] Testing CLI commands in isolated sandbox..."
    Push-Location $sandbox
    try {
        $exe = ".\quant-practice.exe"
        $dataDir = ".\data"

        # Version check
        $verOut = & $exe -version
        if ($verOut -notmatch "0\.1\.0") {
            Write-Host "ERROR: Unexpected version output: $verOut" -ForegroundColor Red
            exit 1
        }

    # Diagnostics check
    Write-Host "    Running offline doctor & diagnostics..."
    $diagOut = & $exe -data-dir $dataDir -diagnostics
    $diagStr = $diagOut -join "`n"
    if ($diagStr -notmatch "ALL SYSTEMS HEALTHY") {
        Write-Host "ERROR: Diagnostics failed to report healthy:`n$diagStr" -ForegroundColor Red
        exit 1
    }
    if ($diagStr -notmatch "137 approved question templates validated offline") {
        Write-Host "ERROR: Diagnostics question template validation mismatch:`n$diagStr" -ForegroundColor Red
        exit 1
    }
    if ($diagStr -notmatch "Active Bank Resolution.*embedded assets") {
        Write-Host "ERROR: Diagnostics did not resolve active bank from embedded assets:`n$diagStr" -ForegroundColor Red
        exit 1
    }

    # List bank check
    Write-Host "    Verifying 137 curriculum questions from embedded assets..."
    $bankOut = & $exe -list-bank
    $bankStr = $bankOut -join "`n"
    if ($bankStr -notmatch "Approved Curriculum Bank \(embedded assets\): 137 questions") {
        Write-Host "ERROR: List-bank failed to report embedded assets 137 questions:`n$bankStr" -ForegroundColor Red
        exit 1
    }
    if ($bankStr -notmatch "binomial_fair_coin_exactly_two" -or $bankStr -notmatch "regression_sales_advertising") {
        Write-Host "ERROR: Key question templates missing from embedded bank output" -ForegroundColor Red
        exit 1
    }

    # High scores check
    Write-Host "    Verifying arcade leaderboard offline..."
    $hsOut = & $exe -data-dir $dataDir -high-scores
    $hsStr = $hsOut -join "`n"
    if ($hsStr -notmatch "HALL OF FAME") {
        Write-Host "ERROR: High-scores output missing header:`n$hsStr" -ForegroundColor Red
        exit 1
    }

    # Math engine derivation check
    Write-Host "    Verifying analytical math engine offline..."
    $mathOut = & $exe -eval-binomial "n=4,p=0.5,k=2"
    $mathStr = $mathOut -join "`n"
    if ($mathStr -notmatch "0\.375") {
        Write-Host "ERROR: Mathematical engine failed binomial derivation:`n$mathStr" -ForegroundColor Red
        exit 1
    }

    # 4. Backup and progress export tests
    Write-Host "--> [4/6] Testing database backup and progress export..."
    $backupPath = Join-Path $sandbox "backup.db"
    & $exe -data-dir $dataDir -backup $backupPath
    if (-not (Test-Path $backupPath) -or (Get-Item $backupPath).Length -eq 0) {
        Write-Host "ERROR: Backup file missing or empty" -ForegroundColor Red
        exit 1
    }

    $exportPath = Join-Path $sandbox "export.json"
    & $exe -data-dir $dataDir -export-data $exportPath
    if (-not (Test-Path $exportPath) -or (Get-Item $exportPath).Length -eq 0) {
        Write-Host "ERROR: Export file missing or empty" -ForegroundColor Red
        exit 1
    }
    $exportContent = Get-Content $exportPath -Raw
    if ($exportContent -notmatch "exported_at") {
        Write-Host "ERROR: Export file missing exported_at timestamp" -ForegroundColor Red
        exit 1
    }

    # 5. Loopback HTTP Server & UI Verification
    Write-Host "--> [5/6] Testing loopback server and embedded UI..."
    $serverLogOut = Join-Path $sandbox "server.log"
    $serverLogErr = Join-Path $sandbox "server_err.log"
    $proc = Start-Process -FilePath $exe -ArgumentList @("-data-dir", $dataDir, "-port", "0", "-no-browser") `
        -RedirectStandardOutput $serverLogOut -RedirectStandardError $serverLogErr -PassThru

    $port = ""
    for ($i = 0; $i -lt 30; $i++) {
        Start-Sleep -Milliseconds 200
        foreach ($logFile in @($serverLogErr, $serverLogOut)) {
            if (Test-Path $logFile) {
                $logContent = Get-Content $logFile -Raw
                if ($logContent -match "127\.0\.0\.1:(\d+)") {
                    $port = $Matches[1]
                    break
                }
            }
        }
        if ($port) { break }
    }

    if (-not $port) {
        Write-Host "ERROR: Server failed to start on loopback port" -ForegroundColor Red
        if (Test-Path $serverLog) { Get-Content $serverLog }
        Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
        exit 1
    }

    Write-Host "    Server listening on 127.0.0.1:$port (PID $($proc.Id))"

    # Query root HTML
    $resp = Invoke-WebRequest -Uri "http://127.0.0.1:$port/" -UseBasicParsing
    if ($resp.Content -notmatch '<div id="root"></div>' -or $resp.Content -notmatch 'Quant Methods Practice') {
        Write-Host "ERROR: Root HTML did not contain embedded SPA app" -ForegroundColor Red
        Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
        exit 1
    }

    # Query public arcade scores API
    $scoresResp = Invoke-WebRequest -Uri "http://127.0.0.1:$port/api/arcade/scores" -UseBasicParsing
    if ($scoresResp.Content -notmatch '"scores"') {
        Write-Host "ERROR: Arcade scores API did not return valid JSON" -ForegroundColor Red
        Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
        exit 1
    }

    # Stop server
    Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
    Start-Sleep -Milliseconds 300

    Write-Host "--> [6/6] Clean teardown of temporary sandbox..."
    } finally {
        Pop-Location
    }
} finally {
    if (Test-Path $sandbox) {
        Remove-Item -Recurse -Force $sandbox -ErrorAction SilentlyContinue
    }
}

Write-Host ""
Write-Host "==========================================================" -ForegroundColor Green
Write-Host "  RELEASE VERIFICATION PASSED! ALL CHECKS SUCCESSFUL.    " -ForegroundColor Green
Write-Host "==========================================================" -ForegroundColor Green
