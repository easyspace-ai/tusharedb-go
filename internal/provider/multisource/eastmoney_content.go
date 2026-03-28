package multisource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// GetNoticeContent 拉取公告正文：优先 np-anotice-stock JSON；正文为空时再抓详情页 HTML 纯文本。
func (s *EastMoneyFullSource) GetNoticeContent(ctx context.Context, stockCode, artCode string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	artCode = strings.TrimSpace(artCode)
	if artCode == "" {
		return "", fmt.Errorf("artCode is required")
	}
	fullCode := normalizeEastMoneyNoticeCode(stockCode)
	code6 := fullCode
	if strings.HasPrefix(fullCode, "sh") || strings.HasPrefix(fullCode, "sz") || strings.HasPrefix(fullCode, "bj") {
		code6 = fullCode[2:]
	}

	apiURL := fmt.Sprintf(
		"https://np-anotice-stock.eastmoney.com/security/ann?ann_id=%s&stock_list=%s",
		url.QueryEscape(artCode), url.QueryEscape(code6),
	)
	body, err := s.reqMgr.GetWithRateLimit("eastmoney.com", apiURL)
	if err != nil {
		return "", fmt.Errorf("notice api: %w", err)
	}

	var result struct {
		Data struct {
			List []struct {
				Title      string `json:"title"`
				Content    string `json:"content"`
				NoticeDate string `json:"notice_date"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("notice json: %w", err)
	}
	if len(result.Data.List) == 0 {
		return "", fmt.Errorf("empty notice list from api")
	}
	n := result.Data.List[0]
	content := strings.TrimSpace(n.Content)
	if content == "" {
		content = s.fetchNoticeDetailHTML(ctx, fullCode, artCode)
	}
	if content == "" {
		detail := fmt.Sprintf("https://data.eastmoney.com/notices/detail/%s/%s.html", fullCode, artCode)
		return fmt.Sprintf("公告链接：%s\n（正文未能直接解析，请打开链接查看。）", detail), nil
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("标题：%s\n", n.Title))
	b.WriteString(fmt.Sprintf("发布日期：%s\n\n", n.NoticeDate))
	b.WriteString("内容：\n")
	b.WriteString(content)
	return b.String(), nil
}

func (s *EastMoneyFullSource) fetchNoticeDetailHTML(ctx context.Context, stockCode, artCode string) string {
	select {
	case <-ctx.Done():
		return ""
	default:
	}
	u := fmt.Sprintf("https://data.eastmoney.com/notices/detail/%s/%s.html", stockCode, artCode)
	body, err := s.reqMgr.GetWithRateLimit("eastmoney.com", u)
	if err != nil || len(body) == 0 {
		return ""
	}
	return stripHTMLToText(string(body))
}

// GetReportContent 拉取研报正文：尝试 reportapi 元数据与详情接口；仍无正文则尝试抓取详情页纯文本。
func (s *EastMoneyFullSource) GetReportContent(ctx context.Context, infoCode string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	infoCode = strings.TrimSpace(infoCode)
	if infoCode == "" {
		return "", fmt.Errorf("infoCode is required")
	}

	// 1) 常见 JSON 接口（字段随东财调整可能变化）
	infoURL := fmt.Sprintf("https://reportapi.eastmoney.com/report/info?infoCode=%s", url.QueryEscape(infoCode))
	body, err := s.reqMgr.GetWithRateLimit("eastmoney.com", infoURL)
	if err == nil && len(body) > 0 {
		var wrap struct {
			Data json.RawMessage `json:"data"`
		}
		if json.Unmarshal(body, &wrap) == nil {
			var asMap map[string]any
			if json.Unmarshal(wrap.Data, &asMap) == nil {
				for _, key := range []string{"content", "reportContent", "summary", "digest", "remark"} {
					if v, ok := asMap[key]; ok {
						if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
							return strings.TrimSpace(s), nil
						}
					}
				}
			}
		}
	}

	// 2) 详情页（query 使用 infoCode）
	detail := fmt.Sprintf("https://data.eastmoney.com/report/zw_stock.jshtml?infoCode=%s", url.QueryEscape(infoCode))
	htmlb, err := s.reqMgr.GetWithRateLimit("eastmoney.com", detail)
	if err == nil && len(htmlb) > 0 {
		txt := stripHTMLToText(string(htmlb))
		if len(strings.TrimSpace(txt)) > 200 {
			return txt, nil
		}
	}

	return "", fmt.Errorf("研报正文无法从当前接口稳定解析，请使用研报列表返回的 Url 在浏览器打开（infoCode=%s）", infoCode)
}

func normalizeEastMoneyNoticeCode(stockCode string) string {
	c := strings.TrimSpace(strings.ToLower(stockCode))
	if c == "" {
		return ""
	}
	if strings.HasPrefix(c, "sh") || strings.HasPrefix(c, "sz") || strings.HasPrefix(c, "bj") {
		return c
	}
	if len(c) == 6 {
		if strings.HasPrefix(c, "6") || strings.HasPrefix(c, "9") || strings.HasPrefix(c, "5") {
			return "sh" + c
		}
		if strings.HasPrefix(c, "0") || strings.HasPrefix(c, "2") || strings.HasPrefix(c, "3") {
			return "sz" + c
		}
		if strings.HasPrefix(c, "4") || strings.HasPrefix(c, "8") {
			return "bj" + c
		}
	}
	return c
}

var (
	scriptStyleRx = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	styleTagRx    = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	tagRx         = regexp.MustCompile(`(?s)<[^>]+>`)
	spaceRx       = regexp.MustCompile(`[ \t\r\n]+`)
)

func stripHTMLToText(html string) string {
	s := scriptStyleRx.ReplaceAllString(html, " ")
	s = styleTagRx.ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = tagRx.ReplaceAllString(s, "\n")
	s = spaceRx.ReplaceAllString(strings.TrimSpace(s), " ")
	lines := strings.Split(s, "\n")
	var b strings.Builder
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	out := b.String()
	if len(out) > 512000 {
		out = out[:512000] + "\n…(truncated)"
	}
	return out
}
