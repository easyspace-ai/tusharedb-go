package main

import (
	"fmt"

	"github.com/easyspace-ai/stock_api/internal/provider/stocksdk"
	"github.com/easyspace-ai/stock_api/pkg/tsdb"
)

func main() {
	fmt.Println("=== StockSDK Advanced Usage Example ===\n")

	// 示例 1: 直接创建 StockSDK Provider
	fmt.Println("1. 直接创建 StockSDK Provider")
	fmt.Println("--------------------------------------------------")
	stocksdkClient := stocksdk.NewClient(stocksdk.Config{
		APIKey: "your-api-key",
	})
	fmt.Printf("✓ Provider Name: %s\n", stocksdkClient.Name())
	fmt.Println()

	// 示例 2: 通过 tsdb.Client 使用 StockSDK 作为数据源
	fmt.Println("2. 通过 tsdb.Client 使用 StockSDK")
	fmt.Println("--------------------------------------------------")
	client, err := tsdb.NewClient(tsdb.Config{
		PrimaryDataSource: tsdb.DataSourceStockSDK,
		StockSDKAPIKey:    "your-api-key",
		DataDir:           "./data",
	})
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
	} else {
		fmt.Println("✓ tsdb.Client 创建成功")
		fmt.Printf("✓ Primary Provider: %s\n", client.PrimaryProvider().Name())
		client.Close()
	}
	fmt.Println()

	// 示例 3: StockSDK Provider 扩展方法清单
	fmt.Println("3. StockSDK Provider 扩展方法清单")
	fmt.Println("--------------------------------------------------")
	fmt.Println("行情接口:")
	fmt.Println("  - GetFullQuotes(ctx, codes) - 获取 A 股 / 指数全量行情")
	fmt.Println("  - GetSimpleQuotes(ctx, codes) - 获取简要行情")
	fmt.Println()

	fmt.Println("K线接口:")
	fmt.Println("  - GetHistoryKline(ctx, symbol, options) - 获取 A 股历史 K 线")
	fmt.Println()

	fmt.Println("资金流向:")
	fmt.Println("  - GetFundFlow(ctx, codes) - 获取资金流向")
	fmt.Println()

	fmt.Println("板块接口:")
	fmt.Println("  - GetIndustryList(ctx) - 获取行业板块列表")
	fmt.Println("  - GetIndustryConstituents(ctx, symbol) - 获取行业板块成分股")
	fmt.Println()

	fmt.Println("期货接口:")
	fmt.Println("  - GetFuturesKline(ctx, symbol, options) - 获取期货历史 K 线")
	fmt.Println("  - GetGlobalFuturesSpot(ctx) - 获取全球期货实时行情")
	fmt.Println()

	fmt.Println("其他接口:")
	fmt.Println("  - Search(ctx, keyword) - 搜索股票")
	fmt.Println("  - GetAShareCodeList(ctx, simple, market) - 获取 A 股代码列表")
	fmt.Println("  - GetTradingCalendar(ctx) - 获取交易日历")
	fmt.Println()

	// 示例 4: K线周期和复权类型
	fmt.Println("4. K线周期和复权类型")
	fmt.Println("--------------------------------------------------")
	fmt.Println("K线周期:")
	fmt.Printf("  - %s: 日线\n", stocksdk.KlinePeriodDaily)
	fmt.Printf("  - %s: 周线\n", stocksdk.KlinePeriodWeekly)
	fmt.Printf("  - %s: 月线\n", stocksdk.KlinePeriodMonthly)
	fmt.Println()

	fmt.Println("复权类型:")
	fmt.Printf("  - %s: 不复权\n", stocksdk.AdjustTypeNone)
	fmt.Printf("  - %s: 前复权\n", stocksdk.AdjustTypeQFQ)
	fmt.Printf("  - %s: 后复权\n", stocksdk.AdjustTypeHFQ)
	fmt.Println()

	fmt.Println("分钟周期:")
	fmt.Printf("  - %s: 1分钟\n", stocksdk.MinutePeriod1)
	fmt.Printf("  - %s: 5分钟\n", stocksdk.MinutePeriod5)
	fmt.Printf("  - %s: 15分钟\n", stocksdk.MinutePeriod15)
	fmt.Printf("  - %s: 30分钟\n", stocksdk.MinutePeriod30)
	fmt.Printf("  - %s: 60分钟\n", stocksdk.MinutePeriod60)
	fmt.Println()

	// 示例 5: 使用建议
	fmt.Println("5. 使用建议")
	fmt.Println("--------------------------------------------------")
	fmt.Println("实现进度:")
	fmt.Println("  ✓ 接口定义完成")
	fmt.Println("  ✓ 数据类型定义完成")
	fmt.Println("  ☐ 具体 API 实现待开发")
	fmt.Println()

	fmt.Println("下一步开发:")
	fmt.Println("  1. 集成腾讯/东方财富/新浪数据源")
	fmt.Println("  2. 实现 HTTP 请求客户端")
	fmt.Println("  3. 实现数据解析逻辑")
	fmt.Println("  4. 添加错误处理和重试机制")
	fmt.Println()

	fmt.Println("✅ StockSDK 高级示例完成!")
	fmt.Println()
	fmt.Println("参考文档: docs/STOCKSDK_API.md")
}
