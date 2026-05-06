@echo off
echo === Building Tick Chart ===

echo Stopping old processes...
taskkill /F /IM tick-chart.exe >nul 2>&1
taskkill /F /IM fen-shi-tu.exe >nul 2>&1
timeout /t 2 /nobreak >nul

echo Cleaning old output...
if exist dist (
    rmdir /S /Q dist
)

echo Setting environment variables...
set "LOCALAPPDATA=%~dp0.local-appdata"
set "CSC_IDENTITY_AUTO_DISCOVERY=false"

echo Building...
call npm run build:dir

if %ERRORLEVEL% EQU 0 (
    echo.
    echo === Build Success ===
    echo Output: dist\win-unpacked\
) else (
    echo.
    echo === Build Failed ===
    pause
    exit /b 1
)

pause
