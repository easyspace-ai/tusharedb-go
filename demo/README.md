# TushareDB-Go 接口演示（面向第三方集成）

本目录按功能分类提供可独立运行的 `main` 示例，用于验证网络环境与集成方式。  
运行前请在仓库根目录执行：`go mod download`。

## 运行方式

在仓库根目录下：

```bash
# 单个示例
go run ./demo/realtime_quotes/

# 批量（可选；含 lake_query 时需本机 CGO 可用）
bash demo/run_all.sh
```

**CGO / DuckDB**：`lake_query` 依赖 `github.com/marcboeker/go-duckdb`，需要启用 CGO 且可链接 DuckDB（与主库一致）。若仅测实时接口，可先 `go run` 其它子目录。

**说明**：部分接口依赖第三方站点可用性（雪球、百度等可能因 Cookie/IP 受限），失败时示例会打印错误并继续后续步骤，便于逐项排查。

## 示例一览

| 目录 | 说明 |
|------|------|
| `realtime_quotes` | 实时行情：Failover、`GetQuotesBySource` 指定数据源 |
| `kline_minute` | K 线、指定数据源 K 线、腾讯分时 |
| `market_universe` | 市场概览、股票/板块/行业/指数列表、涨跌幅榜 |
| `capital_flow` | 个股/板块/市场资金流、北向资金 |
| `news_announcements` | 资讯、公告列表、研报、正文拉取（东财） |
| `global_assets` | 热门美股/港股、全球指数、外汇、期货、加密货币 |
| `market_special` | 龙虎榜、停牌、分红 |
| `lake_query` | Parquet + DuckDB 离线湖：`UnifiedClient` 查询日线（需 CGO） |
| `stockapi_sdk` | 底层 `stockapi`：K 线、分时、批量分时 |
| `marketdata_rest` | 底层 `marketdata`：东财类 REST 封装示例 |
| `client_admin` | `realtimedata`：缓存清理、限流统计 |

## 环境与合规

- 请遵守各数据提供方服务条款与请求频率限制；库内已带域名级限流，仍建议控制并发。
- `TUSHARE_TOKEN`：仅在使用 Tushare 主源的 `UnifiedClient` 配置时需要；`NewAutoClient` 默认 StockSDK 时通常不必填。
