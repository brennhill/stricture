// color.go — Terminal color decisions and severity colorization.
package cli

import (
	"os"
	"strings"
)

func shouldUseColor(forceColor bool, forceNoColor bool, outputPath string) bool {
	if forceNoColor {
		return false
	}
	if forceColor {
		return true
	}
	if strings.TrimSpace(outputPath) != "" {
		return false
	}
	term := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))
	if term == "" || term == "dumb" {
		return false
	}
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func colorizeSeverityLabel(rawSeverity string, label string, enabled bool) string {
	if !enabled {
		return label
	}
	switch strings.ToLower(strings.TrimSpace(rawSeverity)) {
	case "error":
		return "\x1b[31m" + label + "\x1b[0m"
	case "warn", "warning":
		return "\x1b[33m" + label + "\x1b[0m"
	default:
		return label
	}
}
