// archtest.go — Architecture checks: import layering and file-size ratchet.
package archtest

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

// Graph maps a module-relative package path to its module-relative internal imports.
// Only non-test files contribute edges.
type Graph map[string][]string

// StdFlag is the pseudo-package recorded for imports of the standard "flag" package.
const StdFlag = "std:flag"

// LayerRule says which module-internal packages a package family may import.
// Patterns are module-relative; a trailing "/..." matches the package and its subpackages.
type LayerRule struct {
	Pattern string
	Allowed []string
}

// Match reports whether pkg matches pattern.
func Match(pkg, pattern string) bool {
	if base, ok := strings.CutSuffix(pattern, "/..."); ok {
		return pkg == base || strings.HasPrefix(pkg, base+"/")
	}
	return pkg == pattern
}

func matchAny(pkg string, patterns []string) bool {
	for _, p := range patterns {
		if Match(pkg, p) {
			return true
		}
	}
	return false
}

// ruleFor returns the first rule whose pattern matches pkg.
func ruleFor(rules []LayerRule, pkg string) (LayerRule, bool) {
	for _, r := range rules {
		if Match(pkg, r.Pattern) {
			return r, true
		}
	}
	return LayerRule{}, false
}

// CheckLayering returns one message per illegal import and per package no rule covers.
func CheckLayering(g Graph, rules []LayerRule) []string {
	var out []string
	for _, pkg := range sortedKeys(g) {
		rule, ok := ruleFor(rules, pkg)
		if !ok {
			out = append(out, fmt.Sprintf("LAYERING: package %s is not classified. Add it to the layer table in internal/archtest/policy.go and docs/ARCHITECTURE.md", pkg))
			continue
		}
		for _, imp := range g[pkg] {
			if !matchAny(imp, rule.Allowed) {
				out = append(out, fmt.Sprintf("LAYERING: %s imports %s, which layer %q does not allow. Move the code or depend on an allowed package (see docs/ARCHITECTURE.md)", pkg, imp, rule.Pattern))
			}
		}
	}
	return out
}

// CheckRuleImports enforces that individual rule packages are imported only by
// allowed importers. exempt lists transitional importers; a stale entry (one that
// no longer imports a rule package) is reported so the list can only shrink.
func CheckRuleImports(g Graph, rulePkgs, allowedImporters, exempt []string) []string {
	var out []string
	needed := map[string]bool{}
	for _, pkg := range sortedKeys(g) {
		for _, imp := range g[pkg] {
			if !matchAny(imp, rulePkgs) || matchAny(pkg, rulePkgs) || matchAny(pkg, allowedImporters) {
				continue
			}
			if matchAny(pkg, exempt) {
				needed[pkg] = true
				continue
			}
			out = append(out, fmt.Sprintf("REGISTRATION: %s imports rule package %s. Register rules through internal/rules/all instead", pkg, imp))
		}
	}
	for _, e := range exempt {
		if !needed[e] {
			out = append(out, fmt.Sprintf("REGISTRATION: transitional exemption %s is stale. Remove it from internal/archtest/policy.go", e))
		}
	}
	return out
}

// CheckLineCeilings enforces maxLines on every file. Files listed in ceilings may
// exceed maxLines but never grow past their ceiling; once a file fits, its entry is stale.
func CheckLineCeilings(lines map[string]int, maxLines int, ceilings map[string]int) []string {
	var out []string
	for _, f := range sortedKeys(lines) {
		n := lines[f]
		ceiling, listed := ceilings[f]
		switch {
		case n <= maxLines && listed:
			out = append(out, fmt.Sprintf("FILE-SIZE: %s is %d lines and fits the %d limit. Remove its entry from the ratchet in internal/archtest/policy.go", f, n, maxLines))
		case n > maxLines && !listed:
			out = append(out, fmt.Sprintf("FILE-SIZE: %s has %d lines, over the %d limit. Split it into focused files", f, n, maxLines))
		case listed && n > ceiling:
			out = append(out, fmt.Sprintf("FILE-SIZE: %s grew to %d lines, past its ratchet ceiling %d. Split it instead of growing it", f, n, ceiling))
		case listed && n < ceiling:
			out = append(out, fmt.Sprintf("FILE-SIZE: %s shrank to %d lines. Lower its ceiling from %d in internal/archtest/policy.go", f, n, ceiling))
		}
	}
	for _, f := range sortedKeys(ceilings) {
		if _, ok := lines[f]; !ok {
			out = append(out, fmt.Sprintf("FILE-SIZE: ratchet entry %s matches no file. Remove it from internal/archtest/policy.go", f))
		}
	}
	return out
}

// walkGoFiles calls visit with the module-relative path and contents of every non-test Go file
// under roots, skipping testdata directories. Reads are scoped to repoRoot.
func walkGoFiles(repoRoot string, roots []string, visit func(rel string, data []byte) error) error {
	fsys := os.DirFS(repoRoot)
	for _, root := range roots {
		err := fs.WalkDir(fsys, root, func(rel string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == "testdata" {
					return fs.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(rel, ".go") || strings.HasSuffix(rel, "_test.go") {
				return nil
			}
			data, err := fs.ReadFile(fsys, rel)
			if err != nil {
				return fmt.Errorf("read source %s: %w. Check file permissions", rel, err)
			}
			return visit(rel, data)
		})
		if err != nil {
			return fmt.Errorf("walk %s: %w. Run from a valid checkout", root, err)
		}
	}
	return nil
}

// LoadGraph parses imports of non-test Go files under roots (module-relative dirs).
func LoadGraph(repoRoot, modulePath string, roots []string) (Graph, map[string]int, error) {
	g := Graph{}
	lines := map[string]int{}
	fset := token.NewFileSet()
	err := walkGoFiles(repoRoot, roots, func(rel string, data []byte) error {
		lines[rel] = strings.Count(string(data), "\n")
		f, err := parser.ParseFile(fset, rel, data, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("parse imports in %s: %w. Fix the syntax error", rel, err)
		}
		pkg := path.Dir(rel)
		if _, ok := g[pkg]; !ok {
			g[pkg] = nil
		}
		for _, spec := range f.Imports {
			imp, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return fmt.Errorf("read import %s in %s: %w. Fix the import path", spec.Path.Value, rel, err)
			}
			if rest, ok := strings.CutPrefix(imp, modulePath+"/"); ok {
				g[pkg] = appendUnique(g[pkg], rest)
			} else if imp == "flag" {
				// Flag parsing is presentation-only; the layer table must allow it explicitly.
				g[pkg] = appendUnique(g[pkg], StdFlag)
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return g, lines, nil
}

func appendUnique(s []string, v string) []string {
	for _, x := range s {
		if x == v {
			return s
		}
	}
	return append(s, v)
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
