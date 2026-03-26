package stocksdk

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// ============ HTTP 客户端配置 ============

// Config StockSDK 客户端配置
type Config struct {
	APIKey      string        `json:"api_key"`
	Timeout     time.Duration `json:"timeout"`
	Retries     int           `json:"retries"`
	RetryWait   time.Duration `json:"retry_wait"`
	UserAgent   string        `json:"user_agent"`
	BaseURL     string        `json:"base_url"`
}

// HTTPClient HTTP 请求客户端（带重试机制）
type HTTPClient struct {
	cfg        Config
	client     *http.Client
	retryCfg   RetryConfig
	rateLimit  *RateLimiter
	circuitBreaker *CircuitBreaker
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries      int
	BaseDelay       time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableStatus []int
}

// DefaultRetryConfig 默认重试配置
var DefaultRetryConfig = RetryConfig{
	MaxRetries:      DefaultMaxRetries,
	BaseDelay:       time.Duration(DefaultBaseDelay) * time.Millisecond,
	MaxDelay:        time.Duration(DefaultMaxDelay) * time.Millisecond,
	BackoffFactor:   DefaultBackoffMultiplier,
	RetryableStatus: DefaultRetryableStatusCodes,
}

// NewHTTPClient 创建 HTTP 客户端
func NewHTTPClient(cfg Config) *HTTPClient {
	if cfg.Timeout <= 0 {
		cfg.Timeout = time.Duration(DefaultTimeout) * time.Millisecond
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = TencentBaseURL
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "tusharedb-go/1.0 (stocksdk)"
	}

	// 创建自定义 Transport，优化连接配置以避免 EOF 错误
	transport := &http.Transport{
		// 禁用 Keep-Alive 避免连接复用问题
		DisableKeepAlives: true,
		// 强制使用 HTTP/1.1（避免 HTTP/2 的某些兼容性问题）
		ForceAttemptHTTP2: false,
		// 配置 TLS（使用安全的默认值）
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		// 连接超时配置
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: -1, // 禁用 TCP Keep-Alive
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// 限制空闲连接
		MaxIdleConns:        0,
		MaxIdleConnsPerHost: 0,
		IdleConnTimeout:     0,
	}

	return &HTTPClient{
		cfg: cfg,
		client: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		retryCfg: DefaultRetryConfig,
	}
}

// WithRetryConfig 设置重试配置
func (c *HTTPClient) WithRetryConfig(cfg RetryConfig) *HTTPClient {
	c.retryCfg = cfg
	return c
}

// WithRateLimiter 设置限流器
func (c *HTTPClient) WithRateLimiter(rl *RateLimiter) *HTTPClient {
	c.rateLimit = rl
	return c
}

// WithCircuitBreaker 设置熔断器
func (c *HTTPClient) WithCircuitBreaker(cb *CircuitBreaker) *HTTPClient {
	c.circuitBreaker = cb
	return c
}

// Get 发送 GET 请求（使用 backoff 重试机制）
func (c *HTTPClient) Get(ctx context.Context, urlStr string) ([]byte, error) {
	// 检查熔断器
	if c.circuitBreaker != nil && !c.circuitBreaker.Allow() {
		return nil, fmt.Errorf("circuit breaker is open")
	}

	// 限流等待
	if c.rateLimit != nil {
		c.rateLimit.Wait()
	}

	// 使用 backoff 配置
	backoffCfg := HTTPClientBackoffConfig()

	var result []byte
	err := ExecuteWithBackoff(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
		if err != nil {
			return backoff.Permanent(err)
		}

		req.Header.Set("User-Agent", c.cfg.UserAgent)
		req.Header.Set("Referer", "https://quote.eastmoney.com/")
		req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		req.Header.Set("Connection", "close")

		resp, err := c.client.Do(req)
		if err != nil {
			if c.circuitBreaker != nil {
				c.circuitBreaker.Failure()
			}
			return NewRetryableError(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if c.circuitBreaker != nil {
				c.circuitBreaker.Failure()
			}
			if resp.StatusCode >= 500 || resp.StatusCode == 429 {
				return NewRetryableError(fmt.Errorf("HTTP %d", resp.StatusCode))
			}
			return backoff.Permanent(fmt.Errorf("HTTP %d", resp.StatusCode))
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return NewRetryableError(err)
		}

		result = data
		if c.circuitBreaker != nil {
			c.circuitBreaker.Success()
		}
		return nil
	}, backoffCfg)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// ============ 腾讯财经专用方法 ============

// GetTencentQuote 获取腾讯财经行情数据
func (c *HTTPClient) GetTencentQuote(ctx context.Context, codes string) ([]TencentResponseItem, error) {
	urlStr := fmt.Sprintf("%s/?q=%s", c.cfg.BaseURL, url.QueryEscape(codes))

	data, err := c.Get(ctx, urlStr)
	if err != nil {
		return nil, err
	}

	// 解码 GBK
	text, err := DecodeGBK(data)
	if err != nil {
		return nil, fmt.Errorf("decode GBK failed: %w", err)
	}

	// 解析响应
	return ParseTencentResponse(text), nil
}

// ============ 限流器 ============

// RateLimiter 简单的令牌桶限流器
type RateLimiter struct {
	tokens    float64
	capacity  float64
	rate      float64
	mu        sync.Mutex
	lastToken time.Time
}

// NewRateLimiter 创建限流器
// rate: 每秒产生的令牌数
// capacity: 桶容量（最大令牌数）
func NewRateLimiter(rate float64, capacity float64) *RateLimiter {
	return &RateLimiter{
		tokens:    capacity,
		capacity:  capacity,
		rate:      rate,
		lastToken: time.Now(),
	}
}

// Wait 等待获取一个令牌
func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 补充令牌
	now := time.Now()
	elapsed := now.Sub(rl.lastToken).Seconds()
	rl.tokens = math.Min(rl.capacity, rl.tokens+elapsed*rl.rate)
	rl.lastToken = now

	// 如果没有令牌，等待
	if rl.tokens < 1 {
		need := (1 - rl.tokens) / rl.rate
		time.Sleep(time.Duration(need * float64(time.Second)))
		rl.tokens = 0
		rl.lastToken = time.Now()
	} else {
		rl.tokens--
	}
}

// ============ 熔断器 ============

// CircuitState 熔断器状态
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	state           CircuitState
	failureThreshold int
	successThreshold int
	timeout         time.Duration
	failures        int
	successes       int
	lastFailure     time.Time
	mu              sync.Mutex
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:           CircuitClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:         timeout,
	}
}

// Allow 检查是否允许请求
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// 检查是否超时，可以进入半开状态
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.state = CircuitHalfOpen
			cb.successes = 0
			return true
		}
		return false
	case CircuitHalfOpen:
		// 半开状态允许一个请求通过
		return cb.successes == 0
	default:
		return true
	}
}

// Success 记录成功
func (cb *CircuitBreaker) Success() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		cb.failures = 0
	case CircuitHalfOpen:
		cb.successes++
		if cb.successes >= cb.successThreshold {
			cb.state = CircuitClosed
			cb.failures = 0
			cb.successes = 0
		}
	}
}

// Failure 记录失败
func (cb *CircuitBreaker) Failure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		cb.failures++
		if cb.failures >= cb.failureThreshold {
			cb.state = CircuitOpen
			cb.lastFailure = time.Now()
		}
	case CircuitHalfOpen:
		cb.state = CircuitOpen
		cb.lastFailure = time.Now()
		cb.successes = 0
	}
}

// ============ 腾讯数据源函数 ============

// GetFullQuotes 获取 A 股 / 指数全量行情
func GetFullQuotes(ctx context.Context, client *HTTPClient, codes []string) ([]FullQuote, error) {
	if len(codes) == 0 {
		return []FullQuote{}, nil
	}

	// 添加市场前缀
	prefixedCodes := make([]string, len(codes))
	for i, code := range codes {
		prefixedCodes[i] = AddMarketPrefix(code)
	}

	resp, err := client.GetTencentQuote(ctx, strings.Join(prefixedCodes, ","))
	if err != nil {
		return nil, err
	}

	var results []FullQuote
	for _, item := range resp {
		if len(item.Fields) == 0 || item.Fields[0] == "" {
			continue
		}
		results = append(results, ParseFullQuote(item.Fields))
	}

	return results, nil
}

// GetSimpleQuotes 获取简要行情
func GetSimpleQuotes(ctx context.Context, client *HTTPClient, codes []string) ([]SimpleQuote, error) {
	if len(codes) == 0 {
		return []SimpleQuote{}, nil
	}

	// 添加 s_ 前缀和市场前缀
	prefixedCodes := make([]string, len(codes))
	for i, code := range codes {
		prefixedCodes[i] = "s_" + AddMarketPrefix(code)
	}

	resp, err := client.GetTencentQuote(ctx, strings.Join(prefixedCodes, ","))
	if err != nil {
		return nil, err
	}

	var results []SimpleQuote
	for _, item := range resp {
		if len(item.Fields) == 0 || item.Fields[0] == "" {
			continue
		}
		results = append(results, ParseSimpleQuote(item.Fields))
	}

	return results, nil
}

// GetFundFlows 获取资金流向
func GetFundFlows(ctx context.Context, client *HTTPClient, codes []string) ([]FundFlow, error) {
	if len(codes) == 0 {
		return []FundFlow{}, nil
	}

	// 添加 ff_ 前缀和市场前缀
	prefixedCodes := make([]string, len(codes))
	for i, code := range codes {
		prefixedCodes[i] = "ff_" + AddMarketPrefix(code)
	}

	resp, err := client.GetTencentQuote(ctx, strings.Join(prefixedCodes, ","))
	if err != nil {
		return nil, err
	}

	var results []FundFlow
	for _, item := range resp {
		if len(item.Fields) == 0 || item.Fields[0] == "" {
			continue
		}
		results = append(results, ParseFundFlow(item.Fields))
	}

	return results, nil
}

// ============ 补充工具函数（用于解析器） ============

// ioutilReadAll 兼容旧版本 Go 的 io.ReadAll
func ioutilReadAll(r io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	return buf.Bytes(), err
}

// littleEndian 小端序（用于兼容）
var littleEndian = binary.LittleEndian
