package stocksdk

import (
	"testing"
)

func TestParseTencentResponse(t *testing.T) {
	// 模拟腾讯财经响应
	mockResponse := `v_sh600000="1~浦发银行~600000~10.50~10.40~10.45~1000000~500000~500000~10.49~1000~10.48~2000~10.47~3000~10.46~4000~10.45~5000~10.51~1000~10.52~2000~10.53~3000~10.54~4000~10.55~5000~~20240115150000~0.10~0.96~10.60~10.30~10500000~105000000~1.5~8.5~2.1~2000000000~3000000000~1.2~11.0~10.0~1.1~~1.3~1.4~12.0~9.0~1500000000~2000000000";
v_sz000001="0~平安银行~000001~12.50~12.40~12.45~2000000~1000000~1000000~12.49~2000~12.48~3000~12.47~4000~12.46~5000~12.45~6000~12.51~2000~12.52~3000~12.53~4000~12.54~5000~12.55~6000~~20240115150000~0.10~0.81~12.60~12.30~25000000~312500000~2.0~10.5~2.5~2500000000~3500000000~1.5~13.0~12.0~1.2~~1.4~1.6~14.0~11.0~2000000000~2800000000"`

	items := ParseTencentResponse(mockResponse)
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}

	if items[0].Key != "sh600000" {
		t.Errorf("Expected key sh600000, got %s", items[0].Key)
	}

	if items[1].Key != "sz000001" {
		t.Errorf("Expected key sz000001, got %s", items[1].Key)
	}
}

func TestParseSimpleQuote(t *testing.T) {
	// 模拟简要行情字段
	fields := []string{
		"1", "浦发银行", "600000", "10.50", "0.10", "0.96", "1000000", "10500000", "", "3000000000", "stock",
	}

	quote := ParseSimpleQuote(fields)
	if quote.Name != "浦发银行" {
		t.Errorf("Expected name 浦发银行, got %s", quote.Name)
	}
	if quote.Code != "600000" {
		t.Errorf("Expected code 600000, got %s", quote.Code)
	}
	if quote.Price != 10.50 {
		t.Errorf("Expected price 10.50, got %f", quote.Price)
	}
	if quote.Change != 0.10 {
		t.Errorf("Expected change 0.10, got %f", quote.Change)
	}
	if quote.ChangePct != 0.96 {
		t.Errorf("Expected change percent 0.96, got %f", quote.ChangePct)
	}
}

func TestAddMarketPrefix(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"Shanghai stock 600000", "600000", "sh600000"},
		{"Shenzhen stock 000001", "000001", "sz000001"},
		{"Shenzhen GEM 300750", "300750", "sz300750"},
		{"Already has prefix", "sh600000", "sh600000"},
		{"Uppercase prefix", "SH600000", "sh600000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddMarketPrefix(tt.code)
			if result != tt.expected {
				t.Errorf("AddMarketPrefix(%s) = %s, want %s", tt.code, result, tt.expected)
			}
		})
	}
}

func TestNormalizeTSCode(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"Shanghai stock 600000", "600000", "600000.SH"},
		{"Shenzhen stock 000001", "000001", "000001.SZ"},
		{"With sz prefix", "sz000001", "000001.SZ"},
		{"With SZ prefix", "SZ000001", "000001.SZ"},
		{"Already TS format", "600000.SH", "600000.SH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeTSCode(tt.code)
			if result != tt.expected {
				t.Errorf("NormalizeTSCode(%s) = %s, want %s", tt.code, result, tt.expected)
			}
		})
	}
}

func TestSafeNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"Empty string", "", 0},
		{"Dash", "-", 0},
		{"Normal number", "123.45", 123.45},
		{"With comma", "1,234.56", 1234.56},
		{"Invalid", "abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeNumber(tt.input)
			if result != tt.expected {
				t.Errorf("safeNumber(%s) = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}
