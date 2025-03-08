#!/bin/bash

# 设置错误时退出
set -e

# 脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# 项目根目录（向上两级）
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../" && pwd)"
# 源文件路径
SOURCE_FILE="$SCRIPT_DIR/main.go"
# 输出目录
OUTPUT_DIR="$PROJECT_ROOT/bin/example/config/version"

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 检查 Debug 模式
check_debug_mode() {
    echo "=============================================="
    echo "检查直接运行时的 Debug 模式："
    echo "---------------------------------------------"
    go run "$SOURCE_FILE" | grep "Debug Mode"
    echo "---------------------------------------------"
    echo "说明：Debug Mode 为 true 表示当前是开发环境"
    echo "=============================================="
    echo
}

# 获取版本信息
get_version_info() {
    VERSION="1.0.0"
    BUILD_TIME=$(date "+%Y%m%d%H%M%S000")
    
    # 检查是否在 git 仓库中
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        echo "警告：当前目录不是 git 仓库，将使用默认值。"
        GIT_COMMIT="unknown"
        LIB_GIT_COMMIT="unknown"
    else
        GIT_COMMIT=$(git rev-parse HEAD || echo "unknown")
        LIB_GIT_COMMIT=$(cd "$PROJECT_ROOT" && git rev-parse HEAD || echo "unknown")
    fi
    
    LIBRARY_DIR="$PROJECT_ROOT"
    WORKING_DIR="$SCRIPT_DIR"
    GOPATH=$(go env GOPATH)
    GOROOT=$(go env GOROOT)
}

# 主函数
main() {
    # 首先检查 Debug 模式
    check_debug_mode

    echo "开始构建..."
    echo "源文件：$SOURCE_FILE"
    echo "输出目录：$OUTPUT_DIR"
    echo

    # 获取版本信息
    get_version_info

    # 构建当前平台的可执行文件
    echo "正在构建..."
    go build -ldflags "-X github.com/fsyyft-go/kit/go/build.version=${VERSION} \
                      -X github.com/fsyyft-go/kit/go/build.gitVersion=${GIT_COMMIT} \
                      -X github.com/fsyyft-go/kit/go/build.libGitVersion=${LIB_GIT_COMMIT} \
                      -X github.com/fsyyft-go/kit/go/build.buildTimeString=${BUILD_TIME} \
                      -X github.com/fsyyft-go/kit/go/build.buildLibraryDirectory=${LIBRARY_DIR} \
                      -X github.com/fsyyft-go/kit/go/build.buildWorkingDirectory=${WORKING_DIR} \
                      -X github.com/fsyyft-go/kit/go/build.buildGopathDirectory=${GOPATH} \
                      -X github.com/fsyyft-go/kit/go/build.buildGorootDirectory=${GOROOT}" \
             -o "$OUTPUT_DIR/version" "$SOURCE_FILE"
    echo "构建完成！二进制文件位于：$OUTPUT_DIR/version"
    echo "文件列表："
    ls -l "$OUTPUT_DIR"

    echo
    echo "运行构建的程序："
    "$OUTPUT_DIR/version"
}

# 执行主函数
main 