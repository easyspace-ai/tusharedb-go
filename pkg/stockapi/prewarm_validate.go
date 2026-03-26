package stockapi

import (
	"fmt"
	"time"

	"github.com/easyspace-ai/tusharedb-go/internal/provider/stocksdk"
)

// ValidateDailyKlineHistory checks daily K-line completeness (non-fatal warnings).
func ValidateDailyKlineHistory(rows []KlineData, reqStart, reqEnd string, minBars int) []string {
	if len(rows) == 0 {
		return []string{"empty"}
	}
	var warns []string
	minD, maxD := klineDateSpan(rows)
	if minBars > 0 && len(rows) < minBars {
		warns = append(warns, fmt.Sprintf("bars=%d<%d", len(rows), minBars))
	}
	startC := stocksdk.CompactDate(reqStart)
	endC := stocksdk.CompactDate(reqEnd)
	if startC != "" && minD != "" && minD > startC {
		warns = append(warns, fmt.Sprintf("missing_head data_min=%s req_start=%s", minD, startC))
	}
	if endC != "" && maxD != "" && maxD < endC {
		warns = append(warns, fmt.Sprintf("missing_tail data_max=%s req_end=%s", maxD, endC))
	}
	todayC := stocksdk.CompactDate(time.Now().Format("2006-01-02"))
	if maxD != "" && todayC != "" && maxD < todayC {
		if db := calendarDaysBetweenCompact(maxD, todayC); db > 10 {
			warns = append(warns, fmt.Sprintf("stale_tail max=%s behind_today=%dd", maxD, db))
		}
	}
	if span := calendarDaysBetweenCompact(minD, maxD); span > 30 {
		if float64(len(rows)) < float64(span)*0.25 {
			warns = append(warns, fmt.Sprintf("sparse_series span_days=%d bars=%d", span, len(rows)))
		}
	}
	if !isKlineStrictlySortedByDate(rows) {
		warns = append(warns, "dates_not_strictly_sorted")
	}
	return warns
}

func calendarDaysBetweenCompact(a, b string) int {
	if len(a) != 8 || len(b) != 8 {
		return 0
	}
	t1, err1 := time.ParseInLocation("20060102", a, time.Local)
	t2, err2 := time.ParseInLocation("20060102", b, time.Local)
	if err1 != nil || err2 != nil {
		return 0
	}
	if t1.After(t2) {
		t1, t2 = t2, t1
	}
	return int(t2.Sub(t1).Hours() / 24)
}

func isKlineStrictlySortedByDate(rows []KlineData) bool {
	var prev string
	for _, r := range rows {
		c := stocksdk.CompactDate(r.Date)
		if len(c) != 8 {
			continue
		}
		if prev != "" && c <= prev {
			return false
		}
		prev = c
	}
	return true
}
