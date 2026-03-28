package tushare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/easyspace-ai/stock_api/internal/provider"
)

// Client 是 Tushare API 客户端
// 实现 provider.DataProvider 接口
type Client struct {
	cfg        Config
	httpClient *http.Client
}

// NewClient 创建 Tushare 客户端
func NewClient(cfg Config) *Client {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "http://api.tushare.pro"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.Retries <= 0 {
		cfg.Retries = 3
	}
	if cfg.RetryWait <= 0 {
		cfg.RetryWait = time.Second
	}
	if cfg.PageLimit <= 0 {
		cfg.PageLimit = 5000
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "tusharedb-go/0.1"
	}

	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// ============= provider.DataProvider 接口实现 =============

// Name 返回数据源名称
func (c *Client) Name() string {
	return "tushare"
}

// HealthCheck 检查 Tushare API 是否可用
func (c *Client) HealthCheck(ctx context.Context) error {
	// 尝试获取一个小范围的交易日历来测试连接
	_, err := c.FetchTradeCalendar(ctx, "20240101", "20240105")
	return err
}

// Token 返回 API token
func (c *Client) Token() string {
	return c.cfg.Token
}

// Fetch 执行 API 请求
func (c *Client) Fetch(ctx context.Context, req provider.Request, mode provider.FetchMode) (*provider.Response, error) {
	switch mode {
	case provider.FetchModeSingle:
		return c.fetchOne(ctx, req)
	case provider.FetchModePaged, provider.FetchModeByTradeDate:
		return c.FetchAllPages(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported fetch mode: %s", mode)
	}
}

func (c *Client) fetchOne(ctx context.Context, req provider.Request) (*provider.Response, error) {
	var out *provider.Response
	err := Retry(ctx, c.cfg.Retries, c.cfg.RetryWait, func() error {
		resp, err := c.doRequest(ctx, req)
		if err != nil {
			return err
		}
		out = resp
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) doRequest(ctx context.Context, req provider.Request) (*provider.Response, error) {
	params := cloneParams(req.Params)
	if req.Page.Limit > 0 {
		params["limit"] = req.Page.Limit
	}
	if req.Page.Offset > 0 {
		params["offset"] = req.Page.Offset
	}

	payload := apiRequest{
		APIName: req.APIName,
		Token:   c.cfg.Token,
		Params:  params,
		Fields:  strings.Join(req.Fields, ","),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.cfg.UserAgent)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http status: %d", httpResp.StatusCode)
	}

	var resp apiResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if resp.Code != 0 {
		return nil, &APIError{Code: resp.Code, Msg: resp.Msg}
	}

	rows := make([]map[string]any, 0, len(resp.Data.Items))
	for _, item := range resp.Data.Items {
		row := make(map[string]any, len(resp.Data.Fields))
		for i, field := range resp.Data.Fields {
			if i < len(item) {
				row[field] = item[i]
			}
		}
		rows = append(rows, row)
	}

	return &provider.Response{
		Fields: resp.Data.Fields,
		Rows:   rows,
	}, nil
}

func cloneParams(in map[string]any) map[string]any {
	if len(in) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
