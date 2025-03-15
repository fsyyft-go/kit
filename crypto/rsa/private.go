// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package rsa

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

const (
	// BlockTypePrivateKey 定义 PEM 块中 RSA 私钥的类型标识符。
	BlockTypePrivateKey = "RSA PRIVATE KEY"
)

var (
	// ErrDecodePrivateKey 表示私钥解码过程中出现错误的错误信息。
	ErrDecodePrivateKey = errors.New("私钥不正确。")
)

// ConvertPrivateKey 将 PEM 格式的私钥数据转换为 RSA 私钥对象。
//
// 参数：
//   - privateKey：PEM 格式的私钥字节数组。
//
// 返回值：
//   - *rsa.PrivateKey：转换后的 RSA 私钥对象。
//   - error：转换过程中可能发生的错误。
func ConvertPrivateKey(privateKey []byte) (*rsa.PrivateKey, error) {
	// 声明私钥和错误变量。
	var priv *rsa.PrivateKey
	var err error

	// 尝试解码 PEM 格式的私钥数据。
	// 如果解码失败或者块类型不是 RSA PRIVATE KEY，则返回私钥不正确的错误。
	if block, _ := pem.Decode(privateKey); nil == block || block.Type != BlockTypePrivateKey {
		err = ErrDecodePrivateKey
	} else if private, errPrivate := x509.ParsePKCS1PrivateKey(block.Bytes); nil != errPrivate {
		// 尝试将解码后的数据解析为 PKCS1 格式的 RSA 私钥。
		// 如果解析失败，则返回解析过程中的错误。
		err = errPrivate
	} else {
		// 解析成功，将结果赋值给 priv 变量。
		priv = private
	}

	// 返回转换后的私钥对象和可能的错误。
	return priv, err
}
