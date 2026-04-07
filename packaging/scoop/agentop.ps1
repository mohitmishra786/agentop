@echo off
setlocal EnableDelayedExpansion

echo agentop
echo ========

set "VERSION=0.1.0"
set "RELEASE_DIR=https://github.com/mohitmishra786/agentop/releases/download/v%VERSION%"

if "%1"=="/install" goto install
if "%1"=="/upgrade" goto upgrade
goto :eof

:install
echo Downloading agentop...
curl -sSL "%RELEASE_DIR%/agentop_%VERSION%_windows_amd64.zip" -o agentop.zip
echo Extracting...
tar -xf agentop.zip
echo Installing...
copy agentop.exe "%USERPROFILE%\scoop\shims\" >nul 2>&1
echo Cleanup...
del agentop.zip
echo agentop installed!
goto :eof

:upgrade
echo Upgrading agentop...
curl -sSL "%RELEASE_DIR%/agentop_%VERSION%_windows_amd64.zip" -o agentop.zip
echo Extracting...
tar -xf agentop.zip
echo Installing...
copy /Y agentop.exe "%USERPROFILE%\scoop\shims\" >nul 2>&1
echo Cleanup...
del agentop.zip
echo agentop upgraded to version %VERSION%!
goto :eof
