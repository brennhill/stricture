// complexity.go — Cyclomatic complexity per function, and the shrink-only ratchet check.
package archtest

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// FuncComplexities returns cyclomatic complexity keyed by function name ("Name" or "Recv.Name").
// Complexity is 1 plus each if, for, range, non-default case, non-default select clause,
// && and ||. Function literals count toward their enclosing function.
func FuncComplexities(filename string, src []byte) (map[string]int, error) {
	f, err := parser.ParseFile(token.NewFileSet(), filename, src, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w. Fix the syntax error", filename, err)
	}
	out := map[string]int{}
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Body != nil {
				out[uniqueKey(out, funcKey(d))] = 1 + countDecisions(d.Body)
			}
		case *ast.GenDecl:
			addPackageLiterals(out, d)
		}
	}
	return out, nil
}

// uniqueKey keeps repeated names (several init functions) from overwriting each other.
func uniqueKey(m map[string]int, key string) string {
	if _, taken := m[key]; !taken {
		return key
	}
	for i := 2; ; i++ {
		if k := fmt.Sprintf("%s#%d", key, i); m[k] == 0 {
			return k
		}
	}
}

// addPackageLiterals measures package-level `var X = func() {...}` values, keyed by variable name.
func addPackageLiterals(out map[string]int, d *ast.GenDecl) {
	for _, spec := range d.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for i, v := range vs.Values {
			if n := countDecisions(v); n > 0 || isFuncLit(v) {
				name := "var"
				if i < len(vs.Names) {
					name = vs.Names[i].Name
				}
				out[uniqueKey(out, name)] = 1 + n
			}
		}
	}
}

func isFuncLit(e ast.Expr) bool {
	_, ok := e.(*ast.FuncLit)
	return ok
}

func funcKey(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return fn.Name.Name
	}
	t := fn.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	switch r := t.(type) {
	case *ast.Ident:
		return r.Name + "." + fn.Name.Name
	case *ast.IndexExpr: // generic receiver T[K]
		if id, ok := r.X.(*ast.Ident); ok {
			return id.Name + "." + fn.Name.Name
		}
	case *ast.IndexListExpr:
		if id, ok := r.X.(*ast.Ident); ok {
			return id.Name + "." + fn.Name.Name
		}
	}
	return fn.Name.Name
}

func countDecisions(body ast.Node) int {
	n := 0
	ast.Inspect(body, func(node ast.Node) bool {
		switch x := node.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
			n++
		case *ast.CaseClause:
			if x.List != nil {
				n++
			}
		case *ast.CommClause:
			if x.Comm != nil {
				n++
			}
		case *ast.BinaryExpr:
			if x.Op == token.LAND || x.Op == token.LOR {
				n++
			}
		}
		return true
	})
	return n
}

// LoadComplexities measures every function in non-test Go files under roots.
// Keys are "relative/file.go:Func".
func LoadComplexities(repoRoot string, roots []string) (map[string]int, error) {
	out := map[string]int{}
	err := walkGoFiles(repoRoot, roots, func(rel string, data []byte) error {
		funcs, err := FuncComplexities(rel, data)
		if err != nil {
			return err
		}
		for name, c := range funcs {
			out[rel+":"+name] = c
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CheckComplexity enforces maxComplexity on every function. Functions listed in ceilings may
// exceed it but never rise past their ceiling; once a function fits, its entry is stale.
func CheckComplexity(funcs map[string]int, maxComplexity int, ceilings map[string]int) []string {
	var out []string
	for _, k := range sortedKeys(funcs) {
		c := funcs[k]
		ceiling, listed := ceilings[k]
		switch {
		case c <= maxComplexity && listed:
			out = append(out, fmt.Sprintf("COMPLEXITY: %s is %d and fits the %d limit. Remove its entry from the ratchet in internal/archtest/policy.go", k, c, maxComplexity))
		case c > maxComplexity && !listed:
			out = append(out, fmt.Sprintf("COMPLEXITY: %s has complexity %d, over the %d limit. Extract helpers or table-drive the branches", k, c, maxComplexity))
		case listed && c > ceiling:
			out = append(out, fmt.Sprintf("COMPLEXITY: %s rose to %d, past its ratchet ceiling %d. Reduce it instead of growing it", k, c, ceiling))
		case listed && c < ceiling:
			out = append(out, fmt.Sprintf("COMPLEXITY: %s dropped to %d. Lower its ceiling from %d in internal/archtest/policy.go", k, c, ceiling))
		}
	}
	for _, k := range sortedKeys(ceilings) {
		if _, ok := funcs[k]; !ok {
			out = append(out, fmt.Sprintf("COMPLEXITY: ratchet entry %s matches no function. Remove it from internal/archtest/policy.go", k))
		}
	}
	return out
}
