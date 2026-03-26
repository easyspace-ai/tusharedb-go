package syncer

import (
	"context"
	"fmt"
	"time"

	"github.com/easyspace-ai/tusharedb-go/internal/dataset"
	"github.com/easyspace-ai/tusharedb-go/internal/provider"
	"github.com/easyspace-ai/tusharedb-go/internal/query/duckdb"
	"github.com/easyspace-ai/tusharedb-go/internal/storage/meta"
	"github.com/easyspace-ai/tusharedb-go/internal/storage/parquet"
)

type Config struct {
	DataDir string
}

type Syncer struct {
	cfg        Config
	provider   provider.DataProvider
	registry   *dataset.Registry
	checkpoint *meta.CheckpointStore
	writer     *parquet.Writer
	manifests  *parquet.ManifestStore
	engine     *duckdb.Engine
}

func NewSyncer(cfg Config, provider provider.DataProvider, registry *dataset.Registry, checkpoint *meta.CheckpointStore, engine *duckdb.Engine) *Syncer {
	return &Syncer{
		cfg:        cfg,
		provider:   provider,
		registry:   registry,
		checkpoint: checkpoint,
		writer:     parquet.NewWriter(cfg.DataDir),
		manifests:  parquet.NewManifestStore(cfg.DataDir),
		engine:     engine,
	}
}

func (s *Syncer) SyncCore(ctx context.Context) error {
	if err := s.SyncDatasetRange(ctx, "trade_cal", "19900101", "20301231"); err != nil {
		return err
	}
	if err := s.SyncStockBasic(ctx, "L"); err != nil {
		return err
	}
	return nil
}

func (s *Syncer) SyncStockBasic(ctx context.Context, listStatus string) error {
	rows, err := s.provider.FetchStockBasic(ctx, listStatus)
	if err != nil {
		return fmt.Errorf("fetch stock_basic: %w", err)
	}
	file, err := parquet.WriteRecords(s.writer, "stock_basic", map[string]string{"list_status": listStatus}, rows)
	if err != nil {
		return fmt.Errorf("write stock_basic parquet: %w", err)
	}
	if err := s.manifests.Append("stock_basic", file); err != nil {
		return fmt.Errorf("update stock_basic manifest: %w", err)
	}
	return s.checkpoint.Put(meta.DatasetCheckpoint{
		Dataset:          "stock_basic",
		LastSyncedDate:   time.Now().Format("20060102"),
		LastSuccessfulAt: time.Now(),
		SchemaVersion:    "v1",
	})
}

func (s *Syncer) SyncDatasetRange(ctx context.Context, datasetName, startDate, endDate string) error {
	var file string

	switch datasetName {
	case "trade_cal":
		rows, err := s.provider.FetchTradeCalendar(ctx, startDate, endDate)
		if err != nil {
			return fmt.Errorf("fetch trade_cal: %w", err)
		}
		file, err = parquet.WriteRecords(s.writer, "trade_cal", map[string]string{}, rows)
		if err != nil {
			return fmt.Errorf("write trade_cal parquet: %w", err)
		}
		if err := s.manifests.Append("trade_cal", file); err != nil {
			return fmt.Errorf("update trade_cal manifest: %w", err)
		}

	case "daily":
		rows, err := s.provider.FetchDailyRange(ctx, startDate, endDate)
		if err != nil {
			return fmt.Errorf("fetch daily: %w", err)
		}
		// 按年月分区写入
		partitions := make(map[string][]provider.DailyRow)
		for _, row := range rows {
			if len(row.TradeDate) >= 6 {
				year := row.TradeDate[:4]
				month := row.TradeDate[4:6]
				key := year + "-" + month
				partitions[key] = append(partitions[key], row)
			}
		}
		for partKey, partRows := range partitions {
			file, err = parquet.WriteRecords(s.writer, "daily", map[string]string{"partition": partKey}, partRows)
			if err != nil {
				return fmt.Errorf("write daily parquet (partition %s): %w", partKey, err)
			}
			if err := s.manifests.Append("daily", file); err != nil {
				return fmt.Errorf("update daily manifest: %w", err)
			}
		}

	case "adj_factor":
		rows, err := s.provider.FetchAdjFactorRange(ctx, startDate, endDate)
		if err != nil {
			return fmt.Errorf("fetch adj_factor: %w", err)
		}
		partitions := make(map[string][]provider.AdjFactorRow)
		for _, row := range rows {
			if len(row.TradeDate) >= 6 {
				year := row.TradeDate[:4]
				month := row.TradeDate[4:6]
				key := year + "-" + month
				partitions[key] = append(partitions[key], row)
			}
		}
		for partKey, partRows := range partitions {
			file, err = parquet.WriteRecords(s.writer, "adj_factor", map[string]string{"partition": partKey}, partRows)
			if err != nil {
				return fmt.Errorf("write adj_factor parquet (partition %s): %w", partKey, err)
			}
			if err := s.manifests.Append("adj_factor", file); err != nil {
				return fmt.Errorf("update adj_factor manifest: %w", err)
			}
		}

	case "daily_basic":
		rows, err := s.provider.FetchDailyBasicRange(ctx, startDate, endDate)
		if err != nil {
			return fmt.Errorf("fetch daily_basic: %w", err)
		}
		partitions := make(map[string][]provider.DailyBasicRow)
		for _, row := range rows {
			if len(row.TradeDate) >= 6 {
				year := row.TradeDate[:4]
				month := row.TradeDate[4:6]
				key := year + "-" + month
				partitions[key] = append(partitions[key], row)
			}
		}
		for partKey, partRows := range partitions {
			file, err = parquet.WriteRecords(s.writer, "daily_basic", map[string]string{"partition": partKey}, partRows)
			if err != nil {
				return fmt.Errorf("write daily_basic parquet (partition %s): %w", partKey, err)
			}
			if err := s.manifests.Append("daily_basic", file); err != nil {
				return fmt.Errorf("update daily_basic manifest: %w", err)
			}
		}

	default:
		return fmt.Errorf("range sync not implemented for dataset %q", datasetName)
	}

	return s.checkpoint.Put(meta.DatasetCheckpoint{
		Dataset:          datasetName,
		LastSyncedDate:   endDate,
		LastSuccessfulAt: time.Now(),
		SchemaVersion:    "v1",
	})
}

func (s *Syncer) SyncDatasetByDate(ctx context.Context, datasetName, tradeDate string) error {
	var file string

	switch datasetName {
	case "daily":
		rows, err := s.provider.FetchDaily(ctx, tradeDate)
		if err != nil {
			return fmt.Errorf("fetch daily: %w", err)
		}
		year := tradeDate[:4]
		month := tradeDate[4:6]
		file, err = parquet.WriteRecords(s.writer, "daily", map[string]string{"year": year, "month": month}, rows)
		if err != nil {
			return fmt.Errorf("write daily parquet: %w", err)
		}
		if err := s.manifests.Append("daily", file); err != nil {
			return fmt.Errorf("update daily manifest: %w", err)
		}

	case "adj_factor":
		rows, err := s.provider.FetchAdjFactor(ctx, tradeDate)
		if err != nil {
			return fmt.Errorf("fetch adj_factor: %w", err)
		}
		year := tradeDate[:4]
		month := tradeDate[4:6]
		file, err = parquet.WriteRecords(s.writer, "adj_factor", map[string]string{"year": year, "month": month}, rows)
		if err != nil {
			return fmt.Errorf("write adj_factor parquet: %w", err)
		}
		if err := s.manifests.Append("adj_factor", file); err != nil {
			return fmt.Errorf("update adj_factor manifest: %w", err)
		}

	case "daily_basic":
		rows, err := s.provider.FetchDailyBasic(ctx, tradeDate)
		if err != nil {
			return fmt.Errorf("fetch daily_basic: %w", err)
		}
		year := tradeDate[:4]
		month := tradeDate[4:6]
		file, err = parquet.WriteRecords(s.writer, "daily_basic", map[string]string{"year": year, "month": month}, rows)
		if err != nil {
			return fmt.Errorf("write daily_basic parquet: %w", err)
		}
		if err := s.manifests.Append("daily_basic", file); err != nil {
			return fmt.Errorf("update daily_basic manifest: %w", err)
		}

	default:
		return fmt.Errorf("date sync not implemented for dataset %q", datasetName)
	}

	return s.checkpoint.Put(meta.DatasetCheckpoint{
		Dataset:          datasetName,
		LastSyncedDate:   tradeDate,
		LastSuccessfulAt: time.Now(),
		SchemaVersion:    "v1",
	})
}

func (s *Syncer) SyncDatasetIncremental(ctx context.Context, datasetName string) error {
	// 先获取 checkpoint
	cp, ok := s.checkpoint.Get(datasetName)
	if !ok {
		// 如果没有 checkpoint，从一个合理的日期开始
		cp = meta.DatasetCheckpoint{
			LastSyncedDate: "20200101",
		}
	}

	// 获取交易日历，从 LastSyncedDate + 1 到今天
	today := time.Now().Format("20060102")
	calRows, err := s.provider.FetchTradeCalendar(ctx, cp.LastSyncedDate, today)
	if err != nil {
		return fmt.Errorf("fetch trade_cal for incremental: %w", err)
	}

	// 过滤出 is_open = 1 且 > LastSyncedDate 的交易日
	var missingDates []string
	for _, cal := range calRows {
		if cal.IsOpen == "1" && cal.CalDate > cp.LastSyncedDate {
			missingDates = append(missingDates, cal.CalDate)
		}
	}

	// 逐个同步缺失的交易日
	for _, tradeDate := range missingDates {
		if err := s.SyncDatasetByDate(ctx, datasetName, tradeDate); err != nil {
			return fmt.Errorf("sync %s on %s: %w", datasetName, tradeDate, err)
		}
	}

	return nil
}
