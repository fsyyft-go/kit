# math/rand

## 简介

math/rand 包提供了一系列随机数生成的工具函数，包括数值范围随机和中文字符随机生成。在标准库的随机数生成基础上，提供了更加便捷的方式生成指定范围内的随机数，同时提供了中文字符随机生成的功能。

### 主要特性

- 指定范围的随机整数生成（支持 int 和 int64 类型）
- 随机汉字生成（Unicode 范围：[19968, 40869]）
- 随机中文姓氏生成（包含传统单姓和复姓）
- 支持自定义随机数生成器
- 完全兼容标准库 math/rand 包
- 线程安全（当使用自定义线程安全的随机数生成器时）
- 轻量级设计，无外部依赖

### 设计理念

本包设计遵循简单易用的原则，通过扩展标准库 math/rand 的功能，提供更加便捷的范围随机数生成方法和中文字符随机生成功能。API 设计保持一致性，同时兼顾灵活性，允许用户传入自定义的随机数生成器或使用默认生成器。

## 安装

### 前置条件

- Go 版本要求：Go 1.16+
- 依赖要求：
  - 标准库 math/rand

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/math/rand
```

## 快速开始

### 基础用法

```go
package main

import (
    "fmt"
    "github.com/fsyyft-go/kit/math/rand"
)

func main() {
    // 生成范围内的随机数
    num := rand.Intn(nil, 1, 100)  // 生成 [1, 100) 范围内的随机整数
    fmt.Printf("随机整数: %d\n", num)

    // 生成范围内的 int64 随机数
    num64 := rand.Int63n(nil, 1, 1000)  // 生成 [1, 1000) 范围内的随机 int64
    fmt.Printf("随机 int64: %d\n", num64)

    // 生成随机汉字
    ch := rand.Chinese(nil)  // 生成一个随机汉字
    fmt.Printf("随机汉字: %s\n", ch)

    // 生成随机中文姓氏
    lastName := rand.ChineseLastName(nil)  // 生成一个随机中文姓氏
    fmt.Printf("随机姓氏: %s\n", lastName)
}
```

## 详细指南

### 核心概念

本包的核心是对标准库 math/rand 的扩展，提供更加便捷的随机数生成方法。关键概念包括：

1. **随机数生成器**：可以传入自定义的 `*rand.Rand` 实例，也可以传入 nil 使用默认的随机数生成器。
2. **范围随机**：生成指定范围内的随机数，范围为 [min, max)，包含最小值，不包含最大值。
3. **中文字符随机**：生成随机的汉字和中文姓氏，用于需要中文测试数据的场景。

### 常见用例

#### 1. 使用自定义随机数生成器生成范围随机数

```go
package main

import (
    "fmt"
    "math/rand"
    "time"

    kitrand "github.com/fsyyft-go/kit/math/rand"
)

func main() {
    // 创建一个线程安全的随机数生成器
    source := rand.NewSource(time.Now().UnixNano())
    random := rand.New(source)

    // 生成 10 个 [1, 100) 范围内的随机数
    fmt.Println("生成 10 个随机数:")
    for i := 0; i < 10; i++ {
        num := kitrand.Intn(random, 1, 100)
        fmt.Printf("%d ", num)
    }
    fmt.Println()
}
```

#### 2. 生成随机中文名字

```go
package main

import (
    "fmt"
    "math/rand"
    "time"

    kitrand "github.com/fsyyft-go/kit/math/rand"
)

func main() {
    // 创建一个随机数生成器
    source := rand.NewSource(time.Now().UnixNano())
    random := rand.New(source)

    // 生成 5 个随机中文名字
    fmt.Println("生成 5 个随机中文名字:")
    for i := 0; i < 5; i++ {
        // 生成姓氏
        lastName := kitrand.ChineseLastName(random)

        // 生成 1-2 个汉字作为名字
        firstNameLength := kitrand.Intn(random, 1, 3)
        firstName := ""
        for j := 0; j < firstNameLength; j++ {
            firstName += kitrand.Chinese(random)
        }

        // 组合完整名字
        fullName := lastName + firstName
        fmt.Printf("姓名 %d: %s\n", i+1, fullName)
    }
}
```

### 最佳实践

- 对于需要重复使用的场景，创建自定义的随机数生成器并重用，避免每次生成都初始化新的生成器。
- 在多线程环境中，确保传入线程安全的随机数生成器，或者为每个 goroutine 创建独立的生成器。
- 对于高性能场景，预先创建随机数生成器并使用 sync.Pool 进行管理。
- 使用带有时间种子的随机数生成器，避免每次运行程序生成相同的随机序列。
- 使用范围随机时，确保 max 大于 min，否则会导致程序崩溃或不可预期的结果。

## API 文档

### 主要类型

```go
// 本包主要使用标准库中的类型，无特定自定义类型

// 标准库类型引用
import "math/rand"

// *rand.Rand - 随机数生成器
// rand.Source - 随机数源
```

### 关键函数

#### Int63n

在指定范围内生成 int64 类型的随机数。

```go
func Int63n(random *rand.Rand, min, max int64) int64
```

示例：

```go
// 使用默认随机数生成器
num := rand.Int63n(nil, 0, 1000)

// 使用自定义随机数生成器
source := rand.NewSource(time.Now().UnixNano())
random := rand.New(source)
num := rand.Int63n(random, 0, 1000)
```

#### Intn

在指定范围内生成 int 类型的随机数。

```go
func Intn(random *rand.Rand, min, max int) int
```

示例：

```go
// 使用默认随机数生成器
num := rand.Intn(nil, 0, 100)

// 使用自定义随机数生成器
source := rand.NewSource(time.Now().UnixNano())
random := rand.New(source)
num := rand.Intn(random, 0, 100)
```

#### Chinese

生成一个随机的汉字字符串。

```go
func Chinese(random *rand.Rand) string
```

示例：

```go
// 使用默认随机数生成器
char := rand.Chinese(nil)

// 使用自定义随机数生成器
source := rand.NewSource(time.Now().UnixNano())
random := rand.New(source)
char := rand.Chinese(random)
```

#### ChineseLastName

生成一个随机的中文姓氏。

```go
func ChineseLastName(random *rand.Rand) string
```

示例：

```go
// 使用默认随机数生成器
lastName := rand.ChineseLastName(nil)

// 使用自定义随机数生成器
source := rand.NewSource(time.Now().UnixNano())
random := rand.New(source)
lastName := rand.ChineseLastName(random)
```

### 错误处理

本包函数不返回错误，但在以下情况可能会触发 panic：

1. 当 `Intn` 或 `Int63n` 函数的 max 参数小于或等于 min 参数时
2. 当底层标准库 `rand.Intn` 或 `rand.Int63n` 因参数无效而 panic 时

为避免潜在问题，使用前应确保：

- max 参数总是大于 min 参数
- 随机数范围不超过相应类型的最大值

## 性能指标

| 操作            | 性能指标   | 说明                                           |
| --------------- | ---------- | ---------------------------------------------- |
| Intn            | 约 12ns/op | 基于标准库 rand.Intn 的性能，传入 nil 时略慢   |
| Int63n          | 约 12ns/op | 基于标准库 rand.Int63n 的性能，传入 nil 时略慢 |
| Chinese         | 约 25ns/op | 包含随机数生成和 Unicode 转换的时间            |
| ChineseLastName | 约 35ns/op | 包含随机数生成、查表和字符转换的时间           |

## 测试覆盖率

| 包        | 覆盖率 |
| --------- | ------ |
| math/rand | 100%   |

## 调试指南

### 日志级别

本包不包含内置日志功能。如需调试，请在应用层实现日志记录。

### 常见问题排查

#### 随机数生成的范围不正确

问题：生成的随机数超出了预期范围。

解决方案：确保 max 参数大于 min 参数，并注意范围是左闭右开区间 [min, max)。

#### 生成的随机数序列重复

问题：每次运行程序生成的随机数序列相同。

解决方案：使用时间种子初始化随机数生成器：

```go
source := rand.NewSource(time.Now().UnixNano())
random := rand.New(source)
```

#### 多线程环境下的并发问题

问题：在多 goroutine 环境中使用同一个随机数生成器时出现并发问题。

解决方案：

- 为每个 goroutine 创建独立的随机数生成器
- 使用互斥锁保护共享的随机数生成器
- 使用 `sync.Pool` 管理随机数生成器池

## 相关文档

- [标准库 math/rand 文档](https://pkg.go.dev/math/rand)
- [Go 语言中的随机数生成](https://go.dev/doc/effective_go#rand)

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 报告问题
- 提交功能建议
- 提交代码改进
- 完善文档

请参考我们的[贡献指南](../../CONTRIBUTING.md)了解详细信息。

## 许可证

本项目采用 MIT License 许可证。查看 [LICENSE](../../LICENSE) 文件了解更多信息。
