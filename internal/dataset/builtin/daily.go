package builtin

import (
	"github.com/easyspace-ai/tusharedb-go/internal/dataset"
	"github.com/easyspace-ai/tusharedb-go/internal/provider"
)

func init() {
	datasetSpec := dataset.Spec{
		Name:          "daily",
		PrimaryKeys:   []string{"ts_code", "trade_date"},
		PartitionKeys: []string{"year", "month"},
		UpdateMode:    dataset.UpdateModeIncremental,
		FetchStrategy: dataset.FetchStrategy{Mode: provider.FetchModeByTradeDate, PreferByDate: true},
	}
	datasetRegister(datasetSpec)
}
