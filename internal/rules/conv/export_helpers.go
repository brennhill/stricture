// export_helpers.go - Shared export scanning and naming helpers for convention rules.
package conv

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/stricture/stricture/internal/model"
)

type exportSymbol struct {
	Name string
	Kind string
	Line int
}

const (
	namingCamelCase      = "camelCase"
	namingPascalCase     = "PascalCase"
	namingUpperSnakeCase = "UPPER_SNAKE_CASE"
)

var (
	tsExportFunctionRe = regexp.MustCompile(`^\s*export\s+(?:async\s+)?function\s+([A-Za-z_$][A-Za-z0-9_$]*)`)
	tsExportConstRe    = regexp.MustCompile(`^\s*export\s+const\s+([A-Za-z_$][A-Za-z0-9_$]*)`)
	tsExportClassRe    = regexp.MustCompile(`^\s*export\s+class\s+([A-Za-z_$][A-Za-z0-9_$]*)`)
	tsExportTypeRe     = regexp.MustCompile(`^\s*export\s+(?:type|interface|enum)\s+([A-Za-z_$][A-Za-z0-9_$]*)`)
	tsExportDefaultRe  = regexp.MustCompile(`^\s*export\s+default\b`)
	tsExportListRe     = regexp.MustCompile(`^\s*export\s*\{([^}]*)\}`)

	goExportFuncRe     = regexp.MustCompile(`^\s*func\s+(?:\([^)]*\)\s*)?([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	goExportTypeRe     = regexp.MustCompile(`^\s*type\s+([A-Za-z_][A-Za-z0-9_]*)\s+`)
	goExportConstVarRe = regexp.MustCompile(`^\s*(?:const|var)\s+([A-Za-z_][A-Za-z0-9_]*)\b`)
)

func scanExportedSymbols(file *model.UnifiedFileModel) []exportSymbol {
	if file == nil {
		return nil
	}

	symbols := make([]exportSymbol, 0)
	seen := map[string]bool{}

	for _, exp := range file.Exports {
		name := strings.TrimSpace(exp.Name)
		if name == "" {
			continue
		}
		line := exp.StartLine
		if line <= 0 {
			line = 1
		}
		s := exportSymbol{Name: name, Kind: normalizeSymbolKind(exp.Kind), Line: line}
		key := symbolKey(s)
		if seen[key] {
			continue
		}
		seen[key] = true
		symbols = append(symbols, s)
	}

	lang := strings.ToLower(strings.TrimSpace(file.Language))
	var scanned []exportSymbol
	switch lang {
	case "go", "golang":
		scanned = scanGoExports(string(file.Source))
	default:
		scanned = scanTSLikeExports(string(file.Source))
	}

	for _, s := range scanned {
		key := symbolKey(s)
		if seen[key] {
			continue
		}
		seen[key] = true
		symbols = append(symbols, s)
	}

	sort.Slice(symbols, func(i, j int) bool {
		if symbols[i].Line != symbols[j].Line {
			return symbols[i].Line < symbols[j].Line
		}
		if symbols[i].Name != symbols[j].Name {
			return symbols[i].Name < symbols[j].Name
		}
		return symbols[i].Kind < symbols[j].Kind
	})

	return symbols
}

func scanTSLikeExports(source string) []exportSymbol {
	lines := strings.Split(source, "\n")
	out := make([]exportSymbol, 0)

	for i, raw := range lines {
		lineNo := i + 1
		line := stripLineComment(raw)
		if strings.TrimSpace(line) == "" {
			continue
		}

		if m := tsExportFunctionRe.FindStringSubmatch(line); len(m) == 2 {
			out = append(out, exportSymbol{Name: m[1], Kind: "function", Line: lineNo})
			continue
		}
		if m := tsExportConstRe.FindStringSubmatch(line); len(m) == 2 {
			out = append(out, exportSymbol{Name: m[1], Kind: "constant", Line: lineNo})
			continue
		}
		if m := tsExportClassRe.FindStringSubmatch(line); len(m) == 2 {
			out = append(out, exportSymbol{Name: m[1], Kind: "class", Line: lineNo})
			continue
		}
		if m := tsExportTypeRe.FindStringSubmatch(line); len(m) == 2 {
			out = append(out, exportSymbol{Name: m[1], Kind: "type", Line: lineNo})
			continue
		}
		if m := tsExportListRe.FindStringSubmatch(line); len(m) == 2 {
			out = append(out, parseTSExportList(m[1], lineNo)...)
			continue
		}
		if tsExportDefaultRe.MatchString(line) {
			out = append(out, exportSymbol{Name: "default", Kind: "default", Line: lineNo})
		}
	}

	return out
}

func parseTSExportList(content string, lineNo int) []exportSymbol {
	parts := strings.Split(content, ",")
	out := make([]exportSymbol, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		name := token
		kind := "value"
		if strings.Contains(token, " as ") {
			pieces := strings.Split(token, " as ")
			name = strings.TrimSpace(pieces[len(pieces)-1])
		}
		if name == "default" {
			kind = "default"
		}
		out = append(out, exportSymbol{Name: name, Kind: kind, Line: lineNo})
	}
	return out
}

func scanGoExports(source string) []exportSymbol {
	lines := strings.Split(source, "\n")
	out := make([]exportSymbol, 0)
	for i, raw := range lines {
		lineNo := i + 1
		line := stripLineComment(raw)
		if strings.TrimSpace(line) == "" {
			continue
		}

		if m := goExportFuncRe.FindStringSubmatch(line); len(m) == 2 {
			name := m[1]
			if isExportedIdentifier(name) {
				out = append(out, exportSymbol{Name: name, Kind: "function", Line: lineNo})
			}
			continue
		}
		if m := goExportTypeRe.FindStringSubmatch(line); len(m) == 2 {
			name := m[1]
			if isExportedIdentifier(name) {
				out = append(out, exportSymbol{Name: name, Kind: "type", Line: lineNo})
			}
			continue
		}
		if m := goExportConstVarRe.FindStringSubmatch(line); len(m) == 2 {
			name := m[1]
			if isExportedIdentifier(name) {
				out = append(out, exportSymbol{Name: name, Kind: "constant", Line: lineNo})
			}
		}
	}
	return out
}

func normalizeSymbolKind(kind string) string {
	k := strings.ToLower(strings.TrimSpace(kind))
	switch k {
	case "func", "function", "method":
		return "function"
	case "class":
		return "class"
	case "const", "constant", "var", "value":
		return "constant"
	case "type", "interface", "enum", "struct", "alias":
		return "type"
	case "default":
		return "default"
	default:
		if k == "" {
			return "value"
		}
		return k
	}
}

func symbolKey(s exportSymbol) string {
	return s.Name + "|" + s.Kind + "|" + strconv.Itoa(s.Line)
}

func stripLineComment(line string) string {
	if idx := strings.Index(line, "//"); idx >= 0 {
		return line[:idx]
	}
	return line
}

func isExportedIdentifier(name string) bool {
	if name == "" {
		return false
	}
	first, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(first)
}

func normalizeConventionName(raw string) string {
	switch strings.TrimSpace(raw) {
	case namingCamelCase:
		return namingCamelCase
	case namingPascalCase:
		return namingPascalCase
	case namingUpperSnakeCase, "upper_snake_case", "UPPER_SNAKE":
		return namingUpperSnakeCase
	default:
		return ""
	}
}

func matchesNameConvention(name string, convention string) bool {
	if strings.TrimSpace(name) == "" {
		return true
	}
	words := splitIntoWords(name)
	if len(words) == 0 {
		return true
	}

	switch normalizeConventionName(convention) {
	case namingCamelCase:
		return name == joinCamel(words)
	case namingPascalCase:
		return name == joinPascal(words)
	case namingUpperSnakeCase:
		return name == strings.ToUpper(joinSnake(words))
	default:
		return true
	}
}

func suggestNameForConvention(name string, convention string) string {
	words := splitIntoWords(name)
	if len(words) == 0 {
		return name
	}
	switch normalizeConventionName(convention) {
	case namingCamelCase:
		return joinCamel(words)
	case namingPascalCase:
		return joinPascal(words)
	case namingUpperSnakeCase:
		return strings.ToUpper(joinSnake(words))
	default:
		return name
	}
}

func toStringMap(value interface{}) (map[string]interface{}, bool) {
	switch m := value.(type) {
	case map[string]interface{}:
		return m, true
	case map[interface{}]interface{}:
		out := make(map[string]interface{}, len(m))
		for k, v := range m {
			ks, ok := k.(string)
			if !ok {
				continue
			}
			out[ks] = v
		}
		return out, true
	default:
		return nil, false
	}
}

func toStringSlice(value interface{}) []string {
	switch raw := value.(type) {
	case []string:
		out := make([]string, 0, len(raw))
		for _, s := range raw {
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	case []interface{}:
		out := make([]string, 0, len(raw))
		for _, item := range raw {
			s, ok := item.(string)
			if !ok {
				continue
			}
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
