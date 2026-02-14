package server

import "testing"

func TestLoadConfigFromEnvDefaults(t *testing.T) {
	t.Setenv("STRICTURE_SERVER_ADDR", "")
	t.Setenv("STRICTURE_SERVER_DATA_DIR", "")
	t.Setenv("STRICTURE_SERVER_INGEST_TOKEN", "")
	t.Setenv("STRICTURE_SERVER_STORAGE_DRIVER", "")
	t.Setenv("STRICTURE_SERVER_OBJECT_BUCKET", "")
	t.Setenv("STRICTURE_SERVER_OBJECT_PREFIX", "")
	t.Setenv("STRICTURE_SERVER_AUTH_MODE", "")

	cfg := LoadConfigFromEnv()
	if cfg.Addr != ":8085" {
		t.Fatalf("expected default addr :8085, got %q", cfg.Addr)
	}
	if cfg.DataDir != ".stricture-server-data" {
		t.Fatalf("expected default data dir, got %q", cfg.DataDir)
	}
	if cfg.IngestToken != "" {
		t.Fatalf("expected empty token, got %q", cfg.IngestToken)
	}
	if cfg.StorageDriver != "fs" {
		t.Fatalf("expected default storage driver fs, got %q", cfg.StorageDriver)
	}
	if cfg.ObjectPrefix != "stricture" {
		t.Fatalf("expected default object prefix stricture, got %q", cfg.ObjectPrefix)
	}
	if cfg.AuthMode != "none" {
		t.Fatalf("expected default auth mode none, got %q", cfg.AuthMode)
	}
}

func TestLoadConfigFromEnvOverrides(t *testing.T) {
	t.Setenv("STRICTURE_SERVER_ADDR", "127.0.0.1:9091")
	t.Setenv("STRICTURE_SERVER_DATA_DIR", "/tmp/stricture-server")
	t.Setenv("STRICTURE_SERVER_INGEST_TOKEN", "  secret-token  ")
	t.Setenv("STRICTURE_SERVER_STORAGE_DRIVER", "r2")
	t.Setenv("STRICTURE_SERVER_OBJECT_BUCKET", "lineage-bucket")
	t.Setenv("STRICTURE_SERVER_OBJECT_PREFIX", "prod/lineage")
	t.Setenv("STRICTURE_SERVER_AUTH_MODE", "token")

	cfg := LoadConfigFromEnv()
	if cfg.Addr != "127.0.0.1:9091" {
		t.Fatalf("expected addr override, got %q", cfg.Addr)
	}
	if cfg.DataDir != "/tmp/stricture-server" {
		t.Fatalf("expected data dir override, got %q", cfg.DataDir)
	}
	if cfg.IngestToken != "secret-token" {
		t.Fatalf("expected token trim, got %q", cfg.IngestToken)
	}
	if cfg.StorageDriver != "r2" {
		t.Fatalf("expected storage driver override, got %q", cfg.StorageDriver)
	}
	if cfg.ObjectBucket != "lineage-bucket" {
		t.Fatalf("expected object bucket override, got %q", cfg.ObjectBucket)
	}
	if cfg.ObjectPrefix != "prod/lineage" {
		t.Fatalf("expected object prefix override, got %q", cfg.ObjectPrefix)
	}
	if cfg.AuthMode != "token" {
		t.Fatalf("expected auth mode token, got %q", cfg.AuthMode)
	}
}
