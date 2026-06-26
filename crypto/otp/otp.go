// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

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
	// OneTimePasswordOption 定义 NewOneTimePassword 可接收的配置选项。
	//
	// OneTimePasswordOption 包含未导出的 apply 方法，调用方通常应通过
	// WithSHA256、WithSHA512、WithDigits、WithPeriodSeconds、WithWindowSize、
	// WithIssuer 或 WithLabel 创建选项，而不是在包外自行实现。NewOneTimePassword
	// 只会跳过值为 nil 的接口选项；接口值非 nil 的实现都会被调用。
	OneTimePasswordOption interface {
		// apply 将选项应用于 OTP 实例。
		//
		// 参数：
		//   - *oneTimePassword: 待修改的 OTP 实例；实现方应只调整自身负责的配置字段。
		apply(*oneTimePassword)
	}
	// OneTimePasswordOptionFunc 将函数适配为 OneTimePasswordOption。
	//
	// 函数值为 nil 时仍可能作为非 nil 接口值传入；NewOneTimePassword 不会跳过
	// 这种 typed nil 选项，应用时会在 apply 中 panic。
	//
	// 参数：
	//   - *oneTimePassword: 待修改的 OTP 实例；函数应在 NewOneTimePassword 应用选项期间同步执行。
	OneTimePasswordOptionFunc func(*oneTimePassword)
)

// apply 执行函数形式的 OTP 配置选项。
//
// 若接收者是 nil 函数值，调用 apply 会 panic；NewOneTimePassword 不会识别
// 包装在非 nil 接口中的 typed nil OneTimePasswordOptionFunc。
//
// 参数：
//   - password: 待修改的 OTP 实例；调用方应保证其非 nil。
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
	// defaultDigits 默认的原始密码位数配置为 6 位。
	defaultDigits = 6
	// defaultPeriodSeconds 默认的密码有效期为 30 秒。
	defaultPeriodSeconds = 30
	// defaultWindowSize 默认的验证窗口半径为 10。
	defaultWindowSize = 10
)

/**
 * ========== ========== ========== ========== ==========
 * 导入选项方法开始。
 */

// WithSHA256 返回将 TOTP HMAC 哈希算法设置为 SHA256 的选项。
//
// 参数：无。
//
// 返回：
//   - OneTimePasswordOption: 应用于 NewOneTimePassword 的选项；生成 otpauth URL 时会输出 algorithm=SHA256。
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

// WithSHA512 返回将 TOTP HMAC 哈希算法设置为 SHA512 的选项。
//
// 参数：无。
//
// 返回：
//   - OneTimePasswordOption: 应用于 NewOneTimePassword 的选项；生成 otpauth URL 时会输出 algorithm=SHA512。
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

// WithDigits 返回设置一次性密码位数配置的选项。
//
// 参数：
//   - digits: 写入实例的原始位数配置。hmacBasedOneTimePassword 只在计算整数口令时
//     把小于 0 或大于 8 的值局部重置为 8；Password、EffectivePassword 和 VeryfyPassword
//     仍使用原始配置作为 fmt 格式化宽度，GenerateURL 也会按原始配置写入 digits 参数。
//
// 返回：
//   - OneTimePasswordOption: 应用于 NewOneTimePassword 的选项。
func WithDigits(digits int) OneTimePasswordOption {
	// 定义一个函数，用于设置 oneTimePassword 实例的密码长度。
	f := func(password *oneTimePassword) {
		// 设置密码长度。
		password.digits = digits
	}

	// 将函数转换为 OneTimePasswordOptionFunc 类型并返回。
	return (OneTimePasswordOptionFunc)(f)
}

// WithPeriodSeconds 返回设置 TOTP 时间步长的选项。
//
// 参数：
//   - periodSeconds: 时间步长，单位为秒，调用方应传入大于 0 的值。当前实现不会在设置时校验该值；
//     传入 0 会在后续生成或校验口令时因除零而 panic，传入负数会在转换为 uint64 后产生异常计数器语义。
//
// 返回：
//   - OneTimePasswordOption: 应用于 NewOneTimePassword 的选项；非默认值会在 otpauth URL 中输出 period 参数。
func WithPeriodSeconds(periodSeconds int) OneTimePasswordOption {
	// 定义一个函数，用于设置 oneTimePassword 实例的密码有效期。
	f := func(password *oneTimePassword) {
		// 设置密码有效期，单位为秒。
		password.periodSeconds = periodSeconds
	}

	// 将函数转换为 OneTimePasswordOptionFunc 类型并返回。
	return (OneTimePasswordOptionFunc)(f)
}

// WithWindowSize 返回设置 TOTP 验证窗口大小的选项。
//
// 参数：
//   - windowSize: 写入实例的窗口半径配置，实际遍历区间为 [counter-windowSize, counter+windowSize)。
//     0 会使 EffectivePassword 返回空列表且 VeryfyPassword 不接受当前口令；负数未被校验，
//     会在 make 容量计算或转换为 uint64 后产生异常或不可用行为，调用方应避免传入。
//
// 返回：
//   - OneTimePasswordOption: 应用于 NewOneTimePassword 的选项。
func WithWindowSize(windowSize int) OneTimePasswordOption {
	// 定义一个函数，用于设置 oneTimePassword 实例的时间窗口大小。
	f := func(password *oneTimePassword) {
		// 设置时间窗口大小。
		password.windowSize = windowSize
	}

	// 将函数转换为 OneTimePasswordOptionFunc 类型并返回。
	return (OneTimePasswordOptionFunc)(f)
}

// WithIssuer 返回设置 otpauth URL 发行者的选项。
//
// 参数：
//   - issuer: 发行者名称。空字符串表示不在 GenerateURL 结果中写入 issuer 参数；非空值会按 URL 查询参数规则转义。
//
// 返回：
//   - OneTimePasswordOption: 应用于 NewOneTimePassword 的选项。
func WithIssuer(issuer string) OneTimePasswordOption {
	// 定义一个函数，用于设置 oneTimePassword 实例的发行者。
	f := func(password *oneTimePassword) {
		// 设置发行者。
		password.issuer = issuer
	}

	// 将函数转换为 OneTimePasswordOptionFunc 类型并返回。
	return (OneTimePasswordOptionFunc)(f)
}

// WithLabel 返回设置 otpauth URL 标签的选项。
//
// 参数：
//   - label: TOTP 账户标签。空字符串表示 URL 路径中不附加标签；非空值会按 URL 查询转义规则写入 otpauth://totp/ 后的路径段。
//
// 返回：
//   - OneTimePasswordOption: 应用于 NewOneTimePassword 的选项。
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
	// OneTimePassword 定义基于当前时间生成和验证 TOTP 口令的能力。
	//
	// OneTimePassword 实例在 NewOneTimePassword 返回后不维护可变运行状态，可复用执行
	// Password、EffectivePassword、VeryfyPassword 和 GenerateURL。当前接口不包含密钥生成、
	// HOTP 计数器管理或重放检测能力。
	OneTimePassword interface {
		// Password 根据当前时间生成当前时间步的一次性密码。
		//
		// 参数：无。
		//
		// 返回：
		//   - string: 按原始 digits 配置作为宽度左侧补零后的数字口令；生成失败时为空字符串。
		//   - error: 底层 HOTP 生成失败时返回错误。periodSeconds 为 0 会在计算时间步时 panic；为负数会转换为 uint64 并产生异常计数器语义。
		Password() (string, error)

		// EffectivePassword 根据当前时间生成验证窗口内可接受的一次性密码。
		//
		// 参数：无。
		//
		// 返回：
		//   - []string: 按 [counter-windowSize, counter+windowSize) 半开区间生成的口令，并按原始 digits 配置格式化；windowSize 为 0 时返回空切片。
		//   - error: 底层 HOTP 生成失败时返回错误。periodSeconds 为 0 会在计算时间步时 panic；periodSeconds 为负数或 windowSize 为负数会产生异常或不可用行为。
		EffectivePassword() ([]string, error)

		// VeryfyPassword 验证密码是否落在当前配置的时间窗口内。
		//
		// 参数：
		//   - password: 待验证的口令字符串，按当前实现与窗口内生成值进行大小写不敏感比较。
		//
		// 返回：
		//   - bool: 任一 [counter-windowSize, counter+windowSize) 半开区间内、按原始 digits 配置格式化的口令匹配时返回 true；windowSize 为 0、没有匹配值或底层 HOTP 生成失败时返回 false。periodSeconds 为 0 会在计算时间步时 panic；periodSeconds 为负数或 windowSize 为负数会产生异常或不可用行为。
		VeryfyPassword(password string) bool

		// GenerateURL 生成 otpauth://totp/ URL 字符串。
		//
		// 参数：无。
		//
		// 返回：
		//   - string: 包含 secret 参数以及可选 issuer、algorithm、digits 和 period 参数的 URL，可用于生成二维码；digits 和 period 按原始配置值输出，该方法不会校验边界。
		GenerateURL() string
	}

	// oneTimePassword 是 OneTimePassword 接口的具体实现。
	oneTimePassword struct {
		secretKeyBase32 string           // 密钥种子的 Base32 位表示形式。
		secretKey       []byte           // 密钥种子。
		hashCipher      string           // 哈希算法名称。
		hashFunc        func() hash.Hash // 哈希算法。
		digits          int              // 原始密码位数配置。
		periodSeconds   int              // TOTP 时间步长（单位为秒）。
		windowSize      int              // 验证窗口半径配置。

		issuer string // 发行者。
		label  string // 标签。
	}
)

// Password 根据当前时间生成当前时间步的一次性密码。
//
// 参数：无。
//
// 返回：
//   - string: 按原始 digits 配置作为宽度左侧补零后的数字口令；生成失败时为空字符串。
//   - error: 底层 HOTP 生成失败时返回错误。periodSeconds 为 0 会在计算时间步时 panic；为负数会转换为 uint64 并产生异常计数器语义。
func (o *oneTimePassword) Password() (string, error) {
	// 定义返回值。
	var passwordString string
	var err error

	// 生成基于时间的一次性密码。
	if password, errPassword := timeBasedOneTimePassword(o.hashFunc, o.secretKey, o.periodSeconds, o.digits); nil != errPassword {
		// 如果生成过程中出现错误，则返回错误。
		err = errPassword
	} else {
		// 将生成的密码按原始 digits 配置宽度转换为字符串。
		passwordString = fmt.Sprintf("%0*d", o.digits, password)
	}

	// 返回生成的密码和可能的错误。
	return passwordString, err
}

// EffectivePassword 根据当前时间生成验证窗口内可接受的一次性密码。
//
// 参数：无。
//
// 返回：
//   - []string: 按 [counter-windowSize, counter+windowSize) 半开区间生成的口令，并按原始 digits 配置格式化；windowSize 为 0 时返回空切片。
//   - error: 底层 HOTP 生成失败时返回错误。periodSeconds 为 0 会在计算时间步时 panic；periodSeconds 为负数或 windowSize 为负数会产生异常或不可用行为。
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
	// 计算窗口上界计数器值（当前计数器值加上窗口大小，遍历时不包含该值）。
	maxCounter := counter + uint64(o.windowSize) // nolint: gosec

	// 遍历从最小计数器值到窗口上界之前的半开范围。
	for tmpCounter := minCounter; tmpCounter < maxCounter; tmpCounter++ {
		// 生成基于 HMAC 的一次性密码。
		if tmpPassword, errPassword := hmacBasedOneTimePassword(o.hashFunc, o.secretKey, tmpCounter, o.digits); nil != errPassword {
			// 如果生成过程中出现错误，则设置错误并中断循环。
			err = errPassword
			break
		} else {
			// 将生成的密码按原始 digits 配置宽度转换为字符串。
			passwordString := fmt.Sprintf("%0*d", o.digits, tmpPassword)
			// 将密码添加到结果切片中。
			passwordStrings = append(passwordStrings, passwordString)
		}
	}

	// 返回生成的密码切片和可能的错误。
	return passwordStrings, err
}

// VeryfyPassword 验证密码是否落在当前配置的时间窗口内。
//
// 参数：
//   - password: 待验证的口令字符串，按当前实现与窗口内生成值进行大小写不敏感比较。
//
// 返回：
//   - bool: 任一 [counter-windowSize, counter+windowSize) 半开区间内、按原始 digits 配置格式化的口令匹配时返回 true；windowSize 为 0、没有匹配值或底层 HOTP 生成失败时返回 false。periodSeconds 为 0 会在计算时间步时 panic；periodSeconds 为负数或 windowSize 为负数会产生异常或不可用行为。
func (o *oneTimePassword) VeryfyPassword(password string) bool {
	// 定义返回值，默认为 false。
	var resultValue bool

	// 获取当前时间的 Unix 时间戳。
	seconds := uint64(time.Now().Unix()) // nolint: gosec
	// 计算当前的计数器值。
	counter := seconds / uint64(o.periodSeconds) // nolint: gosec

	// 计算最小计数器值（当前计数器值减去窗口大小）。
	minCounter := counter - uint64(o.windowSize) // nolint: gosec
	// 计算窗口上界计数器值（当前计数器值加上窗口大小，遍历时不包含该值）。
	maxCounter := counter + uint64(o.windowSize) // nolint: gosec

	// 遍历从最小计数器值到窗口上界之前的半开范围。
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

// GenerateURL 生成 otpauth://totp/ URL 字符串。
//
// 参数：无。
//
// 返回：
//   - string: 包含 secret 参数以及可选 issuer、algorithm、digits 和 period 参数的 URL，可用于生成二维码；digits 和 period 按原始配置值输出，该方法不会校验边界。
func (o *oneTimePassword) GenerateURL() string {
	// 创建一个字节缓冲区，用于构建 URL。
	buffer := bytes.Buffer{}

	// 添加 URL 前缀。
	buffer.WriteString("otpauth://totp/")

	// 如果存在标签，则将其添加到 URL 中并进行 URL 编码。
	if len(o.label) > 0 {
		buffer.WriteString(url.QueryEscape(o.label))
	}

	// 添加密钥参数。
	buffer.WriteString("?secret=")
	buffer.WriteString(o.secretKeyBase32)
	buffer.WriteString("&")

	// 如果存在发行者，则将其添加到 URL 中并进行 URL 编码。
	if len(o.issuer) > 0 {
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

// NewOneTimePassword 创建使用 Base32 密钥的一次性密码实例。
//
// 参数：
//   - secretKeyBase32: 无填充 Base32 编码的密钥种子；空字符串会被当前 Base32 解码器接受，非法编码会阻止后续选项应用。
//   - options: 可选配置项，按传入顺序应用；密钥解码失败时所有选项都不会应用。只有值为 nil 的接口选项会被跳过，typed nil OneTimePasswordOptionFunc 作为非 nil 接口值传入时仍会被调用并在 apply 中 panic。
//
// 返回：
//   - *oneTimePassword: 创建出的一次性密码实例；即使密钥解码失败也会返回带默认配置的实例，但不应继续用于生成或验证口令。
//   - error: secretKeyBase32 解码失败时返回 Base32 解码错误；当前实现不会校验 periodSeconds、digits 或 windowSize 等选项边界。
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

// VeryfyPassword 使用给定配置验证密码是否落在当前时间窗口内。
//
// 参数：
//   - secretKeyBase32: 无填充 Base32 编码的密钥种子；解码失败时直接返回 false。
//   - password: 待验证的口令字符串，按当前实现与窗口内生成值进行大小写不敏感比较。
//   - options: 可选配置项，按传入顺序应用；只有值为 nil 的接口选项会被忽略，typed nil 选项仍会被调用。periodSeconds 为 0 会在验证过程中 panic；periodSeconds 为负数或 windowSize 为负数会产生异常或不可用行为。
//
// 返回：
//   - bool: 任一 [counter-windowSize, counter+windowSize) 半开区间内、按原始 digits 配置格式化的口令匹配时返回 true；密钥解析失败、windowSize 为 0、没有匹配值或底层 HOTP 生成失败时返回 false。
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

// GenerateURL 使用给定配置生成 otpauth://totp/ URL 字符串。
//
// 参数：
//   - secretKeyBase32: 无填充 Base32 编码的密钥种子；解码失败时直接返回空字符串。
//   - options: 可选配置项，按传入顺序应用；只有值为 nil 的接口选项会被忽略，typed nil 选项仍会被调用。该函数不会校验 digits 或 periodSeconds，非默认值会原样写入 digits 和 period 参数。
//
// 返回：
//   - string: 包含 secret 参数以及可选 issuer、algorithm、digits 和 period 参数的 URL；digits 和 period 按原始配置值输出，secretKeyBase32 解码失败时返回空字符串。
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

// timeBasedOneTimePassword 根据当前 Unix 时间生成 TOTP 整数口令。
//
// 参数：
//   - hashFunc: 用于生成 HMAC 的哈希函数；为 nil 时由 hmacBasedOneTimePassword 回退到 SHA1。
//   - key: 已解码的密钥字节，可为空切片。
//   - periodSeconds: 时间步长，单位为秒，调用方应传入大于 0 的值；当前实现会直接用当前时间戳除以该值，传入 0 会 panic，传入负数会转换为 uint64 并产生异常计数器语义。
//   - digits: 口令位数，会传递给 hmacBasedOneTimePassword；小于 0 或大于 8 时仅在 HOTP 计算中局部重置为 8，本函数不会向调用方返回归一化后的位数。
//
// 返回：
//   - int: 当前时间步对应的一次性密码整数值，调用方负责选择格式化宽度；实例方法当前使用原始 digits 配置补零。
//   - error: 底层 HOTP 写入计数器失败时返回错误。
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

// hmacBasedOneTimePassword 基于 HOTP 动态截断规则生成整数口令。
//
// 参数：
//   - hashFunc: 用于生成 HMAC 的哈希函数；为 nil 时使用 SHA1。
//   - key: 已解码的密钥字节，可为空切片。
//   - counter: HOTP 计数器值；在 TOTP 场景中由 Unix 时间戳除以 periodSeconds 得到。
//   - digits: 口令位数，0 到 8 按原值参与取模；小于 0 或大于 8 时仅在本函数内部重置为 8，等于 0 时返回对 1 取模后的结果。
//
// 返回：
//   - int: 动态截断并按本函数内部使用的位数取模后的口令整数；调用方若补零格式化，需自行选择格式化宽度。
//   - error: 将 counter 写入 HMAC 时失败则返回错误。
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
