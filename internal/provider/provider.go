package provider

import "context"

// FetchMode 数据获取模式
type FetchMode string

const (
	FetchModeSingle      FetchMode = "single"
	FetchModePaged       FetchMode = "paged"
	FetchModeByTradeDate FetchMode = "by_trade_date"
)

// PageRequest 分页请求
type PageRequest struct {
	Limit  int
	Offset int
}

// Request Tushare API 请求
type Request struct {
	APIName string
	Params  map[string]any
	Fields  []string
	Page    PageRequest
}

// Response Tushare API 响应
type Response struct {
	Fields []string
	Rows   []map[string]any
}

// DataProvider 定义统一的数据提供者接口
// 所有数据源（Tushare、StockSDK等）都需要实现此接口
type DataProvider interface {
	// Name 返回数据源名称
	Name() string

	// HealthCheck 检查数据源是否可用
	HealthCheck(ctx context.Context) error

	// TradeCalendarProvider 交易日历提供者
	TradeCalendarProvider

	// StockBasicProvider 股票基础信息提供者
	StockBasicProvider

	// DailyQuoteProvider 日线行情提供者
	DailyQuoteProvider

	// AdjFactorProvider 复权因子提供者
	AdjFactorProvider

	// DailyBasicProvider 每日指标提供者
	DailyBasicProvider
}

// TradeCalendarProvider 交易日历提供者接口
type TradeCalendarProvider interface {
	// FetchTradeCalendar 获取交易日历
	// startDate, endDate: 日期范围 (YYYYMMDD)
	FetchTradeCalendar(ctx context.Context, startDate, endDate string) ([]TradeCalendarRow, error)
}

// StockBasicProvider 股票基础信息提供者接口
type StockBasicProvider interface {
	// FetchStockBasic 获取股票基础信息
	// listStatus: 上市状态 L-上市 D-退市 P-暂停上市
	FetchStockBasic(ctx context.Context, listStatus string) ([]StockBasicRow, error)
}

// DailyQuoteProvider 日线行情提供者接口
type DailyQuoteProvider interface {
	// FetchDaily 获取日线行情（单个交易日横截面）
	FetchDaily(ctx context.Context, tradeDate string) ([]DailyRow, error)

	// FetchDailyRange 获取日线行情（日期范围）
	FetchDailyRange(ctx context.Context, startDate, endDate string) ([]DailyRow, error)
}

// AdjFactorProvider 复权因子提供者接口
type AdjFactorProvider interface {
	// FetchAdjFactor 获取复权因子（单个交易日横截面）
	FetchAdjFactor(ctx context.Context, tradeDate string) ([]AdjFactorRow, error)

	// FetchAdjFactorRange 获取复权因子（日期范围）
	FetchAdjFactorRange(ctx context.Context, startDate, endDate string) ([]AdjFactorRow, error)
}

// DailyBasicProvider 每日指标提供者接口
type DailyBasicProvider interface {
	// FetchDailyBasic 获取每日指标（单个交易日横截面）
	FetchDailyBasic(ctx context.Context, tradeDate string) ([]DailyBasicRow, error)

	// FetchDailyBasicRange 获取每日指标（日期范围）
	FetchDailyBasicRange(ctx context.Context, startDate, endDate string) ([]DailyBasicRow, error)
}

// === 通用数据类型定义 ===

// TradeCalendarRow 交易日历数据行
type TradeCalendarRow struct {
	Exchange     string `json:"exchange" parquet:"exchange"`
	CalDate      string `json:"cal_date" parquet:"cal_date"`
	IsOpen       string `json:"is_open" parquet:"is_open"`
	PretradeDate string `json:"pretrade_date" parquet:"pretrade_date"`
}

// StockBasicRow 股票基础信息数据行
type StockBasicRow struct {
	TSCode     string `json:"ts_code" parquet:"ts_code"`
	Symbol     string `json:"symbol" parquet:"symbol"`
	Name       string `json:"name" parquet:"name"`
	Area       string `json:"area" parquet:"area"`
	Industry   string `json:"industry" parquet:"industry"`
	Fullname   string `json:"fullname" parquet:"fullname"`
	Enname     string `json:"enname" parquet:"enname"`
	Cnspell    string `json:"cnspell" parquet:"cnspell"`
	Market     string `json:"market" parquet:"market"`
	Exchange   string `json:"exchange" parquet:"exchange"`
	CurrType   string `json:"curr_type" parquet:"curr_type"`
	ListStatus string `json:"list_status" parquet:"list_status"`
	ListDate   string `json:"list_date" parquet:"list_date"`
	DelistDate string `json:"delist_date" parquet:"delist_date"`
	IsHS       string `json:"is_hs" parquet:"is_hs"`
}

// DailyRow 日线行情数据行
type DailyRow struct {
	TSCode    string  `json:"ts_code" parquet:"ts_code"`
	TradeDate string  `json:"trade_date" parquet:"trade_date"`
	Open      float64 `json:"open" parquet:"open"`
	High      float64 `json:"high" parquet:"high"`
	Low       float64 `json:"low" parquet:"low"`
	Close     float64 `json:"close" parquet:"close"`
	PreClose  float64 `json:"pre_close" parquet:"pre_close"`
	Change    float64 `json:"change" parquet:"change"`
	PctChg    float64 `json:"pct_chg" parquet:"pct_chg"`
	Vol       float64 `json:"vol" parquet:"vol"`
	Amount    float64 `json:"amount" parquet:"amount"`
}

// AdjFactorRow 复权因子数据行
type AdjFactorRow struct {
	TSCode    string  `json:"ts_code" parquet:"ts_code"`
	TradeDate string  `json:"trade_date" parquet:"trade_date"`
	AdjFactor float64 `json:"adj_factor" parquet:"adj_factor"`
}

// DailyBasicRow 每日基本面指标数据行
type DailyBasicRow struct {
	TSCode         string  `json:"ts_code" parquet:"ts_code"`
	TradeDate      string  `json:"trade_date" parquet:"trade_date"`
	Close          float64 `json:"close" parquet:"close"`
	TurnoverRate   float64 `json:"turnover_rate" parquet:"turnover_rate"`
	TurnoverRateF  float64 `json:"turnover_rate_f" parquet:"turnover_rate_f"`
	VolumeRatio    float64 `json:"volume_ratio" parquet:"volume_ratio"`
	PE             float64 `json:"pe" parquet:"pe"`
	PETTM          float64 `json:"pe_ttm" parquet:"pe_ttm"`
	PB             float64 `json:"pb" parquet:"pb"`
	PS             float64 `json:"ps" parquet:"ps"`
	PSTTM          float64 `json:"ps_ttm" parquet:"ps_ttm"`
	DVRatio        float64 `json:"dv_ratio" parquet:"dv_ratio"`
	DVTTM          float64 `json:"dv_ttm" parquet:"dv_ttm"`
	TotalShare     float64 `json:"total_share" parquet:"total_share"`
	FloatShare     float64 `json:"float_share" parquet:"float_share"`
	FreeShare      float64 `json:"free_share" parquet:"free_share"`
	TotalMV        float64 `json:"total_mv" parquet:"total_mv"`
	CircMV         float64 `json:"circ_mv" parquet:"circ_mv"`
}
