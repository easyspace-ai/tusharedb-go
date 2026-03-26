package parquet

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	goparquet "github.com/parquet-go/parquet-go"
)

type Writer struct {
	dataDir string
}

func NewWriter(dataDir string) *Writer {
	return &Writer{dataDir: dataDir}
}

// WriteRecords writes records to parquet (generic top-level function)
func WriteRecords[T any](w *Writer, dataset string, partition map[string]string, rows []T) (string, error) {
	if len(rows) == 0 {
		return "", nil
	}

	dir := w.datasetDir(dataset, partition)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create dataset dir: %w", err)
	}

	filename := fmt.Sprintf("part-%d.parquet", time.Now().UnixNano())
	finalPath := filepath.Join(dir, filename)
	tempPath := finalPath + ".tmp"

	f, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("create parquet temp file: %w", err)
	}

	writer := goparquet.NewGenericWriter[T](f)
	if _, err := writer.Write(rows); err != nil {
		_ = f.Close()
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("write parquet rows: %w", err)
	}
	if err := writer.Close(); err != nil {
		_ = f.Close()
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("close parquet writer: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("close parquet file: %w", err)
	}

	if err := os.Rename(tempPath, finalPath); err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("commit parquet file: %w", err)
	}

	return finalPath, nil
}

func (w *Writer) datasetDir(dataset string, partition map[string]string) string {
	base := filepath.Join(w.dataDir, "lake", dataset)
	if len(partition) == 0 {
		return base
	}
	if value, ok := partition["exchange"]; ok && value != "" {
		return filepath.Join(base, "exchange="+value)
	}
	if value, ok := partition["list_status"]; ok && value != "" {
		return filepath.Join(base, "list_status="+value)
	}
	if value, ok := partition["year"]; ok && value != "" {
		base = filepath.Join(base, "year="+value)
	}
	if value, ok := partition["month"]; ok && value != "" {
		base = filepath.Join(base, "month="+value)
	}
	return base
}
