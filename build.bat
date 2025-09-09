@echo off
setlocal enableextensions
 
REM Always run from the script's own folder
pushd "%~dp0"
 
REM Load .env into environment
for /F %%A in (.env) do set %%A
 
REM Define the Nmap zip file name
set "NMAP_WIN_ZIP=nmap-%NMAP_VERSION%-win32.zip"
 
REM Check if Nmap zip exists; download if missing
if not exist assets\%NMAP_WIN_ZIP% (
    echo Nmap zip not found, downloading...
    powershell -Command "Invoke-WebRequest https://nmap.org/dist/%NMAP_WIN_ZIP% -OutFile assets\%NMAP_WIN_ZIP%"
) else (
    echo Nmap zip already exists: assets\%NMAP_WIN_ZIP%
)
 
REM Choose package path (root by default; use cmd\discovr if it exists)
set "PKG=."
if exist "cmd\discovr\main.go" set "PKG=cmd\discovr"
 
echo Using package: %PKG%
echo NMAP_VERSION=%NMAP_VERSION%
 
REM Build (verbose)
go build -v -ldflags="-X github.com/Naman1997/discovr/internal.NmapVersion=%NMAP_VERSION%" -o discovr.exe %PKG%
echo Exit code: %ERRORLEVEL%
 
if errorlevel 1 (
  echo Build failed
) else (
  echo Build complete: discovr.exe
)
 
popd
endlocal
 