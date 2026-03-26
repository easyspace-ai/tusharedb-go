package stockapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/easyspace-ai/tusharedb-go/internal/provider/stocksdk"
	goparquet "github.com/parquet-go/parquet-go"
)

// klineParquetRow is the on-disk schema for history K-line (one file per symbol×period×adjust).
type klineParquetRow struct {
	Date   string  `parquet:"date"`
	Open   float64 `parquet:"open"`
	High   float64 `parquet:"high"`
	Low    float64 `parquet:"low"`
	Close  float64 `parquet:"close"`
	Volume float64 `parquet:"volume"`
	Amount float64 `parquet:"amount"`
}

func (c *Client) historyLakeDir() string {
	return filepath.Join(c.config.DataDir, "stockapi-history", "kline")
}

func (c *Client) klineParquetPath(symbol, period, adjust string) string {
	safe := func(s string) string {
		s = strings.ToLower(strings.TrimSpace(s))
		s = strings.ReplaceAll(s, string(filepath.Separator), "_")
		s = strings.ReplaceAll(s, "/", "_")
		s = strings.ReplaceAll(s, "\\", "_")
		return s
	}
	name := fmt.Sprintf("%s_%s_%s.parquet", safe(period), safe(adjust), safe(symbol))
	return filepath.Join(c.historyLakeDir(), name)
}

func (c *Client) withHistoryFileLock(path string, fn func() error) error {
	v, _ := c.historyPathLocks.LoadOrStore(path, &sync.Mutex{})
	m := v.(*sync.Mutex)
	m.Lock()
	defer m.Unlock()
	return fn()
}

func readKlineParquet(path string) ([]KlineData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := goparquet.NewGenericReader[klineParquetRow](f)
	defer r.Close()

	var parquetBuf []klineParquetRow
	chunk := make([]klineParquetRow, 4096)
	for {
		n, err := r.Read(chunk)
		if n > 0 {
			parquetBuf = append(parquetBuf, chunk[:n]...)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if n == 0 {
			break
		}
	}
	out := make([]KlineData, 0, len(parquetBuf))
	for _, row := range parquetBuf {
		out = append(out, KlineData{
			Date:   row.Date,
			Open:   row.Open,
			High:   row.High,
			Low:    row.Low,
			Close:  row.Close,
			Volume: row.Volume,
			Amount: row.Amount,
		})
	}
	return out, nil
}

func writeKlineParquetAtomic(path string, rows []KlineData) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	parquetRows := make([]klineParquetRow, 0, len(rows))
	for _, r := range rows {
		parquetRows = append(parquetRows, klineParquetRow{
			Date:   r.Date,
			Open:   r.Open,
			High:   r.High,
			Low:    r.Low,
			Close:  r.Close,
			Volume: r.Volume,
			Amount: r.Amount,
		})
	}
	tempPath := path + ".tmp." + fmt.Sprintf("%d", time.Now().UnixNano())
	f, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	w := goparquet.NewGenericWriter[klineParquetRow](f)
	if _, err := w.Write(parquetRows); err != nil {
		_ = f.Close()
		_ = os.Remove(tempPath)
		return err
	}
	if err := w.Close(); err != nil {
		_ = f.Close()
		_ = os.Remove(tempPath)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	return nil
}

func klineDateSpan(rows []KlineData) (minCompact, maxCompact string) {
	minD := "99999999"
	maxD := ""
	for _, r := range rows {
		c := stocksdk.CompactDate(r.Date)
		if len(c) != 8 {
			continue
		}
		if c < minD {
			minD = c
		}
		if c > maxD {
			maxD = c
		}
	}
	if maxD == "" {
		return "", ""
	}
	return minD, maxD
}

func klineCoversRange(rows []KlineData, startDate, endDate string) bool {
	if len(rows) == 0 {
		return false
	}
	minD, maxD := klineDateSpan(rows)
	startC := stocksdk.CompactDate(strings.TrimSpace(startDate))
	endC := stocksdk.CompactDate(strings.TrimSpace(endDate))
	if startC != "" && minD != "" && minD > startC {
		return false
	}
	if endC != "" && maxD != "" && maxD < endC {
		return false
	}
	return true
}

func filterKlineRange(rows []KlineData, startDate, endDate string) []KlineData {
	startC := stocksdk.CompactDate(strings.TrimSpace(startDate))
	endC := stocksdk.CompactDate(strings.TrimSpace(endDate))
	if startC == "" && endC == "" {
		return rows
	}
	out := make([]KlineData, 0, len(rows))
	for _, r := range rows {
		c := stocksdk.CompactDate(r.Date)
		if len(c) != 8 {
			continue
		}
		if startC != "" && c < startC {
			continue
		}
		if endC != "" && c > endC {
			continue
		}
		out = append(out, r)
	}
	return out
}

func mergeKlineByDate(existing []KlineData, fromAPI []KlineData) []KlineData {
	m := make(map[string]KlineData, len(existing)+len(fromAPI))
	for _, r := range existing {
		key := stocksdk.CompactDate(r.Date)
		if key == "" {
			continue
		}
		m[key] = r
	}
	for _, r := range fromAPI {
		key := stocksdk.CompactDate(r.Date)
		if key == "" {
			continue
		}
		m[key] = r
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]KlineData, 0, len(keys))
	for _, k := range keys {
		out = append(out, m[k])
	}
	return out
}

func historyFetchParams(disk []KlineData, userStart, userEnd string) (apiStart, apiEnd string) {
	uStart := strings.TrimSpace(userStart)
	uEnd := strings.TrimSpace(userEnd)
	startC := stocksdk.CompactDate(uStart)
	endC := stocksdk.CompactDate(uEnd)
	if len(disk) == 0 {
		if startC == "" && endC == "" {
			return "", ""
		}
		if endC == "" {
			return startC, ""
		}
		return startC, endC
	}
	_, maxD := klineDateSpan(disk)
	minD, _ := klineDateSpan(disk)
	if startC != "" && minD != "" && minD > startC {
		if endC == "" {
			return startC, ""
		}
		return startC, endC
	}
	if endC != "" && maxD != "" && maxD < endC {
		return maxD, endC
	}
	if maxD != "" {
		return maxD, ""
	}
	return startC, endC
}

func (c *Client) getHistoryKlineWithParquet(ctx context.Context, symbol, period, adjust, startDate, endDate string) ([]KlineData, error) {
	period = normalizePeriod(period)
	adjust = strings.TrimSpace(adjust)
	symbol = strings.TrimSpace(symbol)
	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)

	if c.config.CacheMode == CacheModeDisabled {
		return c.fetchHistoryKlineNoStore(ctx, symbol, period, adjust, startDate, endDate)
	}

	path := c.klineParquetPath(symbol, period, adjust)
	var disk []KlineData
	var modTime time.Time
	if fi, err := os.Stat(path); err == nil {
		rows, err := readKlineParquet(path)
		if err == nil {
			disk = rows
			modTime = fi.ModTime()
		}
	}

	covers := klineCoversRange(disk, startDate, endDate)
	fresh := len(disk) == 0 || time.Since(modTime) <= c.config.HistoryTTL
	if len(disk) > 0 && covers && fresh {
		return filterKlineRange(disk, startDate, endDate), nil
	}

	if c.config.CacheMode == CacheModeReadOnly {
		if len(disk) > 0 && covers {
			return filterKlineRange(disk, startDate, endDate), nil
		}
		return nil, errors.New("kline parquet miss or stale in readonly mode")
	}

	apiStart, apiEnd := historyFetchParams(disk, startDate, endDate)
	raw, err := c.sdk.GetHistoryKline(ctx, symbol, &stocksdk.HistoryKlineOptions{
		Period:    stocksdk.KlinePeriod(period),
		Adjust:    stocksdk.AdjustType(adjust),
		StartDate: apiStart,
		EndDate:   apiEnd,
	})
	if err != nil {
		if len(disk) > 0 && covers {
			return filterKlineRange(disk, startDate, endDate), nil
		}
		return nil, err
	}
	apiRows := make([]KlineData, 0, len(raw))
	for _, k := range raw {
		apiRows = append(apiRows, KlineData{
			Date:   k.Date,
			Open:   deref(k.Open),
			High:   deref(k.High),
			Low:    deref(k.Low),
			Close:  deref(k.Close),
			Volume: deref(k.Volume),
			Amount: deref(k.Amount),
		})
	}
	merged := mergeKlineByDate(disk, apiRows)
	if err := c.withHistoryFileLock(path, func() error {
		return writeKlineParquetAtomic(path, merged)
	}); err != nil {
		return filterKlineRange(merged, startDate, endDate), err
	}
	return filterKlineRange(merged, startDate, endDate), nil
}

func (c *Client) fetchHistoryKlineNoStore(ctx context.Context, symbol, period, adjust, startDate, endDate string) ([]KlineData, error) {
	raw, err := c.sdk.GetHistoryKline(ctx, symbol, &stocksdk.HistoryKlineOptions{
		Period:    stocksdk.KlinePeriod(period),
		Adjust:    stocksdk.AdjustType(adjust),
		StartDate: stocksdk.CompactDate(strings.TrimSpace(startDate)),
		EndDate:   stocksdk.CompactDate(strings.TrimSpace(endDate)),
	})
	if err != nil {
		return nil, err
	}
	result := make([]KlineData, 0, len(raw))
	for _, k := range raw {
		result = append(result, KlineData{
			Date:   k.Date,
			Open:   deref(k.Open),
			High:   deref(k.High),
			Low:    deref(k.Low),
			Close:  deref(k.Close),
			Volume: deref(k.Volume),
			Amount: deref(k.Amount),
		})
	}
	return result, nil
}
