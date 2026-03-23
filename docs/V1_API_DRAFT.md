# V1 API Draft

This document defines the proposed public Go API for the first usable version.

## 1. Design Goals

- Keep the public API small and stable
- Match the mental model of the Python version: downloader + reader
- Hide Parquet/DuckDB implementation details from third-party users
- Make future dataset expansion additive, not breaking

## 2. Public Package

Public package:

```text
pkg/tsdb
```

Recommended import path shape:

```go
import "github.com/your-org/tusharedb-go/pkg/tsdb"
```

## 3. Public Entry Point

```go
type Config struct {
    Token   string
    DataDir string

    DuckDBPath string
    TempDir    string

    LogLevel string
}

func NewClient(cfg Config) (*Client, error)
```

```go
type Client struct {
    // internal fields hidden
}

func (c *Client) Close() error
func (c *Client) Downloader() *Downloader
func (c *Client) Reader() *Reader
func (c *Client) Screener() *Screener
func (c *Client) BacktestFeed() *BacktestFeed
```

## 4. Public Common Types

```go
type AdjustType string

const (
    AdjustNone AdjustType = "none"
    AdjustQFQ  AdjustType = "qfq"
    AdjustHFQ  AdjustType = "hfq"
)
```

```go
type SortOrder string

const (
    Asc  SortOrder = "asc"
    Desc SortOrder = "desc"
)
```

## 5. Downloader API

```go
type Downloader struct {
    // hidden
}

func (d *Downloader) SyncCore(ctx context.Context) error
func (d *Downloader) SyncTradeCalendar(ctx context.Context, startDate, endDate string) error
func (d *Downloader) SyncStockBasic(ctx context.Context, listStatus string) error
func (d *Downloader) SyncDailyRange(ctx context.Context, startDate, endDate string) error
func (d *Downloader) SyncDailyByDate(ctx context.Context, tradeDate string) error
func (d *Downloader) SyncDailyIncremental(ctx context.Context) error
func (d *Downloader) SyncAdjFactorRange(ctx context.Context, startDate, endDate string) error
func (d *Downloader) SyncAdjFactorIncremental(ctx context.Context) error
func (d *Downloader) SyncDailyBasicRange(ctx context.Context, startDate, endDate string) error
func (d *Downloader) SyncDailyBasicIncremental(ctx context.Context) error
```

Suggested v1 behavior:

- `SyncCore` does:
  - `trade_cal`
  - `stock_basic`
  - optional latest rolling window for `daily`, `adj_factor`, `daily_basic`
- All methods are idempotent
- All sync methods support rerun without corrupting storage
- Incremental methods use locally stored checkpoints/watermarks
- Incremental sync is driven by `trade_cal`, not by natural calendar days

### 5.1 Incremental Sync Rules

For daily datasets, v1 should support two update paths:

1. Range sync
- Used for initialization, backfill, repair
- Example: `20200101 -> 20241231`

2. Incremental sync
- Used for routine updates
- Reads last successful synced trade date from metadata
- Computes missing open trading dates from `trade_cal`
- Syncs only missing dates

Suggested internal logic for `SyncDailyIncremental`:

1. Read dataset checkpoint for `daily`
2. Read trading calendar from `last_synced_trade_date + 1` to today
3. Filter `is_open = 1`
4. For each missing trade date:
   - fetch one full cross-section by `trade_date`
   - write to Parquet
   - update checkpoint only after successful commit

The same pattern should apply to:

- `adj_factor`
- `daily_basic`

### 5.2 Checkpoint Metadata

Each incremental dataset should persist:

```go
type DatasetCheckpoint struct {
    Dataset            string
    LastSyncedDate     string
    LastSuccessfulAt   time.Time
    SchemaVersion      string
}
```

Suggested checkpoint granularity in v1:

- one checkpoint per dataset
- later can evolve to partition-level checkpoint if needed

### 5.3 Pagination Requirements

Provider layer must explicitly handle Tushare pagination because some APIs:

- have hard row limits per call
- require repeated requests with `limit/offset`
- sometimes support cross-section fetch by date, which is preferable

Recommended provider contract:

```go
type PageRequest struct {
    Limit  int
    Offset int
}
```

```go
type FetchMode string

const (
    FetchModeSingle     FetchMode = "single"
    FetchModePaged      FetchMode = "paged"
    FetchModeByTradeDate FetchMode = "by_trade_date"
)
```

Per dataset, the fetch strategy should be part of `DatasetSpec`.

Examples:

- `trade_cal`: single or small-range fetch
- `stock_basic`: paged or full fetch depending API behavior
- `daily`: prefer `trade_date` cross-section fetch for incremental sync
- `adj_factor`: prefer `trade_date` cross-section fetch for incremental sync
- `daily_basic`: prefer `trade_date` cross-section fetch for incremental sync

### 5.4 Pagination Strategy Priority

When multiple fetch styles are possible, use this order:

1. fetch by `trade_date`
- Best for daily incremental updates
- Natural for cross-sectional screening/backtest datasets

2. fetch by `start_date/end_date`
- Best for controlled historical backfill

3. fetch by `limit/offset`
- Use when the API forces pagination or when result size exceeds hard limits

### 5.5 Safe Commit Behavior

For incremental sync, do not advance checkpoint early.

Required order:

1. fetch page/date batch
2. validate response
3. write Parquet temp file
4. commit manifest/file move
5. update checkpoint

If step 3-4 fails, checkpoint must remain unchanged.

## 6. Reader API

### 6.1 Filters

```go
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
```

### 6.2 Core read methods

```go
type Reader struct {
    // hidden
}

func (r *Reader) GetStockBasic(ctx context.Context, filter StockBasicFilter) (*DataFrame, error)
func (r *Reader) GetTradeCalendar(ctx context.Context, filter TradeCalendarFilter) (*DataFrame, error)
func (r *Reader) GetStockDaily(
    ctx context.Context,
    tsCode string,
    startDate string,
    endDate string,
    adjust AdjustType,
) (*DataFrame, error)

func (r *Reader) GetMultipleStocksDaily(
    ctx context.Context,
    tsCodes []string,
    startDate string,
    endDate string,
    adjust AdjustType,
) (*DataFrame, error)

func (r *Reader) GetAdjFactor(
    ctx context.Context,
    tsCode string,
    startDate string,
    endDate string,
) (*DataFrame, error)

func (r *Reader) GetDailyBasic(
    ctx context.Context,
    tsCode string,
    startDate string,
    endDate string,
) (*DataFrame, error)
```

### 6.3 SQL escape hatch

```go
func (r *Reader) Query(ctx context.Context, sql string, args ...any) (*DataFrame, error)
```

This should exist in v1, but documented as advanced usage.

## 7. Screener API

### 7.1 Request types

```go
type Filter struct {
    Field string
    Op    string
    Value any
}

type Order struct {
    Field string
    Order SortOrder
}

type ScreenRequest struct {
    TradeDate string
    Universe  UniverseSpec
    Filters   []Filter
    OrderBy   []Order
    Limit     int
    Fields    []string
}
```

```go
type UniverseSpec struct {
    ListStatus string
    Markets    []string
    ExcludeST  bool
}
```

### 7.2 Method

```go
type Screener struct {
    // hidden
}

func (s *Screener) Run(ctx context.Context, req ScreenRequest) (*DataFrame, error)
```

Suggested v1 implementation:

- Universe built from `stock_basic`
- Join `daily` + `daily_basic`
- Support basic comparison operators:
  - `=`
  - `!=`
  - `>`
  - `>=`
  - `<`
  - `<=`
  - `in`

## 8. Backtest Feed API

```go
type BarRequest struct {
    TSCodes    []string
    StartDate  string
    EndDate    string
    Adjust     AdjustType
    WithBasic  bool
}

type BacktestFeed struct {
    // hidden
}

func (b *BacktestFeed) LoadBars(ctx context.Context, req BarRequest) (*DataFrame, error)
func (b *BacktestFeed) LoadTradingDates(ctx context.Context, startDate, endDate string) ([]string, error)
```

Suggested v1 behavior:

- `LoadBars` returns long-format bars
- `WithBasic=true` left joins `daily_basic`
- Future PIT industry/universe methods can be added in v2

## 9. Public DataFrame

To keep the library easy to use, expose a simple table type instead of raw DuckDB internals.

```go
type DataFrame struct {
    Columns []string
    Rows    []map[string]any
}

func (df *DataFrame) Len() int
func (df *DataFrame) Empty() bool
func (df *DataFrame) Records() []map[string]any
```

This is not the most efficient shape, but it is stable and library-friendly for v1.

## 10. Error Strategy

Public errors should be domain-oriented:

```go
var (
    ErrNotFound      = errors.New("not found")
    ErrInvalidInput  = errors.New("invalid input")
    ErrNotSynced     = errors.New("dataset not synced")
    ErrQueryFailed   = errors.New("query failed")
    ErrSyncFailed    = errors.New("sync failed")
)
```

## 11. V1 Guaranteed Datasets

The public API should explicitly guarantee these datasets only:

- `trade_cal`
- `stock_basic`
- `daily`
- `adj_factor`
- `daily_basic`

Strongly recommended for next step:

- `index_member_all`
- `index_weight`

## 12. Example Usage

```go
client, err := tsdb.NewClient(tsdb.Config{
    Token:   "...",
    DataDir: "./data",
})
if err != nil {
    panic(err)
}
defer client.Close()

ctx := context.Background()

if err := client.Downloader().SyncCore(ctx); err != nil {
    panic(err)
}

bars, err := client.Reader().GetStockDaily(
    ctx,
    "000001.SZ",
    "20240101",
    "20241231",
    tsdb.AdjustQFQ,
)
if err != nil {
    panic(err)
}

fmt.Println(bars.Len())
```
