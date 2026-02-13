// manifest.go — Manifest model and loading helpers.
package manifest

import (
	"fmt"
	"os"
	"strings"

	"github.com/stricture/stricture/internal/model"
	"gopkg.in/yaml.v3"
)

// Contract describes a declared contract entry in .stricture-manifest.yml.
type Contract struct {
	ID       string `yaml:"id"`
	Endpoint string `yaml:"endpoint"`
	Method   string `yaml:"method"`
}

// Manifest is the top-level manifest declaration.
type Manifest struct {
	ManifestVersion string     `yaml:"manifest_version"`
	Contracts       []Contract `yaml:"contracts"`
}

// Parse parses and validates manifest bytes.
func Parse(data []byte) (Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("parse manifest yaml: %w", model.ErrManifestInvalid)
	}
	if err := Validate(m); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

// Load reads a manifest from disk and validates it.
func Load(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Manifest{}, fmt.Errorf("load manifest file: %w", model.ErrManifestNotFound)
		}
		return Manifest{}, fmt.Errorf("load manifest file: %w", err)
	}
	return Parse(data)
}

// Validate ensures required manifest fields exist.
func Validate(m Manifest) error {
	if strings.TrimSpace(m.ManifestVersion) == "" {
		return fmt.Errorf("validate manifest: %w", model.ErrManifestInvalid)
	}
	if len(m.Contracts) == 0 {
		return fmt.Errorf("validate manifest: %w", model.ErrManifestInvalid)
	}
	for _, c := range m.Contracts {
		if strings.TrimSpace(c.ID) == "" {
			return fmt.Errorf("validate manifest: %w", model.ErrManifestInvalid)
		}
	}
	return nil
}
