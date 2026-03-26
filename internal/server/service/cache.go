package service

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/easyspace-ai/tusharedb-go/internal/provider/stocksdk"
	"github.com/easyspace-ai/tusharedb-go/pkg/tsdb"
)

// CacheService 数据缓存服务
// 负责管理全市场数据的持久化和缓存
type CacheService struct {
	client     *tsdb.UnifiedClient
	stockSDK   *stocksdk.Client
	dataDir    string
	
	// 内存缓存
	cache      map[string]interface{}
	cacheTime  map[string]time.Time
	cacheMutex sync.RWMutex
	
	// 缓存过期时间
	ttl        time.Duration
}

// NewCacheService 创建缓存服务
func NewCacheService(dataDir string, stockSDK *stocksdk.Client) (*CacheService, error) {
	if dataDir == "" {
		dataDir = "./data"
	}

	// 创建 UnifiedClient，使用 Auto 模式
	client, err := tsdb.NewAutoClient(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create unified client: %w", err)
	}

	return &CacheService{
		client:    client,
		stockSDK:  stockSDK,
		dataDir:   dataDir,
		cache:     make(map[string]interface{}),
		cacheTime: make(map[string]time.Time),
		ttl:       5 * time.Minute, // 默认5分钟缓存
	}, nil
}

// Close 关闭缓存服务
func (s *CacheService) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// ==================== 全市场数据缓存 ====================

// GetAllAShareQuotes 获取全部A股行情（带缓存）
func (s *CacheService) GetAllAShareQuotes(ctx context.Context) ([]stocksdk.FullQuote, error) {
	// 1. 先检查内存缓存
	cacheKey := "all_a_share_quotes"
	s.cacheMutex.RLock()
	if cached, ok := s.cache[cacheKey]; ok {
		if cacheTime, ok := s.cacheTime[cacheKey]; ok {
			if time.Since(cacheTime) < s.ttl {
				s.cacheMutex.RUnlock()
				if quotes, ok := cached.([]stocksdk.FullQuote); ok {
					return quotes, nil
				}
			}
		}
	}
	s.cacheMutex.RUnlock()

	// 2. 尝试从 UnifiedClient 获取（本地缓存）
	// 注意：UnifiedClient 目前主要支持日线数据，实时行情需要从 StockSDK 获取
	// 这里我们先实现直接获取，后续可以扩展本地缓存逻辑

	// 3. 从 StockSDK 获取最新数据
	codes, err := s.stockSDK.GetAShareCodeList(ctx, false, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get A-share codes: %w", err)
	}

	// 分批获取行情
	const batchSize = 400
	var allQuotes []stocksdk.FullQuote
	var mu sync.Mutex
	var wg sync.WaitGroup
	
	semaphore := make(chan struct{}, 5)
	
	for i := 0; i < len(codes); i += batchSize {
		end := i + batchSize
		if end > len(codes) {
			end = len(codes)
		}
		batch := codes[i:end]
		
		wg.Add(1)
		go func(batchCodes []string) {
			defer wg.Done()
			
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			quotes, err := s.stockSDK.GetFullQuotes(ctx, batchCodes)
			if err != nil {
				return
			}
			
			mu.Lock()
			allQuotes = append(allQuotes, quotes...)
			mu.Unlock()
		}(batch)
	}
	
	wg.Wait()

	// 4. 更新内存缓存
	s.cacheMutex.Lock()
	s.cache[cacheKey] = allQuotes
	s.cacheTime[cacheKey] = time.Now()
	s.cacheMutex.Unlock()

	return allQuotes, nil
}

// GetAllAShareQuotesWithProgress 获取全部A股行情（带进度回调）
func (s *CacheService) GetAllAShareQuotesWithProgress(
	ctx context.Context,
	onProgress func(completed, total int),
) ([]stocksdk.FullQuote, error) {
	codes, err := s.stockSDK.GetAShareCodeList(ctx, false, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get A-share codes: %w", err)
	}

	total := len(codes)
	const batchSize = 400
	var allQuotes []stocksdk.FullQuote
	var mu sync.Mutex
	var wg sync.WaitGroup
	
	semaphore := make(chan struct{}, 5)
	completed := 0
	
	for i := 0; i < len(codes); i += batchSize {
		end := i + batchSize
		if end > len(codes) {
			end = len(codes)
		}
		batch := codes[i:end]
		
		wg.Add(1)
		go func(batchCodes []string) {
			defer wg.Done()
			
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			quotes, err := s.stockSDK.GetFullQuotes(ctx, batchCodes)
			if err != nil {
				return
			}
			
			mu.Lock()
			allQuotes = append(allQuotes, quotes...)
			completed += len(batchCodes)
			if onProgress != nil {
				onProgress(completed, total)
			}
			mu.Unlock()
		}(batch)
	}
	
	wg.Wait()

	// 更新缓存
	s.cacheMutex.Lock()
	s.cache["all_a_share_quotes"] = allQuotes
	s.cacheTime["all_a_share_quotes"] = time.Now()
	s.cacheMutex.Unlock()

	return allQuotes, nil
}

// ==================== 历史数据缓存 ====================

// SyncStockBasic 同步股票基础信息到本地
func (s *CacheService) SyncStockBasic(ctx context.Context) error {
	return s.client.SyncCore(ctx)
}

// SyncDailyRange 同步日线数据范围
func (s *CacheService) SyncDailyRange(ctx context.Context, startDate, endDate string) error {
	return s.client.SyncDailyRange(ctx, startDate, endDate)
}

// GetStockDaily 从本地缓存获取日线数据
func (s *CacheService) GetStockDaily(ctx context.Context, tsCode, startDate, endDate string, adjust tsdb.AdjustType) (*tsdb.DataFrame, error) {
	return s.client.GetStockDaily(ctx, tsCode, startDate, endDate, adjust)
}

// GetMultipleStocksDaily 获取多只股票日线数据
func (s *CacheService) GetMultipleStocksDaily(ctx context.Context, tsCodes []string, startDate, endDate string, adjust tsdb.AdjustType) (*tsdb.DataFrame, error) {
	return s.client.GetMultipleStocksDaily(ctx, tsCodes, startDate, endDate, adjust)
}

// ==================== 缓存管理 ====================

// ClearCache 清除指定key的缓存
func (s *CacheService) ClearCache(key string) {
	s.cacheMutex.Lock()
	delete(s.cache, key)
	delete(s.cacheTime, key)
	s.cacheMutex.Unlock()
}

// ClearAllCache 清除所有缓存
func (s *CacheService) ClearAllCache() {
	s.cacheMutex.Lock()
	s.cache = make(map[string]interface{})
	s.cacheTime = make(map[string]time.Time)
	s.cacheMutex.Unlock()
}

// SetTTL 设置缓存过期时间
func (s *CacheService) SetTTL(ttl time.Duration) {
	s.ttl = ttl
}

// GetDataDir 获取数据目录
func (s *CacheService) GetDataDir() string {
	return s.dataDir
}

// GetDuckDBPath 获取DuckDB路径
func (s *CacheService) GetDuckDBPath() string {
	return filepath.Join(s.dataDir, "duckdb", "tusharedb.duckdb")
}

// GetLastSyncDate 获取最后同步日期
func (s *CacheService) GetLastSyncDate(dataset string) (string, bool) {
	return s.client.GetLastSyncDate(dataset)
}
