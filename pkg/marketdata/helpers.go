package marketdata

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// todayInChina is calendar date in Asia/Shanghai (A-share trading day alignment when server runs in UTC).
func todayInChina() string {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return time.Now().In(loc).Format("2006-01-02")
}

func defaultUserAgent() string {
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
}

func parseFloat(value string) float64 {
	clean := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(value, ",", ""), "%", ""))
	if clean == "" || clean == "--" {
		return 0
	}
	number, _ := strconv.ParseFloat(clean, 64)
	return number
}

func normalizeStockCode(code string) string {
	value := strings.TrimSpace(strings.ToLower(code))
	value = strings.TrimSuffix(value, ".sh")
	value = strings.TrimSuffix(value, ".sz")
	value = strings.TrimSuffix(value, ".bj")
	replacer := strings.NewReplacer("sh", "", "sz", "", "bj", "", "hk", "", "gb_", "", "us_", "", "us", "")
	return replacer.Replace(value)
}

func buildEastmoneyReportPDF(infoCode string) string {
	if strings.TrimSpace(infoCode) == "" {
		return ""
	}
	return fmt.Sprintf("https://pdf.dfcfw.com/pdf/H3_%s_1.pdf", infoCode)
}

func buildEastmoneyNoticePDF(artCode string) string {
	if strings.TrimSpace(artCode) == "" {
		return ""
	}
	return fmt.Sprintf("https://pdf.dfcfw.com/pdf/H2_%s_1.pdf", artCode)
}

func parseUnixOrNow(value string) time.Time {
	if strings.TrimSpace(value) == "" {
		return time.Now()
	}
	if number, err := strconv.ParseInt(value, 10, 64); err == nil {
		if number > 1e12 {
			return time.UnixMilli(number)
		}
		return time.Unix(number, 0)
	}
	if ts, err := time.Parse(time.DateTime, value); err == nil {
		return ts
	}
	return time.Now()
}

func joinAnyList(value any) string {
	switch data := value.(type) {
	case nil:
		return ""
	case string:
		return data
	case []any:
		parts := make([]string, 0, len(data))
		for _, item := range data {
			text := strings.TrimSpace(fmt.Sprintf("%v", item))
			if text != "" && text != "<nil>" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, ",")
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", value))
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

func paginate[T any](items []T, page, pageSize int) []T {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = len(items)
	}
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []T{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

