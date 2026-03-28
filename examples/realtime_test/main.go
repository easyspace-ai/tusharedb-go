package main

import (
	"context"
	"fmt"
	"log"

	"github.com/easyspace-ai/stock_api/pkg/tsdb"
)

func main() {
	ctx := context.Background()

	// 使用实时模式（禁用缓存），直接获取网络数据
	// 这样不会触发全量同步，直接返回请求的股票数据
	fmt.Println("=== 实时模式测试（禁用缓存）===")

	client, err := tsdb.NewRealtimeClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 获取单只股票日线数据（不复权）
	fmt.Println("\n获取 000001.SZ 日线数据（不复权）...")
	dailyDF, err := client.GetStockDaily(ctx, "000001.SZ", "20240101", "20241231", tsdb.AdjustNone)
	if err != nil {
		log.Printf("获取日线失败: %v", err)
	} else {
		fmt.Printf("获取到 %d 条日线数据\n", len(dailyDF.Rows))
		if len(dailyDF.Rows) > 0 {
			fmt.Printf("第一条: %+v\n", dailyDF.Rows[0])
			fmt.Printf("最后一条: %+v\n", dailyDF.Rows[len(dailyDF.Rows)-1])
		}
	}

	// 获取单只股票日线数据（前复权）
	fmt.Println("\n获取 000001.SZ 日线数据（前复权）...")
	dailyQFQ, err := client.GetStockDaily(ctx, "000001.SZ", "20240101", "20241231", tsdb.AdjustQFQ)
	if err != nil {
		log.Printf("获取日线(QFQ)失败: %v", err)
	} else {
		fmt.Printf("获取到 %d 条日线数据（前复权）\n", len(dailyQFQ.Rows))
	}

	// 获取复权因子
	fmt.Println("\n获取 000001.SZ 复权因子...")
	adjDF, err := client.GetAdjFactor(ctx, "000001.SZ", "20240101", "20241231")
	if err != nil {
		log.Printf("获取复权因子失败: %v", err)
	} else {
		fmt.Printf("获取到 %d 条复权因子数据\n", len(adjDF.Rows))
		if len(adjDF.Rows) > 0 {
			fmt.Printf("第一条: %+v\n", adjDF.Rows[0])
		}
	}

	// 获取每日指标
	fmt.Println("\n获取 000001.SZ 每日指标...")
	basicDF, err := client.GetDailyBasic(ctx, "000001.SZ", "20240101", "20241231")
	if err != nil {
		log.Printf("获取每日指标失败: %v", err)
	} else {
		fmt.Printf("获取到 %d 条每日指标数据\n", len(basicDF.Rows))
		if len(basicDF.Rows) > 0 {
			fmt.Printf("第一条: %+v\n", basicDF.Rows[0])
		}
	}

	fmt.Println("\n测试完成！")
}
