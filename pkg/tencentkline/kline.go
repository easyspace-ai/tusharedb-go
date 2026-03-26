package tencentkline

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Config controls HTTP behavior for Tencent K-line fetches.
type Config struct {
	Timeout time.Duration
}

// KLineItem is one bar from Tencent fqkline API.
type KLineItem struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Volume int64   `json:"volume"`
	Amount float64 `json:"amount"`
	Change float64 `json:"change"`
}

// KLineData is code + name + series.
type KLineData struct {
	Code string      `json:"code"`
	Name string      `json:"name"`
	List []KLineItem `json:"list"`
}

// FetchCommonKLine loads HK/US/CN style K-line from web.ifzq.gtimg.cn (qfq).
func FetchCommonKLine(cfg Config, code string, kLineType string, days int) (*KLineData, error) {
	normalizedCode := normalizeCode(code)
	if normalizedCode == "" {
		return nil, fmt.Errorf("invalid common kline code: %s", code)
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}

	period := normalizeKlineType(kLineType)
	url := fmt.Sprintf("https://web.ifzq.gtimg.cn/appstock/app/fqkline/get?param=%s,%s,,,%d,qfq", normalizedCode, period, days)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Host", "web.ifzq.gtimg.cn")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://gu.qq.com/")

	client := &http.Client{Timeout: cfg.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	codeValue, ok := payload["code"].(float64)
	if ok && int(codeValue) != 0 {
		return nil, fmt.Errorf("tencent kline api error: %v", payload["code"])
	}

	dataMap, _ := payload["data"].(map[string]any)
	quoteMap, _ := dataMap[normalizedCode].(map[string]any)
	if len(quoteMap) == 0 {
		return &KLineData{Code: code, Name: code, List: []KLineItem{}}, nil
	}

	var rows []any
	if qfqDay, ok := quoteMap["qfqday"].([]any); ok && len(qfqDay) > 0 {
		rows = qfqDay
	} else if day, ok := quoteMap["day"].([]any); ok && len(day) > 0 {
		rows = day
	}

	result := KLineData{
		Code: code,
		Name: firstNonEmpty(
			toString(quoteMap["qt"]),
			toString(quoteMap["name"]),
			code,
		),
		List: make([]KLineItem, 0, len(rows)),
	}

	var previousClose float64
	for _, row := range rows {
		values, ok := row.([]any)
		if !ok || len(values) < 6 {
			continue
		}
		open := toFloat(values[1])
		closePrice := toFloat(values[2])
		high := toFloat(values[3])
		low := toFloat(values[4])
		volume := int64(toFloat(values[5]))
		change := 0.0
		if previousClose != 0 {
			change = closePrice - previousClose
		}
		previousClose = closePrice

		result.List = append(result.List, KLineItem{
			Date:   toString(values[0]),
			Open:   open,
			Close:  closePrice,
			High:   high,
			Low:    low,
			Volume: volume,
			Amount: 0,
			Change: change,
		})
	}
	return &result, nil
}

func normalizeKlineType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "day", "d", "101", "":
		return "day"
	case "week", "w", "102":
		return "week"
	case "month", "m", "103":
		return "month"
	default:
		return "day"
	}
}

func normalizeCode(value string) string {
	code := strings.TrimSpace(value)
	if code == "" {
		return ""
	}

	lower := strings.ToLower(code)
	if strings.HasPrefix(lower, "sh") || strings.HasPrefix(lower, "sz") || strings.HasPrefix(lower, "hk") || strings.HasPrefix(lower, "us") {
		return code
	}

	upper := strings.ToUpper(code)
	if strings.Contains(upper, ".") {
		parts := strings.SplitN(upper, ".", 2)
		if len(parts) == 2 {
			switch parts[1] {
			case "SH", "SS":
				return "sh" + parts[0]
			case "SZ":
				return "sz" + parts[0]
			case "HK":
				return "hk" + parts[0]
			case "AM", "N", "OQ", "NYSE", "NASDAQ":
				return "us" + parts[0] + "." + parts[1]
			}
		}
	}

	if len(code) == 6 {
		if strings.HasPrefix(code, "6") {
			return "sh" + code
		}
		return "sz" + code
	}

	return code
}

func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	default:
		return ""
	}
}

func toFloat(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case string:
		var parsed float64
		fmt.Sscanf(v, "%f", &parsed)
		return parsed
	default:
		return 0
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
