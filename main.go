package main

import (
	"fmt"
	"os"

	"idt/internal/cmd"
	"idt/internal/ui"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		ui.PrintExitBranding()
		os.Exit(1)
	}
}
