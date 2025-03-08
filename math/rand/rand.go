// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package rand

import (
	"math/rand"
)

const (
	// minChinese 定义了常用汉字的 Unicode 最小值，用于生成随机汉字的范围下限。
	minChinese = 19968

	// maxChinese 定义了常用汉字的 Unicode 最大值，用于生成随机汉字的范围上限。
	maxChinese = 40869

	// lastNameString 包含了中国常见的姓氏汉字集合，用于随机生成中文姓氏。
	// 包括单姓和复姓，按照传统顺序排列。
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

// Int63n 生成一个范围在 [min, max) 的随机数。
func Int63n(random *rand.Rand, min, max int64) int64 {
	result := max - min
	if nil == random {
		result = rand.Int63n(result)
	} else {
		result = random.Int63n(result)
	}

	result = result + min

	return result
}

// Intn 生成一个范围在 [min, max) 的随机数。
func Intn(random *rand.Rand, min, max int) int {
	result := max - min
	if nil == random {
		result = rand.Intn(result)
	} else {
		result = random.Intn(result)
	}

	result = result + min

	return result
}

// Chinese 生成随机汉字。
func Chinese(random *rand.Rand) string {
	r := rune(Intn(random, minChinese, maxChinese))
	s := string(r)
	return s
}

// ChineseLastName 生成随机汉字的姓。
func ChineseLastName(random *rand.Rand) string {
	var r = []rune(lastNameString)
	idx := Intn(random, 0, len(r))
	lastName := r[idx : idx+1]
	s := string(lastName)
	return s
}
