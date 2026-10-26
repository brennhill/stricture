// baseline.go — Baseline files that suppress known violations so only new ones are reported.
package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/stricture/stricture/internal/model"
)

// State describes what applying a baseline did.
type State struct {
	Enabled      bool
	Path         string
	EntryCount   int
	Suppressed   int
	Bootstrapped bool
	Entries      []Entry
	Added        []model.Violation
	Resolved     []Entry
}

// Options controls baseline application.
type Options struct {
	// BootstrapIfMissing writes a new baseline from current violations when the file does not exist.
	BootstrapIfMissing bool
}

type file struct {
	Version     string  `json:"version"`
	GeneratedAt string  `json:"generatedAt"`
	Entries     []Entry `json:"entries"`
}

// Entry identifies one baselined violation.
type Entry struct {
	RuleID    string `json:"ruleId"`
	FilePath  string `json:"filePath"`
	StartLine int    `json:"startLine"`
	Message   string `json:"message"`
}

// Apply filters *violations against the baseline at pathValue. A blank path disables baselining.
// A missing baseline is created from the current violations when opts.BootstrapIfMissing is set.
func Apply(pathValue string, violations *[]model.Violation, opts Options) (State, error) {
	state := State{}
	if strings.TrimSpace(pathValue) == "" {
		return state, nil
	}
	if violations == nil {
		return state, fmt.Errorf("internal baseline error: violations pointer is nil")
	}

	state.Enabled = true
	state.Path = pathValue

	data, err := os.ReadFile(pathValue)
	if os.IsNotExist(err) {
		if !opts.BootstrapIfMissing {
			return state, fmt.Errorf("baseline %s does not exist (run once without --diff to bootstrap)", pathValue)
		}
		return bootstrap(state, violations)
	}
	if err != nil {
		return state, fmt.Errorf("read baseline %s: %w", pathValue, err)
	}
	return compare(state, data, violations)
}

// bootstrap writes a baseline containing every current violation and suppresses them all.
func bootstrap(state State, violations *[]model.Violation) (State, error) {
	entries := make([]Entry, 0, len(*violations))
	for _, v := range *violations {
		entries = append(entries, Entry{
			RuleID:    strings.TrimSpace(v.RuleID),
			FilePath:  filepath.ToSlash(v.FilePath),
			StartLine: v.StartLine,
			Message:   strings.TrimSpace(v.Message),
		})
	}
	sortEntries(entries)

	doc := file{
		Version:     "1",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Entries:     entries,
	}
	if err := write(state.Path, doc); err != nil {
		return state, err
	}

	state.EntryCount = len(entries)
	state.Suppressed = len(*violations)
	state.Bootstrapped = true
	state.Entries = entries
	state.Resolved = []Entry{}
	state.Added = []model.Violation{}
	*violations = []model.Violation{}
	return state, nil
}

func write(pathValue string, doc file) error {
	encoded, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal baseline %s: %w", pathValue, err)
	}
	encoded = append(encoded, '\n')
	if err := os.MkdirAll(filepath.Dir(pathValue), 0o755); err != nil {
		return fmt.Errorf("create baseline directory for %s: %w", pathValue, err)
	}
	if err := os.WriteFile(pathValue, encoded, 0o644); err != nil {
		return fmt.Errorf("write baseline %s: %w", pathValue, err)
	}
	return nil
}

// compare removes baselined violations from *violations and reports what was added or resolved.
func compare(state State, data []byte, violations *[]model.Violation) (State, error) {
	var doc file
	if err := json.Unmarshal(data, &doc); err != nil {
		return state, fmt.Errorf("parse baseline %s: %w", state.Path, err)
	}

	lookup := map[string]bool{}
	for _, entry := range doc.Entries {
		lookup[keyFromEntry(entry)] = true
	}

	rawCurrent := append([]model.Violation(nil), (*violations)...)
	filtered := make([]model.Violation, 0, len(*violations))
	for _, v := range *violations {
		if lookup[keyFromViolation(v)] {
			state.Suppressed++
			continue
		}
		filtered = append(filtered, v)
	}

	state.EntryCount = len(doc.Entries)
	state.Entries = append([]Entry(nil), doc.Entries...)
	state.Added = append([]model.Violation(nil), filtered...)
	state.Resolved = resolvedEntries(rawCurrent, doc.Entries)
	*violations = filtered
	return state, nil
}

// resolvedEntries returns baseline entries that no longer match any current violation, sorted.
func resolvedEntries(current []model.Violation, entries []Entry) []Entry {
	currentLookup := map[string]bool{}
	for _, v := range current {
		currentLookup[keyFromViolation(v)] = true
	}

	out := make([]Entry, 0)
	for _, entry := range entries {
		if !currentLookup[keyFromEntry(entry)] {
			out = append(out, entry)
		}
	}
	sortEntries(out)
	return out
}

func sortEntries(entries []Entry) {
	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.FilePath != b.FilePath {
			return a.FilePath < b.FilePath
		}
		if a.StartLine != b.StartLine {
			return a.StartLine < b.StartLine
		}
		if a.RuleID != b.RuleID {
			return a.RuleID < b.RuleID
		}
		return a.Message < b.Message
	})
}

func keyFromEntry(entry Entry) string {
	return key(entry.RuleID, entry.FilePath, entry.StartLine, entry.Message)
}

func keyFromViolation(v model.Violation) string {
	return key(v.RuleID, v.FilePath, v.StartLine, v.Message)
}

func key(ruleID, filePath string, line int, message string) string {
	return fmt.Sprintf("%s|%s|%d|%s",
		strings.TrimSpace(ruleID),
		filepath.ToSlash(strings.TrimSpace(filePath)),
		line,
		strings.TrimSpace(message))
}
