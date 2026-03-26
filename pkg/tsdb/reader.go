package tsdb

import (
	"context"

	"github.com/easyspace-ai/tusharedb-go/internal/query/duckdb"
)

type Reader struct {
	client *Client
}

func (r *Reader) GetStockBasic(ctx context.Context, filter StockBasicFilter) (*DataFrame, error) {
	return r.client.engine.GetStockBasic(ctx, duckDBStockBasicFilter(filter))
}

func (r *Reader) GetTradeCalendar(ctx context.Context, filter TradeCalendarFilter) (*DataFrame, error) {
	return r.client.engine.GetTradeCalendar(ctx, duckDBTradeCalendarFilter(filter))
}

func (r *Reader) GetStockDaily(ctx context.Context, tsCode, startDate, endDate string, adjust AdjustType) (*DataFrame, error) {
	return r.client.engine.GetStockDaily(ctx, tsCode, startDate, endDate, string(adjust))
}

func (r *Reader) GetMultipleStocksDaily(ctx context.Context, tsCodes []string, startDate, endDate string, adjust AdjustType) (*DataFrame, error) {
	return r.client.engine.GetMultipleStocksDaily(ctx, tsCodes, startDate, endDate, string(adjust))
}

func (r *Reader) GetAdjFactor(ctx context.Context, tsCode, startDate, endDate string) (*DataFrame, error) {
	return r.client.engine.GetAdjFactor(ctx, tsCode, startDate, endDate)
}

func (r *Reader) GetDailyBasic(ctx context.Context, tsCode, startDate, endDate string) (*DataFrame, error) {
	return r.client.engine.GetDailyBasic(ctx, tsCode, startDate, endDate)
}

func (r *Reader) Query(ctx context.Context, sql string, args ...any) (*DataFrame, error) {
	return r.client.engine.Query(ctx, sql, args...)
}

func duckDBStockBasicFilter(in StockBasicFilter) duckdb.StockBasicFilter {
	return duckdb.StockBasicFilter{
		TSCode:     in.TSCode,
		ListStatus: in.ListStatus,
		Market:     in.Market,
	}
}

func duckDBTradeCalendarFilter(in TradeCalendarFilter) duckdb.TradeCalendarFilter {
	return duckdb.TradeCalendarFilter{
		Exchange:  in.Exchange,
		StartDate: in.StartDate,
		EndDate:   in.EndDate,
		IsOpen:    in.IsOpen,
	}
}
