package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKillCmdKillsOnlyProjectSession(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configDir := filepath.Join(home, ".config", "lmux")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	project := `name = "project-session"

[[windows]]
app = "true"
`
	if err := os.WriteFile(filepath.Join(configDir, "project.toml"), []byte(project), 0o644); err != nil {
		t.Fatal(err)
	}

	argsFile := filepath.Join(home, "tmux-args")
	t.Setenv("LMUX_TMUX_ARGS", argsFile)
	binDir := filepath.Join(home, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	tmuxScript := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$LMUX_TMUX_ARGS\"\n"
	if err := os.WriteFile(filepath.Join(binDir, "tmux"), []byte(tmuxScript), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	input, err := os.CreateTemp(home, "input")
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	if _, err := input.WriteString("y\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := input.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	originalStdin := os.Stdin
	os.Stdin = input
	defer func() { os.Stdin = originalStdin }()

	cmd := newKillCmd()
	cmd.SetArgs([]string{"project"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(args), "kill-session\n-t\nproject-session\n"; got != want {
		t.Fatalf("tmux arguments = %q, want %q", strings.TrimSpace(got), strings.TrimSpace(want))
	}
}

func TestStartCmdHonorsDisabledAttachFromProject(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configDir := filepath.Join(home, ".config", "lmux")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	project := `name = "project-session"
attach = false

[[windows]]
app = "true"
`
	if err := os.WriteFile(filepath.Join(configDir, "project.toml"), []byte(project), 0o644); err != nil {
		t.Fatal(err)
	}

	argsFile := filepath.Join(home, "tmux-args")
	t.Setenv("LMUX_TMUX_ARGS", argsFile)
	binDir := filepath.Join(home, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	tmuxScript := "#!/bin/sh\nif [ \"$1\" = has-session ]; then exit 0; fi\nprintf '%s\\n' \"$@\" > \"$LMUX_TMUX_ARGS\"\n"
	if err := os.WriteFile(filepath.Join(binDir, "tmux"), []byte(tmuxScript), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	cmd := newStartCmd()
	cmd.SetArgs([]string{"project"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(argsFile); !os.IsNotExist(err) {
		t.Fatalf("existing session should not be attached when attach=false, got %v", err)
	}
}

func TestSanitizeNameRejectsPathTraversal(t *testing.T) {
	for _, name := range []string{"../outside", `..\outside`, ".", "..", ""} {
		if got := sanitizeName(name); got != "" {
			t.Errorf("sanitizeName(%q) = %q, want empty", name, got)
		}
	}
}
