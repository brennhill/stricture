// adapter_parity_test.go — Integration checks for inspect language parity.
//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type inspectResult struct {
	Path     string `json:"Path"`
	Language string `json:"Language"`
}

func TestInspectDetectsLanguageByExtension(t *testing.T) {
	tmp := t.TempDir()
	cases := []struct {
		fileName string
		expected string
		content  string
	}{
		{fileName: "a.go", expected: "go", content: "package main\n"},
		{fileName: "a.ts", expected: "typescript", content: "export const x = 1;\n"},
		{fileName: "a.py", expected: "python", content: "x = 1\n"},
		{fileName: "A.java", expected: "java", content: "class A {}\n"},
	}

	for _, tc := range cases {
		t.Run(tc.fileName, func(t *testing.T) {
			p := filepath.Join(tmp, tc.fileName)
			if err := os.WriteFile(p, []byte(tc.content), 0o644); err != nil {
				t.Fatalf("write fixture: %v", err)
			}

			stdout, stderr, code := run(t, "inspect", p)
			if code != 0 {
				t.Fatalf("inspect exit code = %d, stderr=%q", code, stderr)
			}

			var got inspectResult
			if err := json.Unmarshal([]byte(stdout), &got); err != nil {
				t.Fatalf("inspect output must be valid JSON: %v\noutput=%s", err, stdout)
			}
			if got.Language != tc.expected {
				t.Fatalf("language = %q, want %q", got.Language, tc.expected)
			}
		})
	}
}
