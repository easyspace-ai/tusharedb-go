package tushare

import (
	"time"
)

// 内部使用的 Tushare 特定类型
type apiRequest struct {
	APIName string         `json:"api_name"`
	Token   string         `json:"token"`
	Params  map[string]any `json:"params,omitempty"`
	Fields  string         `json:"fields,omitempty"`
}

type apiResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Fields []string        `json:"fields"`
		Items  [][]interface{} `json:"items"`
	} `json:"data"`
}

// Config 定义 Tushare 客户端配置
type Config struct {
	Token      string
	Endpoint   string
	Timeout    time.Duration
	Retries    int
	RetryWait  time.Duration
	PageLimit  int
	UserAgent  string
}
