package main

import (
	"context"
	"fmt"
	"log"

	"github.com/easyspace-ai/stock_api/pkg/tsdb"
)

func main() {
	ctx := context.Background()

	// ============ 使用方式 1：自动模式（推荐） ============
	// 自动模式：本地有数据直接用，没有则自动下载并缓存
	fmt.Println("=== 模式1：自动缓存模式 ===")

	client, err := tsdb.NewUnifiedClient(tsdb.UnifiedConfig{
		PrimaryDataSource: tsdb.DataSourceStockSDK, // 使用 StockSDK 数据源
		DataDir:           "./data",                // 数据存储目录
		CacheMode:         tsdb.CacheModeAuto,      // 自动模式（默认）
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 第一次调用：本地没有数据，会自动下载
	fmt.Println("获取股票基础信息（首次会下载）...")
	df, err := client.GetStockBasic(ctx, tsdb.StockBasicFilter{ListStatus: "L"})
	if err != nil {
		log.Printf("获取失败: %v", err)
	} else {
		fmt.Printf("获取到 %d 只股票\n", len(df.Rows))
	}

	// 第二次调用：直接从本地读取，速度极快
	fmt.Println("再次获取（从缓存读取）...")
	df, err = client.GetStockBasic(ctx, tsdb.StockBasicFilter{ListStatus: "L"})
	if err != nil {
		log.Printf("获取失败: %v", err)
	} else {
		fmt.Printf("从缓存获取到 %d 只股票\n", len(df.Rows))
	}

	// 获取日线数据（自动处理复权）
	fmt.Println("\n获取日线数据...")
	dailyDF, err := client.GetStockDaily(ctx, "000001.SZ", "20240101", "20241231", tsdb.AdjustQFQ)
	if err != nil {
		log.Printf("获取日线失败: %v", err)
	} else {
		fmt.Printf("获取到 %d 条日线数据\n", len(dailyDF.Rows))
		if len(dailyDF.Rows) > 0 {
			fmt.Printf("最新数据: %+v\n", dailyDF.Rows[len(dailyDF.Rows)-1])
		}
	}

	// ============ 使用方式 2：纯离线模式 ============
	fmt.Println("\n=== 模式2：纯离线模式 ===")

	offlineClient, err := tsdb.NewOfflineClient("./data")
	if err != nil {
		log.Fatal(err)
	}
	defer offlineClient.Close()

	// 只读本地，不会触发任何网络请求
	df, err = offlineClient.GetStockBasic(ctx, tsdb.StockBasicFilter{})
	if err != nil {
		fmt.Printf("离线读取失败（数据不存在）: %v\n", err)
	} else {
		fmt.Printf("离线读取成功: %d 条\n", len(df.Rows))
	}

	// ============ 使用方式 3：实时模式 ============
	fmt.Println("\n=== 模式3：实时模式（禁用缓存） ===")

	realtimeClient, err := tsdb.NewRealtimeClient()
	if err != nil {
		log.Fatal(err)
	}
	defer realtimeClient.Close()

	// 总是从网络获取，不读不写本地缓存
	df, err = realtimeClient.GetStockBasic(ctx, tsdb.StockBasicFilter{ListStatus: "L"})
	if err != nil {
		log.Printf("实时获取失败: %v", err)
	} else {
		fmt.Printf("实时获取到 %d 只股票\n", len(df.Rows))
	}

	// ============ 使用方式 4：手动同步控制 ============
	fmt.Println("\n=== 模式4：手动同步控制 ===")

	autoClient, err := tsdb.NewAutoClient("./data")
	if err != nil {
		log.Fatal(err)
	}
	defer autoClient.Close()

	// 检查上次同步日期
	if lastDate, ok := autoClient.GetLastSyncDate("daily"); ok {
		fmt.Printf("日线数据最后同步: %s\n", lastDate)
	} else {
		fmt.Println("日线数据从未同步")
	}

	// 手动触发全量同步（首次部署时使用）
	// err = autoClient.SyncCore(ctx)
	// err = autoClient.SyncDailyRange(ctx, "20200101", "20241231")

	// 手动触发增量同步（每日收盘后运行）
	// err = autoClient.SyncIncremental(ctx)

	fmt.Println("\n所有操作完成！")
}
