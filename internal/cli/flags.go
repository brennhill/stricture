// flags.go — Shared flag types and argument splitting for subcommands.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

type repeatableFlag []string

func (f *repeatableFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *repeatableFlag) Set(value string) error {
	for _, token := range strings.Split(value, ",") {
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			continue
		}
		*f = append(*f, trimmed)
	}
	return nil
}

func (f *repeatableFlag) Values() []string {
	out := make([]string, len(*f))
	copy(out, *f)
	return out
}

func parseFlagSetOrExit(fs *flag.FlagSet, args []string) {
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
}

// lintValueFlags are the lint flags that consume the next argument as their value.
var lintValueFlags = map[string]bool{
	"-format":          true,
	"--format":         true,
	"-config":          true,
	"--config":         true,
	"-rule":            true,
	"--rule":           true,
	"-category":        true,
	"--category":       true,
	"-ext":             true,
	"--ext":            true,
	"-severity":        true,
	"--severity":       true,
	"-concurrency":     true,
	"--concurrency":    true,
	"-output":          true,
	"--output":         true,
	"-max-violations":  true,
	"--max-violations": true,
	"-baseline":        true,
	"--baseline":       true,
}

func splitLintArgs(args []string) ([]string, []string, error) {
	return splitArgs(args, argSpec{valueFlags: lintValueFlags, honorDoubleDash: true})
}

func splitTraceArgs(args []string) (string, []string, error) {
	valueFlags := map[string]bool{
		"-manifest":      true,
		"--manifest":     true,
		"-trace-format":  true,
		"--trace-format": true,
		"-service":       true,
		"--service":      true,
	}
	flagArgs, positional, err := splitArgs(args, argSpec{
		valueFlags:    valueFlags,
		maxPositional: 1,
		tooManyErr:    "trace accepts exactly one file argument",
	})
	if err != nil {
		return "", nil, err
	}
	tracePath := ""
	if len(positional) > 0 {
		tracePath = positional[0]
	}
	return tracePath, flagArgs, nil
}

func splitAuditArgs(args []string) ([]string, []string, error) {
	valueFlags := map[string]bool{
		"-manifest":    true,
		"--manifest":   true,
		"-service":     true,
		"--service":    true,
		"-strictness":  true,
		"--strictness": true,
		"-format":      true,
		"--format":     true,
		"-output":      true,
		"--output":     true,
		"-config":      true,
		"--config":     true,
	}
	return splitArgs(args, argSpec{valueFlags: valueFlags})
}

func verbosef(enabled bool, format string, args ...interface{}) {
	if !enabled {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

// argSpec describes how a subcommand's arguments are tokenized.
type argSpec struct {
	// valueFlags are flags whose next argument is their value (unless written --flag=value).
	valueFlags map[string]bool
	// honorDoubleDash treats everything after "--" as positional.
	honorDoubleDash bool
	// maxPositional rejects more than this many positional arguments with tooManyErr (0 = unlimited).
	maxPositional int
	tooManyErr    string
}

// splitArgs separates flag tokens (with their values) from positional arguments, in order.
func splitArgs(args []string, spec argSpec) (flagArgs, positional []string, err error) {
	flagArgs = make([]string, 0, len(args))
	positional = make([]string, 0, len(args))
	pathsOnly := false

	rest := args
	for len(rest) > 0 {
		token := strings.TrimSpace(rest[0])
		rest = rest[1:]

		switch {
		case token == "":
		case pathsOnly:
			positional = append(positional, token)
		case spec.honorDoubleDash && token == "--":
			pathsOnly = true
		case strings.HasPrefix(token, "-"):
			flagArgs = append(flagArgs, token)
			if strings.Contains(token, "=") || !spec.valueFlags[token] {
				continue
			}
			if len(rest) == 0 {
				return nil, nil, fmt.Errorf("flag %s requires a value", token)
			}
			flagArgs = append(flagArgs, rest[0])
			rest = rest[1:]
		default:
			if spec.maxPositional > 0 && len(positional) >= spec.maxPositional {
				return nil, nil, errors.New(spec.tooManyErr)
			}
			positional = append(positional, token)
		}
	}
	return flagArgs, positional, nil
}
