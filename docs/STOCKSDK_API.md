# StockSDK API 接口文档

本文档整理了 stock-sdk 项目支持的所有接口，用于在 Go 版本的 TushareDB 中实现 StockSDK Provider。

## 目录

1. [行情接口](#行情接口)
2. [K线接口](#k线接口)
3. [资金流向](#资金流向)
4. [行业板块](#行业板块)
5. [概念板块](#概念板块)
6. [搜索接口](#搜索接口)
7. [批量接口](#批量接口)
8. [扩展数据](#扩展数据)
9. [期货接口](#期货接口)
10. [期权接口](#期权接口)
11. [技术指标](#技术指标)

---

## 行情接口

### 获取 A 股 / 指数全量行情
```typescript
getFullQuotes(codes: string[]): Promise<FullQuote[]>
```

**参数:**
- `codes`: 股票代码数组，如 `['sz000858', 'sh600000']`

**返回:** `FullQuote[]`

### 获取简要行情
```typescript
getSimpleQuotes(codes: string[]): Promise<SimpleQuote[]>
```

**参数:**
- `codes`: 股票代码数组，如 `['sz000858', 'sh000001']`

**返回:** `SimpleQuote[]`

### 获取港股扩展行情
```typescript
getHKQuotes(codes: string[]): Promise<HKQuote[]>
```

**参数:**
- `codes`: 港股代码数组，如 `['09988', '00700']`

**返回:** `HKQuote[]`

### 获取美股简要行情
```typescript
getUSQuotes(codes: string[]): Promise<USQuote[]>
```

**参数:**
- `codes`: 美股代码数组，如 `['BABA', 'AAPL']`

**返回:** `USQuote[]`

### 获取公募基金行情
```typescript
getFundQuotes(codes: string[]): Promise<FundQuote[]>
```

**参数:**
- `codes`: 基金代码数组，如 `['000001', '110011']`

**返回:** `FundQuote[]`

---

## K线接口

### 获取 A 股历史 K 线（日/周/月）
```typescript
getHistoryKline(
  symbol: string,
  options?: eastmoney.HistoryKlineOptions
): Promise<HistoryKline[]>
```

**参数:**
- `symbol`: 股票代码
- `options`: 配置选项（周期、复权、日期范围等）

**返回:** `HistoryKline[]`

### 获取 A 股分钟 K 线或分时数据
```typescript
getMinuteKline(
  symbol: string,
  options?: eastmoney.MinuteKlineOptions
): Promise<MinuteTimeline[] | MinuteKline[]>
```

**参数:**
- `symbol`: 股票代码
- `options`: 配置选项（分钟周期等）

**返回:** `MinuteTimeline[] | MinuteKline[]`

### 获取港股历史 K 线（日/周/月）
```typescript
getHKHistoryKline(
  symbol: string,
  options?: eastmoney.HKKlineOptions
): Promise<HKUSHistoryKline[]>
```

**参数:**
- `symbol`: 港股代码
- `options`: 配置选项

**返回:** `HKUSHistoryKline[]`

### 获取美股历史 K 线（日/周/月）
```typescript
getUSHistoryKline(
  symbol: string,
  options?: eastmoney.USKlineOptions
): Promise<HKUSHistoryKline[]>
```

**参数:**
- `symbol`: 美股代码
- `options`: 配置选项

**返回:** `HKUSHistoryKline[]`

### 获取当日分时走势数据
```typescript
getTodayTimeline(code: string): Promise<TodayTimelineResponse>
```

**参数:**
- `code`: 股票代码，如 `'sz000001'` 或 `'sh600000'`

**返回:** `TodayTimelineResponse`

---

## 资金流向

### 获取资金流向
```typescript
getFundFlow(codes: string[]): Promise<FundFlow[]>
```

**参数:**
- `codes`: 股票代码数组，如 `['sz000858', 'sh600000']`

**返回:** `FundFlow[]` - 包含主力流入流出、散户流入流出等

### 获取盘口大单占比
```typescript
getPanelLargeOrder(codes: string[]): Promise<PanelLargeOrder[]>
```

**参数:**
- `codes`: 股票代码数组

**返回:** `PanelLargeOrder[]`

---

## 行业板块

### 获取行业板块名称列表
```typescript
getIndustryList(): Promise<IndustryBoard[]>
```

**返回:** `IndustryBoard[]`

### 获取行业板块实时行情
```typescript
getIndustrySpot(symbol: string): Promise<IndustryBoardSpot[]>
```

**参数:**
- `symbol`: 行业板块名称（如"小金属"）或代码（如"BK1027"）

**返回:** `IndustryBoardSpot[]`

### 获取行业板块成分股
```typescript
getIndustryConstituents(symbol: string): Promise<IndustryBoardConstituent[]>
```

**参数:**
- `symbol`: 行业板块名称或代码

**返回:** `IndustryBoardConstituent[]`

### 获取行业板块历史 K 线
```typescript
getIndustryKline(
  symbol: string,
  options?: eastmoney.IndustryBoardKlineOptions
): Promise<IndustryBoardKline[]>
```

**参数:**
- `symbol`: 行业板块名称或代码
- `options`: 配置选项

**返回:** `IndustryBoardKline[]`

### 获取行业板块分时行情
```typescript
getIndustryMinuteKline(
  symbol: string,
  options?: eastmoney.IndustryBoardMinuteKlineOptions
): Promise<IndustryBoardMinuteTimeline[] | IndustryBoardMinuteKline[]>
```

**参数:**
- `symbol`: 行业板块名称或代码
- `options`: 配置选项

**返回:** `IndustryBoardMinuteTimeline[] | IndustryBoardMinuteKline[]`

---

## 概念板块

### 获取概念板块名称列表
```typescript
getConceptList(): Promise<ConceptBoard[]>
```

**返回:** `ConceptBoard[]`

### 获取概念板块实时行情
```typescript
getConceptSpot(symbol: string): Promise<ConceptBoardSpot[]>
```

**参数:**
- `symbol`: 概念板块名称（如"人工智能"）或代码（如"BK0800"）

**返回:** `ConceptBoardSpot[]`

### 获取概念板块成分股
```typescript
getConceptConstituents(symbol: string): Promise<ConceptBoardConstituent[]>
```

**参数:**
- `symbol`: 概念板块名称或代码

**返回:** `ConceptBoardConstituent[]`

### 获取概念板块历史 K 线
```typescript
getConceptKline(
  symbol: string,
  options?: eastmoney.ConceptBoardKlineOptions
): Promise<ConceptBoardKline[]>
```

**参数:**
- `symbol`: 概念板块名称或代码
- `options`: 配置选项

**返回:** `ConceptBoardKline[]`

### 获取概念板块分时行情
```typescript
getConceptMinuteKline(
  symbol: string,
  options?: eastmoney.ConceptBoardMinuteKlineOptions
): Promise<ConceptBoardMinuteTimeline[] | ConceptBoardMinuteKline[]>
```

**参数:**
- `symbol`: 概念板块名称或代码
- `options`: 配置选项

**返回:** `ConceptBoardMinuteTimeline[] | ConceptBoardMinuteKline[]`

---

## 搜索接口

### 搜索股票
```typescript
search(keyword: string): Promise<SearchResult[]>
```

**参数:**
- `keyword`: 关键词（股票代码、名称、拼音）

**返回:** `SearchResult[]`

---

## 批量接口

### 获取 A 股代码列表
```typescript
getAShareCodeList(
  options?: tencent.GetAShareCodeListOptions | boolean
): Promise<string[]>
```

**参数:**
- `options`: 配置选项，支持 `simple`（不带前缀）、`market`（市场筛选）等

**返回:** `string[]`

### 获取美股代码列表
```typescript
getUSCodeList(
  options?: tencent.GetUSCodeListOptions | boolean
): Promise<string[]>
```

**参数:**
- `options`: 配置选项

**返回:** `string[]`

### 获取港股代码列表
```typescript
getHKCodeList(): Promise<string[]>
```

**返回:** `string[]`

### 获取基金代码列表
```typescript
getFundCodeList(): Promise<string[]>
```

**返回:** `string[]`

### 获取全部 A 股实时行情
```typescript
getAllAShareQuotes(
  options: tencent.GetAllAShareQuotesOptions = {}
): Promise<FullQuote[]>
```

**参数:**
- `options`: 配置选项，支持市场筛选等

**返回:** `FullQuote[]`

### 获取全部港股实时行情
```typescript
getAllHKShareQuotes(
  options: tencent.GetAllAShareQuotesOptions = {}
): Promise<HKQuote[]>
```

**返回:** `HKQuote[]`

### 获取全部美股实时行情
```typescript
getAllUSShareQuotes(
  options: tencent.GetAllUSQuotesOptions = {}
): Promise<USQuote[]>
```

**返回:** `USQuote[]`

### 获取全部股票实时行情（自定义代码列表）
```typescript
getAllQuotesByCodes(
  codes: string[],
  options: tencent.GetAllAShareQuotesOptions = {}
): Promise<FullQuote[]>
```

**参数:**
- `codes`: 股票代码数组
- `options`: 配置选项

**返回:** `FullQuote[]`

---

## 扩展数据

### 获取 A 股交易日历
```typescript
getTradingCalendar(): Promise<string[]>
```

**返回:** `string[]` - 交易日期字符串数组，格式如 `['1990-12-19', '1990-12-20', ...]`

### 获取股票分红派送详情
```typescript
getDividendDetail(symbol: string): Promise<DividendDetail[]>
```

**参数:**
- `symbol`: 股票代码（纯数字或带交易所前缀）

**返回:** `DividendDetail[]` - 分红派送详情列表，按报告日期降序排列

---

## 期货接口

### 获取国内期货历史 K 线
```typescript
getFuturesKline(
  symbol: string,
  options?: eastmoney.FuturesKlineOptions
): Promise<FuturesKline[]>
```

**参数:**
- `symbol`: 合约代码，如 `'rb2605'`（具体合约）或 `'RBM'`（主连）
- `options`: 配置选项

**返回:** `FuturesKline[]`

### 获取全球期货实时行情
```typescript
getGlobalFuturesSpot(
  options?: eastmoney.GlobalFuturesSpotOptions
): Promise<GlobalFuturesQuote[]>
```

**返回:** `GlobalFuturesQuote[]`

### 获取全球期货历史 K 线
```typescript
getGlobalFuturesKline(
  symbol: string,
  options?: eastmoney.GlobalFuturesKlineOptions
): Promise<FuturesKline[]>
```

**参数:**
- `symbol`: 合约代码，如 `'HG00Y'`（COMEX铜连续）
- `options`: 配置选项

**返回:** `FuturesKline[]`

### 获取期货库存品种列表
```typescript
getFuturesInventorySymbols(): Promise<FuturesInventorySymbol[]>
```

**返回:** `FuturesInventorySymbol[]`

### 获取期货库存数据
```typescript
getFuturesInventory(
  symbol: string,
  options?: eastmoney.FuturesInventoryOptions
): Promise<FuturesInventory[]>
```

**参数:**
- `symbol`: 品种代码

**返回:** `FuturesInventory[]`

### 获取 COMEX 黄金/白银库存
```typescript
getComexInventory(
  symbol: 'gold' | 'silver',
  options?: eastmoney.ComexInventoryOptions
): Promise<ComexInventory[]>
```

**参数:**
- `symbol`: `'gold'` 或 `'silver'`

**返回:** `ComexInventory[]`

---

## 期权接口

### 获取中金所股指期权 T 型报价
```typescript
getIndexOptionSpot(
  product: IndexOptionProduct,
  contract: string
): Promise<OptionTQuoteResult>
```

**参数:**
- `product`: 品种代码 `'ho'`(上证50) / `'io'`(沪深300) / `'mo'`(中证1000)
- `contract`: 合约代码，如 `'io2504'`

**返回:** `OptionTQuoteResult`

### 获取中金所股指期权合约日 K 线
```typescript
getIndexOptionKline(symbol: string): Promise<OptionKline[]>
```

**参数:**
- `symbol`: 合约代码，如 `'io2504C3600'`

**返回:** `OptionKline[]`

### 获取中金所全部期权实时行情列表
```typescript
getCFFEXOptionQuotes(
  options?: eastmoney.CFFEXOptionQuotesOptions
): Promise<CFFEXOptionQuote[]>
```

**返回:** `CFFEXOptionQuote[]`

### 获取上交所 ETF 期权到期月份列表
```typescript
getETFOptionMonths(cate: ETFOptionCate): Promise<ETFOptionMonth>
```

**参数:**
- `cate`: 品种名称，如 `'50ETF'`, `'300ETF'`

**返回:** `ETFOptionMonth`

### 获取上交所 ETF 期权到期日与剩余天数
```typescript
getETFOptionExpireDay(
  cate: ETFOptionCate,
  month: string
): Promise<ETFOptionExpireDay>
```

**参数:**
- `cate`: 品种名称
- `month`: 到期月份 `YYYY-MM`

**返回:** `ETFOptionExpireDay`

### 获取上交所 ETF 期权当日分钟行情
```typescript
getETFOptionMinute(code: string): Promise<OptionMinute[]>
```

**参数:**
- `code`: 期权代码（纯数字）

**返回:** `OptionMinute[]`

### 获取上交所 ETF 期权历史日 K 线
```typescript
getETFOptionDailyKline(code: string): Promise<OptionKline[]>
```

**参数:**
- `code`: 期权代码（纯数字）

**返回:** `OptionKline[]`

### 获取上交所 ETF 期权 5 日分钟行情
```typescript
getETFOption5DayMinute(code: string): Promise<OptionMinute[]>
```

**参数:**
- `code`: 期权代码（纯数字）

**返回:** `OptionMinute[]`

### 获取商品期权 T 型报价
```typescript
getCommodityOptionSpot(
  variety: string,
  contract: string
): Promise<OptionTQuoteResult>
```

**参数:**
- `variety`: 品种代码（如 `'au'`, `'cu'`, `'SR'`）
- `contract`: 合约代码，如 `'au2506'`

**返回:** `OptionTQuoteResult`

### 获取商品期权合约日 K 线
```typescript
getCommodityOptionKline(symbol: string): Promise<OptionKline[]>
```

**参数:**
- `symbol`: 合约代码，如 `'m2409C3200'`

**返回:** `OptionKline[]`

### 获取期权龙虎榜
```typescript
getOptionLHB(symbol: string, date: string): Promise<OptionLHBItem[]>
```

**参数:**
- `symbol`: 标的代码 `'510050'` / `'510300'` / `'159919'`
- `date`: 交易日期 `YYYY-MM-DD`

**返回:** `OptionLHBItem[]`

---

## 技术指标

### 获取带技术指标的历史 K 线
```typescript
getKlineWithIndicators(
  symbol: string,
  options: {
    market?: MarketType;
    period?: 'daily' | 'weekly' | 'monthly';
    adjust?: '' | 'qfq' | 'hfq';
    startDate?: string;
    endDate?: string;
    indicators?: IndicatorOptions;
  } = {}
): Promise<KlineWithIndicators<HistoryKline | HKUSHistoryKline>[]>
```

**参数:**
- `symbol`: 股票代码
- `options.market`: 市场类型，不传则自动识别
- `options.period`: K 线周期
- `options.adjust`: 复权类型
- `options.startDate`: 开始日期 `YYYYMMDD`
- `options.endDate`: 结束日期 `YYYYMMDD`
- `options.indicators`: 技术指标配置

**支持的技术指标:**
- `ma`: MA 均线
- `macd`: MACD
- `boll`: BOLL
- `kdj`: KDJ
- `rsi`: RSI
- `wr`: WR
- `bias`: BIAS
- `cci`: CCI
- `atr`: ATR

**返回:** `KlineWithIndicators[]`

---

## 数据类型总览

### 行情数据
- `FullQuote` - A 股 / 指数全量行情
- `SimpleQuote` - 简要行情
- `HKQuote` - 港股扩展行情
- `USQuote` - 美股行情
- `FundQuote` - 公募基金行情

### K线数据
- `HistoryKline` - A 股历史 K 线
- `MinuteTimeline` - A 股分时数据
- `MinuteKline` - A 股分钟 K 线
- `TodayTimelineResponse` - 当日分时走势响应
- `HKUSHistoryKline` - 港股/美股历史 K 线

### 板块数据
- `IndustryBoard` - 行业板块信息
- `IndustryBoardSpot` - 行业板块实时行情
- `IndustryBoardConstituent` - 行业板块成分股
- `IndustryBoardKline` - 行业板块历史 K 线
- `ConceptBoard` - 概念板块信息（同行业板块）

### 期货数据
- `FuturesKline` - 期货历史 K 线
- `GlobalFuturesQuote` - 全球期货实时行情
- `FuturesInventory` - 期货库存数据

### 期权数据
- `OptionTQuoteResult` - 期权 T 型报价结果
- `OptionKline` - 期权日 K 线
- `CFFEXOptionQuote` - 中金所期权实时行情
- `OptionLHBItem` - 期权龙虎榜条目

### 其他
- `FundFlow` - 资金流向
- `DividendDetail` - 分红派送详情
- `SearchResult` - 搜索结果

---

## 数据源提供商

StockSDK 支持多个数据源提供商：

- **腾讯** (`tencent`) - 主要用于 A 股、港股、美股行情
- **东方财富** (`eastmoney`) - 主要用于 K 线、板块、期货
- **新浪** (`sina`) - 主要用于期权

---

## 注意事项

1. **代码格式**: 不同市场的代码格式不同
   - A 股: 带 `sh`/`sz` 前缀，如 `sh600519`、`sz000001`
   - 港股: 5 位数字，如 `00700`、`09988`
   - 美股: 纯字母，如 `AAPL`、`BABA`

2. **日期格式**:
   - 输入支持 `YYYYMMDD` 和 `YYYY-MM-DD`
   - 输出通常为 `YYYY-MM-DD`

3. **复权类型**:
   - `''` 或 `null`: 不复权
   - `'qfq'`: 前复权
   - `'hfq'`: 后复权

