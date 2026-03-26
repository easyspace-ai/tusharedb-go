package tsdb

import "context"

type Downloader struct {
	client *Client
}

func (d *Downloader) SyncCore(ctx context.Context) error {
	return d.client.syncer.SyncCore(ctx)
}

func (d *Downloader) SyncTradeCalendar(ctx context.Context, startDate, endDate string) error {
	return d.client.syncer.SyncDatasetRange(ctx, "trade_cal", startDate, endDate)
}

func (d *Downloader) SyncStockBasic(ctx context.Context, listStatus string) error {
	return d.client.syncer.SyncStockBasic(ctx, listStatus)
}

func (d *Downloader) SyncDailyRange(ctx context.Context, startDate, endDate string) error {
	return d.client.syncer.SyncDatasetRange(ctx, "daily", startDate, endDate)
}

func (d *Downloader) SyncDailyByDate(ctx context.Context, tradeDate string) error {
	return d.client.syncer.SyncDatasetByDate(ctx, "daily", tradeDate)
}

func (d *Downloader) SyncDailyIncremental(ctx context.Context) error {
	return d.client.syncer.SyncDatasetIncremental(ctx, "daily")
}

func (d *Downloader) SyncAdjFactorRange(ctx context.Context, startDate, endDate string) error {
	return d.client.syncer.SyncDatasetRange(ctx, "adj_factor", startDate, endDate)
}

func (d *Downloader) SyncAdjFactorIncremental(ctx context.Context) error {
	return d.client.syncer.SyncDatasetIncremental(ctx, "adj_factor")
}

func (d *Downloader) SyncDailyBasicRange(ctx context.Context, startDate, endDate string) error {
	return d.client.syncer.SyncDatasetRange(ctx, "daily_basic", startDate, endDate)
}

func (d *Downloader) SyncDailyBasicIncremental(ctx context.Context) error {
	return d.client.syncer.SyncDatasetIncremental(ctx, "daily_basic")
}
