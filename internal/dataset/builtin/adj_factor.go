package builtin

import (
	"github.com/easyspace-ai/stock_api/internal/dataset"
	"github.com/easyspace-ai/stock_api/internal/provider"
)

func init() {
	datasetSpec := dataset.Spec{
		Name:          "adj_factor",
		PrimaryKeys:   []string{"ts_code", "trade_date"},
		PartitionKeys: []string{"year", "month"},
		UpdateMode:    dataset.UpdateModeIncremental,
		FetchStrategy: dataset.FetchStrategy{Mode: provider.FetchModeByTradeDate, PreferByDate: true},
	}
	datasetRegister(datasetSpec)
}
