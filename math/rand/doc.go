// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package rand 提供围绕 math/rand 的范围随机数和中文字符随机工具。
//
// 本包在标准库 math/rand 的基础上补充常用的边界随机和中文字符随机能力，适用于需要
// 生成普通伪随机值、测试数据或示例数据的场景。由于底层依赖 math/rand，本包不适用于
// 密码学、安全令牌、密钥材料或其他需要不可预测随机性的场景。
//
// 主要特性：
//
//   - 支持通过 nil random 参数使用 math/rand 的包级随机源。
//   - 支持传入自定义 *rand.Rand，以便调用方控制种子、复现性和并发保护。
//   - 提供 Int63n 和 Intn，用于生成指定左闭右开区间内的整数。
//   - 提供 Chinese 和 ChineseLastName，用于生成随机汉字字符串和姓氏用字。
//
// 基本功能：
//
// Int63n 和 Intn 根据 min、max 生成 [min, max) 范围内的随机数。本包不会额外校验边界，
// 而是直接将 max-min 的结果作为 math/rand.Int63n 或 math/rand.Intn 的上界。正常用法应保证
// max-min 在对应整数类型中可表示且为正；否则行为由整数溢出结果和标准库对非正上界的
// panic 语义共同决定。调用方应在边界来自外部输入或运行时计算时提前校验。
//
// Chinese 从预设 Unicode 汉字码点区间 [19968, 40869) 中随机选择一个 rune，并返回由该
// rune 组成的 UTF-8 字符串。该区间覆盖大量常用汉字，也可能包含调用方业务上不期望的
// 生僻字，展示前应按业务需要过滤。
//
// ChineseLastName 从内置姓氏文本按 rune 展开的集合中随机选择一个 rune，并返回对应的
// UTF-8 字符串。该集合包含复姓文本，但当前实现不会整体返回“欧阳”“司马”等复姓，而是
// 将其中每个汉字作为独立候选值。
//
// 使用建议与注意事项：
//
//   - random 为 nil 时使用 math/rand 的包级随机源；其默认种子和并发语义遵循当前 Go
//     版本标准库的约定。
//   - 传入自定义 *rand.Rand 时，种子设置、结果可复现性和并发安全均由调用方负责；需要
//     可重复结果时可使用固定种子，需要并发共享时应自行加锁或为 goroutine 分配独立实例。
//   - 频繁调用时建议复用随机源，避免在热点路径反复创建 *rand.Rand 或 Source。
//   - 范围随机的结果始终包含 min 且不包含 max，业务上需要闭区间时应由调用方调整上界。
package rand
