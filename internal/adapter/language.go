// language.go — Maps file extensions to language names.
package adapter

import "strings"

// extLanguages maps file extensions to language names.
var extLanguages = map[string]string{
	".go": "go", ".ts": "typescript", ".tsx": "typescript",
	".js": "javascript", ".jsx": "javascript", ".py": "python",
	".java": "java", ".rs": "rust",
}

// DetectLanguage returns the language name for a path by extension, or "unknown".
func DetectLanguage(path string) string {
	for ext, lang := range extLanguages {
		if strings.HasSuffix(path, ext) {
			return lang
		}
	}
	return "unknown"
}
