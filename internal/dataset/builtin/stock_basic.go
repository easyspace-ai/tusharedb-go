package builtin

import (
	"github.com/easyspace-ai/tusharedb-go/internal/dataset"
	"github.com/easyspace-ai/tusharedb-go/internal/provider"
)

func init() {
	datasetSpec := dataset.Spec{
		Name:          "stock_basic",
		PrimaryKeys:   []string{"ts_code"},
		PartitionKeys: []string{"list_status"},
		UpdateMode:    dataset.UpdateModeSnapshot,
		FetchStrategy: dataset.FetchStrategy{Mode: provider.FetchModePaged},
	}
	datasetRegister(datasetSpec)
}
