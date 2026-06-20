// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package time

import (
	"strings"

	"github.com/dromara/carbon/v2"
)

var (
	// defaultDateTimeLayout 指定默认的日期时间格式。
	// 格式遵循 Go 标准时间格式：2006-01-02T15:04:05.000Z
	// 可在编译时通过 -X 参数修改，例如：
	// go build -ldflags "-X 'github.com/fsyyft-go/kit/time.defaultDateTimeLayout=2006-01-02 15:04:05'"
	defaultDateTimeLayout = "2006-01-02T15:04:05.000Z"

	// defaultTimezone 指定默认时区。
	// 使用中国标准时间 (PRC)。
	// 可在编译时通过 -X 参数修改，例如：
	// go build -ldflags "-X 'github.com/fsyyft-go/kit/time.defaultTimezone=UTC'"
	// 常用时区值：
	// - PRC: 中国标准时间
	// - UTC: 协调世界时
	// - Asia/Shanghai: 上海时间
	// - America/New_York: 纽约时间
	defaultTimezone = "PRC"

	// defaultWeekStartAt 指定每周的起始日。
	// 设置为周一 (Monday)。
	// 可在编译时通过 -X 参数修改，例如：
	// go build -ldflags "-X 'github.com/fsyyft-go/kit/time.defaultWeekStartAt=Sunday'"
	// 可选值（大小写不敏感）：
	// - Sunday: 周日
	// - Monday: 周一
	// - Tuesday: 周二
	// - Wednesday: 周三
	// - Thursday: 周四
	// - Friday: 周五
	// - Saturday: 周六
	defaultWeekStartAt = "Monday"

	// defaultLocale 指定默认的语言环境。
	// 使用简体中文 (zh-CN)。
	// 可在编译时通过 -X 参数修改，例如：
	// go build -ldflags "-X 'github.com/fsyyft-go/kit/time.defaultLocale=en'"
	// 支持的语言：
	// - zh-CN: 简体中文
	// - en: 英语
	// - ja: 日语
	// 更多语言支持请参考 carbon 库文档
	defaultLocale = "zh-CN"
)

// init 在包加载时把当前默认变量写入 carbon 的全局默认配置。
// 该初始化只会执行一次；若通过 -ldflags -X 覆盖默认变量，必须在程序启动前完成。
func init() {
	carbon.SetDefault(carbon.Default{
		Layout:       defaultDateTimeLayout,
		Timezone:     defaultTimezone,
		WeekStartsAt: parseWeekStartAt(defaultWeekStartAt),
		Locale:       defaultLocale,
	})
}

// parseWeekStartAt 将字符串形式的周起始日配置转换为 Carbon 使用的 Weekday 类型。
//
// 该函数保留 defaultWeekStartAt 可通过 -X 注入字符串的既有语义，同时兼容新版 carbon
// 将 WeekStartsAt 从 string 改为 carbon.Weekday 的 API 变化。无法识别的值返回 carbon.Monday，
// 与当前项目默认配置保持一致。
func parseWeekStartAt(weekStartAt string) carbon.Weekday {
	switch strings.ToLower(weekStartAt) {
	case "sunday":
		return carbon.Sunday
	case "monday":
		return carbon.Monday
	case "tuesday":
		return carbon.Tuesday
	case "wednesday":
		return carbon.Wednesday
	case "thursday":
		return carbon.Thursday
	case "friday":
		return carbon.Friday
	case "saturday":
		return carbon.Saturday
	default:
		return carbon.Monday
	}
}

// Now 返回当前时间的 Carbon 实例。
//
// 返回值:
//   - *carbon.Carbon: 表示当前时间的 Carbon 实例
//
// 示例:
//
//	now := time.Now()
//	fmt.Println(now.ToDateTimeString()) // 输出类似：2025-01-02 15:04:05
func Now() *carbon.Carbon {
	return carbon.Now()
}

// copyNow 返回当前时间的副本。
//
// 该函数用于避免直接在 carbon.Now() 返回值上执行加减操作：
// 在测试时间被冻结时，carbon.Now() 会返回全局 frozen now 指针，
// 而 Add/Sub 系列方法会原地修改 Carbon 实例，直接调用会污染后续 Now() 结果。
// 同时，invalid Carbon 不能安全调用 Copy()，因此需要先保留错误承载对象并直接返回。
func copyNow() *carbon.Carbon {
	now := carbon.Now()
	if now.IsInvalid() {
		return now
	}
	return now.Copy()
}

// Yesterday 返回昨天同一时间的 Carbon 实例。
//
// 返回值:
//   - *carbon.Carbon: 表示昨天同一时间的 Carbon 实例
//
// 示例:
//
//	yesterday := time.Yesterday()
//	fmt.Println(yesterday.ToDateTimeString()) // 输出昨天的日期和当前时间
func Yesterday() *carbon.Carbon {
	return carbon.Yesterday()
}

// Tomorrow 返回明天同一时间的 Carbon 实例。
//
// 返回值:
//   - *carbon.Carbon: 表示明天同一时间的 Carbon 实例
//
// 示例:
//
//	tomorrow := time.Tomorrow()
//	fmt.Println(tomorrow.ToDateTimeString()) // 输出明天的日期和当前时间
func Tomorrow() *carbon.Carbon {
	return carbon.Tomorrow()
}

// DayAfterTomorrow 返回后天同一时间的 Carbon 实例。
//
// 返回值:
//   - *carbon.Carbon: 表示后天同一时间的 Carbon 实例
//
// 示例:
//
//	dayAfterTomorrow := time.DayAfterTomorrow()
//	fmt.Println(dayAfterTomorrow.ToDateTimeString()) // 输出后天的日期和当前时间
func DayAfterTomorrow() *carbon.Carbon {
	return copyNow().AddDays(2)
}

// DayBeforeYesterday 返回前天同一时间的 Carbon 实例。
//
// 返回值:
//   - *carbon.Carbon: 表示前天同一时间的 Carbon 实例
//
// 示例:
//
//	dayBeforeYesterday := time.DayBeforeYesterday()
//	fmt.Println(dayBeforeYesterday.ToDateTimeString()) // 输出前天的日期和当前时间
func DayBeforeYesterday() *carbon.Carbon {
	return copyNow().SubDays(2)
}

// LastWeek 返回上周同一时间的 Carbon 实例。
//
// 返回值:
//   - *carbon.Carbon: 表示上周同一时间的 Carbon 实例
//
// 示例:
//
//	lastWeek := time.LastWeek()
//	fmt.Println(lastWeek.ToDateTimeString()) // 输出上周的日期和当前时间
func LastWeek() *carbon.Carbon {
	return copyNow().SubWeek()
}

// LastMonth 返回上个月同一时间的 Carbon 实例。
//
// 返回值:
//   - *carbon.Carbon: 表示上个月同一时间的 Carbon 实例
//
// 示例:
//
//	lastMonth := time.LastMonth()
//	fmt.Println(lastMonth.ToDateTimeString()) // 输出上个月的日期和当前时间
func LastMonth() *carbon.Carbon {
	return copyNow().SubMonthNoOverflow()
}

// NextWeek 返回下周同一时间的 Carbon 实例。
//
// 返回值:
//   - *carbon.Carbon: 表示下周同一时间的 Carbon 实例
//
// 示例:
//
//	nextWeek := time.NextWeek()
//	fmt.Println(nextWeek.ToDateTimeString()) // 输出下周的日期和当前时间
func NextWeek() *carbon.Carbon {
	return copyNow().AddWeek()
}

// NextMonth 返回下个月同一时间的 Carbon 实例。
//
// 返回值:
//   - *carbon.Carbon: 表示下个月同一时间的 Carbon 实例
//
// 示例:
//
//	nextMonth := time.NextMonth()
//	fmt.Println(nextMonth.ToDateTimeString()) // 输出下个月的日期和当前时间
func NextMonth() *carbon.Carbon {
	return copyNow().AddMonthNoOverflow()
}

// LastYear 返回去年同一时间的 Carbon 实例。
//
// 返回值:
//   - *carbon.Carbon: 表示去年同一时间的 Carbon 实例
//
// 示例:
//
//	lastYear := time.LastYear()
//	fmt.Println(lastYear.ToDateTimeString()) // 输出去年的日期和当前时间
func LastYear() *carbon.Carbon {
	return copyNow().SubYear()
}

// NextYear 返回明年同一时间的 Carbon 实例。
//
// 返回值:
//   - *carbon.Carbon: 表示明年同一时间的 Carbon 实例
//
// 示例:
//
//	nextYear := time.NextYear()
//	fmt.Println(nextYear.ToDateTimeString()) // 输出明年的日期和当前时间
func NextYear() *carbon.Carbon {
	return copyNow().AddYear()
}
