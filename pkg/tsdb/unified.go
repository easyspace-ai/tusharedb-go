package tsdb

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/easyspace-ai/tusharedb-go/internal/config"
	"github.com/easyspace-ai/tusharedb-go/internal/dataset"
	_ "github.com/easyspace-ai/tusharedb-go/internal/dataset/builtin"
	"github.com/easyspace-ai/tusharedb-go/internal/provider"
	"github.com/easyspace-ai/tusharedb-go/internal/query/duckdb"
	"github.com/easyspace-ai/tusharedb-go/internal/storage/meta"
	"github.com/easyspace-ai/tusharedb-go/internal/syncer"
)

// normalizeConfig 配置归一化（支持离线模式下 Token 可选）
func normalizeConfig(cfg UnifiedConfig) (config.Config, error) {
	result := config.Config{
		Token:      cfg.TushareToken,
		DataDir:    cfg.DataDir,
		DuckDBPath: cfg.DuckDBPath,
		TempDir:    cfg.TempDir,
		LogLevel:   cfg.LogLevel,
	}

	// 设置默认值
	if result.DataDir == "" {
		result.DataDir = "./data"
	}
	if result.TempDir == "" {
		result.TempDir = filepath.Join(result.DataDir, "tmp")
	}
	if result.DuckDBPath == "" {
		result.DuckDBPath = filepath.Join(result.DataDir, "duckdb", "tusharedb.duckdb")
	}
	if result.LogLevel == "" {
		result.LogLevel = "info"
	}

	// StockSDK 数据源不需要 Token（使用公开接口）
	// Tushare 数据源在在线模式下需要 Token
	if cfg.CacheMode != CacheModeReadOnly && cfg.PrimaryDataSource == DataSourceTushare && result.Token == "" {
		return config.Config{}, fmt.Errorf("tushare token is required for online mode")
	}

	return result, nil
}

// CacheMode 缓存模式
type CacheMode string

const (
	// CacheModeDisabled 禁用缓存，总是从网络获取
	CacheModeDisabled CacheMode = "disabled"
	// CacheModeReadOnly 只读缓存，本地没有则报错
	CacheModeReadOnly CacheMode = "readonly"
	// CacheModeAuto 自动模式，本地没有则自动下载并缓存（默认）
	CacheModeAuto CacheMode = "auto"
)

// UnifiedConfig 统一客户端配置
type UnifiedConfig struct {
	// 数据源配置（与 Config 兼容）
	PrimaryDataSource  DataSourceType
	TushareToken       string
	Token              string
	StockSDKAPIKey     string
	FallbackDataSource DataSourceType

	// 数据目录
	DataDir    string
	DuckDBPath string
	TempDir    string
	LogLevel   string

	// 缓存模式
	// disabled: 禁用缓存，总是从网络获取
	// readonly: 只读缓存，本地没有则报错
	// auto: 自动模式，本地没有则自动下载并缓存（默认）
	CacheMode CacheMode
}

// UnifiedClient 统一客户端
// 对外提供统一 API，内部自动处理缓存逻辑
type UnifiedClient struct {
	cfg              config.Config
	cacheMode        CacheMode
	primaryProvider  provider.DataProvider
	fallbackProvider provider.DataProvider
	registry         *dataset.Registry
	checkpoint       *meta.CheckpointStore
	engine           *duckdb.Engine
	syncer           *syncer.Syncer
}

// NewUnifiedClient 创建统一客户端
func NewUnifiedClient(cfg UnifiedConfig) (*UnifiedClient, error) {
	return NewUnifiedClientWithFactory(cfg, DefaultProviderFactory)
}

// NewUnifiedClientWithFactory 使用自定义 Provider 工厂创建统一客户端
func NewUnifiedClientWithFactory(cfg UnifiedConfig, factories map[DataSourceType]ProviderFactory) (*UnifiedClient, error) {
	// 设置默认值
	if cfg.PrimaryDataSource == "" {
		cfg.PrimaryDataSource = DataSourceTushare
	}
	if cfg.CacheMode == "" {
		cfg.CacheMode = CacheModeAuto
	}
	if cfg.TushareToken == "" && cfg.Token != "" {
		cfg.TushareToken = cfg.Token
	}

	// 归一化配置（离线模式下 Token 可选）
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}

	// 创建 Provider
	primaryFactory, ok := factories[cfg.PrimaryDataSource]
	if !ok {
		return nil, fmt.Errorf("unsupported primary data source: %s", cfg.PrimaryDataSource)
	}

	primaryProvider, err := primaryFactory(DataSourceConfig{
		Type:           cfg.PrimaryDataSource,
		TushareToken:   cfg.TushareToken,
		StockSDKAPIKey: cfg.StockSDKAPIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create primary provider: %w", err)
	}

	var fallbackProvider provider.DataProvider
	if cfg.FallbackDataSource != "" {
		fallbackFactory, ok := factories[cfg.FallbackDataSource]
		if !ok {
			return nil, fmt.Errorf("unsupported fallback data source: %s", cfg.FallbackDataSource)
		}
		fallbackProvider, err = fallbackFactory(DataSourceConfig{
			Type:           cfg.FallbackDataSource,
			TushareToken:   cfg.TushareToken,
			StockSDKAPIKey: cfg.StockSDKAPIKey,
		})
		if err != nil {
			return nil, fmt.Errorf("create fallback provider: %w", err)
		}
	}

	// 初始化存储组件
	registry := dataset.NewRegistry()
	dataset.RegisterBuiltins(registry)

	checkpoint, err := meta.NewCheckpointStore(filepath.Join(normalized.DataDir, "meta", "checkpoints.json"))
	if err != nil {
		return nil, fmt.Errorf("init checkpoint store: %w", err)
	}

	engine, err := duckdb.NewEngine(duckdb.Config{
		DuckDBPath: normalized.DuckDBPath,
		DataDir:    normalized.DataDir,
	})
	if err != nil {
		return nil, fmt.Errorf("init duckdb engine: %w", err)
	}

	s := syncer.NewSyncer(syncer.Config{
		DataDir: normalized.DataDir,
	}, primaryProvider, registry, checkpoint, engine)

	return &UnifiedClient{
		cfg:              normalized,
		cacheMode:        cfg.CacheMode,
		primaryProvider:  primaryProvider,
		fallbackProvider: fallbackProvider,
		registry:         registry,
		checkpoint:       checkpoint,
		engine:           engine,
		syncer:           s,
	}, nil
}

// Close 关闭客户端
func (c *UnifiedClient) Close() error {
	if c == nil {
		return nil
	}
	return c.engine.Close()
}

// ==================== 统一 API（自动处理缓存）====================

// GetStockBasic 获取股票基础信息
// 根据 CacheMode 自动处理：
//   - disabled: 直接从网络获取
//   - readonly: 只查本地，没有则报错
//   - auto: 先查本地，没有则下载并缓存
func (c *UnifiedClient) GetStockBasic(ctx context.Context, filter StockBasicFilter) (*DataFrame, error) {
	switch c.cacheMode {
	case CacheModeDisabled:
		return c.fetchStockBasicFromNetwork(ctx, filter)
	case CacheModeReadOnly:
		return c.engine.GetStockBasic(ctx, duckdb.StockBasicFilter(filter))
	case CacheModeAuto:
		fallthrough
	default:
		// 先尝试本地
		df, err := c.engine.GetStockBasic(ctx, duckdb.StockBasicFilter(filter))
		if err == nil && len(df.Rows) > 0 {
			return df, nil
		}
		// 本地没有，尝试同步
		if err := c.syncer.SyncStockBasic(ctx, filter.ListStatus); err != nil {
			return nil, fmt.Errorf("sync stock_basic failed: %w", err)
		}
		// 再次读取
		return c.engine.GetStockBasic(ctx, duckdb.StockBasicFilter(filter))
	}
}

// GetTradeCalendar 获取交易日历
func (c *UnifiedClient) GetTradeCalendar(ctx context.Context, filter TradeCalendarFilter) (*DataFrame, error) {
	switch c.cacheMode {
	case CacheModeDisabled:
		return c.fetchTradeCalendarFromNetwork(ctx, filter)
	case CacheModeReadOnly:
		return c.engine.GetTradeCalendar(ctx, duckdb.TradeCalendarFilter(filter))
	case CacheModeAuto:
		fallthrough
	default:
		df, err := c.engine.GetTradeCalendar(ctx, duckdb.TradeCalendarFilter(filter))
		if err == nil && len(df.Rows) > 0 {
			return df, nil
		}
		// 同步并重新读取
		if err := c.syncer.SyncDatasetRange(ctx, "trade_cal", filter.StartDate, filter.EndDate); err != nil {
			return nil, fmt.Errorf("sync trade_cal failed: %w", err)
		}
		return c.engine.GetTradeCalendar(ctx, duckdb.TradeCalendarFilter(filter))
	}
}

// GetStockDaily 获取单个股票日线数据（支持复权）
func (c *UnifiedClient) GetStockDaily(ctx context.Context, tsCode, startDate, endDate string, adjust AdjustType) (*DataFrame, error) {
	return c.GetMultipleStocksDaily(ctx, []string{tsCode}, startDate, endDate, adjust)
}

// GetMultipleStocksDaily 获取多个股票日线数据（支持复权）
func (c *UnifiedClient) GetMultipleStocksDaily(ctx context.Context, tsCodes []string, startDate, endDate string, adjust AdjustType) (*DataFrame, error) {
	switch c.cacheMode {
	case CacheModeDisabled:
		return c.fetchDailyFromNetwork(ctx, tsCodes, startDate, endDate, adjust)
	case CacheModeReadOnly:
		return c.engine.GetMultipleStocksDaily(ctx, tsCodes, startDate, endDate, string(adjust))
	case CacheModeAuto:
		fallthrough
	default:
		df, err := c.engine.GetMultipleStocksDaily(ctx, tsCodes, startDate, endDate, string(adjust))
		if err == nil && c.checkDailyDataComplete(df, tsCodes, startDate, endDate) {
			return df, nil
		}
		// 数据不完整，需要同步
		if err := c.syncer.SyncDatasetRange(ctx, "daily", startDate, endDate); err != nil {
			// 同步失败但本地有数据，返回本地数据
			if df != nil && len(df.Rows) > 0 {
				return df, nil
			}
			return nil, fmt.Errorf("sync daily failed: %w", err)
		}
		// 如果涉及复权，同步复权因子
		if adjust != AdjustNone {
			if err := c.syncer.SyncDatasetRange(ctx, "adj_factor", startDate, endDate); err != nil {
				// 非致命错误，继续
			}
		}
		return c.engine.GetMultipleStocksDaily(ctx, tsCodes, startDate, endDate, string(adjust))
	}
}

// GetAdjFactor 获取复权因子
func (c *UnifiedClient) GetAdjFactor(ctx context.Context, tsCode, startDate, endDate string) (*DataFrame, error) {
	switch c.cacheMode {
	case CacheModeDisabled:
		return c.fetchAdjFactorFromNetwork(ctx, tsCode, startDate, endDate)
	case CacheModeReadOnly:
		return c.engine.GetAdjFactor(ctx, tsCode, startDate, endDate)
	case CacheModeAuto:
		fallthrough
	default:
		df, err := c.engine.GetAdjFactor(ctx, tsCode, startDate, endDate)
		if err == nil && len(df.Rows) > 0 {
			return df, nil
		}
		if err := c.syncer.SyncDatasetRange(ctx, "adj_factor", startDate, endDate); err != nil {
			return nil, fmt.Errorf("sync adj_factor failed: %w", err)
		}
		return c.engine.GetAdjFactor(ctx, tsCode, startDate, endDate)
	}
}

// GetDailyBasic 获取每日指标
func (c *UnifiedClient) GetDailyBasic(ctx context.Context, tsCode, startDate, endDate string) (*DataFrame, error) {
	switch c.cacheMode {
	case CacheModeDisabled:
		return c.fetchDailyBasicFromNetwork(ctx, tsCode, startDate, endDate)
	case CacheModeReadOnly:
		return c.engine.GetDailyBasic(ctx, tsCode, startDate, endDate)
	case CacheModeAuto:
		fallthrough
	default:
		df, err := c.engine.GetDailyBasic(ctx, tsCode, startDate, endDate)
		if err == nil && len(df.Rows) > 0 {
			return df, nil
		}
		if err := c.syncer.SyncDatasetRange(ctx, "daily_basic", startDate, endDate); err != nil {
			return nil, fmt.Errorf("sync daily_basic failed: %w", err)
		}
		return c.engine.GetDailyBasic(ctx, tsCode, startDate, endDate)
	}
}

// Query 执行自定义 SQL 查询（仅本地数据）
func (c *UnifiedClient) Query(ctx context.Context, sql string, args ...any) (*DataFrame, error) {
	return c.engine.Query(ctx, sql, args...)
}

// ==================== 同步控制方法 ====================

// SyncCore 同步核心数据（stock_basic + trade_cal）
func (c *UnifiedClient) SyncCore(ctx context.Context) error {
	return c.syncer.SyncCore(ctx)
}

// SyncDailyRange 同步日线数据范围
func (c *UnifiedClient) SyncDailyRange(ctx context.Context, startDate, endDate string) error {
	return c.syncer.SyncDatasetRange(ctx, "daily", startDate, endDate)
}

// SyncAdjFactorRange 同步复权因子范围
func (c *UnifiedClient) SyncAdjFactorRange(ctx context.Context, startDate, endDate string) error {
	return c.syncer.SyncDatasetRange(ctx, "adj_factor", startDate, endDate)
}

// SyncDailyBasicRange 同步每日指标范围
func (c *UnifiedClient) SyncDailyBasicRange(ctx context.Context, startDate, endDate string) error {
	return c.syncer.SyncDatasetRange(ctx, "daily_basic", startDate, endDate)
}

// SyncIncremental 增量同步（从上次同步日期到今天）
func (c *UnifiedClient) SyncIncremental(ctx context.Context) error {
	// 同步日线增量
	if err := c.syncer.SyncDatasetIncremental(ctx, "daily"); err != nil {
		return fmt.Errorf("sync daily incremental: %w", err)
	}
	// 同步复权因子增量
	if err := c.syncer.SyncDatasetIncremental(ctx, "adj_factor"); err != nil {
		return fmt.Errorf("sync adj_factor incremental: %w", err)
	}
	// 同步每日指标增量
	if err := c.syncer.SyncDatasetIncremental(ctx, "daily_basic"); err != nil {
		return fmt.Errorf("sync daily_basic incremental: %w", err)
	}
	return nil
}

// GetLastSyncDate 获取数据集最后同步日期
func (c *UnifiedClient) GetLastSyncDate(dataset string) (string, bool) {
	cp, ok := c.checkpoint.Get(dataset)
	if !ok {
		return "", false
	}
	return cp.LastSyncedDate, true
}

// ==================== 内部方法：从网络获取 ====================

func (c *UnifiedClient) fetchStockBasicFromNetwork(ctx context.Context, filter StockBasicFilter) (*DataFrame, error) {
	rows, err := c.primaryProvider.FetchStockBasic(ctx, filter.ListStatus)
	if err != nil && c.fallbackProvider != nil {
		rows, err = c.fallbackProvider.FetchStockBasic(ctx, filter.ListStatus)
	}
	if err != nil {
		return nil, err
	}

	// 转换为 DataFrame
	df := &DataFrame{
		Columns: []string{"ts_code", "symbol", "name", "area", "industry", "market", "list_status", "list_date"},
		Rows:    make([]map[string]any, len(rows)),
	}
	for i, row := range rows {
		df.Rows[i] = map[string]any{
			"ts_code":     row.TSCode,
			"symbol":      row.Symbol,
			"name":        row.Name,
			"area":        row.Area,
			"industry":    row.Industry,
			"market":      row.Market,
			"list_status": row.ListStatus,
			"list_date":   row.ListDate,
		}
	}
	return df, nil
}

func (c *UnifiedClient) fetchTradeCalendarFromNetwork(ctx context.Context, filter TradeCalendarFilter) (*DataFrame, error) {
	rows, err := c.primaryProvider.FetchTradeCalendar(ctx, filter.StartDate, filter.EndDate)
	if err != nil && c.fallbackProvider != nil {
		rows, err = c.fallbackProvider.FetchTradeCalendar(ctx, filter.StartDate, filter.EndDate)
	}
	if err != nil {
		return nil, err
	}

	df := &DataFrame{
		Columns: []string{"exchange", "cal_date", "is_open", "pretrade_date"},
		Rows:    make([]map[string]any, len(rows)),
	}
	for i, row := range rows {
		df.Rows[i] = map[string]any{
			"exchange":      row.Exchange,
			"cal_date":      row.CalDate,
			"is_open":       row.IsOpen,
			"pretrade_date": row.PretradeDate,
		}
	}
	return df, nil
}

func (c *UnifiedClient) fetchDailyFromNetwork(ctx context.Context, tsCodes []string, startDate, endDate string, adjust AdjustType) (*DataFrame, error) {
	rows, err := c.primaryProvider.FetchDailyRange(ctx, startDate, endDate)
	if err != nil && c.fallbackProvider != nil {
		rows, err = c.fallbackProvider.FetchDailyRange(ctx, startDate, endDate)
	}
	if err != nil {
		return nil, err
	}

	// 过滤指定股票
	filtered := make([]provider.DailyRow, 0, len(rows))
	codeSet := make(map[string]bool)
	for _, code := range tsCodes {
		codeSet[code] = true
	}
	for _, row := range rows {
		if codeSet[row.TSCode] {
			filtered = append(filtered, row)
		}
	}

	df := &DataFrame{
		Columns: []string{"ts_code", "trade_date", "open", "high", "low", "close", "pre_close", "change", "pct_chg", "vol", "amount"},
		Rows:    make([]map[string]any, len(filtered)),
	}
	for i, row := range filtered {
		df.Rows[i] = map[string]any{
			"ts_code":    row.TSCode,
			"trade_date": row.TradeDate,
			"open":       row.Open,
			"high":       row.High,
			"low":        row.Low,
			"close":      row.Close,
			"pre_close":  row.PreClose,
			"change":     row.Change,
			"pct_chg":    row.PctChg,
			"vol":        row.Vol,
			"amount":     row.Amount,
		}
	}
	return df, nil
}

func (c *UnifiedClient) fetchAdjFactorFromNetwork(ctx context.Context, tsCode, startDate, endDate string) (*DataFrame, error) {
	rows, err := c.primaryProvider.FetchAdjFactorRange(ctx, startDate, endDate)
	if err != nil && c.fallbackProvider != nil {
		rows, err = c.fallbackProvider.FetchAdjFactorRange(ctx, startDate, endDate)
	}
	if err != nil {
		return nil, err
	}

	// 过滤指定股票
	filtered := make([]provider.AdjFactorRow, 0, len(rows))
	for _, row := range rows {
		if row.TSCode == tsCode {
			filtered = append(filtered, row)
		}
	}

	df := &DataFrame{
		Columns: []string{"ts_code", "trade_date", "adj_factor"},
		Rows:    make([]map[string]any, len(filtered)),
	}
	for i, row := range filtered {
		df.Rows[i] = map[string]any{
			"ts_code":    row.TSCode,
			"trade_date": row.TradeDate,
			"adj_factor": row.AdjFactor,
		}
	}
	return df, nil
}

func (c *UnifiedClient) fetchDailyBasicFromNetwork(ctx context.Context, tsCode, startDate, endDate string) (*DataFrame, error) {
	rows, err := c.primaryProvider.FetchDailyBasicRange(ctx, startDate, endDate)
	if err != nil && c.fallbackProvider != nil {
		rows, err = c.fallbackProvider.FetchDailyBasicRange(ctx, startDate, endDate)
	}
	if err != nil {
		return nil, err
	}

	// 过滤指定股票
	filtered := make([]provider.DailyBasicRow, 0, len(rows))
	for _, row := range rows {
		if row.TSCode == tsCode {
			filtered = append(filtered, row)
		}
	}

	df := &DataFrame{
		Columns: []string{"ts_code", "trade_date", "close", "turnover_rate", "pe", "pe_ttm", "pb", "total_mv", "circ_mv"},
		Rows:    make([]map[string]any, len(filtered)),
	}
	for i, row := range filtered {
		df.Rows[i] = map[string]any{
			"ts_code":       row.TSCode,
			"trade_date":    row.TradeDate,
			"close":         row.Close,
			"turnover_rate": row.TurnoverRate,
			"pe":            row.PE,
			"pe_ttm":        row.PETTM,
			"pb":            row.PB,
			"total_mv":      row.TotalMV,
			"circ_mv":       row.CircMV,
		}
	}
	return df, nil
}

// checkDailyDataComplete 检查日线数据是否完整
func (c *UnifiedClient) checkDailyDataComplete(df *DataFrame, tsCodes []string, startDate, endDate string) bool {
	if df == nil || len(df.Rows) == 0 {
		return false
	}

	// 简单检查：每个股票至少有一条数据
	codeCount := make(map[string]int)
	for _, row := range df.Rows {
		if tsCode, ok := row["ts_code"].(string); ok {
			codeCount[tsCode]++
		}
	}

	for _, code := range tsCodes {
		if codeCount[code] == 0 {
			return false
		}
	}

	return true
}

// ==================== 便捷构造函数 ====================

// NewAutoClient 创建自动缓存模式的客户端（最常用）
func NewAutoClient(dataDir string) (*UnifiedClient, error) {
	return NewUnifiedClient(UnifiedConfig{
		PrimaryDataSource: DataSourceStockSDK,
		DataDir:           dataDir,
		CacheMode:         CacheModeAuto,
	})
}

// NewOfflineClient 创建纯离线客户端（只读本地缓存）
func NewOfflineClient(dataDir string) (*UnifiedClient, error) {
	return NewUnifiedClient(UnifiedConfig{
		PrimaryDataSource: DataSourceStockSDK,
		DataDir:           dataDir,
		CacheMode:         CacheModeReadOnly,
	})
}

// NewRealtimeClient 创建实时客户端（禁用缓存，总是网络获取）
func NewRealtimeClient() (*UnifiedClient, error) {
	return NewUnifiedClient(UnifiedConfig{
		PrimaryDataSource: DataSourceStockSDK,
		CacheMode:         CacheModeDisabled,
	})
}
