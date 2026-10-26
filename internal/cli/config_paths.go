// config_paths.go — Config and plugin path resolution.
package cli

import (
	"os"
	"path/filepath"
	"strings"
)

func resolveConfigPath(configPath string) string {
	if strings.TrimSpace(configPath) == "" || filepath.IsAbs(configPath) {
		return configPath
	}

	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	wd, err := os.Getwd()
	if err != nil {
		return configPath
	}

	current := wd
	for {
		candidate := filepath.Join(current, configPath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return configPath
}

func resolvePluginPaths(configPath string, pluginPaths []string) []string {
	resolved := make([]string, 0, len(pluginPaths))
	configDir := filepath.Dir(configPath)
	for _, pathValue := range pluginPaths {
		p := strings.TrimSpace(pathValue)
		if p == "" {
			continue
		}
		if filepath.IsAbs(p) {
			resolved = append(resolved, p)
			continue
		}
		if strings.Contains(p, "://") {
			resolved = append(resolved, p)
			continue
		}
		resolved = append(resolved, filepath.Join(configDir, p))
	}
	return resolved
}
