// extensions.go — Parses and applies the --ext file extension filter.
package discovery

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ParseExtensionFilter parses a comma-separated extension list (".go,ts") into a lowercase allowlist.
// An empty value yields an empty allowlist, which matches every file.
func ParseExtensionFilter(raw string) (map[string]bool, error) {
	filter := map[string]bool{}
	value := strings.TrimSpace(raw)
	if value == "" {
		return filter, nil
	}

	for _, token := range strings.Split(value, ",") {
		normalized := strings.ToLower(strings.TrimSpace(token))
		if normalized == "" {
			continue
		}
		if !strings.HasPrefix(normalized, ".") {
			normalized = "." + normalized
		}
		if normalized == "." {
			continue
		}
		filter[normalized] = true
	}

	if len(filter) == 0 {
		return nil, fmt.Errorf("invalid --ext %q (expected value like .go or ts)", raw)
	}
	return filter, nil
}

// FilterByExtensions keeps only paths whose extension is in allowlist; an empty allowlist keeps all.
func FilterByExtensions(paths []string, allowlist map[string]bool) []string {
	if len(allowlist) == 0 {
		return paths
	}
	filtered := make([]string, 0, len(paths))
	for _, pathValue := range paths {
		ext := strings.ToLower(filepath.Ext(pathValue))
		if allowlist[ext] {
			filtered = append(filtered, pathValue)
		}
	}
	return filtered
}
