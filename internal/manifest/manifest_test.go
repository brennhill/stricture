// manifest_test.go — Tests for manifest parsing and validation.
package manifest

import (
	"errors"
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
