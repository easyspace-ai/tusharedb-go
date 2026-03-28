package realtimedata

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/easyspace-ai/stock_api/internal/provider/multisource"
	"github.com/easyspace-ai/stock_api/internal/storage/parquet"
)

// Client 实时数据客户端 - 整合多数据源采集 + Parquet存储
type Client struct {
	mgr    *multisource.MultiSourceManager
	reqMgr *multisource.RequestManager
	lake   *parquet.RealTimeLakeManager
	config Config
	mu     sync.RWMutex
}

// Config 配置
type Config struct {
	// 数据目录
	DataDir string
	// 是否启用本地存储
	EnableStorage bool
	// 缓存模式
	CacheMode CacheMode
}

// CacheMode 缓存模式
type CacheMode string

const (
	// CacheModeDisabled 不使用缓存，总是从网络获取
	CacheModeDisabled CacheMode = "disabled"
	// CacheModeReadOnly 只读模式，只从本地获取
	CacheModeReadOnly CacheMode = "readonly"
	// CacheModeAuto 自动模式，优先本地，缺失时下载
	CacheModeAuto CacheMode = "auto"
)

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		DataDir:       "./data",
		EnableStorage: true,
		CacheMode:     CacheModeAuto,
	}
}

// NewClient 创建实时数据客户端
func NewClient(config Config) (*Client, error) {
	reqMgr := multisource.GetRequestManager()
	mgr := multisource.GetMultiSourceManager()
	mgr.RegisterEnhancedSources()

	c := &Client{
		mgr:    mgr,
		reqMgr: reqMgr,
		config: config,
	}

	if config.EnableStorage {
		c.lake = parquet.NewRealTimeLakeManager(config.DataDir)
		if err := c.lake.Init(); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// NewDefaultClient 创建默认客户端
func NewDefaultClient() *Client {
	c, _ := NewClient(DefaultConfig())
	return c
}

// ========== 行情接口 ==========

// StockQuote 股票行情
type StockQuote struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	PrevClose float64 `json:"prevClose"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Volume    float64 `json:"volume"`
	Amount    float64 `json:"amount"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"changePct"`
	Time      string  `json:"time"`
}

// GetQuote 获取单个股票行情
func (c *Client) GetQuote(ctx context.Context, code string) (*StockQuote, error) {
	quotes, err := c.GetQuotes(ctx, []string{code})
	if err != nil {
		return nil, err
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no data for %s", code)
	}
	return &quotes[0], nil
}

// GetQuotes 获取多个股票行情
func (c *Client) GetQuotes(ctx context.Context, codes []string) ([]StockQuote, error) {
	// 尝试从本地读取
	if c.config.EnableStorage && c.config.CacheMode != CacheModeDisabled {
		today := time.Now().Format("20060102")
		local, err := c.lake.ReadQuotes(ctx, today, codes)
		if err == nil && len(local) >= len(codes) {
			result := make([]StockQuote, len(local))
			for i, q := range local {
				result[i] = StockQuote{
					Code:      q.Code,
					Name:      q.Name,
					Price:     q.Price,
					PrevClose: q.PrevClose,
					Open:      q.Open,
					High:      q.High,
					Low:       q.Low,
					Volume:    q.Volume,
					Amount:    q.Amount,
					Change:    q.Change,
					ChangePct: q.ChangePct,
					Time:      q.Time,
				}
			}
			return result, nil
		}
	}

	if c.config.CacheMode == CacheModeReadOnly {
		return nil, fmt.Errorf("cache miss in readonly mode")
	}

	// 从网络获取
	raw, err := c.mgr.GetStockQuotes(ctx, codes)
	if err != nil {
		return nil, err
	}

	// 转换并保存
	result := make([]StockQuote, len(raw))
	records := make([]parquet.QuoteRecord, len(raw))
	now := time.Now().Unix()

	for i, q := range raw {
		result[i] = StockQuote{
			Code:      q.Code,
			Name:      q.Name,
			Price:     q.Price,
			PrevClose: q.PrevClose,
			Open:      q.Open,
			High:      q.High,
			Low:       q.Low,
			Volume:    q.Volume,
			Amount:    q.Amount,
			Change:    q.Change,
			ChangePct: q.ChangePct,
			Time:      q.Time,
		}
		records[i] = parquet.QuoteRecord{
			Code:      q.Code,
			Name:      q.Name,
			Price:     q.Price,
			PrevClose: q.PrevClose,
			Open:      q.Open,
			High:      q.High,
			Low:       q.Low,
			Volume:    q.Volume,
			Amount:    q.Amount,
			Change:    q.Change,
			ChangePct: q.ChangePct,
			Time:      q.Time,
			Timestamp: now,
		}
	}

	if c.config.EnableStorage && c.config.CacheMode != CacheModeDisabled {
		today := time.Now().Format("20060102")
		_ = c.lake.SaveQuotes(ctx, today, records)
	}

	return result, nil
}

// GetQuotesBatch 批量获取行情
func (c *Client) GetQuotesBatch(ctx context.Context, codes []string) (map[string]*StockQuote, map[string]string, error) {
	batch, err := c.mgr.GetStockQuotesBatch(ctx, codes)
	if err != nil {
		return nil, nil, err
	}

	success := make(map[string]*StockQuote)
	for code, q := range batch.Success {
		success[code] = &StockQuote{
			Code:      q.Code,
			Name:      q.Name,
			Price:     q.Price,
			PrevClose: q.PrevClose,
			Open:      q.Open,
			High:      q.High,
			Low:       q.Low,
			Volume:    q.Volume,
			Amount:    q.Amount,
			Change:    q.Change,
			ChangePct: q.ChangePct,
			Time:      q.Time,
		}
	}

	return success, batch.Failed, nil
}

// GetQuotesBySource 仅从指定数据源拉取行情（不读写本地 Parquet lake）。
func (c *Client) GetQuotesBySource(ctx context.Context, src DataSourceType, codes []string) ([]StockQuote, error) {
	raw, err := c.mgr.GetStockQuotesBySource(ctx, src, codes)
	if err != nil {
		return nil, err
	}
	return c.convertQuotes(raw), nil
}

// ========== K线接口 ==========

// KLineItem K线数据
type KLineItem struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
	Amount float64 `json:"amount"`
}

// GetKLine 获取K线
func (c *Client) GetKLine(ctx context.Context, code string, period string, adjust string, startDate string, endDate string) ([]KLineItem, error) {
	// 尝试从本地读取
	if c.config.EnableStorage && c.config.CacheMode != CacheModeDisabled {
		local, err := c.lake.ReadKLines(ctx, code, period, adjust, startDate, endDate)
		if err == nil && len(local) > 0 {
			result := make([]KLineItem, len(local))
			for i, k := range local {
				result[i] = KLineItem{
					Date:   k.Date,
					Open:   k.Open,
					High:   k.High,
					Low:    k.Low,
					Close:  k.Close,
					Volume: k.Volume,
					Amount: k.Amount,
				}
			}
			return result, nil
		}
	}

	if c.config.CacheMode == CacheModeReadOnly {
		return nil, fmt.Errorf("cache miss in readonly mode")
	}

	// 从网络获取
	raw, err := c.mgr.GetKLine(ctx, code, period, adjust, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 转换并保存
	result := make([]KLineItem, len(raw))
	records := make([]parquet.KLineRecord, len(raw))
	now := time.Now().Unix()

	for i, k := range raw {
		result[i] = KLineItem{
			Date:   k.Date,
			Open:   k.Open,
			High:   k.High,
			Low:    k.Low,
			Close:  k.Close,
			Volume: k.Volume,
			Amount: k.Amount,
		}
		records[i] = parquet.KLineRecord{
			Code:      code,
			Date:      k.Date,
			Period:    period,
			Adjust:    adjust,
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
			Amount:    k.Amount,
			Timestamp: now,
		}
	}

	if c.config.EnableStorage && c.config.CacheMode != CacheModeDisabled {
		_ = c.lake.SaveKLines(ctx, code, period, adjust, records)
	}

	return result, nil
}

// GetDailyKLine 获取日K线（简化）
func (c *Client) GetDailyKLine(ctx context.Context, code string, adjust string, startDate string, endDate string) ([]KLineItem, error) {
	return c.GetKLine(ctx, code, "daily", adjust, startDate, endDate)
}

// GetKLineBySource 仅从指定数据源拉取 K 线（不读写本地 lake）。
func (c *Client) GetKLineBySource(ctx context.Context, src DataSourceType, code string, period string, adjust string, startDate string, endDate string) ([]KLineItem, error) {
	raw, err := c.mgr.GetKLineBySource(ctx, src, code, period, adjust, startDate, endDate)
	if err != nil {
		return nil, err
	}
	result := make([]KLineItem, len(raw))
	for i, k := range raw {
		result[i] = KLineItem{
			Date:   k.Date,
			Open:   k.Open,
			High:   k.High,
			Low:    k.Low,
			Close:  k.Close,
			Volume: k.Volume,
			Amount: k.Amount,
		}
	}
	return result, nil
}

// MinuteBar 分时数据点（腾讯）
type MinuteBar struct {
	Time      string  `json:"time"`
	Price     float64 `json:"price"`
	Volume    float64 `json:"volume"`
	ChangePct float64 `json:"changePct"`
}

// GetMinuteData 当日分时。src 传空或 DataSourceTencent；其它源尚不支持。
func (c *Client) GetMinuteData(ctx context.Context, src DataSourceType, code string) ([]MinuteBar, error) {
	raw, err := c.mgr.GetMinuteData(ctx, src, code)
	if err != nil {
		return nil, err
	}
	out := make([]MinuteBar, len(raw))
	for i, b := range raw {
		out[i] = MinuteBar{
			Time:      b.Time,
			Price:     b.Price,
			Volume:    b.Volume,
			ChangePct: b.ChangePct,
		}
	}
	return out, nil
}

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

// GetSectorList 获取板块列表
func (c *Client) GetSectorList(ctx context.Context) ([]Sector, error) {
	raw, err := c.mgr.GetSectorList(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Sector, len(raw))
	for i, s := range raw {
		result[i] = Sector{
			Code:      s.Code,
			Name:      s.Name,
			ChangePct: s.ChangePct,
			LeadStock: s.LeadStock,
			LeadPrice: s.LeadPrice,
		}
	}
	return result, nil
}

// GetIndustryList 获取行业列表
func (c *Client) GetIndustryList(ctx context.Context) ([]Industry, error) {
	raw, err := c.mgr.GetIndustryList(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Industry, len(raw))
	for i, s := range raw {
		result[i] = Industry{
			Code:        s.Code,
			Name:        s.Name,
			MarketCap:   s.MarketCap,
			Pe:          s.Pe,
			ChangePct:   s.ChangePct,
			VolumeRatio: s.VolumeRatio,
		}
	}
	return result, nil
}

// GetStockList 获取股票列表
func (c *Client) GetStockList(ctx context.Context) ([]StockInfo, error) {
	raw, err := c.mgr.GetStockList(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]StockInfo, len(raw))
	for i, s := range raw {
		result[i] = StockInfo{
			Code:     s.Code,
			Name:     s.Name,
			Market:   s.Market,
			Industry: s.Industry,
			ListDate: s.ListDate,
		}
	}
	return result, nil
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

// ========== 扩展类型定义 ==========

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
	Code     string             `json:"code"`
	Name     string             `json:"name"`
	Date     string             `json:"date"`
	Reason   string             `json:"reason"`
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

// ========== 资金流向接口 ==========

// GetMoneyFlow 获取个股资金流向
func (c *Client) GetMoneyFlow(ctx context.Context, code string) (*MoneyFlow, error) {
	raw, err := c.mgr.GetMoneyFlow(ctx, code)
	if err != nil {
		return nil, err
	}
	return &MoneyFlow{
		Code:             raw.Code,
		Name:             raw.Name,
		MainNetInflow:    raw.MainNetInflow,
		MainNetPct:       raw.MainNetPct,
		SuperLargeInflow: raw.SuperLargeInflow,
		SuperLargePct:    raw.SuperLargePct,
		LargeInflow:      raw.LargeInflow,
		LargePct:         raw.LargePct,
		MediumInflow:     raw.MediumInflow,
		MediumPct:        raw.MediumPct,
		SmallInflow:      raw.SmallInflow,
		SmallPct:         raw.SmallPct,
		Time:             raw.Time,
	}, nil
}

// GetSectorMoneyFlow 获取板块资金流向
func (c *Client) GetSectorMoneyFlow(ctx context.Context) ([]SectorMoneyFlow, error) {
	raw, err := c.mgr.GetSectorMoneyFlow(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]SectorMoneyFlow, len(raw))
	for i, r := range raw {
		result[i] = SectorMoneyFlow{
			Code:      r.Code,
			Name:      r.Name,
			NetInflow: r.NetInflow,
			NetPct:    r.NetPct,
			LeadStock: r.LeadStock,
			LeadPct:   r.LeadPct,
		}
	}
	return result, nil
}

// GetMarketMoneyFlow 获取市场资金流向
func (c *Client) GetMarketMoneyFlow(ctx context.Context) (*MarketMoneyFlow, error) {
	raw, err := c.mgr.GetMarketMoneyFlow(ctx)
	if err != nil {
		return nil, err
	}
	return &MarketMoneyFlow{
		SHMainInflow:   raw.SHMainInflow,
		SHRetailInflow: raw.SHRetailInflow,
		SZMainInflow:   raw.SZMainInflow,
		SZRetailInflow: raw.SZRetailInflow,
		TotalInflow:    raw.TotalInflow,
		Time:           raw.Time,
	}, nil
}

// GetNorthboundFlow 获取北向资金
func (c *Client) GetNorthboundFlow(ctx context.Context) (*NorthboundFlow, error) {
	raw, err := c.mgr.GetNorthboundFlow(ctx)
	if err != nil {
		return nil, err
	}
	return &NorthboundFlow{
		SHNetInflow: raw.SHNetInflow,
		SZNetInflow: raw.SZNetInflow,
		Total:       raw.Total,
		Time:        raw.Time,
	}, nil
}

// ========== 新闻资讯接口 ==========

// GetNews 获取市场新闻
func (c *Client) GetNews(ctx context.Context, count int) ([]News, error) {
	raw, err := c.mgr.GetNews(ctx, count)
	if err != nil {
		return nil, err
	}
	return c.convertNews(raw), nil
}

// GetStockNews 获取个股新闻
func (c *Client) GetStockNews(ctx context.Context, code string, count int) ([]News, error) {
	raw, err := c.mgr.GetStockNews(ctx, code, count)
	if err != nil {
		return nil, err
	}
	return c.convertNews(raw), nil
}

// GetStockNotices 获取个股公告
func (c *Client) GetStockNotices(ctx context.Context, code string, count int) ([]StockNotice, error) {
	raw, err := c.mgr.GetStockNotices(ctx, code, count)
	if err != nil {
		return nil, err
	}
	result := make([]StockNotice, len(raw))
	for i, r := range raw {
		result[i] = StockNotice{
			Code:   r.Code,
			Name:   r.Name,
			Title:  r.Title,
			Type:   r.Type,
			Url:    r.Url,
			PdfUrl: r.PdfUrl,
			Time:   r.Time,
		}
	}
	return result, nil
}

// GetStockReports 获取个股研报
func (c *Client) GetStockReports(ctx context.Context, code string, count int) ([]StockReport, error) {
	raw, err := c.mgr.GetStockReports(ctx, code, count)
	if err != nil {
		return nil, err
	}
	result := make([]StockReport, len(raw))
	for i, r := range raw {
		result[i] = StockReport{
			Code:        r.Code,
			Name:        r.Name,
			Title:       r.Title,
			Institution: r.Institution,
			Analyst:     r.Analyst,
			Rating:      r.Rating,
			TargetPrice: r.TargetPrice,
			Url:         r.Url,
			Time:        r.Time,
		}
	}
	return result, nil
}

// GetNoticeContent 公告正文（东方财富）。artCode 来自 GetStockNotices 列表项或与东财公告接口一致；stockCode 为 6 位或带 sh/sz/bj。
func (c *Client) GetNoticeContent(ctx context.Context, stockCode, artCode string) (string, error) {
	return c.mgr.GetNoticeContent(ctx, stockCode, artCode)
}

// GetReportContent 研报正文（东方财富）。infoCode 来自研报列表；部分报告仅靠公开 JSON 无法取全文，需结合 Url 浏览器打开。
func (c *Client) GetReportContent(ctx context.Context, infoCode string) (string, error) {
	return c.mgr.GetReportContent(ctx, infoCode)
}

// GetHotTopics 获取热门话题
func (c *Client) GetHotTopics(ctx context.Context) ([]HotTopic, error) {
	raw, err := c.mgr.GetHotTopics(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]HotTopic, len(raw))
	for i, r := range raw {
		result[i] = HotTopic{
			Title:  r.Title,
			Stocks: r.Stocks,
			Count:  r.Count,
		}
	}
	return result, nil
}

// ========== 全球市场接口 ==========

// GetPopularUSStocks 获取热门美股
func (c *Client) GetPopularUSStocks(ctx context.Context) ([]USStock, error) {
	raw, err := c.mgr.GetPopularUSStocks(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]USStock, len(raw))
	for i, r := range raw {
		result[i] = USStock{
			Symbol:    r.Symbol,
			Name:      r.Name,
			Price:     r.Price,
			Change:    r.Change,
			ChangePct: r.ChangePct,
		}
	}
	return result, nil
}

// GetPopularHKStocks 获取热门港股
func (c *Client) GetPopularHKStocks(ctx context.Context) ([]HKStock, error) {
	raw, err := c.mgr.GetPopularHKStocks(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]HKStock, len(raw))
	for i, r := range raw {
		result[i] = HKStock{
			Code:      r.Code,
			Name:      r.Name,
			Price:     r.Price,
			Change:    r.Change,
			ChangePct: r.ChangePct,
		}
	}
	return result, nil
}

// GetGlobalIndices 获取全球指数
func (c *Client) GetGlobalIndices(ctx context.Context) ([]GlobalIndex, error) {
	raw, err := c.mgr.GetGlobalIndices(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]GlobalIndex, len(raw))
	for i, r := range raw {
		result[i] = GlobalIndex{
			Code:      r.Code,
			Name:      r.Name,
			Price:     r.Price,
			Change:    r.Change,
			ChangePct: r.ChangePct,
			Time:      r.Time,
		}
	}
	return result, nil
}

// GetGlobalNews 获取国际新闻
func (c *Client) GetGlobalNews(ctx context.Context, region string, count int) ([]GlobalNews, error) {
	raw, err := c.mgr.GetGlobalNews(ctx, region, count)
	if err != nil {
		return nil, err
	}
	result := make([]GlobalNews, len(raw))
	for i, r := range raw {
		result[i] = GlobalNews{
			Title:   r.Title,
			Content: r.Content,
			Url:     r.Url,
			Source:  r.Source,
			Region:  r.Region,
			Time:    r.Time,
		}
	}
	return result, nil
}

// ========== 期货/加密接口 ==========

// GetFuturesList 获取期货列表
func (c *Client) GetFuturesList(ctx context.Context) ([]FuturesContract, error) {
	raw, err := c.mgr.GetFuturesList(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]FuturesContract, len(raw))
	for i, r := range raw {
		result[i] = FuturesContract{
			Symbol:   r.Symbol,
			Name:     r.Name,
			Exchange: r.Exchange,
		}
	}
	return result, nil
}

// GetFuturesPrices 获取期货行情
func (c *Client) GetFuturesPrices(ctx context.Context, symbols []string) ([]FuturesPrice, error) {
	raw, err := c.mgr.GetFuturesPrices(ctx, symbols)
	if err != nil {
		return nil, err
	}
	result := make([]FuturesPrice, len(raw))
	for i, r := range raw {
		result[i] = FuturesPrice{
			Symbol:    r.Symbol,
			Name:      r.Name,
			Price:     r.Price,
			PrevClose: r.PrevClose,
			Open:      r.Open,
			High:      r.High,
			Low:       r.Low,
			Change:    r.Change,
			ChangePct: r.ChangePct,
			Volume:    r.Volume,
			OpenInt:   r.OpenInt,
			Time:      r.Time,
		}
	}
	return result, nil
}

// GetFuturesKLine 获取期货K线
func (c *Client) GetFuturesKLine(ctx context.Context, symbol string, period string, startDate string, endDate string) ([]FuturesKLine, error) {
	raw, err := c.mgr.GetFuturesKLine(ctx, symbol, period, startDate, endDate)
	if err != nil {
		return nil, err
	}
	result := make([]FuturesKLine, len(raw))
	for i, r := range raw {
		result[i] = FuturesKLine{
			Symbol:  r.Symbol,
			Date:    r.Date,
			Open:    r.Open,
			High:    r.High,
			Low:     r.Low,
			Close:   r.Close,
			Volume:  r.Volume,
			OpenInt: r.OpenInt,
		}
	}
	return result, nil
}

// GetCryptoList 获取加密货币列表
func (c *Client) GetCryptoList(ctx context.Context) ([]Crypto, error) {
	raw, err := c.mgr.GetCryptoList(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Crypto, len(raw))
	for i, r := range raw {
		result[i] = Crypto{
			Symbol: r.Symbol,
			Name:   r.Name,
		}
	}
	return result, nil
}

// GetCryptoPrices 获取加密货币行情
func (c *Client) GetCryptoPrices(ctx context.Context, symbols []string) ([]CryptoPrice, error) {
	raw, err := c.mgr.GetCryptoPrices(ctx, symbols)
	if err != nil {
		return nil, err
	}
	result := make([]CryptoPrice, len(raw))
	for i, r := range raw {
		result[i] = CryptoPrice{
			Symbol:    r.Symbol,
			Name:      r.Name,
			Price:     r.Price,
			PriceUSD:  r.PriceUSD,
			PriceCNY:  r.PriceCNY,
			Change24h: r.Change24h,
			Volume24h: r.Volume24h,
			MarketCap: r.MarketCap,
			Time:      r.Time,
		}
	}
	return result, nil
}

// GetCryptoKLine 获取加密货币K线
func (c *Client) GetCryptoKLine(ctx context.Context, symbol string, period string, startDate string, endDate string) ([]CryptoKLine, error) {
	raw, err := c.mgr.GetCryptoKLine(ctx, symbol, period, startDate, endDate)
	if err != nil {
		return nil, err
	}
	result := make([]CryptoKLine, len(raw))
	for i, r := range raw {
		result[i] = CryptoKLine{
			Symbol: r.Symbol,
			Time:   r.Time,
			Open:   r.Open,
			High:   r.High,
			Low:    r.Low,
			Close:  r.Close,
			Volume: r.Volume,
		}
	}
	return result, nil
}

// GetForexRates 获取外汇汇率
func (c *Client) GetForexRates(ctx context.Context) ([]ForexRate, error) {
	raw, err := c.mgr.GetForexRates(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]ForexRate, len(raw))
	for i, r := range raw {
		result[i] = ForexRate{
			Pair:   r.Pair,
			Name:   r.Name,
			Rate:   r.Rate,
			Change: r.Change,
		}
	}
	return result, nil
}

// ========== 龙虎榜/停牌接口 ==========

// GetDragonTigerList 获取龙虎榜列表
func (c *Client) GetDragonTigerList(ctx context.Context, date string) ([]DragonTiger, error) {
	raw, err := c.mgr.GetDragonTigerList(ctx, date)
	if err != nil {
		return nil, err
	}
	result := make([]DragonTiger, len(raw))
	for i, r := range raw {
		result[i] = DragonTiger{
			Date:       r.Date,
			Code:       r.Code,
			Name:       r.Name,
			Reason:     r.Reason,
			BuyAmount:  r.BuyAmount,
			SellAmount: r.SellAmount,
			NetAmount:  r.NetAmount,
		}
	}
	return result, nil
}

// GetStockDragonTiger 获取个股龙虎榜
func (c *Client) GetStockDragonTiger(ctx context.Context, code string) ([]DragonTigerDetail, error) {
	raw, err := c.mgr.GetStockDragonTiger(ctx, code)
	if err != nil {
		return nil, err
	}
	result := make([]DragonTigerDetail, len(raw))
	for i, r := range raw {
		buyList := make([]DragonTigerEntry, len(r.BuyList))
		for j, b := range r.BuyList {
			buyList[j] = DragonTigerEntry(b)
		}
		sellList := make([]DragonTigerEntry, len(r.SellList))
		for j, s := range r.SellList {
			sellList[j] = DragonTigerEntry(s)
		}
		result[i] = DragonTigerDetail{
			Code:     r.Code,
			Name:     r.Name,
			Date:     r.Date,
			Reason:   r.Reason,
			BuyList:  buyList,
			SellList: sellList,
		}
	}
	return result, nil
}

// GetSuspendedStocks 获取停牌股票
func (c *Client) GetSuspendedStocks(ctx context.Context) ([]SuspendedStock, error) {
	raw, err := c.mgr.GetSuspendedStocks(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]SuspendedStock, len(raw))
	for i, r := range raw {
		result[i] = SuspendedStock{
			Code:        r.Code,
			Name:        r.Name,
			SuspendDate: r.SuspendDate,
			Reason:      r.Reason,
		}
	}
	return result, nil
}

// GetDividendInfo 获取分红信息
func (c *Client) GetDividendInfo(ctx context.Context, code string) ([]DividendInfo, error) {
	raw, err := c.mgr.GetDividendInfo(ctx, code)
	if err != nil {
		return nil, err
	}
	result := make([]DividendInfo, len(raw))
	for i, r := range raw {
		result[i] = DividendInfo{
			Code:       r.Code,
			Name:       r.Name,
			NoticeDate: r.NoticeDate,
			RecordDate: r.RecordDate,
			ExDate:     r.ExDate,
			Dividend:   r.Dividend,
			Bonus:      r.Bonus,
			Transfer:   r.Transfer,
		}
	}
	return result, nil
}

// ========== 技术指标接口（纯函数封装） ==========

// CalculateMA 计算移动平均
func CalculateMA(data []float64, period int) []float64 {
	return multisource.CalculateMA(data, period)
}

// CalculateEMA 计算指数移动平均
func CalculateEMA(data []float64, period int) []float64 {
	return multisource.CalculateEMA(data, period)
}

// CalculateMACD 计算MACD
func CalculateMACD(data []float64, fastPeriod int, slowPeriod int, signalPeriod int) *MACD {
	raw := multisource.CalculateMACD(data, fastPeriod, slowPeriod, signalPeriod)
	return &MACD{
		DIF:  raw.DIF,
		DEA:  raw.DEA,
		MACD: raw.MACD,
	}
}

// CalculateRSI 计算RSI
func CalculateRSI(data []float64, period int) []float64 {
	return multisource.CalculateRSI(data, period)
}

// CalculateKDJ 计算KDJ
func CalculateKDJ(highs []float64, lows []float64, closes []float64, period int) *KDJ {
	raw := multisource.CalculateKDJ(highs, lows, closes, period)
	return &KDJ{
		K: raw.K,
		D: raw.D,
		J: raw.J,
	}
}

// CalculateBOLL 计算布林带
func CalculateBOLL(data []float64, period int, stdDevTimes float64) *BOLL {
	raw := multisource.CalculateBOLL(data, period, stdDevTimes)
	return &BOLL{
		Upper:  raw.Upper,
		Middle: raw.Middle,
		Lower:  raw.Lower,
	}
}

// ========== 辅助转换函数 ==========

func (c *Client) convertNews(raw []multisource.News) []News {
	result := make([]News, len(raw))
	for i, r := range raw {
		result[i] = News{
			Title:   r.Title,
			Content: r.Content,
			Url:     r.Url,
			Source:  r.Source,
			Time:    r.Time,
		}
	}
	return result
}

// ========== 原有接口保持不变 ==========

// GetMarketOverview 获取市场概览
func (c *Client) GetMarketOverview(ctx context.Context) (*MarketOverview, error) {
	raw, err := c.mgr.GetMarketOverview(ctx)
	if err != nil || raw == nil {
		return nil, err
	}
	return &MarketOverview{
		SHIndex:     raw.SHIndex,
		SHChange:    raw.SHChange,
		SHChangePct: raw.SHChangePct,
		SHVolume:    raw.SHVolume,
		SHAmount:    raw.SHAmount,
		SZIndex:     raw.SZIndex,
		SZChange:    raw.SZChange,
		SZChangePct: raw.SZChangePct,
		SZVolume:    raw.SZVolume,
		SZAmount:    raw.SZAmount,
		LimitUp:     raw.LimitUp,
		LimitDown:   raw.LimitDown,
		RiseCount:   raw.RiseCount,
		FallCount:   raw.FallCount,
		FlatCount:   raw.FlatCount,
		Time:        raw.Time,
	}, nil
}

// GetIndexList 获取指数列表
func (c *Client) GetIndexList(ctx context.Context) ([]IndexInfo, error) {
	raw, err := c.mgr.GetIndexList(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]IndexInfo, len(raw))
	for i, idx := range raw {
		result[i] = IndexInfo{
			Code:      idx.Code,
			Name:      idx.Name,
			Price:     idx.Price,
			Change:    idx.Change,
			ChangePct: idx.ChangePct,
			Volume:    idx.Volume,
			Amount:    idx.Amount,
		}
	}
	return result, nil
}

// ========== 涨跌排行接口 ==========

// GetTopGainers 获取涨幅榜
func (c *Client) GetTopGainers(ctx context.Context, count int) ([]StockQuote, error) {
	raw, err := c.mgr.GetTopGainers(ctx, count)
	if err != nil {
		return nil, err
	}
	return c.convertQuotes(raw), nil
}

// GetTopLosers 获取跌幅榜
func (c *Client) GetTopLosers(ctx context.Context, count int) ([]StockQuote, error) {
	raw, err := c.mgr.GetTopLosers(ctx, count)
	if err != nil {
		return nil, err
	}
	return c.convertQuotes(raw), nil
}

func (c *Client) convertQuotes(raw []multisource.StockQuote) []StockQuote {
	result := make([]StockQuote, len(raw))
	for i, q := range raw {
		result[i] = StockQuote{
			Code:      q.Code,
			Name:      q.Name,
			Price:     q.Price,
			PrevClose: q.PrevClose,
			Open:      q.Open,
			High:      q.High,
			Low:       q.Low,
			Volume:    q.Volume,
			Amount:    q.Amount,
			Change:    q.Change,
			ChangePct: q.ChangePct,
			Time:      q.Time,
		}
	}
	return result
}

// ========== 管理接口 ==========

// ClearCache 清除内存缓存
func (c *Client) ClearCache() {
	c.reqMgr.ClearAllCache()
}

// GetStats 获取状态
func (c *Client) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"cacheMode": c.config.CacheMode,
		"dataDir":   c.config.DataDir,
	}
}

// GetRateLimiterStats 获取限流器状态
func (c *Client) GetRateLimiterStats(domain string) map[string]interface{} {
	rl := multisource.GetRateLimiter()
	return rl.GetStats(domain)
}
