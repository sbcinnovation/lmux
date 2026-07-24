package tmux

import (
	"os"
	"path/filepath"
	"testing"

	cfg "github.com/sbcinnovation/lmux/internal/config"
)

func TestStartProjectUsesConfiguredTmuxCommand(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "custom-tmux")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	project := cfg.Project{
		Name:        "project",
		TmuxCommand: bin,
		Windows:     []cfg.Window{{Name: "app"}},
	}
	if err := StartProject(project, false); err != nil {
		t.Fatalf("StartProject returned %v", err)
	}
}
