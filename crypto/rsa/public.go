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
	// BlockTypePublicKey 是本包接受和输出的 PKIX 公钥 PEM block 类型。
	BlockTypePublicKey = "PUBLIC KEY"
)

var (
	// ErrDecodePublicKey 表示 PEM 公钥解码失败或解析结果不是 RSA 公钥。
	//
	// 当输入不是 PEM、PEM block type 不是 PUBLIC KEY，或 PKIX 公钥可解析但不是 *rsa.PublicKey 时返回该错误。
	// PEM type 正确但 DER 内容非法时会透传 x509 解析错误。调用方可以使用 errors.Is 判断该错误。
	ErrDecodePublicKey = errors.New("公钥不正确。")
)

// convertPublicKey 将 PEM 编码的 PKIX PUBLIC KEY 数据解析为 *rsa.PublicKey。
//
// 参数：
//   - publicKey: PEM 编码的公钥字节切片，且 PEM block 类型必须为 PUBLIC KEY。
//
// 返回：
//   - *rsa.PublicKey: 解析成功后的 RSA 公钥对象。
//   - error: PEM 解码失败、block 类型不匹配、PKIX 解析失败或解析结果不是 RSA 公钥时返回错误；非 RSA 公钥可使用 errors.Is 判断 ErrDecodePublicKey。
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

// ConvertPubKey 将 *rsa.PublicKey 编码为 PKIX PUBLIC KEY PEM 数据。
//
// 返回的字节切片可供本包接受 PEM 公钥输入的解析与加密入口复用。
//
// 参数：
//   - publicKey: 待编码的 RSA 公钥对象，必须包含有效模数和指数。
//
// 返回：
//   - []byte: PKIX PUBLIC KEY 格式的 PEM 字节切片。
//   - error: 公钥对象无效或 PKIX 编码失败时返回错误。
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
