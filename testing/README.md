# Testing 包

`testing` 包提供了一组用于测试时输出日志的辅助函数。这个包封装了标准库 `fmt` 包的功能，并在输出内容前添加统一的日志前缀，使测试输出更加清晰和易于识别。

## 特性

- 统一的日志前缀（`=-=       `）
- 支持普通日志输出和格式化输出
- 自动处理参数间的空格分隔
- 与标准库 `fmt` 包兼容的接口

## 快速开始

### 基本使用

```go
package main

import "github.com/fsyyft-go/kit/testing"

func TestExample(t *testing.T) {
    // 普通日志输出
    testing.Println("测试信息")
    testing.Println("值：", 100, "状态：", "成功")
    
    // 格式化日志输出
    testing.Printf("当前进度：%d%%\n", 50)
    testing.Printf("用户：%s，年龄：%d\n", "张三", 25)
}
```

## 函数说明

### Println

`Println` 函数用于输出带有统一前缀的日志信息，并在末尾自动添加换行符。该函数会在实际内容前添加 `logHeader` 前缀，并使用空格分隔多个参数。

```go
func Println(a ...interface{})
```

参数：
- `a ...interface{}`：要输出的任意类型参数列表，支持多个参数

### Printf

`Printf` 函数用于输出带有统一前缀的格式化日志信息。该函数会在实际内容前添加 `logHeader` 前缀，并根据提供的格式字符串格式化输出内容。

```go
func Printf(format string, a ...interface{})
```

参数：
- `format string`：格式化字符串，支持所有 `fmt.Printf` 的格式化指令
- `a ...interface{}`：要格式化输出的参数列表

## 最佳实践

1. 日志内容
   - 使用清晰、描述性的日志信息
   - 适当使用格式化输出提高可读性
   - 避免过多的无用信息

2. 格式选择
   - 简单信息使用 `Println`
   - 需要格式化的信息使用 `Printf`
   - 确保格式化字符串与参数数量匹配

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个包。

## 许可证

本项目采用 Apache License 2.0 许可证。详见 [LICENSE](../LICENSE) 文件。 