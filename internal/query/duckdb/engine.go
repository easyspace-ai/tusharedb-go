package duckdb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/easyspace-ai/stock_api/internal/frame"

	_ "github.com/marcboeker/go-duckdb"
)

type Config struct {
	DuckDBPath string
	DataDir    string
}

type StockBasicFilter struct {
	TSCode     string
	ListStatus string
	Market     string
}

type TradeCalendarFilter struct {
	Exchange  string
	StartDate string
	EndDate   string
	IsOpen    *bool
}

type Filter struct {
	Field string
	Op    string
	Value any
}

type Order struct {
	Field string
	Order string
}

type UniverseSpec struct {
	ListStatus string
	Markets    []string
	ExcludeST  bool
}

type ScreenRequest struct {
	TradeDate string
	Universe  UniverseSpec
	Filters   []Filter
	OrderBy   []Order
	Limit     int
	Fields    []string
}

type BarRequest struct {
	TSCodes   []string
	StartDate string
	EndDate   string
	Adjust    string
	WithBasic bool
}

type Engine struct {
	cfg Config
	db  *sql.DB
}

func NewEngine(cfg Config) (*Engine, error) {
	if cfg.DuckDBPath != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.DuckDBPath), 0o755); err != nil {
			return nil, fmt.Errorf("create duckdb dir: %w", err)
		}
	}

	db, err := sql.Open("duckdb", cfg.DuckDBPath)
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}

	e := &Engine{
		cfg: cfg,
		db:  db,
	}

	if err := e.initViews(context.Background()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("init views: %w", err)
	}

	return e, nil
}

func (e *Engine) Close() error {
	if e.db != nil {
		return e.db.Close()
	}
	return nil
}

func (e *Engine) initViews(ctx context.Context) error {
	// 注册 Parquet 文件视图
	viewDefs := []struct {
		name    string
		pattern string
	}{
		{"v_trade_cal", "lake/trade_cal/**/*.parquet"},
		{"v_stock_basic", "lake/stock_basic/**/*.parquet"},
		{"v_daily_raw", "lake/daily/**/*.parquet"},
		{"v_adj_factor", "lake/adj_factor/**/*.parquet"},
		{"v_daily_basic", "lake/daily_basic/**/*.parquet"},
	}

	for _, v := range viewDefs {
		parquetPath := filepath.Join(e.cfg.DataDir, v.pattern)
		sql := fmt.Sprintf(`CREATE OR REPLACE VIEW %s AS SELECT * FROM parquet_scan('%s')`, v.name, parquetPath)
		if _, err := e.db.ExecContext(ctx, sql); err != nil {
			// 如果 Parquet 文件不存在，视图依然可以创建（查询时会返回空）
		}
	}

	return nil
}

func (e *Engine) Query(ctx context.Context, sql string, args ...any) (*frame.DataFrame, error) {
	rows, err := e.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	return e.rowsToDataFrame(rows)
}

func (e *Engine) rowsToDataFrame(rows *sql.Rows) (*frame.DataFrame, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	df := &frame.DataFrame{
		Columns: cols,
		Rows:    make([]map[string]any, 0),
	}

	for rows.Next() {
		values := make([]any, len(cols))
		valuePtrs := make([]any, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]any)
		for i, col := range cols {
			row[col] = values[i]
		}
		df.Rows = append(df.Rows, row)
	}

	return df, rows.Err()
}

func (e *Engine) GetStockBasic(ctx context.Context, filter StockBasicFilter) (*frame.DataFrame, error) {
	var conditions []string
	var args []any

	if filter.TSCode != "" {
		conditions = append(conditions, "ts_code = ?")
		args = append(args, filter.TSCode)
	}
	if filter.ListStatus != "" {
		conditions = append(conditions, "list_status = ?")
		args = append(args, filter.ListStatus)
	}
	if filter.Market != "" {
		conditions = append(conditions, "market = ?")
		args = append(args, filter.Market)
	}

	sql := "SELECT * FROM v_stock_basic"
	if len(conditions) > 0 {
		sql += " WHERE " + strings.Join(conditions, " AND ")
	}

	return e.Query(ctx, sql, args...)
}

func (e *Engine) GetTradeCalendar(ctx context.Context, filter TradeCalendarFilter) (*frame.DataFrame, error) {
	var conditions []string
	var args []any

	if filter.Exchange != "" {
		conditions = append(conditions, "exchange = ?")
		args = append(args, filter.Exchange)
	}
	if filter.StartDate != "" {
		conditions = append(conditions, "cal_date >= ?")
		args = append(args, filter.StartDate)
	}
	if filter.EndDate != "" {
		conditions = append(conditions, "cal_date <= ?")
		args = append(args, filter.EndDate)
	}
	if filter.IsOpen != nil {
		if *filter.IsOpen {
			conditions = append(conditions, "is_open = '1'")
		} else {
			conditions = append(conditions, "is_open = '0'")
		}
	}

	sql := "SELECT * FROM v_trade_cal"
	if len(conditions) > 0 {
		sql += " WHERE " + strings.Join(conditions, " AND ")
	}
	sql += " ORDER BY cal_date"

	return e.Query(ctx, sql, args...)
}

func (e *Engine) GetStockDaily(ctx context.Context, tsCode, startDate, endDate, adjust string) (*frame.DataFrame, error) {
	return e.GetMultipleStocksDaily(ctx, []string{tsCode}, startDate, endDate, adjust)
}

func (e *Engine) GetMultipleStocksDaily(ctx context.Context, tsCodes []string, startDate, endDate, adjust string) (*frame.DataFrame, error) {
	if adjust == "none" || adjust == "" {
		return e.getRawDaily(ctx, tsCodes, startDate, endDate)
	}
	return e.getAdjustedDaily(ctx, tsCodes, startDate, endDate, adjust)
}

func (e *Engine) getRawDaily(ctx context.Context, tsCodes []string, startDate, endDate string) (*frame.DataFrame, error) {
	placeholders := make([]string, len(tsCodes))
	args := make([]any, len(tsCodes)+2)
	for i, tsCode := range tsCodes {
		placeholders[i] = "?"
		args[i] = tsCode
	}
	args[len(tsCodes)] = startDate
	args[len(tsCodes)+1] = endDate

	sql := fmt.Sprintf(`
		SELECT * FROM v_daily_raw
		WHERE ts_code IN (%s)
		AND trade_date >= ?
		AND trade_date <= ?
		ORDER BY ts_code, trade_date
	`, strings.Join(placeholders, ","))

	return e.Query(ctx, sql, args...)
}

func (e *Engine) getAdjustedDaily(ctx context.Context, tsCodes []string, startDate, endDate, adjust string) (*frame.DataFrame, error) {
	placeholders := make([]string, len(tsCodes))
	args := make([]any, len(tsCodes)+2)
	for i, tsCode := range tsCodes {
		placeholders[i] = "?"
		args[i] = tsCode
	}
	args[len(tsCodes)] = startDate
	args[len(tsCodes)+1] = endDate

	var sql string
	if adjust == "qfq" {
		// 前复权: price * (adj_factor / latest_factor)
		sql = fmt.Sprintf(`
			WITH latest_adj AS (
				SELECT ts_code, MAX(adj_factor) as last_factor
				FROM v_adj_factor
				WHERE ts_code IN (%s)
				GROUP BY ts_code
			)
			SELECT
				d.ts_code,
				d.trade_date,
				d.open * (a.adj_factor / l.last_factor) as open,
				d.high * (a.adj_factor / l.last_factor) as high,
				d.low * (a.adj_factor / l.last_factor) as low,
				d.close * (a.adj_factor / l.last_factor) as close,
				d.pre_close * (a.adj_factor / l.last_factor) as pre_close,
				d.change * (a.adj_factor / l.last_factor) as change,
				d.pct_chg,
				d.vol,
				d.amount
			FROM v_daily_raw d
			LEFT JOIN v_adj_factor a ON d.ts_code = a.ts_code AND d.trade_date = a.trade_date
			LEFT JOIN latest_adj l ON d.ts_code = l.ts_code
			WHERE d.ts_code IN (%s)
			AND d.trade_date >= ?
			AND d.trade_date <= ?
			ORDER BY d.ts_code, d.trade_date
		`, strings.Join(placeholders, ","), strings.Join(placeholders, ","))
	} else {
		// 后复权: price * adj_factor
		sql = fmt.Sprintf(`
			SELECT
				d.ts_code,
				d.trade_date,
				d.open * a.adj_factor as open,
				d.high * a.adj_factor as high,
				d.low * a.adj_factor as low,
				d.close * a.adj_factor as close,
				d.pre_close * a.adj_factor as pre_close,
				d.change * a.adj_factor as change,
				d.pct_chg,
				d.vol,
				d.amount
			FROM v_daily_raw d
			LEFT JOIN v_adj_factor a ON d.ts_code = a.ts_code AND d.trade_date = a.trade_date
			WHERE d.ts_code IN (%s)
			AND d.trade_date >= ?
			AND d.trade_date <= ?
			ORDER BY d.ts_code, d.trade_date
		`, strings.Join(placeholders, ","))
	}

	return e.Query(ctx, sql, append(args, args...))
}

func (e *Engine) GetAdjFactor(ctx context.Context, tsCode, startDate, endDate string) (*frame.DataFrame, error) {
	sql := `
		SELECT * FROM v_adj_factor
		WHERE ts_code = ?
		AND trade_date >= ?
		AND trade_date <= ?
		ORDER BY trade_date
	`
	return e.Query(ctx, sql, tsCode, startDate, endDate)
}

func (e *Engine) GetDailyBasic(ctx context.Context, tsCode, startDate, endDate string) (*frame.DataFrame, error) {
	sql := `
		SELECT * FROM v_daily_basic
		WHERE ts_code = ?
		AND trade_date >= ?
		AND trade_date <= ?
		ORDER BY trade_date
	`
	return e.Query(ctx, sql, tsCode, startDate, endDate)
}

func (e *Engine) RunScreen(ctx context.Context, req ScreenRequest) (*frame.DataFrame, error) {
	// 选股器实现待完善
	return &frame.DataFrame{}, nil
}

func (e *Engine) LoadBars(ctx context.Context, req BarRequest) (*frame.DataFrame, error) {
	// 回测数据流实现待完善
	return &frame.DataFrame{}, nil
}

func (e *Engine) LoadTradingDates(ctx context.Context, startDate, endDate string) ([]string, error) {
	filter := TradeCalendarFilter{
		StartDate: startDate,
		EndDate:   endDate,
		IsOpen:    &[]bool{true}[0],
	}

	df, err := e.GetTradeCalendar(ctx, filter)
	if err != nil {
		return nil, err
	}

	dates := make([]string, 0, len(df.Rows))
	for _, row := range df.Rows {
		if calDate, ok := row["cal_date"].(string); ok {
			dates = append(dates, calDate)
		}
	}

	return dates, nil
}
