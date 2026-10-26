// git.go — Resolves the set of files touched in a git repository for --changed and --staged.
package discovery

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitScopedFileSet returns canonical keys (see PathKey) of files changed in the git working tree
// and index (changedOnly) or staged in the index (stagedOnly).
func GitScopedFileSet(changedOnly bool, stagedOnly bool) (map[string]bool, error) {
	rootRaw, err := gitOutput("rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("git-scoped lint requires a git repository: %w", err)
	}
	root := strings.TrimSpace(rootRaw)
	if root == "" {
		return nil, fmt.Errorf("unable to resolve git repository root")
	}

	combined := make([]string, 0)
	switch {
	case stagedOnly:
		staged, err := gitOutputLines("diff", "--name-only", "--cached", "--diff-filter=ACMRT")
		if err != nil {
			return nil, err
		}
		combined = append(combined, staged...)
	case changedOnly:
		working, err := gitOutputLines("diff", "--name-only", "--diff-filter=ACMRT")
		if err != nil {
			return nil, err
		}
		staged, err := gitOutputLines("diff", "--name-only", "--cached", "--diff-filter=ACMRT")
		if err != nil {
			return nil, err
		}
		untracked, err := gitOutputLines("ls-files", "--others", "--exclude-standard")
		if err != nil {
			return nil, err
		}
		combined = append(combined, working...)
		combined = append(combined, staged...)
		combined = append(combined, untracked...)
	default:
		return map[string]bool{}, nil
	}

	out := map[string]bool{}
	for _, rel := range combined {
		r := strings.TrimSpace(rel)
		if r == "" {
			continue
		}
		out[PathKey(root, r)] = true
	}
	return out, nil
}

func gitOutputLines(args ...string) ([]string, error) {
	out, err := gitOutput(args...)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(out, "\n")
	trimmed := make([]string, 0, len(lines))
	for _, line := range lines {
		v := strings.TrimSpace(line)
		if v == "" {
			continue
		}
		trimmed = append(trimmed, v)
	}
	return trimmed, nil
}

func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), msg)
	}
	return string(out), nil
}
