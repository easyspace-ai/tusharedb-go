package stockapi

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/easyspace-ai/tusharedb-go/internal/provider/stocksdk"
)

type Client struct {
	sdk              *stocksdk.Client
	config           Config
	mu               sync.RWMutex
	mem              map[string]cacheEntry
	historyPathLocks sync.Map // string (parquet path) -> *sync.Mutex
}

type CacheMode string

const (
	CacheModeDisabled CacheMode = "disabled"
	CacheModeReadOnly CacheMode = "readonly"
	CacheModeAuto     CacheMode = "auto"
)

type Config struct {
	Timeout     time.Duration
	Retries     int
	RetryWait   time.Duration
	UserAgent   string
	CacheMode   CacheMode
	DataDir     string
	QuotesTTL   time.Duration
	HistoryTTL  time.Duration
	TimelineTTL time.Duration
	BatchSize   int
	Concurrency int
}

type cacheEntry struct {
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

type FullQuote struct {
	Code                 string   `json:"code"`
	Name                 string   `json:"name"`
	Price                float64  `json:"price"`
	PrevClose            float64  `json:"prevClose"`
	Open                 float64  `json:"open"`
	High                 float64  `json:"high"`
	Low                  float64  `json:"low"`
	Volume               float64  `json:"volume"`
	Amount               float64  `json:"amount"`
	Change               float64  `json:"change"`
	ChangePercent        float64  `json:"changePercent"`
	TurnoverRate         *float64 `json:"turnoverRate,omitempty"`
	PE                   *float64 `json:"pe,omitempty"`
	PB                   *float64 `json:"pb,omitempty"`
	Amplitude            *float64 `json:"amplitude,omitempty"`
	CirculatingMarketCap *float64 `json:"circulatingMarketCap,omitempty"`
	TotalMarketCap       *float64 `json:"totalMarketCap,omitempty"`
	VolumeRatio          *float64 `json:"volumeRatio,omitempty"`
	AvgPrice             *float64 `json:"avgPrice,omitempty"`
	Time                 string   `json:"time"`
}

type KlineData struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
	Amount float64 `json:"amount"`
}

type TimelineData struct {
	Time     string  `json:"time"`
	Price    float64 `json:"price"`
	AvgPrice float64 `json:"avgPrice"`
	Volume   float64 `json:"volume"`
}

type TimelineResponse struct {
	Symbol    string         `json:"symbol"`
	PrevClose float64        `json:"prevClose"`
	Data      []TimelineData `json:"data"`
}

type TimelineBatchResult struct {
	Success map[string]*TimelineResponse `json:"success"`
	Failed  map[string]string            `json:"failed,omitempty"`
}

// PrewarmReport is filled when PrewarmOptions.Report is non-nil.
type PrewarmReport struct {
	QuotesOK bool

	HistoryTotal   int
	HistorySuccess int
	HistoryFailed  int
	// ValidationWarnSymbols is how many symbols had at least one daily validation warning.
	ValidationWarnSymbols int

	FailedSymbols []string
}

type PrewarmOptions struct {
	WarmQuotes         bool
	WarmHistory        bool
	HistoryPeriod      string
	HistoryAdjust      string
	HistoryStartDate   string
	HistoryEndDate     string
	HistoryConcurrency int

	// Report receives aggregate stats (optional).
	Report *PrewarmReport

	// OnLog is optional; must be safe for concurrent use (e.g. log.Printf).
	OnLog func(msg string)

	// HistoryProgressEvery logs a progress line every N completed symbols (0 = never).
	HistoryProgressEvery int

	// HistoryMaxRetries retries per symbol when API errors or returns empty bars (default 3).
	HistoryMaxRetries int
	// HistoryRetryWait base backoff; attempt k waits k * this duration (default 2s).
	HistoryRetryWait time.Duration

	// ValidateHistory runs daily completeness checks after each successful fetch (warnings only).
	ValidateHistory bool
	// MinDailyBars minimum bar count for validation warning (default 60 when 0).
	MinDailyBars int
}

func DefaultPrewarmOptions() PrewarmOptions {
	return PrewarmOptions{
		WarmQuotes:         true,
		WarmHistory:        true,
		HistoryPeriod:      "daily",
		HistoryAdjust:      "qfq",
		HistoryStartDate:   "",
		HistoryEndDate:     "",
		HistoryConcurrency: 8,
	}
}

func DefaultConfig() Config {
	defaultDir := defaultDataDir()
	return Config{
		Timeout:     30 * time.Second,
		Retries:     3,
		RetryWait:   1 * time.Second,
		UserAgent:   "tusharedb-go/stockapi",
		CacheMode:   CacheModeAuto,
		DataDir:     defaultDir,
		QuotesTTL:   5 * time.Second,
		HistoryTTL:  24 * time.Hour,
		TimelineTTL: 5 * time.Second,
		BatchSize:   400,
		Concurrency: 5,
	}
}

func NewClient() *Client {
	client, _ := NewClientWithConfig(DefaultConfig())
	return client
}

func NewClientWithConfig(cfg Config) (*Client, error) {
	def := DefaultConfig()
	if cfg.Timeout <= 0 {
		cfg.Timeout = def.Timeout
	}
	if cfg.Retries <= 0 {
		cfg.Retries = def.Retries
	}
	if cfg.RetryWait <= 0 {
		cfg.RetryWait = def.RetryWait
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = def.UserAgent
	}
	if cfg.CacheMode == "" {
		cfg.CacheMode = def.CacheMode
	}
	if cfg.DataDir == "" {
		cfg.DataDir = def.DataDir
	}
	if cfg.QuotesTTL <= 0 {
		cfg.QuotesTTL = def.QuotesTTL
	}
	if cfg.HistoryTTL <= 0 {
		cfg.HistoryTTL = def.HistoryTTL
	}
	if cfg.TimelineTTL <= 0 {
		cfg.TimelineTTL = def.TimelineTTL
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = def.BatchSize
	}
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = def.Concurrency
	}

	c := &Client{
		sdk: stocksdk.NewClient(stocksdk.Config{
			Timeout:   cfg.Timeout,
			Retries:   cfg.Retries,
			RetryWait: cfg.RetryWait,
			UserAgent: cfg.UserAgent,
		}),
		config: cfg,
		mem:    make(map[string]cacheEntry),
	}
	if cfg.CacheMode != CacheModeDisabled {
		if err := os.MkdirAll(c.cacheDir(), 0o755); err != nil {
			return nil, err
		}
		if err := os.MkdirAll(c.historyLakeDir(), 0o755); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *Client) GetAllAShareQuotes(ctx context.Context) ([]FullQuote, error) {
	cacheKey := "quotes_all"
	var cached []FullQuote
	if c.getCache(cacheKey, c.config.QuotesTTL, &cached) {
		return cached, nil
	}
	if c.config.CacheMode == CacheModeReadOnly {
		return nil, errors.New("quotes cache miss in readonly mode")
	}

	codes, err := c.sdk.GetAShareCodeList(ctx, false, "")
	if err != nil {
		return nil, err
	}
	var (
		allQuotes []stocksdk.FullQuote
		mu        sync.Mutex
		wg        sync.WaitGroup
		sem       = make(chan struct{}, c.config.Concurrency)
	)
	for i := 0; i < len(codes); i += c.config.BatchSize {
		end := i + c.config.BatchSize
		if end > len(codes) {
			end = len(codes)
		}
		batch := codes[i:end]
		wg.Add(1)
		go func(batchCodes []string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			quotes, err := c.sdk.GetFullQuotes(ctx, batchCodes)
			if err != nil {
				return
			}
			mu.Lock()
			allQuotes = append(allQuotes, quotes...)
			mu.Unlock()
		}(batch)
	}
	wg.Wait()
	raw := allQuotes
	result := make([]FullQuote, 0, len(raw))
	for _, q := range raw {
		result = append(result, FullQuote{
			Code:                 q.Code,
			Name:                 q.Name,
			Price:                q.Price,
			PrevClose:            q.PrevClose,
			Open:                 q.Open,
			High:                 q.High,
			Low:                  q.Low,
			Volume:               q.Volume,
			Amount:               q.Amount,
			Change:               q.Change,
			ChangePercent:        q.ChangePercent,
			TurnoverRate:         q.TurnoverRate,
			PE:                   q.PE,
			PB:                   q.PB,
			Amplitude:            q.Amplitude,
			CirculatingMarketCap: q.CirculatingMarketCap,
			TotalMarketCap:       q.TotalMarketCap,
			VolumeRatio:          q.VolumeRatio,
			AvgPrice:             q.AvgPrice,
			Time:                 q.Time,
		})
	}
	c.setCache(cacheKey, result)
	return result, nil
}

func (c *Client) GetHistoryKline(ctx context.Context, symbol, period, adjust, startDate, endDate string) ([]KlineData, error) {
	return c.getHistoryKlineWithParquet(ctx, symbol, period, adjust, startDate, endDate)
}

func (c *Client) GetTodayTimeline(ctx context.Context, symbol string) (*TimelineResponse, error) {
	cacheKey := "timeline_" + symbol
	var cached TimelineResponse
	if c.getCache(cacheKey, c.config.TimelineTTL, &cached) {
		return &cached, nil
	}
	if c.config.CacheMode == CacheModeReadOnly {
		return nil, errors.New("timeline cache miss in readonly mode")
	}

	raw, err := c.fetchTodayTimelineWithRetry(ctx, symbol)
	if err != nil {
		return nil, err
	}
	result := &TimelineResponse{
		Symbol:    symbol,
		PrevClose: raw.PrevClose,
		Data:      make([]TimelineData, 0, len(raw.Data)),
	}
	for _, item := range raw.Data {
		result.Data = append(result.Data, TimelineData{
			Time:     item.Time,
			Price:    item.Price,
			AvgPrice: item.AvgPrice,
			Volume:   item.Volume,
		})
	}
	c.setCache(cacheKey, result)
	return result, nil
}

func (c *Client) GetTodayTimelineBatch(ctx context.Context, symbols []string) *TimelineBatchResult {
	result := &TimelineBatchResult{
		Success: make(map[string]*TimelineResponse),
		Failed:  make(map[string]string),
	}
	if len(symbols) == 0 {
		return result
	}

	uniq := make([]string, 0, len(symbols))
	seen := make(map[string]struct{}, len(symbols))
	for _, raw := range symbols {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		uniq = append(uniq, s)
	}
	if len(uniq) == 0 {
		return result
	}

	// 批量缓存：同一交易日内，同一 symbols 集合直接命中，减少尾盘重复请求耗时。
	batchCacheKey := c.timelineBatchCacheKey(uniq)
	var cached TimelineBatchResult
	if c.getCache(batchCacheKey, c.timelineBatchTTL(), &cached) {
		if cached.Success != nil {
			result.Success = cached.Success
		}
		if len(cached.Failed) > 0 {
			result.Failed = cached.Failed
		} else {
			result.Failed = nil
		}
		return result
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, c.config.Concurrency)
	var mu sync.Mutex
	for _, symbol := range uniq {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			timeline, err := c.GetTodayTimeline(ctx, sym)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				result.Failed[sym] = err.Error()
				return
			}
			result.Success[sym] = timeline
		}(symbol)
	}
	wg.Wait()
	if len(result.Failed) == 0 {
		result.Failed = nil
	}
	c.setCache(batchCacheKey, result)
	return result
}

func (c *Client) timelineBatchTTL() time.Duration {
	ttl := c.config.TimelineTTL
	if ttl < 2*time.Minute {
		return 2 * time.Minute
	}
	return ttl
}

func (c *Client) timelineBatchCacheKey(symbols []string) string {
	sorted := make([]string, len(symbols))
	copy(sorted, symbols)
	sort.Strings(sorted)
	return fmt.Sprintf("timeline_batch_%s_%s", time.Now().Format("20060102"), strings.Join(sorted, ","))
}

func (c *Client) Prewarm(ctx context.Context, options PrewarmOptions) error {
	opts := DefaultPrewarmOptions()
	if !options.WarmQuotes && !options.WarmHistory {
		return nil
	}
	if options.HistoryPeriod != "" {
		opts.HistoryPeriod = options.HistoryPeriod
	}
	if options.HistoryAdjust != "" {
		opts.HistoryAdjust = options.HistoryAdjust
	}
	if options.HistoryStartDate != "" {
		opts.HistoryStartDate = options.HistoryStartDate
	}
	if options.HistoryEndDate != "" {
		opts.HistoryEndDate = options.HistoryEndDate
	}
	if options.HistoryConcurrency > 0 {
		opts.HistoryConcurrency = options.HistoryConcurrency
	}
	opts.WarmQuotes = options.WarmQuotes
	opts.WarmHistory = options.WarmHistory

	report := options.Report
	logf := func(msg string) {
		if options.OnLog != nil {
			options.OnLog(msg)
		}
	}
	maxRetries := options.HistoryMaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	retryWait := options.HistoryRetryWait
	if retryWait <= 0 {
		retryWait = 2 * time.Second
	}
	progressEvery := options.HistoryProgressEvery
	minBars := options.MinDailyBars
	if minBars <= 0 {
		minBars = 60
	}
	validate := options.ValidateHistory
	periodNorm := normalizePeriod(opts.HistoryPeriod)

	if opts.WarmQuotes {
		logf("stage=quotes msg=fetching_all_ashare")
		if _, err := c.GetAllAShareQuotes(ctx); err != nil {
			return fmt.Errorf("prewarm quotes failed: %w", err)
		}
		if report != nil {
			report.QuotesOK = true
		}
		logf("stage=quotes msg=done")
	}
	if !opts.WarmHistory {
		return nil
	}

	codes, err := c.sdk.GetAShareCodeList(ctx, false, "")
	if err != nil {
		return fmt.Errorf("load symbol list failed: %w", err)
	}
	syms := make([]string, 0, len(codes))
	for _, code := range codes {
		s := toSymbol(code)
		if s != "" {
			syms = append(syms, s)
		}
	}
	total := len(syms)
	if total == 0 {
		return nil
	}
	if report != nil {
		report.HistoryTotal = total
	}

	logf(fmt.Sprintf("stage=history msg=start total=%d period=%s adjust=%s concurrency=%d retries=%d validate_daily=%v min_bars=%d",
		total, periodNorm, opts.HistoryAdjust, opts.HistoryConcurrency, maxRetries, validate && periodNorm == "daily", minBars))

	concurrency := opts.HistoryConcurrency
	if concurrency <= 0 {
		concurrency = c.config.Concurrency
	}
	if concurrency <= 0 {
		concurrency = 5
	}

	var (
		wg            sync.WaitGroup
		failed        int64
		histOK        int64
		completed     int64
		valWarnSyms   int64
		failMu        sync.Mutex
		failedSymbols []string
	)

	sem := make(chan struct{}, concurrency)
	for _, sym := range syms {
		sym := sym
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			defer func() {
				n := atomic.AddInt64(&completed, 1)
				if progressEvery > 0 && n%int64(progressEvery) == 0 {
					logf(fmt.Sprintf("stage=history progress=%d/%d ok=%d fail=%d val_warn_syms=%d",
						n, total, atomic.LoadInt64(&histOK), atomic.LoadInt64(&failed), atomic.LoadInt64(&valWarnSyms)))
				}
			}()

			var rows []KlineData
			var lastErr error
		retryLoop:
			for attempt := 1; attempt <= maxRetries; attempt++ {
				select {
				case <-ctx.Done():
					lastErr = ctx.Err()
					break retryLoop
				default:
				}
				var err error
				rows, err = c.GetHistoryKline(ctx, sym, opts.HistoryPeriod, opts.HistoryAdjust, opts.HistoryStartDate, opts.HistoryEndDate)
				lastErr = err
				if err == nil && len(rows) > 0 {
					break retryLoop
				}
				if attempt < maxRetries {
					logf(fmt.Sprintf("stage=history sym=%s msg=retry attempt=%d/%d err=%v bars=%d", sym, attempt, maxRetries, err, len(rows)))
					select {
					case <-ctx.Done():
						lastErr = ctx.Err()
						break retryLoop
					case <-time.After(retryWait * time.Duration(attempt)):
					}
				}
			}

			if lastErr != nil || len(rows) == 0 {
				atomic.AddInt64(&failed, 1)
				logf(fmt.Sprintf("stage=history sym=%s msg=FAIL err=%v bars=%d", sym, lastErr, len(rows)))
				failMu.Lock()
				if len(failedSymbols) < 200 {
					if lastErr != nil {
						failedSymbols = append(failedSymbols, fmt.Sprintf("%s: %v", sym, lastErr))
					} else {
						failedSymbols = append(failedSymbols, sym+": empty_bars")
					}
				}
				failMu.Unlock()
				return
			}

			atomic.AddInt64(&histOK, 1)
			if validate && periodNorm == "daily" {
				warns := ValidateDailyKlineHistory(rows, opts.HistoryStartDate, opts.HistoryEndDate, minBars)
				if len(warns) > 0 {
					atomic.AddInt64(&valWarnSyms, 1)
					minD, maxD := klineDateSpan(rows)
					logf(fmt.Sprintf("stage=history sym=%s msg=VALIDATION_WARN warns=%v bars=%d range=[%s..%s]",
						sym, warns, len(rows), minD, maxD))
				}
			}
		}()
	}
	wg.Wait()

	if report != nil {
		report.HistorySuccess = int(atomic.LoadInt64(&histOK))
		report.HistoryFailed = int(atomic.LoadInt64(&failed))
		report.ValidationWarnSymbols = int(atomic.LoadInt64(&valWarnSyms))
		failMu.Lock()
		report.FailedSymbols = append([]string(nil), failedSymbols...)
		failMu.Unlock()
	}

	logf(fmt.Sprintf("stage=history msg=summary total=%d ok=%d fail=%d validation_warn_symbols=%d",
		total, atomic.LoadInt64(&histOK), atomic.LoadInt64(&failed), atomic.LoadInt64(&valWarnSyms)))

	if atomic.LoadInt64(&failed) > 0 {
		return fmt.Errorf("prewarm history finished with %d failures (see logs / PrewarmReport.FailedSymbols)", atomic.LoadInt64(&failed))
	}
	return nil
}

func (c *Client) fetchTodayTimelineWithRetry(ctx context.Context, symbol string) (*stocksdk.TodayTimelineResponse, error) {
	var lastErr error
	for i := 0; i < 3; i++ {
		raw, err := c.sdk.GetTodayTimeline(ctx, symbol)
		if err == nil && raw != nil && len(raw.Data) > 0 {
			return raw, nil
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(i+1) * 200 * time.Millisecond):
		}
	}

	// 降级策略：若分时接口不稳定，使用 1 分钟K线拼装分时响应，避免接口直接500。
	minute, err := c.sdk.GetMinuteKline(ctx, symbol, &stocksdk.MinuteKlineOptions{
		Period: stocksdk.MinutePeriod1,
		Adjust: stocksdk.AdjustTypeNone,
	})
	if err != nil || len(minute) == 0 {
		if lastErr != nil {
			return nil, fmt.Errorf("failed to fetch timeline: %w", lastErr)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch timeline via minute fallback: %w", err)
		}
		return nil, errors.New("failed to fetch timeline: empty minute fallback data")
	}

	prevClose := minute[0].Open
	totalAmount := 0.0
	totalVolume := 0.0
	data := make([]stocksdk.TimelineItem, 0, len(minute))
	for _, item := range minute {
		totalAmount += item.Close * item.Volume
		totalVolume += item.Volume
		avg := item.Close
		if totalVolume > 0 {
			avg = totalAmount / totalVolume
		}
		data = append(data, stocksdk.TimelineItem{
			Time:     item.Time,
			Price:    item.Close,
			AvgPrice: avg,
			Volume:   item.Volume,
		})
	}

	return &stocksdk.TodayTimelineResponse{
		Code:      symbol,
		PrevClose: prevClose,
		Data:      data,
	}, nil
}

func (c *Client) cacheDir() string {
	return filepath.Join(c.config.DataDir, "stockapi-cache")
}

func defaultDataDir() string {
	if dir, err := os.UserCacheDir(); err == nil && strings.TrimSpace(dir) != "" {
		return filepath.Join(dir, "stockdb")
	}
	home, err := os.UserHomeDir()
	if err == nil && strings.TrimSpace(home) != "" {
		return filepath.Join(home, ".stockdb")
	}
	return "./data"
}

func (c *Client) cachePath(key string) string {
	sum := sha1.Sum([]byte(key))
	return filepath.Join(c.cacheDir(), hex.EncodeToString(sum[:])+".json")
}

func (c *Client) getCache(key string, ttl time.Duration, out interface{}) bool {
	if c.config.CacheMode == CacheModeDisabled {
		return false
	}
	now := time.Now()
	c.mu.RLock()
	if entry, ok := c.mem[key]; ok && now.Sub(entry.Timestamp) <= ttl {
		_ = json.Unmarshal(entry.Data, out)
		c.mu.RUnlock()
		return true
	}
	c.mu.RUnlock()

	content, err := os.ReadFile(c.cachePath(key))
	if err != nil {
		return false
	}
	var entry cacheEntry
	if json.Unmarshal(content, &entry) != nil {
		return false
	}
	if now.Sub(entry.Timestamp) > ttl {
		return false
	}
	if json.Unmarshal(entry.Data, out) != nil {
		return false
	}
	c.mu.Lock()
	c.mem[key] = entry
	c.mu.Unlock()
	return true
}

func (c *Client) setCache(key string, data interface{}) {
	if c.config.CacheMode == CacheModeDisabled || c.config.CacheMode == CacheModeReadOnly {
		return
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}
	entry := cacheEntry{
		Timestamp: time.Now(),
		Data:      raw,
	}
	c.mu.Lock()
	c.mem[key] = entry
	c.mu.Unlock()

	payload, err := json.Marshal(entry)
	if err != nil {
		return
	}
	_ = os.WriteFile(c.cachePath(key), payload, 0o644)
}

func normalizePeriod(period string) string {
	switch strings.ToLower(strings.TrimSpace(period)) {
	case "weekly", "week", "w":
		return "weekly"
	case "monthly", "month", "m":
		return "monthly"
	default:
		return "daily"
	}
}

func deref(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func toSymbol(code string) string {
	c := strings.TrimSpace(code)
	if c == "" {
		return ""
	}
	if strings.HasPrefix(c, "sh") || strings.HasPrefix(c, "sz") || strings.HasPrefix(c, "bj") {
		return c
	}
	if strings.HasPrefix(c, "6") || strings.HasPrefix(c, "9") {
		return "sh" + c
	}
	if strings.HasPrefix(c, "4") || strings.HasPrefix(c, "8") {
		return "bj" + c
	}
	return "sz" + c
}
