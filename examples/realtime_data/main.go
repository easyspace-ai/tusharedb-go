package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/easyspace-ai/stock_api/pkg/realtimedata"
)

func main() {
	fmt.Println("=== 实时数据采集示例 ===\n")

	// 创建客户端（自动模式：优先本地，缺失时下载）
	config := realtimedata.Config{
		DataDir:       "./data",
		EnableStorage: true,
		CacheMode:     realtimedata.CacheModeAuto,
	}
	client, err := realtimedata.NewClient(config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	defer client.ClearCache()

	ctx := context.Background()

	// 示例1: 获取单个股票行情
	fmt.Println("--- 示例1: 获取单个股票行情 ---")
	quote, err := client.GetQuote(ctx, "000001")
	if err != nil {
		log.Printf("获取行情失败: %v", err)
	} else {
		fmt.Printf("股票: %s (%s)\n", quote.Name, quote.Code)
		fmt.Printf("价格: %.2f (涨跌: %.2f %.2f%%)\n", quote.Price, quote.Change, quote.ChangePct)
		fmt.Printf("开盘: %.2f 最高: %.2f 最低: %.2f 昨收: %.2f\n", quote.Open, quote.High, quote.Low, quote.PrevClose)
		fmt.Printf("成交量: %.0f 成交额: %.0f\n", quote.Volume, quote.Amount)
		fmt.Printf("时间: %s\n\n", quote.Time)
	}

	// 示例2: 批量获取股票行情
	fmt.Println("--- 示例2: 批量获取股票行情 ---")
	codes := []string{"000001", "000002", "600000", "600036", "600519"}
	quotes, err := client.GetQuotes(ctx, codes)
	if err != nil {
		log.Printf("批量获取行情失败: %v", err)
	} else {
		fmt.Printf("获取到 %d 只股票行情:\n", len(quotes))
		for _, q := range quotes {
			fmt.Printf("  %s %-10s 价格: %8.2f 涨跌: %8.2f%%\n", q.Code, q.Name, q.Price, q.ChangePct)
		}
		fmt.Println()
	}

	// 示例3: 获取K线数据
	fmt.Println("--- 示例3: 获取日K线数据 ---")
	klines, err := client.GetDailyKLine(ctx, "000001", "qfq", "20240101", "20241231")
	if err != nil {
		log.Printf("获取K线失败: %v", err)
	} else {
		fmt.Printf("获取到 %d 条K线数据\n", len(klines))
		if len(klines) > 0 {
			fmt.Println("最新5条:")
			start := len(klines) - 5
			if start < 0 {
				start = 0
			}
			for i := start; i < len(klines); i++ {
				k := klines[i]
				fmt.Printf("  %s 开:%.2f 高:%.2f 低:%.2f 收:%.2f 量:%.0f\n",
					k.Date, k.Open, k.High, k.Low, k.Close, k.Volume)
			}
		}
		fmt.Println()
	}

	// 示例4: 查看限流器状态
	fmt.Println("--- 示例4: 限流器状态 ---")
	stats := client.GetRateLimiterStats("eastmoney.com")
	for k, v := range stats {
		fmt.Printf("  %s: %v\n", k, v)
	}
	fmt.Println()

	// 示例5: 纯网络模式（不使用本地存储）
	fmt.Println("--- 示例5: 纯网络模式 ---")
	realtimeConfig := realtimedata.Config{
		DataDir:       "./data",
		EnableStorage: true,
		CacheMode:     realtimedata.CacheModeDisabled,
	}
	realtimeClient, _ := realtimedata.NewClient(realtimeConfig)
	fmt.Println("实时客户端已创建（总是从网络获取）")
	fmt.Println()

	// 示例6: 离线模式（只读本地）
	fmt.Println("--- 示例6: 离线模式 ---")
	offlineConfig := realtimedata.Config{
		DataDir:       "./data",
		EnableStorage: true,
		CacheMode:     realtimedata.CacheModeReadOnly,
	}
	offlineClient, _ := realtimedata.NewClient(offlineConfig)
	fmt.Println("离线客户端已创建（只从本地读取）")
	fmt.Println()

	fmt.Println("=== 示例完成 ===")
	fmt.Println()
	fmt.Println("提示:")
	fmt.Println("  - 数据会自动保存到 ./data/realtime/ 目录")
	fmt.Println("  - 行情按 year/month/day 分区存储为 Parquet")
	fmt.Println("  - K线按股票代码存储为 Parquet")
}

// simpleDemo 简单使用示例
func simpleDemo() {
	// 最简单的方式：使用默认配置
	ctx := context.Background()
	client := realtimedata.NewDefaultClient()

	// 获取行情
	quote, _ := client.GetQuote(ctx, "600519")
	if quote != nil {
		fmt.Printf("茅台价格: %.2f\n", quote.Price)
	}

	// 获取K线
	klines, _ := client.GetDailyKLine(ctx, "600519", "qfq", "20240101", "20241231")
	fmt.Printf("茅台K线: %d 条\n", len(klines))
}

// loopDemo 循环更新示例
func loopDemo() {
	ctx := context.Background()
	client := realtimedata.NewDefaultClient()

	// 每分钟更新一次行情
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	codes := []string{"000001", "000002", "600000"}

	for range ticker.C {
		quotes, err := client.GetQuotes(ctx, codes)
		if err != nil {
			log.Printf("获取行情失败: %v", err)
			continue
		}
		for _, q := range quotes {
			log.Printf("%s: %.2f (%.2f%%)", q.Code, q.Price, q.ChangePct)
		}
	}
}
