package main

import (
	"context"
	"fmt"
	"time"

	"github.com/easyspace-ai/stock_api/internal/provider/stocksdk"
)

func main() {
	fmt.Println("=== StockSDK 腾讯数据源测试 ===\n")

	// 创建 StockSDK 客户端
	client := stocksdk.NewClient(stocksdk.Config{
		Timeout: 30 * time.Second,
	})

	ctx := context.Background()

	// 测试 1: 获取简要行情
	fmt.Println("1. 测试获取简要行情")
	fmt.Println("--------------------------------------------------")
	codes := []string{"000001", "600000"}
	simpleQuotes, err := client.GetSimpleQuotes(ctx, codes)
	if err != nil {
		fmt.Printf("✗ 获取简要行情失败: %v\n", err)
	} else {
		fmt.Printf("✓ 获取到 %d 条简要行情\n", len(simpleQuotes))
		for _, quote := range simpleQuotes {
			fmt.Printf("  - %s (%s): 价格=%.2f, 涨跌=%.2f%%\n",
				quote.Name, quote.Code, quote.Price, quote.ChangePct)
		}
	}
	fmt.Println()

	// 测试 2: 获取交易日历
	fmt.Println("2. 测试获取交易日历")
	fmt.Println("--------------------------------------------------")
	calendar, err := client.GetTradingCalendar(ctx)
	if err != nil {
		fmt.Printf("✗ 获取交易日历失败: %v\n", err)
	} else {
		fmt.Printf("✓ 获取到 %d 个交易日\n", len(calendar))
		if len(calendar) > 5 {
			fmt.Printf("  前5个: %v\n", calendar[:5])
		}
	}
	fmt.Println()

	// 测试 3: 获取A股代码列表
	fmt.Println("3. 测试获取A股代码列表")
	fmt.Println("--------------------------------------------------")
	codeList, err := client.GetAShareCodeList(ctx, false, "")
	if err != nil {
		fmt.Printf("✗ 获取A股代码列表失败: %v\n", err)
	} else {
		fmt.Printf("✓ 获取到 %d 只股票\n", len(codeList))
		if len(codeList) > 5 {
			fmt.Printf("  前5个: %v\n", codeList[:5])
		}
	}
	fmt.Println()

	fmt.Println("✅ 测试完成!")
}
