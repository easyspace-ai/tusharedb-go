// Package multisource provides multi-source data collection with rate limiting,
// proxy support, and automatic failover.
//
// Example:
//
//	mgr := multisource.GetMultiSourceManager()
//
//	quotes, err := mgr.GetStockQuotes(ctx, []string{"000001"})
//
//	klines, err := mgr.GetKLine(ctx, "000001", "daily", "qfq", "20240101", "20241231")
package multisource
