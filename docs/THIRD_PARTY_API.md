# TushareDB-Go 第三方集成说明（实时 + 历史）

本文档面向 **第三方开发者**：说明如何把本仓库作为 **Go 库** 集成使用，并对照 `test/Stock-AI` 中与 **A 股行情相关** 的能力说明 **已覆盖 / 未覆盖** 范围。

> **重要**：本仓库 **不提供** 官方内置的 HTTP REST 服务；下文的「接口」指 **Go 包中的导出类型与方法**。若需要对外开放 HTTP，请在自有服务中封装下列方法，并自行定义 URL 路径（文末给出映射建议）。

- **Go module**：`github.com/easyspace-ai/tusharedb-go`
- **实时与扩展行情**：`pkg/realtimedata`
- **历史湖仓查询与同步**：`pkg/tsdb`

---

## 1. 与 Stock-AI 股票相关 API 的对照

以下仅对比 **A 股行情与研究/资讯主线**（不含 Stock-AI 的自选列表、数据库持仓、AI 对话、自动更新等 **应用层** 能力）。

| Stock-AI（`StockAPI` / `App` 能力） | 本仓库 (`pkg/realtimedata` 等) | 说明 |
|-------------------------------------|--------------------------------|------|
| `GetStockPrice` + 多源轮询（东财/新浪/腾讯/雪球/百度/同花顺等）+ 失败回落 | `GetQuote` / `GetQuotes` / `GetQuotesBatch` | 默认顺序：**东财 → 新浪 → 腾讯 → 雪球 → 百度 → 同花顺**（报价-only 源不参与 K 线 failover） |
| `fetchStockPriceWithTimeout` 多源 **并发抢首包** | 无等价公开 API | 内部 `multisource.MultiSourceManager.GetStockQuotesParallel` 存在，**未** 挂到 `realtimedata.Client` |
| `GetStockPriceFrom*` / `RoundRobin` 指定源 | `GetQuotesBySource(ctx, src, codes)` | `src`：`DataSourceEastMoney` / `Sina` / `Tencent` / `Xueqiu` / `Baidu` / `Tonghuashun`（见 `pkg/realtimedata/sources.go`） |
| `GetKLineData(code, period, count)` | `GetKLine`…；指定源：`GetKLineBySource(ctx, src, …)` | 默认 failover；按源调用时不经本地 lake |
| `GetMinuteData`（分时） | `GetMinuteData(ctx, src, code)` | 与 Stock-AI 相同 **腾讯 minute 接口**；`src` 传空或 `DataSourceTencent` |
| `GetMarketIndex`（大盘指数列表） | **无直接同名** | 可用 `GetMarketOverview`、`GetIndexList`、`GetGlobalIndices` 等 **部分替代**，字段与 Stock-AI 不完全一致 |
| `GetIndustryRank` | **无直接同名** | 有 `GetIndustryList`、`GetSectorList`，**不等价** 于「行业涨跌幅排行」列表 |
| `GetMoneyFlow`（市场级列表） | `GetSectorMoneyFlow`、`GetMarketMoneyFlow`；个股为 `GetMoneyFlow(ctx, code)` | Stock-AI 返回类型为列表模型，本库资金流接口拆分更细 |
| `GetNewsList` | `GetNews` | ✓ |
| `GetResearchReports` | `GetStockReports` | ✓ |
| `GetStockNotices` | `GetStockNotices` | ✓ |
| `GetReportContent` / `GetNoticeContent` | `GetReportContent(ctx, infoCode)` / `GetNoticeContent(ctx, stockCode, artCode)` | 东财 JSON + 公告 HTML 降级抓取；研报部分条目仍可能仅能通过网页查看 |
| `GetLongTigerRank` | `GetDragonTigerList(ctx, date)` | ✓（需传 `date`，格式见方法说明） |
| `GetHotTopics` | `GetHotTopics` | ✓ |
| **自选股 / SQLite / 仓位** | **无** | 属 Stock-AI 应用数据层 |
| **Tushare / AKShare 财务** | `pkg/tsdb` + `provider` 侧 | 与 Stock-AI `UnifiedFinancial` **不同代码路径**；本库以 **同步 Parquet + DuckDB** 为主 |

**结论**：A 股 **实时 / K 线 / 分时 / 指定源 / 公告·研报正文** 已与 Stock-AI 主线对齐或接近；**仍未集成** 的包括：`fetchStockPriceWithTimeout` 式 **并发抢首包**（`GetStockQuotesParallel` 仍未挂到 `Client`）、`GetMarketIndex`/`GetIndustryRank` **同名榜单**、以及应用层自选/仓位等。（网易 / 搜狐 / 和讯等报价源已从库中移除，因 DNS 或接口不稳定。）

---

## 2. 快速开始（Go）

### 2.1 实时客户端 `pkg/realtimedata`

```go
import (
    "context"
    realtime "github.com/easyspace-ai/tusharedb-go/pkg/realtimedata"
)

func main() {
    ctx := context.Background()
    c, err := realtime.NewClient(realtime.Config{
        DataDir:       "./data",
        EnableStorage: true,
        CacheMode:     realtime.CacheModeAuto,
    })
    if err != nil {
        panic(err)
    }
    q, err := c.GetQuote(ctx, "000001.SZ")
    _ = q
}
```

- `NewDefaultClient()`：默认配置创建（忽略 error）。
- `NewClient` 内部会调用 `multisource` 的 `RegisterEnhancedSources()`，保证使用完整数据源实现。

### 2.2 历史数据客户端 `pkg/tsdb`

参见现有文档 `API.md`、`docs/V1_API_DRAFT.md`；典型入口：`NewAutoClient`、`UnifiedClient`、`Reader().GetStockDaily` 等。

---

## 3. 「接口路径」说明（Go 符号路径）

第三方若写 OpenAPI，可将下表 **方法名** 映射为自有 HTTP 路径（示例：`GET /v1/quotes` → `GetQuotes`）。

**包路径**：`github.com/easyspace-ai/tusharedb-go/pkg/realtimedata`

**类型**：`type Client struct { ... }`，通过 `NewClient(cfg Config) (*Client, error)` 构造。

**配置 `Config`**：

| 字段 | 类型 | 含义 |
|------|------|------|
| `DataDir` | `string` | 本地数据根目录（Parquet 等） |
| `EnableStorage` | `bool` | 是否写入/读取本地行情与 K 线缓存 |
| `CacheMode` | `CacheMode` | `disabled` / `readonly` / `auto`（见常量） |

---

## 4. `realtimedata.Client` 方法一览（参数与返回）

以下为 **导出方法**；返回结构体字段名与 **JSON tag** 一致，便于序列化。

### 4.1 行情

| 方法 | 参数 | 返回 | 说明 |
|------|------|------|------|
| `GetQuote` | `ctx`, `code string` | `*StockQuote`, `error` | 单票 |
| `GetQuotes` | `ctx`, `codes []string` | `[]StockQuote`, `error` | 多票；可能先读本地再拉网 |
| `GetQuotesBatch` | `ctx`, `codes []string` | `map[string]*StockQuote`, `map[string]string`（失败原因）, `error` | 批量，按批调用底层 |

**`StockQuote` JSON 字段**：`code`, `name`, `price`, `prevClose`, `open`, `high`, `low`, `volume`, `amount`, `change`, `changePct`, `time`

| `GetQuotesBySource` | `ctx`, `src DataSourceType`, `codes []string` | `[]StockQuote`, `error` | **仅走指定源**，不写本地 lake |

**数据源常量**（`pkg/realtimedata`）：`DataSourceEastMoney`、`DataSourceSina`、`DataSourceTencent`、`DataSourceXueqiu`、`DataSourceBaidu`、`DataSourceTonghuashun`（底层与 `internal/provider/multisource` 一致）。

### 4.2 K 线

| 方法 | 参数 | 返回 |
|------|------|------|
| `GetKLine` | `ctx`, `code`, `period`, `adjust`, `startDate`, `endDate`（日期建议 `YYYYMMDD`） | `[]KLineItem`, `error` |
| `GetDailyKLine` | `ctx`, `code`, `adjust`, `startDate`, `endDate` | 等同 `period="daily"` |

**`period` 常用值**：`daily`/`day`/`d`、`weekly`/`week`/`w`、`monthly`/`month`/`m`（与内部 `GetPeriodCode` 一致）。

**`adjust`**：`none`/`0`、`qfq`/`1`、`hfq`/`2`（东财链路有效；新浪/腾讯 K 线 **不复权**，忽略该参数）。

**`KLineItem` JSON**：`date`, `open`, `high`, `low`, `close`, `volume`, `amount`

| `GetKLineBySource` | `ctx`, `src`, `code`, `period`, `adjust`, `startDate`, `endDate` | `[]KLineItem`, `error` | **仅走指定源**，不写 lake |

### 4.2b 分时

| 方法 | 参数 | 返回 |
|------|------|------|
| `GetMinuteData` | `ctx`, `src DataSourceType`（`""` 或腾讯）, `code string` | `[]MinuteBar`, `error` |

**`MinuteBar` JSON**：`time`, `price`, `volume`, `changePct`

### 4.2c 公告 / 研报正文

| 方法 | 参数 | 返回 |
|------|------|------|
| `GetNoticeContent` | `ctx`, `stockCode string`（6 位或 `sh`/`sz`/`bj`）, `artCode string`（公告 id） | `string`, `error` | 聚合标题、日期与正文或跳转说明 |
| `GetReportContent` | `ctx`, `infoCode string` | `string`, `error` | 成功时为正文纯文本；失败时请用列表中的 `Url` 打开 |

### 4.3 板块 / 行业 / 股票列表

| 方法 | 返回 |
|------|------|
| `GetSectorList` | `[]Sector` → `code`, `name`, `changePct`, `leadStock`, `leadPrice` |
| `GetIndustryList` | `[]Industry` → `code`, `name`, `marketCap`, `pe`, `changePct`, `volumeRatio` |
| `GetStockList` | `[]StockInfo` → `code`, `name`, `market`, `industry`, `listDate` |

### 4.4 市场概览与指数

| 方法 | 参数 | 返回 |
|------|------|------|
| `GetMarketOverview` | `ctx` | `*MarketOverview`（上证/深证涨跌、涨跌停家数、时间等） |
| `GetIndexList` | `ctx` | `[]IndexInfo` |
| `GetTopGainers` | `ctx`, `count int` | `[]StockQuote` |
| `GetTopLosers` | `ctx`, `count int` | `[]StockQuote` |

### 4.5 资金流向

| 方法 | 参数 | 返回 |
|------|------|------|
| `GetMoneyFlow` | `ctx`, `code string` | `*MoneyFlow` |
| `GetSectorMoneyFlow` | `ctx` | `[]SectorMoneyFlow` |
| `GetMarketMoneyFlow` | `ctx` | `*MarketMoneyFlow` |
| `GetNorthboundFlow` | `ctx` | `*NorthboundFlow` |

### 4.6 资讯 / 公告 / 研报 / 热点

| 方法 | 参数 | 返回 |
|------|------|------|
| `GetNews` | `ctx`, `count int` | `[]News` |
| `GetStockNews` | `ctx`, `code`, `count` | `[]News` |
| `GetStockNotices` | `ctx`, `code`, `count` | `[]StockNotice` |
| `GetStockReports` | `ctx`, `code`, `count` | `[]StockReport` |
| `GetHotTopics` | `ctx` | `[]HotTopic` |

### 4.7 全球市场（非纯 A 股，但同在客户端）

| 方法 | 说明 |
|------|------|
| `GetPopularUSStocks` / `GetPopularHKStocks` | 热门美/港股列表 |
| `GetGlobalIndices` | 全球指数 |
| `GetGlobalNews` | `region`, `count` |

### 4.8 期货 / 虚拟币 / 外汇

| 方法 | 主要参数 |
|------|----------|
| `GetFuturesList` | — |
| `GetFuturesPrices` | `symbols []string` |
| `GetFuturesKLine` | `symbol`, `period`, `startDate`, `endDate` |
| `GetCryptoList` / `GetCryptoPrices` / `GetCryptoKLine` | 同上类似 |
| `GetForexRates` | — |

### 4.9 龙虎榜 / 停牌 / 分红

| 方法 | 参数 |
|------|------|
| `GetDragonTigerList` | `date string`（如 `YYYYMMDD`） |
| `GetStockDragonTiger` | `code string` |
| `GetSuspendedStocks` | — |
| `GetDividendInfo` | `code string` |

### 4.10 技术指标（包级函数，非 `Client` 方法）

对 **价格序列** 做纯计算：  
`CalculateMA`, `CalculateEMA`, `CalculateMACD`, `CalculateRSI`, `CalculateKDJ`, `CalculateBOLL`  
入参为 `[]float64` 或高低收序列，返回同包内 `MACD`/`KDJ`/`BOLL` 或与输入等长切片。

### 4.11 运维

| 方法 | 说明 |
|------|------|
| `ClearCache` | 清空 `RequestManager` 内存缓存 |
| `GetStats` | 缓存模式、数据目录等 |
| `GetRateLimiterStats` | `domain string`，查看限流状态 |

---

## 5. `pkg/tsdb` 历史数据 API（摘要）

面向 **Parquet + DuckDB**；典型通过 `*Reader` 调用：

| 方法 | 用途 |
|------|------|
| `GetStockBasic` | 股票基础 |
| `GetTradeCalendar` | 交易日历 |
| `GetStockDaily` | 日线 OHLCV + 复权查询 |
| `GetMultipleStocksDaily` | 多票日线 |
| `GetAdjFactor` | 复权因子 |
| `GetDailyBasic` | 每日指标 |
| `Query` | 原始 SQL |

`*Downloader` / `UnifiedClient` 提供 `SyncCore`、`SyncDailyRange` 等同步入口（详见 `API.md`）。

---

## 6. 自行封装 HTTP 时的路径建议（非官方）

若将 `realtimedata.Client` 暴露为 REST，可参考以下 **仅作命名建议**：

| 建议 HTTP | 对应 Go 方法 |
|-----------|----------------|
| `GET /v1/quotes?codes=000001.SZ,600000.SH` | `GetQuotes` |
| `GET /v1/quotes/{code}` | `GetQuote` |
| `GET /v1/klines?code=&period=&adjust=&start=&end=` | `GetKLine` |
| `GET /v1/quotes/by-source?src=&codes=` | `GetQuotesBySource` |
| `GET /v1/klines/by-source?src=&...` | `GetKLineBySource` |
| `GET /v1/minute?src=&code=` | `GetMinuteData` |
| `GET /v1/notices/content?stock=&art=` | `GetNoticeContent` |
| `GET /v1/reports/content?infoCode=` | `GetReportContent` |
| `GET /v1/market/overview` | `GetMarketOverview` |
| `GET /v1/news?count=` | `GetNews` |

请求/响应体可直接使用本文 **JSON 字段名** 定义 DTO。

---

## 7. 合规与稳定性

- 数据来源为 **公开行情站点**，页面/接口变更可能导致短暂失败；本库通过 **多源切换** 与 **限流** 降低风险。
- 生产环境请自备 **监控、超时、熔断** 与 **用户协议合规** 审查。

---

## 8. 文档索引

| 文档 | 内容 |
|------|------|
| `API.md` | 中文接口说明（偏 `pkg/tsdb`） |
| `docs/V1_API_DRAFT.md` | 公共 API 设计草案 |
| `docs/BOOTSTRAP_CHECKLIST.md` | 实现清单与限制 |
| 本文 `docs/THIRD_PARTY_API.md` | 第三方集成 + Stock-AI 对照 + `realtimedata` 全量方法 |
