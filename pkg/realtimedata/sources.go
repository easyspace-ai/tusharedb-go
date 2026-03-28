package realtimedata

import "github.com/easyspace-ai/stock_api/internal/provider/multisource"

// DataSourceType 数据源枚举（与 multisource 一致）。
type DataSourceType = multisource.DataSourceType

const (
	DataSourceEastMoney   = multisource.DataSourceEastMoney
	DataSourceSina        = multisource.DataSourceSina
	DataSourceTencent     = multisource.DataSourceTencent
	DataSourceXueqiu      = multisource.DataSourceXueqiu
	DataSourceBaidu       = multisource.DataSourceBaidu
	DataSourceTonghuashun = multisource.DataSourceTonghuashun
)
