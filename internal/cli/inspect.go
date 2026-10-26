// inspect.go — The inspect command: print a parsed UnifiedFileModel.
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/stricture/stricture/internal/adapter"
	"github.com/stricture/stricture/internal/adapter/java"
	"github.com/stricture/stricture/internal/adapter/python"
	"github.com/stricture/stricture/internal/adapter/typescript"
	"github.com/stricture/stricture/internal/engine"
	"github.com/stricture/stricture/internal/model"
)

// runInspect parses a file and prints its UnifiedFileModel as JSON.
func runInspect(args []string) {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println("Usage: strict inspect [options] <file>")
		fmt.Println()
		fmt.Println("Parse a file and print its UnifiedFileModel as formatted JSON.")
		fmt.Println("This is a debugging tool for adapter development.")
		fs.PrintDefaults()
	}
	parseFlagSetOrExit(fs, args)

	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Error: inspect requires a file path argument.")
		fs.Usage()
		os.Exit(2)
	}

	filePath := fs.Arg(0)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", filePath)
			os.Exit(2)
		}
		fmt.Fprintf(os.Stderr, "Error: cannot read %s: %v\n", filePath, err)
		os.Exit(2)
	}
	if isLikelyBinary(data) {
		fmt.Fprintf(os.Stderr, "Error: cannot inspect binary file: %s\n", filePath)
		os.Exit(2)
	}

	lang := adapter.DetectLanguage(filePath)
	if lang == "unknown" {
		fmt.Fprintf(os.Stderr, "Error: no language adapter for %q files. Supported: %s\n", filepath.Ext(filePath), strings.Join(supportedInspectLanguages(), ", "))
		os.Exit(2)
	}

	parsed, err := inspectParseFile(filePath, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: inspect parse failed for %s: %v\n", filePath, err)
		os.Exit(2)
	}

	out, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal model: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}

func inspectParseFile(path string, source []byte) (*model.UnifiedFileModel, error) {
	lang := adapter.DetectLanguage(path)
	cfg := adapter.AdapterConfig{}

	switch lang {
	case "typescript", "javascript":
		return (&typescript.Adapter{}).Parse(path, source, cfg)
	case "python":
		return (&python.Adapter{}).Parse(path, source, cfg)
	case "java":
		return (&java.Adapter{}).Parse(path, source, cfg)
	case "go":
		return parseGoInspect(path, source)
	default:
		return nil, fmt.Errorf("unsupported language %q", lang)
	}
}

func parseGoInspect(path string, source []byte) (*model.UnifiedFileModel, error) {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, path, source, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	ufm := &model.UnifiedFileModel{
		Path:       filepath.ToSlash(path),
		Language:   "go",
		Source:     append([]byte(nil), source...),
		LineCount:  engine.CountLines(source),
		IsTestFile: strings.HasSuffix(strings.ToLower(filepath.Base(path)), "_test.go"),
		Imports:    []model.ImportDecl{},
		Functions:  []model.FuncModel{},
		Types:      []model.TypeModel{},
	}

	for _, imp := range parsed.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		name := ""
		if imp.Name != nil {
			name = imp.Name.Name
		}
		ufm.Imports = append(ufm.Imports, model.ImportDecl{
			Path:      importPath,
			Alias:     name,
			StartLine: fset.Position(imp.Pos()).Line,
			EndLine:   fset.Position(imp.End()).Line,
		})
	}

	for _, decl := range parsed.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			fn := model.FuncModel{
				Name:       d.Name.Name,
				IsExported: ast.IsExported(d.Name.Name),
				StartLine:  fset.Position(d.Pos()).Line,
				EndLine:    fset.Position(d.End()).Line,
			}
			ufm.Functions = append(ufm.Functions, fn)
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				ufm.Types = append(ufm.Types, model.TypeModel{
					Name:      ts.Name.Name,
					Exported:  ast.IsExported(ts.Name.Name),
					StartLine: fset.Position(ts.Pos()).Line,
					EndLine:   fset.Position(ts.End()).Line,
				})
			}
		}
	}

	return ufm, nil
}

func isLikelyBinary(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	if !utf8.Valid(data) {
		return true
	}
	sample := data
	if len(sample) > 1024 {
		sample = sample[:1024]
	}
	for _, b := range sample {
		if b == 0x00 {
			return true
		}
	}
	return false
}

func supportedInspectLanguages() []string {
	return []string{"go", "typescript", "javascript", "python", "java"}
}
