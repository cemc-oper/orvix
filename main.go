package main

import (
	"fmt"
	"os"

	"github.com/cemc-oper/orvix/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "orvix:", err)
		os.Exit(1)
	}
}
