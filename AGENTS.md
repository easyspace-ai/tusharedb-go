# TushareDB-Go Project Guide for AI Agents

## Project Overview

TushareDB-Go is a Go implementation of TushareDB - a library for Chinese A-share stock market data management. It provides a unified interface for downloading, storing, and querying financial market data.

**Core Design Principles:**
- Library-first design (can be embedded in applications)
- Parquet as source-of-truth storage format
- DuckDB as SQL query engine
- Optimized for daily backtesting and stock screening
- Multi-datasource support with transparent caching

**Module Path:** `github.com/easyspace-ai/tusharedb-go`

**Go Version:** 1.25.1

## Technology Stack

### Core Dependencies
- **DuckDB:** `github.com/marcboeker/go-duckdb` v1.8.5 - Embedded analytical database
- **Parquet:** `github.com/parquet-go/parquet-go` v0.29.0 - Columnar storage format
- **Apache Arrow:** `github.com/apache/arrow-go/v18` - In-memory columnar format

### Data Sources
1. **StockSDK** (Primary) - Uses East Money/Tencent APIs, no token required
2. **Tushare** (Secondary) - Requires Tushare Pro API token

## Project Structure

```
.
├── cmd/tsdb/              # CLI entry point (minimal)
├── pkg/tsdb/              # Public API package
│   ├── client.go          # Main Client with Downloader/Reader/Screener
│   ├── unified.go         # UnifiedClient with auto-caching
│   ├── types.go           # Public types (Config, DataFrame, filters)
│   ├── downloader.go      # Data synchronization methods
│   ├── reader.go          # Data query methods
│   └── screener.go        # Stock screening methods
├── internal/              # Internal implementation
│   ├── provider/          # Data provider interfaces & implementations
│   │   ├── provider.go    # DataProvider interface definitions
│   │   ├── tushare/       # Tushare API client
│   │   └── stocksdk/      # StockSDK client (East Money/Tencent)
│   ├── dataset/           # Dataset registry and specifications
│   │   ├── spec.go        # DatasetSpec definition
│   │   ├── registry.go    # Dataset registry
│   │   └── builtin/       # Built-in dataset definitions
│   ├── storage/           # Storage layer
│   │   ├── parquet/       # Parquet file I/O
│   │   └── meta/          # Checkpoint/metadata store
│   ├── query/duckdb/      # DuckDB query engine
│   │   ├── engine.go      # SQL engine implementation
│   │   └── views.go       # View definitions
│   ├── syncer/            # Data synchronization logic
│   ├── frame/             # DataFrame type
│   ├── config/            # Configuration
│   └── domain/            # Domain models
│       ├── market/        # Market data (adjustment types)
│       ├── backtest/      # Backtesting feed
│       └── screener/      # Screening filters
├── examples/              # Example programs
├── docs/                  # Documentation
│   ├── V1_API_DRAFT.md    # Public API design
│   ├── API.md             # Interface documentation (Chinese)
│   ├── STOCKSDK_API.md    # StockSDK API reference
│   └── BOOTSTRAP_CHECKLIST.md  # Implementation checklist
└── data/                  # Default data directory (gitignored)
```

## Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│  Public API (pkg/tsdb)                                       │
│  - UnifiedClient (recommended)                               │
│  - Client (Downloader + Reader + Screener)                   │
├─────────────────────────────────────────────────────────────┤
│  Cache Mode Layer                                            │
│  - disabled: always fetch from network                       │
│  - readonly: local only, fail if missing                     │
│  - auto: use local if exists, else download (default)        │
├─────────────────────────────────────────────────────────────┤
│  Syncer (internal/syncer)                                    │
│  - SyncCore: trade_cal + stock_basic                         │
│  - SyncDatasetRange: date range sync                         │
│  - SyncDatasetIncremental: delta sync                        │
├─────────────────────────────────────────────────────────────┤
│  Query Engine (internal/query/duckdb)                        │
│  - SQL queries over Parquet views                            │
│  - Dynamic adjustment calculation (qfq/hfq)                  │
├─────────────────────────────────────────────────────────────┤
│  Storage (internal/storage)                                  │
│  - Parquet files partitioned by year/month                   │
│  - Checkpoint tracking in JSON                               │
├─────────────────────────────────────────────────────────────┤
│  Providers (internal/provider)                               │
│  - StockSDK (East Money/Tencent APIs)                        │
│  - Tushare (Tushare Pro API)                                 │
└─────────────────────────────────────────────────────────────┘
```

## Key Concepts

### Cache Modes
The `UnifiedClient` supports three cache modes:

1. **`CacheModeDisabled`** - Always fetch from network, no local storage
2. **`CacheModeReadOnly`** - Only read from local, fail if data doesn't exist
3. **`CacheModeAuto`** - Use local if available, automatically download if missing (default)

### DataFrame
Simple data container used throughout:
```go
type DataFrame struct {
    Columns []string
    Rows    []map[string]any
}
```

### Adjust Types
- `AdjustNone` / `"none"` - No adjustment
- `AdjustQFQ` / `"qfq"` - Forward adjustment (前复权)
- `AdjustHFQ` / `"hfq"` - Backward adjustment (后复权)

Adjustment is computed dynamically in SQL using adj_factor, not materialized.

### Dataset Storage Layout
```
data/
├── lake/
│   ├── trade_cal/              # Calendar data
│   ├── stock_basic/            # Stock metadata
│   ├── daily/                  # Daily OHLCV
│   │   └── year=2024/month=01/
│   ├── adj_factor/             # Adjustment factors
│   │   └── year=2024/month=01/
│   └── daily_basic/            # Daily fundamentals
│       └── year=2024/month=01/
├── meta/
│   └── checkpoints.json        # Sync tracking
└── duckdb/
    └── tusharedb.duckdb        # DuckDB database
```

## Build Commands

```bash
# Download dependencies
go mod download

# Build CLI
go build -o bin/tsdb ./cmd/tsdb

# Build examples
go build -o bin/unified_client ./examples/unified_client
go build -o bin/basic_download ./examples/basic_download

# Run tests (when implemented)
go test ./...

# Format code
go fmt ./...

# Vet code
go vet ./...
```

## Usage Examples

### Auto Mode (Recommended)
```go
client, err := tsdb.NewAutoClient("./data")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// First call downloads if needed, subsequent calls use cache
df, err := client.GetStockDaily(ctx, "000001.SZ", "20240101", "20241231", tsdb.AdjustQFQ)
```

### Real-time Mode (No Cache)
```go
client, err := tsdb.NewRealtimeClient()
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Always fetches from network
df, err := client.GetStockBasic(ctx, tsdb.StockBasicFilter{ListStatus: "L"})
```

### Offline Mode
```go
client, err := tsdb.NewOfflineClient("./data")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Only reads local data, errors if missing
df, err := client.GetStockDaily(ctx, "000001.SZ", "20240101", "20241231", tsdb.AdjustNone)
```

### Manual Sync Control
```go
client, _ := tsdb.NewAutoClient("./data")
defer client.Close()

// Initial full sync
err := client.SyncCore(ctx)  // trade_cal + stock_basic
err = client.SyncDailyRange(ctx, "20200101", "20241231")

// Daily incremental sync
err = client.SyncIncremental(ctx)  // Sync missing days only
```

## Testing Strategy

**Current Status:** Test files exist but test coverage is minimal (`internal/provider/stocksdk/*_test.go`).

**Recommended Test Areas:**
1. **Unit Tests:**
   - Provider request building
   - Retry logic
   - Pagination handling
   - Dataset registry
   - Partition path generation

2. **Integration Tests:**
   - End-to-end sync + query flow
   - Checkpoint persistence
   - Adjustment formula correctness

**Running Tests:**
```bash
go test ./...
go test -v ./internal/provider/stocksdk/...
```

## Code Style Guidelines

### Naming Conventions
- **Packages:** lowercase, no underscores (e.g., `stocksdk`, `duckdb`)
- **Exported:** PascalCase (e.g., `NewClient`, `GetStockDaily`)
- **Unexported:** camelCase (e.g., `normalizeConfig`, `fetchDaily`)
- **Interfaces:** Provider suffix for data sources (e.g., `DataProvider`)
- **Structs:** Descriptive nouns (e.g., `UnifiedClient`, `CheckpointStore`)

### Error Handling
- Wrap errors with context: `fmt.Errorf("fetch daily: %w", err)`
- Public errors defined in `pkg/tsdb/types.go`:
  - `ErrNotFound`
  - `ErrInvalidInput`
  - `ErrNotSynced`
  - `ErrQueryFailed`
  - `ErrSyncFailed`

### Documentation
- All exported types/functions must have comments
- Comments start with the name being documented
- Chinese comments are acceptable for domain-specific concepts

### Context Usage
- All IO operations accept `context.Context`
- Propagate context through call chain
- Respect cancellation

## Important Implementation Details

### Checkpoint Safety
Checkpoints are updated **after** successful file writes:
1. Fetch data from provider
2. Validate response
3. Write to temp Parquet file
4. Atomic rename to final location
5. Update manifest
6. **Then** update checkpoint

### Adjustment Calculation
Forward adjustment (qfq) formula in SQL:
```sql
price * (adj_factor / last_factor)
```

Backward adjustment (hfq) formula:
```sql
price * adj_factor
```

### Concurrency in StockSDK
StockSDK provider uses limited concurrency (5-10 goroutines) with:
- Semaphore pattern for rate limiting
- Progress tracking with mutex protection
- Small delays between requests (50-200ms)

## Security Considerations

1. **Token Management:** Tushare tokens should be passed via environment variables, not hardcoded
2. **Data Directory:** Ensure data directory has appropriate permissions
3. **Network Calls:** All external calls respect context cancellation
4. **File Permissions:** Parquet files created with 0o755 permissions

## Environment Variables

```bash
# Tushare API token (optional for StockSDK mode)
export TUSHARE_TOKEN="your-token-here"
```

## Documentation References

- `API.md` - Complete interface documentation (Chinese)
- `docs/V1_API_DRAFT.md` - Public API design specification
- `docs/STOCKSDK_API.md` - StockSDK provider API reference
- `docs/BOOTSTRAP_CHECKLIST.md` - Implementation progress tracker

## Current Limitations

1. **Screener:** Basic implementation, needs filter operator expansion
2. **Backtest Feed:** Partial implementation
3. **Test Coverage:** Minimal, needs expansion
4. **DailyBasic:** Some fields not populated from StockSDK (PE, PB from historical data)
5. **Incremental Sync:** Only by date, not true delta

## Future Work (From Checklist)

- [ ] `index_member_all` dataset support
- [ ] `index_weight` dataset support
- [ ] PIT (point-in-time) universe support
- [ ] Enhanced screener with industry filters
- [ ] Complete backtest feed implementation
- [ ] Comprehensive test suite
