// complexity_test.go — Proves cyclomatic complexity counting and the ratchet catch violations.
package archtest

import "testing"

const complexitySrc = `package p

func simple() {}

func branches(a, b int) int {
	if a > 0 && b > 0 || a < -5 { // if + && + ||
		return 1
	}
	for i := 0; i < a; i++ { // for
		switch b {
		case 1: // case
		case 2: // case
		default:
		}
	}
	for range []int{} { // range
	}
	select {
	case <-make(chan int): // comm clause
	default:
	}
	return 0
}

func closure() {
	f := func(x int) {
		if x > 0 {
		}
	}
	f(1)
}
`

func TestFuncComplexities(t *testing.T) {
	got, err := FuncComplexities("p.go", []byte(complexitySrc))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int{
		"simple":   1,
		"branches": 1 + 1 + 2 + 1 + 2 + 1 + 1, // if, &&, ||, for, 2 cases, range, comm
		"closure":  1 + 1,                     // closure body counts toward its parent
	}
	for name, w := range want {
		if got[name] != w {
			t.Errorf("complexity of %s = %d, want %d", name, got[name], w)
		}
	}
}

func TestFuncComplexitiesMethodName(t *testing.T) {
	src := "package p\ntype T struct{}\nfunc (t *T) M() { if true {} }\n"
	got, err := FuncComplexities("p.go", []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if got["T.M"] != 2 {
		t.Errorf("method key T.M missing or wrong: %v", got)
	}
}

func TestCheckComplexity(t *testing.T) {
	funcs := map[string]int{
		"a.go:ok":     15,
		"a.go:big":    16,
		"b.go:legacy": 30,
		"b.go:grew":   32,
		"b.go:shrank": 20,
		"b.go:fits":   10,
	}
	ceilings := map[string]int{
		"b.go:legacy": 30, "b.go:grew": 31, "b.go:shrank": 25, "b.go:fits": 18, "b.go:gone": 20,
	}
	got := CheckComplexity(funcs, 15, ceilings)
	for _, want := range []string{
		"a.go:big has complexity 16, over the 15 limit",
		"b.go:grew rose to 32, past its ratchet ceiling 31",
		"b.go:shrank dropped to 20. Lower its ceiling from 25",
		"b.go:fits is 10 and fits the 15 limit",
		"ratchet entry b.go:gone matches no function",
	} {
		if !contains(got, want) {
			t.Errorf("missing finding %q in %v", want, got)
		}
	}
	if contains(got, "a.go:ok") || contains(got, "b.go:legacy") {
		t.Errorf("compliant or pinned function flagged: %v", got)
	}
}

func TestFuncComplexitiesMeasuresInitsAndPackageLiterals(t *testing.T) {
	src := []byte("package p\n\nvar Handler = func(a, b bool) bool {\n\tif a && b {\n\t\treturn true\n\t}\n\treturn false\n}\n\nfunc init() {\n\tif true {\n\t}\n}\n\nfunc init() {\n\tif true {\n\t}\n\tif true {\n\t}\n}\n")
	got, err := FuncComplexities("p.go", src)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int{"Handler": 3, "init": 2, "init#2": 3}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("complexity[%s] = %d, want %d (all: %v)", k, got[k], v, got)
		}
	}
}
