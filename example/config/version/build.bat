@echo off
setlocal enabledelayedexpansion

:: 设置脚本目录
set "SCRIPT_DIR=%~dp0"
set "SCRIPT_DIR=%SCRIPT_DIR:~0,-1%"

:: 设置项目根目录（向上三级）
for %%I in ("%SCRIPT_DIR%\..\..\..") do set "PROJECT_ROOT=%%~fI"

:: 设置源文件和输出目录
set "SOURCE_FILE=%SCRIPT_DIR%\main.go"
set "OUTPUT_DIR=%PROJECT_ROOT%\bin\example\config\version"

:: 创建输出目录
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

:: 获取版本信息函数
:get_version_info
set "VERSION=1.0.0"
:: 获取构建时间（格式：YYYYMMDDHHmmss000）
for /f "tokens=2 delims==" %%I in ('wmic os get localdatetime /value') do set "datetime=%%I"
set "BUILD_TIME=!datetime:~0,14!000"

:: 获取 Git 提交信息
for /f "tokens=*" %%a in ('git rev-parse HEAD') do set "GIT_COMMIT=%%a"
pushd "%PROJECT_ROOT%"
for /f "tokens=*" %%a in ('git rev-parse HEAD') do set "LIB_GIT_COMMIT=%%a"
set "LIBRARY_DIR=%PROJECT_ROOT%"
popd
set "WORKING_DIR=%SCRIPT_DIR%"

:: 获取 Go 环境信息
for /f "tokens=*" %%a in ('go env GOPATH') do set "GOPATH=%%a"
for /f "tokens=*" %%a in ('go env GOROOT') do set "GOROOT=%%a"
goto :eof

:: 主函数
:main
echo 开始构建...
echo 源文件：%SOURCE_FILE%
echo 输出目录：%OUTPUT_DIR%
echo.

:: 获取版本信息
call :get_version_info

:: 构建当前平台的可执行文件
echo 正在构建...
go build -ldflags "-X github.com/fsyyft-go/kit/go/build.version=%VERSION% ^
                  -X github.com/fsyyft-go/kit/go/build.gitVersion=%GIT_COMMIT% ^
                  -X github.com/fsyyft-go/kit/go/build.libGitVersion=%LIB_GIT_COMMIT% ^
                  -X github.com/fsyyft-go/kit/go/build.buildTimeString=%BUILD_TIME% ^
                  -X github.com/fsyyft-go/kit/go/build.buildLibraryDirectory=%LIBRARY_DIR% ^
                  -X github.com/fsyyft-go/kit/go/build.buildWorkingDirectory=%WORKING_DIR% ^
                  -X github.com/fsyyft-go/kit/go/build.buildGopathDirectory=%GOPATH% ^
                  -X github.com/fsyyft-go/kit/go/build.buildGorootDirectory=%GOROOT%" ^
         -o "%OUTPUT_DIR%\version.exe" "%SOURCE_FILE%"
echo 构建完成！二进制文件位于：%OUTPUT_DIR%\version.exe
echo 文件列表：
dir "%OUTPUT_DIR%"

echo.
echo 运行构建的程序：
"%OUTPUT_DIR%\version.exe"
goto :eof

:: 执行主函数
call :main 