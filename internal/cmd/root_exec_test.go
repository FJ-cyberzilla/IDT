package cmd

import (
	"bytes"
	"testing"
)

func TestRootCmdExecutionPaths(t *testing.T) {
	t.Run("Help flag", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"--help"})

		err := Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if buf.Len() == 0 {
			t.Error("expected help output")
		}
	})

	t.Run("Unknown flag", func(t *testing.T) {
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"--unknown-flag-xyz"})

		err := Execute()
		if err == nil {
			t.Error("expected error for unknown flag, got nil")
		}
	})
}
