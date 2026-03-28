package models

import (
	"time"

	"gorm.io/gorm"
)

// Stock 股票基础信息
type Stock struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Code      string         `gorm:"uniqueIndex;size:20" json:"code"`
	Name      string         `gorm:"size:50" json:"name"`
	Market    string         `gorm:"size:10" json:"market"` // sh, sz, hk, us
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// StockPrice 股票实时价格
type StockPrice struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"changePercent"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	PreClose      float64 `json:"preClose"`
	Volume        int64   `json:"volume"`
	Amount        float64 `json:"amount"`
	UpdateTime    string  `json:"updateTime"`
}

// KLineData K线数据
type KLineData struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
	Code   string  `json:"code"`
}

// TradeLevelDetail AI给出的买卖区间
type TradeLevelDetail struct {
	Buy    float64 `json:"buy"`
	Sell   float64 `json:"sell"`
	Reason string  `json:"reason"`
}

// TradeLevelResult AI交易区间输出
type TradeLevelResult struct {
	ShortTerm TradeLevelDetail `json:"shortTerm"`
	MidTerm   TradeLevelDetail `json:"midTerm"`
	LongTerm  TradeLevelDetail `json:"longTerm"`
	Raw       string           `json:"raw"`
}

// MinuteData 分时数据
type MinuteData struct {
	Time          string  `json:"time"`
	Price         float64 `json:"price"`
	Volume        int64   `json:"volume"`
	ChangePercent float64 `json:"changePercent"`
}

// Fund 基金基础信息
type Fund struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Code      string         `gorm:"uniqueIndex;size:20" json:"code"`
	Name      string         `gorm:"size:100" json:"name"`
	Type      string         `gorm:"size:20" json:"type"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// FundPrice 基金净值/估值
type FundPrice struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Nav           float64 `json:"nav"`           // 单位净值
	Estimate      float64 `json:"estimate"`      // 估算净值
	ChangePercent float64 `json:"changePercent"` // 估算涨跌幅
	UpdateTime    string  `json:"updateTime"`
}

// FundDetail 基金详情
type FundDetail struct {
	Code             string  `json:"code"`
	Name             string  `json:"name"`
	Type             string  `json:"type"`
	RiskLevel        string  `json:"riskLevel"`
	Manager          string  `json:"manager"`
	Company          string  `json:"company"`
	Scale            float64 `json:"scale"`
	ScaleDate        string  `json:"scaleDate"`
	InceptionDate    string  `json:"inceptionDate"`
	NavDate          string  `json:"navDate"`
	Nav              float64 `json:"nav"`
	Estimate         float64 `json:"estimate"`
	OneDayReturn     float64 `json:"oneDayReturn"`
	OneYearReturn    float64 `json:"oneYearReturn"`
	ThreeYearReturn  float64 `json:"threeYearReturn"`
	ThisYearReturn   float64 `json:"thisYearReturn"`
	SinceStartReturn float64 `json:"sinceStartReturn"`
	SharpRatio       float64 `json:"sharpRatio"`
	MaxDrawdown      float64 `json:"maxDrawdown"`
}

// FundPerformancePoint 基金历史净值点
type FundPerformancePoint struct {
	Date          string  `json:"date"`
	Nav           float64 `json:"nav"`
	AccNav        float64 `json:"accNav"`
	ChangePercent float64 `json:"changePercent"`
}

// FundHolding 基金持仓
type FundHolding struct {
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	Ratio    float64 `json:"ratio"`
	Industry string  `json:"industry"`
	Trend    string  `json:"trend"`
	Change   float64 `json:"change"`
	Type     string  `json:"type"`
}

// FundNotice 基金公告
type FundNotice struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Date     string `json:"date"`
	Category string `json:"category"`
	Url      string `json:"url"`
}

// FundOverview 基金详情概览
type FundOverview struct {
	Price         *FundPrice             `json:"price"`
	Detail        *FundDetail            `json:"detail"`
	History       []FundPerformancePoint `json:"history"`
	StockHoldings []FundHolding          `json:"stockHoldings"`
	BondHoldings  []FundHolding          `json:"bondHoldings"`
	Notices       []FundNotice           `json:"notices"`
}

// MarketIndex 市场指数
type MarketIndex struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"changePercent"`
}

// IndustryRank 行业排行
type IndustryRank struct {
	Name          string  `json:"name"`
	ChangePercent float64 `json:"changePercent"`
	LeadStock     string  `json:"leadStock"`
}

// MoneyFlow 资金流向
type MoneyFlow struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	MainFlow  float64 `json:"mainFlow"`  // 主力净流入
	SuperFlow float64 `json:"superFlow"` // 超大单净流入
}

// NewsItem 新闻快讯
type NewsItem struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Time       string `json:"time"`
	Source     string `json:"source"`
	Importance string `json:"importance"` // high, medium, normal
}

// ResearchReport 研报
type ResearchReport struct {
	Title       string `json:"title"`
	StockName   string `json:"stockName"`
	OrgName     string `json:"orgName"`
	PublishDate string `json:"publishDate"`
	Researcher  string `json:"researcher"`
	Rating      string `json:"rating"`
	InfoCode    string `json:"infoCode"` // 用于构造详情URL
	Url         string `json:"url"`      // 详情页URL
}

// StockNotice 公告
type StockNotice struct {
	Title     string `json:"title"`
	Date      string `json:"date"`
	Type      string `json:"type"`
	StockName string `json:"stockName"`
	ArtCode   string `json:"artCode"` // 公告代码
	Url       string `json:"url"`     // 详情页URL
}

// LongTigerItem 龙虎榜
type LongTigerItem struct {
	Rank          int     `json:"rank"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	ChangePercent float64 `json:"changePercent"`
	BuyAmount     string  `json:"buyAmount"`
	SellAmount    string  `json:"sellAmount"`
	Date          string  `json:"date"`
}

// HotTopic 热门话题
type HotTopic struct {
	Rank      int    `json:"rank"`
	Title     string `json:"title"`
	Desc      string `json:"desc"`
	ReadCount int64  `json:"readCount"`
	PostCount int64  `json:"postCount"`
}

// Config 系统配置
type Config struct {
	ID                uint   `gorm:"primarykey" json:"id"`
	RefreshInterval   int    `json:"refreshInterval"`
	ProxyUrl          string `json:"proxyUrl"`
	ProxyPoolEnabled  bool   `json:"proxyPoolEnabled"`
	ProxyProvider     string `json:"proxyProvider"`
	ProxyApiUrl       string `json:"proxyApiUrl"`
	ProxyApiKey       string `json:"proxyApiKey"`
	ProxyApiSecret    string `json:"proxyApiSecret"`
	ProxyRegion       string `json:"proxyRegion"`
	ProxyPoolList     string `json:"proxyPoolList"`
	ProxyPoolProtocol string `json:"proxyPoolProtocol"`
	ProxyPoolTTL      int    `json:"proxyPoolTTL"`
	ProxyPoolSize     int    `json:"proxyPoolSize"`
	Theme             string `json:"theme"`
	CustomPrimary     string `json:"customPrimary"`
	AlertPushEnabled  bool   `json:"alertPushEnabled"`
	WecomWebhook      string `json:"wecomWebhook"`
	DingtalkWebhook   string `json:"dingtalkWebhook"`
	EmailPushEnabled  bool   `json:"emailPushEnabled"`
	EmailSMTP         string `json:"emailSmtp"`
	EmailPort         int    `json:"emailPort"`
	EmailUser         string `json:"emailUser"`
	EmailPassword     string `json:"emailPassword"`
	EmailTo           string `json:"emailTo"`
	AiEnabled         bool   `json:"aiEnabled"`
	AiModel           string `json:"aiModel"`
	AiApiKey          string `json:"aiApiKey"`
	AiApiUrl          string `json:"aiApiUrl"`
	BrowserPath       string `json:"browserPath"`
	// 付费API配置
	PaidApiEnabled  bool   `json:"paidApiEnabled"`
	PaidApiProvider string `json:"paidApiProvider"` // eastmoney, ths, wind, tushare, akshare
	PaidApiKey      string `json:"paidApiKey"`
	PaidApiSecret   string `json:"paidApiSecret"`
	PaidApiUrl      string `json:"paidApiUrl"`
	// Tushare配置
	TushareToken   string `json:"tushareToken"`   // Tushare Pro Token
	TushareEnabled bool   `json:"tushareEnabled"` // 是否启用Tushare
	// AKShare配置
	AkshareEnabled bool `json:"akshareEnabled"` // 是否启用AKShare
	// 数据源优先级
	DataSourcePriority string `json:"dataSourcePriority"` // tushare, akshare（优先使用哪个）
	// AI人设
	ActivePersona string `json:"activePersona"` // 当前激活的AI人设名称
	// 更新策略
	SkipUpdateVersion string `json:"skipUpdateVersion"`
}

// VersionInfo 版本信息
type VersionInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"buildTime"`
}

// UpdateInfo 更新信息
type UpdateInfo struct {
	HasUpdate   bool   `json:"hasUpdate"`
	Version     string `json:"version"`
	CurrentVer  string `json:"currentVersion"`
	Description string `json:"description"`
	DownloadUrl string `json:"downloadUrl"`
	ReleaseUrl  string `json:"releaseUrl"`
	ReleaseDate string `json:"releaseDate"`
	SkipVersion string `json:"skipVersion"`
	Skipped     bool   `json:"skipped"`
}

// DataSourceStatus 数据源状态
type DataSourceStatus struct {
	Key          string        `json:"key"`
	Name         string        `json:"name"`
	Domain       string        `json:"domain"`
	Latency      time.Duration `json:"latency"`
	LatencyLabel string        `json:"latencyLabel"`
	LastChecked  string        `json:"lastChecked"`
	LastSuccess  string        `json:"lastSuccess"`
	Status       string        `json:"status"`
	FailCount    int           `json:"failCount"`
}

// ProxyStatus 代理池状态
type ProxyStatus struct {
	Enabled          bool   `json:"enabled"`
	PoolEnabled      bool   `json:"poolEnabled"`
	Provider         string `json:"provider"`
	ActiveProxies    int    `json:"activeProxies"`
	ExpiresAt        string `json:"expiresAt"`
	ExpiresInSeconds int    `json:"expiresInSeconds"`
	LastFetch        string `json:"lastFetch"`
	LastError        string `json:"lastError"`
}

// DataPipelineStatus 数据通道总览
type DataPipelineStatus struct {
	MarketSources []DataSourceStatus     `json:"marketSources"`
	Financial     map[string]interface{} `json:"financial"`
	Proxy         ProxyStatus            `json:"proxy"`
	GeneratedAt   string                 `json:"generatedAt"`
}

// AIMessage AI聊天消息
type AIMessage struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	SessionID string    `gorm:"index;size:50" json:"sessionId"`
	Role      string    `gorm:"size:20" json:"role"` // user, assistant, system
	Content   string    `gorm:"type:text" json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// AIAnalysisResult AI分析结果
type AIAnalysisResult struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	StockCode  string    `gorm:"index;size:20" json:"stockCode"`
	StockName  string    `gorm:"size:50" json:"stockName"`
	Analysis   string    `gorm:"type:text" json:"analysis"`
	Suggestion string    `gorm:"size:20" json:"suggestion"` // buy, hold, sell
	CreatedAt  time.Time `json:"createdAt"`
}

// ProAnalysisCache 专业分析缓存，保证5秒内可出结果
type ProAnalysisCache struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	StockCode    string    `gorm:"size:20;index:idx_pro_cache,unique" json:"stockCode"`
	AnalysisType string    `gorm:"size:20;index:idx_pro_cache,unique" json:"analysisType"`
	MasterStyle  string    `gorm:"size:30;index:idx_pro_cache,unique" json:"masterStyle"`
	Content      string    `gorm:"type:text" json:"content"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// AIChatRequest AI聊天请求
type AIChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"sessionId"`
	StockCode string `json:"stockCode,omitempty"`
	FundCode  string `json:"fundCode,omitempty"`
}

// AIChatResponse AI聊天响应
type AIChatResponse struct {
	Content   string `json:"content"`
	SessionID string `json:"sessionId"`
	Done      bool   `json:"done"`
}

// Position 持仓信息
type Position struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	StockCode     string         `gorm:"index;size:20" json:"stockCode"`          // 股票代码
	StockName     string         `gorm:"size:50" json:"stockName"`                // 股票名称
	BuyPrice      float64        `json:"buyPrice"`                                // 买入价格
	BuyDate       string         `gorm:"size:20" json:"buyDate"`                  // 买入日期
	Quantity      int            `json:"quantity"`                                // 持仓数量（股）
	CostPrice     float64        `json:"costPrice"`                               // 成本价（含手续费）
	TargetPrice   float64        `json:"targetPrice"`                             // 目标价
	StopLossPrice float64        `json:"stopLossPrice"`                           // 止损价
	Notes         string         `gorm:"type:text" json:"notes"`                  // 备注（买入理由等）
	Status        string         `gorm:"size:20;default:'holding'" json:"status"` // 状态：holding持有, sold已卖出
	SellPrice     float64        `json:"sellPrice"`                               // 卖出价格
	SellDate      string         `gorm:"size:20" json:"sellDate"`                 // 卖出日期
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// FundPosition 基金持仓信息
type FundPosition struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	FundCode    string         `gorm:"index;size:20" json:"fundCode"`
	FundName    string         `gorm:"size:100" json:"fundName"`
	BuyNav      float64        `json:"buyNav"`
	BuyDate     string         `gorm:"size:20" json:"buyDate"`
	Share       float64        `json:"share"`
	CostNav     float64        `json:"costNav"`
	TargetNav   float64        `json:"targetNav"`
	StopLossNav float64        `json:"stopLossNav"`
	Notes       string         `gorm:"type:text" json:"notes"`
	Status      string         `gorm:"size:20;default:'holding'" json:"status"`
	RedeemNav   float64        `json:"redeemNav"`
	RedeemDate  string         `gorm:"size:20" json:"redeemDate"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// ==================== 期货相关模型 ====================

// Futures 期货基础信息
type Futures struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Code      string         `gorm:"uniqueIndex;size:20" json:"code"` // 合约代码，如 AU2406
	Name      string         `gorm:"size:100" json:"name"`            // 合约名称
	Exchange  string         `gorm:"size:20" json:"exchange"`         // 交易所：SHFE上期所, DCE大商所, CZCE郑商所, CFFEX中金所, INE能源中心
	Product   string         `gorm:"size:20" json:"product"`          // 品种代码，如 AU
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// FuturesPrice 期货实时价格
type FuturesPrice struct {
	Code          string  `json:"code"`          // 合约代码
	Name          string  `json:"name"`          // 合约名称
	Price         float64 `json:"price"`         // 最新价
	Change        float64 `json:"change"`        // 涨跌额
	ChangePercent float64 `json:"changePercent"` // 涨跌幅
	Open          float64 `json:"open"`          // 开盘价
	High          float64 `json:"high"`          // 最高价
	Low           float64 `json:"low"`           // 最低价
	PreClose      float64 `json:"preClose"`      // 昨收价
	PreSettle     float64 `json:"preSettle"`     // 昨结算价
	Settle        float64 `json:"settle"`        // 今结算价
	Volume        int64   `json:"volume"`        // 成交量（手）
	Amount        float64 `json:"amount"`        // 成交额
	OpenInterest  int64   `json:"openInterest"`  // 持仓量
	UpdateTime    string  `json:"updateTime"`    // 更新时间
	Exchange      string  `json:"exchange"`      // 交易所
}

// FuturesProduct 期货品种信息
type FuturesProduct struct {
	Code     string `json:"code"`     // 品种代码，如 AU
	Name     string `json:"name"`     // 品种名称，如 黄金
	Exchange string `json:"exchange"` // 交易所
	Unit     string `json:"unit"`     // 交易单位，如 1000克/手
	Margin   string `json:"margin"`   // 保证金比例
}

// ==================== 美股相关模型 ====================

// USStock 美股基础信息
type USStock struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Symbol    string         `gorm:"uniqueIndex;size:20" json:"symbol"` // 股票代码，如 AAPL
	Name      string         `gorm:"size:100" json:"name"`              // 公司名称
	NameCN    string         `gorm:"size:100" json:"nameCn"`            // 中文名称
	Exchange  string         `gorm:"size:20" json:"exchange"`           // 交易所：NYSE, NASDAQ, AMEX
	Sector    string         `gorm:"size:50" json:"sector"`             // 行业
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// USStockPrice 美股实时价格
type USStockPrice struct {
	Symbol        string  `json:"symbol"`        // 股票代码
	Name          string  `json:"name"`          // 公司名称
	NameCN        string  `json:"nameCn"`        // 中文名称
	Price         float64 `json:"price"`         // 最新价（美元）
	Change        float64 `json:"change"`        // 涨跌额
	ChangePercent float64 `json:"changePercent"` // 涨跌幅
	Open          float64 `json:"open"`          // 开盘价
	High          float64 `json:"high"`          // 最高价
	Low           float64 `json:"low"`           // 最低价
	PreClose      float64 `json:"preClose"`      // 昨收价
	Volume        int64   `json:"volume"`        // 成交量
	Amount        float64 `json:"amount"`        // 成交额
	MarketCap     float64 `json:"marketCap"`     // 市值
	PE            float64 `json:"pe"`            // 市盈率
	UpdateTime    string  `json:"updateTime"`    // 更新时间
	Exchange      string  `json:"exchange"`      // 交易所
}

// ==================== 全球指数相关模型 ====================

// GlobalIndex 全球指数
type GlobalIndex struct {
	Code          string  `json:"code"`          // 指数代码
	Name          string  `json:"name"`          // 指数名称
	NameCN        string  `json:"nameCn"`        // 中文名称
	Price         float64 `json:"price"`         // 最新点位
	Change        float64 `json:"change"`        // 涨跌点数
	ChangePercent float64 `json:"changePercent"` // 涨跌幅
	Open          float64 `json:"open"`          // 开盘点位
	High          float64 `json:"high"`          // 最高点位
	Low           float64 `json:"low"`           // 最低点位
	PreClose      float64 `json:"preClose"`      // 昨收点位
	UpdateTime    string  `json:"updateTime"`    // 更新时间
	Region        string  `json:"region"`        // 地区：asia, europe, america
	Country       string  `json:"country"`       // 国家代码：us, jp, kr, hk, tw, uk, de, fr, au, in, sg, ca
	Status        string  `json:"status"`        // 状态：trading, closed
}

// ==================== 港股相关模型 ====================

// HKStock 港股基础信息
type HKStock struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Code      string         `gorm:"uniqueIndex;size:20" json:"code"` // 股票代码，如 00700
	Name      string         `gorm:"size:100" json:"name"`            // 公司名称
	NameCN    string         `gorm:"size:100" json:"nameCn"`          // 中文名称
	Lot       int            `json:"lot"`                             // 每手股数
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// HKStockPrice 港股实时价格
type HKStockPrice struct {
	Code          string  `json:"code"`          // 股票代码
	Name          string  `json:"name"`          // 公司名称
	Price         float64 `json:"price"`         // 最新价（港元）
	Change        float64 `json:"change"`        // 涨跌额
	ChangePercent float64 `json:"changePercent"` // 涨跌幅
	Open          float64 `json:"open"`          // 开盘价
	High          float64 `json:"high"`          // 最高价
	Low           float64 `json:"low"`           // 最低价
	PreClose      float64 `json:"preClose"`      // 昨收价
	Volume        int64   `json:"volume"`        // 成交量
	Amount        float64 `json:"amount"`        // 成交额
	MarketCap     float64 `json:"marketCap"`     // 市值
	PE            float64 `json:"pe"`            // 市盈率
	UpdateTime    string  `json:"updateTime"`    // 更新时间
}

// ==================== 外汇相关模型 ====================

// ForexRate 外汇汇率
type ForexRate struct {
	Pair          string  `json:"pair"`          // 货币对，如 USDCNY
	Name          string  `json:"name"`          // 名称，如 美元/人民币
	Rate          float64 `json:"rate"`          // 汇率
	Change        float64 `json:"change"`        // 涨跌
	ChangePercent float64 `json:"changePercent"` // 涨跌幅
	High          float64 `json:"high"`          // 最高
	Low           float64 `json:"low"`           // 最低
	UpdateTime    string  `json:"updateTime"`    // 更新时间
}

// ==================== 股票提醒相关模型 ====================

// StockAlert 股票价格提醒
type StockAlert struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	StockCode       string         `gorm:"index;size:20" json:"stockCode"` // 股票代码
	StockName       string         `gorm:"size:50" json:"stockName"`       // 股票名称
	AlertType       string         `gorm:"size:20" json:"alertType"`       // 提醒类型：price（股价提醒）、change（涨跌提醒）
	TargetValue     float64        `json:"targetValue"`                    // 目标值（股价或涨跌幅百分比）
	Condition       string         `gorm:"size:10" json:"condition"`       // 条件：above（高于）、below（低于）
	Enabled         bool           `gorm:"default:true" json:"enabled"`    // 是否启用
	Triggered       bool           `gorm:"default:false" json:"triggered"` // 是否已触发
	TriggeredAt     *time.Time     `json:"triggeredAt"`                    // 触发时间
	TriggeredPrice  float64        `json:"triggeredPrice"`                 // 触发时的价格
	TriggeredChange float64        `json:"triggeredChange"`                // 触发时的涨跌幅
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// FundAlert 基金提醒
type FundAlert struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	FundCode        string         `gorm:"index;size:20" json:"fundCode"`
	FundName        string         `gorm:"size:100" json:"fundName"`
	AlertType       string         `gorm:"size:20" json:"alertType"` // nav, change
	TargetValue     float64        `json:"targetValue"`
	Condition       string         `gorm:"size:10" json:"condition"` // above/below
	Enabled         bool           `gorm:"default:true" json:"enabled"`
	Triggered       bool           `gorm:"default:false" json:"triggered"`
	TriggeredAt     *time.Time     `json:"triggeredAt"`
	TriggeredNav    float64        `json:"triggeredNav"`
	TriggeredChange float64        `json:"triggeredChange"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// AlertNotification 提醒通知（用于前端显示）
type AlertNotification struct {
	ID            uint    `json:"id"`
	StockCode     string  `json:"stockCode"`
	StockName     string  `json:"stockName"`
	AlertType     string  `json:"alertType"`
	TargetValue   float64 `json:"targetValue"`
	CurrentPrice  float64 `json:"currentPrice"`
	CurrentChange float64 `json:"currentChange"`
	Message       string  `json:"message"`
	Time          string  `json:"time"`
	AssetType     string  `json:"assetType"`
}
