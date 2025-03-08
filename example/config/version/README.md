# 版本信息管理示例

本示例展示了如何使用 Kit 的版本信息管理功能，实现在编译时注入和运行时获取应用程序的版本信息。

## 功能特性

- 支持在编译时注入版本号、构建时间、Git 提交信息等
- 提供统一的版本信息访问接口
- 支持获取完整的构建环境信息（GOPATH、GOROOT 等）
- 支持调试模式标识

## 设计原理

该功能基于以下三个主要组件：

1. `go/build/context.go`：定义构建上下文接口和基础实现
   - 提供版本信息存储的变量
   - 实现 `BuildingContext` 接口
   - 管理构建时的环境信息

2. `config/version.go`：提供版本信息的访问接口
   - 封装 `CurrentVersion` 全局实例
   - 实现版本信息的格式化输出
   - 提供友好的信息访问方法

3. `example/config/version/main.go`：示例程序
   - 展示如何获取和显示版本信息
   - 演示所有可用的版本信息字段

## 使用方法

### 1. 编译和运行

在 Unix/Linux/macOS 系统上：

```bash
# 添加执行权限
chmod +x build.sh

# 构建和运行
./build.sh
```

在 Windows 系统上：

```cmd
# 构建和运行
build.bat
```

### 2. 输出示例

```
Version: 1.0.0
Git Version: e1a74659dcccc542a787de4d5a8701b58768f42c
Build Time: 20250314210837000
Library Directory: /development/github.com/fsyyft-go/kit
Working Directory: /development/github.com/fsyyft-go/kit/example/config/version
GOPATH Directory: /data/var/lib/go/path
GOROOT Directory: /opt/homebrew/Cellar/go/1.24.1/libexec
Debug Mode: false
```

### 3. 在其他项目中使用

1. 在你的项目中引入包：

```go
import "github.com/fsyyft-go/kit/config"
```

2. 使用 `CurrentVersion` 获取版本信息：

```go
fmt.Printf("Version: %s\n", config.CurrentVersion.Version())
fmt.Printf("Build Time: %s\n", config.CurrentVersion.BuildTimeString())
```

3. 在构建时注入版本信息：

```bash
go build -ldflags "-X github.com/fsyyft-go/kit/go/build.version=1.0.0 \
                  -X github.com/fsyyft-go/kit/go/build.buildTimeString=20250314210837000 \
                  -X github.com/fsyyft-go/kit/go/build.gitVersion=$(git rev-parse HEAD)"
```

## 注意事项

1. 所有版本信息必须在编译时通过 `-X` 链接器标志注入
2. 时间戳格式必须符合 `YYYYMMDDHHmmss000` 格式
3. 目录路径应使用绝对路径以确保准确性
4. Git 提交哈希建议使用完整的 SHA-1 值

## 相关文档

- [Kit 版本管理文档](../../config/README.md)
- [Go 链接器文档](https://golang.org/cmd/link/)

## 许可证

本示例代码采用 MIT 许可证。详见 [LICENSE](../../../LICENSE) 文件。 