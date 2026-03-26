package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/easyspace-ai/tusharedb-go/internal/provider/stocksdk"
)

// KlineService K线服务
type KlineService struct {
	client      *stocksdk.Client
	cache       *KlineCacheService
}

// NewKlineService 创建K线服务
func NewKlineService(client *stocksdk.Client, cache *KlineCacheService) *KlineService {
	return &KlineService{
		client: client,
		cache:  cache,
	}
}

// GetHistoryKline 获取历史K线（带缓存）
func (s *KlineService) GetHistoryKline(ctx context.Context, symbol, period, adjust, startDate, endDate string) ([]stocksdk.HistoryKline, error) {
	// 使用缓存服务获取数据
	if s.cache != nil {
		return s.cache.GetHistoryKline(ctx, symbol, period, adjust, startDate, endDate)
	}
	
	// 无缓存时直接调用API（降级）
	log.Printf("[KlineService] No cache available, fetching directly from API: %s", symbol)
	return s.fetchFromAPI(ctx, symbol, period, adjust, startDate, endDate)
}

// GetMinuteKline 获取分钟K线
func (s *KlineService) GetMinuteKline(ctx context.Context, symbol, period, adjust, startDate, endDate string) ([]stocksdk.MinuteKlineItem, error) {
	// 标准化周期
	var minutePeriod stocksdk.MinutePeriod
	switch period {
	case "5":
		minutePeriod = stocksdk.MinutePeriod5
	case "15":
		minutePeriod = stocksdk.MinutePeriod15
	case "30":
		minutePeriod = stocksdk.MinutePeriod30
	case "60":
		minutePeriod = stocksdk.MinutePeriod60
	default:
		minutePeriod = stocksdk.MinutePeriod1
	}
	
	// 标准化复权类型
	var adjustType stocksdk.AdjustType
	switch adjust {
	case "qfq":
		adjustType = stocksdk.AdjustTypeQFQ
	case "hfq":
		adjustType = stocksdk.AdjustTypeHFQ
	default:
		adjustType = stocksdk.AdjustTypeNone
	}
	
	// 设置默认日期范围
	if startDate == "" {
		startDate = "19700101"
	}
	if endDate == "" {
		endDate = "20500101"
	}
	
	return s.client.GetMinuteKline(ctx, symbol, &stocksdk.MinuteKlineOptions{
		Period:    minutePeriod,
		Adjust:    adjustType,
		StartDate: startDate,
		EndDate:   endDate,
	})
}

// GetTodayTimeline 获取当日分时数据
func (s *KlineService) GetTodayTimeline(ctx context.Context, symbol string) (*stocksdk.TodayTimelineResponse, error) {
	return s.client.GetTodayTimeline(ctx, symbol)
}

// BatchGetKline 批量获取K线数据（用于妖股扫描等批量场景）
func (s *KlineService) BatchGetKline(ctx context.Context, symbols []string, period, adjust string, minDays int) map[string][]stocksdk.HistoryKline {
	if s.cache != nil {
		return s.cache.BatchGetKline(ctx, symbols, period, adjust, minDays)
	}
	
	// 无缓存时的降级处理
	results := make(map[string][]stocksdk.HistoryKline)
	for _, symbol := range symbols {
		data, err := s.GetHistoryKline(ctx, symbol, period, adjust, "", "")
		if err != nil {
			log.Printf("[KlineService] Failed to get kline for %s: %v", symbol, err)
			continue
		}
		if len(data) >= minDays {
			results[symbol] = data
		}
	}
	return results
}

// PrefetchAllKlines 预取所有A股的K线数据（后台任务）
func (s *KlineService) PrefetchAllKlines(ctx context.Context, period, adjust string, onProgress func(completed, total int)) error {
	if s.cache != nil {
		return s.cache.PrefetchAllKlines(ctx, period, adjust, onProgress)
	}
	return nil
}

// GetCacheStats 获取缓存统计
func (s *KlineService) GetCacheStats() map[string]interface{} {
	if s.cache != nil {
		return s.cache.GetCacheStats()
	}
	return map[string]interface{}{"cache_enabled": false}
}

// SyncTradeCalendar 同步交易日历
func (s *KlineService) SyncTradeCalendar(ctx context.Context) error {
	if s.cache != nil {
		return s.cache.SyncTradeCalendar(ctx)
	}
	return nil
}

// CheckDataIntegrity 检查数据完整性
func (s *KlineService) CheckDataIntegrity(symbol, period, adjust string) (*DataIntegrityCheck, error) {
	if s.cache != nil {
		return s.cache.CheckDataIntegrity(symbol, period, adjust)
	}
	return nil, fmt.Errorf("cache not enabled")
}

// BatchCheckIntegrity 批量检查数据完整性
func (s *KlineService) BatchCheckIntegrity(symbols []string, period, adjust string) map[string]*DataIntegrityCheck {
	if s.cache != nil {
		return s.cache.BatchCheckIntegrity(symbols, period, adjust)
	}
	return nil
}

// RepairData 修复数据
func (s *KlineService) RepairData(ctx context.Context, symbol, period, adjust string) error {
	if s.cache != nil {
		return s.cache.RepairData(ctx, symbol, period, adjust)
	}
	return fmt.Errorf("cache not enabled")
}

// GetIncompleteDataList 获取数据不完整的股票列表
func (s *KlineService) GetIncompleteDataList(period, adjust string, threshold float64) ([]string, error) {
	if s.cache != nil {
		return s.cache.GetIncompleteDataList(period, adjust, threshold)
	}
	return nil, fmt.Errorf("cache not enabled")
}

// fetchFromAPI 从API获取数据
func (s *KlineService) fetchFromAPI(ctx context.Context, symbol, period, adjust, startDate, endDate string) ([]stocksdk.HistoryKline, error) {
	var klinePeriod stocksdk.KlinePeriod
	switch strings.ToLower(period) {
	case "weekly", "week":
		klinePeriod = stocksdk.KlinePeriodWeekly
	case "monthly", "month":
		klinePeriod = stocksdk.KlinePeriodMonthly
	default:
		klinePeriod = stocksdk.KlinePeriodDaily
	}
	
	var adjustType stocksdk.AdjustType
	switch adjust {
	case "qfq":
		adjustType = stocksdk.AdjustTypeQFQ
	case "hfq":
		adjustType = stocksdk.AdjustTypeHFQ
	default:
		adjustType = stocksdk.AdjustTypeNone
	}
	
	if startDate == "" {
		startDate = "19700101"
	}
	if endDate == "" {
		endDate = "20500101"
	}
	
	return s.client.GetHistoryKline(ctx, symbol, &stocksdk.HistoryKlineOptions{
		Period:    klinePeriod,
		Adjust:    adjustType,
		StartDate: startDate,
		EndDate:   endDate,
	})
}
