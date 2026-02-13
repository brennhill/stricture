// inspect_errors_test.go — Integration checks for inspect error handling.
//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectUnsupportedExtensionExitsTwo(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "script.rb")
	if err := os.WriteFile(target, []byte("puts 'hi'\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, stderr, code := run(t, "inspect", target)
	if code != 2 {
		t.Fatalf("inspect unsupported extension exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, ".rb") {
		t.Fatalf("stderr should mention extension, got %q", stderr)
	}
}

func TestInspectBinaryFileExitsTwo(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "image.png")
	if err := os.WriteFile(target, []byte{0x89, 0x50, 0x4e, 0x47, 0x00, 0x01}, 0o644); err != nil {
		t.Fatalf("write binary file: %v", err)
	}

	_, stderr, code := run(t, "inspect", target)
	if code != 2 {
		t.Fatalf("inspect binary file exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "binary") {
		t.Fatalf("stderr should mention binary file, got %q", stderr)
	}
}
