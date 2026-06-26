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
	// BlockTypePrivateKey 是本包接受的 PKCS#1 RSA 私钥 PEM block 类型。
	BlockTypePrivateKey = "RSA PRIVATE KEY"
)

var (
	// ErrDecodePrivateKey 表示 PEM 私钥解码失败或 block 类型不是 RSA 私钥。
	//
	// 当输入不是 PEM，或 PEM block type 不是 RSA PRIVATE KEY 时返回该错误。
	// PEM type 正确但 DER 内容非法时会透传 x509 解析错误。调用方可以使用 errors.Is 判断该错误。
	ErrDecodePrivateKey = errors.New("私钥不正确。")
)

// ConvertPrivateKey 将 PEM 编码的 PKCS#1 RSA 私钥解析为 *rsa.PrivateKey。
//
// 输入必须是类型为 RSA PRIVATE KEY 的 PEM block；其他 PEM 类型或解析失败时返回错误。
//
// 参数：
//   - privateKey: PEM 编码的私钥字节切片，且 PEM block 类型必须为 RSA PRIVATE KEY。
//
// 返回：
//   - *rsa.PrivateKey: 解析成功后的 RSA 私钥对象。
//   - error: PEM 解码失败、block 类型不匹配或 PKCS#1 私钥解析失败时返回错误；解码或类型错误可使用 errors.Is 判断 ErrDecodePrivateKey。
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
