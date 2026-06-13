// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package time

import (
	"testing"
	stdtime "time"

	"github.com/dromara/carbon/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInit_DefaultCarbonConfiguration 验证包初始化后的 Carbon 全局默认配置。
//
// 该测试确保 init 会将默认布局、时区、每周起始日和语言环境同步到 Carbon 全局状态。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestInit_DefaultCarbonConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		description string
		give        func() string
		want        string
	}{
		{
			name:        "global-state/layout",
			description: "验证包初始化会把 Carbon 默认日期时间格式设置为项目默认格式。",
			give:        func() string { return carbon.DefaultLayout },
			want:        defaultDateTimeLayout,
		},
		{
			name:        "global-state/timezone",
			description: "验证包初始化会把 Carbon 默认时区设置为项目默认时区。",
			give:        func() string { return carbon.DefaultTimezone },
			want:        defaultTimezone,
		},
		{
			name:        "global-state/week-starts-at",
			description: "验证包初始化会把 Carbon 默认每周起始日设置为项目默认值。",
			give:        func() string { return carbon.DefaultWeekStartsAt },
			want:        defaultWeekStartAt,
		},
		{
			name:        "global-state/locale",
			description: "验证包初始化会把 Carbon 默认语言环境设置为项目默认语言。",
			give:        func() string { return carbon.DefaultLocale },
			want:        defaultLocale,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			assert.Equal(t, tt.want, tt.give())
		})
	}
}

// TestNow_RealClockWithinCallWindow 验证 Now 返回真实当前时间。
//
// 该测试在清理测试时钟后，断言 Now 的结果位于调用前后的真实时间窗口内。
//
// 参数：
//   - t: 测试上下文，用于注册清理逻辑和报告断言失败。
func TestNow_RealClockWithinCallWindow(t *testing.T) {
	// 该用例验证未冻结测试时间时，Now 返回值位于调用前后的真实时间窗口内。
	carbon.CleanTestNow()
	t.Cleanup(carbon.CleanTestNow)

	loc, err := stdtime.LoadLocation(defaultTimezone)
	require.NoError(t, err)

	before := stdtime.Now().In(loc)
	got := Now()
	after := stdtime.Now().In(loc)

	requireValidCarbon(t, got)
	assert.Equal(t, defaultTimezone, got.Timezone())
	assert.False(t, got.StdTime().Before(before), "Now should not be earlier than the call window")
	assert.False(t, got.StdTime().After(after), "Now should not be later than the call window")
}

// TestNow_DefaultFormattingTimezoneAndLocale 验证 Now 的默认格式化、时区和语言环境行为。
//
// 该测试通过冻结时间，断言 Now 返回值继承默认 layout、timezone、locale 和 week start 配置。
//
// 参数：
//   - t: 测试上下文，用于冻结测试时间、注册清理逻辑和报告断言失败。
func TestNow_DefaultFormattingTimezoneAndLocale(t *testing.T) {
	// 该用例验证冻结时间后，Now 返回值继承默认格式、时区、语言环境和周起始日配置。
	freezeNow(t, mustCarbon(t, 2024, 3, 18, 10, 30, 15, 123, defaultTimezone))

	got := Now()
	requireValidCarbon(t, got)

	assert.Equal(t, "2024-03-18T10:30:15.123Z", got.String())
	assert.Equal(t, "2024-03-18 10:30:15", got.ToDateTimeString())
	assert.Equal(t, "2024-03-18", got.ToDateString())
	assert.Equal(t, "10:30:15", got.ToTimeString())
	assert.Equal(t, "2024-03-18 02:30:15", got.Copy().ToDateTimeString(carbon.UTC))
	assert.Equal(t, "2024/03/18 10:30:15", got.Copy().Layout("2006/01/02 15:04:05"))
	assert.Equal(t, defaultTimezone, got.Timezone())
	assert.Equal(t, 8*60*60, got.ZoneOffset())
	assert.Equal(t, defaultLocale, got.Locale())
	assert.Equal(t, defaultWeekStartAt, got.WeekStartsAt())
	assert.Equal(t, stdtime.Monday, got.StdTime().Weekday())
	assert.Equal(t, int((123 * stdtime.Millisecond).Nanoseconds()), got.StdTime().Nanosecond())
}

// TestRelativeAPIs_ReturnExpectedOffsets 验证相对时间 API 的标准偏移结果。
//
// 该测试以固定基准时间驱动所有公开 API，确保本地时间、UTC 时间、星期和毫秒精度符合预期。
//
// 参数：
//   - t: 测试上下文，用于运行表驱动子测试、冻结基准时间和报告断言失败。
func TestRelativeAPIs_ReturnExpectedOffsets(t *testing.T) {
	tests := []struct {
		name         string
		description  string
		giveNow      func(t *testing.T) *carbon.Carbon
		giveFunction func() *carbon.Carbon
		wantDateTime string
		wantUTC      string
		wantWeekday  stdtime.Weekday
	}{
		{
			name:         "success/now",
			description:  "验证 Now 在冻结基准时间下返回基准时间本身，并保留默认时区和毫秒精度。",
			giveNow:      baseCarbon,
			giveFunction: Now,
			wantDateTime: "2024-03-18 10:30:15",
			wantUTC:      "2024-03-18 02:30:15",
			wantWeekday:  stdtime.Monday,
		},
		{
			name:         "success/yesterday",
			description:  "验证 Yesterday 返回冻结基准时间前一天的同一时刻。",
			giveNow:      baseCarbon,
			giveFunction: Yesterday,
			wantDateTime: "2024-03-17 10:30:15",
			wantUTC:      "2024-03-17 02:30:15",
			wantWeekday:  stdtime.Sunday,
		},
		{
			name:         "success/tomorrow",
			description:  "验证 Tomorrow 返回冻结基准时间后一天的同一时刻。",
			giveNow:      baseCarbon,
			giveFunction: Tomorrow,
			wantDateTime: "2024-03-19 10:30:15",
			wantUTC:      "2024-03-19 02:30:15",
			wantWeekday:  stdtime.Tuesday,
		},
		{
			name:         "success/day-after-tomorrow",
			description:  "验证 DayAfterTomorrow 返回冻结基准时间后两天的同一时刻。",
			giveNow:      baseCarbon,
			giveFunction: DayAfterTomorrow,
			wantDateTime: "2024-03-20 10:30:15",
			wantUTC:      "2024-03-20 02:30:15",
			wantWeekday:  stdtime.Wednesday,
		},
		{
			name:         "success/day-before-yesterday",
			description:  "验证 DayBeforeYesterday 返回冻结基准时间前两天的同一时刻。",
			giveNow:      baseCarbon,
			giveFunction: DayBeforeYesterday,
			wantDateTime: "2024-03-16 10:30:15",
			wantUTC:      "2024-03-16 02:30:15",
			wantWeekday:  stdtime.Saturday,
		},
		{
			name:         "success/last-week",
			description:  "验证 LastWeek 返回冻结基准时间前一周的同一时刻。",
			giveNow:      baseCarbon,
			giveFunction: LastWeek,
			wantDateTime: "2024-03-11 10:30:15",
			wantUTC:      "2024-03-11 02:30:15",
			wantWeekday:  stdtime.Monday,
		},
		{
			name:         "success/last-month",
			description:  "验证 LastMonth 返回冻结基准时间前一个自然月的同一时刻。",
			giveNow:      baseCarbon,
			giveFunction: LastMonth,
			wantDateTime: "2024-02-18 10:30:15",
			wantUTC:      "2024-02-18 02:30:15",
			wantWeekday:  stdtime.Sunday,
		},
		{
			name:         "success/next-week",
			description:  "验证 NextWeek 返回冻结基准时间后一周的同一时刻。",
			giveNow:      baseCarbon,
			giveFunction: NextWeek,
			wantDateTime: "2024-03-25 10:30:15",
			wantUTC:      "2024-03-25 02:30:15",
			wantWeekday:  stdtime.Monday,
		},
		{
			name:         "success/next-month",
			description:  "验证 NextMonth 返回冻结基准时间后一个自然月的同一时刻。",
			giveNow:      baseCarbon,
			giveFunction: NextMonth,
			wantDateTime: "2024-04-18 10:30:15",
			wantUTC:      "2024-04-18 02:30:15",
			wantWeekday:  stdtime.Thursday,
		},
		{
			name:         "success/last-year",
			description:  "验证 LastYear 返回冻结基准时间前一年的同一时刻。",
			giveNow:      baseCarbon,
			giveFunction: LastYear,
			wantDateTime: "2023-03-18 10:30:15",
			wantUTC:      "2023-03-18 02:30:15",
			wantWeekday:  stdtime.Saturday,
		},
		{
			name:         "success/next-year",
			description:  "验证 NextYear 返回冻结基准时间后一年的同一时刻。",
			giveNow:      baseCarbon,
			giveFunction: NextYear,
			wantDateTime: "2025-03-18 10:30:15",
			wantUTC:      "2025-03-18 02:30:15",
			wantWeekday:  stdtime.Tuesday,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			freezeNow(t, tt.giveNow(t))

			got := tt.giveFunction()

			requireValidCarbon(t, got)
			assert.Equal(t, tt.wantDateTime, got.ToDateTimeString())
			assert.Equal(t, tt.wantUTC, got.Copy().ToDateTimeString(carbon.UTC))
			assert.Equal(t, tt.wantWeekday, got.StdTime().Weekday())
			assert.Equal(t, int((123 * stdtime.Millisecond).Nanoseconds()), got.StdTime().Nanosecond())
			assert.Equal(t, defaultTimezone, got.Timezone())
		})
	}
}

// TestRelativeAPIs_ReturnInvalidCarbonWithoutPanic 验证相对时间 API 对无效 Carbon 配置的兼容性。
//
// 该测试将默认时区设置为无效值，确保各 API 返回错误承载 Carbon，而不是因复制 invalid Carbon 触发 panic。
//
// 参数：
//   - t: 测试上下文，用于运行表驱动子测试、恢复 Carbon 全局状态和报告断言失败。
func TestRelativeAPIs_ReturnInvalidCarbonWithoutPanic(t *testing.T) {
	carbon.CleanTestNow()
	oldTimezone := carbon.DefaultTimezone
	carbon.DefaultTimezone = "Invalid/Timezone"
	t.Cleanup(func() {
		carbon.DefaultTimezone = oldTimezone
		carbon.CleanTestNow()
	})

	tests := []struct {
		name         string
		description  string
		giveFunction func() *carbon.Carbon
	}{
		{
			name:         "success/now",
			description:  "验证 Now 在默认时区无效时返回错误承载 Carbon，而不是触发 panic。",
			giveFunction: Now,
		},
		{
			name:         "success/yesterday",
			description:  "验证 Yesterday 在默认时区无效时返回错误承载 Carbon，而不是触发 panic。",
			giveFunction: Yesterday,
		},
		{
			name:         "success/tomorrow",
			description:  "验证 Tomorrow 在默认时区无效时返回错误承载 Carbon，而不是触发 panic。",
			giveFunction: Tomorrow,
		},
		{
			name:         "success/day-after-tomorrow",
			description:  "验证 DayAfterTomorrow 在默认时区无效时不会因复制 invalid Carbon 而 panic。",
			giveFunction: DayAfterTomorrow,
		},
		{
			name:         "success/day-before-yesterday",
			description:  "验证 DayBeforeYesterday 在默认时区无效时不会因复制 invalid Carbon 而 panic。",
			giveFunction: DayBeforeYesterday,
		},
		{
			name:         "success/last-week",
			description:  "验证 LastWeek 在默认时区无效时不会因复制 invalid Carbon 而 panic。",
			giveFunction: LastWeek,
		},
		{
			name:         "success/last-month",
			description:  "验证 LastMonth 在默认时区无效时不会因复制 invalid Carbon 而 panic。",
			giveFunction: LastMonth,
		},
		{
			name:         "success/next-week",
			description:  "验证 NextWeek 在默认时区无效时不会因复制 invalid Carbon 而 panic。",
			giveFunction: NextWeek,
		},
		{
			name:         "success/next-month",
			description:  "验证 NextMonth 在默认时区无效时不会因复制 invalid Carbon 而 panic。",
			giveFunction: NextMonth,
		},
		{
			name:         "success/last-year",
			description:  "验证 LastYear 在默认时区无效时不会因复制 invalid Carbon 而 panic。",
			giveFunction: LastYear,
		},
		{
			name:         "success/next-year",
			description:  "验证 NextYear 在默认时区无效时不会因复制 invalid Carbon 而 panic。",
			giveFunction: NextYear,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			var got *carbon.Carbon
			require.NotPanics(t, func() {
				got = tt.giveFunction()
			})

			require.NotNil(t, got)
			require.Error(t, got.Error)
		})
	}
}

// TestRelativeAPIs_DoNotMutateFrozenNow 验证相对时间 API 不会污染冻结的当前时间。
//
// 该测试在 Carbon 测试时间被冻结时调用各相对时间 API，确保返回偏移结果的同时 Now 仍保持原始 frozen now。
//
// 参数：
//   - t: 测试上下文，用于运行表驱动子测试、冻结当前时间和报告断言失败。
func TestRelativeAPIs_DoNotMutateFrozenNow(t *testing.T) {
	tests := []struct {
		name         string
		description  string
		giveFunction func() *carbon.Carbon
		wantDateTime string
	}{
		{
			name:         "success/yesterday",
			description:  "验证 Yesterday 返回偏移结果时不会污染 Carbon 全局 frozen now。",
			giveFunction: Yesterday,
			wantDateTime: "2024-03-17 10:30:15",
		},
		{
			name:         "success/tomorrow",
			description:  "验证 Tomorrow 返回偏移结果时不会污染 Carbon 全局 frozen now。",
			giveFunction: Tomorrow,
			wantDateTime: "2024-03-19 10:30:15",
		},
		{
			name:         "success/day-after-tomorrow",
			description:  "验证 DayAfterTomorrow 使用当前时间副本计算后天，避免修改 frozen now。",
			giveFunction: DayAfterTomorrow,
			wantDateTime: "2024-03-20 10:30:15",
		},
		{
			name:         "success/day-before-yesterday",
			description:  "验证 DayBeforeYesterday 使用当前时间副本计算前天，避免修改 frozen now。",
			giveFunction: DayBeforeYesterday,
			wantDateTime: "2024-03-16 10:30:15",
		},
		{
			name:         "success/last-week",
			description:  "验证 LastWeek 使用当前时间副本计算上周，避免修改 frozen now。",
			giveFunction: LastWeek,
			wantDateTime: "2024-03-11 10:30:15",
		},
		{
			name:         "success/last-month",
			description:  "验证 LastMonth 使用当前时间副本计算上月，避免修改 frozen now。",
			giveFunction: LastMonth,
			wantDateTime: "2024-02-18 10:30:15",
		},
		{
			name:         "success/next-week",
			description:  "验证 NextWeek 使用当前时间副本计算下周，避免修改 frozen now。",
			giveFunction: NextWeek,
			wantDateTime: "2024-03-25 10:30:15",
		},
		{
			name:         "success/next-month",
			description:  "验证 NextMonth 使用当前时间副本计算下月，避免修改 frozen now。",
			giveFunction: NextMonth,
			wantDateTime: "2024-04-18 10:30:15",
		},
		{
			name:         "success/last-year",
			description:  "验证 LastYear 使用当前时间副本计算去年，避免修改 frozen now。",
			giveFunction: LastYear,
			wantDateTime: "2023-03-18 10:30:15",
		},
		{
			name:         "success/next-year",
			description:  "验证 NextYear 使用当前时间副本计算明年，避免修改 frozen now。",
			giveFunction: NextYear,
			wantDateTime: "2025-03-18 10:30:15",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			freezeNow(t, baseCarbon(t))

			got := tt.giveFunction()

			requireValidCarbon(t, got)
			assert.Equal(t, tt.wantDateTime, got.ToDateTimeString())
			assert.Equal(t, "2024-03-18 10:30:15", Now().ToDateTimeString())
		})
	}
}

// TestRelativeAPIs_BoundaryDates 验证相对时间 API 在日期边界上的行为。
//
// 该测试覆盖闰日、跨年、月底夹取和年份偏移溢出场景，确保边界日期计算语义稳定。
//
// 参数：
//   - t: 测试上下文，用于运行边界日期子测试、冻结输入时间和报告断言失败。
func TestRelativeAPIs_BoundaryDates(t *testing.T) {
	tests := []struct {
		name         string
		description  string
		giveYear     int
		giveMonth    int
		giveDay      int
		giveFunction func() *carbon.Carbon
		wantDateTime string
	}{
		{
			name:         "boundary/yesterday-from-march-first-in-leap-year",
			description:  "验证闰年 3 月 1 日调用 Yesterday 会回到 2 月 29 日。",
			giveYear:     2024,
			giveMonth:    3,
			giveDay:      1,
			giveFunction: Yesterday,
			wantDateTime: "2024-02-29 00:00:00",
		},
		{
			name:         "boundary/day-before-yesterday-from-march-first-in-leap-year",
			description:  "验证闰年 3 月 1 日调用 DayBeforeYesterday 会回到 2 月 28 日。",
			giveYear:     2024,
			giveMonth:    3,
			giveDay:      1,
			giveFunction: DayBeforeYesterday,
			wantDateTime: "2024-02-28 00:00:00",
		},
		{
			name:         "boundary/tomorrow-from-leap-day",
			description:  "验证闰日调用 Tomorrow 会进入 3 月 1 日。",
			giveYear:     2024,
			giveMonth:    2,
			giveDay:      29,
			giveFunction: Tomorrow,
			wantDateTime: "2024-03-01 00:00:00",
		},
		{
			name:         "boundary/day-after-tomorrow-from-leap-day",
			description:  "验证闰日调用 DayAfterTomorrow 会进入 3 月 2 日。",
			giveYear:     2024,
			giveMonth:    2,
			giveDay:      29,
			giveFunction: DayAfterTomorrow,
			wantDateTime: "2024-03-02 00:00:00",
		},
		{
			name:         "boundary/last-week-across-year",
			description:  "验证年初日期调用 LastWeek 会正确跨到上一年。",
			giveYear:     2024,
			giveMonth:    1,
			giveDay:      3,
			giveFunction: LastWeek,
			wantDateTime: "2023-12-27 00:00:00",
		},
		{
			name:         "boundary/next-week-across-year",
			description:  "验证年末日期调用 NextWeek 会正确跨到下一年。",
			giveYear:     2024,
			giveMonth:    12,
			giveDay:      29,
			giveFunction: NextWeek,
			wantDateTime: "2025-01-05 00:00:00",
		},
		{
			name:         "boundary/last-month-from-leap-day",
			description:  "验证闰日调用 LastMonth 会回到上一月同一天。",
			giveYear:     2024,
			giveMonth:    2,
			giveDay:      29,
			giveFunction: LastMonth,
			wantDateTime: "2024-01-29 00:00:00",
		},
		{
			name:         "boundary/next-month-from-leap-day",
			description:  "验证闰日调用 NextMonth 会进入下一月同一天。",
			giveYear:     2024,
			giveMonth:    2,
			giveDay:      29,
			giveFunction: NextMonth,
			wantDateTime: "2024-03-29 00:00:00",
		},
		{
			name:         "boundary/last-month-clamps-to-leap-february-end",
			description:  "验证闰年 3 月 31 日调用 LastMonth 会夹取到 2 月 29 日。",
			giveYear:     2024,
			giveMonth:    3,
			giveDay:      31,
			giveFunction: LastMonth,
			wantDateTime: "2024-02-29 00:00:00",
		},
		{
			name:         "boundary/next-month-clamps-to-leap-february-end",
			description:  "验证闰年 1 月 31 日调用 NextMonth 会夹取到 2 月 29 日。",
			giveYear:     2024,
			giveMonth:    1,
			giveDay:      31,
			giveFunction: NextMonth,
			wantDateTime: "2024-02-29 00:00:00",
		},
		{
			name:         "boundary/last-month-clamps-to-common-february-end",
			description:  "验证平年 3 月 31 日调用 LastMonth 会夹取到 2 月 28 日。",
			giveYear:     2023,
			giveMonth:    3,
			giveDay:      31,
			giveFunction: LastMonth,
			wantDateTime: "2023-02-28 00:00:00",
		},
		{
			name:         "boundary/next-month-clamps-to-common-february-end",
			description:  "验证平年 1 月 31 日调用 NextMonth 会夹取到 2 月 28 日。",
			giveYear:     2023,
			giveMonth:    1,
			giveDay:      31,
			giveFunction: NextMonth,
			wantDateTime: "2023-02-28 00:00:00",
		},
		{
			name:         "boundary/last-month-across-year",
			description:  "验证年初月底调用 LastMonth 会正确跨到上一年 12 月。",
			giveYear:     2024,
			giveMonth:    1,
			giveDay:      31,
			giveFunction: LastMonth,
			wantDateTime: "2023-12-31 00:00:00",
		},
		{
			name:         "boundary/next-month-across-year",
			description:  "验证年末月底调用 NextMonth 会正确跨到下一年 1 月。",
			giveYear:     2024,
			giveMonth:    12,
			giveDay:      31,
			giveFunction: NextMonth,
			wantDateTime: "2025-01-31 00:00:00",
		},
		{
			name:         "boundary/last-year-from-leap-day-overflows-to-march-first",
			description:  "验证闰日调用 LastYear 保持 Carbon 年份偏移的溢出语义，落到 3 月 1 日。",
			giveYear:     2024,
			giveMonth:    2,
			giveDay:      29,
			giveFunction: LastYear,
			wantDateTime: "2023-03-01 00:00:00",
		},
		{
			name:         "boundary/next-year-from-leap-day-overflows-to-march-first",
			description:  "验证闰日调用 NextYear 保持 Carbon 年份偏移的溢出语义，落到 3 月 1 日。",
			giveYear:     2024,
			giveMonth:    2,
			giveDay:      29,
			giveFunction: NextYear,
			wantDateTime: "2025-03-01 00:00:00",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			freezeNow(t, mustCarbon(t, tt.giveYear, tt.giveMonth, tt.giveDay, 0, 0, 0, 0, defaultTimezone))

			got := tt.giveFunction()

			requireValidCarbon(t, got)
			assert.Equal(t, tt.wantDateTime, got.ToDateTimeString())
			assert.Equal(t, defaultTimezone, got.Timezone())
		})
	}
}

// baseCarbon 返回相对时间测试使用的统一基准时间。
//
// 该辅助函数集中定义常规用例的 frozen now，确保不同测试用例共享同一基准时间、时区和毫秒精度。
//
// 参数：
//   - t: 测试上下文，用于创建 Carbon 失败时报告错误并标记辅助函数调用栈。
//
// 返回：
//   - *carbon.Carbon: 固定在默认时区的基准 Carbon 时间实例。
func baseCarbon(t *testing.T) *carbon.Carbon {
	t.Helper()
	return mustCarbon(t, 2024, 3, 18, 10, 30, 15, 123, defaultTimezone)
}

// mustCarbon 创建指定日期时间的有效 Carbon 实例。
//
// 该辅助函数用于生成测试输入时间，并立即校验 Carbon 是否创建成功，避免无效测试数据进入断言流程。
//
// 参数：
//   - t: 测试上下文，用于报告 Carbon 创建失败并标记辅助函数调用栈。
//   - year: 年份字段。
//   - month: 月份字段。
//   - day: 日期字段。
//   - hour: 小时字段。
//   - minute: 分钟字段。
//   - second: 秒字段。
//   - millisecond: 毫秒字段。
//   - timezone: 时区名称，用于指定 Carbon 实例所在时区。
//
// 返回：
//   - *carbon.Carbon: 创建成功且无错误的 Carbon 时间实例。
func mustCarbon(t *testing.T, year, month, day, hour, minute, second, millisecond int, timezone string) *carbon.Carbon {
	t.Helper()

	got := carbon.CreateFromDateTimeMilli(year, month, day, hour, minute, second, millisecond, timezone)
	requireValidCarbon(t, got)
	return got
}

// freezeNow 冻结 Carbon 当前时间并在测试结束后清理。
//
// 该辅助函数用于让相对时间 API 的测试保持确定性，同时通过 t.Cleanup 避免 Carbon 全局测试时钟污染后续用例。
//
// 参数：
//   - t: 测试上下文，用于注册清理逻辑、报告无效 frozen now 并标记辅助函数调用栈。
//   - now: 要设置为 Carbon 测试当前时间的有效 Carbon 实例。
func freezeNow(t *testing.T, now *carbon.Carbon) {
	t.Helper()

	requireValidCarbon(t, now)
	carbon.SetTestNow(now)
	t.Cleanup(carbon.CleanTestNow)
}

// requireValidCarbon 校验 Carbon 实例非空且没有错误。
//
// 该辅助函数统一测试前置校验和返回值校验逻辑，使失败位置指向调用方并减少重复断言代码。
//
// 参数：
//   - t: 测试上下文，用于报告 Carbon 校验失败并标记辅助函数调用栈。
//   - got: 待校验的 Carbon 实例。
func requireValidCarbon(t *testing.T, got *carbon.Carbon) {
	t.Helper()

	require.NotNil(t, got)
	require.NoError(t, got.Error)
}
