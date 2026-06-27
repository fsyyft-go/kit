// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"
	"strings"

	kratosconfig "github.com/go-kratos/kratos/v2/config"
	kratosencoding "github.com/go-kratos/kratos/v2/encoding"
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

// WithResolve 为 Decoder 设置解码后的附加处理函数。
//
// 当 Decode 通过 codec 成功写入 target 后，会继续调用该函数处理结果 map。
//
// 参数：
//   - fn：解码完成后的附加处理函数；传入 nil 表示显式禁用默认 Resolve。
//
// 返回值：
//   - DecoderOption：可用于配置 Decoder 的选项函数。
func WithResolve(fn Resolve) DecoderOption {
	return func(o *DecoderOptions) {
		o.Resolve = fn
	}
}

// NewDecoder 创建一个新的配置解码器。
//
// 默认情况下，Decoder 会在 codec 解码成功后执行包级 defaultResolve.Resolve；
// 调用方可通过 WithResolve 覆盖该行为。
//
// 参数：
//   - opts：可选的 DecoderOption 列表；nil 选项会被忽略。
//
// 返回值：
//   - *Decoder：应用默认配置和自定义选项后的解码器实例。
func NewDecoder(opts ...DecoderOption) *Decoder {
	d := Decoder{
		DecoderOptions: DecoderOptions{
			Resolve: defaultResolve.Resolve,
		},
	}

	// 应用所有非空的解码器选项。
	for _, o := range opts {
		if nil == o {
			continue
		}
		o(&d.DecoderOptions)
	}

	return &d
}

// Decode 将 src 解码到 target map。
//
// 当 src.Format 为空时，Decode 会把点分隔的 key 展开为嵌套 map；
// 当 src.Format 非空时，Decode 使用对应的 Kratos codec.Unmarshal 写入 target，随后按需执行 Resolve。
// 当前 API 仅接受 map[string]any 作为 target，不直接解码到结构体指针。
//
// 参数：
//   - src：Kratos 配置源，包含 key、format 和原始 value。
//   - target：解码结果写入的 map[string]any。
//
// 返回值：
//   - error：codec 不存在、反序列化失败或 Resolve 处理失败时返回错误。
func (d *Decoder) Decode(src *kratosconfig.KeyValue, target map[string]any) error {
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
	if codec := kratosencoding.GetCodec(src.Format); nil != codec {
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
