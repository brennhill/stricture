// collect.go — Walks paths and collects lintable source files.
package discovery

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CollectFiles expands files and directories into a sorted, de-duplicated list of lintable
// source files. Skipped directories, generated files, and symlinks resolving outside the
// working directory are excluded.
func CollectFiles(paths []string) ([]string, error) {
	c := &collector{root: projectRoot(), seen: map[string]bool{}}
	for _, raw := range paths {
		pathValue := strings.TrimSpace(raw)
		if pathValue == "" {
			continue
		}
		if err := c.addPath(pathValue); err != nil {
			return nil, err
		}
	}
	sort.Strings(c.files)
	return c.files, nil
}

type collector struct {
	root  string
	seen  map[string]bool
	files []string
}

func (c *collector) addPath(pathValue string) error {
	info, err := os.Stat(pathValue)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return c.addFile(pathValue)
	}
	return filepath.WalkDir(pathValue, c.visit)
}

func (c *collector) visit(current string, entry fs.DirEntry, walkErr error) error {
	if walkErr != nil {
		return walkErr
	}
	if entry.IsDir() {
		if ShouldSkipDir(current) {
			return filepath.SkipDir
		}
		return nil
	}
	return c.addFile(current)
}

func (c *collector) addFile(pathValue string) error {
	if !IsSourceFile(pathValue) {
		return nil
	}
	outside, err := symlinkResolvesOutsideProject(pathValue, c.root)
	if err != nil {
		return err
	}
	if outside {
		return nil
	}
	canonical := filepath.ToSlash(pathValue)
	if !c.seen[canonical] {
		c.seen[canonical] = true
		c.files = append(c.files, canonical)
	}
	return nil
}
