@echo off
setlocal
set SCRIPT_DIR=%~dp0
set ROOT_DIR=%SCRIPT_DIR%..
set BIN_PATH=%ROOT_DIR%\bin\quant-practice.exe

if not exist "%BIN_PATH%" (
    echo Binary not found at %BIN_PATH%. Building first...
    powershell -ExecutionPolicy Bypass -File "%SCRIPT_DIR%build.ps1"
)

echo Starting Quant Methods Practice...
"%BIN_PATH%" %*
