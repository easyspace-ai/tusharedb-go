package tsdb

import (
	"context"

	"github.com/easyspace-ai/stock_api/internal/query/duckdb"
)

type Screener struct {
	client *Client
}

func (s *Screener) Run(ctx context.Context, req ScreenRequest) (*DataFrame, error) {
	return s.client.engine.RunScreen(ctx, duckdb.ScreenRequest{
		TradeDate: req.TradeDate,
		Universe: duckdb.UniverseSpec{
			ListStatus: req.Universe.ListStatus,
			Markets:    req.Universe.Markets,
			ExcludeST:  req.Universe.ExcludeST,
		},
		Filters: screenFilters(req.Filters),
		OrderBy: screenOrders(req.OrderBy),
		Limit:   req.Limit,
		Fields:  req.Fields,
	})
}

type BacktestFeed struct {
	client *Client
}

func (b *BacktestFeed) LoadBars(ctx context.Context, req BarRequest) (*DataFrame, error) {
	return b.client.engine.LoadBars(ctx, duckdb.BarRequest{
		TSCodes:   req.TSCodes,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Adjust:    string(req.Adjust),
		WithBasic: req.WithBasic,
	})
}

func (b *BacktestFeed) LoadTradingDates(ctx context.Context, startDate, endDate string) ([]string, error) {
	return b.client.engine.LoadTradingDates(ctx, startDate, endDate)
}

func screenFilters(in []Filter) []duckdb.Filter {
	out := make([]duckdb.Filter, 0, len(in))
	for _, item := range in {
		out = append(out, duckdb.Filter{
			Field: item.Field,
			Op:    item.Op,
			Value: item.Value,
		})
	}
	return out
}

func screenOrders(in []Order) []duckdb.Order {
	out := make([]duckdb.Order, 0, len(in))
	for _, item := range in {
		out = append(out, duckdb.Order{
			Field: item.Field,
			Order: string(item.Order),
		})
	}
	return out
}
