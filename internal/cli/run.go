// run.go — Command dispatch for the strict CLI.
package cli

import (
	"fmt"
	"os"
	"strings"
)

// commands maps a subcommand name to its handler, which receives the arguments after the name.
var commands = map[string]func(args []string){
	"init":             runInit,
	"fix":              runFix,
	"inspect":          runInspect,
	"inspect-lineage":  runInspectLineage,
	"lineage-export":   runLineageExport,
	"lineage-diff":     runLineageDiff,
	"lineage-escalate": runLineageEscalate,
	"list-rules":       func([]string) { runListRules() },
	"explain":          runExplain,
	"validate-config":  runValidateConfig,
	"lint":             runLint,
	"audit":            runAudit,
	"trace":            runTrace,
	"policy":           runPolicy,
	"--version":        printVersion,
	"-version":         printVersion,
	"version":          printVersion,
	"--help":           func([]string) { printUsage() },
	"-help":            func([]string) { printUsage() },
	"help":             func([]string) { printUsage() },
}

func printVersion([]string) {
	fmt.Printf("strict version %s\n", version)
}

// Run executes the strict CLI. args is the full argument vector including the program name,
// and buildVersion is the version reported by --version.
func Run(args []string, buildVersion string) {
	version = buildVersion
	if len(args) < 2 {
		runLint(nil)
		return
	}

	if handler, ok := commands[args[1]]; ok {
		handler(args[2:])
		return
	}

	first := args[1]
	// No subcommand: treat all args as lint when first arg is a flag or likely path.
	if strings.HasPrefix(first, "-") || looksLikePathArg(first) {
		runLint(args[1:])
		return
	}
	printUnknownCommand(first)
	os.Exit(2)
}
