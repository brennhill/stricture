// policy.go — The enforced architecture: layer table, size limit, and shrink-only ratchets.
package archtest

// ModulePath is the Go module path from go.mod.
const ModulePath = "github.com/stricture/stricture"

// Roots are the module-relative directories under architectural control.
var Roots = []string{"cmd", "internal", "pkg"}

// MaxFileLines is the hard limit for any non-test Go file.
const MaxFileLines = 800

// MaxComplexity is the hard limit on cyclomatic complexity for any non-test function.
const MaxComplexity = 15

// Layers is the import policy. First match wins, so list specific patterns first.
// Keep in sync with docs/ARCHITECTURE.md.
var Layers = []LayerRule{
	// Foundation: no internal dependencies.
	{Pattern: "internal/model", Allowed: nil},
	{Pattern: "pkg/...", Allowed: nil},
	{Pattern: "internal/archtest", Allowed: nil},

	// Leaves: depend on model only.
	{Pattern: "internal/adapter/...", Allowed: []string{"internal/model", "internal/adapter"}},
	{Pattern: "internal/rules/all", Allowed: []string{"internal/model", "internal/rules/...", "internal/rules"}},
	{Pattern: "internal/rules", Allowed: []string{"internal/model"}},
	{Pattern: "internal/rules/...", Allowed: []string{"internal/model", "internal/rules"}},
	{Pattern: "internal/report/...", Allowed: []string{"internal/model"}},
	{Pattern: "internal/reporter", Allowed: []string{"internal/model"}},
	{Pattern: "internal/config", Allowed: []string{"internal/model"}},
	{Pattern: "internal/baseline", Allowed: []string{"internal/model"}},
	{Pattern: "internal/manifest", Allowed: []string{"internal/model"}},
	{Pattern: "internal/suppression", Allowed: []string{"internal/model"}},
	{Pattern: "internal/fix", Allowed: []string{"internal/model"}},
	{Pattern: "internal/lineage/...", Allowed: []string{"internal/model", "internal/lineage"}},
	{Pattern: "internal/discovery", Allowed: []string{"internal/model"}},
	{Pattern: "internal/plugins", Allowed: []string{"internal/model", "pkg/rule"}},
	{Pattern: "internal/server/...", Allowed: []string{"internal/model", "internal/server"}},
	{Pattern: "internal/testutil", Allowed: []string{"internal/model"}},

	// Use-case layer: orchestrates leaves; knows nothing about the CLI.
	{Pattern: "internal/engine", Allowed: []string{
		"internal/model", "internal/adapter/...", "internal/config", "internal/manifest",
		"internal/suppression", "internal/fix", "internal/plugins", "internal/rules",
		"internal/report/...", "internal/reporter", "internal/discovery",
	}},
	{Pattern: "internal/app", Allowed: []string{
		"internal/model", "internal/adapter/...", "internal/config", "internal/manifest",
		"internal/suppression", "internal/fix", "internal/plugins", "internal/rules",
		"internal/report/...", "internal/reporter", "internal/discovery", "internal/engine",
	}},

	// Presentation: the only layer allowed to see everything internal.
	{Pattern: "internal/cli/...", Allowed: []string{"internal/...", StdFlag}},

	// Entry points: each binary may import only its own front-end package. Nothing may import cmd/.
	{Pattern: "cmd/strict", Allowed: []string{"internal/cli"}},
	{Pattern: "cmd/stricture-server", Allowed: []string{"internal/server"}},
	{Pattern: "cmd/demo-pack", Allowed: []string{"internal/lineage", StdFlag}},
}

// RulePackages are the individual rule packages. Only RuleImporters may import them directly.
var RulePackages = []string{
	"internal/rules/arch", "internal/rules/conv", "internal/rules/ctr", "internal/rules/tq",
}

// RuleImporters register rules; everything else goes through them.
var RuleImporters = []string{"internal/rules/all"}

// TransitionalRuleImporters is empty now that rules register through internal/rules/all.
// Do not add entries.
var TransitionalRuleImporters = []string{}

// LineCeilings are shrink-only exceptions to MaxFileLines. A file may not grow past its
// ceiling, and the entry must be deleted once the file fits. Never add new entries.
var LineCeilings = map[string]int{}

// ComplexityCeilings are shrink-only exceptions to MaxComplexity, keyed "file.go:Func" or
// "file.go:Recv.Method". A function may not get more complex than its ceiling, and the entry
// must be deleted once it fits. Never add new entries.
var ComplexityCeilings = map[string]int{}
