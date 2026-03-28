// Package main 演示「全球资产」扩展接口（pkg/realtimedata）。
//
// 涵盖：
//   - GetPopularUSStocks / GetPopularHKStocks
//   - GetGlobalIndices
//   - GetGlobalNews（region 可按库内协议传参，如 us / eu）
//   - GetFuturesList / GetFuturesPrices / GetFuturesKLine
//   - GetCryptoList / GetCryptoPrices / GetCryptoKLine
//   - GetForexRates
//
// 运行：go run ./demo/global_assets/
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

	fmt.Println("=== GetPopularUSStocks（前 5）===")
	us, err := client.GetPopularUSStocks(ctx)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for i := 0; i < 5 && i < len(us); i++ {
			x := us[i]
			fmt.Printf("  %s %s 价=%.2f\n", x.Symbol, x.Name, x.Price)
		}
	}

	fmt.Println("\n=== GetPopularHKStocks（前 5）===")
	hk, err := client.GetPopularHKStocks(ctx)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for i := 0; i < 5 && i < len(hk); i++ {
			x := hk[i]
			fmt.Printf("  %s %s 价=%.2f\n", x.Code, x.Name, x.Price)
		}
	}

	fmt.Println("\n=== GetGlobalIndices（前 8）===")
	gi, err := client.GetGlobalIndices(ctx)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for i := 0; i < 8 && i < len(gi); i++ {
			x := gi[i]
			fmt.Printf("  %s %s %.2f (%.2f%%)\n", x.Code, x.Name, x.Price, x.ChangePct)
		}
	}

	fmt.Println("\n=== GetGlobalNews（region 示例: us，2 条）===")
	gn, err := client.GetGlobalNews(ctx, "us", 2)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for _, n := range gn {
			fmt.Printf("  [%s] %s\n", n.Region, n.Title)
		}
	}

	fmt.Println("\n=== 期货：列表抽样 + 价格 + K 线 ===")
	fl, err := client.GetFuturesList(ctx)
	if err != nil {
		log.Printf("GetFuturesList: %v", err)
	} else if len(fl) > 0 {
		sym := fl[0].Symbol
		fmt.Printf("  示例合约: %s %s\n", sym, fl[0].Name)
		fp, err := client.GetFuturesPrices(ctx, []string{sym})
		if err != nil {
			log.Printf("GetFuturesPrices: %v", err)
		} else if len(fp) > 0 {
			fmt.Printf("  现价=%.2f\n", fp[0].Price)
		}
		fk, err := client.GetFuturesKLine(ctx, sym, "daily", "20240101", "20240201")
		if err != nil {
			log.Printf("GetFuturesKLine: %v", err)
		} else {
			fmt.Printf("  K 线条数=%d\n", len(fk))
		}
	}

	fmt.Println("\n=== 加密货币：列表抽样 + 行情 + K 线 ===")
	cl, err := client.GetCryptoList(ctx)
	if err != nil {
		log.Printf("GetCryptoList: %v", err)
	} else if len(cl) > 0 {
		sym := cl[0].Symbol
		fmt.Printf("  示例: %s\n", sym)
		cp, err := client.GetCryptoPrices(ctx, []string{sym})
		if err != nil {
			log.Printf("GetCryptoPrices: %v", err)
		} else if len(cp) > 0 {
			fmt.Printf("  价=%.4f\n", cp[0].Price)
		}
		ck, err := client.GetCryptoKLine(ctx, sym, "daily", "20240101", "20240201")
		if err != nil {
			log.Printf("GetCryptoKLine: %v", err)
		} else {
			fmt.Printf("  K 线条数=%d\n", len(ck))
		}
	}

	fmt.Println("\n=== GetForexRates（前 6）===")
	fx, err := client.GetForexRates(ctx)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for i := 0; i < 6 && i < len(fx); i++ {
			r := fx[i]
			fmt.Printf("  %s %s 汇率=%.6f 涨跌=%.4f\n", r.Pair, r.Name, r.Rate, r.Change)
		}
	}
}
