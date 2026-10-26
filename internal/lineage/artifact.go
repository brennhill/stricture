// artifact.go - Lineage artifact collection and persistence.
package lineage

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Artifact is the normalized lineage graph emitted for CI/CD drift detection.
type Artifact struct {
	SchemaVersion string       `json:"schema_version"`
	Fields        []Annotation `json:"fields"`
	Overrides     []Override   `json:"overrides,omitempty"`
}

var sourceFileExtensions = map[string]bool{
	".go":   true,
	".ts":   true,
	".tsx":  true,
	".js":   true,
	".jsx":  true,
	".py":   true,
	".java": true,
	".kt":   true,
}

// Collect scans the given paths and builds a deterministic lineage artifact.
func Collect(paths []string) (Artifact, []ParseError, error) {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	acc := &collection{
		fields:    make([]Annotation, 0),
		overrides: make([]Override, 0),
		errors:    make([]ParseError, 0),
	}
	for _, path := range paths {
		if err := acc.collectPath(path); err != nil {
			return Artifact{}, nil, err
		}
	}

	sortCollection(acc)
	return Artifact{
		SchemaVersion: "1",
		Fields:        acc.fields,
		Overrides:     acc.overrides,
	}, acc.errors, nil
}

// collection accumulates parse results across files.
type collection struct {
	fields    []Annotation
	overrides []Override
	errors    []ParseError
}

func (c *collection) collectPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("collect lineage: stat %s: %w", path, err)
	}
	if !info.IsDir() {
		return c.addFile(path)
	}
	if err := filepath.WalkDir(path, c.walk); err != nil {
		return fmt.Errorf("collect lineage: walk %s: %w", path, err)
	}
	return nil
}

func (c *collection) walk(filePath string, d fs.DirEntry, walkErr error) error {
	if walkErr != nil {
		return walkErr
	}
	if d.IsDir() {
		if isSkippedDir(d.Name()) {
			return filepath.SkipDir
		}
		return nil
	}
	return c.addFile(filePath)
}

func isSkippedDir(name string) bool {
	return name == ".git" || name == "node_modules" || name == "bin"
}

// addFile parses one file when it has a supported source extension.
func (c *collection) addFile(path string) error {
	if !isSourceFile(path) {
		return nil
	}
	collected, collectedOverrides, parseErrs, readErr := collectFile(path)
	if readErr != nil {
		return readErr
	}
	c.fields = append(c.fields, collected...)
	c.overrides = append(c.overrides, collectedOverrides...)
	c.errors = append(c.errors, parseErrs...)
	return nil
}

func sortCollection(c *collection) {
	fields, overrides, errors := c.fields, c.overrides, c.errors
	sort.Slice(fields, func(i, j int) bool {
		if fields[i].FieldID != fields[j].FieldID {
			return fields[i].FieldID < fields[j].FieldID
		}
		if fields[i].FilePath != fields[j].FilePath {
			return fields[i].FilePath < fields[j].FilePath
		}
		return fields[i].Line < fields[j].Line
	})

	for i := range fields {
		sort.Slice(fields[i].Sources, func(a, b int) bool {
			return sourceIdentity(fields[i].Sources[a]) < sourceIdentity(fields[i].Sources[b])
		})
	}

	sort.Slice(overrides, func(i, j int) bool {
		if overrides[i].FieldID != overrides[j].FieldID {
			return overrides[i].FieldID < overrides[j].FieldID
		}
		if overrides[i].ChangeType != overrides[j].ChangeType {
			return overrides[i].ChangeType < overrides[j].ChangeType
		}
		if overrides[i].Expires != overrides[j].Expires {
			return overrides[i].Expires < overrides[j].Expires
		}
		if overrides[i].FilePath != overrides[j].FilePath {
			return overrides[i].FilePath < overrides[j].FilePath
		}
		return overrides[i].Line < overrides[j].Line
	})

	sort.Slice(errors, func(i, j int) bool {
		if errors[i].FilePath != errors[j].FilePath {
			return errors[i].FilePath < errors[j].FilePath
		}
		return errors[i].Line < errors[j].Line
	})
}

// LoadArtifact reads a lineage artifact JSON file.
func LoadArtifact(path string) (Artifact, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Artifact{}, fmt.Errorf("load lineage artifact: %w", err)
	}
	var artifact Artifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return Artifact{}, fmt.Errorf("parse lineage artifact: %w", err)
	}
	if artifact.SchemaVersion == "" {
		artifact.SchemaVersion = "1"
	}
	return artifact, nil
}

// WriteArtifact writes a lineage artifact JSON file with deterministic formatting.
func WriteArtifact(path string, artifact Artifact) error {
	if artifact.SchemaVersion == "" {
		artifact.SchemaVersion = "1"
	}
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal lineage artifact: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir artifact dir: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write lineage artifact: %w", err)
	}
	return nil
}

func collectFile(path string) ([]Annotation, []Override, []ParseError, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read %s: %w", path, err)
	}

	annotations, overrides, parseErrs := ParseWithOverrides(content)
	canonicalPath := filepath.ToSlash(path)
	for i := range annotations {
		annotations[i].FilePath = canonicalPath
	}
	for i := range overrides {
		overrides[i].FilePath = canonicalPath
	}
	for i := range parseErrs {
		parseErrs[i].FilePath = canonicalPath
	}

	return annotations, overrides, parseErrs, nil
}

func isSourceFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return sourceFileExtensions[ext]
}

func sourceIdentity(src SourceRef) string {
	return strings.Join([]string{src.Kind, src.Target, src.Path, src.Scope}, "|")
}
