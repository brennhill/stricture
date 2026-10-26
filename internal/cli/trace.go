// trace.go — The trace command: trace artifact validation.
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	manifestpkg "github.com/stricture/stricture/internal/manifest"
)

func runTrace(args []string) {
	if hasHelpFlag(args) {
		printTraceUsage()
		return
	}

	tracePath, flagArgs, argErr := splitTraceArgs(args)
	if argErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", argErr)
		os.Exit(2)
	}

	fs := flag.NewFlagSet("trace", flag.ExitOnError)
	manifestPath := fs.String("manifest", "", "Path to stricture manifest file")
	traceFormat := fs.String("trace-format", "auto", "Trace format (auto, har, otel, custom)")
	service := fs.String("service", "", "Service name that produced the trace")
	strict := fs.Bool("strict", false, "Fail on parse anomalies")
	parseFlagSetOrExit(fs, flagArgs)

	if strings.TrimSpace(tracePath) == "" {
		fmt.Fprintln(os.Stderr, "Error: trace requires a trace file path.")
		os.Exit(2)
	}
	data, err := os.ReadFile(tracePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: read trace %s: %v\n", tracePath, err)
		os.Exit(2)
	}

	manifest := strings.TrimSpace(*manifestPath)
	if manifest == "" {
		manifest = autoDetectManifestPath()
	}
	if manifest != "" {
		if _, err := manifestpkg.Load(manifest); err != nil {
			fmt.Fprintf(os.Stderr, "Error: manifest %s is invalid or unreadable: %v\n", manifest, err)
			os.Exit(2)
		}
	}

	format := strings.ToLower(strings.TrimSpace(*traceFormat))
	if format == "" || format == "auto" {
		format = inferTraceFormat(tracePath, data)
	}
	if err := validateTracePayload(format, data, *strict); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	fmt.Printf("Trace parsed: file=%s format=%s bytes=%d\n", tracePath, format, len(data))
	if strings.TrimSpace(*service) != "" {
		fmt.Printf("Service: %s\n", strings.TrimSpace(*service))
	}
	if *strict {
		fmt.Println("Strict mode: parse checks passed.")
	}
}

func printTraceUsage() {
	fmt.Println("Usage: strict trace <file> [options]")
	fmt.Println()
	fmt.Println("Validate a trace artifact baseline (JSON parse + format sanity checks).")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --manifest <path>      Path to stricture-manifest.yml")
	fmt.Println("  --trace-format <fmt>   Trace format: auto, har, otel, custom")
	fmt.Println("  --service <name>       Service that produced the trace")
	fmt.Println("  --strict               Fail on trace anomalies")
}

func inferTraceFormat(pathValue string, data []byte) string {
	switch strings.ToLower(strings.TrimSpace(filepath.Ext(pathValue))) {
	case ".har":
		return "har"
	case ".otel", ".otlp":
		return "otel"
	}
	envelope, err := parseTraceEnvelope(data)
	if err != nil {
		return "custom"
	}
	if _, ok := envelope["resourceSpans"]; ok {
		return "otel"
	}
	if _, ok := envelope["log"]; ok {
		return "har"
	}
	return "custom"
}

func validateTracePayload(format string, data []byte, strict bool) error {
	envelope, err := parseTraceEnvelope(data)
	if err != nil {
		return err
	}
	switch format {
	case "har":
		logRaw, ok := envelope["log"]
		if !ok {
			return fmt.Errorf("HAR trace must include top-level log object")
		}
		logObj, ok := logRaw.(map[string]interface{})
		if !ok {
			return fmt.Errorf("HAR trace log field must be an object")
		}
		if strict && len(logObj) == 0 {
			return fmt.Errorf("strict HAR validation requires non-empty log object")
		}
	case "otel":
		resourceRaw, ok := envelope["resourceSpans"]
		if !ok {
			return fmt.Errorf("OTEL trace must include resourceSpans array")
		}
		resourceSpans, ok := resourceRaw.([]interface{})
		if !ok {
			return fmt.Errorf("OTEL trace resourceSpans field must be an array")
		}
		if strict && len(resourceSpans) == 0 {
			return fmt.Errorf("strict OTEL validation requires non-empty resourceSpans array")
		}
	case "custom":
		if strict && len(envelope) == 0 {
			return fmt.Errorf("strict custom trace validation requires non-empty JSON object")
		}
	default:
		return fmt.Errorf("unsupported trace format %q (valid: auto, har, otel, custom)", format)
	}
	return nil
}

func parseTraceEnvelope(data []byte) (map[string]interface{}, error) {
	if !json.Valid(data) {
		return nil, fmt.Errorf("trace payload is not valid JSON")
	}
	var envelope map[string]interface{}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("trace payload must be a JSON object")
	}
	return envelope, nil
}
