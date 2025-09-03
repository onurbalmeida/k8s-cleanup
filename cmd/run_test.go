package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func Test_Root_Help_ShowsCommands(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"-h"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("root help execute: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"run", "version", "completion"} {
		if !strings.Contains(out, want) {
			t.Fatalf("root help missing %q\n%s", want, out)
		}
	}
}

func Test_Run_Help_ShowsFlags(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"run", "-h"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("run help execute: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"--dry-run", "--older-than", "--kind", "--namespace",
		"--all-namespaces", "--exclude-ns", "--label-selector",
		"--field-selector", "--completed", "--failed", "--evicted",
		"--protect", "--concurrency", "--output", "--audit-file",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("run help missing flag %q\n%s", want, out)
		}
	}
}
