package marketdata

import "time"

// DTOs for JSON/API responses (no GORM).

type LongTigerRank struct {
	TradeDate        string  `json:"tradeDate"`
	SecurityCode     string  `json:"securityCode"`
	SecuCode         string  `json:"secuCode"`
	SecurityNameAbbr string  `json:"securityNameAbbr"`
	ClosePrice       float64 `json:"closePrice"`
	ChangeRate       float64 `json:"changeRate"`
	AccumAmount      float64 `json:"accumAmount"`
	BillboardBuyAmt  float64 `json:"billboardBuyAmt"`
	BillboardSellAmt float64 `json:"billboardSellAmt"`
	BillboardNetAmt  float64 `json:"billboardNetAmt"`
	BillboardDealAmt float64 `json:"billboardDealAmt"`
	Explanation      string  `json:"explanation"`
	TurnoverRate     float64 `json:"turnoverRate"`
	FreeMarketCap    float64 `json:"freeMarketCap"`
}

type HotStock struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Value      float64 `json:"value"`
	Increment  int     `json:"increment"`
	RankChange int     `json:"rankChange"`
	Percent    float64 `json:"percent"`
	Current    float64 `json:"current"`
	Chg        float64 `json:"chg"`
	Exchange   string  `json:"exchange"`
}

type HotEvent struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Tag         string `json:"tag"`
	Pic         string `json:"pic"`
	Hot         int    `json:"hot"`
	StatusCount int    `json:"statusCount"`
}

type HotTopic struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Hot        int    `json:"hot"`
	StockCount int    `json:"stockCount"`
}

type MarketNews struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Source      string    `json:"source"`
	Url         string    `json:"url"`
	PublishTime time.Time `json:"publishTime"`
	StockCodes  string    `json:"stockCodes"`
	Tags        string    `json:"tags"`
}

type ResearchReport struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	StockCode   string `json:"stockCode"`
	StockName   string `json:"stockName"`
	Author      string `json:"author"`
	OrgName     string `json:"orgName"`
	PublishDate string `json:"publishDate"`
	ReportType  string `json:"reportType"`
	Url         string `json:"url"`
}

type StockNotice struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	StockCode   string `json:"stockCode"`
	StockName   string `json:"stockName"`
	NoticeType  string `json:"noticeType"`
	PublishDate string `json:"publishDate"`
	UpdateTime  string `json:"updateTime"`
	Url         string `json:"url"`
}

type IndustryMoneyRank struct {
	IndustryName  string  `json:"industryName"`
	ChangePct     float64 `json:"changePct"`
	Inflow        float64 `json:"inflow"`
	Outflow       float64 `json:"outflow"`
	NetInflow     float64 `json:"netInflow"`
	NetRatio      float64 `json:"netRatio"`
	LeadStock     string  `json:"leadStock"`
	LeadStockCode string  `json:"leadStockCode"`
	LeadChange    float64 `json:"leadChange"`
	LeadPrice     float64 `json:"leadPrice"`
	LeadNetRatio  float64 `json:"leadNetRatio"`
}

type IndustryRank struct {
	IndustryName  string  `json:"industryName"`
	IndustryCode  string  `json:"industryCode"`
	ChangePct     float64 `json:"changePct"`
	ChangePct5d   float64 `json:"changePct5d"`
	ChangePct20d  float64 `json:"changePct20d"`
	LeadStock     string  `json:"leadStock"`
	LeadStockCode string  `json:"leadStockCode"`
	LeadChange    float64 `json:"leadChange"`
	LeadPrice     float64 `json:"leadPrice"`
}

type StockMoneyRank struct {
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	ChangePct    float64 `json:"changePct"`
	TurnoverRate float64 `json:"turnoverRate"`
	Amount       float64 `json:"amount"`
	OutAmount    float64 `json:"outAmount"`
	InAmount     float64 `json:"inAmount"`
	NetAmount    float64 `json:"netAmount"`
	NetRatio     float64 `json:"netRatio"`
	R0Out        float64 `json:"r0Out"`
	R0In         float64 `json:"r0In"`
	R0Net        float64 `json:"r0Net"`
	R0Ratio      float64 `json:"r0Ratio"`
	R3Out        float64 `json:"r3Out"`
	R3In         float64 `json:"r3In"`
	R3Net        float64 `json:"r3Net"`
	R3Ratio      float64 `json:"r3Ratio"`
}

type GlobalIndex struct {
	Name       string  `json:"name"`
	Code       string  `json:"code"`
	Price      float64 `json:"price"`
	Change     float64 `json:"change"`
	ChangePct  float64 `json:"changePct"`
	UpdateTime string  `json:"updateTime"`
	// Region is the upstream board key: common | america | asia | europe | other
	Region string `json:"region"`
}

type InvestCalendarItem struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

type MoneyFlowInfo struct {
	Date                string  `json:"date"`
	MainNetInflow       float64 `json:"mainNetInflow"`
	MainNetRatio        float64 `json:"mainNetRatio"`
	SuperLargeNetInflow float64 `json:"superLargeNetInflow"`
	LargeNetInflow      float64 `json:"largeNetInflow"`
	MediumNetInflow     float64 `json:"mediumNetInflow"`
	SmallNetInflow      float64 `json:"smallNetInflow"`
}
