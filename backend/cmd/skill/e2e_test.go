package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// contains reports whether substr is within s.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// captureStdout executes f and returns everything written to os.Stdout.
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestCLI_RootHelp verifies that --help lists all 14 custom commands.
func TestCLI_RootHelp(t *testing.T) {
	cli := NewCLI()
	buf := new(bytes.Buffer)
	cli.RootCmd.SetOut(buf)
	cli.RootCmd.SetArgs([]string{"--help"})
	err := cli.RootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	cmds := []string{
		"init", "config", "search", "list", "info",
		"install", "uninstall", "update", "sync", "edit",
		"doctor", "export", "import", "stats",
	}
	for _, cmd := range cmds {
		if !contains(output, cmd) {
			t.Errorf("help output missing command %q", cmd)
		}
	}
}

// TestCLI_InitDryRun verifies init --help output.
func TestCLI_InitDryRun(t *testing.T) {
	cli := NewCLI()
	buf := new(bytes.Buffer)
	cli.RootCmd.SetOut(buf)
	cli.RootCmd.SetArgs([]string{"init", "--help"})
	err := cli.RootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !contains(output, "init") {
		t.Error("init --help output missing 'init'")
	}
	if !contains(output, "Creates") {
		t.Error("init --help output missing 'Creates'")
	}
}

// TestCLI_ListEmpty verifies that list on an empty repo shows expected output.
func TestCLI_ListEmpty(t *testing.T) {
	repoDir := t.TempDir()

	output := captureStdout(func() {
		cli := NewCLI()
		cli.RootCmd.SetArgs([]string{"list", "--pool", repoDir})
		if err := cli.RootCmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})
	if !contains(output, "No skills installed") {
		t.Errorf("expected 'No skills installed', got: %s", output)
	}
}

// TestCLI_ConfigGetSet verifies config get commands against a temp repo.
func TestCLI_ConfigGetSet(t *testing.T) {
	repoDir := t.TempDir()

	// config get repo_path
	output := captureStdout(func() {
		cli := NewCLI()
		cli.RootCmd.SetArgs([]string{"config", "get", "repo_path", "--pool", repoDir})
		if err := cli.RootCmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})
	if output == "" {
		t.Error("expected non-empty output for 'config get repo_path'")
	}

	// config get install_mode
	output2 := captureStdout(func() {
		cli2 := NewCLI()
		cli2.RootCmd.SetArgs([]string{"config", "get", "install_mode", "--pool", repoDir})
		if err := cli2.RootCmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})
	if output2 == "" {
		t.Error("expected non-empty output for 'config get install_mode'")
	}
}

// TestCLI_DoctorHelp verifies doctor --help output.
func TestCLI_DoctorHelp(t *testing.T) {
	cli := NewCLI()
	buf := new(bytes.Buffer)
	cli.RootCmd.SetOut(buf)
	cli.RootCmd.SetArgs([]string{"doctor", "--help"})
	err := cli.RootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !contains(output, "doctor") {
		t.Error("doctor --help output missing 'doctor'")
	}
}

// TestCLI_StatsHelp verifies stats --help output.
func TestCLI_StatsHelp(t *testing.T) {
	cli := NewCLI()
	buf := new(bytes.Buffer)
	cli.RootCmd.SetOut(buf)
	cli.RootCmd.SetArgs([]string{"stats", "--help"})
	err := cli.RootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !contains(output, "stats") {
		t.Error("stats --help output missing 'stats'")
	}
}

// TestCLI_SyncHelp verifies sync --help output.
func TestCLI_SyncHelp(t *testing.T) {
	cli := NewCLI()
	buf := new(bytes.Buffer)
	cli.RootCmd.SetOut(buf)
	cli.RootCmd.SetArgs([]string{"sync", "--help"})
	err := cli.RootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !contains(output, "sync") {
		t.Error("sync --help output missing 'sync'")
	}
}

// TestCLI_UnknownCommand verifies that an unrecognized command returns an error.
func TestCLI_UnknownCommand(t *testing.T) {
	cli := NewCLI()
	cli.RootCmd.SetArgs([]string{"nonexistentcmd"})
	err := cli.RootCmd.Execute()
	if err == nil {
		t.Error("expected error for unknown command, got nil")
	}
}