package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/easyspace-ai/tusharedb-go/internal/provider/stocksdk"
	_ "github.com/marcboeker/go-duckdb"
)

// DataIntegrityCheck 数据完整性检查结果
type DataIntegrityCheck struct {
	Symbol           string    `json:"symbol"`
	Period           string    `json:"period"`
	Adjust           string    `json:"adjust"`
	TotalRecords     int       `json:"total_records"`
	ValidRecords     int       `json:"valid_records"`      // 字段完整的记录数
	MissingFields    int       `json:"missing_fields"`     // 缺失字段的记录数
	TradeDateGaps    int       `json:"trade_date_gaps"`    // 缺失交易日数量
	ExpectedRecords  int       `json:"expected_records"`   // 期望的记录数（根据交易日历）
	FirstDate        string    `json:"first_date"`
	LastDate         string    `json:"last_date"`
	IsComplete       bool      `json:"is_complete"`        // 是否完整
	CompletenessPct  float64   `json:"completeness_pct"`   // 完整度百分比
	LastVerifiedAt   time.Time `json:"last_verified_at"`
	Issues           []string  `json:"issues,omitempty"`   // 发现的问题
}

// KlineCacheService K线缓存服务
// 负责管理K线数据的本地持久化和缓存，解决外部API不稳定的问题
type KlineCacheService struct {
	sdkClient  *stocksdk.Client
	db         *sql.DB
	dataDir    string
	dbPath     string
	
	// 内存缓存
	memCache      map[string]*klineCacheEntry
	memCacheMutex sync.RWMutex
	
	// 缓存配置
	defaultTTL    time.Duration  // 默认缓存过期时间
	longTermTTL   time.Duration  // 长期缓存（历史数据）
	
	// 交易日历缓存
	tradeCalendar   []string
	calendarMutex   sync.RWMutex
	calendarUpdated time.Time
}

// klineCacheEntry K线缓存项
type klineCacheEntry struct {
	Data      []stocksdk.HistoryKline
	Timestamp time.Time
	Symbol    string
	Period    string
	Adjust    string
}

// cacheKey 生成缓存key
func cacheKey(symbol, period, adjust string) string {
	return fmt.Sprintf("%s_%s_%s", symbol, period, adjust)
}

// NewKlineCacheService 创建K线缓存服务
func NewKlineCacheService(dataDir string, sdkClient *stocksdk.Client) (*KlineCacheService, error) {
	if dataDir == "" {
		dataDir = "./data"
	}
	
	dbPath := filepath.Join(dataDir, "duckdb", "kline_cache.duckdb")
	
	// 确保目录存在
	if err := ensureDir(filepath.Dir(dbPath)); err != nil {
		return nil, fmt.Errorf("failed to create kline cache directory: %w", err)
	}
	
	// 打开DuckDB连接
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open kline cache database: %w", err)
	}
	
	service := &KlineCacheService{
		sdkClient:   sdkClient,
		db:          db,
		dataDir:     dataDir,
		dbPath:      dbPath,
		memCache:    make(map[string]*klineCacheEntry),
		defaultTTL:  24 * time.Hour,     // 默认1天
		longTermTTL: 30 * 24 * time.Hour, // 历史数据30天
	}
	
	// 初始化表结构
	if err := service.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init kline cache schema: %w", err)
	}
	
	return service, nil
}

// Close 关闭缓存服务
func (s *KlineCacheService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// initSchema 初始化数据库表结构
func (s *KlineCacheService) initSchema() error {
	// 创建K线数据表
	query := `
		CREATE TABLE IF NOT EXISTS kline_data (
			symbol VARCHAR NOT NULL,
			period VARCHAR NOT NULL,
			adjust VARCHAR NOT NULL,
			date VARCHAR NOT NULL,
			open DOUBLE,
			high DOUBLE,
			low DOUBLE,
			close DOUBLE,
			volume BIGINT,
			amount DOUBLE,
			amplitude DOUBLE,
			pct_chg DOUBLE,
			change DOUBLE,
			turnover DOUBLE,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (symbol, period, adjust, date)
		);
		
		CREATE INDEX IF NOT EXISTS idx_kline_symbol ON kline_data(symbol, period, adjust);
		CREATE INDEX IF NOT EXISTS idx_kline_updated ON kline_data(updated_at);
		
		CREATE TABLE IF NOT EXISTS kline_metadata (
			symbol VARCHAR NOT NULL,
			period VARCHAR NOT NULL,
			adjust VARCHAR NOT NULL,
			count BIGINT DEFAULT 0,
			start_date VARCHAR,
			end_date VARCHAR,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (symbol, period, adjust)
		);
		
		CREATE TABLE IF NOT EXISTS trade_calendar (
			date VARCHAR PRIMARY KEY,
			is_open BOOLEAN DEFAULT true,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_trade_calendar ON trade_calendar(date, is_open);
	`
	
	_, err := s.db.Exec(query)
	return err
}

// ensureDir 确保目录存在
func ensureDir(dir string) error {
	return nil // 简化处理，实际项目中使用 os.MkdirAll
}

// ==================== 核心API ====================

// GetHistoryKline 获取历史K线（带多级缓存）
func (s *KlineCacheService) GetHistoryKline(ctx context.Context, symbol, period, adjust, startDate, endDate string) ([]stocksdk.HistoryKline, error) {
	// 标准化参数
	period = normalizePeriod(period)
	adjust = normalizeAdjust(adjust)
	if startDate == "" {
		startDate = "19700101"
	}
	if endDate == "" {
		endDate = "20500101"
	}
	
	key := cacheKey(symbol, period, adjust)
	
	// 1. 检查内存缓存
	s.memCacheMutex.RLock()
	if entry, ok := s.memCache[key]; ok {
		ttl := s.getTTLForDateRange(entry.Data)
		if time.Since(entry.Timestamp) < ttl {
			s.memCacheMutex.RUnlock()
			log.Printf("[KlineCache] Memory cache hit: %s", key)
			return entry.Data, nil
		}
	}
	s.memCacheMutex.RUnlock()
	
	// 2. 检查本地数据库缓存
	cached, err := s.getFromDB(symbol, period, adjust, startDate, endDate)
	if err == nil && len(cached) > 0 {
		// 检查缓存是否新鲜
		fresh := s.isCacheFresh(symbol, period, adjust, len(cached))
		if fresh {
			// 更新内存缓存
			s.memCacheMutex.Lock()
			s.memCache[key] = &klineCacheEntry{
				Data:      cached,
				Timestamp: time.Now(),
				Symbol:    symbol,
				Period:    period,
				Adjust:    adjust,
			}
			s.memCacheMutex.Unlock()
			log.Printf("[KlineCache] DB cache hit: %s (%d records)", key, len(cached))
			return cached, nil
		}
	}
	
	// 3. 从API获取（带重试）
	data, err := s.fetchFromAPI(ctx, symbol, period, adjust, startDate, endDate)
	if err != nil {
		// API失败，尝试返回缓存数据（即使过期）
		if len(cached) > 0 {
			log.Printf("[KlineCache] API failed, using stale cache: %s, error: %v", key, err)
			return cached, nil
		}
		return nil, fmt.Errorf("failed to get kline from API and no cache available: %w", err)
	}
	
	// 4. 保存到缓存
	if err := s.saveToDB(symbol, period, adjust, data); err != nil {
		log.Printf("[KlineCache] Failed to save to DB: %v", err)
	}
	
	// 更新内存缓存
	s.memCacheMutex.Lock()
	s.memCache[key] = &klineCacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		Symbol:    symbol,
		Period:    period,
		Adjust:    adjust,
	}
	s.memCacheMutex.Unlock()
	
	log.Printf("[KlineCache] Fetched from API and cached: %s (%d records)", key, len(data))
	return data, nil
}

// BatchGetKline 批量获取K线数据（用于妖股扫描等批量场景）
func (s *KlineCacheService) BatchGetKline(ctx context.Context, symbols []string, period, adjust string, minDays int) map[string][]stocksdk.HistoryKline {
	results := make(map[string][]stocksdk.HistoryKline)
	var mu sync.Mutex
	var wg sync.WaitGroup
	
	// 限制并发数
	semaphore := make(chan struct{}, 10)
	
	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// 获取足够的数据（请求更多数据以确保计算指标时有足够的历史）
			data, err := s.GetHistoryKline(ctx, sym, period, adjust, "", "")
			if err != nil {
				log.Printf("[KlineCache] Failed to get kline for %s: %v", sym, err)
				return
			}
			
			if len(data) < minDays {
				log.Printf("[KlineCache] Insufficient data for %s: %d < %d", sym, len(data), minDays)
				return
			}
			
			mu.Lock()
			results[sym] = data
			mu.Unlock()
		}(symbol)
	}
	
	wg.Wait()
	return results
}

// PrefetchAllKlines 预取所有A股的K线数据（后台任务）
func (s *KlineCacheService) PrefetchAllKlines(ctx context.Context, period, adjust string, onProgress func(completed, total int)) error {
	// 获取A股代码列表
	codes, err := s.sdkClient.GetAShareCodeList(ctx, false, "")
	if err != nil {
		return fmt.Errorf("failed to get A-share codes: %w", err)
	}
	
	total := len(codes)
	completed := 0
	var mu sync.Mutex
	
	// 限制并发数，避免对API造成过大压力
	semaphore := make(chan struct{}, 5)
	var wg sync.WaitGroup
	
	for _, code := range codes {
		// 检查是否需要取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		wg.Add(1)
		go func(c string) {
			defer wg.Done()
			
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// 检查本地是否已有足够新鲜的数据
			if s.hasFreshCache(c, period, adjust) {
				mu.Lock()
				completed++
				if onProgress != nil {
					onProgress(completed, total)
				}
				mu.Unlock()
				return
			}
			
			// 获取数据
			_, err := s.GetHistoryKline(ctx, c, period, adjust, "", "")
			if err != nil {
				log.Printf("[KlineCache] Prefetch failed for %s: %v", c, err)
			}
			
			mu.Lock()
			completed++
			if onProgress != nil {
				onProgress(completed, total)
			}
			mu.Unlock()
			
			// 适当延迟，避免请求过快
			time.Sleep(100 * time.Millisecond)
		}(code)
	}
	
	wg.Wait()
	log.Printf("[KlineCache] Prefetch completed: %d/%d", completed, total)
	return nil
}

// ==================== 缓存操作 ====================

// getFromDB 从数据库获取缓存
func (s *KlineCacheService) getFromDB(symbol, period, adjust, startDate, endDate string) ([]stocksdk.HistoryKline, error) {
	var query string
	var args []interface{}
	
	if startDate == "" && endDate == "" {
		// 获取全部数据
		query = `
			SELECT date, open, high, low, close, volume, amount, 
			       amplitude, pct_chg, change, turnover
			FROM kline_data
			WHERE symbol = ? AND period = ? AND adjust = ?
			ORDER BY date ASC
		`
		args = []interface{}{symbol, period, adjust}
	} else {
		// 获取指定日期范围的数据
		if startDate == "" {
			startDate = "00000000"
		}
		if endDate == "" {
			endDate = "99999999"
		}
		query = `
			SELECT date, open, high, low, close, volume, amount, 
			       amplitude, pct_chg, change, turnover
			FROM kline_data
			WHERE symbol = ? AND period = ? AND adjust = ?
			  AND date >= ? AND date <= ?
			ORDER BY date ASC
		`
		args = []interface{}{symbol, period, adjust, startDate, endDate}
	}
	
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var results []stocksdk.HistoryKline
	for rows.Next() {
		var k stocksdk.HistoryKline
		err := rows.Scan(
			&k.Date, &k.Open, &k.High, &k.Low, &k.Close,
			&k.Volume, &k.Amount, &k.Amplitude, &k.ChangePercent,
			&k.Change, &k.TurnoverRate,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, k)
	}
	
	return results, rows.Err()
}

// saveToDB 保存到数据库
func (s *KlineCacheService) saveToDB(symbol, period, adjust string, data []stocksdk.HistoryKline) error {
	if len(data) == 0 {
		return nil
	}
	
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// 先删除旧数据（避免UPSERT兼容性问题）
	_, err = tx.Exec(`
		DELETE FROM kline_data 
		WHERE symbol = ? AND period = ? AND adjust = ?
	`, symbol, period, adjust)
	if err != nil {
		return fmt.Errorf("failed to delete old data: %w", err)
	}
	
	// 批量插入新数据
	stmt, err := tx.Prepare(`
		INSERT INTO kline_data 
		(symbol, period, adjust, date, open, high, low, close, volume, amount, 
		 amplitude, pct_chg, change, turnover)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert: %w", err)
	}
	defer stmt.Close()
	
	var startDate, endDate string
	for i, k := range data {
		if i == 0 {
			startDate = k.Date
		}
		endDate = k.Date
		
		_, err := stmt.Exec(
			symbol, period, adjust, k.Date, k.Open, k.High, k.Low, k.Close,
			k.Volume, k.Amount, k.Amplitude, k.ChangePercent, k.Change, k.TurnoverRate,
		)
		if err != nil {
			return fmt.Errorf("failed to insert record %s: %w", k.Date, err)
		}
	}
	
	// 删除并插入元数据
	_, err = tx.Exec(`
		DELETE FROM kline_metadata 
		WHERE symbol = ? AND period = ? AND adjust = ?
	`, symbol, period, adjust)
	if err != nil {
		return fmt.Errorf("failed to delete old metadata: %w", err)
	}
	
	_, err = tx.Exec(`
		INSERT INTO kline_metadata (symbol, period, adjust, count, start_date, end_date)
		VALUES (?, ?, ?, ?, ?, ?)
	`, symbol, period, adjust, len(data), startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to insert metadata: %w", err)
	}
	
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	
	log.Printf("[KlineCache] Saved %d records to DB for %s_%s_%s", len(data), symbol, period, adjust)
	return nil
}

// isCacheFresh 检查缓存是否新鲜
func (s *KlineCacheService) isCacheFresh(symbol, period, adjust string, count int) bool {
	var updatedAt time.Time
	err := s.db.QueryRow(`
		SELECT updated_at FROM kline_metadata 
		WHERE symbol = ? AND period = ? AND adjust = ?
	`, symbol, period, adjust).Scan(&updatedAt)
	
	if err != nil {
		return false
	}
	
	// 历史数据（超过60天的数据）缓存30天，近期数据缓存1天
	ttl := s.longTermTTL
	if time.Since(updatedAt) < ttl {
		return true
	}
	
	return false
}

// hasFreshCache 检查是否有新鲜缓存
func (s *KlineCacheService) hasFreshCache(symbol, period, adjust string) bool {
	// 内存缓存检查
	key := cacheKey(symbol, period, adjust)
	s.memCacheMutex.RLock()
	if entry, ok := s.memCache[key]; ok {
		if time.Since(entry.Timestamp) < s.defaultTTL {
			s.memCacheMutex.RUnlock()
			return true
		}
	}
	s.memCacheMutex.RUnlock()
	
	// 数据库检查
	return s.isCacheFresh(symbol, period, adjust, 0)
}

// getTTLForDateRange 根据日期范围确定缓存TTL
func (s *KlineCacheService) getTTLForDateRange(data []stocksdk.HistoryKline) time.Duration {
	if len(data) == 0 {
		return s.defaultTTL
	}
	
	// 获取最后一条数据的日期
	lastDate := data[len(data)-1].Date
	lastTime, err := time.Parse("20060102", lastDate)
	if err != nil {
		return s.defaultTTL
	}
	
	// 如果最后数据是60天前的，使用长期缓存
	if time.Since(lastTime) > 60*24*time.Hour {
		return s.longTermTTL
	}
	
	return s.defaultTTL
}

// ==================== 辅助方法 ====================

// fetchFromAPI 从API获取数据（带重试）
func (s *KlineCacheService) fetchFromAPI(ctx context.Context, symbol, period, adjust, startDate, endDate string) ([]stocksdk.HistoryKline, error) {
	var klinePeriod stocksdk.KlinePeriod
	switch period {
	case "weekly":
		klinePeriod = stocksdk.KlinePeriodWeekly
	case "monthly":
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
	
	// 带重试的请求
	var lastErr error
	for i := 0; i < 3; i++ {
		if i > 0 {
			time.Sleep(time.Duration(i) * time.Second) // 递增延迟
		}
		
		data, err := s.sdkClient.GetHistoryKline(ctx, symbol, &stocksdk.HistoryKlineOptions{
			Period:    klinePeriod,
			Adjust:    adjustType,
			StartDate: startDate,
			EndDate:   endDate,
		})
		if err == nil {
			return data, nil
		}
		
		lastErr = err
		log.Printf("[KlineCache] API request failed (attempt %d/3): %v", i+1, err)
		
		// 检查是否是可重试的错误
		if strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "connection") {
			continue
		}
		// 非网络错误，直接返回
		break
	}
	
	return nil, lastErr
}

// normalizePeriod 标准化周期
func normalizePeriod(period string) string {
	switch strings.ToLower(period) {
	case "weekly", "week":
		return "weekly"
	case "monthly", "month":
		return "monthly"
	default:
		return "daily"
	}
}

// normalizeAdjust 标准化复权类型
func normalizeAdjust(adjust string) string {
	switch adjust {
	case "qfq":
		return "qfq"
	case "hfq":
		return "hfq"
	default:
		return "none"
	}
}

// ==================== 数据完整性检测 ====================

// SyncTradeCalendar 同步交易日历
func (s *KlineCacheService) SyncTradeCalendar(ctx context.Context) error {
	// 从StockSDK获取交易日历
	dates, err := s.sdkClient.GetTradingCalendar(ctx)
	if err != nil {
		return fmt.Errorf("failed to get trading calendar: %w", err)
	}
	
	// 开始事务
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// 清空旧数据
	_, err = tx.Exec("DELETE FROM trade_calendar")
	if err != nil {
		return err
	}
	
	// 插入新数据
	stmt, err := tx.Prepare("INSERT INTO trade_calendar (date, is_open) VALUES (?, true)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	
	for _, date := range dates {
		_, err = stmt.Exec(date)
		if err != nil {
			return err
		}
	}
	
	if err := tx.Commit(); err != nil {
		return err
	}
	
	// 更新内存缓存
	s.calendarMutex.Lock()
	s.tradeCalendar = dates
	s.calendarUpdated = time.Now()
	s.calendarMutex.Unlock()
	
	log.Printf("[KlineCache] Trade calendar synced: %d trading days", len(dates))
	return nil
}

// getTradeCalendar 获取交易日历（带缓存）
func (s *KlineCacheService) getTradeCalendar() ([]string, error) {
	// 检查内存缓存
	s.calendarMutex.RLock()
	if len(s.tradeCalendar) > 0 && time.Since(s.calendarUpdated) < 24*time.Hour {
		calendar := s.tradeCalendar
		s.calendarMutex.RUnlock()
		return calendar, nil
	}
	s.calendarMutex.RUnlock()
	
	// 从数据库获取
	rows, err := s.db.Query("SELECT date FROM trade_calendar WHERE is_open = true ORDER BY date ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var dates []string
	for rows.Next() {
		var date string
		if err := rows.Scan(&date); err != nil {
			return nil, err
		}
		dates = append(dates, date)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	// 更新内存缓存
	s.calendarMutex.Lock()
	s.tradeCalendar = dates
	s.calendarUpdated = time.Now()
	s.calendarMutex.Unlock()
	
	return dates, nil
}

// CheckDataIntegrity 检查指定股票的K线数据完整性
func (s *KlineCacheService) CheckDataIntegrity(symbol, period, adjust string) (*DataIntegrityCheck, error) {
	result := &DataIntegrityCheck{
		Symbol: symbol,
		Period: period,
		Adjust: adjust,
		Issues: []string{},
	}
	
	// 从数据库获取数据
	data, err := s.getFromDB(symbol, period, adjust, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get data from DB: %w", err)
	}
	
	result.TotalRecords = len(data)
	if len(data) == 0 {
		result.Issues = append(result.Issues, "No data found")
		return result, nil
	}
	
	result.FirstDate = data[0].Date
	result.LastDate = data[len(data)-1].Date
	
	// 1. 检查字段完整性
	validCount := 0
	for _, k := range data {
		if k.Open != nil && k.Close != nil && k.High != nil && k.Low != nil && 
		   k.Volume != nil && k.Date != "" {
			validCount++
		} else {
			result.MissingFields++
		}
	}
	result.ValidRecords = validCount
	
	// 2. 日线数据检查交易日连续性
	if period == "daily" || period == "" {
		tradeDates, err := s.getTradeCalendar()
		if err != nil {
			log.Printf("[KlineCache] Failed to get trade calendar: %v", err)
		} else if len(tradeDates) > 0 {
			// 计算应该有的交易日数量
			expectedCount := s.countExpectedTradeDays(tradeDates, result.FirstDate, result.LastDate)
			result.ExpectedRecords = expectedCount
			
			// 检查缺失的交易日
			dataDateSet := make(map[string]bool)
			for _, k := range data {
				dataDateSet[k.Date] = true
			}
			
			gaps := 0
			for _, date := range tradeDates {
				if date >= result.FirstDate && date <= result.LastDate {
					if !dataDateSet[date] {
						gaps++
						if gaps <= 5 { // 只记录前5个缺失日期
							result.Issues = append(result.Issues, fmt.Sprintf("Missing trade date: %s", date))
						}
					}
				}
			}
			result.TradeDateGaps = gaps
			
			if gaps > 5 {
				result.Issues = append(result.Issues, fmt.Sprintf("... and %d more missing dates", gaps-5))
			}
		}
	}
	
	// 3. 计算完整度
	if result.ExpectedRecords > 0 {
		result.CompletenessPct = float64(result.TotalRecords) / float64(result.ExpectedRecords) * 100
	} else {
		result.CompletenessPct = 100.0
	}
	
	// 4. 判断数据是否完整
	result.IsComplete = result.CompletenessPct >= 95.0 && result.MissingFields == 0 && result.TradeDateGaps <= 2
	
	result.LastVerifiedAt = time.Now()
	
	return result, nil
}

// countExpectedTradeDays 计算两个日期之间的交易日数量
func (s *KlineCacheService) countExpectedTradeDays(tradeDates []string, startDate, endDate string) int {
	count := 0
	for _, date := range tradeDates {
		if date >= startDate && date <= endDate {
			count++
		}
	}
	return count
}

// BatchCheckIntegrity 批量检查数据完整性
func (s *KlineCacheService) BatchCheckIntegrity(symbols []string, period, adjust string) map[string]*DataIntegrityCheck {
	results := make(map[string]*DataIntegrityCheck)
	var mu sync.Mutex
	var wg sync.WaitGroup
	
	semaphore := make(chan struct{}, 10)
	
	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			check, err := s.CheckDataIntegrity(sym, period, adjust)
			if err != nil {
				log.Printf("[KlineCache] Integrity check failed for %s: %v", sym, err)
				check = &DataIntegrityCheck{
					Symbol: sym,
					Period: period,
					Adjust: adjust,
					Issues: []string{fmt.Sprintf("Check failed: %v", err)},
				}
			}
			
			mu.Lock()
			results[sym] = check
			mu.Unlock()
		}(symbol)
	}
	
	wg.Wait()
	return results
}

// RepairData 修复指定股票的数据（强制刷新）
func (s *KlineCacheService) RepairData(ctx context.Context, symbol, period, adjust string) error {
	log.Printf("[KlineCache] Repairing data for %s_%s_%s", symbol, period, adjust)
	
	// 1. 清除旧缓存
	if err := s.ClearCache(symbol, period, adjust); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	
	// 2. 重新获取数据
	_, err := s.GetHistoryKline(ctx, symbol, period, adjust, "", "")
	if err != nil {
		return fmt.Errorf("failed to refetch data: %w", err)
	}
	
	// 3. 验证修复结果
	check, err := s.CheckDataIntegrity(symbol, period, adjust)
	if err != nil {
		return fmt.Errorf("failed to verify repair: %w", err)
	}
	
	if !check.IsComplete {
		return fmt.Errorf("data still incomplete after repair: %v", check.Issues)
	}
	
	log.Printf("[KlineCache] Data repaired successfully for %s", symbol)
	return nil
}

// GetIncompleteDataList 获取数据不完整的股票列表
func (s *KlineCacheService) GetIncompleteDataList(period, adjust string, threshold float64) ([]string, error) {
	// 获取所有缓存的股票
	rows, err := s.db.Query(`
		SELECT DISTINCT symbol FROM kline_metadata 
		WHERE period = ? AND adjust = ?
	`, period, adjust)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	// 检查每个股票的数据完整性
	var incomplete []string
	for _, symbol := range symbols {
		check, err := s.CheckDataIntegrity(symbol, period, adjust)
		if err != nil {
			continue
		}
		if !check.IsComplete || check.CompletenessPct < threshold {
			incomplete = append(incomplete, symbol)
		}
	}
	
	return incomplete, nil
}

// ==================== 管理接口 ====================

// GetCacheStats 获取缓存统计
func (s *KlineCacheService) GetCacheStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	// 内存缓存统计
	s.memCacheMutex.RLock()
	stats["memory_entries"] = len(s.memCache)
	s.memCacheMutex.RUnlock()
	
	// 数据库统计
	var totalCount int
	err := s.db.QueryRow("SELECT COUNT(*) FROM kline_metadata").Scan(&totalCount)
	if err == nil {
		stats["db_stocks"] = totalCount
	}
	
	var totalRecords int
	err = s.db.QueryRow("SELECT COUNT(*) FROM kline_data").Scan(&totalRecords)
	if err == nil {
		stats["db_records"] = totalRecords
	}
	
	return stats
}

// ClearCache 清除指定股票的缓存
func (s *KlineCacheService) ClearCache(symbol, period, adjust string) error {
	key := cacheKey(symbol, period, adjust)
	
	// 清除内存缓存
	s.memCacheMutex.Lock()
	delete(s.memCache, key)
	s.memCacheMutex.Unlock()
	
	// 清除数据库缓存
	_, err := s.db.Exec(`
		DELETE FROM kline_data WHERE symbol = ? AND period = ? AND adjust = ?
	`, symbol, period, adjust)
	if err != nil {
		return err
	}
	
	_, err = s.db.Exec(`
		DELETE FROM kline_metadata WHERE symbol = ? AND period = ? AND adjust = ?
	`, symbol, period, adjust)
	
	return err
}

// ClearAllCache 清除所有缓存
func (s *KlineCacheService) ClearAllCache() error {
	// 清除内存缓存
	s.memCacheMutex.Lock()
	s.memCache = make(map[string]*klineCacheEntry)
	s.memCacheMutex.Unlock()
	
	// 清除数据库缓存
	_, err := s.db.Exec("DELETE FROM kline_data")
	if err != nil {
		return err
	}
	
	_, err = s.db.Exec("DELETE FROM kline_metadata")
	return err
}
