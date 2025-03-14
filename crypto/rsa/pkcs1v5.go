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

// publicDecrypt 使用 RSA 公钥对签名进行解密，通常用于验证签名。
// 参数：
// - pub: RSA 公钥。
// - hash: 使用的哈希算法。
// - hashed: 已哈希的消息。
// - sig: 签名数据。
// 返回值：
// - out: 解密后的数据。
// - err: 错误信息。
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

// unLeftPad 移除左侧填充，恢复原始数据。
// 参数：
// - input: 带有左侧填充的数据。
// 返回值：
// - out: 移除填充后的原始数据。
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

// leftPad 对输入数据进行左侧填充，使其达到指定大小。
// 参数：
// - input: 原始输入数据。
// - size: 期望的填充后大小。
// 返回值：
// - out: 填充后的数据。
func leftPad(input []byte, size int) (out []byte) {
	// 获取输入数据长度。
	n := len(input)
	// 如果输入长度大于目标大小，则截断。
	if n > size {
		n = size
	}
	// 创建指定大小的切片用于存储结果。
	out = make([]byte, size)
	// 将输入数据复制到结果切片的末尾，前面部分默认为 0，实现左填充。
	copy(out[len(out)-n:], input)
	return
}

// encrypt 执行基本的 RSA 加密操作。
// 参数：
// - c: 用于存储计算结果的大整数。
// - pub: RSA 公钥。
// - m: 待加密的消息（大整数形式）。
// 返回值：
// - 加密后的结果（大整数）。
func encrypt(c *big.Int, pub *rsa.PublicKey, m *big.Int) *big.Int {
	// 创建表示公钥指数的大整数。
	e := big.NewInt(int64(pub.E))
	// 计算 m^e mod N，即 RSA 加密的核心操作。
	c.Exp(m, e, pub.N)
	return c
}

// pkcs1v15HashInfo 获取 PKCS#1 v1.5 中特定哈希算法的信息。
// 参数：
// - hash: 哈希算法类型。
// - inLen: 输入数据的长度。
// 返回值：
// - hashLen: 哈希值长度。
// - prefix: 对应的 ASN.1 DER 编码前缀。
// - err: 错误信息。
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
