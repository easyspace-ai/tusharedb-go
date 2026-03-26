package main

import (
	"context"
	"fmt"
	"os"

	"github.com/easyspace-ai/tusharedb-go/pkg/tsdb"
)

func main() {
	dataDir := "./data"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		fmt.Printf("Error: Data directory %s does not exist\n", dataDir)
		fmt.Println("Please run basic_download first to sync some data")
		os.Exit(1)
	}

	client, err := tsdb.NewClient(tsdb.Config{
		DataDir: dataDir,
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx := context.Background()

	// 示例 1: 查询交易日历
	fmt.Println("=== Example 1: Get Trade Calendar ===")
	isOpen := true
	calFilter := tsdb.TradeCalendarFilter{
		StartDate: "20241201",
		EndDate:   "20241231",
		IsOpen:    &isOpen,
	}
	calDF, err := client.Reader().GetTradeCalendar(ctx, calFilter)
	if err != nil {
		fmt.Printf("GetTradeCalendar failed: %v\n", err)
	} else {
		fmt.Printf("Found %d trading days\n", calDF.Len())
		for _, row := range calDF.Rows {
			fmt.Printf("  %s (is_open=%v)\n", row["cal_date"], row["is_open"])
		}
	}

	// 示例 2: 查询股票基础信息
	fmt.Println("\n=== Example 2: Get Stock Basic ===")
	stockFilter := tsdb.StockBasicFilter{
		ListStatus: "L",
	}
	stockDF, err := client.Reader().GetStockBasic(ctx, stockFilter)
	if err != nil {
		fmt.Printf("GetStockBasic failed: %v\n", err)
	} else {
		fmt.Printf("Found %d stocks\n", stockDF.Len())
		// 显示前 5 只股票
		for i, row := range stockDF.Rows {
			if i >= 5 {
				break
			}
			fmt.Printf("  %s: %s (%s)\n", row["ts_code"], row["name"], row["market"])
		}
		if stockDF.Len() > 5 {
			fmt.Printf("  ... and %d more\n", stockDF.Len()-5)
		}
	}

	// 示例 3: 查询日线数据（不复权）
	fmt.Println("\n=== Example 3: Get Stock Daily (None Adjustment) ===")
	dailyDF, err := client.Reader().GetStockDaily(ctx, "000001.SZ", "20240101", "20241231", tsdb.AdjustNone)
	if err != nil {
		fmt.Printf("GetStockDaily (none) failed: %v\n", err)
	} else {
		fmt.Printf("Found %d daily records (none)\n", dailyDF.Len())
		for i, row := range dailyDF.Rows {
			if i >= 3 {
				break
			}
			fmt.Printf("  %s: open=%.2f close=%.2f vol=%.0f\n",
				row["trade_date"], row["open"], row["close"], row["vol"])
		}
	}

	// 示例 4: 查询日线数据（前复权）
	fmt.Println("\n=== Example 4: Get Stock Daily (QFQ Adjustment) ===")
	qfqDF, err := client.Reader().GetStockDaily(ctx, "000001.SZ", "20240101", "20241231", tsdb.AdjustQFQ)
	if err != nil {
		fmt.Printf("GetStockDaily (qfq) failed: %v\n", err)
	} else {
		fmt.Printf("Found %d daily records (qfq)\n", qfqDF.Len())
		for i, row := range qfqDF.Rows {
			if i >= 3 {
				break
			}
			fmt.Printf("  %s: open=%.2f close=%.2f\n",
				row["trade_date"], row["open"], row["close"])
		}
	}

	fmt.Println("\n✅ Query example completed!")
}
