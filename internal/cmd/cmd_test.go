package cmd

import (
	"bytes"
	"testing"
)

func TestRootCommand(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("expected rootCmd to be initialized")
	}
	if rootCmd.Use != "idt" {
		t.Errorf("expected rootCmd use 'idt', got %s", rootCmd.Use)
	}
	if rootCmd.Short == "" {
		t.Error("expected non-empty short description")
	}
	if rootCmd.Long == "" {
		t.Error("expected non-empty long description")
	}
}

func TestRootCmdHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing root cmd --help: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected help output, got empty string")
	}
}
