// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package otp 提供了基于时间的一次性密码（TOTP）算法的实现，支持各种常见哈希算法和自定义选项。
package otp

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"hash"
	"net/url"
	"strconv"
	"strings"
	"time"
)

/**
 * ========== ========== ========== ========== ==========
 * 定义选项接口和方法开始。
 */

var (
	// 空赋值确保 OneTimePasswordOptionFunc 类型实现了 OneTimePasswordOption 接口。
	_ OneTimePasswordOption = (OneTimePasswordOptionFunc)(nil)
)

type (
	// OneTimePasswordOption OTP 算法实例化时需要的选项。
	OneTimePasswordOption interface {
		// apply 将选项应用于 OTP 算法实例。
		apply(*oneTimePassword)
	}
	// OneTimePasswordOptionFunc OTP 算法实例化时需要的选项的函数表示形式，实现了 OneTimePasswordOption 接口。
	OneTimePasswordOptionFunc func(*oneTimePassword)
)

// apply 将选项应用于 OTP 算法。
//
// 参数：
//   - password：要应用选项的 oneTimePassword 实例。
func (o OneTimePasswordOptionFunc) apply(password *oneTimePassword) {
	// 调用函数本身，将 oneTimePassword 实例传递给函数。
	o(password)
}

/**
 * 定义选项接口和方法结束。
 * ========== ========== ========== ========== ==========
 */

var (
	// defaultHashCipher 默认的哈希算法名称为 SHA1。
	defaultHashCipher = "SHA1"
	// defaultHashFunc 默认的哈希函数为 SHA1。
	defaultHashFunc = sha1.New
	// defaultDigits 默认的密码长度为 6 位。
	defaultDigits = 6
	// defaultPeriodSeconds 默认的密码有效期为 30 秒。
	defaultPeriodSeconds = 30
	// defaultWindowSize 默认的时间窗口大小为 10。
	defaultWindowSize = 10
)

/**
 * ========== ========== ========== ========== ==========
 * 导入选项方法开始。
 */

// WithSHA256 返回哈希算法为 SHA256 选项。
func WithSHA256() OneTimePasswordOption {
	// 定义一个函数，用于设置 oneTimePassword 实例的哈希算法相关属性。
	f := func(password *oneTimePassword) {
		// 设置哈希算法名称为 SHA256。
		password.hashCipher = "SHA256"
		// 设置哈希函数为 SHA256。
		password.hashFunc = sha256.New
	}

	// 将函数转换为 OneTimePasswordOptionFunc 类型并返回。
	return (OneTimePasswordOptionFunc)(f)
}

// WithSHA512 返回哈希算法为 SHA512 选项。
func WithSHA512() OneTimePasswordOption {
	// 定义一个函数，用于设置 oneTimePassword 实例的哈希算法相关属性。
	f := func(password *oneTimePassword) {
		// 设置哈希算法名称为 SHA512。
		password.hashCipher = "SHA512"
		// 设置哈希函数为 SHA512。
		password.hashFunc = sha512.New
	}

	// 将函数转换为 OneTimePasswordOptionFunc 类型并返回。
	return (OneTimePasswordOptionFunc)(f)
}

// WithDigits 返回密码长度选项。
func WithDigits(digits int) OneTimePasswordOption {
	// 定义一个函数，用于设置 oneTimePassword 实例的密码长度。
	f := func(password *oneTimePassword) {
		// 设置密码长度。
		password.digits = digits
	}

	// 将函数转换为 OneTimePasswordOptionFunc 类型并返回。
	return (OneTimePasswordOptionFunc)(f)
}

// WithPeriodSeconds 返回密码的有效期（单位为秒）选项。
func WithPeriodSeconds(periodSeconds int) OneTimePasswordOption {
	// 定义一个函数，用于设置 oneTimePassword 实例的密码有效期。
	f := func(password *oneTimePassword) {
		// 设置密码有效期，单位为秒。
		password.periodSeconds = periodSeconds
	}

	// 将函数转换为 OneTimePasswordOptionFunc 类型并返回。
	return (OneTimePasswordOptionFunc)(f)
}

// WithWindowSize 返回时间窗口选项。
func WithWindowSize(windowSize int) OneTimePasswordOption {
	// 定义一个函数，用于设置 oneTimePassword 实例的时间窗口大小。
	f := func(password *oneTimePassword) {
		// 设置时间窗口大小。
		password.windowSize = windowSize
	}

	// 将函数转换为 OneTimePasswordOptionFunc 类型并返回。
	return (OneTimePasswordOptionFunc)(f)
}

// WithIssuer 返回发行者选项。
func WithIssuer(issuer string) OneTimePasswordOption {
	// 定义一个函数，用于设置 oneTimePassword 实例的发行者。
	f := func(password *oneTimePassword) {
		// 设置发行者。
		password.issuer = issuer
	}

	// 将函数转换为 OneTimePasswordOptionFunc 类型并返回。
	return (OneTimePasswordOptionFunc)(f)
}

// WithLabel 返回标签选项。
func WithLabel(label string) OneTimePasswordOption {
	// 定义一个函数，用于设置 oneTimePassword 实例的标签。
	f := func(password *oneTimePassword) {
		// 设置标签。
		password.label = label
	}

	// 将函数转换为 OneTimePasswordOptionFunc 类型并返回。
	return (OneTimePasswordOptionFunc)(f)
}

/**
 * 导入选项方法结束。
 * ========== ========== ========== ========== ==========
 */

var (
	// base32Encoded 定义一个无填充的 Base32 编码器。
	base32Encoded = base32.StdEncoding.WithPadding(base32.NoPadding)
)

var (
	// 空赋值确保 oneTimePassword 类型实现了 OneTimePassword 接口。
	_ OneTimePassword = (*oneTimePassword)(nil)
)

type (
	// OneTimePassword 定义了一次性密码的接口。
	OneTimePassword interface {
		// Password 根据当前时间生成密码。
		//
		// 返回：
		//   - string：生成的一次性密码字符串。
		//   - error：如果生成过程中出现错误，则返回相应的错误。
		Password() (string, error)

		// EffectivePassword 根据当前时间，生成指定时间窗口内的所有密码。
		//
		// 返回：
		//   - []string：时间窗口内的所有有效密码字符串切片。
		//   - error：如果生成过程中出现错误，则返回相应的错误。
		EffectivePassword() ([]string, error)

		// VeryfyPassword 验证密码是否在指定时间窗口内。
		//
		// 参数：
		//   - password：需要验证的密码字符串。
		//
		// 返回：
		//   - bool：如果密码有效，则返回 true，否则返回 false。
		VeryfyPassword(password string) bool

		// GenerateURL 生成对应的 URL 表示形式的字符串。
		//
		// 返回：
		//   - string：生成的 URL 字符串，可用于设置二维码等。
		GenerateURL() string
	}

	// oneTimePassword 是 OneTimePassword 接口的具体实现。
	oneTimePassword struct {
		secretKeyBase32 string           // 密钥种子的 Base32 位表示形式。
		secretKey       []byte           // 密钥种子。
		hashCipher      string           // 哈希算法名称。
		hashFunc        func() hash.Hash // 哈希算法。
		digits          int              // 密码长度。
		periodSeconds   int              // 密码的有效期（单位为秒）。
		windowSize      int              // 时间窗口。

		issuer string // 发行者。
		label  string // 标签。
	}
)

// Password 根据当前时间生成密码。
func (o *oneTimePassword) Password() (string, error) {
	// 定义返回值。
	var passwordString string
	var err error

	// 生成基于时间的一次性密码。
	if password, errPassword := timeBasedOneTimePassword(o.hashFunc, o.secretKey, o.periodSeconds, o.digits); nil != errPassword {
		// 如果生成过程中出现错误，则返回错误。
		err = errPassword
	} else {
		// 将生成的密码转换为指定长度的字符串。
		passwordString = fmt.Sprintf("%0*d", o.digits, password)
	}

	// 返回生成的密码和可能的错误。
	return passwordString, err
}

// EffectivePassword 根据当前时间，生成指定时间窗口内的所有密码。
func (o *oneTimePassword) EffectivePassword() ([]string, error) {
	// 初始化一个容量为 o.windowSize*2+1 的字符串切片，用于存储生成的密码。
	var passwordStrings = make([]string, 0, o.windowSize*2+1)
	var err error

	// 获取当前时间的 Unix 时间戳。
	seconds := uint64(time.Now().Unix()) // nolint: gosec
	// 计算当前的计数器值。
	counter := seconds / uint64(o.periodSeconds) // nolint: gosec

	// 计算最小计数器值（当前计数器值减去窗口大小）。
	minCounter := counter - uint64(o.windowSize) // nolint: gosec
	// 计算最大计数器值（当前计数器值加上窗口大小）。
	maxCounter := counter + uint64(o.windowSize) // nolint: gosec

	// 遍历从最小计数器值到最大计数器值的范围。
	for tmpCounter := minCounter; tmpCounter < maxCounter; tmpCounter++ {
		// 生成基于 HMAC 的一次性密码。
		if tmpPassword, errPassword := hmacBasedOneTimePassword(o.hashFunc, o.secretKey, tmpCounter, o.digits); nil != errPassword {
			// 如果生成过程中出现错误，则设置错误并中断循环。
			err = errPassword
			break
		} else {
			// 将生成的密码转换为指定长度的字符串。
			passwordString := fmt.Sprintf("%0*d", o.digits, tmpPassword)
			// 将密码添加到结果切片中。
			passwordStrings = append(passwordStrings, passwordString)
		}
	}

	// 返回生成的密码切片和可能的错误。
	return passwordStrings, err
}

// VeryfyPassword 验证密码是否在指定时间窗口内。
//
// 参数：
//   - secretKeyBase32：Base32 编码的密钥种子，用于生成一次性密码。
//   - password：需要验证的密码。
//   - options：可选的配置选项列表，用于自定义 OTP 行为。
//
// 返回：
//   - bool：如果密码有效，则返回 true，否则返回 false。
func (o *oneTimePassword) VeryfyPassword(password string) bool {
	// 定义返回值，默认为 false。
	var resultValue bool

	// 获取当前时间的 Unix 时间戳。
	seconds := uint64(time.Now().Unix()) // nolint: gosec
	// 计算当前的计数器值。
	counter := seconds / uint64(o.periodSeconds) // nolint: gosec

	// 计算最小计数器值（当前计数器值减去窗口大小）。
	minCounter := counter - uint64(o.windowSize) // nolint: gosec
	// 计算最大计数器值（当前计数器值加上窗口大小）。
	maxCounter := counter + uint64(o.windowSize) // nolint: gosec

	// 遍历从最小计数器值到最大计数器值的范围。
	for tmpCounter := minCounter; tmpCounter < maxCounter; tmpCounter++ {
		// 生成基于 HMAC 的一次性密码。
		if tmpPassword, errPassword := hmacBasedOneTimePassword(o.hashFunc, o.secretKey, tmpCounter, o.digits); nil == errPassword {
			// 如果生成成功，将生成的密码转换为指定长度的字符串。
			passwordString := fmt.Sprintf("%0*d", o.digits, tmpPassword)
			// 比较生成的密码与提供的密码是否匹配，比较时忽略大小写。
			if resultValue = strings.EqualFold(passwordString, password); resultValue {
				// 如果匹配，则设置结果为 true 并中断循环。
				break
			}
		}
	}

	// 返回验证结果。
	return resultValue
}

// GenerateURL 生成对应的 URL 表示形式的字符串。
//
// 参数：
//   - secretKeyBase32：Base32 编码的密钥种子，用于生成一次性密码。
//   - options：可选的配置选项列表，用于自定义 OTP 行为。
//
// 返回：
//   - string：生成的 URL 字符串，可用于设置二维码等。
func (o *oneTimePassword) GenerateURL() string {
	// 创建一个字节缓冲区，用于构建 URL。
	buffer := bytes.Buffer{}

	// 添加 URL 前缀。
	buffer.WriteString("otpauth://totp/")

	// 如果存在标签，则将其添加到 URL 中并进行 URL 编码。
	if 0 != len(o.label) {
		buffer.WriteString(url.QueryEscape(o.label))
	}

	// 添加密钥参数。
	buffer.WriteString("?secret=")
	buffer.WriteString(o.secretKeyBase32)
	buffer.WriteString("&")

	// 如果存在发行者，则将其添加到 URL 中并进行 URL 编码。
	if 0 != len(o.issuer) {
		buffer.WriteString("issuer=")
		buffer.WriteString(url.QueryEscape(o.issuer))
		buffer.WriteString("&")
	}

	// 如果哈希算法不是默认值，则将其添加到 URL 中。
	if defaultHashCipher != o.hashCipher {
		buffer.WriteString("algorithm=")
		buffer.WriteString(o.hashCipher)
		buffer.WriteString("&")
	}

	// 如果密码长度不是默认值，则将其添加到 URL 中。
	if defaultDigits != o.digits {
		buffer.WriteString("digits=")
		buffer.WriteString(strconv.Itoa(o.digits))
		buffer.WriteString("&")
	}

	// 如果密码有效期不是默认值，则将其添加到 URL 中。
	if defaultPeriodSeconds != o.periodSeconds {
		buffer.WriteString("period=")
		buffer.WriteString(strconv.Itoa(o.periodSeconds))
		buffer.WriteString("&")
	}

	// 返回生成的 URL 字符串。
	return buffer.String()
}

// NewOneTimePassword 创建一个新的一次性密码实例。
//
// 参数：
//   - secretKeyBase32：Base32 编码的密钥种子，用于生成一次性密码。
//   - options：可选的配置选项列表，用于自定义 OTP 行为。
//
// 返回：
//   - *oneTimePassword：创建的一次性密码实例。
//   - error：如果密钥解码失败，则返回相应的错误。
func NewOneTimePassword(secretKeyBase32 string, options ...OneTimePasswordOption) (*oneTimePassword, error) {
	// 创建一个具有默认值的 oneTimePassword 实例。
	var newOneTimePassword = &oneTimePassword{
		secretKeyBase32: secretKeyBase32,
		hashCipher:      defaultHashCipher,
		hashFunc:        defaultHashFunc,
		digits:          defaultDigits,
		periodSeconds:   defaultPeriodSeconds,
		windowSize:      defaultWindowSize,
	}
	var err error

	// 将 Base32 编码的密钥解码为字节数组。
	newOneTimePassword.secretKey, err = base32Encoded.DecodeString(secretKeyBase32)

	// 如果解码成功且提供了选项，则应用这些选项。
	if nil == err && nil != options && len(options) > 0 {
		for _, option := range options {
			if nil != option {
				option.apply(newOneTimePassword)
			}
		}
	}

	// 返回创建的 oneTimePassword 实例和可能的错误。
	return newOneTimePassword, err
}

// VeryfyPassword 验证密码是否在指定时间窗口内。
//
// 参数：
//   - secretKeyBase32：Base32 编码的密钥种子，用于生成一次性密码。
//   - password：需要验证的密码。
//   - options：可选的配置选项列表，用于自定义 OTP 行为。
//
// 返回：
//   - bool：如果密码有效，则返回 true，否则返回 false。
func VeryfyPassword(secretKeyBase32, password string, options ...OneTimePasswordOption) bool {
	// 定义返回值，默认为 false。
	var resultValue bool

	// 创建一个 oneTimePassword 实例。
	if newOneTimePassword, err := NewOneTimePassword(secretKeyBase32, options...); nil == err {
		// 如果创建成功，则调用实例的 VeryfyPassword 方法进行验证。
		resultValue = newOneTimePassword.VeryfyPassword(password)
	}

	// 返回验证结果。
	return resultValue
}

// GenerateURL 生成对应的 URL 表示形式的字符串。
//
// 参数：
//   - secretKeyBase32：Base32 编码的密钥种子，用于生成一次性密码。
//   - options：可选的配置选项列表，用于自定义 OTP 行为。
//
// 返回：
//   - string：生成的 URL 字符串，可用于设置二维码等。
func GenerateURL(secretKeyBase32 string, options ...OneTimePasswordOption) string {
	// 定义返回值，默认为空字符串。
	var resultValue string

	// 创建一个 oneTimePassword 实例。
	if newOneTimePassword, err := NewOneTimePassword(secretKeyBase32, options...); nil == err {
		// 如果创建成功，则调用实例的 GenerateURL 方法生成 URL。
		resultValue = newOneTimePassword.GenerateURL()
	}

	// 返回生成的 URL 字符串。
	return resultValue
}

// timeBasedOneTimePassword 根据当前时间生成一次性密码。
//
// 参数：
//   - hashFunc：用于生成 HMAC 的哈希函数。
//   - key：密钥种子，用于生成一次性密码。
//   - periodSeconds：密码的有效期（单位为秒）。
//   - digits：密码的长度。
//
// 返回：
//   - int：生成的一次性密码整数值。
//   - error：如果生成过程中出现错误，则返回相应的错误。
func timeBasedOneTimePassword(hashFunc func() hash.Hash, key []byte, periodSeconds int, digits int) (int, error) {
	// 获取当前时间的 Unix 时间戳。
	seconds := uint64(time.Now().Unix()) // nolint: gosec
	// 计算当前的计数器值。
	counter := seconds / uint64(periodSeconds) // nolint: gosec
	// 调用 hmacBasedOneTimePassword 生成密码。
	resultValue, err := hmacBasedOneTimePassword(hashFunc, key, counter, digits)
	// 返回生成的密码和可能的错误。
	return resultValue, err
}

// hmacBasedOneTimePassword 基于 HMAC 算法生成一次性密码。
//
// 参数：
//   - hashFunc：用于生成 HMAC 的哈希函数。
//   - key：密钥种子，用于生成一次性密码。
//   - counter：计数器值，在 TOTP 中基于当前时间计算得出。
//   - digits：密码的长度，范围为 0-8，超出范围会被限制为 8。
//
// 返回：
//   - int：生成的一次性密码整数值。
//   - error：如果生成过程中出现错误，则返回相应的错误。
func hmacBasedOneTimePassword(hashFunc func() hash.Hash, key []byte, counter uint64, digits int) (int, error) {
	// 定义返回值。
	var resultValue int
	var err error

	// 如果未提供哈希函数，则使用默认的 SHA1 哈希函数。
	if nil == hashFunc {
		hashFunc = sha1.New
	}

	// 限制密码长度在 0 到 8 位之间，如果超出范围，则设为 8 位。
	if digits > 8 || digits < 0 {
		// 长度不能超过 8 位。
		digits = 8
	}

	// 创建一个新的 HMAC 对象，使用提供的哈希函数和密钥。
	h := hmac.New(hashFunc, key)
	// 将计数器写入 HMAC 对象。
	if errWrite := binary.Write(h, binary.BigEndian, counter); nil != errWrite {
		// 如果写入过程中出现错误，则返回错误。
		err = errWrite
	} else {
		// 没有发生错误，进入下一步。
		// 计算 HMAC 值。
		sum := h.Sum(nil)
		// 取 Hash 后的最后 4 个 byte。
		offset := sum[len(sum)-1] & 0x0f
		// 从 HMAC 值中提取有效值。
		effectiveValue := binary.BigEndian.Uint32(sum[offset:]) & 0x7FFFFFFF
		// 计算模数，用于截取指定位数的密码。
		effectiveModule := uint32(1)
		for idx := 0; idx < digits; idx++ {
			effectiveModule *= 10
		}
		// 计算最终的密码值。
		resultValue = int(effectiveValue % effectiveModule)
	}

	// 返回生成的密码和可能的错误。
	return resultValue, err
}
