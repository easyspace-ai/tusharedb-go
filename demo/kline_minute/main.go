// Package main 演示「K 线 / 分时」相关 API（pkg/realtimedata）。
//
// 涵盖：
//   - GetKLine / GetDailyKLine：多源 Failover（通常由东财、新浪、腾讯等实现）
//   - GetKLineBySource：指定数据源
//   - GetMinuteData：当前仅腾讯分时有效
//
// period 常用：daily / week / month（具体以库内映射为准）。
// adjust：qfq（前复权） hfq（后复权） none（不复权）。
//
// 运行：go run ./demo/kline_minute/
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/easyspace-ai/stock_api/pkg/realtimedata"
)

func main() {
	cfg := realtimedata.Config{
		DataDir:       "./demo_data",
		EnableStorage: false,
		CacheMode:     realtimedata.CacheModeDisabled,
	}
	client, err := realtimedata.NewClient(cfg)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	code := "600519.SH"
	start := "20240101"
	end := "20240301"

	fmt.Println("=== 1) GetDailyKLine（日线，封装 period=daily）===")
	kl, err := client.GetDailyKLine(ctx, code, "qfq", start, end)
	if err != nil {
		log.Printf("GetDailyKLine: %v", err)
	} else {
		printKPreview("日线", kl, 3)
	}

	fmt.Println("\n=== 2) GetKLine（多源 Failover）===")
	kl2, err := client.GetKLine(ctx, code, "daily", "qfq", start, end)
	if err != nil {
		log.Printf("GetKLine: %v", err)
	} else {
		printKPreview("GetKLine daily", kl2, 3)
	}

	fmt.Println("\n=== 3) GetKLineBySource（指定东财；亦可选 sina / tencent）===")
	kl3, err := client.GetKLineBySource(ctx, realtimedata.DataSourceEastMoney, code, "daily", "qfq", start, end)
	if err != nil {
		log.Printf("GetKLineBySource eastmoney: %v", err)
	} else {
		printKPreview("东财日线", kl3, 3)
	}

	fmt.Println("\n=== 4) GetMinuteData（腾讯分时；src 仅支持 tencent）===")
	bars, err := client.GetMinuteData(ctx, realtimedata.DataSourceTencent, code)
	if err != nil {
		log.Printf("GetMinuteData: %v", err)
	} else {
		n := len(bars)
		if n > 5 {
			n = 5
		}
		fmt.Printf("  共 %d 根分时柱，预览前 %d 根:\n", len(bars), n)
		for i := 0; i < n; i++ {
			b := bars[i]
			fmt.Printf("  时间=%s 价=%.3f 量=%.0f\n", b.Time, b.Price, b.Volume)
		}
	}
}

func printKPreview(label string, kl []realtimedata.KLineItem, max int) {
	if len(kl) == 0 {
		fmt.Printf("  %s 无数据\n", label)
		return
	}
	n := max
	if len(kl) < n {
		n = len(kl)
	}
	fmt.Printf("  %s 条数=%d，末尾 %d 条:\n", label, len(kl), n)
	start := len(kl) - n
	for i := start; i < len(kl); i++ {
		k := kl[i]
		fmt.Printf("  %s O=%.2f H=%.2f L=%.2f C=%.2f V=%.0f\n",
			k.Date, k.Open, k.High, k.Low, k.Close, k.Volume)
	}
}
