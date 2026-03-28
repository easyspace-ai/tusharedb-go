package main

import (
	"context"
	"fmt"
	"os"

	"github.com/easyspace-ai/stock_api/pkg/tsdb"
)

func main() {
	token := os.Getenv("TUSHARE_TOKEN")
	if token == "" {
		fmt.Println("Error: TUSHARE_TOKEN environment variable not set")
		fmt.Println("Please set your Tushare token: export TUSHARE_TOKEN='your-token'")
		os.Exit(1)
	}

	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		panic(fmt.Sprintf("create data dir: %v", err))
	}

	client, err := tsdb.NewClient(tsdb.Config{
		Token:   token,
		DataDir: dataDir,
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx := context.Background()

	fmt.Println("=== Step 1: Sync core data (trade_cal + stock_basic) ===")
	if err := client.Downloader().SyncCore(ctx); err != nil {
		panic(fmt.Sprintf("SyncCore failed: %v", err))
	}
	fmt.Println("✓ Core data synced successfully")

	// 可选：同步最近一个交易日的日线数据
	fmt.Println("\n=== Step 2: Sync daily data (single date) ===")
	// 使用最近的一个交易日（示例：20241231）
	sampleDate := "20241231"
	if err := client.Downloader().SyncDailyByDate(ctx, sampleDate); err != nil {
		fmt.Printf("Warning: SyncDailyByDate failed (date may not be a trading day): %v\n", err)
	} else {
		fmt.Printf("✓ Daily data for %s synced successfully\n", sampleDate)
	}

	// 可选：同步复权因子
	fmt.Println("\n=== Step 3: Sync adj_factor (single date) ===")
	if err := client.Downloader().SyncAdjFactorRange(ctx, sampleDate, sampleDate); err != nil {
		fmt.Printf("Warning: SyncAdjFactorRange failed: %v\n", err)
	} else {
		fmt.Printf("✓ Adj_factor for %s synced successfully\n", sampleDate)
	}

	fmt.Println("\n✅ Download example completed!")
	fmt.Printf("Data stored in: %s\n", dataDir)
}
