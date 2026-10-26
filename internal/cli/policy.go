// policy.go — The policy command: policy URL binding verification.
package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func runPolicy(args []string) {
	if len(args) == 0 {
		printPolicyUsage()
		return
	}

	switch strings.TrimSpace(args[0]) {
	case "-h", "--help", "-help", "help":
		printPolicyUsage()
		return
	case "verify-ref":
		runPolicyVerifyRef(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown policy command %q\n", args[0])
		fmt.Fprintln(os.Stderr, "Valid policy commands: verify-ref")
		os.Exit(2)
	}
}

func printPolicyUsage() {
	fmt.Println("Usage: strict policy <command> [options]")
	fmt.Println()
	fmt.Println("Policy commands:")
	fmt.Println("  verify-ref      Verify repo strict:policy_url matches CI expected URL")
	fmt.Println()
	fmt.Println("Run 'strict policy <command> --help' for details.")
}

func runPolicyVerifyRef(args []string) {
	if hasHelpFlag(args) {
		printPolicyVerifyRefUsage()
		return
	}

	fs := flag.NewFlagSet("policy verify-ref", flag.ExitOnError)
	configPath := fs.String("config", ".stricture.yml", "Path to repo config file")
	expectedURL := fs.String("expected-url", "", "Org-approved strict:policy_url value")
	expectedURLEnv := fs.String("expected-url-env", "STRICTURE_POLICY_URL", "Environment variable containing expected policy URL")
	requireSHA := fs.Bool("require-sha", false, "Require strict:policy_sha256 to be set")
	var allowURLs repeatableFlag
	fs.Var(&allowURLs, "allow-url", "Allowed strict:policy_url (repeatable)")
	parseFlagSetOrExit(fs, args)

	if len(fs.Args()) > 0 {
		fmt.Fprintf(os.Stderr, "Error: verify-ref does not accept positional args: %s\n", strings.Join(fs.Args(), " "))
		os.Exit(2)
	}

	binding, err := loadPolicyBindingFromConfig(resolveConfigPath(*configPath))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	expected, expectedSource := resolveExpectedPolicyURL(strings.TrimSpace(*expectedURL), strings.TrimSpace(*expectedURLEnv))
	allowlist := allowURLs.Values()
	if expected == "" && len(allowlist) == 0 {
		fmt.Fprintln(os.Stderr, "Error: provide --expected-url, --expected-url-env, or at least one --allow-url")
		os.Exit(2)
	}

	if expected != "" && binding.PolicyURL != expected {
		fmt.Fprintf(os.Stderr, "Error: strict:policy_url mismatch: got %q want %q\n", binding.PolicyURL, expected)
		os.Exit(1)
	}
	if len(allowlist) > 0 && !containsString(allowlist, binding.PolicyURL) {
		fmt.Fprintf(os.Stderr, "Error: strict:policy_url %q is not in allowlist\n", binding.PolicyURL)
		os.Exit(1)
	}
	if *requireSHA && binding.PolicySHA256 == "" {
		fmt.Fprintln(os.Stderr, "Error: strict:policy_sha256 is required but missing")
		os.Exit(1)
	}
	if binding.PolicySHA256 != "" && !isValidSHA256Hex(binding.PolicySHA256) {
		fmt.Fprintf(os.Stderr, "Error: strict:policy_sha256 must be 64 hex chars, got %q\n", binding.PolicySHA256)
		os.Exit(1)
	}

	source := "config"
	if expectedSource != "" {
		source = expectedSource
	}
	fmt.Printf("Policy binding verified: strict:policy_url=%s source=%s\n", binding.PolicyURL, source)
}

func printPolicyVerifyRefUsage() {
	fmt.Println("Usage: strict policy verify-ref [options]")
	fmt.Println()
	fmt.Println("CI/CD guard: verify repo strict:policy_url matches the org-approved URL.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --config <path>             Path to repo config file (default: .stricture.yml)")
	fmt.Println("  --expected-url <url>        Expected strict:policy_url value")
	fmt.Println("  --expected-url-env <name>   Env var with expected URL (default: STRICTURE_POLICY_URL)")
	fmt.Println("  --allow-url <url>           Allowed URL (repeatable)")
	fmt.Println("  --require-sha               Require strict:policy_sha256 in config")
}

type policyBinding struct {
	PolicyURL    string
	PolicySHA256 string
}

func loadPolicyBindingFromConfig(configPath string) (policyBinding, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return policyBinding{}, fmt.Errorf("read config %s: %w", configPath, err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return policyBinding{}, fmt.Errorf("parse config %s: %w", configPath, err)
	}

	urlValue := strings.TrimSpace(valueAsString(raw["strict:policy_url"]))
	if urlValue == "" {
		return policyBinding{}, fmt.Errorf("config %s is missing strict:policy_url", configPath)
	}
	shaValue := strings.TrimSpace(valueAsString(raw["strict:policy_sha256"]))
	return policyBinding{
		PolicyURL:    urlValue,
		PolicySHA256: shaValue,
	}, nil
}

func resolveExpectedPolicyURL(expectedURL string, expectedURLEnv string) (string, string) {
	if strings.TrimSpace(expectedURL) != "" {
		return strings.TrimSpace(expectedURL), "flag:expected-url"
	}
	if strings.TrimSpace(expectedURLEnv) == "" {
		return "", ""
	}
	envName := strings.TrimSpace(expectedURLEnv)
	value := strings.TrimSpace(os.Getenv(envName))
	if value == "" {
		return "", ""
	}
	return value, "env:" + envName
}

func valueAsString(value interface{}) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprintf("%v", value)
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == strings.TrimSpace(target) {
			return true
		}
	}
	return false
}

func isValidSHA256Hex(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			continue
		}
		return false
	}
	return true
}
