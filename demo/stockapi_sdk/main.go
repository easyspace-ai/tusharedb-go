// Package main 演示 pkg/stockapi：基于 stocksdk 的行情/K线/分时封装（带可选本地缓存与 Parquet 历史湖）。
//
// 注意：
//   - GetAllAShareQuotes 会拉全市场，数据量大，本示例不默认调用；需要时请自行加超时与机器内存评估。
//
// 运行：go run ./demo/stockapi_sdk/
package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/easyspace-ai/stock_api/pkg/stockapi"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	dataDir := filepath.Join("demo_data", "stockapi_cache")
	c, err := stockapi.NewClientWithConfig(stockapi.Config{
		DataDir:   dataDir,
		CacheMode: stockapi.CacheModeAuto,
	})
	if err != nil {
		log.Fatalf("NewClientWithConfig: %v", err)
	}

	symbol := "600519"

	fmt.Println("=== GetHistoryKline（日线前复权，短区间）===")
	kl, err := c.GetHistoryKline(ctx, symbol, "day", "qfq", "20240201", "20240229")
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		n := len(kl)
		if n > 3 {
			n = 3
		}
		fmt.Printf("  总条数=%d，展示最后 %d 条:\n", len(kl), n)
		for i := len(kl) - n; i < len(kl); i++ {
			b := kl[i]
			fmt.Printf("  %s 收=%.2f 量=%.0f\n", b.Date, b.Close, b.Volume)
		}
	}

	fmt.Println("\n=== GetTodayTimeline（当日分时）===")
	tl, err := c.GetTodayTimeline(ctx, symbol)
	if err != nil {
		log.Printf("失败: %v", err)
	} else if tl != nil {
		fmt.Printf("  昨收=%.2f 点数=%d\n", tl.PrevClose, len(tl.Data))
		if len(tl.Data) > 0 {
			last := tl.Data[len(tl.Data)-1]
			fmt.Printf("  最新分时: 时间=%s 价=%.3f\n", last.Time, last.Price)
		}
	}

	fmt.Println("\n=== GetTodayTimelineBatch（多标的）===")
	batch := c.GetTodayTimelineBatch(ctx, []string{"600519", "000001"})
	for sym, resp := range batch.Success {
		fmt.Printf("  成功 %s: 点数=%d\n", sym, len(resp.Data))
	}
	for sym, msg := range batch.Failed {
		fmt.Printf("  失败 %s: %s\n", sym, msg)
	}
}
