// Package main 演示「数据湖 + DuckDB」统一客户端 pkg/tsdb（UnifiedClient）。
//
// 默认使用 StockSDK（东财/腾讯等公开接口）作为主源，CacheModeAuto：
// 会先读本地 Parquet，缺失时触发同步再查。
//
// 依赖：可写目录 demo_data/lake（示例使用仓库下相对路径）。
// 首次运行可能较慢（需同步日线）。
//
// 运行：go run ./demo/lake_query/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/easyspace-ai/stock_api/pkg/tsdb"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	dataDir := filepath.Join("demo_data", "tsdb_lake")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Fatalf("创建目录: %v", err)
	}

	// NewAutoClient：Primary=StockSDK，自动缓存；若需纯在线可看 NewRealtimeClient
	client, err := tsdb.NewAutoClient(dataDir)
	if err != nil {
		log.Fatalf("NewAutoClient: %v", err)
	}
	defer client.Close()

	fmt.Println("=== GetStockBasic（抽样前 3 行；上市状态 L）===")
	basic, err := client.GetStockBasic(ctx, tsdb.StockBasicFilter{ListStatus: "L"})
	if err != nil {
		log.Fatalf("GetStockBasic: %v", err)
	}
	printFrameHead(basic, 3)

	fmt.Println("\n=== GetTradeCalendar（2024-01 沪市）===")
	cal, err := client.GetTradeCalendar(ctx, tsdb.TradeCalendarFilter{
		Exchange:  "SSE",
		StartDate: "20240101",
		EndDate:   "20240131",
	})
	if err != nil {
		log.Printf("GetTradeCalendar: %v", err)
	} else {
		printFrameHead(cal, 5)
	}

	fmt.Println("\n=== GetStockDaily（000001.SZ，短区间，前复权）===")
	// 短区间降低首次同步耗时
	daily, err := client.GetStockDaily(ctx, "000001.SZ", "20240201", "20240229", tsdb.AdjustQFQ)
	if err != nil {
		log.Printf("GetStockDaily: %v", err)
	} else {
		printFrameHead(daily, 5)
	}

	fmt.Println("\n=== Query（SQL 示例；依赖视图已注册）===")
	// 若本地尚无数据，Query 可能报错，仅作 API 形态演示
	qdf, err := client.Query(ctx, `SELECT COUNT(*) AS n FROM v_stock_basic`)
	if err != nil {
		log.Printf("Query: %v（可能尚未同步 stock_basic）", err)
	} else {
		printFrameHead(qdf, 5)
	}
}

func printFrameHead(df *tsdb.DataFrame, maxRows int) {
	if df == nil {
		fmt.Println("  (nil)")
		return
	}
	fmt.Printf("  列: %v\n", df.Columns)
	n := maxRows
	if len(df.Rows) < n {
		n = len(df.Rows)
	}
	for i := 0; i < n; i++ {
		fmt.Printf("  %v\n", df.Rows[i])
	}
	if len(df.Rows) > n {
		fmt.Printf("  ... 共 %d 行\n", len(df.Rows))
	}
}
