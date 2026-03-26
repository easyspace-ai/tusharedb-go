package builtin

import "github.com/easyspace-ai/tusharedb-go/internal/dataset"

func datasetRegister(spec dataset.Spec) {
	specCopy := spec
	dataset.AddBuiltinRegistrar(func(r *dataset.Registry) {
		r.Register(specCopy)
	})
}
