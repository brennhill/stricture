// main.go — CLI entry point for the artifact-driven demo pack generator.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/stricture/stricture/internal/lineage"
)

func main() {
	artifactPath := flag.String("artifact", "tests/lineage/usecases/current.json", "Path to lineage artifact JSON")
	systemsPath := flag.String("systems", "tests/lineage/usecases/systems.yml", "Path to lineage systems registry YAML")
	outJSON := flag.String("out-json", "site/public/demo/demo-pack.json", "Output JSON file path")
	outTS := flag.String("out-ts", "site/worker/src/generated/demo-pack.ts", "Output TypeScript module path")
	flag.Parse()

	artifact, err := lineage.LoadArtifact(*artifactPath)
	if err != nil {
		fatalf("load artifact: %v", err)
	}

	registry, err := lineage.LoadSystemRegistry(*systemsPath)
	if err != nil {
		fatalf("load systems registry: %v", err)
	}

	pack, err := buildDemoPack(artifact, registry)
	if err != nil {
		fatalf("build demo pack: %v", err)
	}

	if err := writeJSON(*outJSON, pack); err != nil {
		fatalf("write json: %v", err)
	}

	if err := writeTS(*outTS, pack); err != nil {
		fatalf("write ts: %v", err)
	}

	fmt.Printf("generated demo pack: %s and %s\n", *outJSON, *outTS)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}
