package multisource

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SinaFullSource 新浪财经完整实现
type SinaFullSource struct {
	priority int
	reqMgr   *RequestManager
}

// NewSinaSource 创建新浪财经数据源
func NewSinaSource(priority int, reqMgr *RequestManager) *SinaFullSource {
	return &SinaFullSource{
		priority: priority,
		reqMgr:   reqMgr,
	}
}

func (s *SinaFullSource) Name() string {
	return "sina"
}

func (s *SinaFullSource) Type() DataSourceType {
	return DataSourceSina
}

func (s *SinaFullSource) Priority() int {
	return s.priority
}

func (s *SinaFullSource) HealthCheck(ctx context.Context) error {
	_, err := s.GetStockQuotes(ctx, []string{"000001"})
	return err
}

// SinaQuoteURL 新浪行情接口
const SinaQuoteURL = "http://hq.sinajs.cn/list="

// GetStockQuotes 获取股票行情
func (s *SinaFullSource) GetStockQuotes(ctx context.Context, codes []string) ([]StockQuote, error) {
	if len(codes) == 0 {
		return []StockQuote{}, nil
	}

	// 缓存
	cacheKey := CacheKey("sina_quotes", codes...)
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if quotes, ok := cached.([]StockQuote); ok {
			return quotes, nil
		}
	}

	// 转换为新浪格式：sh600000,sz000001
	var sinaCodes []string
	for _, code := range codes {
		sinaCodes = append(sinaCodes, AddMarketPrefix(code))
	}

	url := SinaQuoteURL + strings.Join(sinaCodes, ",")

	data, err := s.reqMgr.GetWithRateLimit("sina.com.cn", url)
	if err != nil {
		return nil, err
	}

	result, err := s.parseSinaQuotes(string(data), codes)
	if err != nil {
		return nil, err
	}

	if len(result) > 0 {
		s.reqMgr.SetCache(cacheKey, result, CacheTimeQuote)
	}

	return result, nil
}

// parseSinaQuotes 解析新浪行情响应
func (s *SinaFullSource) parseSinaQuotes(text string, originalCodes []string) ([]StockQuote, error) {
	var result []StockQuote

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 格式: var hq_str_sh600000="..."
		eqIdx := strings.Index(line, "=")
		if eqIdx == -1 {
			continue
		}

		// 提取数据部分
		dataPart := line[eqIdx+2 : len(line)-1] // 去掉引号
		if dataPart == "" {
			continue
		}

		parts := strings.Split(dataPart, ",")
		if len(parts) < 32 {
			continue
		}

		quote := StockQuote{
			Code:      originalCodes[i%len(originalCodes)],
			Name:      parts[0],
			Open:      parseFloat(parts[1]),
			PrevClose: parseFloat(parts[2]),
			Price:     parseFloat(parts[3]),
			High:      parseFloat(parts[4]),
			Low:       parseFloat(parts[5]),
			Volume:    parseFloat(parts[8]),
			Amount:    parseFloat(parts[9]),
			Time:      parts[30] + " " + parts[31],
		}

		if quote.PrevClose > 0 {
			quote.Change = quote.Price - quote.PrevClose
			quote.ChangePct = (quote.Change / quote.PrevClose) * 100
		}

		result = append(result, quote)
	}

	return result, nil
}

// GetKLine 获取K线（新浪 JSON 接口；adjust 无效，新浪不提供复权参数）
func (s *SinaFullSource) GetKLine(ctx context.Context, code string, period string, adjust string, startDate string, endDate string) ([]KLineItem, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	sym := normalizeCodeForKlineAPI(code)
	if sym == "" {
		return nil, fmt.Errorf("invalid stock code: %s", code)
	}
	p := strings.ToLower(strings.TrimSpace(period))
	if p == "" {
		p = "daily"
	}
	scale := "240"
	switch p {
	case "week", "weekly", "w":
		scale = "1680"
	case "month", "monthly", "m":
		scale = "7200"
	}
	count := klineBarsToFetch(startDate, endDate)
	cacheKey := CacheKey("sina_kline", sym, p, scale, startDate, endDate, strconv.Itoa(count))
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if klines, ok := cached.([]KLineItem); ok {
			return klines, nil
		}
	}

	rawURL := fmt.Sprintf(
		"http://quotes.sina.cn/cn/json_v2.php/CN_MarketDataService.getKLineData?symbol=%s&scale=%s&ma=no&datalen=%d",
		sym, scale, count,
	)
	data, err := s.reqMgr.GetWithRateLimit("sina.com.cn", rawURL)
	if err != nil {
		return nil, err
	}

	var payload []struct {
		Day    string `json:"day"`
		Open   string `json:"open"`
		High   string `json:"high"`
		Low    string `json:"low"`
		Close  string `json:"close"`
		Volume string `json:"volume"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	var out []KLineItem
	for _, item := range payload {
		vol := float64(parseInt64Trim(item.Volume))
		out = append(out, KLineItem{
			Date:   item.Day,
			Open:   parseFloat(item.Open),
			High:   parseFloat(item.High),
			Low:    parseFloat(item.Low),
			Close:  parseFloat(item.Close),
			Volume: vol,
		})
	}
	out = filterKLineByRange(out, startDate, endDate)
	if len(out) > 0 {
		s.reqMgr.SetCache(cacheKey, out, CacheTimeKLine)
	}
	return out, nil
}

// ========== 腾讯财经数据源 ==========

// TencentFullSource 腾讯财经完整实现
type TencentFullSource struct {
	priority int
	reqMgr   *RequestManager
}

// NewTencentSource 创建腾讯财经数据源
func NewTencentSource(priority int, reqMgr *RequestManager) *TencentFullSource {
	return &TencentFullSource{
		priority: priority,
		reqMgr:   reqMgr,
	}
}

func (s *TencentFullSource) Name() string {
	return "tencent"
}

func (s *TencentFullSource) Type() DataSourceType {
	return DataSourceTencent
}

func (s *TencentFullSource) Priority() int {
	return s.priority
}

func (s *TencentFullSource) HealthCheck(ctx context.Context) error {
	_, err := s.GetStockQuotes(ctx, []string{"000001"})
	return err
}

// TencentQuoteURL 腾讯行情接口
const TencentQuoteURL = "http://qt.gtimg.cn/q="

// GetStockQuotes 获取股票行情
func (s *TencentFullSource) GetStockQuotes(ctx context.Context, codes []string) ([]StockQuote, error) {
	if len(codes) == 0 {
		return []StockQuote{}, nil
	}

	cacheKey := CacheKey("tencent_quotes", codes...)
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if quotes, ok := cached.([]StockQuote); ok {
			return quotes, nil
		}
	}

	// 转换为腾讯格式
	var qqCodes []string
	for _, code := range codes {
		prefix := "sz"
		if strings.HasPrefix(RemoveMarketPrefix(code), "6") {
			prefix = "sh"
		}
		qqCodes = append(qqCodes, prefix+RemoveMarketPrefix(code))
	}

	url := TencentQuoteURL + strings.Join(qqCodes, ",")

	data, err := s.reqMgr.GetWithRateLimit("qq.com", url)
	if err != nil {
		return nil, err
	}

	result, err := s.parseTencentQuotes(string(data), codes)
	if err != nil {
		return nil, err
	}

	if len(result) > 0 {
		s.reqMgr.SetCache(cacheKey, result, CacheTimeQuote)
	}

	return result, nil
}

// parseTencentQuotes 解析腾讯行情响应
func (s *TencentFullSource) parseTencentQuotes(text string, originalCodes []string) ([]StockQuote, error) {
	var result []StockQuote

	lines := strings.Split(text, ";")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		eqIdx := strings.Index(line, "=")
		if eqIdx == -1 {
			continue
		}

		dataPart := line[eqIdx+2 : len(line)-1]
		if dataPart == "" {
			continue
		}

		parts := strings.Split(dataPart, "~")
		if len(parts) < 40 {
			continue
		}

		quote := StockQuote{
			Code:      originalCodes[i%len(originalCodes)],
			Name:      parts[1],
			Price:     parseFloat(parts[3]),
			PrevClose: parseFloat(parts[4]),
			Open:      parseFloat(parts[5]),
			High:      parseFloat(parts[33]),
			Low:       parseFloat(parts[34]),
			Volume:    parseFloat(parts[6]),
			Amount:    parseFloat(parts[37]),
			Time:      parts[30],
		}

		if quote.PrevClose > 0 {
			quote.Change = quote.Price - quote.PrevClose
			quote.ChangePct = (quote.Change / quote.PrevClose) * 100
		}

		result = append(result, quote)
	}

	return result, nil
}

// GetKLine 获取K线（腾讯 day/week/month；adjust 无效）
func (s *TencentFullSource) GetKLine(ctx context.Context, code string, period string, adjust string, startDate string, endDate string) ([]KLineItem, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	sym := normalizeCodeForKlineAPI(code)
	if sym == "" {
		return nil, fmt.Errorf("invalid stock code: %s", code)
	}
	klinePeriod := "day"
	switch strings.ToLower(strings.TrimSpace(period)) {
	case "week", "weekly", "w":
		klinePeriod = "week"
	case "month", "monthly", "m":
		klinePeriod = "month"
	}
	count := klineBarsToFetch(startDate, endDate)
	cacheKey := CacheKey("tencent_kline", sym, klinePeriod, startDate, endDate, strconv.Itoa(count))
	if cached, ok := s.reqMgr.GetCache(cacheKey); ok {
		if klines, ok := cached.([]KLineItem); ok {
			return klines, nil
		}
	}

	varName := fmt.Sprintf("kline_%s", klinePeriod)
	param := fmt.Sprintf("%s,%s,,0,%d", sym, klinePeriod, count)
	rawURL := fmt.Sprintf("https://web.ifzq.gtimg.cn/appstock/app/kline/kline?_var=%s&param=%s", varName, param)

	data, err := s.reqMgr.GetWithRateLimit("qq.com", rawURL)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(string(data))
	if idx := strings.Index(content, "="); idx != -1 {
		content = strings.TrimSpace(content[idx+1:])
	}

	var result struct {
		Code int `json:"code"`
		Data map[string]struct {
			Day   [][]string `json:"day"`
			Week  [][]string `json:"week"`
			Month [][]string `json:"month"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, err
	}
	if result.Code != 0 || len(result.Data) == 0 {
		return nil, fmt.Errorf("tencent kline: unexpected response")
	}
	entry, ok := result.Data[sym]
	if !ok {
		return nil, fmt.Errorf("tencent kline: missing series for %s", sym)
	}
	var rows [][]string
	switch klinePeriod {
	case "week":
		rows = entry.Week
	case "month":
		rows = entry.Month
	default:
		rows = entry.Day
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("tencent kline: empty rows")
	}

	var out []KLineItem
	for _, item := range rows {
		if len(item) < 6 {
			continue
		}
		vol := parseFloat(item[5]) * 100 // 手 -> 股
		out = append(out, KLineItem{
			Date:   item[0],
			Open:   parseFloat(item[1]),
			Close:  parseFloat(item[2]),
			High:   parseFloat(item[3]),
			Low:    parseFloat(item[4]),
			Volume: vol,
		})
	}
	out = filterKLineByRange(out, startDate, endDate)
	if len(out) > 0 {
		s.reqMgr.SetCache(cacheKey, out, CacheTimeKLine)
	}
	return out, nil
}

// parseFloat 解析浮点数
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseInt64Trim(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// normalizeCodeForKlineAPI 转为新浪/腾讯 K 线所需的小写 sh/sz 前缀
func normalizeCodeForKlineAPI(code string) string {
	code = strings.TrimSpace(strings.ToLower(code))
	if code == "" {
		return ""
	}
	if strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "sz") || strings.HasPrefix(code, "bj") {
		return code
	}
	if len(code) == 6 {
		if strings.HasPrefix(code, "6") || strings.HasPrefix(code, "9") || strings.HasPrefix(code, "5") {
			return "sh" + code
		}
		if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "2") || strings.HasPrefix(code, "3") {
			return "sz" + code
		}
		if strings.HasPrefix(code, "4") || strings.HasPrefix(code, "8") {
			return "bj" + code
		}
	}
	return code
}

func normKlineDate(d string) string {
	d = strings.TrimSpace(d)
	d = strings.ReplaceAll(d, "-", "")
	d = strings.ReplaceAll(d, "/", "")
	if len(d) >= 8 {
		return d[:8]
	}
	return d
}

func klineBarsToFetch(startDate, endDate string) int {
	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)
	if startDate == "" && endDate == "" {
		return 240
	}
	const layout = "20060102"
	now := time.Now()
	t1 := now
	if endDate != "" {
		if t, err := time.ParseInLocation(layout, endDate, time.Local); err == nil {
			t1 = t
		}
	}
	t0 := t1.AddDate(-2, 0, 0)
	if startDate != "" {
		if t, err := time.ParseInLocation(layout, startDate, time.Local); err == nil {
			t0 = t
		}
	}
	days := int(t1.Sub(t0).Hours() / 24)
	if days < 30 {
		days = 30
	}
	if days > 800 {
		days = 800
	}
	n := days * 5 / 7
	if n < 60 {
		n = 60
	}
	if n > 800 {
		n = 800
	}
	return n
}

func filterKLineByRange(items []KLineItem, startDate, endDate string) []KLineItem {
	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)
	if startDate == "" && endDate == "" {
		return items
	}
	startN := normKlineDate(startDate)
	endN := normKlineDate(endDate)
	var out []KLineItem
	for _, it := range items {
		d := normKlineDate(it.Date)
		if startN != "" && d != "" && d < startN {
			continue
		}
		if endN != "" && d != "" && d > endN {
			continue
		}
		out = append(out, it)
	}
	if len(out) == 0 {
		return items
	}
	return out
}

// ========== 增强的多数据源管理器 ==========

// RegisterEnhancedSources 重置为东财/新浪/腾讯完整数据源（与默认注册一致）。
// 须在已持有写锁时使用 addSourceUnsafe，避免在 RegisterEnhancedSources 内调用 AddSource 导致死锁。
func (m *MultiSourceManager) RegisterEnhancedSources() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sources = make([]DataSource, 0)
	m.sourceMap = make(map[DataSourceType]DataSource)

	m.addSourceUnsafe(NewEastMoneySource(1, m.reqMgr))
	m.addSourceUnsafe(NewSinaSource(2, m.reqMgr))
	m.addSourceUnsafe(NewTencentSource(3, m.reqMgr))
	m.addSourceUnsafe(NewXueqiuSource(4, m.reqMgr))
	m.addSourceUnsafe(NewBaiduSource(5, m.reqMgr))
	m.addSourceUnsafe(NewTonghuashunSource(6, m.reqMgr))
}

// ========== 高级功能 ==========

// StockQuoteBatch 批量行情响应
type StockQuoteBatch struct {
	Success map[string]*StockQuote `json:"success"`
	Failed  map[string]string      `json:"failed,omitempty"`
}

// GetStockQuotesBatch 批量获取股票行情
func (m *MultiSourceManager) GetStockQuotesBatch(ctx context.Context, codes []string) (*StockQuoteBatch, error) {
	result := &StockQuoteBatch{
		Success: make(map[string]*StockQuote),
		Failed:  make(map[string]string),
	}

	// 分批获取，每批最多 80 个（新浪限制）
	batchSize := 80
	for i := 0; i < len(codes); i += batchSize {
		end := i + batchSize
		if end > len(codes) {
			end = len(codes)
		}
		batch := codes[i:end]

		quotes, err := m.GetStockQuotes(ctx, batch)
		if err != nil {
			// 整批失败
			for _, code := range batch {
				result.Failed[code] = err.Error()
			}
			continue
		}

		// 按代码索引
		quoteMap := make(map[string]*StockQuote)
		for j := range quotes {
			quoteMap[quotes[j].Code] = &quotes[j]
		}

		for _, code := range batch {
			if q, ok := quoteMap[code]; ok {
				result.Success[code] = q
			} else {
				result.Failed[code] = "no data"
			}
		}
	}

	if len(result.Failed) == 0 {
		result.Failed = nil
	}

	return result, nil
}

// ========== JSON 序列化辅助 ==========

// ToJSON 转换为JSON
func (sq *StockQuote) ToJSON() string {
	data, _ := json.Marshal(sq)
	return string(data)
}

// ToJSON 转换为JSON
func (kl *KLineItem) ToJSON() string {
	data, _ := json.Marshal(kl)
	return string(data)
}
