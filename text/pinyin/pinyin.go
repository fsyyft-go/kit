// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package pinyin

import (
	py "github.com/mozillazg/go-pinyin"
)

type (
	// Pinyin 接口定义了中文文本转拼音的标准方法。
	//
	// 该接口用于将输入的中文文本转换为拼音，支持多音字。
	Pinyin interface {
		// Pinyin 将输入的中文文本转换为拼音。
		//
		// 参数：
		//   - s：需要转换的中文文本。
		//
		// 返回：
		//   - 返回一个二维数组，支持多音字。
		Pinyin(s string) [][]string
	}

	// pinyin 结构体实现了 Pinyin 接口，封装了拼音转换的参数。
	pinyin struct {
		// args 用于存储拼音转换的参数。
		args py.Args
	}

	// Option 定义了用于配置 pinyin 结构体的函数类型。
	Option func(*pinyin)
)

// 拼音风格常量定义（推荐使用）。
const (
	Normal      = py.Normal      // 普通风格，不带声调（默认风格）。如： zhong guo
	Tone        = py.Tone        // 声调风格 1，拼音声调在韵母第一个字母上。如： zhōng guó
	Tone2       = py.Tone2       // 声调风格 2，拼音声调在各个韵母之后，用数字 [1-4] 表示。如： zho1ng guo2
	Tone3       = py.Tone3       // 声调风格 3，拼音声调在各个拼音之后，用数字 [1-4] 表示。如： zhong1 guo2
	Initials    = py.Initials    // 声母风格，只返回各个拼音的声母部分。如： zh g 。注意：不是所有的拼音都有声母。
	FirstLetter = py.FirstLetter // 首字母风格，只返回拼音的首字母部分。如： z g
	Finals      = py.Finals      // 韵母风格，只返回各个拼音的韵母部分，不带声调。如： ong uo
	FinalsTone  = py.FinalsTone  // 韵母风格 1，带声调，声调在韵母第一个字母上。如： ōng uó
	FinalsTone2 = py.FinalsTone2 // 韵母风格 2，带声调，声调在各个韵母之后，用数字 [1-4] 表示。如： o1ng uo2
	FinalsTone3 = py.FinalsTone3 // 韵母风格 3，带声调，声调在各个拼音之后，用数字 [1-4] 表示。如： ong1 uo2
)

var (
	// defaultStyle 默认拼音风格。
	defaultStyle = Normal
	// defaultHeteronym 默认是否启用多音字。
	defaultHeteronym = false
	// defaultSeparator 默认分隔符。
	defaultSeparator = " "
)

// NewPinyin 创建一个新的 Pinyin 实例。
//
// 参数：
//   - opts：可选参数，用于自定义拼音转换的行为。
//
// 返回：
//   - 返回实现了 Pinyin 接口的实例。
func NewPinyin(opts ...Option) Pinyin {
	p := &pinyin{
		args: py.Args{
			Style:     defaultStyle,
			Heteronym: defaultHeteronym,
			Separator: defaultSeparator,
		},
	}

	// 如果有自定义选项，则依次应用。
	if nil != opts && len(opts) > 0 {
		for _, opt := range opts {
			opt(p)
		}
	}

	return p
}

// WithStyle 设置拼音风格。
//
// 参数：
//   - style：拼音风格常量。
//
// 返回：
//   - 返回 Option，用于配置 pinyin。
func WithStyle(style int) Option {
	return func(p *pinyin) {
		p.args.Style = style
	}
}

// WithHeteronym 设置是否启用多音字。
//
// 参数：
//   - heteronym：布尔值，true 表示启用多音字。
//
// 返回：
//   - 返回 Option，用于配置 pinyin。
func WithHeteronym(heteronym bool) Option {
	return func(p *pinyin) {
		p.args.Heteronym = heteronym
	}
}

// WithSeparator 设置拼音分隔符。
//
// 参数：
//   - separator：分隔符字符串。
//
// 返回：
//   - 返回 Option，用于配置 pinyin。
func WithSeparator(separator string) Option {
	return func(p *pinyin) {
		p.args.Separator = separator
	}
}

// Pinyin 实现了 Pinyin 接口的方法，将中文文本转换为拼音。
//
// 参数：
//   - s：需要转换的中文文本。
//
// 返回：
//   - 返回一个二维数组，支持多音字。
func (p *pinyin) Pinyin(s string) [][]string {
	return py.Pinyin(s, p.args)
}
