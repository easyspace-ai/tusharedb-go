# Bootstrap Checklist

This document is the implementation checklist for initializing the Go project.

## 1. Repository Setup

- Create `go.mod`
- Decide module path
- Add `.gitignore`
- Add root `README.md`
- Add `LICENSE` if needed
- Add `Makefile`

Suggested commands:

```bash
go mod init github.com/your-org/tusharedb-go
```

## 2. Initial Directory Structure

Create:

```text
cmd/tsdb
examples/basic_download
examples/basic_query
pkg/tsdb
internal/provider/tushare
internal/dataset
internal/dataset/builtin
internal/storage/parquet
internal/storage/meta
internal/query/duckdb
internal/syncer
internal/domain/market
internal/domain/backtest
internal/domain/screener
internal/config
internal/logging
docs/adr
```

## 3. First Files To Create

### Root

- `go.mod`
- `README.md`
- `Makefile`

### Public package

- `pkg/tsdb/client.go`
- `pkg/tsdb/downloader.go`
- `pkg/tsdb/reader.go`
- `pkg/tsdb/screener.go`
- `pkg/tsdb/types.go`

### Internal provider

- `internal/provider/tushare/client.go`
- `internal/provider/tushare/retry.go`
- `internal/provider/tushare/paginator.go`
- `internal/provider/tushare/errors.go`
- `internal/provider/tushare/models.go`

### Dataset registry

- `internal/dataset/spec.go`
- `internal/dataset/registry.go`
- `internal/dataset/builtin/trade_cal.go`
- `internal/dataset/builtin/stock_basic.go`
- `internal/dataset/builtin/daily.go`
- `internal/dataset/builtin/adj_factor.go`
- `internal/dataset/builtin/daily_basic.go`

### Storage

- `internal/storage/parquet/writer.go`
- `internal/storage/parquet/partition.go`
- `internal/storage/parquet/manifest.go`
- `internal/storage/meta/store.go`

### Query

- `internal/query/duckdb/engine.go`
- `internal/query/duckdb/views.go`

### Sync

- `internal/syncer/syncer.go`
- `internal/syncer/jobs.go`

## 4. v1 Technical Decisions To Lock Early

- Parquet is source-of-truth storage
- DuckDB is query engine only
- Daily data partitioned by `year/month`
- Public API is exposed only from `pkg/tsdb`
- Internal packages are not public contract
- Adjusted prices are computed dynamically, not materialized

## 5. Provider Layer Tasks

- Implement generic POST client for Tushare
- Add context support
- Add retry/backoff
- Add timeout config
- Add pagination support
- Add fetch strategy selection: single / paged / by-trade-date
- Normalize API errors

Borrow/reference priority:

- `ok/go-tushare/client.go`
- `ok/go-tushare/stock/basic/*`
- `ok/go-tushare/stock/market/*`

## 6. Dataset Registry Tasks

- Define `DatasetSpec`
- Define `FetchFunc`
- Define fetch mode
- Define primary keys
- Define partition keys
- Define update strategy
- Define incremental checkpoint strategy
- Register v1 builtin datasets

Required builtin datasets:

- `trade_cal`
- `stock_basic`
- `daily`
- `adj_factor`
- `daily_basic`

## 7. Storage Tasks

- Pick Parquet library
- Implement partition path builder
- Implement temp write + atomic rename
- Implement manifest file writer
- Implement dedupe-by-primary-key workflow
- Implement checkpoint / watermark metadata
- Ensure checkpoint is updated only after successful write commit

Recommended metadata to persist:

- dataset name
- schema version
- last synced date
- active files
- partition info
- last successful sync timestamp

## 8. DuckDB Query Layer Tasks

- Open DuckDB connection
- Register dataset views from Parquet paths
- Create base views:
  - `v_trade_cal`
  - `v_stock_basic`
  - `v_daily_raw`
  - `v_adj_factor`
  - `v_daily_basic`
- Create derived views:
  - `v_daily_qfq`
  - `v_daily_hfq`

## 9. Downloader Tasks

- Implement `SyncTradeCalendar`
- Implement `SyncStockBasic`
- Implement `SyncDailyRange`
- Implement `SyncDailyByDate`
- Implement `SyncDailyIncremental`
- Implement `SyncAdjFactorRange`
- Implement `SyncAdjFactorIncremental`
- Implement `SyncDailyBasicRange`
- Implement `SyncDailyBasicIncremental`
- Implement `SyncCore`

Incremental sync rules:

- Use `trade_cal` as the source of truth for missing trade dates
- Use date-driven cross-section sync for `daily`, `adj_factor`, `daily_basic`
- Persist checkpoint after each successfully committed trade date
- Support rerun after partial failure without duplicate corruption

## 10. Reader Tasks

- Implement `GetTradeCalendar`
- Implement `GetStockBasic`
- Implement `GetStockDaily`
- Implement `GetMultipleStocksDaily`
- Implement `GetAdjFactor`
- Implement `GetDailyBasic`
- Implement `Query`

## 11. Screener Tasks

- Build simple universe from `stock_basic`
- Join daily + daily_basic on `trade_date`
- Support basic filters
- Support ordering
- Support limit

## 12. Backtest Feed Tasks

- Implement `LoadTradingDates`
- Implement `LoadBars`
- Support `AdjustNone`
- Support `AdjustQFQ`
- Support `AdjustHFQ`
- Optional join with `daily_basic`

## 13. Example Programs

Create examples that act as executable documentation:

- `examples/basic_download/main.go`
- `examples/basic_query/main.go`
- `examples/backtest_feed/main.go`
- `examples/screener/main.go`

Each example should compile and stay minimal.

## 14. Testing Checklist

- Unit test provider request building
- Unit test retry logic
- Unit test pagination loop
- Unit test dataset registry
- Unit test partition path generation
- Unit test manifest persistence
- Unit test checkpoint persistence
- Unit test adjustment formulas
- Integration test sync + query for v1 datasets

Minimum useful test matrix:

- sync `trade_cal`
- sync `stock_basic`
- sync `daily`
- incremental sync `daily`
- query raw daily
- query qfq daily
- query daily_basic
- run one screen

## 15. Documentation Checklist

- Root README with quick start
- Public API examples
- Dataset support table
- Storage layout explanation
- ADR for `Parquet + DuckDB`
- ADR for public package boundary

## 16. Suggested Build Targets

Suggested `Makefile` targets:

- `make test`
- `make lint`
- `make build`
- `make example`

## 17. Suggested First Milestone

Milestone 1 should be considered complete when:

- project compiles
- `SyncCore` works
- `GetStockDaily` works
- `AdjustQFQ/HFQ` works
- one screener example runs
- public package usage is documented

## 18. Suggested Next Milestone

Milestone 2:

- `index_member_all`
- `index_weight`
- PIT universe support
- backtest feed improvements
- industry-aware screening
