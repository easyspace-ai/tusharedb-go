package marketdata

import (
	"time"

	"github.com/go-resty/resty/v2"
)

// Config controls HTTP behavior for market fetches.
type Config struct {
	Timeout   time.Duration
	UserAgent string
}

// DefaultConfig returns sensible defaults (aligned with legacy server market repo).
func DefaultConfig() Config {
	return Config{
		Timeout:   15 * time.Second,
		UserAgent: defaultUserAgent(),
	}
}

// Client fetches third-party market data (East Money, Sina, Xueqiu, etc.).
type Client struct {
	client *resty.Client
	cfg    Config
}

// NewClient builds a Client. Zero Timeout uses DefaultConfig().Timeout.
func NewClient(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultConfig().Timeout
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = defaultUserAgent()
	}
	rc := resty.New().SetTimeout(cfg.Timeout)
	return &Client{client: rc, cfg: cfg}
}

func (c *Client) ua() string {
	return c.cfg.UserAgent
}

// rJSON begins a request whose JSON body should be unmarshaled via SetResult even when
// the server returns Content-Type: text/plain (East Money datacenter APIs do this).
func (c *Client) rJSON() *resty.Request {
	return c.client.R().ForceContentType("application/json")
}
