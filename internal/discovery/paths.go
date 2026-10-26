// paths.go — Path classification, normalization, and symlink safety for discovery.
package discovery

import (
	"os"
	"path/filepath"
	"strings"
)

// ShouldSkipDir reports whether a directory is excluded from discovery.
func ShouldSkipDir(dir string) bool {
	base := filepath.Base(dir)
	switch base {
	case ".git", "node_modules", "bin", ".stricture-cache", "docs", "tests":
		return true
	}

	normalized := filepath.ToSlash(dir)
	return strings.Contains(normalized, "tests/fixtures") || strings.Contains(normalized, "tests/benchmark")
}

// IsSourceFile reports whether path is a non-generated source file in a supported language.
func IsSourceFile(path string) bool {
	if isGeneratedSourceFile(path) {
		return false
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".java", ".kt", ".rs":
		return true
	default:
		return false
	}
}

func isGeneratedSourceFile(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	return strings.Contains(name, ".generated.") ||
		strings.HasSuffix(name, ".pb.go") ||
		strings.HasSuffix(name, ".pb.ts")
}

// PathKey returns a canonical absolute, symlink-resolved, slash-separated key for pathValue
// interpreted relative to baseDir.
func PathKey(baseDir string, pathValue string) string {
	candidate := pathValue
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(baseDir, candidate)
	}
	abs, err := filepath.Abs(candidate)
	if err == nil {
		candidate = abs
	}
	if resolved, err := filepath.EvalSymlinks(candidate); err == nil {
		candidate = resolved
	}
	return filepath.ToSlash(filepath.Clean(candidate))
}

func projectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	abs, err := filepath.Abs(wd)
	if err != nil {
		return filepath.Clean(wd)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err == nil {
		return filepath.Clean(resolved)
	}
	return filepath.Clean(abs)
}

func symlinkResolvesOutsideProject(pathValue string, root string) (bool, error) {
	if strings.TrimSpace(root) == "" {
		return false, nil
	}

	info, err := os.Lstat(pathValue)
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}

	resolved, err := filepath.EvalSymlinks(pathValue)
	if err != nil {
		// Broken or inaccessible symlink should not be linted.
		//nolint:nilerr // Returning nil error is intentional because this is a soft-skip condition.
		return true, nil
	}
	absResolved, err := filepath.Abs(resolved)
	if err != nil {
		return false, err
	}
	return !isPathWithinRoot(absResolved, root), nil
}

func isPathWithinRoot(pathValue string, root string) bool {
	rel, err := filepath.Rel(root, pathValue)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	if rel == ".." {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
