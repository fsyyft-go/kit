// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package pinyin

import (
	py "github.com/mozillazg/go-pinyin"
)

type (
	// Pinyin 定义中文文本转拼音转换器的公共契约。
	//
	// 实现应返回按可转换字符排列的拼音候选列表，并保留底层词典对非汉字字符和多音字的处理规则。
	Pinyin interface {
		// Pinyin 将文本转换为按可转换字符排列的拼音候选二维切片。
		//
		// 参数：
		//   - s: 待转换文本；非汉字字符是否产生结果由底层 go-pinyin 转换规则决定。
		//
		// 返回：
		//   - [][]string: 每个外层元素对应一个可转换字符的拼音候选；空输入或不含可转换字符时返回空结果。
		Pinyin(s string) [][]string
	}

	// pinyin 保存调用底层 go-pinyin 所需的转换参数。
	pinyin struct {
		// args 是传递给底层 go-pinyin 的转换配置。
		args py.Args
	}

	// Option 表示 NewPinyin 初始化转换器时应用的配置函数。
	//
	// 参数：
	//   - *pinyin: NewPinyin 创建的内部配置对象；Option 通过修改该对象影响后续转换。
	Option func(*pinyin)
)

const (
	// Normal 表示不带声调的完整拼音风格，是 NewPinyin 的默认 style。
	Normal = py.Normal
	// Tone 表示使用声调符号的完整拼音风格，例如 zhōng、guó。
	Tone = py.Tone
	// Tone2 表示将声调数字放在韵母内部约定位置的完整拼音风格，例如 zho1ng、guo2。
	Tone2 = py.Tone2
	// Tone3 表示将声调数字放在拼音末尾的完整拼音风格，例如 zhong1、guo2。
	Tone3 = py.Tone3
	// Initials 表示只返回声母部分的拼音风格；没有声母的拼音按底层 go-pinyin 规则处理。
	Initials = py.Initials
	// FirstLetter 表示只返回拼音首字母的风格。
	FirstLetter = py.FirstLetter
	// Finals 表示只返回不带声调韵母的拼音风格。
	Finals = py.Finals
	// FinalsTone 表示只返回带声调符号韵母的拼音风格。
	FinalsTone = py.FinalsTone
	// FinalsTone2 表示只返回带声调数字韵母的拼音风格，声调数字位于韵母内部约定位置。
	FinalsTone2 = py.FinalsTone2
	// FinalsTone3 表示只返回带声调数字韵母的拼音风格，声调数字位于韵母末尾。
	FinalsTone3 = py.FinalsTone3
)

var (
	// defaultStyle 是 NewPinyin 未显式配置 style 时使用的默认拼音风格。
	defaultStyle = Normal
	// defaultHeteronym 是 NewPinyin 未显式配置 heteronym 时使用的多音字开关。
	defaultHeteronym = false
	// defaultSeparator 是写入底层 go-pinyin Args 的默认分隔符；当前 Pinyin 方法不会将结果拼接为字符串。
	defaultSeparator = " "
)

// NewPinyin 创建使用默认配置并按顺序应用选项的拼音转换器。
//
// 参数：
//   - opts: 可选配置函数，按传入顺序应用；未传入时使用 Normal、关闭多音字和底层空格分隔符配置。传入 nil Option
//     会在应用时引发 panic。
//
// 返回：
//   - Pinyin: 初始化完成的拼音转换器，后续转换使用创建时保存的配置。
func NewPinyin(opts ...Option) Pinyin {
	p := &pinyin{
		args: py.Args{
			Style:     defaultStyle,
			Heteronym: defaultHeteronym,
			Separator: defaultSeparator,
		},
	}

	// 按调用方传入顺序应用选项，使后传入的配置可以覆盖先前配置。
	if nil != opts && len(opts) > 0 {
		for _, opt := range opts {
			opt(p)
		}
	}

	return p
}

// WithStyle 返回设置拼音风格的配置选项。
//
// 参数：
//   - style: 拼音风格枚举值；本函数不校验该值，会原样写入底层 go-pinyin
//     Args.Style。本包导出的取值包括：
//   - Normal: 不带声调的完整拼音，也是 NewPinyin 的默认风格，例如 zhong、guo。
//   - Tone: 使用声调符号的完整拼音，例如 zhōng、guó。
//   - Tone2: 将声调数字放在韵母内部约定位置的完整拼音，例如 zho1ng、guo2。
//   - Tone3: 将声调数字放在拼音末尾的完整拼音，例如 zhong1、guo2。
//   - Initials: 只返回声母部分；没有声母的拼音按底层 go-pinyin 规则处理。
//   - FirstLetter: 只返回拼音首字母。
//   - Finals: 只返回不带声调的韵母。
//   - FinalsTone: 只返回带声调符号的韵母。
//   - FinalsTone2: 只返回带声调数字的韵母，声调数字位于韵母内部约定位置。
//   - FinalsTone3: 只返回带声调数字的韵母，声调数字位于韵母末尾。
//     未列出的整数也会原样传递，效果由底层 go-pinyin 实现决定。
//
// 返回：
//   - Option: NewPinyin 应用后将转换器 style 设置为 style 的配置函数。
func WithStyle(style int) Option {
	return func(p *pinyin) {
		p.args.Style = style
	}
}

// WithHeteronym 返回设置多音字模式的配置选项。
//
// 参数：
//   - heteronym: true 表示保留多音字候选，false 表示使用底层默认读音。
//
// 返回：
//   - Option: NewPinyin 应用后将转换器 heteronym 设置为 heteronym 的配置函数。
func WithHeteronym(heteronym bool) Option {
	return func(p *pinyin) {
		p.args.Heteronym = heteronym
	}
}

// WithSeparator 返回设置底层分隔符字段的配置选项。
//
// 参数：
//   - separator: 写入底层 go-pinyin Args.Separator 的分隔符；当前 Pinyin 方法返回 [][]string，不会把结果拼接成使用该分隔符的
//     字符串。
//
// 返回：
//   - Option: NewPinyin 应用后将转换器底层 separator 设置为 separator 的配置函数。
func WithSeparator(separator string) Option {
	return func(p *pinyin) {
		p.args.Separator = separator
	}
}

// Pinyin 将文本转换为按可转换字符排列的拼音候选二维切片。
//
// 参数：
//   - s: 待转换文本；非汉字字符是否产生结果由底层 go-pinyin 转换规则决定。
//
// 返回：
//   - [][]string: 每个外层元素对应一个可转换字符的拼音候选；关闭多音字时通常每项一个读音，开启多音字时每项可能包含多个候选，空输入或不含可转换字符时返回空结果。
func (p *pinyin) Pinyin(s string) [][]string {
	return py.Pinyin(s, p.args)
}
