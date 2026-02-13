//go:build !windows

package plugins

import (
	"fmt"
	"plugin"

	"github.com/stricture/stricture/internal/model"
	plugapi "github.com/stricture/stricture/pkg/rule"
)

func loadGoPluginRules(pathValue string) ([]model.Rule, error) {
	plug, err := plugin.Open(pathValue)
	if err != nil {
		return nil, fmt.Errorf("open plugin %s: %w", pathValue, err)
	}
	sym, err := plug.Lookup("Rule")
	if err != nil {
		return nil, fmt.Errorf("plugin %s missing exported symbol Rule: %w", pathValue, err)
	}

	switch v := sym.(type) {
	case *plugapi.Definition:
		return []model.Rule{&goPluginRule{definition: v}}, nil
	default:
		return nil, fmt.Errorf("plugin %s Rule symbol must be *rule.Definition, got %T", pathValue, sym)
	}
}
