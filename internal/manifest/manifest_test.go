// manifest_test.go — Tests for manifest parsing and validation.
package manifest

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stricture/stricture/internal/model"
)

func TestParseValidManifest(t *testing.T) {
	data := []byte("manifest_version: v1\ncontracts:\n  - id: users.v1\n    endpoint: /users\n    method: GET\n")
	m, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if m.ManifestVersion != "v1" {
		t.Fatalf("manifest version = %q, want v1", m.ManifestVersion)
	}
	if len(m.Contracts) != 1 || m.Contracts[0].ID != "users.v1" {
		t.Fatalf("unexpected contracts: %+v", m.Contracts)
	}
}

func TestParseInvalidManifest(t *testing.T) {
	_, err := Parse([]byte("manifest_version: \"\"\ncontracts: []\n"))
	if !errors.Is(err, model.ErrManifestInvalid) {
		t.Fatalf("error = %v, want ErrManifestInvalid", err)
	}
}

func TestLoadMissingManifest(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.yml"))
	if !errors.Is(err, model.ErrManifestNotFound) {
		t.Fatalf("error = %v, want ErrManifestNotFound", err)
	}
}

func TestLoadUnreadablePathReturnsWrappedError(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "dir")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	_, err := Load(dir)
	if err == nil {
		t.Fatalf("expected load error for directory path")
	}
	if errors.Is(err, model.ErrManifestNotFound) {
		t.Fatalf("expected non-not-found error, got %v", err)
	}
}

func TestValidateManifestBranches(t *testing.T) {
	if err := Validate(Manifest{}); !errors.Is(err, model.ErrManifestInvalid) {
		t.Fatalf("validate empty manifest error = %v, want ErrManifestInvalid", err)
	}

	if err := Validate(Manifest{ManifestVersion: "v1"}); !errors.Is(err, model.ErrManifestInvalid) {
		t.Fatalf("validate empty contracts error = %v, want ErrManifestInvalid", err)
	}

	m := Manifest{
		ManifestVersion: "v1",
		Contracts: []Contract{
			{ID: "   ", Endpoint: "/users", Method: "GET"},
		},
	}
	if err := Validate(m); !errors.Is(err, model.ErrManifestInvalid) {
		t.Fatalf("validate empty contract id error = %v, want ErrManifestInvalid", err)
	}

	m.Contracts[0].ID = "users.v1"
	if err := Validate(m); err != nil {
		t.Fatalf("validate valid manifest returned error: %v", err)
	}
}
