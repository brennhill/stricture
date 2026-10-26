// lint_options_test.go — Tests lint flag validation, severity resolution, and dispatch table coverage.
package cli

import (
	"strings"
	"testing"
)

func validRaw() rawLintFlags {
	return rawLintFlags{format: "text", concurrency: 2}
}

func firstFailure(raw rawLintFlags) string {
	for _, c := range usageChecks(raw) {
		if c.failed {
			return c.message
		}
	}
	return ""
}

func TestUsageChecksAcceptDefaults(t *testing.T) {
	if msg := firstFailure(validRaw()); msg != "" {
		t.Fatalf("defaults should be valid, got %q", msg)
	}
}

func TestUsageChecksRejectConflicts(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*rawLintFlags)
		want   string
	}{
		{"fix and dry-run", func(r *rawLintFlags) { r.fixApply, r.fixDryRun = true, true }, "--fix and --fix-dry-run are mutually exclusive"},
		{"backup without fix", func(r *rawLintFlags) { r.fixBackup = true }, "--fix-backup requires --fix"},
		{"changed and staged", func(r *rawLintFlags) { r.changedOnly, r.stagedOnly = true, true }, "--changed and --staged are mutually exclusive"},
		{"diff without baseline", func(r *rawLintFlags) { r.diffMode = true; r.baselinePath = "  " }, "--diff requires --baseline"},
		{"bad format", func(r *rawLintFlags) { r.format = "xml" }, `invalid format "xml" (valid: text, json, sarif, junit)`},
		{"negative max violations", func(r *rawLintFlags) { r.maxViolations = -1 }, "--max-violations must be >= 0"},
		{"color conflict", func(r *rawLintFlags) { r.forceColor, r.forceNoColor = true, true }, "--color and --no-color are mutually exclusive"},
		{"cache conflict", func(r *rawLintFlags) { r.cacheEnabled, r.noCache = true, true }, "--cache and --no-cache are mutually exclusive"},
		{"zero concurrency", func(r *rawLintFlags) { r.concurrency = 0 }, "--concurrency must be >= 1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := validRaw()
			tc.mutate(&raw)
			if got := firstFailure(raw); got != tc.want {
				t.Errorf("first failure = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestUsageChecksReportFirstConflictInOrder(t *testing.T) {
	raw := validRaw()
	raw.fixApply, raw.fixDryRun, raw.format = true, true, "xml"
	if got := firstFailure(raw); !strings.Contains(got, "--fix and --fix-dry-run") {
		t.Errorf("earlier conflict should win, got %q", got)
	}
}

func TestResolveMinSeverity(t *testing.T) {
	cases := []struct {
		severity string
		quiet    bool
		want     string
		wantErr  string
	}{
		{"", false, "", ""},
		{" Warn ", false, "warn", ""},
		{"ERROR", false, "error", ""},
		{"", true, "error", ""},
		{"error", true, "error", ""},
		{"warn", true, "", "--quiet cannot be combined with --severity=warn"},
		{"fatal", false, "", `invalid severity "fatal" (valid: error, warn)`},
	}
	for _, tc := range cases {
		got, err := resolveMinSeverity(tc.severity, tc.quiet)
		if tc.wantErr != "" {
			if err == nil || err.Error() != tc.wantErr {
				t.Errorf("resolveMinSeverity(%q,%v) error = %v, want %q", tc.severity, tc.quiet, err, tc.wantErr)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Errorf("resolveMinSeverity(%q,%v) = %q, %v; want %q", tc.severity, tc.quiet, got, err, tc.want)
		}
	}
}

func TestValidateLintFlagsBuildsOptions(t *testing.T) {
	raw := validRaw()
	raw.noCache, raw.cacheEnabled, raw.severity, raw.extFilter = true, false, "warn", "go,.ts"
	raw.ruleFilters = repeatableFlag{"A", "B"}

	opts := validateLintFlags(raw, nil)
	if len(opts.paths) != 1 || opts.paths[0] != "." {
		t.Errorf("paths = %v, want default [.]", opts.paths)
	}
	if opts.cacheActive {
		t.Error("--no-cache should disable the cache")
	}
	if opts.minSeverity != "warn" || !opts.extensions[".go"] || !opts.extensions[".ts"] {
		t.Errorf("unexpected severity/extensions: %q %v", opts.minSeverity, opts.extensions)
	}
	if len(opts.ruleFilters) != 2 {
		t.Errorf("ruleFilters = %v", opts.ruleFilters)
	}

	raw.noCache, raw.cacheEnabled = false, true
	withCache := validateLintFlags(raw, []string{"src"})
	if !withCache.cacheActive || withCache.paths[0] != "src" {
		t.Errorf("--cache should enable the cache and keep explicit paths: %+v", withCache)
	}
}

func TestCommandsTableCoversAdvertisedSubcommands(t *testing.T) {
	for _, name := range []string{
		"init", "fix", "inspect", "inspect-lineage", "lineage-export", "lineage-diff", "lineage-escalate",
		"list-rules", "explain", "validate-config", "lint", "audit", "trace", "policy",
		"version", "--version", "-version", "help", "--help", "-help",
	} {
		if commands[name] == nil {
			t.Errorf("command %q is not dispatched", name)
		}
	}
}

func TestSplitArgsHonorsDoubleDashOnlyWhenAsked(t *testing.T) {
	spec := argSpec{valueFlags: map[string]bool{"--out": true}, honorDoubleDash: true}
	flags, positional, err := splitArgs([]string{"--out", "x", "--", "--not-a-flag", " ", "p"}, spec)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(flags, ",") != "--out,x" || strings.Join(positional, ",") != "--not-a-flag,p" {
		t.Errorf("flags=%v positional=%v", flags, positional)
	}

	flags, _, err = splitArgs([]string{"--"}, argSpec{})
	if err != nil || len(flags) != 1 || flags[0] != "--" {
		t.Errorf("without honorDoubleDash, -- is an ordinary flag token: %v %v", flags, err)
	}
}
