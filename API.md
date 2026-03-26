# TushareDB-Go 接口文档

本文档列出 TushareDB-Go 已实现的所有接口，按架构层次分类。

## 目录

1. [Provider 层](#provider-层) - 数据源接入
2. [Syncer 层](#syncer-层) - 数据同步
3. [Engine 层](#engine-层) - 本地查询引擎
4. [UnifiedClient 层](#unifiedclient-层) - 统一客户端（推荐）
5. [配置与类型](#配置与类型)

---

## Provider 层

数据源接入接口，实现 `provider.DataProvider` 接口即可接入新的数据源。

### 接口定义

```go
type DataProvider interface {
    Name() string

    // 股票基础
    FetchStockBasic(ctx context.Context, listStatus string) ([]StockBasicRow, error)

    // 交易日历
    FetchTradeCalendar(ctx context.Context, startDate, endDate string) ([]TradeCalendarRow, error)

    // 日线数据
    FetchDaily(ctx context.Context, tradeDate string) ([]DailyRow, error)
    FetchDailyRange(ctx context.Context, startDate, endDate string) ([]DailyRow, error)

    // 复权因子
    FetchAdjFactor(ctx context.Context, tradeDate string) ([]AdjFactorRow, error)
    FetchAdjFactorRange(ctx context.Context, startDate, endDate string) ([]AdjFactorRow, error)

    // 每日指标
    FetchDailyBasic(ctx context.Context, tradeDate string) ([]DailyBasicRow, error)
    FetchDailyBasicRange(ctx context.Context, startDate, endDate string) ([]DailyBasicRow, error)
}
```

### 已实现的数据源

| 数据源 | 包路径 | 说明 |
|-------|--------|------|
| StockSDK | `internal/provider/stocksdk` | 主要数据源，支持 A 股、期货、全球市场 |
| Tushare | `internal/provider/tushare` | 备用数据源，Tushare Pro API |

### StockSDK 特有接口

除了标准 DataProvider 接口，StockSDK 还提供以下扩展功能：

```go
// 实时行情
type RealtimeQuoter interface {
    GetRealtimeQuote(ctx context.Context, tsCode string) (*Quote, error)
    GetBatchRealtimeQuote(ctx context.Context, tsCodes []string) ([]*Quote, error)
}

// K 线数据
type KlineFetcher interface {
    GetStockKline(ctx context.Context, tsCode string, start, end string) ([]Kline, error)
    GetIndustryBoardKline(ctx context.Context, code string, start, end string) ([]BoardKline, error)
    GetFuturesKline(ctx context.Context, variety string, start, end string, contract string) ([]FuturesKline, error)
}

// 股票列表
type StockLister interface {
    GetAllStockCodes(ctx context.Context) ([]string, error)
    GetAShareStockCodes(ctx context.Context, markets ...AShareMarket) ([]string, error)
}

// 全球期货行情
type GlobalFuturesQuoter interface {
    GetGlobalFuturesSpot(ctx context.Context) ([]GlobalFuturesSpot, error)
}
```

---

## Syncer 层

数据同步层，负责从 Provider 拉取数据并写入本地 Parquet 文件。

### Syncer 结构

```go
type Syncer struct {
    // 内部字段...
}

// 创建 Syncer
func NewSyncer(cfg Config, provider provider.DataProvider, registry *dataset.Registry,
    checkpoint *meta.CheckpointStore, engine *duckdb.Engine) *Syncer
```

### 同步方法

| 方法 | 签名 | 功能说明 |
|-----|------|---------|
| `SyncCore` | `SyncCore(ctx context.Context) error` | 同步核心数据（stock_basic + trade_cal） |
| `SyncStockBasic` | `SyncStockBasic(ctx context.Context, listStatus string) error` | 同步股票基础信息 |
| `SyncDatasetRange` | `SyncDatasetRange(ctx context.Context, dataset, startDate, endDate string) error` | 按日期范围同步指定数据集 |
| `SyncDatasetIncremental` | `SyncDatasetIncremental(ctx context.Context, dataset string) error` | 增量同步（从上次同步日期到最新） |
| `SyncTradeCalendar` | `SyncTradeCalendar(ctx context.Context, start, end string) error` | 同步交易日历 |

### 支持的数据集名称

在 `SyncDatasetRange` 和 `SyncDatasetIncremental` 中使用：

- `"daily"` - 日线行情
- `"adj_factor"` - 复权因子
- `"daily_basic"` - 每日指标
- `"trade_cal"` - 交易日历

---

## Engine 层

本地查询引擎，基于 DuckDB 提供 SQL 查询能力。

### Engine 结构

```go
type Engine struct {
    // 内部字段...
}

// 创建 Engine
func NewEngine(cfg Config) (*Engine, error)

// 关闭 Engine
func (e *Engine) Close() error
```

### 查询方法

| 方法 | 签名 | 功能说明 |
|-----|------|---------|
| `Query` | `Query(ctx context.Context, sql string, args ...any) (*frame.DataFrame, error)` | 执行自定义 SQL 查询 |
| `GetStockBasic` | `GetStockBasic(ctx context.Context, filter StockBasicFilter) (*frame.DataFrame, error)` | 查询股票基础信息 |
| `GetTradeCalendar` | `GetTradeCalendar(ctx context.Context, filter TradeCalendarFilter) (*frame.DataFrame, error)` | 查询交易日历 |
| `GetStockDaily` | `GetStockDaily(ctx context.Context, tsCode, startDate, endDate, adjust string) (*frame.DataFrame, error)` | 查询单只股票日线 |
| `GetMultipleStocksDaily` | `GetMultipleStocksDaily(ctx context.Context, tsCodes []string, startDate, endDate, adjust string) (*frame.DataFrame, error)` | 查询多只股票日线 |
| `GetAdjFactor` | `GetAdjFactor(ctx context.Context, tsCode, startDate, endDate string) (*frame.DataFrame, error)` | 查询复权因子 |
| `GetDailyBasic` | `GetDailyBasic(ctx context.Context, tsCode, startDate, endDate string) (*frame.DataFrame, error)` | 查询每日指标 |
| `LoadTradingDates` | `LoadTradingDates(ctx context.Context, startDate, endDate string) ([]string, error)` | 加载交易日列表 |
| `RunScreen` | `RunScreen(ctx context.Context, req ScreenRequest) (*frame.DataFrame, error)` | 选股器（待完善） |
| `LoadBars` | `LoadBars(ctx context.Context, req BarRequest) (*frame.DataFrame, error)` | 批量加载 K 线（待完善） |

### 过滤器类型

```go
// 股票基础信息过滤
type StockBasicFilter struct {
    TSCode     string  // 股票代码（精确匹配）
    ListStatus string  // 上市状态：L(上市), D(退市), P(暂停)
    Market     string  // 市场：主板、创业板、科创板等
}

// 交易日历过滤
type TradeCalendarFilter struct {
    Exchange  string   // 交易所：SSE, SZSE
    StartDate string   // 开始日期 YYYYMMDD
    EndDate   string   // 结束日期 YYYYMMDD
    IsOpen    *bool    // 是否交易日（nil 表示不限）
}
```

### 复权类型

在 `GetStockDaily` 和 `GetMultipleStocksDaily` 中使用：

- `"none"` / `""` - 不复权
- `"qfq"` - 前复权
- `"hfq"` - 后复权

---

## UnifiedClient 层

统一客户端（推荐），对外提供透明缓存的 API。

### 缓存模式

```go
type CacheMode string

const (
    CacheModeDisabled CacheMode = "disabled"  // 禁用缓存，总是从网络获取
    CacheModeReadOnly CacheMode = "readonly"  // 只读缓存，本地没有则报错
    CacheModeAuto     CacheMode = "auto"      // 自动模式，本地没有则下载并缓存（默认）
)
```

### 构造方法

| 方法 | 签名 | 功能说明 |
|-----|------|---------|
| `NewUnifiedClient` | `NewUnifiedClient(cfg UnifiedConfig) (*UnifiedClient, error)` | 创建统一客户端（标准） |
| `NewUnifiedClientWithFactory` | `NewUnifiedClientWithFactory(cfg UnifiedConfig, factories map[DataSourceType]ProviderFactory) (*UnifiedClient, error)` | 使用自定义 Provider 工厂 |
| `NewAutoClient` | `NewAutoClient(dataDir string) (*UnifiedClient, error)` | 快速创建自动缓存客户端 |
| `NewOfflineClient` | `NewOfflineClient(dataDir string) (*UnifiedClient, error)` | 快速创建离线客户端 |
| `NewRealtimeClient` | `NewRealtimeClient() (*UnifiedClient, error)` | 快速创建实时客户端 |
| `Close` | `(c *UnifiedClient) Close() error` | 关闭客户端 |

### 数据查询方法（自动处理缓存）

| 方法 | 签名 | 功能说明 |
|-----|------|---------|
| `GetStockBasic` | `GetStockBasic(ctx context.Context, filter StockBasicFilter) (*DataFrame, error)` | 获取股票基础信息 |
| `GetTradeCalendar` | `GetTradeCalendar(ctx context.Context, filter TradeCalendarFilter) (*DataFrame, error)` | 获取交易日历 |
| `GetStockDaily` | `GetStockDaily(ctx context.Context, tsCode, startDate, endDate string, adjust AdjustType) (*DataFrame, error)` | 获取单只股票日线 |
| `GetMultipleStocksDaily` | `GetMultipleStocksDaily(ctx context.Context, tsCodes []string, startDate, endDate string, adjust AdjustType) (*DataFrame, error)` | 获取多只股票日线 |
| `GetAdjFactor` | `GetAdjFactor(ctx context.Context, tsCode, startDate, endDate string) (*DataFrame, error)` | 获取复权因子 |
| `GetDailyBasic` | `GetDailyBasic(ctx context.Context, tsCode, startDate, endDate string) (*DataFrame, error)` | 获取每日指标 |
| `Query` | `Query(ctx context.Context, sql string, args ...any) (*DataFrame, error)` | 执行自定义 SQL（仅本地） |

### 同步控制方法

| 方法 | 签名 | 功能说明 |
|-----|------|---------|
| `SyncCore` | `SyncCore(ctx context.Context) error` | 手动同步核心数据 |
| `SyncDailyRange` | `SyncDailyRange(ctx context.Context, startDate, endDate string) error` | 手动同步日线范围 |
| `SyncAdjFactorRange` | `SyncAdjFactorRange(ctx context.Context, startDate, endDate string) error` | 手动同步复权因子范围 |
| `SyncDailyBasicRange` | `SyncDailyBasicRange(ctx context.Context, startDate, endDate string) error` | 手动同步每日指标范围 |
| `SyncIncremental` | `SyncIncremental(ctx context.Context) error` | 增量同步（日线+复权因子+每日指标） |
| `GetLastSyncDate` | `GetLastSyncDate(dataset string) (string, bool)` | 获取数据集最后同步日期 |

### AdjustType 复权类型

```go
type AdjustType string

const (
    AdjustNone AdjustType = "none"  // 不复权
    AdjustQFQ  AdjustType = "qfq"   // 前复权
    AdjustHFQ  AdjustType = "hfq"   // 后复权
)
```

---

## 配置与类型

### UnifiedConfig 统一客户端配置

```go
type UnifiedConfig struct {
    // 数据源配置
    PrimaryDataSource  DataSourceType  // 主数据源
    TushareToken       string          // Tushare Token
    Token              string          // Token 别名（与 TushareToken 兼容）
    StockSDKAPIKey     string          // StockSDK API Key
    FallbackDataSource DataSourceType  // 备用数据源

    // 数据目录
    DataDir    string  // 数据存储目录，默认 "./data"
    DuckDBPath string  // DuckDB 文件路径
    TempDir    string  // 临时目录
    LogLevel   string  // 日志级别

    // 缓存模式
    CacheMode CacheMode  // disabled / readonly / auto
}
```

### DataSourceType 数据源类型

```go
type DataSourceType string

const (
    DataSourceTushare  DataSourceType = "tushare"   // Tushare Pro API
    DataSourceStockSDK DataSourceType = "stocksdk"  // StockSDK API
)
```

### DataFrame 数据帧

```go
type DataFrame struct {
    Columns []string
    Rows    []map[string]any
}
```

---

## 使用示例

### 自动缓存模式（推荐）

```go
client, err := tsdb.NewAutoClient("./data")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// 自动处理：本地有则读本地，没有则下载
df, err := client.GetStockDaily(ctx, "000001.SZ", "20240101", "20241231", tsdb.AdjustQFQ)
```

### 离线模式

```go
client, err := tsdb.NewOfflineClient("./data")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// 只读本地，不会触发网络请求
df, err := client.GetStockBasic(ctx, tsdb.StockBasicFilter{ListStatus: "L"})
```

### 实时模式

```go
client, err := tsdb.NewRealtimeClient()
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// 总是从网络获取，不读不写本地
df, err := client.GetStockBasic(ctx, tsdb.StockBasicFilter{})
```

### 手动同步

```go
client, _ := tsdb.NewAutoClient("./data")
defer client.Close()

// 首次全量同步
err := client.SyncDailyRange(ctx, "20200101", "20241231")

// 每日增量同步
err = client.SyncIncremental(ctx)
```

---

## 架构层次关系

```
┌─────────────────────────────────────────────────────────────┐
│                    UnifiedClient 层                          │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │  自动缓存逻辑：CacheMode (disabled/readonly/auto)        │ │
│  └─────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                    Syncer 层                                 │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │  数据同步：从 Provider → Parquet 文件                    │ │
│  └─────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                    Engine 层 (DuckDB)                        │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │  SQL 查询：视图 → Parquet 文件扫描                       │ │
│  └─────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                    Provider 层                               │
│  ┌───────────────┐  ┌───────────────┐                       │
│  │   StockSDK    │  │   Tushare     │                       │
│  └───────────────┘  └───────────────┘                       │
└─────────────────────────────────────────────────────────────┘
```

---

## 注意事项

1. **缓存模式行为**：
   - `disabled`：适合实时行情，每次查询都走网络
   - `readonly`：适合生产环境查询，保证不触发网络请求
   - `auto`：适合开发/分析，首次慢（下载），后续快（本地）

2. **数据同步**：
   - `SyncIncremental` 依赖 `checkpoints.json` 记录的最后同步日期
   - 日线数据按年/月分区，增量同步只会下载缺失的月份

3. **复权计算**：
   - 前复权在 SQL 层通过 `adj_factor / last_factor` 实时计算
   - 需要同时存在日线和复权因子数据
