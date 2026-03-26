package stocksdk

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"strconv"
	"strings"
	"sync"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// ============ 全局缓存 ============

var (
	cachedAShareCodes          []string
	cachedAShareCodesNoExchange []string
	cachedTradeCalendar        []string
	cacheMutex                 sync.RWMutex
)

// ============ 腾讯财经响应解析器 ============

// DecodeGBK 将 GBK 编码的字节数组解码为 UTF-8 字符串
func DecodeGBK(data []byte) (string, error) {
	reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GBK.NewDecoder())
	result, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// ParseTencentResponse 解析腾讯财经响应文本
// 按 `;` 拆行，提取 `v_xxx="..."` 里的内容，返回 []TencentResponseItem
func ParseTencentResponse(text string) []TencentResponseItem {
	var results []TencentResponseItem

	// 按分号分割行
	lines := strings.Split(text, ";")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 查找等号位置
		eqIdx := strings.Index(line, "=")
		if eqIdx < 0 {
			continue
		}

		// 提取 key（去掉 v_ 前缀）
		key := strings.TrimSpace(line[:eqIdx])
		if strings.HasPrefix(key, "v_") {
			key = key[2:]
		}

		// 提取 raw 内容（去掉引号）
		raw := strings.TrimSpace(line[eqIdx+1:])
		if strings.HasPrefix(raw, "\"") && strings.HasSuffix(raw, "\"") {
			raw = raw[1 : len(raw)-1]
		}

		// 按 ~ 分割字段
		fields := strings.Split(raw, "~")

		results = append(results, TencentResponseItem{
			Key:    key,
			Fields: fields,
		})
	}

	return results
}

// TencentResponseItem 腾讯财经响应单项
type TencentResponseItem struct {
	Key    string
	Fields []string
}

// ============ 行情数据解析器 ============

// ParseFullQuote 解析 A 股全量行情
func ParseFullQuote(fields []string) FullQuote {
	// 买卖盘口 - 买盘
	bid := make([]BidAskItem, 0, 5)
	for i := 0; i < 5; i++ {
		priceIdx := 9 + i*2
		volumeIdx := 10 + i*2
		bid = append(bid, BidAskItem{
			Price:  safeNumber(getField(fields, priceIdx)),
			Volume: safeNumber(getField(fields, volumeIdx)),
		})
	}

	// 买卖盘口 - 卖盘
	ask := make([]BidAskItem, 0, 5)
	for i := 0; i < 5; i++ {
		priceIdx := 19 + i*2
		volumeIdx := 20 + i*2
		ask = append(ask, BidAskItem{
			Price:  safeNumber(getField(fields, priceIdx)),
			Volume: safeNumber(getField(fields, volumeIdx)),
		})
	}

	return FullQuote{
		MarketID:             getField(fields, 0),
		Name:                 getField(fields, 1),
		Code:                 getField(fields, 2),
		Price:                safeNumber(getField(fields, 3)),
		PrevClose:            safeNumber(getField(fields, 4)),
		Open:                 safeNumber(getField(fields, 5)),
		Volume:               safeNumber(getField(fields, 6)),
		OuterVolume:          safeNumber(getField(fields, 7)),
		InnerVolume:          safeNumber(getField(fields, 8)),
		Bid:                  bid,
		Ask:                  ask,
		Time:                 getField(fields, 30),
		Change:               safeNumber(getField(fields, 31)),
		ChangePercent:        safeNumber(getField(fields, 32)),
		High:                 safeNumber(getField(fields, 33)),
		Low:                  safeNumber(getField(fields, 34)),
		Volume2:              safeNumber(getField(fields, 36)),
		Amount:               safeNumber(getField(fields, 37)),
		TurnoverRate:         safeNumberOrNull(getField(fields, 38)),
		PE:                   safeNumberOrNull(getField(fields, 39)),
		Amplitude:            safeNumberOrNull(getField(fields, 43)),
		CirculatingMarketCap: safeNumberOrNull(getField(fields, 44)),
		TotalMarketCap:       safeNumberOrNull(getField(fields, 45)),
		PB:                   safeNumberOrNull(getField(fields, 46)),
		LimitUp:              safeNumberOrNull(getField(fields, 47)),
		LimitDown:            safeNumberOrNull(getField(fields, 48)),
		VolumeRatio:          safeNumberOrNull(getField(fields, 49)),
		AvgPrice:             safeNumberOrNull(getField(fields, 51)),
		PEStatic:             safeNumberOrNull(getField(fields, 52)),
		PEDynamic:            safeNumberOrNull(getField(fields, 53)),
		High52W:              safeNumberOrNull(getField(fields, 67)),
		Low52W:               safeNumberOrNull(getField(fields, 68)),
		CirculatingShares:    safeNumberOrNull(getField(fields, 72)),
		TotalShares:          safeNumberOrNull(getField(fields, 73)),
		Raw:                  fields,
	}
}

// ParseSimpleQuote 解析简要行情
func ParseSimpleQuote(fields []string) SimpleQuote {
	return SimpleQuote{
		MarketID:   getField(fields, 0),
		Name:       getField(fields, 1),
		Code:       getField(fields, 2),
		Price:      safeNumber(getField(fields, 3)),
		Change:     safeNumber(getField(fields, 4)),
		ChangePct:  safeNumber(getField(fields, 5)),
		Volume:     safeNumber(getField(fields, 6)),
		Amount:     safeNumber(getField(fields, 7)),
		MarketCap:  safeNumberOrNull(getField(fields, 9)),
		MarketType: getField(fields, 10),
		Raw:        fields,
	}
}

// ParseFundFlow 解析资金流向
func ParseFundFlow(fields []string) FundFlow {
	return FundFlow{
		Code:           getField(fields, 0),
		MainInflow:     safeNumber(getField(fields, 1)),
		MainOutflow:    safeNumber(getField(fields, 2)),
		MainNet:        safeNumber(getField(fields, 3)),
		MainNetRatio:   safeNumber(getField(fields, 4)),
		RetailInflow:   safeNumber(getField(fields, 5)),
		RetailOutflow:  safeNumber(getField(fields, 6)),
		RetailNet:      safeNumber(getField(fields, 7)),
		RetailNetRatio: safeNumber(getField(fields, 8)),
		TotalFlow:      safeNumber(getField(fields, 9)),
		Name:           getField(fields, 12),
		Date:           getField(fields, 13),
		Raw:            fields,
	}
}

// ============ 补充类型定义（用于解析器） ============

// PanelLargeOrder 盘口大单占比
type PanelLargeOrder struct {
	BuyLargeRatio  float64   `json:"buy_large_ratio"`
	BuySmallRatio  float64   `json:"buy_small_ratio"`
	SellLargeRatio float64   `json:"sell_large_ratio"`
	SellSmallRatio float64   `json:"sell_small_ratio"`
	Raw            []string `json:"raw"`
}

// HKQuote 港股行情
type HKQuote struct {
	MarketID             string   `json:"market_id"`
	Name                 string   `json:"name"`
	Code                 string   `json:"code"`
	Price                float64  `json:"price"`
	PrevClose            float64  `json:"prev_close"`
	Open                 float64  `json:"open"`
	Volume               float64  `json:"volume"`
	Time                 string   `json:"time"`
	Change               float64  `json:"change"`
	ChangePercent        float64  `json:"change_percent"`
	High                 float64  `json:"high"`
	Low                  float64  `json:"low"`
	Amount               float64  `json:"amount"`
	LotSize              *float64 `json:"lot_size"`
	CirculatingMarketCap *float64 `json:"circulating_market_cap"`
	TotalMarketCap       *float64 `json:"total_market_cap"`
	Currency             string   `json:"currency"`
	Raw                  []string `json:"raw"`
}

// USQuote 美股行情
type USQuote struct {
	MarketID      string   `json:"market_id"`
	Name          string   `json:"name"`
	Code          string   `json:"code"`
	Price         float64  `json:"price"`
	PrevClose     float64  `json:"prev_close"`
	Open          float64  `json:"open"`
	Volume        float64  `json:"volume"`
	Time          string   `json:"time"`
	Change        float64  `json:"change"`
	ChangePercent float64  `json:"change_percent"`
	High          float64  `json:"high"`
	Low           float64  `json:"low"`
	Amount        float64  `json:"amount"`
	TurnoverRate  *float64 `json:"turnover_rate"`
	PE            *float64 `json:"pe"`
	Amplitude     *float64 `json:"amplitude"`
	TotalMarketCap *float64 `json:"total_market_cap"`
	PB            *float64 `json:"pb"`
	High52W       *float64 `json:"high_52w"`
	Low52W        *float64 `json:"low_52w"`
	Raw           []string `json:"raw"`
}

// FundQuote 基金行情
type FundQuote struct {
	Code    string   `json:"code"`
	Name    string   `json:"name"`
	NAV     float64  `json:"nav"`
	AccNAV  float64  `json:"acc_nav"`
	Change  float64  `json:"change"`
	NAVDate string   `json:"nav_date"`
	Raw     []string `json:"raw"`
}

// ============ 东方财富 K 线类型 ============

// EmKlineResponse 东方财富 K 线 API 响应
type EmKlineResponse struct {
	Data *EmKlineData `json:"data"`
}

// EmKlineData 东方财富 K 线数据
type EmKlineData struct {
	Klines []string `json:"klines"`
	Name   string   `json:"name"`
	Code   string   `json:"code"`
}

// EmKlineItem 东方财富 K 线 CSV 解析后的数据
type EmKlineItem struct {
	Date          string   `json:"date"`
	Open          *float64 `json:"open"`
	Close         *float64 `json:"close"`
	High          *float64 `json:"high"`
	Low           *float64 `json:"low"`
	Volume        *float64 `json:"volume"`
	Amount        *float64 `json:"amount"`
	Amplitude     *float64 `json:"amplitude"`
	ChangePercent *float64 `json:"change_percent"`
	Change        *float64 `json:"change"`
	TurnoverRate  *float64 `json:"turnover_rate"`
}

// ParseEmKlineCsv 解析东方财富 K 线 CSV 行
func ParseEmKlineCsv(line string) EmKlineItem {
	parts := strings.Split(line, ",")
	var item EmKlineItem

	if len(parts) > 0 {
		item.Date = parts[0]
	}
	if len(parts) > 1 {
		item.Open = toNumberOrNull(parts[1])
	}
	if len(parts) > 2 {
		item.Close = toNumberOrNull(parts[2])
	}
	if len(parts) > 3 {
		item.High = toNumberOrNull(parts[3])
	}
	if len(parts) > 4 {
		item.Low = toNumberOrNull(parts[4])
	}
	if len(parts) > 5 {
		item.Volume = toNumberOrNull(parts[5])
	}
	if len(parts) > 6 {
		item.Amount = toNumberOrNull(parts[6])
	}
	if len(parts) > 7 {
		item.Amplitude = toNumberOrNull(parts[7])
	}
	if len(parts) > 8 {
		item.ChangePercent = toNumberOrNull(parts[8])
	}
	if len(parts) > 9 {
		item.Change = toNumberOrNull(parts[9])
	}
	if len(parts) > 10 {
		item.TurnoverRate = toNumberOrNull(parts[10])
	}

	return item
}

// toNumberOrNull 将字符串转换为 *float64，空值返回 nil
func toNumberOrNull(s string) *float64 {
	if s == "" || s == "-" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil
	}
	return &f
}

// ============ Linkdiary 数据解析器 ============

// ParseStockListResponse 解析股票列表 JSON 响应
func ParseStockListResponse(data []byte) (*StockListResponse, error) {
	var resp StockListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ParseTradeCalendar 解析交易日历文本（逗号分隔）
func ParseTradeCalendar(text string) []string {
	if text == "" || strings.TrimSpace(text) == "" {
		return []string{}
	}
	parts := strings.Split(strings.TrimSpace(text), ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// MatchMarket 根据市场类型筛选股票代码
// code: 带交易所前缀的股票代码，如 'sh600000'
func MatchMarket(code string, market AShareMarket) bool {
	// 提取纯数字代码
	pureCode := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(code, "sh", ""), "sz", ""), "bj", "")

	switch market {
	case AShareMarketSH:
		// 上交所：6 开头
		return strings.HasPrefix(pureCode, "6")
	case AShareMarketSZ:
		// 深交所：0 开头或 3 开头（包含创业板）
		return strings.HasPrefix(pureCode, "0") || strings.HasPrefix(pureCode, "3")
	case AShareMarketBJ:
		// 北交所：92 开头
		return strings.HasPrefix(pureCode, "92")
	case AShareMarketKC:
		// 科创板：688 开头
		return strings.HasPrefix(pureCode, "688")
	case AShareMarketCY:
		// 创业板：30 开头
		return strings.HasPrefix(pureCode, "30")
	default:
		return true
	}
}

// ParsePanelLargeOrder 解析盘口大单
func ParsePanelLargeOrder(fields []string) PanelLargeOrder {
	return PanelLargeOrder{
		BuyLargeRatio:  safeNumber(getField(fields, 0)),
		BuySmallRatio:  safeNumber(getField(fields, 1)),
		SellLargeRatio: safeNumber(getField(fields, 2)),
		SellSmallRatio: safeNumber(getField(fields, 3)),
		Raw:            fields,
	}
}

// ParseHKQuote 解析港股行情
func ParseHKQuote(fields []string) HKQuote {
	var currency string
	if len(fields) >= 3 {
		currency = getField(fields, len(fields)-3)
	}

	return HKQuote{
		MarketID:             getField(fields, 0),
		Name:                 getField(fields, 1),
		Code:                 getField(fields, 2),
		Price:                safeNumber(getField(fields, 3)),
		PrevClose:            safeNumber(getField(fields, 4)),
		Open:                 safeNumber(getField(fields, 5)),
		Volume:               safeNumber(getField(fields, 6)),
		Time:                 getField(fields, 30),
		Change:               safeNumber(getField(fields, 31)),
		ChangePercent:        safeNumber(getField(fields, 32)),
		High:                 safeNumber(getField(fields, 33)),
		Low:                  safeNumber(getField(fields, 34)),
		Amount:               safeNumber(getField(fields, 37)),
		LotSize:              safeNumberOrNull(getField(fields, 40)),
		CirculatingMarketCap: safeNumberOrNull(getField(fields, 44)),
		TotalMarketCap:       safeNumberOrNull(getField(fields, 45)),
		Currency:             currency,
		Raw:                  fields,
	}
}

// ParseUSQuote 解析美股行情
func ParseUSQuote(fields []string) USQuote {
	return USQuote{
		MarketID:       getField(fields, 0),
		Name:           getField(fields, 1),
		Code:           getField(fields, 2),
		Price:          safeNumber(getField(fields, 3)),
		PrevClose:      safeNumber(getField(fields, 4)),
		Open:           safeNumber(getField(fields, 5)),
		Volume:         safeNumber(getField(fields, 6)),
		Time:           getField(fields, 30),
		Change:         safeNumber(getField(fields, 31)),
		ChangePercent:  safeNumber(getField(fields, 32)),
		High:           safeNumber(getField(fields, 33)),
		Low:            safeNumber(getField(fields, 34)),
		Amount:         safeNumber(getField(fields, 37)),
		TurnoverRate:   safeNumberOrNull(getField(fields, 38)),
		PE:             safeNumberOrNull(getField(fields, 39)),
		Amplitude:      safeNumberOrNull(getField(fields, 43)),
		TotalMarketCap: safeNumberOrNull(getField(fields, 45)),
		PB:             safeNumberOrNull(getField(fields, 47)),
		High52W:        safeNumberOrNull(getField(fields, 48)),
		Low52W:         safeNumberOrNull(getField(fields, 49)),
		Raw:            fields,
	}
}

// ParseFundQuote 解析基金行情
func ParseFundQuote(fields []string) FundQuote {
	return FundQuote{
		Code:    getField(fields, 0),
		Name:    getField(fields, 1),
		NAV:     safeNumber(getField(fields, 5)),
		AccNAV:  safeNumber(getField(fields, 6)),
		Change:  safeNumber(getField(fields, 7)),
		NAVDate: getField(fields, 8),
		Raw:     fields,
	}
}

// ============ 东方财富板块响应类型 ============

// EmBoardListResponse 东方财富板块列表 API 响应
type EmBoardListResponse struct {
	Data *EmBoardListData `json:"data"`
}

// EmBoardListData 东方财富板块列表数据
type EmBoardListData struct {
	Total int                      `json:"total"`
	Diff  []map[string]interface{} `json:"diff"`
}

// EmBoardSpotResponse 东方财富板块实时行情 API 响应
type EmBoardSpotResponse struct {
	Data map[string]interface{} `json:"data"`
}

// EmBoardKlineResponse 东方财富板块 K 线 API 响应
type EmBoardKlineResponse struct {
	Data *EmBoardKlineData `json:"data"`
}

// EmBoardKlineData 东方财富板块 K 线数据
type EmBoardKlineData struct {
	Klines []string `json:"klines"`
	Name   string   `json:"name"`
	Code   string   `json:"code"`
}

// ============ 东方财富板块解析器 ============

// BoardTypeConfig 板块类型配置
type BoardTypeConfig struct {
	Type         string
	FsFilter     string
	ListURL      string
	SpotURL      string
	ConsURL      string
	KlineURL     string
	TrendsURL    string
	ErrorPrefix  string
}

// IndustryConfig 行业板块配置
var IndustryConfig = BoardTypeConfig{
	Type:        "industry",
	FsFilter:    "m:90 t:2 f:!50",
	ListURL:     EMBoardListURL,
	SpotURL:     EMBoardSpotURL,
	ConsURL:     EMBoardConsURL,
	KlineURL:    EMBoardKlineURL,
	TrendsURL:   EMBoardTrendsURL,
	ErrorPrefix: "未找到行业板块",
}

// ConceptConfig 概念板块配置
var ConceptConfig = BoardTypeConfig{
	Type:        "concept",
	FsFilter:    "m:90 t:3 f:!50",
	ListURL:     EMConceptListURL,
	SpotURL:     EMConceptSpotURL,
	ConsURL:     EMConceptConsURL,
	KlineURL:    EMConceptKlineURL,
	TrendsURL:   EMConceptTrendsURL,
	ErrorPrefix: "未找到概念板块",
}

// parseBoardListItem 解析板块列表项
func parseBoardListItem(item map[string]interface{}, index int, isConcept bool) IndustryBoard {
	var result IndustryBoard
	result.Rank = index + 1

	if name, ok := item["f14"].(string); ok {
		result.Name = name
	}
	if code, ok := item["f12"].(string); ok {
		result.Code = code
	}
	result.Price = safeNumberFromInterface(item["f2"])
	result.Change = safeNumberFromInterface(item["f4"])
	result.ChangePercent = safeNumberFromInterface(item["f3"])
	result.TotalMarketCap = safeNumberFromInterface(item["f20"])
	result.TurnoverRate = safeNumberFromInterface(item["f8"])
	result.RiseCount = safeIntFromInterface(item["f104"])
	result.FallCount = safeIntFromInterface(item["f105"])
	if leadingStock, ok := item["f128"].(string); ok {
		result.LeadingStock = &leadingStock
	}
	result.LeadingStockChangePct = safeNumberFromInterface(item["f136"])

	return result
}

// parseBoardConstituentItem 解析板块成分股项
func parseBoardConstituentItem(item map[string]interface{}, index int) IndustryBoardConstituent {
	var result IndustryBoardConstituent
	result.Rank = index + 1

	if code, ok := item["f12"].(string); ok {
		result.Code = code
	}
	if name, ok := item["f14"].(string); ok {
		result.Name = name
	}
	result.Price = safeNumberFromInterface(item["f2"])
	result.ChangePercent = safeNumberFromInterface(item["f3"])
	result.Change = safeNumberFromInterface(item["f4"])
	result.Volume = safeNumberFromInterface(item["f5"])
	result.Amount = safeNumberFromInterface(item["f6"])
	result.Amplitude = safeNumberFromInterface(item["f7"])
	result.High = safeNumberFromInterface(item["f15"])
	result.Low = safeNumberFromInterface(item["f16"])
	result.Open = safeNumberFromInterface(item["f17"])
	result.PrevClose = safeNumberFromInterface(item["f18"])
	result.TurnoverRate = safeNumberFromInterface(item["f8"])
	result.PE = safeNumberFromInterface(item["f9"])
	result.PB = safeNumberFromInterface(item["f23"])

	return result
}

// ParseIndustryBoardKlineCsv 解析行业板块 K 线 CSV 行
func ParseIndustryBoardKlineCsv(line string) IndustryBoardKline {
	parts := strings.Split(line, ",")
	var item IndustryBoardKline

	if len(parts) > 0 {
		item.Date = parts[0]
	}
	if len(parts) > 1 {
		item.Open = toNumberOrNull(parts[1])
	}
	if len(parts) > 2 {
		item.Close = toNumberOrNull(parts[2])
	}
	if len(parts) > 3 {
		item.High = toNumberOrNull(parts[3])
	}
	if len(parts) > 4 {
		item.Low = toNumberOrNull(parts[4])
	}
	if len(parts) > 5 {
		item.Volume = toNumberOrNull(parts[5])
	}
	if len(parts) > 6 {
		item.Amount = toNumberOrNull(parts[6])
	}
	if len(parts) > 7 {
		item.Amplitude = toNumberOrNull(parts[7])
	}
	if len(parts) > 8 {
		item.ChangePercent = toNumberOrNull(parts[8])
	}
	if len(parts) > 9 {
		item.Change = toNumberOrNull(parts[9])
	}
	if len(parts) > 10 {
		item.TurnoverRate = toNumberOrNull(parts[10])
	}

	return item
}

// ParseFuturesKlineCsv 解析期货 K 线 CSV 行（包含持仓量）
func ParseFuturesKlineCsv(line string) (FuturesKline, string, string) {
	parts := strings.Split(line, ",")
	var item FuturesKline

	if len(parts) > 0 {
		item.Date = parts[0]
	}
	if len(parts) > 1 {
		item.Open = toNumberOrNull(parts[1])
	}
	if len(parts) > 2 {
		item.Close = toNumberOrNull(parts[2])
	}
	if len(parts) > 3 {
		item.High = toNumberOrNull(parts[3])
	}
	if len(parts) > 4 {
		item.Low = toNumberOrNull(parts[4])
	}
	if len(parts) > 5 {
		item.Volume = toNumberOrNull(parts[5])
	}
	if len(parts) > 6 {
		item.Amount = toNumberOrNull(parts[6])
	}
	if len(parts) > 7 {
		item.Amplitude = toNumberOrNull(parts[7])
	}
	if len(parts) > 8 {
		item.ChangePct = toNumberOrNull(parts[8])
	}
	if len(parts) > 9 {
		item.Change = toNumberOrNull(parts[9])
	}
	if len(parts) > 10 {
		item.TurnoverRate = toNumberOrNull(parts[10])
	}
	if len(parts) > 12 {
		item.OpenInterest = toNumberOrNull(parts[12])
	}

	return item, "", ""
}

// GlobalFuturesSpotResponse 全球期货实时行情 API 响应
type GlobalFuturesSpotResponse struct {
	List  []GlobalFuturesSpotItem `json:"list"`
	Total int                     `json:"total"`
}

// GlobalFuturesSpotItem 全球期货实时行情项
type GlobalFuturesSpotItem struct {
	Dm    string  `json:"dm"`
	Name  string  `json:"name"`
	P     float64 `json:"p"`
	Zde   float64 `json:"zde"`
	Zdf   float64 `json:"zdf"`
	O     float64 `json:"o"`
	H     float64 `json:"h"`
	L     float64 `json:"l"`
	Zjsj  float64 `json:"zjsj"`
	Vol   float64 `json:"vol"`
	Wp    float64 `json:"wp"`
	Np    float64 `json:"np"`
	Ccl   float64 `json:"ccl"`
	Sc    float64 `json:"sc"`
	Zsjd  float64 `json:"zsjd"`
}

// MapGlobalFuturesSpotItem 映射全球期货行情项
func MapGlobalFuturesSpotItem(item GlobalFuturesSpotItem) GlobalFuturesQuote {
	return GlobalFuturesQuote{
		Code:       item.Dm,
		Name:       item.Name,
		Price:      safeNumberPtrFromFloat(item.P),
		Change:     safeNumberPtrFromFloat(item.Zde),
		ChangePct:  safeNumberPtrFromFloat(item.Zdf),
		Open:       safeNumberPtrFromFloat(item.O),
		High:       safeNumberPtrFromFloat(item.H),
		Low:        safeNumberPtrFromFloat(item.L),
		PrevSettle: safeNumberPtrFromFloat(item.Zjsj),
		Volume:     safeNumberPtrFromFloat(item.Vol),
		BuyVolume:  safeNumberPtrFromFloat(item.Wp),
		SellVolume: safeNumberPtrFromFloat(item.Np),
		OpenInt:    safeNumberPtrFromFloat(item.Ccl),
	}
}

func safeNumberPtrFromFloat(f float64) *float64 {
	if f == 0 {
		return nil
	}
	return &f
}

// safeNumberFromInterface 从 interface{} 安全转换为 *float64
func safeNumberFromInterface(v interface{}) *float64 {
	switch val := v.(type) {
	case float64:
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return nil
		}
		return &val
	case float32:
		f := float64(val)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil
		}
		return &f
	case int:
		f := float64(val)
		return &f
	case int64:
		f := float64(val)
		return &f
	case string:
		return toNumberOrNull(val)
	default:
		return nil
	}
}

// safeIntFromInterface 从 interface{} 安全转换为 *int
func safeIntFromInterface(v interface{}) *int {
	switch val := v.(type) {
	case int:
		return &val
	case float64:
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return nil
		}
		i := int(val)
		return &i
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return &i
		}
		return nil
	default:
		return nil
	}
}
