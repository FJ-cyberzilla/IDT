package cmd

import (
	"fmt"
	"os"

	"idt/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "idt",
	Short: "IDT - Professional Internet Dialer Test & Command Center",
	Long:  `IDT is a professional network testing, DNS diagnostics, and latency benchmarking command center built with Cobra and Lip Gloss.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("unknown command or argument %q", args[0])
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(ui.InitialModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running command center: %v\n", err)
			ui.PrintExitBranding()
			os.Exit(1)
		}
		ui.PrintExitBranding()
	},
}

func Execute() error {
	return rootCmd.Execute()
}
