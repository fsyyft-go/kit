 # 版本信息管理示例

本示例展示了如何使用 Kit 的版本信息管理功能，实现在编译时注入和运行时获取应用程序的版本信息。

## 功能特性

- 支持在编译时注入版本号、Git 提交信息
- 支持获取构建时间信息
- 支持获取构建环境信息（工作目录、GOPATH、GOROOT 等）
- 提供统一的版本信息访问接口
- 支持调试模式标识

## 设计原理

Kit 的版本信息管理模块采用了以下设计：

- 通过 Go 的链接标志（ldflags）在编译时注入版本信息
- 使用 `BuildingContext` 接口定义统一的版本信息访问方法
- 通过 `config.CurrentVersion` 提供全局版本信息访问点
- 支持完整版本和短版本格式

这种设计使得应用程序可以在运行时轻松获取编译时的各种信息，便于问题诊断和版本追踪。

## 使用方法

### 1. 编译和运行

在 Unix/Linux/macOS 系统上：

```bash
# 添加执行权限
chmod +x build.sh

# 构建和运行
./build.sh
```

### 2. 代码示例

```go
package main

import (
	"fmt"

	"github.com/fsyyft-go/kit/config"
)

func main() {
	// 获取基本版本信息
	fmt.Printf("Version: %s\n", config.CurrentVersion.Version())
	fmt.Printf("Git Version: %s\n", config.CurrentVersion.GitVersion())
	
	// 获取构建时间
	fmt.Printf("Build Time: %s\n", config.CurrentVersion.BuildTimeString())
	
	// 获取构建环境信息
	fmt.Printf("Library Directory: %s\n", config.CurrentVersion.BuildLibraryDirectory())
	fmt.Printf("Working Directory: %s\n", config.CurrentVersion.BuildWorkingDirectory())
	fmt.Printf("GOPATH Directory: %s\n", config.CurrentVersion.BuildGopathDirectory())
	fmt.Printf("GOROOT Directory: %s\n", config.CurrentVersion.BuildGorootDirectory())
	
	// 获取调试模式状态
	fmt.Printf("Debug Mode: %v\n", config.CurrentVersion.Debug())
}
```

### 3. 输出示例

```
Version: 1.0.0
Git Version: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
Build Time: 20240315100000000
Library Directory: /path/to/fsyyft-go/kit
Working Directory: /path/to/your/project
GOPATH Directory: /path/to/gopath
GOROOT Directory: /path/to/goroot
Debug Mode: false
```

### 4. 在其他项目中使用

在你的项目中，可以通过以下方式使用版本信息功能：

```go
package main

import (
	"fmt"

	"github.com/fsyyft-go/kit/config"
)

func main() {
	// 打印简短版本信息
	fmt.Println("应用版本:", config.CurrentVersion)
	
	// 打印详细版本信息
	fmt.Printf("%+v\n", config.CurrentVersion)
	
	// 在日志中使用版本信息
	log.Printf("启动应用 %s", config.CurrentVersion.Version())
	
	// 在 API 响应中包含版本信息
	response := map[string]interface{}{
		"status":  "ok",
		"version": config.CurrentVersion.Version(),
		"build":   config.CurrentVersion.BuildTimeString(),
	}
	
	// ... 其他代码
}
```

在编译时注入版本信息：

```bash
go build -ldflags "
    -X github.com/fsyyft-go/kit/go/build.version=1.2.3
    -X github.com/fsyyft-go/kit/go/build.gitVersion=$(git rev-parse HEAD)
    -X github.com/fsyyft-go/kit/go/build.buildTimeString=$(date '+%Y%m%d%H%M%S000')
" -o myapp main.go
```

## 注意事项

- 版本信息只有在编译时通过 ldflags 注入才能正确显示
- 如果没有注入版本信息，相关字段将显示为空字符串或默认值
- 短版本格式（GitShortVersion）只显示 Git 哈希的前 8 位字符
- 构建时间使用 "20060102150405000" 格式（年月日时分秒毫秒）
- 在 CI/CD 环境中使用时，确保正确设置了所有必要的环境变量

## 相关文档

- [Go 编译链接标志文档](https://golang.org/cmd/link/)
- [Kit 配置模块文档](../../config/README.md)
- [Kit 构建工具文档](../../go/build/README.md)

## 许可证

本示例代码采用 MIT 许可证。详见 [LICENSE](../../../LICENSE) 文件。