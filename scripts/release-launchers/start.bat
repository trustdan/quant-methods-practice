@echo off
setlocal
cd /d "%~dp0"
if not exist "%~dp0quant-practice.exe" (
    echo Error: quant-practice.exe not found in this directory.
    pause
    exit /b 1
)
echo Launching Quant Methods Practice...
"%~dp0quant-practice.exe" %*
