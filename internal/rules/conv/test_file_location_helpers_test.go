package conv

import "testing"

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "golang", want: "go"},
		{in: " TSX ", want: "typescript"},
		{in: "jsx", want: "javascript"},
		{in: "python", want: "python"},
		{in: "custom", want: "custom"},
	}

	for _, tt := range tests {
		got := normalizeLanguage(tt.in)
		if got != tt.want {
			t.Fatalf("normalizeLanguage(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestResolveTestSuffixes(t *testing.T) {
	got := resolveTestSuffixes("typescript", nil)
	if len(got) == 0 || got[0] != ".test.ts" {
		t.Fatalf("resolveTestSuffixes default failed: %#v", got)
	}

	custom := resolveTestSuffixes("typescript", map[string]interface{}{
		"suffixes": map[string]interface{}{
			"typescript": []interface{}{".it.ts", ".spec.ts"},
		},
	})
	if len(custom) != 2 || custom[0] != ".it.ts" || custom[1] != ".spec.ts" {
		t.Fatalf("resolveTestSuffixes custom failed: %#v", custom)
	}

	invalid := resolveTestSuffixes("typescript", map[string]interface{}{
		"suffixes": "not-a-map",
	})
	if len(invalid) == 0 || invalid[0] != ".test.ts" {
		t.Fatalf("resolveTestSuffixes invalid options should fall back to defaults: %#v", invalid)
	}
}

func TestLocationMatchesStrategy(t *testing.T) {
	if !locationMatchesStrategy("src/service/user.test.ts", testStrategyColocated) {
		t.Fatal("expected colocated strategy to match non-tests root path")
	}
	if locationMatchesStrategy("tests/service/user.test.ts", testStrategyColocated) {
		t.Fatal("expected colocated strategy to reject tests/ root path")
	}
	if !locationMatchesStrategy("tests/service/user.test.ts", testStrategyMirrored) {
		t.Fatal("expected mirrored strategy to match tests/ root path")
	}
	if locationMatchesStrategy("src/service/user.test.ts", testStrategyMirrored) {
		t.Fatal("expected mirrored strategy to reject src/ path")
	}
	if !locationMatchesStrategy("src/service/__tests__/user.test.ts", testStrategySubfolder) {
		t.Fatal("expected subfolder strategy to match __tests__ path")
	}
}

func TestExpectedTestPath(t *testing.T) {
	if got := expectedTestPath("tests/service/user.test.ts", testStrategyColocated); got != "src/service/user.test.ts" {
		t.Fatalf("colocated expected path = %q", got)
	}
	if got := expectedTestPath("src/service/user.test.ts", testStrategyMirrored); got != "tests/service/user.test.ts" {
		t.Fatalf("mirrored expected path = %q", got)
	}
	if got := expectedTestPath("src/service/user.test.ts", testStrategySubfolder); got != "src/service/__tests__/user.test.ts" {
		t.Fatalf("subfolder expected path = %q", got)
	}
	if got := expectedTestPath("tests/service/user.test.ts", testStrategySubfolder); got != "src/service/__tests__/user.test.ts" {
		t.Fatalf("subfolder expected path from tests root = %q", got)
	}
}

func TestLooksLikeTestFile(t *testing.T) {
	if !looksLikeTestFile("pkg/user_test.go", "go", defaultTestSuffixes["go"], false) {
		t.Fatal("expected go _test file to be recognized")
	}
	if !looksLikeTestFile("tests/user_test.py", "python", defaultTestSuffixes["python"], false) {
		t.Fatal("expected python test file to be recognized")
	}
	if !looksLikeTestFile("src/user.spec.ts", "typescript", defaultTestSuffixes["typescript"], false) {
		t.Fatal("expected ts spec file to be recognized")
	}
	if looksLikeTestFile("src/user.ts", "typescript", defaultTestSuffixes["typescript"], false) {
		t.Fatal("expected non-test file not to be recognized")
	}
	if !looksLikeTestFile("src/user.ts", "typescript", defaultTestSuffixes["typescript"], true) {
		t.Fatal("isTestFile flag should force recognition")
	}
}
