// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package rand_test 提供 rand 包公开随机工具函数的单元测试。
//
// 测试通过固定种子的自定义随机源验证确定性输出，并通过 nil 随机源用例覆盖默认随机源分支的范围、
// UTF-8、rune 与集合成员语义。
package rand_test

import (
	"math/rand"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kitrand "github.com/fsyyft-go/kit/math/rand"
)

const (
	minChineseRune = 19968
	maxChineseRune = 40869
	// testLastNames 有意镜像生产代码中的姓氏字符集，用于验证 ChineseLastName 的公开集合契约。
	testLastNames = "赵钱孙李周吴郑王冯陈褚卫蒋沈韩杨朱秦尤许何吕施张孔曹严华金魏陶姜戚谢邹喻柏水窦章云苏潘葛奚范彭郎" +
		"鲁韦昌马苗凤花方俞任袁柳酆鲍史唐费廉岑薛雷贺倪汤滕殷罗毕郝邬安常乐于时傅皮卞齐康伍余元卜顾孟平黄" +
		"和穆萧尹姚邵湛汪祁毛禹狄米贝明臧计伏成戴谈宋茅庞熊纪舒屈项祝董梁杜阮蓝闵席季麻强贾路娄危江童颜郭" +
		"梅盛林刁锺徐邱骆高夏蔡田樊胡凌霍虞万支柯昝管卢莫经房裘缪干解应宗丁宣贲邓郁单杭洪包诸左石崔吉钮龚" +
		"程嵇邢滑裴陆荣翁荀羊於惠甄麹家封芮羿储靳汲邴糜松井段富巫乌焦巴弓牧隗山谷车侯宓蓬全郗班仰秋仲伊宫" +
		"甯仇栾暴甘钭厉戎祖武符刘景詹束龙叶幸司韶郜黎蓟薄印宿白怀蒲邰从鄂索咸籍赖卓蔺屠蒙池乔阴鬱胥能苍双" +
		"闻莘党翟谭贡劳逄姬申扶堵冉宰郦雍郤璩桑桂濮牛寿通边扈燕冀郏浦尚农温别庄晏柴瞿阎充慕连茹习宦艾鱼容" +
		"向古易慎戈廖庾终暨居衡步都耿满弘匡国文寇广禄阙东欧殳沃利蔚越夔隆师巩厍聂晁勾敖融冷訾辛阚那简饶空" +
		"曾毋沙乜养鞠须丰巢关蒯相查后荆红游竺权逯盖益桓公万俟司马上官欧阳夏侯诸葛闻人东方赫连皇甫尉迟公羊" +
		"澹台公冶宗政濮阳淳于单于太叔申屠公孙仲孙轩辕令狐锺离宇文长孙慕容鲜于闾丘司徒司空亓官司寇仉督子车" +
		"颛孙端木巫马公西漆雕乐正壤驷公良拓跋夹谷宰父穀梁晋楚闫法汝鄢涂钦段干百里东郭南门呼延归海羊舌微生" +
		"岳帅缑亢况後有琴梁丘左丘东门西门商牟佘佴伯赏南宫墨哈谯笪年爱阳佟"
)

// TestInt63n_CustomRandomDeterministic 验证 Int63n 使用自定义随机源时的确定性范围映射。
//
// 该测试通过表驱动用例覆盖正数范围、跨零范围、单值范围和大数边界，确保固定种子的随机序列被稳定平移到 [min, max) 区间。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestInt63n_CustomRandomDeterministic(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveSeed    int64
		giveMin     int64
		giveMax     int64
		want        int64
	}{
		{
			name:        "success/positive-range",
			description: "验证 Int63n 使用固定种子的自定义随机源时，将结果稳定映射到正数半开区间。",
			giveSeed:    11,
			giveMin:     0,
			giveMax:     100,
			want:        65,
		},
		{
			name:        "success/cross-zero-range",
			description: "验证 Int63n 使用固定种子的自定义随机源时，将结果稳定映射到跨零半开区间。",
			giveSeed:    12,
			giveMin:     -5,
			giveMax:     5,
			want:        -3,
		},
		{
			name:        "boundary/single-value-range",
			description: "验证 Int63n 在仅包含一个可取值的半开区间内始终返回下界。",
			giveSeed:    13,
			giveMin:     42,
			giveMax:     43,
			want:        42,
		},
		{
			name:        "boundary/large-int64-range",
			description: "验证 Int63n 在较大的 int64 边界区间内仍能稳定平移随机偏移量。",
			giveSeed:    14,
			giveMin:     1 << 40,
			giveMax:     (1 << 40) + 256,
			want:        1099511628009,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			random := newTestRand(tt.giveSeed)
			got := kitrand.Int63n(random, tt.giveMin, tt.giveMax)

			assert.Equal(t, tt.want, got)
			assert.GreaterOrEqual(t, got, tt.giveMin)
			assert.Less(t, got, tt.giveMax)
		})
	}
}

// TestInt63n_NilRandomRange 验证 Int63n 使用默认随机源时的范围契约。
//
// 该测试覆盖 nil random 分支在正数、跨零和单值边界区间内的半开区间语义，避免依赖 Go 运行时默认随机源的具体种子。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestInt63n_NilRandomRange(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveMin     int64
		giveMax     int64
		giveTimes   int
	}{
		{
			name:        "success/positive-range",
			description: "验证 Int63n 使用默认随机源时返回值始终位于正数半开区间。",
			giveMin:     0,
			giveMax:     100,
			giveTimes:   64,
		},
		{
			name:        "success/cross-zero-range",
			description: "验证 Int63n 使用默认随机源时返回值始终位于跨零半开区间。",
			giveMin:     -1,
			giveMax:     1,
			giveTimes:   64,
		},
		{
			name:        "boundary/single-value-range",
			description: "验证 Int63n 使用默认随机源时在单值区间内稳定返回下界。",
			giveMin:     7,
			giveMax:     8,
			giveTimes:   8,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			for i := 0; i < tt.giveTimes; i++ {
				got := kitrand.Int63n(nil, tt.giveMin, tt.giveMax)

				assert.GreaterOrEqual(t, got, tt.giveMin)
				assert.Less(t, got, tt.giveMax)
			}
		})
	}
}

// TestInt63n_Panic 验证 Int63n 在非法区间上保持 math/rand 的 panic 契约。
//
// 该测试覆盖 nil random 与自定义 random 两条分支，并验证 max <= min 时底层随机函数会拒绝非正范围。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestInt63n_Panic(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveRandom  *rand.Rand
		giveMin     int64
		giveMax     int64
	}{
		{
			name:        "panic/nil-random-equal-bound",
			description: "验证 Int63n 使用默认随机源且 max 等于 min 时触发 panic。",
			giveRandom:  nil,
			giveMin:     3,
			giveMax:     3,
		},
		{
			name:        "panic/nil-random-inverted-bound",
			description: "验证 Int63n 使用默认随机源且 max 小于 min 时触发 panic。",
			giveRandom:  nil,
			giveMin:     4,
			giveMax:     3,
		},
		{
			name:        "panic/custom-random-equal-bound",
			description: "验证 Int63n 使用自定义随机源且 max 等于 min 时触发 panic。",
			giveRandom:  newTestRand(15),
			giveMin:     3,
			giveMax:     3,
		},
		{
			name:        "panic/custom-random-inverted-bound",
			description: "验证 Int63n 使用自定义随机源且 max 小于 min 时触发 panic。",
			giveRandom:  newTestRand(16),
			giveMin:     4,
			giveMax:     3,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			require.Panics(t, func() {
				_ = kitrand.Int63n(tt.giveRandom, tt.giveMin, tt.giveMax)
			})
		})
	}
}

// TestIntn_CustomRandomDeterministic 验证 Intn 使用自定义随机源时的确定性范围映射。
//
// 该测试通过表驱动用例覆盖正数范围、跨零范围、单值范围和较大 int 区间，确保固定种子的随机序列被稳定平移到 [min, max) 区间。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestIntn_CustomRandomDeterministic(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveSeed    int64
		giveMin     int
		giveMax     int
		want        int
	}{
		{
			name:        "success/positive-range",
			description: "验证 Intn 使用固定种子的自定义随机源时，将结果稳定映射到正数半开区间。",
			giveSeed:    21,
			giveMin:     0,
			giveMax:     100,
			want:        0,
		},
		{
			name:        "success/cross-zero-range",
			description: "验证 Intn 使用固定种子的自定义随机源时，将结果稳定映射到跨零半开区间。",
			giveSeed:    22,
			giveMin:     -5,
			giveMax:     5,
			want:        4,
		},
		{
			name:        "boundary/single-value-range",
			description: "验证 Intn 在仅包含一个可取值的半开区间内始终返回下界。",
			giveSeed:    23,
			giveMin:     42,
			giveMax:     43,
			want:        42,
		},
		{
			name:        "boundary/large-int-range",
			description: "验证 Intn 在较大的 int 边界区间内仍能稳定平移随机偏移量。",
			giveSeed:    24,
			giveMin:     1 << 20,
			giveMax:     (1 << 20) + 128,
			want:        1048617,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			random := newTestRand(tt.giveSeed)
			got := kitrand.Intn(random, tt.giveMin, tt.giveMax)

			assert.Equal(t, tt.want, got)
			assert.GreaterOrEqual(t, got, tt.giveMin)
			assert.Less(t, got, tt.giveMax)
		})
	}
}

// TestIntn_NilRandomRange 验证 Intn 使用默认随机源时的范围契约。
//
// 该测试覆盖 nil random 分支在正数、跨零和单值边界区间内的半开区间语义，避免依赖 Go 运行时默认随机源的具体种子。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestIntn_NilRandomRange(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveMin     int
		giveMax     int
		giveTimes   int
	}{
		{
			name:        "success/positive-range",
			description: "验证 Intn 使用默认随机源时返回值始终位于正数半开区间。",
			giveMin:     0,
			giveMax:     100,
			giveTimes:   64,
		},
		{
			name:        "success/cross-zero-range",
			description: "验证 Intn 使用默认随机源时返回值始终位于跨零半开区间。",
			giveMin:     -1,
			giveMax:     1,
			giveTimes:   64,
		},
		{
			name:        "boundary/single-value-range",
			description: "验证 Intn 使用默认随机源时在单值区间内稳定返回下界。",
			giveMin:     7,
			giveMax:     8,
			giveTimes:   8,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			for i := 0; i < tt.giveTimes; i++ {
				got := kitrand.Intn(nil, tt.giveMin, tt.giveMax)

				assert.GreaterOrEqual(t, got, tt.giveMin)
				assert.Less(t, got, tt.giveMax)
			}
		})
	}
}

// TestIntn_Panic 验证 Intn 在非法区间上保持 math/rand 的 panic 契约。
//
// 该测试覆盖 nil random 与自定义 random 两条分支，并验证 max <= min 时底层随机函数会拒绝非正范围。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestIntn_Panic(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveRandom  *rand.Rand
		giveMin     int
		giveMax     int
	}{
		{
			name:        "panic/nil-random-equal-bound",
			description: "验证 Intn 使用默认随机源且 max 等于 min 时触发 panic。",
			giveRandom:  nil,
			giveMin:     3,
			giveMax:     3,
		},
		{
			name:        "panic/nil-random-inverted-bound",
			description: "验证 Intn 使用默认随机源且 max 小于 min 时触发 panic。",
			giveRandom:  nil,
			giveMin:     4,
			giveMax:     3,
		},
		{
			name:        "panic/custom-random-equal-bound",
			description: "验证 Intn 使用自定义随机源且 max 等于 min 时触发 panic。",
			giveRandom:  newTestRand(25),
			giveMin:     3,
			giveMax:     3,
		},
		{
			name:        "panic/custom-random-inverted-bound",
			description: "验证 Intn 使用自定义随机源且 max 小于 min 时触发 panic。",
			giveRandom:  newTestRand(26),
			giveMin:     4,
			giveMax:     3,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			require.Panics(t, func() {
				_ = kitrand.Intn(tt.giveRandom, tt.giveMin, tt.giveMax)
			})
		})
	}
}

// TestChinese_CustomRandomDeterministic 验证 Chinese 使用自定义随机源时生成确定的单个汉字。
//
// 该测试通过固定种子覆盖多个确定性输出，并同时断言结果是有效 UTF-8、单 rune、Han 字符且位于约定 Unicode 半开区间。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestChinese_CustomRandomDeterministic(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveSeed    int64
		want        string
		wantRune    rune
	}{
		{
			name:        "success/fixed-seed-31",
			description: "验证 Chinese 使用固定种子 31 时返回确定的单个汉字。",
			giveSeed:    31,
			want:        "鐜",
			wantRune:    '鐜',
		},
		{
			name:        "success/fixed-seed-32",
			description: "验证 Chinese 使用固定种子 32 时返回确定的单个汉字。",
			giveSeed:    32,
			want:        "赖",
			wantRune:    '赖',
		},
		{
			name:        "success/fixed-seed-33",
			description: "验证 Chinese 使用固定种子 33 时返回确定的单个汉字。",
			giveSeed:    33,
			want:        "焣",
			wantRune:    '焣',
		},
		{
			name:        "success/fixed-seed-34",
			description: "验证 Chinese 使用固定种子 34 时返回确定的单个汉字。",
			giveSeed:    34,
			want:        "抬",
			wantRune:    '抬',
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := kitrand.Chinese(newTestRand(tt.giveSeed))

			assert.Equal(t, tt.want, got)
			assertChineseRune(t, got, tt.wantRune)
		})
	}
}

// TestChinese_NilRandomRuneContract 验证 Chinese 使用默认随机源时的汉字 rune 契约。
//
// 该测试覆盖 nil random 分支，断言返回值始终是有效 UTF-8 编码的单个 Han rune，并位于公开函数约定的常用汉字范围内。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestChinese_NilRandomRuneContract(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveTimes   int
	}{
		{
			name:        "success/default-random-rune-contract",
			description: "验证 Chinese 使用默认随机源时多次返回值均满足 UTF-8、单 rune 与 Han 范围契约。",
			giveTimes:   64,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			for i := 0; i < tt.giveTimes; i++ {
				got := kitrand.Chinese(nil)

				assertChineseRune(t, got, 0)
			}
		})
	}
}

// TestChineseLastName_CustomRandomDeterministic 验证 ChineseLastName 使用自定义随机源时生成确定的姓氏字符。
//
// 该测试通过固定种子覆盖多个确定性输出，并同时断言结果是有效 UTF-8、单 rune、Han 字符且属于内置姓氏集合。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestChineseLastName_CustomRandomDeterministic(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveSeed    int64
		want        string
		wantRune    rune
	}{
		{
			name:        "success/fixed-seed-41",
			description: "验证 ChineseLastName 使用固定种子 41 时返回确定的姓氏字符。",
			giveSeed:    41,
			want:        "干",
			wantRune:    '干',
		},
		{
			name:        "success/fixed-seed-42",
			description: "验证 ChineseLastName 使用固定种子 42 时返回确定的姓氏字符。",
			giveSeed:    42,
			want:        "盛",
			wantRune:    '盛',
		},
		{
			name:        "success/fixed-seed-43",
			description: "验证 ChineseLastName 使用固定种子 43 时返回确定的姓氏字符。",
			giveSeed:    43,
			want:        "弘",
			wantRune:    '弘',
		},
		{
			name:        "success/fixed-seed-44",
			description: "验证 ChineseLastName 使用固定种子 44 时返回确定的姓氏字符。",
			giveSeed:    44,
			want:        "伏",
			wantRune:    '伏',
		},
		{
			name:        "success/fixed-seed-45",
			description: "验证 ChineseLastName 使用固定种子 45 时返回确定的姓氏字符。",
			giveSeed:    45,
			want:        "胡",
			wantRune:    '胡',
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := kitrand.ChineseLastName(newTestRand(tt.giveSeed))

			assert.Equal(t, tt.want, got)
			assertLastNameRune(t, got, tt.wantRune)
		})
	}
}

// TestChineseLastName_NilRandomSetContract 验证 ChineseLastName 使用默认随机源时的姓氏集合契约。
//
// 该测试覆盖 nil random 分支，断言返回值始终是有效 UTF-8 编码的单个 Han rune，并属于公开函数使用的中文姓氏字符集合。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestChineseLastName_NilRandomSetContract(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveTimes   int
	}{
		{
			name:        "success/default-random-set-contract",
			description: "验证 ChineseLastName 使用默认随机源时多次返回值均满足 UTF-8、单 rune、Han 与姓氏集合契约。",
			giveTimes:   64,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			for i := 0; i < tt.giveTimes; i++ {
				got := kitrand.ChineseLastName(nil)

				assertLastNameRune(t, got, 0)
			}
		})
	}
}

// newTestRand 构造使用固定种子的随机数生成器。
//
// 该辅助函数集中创建可复现的随机源，确保依赖自定义 random 分支的测试不会受当前时间或全局随机源影响。
//
// 参数：
//   - seed: 用于初始化 math/rand Source 的固定种子。
//
// 返回：
//   - *rand.Rand: 使用指定种子初始化的随机数生成器。
func newTestRand(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

// assertChineseRune 断言字符串满足 Chinese 函数的单个汉字 rune 契约。
//
// 该辅助函数验证字符串为有效 UTF-8、仅包含一个 rune、属于 Han 字符集，并落在 Chinese 函数约定的 Unicode 半开区间内。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败并标记辅助函数调用栈。
//   - got: 被验证的字符串结果。
//   - wantRune: 期望 rune；传入 0 时仅验证通用汉字契约。
func assertChineseRune(t *testing.T, got string, wantRune rune) {
	t.Helper()

	runes := []rune(got)
	require.Len(t, runes, 1)
	assert.True(t, utf8.ValidString(got))

	gotRune := runes[0]
	if wantRune != 0 {
		assert.Equal(t, wantRune, gotRune)
	}
	assert.True(t, unicode.Is(unicode.Han, gotRune))
	assert.GreaterOrEqual(t, int(gotRune), minChineseRune)
	assert.Less(t, int(gotRune), maxChineseRune)
}

// assertLastNameRune 断言字符串满足 ChineseLastName 函数的单个中文姓氏 rune 契约。
//
// 该辅助函数验证字符串为有效 UTF-8、仅包含一个 rune、属于 Han 字符集，并存在于测试镜像的中文姓氏字符集合中。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败并标记辅助函数调用栈。
//   - got: 被验证的字符串结果。
//   - wantRune: 期望 rune；传入 0 时仅验证通用姓氏集合契约。
func assertLastNameRune(t *testing.T, got string, wantRune rune) {
	t.Helper()

	runes := []rune(got)
	require.Len(t, runes, 1)
	assert.True(t, utf8.ValidString(got))

	gotRune := runes[0]
	if wantRune != 0 {
		assert.Equal(t, wantRune, gotRune)
	}
	assert.True(t, unicode.Is(unicode.Han, gotRune))
	assert.True(t, strings.ContainsRune(testLastNames, gotRune))
}
