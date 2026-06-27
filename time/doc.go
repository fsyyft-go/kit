// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package time 基于 github.com/dromara/carbon/v2 提供相对当前时间的便捷函数。
//
// 包加载时会把默认日期时间布局、时区、每周起始日和语言环境写入 carbon 的全局默认配置。
// defaultDateTimeLayout、defaultTimezone、defaultWeekStartAt 和 defaultLocale 可以通过 -ldflags -X
// 在程序启动前覆盖；这些设置会影响同一进程中依赖 carbon 全局默认值的代码。
//
// defaultWeekStartAt 支持以下大小写不敏感的枚举字符串，空字符串或无法识别的值会回退为 Monday：
//   - Sunday: 以星期日作为每周起始日，对应 carbon.Sunday。
//   - Monday: 以星期一作为每周起始日，对应 carbon.Monday。
//   - Tuesday: 以星期二作为每周起始日，对应 carbon.Tuesday。
//   - Wednesday: 以星期三作为每周起始日，对应 carbon.Wednesday。
//   - Thursday: 以星期四作为每周起始日，对应 carbon.Thursday。
//   - Friday: 以星期五作为每周起始日，对应 carbon.Friday。
//   - Saturday: 以星期六作为每周起始日，对应 carbon.Saturday。
//
// Now、Yesterday 和 Tomorrow 直接返回 carbon 当前时间或相邻日期；DayAfterTomorrow、
// DayBeforeYesterday、LastWeek、NextWeek、LastMonth、NextMonth、LastYear 和 NextYear
// 基于当前时间副本计算，避免在 Carbon 测试时间被冻结时修改全局 frozen now。函数均返回
// *carbon.Carbon；当 carbon 默认配置无效时，返回值会携带 Carbon 错误，调用方应检查 Error
// 或 IsInvalid。更完整的解析、格式化和日历能力由 carbon API 提供。
package time
