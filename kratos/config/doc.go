// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package config 提供 Kratos 配置项到 map[string]any 的解码与后处理能力。
//
// NewDecoder 返回可选 Resolve 钩子的解码器。Decode 在 src.Format 为空时会把点分 key
// 展开为嵌套 map；在 src.Format 非空时委托 Kratos codec 解码到 map[string]any。
// RegisterResolve 用于扩展包级默认解析器，当前内置 .b64、.des 和 .env 后缀处理。
// 包级解析器注册会修改全局 map，应在程序初始化阶段或并发解码开始前完成。
package config
