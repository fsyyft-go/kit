// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package otp 提供了一次性密码（One-Time Password）的实现，支持 HOTP（HMAC-based OTP）和 TOTP（Time-based OTP）。

主要功能：

1. HOTP 实现：
  - 基于 HMAC-SHA1 算法
  - 支持自定义计数器
  - 可配置密码长度
  - 验证窗口设置

2. TOTP 实现：
  - 基于时间戳
  - 支持自定义时间步长
  - 兼容 Google Authenticator
  - 时间偏差处理

3. 密钥管理：
  - 密钥生成
  - Base32 编码
  - 密钥验证
  - 安全存储建议

基本用法：

1. HOTP 生成和验证：

	// 生成 HOTP 密码
	code, err := otp.GenerateHOTP(key, counter, 6)
	if err != nil {
	    // 处理错误
	}

	// 验证 HOTP 密码
	valid := otp.ValidateHOTP(code, key, counter, 6)

2. TOTP 生成和验证：

	// 生成 TOTP 密码
	code, err := otp.GenerateTOTP(key, time.Now(), 30)
	if err != nil {
	    // 处理错误
	}

	// 验证 TOTP 密码
	valid := otp.ValidateTOTP(code, key, time.Now(), 30)

安全特性：

1. 密钥保护：
  - 安全的密钥生成
  - 密钥格式验证
  - 密钥长度检查

2. 验证机制：
  - 防重放攻击
  - 时间窗口控制
  - 计数器同步

3. 错误处理：
  - 安全的错误信息
  - 完整的错误检查
  - panic 保护机制

性能优化：

1. 内存使用：
  - 避免不必要的内存分配
  - 使用适当的缓冲区大小
  - 及时释放资源

2. 计算优化：
  - 高效的 HMAC 计算
  - 优化的时间处理
  - 减少内存拷贝

注意事项：

1. 密钥管理：
  - 安全生成密钥
  - 安全传输密钥
  - 安全存储密钥

2. 时间同步：
  - 服务器时间同步
  - 客户端时间校准
  - 处理时间偏差

3. 最佳实践：
  - 使用足够长的密码
  - 合理设置验证窗口
  - 注意并发安全
*/
package otp
