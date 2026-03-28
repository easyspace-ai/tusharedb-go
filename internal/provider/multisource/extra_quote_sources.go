package multisource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func toShSzLower(code string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", fmt.Errorf("empty code")
	}
	u := strings.ToUpper(code)
	if i := strings.Index(u, "."); i > 0 {
		sym := strings.TrimSpace(u[:i])
		ex := strings.TrimSpace(u[i+1:])
		switch ex {
		case "SH":
			return "sh" + sym, nil
		case "SZ":
			return "sz" + sym, nil
		case "BJ":
			return "bj" + sym, nil
		}
	}
	return AddMarketPrefix(RemoveMarketPrefix(code)), nil
}

func quoteMapToSlice(m map[string]StockQuote, order []string) []StockQuote {
	var out []StockQuote
	for _, k := range order {
		if q, ok := m[k]; ok {
			out = append(out, q)
		}
	}
	return out
}

// --- 雪球 ---

type XueqiuFullSource struct {
	priority int
	reqMgr   *RequestManager
}

func NewXueqiuSource(priority int, reqMgr *RequestManager) *XueqiuFullSource {
	return &XueqiuFullSource{priority: priority, reqMgr: reqMgr}
}

func (s *XueqiuFullSource) Name() string         { return "xueqiu" }
func (s *XueqiuFullSource) Type() DataSourceType { return DataSourceXueqiu }
func (s *XueqiuFullSource) Priority() int        { return s.priority }
func (s *XueqiuFullSource) HealthCheck(ctx context.Context) error {
	_, err := s.GetStockQuotes(ctx, []string{"600519.SH"})
	return err
}
func (s *XueqiuFullSource) GetKLine(ctx context.Context, code string, period string, adjust string, startDate string, endDate string) ([]KLineItem, error) {
	return nil, ErrKLineNotSupported
}

func (s *XueqiuFullSource) GetStockQuotes(ctx context.Context, codes []string) ([]StockQuote, error) {
	var order []string
	var symList []string
	codeMap := make(map[string]string)
	for _, code := range codes {
		sz, err := toShSzLower(code)
		if err != nil || strings.HasPrefix(sz, "bj") {
			continue
		}
		xCode := strings.ToUpper(sz[:2]) + sz[2:]
		symList = append(symList, xCode)
		codeMap[xCode] = sz
		order = append(order, sz)
	}
	if len(symList) == 0 {
		return nil, fmt.Errorf("xueqiu: no codes")
	}
	rawURL := fmt.Sprintf("https://stock.xueqiu.com/v5/stock/realtime/quotec.json?symbol=%s", url.QueryEscape(strings.Join(symList, ",")))
	data, err := s.reqMgr.GetWithRateLimitEx("stock.xueqiu.com", rawURL, "https://xueqiu.com/", map[string]string{
		"Cookie": "device_id=00000000-0000-0000-0000-000000000001",
	})
	if err != nil {
		return nil, err
	}
	var jsonResp struct {
		Data []struct {
			Symbol    string  `json:"symbol"`
			Current   float64 `json:"current"`
			Percent   float64 `json:"percent"`
			Chg       float64 `json:"chg"`
			High      float64 `json:"high"`
			Low       float64 `json:"low"`
			Open      float64 `json:"open"`
			LastClose float64 `json:"last_close"`
			Volume    int64   `json:"volume"`
			Amount    float64 `json:"amount"`
		} `json:"data"`
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_description"`
	}
	if err := json.Unmarshal(data, &jsonResp); err != nil {
		return nil, err
	}
	if jsonResp.ErrorCode != 0 {
		return nil, fmt.Errorf("xueqiu api: %s", jsonResp.ErrorMessage)
	}
	res := make(map[string]StockQuote)
	for _, item := range jsonResp.Data {
		orig, ok := codeMap[item.Symbol]
		if !ok {
			continue
		}
		res[orig] = StockQuote{
			Code: orig, Price: item.Current, PrevClose: item.LastClose, Open: item.Open,
			High: item.High, Low: item.Low, Volume: float64(item.Volume), Amount: item.Amount,
			Change: item.Chg, ChangePct: item.Percent,
			Time: time.Now().Format("2006-01-02 15:04:05"),
		}
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("xueqiu: empty")
	}
	return quoteMapToSlice(res, order), nil
}

// --- 百度股市通 ---

type BaiduFullSource struct {
	priority int
	reqMgr   *RequestManager
}

func NewBaiduSource(priority int, reqMgr *RequestManager) *BaiduFullSource {
	return &BaiduFullSource{priority: priority, reqMgr: reqMgr}
}

func (s *BaiduFullSource) Name() string         { return "baidu" }
func (s *BaiduFullSource) Type() DataSourceType { return DataSourceBaidu }
func (s *BaiduFullSource) Priority() int        { return s.priority }
func (s *BaiduFullSource) HealthCheck(ctx context.Context) error {
	_, err := s.GetStockQuotes(ctx, []string{"000001.SZ"})
	return err
}
func (s *BaiduFullSource) GetKLine(ctx context.Context, code string, period string, adjust string, startDate string, endDate string) ([]KLineItem, error) {
	return nil, ErrKLineNotSupported
}

func (s *BaiduFullSource) GetStockQuotes(ctx context.Context, codes []string) ([]StockQuote, error) {
	list := make([]string, 0, len(codes))
	order := make([]string, 0, len(codes))
	for _, code := range codes {
		sz, err := toShSzLower(code)
		if err != nil {
			continue
		}
		list = append(list, sz)
		order = append(order, sz)
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("baidu: no codes")
	}
	rawURL := fmt.Sprintf(
		"https://finance.pae.baidu.com/selfselect/getstockquotation?all=1&code=%s&isIndex=false&isBk=false&isBlock=false&isFutures=false&isStock=true&newFormat=1&is_kc=0",
		url.QueryEscape(strings.Join(list, ",")),
	)
	data, err := s.reqMgr.GetWithRateLimitEx("finance.pae.baidu.com", rawURL, "https://gushitong.baidu.com/", nil)
	if err != nil {
		return nil, err
	}
	rows, err := parseBaiduStockQuotationJSON(data)
	if err != nil {
		return nil, err
	}
	res := make(map[string]StockQuote)
	for _, row := range rows {
		code := strings.TrimSpace(stringFromAny(baiduPick(row, "code", "Code")))
		if code == "" {
			continue
		}
		ex := strings.ToLower(strings.TrimSpace(stringFromAny(baiduPick(row, "exchange", "Exchange", "market"))))
		var orig string
		switch ex {
		case "sh":
			orig = "sh" + code
		case "sz":
			orig = "sz" + code
		default:
			orig = code
		}
		ratio := strings.TrimSuffix(strings.TrimSpace(stringFromAny(baiduPick(row, "ratio", "Ratio", "percent"))), "%")
		name := stringFromAny(baiduPick(row, "name", "Name", "stockName"))
		res[orig] = StockQuote{
			Code: orig, Name: name,
			Price:     parseFloat(stringFromAny(baiduPick(row, "price", "Price", "current", "last"))),
			PrevClose: parseFloat(stringFromAny(baiduPick(row, "preClose", "preclose", "yesterday_close", "prev_close"))),
			Open:      parseFloat(stringFromAny(baiduPick(row, "open", "Open"))),
			High:      parseFloat(stringFromAny(baiduPick(row, "high", "High"))),
			Low:       parseFloat(stringFromAny(baiduPick(row, "low", "Low"))),
			Volume:    float64(parseInt64Trim(stringFromAny(baiduPick(row, "volume", "Volume", "vol")))),
			Amount:    parseFloat(stringFromAny(baiduPick(row, "amount", "Amount", "turnover"))),
			Change:    parseFloat(stringFromAny(baiduPick(row, "increase", "Increase", "change", "px_change"))),
			ChangePct: parseFloat(ratio),
			Time:      time.Now().Format("2006-01-02 15:04:05"),
		}
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("baidu: empty")
	}
	return quoteMapToSlice(res, order), nil
}

// parseBaiduStockQuotationJSON 兼容百度股市通 selfselect/getstockquotation：
// Result 可能为 { "list": [ {...} ] }，也可能直接为 [ {...} ]。
// 外层用 map 解包，避免 ResultCode 等字段在 JSON 中为数字导致整段反序列化失败。
func parseBaiduStockQuotationJSON(data []byte) ([]map[string]interface{}, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, err
	}
	raw, ok := top["Result"]
	if !ok || len(raw) == 0 {
		raw, ok = top["result"]
	}
	if !ok || len(raw) == 0 {
		return nil, fmt.Errorf("baidu: missing Result")
	}
	// 形态一：{ "list": [ {...} ] }
	var wrapped struct {
		List []map[string]interface{} `json:"list"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil && len(wrapped.List) > 0 {
		return wrapped.List, nil
	}
	// 形态二：直接数组
	var arr []map[string]interface{}
	if err := json.Unmarshal(raw, &arr); err == nil && len(arr) > 0 {
		return arr, nil
	}
	if rc, ok := top["ResultCode"]; ok {
		return nil, fmt.Errorf("baidu ResultCode=%s", strings.TrimSpace(string(rc)))
	}
	return nil, fmt.Errorf("baidu: cannot parse Result JSON")
}

func baiduPick(row map[string]interface{}, keys ...string) interface{} {
	for _, k := range keys {
		if v, ok := row[k]; ok && v != nil {
			return v
		}
	}
	return nil
}

func stringFromAny(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return string(t)
	case bool:
		if t {
			return "1"
		}
		return "0"
	default:
		return fmt.Sprint(t)
	}
}

// --- 同花顺 (last.js 逐只) ---

type TonghuashunFullSource struct {
	priority int
	reqMgr   *RequestManager
}

func NewTonghuashunSource(priority int, reqMgr *RequestManager) *TonghuashunFullSource {
	return &TonghuashunFullSource{priority: priority, reqMgr: reqMgr}
}

func (s *TonghuashunFullSource) Name() string         { return "tonghuashun" }
func (s *TonghuashunFullSource) Type() DataSourceType { return DataSourceTonghuashun }
func (s *TonghuashunFullSource) Priority() int        { return s.priority }

func (s *TonghuashunFullSource) HealthCheck(ctx context.Context) error {
	_, err := s.GetStockQuotes(ctx, []string{"600519.SH"})
	return err
}
func (s *TonghuashunFullSource) GetKLine(ctx context.Context, code string, period string, adjust string, startDate string, endDate string) ([]KLineItem, error) {
	return nil, ErrKLineNotSupported
}

func thsKindPrefix(sz string) string {
	sz = strings.ToLower(sz)
	if strings.HasPrefix(sz, "sh") {
		return "hs_" + sz[2:]
	}
	if strings.HasPrefix(sz, "sz") {
		return "sz_" + sz[2:]
	}
	return ""
}

func thsExtractInner(root map[string]interface{}) map[string]interface{} {
	for _, v := range root {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

func thsPrevClose(inner map[string]interface{}) float64 {
	if s, ok := inner["pre"].(string); ok && strings.TrimSpace(s) != "" {
		return parseFloat(strings.TrimSpace(s))
	}
	return floatFromObj(inner, "pre")
}

// thsParseDataBars 解析同花顺 last.js 内 data 字段：time,price,volume,... 以分号分条。
func thsParseDataBars(dataStr string) (lastPrice, open, high, low, volSum float64, lastHHMM string, ok bool) {
	segs := strings.Split(dataStr, ";")
	var firstOpen bool
	high = -1e18
	low = 1e18
	for _, seg := range segs {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		f := strings.Split(seg, ",")
		if len(f) < 3 {
			continue
		}
		px := parseFloat(f[1])
		v := parseFloat(f[2])
		if !firstOpen {
			open = px
			firstOpen = true
		}
		if px > high {
			high = px
		}
		if px < low {
			low = px
		}
		volSum += v
		lastPrice = px
		lastHHMM = strings.TrimSpace(f[0])
	}
	if !firstOpen {
		return 0, 0, 0, 0, 0, "", false
	}
	return lastPrice, open, high, low, volSum, lastHHMM, true
}

func (s *TonghuashunFullSource) GetStockQuotes(ctx context.Context, codes []string) ([]StockQuote, error) {
	var out []StockQuote
	for _, code := range codes {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		sz, err := toShSzLower(code)
		if err != nil || strings.HasPrefix(sz, "bj") {
			continue
		}
		kind := thsKindPrefix(sz)
		if kind == "" {
			continue
		}
		rawURL := fmt.Sprintf("https://d.10jqka.com.cn/v6/time/%s/last.js", kind)
		data, err := s.reqMgr.GetWithRateLimit("d.10jqka.com.cn", rawURL)
		if err != nil {
			continue
		}
		bodyStr := string(data)
		i := strings.Index(bodyStr, "(")
		j := strings.LastIndex(bodyStr, ")")
		if i < 0 || j <= i+1 {
			continue
		}
		var root map[string]interface{}
		if err := json.Unmarshal([]byte(bodyStr[i+1:j]), &root); err != nil {
			continue
		}
		inner := thsExtractInner(root)
		if inner == nil {
			continue
		}
		name := stringFromObj(inner, "name")
		prev := thsPrevClose(inner)
		dataStr, _ := inner["data"].(string)
		price, openBar, high, low, volSum, lastHHMM, barOK := thsParseDataBars(dataStr)
		openMeta := floatFromObj(inner, "open")
		open := openMeta
		if open <= 0 && barOK {
			open = openBar
		}
		if !barOK && price <= 0 {
			continue
		}
		if price <= 0 && prev > 0 {
			price = prev
		}
		dateStr, _ := inner["date"].(string)
		tm := time.Now().Format("2006-01-02 15:04:05")
		if len(dateStr) == 8 && len(lastHHMM) == 4 {
			if loc, err := time.LoadLocation("Asia/Shanghai"); err == nil {
				if t, err := time.ParseInLocation("20060102150405", dateStr+lastHHMM+"00", loc); err == nil {
					tm = t.Format("2006-01-02 15:04:05")
				}
			}
		}
		q := StockQuote{
			Code: sz, Name: name, Price: price, PrevClose: prev,
			Open: open, High: high, Low: low, Volume: volSum, Amount: 0,
			Time: tm,
		}
		if prev > 0 {
			q.Change = price - prev
			q.ChangePct = q.Change / prev * 100
		}
		out = append(out, q)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("tonghuashun: no quotes (interface may have changed)")
	}
	return out, nil
}

func stringFromObj(obj map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := obj[k]; ok {
			switch t := v.(type) {
			case string:
				return t
			case float64:
				return strconv.FormatFloat(t, 'f', -1, 64)
			}
		}
	}
	return ""
}

func floatFromObj(obj map[string]interface{}, keys ...string) float64 {
	for _, k := range keys {
		if v, ok := obj[k]; ok {
			switch t := v.(type) {
			case float64:
				return t
			case int:
				return float64(t)
			case int64:
				return float64(t)
			case string:
				return parseFloat(t)
			}
		}
	}
	return 0
}
