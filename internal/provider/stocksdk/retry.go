package stocksdk

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// BackoffRetryConfig 基于 backoff 库的重试配置
type BackoffRetryConfig struct {
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	RandomizationFactor float64
	UseExponentialBackoff bool // true: 指数退避, false: 固定间隔
}

// DefaultBackoffRetryConfig 默认 backoff 重试配置
var DefaultBackoffRetryConfig = BackoffRetryConfig{
	MaxRetries:            3,
	InitialInterval:       500 * time.Millisecond,
	MaxInterval:           10 * time.Second,
	Multiplier:            2.0,
	RandomizationFactor:   0.1,
	UseExponentialBackoff: true,
}

// AggressiveBackoffRetryConfig 激进的重试配置（适合不稳定网络）
var AggressiveBackoffRetryConfig = BackoffRetryConfig{
	MaxRetries:            5,
	InitialInterval:       100 * time.Millisecond,
	MaxInterval:           30 * time.Second,
	Multiplier:            2.0,
	RandomizationFactor:   0.1,
	UseExponentialBackoff: true,
}

// ConservativeBackoffRetryConfig 保守的重试配置（减少服务器压力）
var ConservativeBackoffRetryConfig = BackoffRetryConfig{
	MaxRetries:            2,
	InitialInterval:       1 * time.Second,
	MaxInterval:           5 * time.Second,
	Multiplier:            2.0,
	RandomizationFactor:   0.1,
	UseExponentialBackoff: true,
}

// RetryableError 可重试的错误
type RetryableError struct {
	Err        error
	RetryAfter time.Duration
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// IsRetryableError 检查错误是否可重试
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	// 检查是否是 RetryableError 类型
	_, ok := err.(*RetryableError)
	if ok {
		return true
	}
	// 根据错误内容判断
	errStr := err.Error()
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "EOF") ||
		strings.Contains(errStr, "temporary")
}

// NewRetryableError 创建可重试错误
func NewRetryableError(err error) error {
	return &RetryableError{Err: err}
}

// ExecuteWithBackoff 使用 backoff 执行操作
func ExecuteWithBackoff(ctx context.Context, operation func() error, cfg BackoffRetryConfig) error {
	var b backoff.BackOff

	if cfg.UseExponentialBackoff {
		expBackoff := backoff.NewExponentialBackOff()
		expBackoff.InitialInterval = cfg.InitialInterval
		expBackoff.MaxInterval = cfg.MaxInterval
		expBackoff.Multiplier = cfg.Multiplier
		expBackoff.RandomizationFactor = cfg.RandomizationFactor
		b = expBackoff
	} else {
		b = backoff.NewConstantBackOff(cfg.InitialInterval)
	}

	b = backoff.WithMaxRetries(b, uint64(cfg.MaxRetries))
	b = backoff.WithContext(b, ctx)

	return backoff.RetryNotify(
		func() error {
			err := operation()
			if err == nil {
				return nil
			}
			// 如果错误不可重试，使用 Permanent 终止重试
			if !IsRetryableError(err) {
				return backoff.Permanent(err)
			}
			return err
		},
		b,
		func(err error, duration time.Duration) {
			// 重试通知，可以在这里添加日志
			// fmt.Printf("[Retry] 将在 %v 后重试，错误: %v\n", duration, err)
		},
	)
}

// HTTPRequestWithBackoff 使用 backoff 发送 HTTP 请求
func HTTPRequestWithBackoff(
	ctx context.Context,
	client *http.Client,
	req *http.Request,
	cfg BackoffRetryConfig,
) (*http.Response, error) {
	var resp *http.Response
	var lastErr error

	err := ExecuteWithBackoff(ctx, func() error {
		// 创建新的请求（因为 Body 可能已被读取）
		r, err := http.NewRequestWithContext(ctx, req.Method, req.URL.String(), req.Body)
		if err != nil {
			return backoff.Permanent(err)
		}
		// 复制 Header
		for k, v := range req.Header {
			r.Header[k] = v
		}

		resp, err = client.Do(r)
		if err != nil {
			lastErr = err
			return NewRetryableError(err)
		}

		// 检查 HTTP 状态码
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			return NewRetryableError(lastErr)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return backoff.Permanent(fmt.Errorf("HTTP %d", resp.StatusCode))
		}

		return nil
	}, cfg)

	if err != nil {
		if lastErr != nil {
			return nil, fmt.Errorf("request failed after retries: %w", lastErr)
		}
		return nil, err
	}

	return resp, nil
}

// HTTPClientBackoffConfig 适用于东方财富 API 的重试配置
// 考虑到东方财富接口的限制，使用较短的初始间隔和适中的重试次数
func HTTPClientBackoffConfig() BackoffRetryConfig {
	return BackoffRetryConfig{
		MaxRetries:            5,
		InitialInterval:       500 * time.Millisecond,
		MaxInterval:           10 * time.Second,
		Multiplier:            2.0,
		RandomizationFactor:   0.3,
		UseExponentialBackoff: true,
	}
}
