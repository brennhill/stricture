// lineage_commands.go — Lineage commands: inspect-lineage, lineage-export, lineage-diff, lineage-escalate.
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/stricture/stricture/internal/lineage"
)

// runInspectLineage parses strict-source annotations and prints JSON output.
func runInspectLineage(args []string) {
	fs := flag.NewFlagSet("inspect-lineage", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println("Usage: strict inspect-lineage <file>")
		fmt.Println()
		fmt.Println("Parse strict-source annotations from comments and print them as JSON.")
	}
	parseFlagSetOrExit(fs, args)

	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Error: inspect-lineage requires a file path argument.")
		fs.Usage()
		os.Exit(2)
	}

	filePath := fs.Arg(0)
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read %s: %v\n", filePath, err)
		os.Exit(1)
	}

	annotations, parseErrs := lineage.Parse(data)
	payload := map[string]interface{}{
		"file":        filePath,
		"annotations": annotations,
		"errors":      parseErrs,
	}

	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to serialize lineage result: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))

	if len(parseErrs) > 0 {
		os.Exit(1)
	}
}

// runLineageExport builds a normalized lineage artifact from source files.
func runLineageExport(args []string) {
	fs := flag.NewFlagSet("lineage-export", flag.ExitOnError)
	outPath := fs.String("out", "", "Write artifact JSON to this path (stdout if empty)")
	strict := fs.Bool("strict", true, "Exit non-zero if parse errors are found")
	profileRaw := fs.String("profile", string(lineage.ProfileStricture), "Export profile (stricture, openlineage, otel, openapi, asyncapi)")
	fs.Usage = func() {
		fmt.Println("Usage: strict lineage-export [options] [paths...]")
		fmt.Println()
		fmt.Println("Build a normalized lineage artifact from source files.")
		fs.PrintDefaults()
	}
	parseFlagSetOrExit(fs, args)
	profile, err := lineage.ParseExportProfile(*profileRaw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	paths := fs.Args()
	if len(paths) == 0 {
		paths = []string{"."}
	}

	artifact, parseErrs, err := lineage.Collect(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: lineage-export failed: %v\n", err)
		os.Exit(1)
	}

	if *outPath != "" {
		if err := lineage.WriteArtifactForProfile(*outPath, artifact, profile); err != nil {
			fmt.Fprintf(os.Stderr, "Error: write artifact: %v\n", err)
			os.Exit(1)
		}
	} else {
		out, err := lineage.MarshalArtifactForProfile(artifact, profile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: marshal artifact: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(out))
	}

	if len(parseErrs) > 0 {
		errOut, err := json.MarshalIndent(parseErrs, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Lineage parse errors (%d)\n", len(parseErrs))
		} else {
			fmt.Fprintf(os.Stderr, "Lineage parse errors (%d):\n%s\n", len(parseErrs), string(errOut))
		}
		if *strict {
			os.Exit(1)
		}
	}
}

// runLineageDiff diffs two lineage artifacts and classifies drift severity.
func runLineageDiff(args []string) {
	fs := flag.NewFlagSet("lineage-diff", flag.ExitOnError)
	basePath := fs.String("base", "", "Path to base lineage artifact JSON")
	headPath := fs.String("head", "", "Path to head lineage artifact JSON")
	outPath := fs.String("out", "", "Write diff JSON to this path (stdout if empty)")
	failOn := fs.String("fail-on", "high", "Fail when drift at/above severity (high|medium|low|info|none)")
	modeRaw := fs.String("mode", string(lineage.ModeBlock), "Enforcement mode: block (exit non-zero) or warn (always exit zero)")
	fs.Usage = func() {
		fmt.Println("Usage: strict lineage-diff --base <file> --head <file> [options]")
		fmt.Println()
		fmt.Println("Diff two lineage artifacts and classify drift severity.")
		fs.PrintDefaults()
	}
	parseFlagSetOrExit(fs, args)

	if strings.TrimSpace(*basePath) == "" || strings.TrimSpace(*headPath) == "" {
		fmt.Fprintln(os.Stderr, "Error: --base and --head are required")
		fs.Usage()
		os.Exit(2)
	}

	base, err := lineage.LoadArtifact(*basePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: load base artifact: %v\n", err)
		os.Exit(1)
	}
	head, err := lineage.LoadArtifact(*headPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: load head artifact: %v\n", err)
		os.Exit(1)
	}

	result := lineage.DiffArtifacts(base, head)
	threshold, err := lineage.ParseSeverity(*failOn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}
	mode, err := lineage.ParseEnforcementMode(*modeRaw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: marshal diff result: %v\n", err)
		os.Exit(1)
	}

	if *outPath != "" {
		if err := os.WriteFile(*outPath, append(out, '\n'), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: write diff output: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println(string(out))
	}

	thresholdExceeded := lineage.ShouldFailAtThreshold(result, threshold)
	if thresholdExceeded && mode == lineage.ModeWarn {
		fmt.Fprintf(os.Stderr, "WARN: drift at/above %s detected, but mode=warn so exit code remains 0\n", threshold)
	}
	if thresholdExceeded && lineage.ShouldFailAtThresholdWithMode(result, threshold, mode) {
		os.Exit(1)
	}
}

// runLineageEscalate resolves emergency contacts upstream from a service.
func runLineageEscalate(args []string) {
	fs := flag.NewFlagSet("lineage-escalate", flag.ExitOnError)
	serviceID := fs.String("service", "", "Service/system ID to investigate")
	artifactPath := fs.String("artifact", "", "Path to lineage artifact JSON")
	systemsPath := fs.String("systems", "", "Path to system registry YAML (optional)")
	maxDepth := fs.Int("max-depth", 8, "Maximum upstream depth to traverse")
	fs.Usage = func() {
		fmt.Println("Usage: strict lineage-escalate --service <id> --artifact <file> [options]")
		fmt.Println()
		fmt.Println("Show emergency contacts for a service and its upstream dependencies.")
		fs.PrintDefaults()
	}
	parseFlagSetOrExit(fs, args)

	if strings.TrimSpace(*serviceID) == "" || strings.TrimSpace(*artifactPath) == "" {
		fmt.Fprintln(os.Stderr, "Error: --service and --artifact are required")
		fs.Usage()
		os.Exit(2)
	}

	artifact, err := lineage.LoadArtifact(*artifactPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: load artifact: %v\n", err)
		os.Exit(1)
	}

	registry := lineage.SystemRegistry{}
	if strings.TrimSpace(*systemsPath) != "" {
		registry, err = lineage.LoadSystemRegistry(*systemsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: load systems registry: %v\n", err)
			os.Exit(1)
		}
	}

	steps, err := lineage.BuildEscalationChain(*serviceID, artifact, registry, *maxDepth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: build escalation chain: %v\n", err)
		os.Exit(1)
	}

	payload := map[string]interface{}{
		"service": *serviceID,
		"steps":   steps,
	}
	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: marshal escalation chain: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}
