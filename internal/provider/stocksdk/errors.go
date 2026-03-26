package stocksdk

import "fmt"

// APIError StockSDK API 错误
type APIError struct {
	Code    int
	Msg     string
	RawResp []byte
}

func (e *APIError) Error() string {
	return fmt.Sprintf("stocksdk api error: code=%d, msg=%s", e.Code, e.Msg)
}
