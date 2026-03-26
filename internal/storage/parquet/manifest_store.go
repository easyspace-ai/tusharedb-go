package parquet

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type ManifestStore struct {
	path string
	mu   sync.Mutex
}

func NewManifestStore(dataDir string) *ManifestStore {
	return &ManifestStore{
		path: filepath.Join(dataDir, "meta", "manifests.json"),
	}
}

func (s *ManifestStore) Append(dataset string, file string) error {
	if file == "" {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	manifests := map[string]Manifest{}
	if b, err := os.ReadFile(s.path); err == nil {
		_ = json.Unmarshal(b, &manifests)
	} else if !os.IsNotExist(err) {
		return err
	}

	manifest := manifests[dataset]
	manifest.Dataset = dataset
	manifest.Files = append(manifest.Files, file)
	manifests[dataset] = manifest

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(manifests, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o644)
}
