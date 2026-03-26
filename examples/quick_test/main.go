package main

import (
	"context"
	"fmt"
	"log"

	"github.com/easyspace-ai/tusharedb-go/pkg/tsdb"
)

func main() {
	ctx := context.Background()

	// 创建自动缓存客户端
	client, err := tsdb.NewAutoClient("./data")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 获取股票基础信息
	fmt.Println("获取股票基础信息...")
	df, err := client.GetStockBasic(ctx, tsdb.StockBasicFilter{ListStatus: "L"})
	if err != nil {
		log.Printf("获取失败: %v", err)
	} else {
		fmt.Printf("获取到 %d 只股票\n", len(df.Rows))
	}

	// 获取单只股票日线数据（仅获取一只股票）
	fmt.Println("\n获取 000001.SZ 日线数据...")
	dailyDF, err := client.GetStockDaily(ctx, "000001.SZ", "20240101", "20241231", tsdb.AdjustQFQ)
	if err != nil {
		log.Printf("获取日线失败: %v", err)
	} else {
		fmt.Printf("获取到 %d 条日线数据\n", len(dailyDF.Rows))
		if len(dailyDF.Rows) > 0 {
			fmt.Printf("第一条: %+v\n", dailyDF.Rows[0])
			fmt.Printf("最后一条: %+v\n", dailyDF.Rows[len(dailyDF.Rows)-1])
		}
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

	// 检查同步状态
	fmt.Println("\n检查同步状态...")
	if lastDate, ok := client.GetLastSyncDate("daily"); ok {
		fmt.Printf("日线数据最后同步: %s\n", lastDate)
	} else {
		fmt.Println("日线数据从未同步")
	}

	fmt.Println("\n测试完成！")
}
