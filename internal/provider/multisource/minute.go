package multisource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// MinuteBar 分时成交/价格点（腾讯 minute 接口）
type MinuteBar struct {
	Time      string  `json:"time"`
	Price     float64 `json:"price"`
	Volume    float64 `json:"volume"`
	ChangePct float64 `json:"changePct"`
}

// GetMinuteData 获取当日分时数据（腾讯行情；code 支持 6 位或带 sh/sz 前缀）。
func (s *TencentFullSource) GetMinuteData(ctx context.Context, code string) ([]MinuteBar, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	sym := normalizeCodeForKlineAPI(strings.TrimSpace(code))
	if sym == "" {
		return nil, fmt.Errorf("invalid stock code: %s", code)
	}
	rawURL := fmt.Sprintf("https://web.ifzq.gtimg.cn/appstock/app/minute/query?code=%s", url.QueryEscape(sym))
	data, err := s.reqMgr.GetWithRateLimit("qq.com", rawURL)
	if err != nil {
		return nil, err
	}

	var result struct {
		Code int `json:"code"`
		Data map[string]struct {
			Data struct {
				Data []string `json:"data"`
			} `json:"data"`
			Qt struct {
				Stock []string `json:"stock"`
			} `json:"qt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	var out []MinuteBar
	for _, v := range result.Data {
		preClose := 0.0
		if len(v.Qt.Stock) > 4 {
			preClose = parseFloat(v.Qt.Stock[4])
		}
		for _, item := range v.Data.Data {
			parts := strings.Fields(item)
			if len(parts) < 3 {
				continue
			}
			price := parseFloat(parts[1])
			vol := float64(parseInt64Trim(parts[2]))
			chg := 0.0
			if preClose > 0 {
				chg = (price - preClose) / preClose * 100
			}
			out = append(out, MinuteBar{
				Time:      parts[0],
				Price:     price,
				Volume:    vol,
				ChangePct: chg,
			})
		}
		break
	}
	return out, nil
}
