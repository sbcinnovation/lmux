package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cfg "github.com/sbcinnovation/lmux/internal/config"
)

func TestOpenInEditorPassesArgumentsWithoutShell(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	argsFile := filepath.Join(home, "editor-args")
	t.Setenv("LMUX_EDITOR_ARGS", argsFile)
	binDir := filepath.Join(home, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	editorPath := filepath.Join(binDir, "editor")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$LMUX_EDITOR_ARGS\"\n"
	if err := os.WriteFile(editorPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := cfg.SaveSettings(cfg.Settings{Editor: editorPath + " --wait"}); err != nil {
		t.Fatal(err)
	}

	file := filepath.Join(home, "project.toml")
	if err := OpenInEditor(file); err != nil {
		t.Fatal(err)
	}
	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(args), "--wait\n"+file+"\n"; got != want {
		t.Fatalf("editor arguments = %q, want %q", strings.TrimSpace(got), strings.TrimSpace(want))
	}
}
