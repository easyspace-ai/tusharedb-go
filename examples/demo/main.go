package main

import (
	"context"
	"fmt"
	"github.com/easyspace-ai/tusharedb-go/internal/provider/stocksdk"
)

func main() {
	// 测试健康检查（预期返回未实现错误）
	ctx := context.Background()
	testProvider := stocksdk.NewClient(stocksdk.Config{
		APIKey: "test-key",
	})

	a, err := testProvider.FetchStockBasic(ctx, "L")
	fmt.Printf("✓ FetchStockBasic: %v\n", err)
	if err != nil {
	}

	fmt.Println(a)
}
