package stocksdk

// ============ API 端点常量 ============

const (
	// 腾讯财经 API
	TencentBaseURL    = "https://qt.gtimg.cn"
	TencentMinuteURL  = "https://web.ifzq.gtimg.cn/appstock/app/minute/query"

	// 股票代码列表
	AShareListURL     = "https://assets.linkdiary.cn/shares/zh_a_list.json"
	USListURL         = "https://assets.linkdiary.cn/shares/us_list.json"
	HKListURL         = "https://assets.linkdiary.cn/shares/hk_list.json"
	FundListURL       = "https://assets.linkdiary.cn/shares/fund_list"

	// A 股交易日历
	TradeCalendarURL  = "https://assets.linkdiary.cn/shares/trade-data-list.txt"

	// 东方财富 API
	EMKlineURL        = "https://push2his.eastmoney.com/api/qt/stock/kline/get"
	EMTrendsURL       = "https://push2his.eastmoney.com/api/qt/stock/trends2/get"
	EMHKKlineURL      = "https://33.push2his.eastmoney.com/api/qt/stock/kline/get"
	EMUSKlineURL      = "https://63.push2his.eastmoney.com/api/qt/stock/kline/get"

	// 东方财富备用端点（用于轮询/故障转移）
	EMKlineAltURL1 = "https://33.push2his.eastmoney.com/api/qt/stock/kline/get"
	EMKlineAltURL2 = "https://63.push2his.eastmoney.com/api/qt/stock/kline/get"
	EMKlineAltURL3 = "https://7.push2his.eastmoney.com/api/qt/stock/kline/get"
	EMKlineAltURL4 = "https://91.push2his.eastmoney.com/api/qt/stock/kline/get"

	// 东方财富行业板块 API
	EMBoardListURL    = "https://17.push2.eastmoney.com/api/qt/clist/get"
	EMBoardSpotURL    = "https://91.push2.eastmoney.com/api/qt/stock/get"
	EMBoardConsURL    = "https://29.push2.eastmoney.com/api/qt/clist/get"
	EMBoardKlineURL   = "https://7.push2his.eastmoney.com/api/qt/stock/kline/get"
	EMBoardTrendsURL  = "https://push2his.eastmoney.com/api/qt/stock/trends2/get"

	// 东方财富概念板块 API
	EMConceptListURL   = "https://79.push2.eastmoney.com/api/qt/clist/get"
	EMConceptSpotURL   = "https://91.push2.eastmoney.com/api/qt/stock/get"
	EMConceptConsURL   = "https://29.push2.eastmoney.com/api/qt/clist/get"
	EMConceptKlineURL  = "https://91.push2his.eastmoney.com/api/qt/stock/kline/get"
	EMConceptTrendsURL = "https://push2his.eastmoney.com/api/qt/stock/trends2/get"

	// 东方财富数据中心 API
	EMDatacenterURL    = "https://datacenter-web.eastmoney.com/api/data/v1/get"

	// 东方财富期货 API
	EMFuturesKlineURL         = "https://push2his.eastmoney.com/api/qt/stock/kline/get"
	EMFuturesGlobalSpotURL    = "https://futsseapi.eastmoney.com/list/COMEX,NYMEX,COBOT,SGX,NYBOT,LME,MDEX,TOCOM,IPE"
	EMFuturesGlobalSpotToken  = "58b2fa8f54638b60b87d69b31969089c"
)

// FuturesExchangeMap 国内期货交易所 market code 映射
var FuturesExchangeMap = map[string]int{
	"SHFE":  113,
	"DCE":   114,
	"CZCE":  115,
	"INE":   142,
	"CFFEX": 220,
	"GFEX":  225,
}

// FuturesVarietyExchangeMap 品种代码 -> 交易所映射
var FuturesVarietyExchangeMap = map[string]string{
	// SHFE 上海期货交易所
	"cu": "SHFE", "al": "SHFE", "zn": "SHFE", "pb": "SHFE", "au": "SHFE", "ag": "SHFE",
	"rb": "SHFE", "wr": "SHFE", "fu": "SHFE", "ru": "SHFE", "bu": "SHFE", "hc": "SHFE",
	"ni": "SHFE", "sn": "SHFE", "sp": "SHFE", "ss": "SHFE", "ao": "SHFE", "br": "SHFE",
	// DCE 大连商品交易所
	"c": "DCE", "a": "DCE", "b": "DCE", "m": "DCE", "y": "DCE", "p": "DCE",
	"l": "DCE", "v": "DCE", "j": "DCE", "jm": "DCE", "i": "DCE", "jd": "DCE",
	"pp": "DCE", "cs": "DCE", "eg": "DCE", "eb": "DCE", "pg": "DCE", "lh": "DCE",
	// CZCE 郑州商品交易所
	"WH": "CZCE", "CF": "CZCE", "SR": "CZCE", "TA": "CZCE", "OI": "CZCE", "MA": "CZCE",
	"FG": "CZCE", "RM": "CZCE", "SF": "CZCE", "SM": "CZCE", "ZC": "CZCE", "AP": "CZCE",
	"CJ": "CZCE", "UR": "CZCE", "SA": "CZCE", "PF": "CZCE", "PK": "CZCE", "PX": "CZCE",
	"SH": "CZCE",
	// INE 上海国际能源交易中心
	"sc": "INE", "nr": "INE", "lu": "INE", "bc": "INE", "ec": "INE",
	// CFFEX 中国金融期货交易所
	"IF": "CFFEX", "IC": "CFFEX", "IH": "CFFEX", "IM": "CFFEX",
	"TS": "CFFEX", "TF": "CFFEX", "T": "CFFEX", "TL": "CFFEX",
	// GFEX 广州期货交易所
	"si": "GFEX", "lc": "GFEX", "ps": "GFEX", "pt": "GFEX", "pd": "GFEX",
}

// GlobalFuturesMarketMap 全球期货市场代码映射
var GlobalFuturesMarketMap = map[string]int{
	"HG": 101, "GC": 101, "SI": 101, "QI": 101, "QO": 101, "MGC": 101,
	"CL": 102, "NG": 102, "RB": 102, "HO": 102, "PA": 102, "PL": 102,
	"ZW": 103, "ZM": 103, "ZS": 103, "ZC": 103, "ZL": 103, "ZR": 103,
	"YM": 103, "NQ": 103, "ES": 103,
	"SB": 108, "CT": 108,
	"LCPT": 109, "LZNT": 109, "LALT": 109,
}

// ============ 东方财富 K 线相关常量 ============

// GetMarketCode 根据股票代码获取东方财富市场代码
// 支持带前缀(sh/sz/bj)或纯代码
func GetMarketCode(symbol string) string {
	// 如果有前缀，直接根据前缀判断
	if len(symbol) >= 2 {
		prefix := symbol[:2]
		if prefix == "sh" || prefix == "SH" {
			return "1"
		}
		if prefix == "sz" || prefix == "SZ" || prefix == "bj" || prefix == "BJ" {
			return "0"
		}
	}
	// 纯代码：6 开头为上海(1)，其他为深圳/北交所(0)
	if len(symbol) >= 1 && symbol[0] == '6' {
		return "1"
	}
	return "0"
}

// GetPeriodCode 获取 K 线周期代码
func GetPeriodCode(period KlinePeriod) string {
	switch period {
	case KlinePeriodDaily:
		return "101"
	case KlinePeriodWeekly:
		return "102"
	case KlinePeriodMonthly:
		return "103"
	default:
		return "101"
	}
}

// GetAdjustCode 获取复权类型代码
func GetAdjustCode(adjust AdjustType) string {
	switch adjust {
	case AdjustTypeNone:
		return "0"
	case AdjustTypeQFQ:
		return "1"
	case AdjustTypeHFQ:
		return "2"
	default:
		return "0"
	}
}

// ============ 默认配置 ============

const (
	DefaultTimeout         = 30000 // 30秒
	DefaultMaxRetries      = 3
	DefaultBaseDelay       = 1000  // 1秒
	DefaultMaxDelay       = 30000 // 30秒
	DefaultBackoffMultiplier = 2
	DefaultBatchSize      = 500
	MaxBatchSize          = 500
	DefaultConcurrency    = 7
)

// ============ 可重试的 HTTP 状态码 ============

var DefaultRetryableStatusCodes = []int{408, 429, 500, 502, 503, 504}
