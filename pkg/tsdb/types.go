package tsdb

import (
	"errors"

	"github.com/easyspace-ai/tusharedb-go/internal/frame"
	"github.com/easyspace-ai/tusharedb-go/internal/provider"
)

type AdjustType string

const (
	AdjustNone AdjustType = "none"
	AdjustQFQ  AdjustType = "qfq"
	AdjustHFQ  AdjustType = "hfq"
)

type SortOrder string

const (
	Asc  SortOrder = "asc"
	Desc SortOrder = "desc"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrNotSynced    = errors.New("dataset not synced")
	ErrQueryFailed  = errors.New("query failed")
	ErrSyncFailed   = errors.New("sync failed")
)

// DataSourceType 数据源类型
type DataSourceType string

const (
	// DataSourceTushare Tushare 数据源
	DataSourceTushare DataSourceType = "tushare"
	// DataSourceStockSDK StockSDK 数据源
	DataSourceStockSDK DataSourceType = "stocksdk"
)

// Config 客户端配置
type Config struct {
	// 主数据源类型 (默认为 tushare)
	PrimaryDataSource DataSourceType
	// Tushare 配置 (当 PrimaryDataSource = tushare 时使用)
	TushareToken string
	// Token 是 TushareToken 的别名 (向后兼容)
	Token string
	// StockSDK 配置 (当 PrimaryDataSource = stocksdk 时使用)
	StockSDKAPIKey string
	// 备用数据源类型 (可选)
	FallbackDataSource DataSourceType
	// 数据目录
	DataDir string
	// DuckDB 路径
	DuckDBPath string
	// 临时目录
	TempDir string
	// 日志级别
	LogLevel string
}

// DataSourceConfig 数据源配置 (用于运行时动态创建 Provider)
type DataSourceConfig struct {
	// 数据源类型
	Type DataSourceType
	// Tushare 配置
	TushareToken string
	// StockSDK 配置
	StockSDKAPIKey string
}

// ProviderFactory Provider 工厂函数
type ProviderFactory func(cfg DataSourceConfig) (provider.DataProvider, error)

type DataFrame = frame.DataFrame

type StockBasicFilter struct {
	TSCode     string
	ListStatus string
	Market     string
}

type TradeCalendarFilter struct {
	Exchange  string
	StartDate string
	EndDate   string
	IsOpen    *bool
}

type Filter struct {
	Field string
	Op    string
	Value any
}

type Order struct {
	Field string
	Order SortOrder
}

type UniverseSpec struct {
	ListStatus string
	Markets    []string
	ExcludeST  bool
}

type ScreenRequest struct {
	TradeDate string
	Universe  UniverseSpec
	Filters   []Filter
	OrderBy   []Order
	Limit     int
	Fields    []string
}

type BarRequest struct {
	TSCodes   []string
	StartDate string
	EndDate   string
	Adjust    AdjustType
	WithBasic bool
}
