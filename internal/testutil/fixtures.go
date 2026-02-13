// fixtures.go — Test fixture loading from validation set directories.
package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// FixtureSet holds validation set fixtures grouped by status.
type FixtureSet struct {
	// Perfect holds files that should produce zero violations.
	Perfect []FixtureFile
	// Bugs holds files organized by bug level (B01, B02, etc.) that should trigger violations.
	Bugs map[string][]FixtureFile
}

// FixtureFile represents a loaded test fixture.
type FixtureFile struct {
	Path    string
	Content []byte
}

// LoadFixtureSet loads a validation set from tests/fixtures/{name}/.
// The directory must contain:
//   - pass/ or perfect/ — files that should not trigger violations
//   - fail-b*/ or bug-*/ — files that should trigger violations
func LoadFixtureSet(t *testing.T, name string) *FixtureSet {
	t.Helper()

	root := FindProjectRoot(t)
	dir := filepath.Join(root, "tests", "fixtures", name)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skipf("fixture set %q not found at %s", name, dir)
	}

	set := &FixtureSet{
		Bugs: make(map[string][]FixtureFile),
	}

	entries, err := os.ReadDir(dir)
	require.NoError(t, err, "reading fixture directory %s", dir)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subName := entry.Name()
		subDir := filepath.Join(dir, subName)

		files := loadFilesFromDir(t, subDir)

		switch {
		case subName == "pass" || subName == "perfect":
			set.Perfect = append(set.Perfect, files...)
		case strings.HasPrefix(subName, "fail-") || strings.HasPrefix(subName, "bug-"):
			// Extract bug level: "fail-b01" -> "B01", "bug-01" -> "01"
			key := strings.TrimPrefix(subName, "fail-")
			key = strings.TrimPrefix(key, "bug-")
			key = strings.ToUpper(key)
			set.Bugs[key] = append(set.Bugs[key], files...)
		}
	}

	return set
}

// LoadFixtureFile loads a single file from the fixture directory.
func LoadFixtureFile(t *testing.T, relativePath string) FixtureFile {
	t.Helper()

	root := FindProjectRoot(t)
	fullPath := filepath.Join(root, relativePath)

	content, err := os.ReadFile(fullPath)
	require.NoError(t, err, "loading fixture %s", fullPath)

	return FixtureFile{
		Path:    fullPath,
		Content: content,
	}
}

// FindProjectRoot walks up from the current working directory to find go.mod.
func FindProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err, "getting working directory")

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (no go.mod found)")
		}
		dir = parent
	}
}

func loadFilesFromDir(t *testing.T, dir string) []FixtureFile {
	t.Helper()

	var files []FixtureFile
	entries, err := os.ReadDir(dir)
	require.NoError(t, err, "reading directory %s", dir)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(path)
		require.NoError(t, err, "reading %s", path)
		files = append(files, FixtureFile{
			Path:    path,
			Content: content,
		})
	}
	return files
}
