// Package main 演示「资金流向」相关 API（pkg/realtimedata）。
//
// 涵盖：
//   - GetMoneyFlow：个股资金流向
//   - GetSectorMoneyFlow：板块资金流排名
//   - GetMarketMoneyFlow：全市场统计
//   - GetNorthboundFlow：北向（沪深港通）净流
//
// 运行：go run ./demo/capital_flow/
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
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("=== GetMoneyFlow（贵州茅台示例）===")
	mf, err := client.GetMoneyFlow(ctx, "600519.SH")
	if err != nil {
		log.Printf("失败: %v", err)
	} else if mf != nil {
		fmt.Printf("  主力净流入=%.2f 万（单位以接口字段含义为准）\n", mf.MainNetInflow)
	}

	fmt.Println("\n=== GetSectorMoneyFlow（前 5 板块）===")
	smf, err := client.GetSectorMoneyFlow(ctx)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for i := 0; i < 5 && i < len(smf); i++ {
			s := smf[i]
			fmt.Printf("  %s 净流入=%.2f\n", s.Name, s.NetInflow)
		}
	}

	fmt.Println("\n=== GetMarketMoneyFlow ===")
	mm, err := client.GetMarketMoneyFlow(ctx)
	if err != nil {
		log.Printf("失败: %v", err)
	} else if mm != nil {
		fmt.Printf("  上证主力=%.2f 上证散户=%.2f 深证主力=%.2f 深证散户=%.2f 合计估算=%.2f 时间=%s\n",
			mm.SHMainInflow, mm.SHRetailInflow, mm.SZMainInflow, mm.SZRetailInflow, mm.TotalInflow, mm.Time)
	}

	fmt.Println("\n=== GetNorthboundFlow ===")
	nb, err := client.GetNorthboundFlow(ctx)
	if err != nil {
		log.Printf("失败: %v", err)
	} else if nb != nil {
		fmt.Printf("  沪股通净流入=%.2f 深股通净流入=%.2f 合计=%.2f 时间=%s\n",
			nb.SHNetInflow, nb.SZNetInflow, nb.Total, nb.Time)
	}
}
