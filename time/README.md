# Time 包

`time` 包基于 [carbon](https://github.com/dromara/carbon) 库提供了一组简单的时间处理工具。

## 特性

- 基于 carbon 库的时间处理
- 支持编译时配置（时区、格式、语言等）
- 提供常用的相对时间获取功能

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/fsyyft-go/kit/time"
)

func main() {
    // 获取当前时间
    now := time.Now()
    fmt.Println(now.ToDateTimeString())

    // 获取相对时间
    yesterday := time.Yesterday()
    tomorrow := time.Tomorrow()
    lastWeek := time.LastWeek()
    nextMonth := time.NextMonth()
}
```

### 编译时配置

可以通过 `-ldflags` 参数在编译时配置包的默认行为：

```bash
go build -ldflags "
    -X 'github.com/fsyyft/fsyyft-go/time.defaultDateTimeLayout=2006-01-02 15:04:05'
    -X 'github.com/fsyyft/fsyyft-go/time.defaultTimezone=UTC'
    -X 'github.com/fsyyft/fsyyft-go/time.defaultWeekStartAt=Sunday'
    -X 'github.com/fsyyft/fsyyft-go/time.defaultLocale=en'
"
```

## 功能列表

### 相对时间

- 日级别
  - Yesterday() - 昨天
  - Tomorrow() - 明天
  - DayBeforeYesterday() - 前天
  - DayAfterTomorrow() - 后天

- 周级别
  - LastWeek() - 上周
  - NextWeek() - 下周

- 月级别
  - LastMonth() - 上月
  - NextMonth() - 下月

- 年级别
  - LastYear() - 去年
  - NextYear() - 明年

### 配置选项

- defaultDateTimeLayout - 默认日期时间格式
  - 默认值：`2006-01-02T15:04:05.000Z`

- defaultTimezone - 默认时区
  - 默认值：`PRC`（中国标准时间）
  - 可选值：`UTC`、`Asia/Shanghai`、`America/New_York` 等

- defaultWeekStartAt - 每周起始日
  - 默认值：`Monday`（周一）
  - 可选值：`Sunday`（周日）

- defaultLocale - 语言环境
  - 默认值：`zh-CN`（简体中文）
  - 可选值：`en`（英语）、`ja`（日语）等

## 最佳实践

1. 时区处理
   - 明确指定时区，避免依赖系统默认时区
   - 在跨时区应用中使用 UTC 时间

2. 格式化
   - 使用标准的时间格式
   - 注意时区信息的保留

3. 性能优化
   - 重用 Carbon 实例
   - 避免频繁的时区转换

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个包。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../LICENSE) 文件。 