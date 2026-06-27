// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package time

import (
	"strings"

	"github.com/dromara/carbon/v2"
)

var (
	// defaultDateTimeLayout 指定写入 carbon 全局默认配置的日期时间布局。
	//
	// 布局遵循 Go 标准时间格式，默认值为 2006-01-02T15:04:05.000Z。
	// 可在编译时通过 -ldflags -X 覆盖，例如：
	// go build -ldflags "-X 'github.com/fsyyft-go/kit/time.defaultDateTimeLayout=2006-01-02 15:04:05'"
	defaultDateTimeLayout = "2006-01-02T15:04:05.000Z"

	// defaultTimezone 指定写入 carbon 全局默认配置的时区名称。
	//
	// 默认值 PRC 表示中国标准时间。可在编译时通过 -ldflags -X 覆盖，例如：
	// go build -ldflags "-X 'github.com/fsyyft-go/kit/time.defaultTimezone=UTC'"
	// 常用时区值包括 PRC、UTC、Asia/Shanghai 和 America/New_York。
	defaultTimezone = "PRC"

	// defaultWeekStartAt 指定写入 carbon 全局默认配置的每周起始日。
	//
	// 默认值为 Monday。可在编译时通过 -ldflags -X 覆盖，例如：
	// go build -ldflags "-X 'github.com/fsyyft-go/kit/time.defaultWeekStartAt=Sunday'"
	// 支持以下大小写不敏感的枚举字符串，空字符串或无法识别的值会回退为 Monday：
	//   - Sunday: 以星期日作为每周起始日，对应 carbon.Sunday。
	//   - Monday: 以星期一作为每周起始日，对应 carbon.Monday。
	//   - Tuesday: 以星期二作为每周起始日，对应 carbon.Tuesday。
	//   - Wednesday: 以星期三作为每周起始日，对应 carbon.Wednesday。
	//   - Thursday: 以星期四作为每周起始日，对应 carbon.Thursday。
	//   - Friday: 以星期五作为每周起始日，对应 carbon.Friday。
	//   - Saturday: 以星期六作为每周起始日，对应 carbon.Saturday。
	defaultWeekStartAt = "Monday"

	// defaultLocale 指定写入 carbon 全局默认配置的语言环境。
	//
	// 默认值为 zh-CN。可在编译时通过 -ldflags -X 覆盖，例如：
	// go build -ldflags "-X 'github.com/fsyyft-go/kit/time.defaultLocale=en'"
	// 具体语言支持范围由 carbon 库决定。
	defaultLocale = "zh-CN"
)

// init 在包加载时把当前默认变量写入 carbon 的全局默认配置。
//
// 该初始化只会执行一次；若通过 -ldflags -X 覆盖默认变量，必须在程序启动前完成。
// 写入的是 carbon 全局状态，会影响同一进程中依赖 carbon 默认配置的代码。
//
// 参数：无。
func init() {
	carbon.SetDefault(carbon.Default{
		Layout:       defaultDateTimeLayout,
		Timezone:     defaultTimezone,
		WeekStartsAt: parseWeekStartAt(defaultWeekStartAt),
		Locale:       defaultLocale,
	})
}

// parseWeekStartAt 将字符串形式的周起始日配置转换为 carbon 使用的 Weekday 类型。
//
// 该函数保留 defaultWeekStartAt 可通过 -ldflags -X 注入字符串的既有语义，同时兼容新版 carbon
// 将 WeekStartsAt 从 string 改为 carbon.Weekday 的 API 变化。无法识别的值返回 carbon.Monday，
// 与当前项目默认配置保持一致。
//
// weekStartAt 支持以下大小写不敏感的枚举值：
//   - Sunday: 以星期日作为每周起始日，对应 carbon.Sunday。
//   - Monday: 以星期一作为每周起始日，对应 carbon.Monday。
//   - Tuesday: 以星期二作为每周起始日，对应 carbon.Tuesday。
//   - Wednesday: 以星期三作为每周起始日，对应 carbon.Wednesday。
//   - Thursday: 以星期四作为每周起始日，对应 carbon.Thursday。
//   - Friday: 以星期五作为每周起始日，对应 carbon.Friday。
//   - Saturday: 以星期六作为每周起始日，对应 carbon.Saturday。
//
// 参数：
//   - weekStartAt: 周起始日名称；空字符串或无法识别的值会回退为 carbon.Monday。
//
// 返回：
//   - carbon.Weekday: weekStartAt 对应的 carbon 星期枚举；空字符串或无法识别的值返回 carbon.Monday。
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
// 返回值继承 carbon 全局默认布局、时区、每周起始日和语言环境；当默认配置无效时，返回值会携带
// Carbon 错误，调用方应在格式化或继续计算前检查 Error 或 IsInvalid。
//
// 参数：无。
//
// 返回：
//   - *carbon.Carbon: 表示当前时间的 Carbon 实例，可能因无效 carbon 默认配置而处于 invalid 状态。
func Now() *carbon.Carbon {
	return carbon.Now()
}

// copyNow 返回当前时间的副本。
//
// 该函数用于避免直接在 carbon.Now() 返回值上执行加减操作：在测试时间被冻结时，carbon.Now()
// 会返回全局 frozen now 指针，而 Add/Sub 系列方法会原地修改 Carbon 实例，直接调用会污染后续
// Now 结果。invalid Carbon 不能安全调用 Copy，因此该函数会保留错误承载对象并直接返回。
//
// 参数：无。
//
// 返回：
//   - *carbon.Carbon: 有效当前时间的副本；如果 carbon.Now 返回 invalid Carbon，则返回原始错误承载实例。
func copyNow() *carbon.Carbon {
	now := carbon.Now()
	if now.IsInvalid() {
		return now
	}
	return now.Copy()
}

// Yesterday 返回昨天同一时刻的 Carbon 实例。
//
// 返回值使用 carbon 的昨天计算语义，并继承 carbon 全局默认配置；当默认配置无效时，返回值会携带
// Carbon 错误。
//
// 参数：无。
//
// 返回：
//   - *carbon.Carbon: 当前时间前一天同一时刻的 Carbon 实例，可能因无效 carbon 默认配置而处于 invalid 状态。
func Yesterday() *carbon.Carbon {
	return carbon.Yesterday()
}

// Tomorrow 返回明天同一时刻的 Carbon 实例。
//
// 返回值使用 carbon 的明天计算语义，并继承 carbon 全局默认配置；当默认配置无效时，返回值会携带
// Carbon 错误。
//
// 参数：无。
//
// 返回：
//   - *carbon.Carbon: 当前时间后一天同一时刻的 Carbon 实例，可能因无效 carbon 默认配置而处于 invalid 状态。
func Tomorrow() *carbon.Carbon {
	return carbon.Tomorrow()
}

// DayAfterTomorrow 返回后天同一时刻的 Carbon 实例。
//
// 该函数基于当前时间副本加两天，避免 Carbon 测试时间被冻结时修改全局 frozen now；当默认配置无效时，
// 返回值会携带 Carbon 错误。
//
// 参数：无。
//
// 返回：
//   - *carbon.Carbon: 当前时间后两天同一时刻的 Carbon 实例，可能因无效 carbon 默认配置而处于 invalid 状态。
func DayAfterTomorrow() *carbon.Carbon {
	return copyNow().AddDays(2)
}

// DayBeforeYesterday 返回前天同一时刻的 Carbon 实例。
//
// 该函数基于当前时间副本减两天，避免 Carbon 测试时间被冻结时修改全局 frozen now；当默认配置无效时，
// 返回值会携带 Carbon 错误。
//
// 参数：无。
//
// 返回：
//   - *carbon.Carbon: 当前时间前两天同一时刻的 Carbon 实例，可能因无效 carbon 默认配置而处于 invalid 状态。
func DayBeforeYesterday() *carbon.Carbon {
	return copyNow().SubDays(2)
}

// LastWeek 返回上周同一时刻的 Carbon 实例。
//
// 该函数基于当前时间副本减一周，避免 Carbon 测试时间被冻结时修改全局 frozen now；当默认配置无效时，
// 返回值会携带 Carbon 错误。
//
// 参数：无。
//
// 返回：
//   - *carbon.Carbon: 当前时间前一周同一时刻的 Carbon 实例，可能因无效 carbon 默认配置而处于 invalid 状态。
func LastWeek() *carbon.Carbon {
	return copyNow().SubWeek()
}

// LastMonth 返回上个月对应时刻的 Carbon 实例。
//
// 该函数基于当前时间副本减一个自然月，并使用 carbon.SubMonthNoOverflow 在目标月份天数不足时夹取到月末；
// 当默认配置无效时，返回值会携带 Carbon 错误。
//
// 参数：无。
//
// 返回：
//   - *carbon.Carbon: 当前时间前一个自然月的 Carbon 实例，月底溢出时按 carbon 无溢出语义夹取到目标月末。
func LastMonth() *carbon.Carbon {
	return copyNow().SubMonthNoOverflow()
}

// NextWeek 返回下周同一时刻的 Carbon 实例。
//
// 该函数基于当前时间副本加一周，避免 Carbon 测试时间被冻结时修改全局 frozen now；当默认配置无效时，
// 返回值会携带 Carbon 错误。
//
// 参数：无。
//
// 返回：
//   - *carbon.Carbon: 当前时间后一周同一时刻的 Carbon 实例，可能因无效 carbon 默认配置而处于 invalid 状态。
func NextWeek() *carbon.Carbon {
	return copyNow().AddWeek()
}

// NextMonth 返回下个月对应时刻的 Carbon 实例。
//
// 该函数基于当前时间副本加一个自然月，并使用 carbon.AddMonthNoOverflow 在目标月份天数不足时夹取到月末；
// 当默认配置无效时，返回值会携带 Carbon 错误。
//
// 参数：无。
//
// 返回：
//   - *carbon.Carbon: 当前时间后一个自然月的 Carbon 实例，月底溢出时按 carbon 无溢出语义夹取到目标月末。
func NextMonth() *carbon.Carbon {
	return copyNow().AddMonthNoOverflow()
}

// LastYear 返回去年对应时刻的 Carbon 实例。
//
// 该函数基于当前时间副本减一年，并保留 carbon.SubYear 的年份偏移语义；例如闰日偏移到平年时会按
// carbon 规则溢出到 3 月 1 日。当默认配置无效时，返回值会携带 Carbon 错误。
//
// 参数：无。
//
// 返回：
//   - *carbon.Carbon: 当前时间前一年的 Carbon 实例，年份边界和闰日由 carbon.SubYear 决定。
func LastYear() *carbon.Carbon {
	return copyNow().SubYear()
}

// NextYear 返回明年对应时刻的 Carbon 实例。
//
// 该函数基于当前时间副本加一年，并保留 carbon.AddYear 的年份偏移语义；例如闰日偏移到平年时会按
// carbon 规则溢出到 3 月 1 日。当默认配置无效时，返回值会携带 Carbon 错误。
//
// 参数：无。
//
// 返回：
//   - *carbon.Carbon: 当前时间后一年的 Carbon 实例，年份边界和闰日由 carbon.AddYear 决定。
func NextYear() *carbon.Carbon {
	return copyNow().AddYear()
}
