#!/bin/bash

# 设置错误时退出
set -e

# 脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# 项目根目录（向上两级）
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../" && pwd)"
# 源文件路径
SOURCE_FILE="$SCRIPT_DIR/main.go"
# 输出目录
OUTPUT_DIR="$PROJECT_ROOT/bin/example/log"

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 显示日志文件内容
show_logs() {
    echo
    echo "日志文件内容："
    echo "----------------------------------------"
    
    # 显示 app.log 的内容（如果存在）
    if [ -f "$SCRIPT_DIR/app.log" ]; then
        echo "=== app.log 内容 ==="
        cat "$SCRIPT_DIR/app.log"
        echo
    fi
    
    # 查找并显示最新的带时间戳的日志文件内容
    local latest_log=$(ls -t "$SCRIPT_DIR"/app-*.log 2>/dev/null | head -n 1)
    if [ ! -z "$latest_log" ]; then
        echo "=== 最新的时间戳日志文件 $(basename "$latest_log") 内容 ==="
        cat "$latest_log"
    fi
    
    echo "----------------------------------------"
}

# 主函数
main() {
    echo "开始构建..."
    echo "源文件：$SOURCE_FILE"
    echo "输出目录：$OUTPUT_DIR"
    echo

    # 构建当前平台的可执行文件
    echo "正在构建..."
    go build -o "$OUTPUT_DIR/log" "$SOURCE_FILE"
    echo "构建完成！二进制文件位于：$OUTPUT_DIR/log"
    echo "文件列表："
    ls -l "$OUTPUT_DIR"

    echo
    echo "运行构建的程序："
    "$OUTPUT_DIR/log"
    
    # 显示生成的日志文件内容
    show_logs
}

# 执行主函数
main 