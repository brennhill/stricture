// repo_test.go — Enforces the architecture policy against the real repository.
package archtest

import (
	"os"
	"path/filepath"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("locate repo: %v. Run tests from inside the repository", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("locate repo: go.mod not found. Run tests from inside the repository")
		}
		dir = parent
	}
}

func loadRepo(t *testing.T) (Graph, map[string]int) {
	t.Helper()
	g, lines, err := LoadGraph(repoRoot(t), ModulePath, Roots)
	if err != nil {
		t.Fatal(err)
	}
	return g, lines
}

func report(t *testing.T, findings []string) {
	t.Helper()
	for _, f := range findings {
		t.Error(f)
	}
}

func TestRepoLayering(t *testing.T) {
	g, _ := loadRepo(t)
	report(t, CheckLayering(g, Layers))
}

func TestRepoRuleRegistration(t *testing.T) {
	g, _ := loadRepo(t)
	report(t, CheckRuleImports(g, RulePackages, RuleImporters, TransitionalRuleImporters))
}

func TestRepoFileSize(t *testing.T) {
	_, lines := loadRepo(t)
	report(t, CheckLineCeilings(lines, MaxFileLines, LineCeilings))
}

func TestRepoComplexity(t *testing.T) {
	funcs, err := LoadComplexities(repoRoot(t), Roots)
	if err != nil {
		t.Fatal(err)
	}
	report(t, CheckComplexity(funcs, MaxComplexity, ComplexityCeilings))
}
