// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package pinyin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewPinyin_DefaultAndBoundaryInputs 验证默认 Pinyin 实例对标准中文和边界文本的转换行为。
//
// 该测试通过表驱动用例覆盖默认 Normal 风格、空输入、无拼音字符、空白字符、中文标点以及 ASCII/数字/emoji 混合文本，
// 确保包装层保持 go-pinyin 对 Unicode 文本的公开转换契约。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewPinyin_DefaultAndBoundaryInputs(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveText    string
		want        [][]string
	}{
		{
			name:        "success/default-normal-chinese",
			description: "验证默认 NewPinyin 实例使用 Normal 风格转换连续中文文本。",
			giveText:    "中国",
			want:        [][]string{{"zhong"}, {"guo"}},
		},
		{
			name:        "boundary/empty-string",
			description: "验证空字符串不会产生任何拼音结果。",
			giveText:    "",
			want:        [][]string{},
		},
		{
			name:        "boundary/non-chinese-ascii-and-digits",
			description: "验证纯 ASCII 字母和数字会被默认转换策略忽略。",
			giveText:    "abc123",
			want:        [][]string{},
		},
		{
			name:        "boundary/whitespace-only",
			description: "验证仅包含空白字符的输入不会产生拼音结果。",
			giveText:    " \t\n",
			want:        [][]string{},
		},
		{
			name:        "success/chinese-punctuation-is-ignored",
			description: "验证中文标点不会打断汉字转换，且不会作为结果项返回。",
			giveText:    "中，国。",
			want:        [][]string{{"zhong"}, {"guo"}},
		},
		{
			name:        "success/mixed-unicode-text-keeps-chinese-order",
			description: "验证 ASCII、数字和 emoji 混合文本中仅汉字被转换，并保持原有汉字顺序。",
			giveText:    "Hi中国123😊",
			want:        [][]string{{"zhong"}, {"guo"}},
		},
	}

	converter := NewPinyin()
	require.NotNil(t, converter)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := converter.Pinyin(tt.giveText)

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestNewPinyin_StyleOptions 验证 WithStyle 对主要拼音风格的输出配置行为。
//
// 该测试通过表驱动用例覆盖 Normal、Tone、Tone2、Tone3、Initials、FirstLetter、Finals 及各韵母声调风格，
// 确保每种公开 style 常量都能传递到底层转换器并产生稳定的可观察输出。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewPinyin_StyleOptions(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveStyle   int
		want        [][]string
	}{
		{
			name:        "success/normal",
			description: "验证 Normal 风格返回不带声调的完整拼音。",
			giveStyle:   Normal,
			want:        [][]string{{"zhong"}, {"guo"}},
		},
		{
			name:        "success/tone",
			description: "验证 Tone 风格返回带声调符号的完整拼音。",
			giveStyle:   Tone,
			want:        [][]string{{"zhōng"}, {"guó"}},
		},
		{
			name:        "success/tone2",
			description: "验证 Tone2 风格将声调数字放在韵母后。",
			giveStyle:   Tone2,
			want:        [][]string{{"zho1ng"}, {"guo2"}},
		},
		{
			name:        "success/tone3",
			description: "验证 Tone3 风格将声调数字放在拼音末尾。",
			giveStyle:   Tone3,
			want:        [][]string{{"zhong1"}, {"guo2"}},
		},
		{
			name:        "success/initials",
			description: "验证 Initials 风格仅返回每个汉字拼音的声母。",
			giveStyle:   Initials,
			want:        [][]string{{"zh"}, {"g"}},
		},
		{
			name:        "success/first-letter",
			description: "验证 FirstLetter 风格仅返回每个汉字拼音的首字母。",
			giveStyle:   FirstLetter,
			want:        [][]string{{"z"}, {"g"}},
		},
		{
			name:        "success/finals",
			description: "验证 Finals 风格仅返回不带声调的韵母。",
			giveStyle:   Finals,
			want:        [][]string{{"ong"}, {"uo"}},
		},
		{
			name:        "success/finals-tone",
			description: "验证 FinalsTone 风格返回带声调符号的韵母。",
			giveStyle:   FinalsTone,
			want:        [][]string{{"ōng"}, {"uó"}},
		},
		{
			name:        "success/finals-tone2",
			description: "验证 FinalsTone2 风格将声调数字放在韵母内部约定位置。",
			giveStyle:   FinalsTone2,
			want:        [][]string{{"o1ng"}, {"uo2"}},
		},
		{
			name:        "success/finals-tone3",
			description: "验证 FinalsTone3 风格将声调数字放在韵母末尾。",
			giveStyle:   FinalsTone3,
			want:        [][]string{{"ong1"}, {"uo2"}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			converter := NewPinyin(WithStyle(tt.giveStyle))
			require.NotNil(t, converter)

			got := converter.Pinyin("中国")

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestNewPinyin_HeteronymOption 验证 WithHeteronym 对多音字候选结果的开关行为。
//
// 该测试通过表驱动用例覆盖关闭和开启多音字两种配置，断言关闭时只返回默认读音，开启时返回多个候选和关键读音，
// 避免将测试耦合到底层词典的完整候选顺序或候选总数。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewPinyin_HeteronymOption(t *testing.T) {
	tests := []struct {
		name          string
		description   string
		giveHeteronym bool
		wantExact     []string
		wantContains  []string
		wantMultiple  bool
	}{
		{
			name:          "success/heteronym-disabled",
			description:   "验证关闭多音字时仅返回默认读音。",
			giveHeteronym: false,
			wantExact:     []string{"zhong"},
		},
		{
			name:          "success/heteronym-enabled",
			description:   "验证开启多音字时返回多个候选读音，并包含重字的关键读音。",
			giveHeteronym: true,
			wantContains:  []string{"zhong", "chong"},
			wantMultiple:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			converter := NewPinyin(WithHeteronym(tt.giveHeteronym))
			require.NotNil(t, converter)

			got := converter.Pinyin("重")

			require.Len(t, got, 1)
			if tt.wantExact != nil {
				assert.Equal(t, tt.wantExact, got[0])
			}
			if tt.wantMultiple {
				assert.Greater(t, len(got[0]), 1)
			}
			for _, want := range tt.wantContains {
				assert.Contains(t, got[0], want)
			}
		})
	}
}

// TestNewPinyin_OptionOrderAndInstanceIsolation 验证 Option 应用顺序和实例配置隔离行为。
//
// 该测试通过表驱动用例覆盖后传入的 WithStyle 覆盖先前配置，以及不同 Pinyin 实例之间不共享可变配置，
// 确保配置组合的公开行为稳定且互不污染。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewPinyin_OptionOrderAndInstanceIsolation(t *testing.T) {
	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T)
	}{
		{
			name:        "success/later-style-overrides-earlier-style",
			description: "验证多个 WithStyle 按传入顺序应用，后传入的 style 覆盖先前 style。",
			assert: func(t *testing.T) {
				converter := NewPinyin(WithStyle(Tone), WithStyle(Normal))
				require.NotNil(t, converter)

				got := converter.Pinyin("中国")

				assert.Equal(t, [][]string{{"zhong"}, {"guo"}}, got)
			},
		},
		{
			name:        "success/instances-keep-independent-styles",
			description: "验证不同 NewPinyin 实例分别保留自身 style 配置，不会发生实例间污染。",
			assert: func(t *testing.T) {
				toneConverter := NewPinyin(WithStyle(Tone))
				normalConverter := NewPinyin(WithStyle(Normal))
				require.NotNil(t, toneConverter)
				require.NotNil(t, normalConverter)

				gotTone := toneConverter.Pinyin("中国")
				gotNormal := normalConverter.Pinyin("中国")

				assert.Equal(t, [][]string{{"zhōng"}, {"guó"}}, gotTone)
				assert.Equal(t, [][]string{{"zhong"}, {"guo"}}, gotNormal)
			},
		},
		{
			name:        "success/instances-keep-independent-heteronym-settings",
			description: "验证不同 NewPinyin 实例分别保留自身多音字配置，不会发生实例间污染。",
			assert: func(t *testing.T) {
				heteronymConverter := NewPinyin(WithHeteronym(true))
				defaultConverter := NewPinyin()
				require.NotNil(t, heteronymConverter)
				require.NotNil(t, defaultConverter)

				gotHeteronym := heteronymConverter.Pinyin("重")
				gotDefault := defaultConverter.Pinyin("重")

				require.Len(t, gotHeteronym, 1)
				require.Len(t, gotDefault, 1)
				assert.Greater(t, len(gotHeteronym[0]), len(gotDefault[0]))
				assert.Equal(t, []string{"zhong"}, gotDefault[0])
				assert.Contains(t, gotHeteronym[0], "chong")
			},
		},
		{
			name:        "success/separator-does-not-change-pinyin-output",
			description: "验证 WithSeparator 只配置底层分隔符，不改变 Pinyin 方法按汉字返回二维切片的转换结果。",
			assert: func(t *testing.T) {
				converter := NewPinyin(WithSeparator("-"))
				require.NotNil(t, converter)

				got := converter.Pinyin("中国")

				assert.Equal(t, [][]string{{"zhong"}, {"guo"}}, got)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			tt.assert(t)
		})
	}
}
