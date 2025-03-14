// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package rand 提供了一系列随机数生成的工具函数，包括数值范围随机和中文字符随机生成。
package rand

import (
	"math/rand"
)

const (
	// minChinese 定义了常用汉字的 Unicode 最小值，用于生成随机汉字的范围下限。
	minChinese = 19968

	// maxChinese 定义了常用汉字的 Unicode 最大值，用于生成随机汉字的范围上限。
	maxChinese = 40869

	// lastNameString 包含了中国传统姓氏的字符集，包括单姓和复姓，按照传统顺序排列。
	// 此字符集收录了中国最常见的姓氏，用于随机生成具有真实性的中文姓氏。
	lastNameString = "赵钱孙李周吴郑王冯陈褚卫蒋沈韩杨朱秦尤许何吕施张孔曹严华金魏陶姜戚谢邹喻柏水窦章云苏潘葛奚范彭郎" +
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

// Int63n 生成一个范围在 [min, max) 之间的 int64 类型随机数。
//
// 参数：
//   - random：随机数生成器，如果为 nil 则使用默认的随机数生成器。
//   - min：随机数范围的最小值（包含）。
//   - max：随机数范围的最大值（不包含）。
//
// 返回值：
//   - int64：返回一个范围在 [min, max) 之间的 int64 类型随机数。
func Int63n(random *rand.Rand, min, max int64) int64 {
	// 计算随机数范围。
	result := max - min
	if nil == random {
		// 如果未提供随机数生成器，使用默认的随机数生成器。
		result = rand.Int63n(result)
	} else {
		// 使用提供的随机数生成器。
		result = random.Int63n(result)
	}

	// 将结果调整到指定范围内。
	result = result + min

	return result
}

// Intn 生成一个范围在 [min, max) 之间的 int 类型随机数。
//
// 参数：
//   - random：随机数生成器，如果为 nil 则使用默认的随机数生成器。
//   - min：随机数范围的最小值（包含）。
//   - max：随机数范围的最大值（不包含）。
//
// 返回值：
//   - int：返回一个范围在 [min, max) 之间的 int 类型随机数。
func Intn(random *rand.Rand, min, max int) int {
	// 计算随机数范围。
	result := max - min
	if nil == random {
		// 如果未提供随机数生成器，使用默认的随机数生成器。
		result = rand.Intn(result)
	} else {
		// 使用提供的随机数生成器。
		result = random.Intn(result)
	}

	// 将结果调整到指定范围内。
	result = result + min

	return result
}

// Chinese 生成一个随机的汉字字符串。
//
// 参数：
//   - random：随机数生成器，如果为 nil 则使用默认的随机数生成器。
//
// 返回值：
//   - string：返回一个随机生成的汉字字符串。
func Chinese(random *rand.Rand) string {
	// 生成一个在常用汉字 Unicode 范围内的随机数。
	r := rune(Intn(random, minChinese, maxChinese))
	// 将 Unicode 码点转换为字符串。
	s := string(r)
	return s
}

// ChineseLastName 生成一个随机的中文姓氏。
//
// 参数：
//   - random：随机数生成器，如果为 nil 则使用默认的随机数生成器。
//
// 返回值：
//   - string：返回一个随机选择的中文姓氏字符串。
func ChineseLastName(random *rand.Rand) string {
	// 将姓氏字符串转换为 rune 切片。
	var r = []rune(lastNameString)
	// 随机选择一个姓氏的索引。
	idx := Intn(random, 0, len(r))
	// 获取选中的姓氏字符。
	lastName := r[idx : idx+1]
	// 将 rune 切片转换为字符串。
	s := string(lastName)
	return s
}
