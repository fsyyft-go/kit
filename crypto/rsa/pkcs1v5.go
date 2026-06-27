// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package rsa

import (
	"crypto"
	"crypto/rsa"
	"errors"
	"fmt"
	"math/big"
)

// 下面的代码来自 SDK 的：crypt/rsa/pkcs1v5.go 。

// hashPrefixes 是一个映射表，存储了各种哈希算法对应的 ASN.1 DER 编码前缀。
// 这些前缀在 PKCS#1 v1.5 签名验证过程中用于标识使用的哈希算法。
var (
	hashPrefixes = map[crypto.Hash][]byte{
		crypto.MD5:       {0x30, 0x20, 0x30, 0x0c, 0x06, 0x08, 0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x02, 0x05, 0x05, 0x00, 0x04, 0x10},
		crypto.SHA1:      {0x30, 0x21, 0x30, 0x09, 0x06, 0x05, 0x2b, 0x0e, 0x03, 0x02, 0x1a, 0x05, 0x00, 0x04, 0x14},
		crypto.SHA224:    {0x30, 0x2d, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x04, 0x05, 0x00, 0x04, 0x1c},
		crypto.SHA256:    {0x30, 0x31, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x01, 0x05, 0x00, 0x04, 0x20},
		crypto.SHA384:    {0x30, 0x41, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x02, 0x05, 0x00, 0x04, 0x30},
		crypto.SHA512:    {0x30, 0x51, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x03, 0x05, 0x00, 0x04, 0x40},
		crypto.MD5SHA1:   {},
		crypto.RIPEMD160: {0x30, 0x20, 0x30, 0x08, 0x06, 0x06, 0x28, 0xcf, 0x06, 0x03, 0x00, 0x31, 0x04, 0x14},
	}
)

// publicDecrypt 执行与历史私钥原始操作配套的 RSA 公钥恢复流程。
//
// 当 hash 非 0 且输入满足本函数的前置约束时，返回值是恢复出的 DigestInfo；当 hash 为 0 时，
// 返回值是恢复出的原始消息块。本函数只使用 hashed 校验摘要长度，不会验证恢复内容是否等于 prefix+hashed。
//
// 参数：
//   - pub: RSA 公钥，必须非 nil 且包含有效模数和指数。
//   - hash: 使用的摘要算法；必须是 hash 可选值列表中的值之一。
//   - hashed: 待匹配的摘要数据；hash 非 0 时长度必须与摘要算法匹配。
//   - sig: 待恢复的签名数据。
//
// hash 可选值：
//   - crypto.Hash(0): 按原始消息块恢复，不附加 ASN.1 DER 前缀。
//   - crypto.MD5: 使用 MD5 DigestInfo 前缀。
//   - crypto.SHA1: 使用 SHA-1 DigestInfo 前缀。
//   - crypto.SHA224: 使用 SHA-224 DigestInfo 前缀。
//   - crypto.SHA256: 使用 SHA-256 DigestInfo 前缀。
//   - crypto.SHA384: 使用 SHA-384 DigestInfo 前缀。
//   - crypto.SHA512: 使用 SHA-512 DigestInfo 前缀。
//   - crypto.MD5SHA1: 使用无前缀的 MD5+SHA-1 摘要兼容形式。
//   - crypto.RIPEMD160: 使用 RIPEMD-160 DigestInfo 前缀。
//
// 返回：
//   - []byte: hash 非 0 时为恢复出的 DigestInfo；hash 为 0 时为恢复出的原始消息块。
//   - error: 输入摘要长度、哈希类型或密钥长度不满足要求时返回错误。
func publicDecrypt(pub *rsa.PublicKey, hash crypto.Hash, hashed []byte, sig []byte) (out []byte, err error) {
	// 获取哈希算法的相关信息，包括哈希长度和前缀。
	hashLen, prefix, err := pkcs1v15HashInfo(hash, len(hashed))
	if err != nil {
		return nil, err
	}

	// 计算 T 长度（前缀长度 + 哈希长度）。
	tLen := len(prefix) + hashLen
	// 计算密钥长度（以字节为单位）。
	k := (pub.N.BitLen() + 7) / 8
	// 检查密钥长度是否足够（PKCS#1 v1.5 要求至少有 11 字节的开销）。
	if k < tLen+11 {
		return nil, fmt.Errorf("length illegal")
	}

	// 将签名转换为大整数。
	c := new(big.Int).SetBytes(sig)
	// 使用公钥进行加密操作（在 RSA 中，验证签名实际上是用公钥加密签名数据）。
	m := encrypt(new(big.Int), pub, c)
	// 对结果进行左填充，确保长度正确。
	em := leftPad(m.Bytes(), k)
	// 移除填充，获取原始数据。
	out = unLeftPad(em)

	err = nil
	return
}

// unLeftPad 按历史兼容规则移除 PKCS#1 v1.5 风格左侧填充。
//
// 参数：
//   - input: 带左侧填充的编码块；调用方应保证长度足以包含前导标记和填充长度信息。
//
// 返回：
//   - []byte: 移除填充后得到的 payload；无法识别连续 0xff 填充时按历史长度字节规则截取。
func unLeftPad(input []byte) (out []byte) {
	// 获取输入数据长度。
	n := len(input)
	// 初始填充长度为 2。
	t := 2
	// 从第三个字节开始遍历。
	for i := 2; i < n; i++ {
		// 如果当前字节是 0xff，说明是填充字节，增加填充计数。
		if input[i] == 0xff {
			t = t + 1
		} else {
			// 如果当前字节等于第一个字节，则填充字节数等于第二个字节的值。
			if input[i] == input[0] {
				t = t + int(input[1])
			}
			// 跳出循环，找到了填充结束的位置。
			break
		}
	}
	// 创建用于存储结果的切片，大小为原始数据长度减去填充长度。
	out = make([]byte, n-t)
	// 复制原始数据到结果切片中。
	copy(out, input[t:])
	return
}

// leftPad 对输入数据进行左侧零填充，使其达到指定大小。
//
// 参数：
//   - input: 原始输入数据。
//   - size: 期望的输出大小，必须非负；小于 input 长度时按当前实现保留 input 前缀字节。
//
// 返回：
//   - []byte: 长度为 size 的填充结果；当 input 短于 size 时，结果左侧以 0 填充。
func leftPad(input []byte, size int) (out []byte) {
	// 获取输入数据长度。
	n := len(input)
	// 如果输入长度大于目标大小，则截断为目标大小。
	// 当前实现通过后续 copy(out[len(out)-n:], input) 保留 input 的前缀字节。
	if n > size {
		n = size
	}
	// 创建指定大小的切片用于存储结果。
	out = make([]byte, size)
	// 将输入数据复制到结果切片的末尾，前面部分默认为 0，实现左填充。
	copy(out[len(out)-n:], input)
	return
}

// encrypt 执行基本的 RSA 公钥指数运算。
//
// 参数：
//   - c: 用于存储计算结果的大整数，必须非 nil。
//   - pub: RSA 公钥，必须非 nil 且包含有效模数和指数。
//   - m: 待处理的消息整数，必须非 nil。
//
// 返回：
//   - *big.Int: 计算得到的 m^e mod N；返回值与 c 指向同一对象。
func encrypt(c *big.Int, pub *rsa.PublicKey, m *big.Int) *big.Int {
	// 创建表示公钥指数的大整数。
	e := big.NewInt(int64(pub.E))
	// 计算 m^e mod N，即 RSA 加密的核心操作。
	c.Exp(m, e, pub.N)
	return c
}

// pkcs1v15HashInfo 返回 PKCS#1 v1.5 签名编码所需的哈希长度和 ASN.1 DER 前缀。
//
// 参数：
//   - hash: 哈希算法；必须是 hash 可选值列表中的值之一。
//   - inLen: 输入摘要长度；hash 非 0 时必须等于 hash.Size()。
//
// hash 可选值：
//   - crypto.Hash(0): 表示输入是原始消息块，不使用 ASN.1 DER 前缀。
//   - crypto.MD5: 返回 MD5 摘要长度和 DigestInfo 前缀。
//   - crypto.SHA1: 返回 SHA-1 摘要长度和 DigestInfo 前缀。
//   - crypto.SHA224: 返回 SHA-224 摘要长度和 DigestInfo 前缀。
//   - crypto.SHA256: 返回 SHA-256 摘要长度和 DigestInfo 前缀。
//   - crypto.SHA384: 返回 SHA-384 摘要长度和 DigestInfo 前缀。
//   - crypto.SHA512: 返回 SHA-512 摘要长度和 DigestInfo 前缀。
//   - crypto.MD5SHA1: 返回 MD5+SHA-1 组合摘要长度和空前缀。
//   - crypto.RIPEMD160: 返回 RIPEMD-160 摘要长度和 DigestInfo 前缀。
//
// 返回：
//   - hashLen: hash 为 0 时等于 inLen；否则等于 hash.Size()。
//   - prefix: hash 对应的 ASN.1 DER 前缀；hash 为 0 或 crypto.MD5SHA1 时为空。
//   - error: 输入长度与 hash 不匹配或 hash 不在支持列表中时返回错误。
func pkcs1v15HashInfo(hash crypto.Hash, inLen int) (hashLen int, prefix []byte, err error) {
	// 如果没有指定哈希算法，则直接使用输入长度作为哈希长度，不使用前缀。
	if hash == 0 {
		return inLen, nil, nil
	}

	// 获取指定哈希算法的输出长度。
	hashLen = hash.Size()
	// 验证输入长度是否匹配哈希长度。
	if inLen != hashLen {
		return 0, nil, errors.New("crypto/rsa: input must be hashed message")
	}
	// 从预定义映射中获取对应的前缀。
	prefix, ok := hashPrefixes[hash]
	if !ok {
		return 0, nil, errors.New("crypto/rsa: unsupported hash function")
	}
	return
}
