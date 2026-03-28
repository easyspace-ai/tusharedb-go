package dataset

import "github.com/easyspace-ai/stock_api/internal/provider"

type UpdateMode string

const (
	UpdateModeSnapshot    UpdateMode = "snapshot"
	UpdateModeAppend      UpdateMode = "append"
	UpdateModeIncremental UpdateMode = "incremental"
)

type FetchStrategy struct {
	Mode         provider.FetchMode
	PreferByDate bool
}

type Spec struct {
	Name          string
	PrimaryKeys   []string
	PartitionKeys []string
	UpdateMode    UpdateMode
	FetchStrategy FetchStrategy
}
