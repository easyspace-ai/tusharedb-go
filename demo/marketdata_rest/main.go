// Package main 演示 pkg/marketdata：面向东财等站点的 REST 封装（轻量 HTTP 客户端）。
// 与 pkg/realtimedata 相比更底层，适合只需某一类接口、希望自行组合的场景。
//
// 运行：go run ./demo/marketdata_rest/
package main

import (
	"fmt"
	"log"

	"github.com/easyspace-ai/tusharedb-go/pkg/marketdata"
)

func main() {
	c := marketdata.NewClient(marketdata.DefaultConfig())

	fmt.Println("=== GetGlobalIndexes（全球指数，前 6 条）===")
	idx, err := c.GetGlobalIndexes()
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for i := 0; i < 6 && i < len(idx); i++ {
			x := idx[i]
			fmt.Printf("  %s %s 价=%.2f\n", x.Code, x.Name, x.Price)
		}
	}

	fmt.Println("\n=== GetHotTopics（热点，前 5 条）===")
	topics, err := c.GetHotTopics()
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for i := 0; i < 5 && i < len(topics); i++ {
			t := topics[i]
			fmt.Printf("  %s\n", t.Title)
		}
	}

	fmt.Println("\n=== GetNews24h（7x24 快讯，2 条）===")
	news, total, err := c.GetNews24h(1, 2)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		fmt.Printf("  total=%d\n", total)
		for _, n := range news {
			fmt.Printf("  %s\n", n.Title)
		}
	}
}
