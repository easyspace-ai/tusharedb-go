package plugin

import (
	"encoding/json"
	"time"
)

// PluginType 插件类型
type PluginType string

const (
	PluginTypeDatasource   PluginType = "datasource"
	PluginTypeNotification PluginType = "notification"
	PluginTypeAI           PluginType = "ai"
)

// Plugin 插件基础结构
type Plugin struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Type        PluginType      `json:"type"`
	Version     string          `json:"version"`
	Author      string          `json:"author"`
	Description string          `json:"description"`
	Homepage    string          `json:"homepage"`
	Enabled     bool            `json:"enabled"`
	Config      json.RawMessage `json:"config"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

// NotificationConfig 通知插件配置
type NotificationConfig struct {
	URL          string            `json:"url"`
	Method       string            `json:"method"`
	Headers      map[string]string `json:"headers"`
	Params       map[string]string `json:"params"`
	BodyTemplate interface{}       `json:"bodyTemplate"`
	ContentType  string            `json:"contentType"` // json, form, text
}

// DatasourceConfig 数据源插件配置
type DatasourceConfig struct {
	BaseURL   string            `json:"baseUrl"`
	Headers   map[string]string `json:"headers"`
	Params    map[string]string `json:"params"`
	Endpoints DatasourceEndpoints `json:"endpoints"`
	Mapping   DatasourceMapping   `json:"mapping"`
}

// DatasourceEndpoints 数据源端点配置
type DatasourceEndpoints struct {
	Quote  string `json:"quote"`  // 实时行情: /quote?code={code}
	KLine  string `json:"kline"`  // K线数据: /kline?code={code}&period={period}
	Minute string `json:"minute"` // 分时数据: /minute?code={code}
}

// DatasourceMapping 数据源字段映射
type DatasourceMapping struct {
	// 实时行情字段映射
	Price        string `json:"price"`        // 当前价格
	Change       string `json:"change"`       // 涨跌额
	ChangePercent string `json:"changePercent"` // 涨跌幅
	Volume       string `json:"volume"`       // 成交量
	Amount       string `json:"amount"`       // 成交额
	High         string `json:"high"`         // 最高价
	Low          string `json:"low"`          // 最低价
	Open         string `json:"open"`         // 开盘价
	PreClose     string `json:"preClose"`     // 昨收价
	Name         string `json:"name"`         // 股票名称
	// K线字段映射
	KLineTime   string `json:"klineTime"`   // 时间
	KLineOpen   string `json:"klineOpen"`   // 开盘
	KLineClose  string `json:"klineClose"`  // 收盘
	KLineHigh   string `json:"klineHigh"`   // 最高
	KLineLow    string `json:"klineLow"`    // 最低
	KLineVolume string `json:"klineVolume"` // 成交量
}

// DatasourceResult 数据源返回结果
type DatasourceResult struct {
	Price        float64 `json:"price"`
	Change       float64 `json:"change"`
	ChangePercent float64 `json:"changePercent"`
	Volume       float64 `json:"volume"`
	Amount       float64 `json:"amount"`
	High         float64 `json:"high"`
	Low          float64 `json:"low"`
	Open         float64 `json:"open"`
	PreClose     float64 `json:"preClose"`
	Name         string  `json:"name"`
}

// AIConfig AI模型插件配置
type AIConfig struct {
	Provider     string            `json:"provider"` // openai-compatible, custom
	BaseURL      string            `json:"baseUrl"`
	APIKey       string            `json:"apiKey"`
	Model        string            `json:"model"`
	MaxTokens    int               `json:"maxTokens"`
	Temperature  float64           `json:"temperature"`
	SystemPrompt string            `json:"systemPrompt"`
	Headers      map[string]string `json:"headers"`
}

// AIChatMessage AI聊天消息
type AIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIChatRequest AI聊天请求
type AIChatRequest struct {
	Model       string          `json:"model"`
	Messages    []AIChatMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

// AIChatResponse AI聊天响应
type AIChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message AIChatMessage `json:"message"`
		Delta   AIChatMessage `json:"delta"`
	} `json:"choices"`
}

// NotificationData 通知数据
type NotificationData struct {
	StockCode    string  `json:"stockCode"`
	StockName    string  `json:"stockName"`
	AlertType    string  `json:"alertType"`
	CurrentPrice float64 `json:"currentPrice"`
	Condition    string  `json:"condition"`
	TargetValue  float64 `json:"targetValue"`
	TriggerTime  string  `json:"triggerTime"`
	Change       float64 `json:"change"`
	ChangePercent float64 `json:"changePercent"`
}

// PluginConfig 插件配置文件结构
type PluginConfig struct {
	Version string        `json:"version"`
	Plugins []PluginEntry `json:"plugins"`
}

// PluginEntry 插件条目
type PluginEntry struct {
	ID      string     `json:"id"`
	Type    PluginType `json:"type"`
	Path    string     `json:"path"`
	Enabled bool       `json:"enabled"`
}

// NotificationTemplate 预置通知模板
type NotificationTemplate struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Config      NotificationConfig `json:"config"`
}

// 预置通知模板
var NotificationTemplates = []NotificationTemplate{
	{
		ID:          "dingtalk",
		Name:        "钉钉机器人",
		Description: "通过钉钉群机器人发送通知",
		Config: NotificationConfig{
			URL:    "",
			Method: "POST",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			ContentType: "json",
			BodyTemplate: map[string]interface{}{
				"msgtype": "markdown",
				"markdown": map[string]string{
					"title": "股票提醒",
					"text":  "### {stockName}({stockCode})\n\n**{alertType}**\n\n- 当前价格: {currentPrice}\n- 触发条件: {condition} {targetValue}\n- 触发时间: {triggerTime}",
				},
			},
		},
	},
	{
		ID:          "wechat",
		Name:        "企业微信机器人",
		Description: "通过企业微信群机器人发送通知",
		Config: NotificationConfig{
			URL:    "",
			Method: "POST",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			ContentType: "json",
			BodyTemplate: map[string]interface{}{
				"msgtype": "markdown",
				"markdown": map[string]string{
					"content": "### 股票提醒\n**{stockName}**({stockCode})\n> {alertType}\n> 当前价格: <font color=\"warning\">{currentPrice}</font>\n> 触发条件: {condition} {targetValue}",
				},
			},
		},
	},
	{
		ID:          "feishu",
		Name:        "飞书机器人",
		Description: "通过飞书群机器人发送通知",
		Config: NotificationConfig{
			URL:    "",
			Method: "POST",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			ContentType: "json",
			BodyTemplate: map[string]interface{}{
				"msg_type": "interactive",
				"card": map[string]interface{}{
					"header": map[string]interface{}{
						"title": map[string]string{
							"tag":     "plain_text",
							"content": "股票提醒: {stockName}",
						},
						"template": "orange",
					},
					"elements": []map[string]interface{}{
						{
							"tag": "div",
							"text": map[string]string{
								"tag":     "lark_md",
								"content": "**{alertType}**\n当前价格: {currentPrice}\n触发条件: {condition} {targetValue}\n触发时间: {triggerTime}",
							},
						},
					},
				},
			},
		},
	},
	{
		ID:          "bark",
		Name:        "Bark推送",
		Description: "通过Bark发送iOS推送通知",
		Config: NotificationConfig{
			URL:    "https://api.day.app/{deviceKey}/{title}/{body}",
			Method: "GET",
			Params: map[string]string{
				"deviceKey": "",
				"title":     "股票提醒: {stockName}",
				"body":      "{alertType} - 当前价格: {currentPrice}",
			},
			ContentType: "text",
		},
	},
	{
		ID:          "serverchan",
		Name:        "Server酱",
		Description: "通过Server酱发送微信通知",
		Config: NotificationConfig{
			URL:    "https://sctapi.ftqq.com/{sendKey}.send",
			Method: "POST",
			Headers: map[string]string{
				"Content-Type": "application/x-www-form-urlencoded",
			},
			Params: map[string]string{
				"sendKey": "",
			},
			ContentType:  "form",
			BodyTemplate: "title=股票提醒: {stockName}&desp={alertType}，当前价格: {currentPrice}，触发条件: {condition} {targetValue}",
		},
	},
	{
		ID:          "pushplus",
		Name:        "PushPlus",
		Description: "通过PushPlus发送微信通知",
		Config: NotificationConfig{
			URL:    "http://www.pushplus.plus/send",
			Method: "POST",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Params: map[string]string{
				"token": "",
			},
			ContentType: "json",
			BodyTemplate: map[string]interface{}{
				"token":    "{token}",
				"title":    "股票提醒: {stockName}",
				"content":  "<h3>{stockName}({stockCode})</h3><p><b>{alertType}</b></p><p>当前价格: {currentPrice}</p><p>触发条件: {condition} {targetValue}</p><p>触发时间: {triggerTime}</p>",
				"template": "html",
			},
		},
	},
	{
		ID:          "webhook",
		Name:        "自定义Webhook",
		Description: "发送到自定义HTTP接口",
		Config: NotificationConfig{
			URL:    "",
			Method: "POST",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			ContentType: "json",
			BodyTemplate: map[string]interface{}{
				"stockCode":    "{stockCode}",
				"stockName":    "{stockName}",
				"alertType":    "{alertType}",
				"currentPrice": "{currentPrice}",
				"condition":    "{condition}",
				"targetValue":  "{targetValue}",
				"triggerTime":  "{triggerTime}",
			},
		},
	},
}
