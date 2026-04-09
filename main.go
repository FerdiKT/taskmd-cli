package main

import (
	"fmt"
	"os"

	"github.com/ferdikt/taskmd-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
