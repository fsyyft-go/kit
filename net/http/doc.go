// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package http 提供可配置的 HTTP client、请求 Hook，以及 HTTPS 证书辅助函数。
//
// NewClient 基于标准库 http.Client 组装超时、连接池、代理和日志/trace Hook。
// 当未通过 WithTransport 显式提供自定义 Transport 时，默认 Transport 会将
// TLSClientConfig.InsecureSkipVerify 设为 true，也就是默认跳过 TLS 证书校验；
// 如需启用证书校验，调用方需要通过 WithTransport 显式调整 TLS 配置。
// GetCertificates 与 GetCertificatesExpirestime 用于发起 HTTPS 请求并提取对端证书链及剩余有效期。
package http
