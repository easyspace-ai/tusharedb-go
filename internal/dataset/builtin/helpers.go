package builtin

import "github.com/easyspace-ai/stock_api/internal/dataset"

func datasetRegister(spec dataset.Spec) {
	specCopy := spec
	dataset.AddBuiltinRegistrar(func(r *dataset.Registry) {
		r.Register(specCopy)
	})
}
