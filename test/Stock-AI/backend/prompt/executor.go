package prompt

import (
	"fmt"
	"strings"
)

// StockData 股票数据（用于提示词变量替换）
type StockData struct {
	Code          string      `json:"code"`
	Name          string      `json:"name"`
	Price         float64     `json:"price"`
	Change        float64     `json:"change"`
	ChangePercent float64     `json:"changePercent"`
	Volume        float64     `json:"volume"`
	Amount        float64     `json:"amount"`
	High          float64     `json:"high"`
	Low           float64     `json:"low"`
	Open          float64     `json:"open"`
	PreClose      float64     `json:"preClose"`
	KLines        []KLineData `json:"klines,omitempty"`
}

// KLineData K线数据
type KLineData struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Volume float64 `json:"volume"`
}

// ScreenerResult 选股结果
type ScreenerResult struct {
	Stocks  []ScreenerStock `json:"stocks"`  // 筛选出的股票
	Summary string          `json:"summary"` // AI分析摘要
	Raw     string          `json:"raw"`     // AI原始输出
}

// ScreenerStock 选股结果中的单只股票
type ScreenerStock struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Reason string `json:"reason"` // 入选原因
	Signal string `json:"signal"` // buy/sell/hold
}

// ReviewResult 复盘结果
type ReviewResult struct {
	Summary      string        `json:"summary"`      // 总体摘要
	Performance  string        `json:"performance"`  // 表现评价
	Suggestions  []string      `json:"suggestions"`  // 操作建议
	StockReviews []StockReview `json:"stockReviews"` // 各股票复盘
	Raw          string        `json:"raw"`          // AI原始输出
}

// StockReview 单只股票的复盘
type StockReview struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Action      string  `json:"action"`      // 建议操作：hold/sell/add
	Reason      string  `json:"reason"`      // 原因
	TargetPrice float64 `json:"targetPrice"` // 目标价
}

// IndicatorResult 指标分析结果
type IndicatorResult struct {
	Signal string  `json:"signal"` // buy/sell/neutral
	Value  float64 `json:"value"`  // 数值结果
	Text   string  `json:"text"`   // 文本分析
	Raw    string  `json:"raw"`    // AI原始输出
}

// BuildPrompt 构建提示词（替换变量）
func BuildPrompt(template string, stock *StockData) string {
	prompt := template

	if stock != nil {
		prompt = strings.ReplaceAll(prompt, "{code}", stock.Code)
		prompt = strings.ReplaceAll(prompt, "{name}", stock.Name)
		prompt = strings.ReplaceAll(prompt, "{price}", fmt.Sprintf("%.2f", stock.Price))
		prompt = strings.ReplaceAll(prompt, "{change}", fmt.Sprintf("%.2f", stock.Change))
		prompt = strings.ReplaceAll(prompt, "{changePercent}", fmt.Sprintf("%.2f%%", stock.ChangePercent))
		prompt = strings.ReplaceAll(prompt, "{volume}", fmt.Sprintf("%.0f", stock.Volume))
		prompt = strings.ReplaceAll(prompt, "{amount}", fmt.Sprintf("%.0f", stock.Amount))
		prompt = strings.ReplaceAll(prompt, "{high}", fmt.Sprintf("%.2f", stock.High))
		prompt = strings.ReplaceAll(prompt, "{low}", fmt.Sprintf("%.2f", stock.Low))
		prompt = strings.ReplaceAll(prompt, "{open}", fmt.Sprintf("%.2f", stock.Open))
		prompt = strings.ReplaceAll(prompt, "{preClose}", fmt.Sprintf("%.2f", stock.PreClose))

		// K线数据
		if len(stock.KLines) > 0 {
			klineStr := FormatKLineData(stock.KLines)
			prompt = strings.ReplaceAll(prompt, "{klines}", klineStr)
		}
	}

	return prompt
}

// BuildPromptWithStockList 构建带股票列表的提示词
func BuildPromptWithStockList(template string, stocks []*StockData) string {
	prompt := template

	// 构建股票列表文本
	var stockListLines []string
	stockListLines = append(stockListLines, "代码,名称,价格,涨跌幅,成交量")
	for _, s := range stocks {
		stockListLines = append(stockListLines, fmt.Sprintf("%s,%s,%.2f,%.2f%%,%.0f",
			s.Code, s.Name, s.Price, s.ChangePercent, s.Volume))
	}
	stockListStr := strings.Join(stockListLines, "\n")

	prompt = strings.ReplaceAll(prompt, "{stockList}", stockListStr)
	prompt = strings.ReplaceAll(prompt, "{stockCount}", fmt.Sprintf("%d", len(stocks)))

	return prompt
}

// BuildPromptWithPortfolio 构建带持仓的提示词
func BuildPromptWithPortfolio(template string, positions []*PositionData) string {
	prompt := template

	// 构建持仓列表文本
	var positionLines []string
	positionLines = append(positionLines, "代码,名称,持仓数量,成本价,当前价,盈亏比例")
	totalProfit := 0.0
	for _, p := range positions {
		profitPercent := 0.0
		if p.CostPrice > 0 {
			profitPercent = (p.CurrentPrice - p.CostPrice) / p.CostPrice * 100
		}
		totalProfit += (p.CurrentPrice - p.CostPrice) * float64(p.Quantity)
		positionLines = append(positionLines, fmt.Sprintf("%s,%s,%d,%.2f,%.2f,%.2f%%",
			p.Code, p.Name, p.Quantity, p.CostPrice, p.CurrentPrice, profitPercent))
	}
	positionStr := strings.Join(positionLines, "\n")

	prompt = strings.ReplaceAll(prompt, "{portfolio}", positionStr)
	prompt = strings.ReplaceAll(prompt, "{positionCount}", fmt.Sprintf("%d", len(positions)))
	prompt = strings.ReplaceAll(prompt, "{totalProfit}", fmt.Sprintf("%.2f", totalProfit))

	return prompt
}

// PositionData 持仓数据
type PositionData struct {
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Quantity     int     `json:"quantity"`
	CostPrice    float64 `json:"costPrice"`
	CurrentPrice float64 `json:"currentPrice"`
}

// FormatKLineData 格式化K线数据为CSV文本
func FormatKLineData(klines []KLineData) string {
	var lines []string
	lines = append(lines, "日期,开盘,收盘,最高,最低,成交量")
	for _, k := range klines {
		lines = append(lines, fmt.Sprintf("%s,%.2f,%.2f,%.2f,%.2f,%.0f",
			k.Date, k.Open, k.Close, k.High, k.Low, k.Volume))
	}
	return strings.Join(lines, "\n")
}

// ParseSignal 从AI响应中解析信号
func ParseSignal(response string) string {
	responseLower := strings.ToLower(response)

	// 强信号
	if strings.Contains(responseLower, "强烈买入") || strings.Contains(responseLower, "strong buy") {
		return "strong_buy"
	}
	if strings.Contains(responseLower, "强烈卖出") || strings.Contains(responseLower, "strong sell") {
		return "strong_sell"
	}

	// 普通信号
	if strings.Contains(responseLower, "买入") || strings.Contains(responseLower, "buy") ||
		strings.Contains(responseLower, "建仓") || strings.Contains(responseLower, "加仓") ||
		strings.Contains(responseLower, "【买入】") {
		return "buy"
	}
	if strings.Contains(responseLower, "卖出") || strings.Contains(responseLower, "sell") ||
		strings.Contains(responseLower, "清仓") || strings.Contains(responseLower, "减仓") ||
		strings.Contains(responseLower, "【卖出】") {
		return "sell"
	}
	if strings.Contains(responseLower, "持有") || strings.Contains(responseLower, "hold") ||
		strings.Contains(responseLower, "观望") || strings.Contains(responseLower, "【持有】") ||
		strings.Contains(responseLower, "【观望】") {
		return "hold"
	}

	return "neutral"
}

// TruncateText 截断文本
func TruncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
