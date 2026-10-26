// discovery_test.go — Tests extension filtering, directory skipping, and source file classification.
package discovery

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseExtensionFilter(t *testing.T) {
	t.Parallel()

	got, err := ParseExtensionFilter("go, .TS,py")
	if err != nil {
		t.Fatalf("parseExtensionFilter returned error: %v", err)
	}
	want := map[string]bool{
		".go": true,
		".ts": true,
		".py": true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseExtensionFilter result = %#v, want %#v", got, want)
	}
}

func TestParseExtensionFilterInvalid(t *testing.T) {
	t.Parallel()

	if _, err := ParseExtensionFilter(" , , . "); err == nil {
		t.Fatalf("expected invalid extension filter error")
	}
}

func TestFilterByExtensions(t *testing.T) {
	t.Parallel()

	paths := []string{"a.go", "b.ts", "c.py", "README.md"}
	allowlist := map[string]bool{".go": true, ".py": true}
	got := FilterByExtensions(paths, allowlist)
	want := []string{"a.go", "c.py"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filterFilePathsByExtensions = %#v, want %#v", got, want)
	}
}

func TestIsPathWithinRoot(t *testing.T) {
	t.Parallel()

	root := filepath.FromSlash("/workspace/project")
	inside := filepath.FromSlash("/workspace/project/src/main.go")
	outside := filepath.FromSlash("/workspace/other/file.go")
	if !isPathWithinRoot(inside, root) {
		t.Fatalf("inside path should be within root")
	}
	if isPathWithinRoot(outside, root) {
		t.Fatalf("outside path should not be within root")
	}
}

func TestShouldSkipDir(t *testing.T) {
	t.Parallel()

	if !ShouldSkipDir("node_modules") {
		t.Fatalf("node_modules should be skipped")
	}
	if !ShouldSkipDir("tests/fixtures/sample") {
		t.Fatalf("tests/fixtures path should be skipped")
	}
	if ShouldSkipDir("src") {
		t.Fatalf("src should not be skipped")
	}
}

func TestSourceFileHeuristics(t *testing.T) {
	t.Parallel()

	if IsSourceFile("x.generated.ts") {
		t.Fatalf("generated file should not be lint source")
	}
	if IsSourceFile("x.pb.go") {
		t.Fatalf("protobuf-generated file should not be lint source")
	}
	if !IsSourceFile("x.go") {
		t.Fatalf("go file should be lint source")
	}
	if !IsSourceFile("x.ts") {
		t.Fatalf("ts file should be lint source")
	}
	if IsSourceFile("x.md") {
		t.Fatalf("markdown should not be lint source")
	}
}

func TestCollectFilesSortsDedupesAndSkips(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()
	files := map[string]string{
		"b.go":                 "package b\n",
		"a.ts":                 "export {}\n",
		"sub/c.py":             "x = 1\n",
		"node_modules/skip.js": "x\n",
		"gen.pb.go":            "package g\n",
		"notes.txt":            "n\n",
	}
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := CollectFiles([]string{".", "b.go", " ", "sub"})
	if err != nil {
		t.Fatalf("CollectFiles error: %v", err)
	}
	want := []string{"a.ts", "b.go", "sub/c.py"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CollectFiles = %v, want %v", got, want)
	}

	if _, err := CollectFiles([]string{"missing.go"}); err == nil {
		t.Fatal("CollectFiles should fail for a missing path")
	}
}
