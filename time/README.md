# time

## 简介

time 包是一个基于 carbon 库的时间处理工具包，提供了丰富的时间操作功能和灵活的配置选项。该包旨在简化 Go 语言中的时间处理操作，提供了一系列直观且易用的 API。

### 主要特性

- 支持自定义日期时间格式
- 灵活的时区配置
- 可配置的周起始日（周一或周日）
- 多语言环境支持（中文、英文、日文等）
- 丰富的时间计算功能（昨天、明天、上周、下月等）
- 编译时可配置的默认参数

### 设计理念

time 包的设计遵循以下原则：

1. **简单易用**：提供直观的 API，减少开发者的学习成本
2. **灵活配置**：支持通过编译参数自定义包的行为
3. **国际化支持**：内置多语言支持，满足不同地区的需求
4. **可扩展性**：基于成熟的 carbon 库，保证功能的可靠性和可扩展性

## 安装

### 前置条件

- Go 版本要求：>= 1.18
- 依赖要求：
  - github.com/dromara/carbon/v2

### 安装命令

```bash
go get -u github.com/fsyyft/fsyyft-go/time
```

## 快速开始

### 基础用法

```go
package main

import (
    "fmt"
    "github.com/fsyyft/fsyyft-go/time"
)

func main() {
    // 获取当前时间
    now := time.Now()
    fmt.Println(now.ToDateTimeString())

    // 获取昨天的时间
    yesterday := time.Yesterday()
    fmt.Println(yesterday.ToDateTimeString())

    // 获取明天的时间
    tomorrow := time.Tomorrow()
    fmt.Println(tomorrow.ToDateTimeString())
}
```

### 配置选项

可以通过编译时参数配置包的默认行为：

```bash
go build -ldflags "
    -X 'github.com/fsyyft/fsyyft-go/time.defaultDateTimeLayout=2006-01-02 15:04:05'
    -X 'github.com/fsyyft/fsyyft-go/time.defaultTimezone=UTC'
    -X 'github.com/fsyyft/fsyyft-go/time.defaultWeekStartAt=Sunday'
    -X 'github.com/fsyyft/fsyyft-go/time.defaultLocale=en'
"
```

## 详细指南

### 核心概念

1. **时间格式化**：使用 Go 的标准时间格式字符串
2. **时区处理**：支持多种时区配置
3. **语言环境**：支持多语言显示
4. **时间计算**：提供丰富的时间计算函数

### 常见用例

#### 1. 基本时间获取

```go
// 获取当前时间
now := time.Now()

// 获取昨天和明天
yesterday := time.Yesterday()
tomorrow := time.Tomorrow()

// 获取前天和后天
dayBefore := time.DayBeforeYesterday()
dayAfter := time.DayAfterTomorrow()
```

#### 2. 周期性时间计算

```go
// 获取上周和下周
lastWeek := time.LastWeek()
nextWeek := time.NextWeek()

// 获取上月和下月
lastMonth := time.LastMonth()
nextMonth := time.NextMonth()

// 获取去年和明年
lastYear := time.LastYear()
nextYear := time.NextYear()
```

### 最佳实践

- 使用编译时配置来设置全局默认值
- 在应用初始化时确认时区设置
- 使用适合目标用户的语言环境
- 注意处理跨时区的时间计算

## API 文档

### 主要类型

time 包主要使用 `carbon.Carbon` 类型作为时间表示：

```go
type Carbon struct {
    // 内部字段由 carbon 库管理
}
```

### 关键函数

#### Now()

返回当前时间的 Carbon 实例。

```go
func Now() carbon.Carbon
```

示例：
```go
now := time.Now()
fmt.Println(now.ToDateTimeString())
```

#### Yesterday()

返回昨天同一时间的 Carbon 实例。

```go
func Yesterday() carbon.Carbon
```

示例：
```go
yesterday := time.Yesterday()
fmt.Println(yesterday.ToDateTimeString())
```

### 错误处理

time 包的函数返回 `carbon.Carbon` 实例，不会返回错误。如果需要进行错误处理，请参考 carbon 库的文档。

## 性能指标

| 操作 | 性能指标 | 说明 |
|------|----------|------|
| 基本时间操作 | O(1) | 常量时间复杂度 |
| 时间计算 | O(1) | 常量时间复杂度 |
| 格式化输出 | O(n) | n 为输出字符串长度 |

## 测试覆盖率

| 包 | 覆盖率 |
|------|--------|
| time | 待补充 |

## 调试指南

### 常见问题排查

#### 时区不正确

问题：时间显示的时区与预期不符
解决方案：检查 `defaultTimezone` 配置，确保使用正确的时区标识符

#### 格式化输出异常

问题：时间格式化输出不符合预期
解决方案：检查 `defaultDateTimeLayout` 配置，确保使用正确的格式字符串

## 相关文档

- [Carbon 库文档](https://github.com/dromara/carbon)
- [Go 时间格式化文档](https://golang.org/pkg/time/#pkg-constants)

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 报告问题
- 提交功能建议
- 提交代码改进
- 完善文档

请参考我们的[贡献指南](../CONTRIBUTING.md)了解详细信息。

## 许可证

本项目采用 MIT 许可证。查看 [LICENSE](../LICENSE) 文件了解更多信息。