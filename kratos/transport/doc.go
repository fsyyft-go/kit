// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package transport 汇总 Kratos 传输层扩展子包。
//
// 当前实现集中在 http 子包，提供 Kratos HTTP Server 与 Gin Engine 之间的
// 路由桥接能力。本包仅作为传输层相关子包的分类入口，不直接导出服务构造函数。
// 具体路由提取、路径转换和 unsafe 访问 Kratos 内部结构的约束在 http 子包中说明。
package transport
