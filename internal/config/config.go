package config

import (
	"fmt"
	"path/filepath"
)

type Config struct {
	Token      string
	DataDir    string
	DuckDBPath string
	TempDir    string
	LogLevel   string
}

func Normalize(cfg Config) (Config, error) {
	if cfg.DataDir == "" {
		cfg.DataDir = "./data"
	}
	if cfg.TempDir == "" {
		cfg.TempDir = filepath.Join(cfg.DataDir, "tmp")
	}
	if cfg.DuckDBPath == "" {
		cfg.DuckDBPath = filepath.Join(cfg.DataDir, "duckdb", "tusharedb.duckdb")
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.Token == "" {
		return Config{}, fmt.Errorf("token is required")
	}
	return cfg, nil
}
