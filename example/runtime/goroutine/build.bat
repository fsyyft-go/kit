@echo off
setlocal enabledelayedexpansion

:: 设置路径
set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%SCRIPT_DIR%..\..\..\"
set "SOURCE_FILE=%SCRIPT_DIR%main.go"
set "OUTPUT_DIR=%PROJECT_ROOT%bin\example\runtime\goroutine"

:: 创建输出目录
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

echo Starting cross-platform build...
echo Source file: %SOURCE_FILE%
echo Output directory: %OUTPUT_DIR%
echo.

:: Windows 平台
call :build windows amd64 .exe
call :build windows arm64 .exe
call :build windows 386 .exe

:: Linux 平台
call :build linux amd64
call :build linux arm64
call :build linux arm
call :build linux 386
call :build linux mips
call :build linux mips64
call :build linux mips64le
call :build linux ppc64
call :build linux ppc64le
call :build linux s390x
call :build linux riscv64

:: Darwin 平台
call :build darwin amd64
call :build darwin arm64

echo.
echo Build complete! Binaries are available in: %OUTPUT_DIR%
echo Files:
dir /B "%OUTPUT_DIR%"

goto :eof

:build
set "GOOS=%~1"
set "GOARCH=%~2"
set "SUFFIX=%~3"
set "OUTPUT=%OUTPUT_DIR%\goid_%GOOS%_%GOARCH%%SUFFIX%"

echo Building for %GOOS%/%GOARCH%...
set "GOOS=%GOOS%"
set "GOARCH=%GOARCH%"
go build -o "%OUTPUT%" "%SOURCE_FILE%"
echo Built: %OUTPUT%
goto :eof 