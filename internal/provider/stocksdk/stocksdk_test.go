package stocksdk

import (
	"context"
	"testing"
	"time"

	"github.com/easyspace-ai/tusharedb-go/internal/provider"
)

// TestClientImplementsDataProvider 验证 Client 是否实现了 DataProvider 接口
func TestClientImplementsDataProvider(t *testing.T) {
	var _ provider.DataProvider = (*Client)(nil)
	t.Log("✓ Client implements provider.DataProvider interface")
}

// TestClientName 测试 Name 方法
func TestClientName(t *testing.T) {
	client := NewClient(Config{})
	name := client.Name()
	if name != "stocksdk" {
		t.Errorf("Expected name 'stocksdk', got '%s'", name)
	}
	t.Logf("✓ Client.Name() = '%s'", name)
}

// TestClientMethods 测试所有接口方法
func TestClientMethods(t *testing.T) {
	ctx := context.Background()
	client := NewClient(Config{APIKey: "test-key"})

	// 测试那些已经实现的方法不应该返回错误
	t.Run("FetchTradeCalendar", func(t *testing.T) {
		_, err := client.FetchTradeCalendar(ctx, "20240101", "20241231")
		if err != nil {
			t.Errorf("FetchTradeCalendar unexpected error: %v", err)
		} else {
			t.Log("✓ FetchTradeCalendar implemented")
		}
	})

	// 测试那些仍然返回未实现错误的方法
	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "FetchDaily",
			fn: func() error {
				_, err := client.FetchDaily(ctx, "20241231")
				return err
			},
		},
		{
			name: "FetchDailyRange",
			fn: func() error {
				_, err := client.FetchDailyRange(ctx, "20240101", "20241231")
				return err
			},
		},
		{
			name: "FetchAdjFactor",
			fn: func() error {
				_, err := client.FetchAdjFactor(ctx, "20241231")
				return err
			},
		},
		{
			name: "FetchAdjFactorRange",
			fn: func() error {
				_, err := client.FetchAdjFactorRange(ctx, "20240101", "20241231")
				return err
			},
		},
		{
			name: "FetchDailyBasicRange",
			fn: func() error {
				_, err := client.FetchDailyBasicRange(ctx, "20240101", "20241231")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
				return
			}
			t.Logf("✓ %s returns expected error: %v", tt.name, err)
		})
	}
}

// TestNewClient 测试 NewClient 构造函数
func TestNewClient(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "Empty config",
			cfg:  Config{},
		},
		{
			name: "With API key",
			cfg: Config{
				APIKey: "test-api-key",
			},
		},
		{
			name: "Full config",
			cfg: Config{
				BaseURL:   "https://api.example.com",
				Timeout:   60 * time.Second,
				Retries:   5,
				RetryWait: 2 * time.Second,
				UserAgent: "test-agent",
				APIKey:    "test-key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.cfg)
			if client == nil {
				t.Fatal("NewClient returned nil")
			}
			t.Logf("✓ NewClient(%s) created successfully", tt.name)
		})
	}
}

// TestParseStockListResponse 测试股票列表响应解析
func TestParseStockListResponse(t *testing.T) {
	jsonData := []byte(`{"success":true,"list":["sh600000","sz000001","sh600519"]}`)

	resp, err := ParseStockListResponse(jsonData)
	if err != nil {
		t.Fatalf("ParseStockListResponse failed: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success=true")
	}

	expectedCodes := []string{"sh600000", "sz000001", "sh600519"}
	if len(resp.List) != len(expectedCodes) {
		t.Errorf("Expected %d codes, got %d", len(expectedCodes), len(resp.List))
	}

	for i, code := range expectedCodes {
		if resp.List[i] != code {
			t.Errorf("Expected code %d to be '%s', got '%s'", i, code, resp.List[i])
		}
	}

	t.Log("✓ ParseStockListResponse works correctly")
}

// TestParseTradeCalendar 测试交易日历解析
func TestParseTradeCalendar(t *testing.T) {
	text := "2024-01-02,2024-01-03,2024-01-04,2024-01-05"

	dates := ParseTradeCalendar(text)
	expectedDates := []string{"2024-01-02", "2024-01-03", "2024-01-04", "2024-01-05"}

	if len(dates) != len(expectedDates) {
		t.Errorf("Expected %d dates, got %d", len(expectedDates), len(dates))
	}

	for i, date := range expectedDates {
		if dates[i] != date {
			t.Errorf("Expected date %d to be '%s', got '%s'", i, date, dates[i])
		}
	}

	// 测试空输入
	emptyDates := ParseTradeCalendar("")
	if len(emptyDates) != 0 {
		t.Error("Expected empty slice for empty input")
	}

	t.Log("✓ ParseTradeCalendar works correctly")
}

// TestMatchMarket 测试市场筛选
func TestMatchMarket(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		market AShareMarket
		expect bool
	}{
		{"Shanghai stock with SH market", "sh600000", AShareMarketSH, true},
		{"Shenzhen stock with SZ market", "sz000001", AShareMarketSZ, true},
		{"GEM stock with CY market", "sz300750", AShareMarketCY, true},
		{"STAR market with KC market", "sh688001", AShareMarketKC, true},
		{"Shanghai stock should not match SZ", "sh600000", AShareMarketSZ, false},
		{"Shenzhen stock should not match SH", "sz000001", AShareMarketSH, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchMarket(tt.code, tt.market)
			if result != tt.expect {
				t.Errorf("MatchMarket(%s, %s) = %v, expect %v", tt.code, tt.market, result, tt.expect)
			}
		})
	}

	t.Log("✓ MatchMarket works correctly")
}

// TestParseIndustryBoardKlineCsv 测试行业板块 K 线 CSV 解析
func TestParseIndustryBoardKlineCsv(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantDate string
		wantOpen float64
	}{
		{
			name:     "Valid line",
			line:     "2024-01-15,100.5,101.2,102.3,99.8,1000000,50000000,2.5,1.2,1.5,5.2",
			wantDate: "2024-01-15",
			wantOpen: 100.5,
		},
		{
			name:     "Empty line",
			line:     "",
			wantDate: "",
			wantOpen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseIndustryBoardKlineCsv(tt.line)
			if result.Date != tt.wantDate {
				t.Errorf("ParseIndustryBoardKlineCsv() Date = %v, want %v", result.Date, tt.wantDate)
			}
			if tt.wantOpen > 0 && (result.Open == nil || *result.Open != tt.wantOpen) {
				t.Errorf("ParseIndustryBoardKlineCsv() Open = %v, want %v", result.Open, tt.wantOpen)
			}
		})
	}
	t.Log("✓ ParseIndustryBoardKlineCsv works correctly")
}

// TestParseFuturesKlineCsv 测试期货 K 线 CSV 解析
func TestParseFuturesKlineCsv(t *testing.T) {
	line := "2024-01-15,3800.5,3820.3,3850.0,3780.2,150000,285000000,2.1,1.05,40.2,25.5,3800.0,50000"
	result, _, _ := ParseFuturesKlineCsv(line)

	if result.Date != "2024-01-15" {
		t.Errorf("Expected date '2024-01-15', got '%s'", result.Date)
	}
	if result.Open == nil || *result.Open != 3800.5 {
		t.Errorf("Expected open 3800.5, got %v", result.Open)
	}
	if result.OpenInterest == nil || *result.OpenInterest != 50000 {
		t.Errorf("Expected open interest 50000, got %v", result.OpenInterest)
	}
	t.Log("✓ ParseFuturesKlineCsv works correctly")
}

// TestExtractFuturesVariety 测试期货品种提取
func TestExtractFuturesVariety(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"rb2605", "rb"},
		{"IF2406", "IF"},
		{"au2408", "au"},
		{"RBM", "RBM"},
		{"T", "T"},
	}

	for _, tt := range tests {
		result := extractFuturesVariety(tt.input)
		if result != tt.expected {
			t.Errorf("extractFuturesVariety(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
	t.Log("✓ extractFuturesVariety works correctly")
}

// TestGetFuturesMarketCode 测试期货 market code 获取
func TestGetFuturesMarketCode(t *testing.T) {
	tests := []struct {
		variety  string
		expected int
		wantErr  bool
	}{
		{"rb", 113, false},   // SHFE
		{"IF", 220, false},   // CFFEX
		{"c", 114, false},    // DCE
		{"TA", 115, false},   // CZCE
		{"sc", 142, false},   // INE
		{"si", 225, false},   // GFEX
		{"XXX", 0, true},     // Unknown
	}

	for _, tt := range tests {
		result, err := getFuturesMarketCode(tt.variety)
		if tt.wantErr {
			if err == nil {
				t.Errorf("getFuturesMarketCode(%s) expected error, got none", tt.variety)
			}
		} else {
			if err != nil {
				t.Errorf("getFuturesMarketCode(%s) unexpected error: %v", tt.variety, err)
			}
			if result != tt.expected {
				t.Errorf("getFuturesMarketCode(%s) = %d, want %d", tt.variety, result, tt.expected)
			}
		}
	}
	t.Log("✓ getFuturesMarketCode works correctly")
}

// TestMapGlobalFuturesSpotItem 测试全球期货行情映射
func TestMapGlobalFuturesSpotItem(t *testing.T) {
	item := GlobalFuturesSpotItem{
		Dm:   "HG00Y",
		Name: "COMEX铜连续",
		P:    4.25,
		Zde:  0.05,
		Zdf:  1.19,
		O:    4.20,
		H:    4.30,
		L:    4.18,
		Zjsj: 4.20,
		Vol:  150000,
		Wp:   75000,
		Np:   75000,
		Ccl:  200000,
	}

	result := MapGlobalFuturesSpotItem(item)

	if result.Code != "HG00Y" {
		t.Errorf("Expected code 'HG00Y', got '%s'", result.Code)
	}
	if result.Name != "COMEX铜连续" {
		t.Errorf("Expected name 'COMEX铜连续', got '%s'", result.Name)
	}
	if result.Price == nil || *result.Price != 4.25 {
		t.Errorf("Expected price 4.25, got %v", result.Price)
	}
	t.Log("✓ MapGlobalFuturesSpotItem works correctly")
}

func float64Ptr(f float64) *float64 {
	return &f
}
