// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package log 提供统一的 Logger 接口、全局默认日志器访问，以及基于标准库和 Logrus 的实现。
//
// Logger 统一封装 Debug、Info、Warn、Error、Fatal、级别控制和结构化字段追加。
// InitLogger 配置包级默认日志器，Debug、Info、Warn、Error、Fatal 等全局函数都委托到该实例；
// 若调用方未显式初始化，首次调用 GetLogger 或全局函数时会惰性创建一个输出到标准输出的 Std logger。
// NewLogger 用于创建独立日志器，可通过 Option 选择 Std、Console 或 Logrus 实现，
// 并配置级别、输出路径、输出格式和日志轮转。JSONFormat 与 TextFormat 仅影响 Logrus 格式化；
// 当前 Logger 接口不提供 Close 方法，调用方也无法显式关闭文件型实现。
package log
