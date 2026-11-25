package calendar

// 六十甲子纳音，按序号每两干支共用一音（甲子乙丑为第一组）。
var naYinNames = [30]string{
	"海中金", "炉中火", "大林木", "路旁土", "剑锋金",
	"山头火", "涧下水", "城头土", "白蜡金", "杨柳木",
	"泉中水", "屋上土", "霹雳火", "松柏木", "长流水",
	"沙中金", "山下火", "平地木", "壁上土", "金箔金",
	"覆灯火", "天河水", "大驿土", "钗钏金", "桑柘木",
	"大溪水", "沙中土", "天上火", "石榴木", "大海水",
}

// NaYin 返回干支的纳音五行名，如 "甲子" 返回 "海中金"。
// 非法干支返回空串。
func NaYin(gz GanZhi) string {
	idx := gz.Index()
	if idx < 0 {
		return ""
	}
	return naYinNames[idx/2]
}

// NaYinWuXing 返回纳音对应的五行。
// 纳音名末字即五行（金木水火土），如 "海中金" 为金。
func NaYinWuXing(gz GanZhi) (WuXing, bool) {
	name := NaYin(gz)
	if name == "" {
		return 0, false
	}
	runes := []rune(name)
	switch string(runes[len(runes)-1]) {
	case "金":
		return WuXingJin, true
	case "木":
		return WuXingMu, true
	case "水":
		return WuXingShui, true
	case "火":
		return WuXingHuo, true
	case "土":
		return WuXingTu, true
	}
	return 0, false
}
