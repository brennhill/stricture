// run.go — Runs rules over files sequentially or in parallel, honoring suppressions and limits.
package engine

import (
	"fmt"
	"strings"
	"sync"

	"github.com/stricture/stricture/internal/model"
	"github.com/stricture/stricture/internal/suppression"
)

// RunRules checks every file against every rule and returns the violations that are not suppressed.
// Files are checked concurrently unless maxViolations > 0, which requires sequential fail-fast order.
func RunRules(files []*model.UnifiedFileModel, rules []model.Rule, ctx *model.ProjectContext, maxViolations int, concurrency int) []model.Violation {
	if concurrency <= 1 || len(files) <= 1 || maxViolations > 0 {
		// Preserve exact fail-fast behavior when maxViolations is configured.
		return runLintRulesSequential(files, rules, ctx, maxViolations)
	}
	return runLintRulesParallel(files, rules, ctx, concurrency)
}

func runLintRulesSequential(files []*model.UnifiedFileModel, rules []model.Rule, ctx *model.ProjectContext, maxViolations int) []model.Violation {
	violations := make([]model.Violation, 0)
	for _, file := range files {
		budget := 0
		if maxViolations > 0 {
			budget = maxViolations - len(violations)
			if budget <= 0 {
				break
			}
		}
		violations = append(violations, runLintRulesForFile(file, rules, ctx, budget)...)
	}
	return violations
}

func runLintRulesParallel(files []*model.UnifiedFileModel, rules []model.Rule, ctx *model.ProjectContext, concurrency int) []model.Violation {
	workerCount := concurrency
	if workerCount > len(files) {
		workerCount = len(files)
	}

	jobs := make(chan *model.UnifiedFileModel)
	results := make(chan []model.Violation, len(files))

	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for file := range jobs {
			results <- runLintRulesForFile(file, rules, ctx, 0)
		}
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker()
	}

	for _, file := range files {
		jobs <- file
	}
	close(jobs)

	wg.Wait()
	close(results)

	violations := make([]model.Violation, 0)
	for chunk := range results {
		violations = append(violations, chunk...)
	}
	return violations
}

// runLintRulesForFile runs all rules on one file, stopping once maxViolations (when > 0) are collected.
func runLintRulesForFile(file *model.UnifiedFileModel, rules []model.Rule, ctx *model.ProjectContext, maxViolations int) []model.Violation {
	violations := make([]model.Violation, 0)
	policy := suppression.Compile(file.Source)
	for _, rule := range rules {
		limit := 0
		if maxViolations > 0 {
			limit = maxViolations - len(violations)
		}
		violations = append(violations, checkRule(rule, file, ctx, policy, limit)...)
		if maxViolations > 0 && len(violations) >= maxViolations {
			break
		}
	}
	return violations
}

// checkRule runs one rule on one file and returns its unsuppressed violations, at most limit when
// limit > 0. A panicking rule yields a single error violation instead of crashing the run.
func checkRule(rule model.Rule, file *model.UnifiedFileModel, ctx *model.ProjectContext, policy *suppression.Policy, limit int) (out []model.Violation) {
	cfg := model.RuleConfig{Severity: rule.DefaultSeverity(), Options: map[string]interface{}{}}
	if bound, ok := rule.(RuleWithConfig); ok {
		rule, cfg = bound.Rule, bound.Config
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			out = append(out, model.Violation{
				RuleID:    rule.ID(),
				Severity:  "error",
				Message:   fmt.Sprintf("Rule panicked: %v", recovered),
				FilePath:  file.Path,
				StartLine: 1,
			})
		}
	}()

	for _, v := range rule.Check(file, ctx, cfg) {
		ruleID := strings.TrimSpace(v.RuleID)
		if ruleID == "" {
			ruleID = rule.ID()
			v.RuleID = ruleID
		}
		line := v.StartLine
		if line <= 0 {
			line = 1
		}
		if policy.Suppressed(ruleID, line) {
			continue
		}
		out = append(out, v)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}
