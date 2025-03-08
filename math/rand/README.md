# Rand 包

`rand` 包提供了一系列随机数生成的工具函数，包括数值范围随机和中文字符随机生成。

## 特性

- 支持范围内的随机数生成（int 和 int64）
- 支持随机汉字生成（Unicode 范围：\[19968, 40869\]）
- 支持随机中文姓氏生成（包含单姓和复姓）
- 支持自定义随机数生成器
- 线程安全
- 使用标准库 math/rand 实现
- 轻量级设计，无外部依赖
- 完整的单元测试覆盖

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/fsyyft-go/kit/math/rand"
)

func main() {
    // 生成范围内的随机数
    num := rand.Intn(nil, 0, 100)     // 生成 [0, 100) 范围内的随机整数
    num64 := rand.Int63n(nil, 0, 100) // 生成 [0, 100) 范围内的 int64 随机数
    
    // 生成随机汉字
    char := rand.Chinese(nil)          // 生成一个随机汉字
    
    // 生成随机姓氏
    lastName := rand.ChineseLastName(nil) // 生成一个随机中文姓氏
    
    fmt.Printf("随机数: %d\n", num)
    fmt.Printf("随机 int64: %d\n", num64)
    fmt.Printf("随机汉字: %s\n", char)
    fmt.Printf("随机姓氏: %s\n", lastName)
}
```

### 使用自定义随机数生成器

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
    
    kitrand "github.com/fsyyft-go/kit/math/rand"
)

func main() {
    // 创建自定义随机数生成器
    source := rand.NewSource(time.Now().UnixNano())
    random := rand.New(source)
    
    // 使用自定义随机数生成器
    num := kitrand.Intn(random, 0, 100)
    char := kitrand.Chinese(random)
    lastName := kitrand.ChineseLastName(random)
    
    fmt.Printf("使用自定义生成器的随机数: %d\n", num)
    fmt.Printf("使用自定义生成器的随机汉字: %s\n", char)
    fmt.Printf("使用自定义生成器的随机姓氏: %s\n", lastName)
}
```

## API 说明

### Int63n

```go
func Int63n(random *rand.Rand, min, max int64) int64
```

生成一个范围在 [min, max) 之间的 int64 类型随机数。

参数：
- `random`：随机数生成器，如果为 nil 则使用默认的随机数生成器
- `min`：随机数范围的最小值（包含）
- `max`：随机数范围的最大值（不包含）

返回值：
- 返回一个范围在 [min, max) 之间的 int64 类型随机数

### Intn

```go
func Intn(random *rand.Rand, min, max int) int
```

生成一个范围在 [min, max) 之间的 int 类型随机数。

参数：
- `random`：随机数生成器，如果为 nil 则使用默认的随机数生成器
- `min`：随机数范围的最小值（包含）
- `max`：随机数范围的最大值（不包含）

返回值：
- 返回一个范围在 [min, max) 之间的 int 类型随机数

### Chinese

```go
func Chinese(random *rand.Rand) string
```

生成一个随机的汉字字符串。

参数：
- `random`：随机数生成器，如果为 nil 则使用默认的随机数生成器

返回值：
- 返回一个随机生成的汉字字符串

### ChineseLastName

```go
func ChineseLastName(random *rand.Rand) string
```

生成一个随机的中文姓氏。

参数：
- `random`：随机数生成器，如果为 nil 则使用默认的随机数生成器

返回值：
- 返回一个随机选择的中文姓氏字符串

## 最佳实践

1. 随机数生成器的选择
   - 对于一次性使用，使用 nil 作为随机数生成器参数
   - 对于需要重复使用的场景，创建自定义随机数生成器并重用
   - 需要可重现的随机序列时，使用固定种子的随机数生成器

2. 范围随机数生成
   - 确保 max 大于 min
   - 注意范围是左闭右开区间 [min, max)
   - 对于大范围随机数，优先使用 Int63n

3. 汉字生成
   - Chinese 函数生成的是常用汉字范围内的字符（Unicode 范围：\[19968, 40869\]）
   - ChineseLastName 函数生成的是传统中文姓氏，包含单姓和复姓
   - 生成的汉字都是有效的 Unicode 字符

4. 性能考虑
   - 频繁调用时重用随机数生成器
   - 避免在性能关键路径中频繁创建新的随机数生成器
   - 根据实际需求选择合适的随机数范围

## 测试

包中提供了完整的单元测试，包括：

1. 随机数生成测试
   - 正常范围测试
   - 边界值测试
   - nil random 测试
   - 多次重复测试以验证分布

2. 汉字生成测试
   - 验证生成字符的有效性
   - 验证 Unicode 范围
   - 验证单字符输出

运行测试：
```bash
# 运行所有测试
go test ./math/rand

# 查看测试覆盖率
go test ./math/rand -cover

# 生成覆盖率报告
go test ./math/rand -coverprofile=coverage.out
```

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个包。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../../LICENSE) 文件。 