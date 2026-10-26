// git_symlink_test.go — Tests git scoping, path keys, and symlink safety using real temp repos.
package discovery

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestGitScopedFileSet(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	write(t, filepath.Join(dir, "committed.go"), "package a\n")
	write(t, filepath.Join(dir, "staged.go"), "package a\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "init")

	write(t, filepath.Join(dir, "committed.go"), "package a // edited\n")
	write(t, filepath.Join(dir, "staged.go"), "package a // staged edit\n")
	git(t, dir, "add", "staged.go")
	write(t, filepath.Join(dir, "untracked.go"), "package a\n")
	chdir(t, dir)

	key := func(name string) string { return PathKey(dir, name) }

	changed, err := GitScopedFileSet(true, false)
	if err != nil {
		t.Fatalf("changed: %v", err)
	}
	for _, name := range []string{"committed.go", "staged.go", "untracked.go"} {
		if !changed[key(name)] {
			t.Errorf("changed set missing %s: %v", name, changed)
		}
	}

	staged, err := GitScopedFileSet(false, true)
	if err != nil {
		t.Fatalf("staged: %v", err)
	}
	if !staged[key("staged.go")] || staged[key("committed.go")] || staged[key("untracked.go")] {
		t.Errorf("staged set should contain only staged.go: %v", staged)
	}

	none, err := GitScopedFileSet(false, false)
	if err != nil || len(none) != 0 {
		t.Errorf("no scope should yield an empty set, got %v err=%v", none, err)
	}
}

func TestGitScopedFileSetOutsideRepoFails(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	chdir(t, dir)
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	if _, err := GitScopedFileSet(true, false); err == nil || !strings.Contains(err.Error(), "requires a git repository") {
		t.Fatalf("error = %v, want git repository requirement", err)
	}
}

func TestPathKeyResolvesRelativeAndSymlinks(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "real")
	write(t, filepath.Join(target, "f.go"), "package a\n")
	link := filepath.Join(dir, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
	if PathKey(dir, "link/f.go") != PathKey(dir, "real/f.go") {
		t.Errorf("symlinked and real paths should share a key")
	}
	if PathKey("/ignored", filepath.Join(target, "f.go")) != PathKey(dir, "real/f.go") {
		t.Errorf("absolute paths should ignore the base dir")
	}
}

func TestCollectFilesSkipsSymlinksEscapingProject(t *testing.T) {
	project := t.TempDir()
	outside := t.TempDir()
	write(t, filepath.Join(outside, "secret.go"), "package s\n")
	write(t, filepath.Join(project, "ok.go"), "package ok\n")
	if err := os.Symlink(filepath.Join(outside, "secret.go"), filepath.Join(project, "escape.go")); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
	if err := os.Symlink(filepath.Join(project, "missing.go"), filepath.Join(project, "broken.go")); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
	chdir(t, project)

	got, err := CollectFiles([]string{"."})
	if err != nil {
		t.Fatalf("CollectFiles: %v", err)
	}
	if len(got) != 1 || got[0] != "ok.go" {
		t.Fatalf("CollectFiles = %v, want only ok.go (escaping and broken symlinks skipped)", got)
	}
}

func TestIsSourceFileAndGenerated(t *testing.T) {
	for path, want := range map[string]bool{
		"a.go": true, "a.TS": true, "a.kt": true, "a.pb.go": false, "x.generated.ts": false, "a.txt": false, "a.pb.ts": false,
	} {
		if got := IsSourceFile(path); got != want {
			t.Errorf("IsSourceFile(%q) = %v, want %v", path, got, want)
		}
	}
}
