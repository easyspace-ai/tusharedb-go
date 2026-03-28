package multisource

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrKLineNotSupported 表示该数据源不提供 K 线；GetKLine 自动切换时不应记为失败。
var ErrKLineNotSupported = errors.New("multisource: kline not supported by this data source")

// DataSourceType 数据源类型
type DataSourceType string

const (
	// DataSourceEastMoney 东方财富
	DataSourceEastMoney DataSourceType = "eastmoney"
	// DataSourceSina 新浪财经
	DataSourceSina DataSourceType = "sina"
	// DataSourceTencent 腾讯财经
	DataSourceTencent DataSourceType = "tencent"
	// DataSourceTushare Tushare
	DataSourceTushare DataSourceType = "tushare"
	// DataSourceAKShare AKShare
	DataSourceAKShare DataSourceType = "akshare"
	// DataSourceXueqiu 雪球
	DataSourceXueqiu DataSourceType = "xueqiu"
	// DataSourceBaidu 百度股市通
	DataSourceBaidu DataSourceType = "baidu"
	// DataSourceTonghuashun 同花顺
	DataSourceTonghuashun DataSourceType = "tonghuashun"
)

// DataSource 数据源接口
type DataSource interface {
	// Name 返回数据源名称
	Name() string
	// Type 返回数据源类型
	Type() DataSourceType
	// Priority 返回优先级（数字越小优先级越高）
	Priority() int
	// HealthCheck 检查数据源是否可用
	HealthCheck(ctx context.Context) error
	// GetStockQuotes 获取股票实时行情
	GetStockQuotes(ctx context.Context, codes []string) ([]StockQuote, error)
	// GetKLine 获取K线数据
	GetKLine(ctx context.Context, code string, period string, adjust string, startDate string, endDate string) ([]KLineItem, error)
}

// StockQuote 股票行情数据
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

// KLineItem K线数据项
type KLineItem struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
	Amount float64 `json:"amount"`
}

// MultiSourceManager 多数据源管理器
type MultiSourceManager struct {
	sources   []DataSource
	sourceMap map[DataSourceType]DataSource
	mu        sync.RWMutex
	firstLoad bool
	loadOnce  sync.Once
	reqMgr    *RequestManager
}

var (
	globalManager *MultiSourceManager
	managerOnce   sync.Once
)

// GetMultiSourceManager 获取多数据源管理器单例
func GetMultiSourceManager() *MultiSourceManager {
	managerOnce.Do(func() {
		globalManager = NewMultiSourceManager()
	})
	return globalManager
}

// NewMultiSourceManager 创建多数据源管理器
func NewMultiSourceManager() *MultiSourceManager {
	mgr := &MultiSourceManager{
		sources:   make([]DataSource, 0),
		sourceMap: make(map[DataSourceType]DataSource),
		firstLoad: true,
		reqMgr:    GetRequestManager(),
	}

	// 注册默认数据源
	mgr.registerDefaultSources()

	return mgr
}

// registerDefaultSources 注册默认数据源（东财 / 新浪 / 腾讯 + 雪球 / 百度 / 同花顺 备选）
func (m *MultiSourceManager) registerDefaultSources() {
	m.AddSource(NewEastMoneySource(1, m.reqMgr))
	m.AddSource(NewSinaSource(2, m.reqMgr))
	m.AddSource(NewTencentSource(3, m.reqMgr))
	m.AddSource(NewXueqiuSource(4, m.reqMgr))
	m.AddSource(NewBaiduSource(5, m.reqMgr))
	m.AddSource(NewTonghuashunSource(6, m.reqMgr))
}

// addSourceUnsafe 追加数据源并按 Priority 排序；调用方必须已持有 m.mu（写锁）。
func (m *MultiSourceManager) addSourceUnsafe(source DataSource) {
	m.sources = append(m.sources, source)
	m.sourceMap[source.Type()] = source
	for i := range m.sources {
		for j := i + 1; j < len(m.sources); j++ {
			if m.sources[i].Priority() > m.sources[j].Priority() {
				m.sources[i], m.sources[j] = m.sources[j], m.sources[i]
			}
		}
	}
}

// AddSource 添加数据源
func (m *MultiSourceManager) AddSource(source DataSource) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addSourceUnsafe(source)
}

// IsFirstLoad 是否首次加载
func (m *MultiSourceManager) IsFirstLoad() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.firstLoad
}

// SetFirstLoadComplete 设置首次加载完成
func (m *MultiSourceManager) SetFirstLoadComplete() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.firstLoad = false
}

// GetAvailableSources 获取可用的数据源列表
func (m *MultiSourceManager) GetAvailableSources() []DataSource {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var available []DataSource
	for _, src := range m.sources {
		if m.reqMgr.IsSourceAvailable(src.Name()) {
			available = append(available, src)
		}
	}
	return available
}

// GetStockQuotes 获取股票行情（多数据源自动切换）
func (m *MultiSourceManager) GetStockQuotes(ctx context.Context, codes []string) ([]StockQuote, error) {
	sources := m.GetAvailableSources()
	if len(sources) == 0 {
		return nil, fmt.Errorf("no available data sources")
	}

	var lastErr error
	for _, src := range sources {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		result, err := src.GetStockQuotes(ctx, codes)
		if err == nil && len(result) > 0 {
			m.reqMgr.MarkSourceSuccess(src.Name())
			if m.IsFirstLoad() {
				m.SetFirstLoadComplete()
			}
			return result, nil
		}

		m.reqMgr.MarkSourceFailed(src.Name())
		lastErr = err
	}

	return nil, fmt.Errorf("all sources failed: %w", lastErr)
}

// GetKLine 获取K线数据（多数据源自动切换）
func (m *MultiSourceManager) GetKLine(ctx context.Context, code string, period string, adjust string, startDate string, endDate string) ([]KLineItem, error) {
	sources := m.GetAvailableSources()
	if len(sources) == 0 {
		return nil, fmt.Errorf("no available data sources")
	}

	var lastErr error
	for _, src := range sources {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		result, err := src.GetKLine(ctx, code, period, adjust, startDate, endDate)
		if err != nil {
			if errors.Is(err, ErrKLineNotSupported) {
				continue
			}
			m.reqMgr.MarkSourceFailed(src.Name())
			lastErr = err
			continue
		}
		if len(result) > 0 {
			m.reqMgr.MarkSourceSuccess(src.Name())
			if m.IsFirstLoad() {
				m.SetFirstLoadComplete()
			}
			return result, nil
		}

		m.reqMgr.MarkSourceFailed(src.Name())
		lastErr = err
	}

	return nil, fmt.Errorf("all sources failed: %w", lastErr)
}

// GetStockQuotesParallel 并行获取股票行情（首次加载时使用）
func (m *MultiSourceManager) GetStockQuotesParallel(ctx context.Context, codes []string) ([]StockQuote, error) {
	if !m.IsFirstLoad() {
		return m.GetStockQuotes(ctx, codes)
	}

	sources := m.GetAvailableSources()
	if len(sources) == 0 {
		return nil, fmt.Errorf("no available data sources")
	}

	type result struct {
		data []StockQuote
		err  error
		src  string
	}

	resultChan := make(chan result, len(sources))
	var wg sync.WaitGroup

	for _, src := range sources {
		wg.Add(1)
		go func(s DataSource) {
			defer wg.Done()
			data, err := s.GetStockQuotes(ctx, codes)
			resultChan <- result{data: data, err: err, src: s.Name()}
		}(src)
	}

	// 等待所有完成或第一个成功
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var lastErr error
	for res := range resultChan {
		if res.err == nil && len(res.data) > 0 {
			m.reqMgr.MarkSourceSuccess(res.src)
			m.SetFirstLoadComplete()
			return res.data, nil
		}
		lastErr = res.err
		m.reqMgr.MarkSourceFailed(res.src)
	}

	return nil, fmt.Errorf("all sources failed: %w", lastErr)
}

// eastMoney 返回已注册的东方财富完整数据源实现。
func (m *MultiSourceManager) eastMoney() (*EastMoneyFullSource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sourceMap[DataSourceEastMoney]
	if !ok {
		return nil, fmt.Errorf("eastmoney source not registered")
	}
	em, ok := s.(*EastMoneyFullSource)
	if !ok {
		return nil, fmt.Errorf("eastmoney source has unexpected type %T", s)
	}
	return em, nil
}

// GetStockQuotesBySource 仅从指定数据源拉取实时行情（不参与自动 failover）。
func (m *MultiSourceManager) GetStockQuotesBySource(ctx context.Context, src DataSourceType, codes []string) ([]StockQuote, error) {
	m.mu.RLock()
	s, ok := m.sourceMap[src]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("data source not registered: %s", src)
	}
	return s.GetStockQuotes(ctx, codes)
}

// GetKLineBySource 仅从指定数据源拉取 K 线。
func (m *MultiSourceManager) GetKLineBySource(ctx context.Context, src DataSourceType, code, period, adjust, startDate, endDate string) ([]KLineItem, error) {
	m.mu.RLock()
	s, ok := m.sourceMap[src]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("data source not registered: %s", src)
	}
	return s.GetKLine(ctx, code, period, adjust, startDate, endDate)
}

// GetMinuteData 分时数据。当前仅支持腾讯（source 为空或 DataSourceTencent）。
func (m *MultiSourceManager) GetMinuteData(ctx context.Context, src DataSourceType, code string) ([]MinuteBar, error) {
	if src != "" && src != DataSourceTencent {
		return nil, fmt.Errorf("minute data: source %s not supported (use %s)", src, DataSourceTencent)
	}
	m.mu.RLock()
	s, ok := m.sourceMap[DataSourceTencent]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("tencent source not registered")
	}
	tenc, ok := s.(*TencentFullSource)
	if !ok {
		return nil, fmt.Errorf("tencent source has unexpected type %T", s)
	}
	return tenc.GetMinuteData(ctx, code)
}

// GetNoticeContent 公告正文（东方财富）。
func (m *MultiSourceManager) GetNoticeContent(ctx context.Context, stockCode, artCode string) (string, error) {
	em, err := m.eastMoney()
	if err != nil {
		return "", err
	}
	return em.GetNoticeContent(ctx, stockCode, artCode)
}

// GetReportContent 研报正文（东方财富；部分研报可能仅能通过网页/PDF 查看）。
func (m *MultiSourceManager) GetReportContent(ctx context.Context, infoCode string) (string, error) {
	em, err := m.eastMoney()
	if err != nil {
		return "", err
	}
	return em.GetReportContent(ctx, infoCode)
}

// ========== 通用常量 ==========

// 缓存时间配置
const (
	CacheTimeQuote = 30 * time.Second // 行情数据缓存30秒
	CacheTimeIndex = 30 * time.Second // 指数数据缓存30秒
	CacheTimeKLine = 5 * time.Minute  // K线数据缓存5分钟
	CacheTimeNews  = 3 * time.Minute  // 新闻缓存3分钟
)
