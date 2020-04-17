package main

import (
	"fmt"
	"os"

	"github.com/sector-f/wireconnect/cmd/wireconnect/cmd"
)

func main() {
	rootCmd := cmd.Root()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
