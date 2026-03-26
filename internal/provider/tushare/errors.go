package tushare

import "fmt"

type APIError struct {
	Code int
	Msg  string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("tushare api error: code=%d, msg=%s", e.Code, e.Msg)
}
