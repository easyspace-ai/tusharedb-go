package tsdb

import (
	"fmt"
	"path/filepath"

	"github.com/easyspace-ai/stock_api/internal/config"
	"github.com/easyspace-ai/stock_api/internal/dataset"
	_ "github.com/easyspace-ai/stock_api/internal/dataset/builtin"
	"github.com/easyspace-ai/stock_api/internal/provider"
	"github.com/easyspace-ai/stock_api/internal/provider/stocksdk"
	"github.com/easyspace-ai/stock_api/internal/provider/tushare"
	"github.com/easyspace-ai/stock_api/internal/query/duckdb"
	"github.com/easyspace-ai/stock_api/internal/storage/meta"
	"github.com/easyspace-ai/stock_api/internal/syncer"
)

// Client 是 TushareDB 的主客户端
// 支持多数据源配置，提供数据下载、查询和筛选功能
type Client struct {
	cfg              config.Config
	primaryProvider  provider.DataProvider
	fallbackProvider provider.DataProvider
	registry         *dataset.Registry
	checkpoint       *meta.CheckpointStore
	engine           *duckdb.Engine
	syncer           *syncer.Syncer
}

// DefaultProviderFactory 默认的 Provider 工厂
var DefaultProviderFactory = map[DataSourceType]ProviderFactory{
	DataSourceTushare: func(cfg DataSourceConfig) (provider.DataProvider, error) {
		if cfg.TushareToken == "" {
			return nil, fmt.Errorf("tushare token is required")
		}
		return tushare.NewClient(tushare.Config{
			Token: cfg.TushareToken,
		}), nil
	},
	DataSourceStockSDK: func(cfg DataSourceConfig) (provider.DataProvider, error) {
		return stocksdk.NewClient(stocksdk.Config{
			APIKey: cfg.StockSDKAPIKey,
		}), nil
	},
}

// NewClient 创建 TushareDB 客户端
func NewClient(cfg Config) (*Client, error) {
	return NewClientWithFactory(cfg, DefaultProviderFactory)
}

// NewClientWithFactory 使用自定义 Provider 工厂创建客户端
func NewClientWithFactory(cfg Config, factories map[DataSourceType]ProviderFactory) (*Client, error) {
	// 设置默认数据源
	if cfg.PrimaryDataSource == "" {
		cfg.PrimaryDataSource = DataSourceTushare
	}

	// 向后兼容：如果设置了 Token 但没有设置 TushareToken，则使用 Token
	if cfg.TushareToken == "" && cfg.Token != "" {
		cfg.TushareToken = cfg.Token
	}

	// 归一化配置
	normalized, err := config.Normalize(config.Config{
		Token:      cfg.TushareToken,
		DataDir:    cfg.DataDir,
		DuckDBPath: cfg.DuckDBPath,
		TempDir:    cfg.TempDir,
		LogLevel:   cfg.LogLevel,
	})
	if err != nil {
		return nil, err
	}

	// 创建主数据源 Provider
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

	// 创建备用数据源 Provider (如果配置了)
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

	return &Client{
		cfg:              normalized,
		primaryProvider:  primaryProvider,
		fallbackProvider: fallbackProvider,
		registry:         registry,
		checkpoint:       checkpoint,
		engine:           engine,
		syncer:           s,
	}, nil
}

// PrimaryProvider 返回主数据源 Provider
func (c *Client) PrimaryProvider() provider.DataProvider {
	return c.primaryProvider
}

// FallbackProvider 返回备用数据源 Provider (如果有)
func (c *Client) FallbackProvider() provider.DataProvider {
	return c.fallbackProvider
}

// HasFallbackProvider 检查是否配置了备用数据源
func (c *Client) HasFallbackProvider() bool {
	return c.fallbackProvider != nil
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	return c.engine.Close()
}

func (c *Client) Downloader() *Downloader {
	return &Downloader{client: c}
}

func (c *Client) Reader() *Reader {
	return &Reader{client: c}
}

func (c *Client) Screener() *Screener {
	return &Screener{client: c}
}

func (c *Client) BacktestFeed() *BacktestFeed {
	return &BacktestFeed{client: c}
}
