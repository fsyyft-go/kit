// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package time 基于 carbon 提供一组“相对当前时间”的便捷函数，并在包初始化时设置默认布局、时区、周起始日和语言环境。
//
// 当前导出 API 主要包含 Now、Yesterday、Tomorrow、DayAfterTomorrow、DayBeforeYesterday、
// LastWeek、NextWeek、LastMonth、NextMonth、LastYear 和 NextYear 等 helpers。
// defaultDateTimeLayout、defaultTimezone、defaultWeekStartAt 和 defaultLocale 可通过
// -ldflags -X 在编译时覆盖；包加载时会将这些值写入 carbon.SetDefault。
// 这些函数返回 *carbon.Carbon；更完整的解析、格式化和日历能力请直接使用 carbon API。
package time
