// adapter_test.go — Tests for the TypeScript adapter.
package typescript

import (
	"testing"

	"github.com/stricture/stricture/internal/adapter"
)

func TestAdapterMetadata(t *testing.T) {
	a := &Adapter{}
	if a.Name() != "typescript" {
		t.Fatalf("name = %q, want typescript", a.Name())
	}
	ext := a.Extensions()
	if len(ext) != 4 {
		t.Fatalf("extensions len = %d, want 4", len(ext))
	}
}

func TestAdapterIsTestFile(t *testing.T) {
	a := &Adapter{}
	if !a.IsTestFile("user.test.ts") {
		t.Fatal("expected user.test.ts to be test file")
	}
	if a.IsTestFile("user.ts") {
		t.Fatal("expected user.ts to be non-test file")
	}
}

func TestAdapterParse(t *testing.T) {
	a := &Adapter{}
	source := []byte("export function CreateUser() {}\n")
	parsed, err := a.Parse("api/user.ts", source, adapter.AdapterConfig{})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if parsed.Language != "typescript" {
		t.Fatalf("language = %q, want typescript", parsed.Language)
	}
	if parsed.Path != "api/user.ts" {
		t.Fatalf("path = %q, want api/user.ts", parsed.Path)
	}
	if parsed.LineCount != 2 {
		t.Fatalf("line count = %d, want 2", parsed.LineCount)
	}
	if len(parsed.Exports) != 1 || parsed.Exports[0].Name != "CreateUser" {
		t.Fatalf("unexpected exports: %+v", parsed.Exports)
	}
}
