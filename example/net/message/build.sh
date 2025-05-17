#!/bin/bash

# ===== 基础设置 =====
# 设置错误时退出
set -e

# ===== 路径配置 =====
# 脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# 项目根目录（向上两级）
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
# 源文件路径
SOURCE_FILE="$SCRIPT_DIR/main.go"
# 输出目录
OUTPUT_DIR="$PROJECT_ROOT/bin/example/net/message"
# 输出文件名（默认使用目录名）
OUTPUT_NAME="message"

# ===== 可选功能配置 =====
# 是否显示日志文件
SHOW_LOGS=true
# 是否在构建后运行
RUN_AFTER_BUILD=true

# ===== 颜色配置 =====
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ===== 辅助函数 =====

# 打印带颜色的信息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 创建输出目录
create_output_dir() {
    print_info "创建输出目录：$OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR"
}

# 显示日志文件内容
show_logs() {
    if [ "$SHOW_LOGS" = true ]; then
        print_info "显示日志文件内容..."
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
    fi
}

# 单平台构建
build_single() {
    print_info "开始构建..."
    go build -o "$OUTPUT_DIR/$OUTPUT_NAME" "$SOURCE_FILE"
    print_success "构建完成！二进制文件位于：$OUTPUT_DIR/$OUTPUT_NAME"
    echo "文件列表："
    ls -l "$OUTPUT_DIR"
}

# ===== 主流程 =====
main() {
    create_output_dir
    show_logs
    build_single
    if [ "$RUN_AFTER_BUILD" = true ]; then
        print_info "运行构建后的程序..."
        "$OUTPUT_DIR/$OUTPUT_NAME"
    fi
}

main "$@" 