package main

import (
	"fmt"
	"os"

	"github.com/lee-mcfaul2/pii-tokenizer/tools/rotate-kmaster/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
