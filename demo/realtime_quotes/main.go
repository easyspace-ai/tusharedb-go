// Package main 演示「实时行情」相关 API（pkg/realtimedata）。
//
// 涵盖：
//   - GetQuote / GetQuotes：多源自动 Failover
//   - GetQuotesBySource：指定数据源（东财、新浪、腾讯、雪球、百度、同花顺等）
//
// 运行：在仓库根目录执行 go run ./demo/realtime_quotes/
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/easyspace-ai/stock_api/pkg/realtimedata"
)

func main() {
	// 关闭本地 Parquet 快照写入，纯拉网测试；如需落盘可设为 true 并配置 DataDir
	cfg := realtimedata.Config{
		DataDir:       "./demo_data",
		EnableStorage: false,
		CacheMode:     realtimedata.CacheModeDisabled,
	}
	client, err := realtimedata.NewClient(cfg)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 多源逐个探测 + 限流随机等待，60s 容易在末段触发 deadline；演示场景放宽到 3 分钟
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	codes := []string{"600519.SH", "000001.SZ"}
	fmt.Println("=== 1) GetQuotes（自动 Failover，按已注册源顺序尝试）===")
	quotes, err := client.GetQuotes(ctx, codes)
	if err != nil {
		log.Printf("GetQuotes 失败: %v", err)
	} else {
		for _, q := range quotes {
			fmt.Printf("  %s %s 现价=%.2f 涨跌幅=%.2f%%\n", q.Code, q.Name, q.Price, q.ChangePct)
		}
	}

	// 与 multisource.DataSourceType 一致的别名
	sources := []realtimedata.DataSourceType{
		realtimedata.DataSourceEastMoney,
		realtimedata.DataSourceSina,
		realtimedata.DataSourceTencent,
		realtimedata.DataSourceXueqiu,
		realtimedata.DataSourceBaidu,
		realtimedata.DataSourceTonghuashun,
	}

	fmt.Println("\n=== 2) GetQuotesBySource（逐数据源；部分源可能因风控失败，属正常现象）===")
	for _, src := range sources {
		qb, err := client.GetQuotesBySource(ctx, src, []string{"600519.SH"})
		if err != nil {
			fmt.Printf("  %-12s 错误: %v\n", src, err)
			continue
		}
		if len(qb) == 0 {
			fmt.Printf("  %-12s 无数据\n", src)
			continue
		}
		q := qb[0]
		fmt.Printf("  %-12s %s 现价=%.2f\n", src, q.Name, q.Price)
	}

	fmt.Println("\n=== 3) GetQuotesBatch（_partial 成功也返回 map）===")
	ok, bad, err := client.GetQuotesBatch(ctx, codes)
	if err != nil {
		log.Printf("GetQuotesBatch 整体错误: %v", err)
	}
	for k, v := range ok {
		fmt.Printf("  成功 %s: %.2f\n", k, v.Price)
	}
	for k, msg := range bad {
		fmt.Printf("  失败 %s: %s\n", k, msg)
	}
}
