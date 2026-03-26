package stocksdk

// ============ 行情数据类型 ============

// FullQuote A 股 / 指数 全量行情
type FullQuote struct {
	MarketID             string                 `json:"market_id"`
	Name                 string                 `json:"name"`
	Code                 string                 `json:"code"`
	Price                float64                `json:"price"`
	PrevClose            float64                `json:"prev_close"`
	Open                 float64                `json:"open"`
	Volume               float64                `json:"volume"`
	OuterVolume          float64                `json:"outer_volume"`
	InnerVolume          float64                `json:"inner_volume"`
	Bid                  []BidAskItem           `json:"bid"`
	Ask                  []BidAskItem           `json:"ask"`
	Time                 string                 `json:"time"`
	Change               float64                `json:"change"`
	ChangePercent        float64                `json:"change_percent"`
	High                 float64                `json:"high"`
	Low                  float64                `json:"low"`
	Volume2              float64                `json:"volume2"`
	Amount               float64                `json:"amount"`
	TurnoverRate         *float64               `json:"turnover_rate"`
	PE                   *float64               `json:"pe"`
	Amplitude            *float64               `json:"amplitude"`
	CirculatingMarketCap *float64               `json:"circulating_market_cap"`
	TotalMarketCap       *float64               `json:"total_market_cap"`
	PB                   *float64               `json:"pb"`
	LimitUp              *float64               `json:"limit_up"`
	LimitDown            *float64               `json:"limit_down"`
	VolumeRatio          *float64               `json:"volume_ratio"`
	AvgPrice             *float64               `json:"avg_price"`
	PEStatic             *float64               `json:"pe_static"`
	PEDynamic            *float64               `json:"pe_dynamic"`
	High52W              *float64               `json:"high_52w"`
	Low52W               *float64               `json:"low_52w"`
	CirculatingShares    *float64               `json:"circulating_shares"`
	TotalShares          *float64               `json:"total_shares"`
	Raw                  []string               `json:"raw"`
}

// BidAskItem 买卖盘口项
type BidAskItem struct {
	Price  float64 `json:"price"`
	Volume float64 `json:"volume"`
}

// SimpleQuote 简要行情（股票 / 指数）
type SimpleQuote struct {
	MarketID   string   `json:"market_id"`
	Name       string   `json:"name"`
	Code       string   `json:"code"`
	Price      float64  `json:"price"`
	Change     float64  `json:"change"`
	ChangePct  float64  `json:"change_pct"`
	Volume     float64  `json:"volume"`
	Amount     float64  `json:"amount"`
	MarketCap  *float64 `json:"market_cap"`
	MarketType string   `json:"market_type"`
	Raw        []string `json:"raw"`
}

// FundFlow 资金流向
type FundFlow struct {
	Code           string   `json:"code"`
	MainInflow     float64  `json:"main_inflow"`
	MainOutflow    float64  `json:"main_outflow"`
	MainNet        float64  `json:"main_net"`
	MainNetRatio   float64  `json:"main_net_ratio"`
	RetailInflow   float64  `json:"retail_inflow"`
	RetailOutflow  float64  `json:"retail_outflow"`
	RetailNet      float64  `json:"retail_net"`
	RetailNetRatio float64  `json:"retail_net_ratio"`
	TotalFlow      float64  `json:"total_flow"`
	Name           string   `json:"name"`
	Date           string   `json:"date"`
	Raw            []string `json:"raw"`
}

// HistoryKline A 股历史 K 线（日/周/月）
type HistoryKline struct {
	Date          string   `json:"date"`
	Code          string   `json:"code"`
	Open          *float64 `json:"open"`
	Close         *float64 `json:"close"`
	High          *float64 `json:"high"`
	Low           *float64 `json:"low"`
	Volume        *float64 `json:"volume"`
	Amount        *float64 `json:"amount"`
	Amplitude     *float64 `json:"amplitude"`
	ChangePercent *float64 `json:"change_percent"`
	Change        *float64 `json:"change"`
	TurnoverRate  *float64 `json:"turnover_rate"`
}

// SearchResult 股票搜索结果
type SearchResult struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Market string `json:"market"`
	Type   string `json:"type"`
}

// ============ K线周期类型 ============

// KlinePeriod K线周期
type KlinePeriod string

const (
	KlinePeriodDaily   KlinePeriod = "daily"
	KlinePeriodWeekly  KlinePeriod = "weekly"
	KlinePeriodMonthly KlinePeriod = "monthly"
)

// MinutePeriod 分钟 K线周期
type MinutePeriod string

const (
	MinutePeriod1  MinutePeriod = "1"
	MinutePeriod5  MinutePeriod = "5"
	MinutePeriod15 MinutePeriod = "15"
	MinutePeriod30 MinutePeriod = "30"
	MinutePeriod60 MinutePeriod = "60"
)

// AdjustType 复权类型
type AdjustType string

const (
	AdjustTypeNone AdjustType = ""
	AdjustTypeQFQ  AdjustType = "qfq"
	AdjustTypeHFQ  AdjustType = "hfq"
)

// MarketType 市场类型
type MarketType string

const (
	MarketTypeA  MarketType = "A"
	MarketTypeHK MarketType = "HK"
	MarketTypeUS MarketType = "US"
)

// ============ 板块数据类型 ============

// IndustryBoard 行业板块信息
type IndustryBoard struct {
	Rank                    int      `json:"rank"`
	Name                    string   `json:"name"`
	Code                    string   `json:"code"`
	Price                   *float64 `json:"price"`
	Change                  *float64 `json:"change"`
	ChangePercent           *float64 `json:"change_percent"`
	TotalMarketCap          *float64 `json:"total_market_cap"`
	TurnoverRate            *float64 `json:"turnover_rate"`
	RiseCount               *int     `json:"rise_count"`
	FallCount               *int     `json:"fall_count"`
	LeadingStock            *string  `json:"leading_stock"`
	LeadingStockChangePct   *float64 `json:"leading_stock_change_pct"`
}

// IndustryBoardConstituent 行业板块成分股
type IndustryBoardConstituent struct {
	Rank          int      `json:"rank"`
	Code          string   `json:"code"`
	Name          string   `json:"name"`
	Price         *float64 `json:"price"`
	ChangePercent *float64 `json:"change_percent"`
	Change        *float64 `json:"change"`
	Volume        *float64 `json:"volume"`
	Amount        *float64 `json:"amount"`
	Amplitude     *float64 `json:"amplitude"`
	High          *float64 `json:"high"`
	Low           *float64 `json:"low"`
	Open          *float64 `json:"open"`
	PrevClose     *float64 `json:"prev_close"`
	TurnoverRate  *float64 `json:"turnover_rate"`
	PE            *float64 `json:"pe"`
	PB            *float64 `json:"pb"`
}

// IndustryBoardSpot 行业板块实时行情
type IndustryBoardSpot struct {
	Item  string   `json:"item"`
	Value *float64 `json:"value"`
}

// IndustryBoardKline 行业板块历史 K 线
type IndustryBoardKline struct {
	Date          string   `json:"date"`
	Open          *float64 `json:"open"`
	Close         *float64 `json:"close"`
	High          *float64 `json:"high"`
	Low           *float64 `json:"low"`
	ChangePercent *float64 `json:"change_percent"`
	Change        *float64 `json:"change"`
	Volume        *float64 `json:"volume"`
	Amount        *float64 `json:"amount"`
	Amplitude     *float64 `json:"amplitude"`
	TurnoverRate  *float64 `json:"turnover_rate"`
}

// IndustryBoardMinuteTimeline 行业板块 1 分钟分时数据
type IndustryBoardMinuteTimeline struct {
	Time   string   `json:"time"`
	Open   *float64 `json:"open"`
	Close  *float64 `json:"close"`
	High   *float64 `json:"high"`
	Low    *float64 `json:"low"`
	Volume *float64 `json:"volume"`
	Amount *float64 `json:"amount"`
	Price  *float64 `json:"price"`
}

// IndustryBoardMinuteKline 行业板块分钟 K 线
type IndustryBoardMinuteKline struct {
	Time          string   `json:"time"`
	Open          *float64 `json:"open"`
	Close         *float64 `json:"close"`
	High          *float64 `json:"high"`
	Low           *float64 `json:"low"`
	ChangePercent *float64 `json:"change_percent"`
	Change        *float64 `json:"change"`
	Volume        *float64 `json:"volume"`
	Amount        *float64 `json:"amount"`
	Amplitude     *float64 `json:"amplitude"`
	TurnoverRate  *float64 `json:"turnover_rate"`
}

// ConceptBoard 概念板块信息
type ConceptBoard = IndustryBoard

// ConceptBoardSpot 概念板块实时行情
type ConceptBoardSpot = IndustryBoardSpot

// ConceptBoardConstituent 概念板块成分股
type ConceptBoardConstituent = IndustryBoardConstituent

// ConceptBoardKline 概念板块历史 K 线
type ConceptBoardKline = IndustryBoardKline

// ConceptBoardMinuteTimeline 概念板块 1 分钟分时数据
type ConceptBoardMinuteTimeline = IndustryBoardMinuteTimeline

// ConceptBoardMinuteKline 概念板块分钟 K 线
type ConceptBoardMinuteKline = IndustryBoardMinuteKline

// BoardKlineOptions 板块 K 线选项
type BoardKlineOptions struct {
	Period    KlinePeriod `json:"period"`
	Adjust    AdjustType  `json:"adjust"`
	StartDate string      `json:"start_date"`
	EndDate   string      `json:"end_date"`
}

// BoardMinuteKlineOptions 板块分钟 K 线选项
type BoardMinuteKlineOptions struct {
	Period MinutePeriod `json:"period"`
}

// ============ 期货数据类型 ============

// FuturesKline 期货历史 K 线
type FuturesKline struct {
	Date         string   `json:"date"`
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	Open         *float64 `json:"open"`
	Close        *float64 `json:"close"`
	High         *float64 `json:"high"`
	Low          *float64 `json:"low"`
	Volume       *float64 `json:"volume"`
	Amount       *float64 `json:"amount"`
	Amplitude    *float64 `json:"amplitude"`
	ChangePct    *float64 `json:"change_pct"`
	Change       *float64 `json:"change"`
	TurnoverRate *float64 `json:"turnover_rate"`
	OpenInterest *float64 `json:"open_interest"`
}

// GlobalFuturesQuote 全球期货实时行情
type GlobalFuturesQuote struct {
	Code       string   `json:"code"`
	Name       string   `json:"name"`
	Price      *float64 `json:"price"`
	Change     *float64 `json:"change"`
	ChangePct  *float64 `json:"change_pct"`
	Open       *float64 `json:"open"`
	High       *float64 `json:"high"`
	Low        *float64 `json:"low"`
	PrevSettle *float64 `json:"prev_settle"`
	Volume     *float64 `json:"volume"`
	BuyVolume  *float64 `json:"buy_volume"`
	SellVolume *float64 `json:"sell_volume"`
	OpenInt    *float64 `json:"open_int"`
}

// ============ 配置和请求类型 ============

// HistoryKlineOptions 历史 K线请求选项
type HistoryKlineOptions struct {
	Period    KlinePeriod `json:"period"`
	Adjust    AdjustType  `json:"adjust"`
	StartDate string      `json:"start_date"`
	EndDate   string      `json:"end_date"`
}

// MinuteKlineOptions 分钟 K线请求选项
type MinuteKlineOptions struct {
	Period    MinutePeriod `json:"period"`
	Adjust    AdjustType   `json:"adjust"`
	StartDate string       `json:"start_date"`
	EndDate   string       `json:"end_date"`
}

// MinuteKlineItem 分钟K线数据
type MinuteKlineItem struct {
	Time   string  `json:"time"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Volume float64 `json:"volume"`
}

// TimelineItem 分时数据项
type TimelineItem struct {
	Time     string  `json:"time"`
	Price    float64 `json:"price"`
	AvgPrice float64 `json:"avgPrice"`
	Volume   float64 `json:"volume"`
}

// TodayTimelineResponse 当日分时响应
type TodayTimelineResponse struct {
	Code      string         `json:"code"`
	PrevClose float64        `json:"prevClose"`
	Data      []TimelineItem `json:"data"`
}

// ============ API 响应类型 ============

// StockListResponse 股票列表接口响应格式
type StockListResponse struct {
	Success bool     `json:"success"`
	List    []string `json:"list"`
}

// AShareMarket A 股市场/板块类型
// - sh: 上交所（6 开头）
// - sz: 深交所（0 和 3 开头，包含创业板）
// - bj: 北交所（92 开头）
// - kc: 科创板（688 开头）
// - cy: 创业板（30 开头）
type AShareMarket string

const (
	AShareMarketSH AShareMarket = "sh"
	AShareMarketSZ AShareMarket = "sz"
	AShareMarketBJ AShareMarket = "bj"
	AShareMarketKC AShareMarket = "kc"
	AShareMarketCY AShareMarket = "cy"
)
