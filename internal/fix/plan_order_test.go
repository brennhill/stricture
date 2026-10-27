// plan_order_test.go — Pins that Plan output order does not depend on violation order.
package fix

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stricture/stricture/internal/model"
)

func TestPlanOrderIsIndependentOfViolationOrder(t *testing.T) {
	dir := t.TempDir()
	names := []string{"b_service.ts", "a_service.ts", "c_service.ts"}
	var violations []model.Violation
	for _, n := range names {
		path := filepath.Join(dir, n)
		if err := os.WriteFile(path, []byte("export const v = 1;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		violations = append(violations, model.Violation{RuleID: "CONV-file-header", FilePath: path, StartLine: 1})
	}
	want, err := Plan(violations)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		shuffled := append([]model.Violation(nil), violations...)
		shuffled[0], shuffled[i%len(shuffled)] = shuffled[i%len(shuffled)], shuffled[0]
		got, err := Plan(shuffled)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("plan order changed with violation order:\n got %v\nwant %v", got, want)
		}
	}
}
