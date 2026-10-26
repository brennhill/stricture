// exit.go — Helpers that print a CLI error to stderr and terminate with an exit code.
package cli

import (
	"fmt"
	"os"
)

// exitf prints "Error: <message>" to stderr and exits with code.
func exitf(code int, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(code)
}
