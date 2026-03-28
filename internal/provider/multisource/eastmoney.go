package multisource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// EastMoneyFullSource 东方财富完整实现
type EastMoneyFullSource struct {
	priority int
	reqMgr   *RequestManager
}

// NewEastMoneySource 创建东方财富数据源
func NewEastMoneySource(priority int, reqMgr *RequestManager) *EastMoneyFullSource {
	return &EastMoneyFullSource{
		priority: priority,
		reqMgr:   reqMgr,
	}
}

func (s *EastMoneyFullSource) Name() string {
	return "eastmoney"
}

func (s *EastMoneyFullSource) Type() DataSourceType {
	return DataSourceEastMoney
}

func (s *EastMoneyFullSource) Priority() int {
	return s.priority
}

func (s *EastMoneyFullSource) HealthCheck(ctx context.Context) error {
	_, err := s.GetStockQuotes(ctx, []string{"000001"})
	return err
}

// ========== 东方财富API配置 ==========

const (
	// EMKlineURL 东方财富K线接口
	EMKlineURL = "https://push2his.eastmoney.com/api/qt/stock/kline/get"
	// EMQuoteURL 东方财富行情接口
	EMQuoteURL = "https://push2.eastmoney.com/api/qt/stock/get"
	// EMBatchQuoteURL 东方财富批量行情接口
	EMBatchQuoteURL = "https://push2.eastmoney.com/api/qt/clist/get"
	// EMKlineAltURL1 备用K线接口
	EMKlineAltURL1 = "https://push2his.eastmoney.com/api/qt/stock/kline/get"
	// EMKlineAltURL2 备用K线接口2
	EMKlineAltURL2 = "https://his.eastmoney.com/api/qt/stock/kline/get"
)

// GetMarketCode 获取市场代码
func GetMarketCode(code string) string {
	code = strings.TrimSpace(code)
	if strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "SH") {
		return "1"
	}
	if strings.HasPrefix(code, "sz") || strings.HasPrefix(code, "SZ") {
		return "0"
	}
	if strings.HasPrefix(code, "bj") || strings.HasPrefix(code, "BJ") {
		return "2"
	}
	if len(code) == 6 {
		if strings.HasPrefix(code, "6") || strings.HasPrefix(code, "9") || strings.HasPrefix(code, "5") || strings.HasPrefix(code, "11") || strings.HasPrefix(code, "13") || strings.HasPrefix(code, "204") || strings.HasPrefix(code, "511") || strings.HasPrefix(code, "51") || strings.HasPrefix(code, "58") || strings.HasPrefix(code, "60") || strings.HasPrefix(code, "68") || strings.HasPrefix(code, "78") || strings.HasPrefix(code, "88") || strings.HasPrefix(code, "90") || strings.HasPrefix(code, "1") || strings.HasPrefix(code, "50") || strings.HasPrefix(code, "55") || strings.HasPrefix(code, "56") || strings.HasPrefix(code, "73") || strings.HasPrefix(code, "10") || strings.HasPrefix(code, "11") || strings.HasPrefix(code, "12") || strings.HasPrefix(code, "13") || strings.HasPrefix(code, "14") || strings.HasPrefix(code, "20") || strings.HasPrefix(code, "70") || strings.HasPrefix(code, "71") || strings.HasPrefix(code, "72") || strings.HasPrefix(code, "90") || strings.HasPrefix(code, "91") || strings.HasPrefix(code, "92") || strings.HasPrefix(code, "93") || strings.HasPrefix(code, "94") || strings.HasPrefix(code, "95") || strings.HasPrefix(code, "96") || strings.HasPrefix(code, "97") || strings.HasPrefix(code, "98") || strings.HasPrefix(code, "99") {
			return "1"
		}
		if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "2") || strings.HasPrefix(code, "3") {
			return "0"
		}
		if strings.HasPrefix(code, "4") || strings.HasPrefix(code, "8") {
			return "2"
		}
	}
	// 默认上海
	return "1"
}

// RemoveMarketPrefix 移除市场前缀
func RemoveMarketPrefix(code string) string {
	code = strings.TrimSpace(code)
	if strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "sz") || strings.HasPrefix(code, "bj") ||
		strings.HasPrefix(code, "SH") || strings.HasPrefix(code, "SZ") || strings.HasPrefix(code, "BJ") {
		return code[2:]
	}
	return code
}

// AddMarketPrefix 添加市场前缀
func AddMarketPrefix(code string) string {
	code = strings.TrimSpace(code)
	if strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "sz") || strings.HasPrefix(code, "bj") ||
		strings.HasPrefix(code, "SH") || strings.HasPrefix(code, "SZ") || strings.HasPrefix(code, "BJ") {
		return strings.ToLower(code)
	}
	if len(code) == 6 {
		if strings.HasPrefix(code, "6") || strings.HasPrefix(code, "5") || strings.HasPrefix(code, "9") {
			return "sh" + code
		}
		if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "2") || strings.HasPrefix(code, "3") {
			return "sz" + code
		}
		if strings.HasPrefix(code, "4") || strings.HasPrefix(code, "8") {
			return "bj" + code
		}
	}
	return "sz" + code
}

// GetPeriodCode 获取K线周期代码
func GetPeriodCode(period string) string {
	switch period {
	case "daily", "day", "d", "1d", "":
		return "101"
	case "weekly", "week", "w":
		return "102"
	case "monthly", "month", "m":
		return "103"
	case "1min", "1m":
		return "1"
	case "5min", "5m":
		return "5"
	case "15min", "15m":
		return "15"
	case "30min", "30m":
		return "30"
	case "60min", "60m", "1h":
		return "60"
	default:
		return "101"
	}
}

// GetAdjustCode 获取复权方式代码
func GetAdjustCode(adjust string) string {
	switch adjust {
	case "none", "None", "0", "":
		return "0"
	case "qfq", "QFQ", "前复权", "1":
		return "1"
	case "hfq", "HFQ", "后复权", "2":
		return "2"
	default:
		return "1"
	}
}

// EmKlineResponse 东方财富K线响应
type EmKlineResponse struct {
	RC   int    `json:"rc"`
	RT   int    `json:"rt"`
	Svr  int    `json:"svr"`
	LT   int    `json:"lt"`
	Full int    `json:"full"`
	Dlmk int    `json:"dlmk"`
	Data *struct {
		Code   string   `json:"code"`
		Market int      `json:"market"`
		Name   string   `json:"name"`
		Klines []string `json:"klines"`
	} `json:"data"`
}

// ParseEmKlineCsv 解析东方财富K线CSV行
func ParseEmKlineCsv(line string) *ParsedEmKline {
	parts := splitCSVLine(line)
	if len(parts) < 11 {
		return nil
	}

	result := &ParsedEmKline{
		Date:          parts[0],
		Open:          parseFloatPtr(parts[1]),
		Close:         parseFloatPtr(parts[2]),
		High:          parseFloatPtr(parts[3]),
		Low:           parseFloatPtr(parts[4]),
		Volume:        parseFloatPtr(parts[5]),
		Amount:        parseFloatPtr(parts[6]),
		Amplitude:     parseFloatPtr(parts[7]),
		ChangePercent: parseFloatPtr(parts[8]),
		Change:        parseFloatPtr(parts[9]),
		TurnoverRate:  parseFloatPtr(parts[10]),
	}

	return result
}

// ParsedEmKline 解析后的K线数据
type ParsedEmKline struct {
	Date          string
	Open          *float64
	Close         *float64
	High          *float64
	Low           *float64
	Volume        *float64
	Amount        *float64
	Amplitude     *float64
	ChangePercent *float64
	Change        *float64
	TurnoverRate  *float64
}

func splitCSVLine(line string) []string {
	// 简单的逗号分隔，东方财富的CSV没有引号
	return strings.Split(line, ",")
}

func parseFloatPtr(s string) *float64 {
	if s == "" || s == "None" || s == "null" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &f
}

func safeFloatFromPtr(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

// GetStockQuotes 获取股票行情
func (s *EastMoneyFullSource) GetStockQuotes(ctx context.Context, codes []string) ([]StockQuote, error) {
	if len(codes) == 0 {
		return []StockQuote{}, nil
	}

	// 尝试缓存
	cacheKey := CacheKey("em_quotes", codes...)
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if quotes, ok := cached.([]StockQuote); ok {
			return quotes, nil
		}
	}

	var result []StockQuote
	var err error

	// 逐个获取（东方财富批量接口较复杂）
	for _, code := range codes {
		quote, qErr := s.getSingleQuote(ctx, code)
		if qErr == nil {
			result = append(result, quote)
		} else {
			err = qErr
		}
	}

	if len(result) > 0 {
		s.reqMgr.SetCache(cacheKey, result, CacheTimeQuote)
		return result, nil
	}

	return nil, err
}

// getSingleQuote 获取单个股票行情
func (s *EastMoneyFullSource) getSingleQuote(ctx context.Context, code string) (StockQuote, error) {
	pureCode := RemoveMarketPrefix(code)
	marketCode := GetMarketCode(code)
	secid := fmt.Sprintf("%s.%s", marketCode, pureCode)

	params := url.Values{}
	params.Set("secid", secid)
	params.Set("fields", "f43,f44,f45,f46,f47,f48,f49,f50,f51,f52,f57,f58,f60,f107,f116,f117,f127,f152,f161,f162,f163,f164,f165,f168,f169,f170,f171")
	params.Set("ut", "bd1d9ddb04089700cf9c27f6f7426281")

	fullURL := fmt.Sprintf("%s?%s", EMQuoteURL, params.Encode())

	data, err := s.reqMgr.GetWithRateLimit("eastmoney.com", fullURL)
	if err != nil {
		return StockQuote{}, err
	}

	var resp EmQuoteResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return StockQuote{}, err
	}

	if resp.Data == nil {
		return StockQuote{}, fmt.Errorf("no data")
	}

	return StockQuote{
		Code:      code,
		Name:      resp.Data.Name,
		Price:     safeFloatFromPtr(resp.Data.Price),
		PrevClose: safeFloatFromPtr(resp.Data.PrevClose),
		Open:      safeFloatFromPtr(resp.Data.Open),
		High:      safeFloatFromPtr(resp.Data.High),
		Low:       safeFloatFromPtr(resp.Data.Low),
		Volume:    safeFloatFromPtr(resp.Data.Volume),
		Amount:    safeFloatFromPtr(resp.Data.Amount),
		Change:    safeFloatFromPtr(resp.Data.Change),
		ChangePct: safeFloatFromPtr(resp.Data.ChangePct),
		Time:      time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// EmQuoteResponse 东方财富行情响应
type EmQuoteResponse struct {
	Data *struct {
		Code      string   `json:"f57"`
		Name      string   `json:"f58"`
		Price     *float64 `json:"f43"`
		PrevClose *float64 `json:"f60"`
		Open      *float64 `json:"f46"`
		High      *float64 `json:"f44"`
		Low       *float64 `json:"f45"`
		Volume    *float64 `json:"f47"`
		Amount    *float64 `json:"f48"`
		Change    *float64 `json:"f169"`
		ChangePct *float64 `json:"f170"`
	} `json:"data"`
}

// GetKLine 获取K线数据
func (s *EastMoneyFullSource) GetKLine(ctx context.Context, code string, period string, adjust string, startDate string, endDate string) ([]KLineItem, error) {
	if startDate == "" {
		startDate = "19700101"
	}
	if endDate == "" {
		endDate = "20500101"
	}

	// 缓存
	cacheKey := CacheKey("em_kline", code, period, adjust, startDate, endDate)
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if klines, ok := cached.([]KLineItem); ok {
			return klines, nil
		}
	}

	pureSymbol := RemoveMarketPrefix(code)
	secid := fmt.Sprintf("%s.%s", GetMarketCode(code), pureSymbol)

	params := url.Values{}
	params.Set("fields1", "f1,f2,f3,f4,f5,f6")
	params.Set("fields2", "f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61,f116")
	params.Set("ut", "7eea3edcaed734bea9cbfc24409ed989")
	params.Set("klt", GetPeriodCode(period))
	params.Set("fqt", GetAdjustCode(adjust))
	params.Set("secid", secid)
	params.Set("beg", startDate)
	params.Set("end", endDate)

	endpoints := []string{
		EMKlineURL,
		EMKlineAltURL1,
		EMKlineAltURL2,
	}

	var data []byte
	var err error

	for i, endpoint := range endpoints {
		fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())
		data, err = s.reqMgr.GetWithRateLimit("eastmoney.com", fullURL)
		if err == nil {
			break
		}
		if i < len(endpoints)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	if err != nil {
		return nil, err
	}

	var resp EmKlineResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if resp.Data == nil || len(resp.Data.Klines) == 0 {
		return []KLineItem{}, nil
	}

	var results []KLineItem
	for _, line := range resp.Data.Klines {
		parsed := ParseEmKlineCsv(line)
		if parsed == nil {
			continue
		}
		results = append(results, KLineItem{
			Date:   parsed.Date,
			Open:   safeFloatFromPtr(parsed.Open),
			High:   safeFloatFromPtr(parsed.High),
			Low:    safeFloatFromPtr(parsed.Low),
			Close:  safeFloatFromPtr(parsed.Close),
			Volume: safeFloatFromPtr(parsed.Volume),
			Amount: safeFloatFromPtr(parsed.Amount),
		})
	}

	if len(results) > 0 {
		s.reqMgr.SetCache(cacheKey, results, CacheTimeKLine)
	}

	return results, nil
}
