package meta

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type DatasetCheckpoint struct {
	Dataset          string    `json:"dataset"`
	LastSyncedDate   string    `json:"last_synced_date"`
	LastSuccessfulAt time.Time `json:"last_successful_at"`
	SchemaVersion    string    `json:"schema_version"`
}

type CheckpointStore struct {
	path string
	mu   sync.RWMutex
	data map[string]DatasetCheckpoint
}

func NewCheckpointStore(path string) (*CheckpointStore, error) {
	store := &CheckpointStore{
		path: path,
		data: make(map[string]DatasetCheckpoint),
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *CheckpointStore) Get(dataset string) (DatasetCheckpoint, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp, ok := s.data[dataset]
	return cp, ok
}

func (s *CheckpointStore) Put(cp DatasetCheckpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[cp.Dataset] = cp
	return s.saveLocked()
}

func (s *CheckpointStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(b, &s.data)
}

func (s *CheckpointStore) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o644)
}
