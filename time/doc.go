// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package time 提供了基于 carbon 库的时间处理工具包，支持丰富的时间操作和格式化功能。

包配置：
本包提供了多个可在编译时配置的变量，可以通过 go build 的 -ldflags 参数进行修改。
完整的编译示例：

go build -ldflags "

	-X 'github.com/fsyyft/fsyyft-go/time.defaultDateTimeLayout=2006-01-02 15:04:05'
	-X 'github.com/fsyyft/fsyyft-go/time.defaultTimezone=UTC'
	-X 'github.com/fsyyft/fsyyft-go/time.defaultWeekStartAt=Sunday'
	-X 'github.com/fsyyft/fsyyft-go/time.defaultLocale=en'

主要特性：

  - 基于 Carbon 库，提供强大的时间处理能力
  - 支持多种时间格式的解析和转换
  - 提供丰富的时间计算和比较功能
  - 支持时区处理
  - 支持自然语言时间描述
  - 链式操作支持

基本使用：

1. 创建时间：

	// 获取当前时间
	now := time.Now()

	// 从字符串创建时间
	t1, err := time.Parse("2006-01-02 15:04:05", "2024-03-18 10:30:00")
	if err != nil {
	    panic(err)
	}

	// 从时间戳创建
	t2 := time.FromTimestamp(1710728400)

2. 时间格式化：

	now := time.Now()

	// 标准格式
	fmt.Println(now.Format("Y-m-d H:i:s"))  // 2024-03-18 10:30:00
	fmt.Println(now.ToDateTimeString())      // 2024-03-18 10:30:00
	fmt.Println(now.ToDateString())          // 2024-03-18
	fmt.Println(now.ToTimeString())          // 10:30:00

3. 时间计算：

	now := time.Now()

	// 增加时间
	future := now.AddDays(7)
	future = now.AddHours(24)
	future = now.AddMinutes(30)

	// 减少时间
	past := now.SubDays(7)
	past = now.SubHours(24)
	past = now.SubMinutes(30)

4. 时间比较：

	t1 := time.Now()
	t2 := time.Now().AddDays(1)

	// 比较
	if t1.Lt(t2) {
	    fmt.Println("t1 早于 t2")
	}

	if t2.Gt(t1) {
	    fmt.Println("t2 晚于 t1")
	}

	if t1.Eq(t1) {
	    fmt.Println("时间相等")
	}

5. 时间范围判断：

	now := time.Now()

	// 判断是否在某个时间范围内
	start := now.SubDays(7)
	end := now.AddDays(7)

	if now.Between(start, end) {
	    fmt.Println("在时间范围内")
	}

6. 时区处理：

	// 设置时区
	t := time.Now().SetTimezone("Asia/Shanghai")

	// 转换时区
	utc := t.ToUTC()
	local := utc.ToLocal()

7. 自然语言：

	now := time.Now()

	// 获取自然语言描述
	fmt.Println(now.DiffForHumans())  // 1 分钟前

	// 获取日期描述
	fmt.Println(now.ToFormattedDateString())  // 2024年03月18日

8. 日期判断：

	now := time.Now()

	// 判断是否是特定日期
	if now.IsMonday() {
	    fmt.Println("今天是星期一")
	}

	if now.IsWeekend() {
	    fmt.Println("今天是周末")
	}

	if now.IsLeapYear() {
	    fmt.Println("今年是闰年")
	}

注意事项：

1. 时区处理：
  - 注意时区转换可能带来的时间差异
  - 建议统一使用 UTC 时间进行存储
  - 仅在展示时转换为本地时间

2. 格式化：
  - 使用预定义的格式化字符串避免错误
  - 注意不同地区的日期格式差异
  - 处理时注意年份的表示范围

3. 性能优化：
  - 避免频繁的时区转换
  - 合理使用时间缓存
  - 注意时间计算的精度要求

更多示例和最佳实践请参考 example/time 目录。
*/
package time
