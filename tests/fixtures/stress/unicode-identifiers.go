// unicode-identifiers.go — Stress test: Unicode in identifiers and strings.
package stress

import "fmt"

type Ünïcödë struct {
	Nàme   string `json:"name"`
	Ëmäïl  string `json:"email"`
}

func NéwÜsér(n string) *Ünïcödë {
	return &Ünïcödë{Nàme: n, Ëmäïl: fmt.Sprintf("%s@example.com", n)}
}

// CJK identifiers
type 用户 struct {
	名前 string
}

func 新用户(name string) *用户 {
	return &用户{名前: name}
}

// Emoji in strings (not identifiers — Go doesn't allow emoji identifiers)
func Greet() string {
	return "Hello 🌍 World 🎉"
}
