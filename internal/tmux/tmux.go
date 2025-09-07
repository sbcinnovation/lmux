package tmux

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	cfg "github.com/sbcinnovation/lmux/internal/config"
)

// CheckTmuxInstalled verifies tmux is installed.
func CheckTmuxInstalled() error {
	if _, err := exec.LookPath("tmux"); err != nil {
		return errors.New("tmux not found in PATH")
	}
	return nil
}

// TmuxVersion returns the tmux version string.
func TmuxVersion() (string, error) {
	out, err := exec.Command("tmux", "-V").CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// StartProject creates a tmux session for the given project and optionally attaches.
func StartProject(project cfg.Project, attach bool) error {
	if err := CheckTmuxInstalled(); err != nil {
		return err
	}

	tmuxCmd := project.TmuxCommand
	if tmuxCmd == "" {
		tmuxCmd = "tmux"
	}

	// If session already exists, attach and return
	if hasSession(tmuxCmd, project.Name) {
		return runAttach(tmuxCmd, project.Name)
	}

	// Create detached session with first window
	firstWindowName := ""
	if len(project.Windows) > 0 {
		firstWindowName = project.Windows[0].Name
	}
	createArgs := []string{"new-session", "-d", "-s", project.Name}
	if firstWindowName != "" {
		createArgs = append(createArgs, "-n", firstWindowName)
	}
	if project.Root != "" {
		createArgs = append(createArgs, "-c", project.Root)
	}
	if project.TmuxOptions != "" {
		// Tmux options like -f need to be passed when invoking tmux, not subcommand
		// For simplicity we ignore custom tmux options here; TODO in future.
	}

	if err := run(tmuxCmd, createArgs...); err != nil {
		return fmt.Errorf("failed creating session: %w", err)
	}

	// Build subsequent windows
	for i, w := range project.Windows {
		if i == 0 {
			// already created with session
			if err := setupWindow(tmuxCmd, project.Name, i, w); err != nil {
				return err
			}
			continue
		}
		args := []string{"new-window", "-t", project.Name}
		if w.Name != "" {
			args = append(args, "-n", w.Name)
		}
		if w.Root != "" {
			args = append(args, "-c", w.Root)
		}
		if err := run(tmuxCmd, args...); err != nil {
			return fmt.Errorf("failed creating window %s: %w", w.Name, err)
		}
		if err := setupWindow(tmuxCmd, project.Name, i, w); err != nil {
			return err
		}
	}

	// Select startup window if specified
	if project.StartupWindow != "" {
		if err := run(tmuxCmd, "select-window", "-t", fmt.Sprintf("%s:%s", project.Name, project.StartupWindow)); err != nil {
			return err
		}
		if project.StartupPane > 0 {
			if err := run(tmuxCmd, "select-pane", "-t", fmt.Sprintf("%s:%s.%d", project.Name, project.StartupWindow, project.StartupPane)); err != nil {
				return err
			}
		}
	}

	if attach {
		return runAttach(tmuxCmd, project.Name)
	}
	return nil
}

func setupWindow(tmuxCmd, session string, index int, w cfg.Window) error {
	target := windowTarget(session, index, w)
	// Layout not fully supported; best effort
	if w.Layout != "" {
		_ = run(tmuxCmd, "select-layout", "-t", target, w.Layout)
	}

	// If panes specified, split and run commands per pane; otherwise run window commands
	if len(w.Panes) > 0 {
		// Ensure we have the right number of panes
		for i := 1; i < len(w.Panes); i++ {
			if err := run(tmuxCmd, "split-window", "-t", target, "-h"); err != nil {
				return err
			}
		}
		// Set tiled layout to spread panes
		_ = run(tmuxCmd, "select-layout", "-t", target, "tiled")

		// Run commands in each pane
		paneBase := getPaneBaseIndex(tmuxCmd)
		for paneIndex, pane := range w.Panes {
			if len(pane.Commands) == 0 {
				continue
			}
			for _, cmd := range pane.Commands {
				if err := run(tmuxCmd, "send-keys", "-t", fmt.Sprintf("%s.%d", target, paneBase+paneIndex), cmd, "C-m"); err != nil {
					return err
				}
			}
		}
	} else if len(w.Commands) > 0 {
		for _, cmd := range w.Commands {
			if err := run(tmuxCmd, "send-keys", "-t", target, cmd, "C-m"); err != nil {
				return err
			}
		}
	}
	return nil
}

// windowTarget returns a tmux target for the window, preferring name to avoid base-index issues.
func windowTarget(session string, index int, w cfg.Window) string {
	if strings.TrimSpace(w.Name) != "" {
		return fmt.Sprintf("%s:%s", session, w.Name)
	}
	// Fallback to numeric index (may fail if base-index != 0)
	return fmt.Sprintf("%s:%d", session, index)
}

func runAttach(tmuxCmd, session string) error {
	// If already inside tmux, switch client instead of attaching
	if os.Getenv("TMUX") != "" {
		return run(tmuxCmd, "switch-client", "-t", session)
	}
	c := exec.Command(tmuxCmd, "attach-session", "-t", session)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		// Likely non-interactive shell; print hint but do not fail
		fmt.Fprintf(os.Stderr, "Note: could not attach automatically. Run: tmux attach -t %s\n", session)
		return nil
	}
	return nil
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("%v: %s", err, strings.TrimSpace(stderr.String()))
		}
		return err
	}
	return nil
}

// getPaneBaseIndex returns tmux's pane-base-index (default 0 if unknown).
func getPaneBaseIndex(tmuxCmd string) int {
	out, err := exec.Command(tmuxCmd, "show", "-gv", "pane-base-index").CombinedOutput()
	if err != nil {
		return 0
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

// hasSession checks whether a tmux session exists
func hasSession(tmuxCmd, name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	cmd := exec.Command(tmuxCmd, "has-session", "-t", name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
