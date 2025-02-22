@echo off
setlocal enabledelayedexpansion

rem 设置编码为 UTF-8
chcp 65001 > nul

rem 脚本所在目录
set "SCRIPT_DIR=%~dp0"
rem 删除路径末尾的反斜杠
set "SCRIPT_DIR=%SCRIPT_DIR:~0,-1%"

rem 项目根目录（向上两级）
for %%I in ("%SCRIPT_DIR%\..\..\") do set "PROJECT_ROOT=%%~fI"

rem 源文件路径
set "SOURCE_FILE=%SCRIPT_DIR%\main.go"
rem 输出目录
set "OUTPUT_DIR=%PROJECT_ROOT%\bin\example\cache"

rem 创建输出目录
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

rem 主函数
:main
echo 开始构建...
echo 源文件：%SOURCE_FILE%
echo 输出目录：%OUTPUT_DIR%
echo.

rem 构建当前平台的可执行文件
echo 正在构建...
go build -o "%OUTPUT_DIR%\cache.exe" "%SOURCE_FILE%"
echo 构建完成！二进制文件位于：%OUTPUT_DIR%\cache.exe
echo 文件列表：
dir "%OUTPUT_DIR%"

echo.
echo 运行构建的程序：
"%OUTPUT_DIR%\cache.exe"

endlocal 