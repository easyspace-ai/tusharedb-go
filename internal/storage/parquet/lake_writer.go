package parquet

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	goparquet "github.com/parquet-go/parquet-go"
)

// ========== 实时行情数据存储 ==========

// QuoteRecord 行情记录（用于Parquet存储）
type QuoteRecord struct {
	Code      string  `parquet:"code,plain"`
	Name      string  `parquet:"name,plain"`
	Price     float64 `parquet:"price,plain"`
	PrevClose float64 `parquet:"prev_close,plain"`
	Open      float64 `parquet:"open,plain"`
	High      float64 `parquet:"high,plain"`
	Low       float64 `parquet:"low,plain"`
	Volume    float64 `parquet:"volume,plain"`
	Amount    float64 `parquet:"amount,plain"`
	Change    float64 `parquet:"change,plain"`
	ChangePct float64 `parquet:"change_pct,plain"`
	Time      string  `parquet:"time,plain"`
	Timestamp int64   `parquet:"timestamp,plain"`
}

// KLineRecord K线记录（用于Parquet存储）
type KLineRecord struct {
	Code      string  `parquet:"code,plain"`
	Date      string  `parquet:"date,plain"`
	Period    string  `parquet:"period,plain"`
	Adjust    string  `parquet:"adjust,plain"`
	Open      float64 `parquet:"open,plain"`
	High      float64 `parquet:"high,plain"`
	Low       float64 `parquet:"low,plain"`
	Close     float64 `parquet:"close,plain"`
	Volume    float64 `parquet:"volume,plain"`
	Amount    float64 `parquet:"amount,plain"`
	Timestamp int64   `parquet:"timestamp,plain"`
}

// ========== 实时数据湖存储管理 ==========

// RealTimeLakeManager 实时数据湖管理器
type RealTimeLakeManager struct {
	baseDir   string
	quotePath string
	klinePath string
	mu        sync.RWMutex
}

// NewRealTimeLakeManager 创建实时数据湖管理器
func NewRealTimeLakeManager(baseDir string) *RealTimeLakeManager {
	return &RealTimeLakeManager{
		baseDir:   baseDir,
		quotePath: filepath.Join(baseDir, "realtime", "quotes"),
		klinePath: filepath.Join(baseDir, "realtime", "klines"),
	}
}

// Init 初始化目录
func (m *RealTimeLakeManager) Init() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(m.quotePath, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(m.klinePath, 0o755); err != nil {
		return err
	}
	return nil
}

// ========== 行情存储 ==========

// getQuotePartitionPath 获取行情分区路径
func (m *RealTimeLakeManager) getQuotePartitionPath(date string) string {
	year := date[:4]
	month := date[4:6]
	day := date[6:8]
	return filepath.Join(m.quotePath, fmt.Sprintf("year=%s/month=%s/day=%s", year, month, day))
}

// SaveQuotes 保存行情数据到Parquet
func (m *RealTimeLakeManager) SaveQuotes(ctx context.Context, date string, quotes []QuoteRecord) error {
	if len(quotes) == 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	partitionPath := m.getQuotePartitionPath(date)
	if err := os.MkdirAll(partitionPath, 0o755); err != nil {
		return err
	}

	// 生成文件名：timestamp.parquet
	timestamp := time.Now().UnixNano()
	fileName := fmt.Sprintf("%d.parquet", timestamp)
	filePath := filepath.Join(partitionPath, fileName)

	// 写入临时文件
	tmpPath := filePath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// 写入Parquet
	w := goparquet.NewWriter(f, goparquet.SchemaOf(QuoteRecord{}))
	if err := w.Write(quotes); err != nil {
		w.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := w.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// 原子重命名
	return os.Rename(tmpPath, filePath)
}

// ReadQuotes 读取行情数据
func (m *RealTimeLakeManager) ReadQuotes(ctx context.Context, date string, codes []string) ([]QuoteRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	partitionPath := m.getQuotePartitionPath(date)
	if _, err := os.Stat(partitionPath); os.IsNotExist(err) {
		return []QuoteRecord{}, nil
	}

	// 列出所有parquet文件
	entries, err := os.ReadDir(partitionPath)
	if err != nil {
		return nil, err
	}

	var records []QuoteRecord
	codeSet := make(map[string]bool)
	for _, c := range codes {
		codeSet[c] = true
	}
	filterByCode := len(codes) > 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".parquet") {
			continue
		}

		filePath := filepath.Join(partitionPath, entry.Name())
		f, err := os.Open(filePath)
		if err != nil {
			continue
		}

		var batch []QuoteRecord
		r := goparquet.NewReader(f, goparquet.SchemaOf(QuoteRecord{}))
		if err := r.Read(&batch); err == nil {
			for _, r := range batch {
				if filterByCode && !codeSet[r.Code] {
					continue
				}
				records = append(records, r)
			}
		}
		r.Close()
		f.Close()
	}

	// 按时间戳排序，保留最新的
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp > records[j].Timestamp
	})

	// 去重，保留每个code的最新记录
	seen := make(map[string]bool)
	var result []QuoteRecord
	for _, r := range records {
		if !seen[r.Code] {
			seen[r.Code] = true
			result = append(result, r)
		}
	}

	return result, nil
}

// ========== K线存储 ==========

// getKLinePartitionPath 获取K线分区路径
func (m *RealTimeLakeManager) getKLinePartitionPath(code string) string {
	// 按股票代码分目录
	return filepath.Join(m.klinePath, fmt.Sprintf("code=%s", code))
}

// SaveKLines 保存K线数据到Parquet
func (m *RealTimeLakeManager) SaveKLines(ctx context.Context, code string, period string, adjust string, klines []KLineRecord) error {
	if len(klines) == 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	partitionPath := m.getKLinePartitionPath(code)
	if err := os.MkdirAll(partitionPath, 0o755); err != nil {
		return err
	}

	// 文件名：period_adjust.parquet
	fileName := fmt.Sprintf("%s_%s.parquet", period, adjust)
	filePath := filepath.Join(partitionPath, fileName)

	// 先读取现有数据
	var existing []KLineRecord
	if _, err := os.Stat(filePath); err == nil {
		f, err := os.Open(filePath)
		if err == nil {
			r := goparquet.NewReader(f, goparquet.SchemaOf(KLineRecord{}))
			r.Read(&existing)
			r.Close()
			f.Close()
		}
	}

	// 合并去重
	dateMap := make(map[string]KLineRecord)
	for _, k := range existing {
		dateMap[k.Date] = k
	}
	for _, k := range klines {
		dateMap[k.Date] = k
	}

	// 转换为列表
	var merged []KLineRecord
	for _, k := range dateMap {
		merged = append(merged, k)
	}

	// 按日期排序
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Date < merged[j].Date
	})

	// 写入临时文件
	tmpPath := filePath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := goparquet.NewWriter(f, goparquet.SchemaOf(KLineRecord{}))
	if err := w.Write(merged); err != nil {
		w.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := w.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return os.Rename(tmpPath, filePath)
}

// ReadKLines 读取K线数据
func (m *RealTimeLakeManager) ReadKLines(ctx context.Context, code string, period string, adjust string, startDate string, endDate string) ([]KLineRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	partitionPath := m.getKLinePartitionPath(code)
	fileName := fmt.Sprintf("%s_%s.parquet", period, adjust)
	filePath := filepath.Join(partitionPath, fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []KLineRecord{}, nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var records []KLineRecord
	r := goparquet.NewReader(f, goparquet.SchemaOf(KLineRecord{}))
	if err := r.Read(&records); err != nil {
		return nil, err
	}

	// 过滤日期范围
	var result []KLineRecord
	for _, r := range records {
		if startDate != "" && r.Date < startDate {
			continue
		}
		if endDate != "" && r.Date > endDate {
			continue
		}
		result = append(result, r)
	}

	return result, nil
}

// ========== 辅助函数 ==========

// ListQuoteDates 列出有行情数据的日期
func (m *RealTimeLakeManager) ListQuoteDates() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	yearDirs, err := os.ReadDir(m.quotePath)
	if err != nil {
		return nil, err
	}

	var dates []string
	for _, yearDir := range yearDirs {
		if !yearDir.IsDir() || !filepath.HasPrefix(yearDir.Name(), "year=") {
			continue
		}
		year := yearDir.Name()[5:]
		yearPath := filepath.Join(m.quotePath, yearDir.Name())

		monthDirs, err := os.ReadDir(yearPath)
		if err != nil {
			continue
		}
		for _, monthDir := range monthDirs {
			if !monthDir.IsDir() || !filepath.HasPrefix(monthDir.Name(), "month=") {
				continue
			}
			month := monthDir.Name()[6:]
			monthPath := filepath.Join(yearPath, monthDir.Name())

			dayDirs, err := os.ReadDir(monthPath)
			if err != nil {
				continue
			}
			for _, dayDir := range dayDirs {
				if !dayDir.IsDir() || !filepath.HasPrefix(dayDir.Name(), "day=") {
					continue
				}
				day := dayDir.Name()[4:]
				dates = append(dates, year+month+day)
			}
		}
	}

	sort.Strings(dates)
	return dates, nil
}

// ClearOldQuotes 清理旧的行情数据（保留最近N天）
func (m *RealTimeLakeManager) ClearOldQuotes(ctx context.Context, keepDays int) error {
	if keepDays <= 0 {
		keepDays = 7
	}

	cutoffDate := time.Now().AddDate(0, 0, -keepDays).Format("20060102")

	dates, err := m.ListQuoteDates()
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, date := range dates {
		if date < cutoffDate {
			partitionPath := m.getQuotePartitionPath(date)
			os.RemoveAll(partitionPath)
		}
	}

	return nil
}
