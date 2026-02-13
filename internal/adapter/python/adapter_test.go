// adapter_test.go — Tests for the Python adapter.
package python

import (
	"testing"

	"github.com/stricture/stricture/internal/adapter"
)

func TestAdapterMetadata(t *testing.T) {
	a := &Adapter{}
	if a.Name() != "python" {
		t.Fatalf("name = %q, want python", a.Name())
	}
	ext := a.Extensions()
	if len(ext) != 1 || ext[0] != ".py" {
		t.Fatalf("unexpected extensions: %v", ext)
	}
}

func TestAdapterParse(t *testing.T) {
	a := &Adapter{}
	source := []byte("def create_user(name):\n    return name\n")
	parsed, err := a.Parse("service/user.py", source, adapter.AdapterConfig{})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if parsed.Language != "python" {
		t.Fatalf("language = %q, want python", parsed.Language)
	}
	if len(parsed.Functions) != 1 || parsed.Functions[0].Name != "create_user" {
		t.Fatalf("unexpected functions: %+v", parsed.Functions)
	}
}
