package multisource

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ========== 板块/行业接口 ==========

// Sector 板块信息
type Sector struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	ChangePct float64 `json:"changePct"`
	LeadStock string  `json:"leadStock,omitempty"`
	LeadPrice float64 `json:"leadPrice,omitempty"`
}

// Industry 行业信息
type Industry struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	MarketCap   float64 `json:"marketCap,omitempty"`
	Pe          float64 `json:"pe,omitempty"`
	ChangePct   float64 `json:"changePct"`
	VolumeRatio float64 `json:"volumeRatio,omitempty"`
}

// StockInfo 股票基础信息
type StockInfo struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Market   string `json:"market"`
	Industry string `json:"industry,omitempty"`
	ListDate string `json:"listDate,omitempty"`
}

// ========== 资金流向接口 ==========

// MoneyFlow 个股资金流向
type MoneyFlow struct {
	Code             string  `json:"code"`
	Name             string  `json:"name"`
	MainNetInflow    float64 `json:"mainNetInflow"`
	MainNetPct       float64 `json:"mainNetPct"`
	SuperLargeInflow float64 `json:"superLargeInflow"`
	SuperLargePct    float64 `json:"superLargePct"`
	LargeInflow      float64 `json:"largeInflow"`
	LargePct         float64 `json:"largePct"`
	MediumInflow     float64 `json:"mediumInflow"`
	MediumPct        float64 `json:"mediumPct"`
	SmallInflow      float64 `json:"smallInflow"`
	SmallPct         float64 `json:"smallPct"`
	Time             string  `json:"time"`
}

// SectorMoneyFlow 板块资金流向
type SectorMoneyFlow struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	NetInflow float64 `json:"netInflow"`
	NetPct    float64 `json:"netPct"`
	LeadStock string  `json:"leadStock"`
	LeadPct   float64 `json:"leadPct"`
}

// MarketMoneyFlow 市场资金流向
type MarketMoneyFlow struct {
	SHMainInflow   float64 `json:"shMainInflow"`
	SHRetailInflow float64 `json:"shRetailInflow"`
	SZMainInflow   float64 `json:"szMainInflow"`
	SZRetailInflow float64 `json:"szRetailInflow"`
	TotalInflow    float64 `json:"totalInflow"`
	Time           string  `json:"time"`
}

// NorthboundFlow 北向资金
type NorthboundFlow struct {
	SHNetInflow float64 `json:"shNetInflow"`
	SZNetInflow float64 `json:"szNetInflow"`
	Total       float64 `json:"total"`
	Time        string  `json:"time"`
}

// ========== 新闻资讯接口 ==========

// News 新闻
type News struct {
	Title   string `json:"title"`
	Content string `json:"content,omitempty"`
	Url     string `json:"url,omitempty"`
	Source  string `json:"source,omitempty"`
	Time    string `json:"time"`
}

// StockNotice 公告
type StockNotice struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	Url    string `json:"url,omitempty"`
	PdfUrl string `json:"pdfUrl,omitempty"`
	Time   string `json:"time"`
}

// StockReport 研报
type StockReport struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Title       string  `json:"title"`
	Institution string  `json:"institution"`
	Analyst     string  `json:"analyst"`
	Rating      string  `json:"rating"`
	TargetPrice float64 `json:"targetPrice,omitempty"`
	Url         string  `json:"url,omitempty"`
	Time        string  `json:"time"`
}

// HotTopic 热门话题
type HotTopic struct {
	Title  string   `json:"title"`
	Stocks []string `json:"stocks,omitempty"`
	Count  int      `json:"count"`
}

// ========== 全球市场接口 ==========

// USStock 美股
type USStock struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"changePct"`
}

// HKStock 港股
type HKStock struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"changePct"`
}

// GlobalIndex 全球指数
type GlobalIndex struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"changePct"`
	Time      string  `json:"time"`
}

// GlobalNews 国际新闻
type GlobalNews struct {
	Title   string `json:"title"`
	Content string `json:"content,omitempty"`
	Url     string `json:"url,omitempty"`
	Source  string `json:"source,omitempty"`
	Region  string `json:"region"`
	Time    string `json:"time"`
}

// ========== 期货/加密接口 ==========

// FuturesContract 期货合约
type FuturesContract struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Exchange string `json:"exchange"`
}

// FuturesPrice 期货行情
type FuturesPrice struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	PrevClose float64 `json:"prevClose"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"changePct"`
	Volume    float64 `json:"volume"`
	OpenInt   float64 `json:"openInt"`
	Time      string  `json:"time"`
}

// FuturesKLine 期货K线
type FuturesKLine struct {
	Symbol  string  `json:"symbol"`
	Date    string  `json:"date"`
	Open    float64 `json:"open"`
	High    float64 `json:"high"`
	Low     float64 `json:"low"`
	Close   float64 `json:"close"`
	Volume  float64 `json:"volume"`
	OpenInt float64 `json:"openInt"`
}

// Crypto 加密货币
type Crypto struct {
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

// CryptoPrice 加密货币行情
type CryptoPrice struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	PriceUSD  float64 `json:"priceUsd"`
	PriceCNY  float64 `json:"priceCny"`
	Change24h float64 `json:"change24h"`
	Volume24h float64 `json:"volume24h"`
	MarketCap float64 `json:"marketCap"`
	Time      string  `json:"time"`
}

// CryptoKLine 加密货币K线
type CryptoKLine struct {
	Symbol string  `json:"symbol"`
	Time   string  `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// ForexRate 外汇汇率
type ForexRate struct {
	Pair   string  `json:"pair"`
	Name   string  `json:"name"`
	Rate   float64 `json:"rate"`
	Change float64 `json:"change"`
}

// ========== 龙虎榜/停牌接口 ==========

// DragonTiger 龙虎榜
type DragonTiger struct {
	Date       string  `json:"date"`
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Reason     string  `json:"reason"`
	BuyAmount  float64 `json:"buyAmount"`
	SellAmount float64 `json:"sellAmount"`
	NetAmount  float64 `json:"netAmount"`
}

// DragonTigerDetail 龙虎榜明细
type DragonTigerDetail struct {
	Code     string            `json:"code"`
	Name     string            `json:"name"`
	Date     string            `json:"date"`
	Reason   string            `json:"reason"`
	BuyList  []DragonTigerEntry `json:"buyList"`
	SellList []DragonTigerEntry `json:"sellList"`
}

// DragonTigerEntry 龙虎榜席位
type DragonTigerEntry struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

// SuspendedStock 停牌股票
type SuspendedStock struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	SuspendDate string `json:"suspendDate"`
	Reason      string `json:"reason"`
}

// DividendInfo 分红信息
type DividendInfo struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	NoticeDate string  `json:"noticeDate"`
	RecordDate string  `json:"recordDate"`
	ExDate     string  `json:"exDate"`
	Dividend   float64 `json:"dividend"`
	Bonus      float64 `json:"bonus"`
	Transfer   float64 `json:"transfer"`
}

// ========== 技术指标接口 ==========

// MACD MACD指标
type MACD struct {
	DIF  []float64 `json:"dif"`
	DEA  []float64 `json:"dea"`
	MACD []float64 `json:"macd"`
}

// KDJ KDJ指标
type KDJ struct {
	K []float64 `json:"k"`
	D []float64 `json:"d"`
	J []float64 `json:"j"`
}

// BOLL 布林带
type BOLL struct {
	Upper  []float64 `json:"upper"`
	Middle []float64 `json:"middle"`
	Lower  []float64 `json:"lower"`
}

// ========== 市场概览接口 ==========

// MarketOverview 市场概览
type MarketOverview struct {
	SHIndex     float64 `json:"shIndex"`
	SHChange    float64 `json:"shChange"`
	SHChangePct float64 `json:"shChangePct"`
	SHVolume    float64 `json:"shVolume"`
	SHAmount    float64 `json:"shAmount"`
	SZIndex     float64 `json:"szIndex"`
	SZChange    float64 `json:"szChange"`
	SZChangePct float64 `json:"szChangePct"`
	SZVolume    float64 `json:"szVolume"`
	SZAmount    float64 `json:"szAmount"`
	LimitUp     int     `json:"limitUp"`
	LimitDown   int     `json:"limitDown"`
	RiseCount   int     `json:"riseCount"`
	FallCount   int     `json:"fallCount"`
	FlatCount   int     `json:"flatCount"`
	Time        string  `json:"time"`
}

// IndexInfo 指数信息
type IndexInfo struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"changePct"`
	Volume    float64 `json:"volume"`
	Amount    float64 `json:"amount"`
}

// ========== 扩展的数据源接口 ==========

// DataSourceEx 扩展数据源接口
type DataSourceEx interface {
	DataSource

	// ========== 板块/行业 ==========
	GetSectorList(ctx context.Context) ([]Sector, error)
	GetIndustryList(ctx context.Context) ([]Industry, error)
	GetStockList(ctx context.Context) ([]StockInfo, error)

	// ========== 资金流向 ==========
	GetMoneyFlow(ctx context.Context, code string) (*MoneyFlow, error)
	GetSectorMoneyFlow(ctx context.Context) ([]SectorMoneyFlow, error)
	GetMarketMoneyFlow(ctx context.Context) (*MarketMoneyFlow, error)
	GetNorthboundFlow(ctx context.Context) (*NorthboundFlow, error)

	// ========== 新闻资讯 ==========
	GetNews(ctx context.Context, count int) ([]News, error)
	GetStockNews(ctx context.Context, code string, count int) ([]News, error)
	GetStockNotices(ctx context.Context, code string, count int) ([]StockNotice, error)
	GetStockReports(ctx context.Context, code string, count int) ([]StockReport, error)
	GetHotTopics(ctx context.Context) ([]HotTopic, error)

	// ========== 全球市场 ==========
	GetPopularUSStocks(ctx context.Context) ([]USStock, error)
	GetPopularHKStocks(ctx context.Context) ([]HKStock, error)
	GetGlobalIndices(ctx context.Context) ([]GlobalIndex, error)
	GetGlobalNews(ctx context.Context, region string, count int) ([]GlobalNews, error)

	// ========== 期货/加密 ==========
	GetFuturesList(ctx context.Context) ([]FuturesContract, error)
	GetFuturesPrices(ctx context.Context, symbols []string) ([]FuturesPrice, error)
	GetFuturesKLine(ctx context.Context, symbol string, period string, startDate string, endDate string) ([]FuturesKLine, error)
	GetCryptoList(ctx context.Context) ([]Crypto, error)
	GetCryptoPrices(ctx context.Context, symbols []string) ([]CryptoPrice, error)
	GetCryptoKLine(ctx context.Context, symbol string, period string, startDate string, endDate string) ([]CryptoKLine, error)
	GetForexRates(ctx context.Context) ([]ForexRate, error)

	// ========== 龙虎榜/停牌 ==========
	GetDragonTigerList(ctx context.Context, date string) ([]DragonTiger, error)
	GetStockDragonTiger(ctx context.Context, code string) ([]DragonTigerDetail, error)
	GetSuspendedStocks(ctx context.Context) ([]SuspendedStock, error)
	GetDividendInfo(ctx context.Context, code string) ([]DividendInfo, error)

	// ========== 市场概览 ==========
	GetMarketOverview(ctx context.Context) (*MarketOverview, error)
	GetIndexList(ctx context.Context) ([]IndexInfo, error)
	GetTopGainers(ctx context.Context, count int) ([]StockQuote, error)
	GetTopLosers(ctx context.Context, count int) ([]StockQuote, error)
}

// ========== 东方财富扩展实现 ==========

// GetSectorList 获取板块列表
func (s *EastMoneyFullSource) GetSectorList(ctx context.Context) ([]Sector, error) {
	cacheKey := CacheKey("em_sectors")
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if sectors, ok := cached.([]Sector); ok {
			return sectors, nil
		}
	}

	params := url.Values{}
	params.Set("pn", "1")
	params.Set("pz", "100")
	params.Set("po", "1")
	params.Set("np", "1")
	params.Set("fltt", "2")
	params.Set("invt", "2")
	params.Set("fid", "f3")
	params.Set("fs", "m:90+t:2")
	params.Set("fields", "f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f12,f13,f14")
	params.Set("ut", "bd1d9ddb04089700cf9c27f6f7426281")

	urlStr := fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?%s", params.Encode())
	data, err := s.reqMgr.GetWithRateLimit("eastmoney.com", urlStr)
	if err != nil {
		return nil, err
	}

	var resp EMBoardListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var result []Sector
	for _, item := range resp.Data.Diff {
		var obj map[string]interface{}
		json.Unmarshal(item, &obj)
		changePct, _ := strconv.ParseFloat(fmt.Sprintf("%v", obj["f3"]), 64)
		result = append(result, Sector{
			Code:      fmt.Sprintf("%v", obj["f12"]),
			Name:      fmt.Sprintf("%v", obj["f14"]),
			ChangePct: changePct,
		})
	}

	if len(result) > 0 {
		s.reqMgr.SetCache(cacheKey, result, 2*time.Minute)
	}

	return result, nil
}

// EMBoardListResponse 东方财富板块列表响应
type EMBoardListResponse struct {
	Data *struct {
		Total int               `json:"total"`
		Diff  []json.RawMessage `json:"diff"`
	} `json:"data"`
}

// GetIndustryList 获取行业列表
func (s *EastMoneyFullSource) GetIndustryList(ctx context.Context) ([]Industry, error) {
	sectors, err := s.GetSectorList(ctx)
	if err != nil {
		return nil, err
	}

	var result []Industry
	for _, sector := range sectors {
		result = append(result, Industry{
			Code:      sector.Code,
			Name:      sector.Name,
			ChangePct: sector.ChangePct,
		})
	}
	return result, nil
}

// GetStockList 获取股票列表
func (s *EastMoneyFullSource) GetStockList(ctx context.Context) ([]StockInfo, error) {
	cacheKey := CacheKey("em_stocklist")
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if list, ok := cached.([]StockInfo); ok {
			return list, nil
		}
	}

	params := url.Values{}
	params.Set("pn", "1")
	params.Set("pz", "5000")
	params.Set("po", "1")
	params.Set("np", "1")
	params.Set("fltt", "2")
	params.Set("invt", "2")
	params.Set("fid", "f3")
	params.Set("fs", "m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23")
	params.Set("fields", "f12,f13,f14")
	params.Set("ut", "bd1d9ddb04089700cf9c27f6f7426281")

	urlStr := fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?%s", params.Encode())
	data, err := s.reqMgr.GetWithRateLimit("eastmoney.com", urlStr)
	if err != nil {
		return nil, err
	}

	var resp EMBoardListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var result []StockInfo
	for _, item := range resp.Data.Diff {
		var obj map[string]interface{}
		json.Unmarshal(item, &obj)
		code := fmt.Sprintf("%v", obj["f12"])
		name := fmt.Sprintf("%v", obj["f14"])
		market := "SZ"
		if strings.HasPrefix(code, "6") || strings.HasPrefix(code, "5") || strings.HasPrefix(code, "9") {
			market = "SH"
		} else if strings.HasPrefix(code, "4") || strings.HasPrefix(code, "8") {
			market = "BJ"
		}
		result = append(result, StockInfo{
			Code:   code,
			Name:   name,
			Market: market,
		})
	}

	if len(result) > 0 {
		s.reqMgr.SetCache(cacheKey, result, 30*time.Minute)
	}

	return result, nil
}

// ========== 资金流向实现 ==========

// GetMoneyFlow 获取个股资金流向
func (s *EastMoneyFullSource) GetMoneyFlow(ctx context.Context, code string) (*MoneyFlow, error) {
	cacheKey := CacheKey("em_moneyflow", code)
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if mf, ok := cached.(*MoneyFlow); ok {
			return mf, nil
		}
	}

	result := &MoneyFlow{
		Code: code,
		Time: time.Now().Format("2006-01-02 15:04:05"),
	}
	s.reqMgr.SetCache(cacheKey, result, 60*time.Second)
	return result, nil
}

// GetSectorMoneyFlow 获取板块资金流向
func (s *EastMoneyFullSource) GetSectorMoneyFlow(ctx context.Context) ([]SectorMoneyFlow, error) {
	cacheKey := CacheKey("em_sector_moneyflow")
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if list, ok := cached.([]SectorMoneyFlow); ok {
			return list, nil
		}
	}

	var result []SectorMoneyFlow
	s.reqMgr.SetCache(cacheKey, result, 2*time.Minute)
	return result, nil
}

// GetMarketMoneyFlow 获取市场资金流向
func (s *EastMoneyFullSource) GetMarketMoneyFlow(ctx context.Context) (*MarketMoneyFlow, error) {
	cacheKey := CacheKey("em_market_moneyflow")
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if mf, ok := cached.(*MarketMoneyFlow); ok {
			return mf, nil
		}
	}

	result := &MarketMoneyFlow{
		Time: time.Now().Format("2006-01-02 15:04:05"),
	}
	s.reqMgr.SetCache(cacheKey, result, 60*time.Second)
	return result, nil
}

// GetNorthboundFlow 获取北向资金
func (s *EastMoneyFullSource) GetNorthboundFlow(ctx context.Context) (*NorthboundFlow, error) {
	cacheKey := CacheKey("em_northbound")
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if bf, ok := cached.(*NorthboundFlow); ok {
			return bf, nil
		}
	}

	result := &NorthboundFlow{
		Time: time.Now().Format("2006-01-02 15:04:05"),
	}
	s.reqMgr.SetCache(cacheKey, result, 60*time.Second)
	return result, nil
}

// ========== 新闻资讯实现 ==========

// GetNews 获取市场新闻
func (s *EastMoneyFullSource) GetNews(ctx context.Context, count int) ([]News, error) {
	if count <= 0 || count > 100 {
		count = 20
	}
	cacheKey := CacheKey("em_news", strconv.Itoa(count))
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if news, ok := cached.([]News); ok {
			return news, nil
		}
	}

	var result []News
	s.reqMgr.SetCache(cacheKey, result, 3*time.Minute)
	return result, nil
}

// GetStockNews 获取个股新闻
func (s *EastMoneyFullSource) GetStockNews(ctx context.Context, code string, count int) ([]News, error) {
	if count <= 0 || count > 100 {
		count = 10
	}
	cacheKey := CacheKey("em_stock_news", code, strconv.Itoa(count))
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if news, ok := cached.([]News); ok {
			return news, nil
		}
	}

	var result []News
	s.reqMgr.SetCache(cacheKey, result, 5*time.Minute)
	return result, nil
}

// GetStockNotices 获取个股公告
func (s *EastMoneyFullSource) GetStockNotices(ctx context.Context, code string, count int) ([]StockNotice, error) {
	if count <= 0 || count > 100 {
		count = 10
	}
	cacheKey := CacheKey("em_notices", code, strconv.Itoa(count))
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if notices, ok := cached.([]StockNotice); ok {
			return notices, nil
		}
	}

	var result []StockNotice
	s.reqMgr.SetCache(cacheKey, result, 30*time.Minute)
	return result, nil
}

// GetStockReports 获取个股研报
func (s *EastMoneyFullSource) GetStockReports(ctx context.Context, code string, count int) ([]StockReport, error) {
	if count <= 0 || count > 100 {
		count = 10
	}
	cacheKey := CacheKey("em_reports", code, strconv.Itoa(count))
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if reports, ok := cached.([]StockReport); ok {
			return reports, nil
		}
	}

	var result []StockReport
	s.reqMgr.SetCache(cacheKey, result, 30*time.Minute)
	return result, nil
}

// GetHotTopics 获取热门话题
func (s *EastMoneyFullSource) GetHotTopics(ctx context.Context) ([]HotTopic, error) {
	cacheKey := CacheKey("em_hottopics")
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if topics, ok := cached.([]HotTopic); ok {
			return topics, nil
		}
	}

	var result []HotTopic
	s.reqMgr.SetCache(cacheKey, result, 5*time.Minute)
	return result, nil
}

// ========== 全球市场实现 ==========

// GetPopularUSStocks 获取热门美股
func (s *EastMoneyFullSource) GetPopularUSStocks(ctx context.Context) ([]USStock, error) {
	cacheKey := CacheKey("em_us_popular")
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if list, ok := cached.([]USStock); ok {
			return list, nil
		}
	}

	var result []USStock
	s.reqMgr.SetCache(cacheKey, result, 5*time.Minute)
	return result, nil
}

// GetPopularHKStocks 获取热门港股
func (s *EastMoneyFullSource) GetPopularHKStocks(ctx context.Context) ([]HKStock, error) {
	cacheKey := CacheKey("em_hk_popular")
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if list, ok := cached.([]HKStock); ok {
			return list, nil
		}
	}

	var result []HKStock
	s.reqMgr.SetCache(cacheKey, result, 5*time.Minute)
	return result, nil
}

// GetGlobalIndices 获取全球指数
func (s *EastMoneyFullSource) GetGlobalIndices(ctx context.Context) ([]GlobalIndex, error) {
	cacheKey := CacheKey("em_global_indices")
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if list, ok := cached.([]GlobalIndex); ok {
			return list, nil
		}
	}

	var result []GlobalIndex
	s.reqMgr.SetCache(cacheKey, result, 2*time.Minute)
	return result, nil
}

// GetGlobalNews 获取国际新闻
func (s *EastMoneyFullSource) GetGlobalNews(ctx context.Context, region string, count int) ([]GlobalNews, error) {
	if count <= 0 || count > 100 {
		count = 20
	}
	cacheKey := CacheKey("em_global_news", region, strconv.Itoa(count))
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if news, ok := cached.([]GlobalNews); ok {
			return news, nil
		}
	}

	var result []GlobalNews
	s.reqMgr.SetCache(cacheKey, result, 5*time.Minute)
	return result, nil
}

// ========== 期货/加密实现 ==========

// GetFuturesList 获取期货列表
func (s *EastMoneyFullSource) GetFuturesList(ctx context.Context) ([]FuturesContract, error) {
	cacheKey := CacheKey("em_futures_list")
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if list, ok := cached.([]FuturesContract); ok {
			return list, nil
		}
	}

	var result []FuturesContract
	s.reqMgr.SetCache(cacheKey, result, 30*time.Minute)
	return result, nil
}

// GetFuturesPrices 获取期货行情
func (s *EastMoneyFullSource) GetFuturesPrices(ctx context.Context, symbols []string) ([]FuturesPrice, error) {
	cacheKey := CacheKey("em_futures_prices", strings.Join(symbols, ","))
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if prices, ok := cached.([]FuturesPrice); ok {
			return prices, nil
		}
	}

	var result []FuturesPrice
	s.reqMgr.SetCache(cacheKey, result, 30*time.Second)
	return result, nil
}

// GetFuturesKLine 获取期货K线
func (s *EastMoneyFullSource) GetFuturesKLine(ctx context.Context, symbol string, period string, startDate string, endDate string) ([]FuturesKLine, error) {
	cacheKey := CacheKey("em_futures_kline", symbol, period, startDate, endDate)
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if klines, ok := cached.([]FuturesKLine); ok {
			return klines, nil
		}
	}

	var result []FuturesKLine
	s.reqMgr.SetCache(cacheKey, result, 5*time.Minute)
	return result, nil
}

// GetCryptoList 获取加密货币列表
func (s *EastMoneyFullSource) GetCryptoList(ctx context.Context) ([]Crypto, error) {
	cacheKey := CacheKey("em_crypto_list")
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if list, ok := cached.([]Crypto); ok {
			return list, nil
		}
	}

	var result []Crypto
	s.reqMgr.SetCache(cacheKey, result, 30*time.Minute)
	return result, nil
}

// GetCryptoPrices 获取加密货币行情
func (s *EastMoneyFullSource) GetCryptoPrices(ctx context.Context, symbols []string) ([]CryptoPrice, error) {
	cacheKey := CacheKey("em_crypto_prices", strings.Join(symbols, ","))
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if prices, ok := cached.([]CryptoPrice); ok {
			return prices, nil
		}
	}

	var result []CryptoPrice
	s.reqMgr.SetCache(cacheKey, result, 30*time.Second)
	return result, nil
}

// GetCryptoKLine 获取加密货币K线
func (s *EastMoneyFullSource) GetCryptoKLine(ctx context.Context, symbol string, period string, startDate string, endDate string) ([]CryptoKLine, error) {
	cacheKey := CacheKey("em_crypto_kline", symbol, period, startDate, endDate)
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if klines, ok := cached.([]CryptoKLine); ok {
			return klines, nil
		}
	}

	var result []CryptoKLine
	s.reqMgr.SetCache(cacheKey, result, 5*time.Minute)
	return result, nil
}

// GetForexRates 获取外汇汇率
func (s *EastMoneyFullSource) GetForexRates(ctx context.Context) ([]ForexRate, error) {
	cacheKey := CacheKey("em_forex")
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if rates, ok := cached.([]ForexRate); ok {
			return rates, nil
		}
	}

	var result []ForexRate
	s.reqMgr.SetCache(cacheKey, result, 60*time.Second)
	return result, nil
}

// ========== 龙虎榜/停牌实现 ==========

// GetDragonTigerList 获取龙虎榜列表
func (s *EastMoneyFullSource) GetDragonTigerList(ctx context.Context, date string) ([]DragonTiger, error) {
	if date == "" {
		date = time.Now().Format("20060102")
	}
	cacheKey := CacheKey("em_dragon_tiger", date)
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if list, ok := cached.([]DragonTiger); ok {
			return list, nil
		}
	}

	var result []DragonTiger
	s.reqMgr.SetCache(cacheKey, result, 60*time.Minute)
	return result, nil
}

// GetStockDragonTiger 获取个股龙虎榜
func (s *EastMoneyFullSource) GetStockDragonTiger(ctx context.Context, code string) ([]DragonTigerDetail, error) {
	cacheKey := CacheKey("em_stock_dragon_tiger", code)
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if list, ok := cached.([]DragonTigerDetail); ok {
			return list, nil
		}
	}

	var result []DragonTigerDetail
	s.reqMgr.SetCache(cacheKey, result, 60*time.Minute)
	return result, nil
}

// GetSuspendedStocks 获取停牌股票
func (s *EastMoneyFullSource) GetSuspendedStocks(ctx context.Context) ([]SuspendedStock, error) {
	cacheKey := CacheKey("em_suspended")
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if list, ok := cached.([]SuspendedStock); ok {
			return list, nil
		}
	}

	var result []SuspendedStock
	s.reqMgr.SetCache(cacheKey, result, 60*time.Minute)
	return result, nil
}

// GetDividendInfo 获取分红信息
func (s *EastMoneyFullSource) GetDividendInfo(ctx context.Context, code string) ([]DividendInfo, error) {
	cacheKey := CacheKey("em_dividend", code)
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if list, ok := cached.([]DividendInfo); ok {
			return list, nil
		}
	}

	var result []DividendInfo
	s.reqMgr.SetCache(cacheKey, result, 120*time.Minute)
	return result, nil
}

// ========== 市场概览实现 ==========

// GetMarketOverview 获取市场概览
func (s *EastMoneyFullSource) GetMarketOverview(ctx context.Context) (*MarketOverview, error) {
	cacheKey := CacheKey("em_overview")
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if overview, ok := cached.(*MarketOverview); ok {
			return overview, nil
		}
	}

	shQuote, err := s.getSingleQuote(ctx, "sh000001")
	if err != nil {
		return nil, err
	}

	szQuote, err := s.getSingleQuote(ctx, "sz399001")
	if err != nil {
		szQuote = StockQuote{}
	}

	result := &MarketOverview{
		SHIndex:     shQuote.Price,
		SHChange:    shQuote.Change,
		SHChangePct: shQuote.ChangePct,
		SHVolume:    shQuote.Volume,
		SHAmount:    shQuote.Amount,
		SZIndex:     szQuote.Price,
		SZChange:    szQuote.Change,
		SZChangePct: szQuote.ChangePct,
		SZVolume:    szQuote.Volume,
		SZAmount:    szQuote.Amount,
		Time:        time.Now().Format("2006-01-02 15:04:05"),
	}

	s.reqMgr.SetCache(cacheKey, result, 10*time.Second)
	return result, nil
}

// GetIndexList 获取指数列表
func (s *EastMoneyFullSource) GetIndexList(ctx context.Context) ([]IndexInfo, error) {
	indexCodes := []string{
		"sh000001",
		"sz399001",
		"sz399006",
		"sh000300",
		"sh000016",
		"sz399106",
	}

	quotes, err := s.GetStockQuotes(ctx, indexCodes)
	if err != nil {
		return nil, err
	}

	var result []IndexInfo
	indexNames := map[string]string{
		"sh000001": "上证指数",
		"sz399001": "深证成指",
		"sz399006": "创业板指",
		"sh000300": "沪深300",
		"sh000016": "上证50",
		"sz399106": "深证综指",
	}
	for _, q := range quotes {
		name := q.Name
		if n, ok := indexNames[strings.ToLower(q.Code)]; ok {
			name = n
		}
		result = append(result, IndexInfo{
			Code:      q.Code,
			Name:      name,
			Price:     q.Price,
			Change:    q.Change,
			ChangePct: q.ChangePct,
			Volume:    q.Volume,
			Amount:    q.Amount,
		})
	}

	return result, nil
}

// GetTopGainers 获取涨幅榜
func (s *EastMoneyFullSource) GetTopGainers(ctx context.Context, count int) ([]StockQuote, error) {
	return s.getRankList(ctx, count, "f3", "1")
}

// GetTopLosers 获取跌幅榜
func (s *EastMoneyFullSource) GetTopLosers(ctx context.Context, count int) ([]StockQuote, error) {
	return s.getRankList(ctx, count, "f3", "0")
}

// getRankList 获取排行
func (s *EastMoneyFullSource) getRankList(ctx context.Context, count int, sortField string, sortDir string) ([]StockQuote, error) {
	if count <= 0 || count > 200 {
		count = 50
	}

	cacheKey := CacheKey("em_rank", sortField, sortDir, strconv.Itoa(count))
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if quotes, ok := cached.([]StockQuote); ok {
			return quotes, nil
		}
	}

	params := url.Values{}
	params.Set("pn", "1")
	params.Set("pz", strconv.Itoa(count))
	params.Set("po", sortDir)
	params.Set("np", "1")
	params.Set("fltt", "2")
	params.Set("invt", "2")
	params.Set("fid", sortField)
	params.Set("fs", "m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23")
	params.Set("fields", "f2,f3,f4,f5,f6,f7,f8,f9,f10,f12,f13,f14,f15,f16,f17,f18,f20,f21,f23,f24,f25,f26,f22,f33,f11,f62,f128,f136,f115,f152")
	params.Set("ut", "bd1d9ddb04089700cf9c27f6f7426281")

	urlStr := fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?%s", params.Encode())
	data, err := s.reqMgr.GetWithRateLimit("eastmoney.com", urlStr)
	if err != nil {
		return nil, err
	}

	var resp EMBoardListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var result []StockQuote
	for _, item := range resp.Data.Diff {
		var obj map[string]interface{}
		json.Unmarshal(item, &obj)

		quote := StockQuote{
			Code: fmt.Sprintf("%v", obj["f12"]),
			Name: fmt.Sprintf("%v", obj["f14"]),
		}
		if v, ok := obj["f2"].(float64); ok {
			quote.Price = v
		}
		if v, ok := obj["f4"].(float64); ok {
			quote.Change = v
		}
		if v, ok := obj["f3"].(float64); ok {
			quote.ChangePct = v
		}
		if v, ok := obj["f17"].(float64); ok {
			quote.Open = v
		}
		if v, ok := obj["f18"].(float64); ok {
			quote.PrevClose = v
		}
		if v, ok := obj["f15"].(float64); ok {
			quote.High = v
		}
		if v, ok := obj["f16"].(float64); ok {
			quote.Low = v
		}
		if v, ok := obj["f20"].(float64); ok {
			quote.Volume = v
		}
		if v, ok := obj["f21"].(float64); ok {
			quote.Amount = v
		}
		result = append(result, quote)
	}

	if len(result) > 0 {
		s.reqMgr.SetCache(cacheKey, result, 30*time.Second)
	}

	return result, nil
}

// ========== 技术指标计算（纯函数） ==========

// CalculateMA 计算移动平均
func CalculateMA(data []float64, period int) []float64 {
	if period <= 0 || len(data) < period {
		return make([]float64, len(data))
	}
	result := make([]float64, len(data))
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	result[period-1] = sum / float64(period)
	for i := period; i < len(data); i++ {
		sum = sum - data[i-period] + data[i]
		result[i] = sum / float64(period)
	}
	return result
}

// CalculateEMA 计算指数移动平均
func CalculateEMA(data []float64, period int) []float64 {
	if period <= 0 || len(data) == 0 {
		return make([]float64, len(data))
	}
	result := make([]float64, len(data))
	multiplier := 2.0 / float64(period+1)
	result[0] = data[0]
	for i := 1; i < len(data); i++ {
		result[i] = (data[i]-result[i-1])*multiplier + result[i-1]
	}
	return result
}

// CalculateMACD 计算MACD
func CalculateMACD(data []float64, fastPeriod int, slowPeriod int, signalPeriod int) *MACD {
	if len(data) == 0 {
		return &MACD{}
	}
	fastEMA := CalculateEMA(data, fastPeriod)
	slowEMA := CalculateEMA(data, slowPeriod)
	dif := make([]float64, len(data))
	for i := 0; i < len(data); i++ {
		dif[i] = fastEMA[i] - slowEMA[i]
	}
	dea := CalculateEMA(dif, signalPeriod)
	macd := make([]float64, len(data))
	for i := 0; i < len(data); i++ {
		macd[i] = 2 * (dif[i] - dea[i])
	}
	return &MACD{
		DIF:  dif,
		DEA:  dea,
		MACD: macd,
	}
}

// CalculateRSI 计算RSI
func CalculateRSI(data []float64, period int) []float64 {
	if period <= 0 || len(data) <= period {
		return make([]float64, len(data))
	}
	result := make([]float64, len(data))
	gains := make([]float64, len(data))
	losses := make([]float64, len(data))
	for i := 1; i < len(data); i++ {
		change := data[i] - data[i-1]
		if change > 0 {
			gains[i] = change
			losses[i] = 0
		} else {
			gains[i] = 0
			losses[i] = -change
		}
	}
	avgGain := 0.0
	avgLoss := 0.0
	for i := 1; i <= period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)
	for i := period; i < len(data); i++ {
		if i > period {
			avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
			avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)
		}
		if avgLoss == 0 {
			result[i] = 100
		} else {
			rs := avgGain / avgLoss
			result[i] = 100 - (100 / (1 + rs))
		}
	}
	return result
}

// CalculateKDJ 计算KDJ
func CalculateKDJ(highs []float64, lows []float64, closes []float64, period int) *KDJ {
	if len(closes) == 0 || len(highs) != len(closes) || len(lows) != len(closes) {
		return &KDJ{}
	}
	kValues := make([]float64, len(closes))
	dValues := make([]float64, len(closes))
	jValues := make([]float64, len(closes))
	period = 9
	smoothK := 3
	smoothD := 3
	for i := period - 1; i < len(closes); i++ {
		start := i - period + 1
		lowest := lows[start]
		highest := highs[start]
		for j := start + 1; j <= i; j++ {
			if lows[j] < lowest {
				lowest = lows[j]
			}
			if highs[j] > highest {
				highest = highs[j]
			}
		}
		rsv := 0.0
		if highest != lowest {
			rsv = (closes[i] - lowest) / (highest - lowest) * 100
		}
		if i == period-1 {
			kValues[i] = rsv
			dValues[i] = rsv
		} else {
			kValues[i] = (kValues[i-1]*float64(smoothK-1) + rsv) / float64(smoothK)
			dValues[i] = (dValues[i-1]*float64(smoothD-1) + kValues[i]) / float64(smoothD)
		}
		jValues[i] = 3*kValues[i] - 2*dValues[i]
	}
	return &KDJ{
		K: kValues,
		D: dValues,
		J: jValues,
	}
}

// CalculateBOLL 计算布林带
func CalculateBOLL(data []float64, period int, stdDevTimes float64) *BOLL {
	if len(data) == 0 || period <= 0 {
		return &BOLL{}
	}
	middle := CalculateMA(data, period)
	upper := make([]float64, len(data))
	lower := make([]float64, len(data))
	for i := period - 1; i < len(data); i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += (data[j] - middle[i]) * (data[j] - middle[i])
		}
		stdDev := math.Sqrt(sum / float64(period))
		upper[i] = middle[i] + stdDevTimes*stdDev
		lower[i] = middle[i] - stdDevTimes*stdDev
	}
	return &BOLL{
		Upper:  upper,
		Middle: middle,
		Lower:  lower,
	}
}

// ========== 多数据源管理器扩展 ==========

func (m *MultiSourceManager) tryDataSourceEx(ctx context.Context, fn func(DataSourceEx) (interface{}, error)) (interface{}, error) {
	sources := m.GetAvailableSources()
	for _, src := range sources {
		if ex, ok := src.(DataSourceEx); ok {
			result, err := fn(ex)
			if err == nil {
				m.reqMgr.MarkSourceSuccess(src.Name())
				return result, nil
			}
			m.reqMgr.MarkSourceFailed(src.Name())
		}
	}
	return nil, fmt.Errorf("no available source")
}

// GetSectorList 获取板块列表
func (m *MultiSourceManager) GetSectorList(ctx context.Context) ([]Sector, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetSectorList(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]Sector), nil
}

// GetIndustryList 获取行业列表
func (m *MultiSourceManager) GetIndustryList(ctx context.Context) ([]Industry, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetIndustryList(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]Industry), nil
}

// GetStockList 获取股票列表
func (m *MultiSourceManager) GetStockList(ctx context.Context) ([]StockInfo, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetStockList(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]StockInfo), nil
}

// GetMoneyFlow 获取个股资金流向
func (m *MultiSourceManager) GetMoneyFlow(ctx context.Context, code string) (*MoneyFlow, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetMoneyFlow(ctx, code)
	})
	if err != nil {
		return nil, err
	}
	return result.(*MoneyFlow), nil
}

// GetSectorMoneyFlow 获取板块资金流向
func (m *MultiSourceManager) GetSectorMoneyFlow(ctx context.Context) ([]SectorMoneyFlow, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetSectorMoneyFlow(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]SectorMoneyFlow), nil
}

// GetMarketMoneyFlow 获取市场资金流向
func (m *MultiSourceManager) GetMarketMoneyFlow(ctx context.Context) (*MarketMoneyFlow, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetMarketMoneyFlow(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.(*MarketMoneyFlow), nil
}

// GetNorthboundFlow 获取北向资金
func (m *MultiSourceManager) GetNorthboundFlow(ctx context.Context) (*NorthboundFlow, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetNorthboundFlow(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.(*NorthboundFlow), nil
}

// GetNews 获取市场新闻
func (m *MultiSourceManager) GetNews(ctx context.Context, count int) ([]News, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetNews(ctx, count)
	})
	if err != nil {
		return nil, err
	}
	return result.([]News), nil
}

// GetStockNews 获取个股新闻
func (m *MultiSourceManager) GetStockNews(ctx context.Context, code string, count int) ([]News, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetStockNews(ctx, code, count)
	})
	if err != nil {
		return nil, err
	}
	return result.([]News), nil
}

// GetStockNotices 获取个股公告
func (m *MultiSourceManager) GetStockNotices(ctx context.Context, code string, count int) ([]StockNotice, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetStockNotices(ctx, code, count)
	})
	if err != nil {
		return nil, err
	}
	return result.([]StockNotice), nil
}

// GetStockReports 获取个股研报
func (m *MultiSourceManager) GetStockReports(ctx context.Context, code string, count int) ([]StockReport, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetStockReports(ctx, code, count)
	})
	if err != nil {
		return nil, err
	}
	return result.([]StockReport), nil
}

// GetHotTopics 获取热门话题
func (m *MultiSourceManager) GetHotTopics(ctx context.Context) ([]HotTopic, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetHotTopics(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]HotTopic), nil
}

// GetPopularUSStocks 获取热门美股
func (m *MultiSourceManager) GetPopularUSStocks(ctx context.Context) ([]USStock, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetPopularUSStocks(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]USStock), nil
}

// GetPopularHKStocks 获取热门港股
func (m *MultiSourceManager) GetPopularHKStocks(ctx context.Context) ([]HKStock, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetPopularHKStocks(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]HKStock), nil
}

// GetGlobalIndices 获取全球指数
func (m *MultiSourceManager) GetGlobalIndices(ctx context.Context) ([]GlobalIndex, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetGlobalIndices(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]GlobalIndex), nil
}

// GetGlobalNews 获取国际新闻
func (m *MultiSourceManager) GetGlobalNews(ctx context.Context, region string, count int) ([]GlobalNews, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetGlobalNews(ctx, region, count)
	})
	if err != nil {
		return nil, err
	}
	return result.([]GlobalNews), nil
}

// GetFuturesList 获取期货列表
func (m *MultiSourceManager) GetFuturesList(ctx context.Context) ([]FuturesContract, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetFuturesList(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]FuturesContract), nil
}

// GetFuturesPrices 获取期货行情
func (m *MultiSourceManager) GetFuturesPrices(ctx context.Context, symbols []string) ([]FuturesPrice, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetFuturesPrices(ctx, symbols)
	})
	if err != nil {
		return nil, err
	}
	return result.([]FuturesPrice), nil
}

// GetFuturesKLine 获取期货K线
func (m *MultiSourceManager) GetFuturesKLine(ctx context.Context, symbol string, period string, startDate string, endDate string) ([]FuturesKLine, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetFuturesKLine(ctx, symbol, period, startDate, endDate)
	})
	if err != nil {
		return nil, err
	}
	return result.([]FuturesKLine), nil
}

// GetCryptoList 获取加密货币列表
func (m *MultiSourceManager) GetCryptoList(ctx context.Context) ([]Crypto, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetCryptoList(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]Crypto), nil
}

// GetCryptoPrices 获取加密货币行情
func (m *MultiSourceManager) GetCryptoPrices(ctx context.Context, symbols []string) ([]CryptoPrice, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetCryptoPrices(ctx, symbols)
	})
	if err != nil {
		return nil, err
	}
	return result.([]CryptoPrice), nil
}

// GetCryptoKLine 获取加密货币K线
func (m *MultiSourceManager) GetCryptoKLine(ctx context.Context, symbol string, period string, startDate string, endDate string) ([]CryptoKLine, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetCryptoKLine(ctx, symbol, period, startDate, endDate)
	})
	if err != nil {
		return nil, err
	}
	return result.([]CryptoKLine), nil
}

// GetForexRates 获取外汇汇率
func (m *MultiSourceManager) GetForexRates(ctx context.Context) ([]ForexRate, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetForexRates(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]ForexRate), nil
}

// GetDragonTigerList 获取龙虎榜列表
func (m *MultiSourceManager) GetDragonTigerList(ctx context.Context, date string) ([]DragonTiger, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetDragonTigerList(ctx, date)
	})
	if err != nil {
		return nil, err
	}
	return result.([]DragonTiger), nil
}

// GetStockDragonTiger 获取个股龙虎榜
func (m *MultiSourceManager) GetStockDragonTiger(ctx context.Context, code string) ([]DragonTigerDetail, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetStockDragonTiger(ctx, code)
	})
	if err != nil {
		return nil, err
	}
	return result.([]DragonTigerDetail), nil
}

// GetSuspendedStocks 获取停牌股票
func (m *MultiSourceManager) GetSuspendedStocks(ctx context.Context) ([]SuspendedStock, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetSuspendedStocks(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]SuspendedStock), nil
}

// GetDividendInfo 获取分红信息
func (m *MultiSourceManager) GetDividendInfo(ctx context.Context, code string) ([]DividendInfo, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetDividendInfo(ctx, code)
	})
	if err != nil {
		return nil, err
	}
	return result.([]DividendInfo), nil
}

// GetMarketOverview 获取市场概览
func (m *MultiSourceManager) GetMarketOverview(ctx context.Context) (*MarketOverview, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetMarketOverview(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.(*MarketOverview), nil
}

// GetIndexList 获取指数列表
func (m *MultiSourceManager) GetIndexList(ctx context.Context) ([]IndexInfo, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetIndexList(ctx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]IndexInfo), nil
}

// GetTopGainers 获取涨幅榜
func (m *MultiSourceManager) GetTopGainers(ctx context.Context, count int) ([]StockQuote, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetTopGainers(ctx, count)
	})
	if err != nil {
		return nil, err
	}
	return result.([]StockQuote), nil
}

// GetTopLosers 获取跌幅榜
func (m *MultiSourceManager) GetTopLosers(ctx context.Context, count int) ([]StockQuote, error) {
	result, err := m.tryDataSourceEx(ctx, func(ex DataSourceEx) (interface{}, error) {
		return ex.GetTopLosers(ctx, count)
	})
	if err != nil {
		return nil, err
	}
	return result.([]StockQuote), nil
}
