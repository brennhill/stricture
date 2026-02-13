// adapter_test.go — Tests for the Java adapter.
package java

import (
	"testing"

	"github.com/stricture/stricture/internal/adapter"
)

func TestAdapterMetadata(t *testing.T) {
	a := &Adapter{}
	if a.Name() != "java" {
		t.Fatalf("name = %q, want java", a.Name())
	}
	ext := a.Extensions()
	if len(ext) != 1 || ext[0] != ".java" {
		t.Fatalf("unexpected extensions: %v", ext)
	}
}

func TestAdapterParse(t *testing.T) {
	a := &Adapter{}
	source := []byte("public class UserService {}\n")
	parsed, err := a.Parse("service/UserService.java", source, adapter.AdapterConfig{})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if parsed.Language != "java" {
		t.Fatalf("language = %q, want java", parsed.Language)
	}
	if len(parsed.Classes) != 1 || parsed.Classes[0].Name != "UserService" {
		t.Fatalf("unexpected classes: %+v", parsed.Classes)
	}
}
