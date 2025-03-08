// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package config 提供配置解码和处理的功能。
package config

import (
	"fmt"
	"strings"

	kratos_config "github.com/go-kratos/kratos/v2/config"
	kratos_encoding "github.com/go-kratos/kratos/v2/encoding"
)

type (
	// Resolve 是一个函数类型，用于在配置解码后对目标映射进行额外处理。
	// 它接收配置目标映射作为参数，返回处理过程中可能发生的错误。
	Resolve func(target map[string]interface{}) error

	// DecoderOption 是一个函数类型，用于配置 Decoder 的选项。
	// 它接收 DecoderOptions 指针作为参数，用于修改解码器的配置。
	DecoderOption func(*DecoderOptions)

	// DecoderOptions 包含解码器的配置选项。
	DecoderOptions struct {
		// Resolve 是一个可选的解析函数，用于在解码完成后对配置进行额外处理。
		Resolve Resolve
	}

	// Decoder 是配置解码器，用于将配置从源格式解码到目标映射。
	// 它嵌入了 DecoderOptions 结构体，继承了其所有字段和方法。
	Decoder struct {
		DecoderOptions
	}
)

// WithResolve 返回一个 DecoderOption，用于设置解码器的 Resolve 函数。
// 该函数在配置解码完成后被调用，可以对解码后的配置进行额外处理。
//
// 参数：
//   - fn: 要设置的解析函数。
//
// 返回值：
//   - DecoderOption: 可用于配置解码器的选项函数。
func WithResolve(fn Resolve) DecoderOption {
	return func(o *DecoderOptions) {
		o.Resolve = fn
	}
}

// NewDecoder 创建并返回一个新的 Decoder 实例。
//
// 参数：
//   - opts: 可选的解码器配置选项列表。
//
// 返回值：
//   - *Decoder: 配置好的解码器实例。
func NewDecoder(opts ...DecoderOption) *Decoder {
	d := Decoder{
		DecoderOptions: DecoderOptions{
			Resolve: defaultResolve.Resolve,
		},
	}

	// 应用所有非空的解码器选项。
	for _, o := range opts {
		if nil != o {
			o(&d.DecoderOptions)
		}
	}

	return &d
}

// Decode 将配置从源 KeyValue 解码到目标映射。
// 如果 src.Format 为空，则将点分隔的键展开为嵌套映射。
// 如果 src.Format 不为空，则使用对应的编解码器解码配置值。
//
// 参数：
//   - src: 源配置的 KeyValue 对象。
//   - target: 解码后的配置将存储在此映射中。
//
// 返回值：
//   - error: 解码过程中可能发生的错误，成功时返回 nil。
func (d *Decoder) Decode(src *kratos_config.KeyValue, target map[string]any) error {
	if src.Format == "" {
		// 当格式为空时，将键 "aaa.bbb" 展开为 map[aaa]map[bbb]interface{}。
		keys := strings.Split(src.Key, ".")
		for i, k := range keys {
			if i == len(keys)-1 {
				// 如果是最后一个键，直接设置值。
				target[k] = src.Value
			} else {
				// 否则创建一个子映射并继续处理。
				sub := make(map[string]any)
				target[k] = sub
				target = sub
			}
		}
		return nil
	}
	if codec := kratos_encoding.GetCodec(src.Format); codec != nil {
		// 使用对应格式的编解码器解码配置值。
		if err := codec.Unmarshal(src.Value, &target); nil != err {
			return err
		}
		// 如果设置了解析函数，则调用它进行额外处理。
		if nil != d.Resolve {
			if err := d.Resolve(target); nil != err {
				return err
			}
		}

		return nil
	}
	// 如果不支持该格式，返回错误。
	return fmt.Errorf("unsupported key: %s format: %s", src.Key, src.Format)
}
