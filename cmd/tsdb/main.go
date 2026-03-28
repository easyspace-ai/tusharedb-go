package main

import (
	"fmt"
	"os"

	"github.com/easyspace-ai/stock_api/pkg/tsdb"
)

func main() {
	client, err := tsdb.NewClient(tsdb.Config{
		Token:   getenv("TUSHARE_TOKEN", "placeholder-token"),
		DataDir: "./data",
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	fmt.Println("tusharedb-go scaffold initialized")
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
