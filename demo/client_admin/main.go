// Package main 演示 pkg/realtimedata 客户端运维向接口：缓存清理、配置快照、域名限流统计。
//
// 运行：go run ./demo/client_admin/
package main

import (
	"fmt"
	"log"

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
		log.Fatalf("创建客户端: %v", err)
	}

	fmt.Println("=== GetStats（当前 Client 配置摘要）===")
	for k, v := range client.GetStats() {
		fmt.Printf("  %s = %v\n", k, v)
	}

	fmt.Println("\n=== GetRateLimiterStats（示例域名 eastmoney.com）===")
	st := client.GetRateLimiterStats("eastmoney.com")
	for k, v := range st {
		fmt.Printf("  %s = %v\n", k, v)
	}

	fmt.Println("\n=== ClearAllCache（清空 RequestManager 内存缓存）===")
	client.ClearCache()
	fmt.Println("  已调用 ClearCache()")
}
