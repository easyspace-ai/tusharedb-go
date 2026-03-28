package builtin

import (
	"github.com/easyspace-ai/stock_api/internal/dataset"
	"github.com/easyspace-ai/stock_api/internal/provider"
)

func init() {
	datasetSpec := dataset.Spec{
		Name:          "trade_cal",
		PrimaryKeys:   []string{"exchange", "cal_date"},
		PartitionKeys: []string{"exchange"},
		UpdateMode:    dataset.UpdateModeIncremental,
		FetchStrategy: dataset.FetchStrategy{Mode: provider.FetchModeSingle},
	}
	datasetRegister(datasetSpec)
}
