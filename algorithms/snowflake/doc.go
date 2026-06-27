// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package snowflake 提供基于 Snowflake 位布局的分布式唯一 ID 生成能力。
//
// NewNode 使用当前的 Epoch、NodeBits 和 StepBits 创建生成节点。调用方需要为每个
// 节点分配唯一的 nodeid，并在调整 Epoch 或位宽配置后重新创建节点，确保生成与解析
// 使用同一组位布局。
//
// 生成的 ID 支持十进制、Base2、z-base-32、Base36、Base58、Base64 以及字节表示，
// 并可通过对应的 Parse* 函数恢复。Time、Node 和 Step 等字段提取方法保留用于兼容旧
// 版本，它们依赖当前的全局位宽配置。
package snowflake
