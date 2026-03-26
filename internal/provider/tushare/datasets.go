package tushare

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/easyspace-ai/tusharedb-go/internal/provider"
)

func (c *Client) FetchTradeCalendar(ctx context.Context, startDate, endDate string) ([]provider.TradeCalendarRow, error) {
	resp, err := c.Fetch(ctx, provider.Request{
		APIName: "trade_cal",
		Params: map[string]any{
			"start_date": startDate,
			"end_date":   endDate,
		},
		Fields: []string{"exchange", "cal_date", "is_open", "pretrade_date"},
	}, provider.FetchModeSingle)
	if err != nil {
		return nil, err
	}
	return decodeRows[provider.TradeCalendarRow](resp.Rows)
}

func (c *Client) FetchStockBasic(ctx context.Context, listStatus string) ([]provider.StockBasicRow, error) {
	resp, err := c.Fetch(ctx, provider.Request{
		APIName: "stock_basic",
		Params: map[string]any{
			"list_status": listStatus,
		},
		Fields: []string{
			"ts_code", "symbol", "name", "area", "industry",
			"fullname", "enname", "cnspell", "market", "exchange",
			"curr_type", "list_status", "list_date", "delist_date", "is_hs",
		},
	}, provider.FetchModePaged)
	if err != nil {
		return nil, err
	}
	return decodeRows[provider.StockBasicRow](resp.Rows)
}

// FetchDaily 获取日线行情数据（按 trade_date 横截面抓取）
func (c *Client) FetchDaily(ctx context.Context, tradeDate string) ([]provider.DailyRow, error) {
	resp, err := c.Fetch(ctx, provider.Request{
		APIName: "daily",
		Params: map[string]any{
			"trade_date": tradeDate,
		},
		Fields: []string{
			"ts_code", "trade_date", "open", "high", "low", "close",
			"pre_close", "change", "pct_chg", "vol", "amount",
		},
	}, provider.FetchModeByTradeDate)
	if err != nil {
		return nil, err
	}
	return decodeRows[provider.DailyRow](resp.Rows)
}

// FetchDailyRange 获取日线行情数据（按日期范围）
func (c *Client) FetchDailyRange(ctx context.Context, startDate, endDate string) ([]provider.DailyRow, error) {
	resp, err := c.Fetch(ctx, provider.Request{
		APIName: "daily",
		Params: map[string]any{
			"start_date": startDate,
			"end_date":   endDate,
		},
		Fields: []string{
			"ts_code", "trade_date", "open", "high", "low", "close",
			"pre_close", "change", "pct_chg", "vol", "amount",
		},
	}, provider.FetchModePaged)
	if err != nil {
		return nil, err
	}
	return decodeRows[provider.DailyRow](resp.Rows)
}

// FetchAdjFactor 获取复权因子数据（按 trade_date 横截面抓取）
func (c *Client) FetchAdjFactor(ctx context.Context, tradeDate string) ([]provider.AdjFactorRow, error) {
	resp, err := c.Fetch(ctx, provider.Request{
		APIName: "adj_factor",
		Params: map[string]any{
			"trade_date": tradeDate,
		},
		Fields: []string{"ts_code", "trade_date", "adj_factor"},
	}, provider.FetchModeByTradeDate)
	if err != nil {
		return nil, err
	}
	return decodeRows[provider.AdjFactorRow](resp.Rows)
}

// FetchAdjFactorRange 获取复权因子数据（按日期范围）
func (c *Client) FetchAdjFactorRange(ctx context.Context, startDate, endDate string) ([]provider.AdjFactorRow, error) {
	resp, err := c.Fetch(ctx, provider.Request{
		APIName: "adj_factor",
		Params: map[string]any{
			"start_date": startDate,
			"end_date":   endDate,
		},
		Fields: []string{"ts_code", "trade_date", "adj_factor"},
	}, provider.FetchModePaged)
	if err != nil {
		return nil, err
	}
	return decodeRows[provider.AdjFactorRow](resp.Rows)
}

// FetchDailyBasic 获取每日基本面指标数据（按 trade_date 横截面抓取）
func (c *Client) FetchDailyBasic(ctx context.Context, tradeDate string) ([]provider.DailyBasicRow, error) {
	resp, err := c.Fetch(ctx, provider.Request{
		APIName: "daily_basic",
		Params: map[string]any{
			"trade_date": tradeDate,
		},
		Fields: []string{
			"ts_code", "trade_date", "close", "turnover_rate", "turnover_rate_f",
			"volume_ratio", "pe", "pe_ttm", "pb", "ps", "ps_ttm", "dv_ratio", "dv_ttm",
			"total_share", "float_share", "free_share", "total_mv", "circ_mv",
		},
	}, provider.FetchModeByTradeDate)
	if err != nil {
		return nil, err
	}
	return decodeRows[provider.DailyBasicRow](resp.Rows)
}

// FetchDailyBasicRange 获取每日基本面指标数据（按日期范围）
func (c *Client) FetchDailyBasicRange(ctx context.Context, startDate, endDate string) ([]provider.DailyBasicRow, error) {
	resp, err := c.Fetch(ctx, provider.Request{
		APIName: "daily_basic",
		Params: map[string]any{
			"start_date": startDate,
			"end_date":   endDate,
		},
		Fields: []string{
			"ts_code", "trade_date", "close", "turnover_rate", "turnover_rate_f",
			"volume_ratio", "pe", "pe_ttm", "pb", "ps", "ps_ttm", "dv_ratio", "dv_ttm",
			"total_share", "float_share", "free_share", "total_mv", "circ_mv",
		},
	}, provider.FetchModePaged)
	if err != nil {
		return nil, err
	}
	return decodeRows[provider.DailyBasicRow](resp.Rows)
}

func decodeRows[T any](rows []map[string]any) ([]T, error) {
	if len(rows) == 0 {
		return []T{}, nil
	}
	data, err := json.Marshal(rows)
	if err != nil {
		return nil, fmt.Errorf("marshal rows: %w", err)
	}
	var out []T
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("unmarshal rows: %w", err)
	}
	return out, nil
}
