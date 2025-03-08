// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package rsa 提供了 RSA 加密算法相关的功能实现。
package rsa

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

const (
	// BlockTypePublicKey 定义了 PEM 编码中公钥块的类型标识。
	BlockTypePublicKey = "PUBLIC KEY"
)

var (
	// ErrDecodePublicKey 表示公钥解析失败时返回的错误。
	ErrDecodePublicKey = errors.New("公钥不正确。")
)

// convertPublicKey 将 PEM 格式的公钥字节数据转换为 RSA 公钥对象。
//
// 参数：
//   - publicKey：PEM 格式的公钥字节数据。
//
// 返回值：
//   - *rsa.PublicKey：转换后的 RSA 公钥对象。
//   - error：转换过程中可能发生的错误。
func convertPublicKey(publicKey []byte) (*rsa.PublicKey, error) {
	var pub *rsa.PublicKey
	var err error

	// 解码 PEM 格式的公钥数据，并验证块类型是否为 PUBLIC KEY。
	if block, _ := pem.Decode(publicKey); nil == block || block.Type != BlockTypePublicKey {
		err = ErrDecodePublicKey
	} else if public, errPublic := x509.ParsePKIXPublicKey(block.Bytes); nil != errPublic {
		// 解析 PKIX 格式的公钥失败。
		err = errPublic
	} else if tmp, ok := public.(*rsa.PublicKey); !ok {
		// 类型断言失败，不是 RSA 公钥类型。
		err = ErrDecodePublicKey
	} else {
		// 成功获取 RSA 公钥。
		pub = tmp
	}

	return pub, err
}

// ConvertPubKey 将 RSA 公钥对象转换为 PEM 格式的字节数据。
//
// 参数：
//   - publicKey：RSA 公钥对象。
//
// 返回值：
//   - []byte：转换后的 PEM 格式公钥字节数据。
//   - error：转换过程中可能发生的错误。
func ConvertPubKey(publicKey *rsa.PublicKey) ([]byte, error) {
	var pubKey []byte
	var err error

	// 将公钥对象编码为 PKIX 格式。
	if bs, errMarshal := x509.MarshalPKIXPublicKey(publicKey); nil != errMarshal {
		err = errMarshal
	} else {
		// 创建 PEM 块并编码。
		block := pem.Block{
			Type:  BlockTypePublicKey,
			Bytes: bs,
		}
		pubKey = pem.EncodeToMemory(&block)
	}

	return pubKey, err
}
