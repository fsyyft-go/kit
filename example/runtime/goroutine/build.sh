#!/bin/bash

# ===== 基础设置 =====
# 设置错误时退出
set -e

# ===== 路径配置 =====
# 脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# 项目根目录（向上三级）
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../" && pwd)"
# 源文件路径
SOURCE_FILE="$SCRIPT_DIR/main.go"
# 输出目录
OUTPUT_DIR="$PROJECT_ROOT/bin/example/runtime/goroutine"
# 输出文件名（默认使用目录名）
OUTPUT_NAME="goid"

# ===== 可选功能配置 =====
# 是否启用版本信息
ENABLE_VERSION_INFO=true
# 是否启用跨平台构建
ENABLE_CROSS_PLATFORM=true
# 是否显示日志文件
SHOW_LOGS=true
# 是否在构建后运行
RUN_AFTER_BUILD=true

# ===== 版本信息配置 =====
VERSION="1.0.0"
BUILD_TIME=$(date "+%Y%m%d%H%M%S000")
GIT_COMMIT="unknown"
LIB_GIT_COMMIT="unknown"

# ===== 跨平台构建配置 =====
# 支持的操作系统和架构
OS_LIST=("windows" "linux" "darwin")
ARCH_LIST=("amd64" "arm64" "386")
# 扩展的 Linux 架构
LINUX_EXTRA_ARCH=("arm" "mips" "mips64" "mips64le" "ppc64" "ppc64le" "s390x" "riscv64")

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

# 检查 Debug 模式
check_debug_mode() {
    if [ "$ENABLE_VERSION_INFO" = true ]; then
        print_info "检查 Debug 模式..."
        echo "=============================================="
        echo "检查直接运行时的 Debug 模式："
        echo "---------------------------------------------"
        go run "$SOURCE_FILE" | grep "Debug Mode" || echo "未找到 Debug Mode 信息"
        echo "---------------------------------------------"
        echo "说明：Debug Mode 为 true 表示当前是开发环境"
        echo "=============================================="
        echo
    fi
}

# 获取版本信息
get_version_info() {
    if [ "$ENABLE_VERSION_INFO" = true ]; then
        print_info "获取版本信息..."
        
        # 检查是否在 git 仓库中
        if ! git rev-parse --git-dir > /dev/null 2>&1; then
            print_warning "当前目录不是 git 仓库，将使用默认值。"
        else
            GIT_COMMIT=$(git rev-parse HEAD || echo "unknown")
            LIB_GIT_COMMIT=$(cd "$PROJECT_ROOT" && git rev-parse HEAD || echo "unknown")
        fi
        
        LIBRARY_DIR="$PROJECT_ROOT"
        WORKING_DIR="$SCRIPT_DIR"
        GOPATH=$(go env GOPATH)
        GOROOT=$(go env GOROOT)
        
        print_info "版本: $VERSION"
        print_info "构建时间: $BUILD_TIME"
        print_info "Git 提交: $GIT_COMMIT"
        print_info "库 Git 提交: $LIB_GIT_COMMIT"
    fi
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
    
    local ldflags=""
    
    # 如果启用版本信息，添加版本信息到 ldflags
    if [ "$ENABLE_VERSION_INFO" = true ]; then
        ldflags="-X github.com/fsyyft-go/kit/go/build.version=${VERSION} \
                 -X github.com/fsyyft-go/kit/go/build.gitVersion=${GIT_COMMIT} \
                 -X github.com/fsyyft-go/kit/go/build.libGitVersion=${LIB_GIT_COMMIT} \
                 -X github.com/fsyyft-go/kit/go/build.buildTimeString=${BUILD_TIME} \
                 -X github.com/fsyyft-go/kit/go/build.buildLibraryDirectory=${LIBRARY_DIR} \
                 -X github.com/fsyyft-go/kit/go/build.buildWorkingDirectory=${WORKING_DIR} \
                 -X github.com/fsyyft-go/kit/go/build.buildGopathDirectory=${GOPATH} \
                 -X github.com/fsyyft-go/kit/go/build.buildGorootDirectory=${GOROOT}"
        
        print_info "构建带版本信息的可执行文件..."
        go build -ldflags "$ldflags" -o "$OUTPUT_DIR/$OUTPUT_NAME" "$SOURCE_FILE"
    else
        print_info "构建可执行文件..."
        go build -o "$OUTPUT_DIR/$OUTPUT_NAME" "$SOURCE_FILE"
    fi
    
    print_success "构建完成！二进制文件位于：$OUTPUT_DIR/$OUTPUT_NAME"
    echo "文件列表："
    ls -l "$OUTPUT_DIR"
}

# 编译单个平台/架构
build_for_platform() {
    local os=$1
    local arch=$2
    local suffix=$3
    local output="$OUTPUT_DIR/${OUTPUT_NAME}_${os}_${arch}${suffix}"
    
    local ldflags=""
    if [ "$ENABLE_VERSION_INFO" = true ]; then
        ldflags="-X github.com/fsyyft-go/kit/go/build.version=${VERSION} \
                 -X github.com/fsyyft-go/kit/go/build.gitVersion=${GIT_COMMIT} \
                 -X github.com/fsyyft-go/kit/go/build.libGitVersion=${LIB_GIT_COMMIT} \
                 -X github.com/fsyyft-go/kit/go/build.buildTimeString=${BUILD_TIME} \
                 -X github.com/fsyyft-go/kit/go/build.buildLibraryDirectory=${LIBRARY_DIR} \
                 -X github.com/fsyyft-go/kit/go/build.buildWorkingDirectory=${WORKING_DIR} \
                 -X github.com/fsyyft-go/kit/go/build.buildGopathDirectory=${GOPATH} \
                 -X github.com/fsyyft-go/kit/go/build.buildGorootDirectory=${GOROOT}"
    fi
    
    print_info "构建 $os/$arch..."
    GOOS=$os GOARCH=$arch go build -ldflags "$ldflags" -o "$output" "$SOURCE_FILE"
    print_success "已构建: $output"
}

# 跨平台构建
build_cross_platform() {
    if [ "$ENABLE_CROSS_PLATFORM" = true ]; then
        print_info "开始跨平台构建..."
        
        # 构建常规平台
        for os in "${OS_LIST[@]}"; do
            for arch in "${ARCH_LIST[@]}"; do
                # 跳过不支持的组合
                if [[ "$os" == "darwin" && "$arch" != "amd64" && "$arch" != "arm64" ]]; then
                    continue
                fi
                
                local suffix=""
                if [ "$os" = "windows" ]; then
                    suffix=".exe"
                fi
                
                build_for_platform "$os" "$arch" "$suffix"
            done
        done
        
        # 构建额外的 Linux 架构
        for arch in "${LINUX_EXTRA_ARCH[@]}"; do
            build_for_platform "linux" "$arch" ""
        done
        
        print_success "跨平台构建完成！二进制文件位于：$OUTPUT_DIR"
        echo "文件列表："
        ls -l "$OUTPUT_DIR"
    else
        # 如果不启用跨平台构建，则执行单平台构建
        build_single
    fi
}

# 运行构建后的程序
run_built_program() {
    if [ "$RUN_AFTER_BUILD" = true ]; then
        print_info "运行构建的程序..."
        
        if [ "$ENABLE_CROSS_PLATFORM" = true ]; then
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
            local current_binary="$OUTPUT_DIR/${OUTPUT_NAME}_${current_os}_${current_arch}${suffix}"
            
            print_info "运行当前平台 ($current_os/$current_arch) 的二进制文件:"
            "$current_binary"
        else
            print_info "运行构建的程序:"
            "$OUTPUT_DIR/$OUTPUT_NAME"
        fi
    fi
}

# ===== 主函数 =====
main() {
    # 创建输出目录
    create_output_dir
    
    # 检查 Debug 模式
    check_debug_mode
    
    # 获取版本信息
    get_version_info
    
    # 构建程序
    if [ "$ENABLE_CROSS_PLATFORM" = true ]; then
        build_cross_platform
    else
        build_single
    fi
    
    # 运行构建后的程序
    run_built_program
    
    # 显示日志文件
    show_logs
    
    print_success "所有操作完成！"
}

# ===== 执行主函数 =====
main 