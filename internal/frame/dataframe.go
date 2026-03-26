package frame

type DataFrame struct {
	Columns []string
	Rows    []map[string]any
}

func (df *DataFrame) Len() int {
	if df == nil {
		return 0
	}
	return len(df.Rows)
}

func (df *DataFrame) Empty() bool {
	return df.Len() == 0
}

func (df *DataFrame) Records() []map[string]any {
	if df == nil {
		return nil
	}
	return df.Rows
}
