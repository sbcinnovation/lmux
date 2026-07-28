package tmux

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cfg "github.com/sbcinnovation/lmux/internal/config"
)

func TestSendPaneCommandUsesEnterNotCtrlM(t *testing.T) {
	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args")
	bin := filepath.Join(dir, "fake-tmux")
	script := `#!/bin/sh
case "$1" in
has-session) exit 1 ;;
show) echo 0 ;;
*) printf '%s\n' "$@" >> "` + argsFile + `" ;;
esac
`
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	project := cfg.Project{
		Name:        "proj",
		TmuxCommand: bin,
		Windows: []cfg.Window{{
			Name:     "lazygit",
			Commands: []string{"lazygit"},
		}},
	}
	if err := StartProject(project, false); err != nil {
		t.Fatalf("StartProject returned %v", err)
	}

	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	got := string(args)
	if strings.Contains(got, "C-m") {
		t.Fatalf("send-keys should not use C-m on Windows/psmux; got %q", got)
	}
	if !strings.Contains(got, "Enter") {
		t.Fatalf("send-keys should submit with Enter; got %q", got)
	}
	if !strings.Contains(got, "lazygit") {
		t.Fatalf("send-keys should include command text; got %q", got)
	}
}

func TestSendPaneCommandStripsCarriageReturns(t *testing.T) {
	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args")
	bin := filepath.Join(dir, "fake-tmux")
	script := `#!/bin/sh
case "$1" in
has-session) exit 1 ;;
show) echo 0 ;;
*) printf '%s\n' "$@" >> "` + argsFile + `" ;;
esac
`
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	project := cfg.Project{
		Name:        "proj",
		TmuxCommand: bin,
		Windows: []cfg.Window{{
			Name:     "app",
			Commands: []string{"lazygit\r"},
		}},
	}
	if err := StartProject(project, false); err != nil {
		t.Fatalf("StartProject returned %v", err)
	}
	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(args), "lazygit\r") {
		t.Fatalf("carriage returns should be stripped from commands; got %q", string(args))
	}
}

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
