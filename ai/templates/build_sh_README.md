# Build.sh 模板使用指南

## 概述

这是一个用于 Go 项目的标准化构建脚本模板。该模板提供了一个模块化、可扩展的框架，用于自动化 Go 项目的构建、测试和运行过程。模板包含了多个可选功能模块，可以根据项目需求进行自定义。

## 特性

- **基础构建功能**：自动编译 Go 源代码并生成可执行文件
- **版本信息管理**：支持从 Git 获取版本号、提交哈希等信息
- **跨平台构建**：支持为多种操作系统和架构构建可执行文件
- **日志文件处理**：支持显示和管理日志文件
- **彩色输出**：使用彩色文本增强输出可读性
- **模块化结构**：清晰的函数组织，便于理解和自定义
- **错误处理**：内置错误检测和处理机制

## 使用方法

### 基本使用

1. 将 `build.sh` 模板复制到你的 Go 项目目录中
2. 根据项目需求修改配置部分
3. 确保脚本具有执行权限：`chmod +x build.sh`
4. 运行脚本：`./build.sh`

### 配置选项

模板顶部包含多个配置选项，可以根据项目需求进行调整：

#### 路径配置

```bash
# 项目根目录（根据实际项目结构调整）
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../" && pwd)"  # 默认向上一级，根据需要调整
# 源文件路径
SOURCE_FILE="$SCRIPT_DIR/main.go"
# 输出目录（根据实际项目结构调整）
OUTPUT_DIR="$PROJECT_ROOT/bin/$(basename "$(dirname "$SCRIPT_DIR")")/$(basename "$SCRIPT_DIR")"
# 输出文件名（默认使用目录名）
OUTPUT_NAME="$(basename "$SCRIPT_DIR")"
```

#### 功能开关

```bash
# 是否启用版本信息
ENABLE_VERSION_INFO=false
# 是否启用跨平台构建
ENABLE_CROSS_PLATFORM=false
# 是否显示日志文件
SHOW_LOGS=false
# 是否在构建后运行
RUN_AFTER_BUILD=true
```

#### 版本信息配置

```bash
VERSION="1.0.0"
BUILD_TIME=$(date "+%Y%m%d%H%M%S000")
```

#### 跨平台构建配置

```bash
# 支持的操作系统和架构
OS_LIST=("linux" "darwin" "windows")
ARCH_LIST=("amd64" "arm64" "386")
```

## 功能模块说明

### 版本信息管理

当 `ENABLE_VERSION_INFO=true` 时，脚本会：

1. 从 Git 获取当前提交哈希
2. 使用 ldflags 将版本信息注入到编译的二进制文件中
3. 显示版本相关信息

这要求你的 Go 代码中有相应的变量来接收这些信息，例如：

```go
package build

var (
    version           string
    gitVersion        string
    libGitVersion     string
    buildTimeString   string
    buildLibraryDirectory string
    buildWorkingDirectory string
    buildGopathDirectory  string
    buildGorootDirectory  string
)

func GetVersion() string {
    return version
}

// 其他 getter 函数...
```

### 跨平台构建

当 `ENABLE_CROSS_PLATFORM=true` 时，脚本会：

1. 为配置的所有操作系统和架构组合构建可执行文件
2. 自动跳过不支持的组合（如 darwin/386）
3. 为 Windows 平台添加 .exe 后缀
4. 在运行时自动检测当前平台并运行相应的二进制文件

### 日志文件处理

当 `SHOW_LOGS=true` 时，脚本会：

1. 显示当前目录下的 app.log 文件内容（如果存在）
2. 显示最新的带时间戳的日志文件内容（如 app-20250314.log）

## 自定义和扩展

### 添加新的功能模块

要添加新的功能模块，请遵循以下步骤：

1. 在配置部分添加新的开关变量
   ```bash
   # 是否启用新功能
   ENABLE_NEW_FEATURE=false
   ```

2. 创建实现该功能的函数
   ```bash
   # 新功能实现
   new_feature() {
       if [ "$ENABLE_NEW_FEATURE" = true ]; then
           print_info "执行新功能..."
           # 实现代码...
       fi
   }
   ```

3. 在 main 函数中的适当位置调用该函数
   ```bash
   main() {
       # ...
       # 调用新功能
       new_feature
       # ...
   }
   ```

### 修改构建参数

要修改 Go 构建参数，可以编辑 `build_single` 和 `build_cross_platform` 函数中的 `go build` 命令。例如，添加 `-race` 标志来启用竞态检测：

```bash
go build -race -o "$OUTPUT_DIR/$OUTPUT_NAME" "$SOURCE_FILE"
```

## 故障排除

### 脚本无法执行

确保脚本具有执行权限：

```bash
chmod +x build.sh
```

### 构建失败

1. 检查 Go 环境是否正确设置
2. 确保所有依赖项都已安装
3. 检查源文件路径是否正确

### 版本信息不显示

1. 确保 `ENABLE_VERSION_INFO=true`
2. 确保项目在 Git 仓库中
3. 确保 Go 代码中有相应的变量来接收版本信息

### 跨平台构建问题

1. 确保已安装所需的交叉编译工具链
2. 检查 OS_LIST 和 ARCH_LIST 是否包含支持的组合

## 示例

### 基本构建

```bash
./build.sh
```

### 启用版本信息

编辑 build.sh，设置 `ENABLE_VERSION_INFO=true`，然后运行：

```bash
./build.sh
```

### 跨平台构建

编辑 build.sh，设置 `ENABLE_CROSS_PLATFORM=true`，然后运行：

```bash
./build.sh
```

### 自定义输出目录和文件名

编辑 build.sh：

```bash
OUTPUT_DIR="$PROJECT_ROOT/dist"
OUTPUT_NAME="myapp"
```

然后运行：

```bash
./build.sh
```

## 最佳实践

1. **保持模板更新**：定期检查和更新模板，以包含新的最佳实践
2. **版本控制**：将自定义的 build.sh 纳入版本控制系统
3. **文档化**：记录对模板的任何重大更改
4. **参数化**：考虑使用命令行参数而不是硬编码配置
5. **CI/CD 集成**：将脚本集成到 CI/CD 流程中 