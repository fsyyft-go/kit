// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package crypto 汇总本项目中与密码学和摘要算法相关的子包。
//
// 本包不提供根级别的加密、哈希或一次性密码 API，主要用于在 Go 文档中
// 说明 crypto 目录的组织方式。具体能力由下级子包提供，调用方应直接导入
// 所需子包，例如 aes、des、rsa、md5、sha 或 otp 相关实现。
//
// 使用这些子包时，调用方需要结合各子包文档处理密钥来源、随机数、密文
// 编码、错误返回和兼容性要求。涉及新业务安全设计时，应优先选择当前
// 推荐的算法和足够强度的密钥，不应将过时摘要算法用于密码存储或认证。
package crypto
