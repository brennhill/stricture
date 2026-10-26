// rules_commands.go — Rule introspection commands: list-rules, explain, validate-config.
package cli

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/stricture/stricture/internal/rules/all"

	"github.com/stricture/stricture/internal/model"
	"gopkg.in/yaml.v3"
)

// runListRules prints a table of all registered rules.
func runListRules() {
	registry := all.Registry()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCATEGORY\tDEFAULT\tFIXABLE\tDESCRIPTION")
	fmt.Fprintln(w, "--\t--------\t-------\t-------\t-----------")
	for _, r := range sortedRulesForDisplay(registry) {
		meta := ruleMetadata(r.ID())
		desc := r.Description()
		if meta.RequiresManifest {
			desc += " (requires manifest)"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			r.ID(), strings.ToUpper(r.Category()), r.DefaultSeverity(), meta.Fixability, desc)
	}
	w.Flush()
	fmt.Printf("\n%d rules registered.\n", len(registry.All()))
}

func runExplain(args []string) {
	fs := flag.NewFlagSet("explain", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println("Usage: strict explain <rule-id>")
		fmt.Println()
		fmt.Println("Show details for a specific rule.")
	}
	parseFlagSetOrExit(fs, args)

	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Error: explain requires a rule ID argument.")
		fs.Usage()
		os.Exit(2)
	}

	ruleID := strings.TrimSpace(fs.Arg(0))
	registry := all.Registry()
	ruleDef, ok := registry.ByID(ruleID)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: unknown rule %q\n", ruleID)
		fmt.Fprintln(os.Stderr, "Run 'strict list-rules' to see available rules.")
		os.Exit(2)
	}

	meta := ruleMetadata(ruleDef.ID())
	requiresManifest := "No"
	if meta.RequiresManifest {
		requiresManifest = "Yes"
	}

	fmt.Printf("ID: %s\n", ruleDef.ID())
	fmt.Printf("Category: %s\n", strings.ToUpper(ruleDef.Category()))
	fmt.Printf("Default Severity: %s\n", ruleDef.DefaultSeverity())
	fmt.Printf("Fixable: %s\n", meta.Fixability)
	fmt.Printf("Needs Project Context: %t\n", ruleDef.NeedsProjectContext())
	fmt.Printf("Requires Manifest: %s\n", requiresManifest)
	fmt.Printf("Description: %s\n", ruleDef.Description())
	fmt.Printf("Why: %s\n", ruleDef.Why())
}

// runValidateConfig checks that a config file is valid YAML with recognized rule IDs.
func runValidateConfig(args []string) {
	fs := flag.NewFlagSet("validate-config", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println("Usage: strict validate-config [path]")
		fmt.Println()
		fmt.Println("Validate a .stricture.yml configuration file.")
		fmt.Println("Checks YAML syntax and verifies all rule IDs are recognized.")
	}
	parseFlagSetOrExit(fs, args)

	configPath := ".stricture.yml"
	if fs.NArg() > 0 {
		configPath = fs.Arg(0)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read %s: %v\n", configPath, err)
		os.Exit(1)
	}

	var cfg struct {
		Rules   map[string]string                 `yaml:"rules"`
		Options map[string]map[string]interface{} `yaml:"options"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid YAML in %s: %v\n", configPath, err)
		os.Exit(1)
	}

	registry := all.Registry()
	var unknown []string
	for id := range cfg.Rules {
		if _, ok := registry.ByID(id); !ok {
			unknown = append(unknown, id)
		}
	}

	if len(unknown) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: %d unrecognized rule(s): %s\n",
			len(unknown), strings.Join(unknown, ", "))
		fmt.Fprintf(os.Stderr, "(These may be valid rules not yet registered in this build.)\n")
	}

	fmt.Printf("Config %s: valid YAML, %d rules configured.\n", configPath, len(cfg.Rules))
}

type ruleMeta struct {
	Fixability       string
	RequiresManifest bool
}

func ruleMetadata(ruleID string) ruleMeta {
	switch ruleID {
	case "CONV-file-header", "CONV-file-naming", "CONV-test-file-location":
		return ruleMeta{Fixability: "Yes"}
	case "TQ-mock-scope":
		return ruleMeta{Fixability: "Partial"}
	case "CTR-strictness-parity", "CTR-manifest-conformance":
		return ruleMeta{Fixability: "No", RequiresManifest: true}
	default:
		return ruleMeta{Fixability: "No"}
	}
}

func sortedRulesForDisplay(registry *model.RuleRegistry) []model.Rule {
	all := append([]model.Rule(nil), registry.All()...)
	sort.SliceStable(all, func(i, j int) bool {
		ci := categoryOrder(strings.ToLower(all[i].Category()))
		cj := categoryOrder(strings.ToLower(all[j].Category()))
		if ci != cj {
			return ci < cj
		}
		return all[i].ID() < all[j].ID()
	})
	return all
}

func categoryOrder(category string) int {
	switch strings.ToLower(category) {
	case "tq":
		return 0
	case "arch":
		return 1
	case "conv":
		return 2
	case "ctr":
		return 3
	default:
		return 4
	}
}
