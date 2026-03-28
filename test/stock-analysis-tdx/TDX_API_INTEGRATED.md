# Gusheng 已集成 TDX API 清单与用法

通达信相关接口由主服务进程内模块 `internal/tdxapi` 提供，挂载在 **`/api/tdx`** 下（不再单独起 `stock-web`、也不再走反代）。

## 前缀与运行条件

| 项 | 说明 |
|----|------|
| **Base URL** | `http://<主机>:<PORT>/api/tdx`；默认主服务 `PORT` 未设时为 **8787**（见 `cmd/server/main.go`）。 |
| **路径规则** | 原独立服务文档里的 `/api/xxx` → 现 **`/api/tdx/xxx`**（只多一层 `/api/tdx`，不再多出 `/api`）。 |
| **开关** | 环境变量 **`TDX_ENABLED=0`** 时：不连接通达信、不注册 `/api/tdx/*`。 |
| **依赖** | `go.mod` 中 `replace github.com/injoyai/tdx => ../test/tdx-api`（与仓库内 TDX 源码一致）。 |

## 代码布局（扩展时对照上游）

| 文件 | 说明 |
|------|------|
| `internal/tdxapi/service.go` | `NewService`、`mux()` 路由表 |
| `internal/tdxapi/service_types.go` | `Service` 结构体 |
| `internal/tdxapi/response.go` | 统一 `Response` / `successResponse` / `errorResponse` |
| `internal/tdxapi/handlers_core.go` | 对应 `test/tdx-api/web/server.go` 中的基础 handler |
| `internal/tdxapi/handlers_extended.go` | 对应 `test/tdx-api/web/server_api_extended.go` |
| `internal/tdxapi/tasks.go` | 异步任务管理 |

从上游合并新接口时：在 `handlers_extended.go`（或 core）增加 `(*Service)` 方法，并在 `service.go` 的 `mux()` 里增加一行 `reg(...)`。

## 响应格式

多数接口返回统一 JSON：

```json
{ "code": 0, "message": "success", "data": … }
```

- `code === 0` 表示成功；非 0 时 `message` 为错误说明（`data` 常为 `null`）。
- **例外**：`GET /api/tdx/health` 直接返回 `{ "status": "healthy", "time": "<unix秒字符串>" }`，无 `code/message/data` 包裹。

## API 一览

以下为**实际注册**的方法与路径（与 `internal/tdxapi/service.go` 一致）。

### 基础行情与个股

| 方法 | 路径 | 说明 | 主要参数 |
|------|------|------|----------|
| GET | `/api/tdx/quote` | 五档行情 | `code`：单码或多个逗号分隔 |
| GET | `/api/tdx/kline` | K 线（日 K 等为前复权链路） | `code` 必填；`type`：`minute1`/`minute5`/`minute15`/`minute30`/`hour`/`day`/`week`/`month`，缺省 `day` |
| GET | `/api/tdx/minute` | 分时 | `code`；`date` 可选 `YYYYMMDD` |
| GET | `/api/tdx/trade` | 分时成交 | `code`；`date` 空则当日 |
| GET | `/api/tdx/search` | 按代码/名称模糊搜股票 | `keyword` |
| GET | `/api/tdx/stock-info` | 行情 + 日 K 摘要 + 分时 等综合 | `code` |

### 列表、批量与指数

| 方法 | 路径 | 说明 | 主要参数 |
|------|------|------|----------|
| GET | `/api/tdx/codes` | 股票代码列表（带交易所统计） | `exchange`：`sh`/`sz`/`bj`/`all`，可空 |
| POST | `/api/tdx/batch-quote` | 批量五档（≤50） | JSON：`{"codes":["000001","600519"]}` |
| GET | `/api/tdx/kline-history` | 指定类型历史 K 线（截断根数） | `code`，`type` 同 kline，`limit` 默认 100、最大 800 |
| GET | `/api/tdx/index` | 指数 K 线 | `code`，`type`，`limit` |
| GET | `/api/tdx/index/all` | 指数全历史（再可按 `limit` 取尾部） | `code`；`type` 默认 `day`；`limit` 可选 |
| GET | `/api/tdx/market-stats` | 市场涨跌平统计 | 无 |
| GET | `/api/tdx/market-count` | 各交易所证券数量 | 无 |

### ETF 与代码表

| 方法 | 路径 | 说明 | 主要参数 |
|------|------|------|----------|
| GET | `/api/tdx/etf` | ETF 列表 | `exchange`；`limit` |
| GET | `/api/tdx/stock-codes` | 全市场股票代码（缓存） | `limit`；`prefix`：`false` 时去掉交易所前缀 |
| GET | `/api/tdx/etf-codes` | 全 ETF 代码（缓存） | 同上 |

### 成交、K 线全集与交易日

| 方法 | 路径 | 说明 | 主要参数 |
|------|------|------|----------|
| GET | `/api/tdx/trade-history` | 历史分时成交分页 | `code`，`date`；`start`，`count`（默认 2000、最大 2000） |
| GET | `/api/tdx/minute-trade-all` | 全天分时成交 | `code`；`date` 可空 |
| GET | `/api/tdx/trade-history/full` | 上市以来分时成交聚合 | `code`；`start_date`/`end_date`/`before`；`include_today`；`limit` |
| GET | `/api/tdx/kline-all` | 股票全历史 K（通达信源，同下 `/kline-all/tdx`） | `code`；`type`；`limit` |
| GET | `/api/tdx/kline-all/tdx` | 同上，明确 TDX | 同上 |
| GET | `/api/tdx/kline-all/ths` | 同花顺前复权全 K（日/周/月类） | `code`；`type`；`limit` |
| GET | `/api/tdx/workday` | 是否交易日、前后邻交易日 | `date`；`count`（默认 1～30） |
| GET | `/api/tdx/workday/range` | 区间内交易日列表 | `start`，`end`：`YYYYMMDD` 或 `YYYY-MM-DD` |
| GET | `/api/tdx/income` | 基于前复权日 K 的区间收益 | `code`，`start_date`；`days` 如 `5,10,20` 逗号分隔，有默认档位 |

### 运维与任务

| 方法 | 路径 | 说明 | 主要参数 / 体 |
|------|------|------|----------------|
| GET | `/api/tdx/server-status` | 占位状态信息 | 无 |
| GET | `/api/tdx/health` | 存活探测 | 无 |
| POST | `/api/tdx/tasks/pull-kline` | 异步：批量 K 线入库 | JSON：`codes`，`tables`，`dir`，`limit`，`start_date` |
| POST | `/api/tdx/tasks/pull-trade` | 异步：分时成交入库 | JSON：`code`，`dir`，`start_year`，`end_year` |
| GET | `/api/tdx/tasks` | 任务列表 | 无 |
| GET | `/api/tdx/tasks/{id}` | 任务详情 | 路径：`id` |
| POST | `/api/tdx/tasks/{id}/cancel` | 取消任务 | 路径：`id` |

任务路径须带尾部子路径（如 `/api/tdx/tasks/<uuid>`、`/api/tdx/tasks/<uuid>/cancel`），与独立版 `/api/tasks/...` 行为一致，仅前缀变为 `/api/tdx`。

## curl 示例（主服务在 8787）

```bash
# 五档
curl "http://127.0.0.1:8787/api/tdx/quote?code=000001"

# 日 K
curl "http://127.0.0.1:8787/api/tdx/kline?code=000001&type=day"

# 批量行情
curl -X POST "http://127.0.0.1:8787/api/tdx/batch-quote" \
  -H "Content-Type: application/json" \
  -d '{"codes":["000001","600519"]}'

# 任务列表
curl "http://127.0.0.1:8787/api/tdx/tasks"
```

## 与《API_集成指南》对照

独立仓库文档中的 **`http://host:8080/api/...`** 示例，请改为 **`http://<主服务>/api/tdx/...`**，方法与 Query/Body 不变。更细的字段说明仍以 `test/tdx-api/API_集成指南.md`、`API_接口文档.md` 为准。
