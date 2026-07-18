package utils

var cnPinyinMap = map[rune]string{
	'张': "zhang", '三': "san", '李': "li", '四': "si", '王': "wang",
	'五': "wu", '赵': "zhao", '钱': "qian", '孙': "sun", '周': "zhou",
	'吴': "wu", '郑': "zheng", '冯': "feng", '陈': "chen", '褚': "chu",
	'卫': "wei", '蒋': "jiang", '沈': "shen", '韩': "han", '杨': "yang",
	'朱': "zhu", '秦': "qin", '尤': "you", '许': "xu", '何': "he",
	'吕': "lv", '施': "shi", '孔': "kong", '曹': "cao", '严': "yan",
	'华': "hua", '金': "jin", '魏': "wei", '陶': "tao", '姜': "jiang",
	'戚': "qi", '谢': "xie", '邹': "zou", '苏': "su", '潘': "pan",
	'葛': "ge", '范': "fan", '彭': "peng", '鲁': "lu", '马': "ma",
	'方': "fang", '袁': "yuan", '柳': "liu", '唐': "tang", '史': "shi",
	'费': "fei", '薛': "xue", '雷': "lei", '贺': "he", '倪': "ni",
	'汤': "tang", '罗': "luo", '郝': "hao", '安': "an", '常': "chang",
	'乐': "le", '于': "yu", '时': "shi", '傅': "fu", '齐': "qi",
	'康': "kang", '余': "yu", '顾': "gu", '孟': "meng", '黄': "huang",
	'穆': "mu", '尹': "yin", '姚': "yao", '汪': "wang", '毛': "mao",
	'戴': "dai", '宋': "song", '庞': "pang", '梁': "liang", '董': "dong",
	'贾': "jia", '郭': "guo", '林': "lin", '钟': "zhong", '徐': "xu",
	'高': "gao", '蔡': "cai", '田': "tian", '樊': "fan", '胡': "hu",
	'万': "wan", '卢': "lu", '管': "guan", '叶': "ye", '刘': "liu",
	'白': "bai", '文': "wen", '关': "guan", '邓': "deng", '曾': "zeng",
	'任': "ren", '龙': "long", '武': "wu", '段': "duan", '乔': "qiao",
	'丁': "ding", '肖': "xiao", '明': "ming", '杰': "jie", '伟': "wei",
	'强': "qiang", '娜': "na", '敏': "min", '静': "jing", '丽': "li",
	'芳': "fang", '玲': "ling", '婷': "ting", '阳': "yang", '洋': "yang",
	'勇': "yong", '军': "jun", '磊': "lei", '峰': "feng", '鑫': "xin",
	'鹏': "peng", '浩': "hao", '然': "ran", '博': "bo",
	'辉': "hui", '超': "chao", '瑞': "rui", '凯': "kai", '宇': "yu",
	'翔': "xiang", '飞': "fei", '斌': "bin", '旭': "xu", '志': "zhi",
	'海': "hai", '涛': "tao", '波': "bo", '建': "jian", '国': "guo",
	'平': "ping", '庆': "qing", '永': "yong", '健': "jian",
	'春': "chun", '秋': "qiu", '冬': "dong", '雪': "xue", '云': "yun",
	'霞': "xia", '虹': "hong", '燕': "yan", '凤': "feng",
	'梅': "mei", '兰': "lan", '竹': "zhu", '菊': "ju", '松': "song",
	'柏': "bai", '红': "hong", '丹': "dan", '青': "qing", '翠': "cui",
	'玉': "yu", '珍': "zhen", '珠': "zhu", '宝': "bao", '秀': "xiu",
	'英': "ying", '雄': "xiong", '豪': "hao", '俊': "jun",
	'帅': "shuai", '才': "cai", '智': "zhi", '慧': "hui", '德': "de",
	'仁': "ren", '义': "yi", '礼': "li", '信': "xin", '忠': "zhong",
	'孝': "xiao", '诚': "cheng", '善': "shan", '良': "liang", '美': "mei",
	'佳': "jia", '欣': "xin", '怡': "yi", '悦': "yue", '欢': "huan",
	'梦': "meng", '思': "si", '念': "nian", '爱': "ai",
	'小': "xiao", '大': "da", '一': "yi", '二': "er", '六': "liu",
	'七': "qi", '八': "ba", '九': "jiu", '十': "shi", '百': "bai",
	'千': "qian", '东': "dong", '南': "nan", '西': "xi", '北': "bei",
	'中': "zhong", '新': "xin", '胜': "sheng", '利': "li",
	'星': "xing", '光': "guang", '天': "tian", '地': "di", '日': "ri",
	'月': "yue", '风': "feng", '雨': "yu", '电': "dian",
	'和': "he", '书': "shu", '学': "xue", '吉': "ji", '祥': "xiang",
	'如': "ru", '意': "yi", '程': "cheng", '远': "yuan", '宁': "ning",
	'晨': "chen", '宏': "hong", '凡': "fan", '颖': "ying", '亮': "liang",
	'源': "yuan", '成': "cheng", '立': "li", '发': "fa", '兴': "xing",
	'易': "yi", '盛': "sheng", '通': "tong", '达': "da", '顺': "shun",
	'昌': "chang", '富': "fu", '贵': "gui", '邦': "bang",
	'恒': "heng", '延': "yan", '弘': "hong", '振': "zhen", '启': "qi",
	'开': "kai", '承': "cheng", '继': "ji", '世': "shi", '家': "jia",
	'宗': "zong", '泽': "ze", '润': "run", '泉': "quan", '江': "jiang",
	'河': "he", '湖': "hu", '清': "qing", '淑': "shu", '雅': "ya",
	'芬': "fen", '茹': "ru", '若': "ruo", '薇': "wei", '萍': "ping",
	'政': "zheng", '哲': "zhe", '维': "wei",
	'璇': "xuan", '翰': "han", '瑶': "yao", '珑': "long",
}

func ConvertCNToPinyin(name string) string {
	if name == "" {
		return ""
	}
	result := make([]byte, 0, len(name)*4)
	for _, r := range name {
		if pinyin, ok := cnPinyinMap[r]; ok {
			result = append(result, pinyin...)
		}
	}
	if len(result) == 0 {
		return ""
	}
	return string(result)
}
