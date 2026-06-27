// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package rsa

// 本文件提供公钥加密/私钥解密入口，并保留与历史私钥操作/公钥恢复格式兼容的包装函数。
// 历史兼容流程参考：https://github.com/wenzhenxi/gorsa 。

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
	// ErrNilHash 表示 OAEP 加密或解密调用未提供 hash.Hash。
	//
	// 结构体入口在 hash 为 nil 时直接返回该错误；PEM wrapper 会在密钥解析成功后透传该错误。
	// 调用方可以使用 errors.Is 判断该错误。
	ErrNilHash = errors.New("OAEP 哈希函数不能为空。")
)

// EncryptPubKey 使用 PEM 公钥按 PKCS#1 v1.5 encryption 对数据进行加密。
//
// publicKey 必须是 PUBLIC KEY 类型的 PKIX PEM 数据。明文长度受 RSA 模数字节数和
// PKCS#1 v1.5 填充开销限制；本函数不做分块。
//
// 兼容性说明：本函数的 recover 仅阻止 panic 继续传播；由于返回值未命名，
// recover 分支设置的局部 err 不会写入返回值。EncryptPublicKey 自身 recover 到的
// panic 仍会通过该函数的具名返回值以 error 形式透传。
//
// 参数：
//   - publicKey: PEM 编码的 RSA 公钥数据。
//   - dataClear: 需要加密的明文数据，长度不能超过当前密钥允许的 PKCS#1 v1.5 最大明文长度。
//
// 返回：
//   - []byte: 加密后的 PKCS#1 v1.5 密文。
//   - error: 公钥解析失败时返回错误；EncryptPublicKey 返回的明文过长、底层加密失败或 panic recover 错误会原样透传。
//
// Deprecated: PKCS#1 v1.5 encryption 不推荐新代码使用；新代码请使用
// EncryptPubKeyOAEP。默认 OAEP 函数使用 SHA-256 和 nil label；如需指定 OAEP
// hash 或 label，请使用 EncryptPubKeyOAEPWithHash。该旧函数仅用于兼容历史密文格式或既有协议。
func EncryptPubKey(publicKey, dataClear []byte) ([]byte, error) {
	// 声明密文变量和错误变量。
	var dataCipher []byte
	var err error

	// 使用 defer 和 recover 尝试拦截底层 panic，避免 panic 继续向外传播。
	// 本函数使用未命名返回值，recover 分支保存到局部 err 的错误信息不会写入返回值。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("公钥加密发生错误：%v", r)
		}
	}()

	// 将字节切片形式的公钥转换为 rsa.PublicKey 结构。
	if pub, errPub := convertPublicKey(publicKey); errPub != nil {
		// 转换失败时保存错误信息，并在正常 return 路径中返回。
		err = errPub
	} else {
		// 转换成功后调用 EncryptPublicKey 函数进行加密。
		dataCipher, err = EncryptPublicKey(pub, dataClear)
	}

	// 返回正常执行路径得到的密文和可能的错误。
	return dataCipher, err
}

// EncryptPublicKey 使用 RSA 公钥结构按 PKCS#1 v1.5 encryption 对数据进行加密。
//
// 明文长度受 RSA 模数字节数和 PKCS#1 v1.5 填充开销限制；本函数不做分块。
//
// 参数：
//   - pubKey: RSA 公钥对象，必须非 nil 且包含有效模数和指数。
//   - dataClear: 需要加密的明文数据，长度不能超过当前密钥允许的 PKCS#1 v1.5 最大明文长度。
//
// 返回：
//   - []byte: 使用 PKCS#1 v1.5 填充方案加密后的密文数据。
//   - error: 公钥无效、明文过长、底层加密失败或 panic 被拦截时返回错误。
//
// Deprecated: PKCS#1 v1.5 encryption 不推荐新代码使用；新代码请使用
// EncryptPublicKeyOAEP。默认 OAEP 函数使用 SHA-256 和 nil label；如需指定 OAEP
// hash 或 label，请使用 EncryptPublicKeyOAEPWithHash。该旧函数仅用于兼容历史密文格式或既有协议。
func EncryptPublicKey(pubKey *rsa.PublicKey, dataClear []byte) (dataCipher []byte, err error) {
	// 使用 defer 和 recover 尝试拦截底层 panic，避免 panic 继续向外传播。
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
// 本函数不做分块，明文长度不能超过当前密钥和 SHA-256 OAEP 允许的最大值。
//
// 参数：
//   - publicKey: PEM 编码的 RSA 公钥数据。
//   - dataClear: 需要加密的明文数据。
//
// 返回：
//   - []byte: 使用 OAEP 加密后的密文数据。
//   - error: 公钥解析失败、明文过长或 OAEP 加密失败时返回错误。
func EncryptPubKeyOAEP(publicKey, dataClear []byte) ([]byte, error) {
	return EncryptPubKeyOAEPWithHash(publicKey, dataClear, sha256.New(), nil)
}

// EncryptPubKeyOAEPWithHash 使用 PEM 公钥按指定 hash 和 label 的 RSA-OAEP 对数据进行加密。
//
// hash 不能为空，否则在公钥解析成功后返回 ErrNilHash。label 可以为 nil；当提供非 nil
// label 时，加密和解密必须使用完全一致的 hash 和 label，否则解密会失败。默认 OAEP
// 加密请使用 EncryptPubKeyOAEP，其默认参数为 SHA-256 和 nil label。
//
// 参数：
//   - publicKey: PEM 编码的 RSA 公钥数据。
//   - dataClear: 需要加密的明文数据，长度受当前密钥和 hash 输出长度共同限制。
//   - hash: OAEP 使用的哈希函数，不能为 nil；加密和解密必须使用相同算法。
//   - label: OAEP 使用的标签，可以为 nil；加密和解密必须逐字节一致。
//
// 返回：
//   - []byte: 使用 OAEP 加密后的密文数据。
//   - error: 公钥解析失败、hash 为 nil、明文过长或 OAEP 加密失败时返回错误；hash 为 nil 时可使用 errors.Is 判断 ErrNilHash。
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
// 本函数不做分块，明文长度不能超过当前密钥和 SHA-256 OAEP 允许的最大值。
//
// 参数：
//   - pubKey: RSA 公钥对象，必须非 nil 且包含有效模数和指数。
//   - dataClear: 需要加密的明文数据。
//
// 返回：
//   - []byte: 使用 OAEP 加密后的密文数据。
//   - error: 公钥无效、明文过长或 OAEP 加密失败时返回错误。
func EncryptPublicKeyOAEP(pubKey *rsa.PublicKey, dataClear []byte) ([]byte, error) {
	return EncryptPublicKeyOAEPWithHash(pubKey, dataClear, sha256.New(), nil)
}

// EncryptPublicKeyOAEPWithHash 使用 RSA 公钥结构按指定 hash 和 label 的 RSA-OAEP 对数据进行加密。
//
// hash 不能为空，否则返回 ErrNilHash。label 可以为 nil；当提供非 nil label 时，
// 加密和解密必须使用完全一致的 hash 和 label，否则解密会失败。默认 OAEP 加密请使用
// EncryptPublicKeyOAEP，其默认参数为 SHA-256 和 nil label。
//
// 参数：
//   - pubKey: RSA 公钥对象，必须非 nil 且包含有效模数和指数。
//   - dataClear: 需要加密的明文数据，长度受当前密钥和 hash 输出长度共同限制。
//   - hash: OAEP 使用的哈希函数，不能为 nil；加密和解密必须使用相同算法。
//   - label: OAEP 使用的标签，可以为 nil；加密和解密必须逐字节一致。
//
// 返回：
//   - []byte: 使用 OAEP 加密后的密文数据。
//   - error: hash 为 nil、明文过长、公钥无效、底层加密失败或 panic 被拦截时返回错误；hash 为 nil 时可使用 errors.Is 判断 ErrNilHash。
func EncryptPublicKeyOAEPWithHash(pubKey *rsa.PublicKey, dataClear []byte, hash hash.Hash, label []byte) (dataCipher []byte, err error) {
	if hash == nil {
		return nil, ErrNilHash
	}

	// 使用 defer 和 recover 尝试拦截底层 panic，避免 panic 继续向外传播。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("OAEP 公钥加密发生错误：%v", r)
		}
	}()

	dataCipher, err = rsa.EncryptOAEP(hash, rand.Reader, pubKey, dataClear, label)

	// 返回加密后的密文和可能的错误。
	return dataCipher, err
}

// DecryptPubKey 使用 PEM 公钥执行与 EncryptPrivKey 兼容的历史公钥恢复操作。
//
// 该函数不会先做哈希，也不执行标准意义上的签名验签。它仅适用于兼容历史
// “私钥加密、公钥解密”数据格式；调用方需要自行处理认证和协议校验。
//
// 兼容性说明：本函数的 recover 仅阻止 panic 继续传播；由于返回值未命名，
// recover 分支设置的局部 err 不会写入返回值。DecryptPublicKey 自身 recover 到的
// panic 仍会通过该函数的具名返回值以 error 形式透传。
//
// 参数：
//   - publicKey: PEM 编码的 RSA 公钥数据。
//   - dataCipher: 由历史私钥操作产生、待恢复的输入数据。
//
// 返回：
//   - []byte: 恢复出的原始消息块。
//   - error: 公钥解析失败时返回错误；DecryptPublicKey 返回的公钥恢复失败或 panic recover 错误会原样透传。
func DecryptPubKey(publicKey, dataCipher []byte) ([]byte, error) {
	// 声明明文变量和错误变量。
	var dataClear []byte
	var err error

	// 使用 defer 和 recover 尝试拦截底层 panic，避免 panic 继续向外传播。
	// 本函数使用未命名返回值，recover 分支保存到局部 err 的错误信息不会写入返回值。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("公钥解密发生错误：%v", r)
		}
	}()

	// 将字节切片形式的公钥转换为 rsa.PublicKey 结构。
	if pub, errPub := convertPublicKey(publicKey); errPub != nil {
		// 转换失败时保存错误信息，并在正常 return 路径中返回。
		err = errPub
	} else {
		// 转换成功后调用 DecryptPublicKey 函数进行解密。
		dataClear, err = DecryptPublicKey(pub, dataCipher)
	}

	// 返回正常执行路径得到的明文和可能的错误。
	return dataClear, err
}

// DecryptPublicKey 使用 RSA 公钥执行与 EncryptPrivateKey 兼容的历史公钥恢复操作。
//
// 该函数调用原始 PKCS#1 v1.5 公钥恢复流程，不会对输入做哈希，也不等同于标准验签 API。
// 调用方需要自行判断恢复出的内容是否符合上层协议。
//
// 参数：
//   - publicKey: RSA 公钥对象，必须非 nil 且包含有效模数和指数。
//   - dataCipher: 由历史私钥操作产生、待恢复的输入数据。
//
// 返回：
//   - []byte: 恢复出的原始消息块。
//   - error: 公钥恢复失败或 panic 被拦截时返回错误。
func DecryptPublicKey(publicKey *rsa.PublicKey, dataCipher []byte) (dataClear []byte, err error) {
	// 使用 defer 和 recover 尝试拦截底层 panic，避免 panic 继续向外传播。
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

// EncryptPrivKey 使用 PEM 私钥执行与 DecryptPubKey 兼容的历史私钥操作。
//
// 该函数不会先对 dataClear 做哈希，也不等同于标准签名 API。dataClear 长度受 RSA 模数字节数和
// PKCS#1 v1.5 填充开销限制；需要标准签名语义时，应直接使用 crypto/rsa 的签名函数。
//
// 兼容性说明：本函数的 recover 仅阻止 panic 继续传播；由于返回值未命名，
// recover 分支设置的局部 err 不会写入返回值。EncryptPrivateKey 自身 recover 到的
// panic 仍会通过该函数的具名返回值以 error 形式透传。
//
// 参数：
//   - privateKey: PEM 编码的 RSA 私钥数据。
//   - dataClear: 待处理的原始消息块。
//
// 返回：
//   - []byte: 与历史公钥恢复流程兼容的输出数据。
//   - error: 私钥解析失败时返回错误；EncryptPrivateKey 返回的消息过长、私钥操作失败或 panic recover 错误会原样透传。
func EncryptPrivKey(privateKey, dataClear []byte) ([]byte, error) {
	// 声明密文变量和错误变量。
	var dataCipher []byte
	var err error

	// 使用 defer 和 recover 尝试拦截底层 panic，避免 panic 继续向外传播。
	// 本函数使用未命名返回值，recover 分支保存到局部 err 的错误信息不会写入返回值。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("私钥加密发生错误：%v", r)
		}
	}()

	// 将字节切片形式的私钥转换为 rsa.PrivateKey 结构。
	if priv, errPri := ConvertPrivateKey(privateKey); errPri != nil {
		// 转换失败时保存错误信息，并在正常 return 路径中返回。
		err = errPri
	} else {
		// 转换成功后调用 EncryptPrivateKey 函数进行加密。
		dataCipher, err = EncryptPrivateKey(priv, dataClear)
	}

	// 返回正常执行路径得到的密文和可能的错误。
	return dataCipher, err
}

// EncryptPrivateKey 使用 RSA 私钥执行与 DecryptPublicKey 兼容的历史私钥操作。
//
// 它基于 PKCS#1 v1.5 原始私钥流程处理输入数据，不会先做哈希。dataClear 长度受 RSA 模数字节数和
// PKCS#1 v1.5 填充开销限制；如需标准签名语义，应使用 crypto/rsa 的签名函数。
//
// 参数：
//   - privateKey: RSA 私钥对象，必须非 nil 且包含有效参数。
//   - dataClear: 待处理的原始消息块。
//
// 返回：
//   - []byte: 与历史公钥恢复流程兼容的输出数据。
//   - error: 消息过长、私钥操作失败或 panic 被拦截时返回错误。
func EncryptPrivateKey(privateKey *rsa.PrivateKey, dataClear []byte) (dataCipher []byte, err error) {
	// 使用 defer 和 recover 尝试拦截底层 panic，避免 panic 继续向外传播。
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
// privateKey 必须是 RSA PRIVATE KEY 类型的 PKCS#1 PEM 数据。本函数仅用于兼容历史
// PKCS#1 v1.5 encryption 密文格式，不提供 OAEP 的 label/hash 校验语义。
//
// 兼容性说明：本函数的 recover 仅阻止 panic 继续传播；由于返回值未命名，
// recover 分支设置的局部 err 不会写入返回值。DecryptPrivateKey 自身 recover 到的
// panic 仍会通过该函数的具名返回值以 error 形式透传。
//
// 参数：
//   - privateKey: PEM 编码的 RSA 私钥数据。
//   - dataCipher: 需要解密的 PKCS#1 v1.5 密文数据，通常由对应公钥加密得到。
//
// 返回：
//   - []byte: 解密后的明文数据。
//   - error: 私钥解析失败时返回错误；DecryptPrivateKey 返回的密文格式非法、底层解密失败或 panic recover 错误会原样透传。
//
// Deprecated: PKCS#1 v1.5 encryption 不推荐新代码使用；新代码请使用
// DecryptPrivKeyOAEP。默认 OAEP 函数使用 SHA-256 和 nil label；如需指定 OAEP
// hash 或 label，请使用 DecryptPrivKeyOAEPWithHash。该旧函数仅用于兼容历史密文格式或既有协议。
func DecryptPrivKey(privateKey, dataCipher []byte) ([]byte, error) {
	// 声明明文变量和错误变量。
	var dataClear []byte
	var err error

	// 使用 defer 和 recover 尝试拦截底层 panic，避免 panic 继续向外传播。
	// 本函数使用未命名返回值，recover 分支保存到局部 err 的错误信息不会写入返回值。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("私钥解密发生错误：%v", r)
		}
	}()

	// 将字节切片形式的私钥转换为 rsa.PrivateKey 结构。
	if priv, errPri := ConvertPrivateKey(privateKey); errPri != nil {
		// 转换失败时保存错误信息，并在正常 return 路径中返回。
		err = errPri
	} else {
		// 转换成功后调用 DecryptPrivateKey 函数进行解密。
		dataClear, err = DecryptPrivateKey(priv, dataCipher)
	}

	// 返回正常执行路径得到的明文和可能的错误。
	return dataClear, err
}

// DecryptPrivateKey 使用 RSA 私钥结构按 PKCS#1 v1.5 encryption 对应格式解密数据。
//
// 本函数仅用于兼容历史 PKCS#1 v1.5 encryption 密文格式，不提供 OAEP 的 label/hash 校验语义。
//
// 参数：
//   - privateKey: RSA 私钥对象，必须非 nil 且包含有效参数。
//   - dataCipher: 需要解密的 PKCS#1 v1.5 密文数据，通常由对应公钥加密得到。
//
// 返回：
//   - []byte: 使用 PKCS#1 v1.5 填充方案解密后的明文数据。
//   - error: 私钥无效、密文格式非法、底层解密失败或 panic 被拦截时返回错误。
//
// Deprecated: PKCS#1 v1.5 encryption 不推荐新代码使用；新代码请使用
// DecryptPrivateKeyOAEP。默认 OAEP 函数使用 SHA-256 和 nil label；如需指定 OAEP
// hash 或 label，请使用 DecryptPrivateKeyOAEPWithHash。该旧函数仅用于兼容历史密文格式或既有协议。
func DecryptPrivateKey(privateKey *rsa.PrivateKey, dataCipher []byte) (dataClear []byte, err error) {
	// 使用 defer 和 recover 尝试拦截底层 panic，避免 panic 继续向外传播。
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
//   - privateKey: PEM 编码的 RSA 私钥数据。
//   - dataCipher: 需要解密的 OAEP 密文数据。
//
// 返回：
//   - []byte: 解密后的明文数据。
//   - error: 私钥解析失败、密文格式非法或 OAEP 解密失败时返回错误。
func DecryptPrivKeyOAEP(privateKey, dataCipher []byte) ([]byte, error) {
	return DecryptPrivKeyOAEPWithHash(privateKey, dataCipher, sha256.New(), nil)
}

// DecryptPrivKeyOAEPWithHash 使用 PEM 私钥按指定 hash 和 label 的 RSA-OAEP 对数据进行解密。
//
// hash 不能为空，否则在私钥解析成功后返回 ErrNilHash。label 可以为 nil；解密时使用的
// hash 和 label 必须与加密时完全一致，否则解密会失败。默认 OAEP 解密请使用
// DecryptPrivKeyOAEP，其默认参数为 SHA-256 和 nil label。
//
// 参数：
//   - privateKey: PEM 编码的 RSA 私钥数据。
//   - dataCipher: 需要解密的 OAEP 密文数据。
//   - hash: OAEP 使用的哈希函数，不能为 nil；加密和解密必须使用相同算法。
//   - label: OAEP 使用的标签，可以为 nil；加密和解密必须逐字节一致。
//
// 返回：
//   - []byte: 解密后的明文数据。
//   - error: 私钥解析失败、hash 为 nil、密文格式非法或 OAEP 解密失败时返回错误；hash 为 nil 时可使用 errors.Is 判断 ErrNilHash。
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
//   - privateKey: RSA 私钥对象，必须非 nil 且包含有效参数。
//   - dataCipher: 需要解密的 OAEP 密文数据。
//
// 返回：
//   - []byte: 解密后的明文数据。
//   - error: 私钥无效、密文格式非法或 OAEP 解密失败时返回错误。
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
//   - privateKey: RSA 私钥对象，必须非 nil 且包含有效参数。
//   - dataCipher: 需要解密的 OAEP 密文数据。
//   - hash: OAEP 使用的哈希函数，不能为 nil；加密和解密必须使用相同算法。
//   - label: OAEP 使用的标签，可以为 nil；加密和解密必须逐字节一致。
//
// 返回：
//   - []byte: 解密后的明文数据。
//   - error: hash 为 nil、私钥无效、密文格式非法、OAEP 解密失败或 panic 被拦截时返回错误；hash 为 nil 时可使用 errors.Is 判断 ErrNilHash。
func DecryptPrivateKeyOAEPWithHash(privateKey *rsa.PrivateKey, dataCipher []byte, hash hash.Hash, label []byte) (dataClear []byte, err error) {
	if hash == nil {
		return nil, ErrNilHash
	}

	// 使用 defer 和 recover 尝试拦截底层 panic，避免 panic 继续向外传播。
	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("OAEP 私钥解密发生错误：%v", r)
		}
	}()

	dataClear, err = rsa.DecryptOAEP(hash, rand.Reader, privateKey, dataCipher, label)

	// 返回解密后的明文和可能的错误。
	return dataClear, err
}
