// Package main 演示「特色数据」：龙虎榜、停牌、分红（pkg/realtimedata）。
//
// 运行：go run ./demo/market_special/
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/easyspace-ai/tusharedb-go/pkg/realtimedata"
)

func main() {
	cfg := realtimedata.Config{DataDir: "./demo_data", EnableStorage: false, CacheMode: realtimedata.CacheModeDisabled}
	client, err := realtimedata.NewClient(cfg)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// 龙虎榜日期：建议传最近交易日 YYYYMMDD；空字符串由服务端默认处理时可能失败，此处用固定示例日并允许错误。
	date := time.Now().AddDate(0, 0, -3).Format("20060102")

	fmt.Printf("=== GetDragonTigerList（日期=%s，前 5 条）===\n", date)
	list, err := client.GetDragonTigerList(ctx, date)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for i := 0; i < 5 && i < len(list); i++ {
			d := list[i]
			fmt.Printf("  %s %s 净额=%.0f\n", d.Code, d.Name, d.NetAmount)
		}
	}

	fmt.Println("\n=== GetStockDragonTiger（600519，近期席位）===")
	dt, err := client.GetStockDragonTiger(ctx, "600519.SH")
	if err != nil {
		log.Printf("失败: %v", err)
	} else if len(dt) > 0 {
		d := dt[0]
		fmt.Printf("  %s %s 原因=%s 买方席位=%d\n", d.Code, d.Name, d.Reason, len(d.BuyList))
	}

	fmt.Println("\n=== GetSuspendedStocks（前 5 条）===")
	sus, err := client.GetSuspendedStocks(ctx)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for i := 0; i < 5 && i < len(sus); i++ {
			s := sus[i]
			fmt.Printf("  %s %s 复牌相关: %s\n", s.Code, s.Name, s.Reason)
		}
	}

	fmt.Println("\n=== GetDividendInfo（600519）===")
	divs, err := client.GetDividendInfo(ctx, "600519.SH")
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for i, d := range divs {
			if i >= 3 {
				break
			}
			fmt.Printf("  除权日=%s 派息=%.4f 送转=%.2f\n", d.ExDate, d.Dividend, d.Bonus)
		}
	}
}
