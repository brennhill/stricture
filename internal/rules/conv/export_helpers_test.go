package conv

import (
	"reflect"
	"testing"
)

func TestParseTSExportList_AliasAndDefault(t *testing.T) {
	got := parseTSExportList("Foo, Bar as Baz, default,  ", 7)
	want := []exportSymbol{
		{Name: "Foo", Kind: "value", Line: 7},
		{Name: "Baz", Kind: "value", Line: 7},
		{Name: "default", Kind: "default", Line: 7},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseTSExportList() = %#v, want %#v", got, want)
	}
}

func TestScanTSLikeExports_ExportListAndDefault(t *testing.T) {
	src := "export { Foo, Bar as Baz }\nexport default service\n"
	got := scanTSLikeExports(src)
	want := []exportSymbol{
		{Name: "Foo", Kind: "value", Line: 1},
		{Name: "Baz", Kind: "value", Line: 1},
		{Name: "default", Kind: "default", Line: 2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("scanTSLikeExports() = %#v, want %#v", got, want)
	}
}

func TestToStringMap_Conversions(t *testing.T) {
	got, ok := toStringMap(map[interface{}]interface{}{
		"a": 1,
		2:   "skip",
	})
	if !ok {
		t.Fatal("toStringMap() expected ok=true for map[interface{}]interface{}")
	}
	if len(got) != 1 || got["a"] != 1 {
		t.Fatalf("toStringMap() = %#v, want map with key 'a' only", got)
	}

	direct, ok := toStringMap(map[string]interface{}{"x": "y"})
	if !ok || direct["x"] != "y" {
		t.Fatalf("toStringMap() direct map failed: ok=%v got=%#v", ok, direct)
	}

	if _, ok := toStringMap("nope"); ok {
		t.Fatal("toStringMap() expected ok=false for non-map input")
	}
}

func TestToStringSlice_TrimsAndFilters(t *testing.T) {
	got := toStringSlice([]interface{}{" one ", 42, "", "two"})
	want := []string{"one", "two"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("toStringSlice() = %#v, want %#v", got, want)
	}

	got = toStringSlice([]string{" alpha ", "", "beta"})
	want = []string{"alpha", "beta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("toStringSlice() = %#v, want %#v", got, want)
	}

	if out := toStringSlice(123); out != nil {
		t.Fatalf("toStringSlice() expected nil for invalid input, got %#v", out)
	}
}
