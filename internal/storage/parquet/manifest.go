package parquet

type Manifest struct {
	Dataset string   `json:"dataset"`
	Files   []string `json:"files"`
}
