package stocksdk

import (
	"math"
	"strconv"
	"strings"
)

// ============ 数值解析工具 ============

// safeNumber 安全解析字符串为 float64，失败返回 0
func safeNumber(s string) float64 {
	if s == "" || s == "-" {
		return 0
	}
	// 移除可能的逗号千位分隔符
	s = strings.ReplaceAll(s, ",", "")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// safeNumberOrNull 安全解析字符串为 *float64，失败返回 nil
func safeNumberOrNull(s string) *float64 {
	if s == "" || s == "-" {
		return nil
	}
	// 移除可能的逗号千位分隔符
	s = strings.ReplaceAll(s, ",", "")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	// 检查是否为 NaN 或无穷大
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil
	}
	return &f
}

// safeInt 安全解析字符串为 int，失败返回 0
func safeInt(s string) int {
	if s == "" || s == "-" {
		return 0
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// safeIntOrNull 安全解析字符串为 *int，失败返回 nil
func safeIntOrNull(s string) *int {
	if s == "" || s == "-" {
		return nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &i
}

// ============ 股票代码工具 ============

// AddMarketPrefix 给股票代码添加市场前缀
// 6位数字代码：0/3 开头加 sz，6/9 开头加 sh
func AddMarketPrefix(code string) string {
	// 检查是否已有前缀
	if strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "sz") ||
		strings.HasPrefix(code, "SH") || strings.HasPrefix(code, "SZ") {
		return strings.ToLower(code)
	}
	if len(code) != 6 {
		return code
	}
	// 0/3 开头是深圳
	if code[0] == '0' || code[0] == '3' {
		return "sz" + code
	}
	// 6/9 开头是上海
	if code[0] == '6' || code[0] == '9' {
		return "sh" + code
	}
	return code
}

// RemoveMarketPrefix 移除股票代码的市场前缀
func RemoveMarketPrefix(code string) string {
	code = strings.ToLower(code)
	// 处理点号分隔格式 (000001.sz)
	if idx := strings.Index(code, "."); idx > 0 {
		return code[:idx]
	}
	// 处理无分隔符前缀格式 (sz000001)
	if strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "sz") || strings.HasPrefix(code, "bj") {
		return code[2:]
	}
	return code
}

// NormalizeTSCode 标准化 Tushare 格式代码 (000001.SZ)
func NormalizeTSCode(code string) string {
	code = strings.ToUpper(code)
	if strings.Contains(code, ".") {
		return code
	}
	if len(code) < 6 {
		return code
	}
	// 检查是否有前缀
	if strings.HasPrefix(code, "SH") || strings.HasPrefix(code, "SZ") {
		return code[2:] + "." + code[:2]
	}
	// 0/3 开头是深圳
	if code[0] == '0' || code[0] == '3' {
		return code + ".SZ"
	}
	// 6/9 开头是上海
	if code[0] == '6' || code[0] == '9' {
		return code + ".SH"
	}
	return code
}

// ============ 日期工具 ============

// NormalizeDate 标准化日期格式为 YYYY-MM-DD
func NormalizeDate(date string) string {
	if date == "" {
		return ""
	}
	// 如果已经是 YYYY-MM-DD 格式
	if len(date) == 10 && date[4] == '-' && date[7] == '-' {
		return date
	}
	// 如果是 YYYYMMDD 格式
	if len(date) == 8 {
		return date[:4] + "-" + date[4:6] + "-" + date[6:8]
	}
	return date
}

// CompactDate 压缩日期格式为 YYYYMMDD
func CompactDate(date string) string {
	if date == "" {
		return ""
	}
	// 如果已经是 YYYYMMDD 格式
	if len(date) == 8 {
		return date
	}
	// 如果是 YYYY-MM-DD 格式
	if len(date) == 10 && date[4] == '-' && date[7] == '-' {
		return date[:4] + date[5:7] + date[8:10]
	}
	return date
}

// ============ 字符串工具 ============

// safeString 安全获取字符串，空字符串返回默认值
func safeString(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// getField 安全获取数组元素，越界返回空字符串
func getField(fields []string, index int) string {
	if index < 0 || index >= len(fields) {
		return ""
	}
	return fields[index]
}

// splitCSVLine 分割CSV行
func splitCSVLine(line string) []string {
	return strings.Split(line, ",")
}

// parseFloat 解析字符串为float64
func parseFloat(s string) float64 {
	return safeNumber(s)
}
