 # Goroutine ID 获取示例

本示例展示了如何使用 Kit 的 runtime/goroutine 模块获取 Go 协程（goroutine）的唯一标识符（ID）。

## 功能特性

- 支持获取当前 goroutine 的 ID
- 提供高性能的 ID 获取方法（GetGoID）
- 提供通用兼容的 ID 获取方法（GetGoIDSlow）
- 支持多种 CPU 架构（AMD64、ARM64 等）

## 设计原理

Kit 的 goroutine ID 获取模块采用了两种实现方式：

1. **快速路径（GetGoID）**：
   - 在支持的架构（AMD64、ARM64）上，通过直接访问 TLS（Thread Local Storage）获取 goroutine ID
   - 使用汇编语言实现，性能最优，几乎没有额外开销

2. **慢速路径（GetGoIDSlow）**：
   - 通过解析 goroutine 堆栈信息来获取 ID
   - 适用于所有平台，但性能相对较低
   - 作为快速路径的降级方案

这种设计在保证功能可用性的同时，针对主流平台进行了性能优化。

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
	"sync"

	"github.com/fsyyft-go/kit/runtime/goroutine"
)

func main() {
	// 创建一个等待组，用于等待所有 goroutine 完成
	var wg sync.WaitGroup

	// 获取主 goroutine 的 ID
	fmt.Printf("主 goroutine ID: %d\n", goroutine.GetGoID())

	// 启动多个 goroutine 并获取它们的 ID
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			// 使用快速方法获取 ID
			id := goroutine.GetGoID()
			
			// 使用通用方法获取 ID（作为对比）
			idSlow := goroutine.GetGoIDSlow()
			
			fmt.Printf("goroutine %d 的 ID: %d %d\n", index+1, id, idSlow)
		}(i)
	}

	// 等待所有 goroutine 完成
	wg.Wait()
}
```

### 3. 输出示例

```
主 goroutine ID: 1
goroutine 1 的 ID: 18 18
goroutine 2 的 ID: 19 19
goroutine 3 的 ID: 20 20
```

### 4. 在其他项目中使用

```go
package main

import (
	"fmt"
	"log"
	
	"github.com/fsyyft-go/kit/runtime/goroutine"
)

func main() {
	// 在日志中包含 goroutine ID
	goid := goroutine.GetGoID()
	log.Printf("[goroutine-%d] 应用启动", goid)
	
	// 在多个 goroutine 中使用
	for i := 0; i < 5; i++ {
		go func(n int) {
			goid := goroutine.GetGoID()
			log.Printf("[goroutine-%d] 处理任务 %d", goid, n)
		}(i)
	}
	
	// ... 其他代码
}
```

## 注意事项

- Go 语言官方不推荐依赖 goroutine ID 进行业务逻辑处理
- 该功能主要用于调试、日志追踪等辅助场景
- 在不同的 Go 版本中，获取 goroutine ID 的实现细节可能会发生变化
- 在不支持快速路径的平台上，会自动降级使用慢速路径
- 如果需要频繁获取 goroutine ID，建议缓存结果而不是重复获取

## 性能考虑

- 在支持的架构上（AMD64、ARM64），`GetGoID()` 性能接近于直接内存访问
- 在不支持的架构上，`GetGoIDSlow()` 会有一定性能开销，因为需要解析堆栈信息
- 两种方法的性能差异可能在高并发场景下更为明显

## 支持的平台

### 快速路径支持：
- AMD64 架构（Windows、Linux、macOS）
- ARM64 架构（Linux、macOS，不包括 Windows）

### 通用支持（慢速路径）：
- 所有 Go 支持的平台

## 相关文档

- [Kit Runtime 模块文档](../../runtime/README.md)
- [Kit Goroutine 模块文档](../../runtime/goroutine/README.md)
- [Go Runtime 包文档](https://golang.org/pkg/runtime/)

## 许可证

本示例代码采用 MIT 许可证。详见 [LICENSE](../../../LICENSE) 文件。