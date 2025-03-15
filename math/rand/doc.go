// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package rand 提供了一系列随机数生成的工具函数，包括数值范围随机和中文字符随机生成。

主要特性：

  - 支持自定义随机数生成器
  - 提供整数范围随机数生成
  - 支持中文字符随机生成
  - 内置中文姓氏随机生成

基本功能：

1. 整数范围随机：

	// 生成 int64 范围随机数
	num1 := rand.Int63n(nil, 1, 100)  // 使用默认随机源

	// 使用自定义随机源
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	num2 := rand.Int63n(random, 1, 100)

	// 生成 int 范围随机数
	num3 := rand.Intn(nil, 1, 10)     // 使用默认随机源
	num4 := rand.Intn(random, 1, 10)   // 使用自定义随机源

2. 中文字符生成：

	// 生成随机汉字
	char1 := rand.Chinese(nil)         // 使用默认随机源
	char2 := rand.Chinese(random)      // 使用自定义随机源

	// 生成随机中文姓氏
	name1 := rand.ChineseLastName(nil)    // 使用默认随机源
	name2 := rand.ChineseLastName(random) // 使用自定义随机源

功能说明：

1. Int63n：
  - 生成 [min, max) 范围内的 int64 随机数
  - 支持自定义随机源
  - 使用 nil 随机源时采用默认随机源

2. Intn：
  - 生成 [min, max) 范围内的 int 随机数
  - 支持自定义随机源
  - 使用 nil 随机源时采用默认随机源

3. Chinese：
  - 生成随机汉字字符串
  - 使用 Unicode 范围 [19968, 40869]
  - 覆盖常用汉字范围

4. ChineseLastName：
  - 生成随机中文姓氏
  - 包含常见单姓和复姓
  - 使用预定义的姓氏字符集

使用建议：

1. 随机源：
  - 需要重复结果时使用固定种子
  - 需要随机结果时使用时间戳作为种子
  - 高并发场景建议使用自定义随机源

2. 范围随机：
  - 注意范围参数的有效性
  - max 必须大于 min
  - 结果范围是左闭右开区间 [min, max)

3. 中文生成：
  - 随机汉字可能包含生僻字
  - 姓氏生成仅包含预定义集合
  - 输出为 UTF-8 编码字符串

注意事项：

1. 性能考虑：
  - 频繁调用时建议复用随机源
  - 避免在热点代码中频繁创建随机源
  - 合理使用默认随机源

2. 并发安全：
  - 默认随机源是并发安全的
  - 自定义随机源需要自行保证并发安全
  - 高并发场景建议使用独立随机源

更多示例请参考 example/math/rand 目录。
*/
package rand
