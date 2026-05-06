@echo off
echo === Running Tick Chart ===

echo Setting environment variables...
set "LOCALAPPDATA=%~dp0.local-appdata"

echo Starting application...
npm start
