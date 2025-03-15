// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package time

import (
	"github.com/dromara/carbon/v2"
)

var (
	// defaultDateTimeLayout 指定默认的日期时间格式。
	// 格式遵循 Go 标准时间格式：2006-01-02T15:04:05.000Z
	// 可在编译时通过 -X 参数修改，例如：
	// go build -ldflags "-X 'github.com/fsyyft/fsyyft-go/time.defaultDateTimeLayout=2006-01-02 15:04:05'"
	defaultDateTimeLayout = "2006-01-02T15:04:05.000Z"

	// defaultTimezone 指定默认时区。
	// 使用中国标准时间 (PRC)。
	// 可在编译时通过 -X 参数修改，例如：
	// go build -ldflags "-X 'github.com/fsyyft/fsyyft-go/time.defaultTimezone=UTC'"
	// 常用时区值：
	// - PRC: 中国标准时间
	// - UTC: 协调世界时
	// - Asia/Shanghai: 上海时间
	// - America/New_York: 纽约时间
	defaultTimezone = "PRC"

	// defaultWeekStartAt 指定每周的起始日。
	// 设置为周一 (Monday)。
	// 可在编译时通过 -X 参数修改，例如：
	// go build -ldflags "-X 'github.com/fsyyft/fsyyft-go/time.defaultWeekStartAt=Sunday'"
	// 可选值：
	// - Monday: 周一
	// - Sunday: 周日
	defaultWeekStartAt = "Monday"

	// defaultLocale 指定默认的语言环境。
	// 使用简体中文 (zh-CN)。
	// 可在编译时通过 -X 参数修改，例如：
	// go build -ldflags "-X 'github.com/fsyyft/fsyyft-go/time.defaultLocale=en'"
	// 支持的语言：
	// - zh-CN: 简体中文
	// - en: 英语
	// - ja: 日语
	// 更多语言支持请参考 carbon 库文档
	defaultLocale = "zh-CN"
)

// init 初始化包的默认配置。
// 该函数会在包首次加载时自动执行，设置 carbon 库的默认参数，包括：
// - 日期时间格式
// - 时区
// - 每周起始日
// - 语言环境
func init() {
	carbon.SetDefault(carbon.Default{
		Layout:       defaultDateTimeLayout,
		Timezone:     defaultTimezone,
		WeekStartsAt: defaultWeekStartAt,
		Locale:       defaultLocale,
	})
}

// Now 返回当前时间的 Carbon 实例。
//
// 返回值:
//   - carbon.Carbon: 表示当前时间的 Carbon 实例
//
// 示例:
//
//	now := time.Now()
//	fmt.Println(now.ToDateTimeString()) // 输出类似：2025-01-02 15:04:05
func Now() carbon.Carbon {
	return carbon.Now()
}

// Yesterday 返回昨天同一时间的 Carbon 实例。
//
// 返回值:
//   - carbon.Carbon: 表示昨天同一时间的 Carbon 实例
//
// 示例:
//
//	yesterday := time.Yesterday()
//	fmt.Println(yesterday.ToDateTimeString()) // 输出昨天的日期和当前时间
func Yesterday() carbon.Carbon {
	return carbon.Yesterday()
}

// Tomorrow 返回明天同一时间的 Carbon 实例。
//
// 返回值:
//   - carbon.Carbon: 表示明天同一时间的 Carbon 实例
//
// 示例:
//
//	tomorrow := time.Tomorrow()
//	fmt.Println(tomorrow.ToDateTimeString()) // 输出明天的日期和当前时间
func Tomorrow() carbon.Carbon {
	return carbon.Tomorrow()
}

// DayAfterTomorrow 返回后天同一时间的 Carbon 实例。
//
// 返回值:
//   - carbon.Carbon: 表示后天同一时间的 Carbon 实例
//
// 示例:
//
//	dayAfterTomorrow := time.DayAfterTomorrow()
//	fmt.Println(dayAfterTomorrow.ToDateTimeString()) // 输出后天的日期和当前时间
func DayAfterTomorrow() carbon.Carbon {
	return carbon.Now().AddDays(2)
}

// DayBeforeYesterday 返回前天同一时间的 Carbon 实例。
//
// 返回值:
//   - carbon.Carbon: 表示前天同一时间的 Carbon 实例
//
// 示例:
//
//	dayBeforeYesterday := time.DayBeforeYesterday()
//	fmt.Println(dayBeforeYesterday.ToDateTimeString()) // 输出前天的日期和当前时间
func DayBeforeYesterday() carbon.Carbon {
	return carbon.Now().SubDays(2)
}

// LastWeek 返回上周同一时间的 Carbon 实例。
//
// 返回值:
//   - carbon.Carbon: 表示上周同一时间的 Carbon 实例
//
// 示例:
//
//	lastWeek := time.LastWeek()
//	fmt.Println(lastWeek.ToDateTimeString()) // 输出上周的日期和当前时间
func LastWeek() carbon.Carbon {
	return carbon.Now().SubWeek()
}

// LastMonth 返回上个月同一时间的 Carbon 实例。
//
// 返回值:
//   - carbon.Carbon: 表示上个月同一时间的 Carbon 实例
//
// 示例:
//
//	lastMonth := time.LastMonth()
//	fmt.Println(lastMonth.ToDateTimeString()) // 输出上个月的日期和当前时间
func LastMonth() carbon.Carbon {
	return carbon.Now().SubMonth()
}

// NextWeek 返回下周同一时间的 Carbon 实例。
//
// 返回值:
//   - carbon.Carbon: 表示下周同一时间的 Carbon 实例
//
// 示例:
//
//	nextWeek := time.NextWeek()
//	fmt.Println(nextWeek.ToDateTimeString()) // 输出下周的日期和当前时间
func NextWeek() carbon.Carbon {
	return carbon.Now().AddWeek()
}

// NextMonth 返回下个月同一时间的 Carbon 实例。
//
// 返回值:
//   - carbon.Carbon: 表示下个月同一时间的 Carbon 实例
//
// 示例:
//
//	nextMonth := time.NextMonth()
//	fmt.Println(nextMonth.ToDateTimeString()) // 输出下个月的日期和当前时间
func NextMonth() carbon.Carbon {
	return carbon.Now().AddMonth()
}

// LastYear 返回去年同一时间的 Carbon 实例。
//
// 返回值:
//   - carbon.Carbon: 表示去年同一时间的 Carbon 实例
//
// 示例:
//
//	lastYear := time.LastYear()
//	fmt.Println(lastYear.ToDateTimeString()) // 输出去年的日期和当前时间
func LastYear() carbon.Carbon {
	return carbon.Now().SubYear()
}

// NextYear 返回明年同一时间的 Carbon 实例。
//
// 返回值:
//   - carbon.Carbon: 表示明年同一时间的 Carbon 实例
//
// 示例:
//
//	nextYear := time.NextYear()
//	fmt.Println(nextYear.ToDateTimeString()) // 输出明年的日期和当前时间
func NextYear() carbon.Carbon {
	return carbon.Now().AddYear()
}
