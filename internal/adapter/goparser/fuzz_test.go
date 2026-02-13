// fuzz_test.go — Fuzz tests for the Go language adapter.
//
// These tests feed random/mutated Go source to the parser and verify it
// never panics, never returns corrupt data, and always returns either a
// valid UnifiedFileModel or a clean error.
//
// Run: go test -fuzz=FuzzGoParser -fuzztime=60s ./internal/adapter/goparser/...

package goparser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stricture/stricture/internal/adapter"
	"github.com/stricture/stricture/internal/model"
)

// FuzzGoParser feeds random bytes as Go source to the parser.
// It must never panic. It must return either a valid model or an error.
func FuzzGoParser(f *testing.F) {
	// Seed corpus: valid Go files of increasing complexity
	seeds := []string{
		// Minimal valid Go file
		`package main`,

		// With function
		`package main

func hello() string {
	return "world"
}`,

		// With struct and json tags
		`package main

type User struct {
	ID   int    ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}`,

		// With imports
		`package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("hello")
	http.ListenAndServe(":8080", nil)
}`,

		// Test file
		`package main

import "testing"

func TestHello(t *testing.T) {
	result := hello()
	if result != "world" {
		t.Errorf("expected world, got %s", result)
	}
}`,

		// With error handling
		`package main

import "fmt"

func risky() (string, error) {
	return "", fmt.Errorf("parse input: unexpected EOF. Check input format.")
}`,

		// With interface
		`package main

type Reader interface {
	Read(p []byte) (n int, err error)
}`,

		// Edge case: empty file
		``,

		// Edge case: just a comment
		`// just a comment`,

		// Edge case: syntax error
		`package main
func broken( {`,

		// Edge case: unicode
		`package main
// 日本語のコメント
func こんにちは() string { return "世界" }`,
	}

	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		a := &GoAdapter{}
		cfg := adapter.AdapterConfig{}

		// Must not panic — this is the primary assertion.
		// A recovered panic will fail the fuzz test automatically.
		result, err := a.Parse("/fuzz/test.go", data, cfg)

		if err != nil {
			// Error is fine — parser rejected invalid input.
			// But the error must be non-nil and the result must be nil.
			if result != nil {
				t.Error("Parse returned both result and error")
			}
			return
		}

		// If no error, result must be valid:
		if result == nil {
			t.Fatal("Parse returned nil result with nil error")
		}

		// Basic invariants that must hold for any successful parse:
		if result.Language != "go" {
			t.Errorf("Language = %q, want 'go'", result.Language)
		}
		if result.Path != "/fuzz/test.go" {
			t.Errorf("Path = %q, want '/fuzz/test.go'", result.Path)
		}

		// Functions must have names
		for i, fn := range result.Functions {
			if fn.Name == "" {
				t.Errorf("Function[%d] has empty name", i)
			}
		}

		// Types must have names and valid kinds
		validKinds := map[string]bool{
			"struct": true, "interface": true, "type": true, "enum": true,
		}
		for i, typ := range result.Types {
			if typ.Name == "" {
				t.Errorf("Type[%d] has empty name", i)
			}
			if !validKinds[typ.Kind] {
				t.Errorf("Type[%d] has invalid kind %q", i, typ.Kind)
			}
		}

		// Line numbers must be positive
		for i, imp := range result.Imports {
			if imp.StartLine < 1 {
				t.Errorf("Import[%d] has invalid line %d", i, imp.StartLine)
			}
		}
	})
}

// FuzzGoAdapterIsTestFile checks that IsTestFile never panics.
func FuzzGoAdapterIsTestFile(f *testing.F) {
	f.Add("/project/main.go")
	f.Add("/project/main_test.go")
	f.Add("")
	f.Add("/")
	f.Add("test.go")
	f.Add("_test.go")

	f.Fuzz(func(t *testing.T, path string) {
		a := &GoAdapter{}
		// Must not panic. Result doesn't matter.
		_ = a.IsTestFile(path)
	})
}

// Placeholder — actual adapter not implemented yet.
// This struct exists so the fuzz tests compile.
type GoAdapter struct{}

func (a *GoAdapter) Language() string            { return "go" }
func (a *GoAdapter) Extensions() []string        { return []string{".go"} }
func (a *GoAdapter) IsTestFile(path string) bool { return strings.HasSuffix(path, "_test.go") }
func (a *GoAdapter) Parse(path string, source []byte, config adapter.AdapterConfig) (*model.UnifiedFileModel, error) {
	return nil, fmt.Errorf("parse Go file: %w", model.ErrParseFailure)
}
func (a *GoAdapter) ResolveImport(importPath string, fromFile string) string { return "" }
