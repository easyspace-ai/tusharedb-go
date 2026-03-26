package main

import (
	"context"
	"fmt"
	"os"

	"github.com/easyspace-ai/tusharedb-go/internal/provider"
	"github.com/easyspace-ai/tusharedb-go/pkg/tsdb"
)

func main() {
	fmt.Println("=== Multi-DataSource Test ===\n")

	// 示例 1: 演示不同的数据源配置方式
	fmt.Println("1. 数据源配置方式")
	fmt.Println("--------------------------------------------------")

	// 方式 A: 使用默认 Tushare 数据源（向后兼容）
	fmt.Println("A. 使用默认 Tushare 数据源（向后兼容）:")
	fmt.Println(`   client, err := tsdb.NewClient(tsdb.Config{
       Token:   "your-token",
       DataDir: "./data",
   })`)
	fmt.Println()

	// 方式 B: 显式指定 Tushare 数据源
	fmt.Println("B. 显式指定 Tushare 数据源:")
	fmt.Println(`   client, err := tsdb.NewClient(tsdb.Config{
       PrimaryDataSource: tsdb.DataSourceTushare,
       TushareToken:      "your-token",
       DataDir:           "./data",
   })`)
	fmt.Println()

	// 方式 C: 使用 StockSDK 数据源
	fmt.Println("C. 使用 StockSDK 数据源:")
	fmt.Println(`   client, err := tsdb.NewClient(tsdb.Config{
       PrimaryDataSource: tsdb.DataSourceStockSDK,
       StockSDKAPIKey:    "your-api-key",
       DataDir:           "./data",
   })`)
	fmt.Println()

	// 方式 D: 使用主备数据源
	fmt.Println("D. 使用主备数据源:")
	fmt.Println(`   client, err := tsdb.NewClient(tsdb.Config{
       PrimaryDataSource:  tsdb.DataSourceStockSDK,
       FallbackDataSource: tsdb.DataSourceTushare,
       StockSDKAPIKey:     "your-api-key",
       TushareToken:       "your-token",
       DataDir:            "./data",
   })`)
	fmt.Println()

	// 示例 2: 从环境变量获取配置
	fmt.Println("2. 从环境变量获取配置")
	fmt.Println("--------------------------------------------------")
	token := os.Getenv("TUSHARE_TOKEN")
	apiKey := os.Getenv("STOCKSDK_API_KEY")

	if token != "" {
		fmt.Printf("✓ TUSHARE_TOKEN 已设置: %s\n", maskKey(token))
	} else {
		fmt.Println("✗ TUSHARE_TOKEN 未设置")
	}

	if apiKey != "" {
		fmt.Printf("✓ STOCKSDK_API_KEY 已设置: %s\n", maskKey(apiKey))
	} else {
		fmt.Println("✗ STOCKSDK_API_KEY 未设置")
	}
	fmt.Println()

	// 示例 3: 检测并使用可用的 Provider
	fmt.Println("3. 检测并使用可用的 Provider")
	fmt.Println("--------------------------------------------------")

	// 创建一个测试用的配置
	testConfigs := []struct {
		name   string
		source tsdb.DataSourceType
		token  string
	}{
		{"Tushare", tsdb.DataSourceTushare, "test-tushare-token"},
		{"StockSDK", tsdb.DataSourceStockSDK, "test-stocksdk-key"},
	}

	for _, tc := range testConfigs {
		fmt.Printf("测试 %s Provider:\n", tc.name)

		// 使用工厂创建 Provider
		if factory, ok := tsdb.DefaultProviderFactory[tc.source]; ok {
			p, err := factory(tsdb.DataSourceConfig{
				Type:           tc.source,
				TushareToken:   tc.token,
				StockSDKAPIKey: tc.token,
			})
			if err != nil {
				fmt.Printf("  ✗ 创建失败: %v\n", err)
			} else {
				fmt.Printf("  ✓ Provider 名称: %s\n", p.Name())
				testProviderMethods(p)
			}
		}
		fmt.Println()
	}

	// 示例 4: Provider 接口方法清单
	fmt.Println("4. Provider 接口方法清单")
	fmt.Println("--------------------------------------------------")
	fmt.Println("DataProvider 接口包含以下方法:")
	fmt.Println("  - Name() string")
	fmt.Println("  - HealthCheck(ctx context.Context) error")
	fmt.Println()
	fmt.Println("TradeCalendarProvider:")
	fmt.Println("  - FetchTradeCalendar(ctx, startDate, endDate string)")
	fmt.Println()
	fmt.Println("StockBasicProvider:")
	fmt.Println("  - FetchStockBasic(ctx, listStatus string)")
	fmt.Println()
	fmt.Println("DailyQuoteProvider:")
	fmt.Println("  - FetchDaily(ctx, tradeDate string)")
	fmt.Println("  - FetchDailyRange(ctx, startDate, endDate string)")
	fmt.Println()
	fmt.Println("AdjFactorProvider:")
	fmt.Println("  - FetchAdjFactor(ctx, tradeDate string)")
	fmt.Println("  - FetchAdjFactorRange(ctx, startDate, endDate string)")
	fmt.Println()
	fmt.Println("DailyBasicProvider:")
	fmt.Println("  - FetchDailyBasic(ctx, tradeDate string)")
	fmt.Println("  - FetchDailyBasicRange(ctx, startDate, endDate string)")
	fmt.Println()

	fmt.Println("✅ 多数据源测试完成!")
}

// testProviderMethods 测试 Provider 的各个接口方法
func testProviderMethods(p provider.DataProvider) {
	ctx := context.Background()

	// 测试 HealthCheck
	err := p.HealthCheck(ctx)
	fmt.Printf("  - HealthCheck: %v\n", formatError(err))

	// 测试 FetchTradeCalendar
	_, err = p.FetchTradeCalendar(ctx, "20240101", "20241231")
	fmt.Printf("  - FetchTradeCalendar: %v\n", formatError(err))

	// 测试 FetchStockBasic
	_, err = p.FetchStockBasic(ctx, "L")
	fmt.Printf("  - FetchStockBasic: %v\n", formatError(err))

	// 测试 FetchDaily
	_, err = p.FetchDaily(ctx, "20241231")
	fmt.Printf("  - FetchDaily: %v\n", formatError(err))

	// 测试 FetchAdjFactor
	_, err = p.FetchAdjFactor(ctx, "20241231")
	fmt.Printf("  - FetchAdjFactor: %v\n", formatError(err))

	// 测试 FetchDailyBasic
	_, err = p.FetchDailyBasic(ctx, "20241231")
	fmt.Printf("  - FetchDailyBasic: %v\n", formatError(err))
}

// formatError 格式化错误输出
func formatError(err error) string {
	if err == nil {
		return "✓ ok"
	}
	return fmt.Sprintf("✗ %v", err)
}

// maskKey 用于在输出中隐藏敏感信息
func maskKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "****"
	}
	return fmt.Sprintf("%s****", key[:4])
}
