package dataset

import "sync"

var builtinRegistrars []func(*Registry)

type Registry struct {
	mu    sync.RWMutex
	specs map[string]Spec
}

func NewRegistry() *Registry {
	return &Registry{
		specs: make(map[string]Spec),
	}
}

func (r *Registry) Register(spec Spec) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.specs[spec.Name] = spec
}

func (r *Registry) Get(name string) (Spec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	spec, ok := r.specs[name]
	return spec, ok
}

func (r *Registry) List() []Spec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Spec, 0, len(r.specs))
	for _, spec := range r.specs {
		out = append(out, spec)
	}
	return out
}

func RegisterBuiltins(r *Registry) {
	for _, fn := range builtinRegistrars {
		fn(r)
	}
}

func AddBuiltinRegistrar(fn func(*Registry)) {
	builtinRegistrars = append(builtinRegistrars, fn)
}
