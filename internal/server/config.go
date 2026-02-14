package server

import (
	"os"
	"strings"
)

// Config controls stricture-server runtime settings.
type Config struct {
	Addr          string
	DataDir       string
	IngestToken   string
	StorageDriver string
	ObjectBucket  string
	ObjectPrefix  string
	AuthMode      string
}

// LoadConfigFromEnv builds server config from environment variables.
func LoadConfigFromEnv() Config {
	cfg := Config{
		Addr:          ":8085",
		DataDir:       ".stricture-server-data",
		IngestToken:   strings.TrimSpace(os.Getenv("STRICTURE_SERVER_INGEST_TOKEN")),
		StorageDriver: "fs",
		ObjectPrefix:  "stricture",
	}

	if value := strings.TrimSpace(os.Getenv("STRICTURE_SERVER_ADDR")); value != "" {
		cfg.Addr = value
	}
	if value := strings.TrimSpace(os.Getenv("STRICTURE_SERVER_DATA_DIR")); value != "" {
		cfg.DataDir = value
	}
	if value := strings.TrimSpace(os.Getenv("STRICTURE_SERVER_STORAGE_DRIVER")); value != "" {
		cfg.StorageDriver = strings.ToLower(value)
	}
	if value := strings.TrimSpace(os.Getenv("STRICTURE_SERVER_OBJECT_BUCKET")); value != "" {
		cfg.ObjectBucket = value
	}
	if value := strings.TrimSpace(os.Getenv("STRICTURE_SERVER_OBJECT_PREFIX")); value != "" {
		cfg.ObjectPrefix = value
	}
	if value := strings.TrimSpace(os.Getenv("STRICTURE_SERVER_AUTH_MODE")); value != "" {
		cfg.AuthMode = strings.ToLower(value)
	}
	if cfg.AuthMode == "" {
		if cfg.IngestToken == "" {
			cfg.AuthMode = "none"
		} else {
			cfg.AuthMode = "token"
		}
	}
	return cfg
}
