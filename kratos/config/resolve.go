// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package config 提供配置解析和处理的功能，特别是对特殊格式配置值的处理。
package config

import (
	// 导入 base64 包，用于解码 base64 编码的配置值。
	"encoding/base64"
	// 导入 strings 包，用于字符串处理操作。
	"strings"
)

var (
	// defaultResolve 是默认的解析器实例，用于全局配置解析。
	defaultResolve *resolve
)

const (
	// suffixBase64 定义了 base64 编码值的后缀标识，用于识别需要 base64 解码的配置项。
	suffixBase64 = ".b64"
)

// init 函数在包初始化时执行，创建默认解析器并注册 base64 解析处理器。
func init() {
	defaultResolve = newResolve()
	defaultResolve.register(suffixBase64, registerResolveBase64)
}

type (
	// ResolveItem 是配置解析处理函数类型，用于处理特定格式的配置项。
	// 参数：
	//   - target: 配置目标映射，存储解析后的配置。
	//   - key: 当前处理的配置键名。
	//   - val: 当前处理的配置值。
	// 返回值：
	//   - error: 处理过程中可能发生的错误，成功时返回 nil。
	ResolveItem func(target map[string]interface{}, key, val string) error

	// resolve 结构体是配置解析器，管理多个解析处理函数。
	resolve struct {
		// resolvers 存储注册的解析处理函数，键为处理函数标识，值为处理函数。
		resolvers map[string]ResolveItem
	}
)

// newResolve 创建并返回一个新的 resolve 实例。
// 返回值：
//   - *resolve: 初始化后的解析器实例。
func newResolve() *resolve {
	return &resolve{
		resolvers: make(map[string]ResolveItem),
	}
}

// Resolve 递归处理配置映射中的所有项，对字符串类型的值应用所有注册的解析处理函数。
// 参数：
//   - target: 需要处理的配置映射。
//
// 返回值：
//   - error: 处理过程中可能发生的错误，成功时返回 nil。
func (r *resolve) Resolve(target map[string]interface{}) error {
	// 遍历配置映射中的所有键值对。
	for k, v := range target {
		switch vv := v.(type) {
		// 如果值是嵌套的映射，递归处理该映射。
		case map[string]interface{}:
			if err := r.Resolve(vv); nil != err {
				return err
			}
		// 如果值是数组，检查数组中的每个元素，对映射类型的元素递归处理。
		case []interface{}:
			for _, vvv := range vv {
				if vvvv, ok := vvv.(map[string]interface{}); ok {
					if err := r.Resolve(vvvv); nil != err {
						return err
					}
				}
			}
		// 如果值是字符串，应用所有注册的解析处理函数。
		case string:
			if nil != r.resolvers && len(r.resolvers) > 0 {
				for _, resolver := range r.resolvers {
					if err := resolver(target, k, vv); nil != err {
						return err
					}
				}
			}
		}
	}

	return nil
}

// register 向解析器注册一个解析处理函数。
// 参数：
//   - key: 处理函数的标识。
//   - item: 要注册的解析处理函数。
func (r *resolve) register(key string, item ResolveItem) {
	r.resolvers[key] = item
}

// RegisterResolve 向默认解析器注册一个解析处理函数。
// 参数：
//   - key: 处理函数的标识。
//   - item: 要注册的解析处理函数。
func RegisterResolve(key string, item ResolveItem) {
	defaultResolve.register(key, item)
}

// registerResolveBase64 是处理 base64 编码配置值的解析函数。
// 当配置键以 .b64 后缀结尾时，尝试将其值解码为 base64，并将解码后的值存储到去除后缀的键中。
// 参数：
//   - target: 配置目标映射，存储解析后的配置。
//   - key: 当前处理的配置键名。
//   - val: 当前处理的配置值。
//
// 返回值：
//   - error: 处理过程中可能发生的错误，成功时返回 nil。
func registerResolveBase64(target map[string]interface{}, key, val string) error {
	// 检查键名是否以 .b64 后缀结尾。
	if strings.HasSuffix(key, suffixBase64) {
		// 尝试解码 base64 值。
		if v, err := base64.StdEncoding.DecodeString(val); nil != err {
			// 解码失败时，将错误信息存储到去除后缀的键中。
			target[strings.TrimSuffix(key, suffixBase64)] = err.Error()
			return err
		} else {
			// 解码成功时，将解码后的字符串存储到去除后缀的键中。
			target[strings.TrimSuffix(key, suffixBase64)] = string(v)
		}
	}

	return nil
}
