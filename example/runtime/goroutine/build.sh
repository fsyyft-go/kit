#!/bin/bash

# 设置错误时退出
set -e

# 脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# 项目根目录（向上三级）
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../" && pwd)"
# 源文件路径
SOURCE_FILE="$SCRIPT_DIR/main.go"
# 输出目录
OUTPUT_DIR="$PROJECT_ROOT/bin/example/runtime/goroutine"

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 编译函数
build() {
    local os=$1
    local arch=$2
    local suffix=$3
    local output="$OUTPUT_DIR/goid_${os}_${arch}${suffix}"
    
    echo "Building for $os/$arch..."
    GOOS=$os GOARCH=$arch go build -o "$output" "$SOURCE_FILE"
    echo "Built: $output"
}

# 主函数
main() {
    echo "Starting cross-platform build..."
    echo "Source file: $SOURCE_FILE"
    echo "Output directory: $OUTPUT_DIR"
    echo

    # Windows 平台
    build "windows" "amd64" ".exe"
    build "windows" "arm64" ".exe"
    build "windows" "386" ".exe"

    # Linux 平台
    build "linux" "amd64" ""
    build "linux" "arm64" ""
    build "linux" "arm" ""
    build "linux" "386" ""
    build "linux" "mips" ""
    build "linux" "mips64" ""
    build "linux" "mips64le" ""
    build "linux" "ppc64" ""
    build "linux" "ppc64le" ""
    build "linux" "s390x" ""
    build "linux" "riscv64" ""

    # Darwin 平台
    build "darwin" "amd64" ""
    build "darwin" "arm64" ""

    echo
    echo "Build complete! Binaries are available in: $OUTPUT_DIR"
    echo "Files:"
    ls -l "$OUTPUT_DIR"

    # 获取当前系统信息
    local current_os
    local current_arch
    case "$(uname -s)" in
        Darwin*) current_os="darwin" ;;
        Linux*)  current_os="linux" ;;
        MINGW*|MSYS*|CYGWIN*) current_os="windows" ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64) current_arch="amd64" ;;
        arm64|aarch64) current_arch="arm64" ;;
        i386|i686) current_arch="386" ;;
        armv7*) current_arch="arm" ;;
    esac

    # 构建当前平台的可执行文件名
    local suffix=""
    if [ "$current_os" = "windows" ]; then
        suffix=".exe"
    fi
    local current_binary="$OUTPUT_DIR/goid_${current_os}_${current_arch}${suffix}"

    echo
    echo "Running binary for current platform ($current_os/$current_arch):"
    "$current_binary"
}

# 执行主函数
main 