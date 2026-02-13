// main.go — Stricture CLI entry point.
package main

import (
	"flag"
	"fmt"
	"os"
)

var version = "0.1.0-dev"

func main() {
	var (
		showVersion = flag.Bool("version", false, "Print version and exit")
		showHelp    = flag.Bool("help", false, "Print help and exit")
		format      = flag.String("format", "text", "Output format (text, json, sarif, junit)")
		configPath  = flag.String("config", ".stricture.yml", "Path to configuration file")
		rule        = flag.String("rule", "", "Run a single rule by ID")
		category    = flag.String("category", "", "Run all rules in a category")
		changed     = flag.Bool("changed", false, "Lint only changed files")
		staged      = flag.Bool("staged", false, "Lint only staged files")
		noCache     = flag.Bool("no-cache", false, "Disable caching")
	)

	flag.Parse()

	if *showVersion {
		fmt.Printf("stricture version %s\n", version)
		os.Exit(0)
	}

	if *showHelp {
		fmt.Println("Stricture — A fast, language-agnostic linter")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  stricture [options] [paths...]")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Validate format
	validFormats := map[string]bool{
		"text":  true,
		"json":  true,
		"sarif": true,
		"junit": true,
	}
	if !validFormats[*format] {
		fmt.Fprintf(os.Stderr, "Error: invalid format %q (valid: text, json, sarif, junit)\n", *format)
		os.Exit(2)
	}

	// Placeholder for actual implementation
	fmt.Println("not yet implemented")
	fmt.Printf("Config: %s\n", *configPath)
	fmt.Printf("Format: %s\n", *format)
	if *rule != "" {
		fmt.Printf("Rule: %s\n", *rule)
	}
	if *category != "" {
		fmt.Printf("Category: %s\n", *category)
	}
	if *changed {
		fmt.Println("Mode: changed files only")
	}
	if *staged {
		fmt.Println("Mode: staged files only")
	}
	if *noCache {
		fmt.Println("Cache: disabled")
	}

	os.Exit(0)
}
