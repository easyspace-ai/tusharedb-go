package multisource

import (
	"compress/gzip"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// RequestManager 请求管理器
type RequestManager struct {
	client       *http.Client
	cache        *Cache
	sourceStatus map[string]*SourceStatus
	rateLimiter  *RateLimiter
	proxyPool    *ProxyPoolState
	mu           sync.RWMutex
}

// SourceStatus 数据源状态
type SourceStatus struct {
	FailCount    int
	LastFailTime time.Time
	LastSuccess  time.Time
	LastChecked  time.Time
	LastLatency  time.Duration
	Disabled     bool
}

// ProxyPoolState 代理池状态
type ProxyPoolState struct {
	proxies   []string
	expiresAt time.Time
	current   int
	mu        sync.Mutex
	lastFetch time.Time
	lastError string
}

func (ps *ProxyPoolState) reset() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.proxies = nil
	ps.expiresAt = time.Time{}
	ps.current = 0
	ps.lastFetch = time.Time{}
	ps.lastError = ""
}

// Cache 缓存管理
type Cache struct {
	data map[string]*CacheItem
	mu   sync.RWMutex
}

// CacheItem 缓存项
type CacheItem struct {
	Data      interface{}
	ExpireAt  time.Time
	CacheTime time.Duration
}

// Config 请求管理器配置
type Config struct {
	ProxyURL          string
	ProxyPoolEnabled  bool
	ProxyPoolList     string
	ProxyProvider     string
	ProxyApiUrl       string
	ProxyApiKey       string
	ProxyApiSecret    string
	ProxyPoolSize     int
	ProxyPoolTTL      int
	ProxyPoolProtocol string
	ProxyRegion       string
}

// User-Agent池
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
}

var (
	globalRequestManager *RequestManager
	once                 sync.Once
)

// GetRequestManager 获取全局请求管理器
func GetRequestManager() *RequestManager {
	once.Do(func() {
		globalRequestManager = NewRequestManager()
	})
	return globalRequestManager
}

// NewRequestManager 创建请求管理器
func NewRequestManager() *RequestManager {
	rm := &RequestManager{
		cache: &Cache{
			data: make(map[string]*CacheItem),
		},
		sourceStatus: make(map[string]*SourceStatus),
		rateLimiter:  GetRateLimiter(),
		proxyPool:    &ProxyPoolState{},
	}
	rm.initClient(nil)

	// 启动缓存清理调度器
	go rm.startCacheCleanupScheduler()

	return rm
}

// UpdateConfig 更新配置
func (rm *RequestManager) UpdateConfig(cfg *Config) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if rm.proxyPool != nil {
		rm.proxyPool.reset()
	}
	rm.initClient(cfg)
}

// initClient 初始化HTTP客户端
func (rm *RequestManager) initClient(cfg *Config) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{},
		MaxIdleConns:    100,
		IdleConnTimeout: 90 * time.Second,
	}

	rm.client = &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}

// GetRandomUA 获取随机User-Agent
func (rm *RequestManager) GetRandomUA() string {
	return userAgents[rand.Intn(len(userAgents))]
}

// SetRequestHeaders 设置完整的请求头
func (rm *RequestManager) SetRequestHeaders(req *http.Request, referer string) {
	req.Header.Set("User-Agent", rm.GetRandomUA())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "no-cache")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
}

// GetClient 获取HTTP客户端
func (rm *RequestManager) GetClient() *http.Client {
	return rm.client
}

// GetCache 获取缓存数据
func (rm *RequestManager) GetCache(key string) (interface{}, bool) {
	rm.cache.mu.RLock()
	defer rm.cache.mu.RUnlock()

	item, exists := rm.cache.data[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.ExpireAt) {
		return nil, false
	}

	return item.Data, true
}

// SetCache 设置缓存数据
func (rm *RequestManager) SetCache(key string, data interface{}, duration time.Duration) {
	rm.cache.mu.Lock()
	defer rm.cache.mu.Unlock()

	rm.cache.data[key] = &CacheItem{
		Data:      data,
		ExpireAt:  time.Now().Add(duration),
		CacheTime: duration,
	}
}

// ClearCache 清除缓存
func (rm *RequestManager) ClearCache(key string) {
	rm.cache.mu.Lock()
	defer rm.cache.mu.Unlock()
	delete(rm.cache.data, key)
}

// ClearAllCache 清除所有缓存
func (rm *RequestManager) ClearAllCache() {
	rm.cache.mu.Lock()
	defer rm.cache.mu.Unlock()
	rm.cache.data = make(map[string]*CacheItem)
}

// MarkSourceFailed 标记数据源失败
func (rm *RequestManager) MarkSourceFailed(sourceName string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	status, exists := rm.sourceStatus[sourceName]
	if !exists {
		status = &SourceStatus{}
		rm.sourceStatus[sourceName] = status
	}

	status.FailCount++
	status.LastFailTime = time.Now()

	// 连续失败3次，禁用5分钟
	if status.FailCount >= 3 {
		status.Disabled = true
	}
}

// MarkSourceSuccess 标记数据源成功
func (rm *RequestManager) MarkSourceSuccess(sourceName string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if status, exists := rm.sourceStatus[sourceName]; exists {
		status.FailCount = 0
		status.Disabled = false
		status.LastSuccess = time.Now()
	}
}

// IsSourceAvailable 检查数据源是否可用
func (rm *RequestManager) IsSourceAvailable(sourceName string) bool {
	rm.mu.RLock()
	status, exists := rm.sourceStatus[sourceName]
	if !exists {
		rm.mu.RUnlock()
		return true
	}
	disabled := status.Disabled
	lastFail := status.LastFailTime
	rm.mu.RUnlock()

	if !disabled {
		return true
	}

	// 5分钟后自动重置
	if time.Since(lastFail) > 5*time.Minute {
		rm.mu.Lock()
		if status, ok := rm.sourceStatus[sourceName]; ok {
			status.Disabled = false
			status.FailCount = 0
		}
		rm.mu.Unlock()
		return true
	}

	return false
}

// DoRequestWithRateLimit 带限流的HTTP请求
func (rm *RequestManager) DoRequestWithRateLimit(domain string, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	// 使用限流
	err = rm.rateLimiter.ExecuteWithRateLimit(domain, func() error {
		resp, err = rm.client.Do(req)
		return err
	})

	return resp, err
}

// GetWithRateLimit 带限流的GET请求
func (rm *RequestManager) GetWithRateLimit(domain string, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	rm.SetRequestHeaders(req, "")

	resp, err := rm.DoRequestWithRateLimit(domain, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体，处理gzip压缩
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer gzReader.Close()
		reader = gzReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// 检测并转换编码
	charset := detectCharset(resp.Header.Get("Content-Type"), body)
	if charset == "gbk" || charset == "gb2312" || charset == "gb18030" {
		// 将GBK编码转换为UTF-8
		decoder := simplifiedchinese.GBK.NewDecoder()
		utf8Body, _, err := transform.Bytes(decoder, body)
		if err == nil {
			body = utf8Body
		}
	}

	return body, nil
}

// GetWithRateLimitEx 带限流的 GET，可设置 Referer 与额外请求头（如雪球 Cookie、百度 Referer）。
func (rm *RequestManager) GetWithRateLimitEx(domain string, rawURL string, referer string, extra map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	rm.SetRequestHeaders(req, referer)
	for k, v := range extra {
		if strings.TrimSpace(k) != "" && v != "" {
			req.Header.Set(k, v)
		}
	}

	resp, err := rm.DoRequestWithRateLimit(domain, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer gzReader.Close()
		reader = gzReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	charset := detectCharset(resp.Header.Get("Content-Type"), body)
	if charset == "gbk" || charset == "gb2312" || charset == "gb18030" {
		decoder := simplifiedchinese.GBK.NewDecoder()
		utf8Body, _, err := transform.Bytes(decoder, body)
		if err == nil {
			body = utf8Body
		}
	}

	return body, nil
}

// Get 简单的GET请求（不带限流）
func (rm *RequestManager) Get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	rm.SetRequestHeaders(req, "")

	resp, err := rm.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// detectCharset 检测网页编码
func detectCharset(contentType string, body []byte) string {
	// 1. 从Content-Type头检测
	contentType = strings.ToLower(contentType)
	if strings.Contains(contentType, "charset=") {
		parts := strings.Split(contentType, "charset=")
		if len(parts) > 1 {
			charset := strings.TrimSpace(parts[1])
			charset = strings.Split(charset, ";")[0]
			charset = strings.Trim(charset, "\"'")
			return strings.ToLower(charset)
		}
	}

	// 2. 从HTML meta标签检测
	htmlStr := string(body)
	metaCharsetRe := regexp.MustCompile(`(?i)<meta[^>]+charset=["']?([^"'\s>]+)`)
	if matches := metaCharsetRe.FindStringSubmatch(htmlStr); len(matches) > 1 {
		return strings.ToLower(matches[1])
	}

	// 默认返回utf-8
	return "utf-8"
}

// startCacheCleanupScheduler 启动缓存清理调度器
func (rm *RequestManager) startCacheCleanupScheduler() {
	// 每分钟检查一次过期缓存
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rm.cleanupExpiredCache()
	}
}

// cleanupExpiredCache 清理过期的缓存
func (rm *RequestManager) cleanupExpiredCache() {
	rm.cache.mu.Lock()
	defer rm.cache.mu.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, item := range rm.cache.data {
		if now.After(item.ExpireAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(rm.cache.data, key)
	}
}

// CacheKey 生成缓存键
func CacheKey(prefix string, parts ...string) string {
	full := prefix + ":" + strings.Join(parts, ":")
	sum := sha1.Sum([]byte(full))
	return prefix + "_" + hex.EncodeToString(sum[:8])
}

// IsTradingTime 检查是否为交易时间
func IsTradingTime() bool {
	now := time.Now()
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now = now.In(loc)

	// 检查是否为工作日（周一到周五）
	weekday := now.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	// 检查时间段
	hour := now.Hour()
	minute := now.Minute()
	currentMinutes := hour*60 + minute

	// 上午交易时间：9:30 - 11:30
	morningStart := 9*60 + 30
	morningEnd := 11*60 + 30

	// 下午交易时间：13:00 - 15:00
	afternoonStart := 13 * 60
	afternoonEnd := 15 * 60

	return (currentMinutes >= morningStart && currentMinutes <= morningEnd) ||
		(currentMinutes >= afternoonStart && currentMinutes <= afternoonEnd)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
