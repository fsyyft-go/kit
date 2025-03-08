// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package rsa

/**
 * SDK 中只包含公钥加密和私钥解密；
 * 私钥加密和公钥解密，参考：https://github.com/wenzhenxi/gorsa 。
 */

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
)

// EncryptPubKey 使用公钥对数据进行加密。
//
// 参数：
//   - publieKey：字节切片形式的 RSA 公钥数据。
//   - dataClear：需要加密的明文数据。
//
// 返回值：
//   - []byte：加密后的密文数据。
//   - error：加密过程中可能发生的错误，如公钥格式错误或加密失败。
func EncryptPubKey(publieKey, dataClear []byte) ([]byte, error) {
	// 声明密文变量和错误变量。
	var dataCipher []byte
	var err error

	// 使用 defer 和 recover 捕获可能发生的 panic，并转换为错误返回。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("公钥加密发生错误：%v", r)
		}
	}()

	// 将字节切片形式的公钥转换为 rsa.PublicKey 结构
	if pub, errPub := convertPublicKey(publieKey); errPub != nil {
		// 转换失败时保存错误信息。
		err = errPub
	} else {
		// 转换成功后调用 EncryptPublicKey 函数进行加密。
		dataCipher, err = EncryptPublicKey(pub, dataClear)
	}

	// 返回加密后的密文和可能的错误。
	return dataCipher, err
}

// EncryptPublicKey 使用公钥对数据进行加密。
//
// 参数：
//   - pubKey：RSA 公钥结构指针。
//   - dataClear：需要加密的明文数据。
//
// 返回值：
//   - []byte：使用 PKCS#1 v1.5 填充方案加密后的密文数据。
//   - error：加密过程中可能发生的错误。
func EncryptPublicKey(pubKey *rsa.PublicKey, dataClear []byte) ([]byte, error) {
	// 声明密文变量和错误变量。
	var dataCipher []byte
	var err error

	// 使用 defer 和 recover 捕获可能发生的 panic，并转换为错误返回。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("公钥加密发生错误：%v", r)
		}
	}()

	// 使用标准库函数 rsa.EncryptPKCS1v15 进行加密操作
	dataCipher, err = rsa.EncryptPKCS1v15(rand.Reader, pubKey, dataClear)

	// 返回加密后的密文和可能的错误。
	return dataCipher, err
}

// DecryptPubKey 使用公钥对数据进行解密（通常用于验证签名场景）。
//
// 参数：
//   - publieKey：字节切片形式的 RSA 公钥数据。
//   - dataCipher：需要解密的密文数据，通常是由对应私钥加密的数据。
//
// 返回值：
//   - []byte：解密后的明文数据。
//   - error：解密过程中可能发生的错误，如公钥格式错误或解密失败。
func DecryptPubKey(publieKey, dataCipher []byte) ([]byte, error) {
	// 声明明文变量和错误变量。
	var dataClear []byte
	var err error

	// 使用 defer 和 recover 捕获可能发生的 panic，并转换为错误返回。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("公钥解密发生错误：%v", r)
		}
	}()

	// 将字节切片形式的公钥转换为 rsa.PublicKey 结构
	if pub, errPub := convertPublicKey(publieKey); errPub != nil {
		// 转换失败时保存错误信息。
		err = errPub
	} else {
		// 转换成功后调用 DecryptPublicKey 函数进行解密。
		dataClear, err = DecryptPublicKey(pub, dataCipher)
	}

	// 返回解密后的明文和可能的错误。
	return dataClear, err
}

// DecryptPublicKey 使用公钥对数据进行解密（通常用于验证签名场景）。
//
// 参数：
//   - publieKey：RSA 公钥结构指针。
//   - dataCipher：需要解密的密文数据，通常是由对应私钥加密的数据。
//
// 返回值：
//   - []byte：解密后的明文数据。
//   - error：解密过程中可能发生的错误。
func DecryptPublicKey(publieKey *rsa.PublicKey, dataCipher []byte) ([]byte, error) {
	// 声明明文变量和错误变量。
	var dataClear []byte
	var err error

	// 使用 defer 和 recover 捕获可能发生的 panic，并转换为错误返回。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("公钥解密发生错误：%v", r)
		}
	}()

	// 使用 publicDecrypt 函数进行公钥解密。
	// 参数中的 crypto.Hash(0) 表示不使用任何哈希算法。
	dataClear, err = publicDecrypt(publieKey, crypto.Hash(0), nil, dataCipher)

	// 返回解密后的明文和可能的错误。
	return dataClear, err
}

// EncryptPrivKey 使用私钥对数据进行加密（通常用于数字签名场景）。
//
// 参数：
//   - privateKey：字节切片形式的 RSA 私钥数据。
//   - dataClear：需要加密的明文数据。
//
// 返回值：
//   - []byte：加密后的密文数据，实际上是一个签名。
//   - error：加密过程中可能发生的错误，如私钥格式错误或加密失败。
func EncryptPrivKey(privateKey, dataClear []byte) ([]byte, error) {
	// 声明密文变量和错误变量。
	var dataCipher []byte
	var err error

	// 使用 defer 和 recover 捕获可能发生的 panic，并转换为错误返回。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("私钥加密发生错误：%v", r)
		}
	}()

	// 将字节切片形式的私钥转换为 rsa.PrivateKey 结构。
	if priv, errPri := ConvertPrivateKey(privateKey); errPri != nil {
		// 转换失败时保存错误信息。
		err = errPri
	} else {
		// 转换成功后调用 EncryptPrivateKey 函数进行加密。
		dataCipher, err = EncryptPrivateKey(priv, dataClear)
	}

	// 返回加密后的密文和可能的错误。
	return dataCipher, err
}

// EncryptPrivateKey 使用私钥对数据进行加密（通常用于数字签名场景）。
//
// 参数：
//   - privateKey：RSA 私钥结构指针。
//   - dataClear：需要加密的明文数据。
//
// 返回值：
//   - []byte：使用 PKCS#1 v1.5 签名算法加密后的密文数据，实际上是一个签名。
//   - error：加密过程中可能发生的错误。
func EncryptPrivateKey(privateKey *rsa.PrivateKey, dataClear []byte) ([]byte, error) {
	// 声明密文变量和错误变量。
	var dataCipher []byte
	var err error

	// 使用 defer 和 recover 捕获可能发生的 panic，并转换为错误返回。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("私钥加密发生错误：%v", r)
		}
	}()

	// 使用 rsa.SignPKCS1v15 函数进行私钥加密（签名）。
	// 参数中的 crypto.Hash(0) 表示不使用任何哈希算法。
	dataCipher, err = rsa.SignPKCS1v15(nil, privateKey, crypto.Hash(0), dataClear)

	// 返回加密后的密文和可能的错误。
	return dataCipher, err
}

// DecryptPrivKey 使用私钥对数据进行解密。
//
// 参数：
//   - privateKey：字节切片形式的 RSA 私钥数据。
//   - dataCipher：需要解密的密文数据，通常是由对应公钥加密的数据。
//
// 返回值：
//   - []byte：解密后的明文数据。
//   - error：解密过程中可能发生的错误，如私钥格式错误或解密失败。
func DecryptPrivKey(privateKey, dataCipher []byte) ([]byte, error) {
	// 声明明文变量和错误变量。
	var dataClear []byte
	var err error

	// 使用 defer 和 recover 捕获可能发生的 panic，并转换为错误返回。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("私钥解密发生错误：%v", r)
		}
	}()

	// 将字节切片形式的私钥转换为 rsa.PrivateKey 结构。
	if priv, errPri := ConvertPrivateKey(privateKey); errPri != nil {
		// 转换失败时保存错误信息。
		err = errPri
	} else {
		// 转换成功后调用 DecryptPrivateKey 函数进行解密。
		dataClear, err = DecryptPrivateKey(priv, dataCipher)
	}

	// 返回解密后的明文和可能的错误。
	return dataClear, err
}

// DecryptPrivateKey 使用私钥对数据进行解密。
//
// 参数：
//   - privateKey：RSA 私钥结构指针。
//   - dataCipher：需要解密的密文数据，通常是由对应公钥加密的数据。
//
// 返回值：
//   - []byte：使用 PKCS#1 v1.5 填充方案解密后的明文数据。
//   - error：解密过程中可能发生的错误。
func DecryptPrivateKey(privateKey *rsa.PrivateKey, dataCipher []byte) ([]byte, error) {
	// 声明明文变量和错误变量。
	var dataClear []byte
	var err error

	// 使用 defer 和 recover 捕获可能发生的 panic，并转换为错误返回。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("私钥解密发生错误：%v", r)
		}
	}()

	// 使用标准库函数 rsa.DecryptPKCS1v15 进行解密操作。
	dataClear, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, dataCipher)

	// 返回解密后的明文和可能的错误。
	return dataClear, err
}
