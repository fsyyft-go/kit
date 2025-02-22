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
OUTPUT_DIR="$PROJECT_ROOT/bin/example/cache"

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 主函数
main() {
    echo "开始构建..."
    echo "源文件：$SOURCE_FILE"
    echo "输出目录：$OUTPUT_DIR"
    echo

    # 构建当前平台的可执行文件
    echo "正在构建..."
    go build -o "$OUTPUT_DIR/cache" "$SOURCE_FILE"
    echo "构建完成！二进制文件位于：$OUTPUT_DIR/cache"
    echo "文件列表："
    ls -l "$OUTPUT_DIR"

    echo
    echo "运行构建的程序："
    "$OUTPUT_DIR/cache"
}

# 执行主函数
main 