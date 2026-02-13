// file.go — UnifiedFileModel and all language-agnostic parsed file types.
package model

// UnifiedFileModel represents a parsed file in any supported language.
type UnifiedFileModel struct {
	Path        string
	Language    string
	IsTestFile  bool
	Source      []byte
	LineCount   int
	Imports     []ImportDecl
	Exports     []ExportDecl
	Functions   []FuncModel
	Types       []TypeModel
	Classes     []ClassModel
	TestCases   []TestCase
	TestTargets []string
	JSONTags    []JSONTag
}

// ImportDecl represents an import statement.
type ImportDecl struct {
	Path      string
	Alias     string
	Names     []string
	IsDefault bool
	StartLine int
	EndLine   int
}

// ExportDecl represents an export statement.
type ExportDecl struct {
	Name      string
	Kind      string
	IsDefault bool
	StartLine int
	EndLine   int
}

// FuncModel represents a function or method.
type FuncModel struct {
	Name        string
	Receiver    string
	Params      []ParamModel
	Returns     []string
	IsExported  bool
	IsTest      bool
	Calls       []string
	ErrorExits  []ErrorExit
	LineCount   int
	Complexity  int
	StartLine   int
	EndLine     int
}

// ParamModel represents a function parameter.
type ParamModel struct {
	Name string
	Type string
}

// TypeModel represents a type definition (struct, interface, type alias).
type TypeModel struct {
	Name      string
	Kind      string
	Fields    []FieldModel
	Methods   []string
	Exported  bool
	StartLine int
	EndLine   int
}

// FieldModel represents a struct field or interface method.
type FieldModel struct {
	Name      string
	Type      string
	Exported  bool
	JSONTag   string
	StartLine int
}

// ClassModel represents a class (for OOP languages).
type ClassModel struct {
	Name       string
	Exported   bool
	Methods    []FuncModel
	Fields     []FieldModel
	Implements []string
	StartLine  int
	EndLine    int
}

// TestCase represents a test function.
type TestCase struct {
	Name       string
	FuncName   string
	Assertions []Assertion
	Mocks      []Mock
	StartLine  int
	EndLine    int
}

// Assertion represents a test assertion.
type Assertion struct {
	Kind      string
	Subject   string
	Expected  string
	StartLine int
}

// Mock represents a mocked dependency in a test.
type Mock struct {
	Target    string
	Kind      string
	StartLine int
}

// ErrorExit represents an error return or throw statement.
type ErrorExit struct {
	Kind      string
	Message   string
	StartLine int
}

// JSONTag represents a JSON tag on a struct field.
type JSONTag struct {
	FieldName string
	JSONName  string
	Options   []string
	StartLine int
}
