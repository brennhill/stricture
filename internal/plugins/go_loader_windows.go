//go:build windows

package plugins

import (
	"fmt"

	"github.com/stricture/stricture/internal/model"
)

func loadGoPluginRules(pathValue string) ([]model.Rule, error) {
	return nil, fmt.Errorf("go plugins are not supported on windows: %s", pathValue)
}
