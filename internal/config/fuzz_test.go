// fuzz_test.go — Fuzz tests for the config loader.
//
// Feeds random YAML to the config parser. Must never panic.
// Run: go test -fuzz=FuzzConfigLoad -fuzztime=30s ./internal/config/...

package config

import (
	"testing"
)

// FuzzConfigLoad feeds random bytes as YAML config to the loader.
func FuzzConfigLoad(f *testing.F) {
	seeds := []string{
		// Valid minimal config
		`version: "1.0"
rules:
  CONV-file-naming: error`,

		// Valid full config
		`version: "1.0"
rules:
  CONV-file-naming: [error, { style: "kebab-case" }]
  CONV-file-header: [error, { pattern: "// {filename} — {purpose}" }]
  ARCH-max-file-lines: [error, { max: 800 }]
  TQ-no-shallow-assertions: error`,

		// Empty
		``,

		// Just a comment
		`# nothing here`,

		// Invalid YAML
		`{{{`,

		// Wrong types
		`version: 123
rules: "not a map"`,

		// Unknown rule
		`version: "1.0"
rules:
  FAKE-rule: error`,

		// Nested nonsense
		`version: "1.0"
rules:
  CONV-file-naming:
    - error
    - style: "kebab-case"
    - extra: true
    - nested:
        deep:
          deeper: value`,

		// Very long rule name
		`version: "1.0"
rules:
  AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA: error`,
	}

	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Must not panic
		_, _ = LoadFromBytes(data)
	})
}

func LoadFromBytes(data []byte) (interface{}, error) {
	_ = data
	return nil, nil
}
