// Package main 演示「市场总览 / 标的池」类 API（pkg/realtimedata）。
//
// 涵盖：
//   - GetMarketOverview：大盘概览
//   - GetIndexList：指数列表与点位
//   - GetStockList：A 股基础列表（体量可能较大，只做条数统计 + 抽样）
//   - GetSectorList / GetIndustryList：板块、行业
//   - GetTopGainers / GetTopLosers：涨跌榜
//
// 运行：go run ./demo/market_universe/
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/easyspace-ai/stock_api/pkg/realtimedata"
)

func main() {
	cfg := realtimedata.Config{DataDir: "./demo_data", EnableStorage: false, CacheMode: realtimedata.CacheModeDisabled}
	client, err := realtimedata.NewClient(cfg)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	fmt.Println("=== GetMarketOverview ===")
	ov, err := client.GetMarketOverview(ctx)
	if err != nil {
		log.Printf("失败: %v", err)
	} else if ov != nil {
		fmt.Printf("  上证=%.2f (%.2f%%) 深成=%.2f (%.2f%%) 涨/跌家数=%d/%d 时间=%s\n",
			ov.SHIndex, ov.SHChangePct, ov.SZIndex, ov.SZChangePct, ov.RiseCount, ov.FallCount, ov.Time)
	}

	fmt.Println("\n=== GetIndexList（前 8 条抽样）===")
	idx, err := client.GetIndexList(ctx)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		n := 8
		if len(idx) < n {
			n = len(idx)
		}
		for i := 0; i < n; i++ {
			x := idx[i]
			fmt.Printf("  %s %s 最新=%.2f\n", x.Code, x.Name, x.Price)
		}
		if len(idx) > n {
			fmt.Printf("  ... 共 %d 条\n", len(idx))
		}
	}

	fmt.Println("\n=== GetStockList（条数 + 前 5 条抽样）===")
	stocks, err := client.GetStockList(ctx)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		fmt.Printf("  总条数: %d\n", len(stocks))
		for i := 0; i < 5 && i < len(stocks); i++ {
			s := stocks[i]
			fmt.Printf("  %s %s [%s]\n", s.Code, s.Name, s.Market)
		}
	}

	fmt.Println("\n=== GetSectorList（前 5 条）===")
	sectors, err := client.GetSectorList(ctx)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for i := 0; i < 5 && i < len(sectors); i++ {
			s := sectors[i]
			fmt.Printf("  %s %s 涨跌幅=%.2f%%\n", s.Code, s.Name, s.ChangePct)
		}
	}

	fmt.Println("\n=== GetIndustryList（前 5 条）===")
	inds, err := client.GetIndustryList(ctx)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for i := 0; i < 5 && i < len(inds); i++ {
			s := inds[i]
			fmt.Printf("  %s %s PE=%.2f\n", s.Code, s.Name, s.Pe)
		}
	}

	fmt.Println("\n=== GetTopGainers / GetTopLosers（各 5 条）===")
	gain, err := client.GetTopGainers(ctx, 5)
	if err != nil {
		log.Printf("涨幅榜: %v", err)
	} else {
		for _, q := range gain {
			fmt.Printf("  涨 %s %.2f%%\n", q.Name, q.ChangePct)
		}
	}
	lose, err := client.GetTopLosers(ctx, 5)
	if err != nil {
		log.Printf("跌幅榜: %v", err)
	} else {
		for _, q := range lose {
			fmt.Printf("  跌 %s %.2f%%\n", q.Name, q.ChangePct)
		}
	}
}
