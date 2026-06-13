// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package rsa

// SDK 中只包含公钥加密和私钥解密；
// 私钥加密和公钥解密，参考：https://github.com/wenzhenxi/gorsa 。

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
)

var (
	// ErrNilHash 表示 OAEP 加解密时没有提供哈希函数。
	ErrNilHash = errors.New("OAEP 哈希函数不能为空。")
)

// EncryptPubKey 使用 PEM 公钥按 PKCS#1 v1.5 encryption 对数据进行加密。
//
// 参数：
//   - publicKey：字节切片形式的 RSA 公钥数据。
//   - dataClear：需要加密的明文数据。
//
// 返回值：
//   - []byte：加密后的密文数据。
//   - error：加密过程中可能发生的错误，如公钥格式错误或加密失败。
//
// Deprecated: PKCS#1 v1.5 encryption 不推荐新代码使用；新代码请使用
// EncryptPubKeyOAEP。默认 OAEP 函数使用 SHA-256 和 nil label；如需指定 OAEP
// hash 或 label，请使用 EncryptPubKeyOAEPWithHash。该旧函数仅用于兼容历史密文格式或既有协议。
func EncryptPubKey(publicKey, dataClear []byte) ([]byte, error) {
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
	if pub, errPub := convertPublicKey(publicKey); errPub != nil {
		// 转换失败时保存错误信息。
		err = errPub
	} else {
		// 转换成功后调用 EncryptPublicKey 函数进行加密。
		dataCipher, err = EncryptPublicKey(pub, dataClear)
	}

	// 返回加密后的密文和可能的错误。
	return dataCipher, err
}

// EncryptPublicKey 使用 RSA 公钥结构按 PKCS#1 v1.5 encryption 对数据进行加密。
//
// 参数：
//   - pubKey：RSA 公钥结构指针。
//   - dataClear：需要加密的明文数据。
//
// 返回值：
//   - []byte：使用 PKCS#1 v1.5 填充方案加密后的密文数据。
//   - error：加密过程中可能发生的错误。
//
// Deprecated: PKCS#1 v1.5 encryption 不推荐新代码使用；新代码请使用
// EncryptPublicKeyOAEP。默认 OAEP 函数使用 SHA-256 和 nil label；如需指定 OAEP
// hash 或 label，请使用 EncryptPublicKeyOAEPWithHash。该旧函数仅用于兼容历史密文格式或既有协议。
func EncryptPublicKey(pubKey *rsa.PublicKey, dataClear []byte) (dataCipher []byte, err error) {
	// 使用 defer 和 recover 捕获可能发生的 panic，并转换为错误返回。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("公钥加密发生错误：%v", r)
		}
	}()

	// 使用标准库函数 rsa.EncryptPKCS1v15 进行加密操作。
	//lint:ignore SA1019 仅用于兼容历史 PKCS#1 v1.5 密文格式；新代码请使用 EncryptPublicKeyOAEP 或 EncryptPublicKeyOAEPWithHash。
	dataCipher, err = rsa.EncryptPKCS1v15(rand.Reader, pubKey, dataClear)

	// 返回加密后的密文和可能的错误。
	return dataCipher, err
}

// EncryptPubKeyOAEP 使用 PEM 公钥按 RSA-OAEP 对数据进行加密。
//
// 默认使用 SHA-256 作为 OAEP 哈希函数，并使用 nil label。解密时必须使用相同的
// hash 和 label；如需指定 OAEP hash 或 label，请使用 EncryptPubKeyOAEPWithHash。
//
// 参数：
//   - publicKey：字节切片形式的 RSA 公钥数据。
//   - dataClear：需要加密的明文数据。
//
// 返回值：
//   - []byte：使用 OAEP 加密后的密文数据。
//   - error：加密过程中可能发生的错误，如公钥格式错误或加密失败。
func EncryptPubKeyOAEP(publicKey, dataClear []byte) ([]byte, error) {
	return EncryptPubKeyOAEPWithHash(publicKey, dataClear, sha256.New(), nil)
}

// EncryptPubKeyOAEPWithHash 使用 PEM 公钥按指定 hash 和 label 的 RSA-OAEP 对数据进行加密。
//
// hash 不能为空，否则返回 ErrNilHash。label 可以为 nil；当提供非 nil label 时，加密和解密
// 必须使用完全一致的 hash 和 label，否则解密会失败。默认 OAEP 加密请使用 EncryptPubKeyOAEP，
// 其默认参数为 SHA-256 和 nil label。
//
// 参数：
//   - publicKey：字节切片形式的 RSA 公钥数据。
//   - dataClear：需要加密的明文数据。
//   - hash：OAEP 使用的哈希函数，加密和解密必须一致。
//   - label：OAEP 使用的标签，加密和解密必须一致。
//
// 返回值：
//   - []byte：使用 OAEP 加密后的密文数据。
//   - error：加密过程中可能发生的错误，如公钥格式错误、hash 为空或加密失败。
func EncryptPubKeyOAEPWithHash(publicKey, dataClear []byte, hash hash.Hash, label []byte) ([]byte, error) {
	if pub, errPub := convertPublicKey(publicKey); errPub != nil {
		return nil, errPub
	} else {
		return EncryptPublicKeyOAEPWithHash(pub, dataClear, hash, label)
	}
}

// EncryptPublicKeyOAEP 使用 RSA 公钥结构按 RSA-OAEP 对数据进行加密。
//
// 默认使用 SHA-256 作为 OAEP 哈希函数，并使用 nil label。解密时必须使用相同的
// hash 和 label；如需指定 OAEP hash 或 label，请使用 EncryptPublicKeyOAEPWithHash。
//
// 参数：
//   - pubKey：RSA 公钥结构指针。
//   - dataClear：需要加密的明文数据。
//
// 返回值：
//   - []byte：使用 OAEP 加密后的密文数据。
//   - error：加密过程中可能发生的错误。
func EncryptPublicKeyOAEP(pubKey *rsa.PublicKey, dataClear []byte) ([]byte, error) {
	return EncryptPublicKeyOAEPWithHash(pubKey, dataClear, sha256.New(), nil)
}

// EncryptPublicKeyOAEPWithHash 使用 RSA 公钥结构按指定 hash 和 label 的 RSA-OAEP 对数据进行加密。
//
// hash 不能为空，否则返回 ErrNilHash。label 可以为 nil；当提供非 nil label 时，加密和解密
// 必须使用完全一致的 hash 和 label，否则解密会失败。默认 OAEP 加密请使用 EncryptPublicKeyOAEP，
// 其默认参数为 SHA-256 和 nil label。
//
// 参数：
//   - pubKey：RSA 公钥结构指针。
//   - dataClear：需要加密的明文数据。
//   - hash：OAEP 使用的哈希函数，加密和解密必须一致。
//   - label：OAEP 使用的标签，加密和解密必须一致。
//
// 返回值：
//   - []byte：使用 OAEP 加密后的密文数据。
//   - error：加密过程中可能发生的错误，如 hash 为空或加密失败。
func EncryptPublicKeyOAEPWithHash(pubKey *rsa.PublicKey, dataClear []byte, hash hash.Hash, label []byte) (dataCipher []byte, err error) {
	if hash == nil {
		return nil, ErrNilHash
	}

	// 使用 defer 和 recover 捕获可能发生的 panic，并转换为错误返回。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("OAEP 公钥加密发生错误：%v", r)
		}
	}()

	dataCipher, err = rsa.EncryptOAEP(hash, rand.Reader, pubKey, dataClear, label)

	// 返回加密后的密文和可能的错误。
	return dataCipher, err
}

// DecryptPubKey 使用公钥对数据进行解密（通常用于验证签名场景）。
//
// 参数：
//   - publicKey：字节切片形式的 RSA 公钥数据。
//   - dataCipher：需要解密的密文数据，通常是由对应私钥加密的数据。
//
// 返回值：
//   - []byte：解密后的明文数据。
//   - error：解密过程中可能发生的错误，如公钥格式错误或解密失败。
func DecryptPubKey(publicKey, dataCipher []byte) ([]byte, error) {
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
	if pub, errPub := convertPublicKey(publicKey); errPub != nil {
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
//   - publicKey：RSA 公钥结构指针。
//   - dataCipher：需要解密的密文数据，通常是由对应私钥加密的数据。
//
// 返回值：
//   - []byte：解密后的明文数据。
//   - error：解密过程中可能发生的错误。
func DecryptPublicKey(publicKey *rsa.PublicKey, dataCipher []byte) (dataClear []byte, err error) {
	// 使用 defer 和 recover 捕获可能发生的 panic，并转换为错误返回。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("公钥解密发生错误：%v", r)
		}
	}()

	// 使用 publicDecrypt 函数进行公钥解密。
	// 参数中的 crypto.Hash(0) 表示不使用任何哈希算法。
	dataClear, err = publicDecrypt(publicKey, crypto.Hash(0), nil, dataCipher)

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
func EncryptPrivateKey(privateKey *rsa.PrivateKey, dataClear []byte) (dataCipher []byte, err error) {
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

// DecryptPrivKey 使用 PEM 私钥按 PKCS#1 v1.5 encryption 对应格式解密数据。
//
// 参数：
//   - privateKey：字节切片形式的 RSA 私钥数据。
//   - dataCipher：需要解密的密文数据，通常是由对应公钥加密的数据。
//
// 返回值：
//   - []byte：解密后的明文数据。
//   - error：解密过程中可能发生的错误，如私钥格式错误或解密失败。
//
// Deprecated: PKCS#1 v1.5 encryption 不推荐新代码使用；新代码请使用
// DecryptPrivKeyOAEP。默认 OAEP 函数使用 SHA-256 和 nil label；如需指定 OAEP
// hash 或 label，请使用 DecryptPrivKeyOAEPWithHash。该旧函数仅用于兼容历史密文格式或既有协议。
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

// DecryptPrivateKey 使用 RSA 私钥结构按 PKCS#1 v1.5 encryption 对应格式解密数据。
//
// 参数：
//   - privateKey：RSA 私钥结构指针。
//   - dataCipher：需要解密的密文数据，通常是由对应公钥加密的数据。
//
// 返回值：
//   - []byte：使用 PKCS#1 v1.5 填充方案解密后的明文数据。
//   - error：解密过程中可能发生的错误。
//
// Deprecated: PKCS#1 v1.5 encryption 不推荐新代码使用；新代码请使用
// DecryptPrivateKeyOAEP。默认 OAEP 函数使用 SHA-256 和 nil label；如需指定 OAEP
// hash 或 label，请使用 DecryptPrivateKeyOAEPWithHash。该旧函数仅用于兼容历史密文格式或既有协议。
func DecryptPrivateKey(privateKey *rsa.PrivateKey, dataCipher []byte) (dataClear []byte, err error) {
	// 使用 defer 和 recover 捕获可能发生的 panic，并转换为错误返回。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("私钥解密发生错误：%v", r)
		}
	}()

	// 使用标准库函数 rsa.DecryptPKCS1v15 进行解密操作。
	//lint:ignore SA1019 仅用于兼容历史 PKCS#1 v1.5 密文格式；新代码请使用 DecryptPrivateKeyOAEP 或 DecryptPrivateKeyOAEPWithHash。
	dataClear, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, dataCipher)

	// 返回解密后的明文和可能的错误。
	return dataClear, err
}

// DecryptPrivKeyOAEP 使用 PEM 私钥按 RSA-OAEP 对数据进行解密。
//
// 默认使用 SHA-256 作为 OAEP 哈希函数，并使用 nil label。加密时必须使用相同的
// hash 和 label；如需指定 OAEP hash 或 label，请使用 DecryptPrivKeyOAEPWithHash。
//
// 参数：
//   - privateKey：字节切片形式的 RSA 私钥数据。
//   - dataCipher：需要解密的 OAEP 密文数据。
//
// 返回值：
//   - []byte：解密后的明文数据。
//   - error：解密过程中可能发生的错误，如私钥格式错误或解密失败。
func DecryptPrivKeyOAEP(privateKey, dataCipher []byte) ([]byte, error) {
	return DecryptPrivKeyOAEPWithHash(privateKey, dataCipher, sha256.New(), nil)
}

// DecryptPrivKeyOAEPWithHash 使用 PEM 私钥按指定 hash 和 label 的 RSA-OAEP 对数据进行解密。
//
// hash 不能为空，否则返回 ErrNilHash。label 可以为 nil；解密时使用的 hash 和 label
// 必须与加密时完全一致，否则解密会失败。默认 OAEP 解密请使用 DecryptPrivKeyOAEP，
// 其默认参数为 SHA-256 和 nil label。
//
// 参数：
//   - privateKey：字节切片形式的 RSA 私钥数据。
//   - dataCipher：需要解密的 OAEP 密文数据。
//   - hash：OAEP 使用的哈希函数，加密和解密必须一致。
//   - label：OAEP 使用的标签，加密和解密必须一致。
//
// 返回值：
//   - []byte：解密后的明文数据。
//   - error：解密过程中可能发生的错误，如私钥格式错误、hash 为空或解密失败。
func DecryptPrivKeyOAEPWithHash(privateKey, dataCipher []byte, hash hash.Hash, label []byte) ([]byte, error) {
	if priv, errPri := ConvertPrivateKey(privateKey); errPri != nil {
		return nil, errPri
	} else {
		return DecryptPrivateKeyOAEPWithHash(priv, dataCipher, hash, label)
	}
}

// DecryptPrivateKeyOAEP 使用 RSA 私钥结构按 RSA-OAEP 对数据进行解密。
//
// 默认使用 SHA-256 作为 OAEP 哈希函数，并使用 nil label。加密时必须使用相同的
// hash 和 label；如需指定 OAEP hash 或 label，请使用 DecryptPrivateKeyOAEPWithHash。
//
// 参数：
//   - privateKey：RSA 私钥结构指针。
//   - dataCipher：需要解密的 OAEP 密文数据。
//
// 返回值：
//   - []byte：解密后的明文数据。
//   - error：解密过程中可能发生的错误。
func DecryptPrivateKeyOAEP(privateKey *rsa.PrivateKey, dataCipher []byte) ([]byte, error) {
	return DecryptPrivateKeyOAEPWithHash(privateKey, dataCipher, sha256.New(), nil)
}

// DecryptPrivateKeyOAEPWithHash 使用 RSA 私钥结构按指定 hash 和 label 的 RSA-OAEP 对数据进行解密。
//
// hash 不能为空，否则返回 ErrNilHash。label 可以为 nil；解密时使用的 hash 和 label
// 必须与加密时完全一致，否则解密会失败。默认 OAEP 解密请使用 DecryptPrivateKeyOAEP，
// 其默认参数为 SHA-256 和 nil label。
//
// 参数：
//   - privateKey：RSA 私钥结构指针。
//   - dataCipher：需要解密的 OAEP 密文数据。
//   - hash：OAEP 使用的哈希函数，加密和解密必须一致。
//   - label：OAEP 使用的标签，加密和解密必须一致。
//
// 返回值：
//   - []byte：解密后的明文数据。
//   - error：解密过程中可能发生的错误，如 hash 为空或解密失败。
func DecryptPrivateKeyOAEPWithHash(privateKey *rsa.PrivateKey, dataCipher []byte, hash hash.Hash, label []byte) (dataClear []byte, err error) {
	if hash == nil {
		return nil, ErrNilHash
	}

	// 使用 defer 和 recover 捕获可能发生的 panic，并转换为错误返回。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("OAEP 私钥解密发生错误：%v", r)
		}
	}()

	dataClear, err = rsa.DecryptOAEP(hash, rand.Reader, privateKey, dataCipher, label)

	// 返回解密后的明文和可能的错误。
	return dataClear, err
}
