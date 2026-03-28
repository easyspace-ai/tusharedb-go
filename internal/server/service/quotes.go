package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/easyspace-ai/stock_api/internal/provider/stocksdk"
)

// QuotesService 行情服务
type QuotesService struct {
	client   *stocksdk.Client
	cacheSvc *CacheService
}

// NewQuotesService 创建行情服务
func NewQuotesService(client *stocksdk.Client, cacheSvc *CacheService) *QuotesService {
	return &QuotesService{
		client:   client,
		cacheSvc: cacheSvc,
	}
}

// GetFullQuotes 获取批量行情
func (s *QuotesService) GetFullQuotes(ctx context.Context, codes []string) ([]stocksdk.FullQuote, error) {
	if len(codes) == 0 {
		return nil, fmt.Errorf("codes cannot be empty")
	}
	return s.client.GetFullQuotes(ctx, codes)
}

// GetAllAShareQuotes 获取全部A股行情（使用缓存）
func (s *QuotesService) GetAllAShareQuotes(ctx context.Context) ([]stocksdk.FullQuote, error) {
	// 优先使用缓存服务获取
	if s.cacheSvc != nil {
		return s.cacheSvc.GetAllAShareQuotes(ctx)
	}

	// 降级：直接从 StockSDK 获取
	codes, err := s.client.GetAShareCodeList(ctx, false, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get A-share codes: %w", err)
	}

	// 分批获取行情，每批400只
	const batchSize = 400
	var allQuotes []stocksdk.FullQuote
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 限制并发数
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

			quotes, err := s.client.GetFullQuotes(ctx, batchCodes)
			if err != nil {
				return
			}

			mu.Lock()
			allQuotes = append(allQuotes, quotes...)
			mu.Unlock()
		}(batch)
	}

	wg.Wait()
	return allQuotes, nil
}

// GetAllAShareQuotesWithProgress 获取全部A股行情（带进度回调）
func (s *QuotesService) GetAllAShareQuotesWithProgress(
	ctx context.Context,
	onProgress func(completed, total int),
) ([]stocksdk.FullQuote, error) {
	// 优先使用缓存服务
	if s.cacheSvc != nil {
		return s.cacheSvc.GetAllAShareQuotesWithProgress(ctx, onProgress)
	}

	// 降级处理：直接从 StockSDK 获取
	codes, err := s.client.GetAShareCodeList(ctx, false, "")
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

			quotes, err := s.client.GetFullQuotes(ctx, batchCodes)
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
	return allQuotes, nil
}

// GetSimpleQuotes 获取简要行情
func (s *QuotesService) GetSimpleQuotes(ctx context.Context, codes []string) ([]stocksdk.SimpleQuote, error) {
	if len(codes) == 0 {
		return nil, fmt.Errorf("codes cannot be empty")
	}
	return s.client.GetSimpleQuotes(ctx, codes)
}
