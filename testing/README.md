# testing

## 简介

`testing` 包提供了一组用于测试时输出日志的辅助函数。这个包封装了标准库 `fmt` 包的功能，并在输出内容前添加统一的日志前缀，使测试输出更加清晰和易于识别。

### 主要特性

- 统一的日志输出格式，便于识别测试日志
- 支持 `Println` 和 `Printf` 两种输出方式
- 自动添加换行符，保持输出格式整洁
- 支持任意类型参数的输出
- 与标准库 `fmt` 包完全兼容的格式化功能

### 设计理念

该包的设计遵循以下原则：
- 简单性：提供简单直观的 API
- 一致性：保持与标准库 `fmt` 包相似的使用方式
- 可识别性：通过统一的日志前缀，使测试输出易于识别
- 灵活性：支持多种输出格式和参数类型

## 安装

### 前置条件

- Go 版本要求：Go 1.16 或更高版本
- 依赖要求：无外部依赖

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/testing
```

## 快速开始

### 基础用法

```go
package main

import "github.com/fsyyft-go/kit/testing"

func main() {
    // 使用 Println 输出简单信息
    testing.Println("开始测试")
    
    // 使用 Printf 输出格式化信息
    testing.Printf("测试进度：%d%%\n", 50)
}
```

### 配置选项

该包无需配置，直接使用即可。

## 详细指南

### 核心概念

- `logHeader`：统一的日志前缀，用于标识测试输出
- `Println`：支持多参数输出的日志函数
- `Printf`：支持格式化输出的日志函数

### 常见用例

#### 1. 输出简单测试信息

```go
testing.Println("测试用例 1 开始执行")
testing.Println("输入参数：", 100, "期望结果：", "成功")
```

#### 2. 输出格式化测试信息

```go
testing.Printf("测试进度：%d%%\n", 75)
testing.Printf("用户 %s 的测试结果：%s\n", "张三", "通过")
```

### 最佳实践

- 在测试开始时使用 `Println` 输出测试用例信息
- 使用 `Printf` 输出需要格式化的详细信息
- 保持日志信息的简洁性和可读性
- 在关键测试步骤添加日志输出

## API 文档

### 主要类型

```go
const logHeader = "=-=       "
```

### 关键函数

#### Println

输出带有统一前缀的日志信息，并在末尾自动添加换行符。

```go
func Println(a ...interface{})
```

示例：
```go
testing.Println("测试信息")
testing.Println("值：", 100, "状态：", "成功")
```

#### Printf

输出带有统一前缀的格式化日志信息。

```go
func Printf(format string, a ...interface{})
```

示例：
```go
testing.Printf("当前进度：%d%%\n", 50)
testing.Printf("用户：%s，年龄：%d\n", "张三", 25)
```

### 错误处理

该包不涉及错误处理，所有函数都是安全的，不会返回错误。

## 性能指标

| 操作 | 性能指标 | 说明 |
|------|----------|------|
| Println | O(n) | n 为参数数量 |
| Printf | O(n) | n 为参数数量 |

## 测试覆盖率

| 包 | 覆盖率 |
|------|--------|
| testing | 100% |

## 调试指南

### 日志级别

该包不提供日志级别功能，所有输出都使用统一的格式。

### 常见问题排查

#### 输出格式不符合预期

检查是否正确使用了格式化字符串，确保与 `fmt.Printf` 的用法一致。

## 相关文档

- [Go 标准库 fmt 包文档](https://pkg.go.dev/fmt)
- [Go 测试指南](https://pkg.go.dev/testing)

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 报告问题
- 提交功能建议
- 提交代码改进
- 完善文档

请参考我们的[贡献指南](../CONTRIBUTING.md)了解详细信息。

## 许可证

本项目采用 MIT 许可证。查看 [LICENSE](../LICENSE) 文件了解更多信息。

## 补充说明

本文的大部分信息，由 AI 使用[模板](../ai/templates/docs/package_readme_template.md)根据[提示词](../ai/prompts/docs/package_readme_generator.md)自动生成，如有任何问题，请随时联系我。 