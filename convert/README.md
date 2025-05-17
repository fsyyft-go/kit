# convert

## 简介

convert 包提供了通用且强大的数据类型转换工具，支持任意类型与基础类型、切片、Map、结构体之间的安全转换。基于 [gconv](https://github.com/gogf/gf) 实现，兼容多种常见场景，适用于数据解析、配置加载、接口适配等需求。

### 主要特性

- 支持 int、uint、float、bool、string、time.Time、time.Duration 等基础类型的安全转换
- 支持切片、Map、结构体的自动转换
- 提供带错误返回和无错误返回的两套 API，兼顾安全性与便捷性
- 兼容 gconv，支持多种输入格式（字符串、数字、布尔、时间戳等）
- 支持结构体与 Map 互转、切片批量转换
- 完善的单元测试覆盖，健壮性强

### 设计理念

convert 包遵循"统一、健壮、易用、类型安全"的设计理念，封装 gconv 能力，提供更直观的 API。通过分层 API（带 error/无 error），兼顾类型安全和开发效率，适合多种业务场景。

## 安装

### 前置条件

- Go 版本要求：Go 1.18+
- 依赖要求：
  - github.com/gogf/gf/v2

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/convert
```

## 快速开始

### 基础用法

```go
package main

import (
    "fmt"
    "github.com/fsyyft-go/kit/convert"
)

func main() {
    // 基础类型转换
    i, err := convert.ToInt("123")
    f, err := convert.ToFloat64("123.45")
    b, err := convert.ToBool(1)
    s, err := convert.ToString(456)
    t, err := convert.ToTime("2025-01-01 12:00:00")
    d, err := convert.ToDuration("2h30m")
    fmt.Println(i, f, b, s, t, d)

    // 切片与 Map 转换
    ints, err := convert.ToSliceInt([]any{"1", 2, 3})
    m, err := convert.ToMap(struct{A int}{A: 1})
    fmt.Println(ints, m)

    // 结构体转换
    type User struct {
        Name string `json:"name"`
        Age  int    `json:"age"`
    }
    var user User
    err = convert.ToStruct(map[string]any{"name": "Tom", "age": 20}, &user)
    fmt.Println(user)
}
```

### 无错误返回版本

```go
v := convert.Int("123")      // 123
b := convert.Bool("true")    // true
s := convert.String(456)      // "456"
sl := convert.SliceInt([]any{"1", 2, 3}) // []int{1,2,3}
```

## 详细指南

### 核心概念

- **类型安全**：所有 ToXxx 方法均返回 error，适合对输入不确定场景；Xxx 方法忽略错误，适合已知输入。
- **结构体与 Map 互转**：支持 map[string]any <-> struct，支持切片批量转换。
- **切片批量转换**：支持各种类型切片的自动转换。
- **gconv 兼容**：底层基于 gconv，支持丰富的输入格式。

### 常见用例

#### 1. 字符串转 int/float/bool

```go
i, err := convert.ToInt("123")
f, err := convert.ToFloat64("123.45")
b, err := convert.ToBool("true")
```

#### 2. 任意类型转 string

```go
s, err := convert.ToString(123.45) // "123.45"
```

#### 3. 切片批量转换

```go
ints, err := convert.ToSliceInt([]any{"1", 2, 3}) // []int{1,2,3}
```

#### 4. Map 与结构体互转

```go
type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
var user User
err := convert.ToStruct(map[string]any{"name": "Tom", "age": 20}, &user)
// 结构体转 map
m, err := convert.ToMap(user)
```

#### 5. 无错误返回的便捷用法

```go
v := convert.Int("abc") // 0，转换失败返回零值
```

### 最佳实践

- 推荐优先使用 ToXxx 带 error 的方法，保证类型安全
- 对于已知输入可用 Xxx 方法，简化代码
- 结构体转换时字段需导出且 tag 匹配
- 切片/Map 转换建议输入为 []any/map[string]any
- 错误输入建议始终检查 error

## API 文档

### 主要类型

convert 包不定义特殊类型，直接使用 Go 基础类型和标准库类型。

### 关键函数

#### ToInt/ToFloat64/ToBool/ToString/ToTime/ToDuration

```go
func ToInt(v any) (int, error)
func ToFloat64(v any) (float64, error)
func ToBool(v any) (bool, error)
func ToString(v any) (string, error)
func ToTime(v any) (time.Time, error)
func ToDuration(v any) (time.Duration, error)
```

#### ToSliceInt/ToSliceFloat64/ToSliceStr/ToMap/ToStruct

```go
func ToSliceInt(v any) ([]int, error)
func ToSliceFloat64(v any) ([]float64, error)
func ToSliceStr(v any) ([]string, error)
func ToMap(v any) (map[string]any, error)
func ToStruct(v any, out any) error
```

#### 无错误返回版本

```go
func Int(v any) int
func Float64(v any) float64
func Bool(v any) bool
func String(v any) string
func SliceInt(v any) []int
func Map(v any) map[string]any
```

### 错误处理

- ToXxx 方法遇到无法转换时返回 error，Xxx 方法返回类型零值
- 结构体转换字段不匹配时返回 error
- 切片/Map 转换输入类型不符时返回 error

## 性能指标

| 操作 | 性能指标 | 说明 |
|------|----------|------|
| 基础类型转换 | O(1) | 常量时间 |
| 切片/Map/结构体转换 | O(n) | n 为元素数量 |

## 测试覆盖率

| 包 | 覆盖率 |
|------|--------|
| convert | 100% |

所有导出方法均有单元测试，覆盖常见、边界和错误场景。

## 调试指南

### 常见问题排查

#### 转换失败返回零值

- 建议优先使用 ToXxx 方法，检查 error
- 输入类型不符、字段不匹配、格式错误等均会导致转换失败

#### 结构体/Map 转换异常

- 检查字段名、tag 是否匹配，字段需导出
- 切片批量转换时元素类型需一致

## 相关文档

- [gconv 官方文档](https://github.com/gogf/gf/tree/main/util/gconv)
- [golang每日一库之gconv](http://blog.dd95828.com/articles/2025/04/07/1743987204778.html)

## 贡献指南

欢迎任何形式的贡献，包括但不限于：

- 报告问题
- 提交功能建议
- 提交代码改进
- 完善文档

请参考[贡献指南](../CONTRIBUTING.md)了解详细信息。

## 许可证

本项目采用 MIT License 许可证。查看 [LICENSE](../LICENSE) 文件了解更多信息。