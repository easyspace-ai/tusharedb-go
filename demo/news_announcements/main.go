// Package main 演示「资讯 / 公告 / 研报」相关 API（pkg/realtimedata）。
//
// 涵盖：
//   - GetNews / GetStockNews：市场资讯、个股资讯
//   - GetStockNotices / GetStockReports：公告与研报列表
//   - GetNoticeContent / GetReportContent：东财正文（需 URL 中带 artCode/infoCode 或可解析）
//   - GetHotTopics：热门话题
//
// 运行：go run ./demo/news_announcements/
package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/easyspace-ai/stock_api/pkg/realtimedata"
)

var reArtCode = regexp.MustCompile(`(?i)art[_-]?code=([^&]+)`)
var reInfoCode = regexp.MustCompile(`(?i)infoCode=([^&]+)`)

func main() {
	cfg := realtimedata.Config{DataDir: "./demo_data", EnableStorage: false, CacheMode: realtimedata.CacheModeDisabled}
	client, err := realtimedata.NewClient(cfg)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	fmt.Println("=== GetNews（市场资讯 3 条）===")
	news, err := client.GetNews(ctx, 3)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for _, n := range news {
			fmt.Printf("  [%s] %s\n", n.Source, truncate(n.Title, 60))
		}
	}

	fmt.Println("\n=== GetStockNews（600519，2 条）===")
	sn, err := client.GetStockNews(ctx, "600519.SH", 2)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for _, n := range sn {
			fmt.Printf("  %s\n", truncate(n.Title, 70))
		}
	}

	fmt.Println("\n=== GetStockNotices（600519，3 条；若 URL 含 art_code 则尝试拉正文）===")
	notices, err := client.GetStockNotices(ctx, "600519.SH", 3)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for _, n := range notices {
			fmt.Printf("  %s | %s\n", n.Time, truncate(n.Title, 50))
			art := pickCode(n.Url, n.PdfUrl, reArtCode)
			if art == "" {
				continue
			}
			body, err := client.GetNoticeContent(ctx, "600519", art)
			if err != nil {
				log.Printf("    正文: %v", err)
				continue
			}
			fmt.Printf("    正文预览: %s\n", truncate(stripHTML(body), 120))
			break // 仅演示一条，避免请求过多
		}
	}

	fmt.Println("\n=== GetStockReports（600519，2 条；若 URL 含 infoCode 则尝试拉正文）===")
	reports, err := client.GetStockReports(ctx, "600519.SH", 2)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for _, r := range reports {
			fmt.Printf("  %s [%s] %s\n", r.Time, r.Institution, truncate(r.Title, 50))
			info := pickCode(r.Url, "", reInfoCode)
			if info == "" {
				continue
			}
			body, err := client.GetReportContent(ctx, info)
			if err != nil {
				log.Printf("    研报正文: %v", err)
				continue
			}
			fmt.Printf("    正文预览: %s\n", truncate(stripHTML(body), 120))
			break
		}
	}

	fmt.Println("\n=== GetHotTopics（前 5 个）===")
	topics, err := client.GetHotTopics(ctx)
	if err != nil {
		log.Printf("失败: %v", err)
	} else {
		for i := 0; i < 5 && i < len(topics); i++ {
			t := topics[i]
			fmt.Printf("  %s (关联标的数=%d)\n", truncate(t.Title, 60), t.Count)
		}
	}
}

func pickCode(primary, secondary string, re *regexp.Regexp) string {
	for _, raw := range []string{primary, secondary} {
		if raw == "" {
			continue
		}
		if m := re.FindStringSubmatch(raw); len(m) > 1 {
			return strings.TrimSpace(m[1])
		}
		if u, err := url.Parse(raw); err == nil {
			q := u.Query()
			for _, key := range []string{"art_code", "artCode", "infoCode", "infocode"} {
				if v := q.Get(key); v != "" {
					return v
				}
			}
		}
	}
	return ""
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func stripHTML(s string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	return strings.TrimSpace(re.ReplaceAllString(s, ""))
}
