package main

import (
	"context"
	"fmt"

	"github.com/easyspace-ai/tusharedb-go/internal/provider"
	"github.com/easyspace-ai/tusharedb-go/internal/provider/stocksdk"
	"github.com/easyspace-ai/tusharedb-go/pkg/tsdb"
)

func main() {
	fmt.Println("=== StockSDK Provider Test ===\n")

	// 示例 1: 基本配置 - 使用 StockSDK 作为主数据源
	fmt.Println("1. 基本配置 - 使用 StockSDK 作为主数据源")
	fmt.Println("--------------------------------------------------")
	client, err := tsdb.NewClient(tsdb.Config{
		PrimaryDataSource: tsdb.DataSourceStockSDK,
		StockSDKAPIKey:    "your-stocksdk-api-key", // 替换为实际的 API Key
		DataDir:           "./data",
	})
	if err != nil {
		fmt.Printf("创建 StockSDK 客户端失败: %v\n", err)
	} else {
		fmt.Println("✓ StockSDK 客户端创建成功")
		client.Close()
	}
	fmt.Println()

	// 示例 2: 使用自定义 Provider 工厂
	fmt.Println("2. 使用自定义 Provider 工厂")
	fmt.Println("--------------------------------------------------")
	customFactories := map[tsdb.DataSourceType]tsdb.ProviderFactory{
		tsdb.DataSourceTushare: func(cfg tsdb.DataSourceConfig) (provider.DataProvider, error) {
			fmt.Println("  → 使用自定义 Tushare Provider 工厂")
			return nil, fmt.Errorf("custom tushare not implemented")
		},
		tsdb.DataSourceStockSDK: func(cfg tsdb.DataSourceConfig) (provider.DataProvider, error) {
			fmt.Println("  → 使用自定义 StockSDK Provider 工厂")
			return nil, fmt.Errorf("custom stocksdk not implemented")
		},
	}

	// 测试工厂调用
	if factory, ok := customFactories[tsdb.DataSourceStockSDK]; ok {
		_, err := factory(tsdb.DataSourceConfig{
			Type:           tsdb.DataSourceStockSDK,
			StockSDKAPIKey: "test-key",
		})
		fmt.Printf("  自定义工厂调用结果: %v\n", err)
	}
	fmt.Println()

	// 示例 3: 主备数据源配置
	fmt.Println("3. 主备数据源配置")
	fmt.Println("--------------------------------------------------")
	fmt.Println("  配置说明:")
	fmt.Println("  - PrimaryDataSource: StockSDK (主)")
	fmt.Println("  - FallbackDataSource: Tushare (备)")
	fmt.Println()

	// 示例配置结构
	config := tsdb.Config{
		PrimaryDataSource:  tsdb.DataSourceStockSDK,
		FallbackDataSource: tsdb.DataSourceTushare,
		StockSDKAPIKey:     "stocksdk-key",
		TushareToken:       "tushare-token",
		DataDir:            "./data",
	}

	fmt.Printf("  配置详情:\n")
	fmt.Printf("  - Primary: %s\n", config.PrimaryDataSource)
	fmt.Printf("  - Fallback: %s\n", config.FallbackDataSource)
	fmt.Printf("  - Has StockSDK Key: %s\n", maskKey(config.StockSDKAPIKey))
	fmt.Printf("  - Has Tushare Token: %s\n", maskKey(config.TushareToken))
	fmt.Println()

	// 示例 4: Provider 接口方法测试
	fmt.Println("4. Provider 接口方法测试")
	fmt.Println("--------------------------------------------------")

	// 创建一个真实的 StockSDK Provider 来测试接口
	testProvider := stocksdk.NewClient(stocksdk.Config{
		APIKey: "test-key",
	})

	fmt.Printf("✓ Provider Name: %s\n", testProvider.Name())

	// 测试健康检查（预期返回未实现错误）
	ctx := context.Background()
	err = testProvider.HealthCheck(ctx)
	fmt.Printf("✓ HealthCheck: %v\n", err)

	// 测试其他接口方法
	_, err = testProvider.FetchTradeCalendar(ctx, "20240101", "20241231")
	fmt.Printf("✓ FetchTradeCalendar: %v\n", err)

	_, err = testProvider.FetchStockBasic(ctx, "L")
	fmt.Printf("✓ FetchStockBasic: %v\n", err)

	_, err = testProvider.FetchDaily(ctx, "20241231")
	fmt.Printf("✓ FetchDaily: %v\n", err)

	_, err = testProvider.FetchAdjFactor(ctx, "20241231")
	fmt.Printf("✓ FetchAdjFactor: %v\n", err)

	_, err = testProvider.FetchDailyBasic(ctx, "20241231")
	fmt.Printf("✓ FetchDailyBasic: %v\n", err)
	fmt.Println()

	// 示例 5: 数据源类型比较
	fmt.Println("5. 数据源类型比较")
	fmt.Println("--------------------------------------------------")
	dataSources := []tsdb.DataSourceType{
		tsdb.DataSourceTushare,
		tsdb.DataSourceStockSDK,
	}

	for _, ds := range dataSources {
		fmt.Printf("  - %s\n", ds)
	}
	fmt.Println()

	fmt.Println("✅ StockSDK Provider 测试完成!")
	fmt.Println()
	fmt.Println("提示: StockSDK Provider 的具体实现需要根据")
	fmt.Println("      实际的 StockSDK API 文档进行开发。")
}

// maskKey 用于在输出中隐藏敏感信息
func maskKey(key string) string {
	if key == "" {
		return "no"
	}
	if len(key) <= 8 {
		return "yes (****)"
	}
	return fmt.Sprintf("yes (%s****)", key[:4])
}
