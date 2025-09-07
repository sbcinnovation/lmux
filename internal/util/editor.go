// Package util contains helper utilities for lmux.
package util

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	cfg "github.com/sbcinnovation/lmux/lmux/internal/config"
)

// OpenInEditor opens the given file path in $EDITOR if set, otherwise uses open -t on macOS.
func OpenInEditor(path string) error {
	// Prefer persisted settings
	var editor string
	if settings, err := cfg.LoadSettings(); err == nil {
		editor = strings.TrimSpace(settings.Editor)
	}
	// If no persisted editor, fall back to $EDITOR and persist it for future runs
	if editor == "" {
		if env := strings.TrimSpace(os.Getenv("EDITOR")); env != "" {
			editor = env
			// Best-effort persistence; ignore errors
			if settings, err := cfg.LoadSettings(); err == nil {
				settings.Editor = editor
				_ = cfg.SaveSettings(settings)
			}
		}
	}
	if editor != "" {
		// Support EDITOR commands with arguments (e.g., "code -w")
		if strings.Contains(editor, " ") {
			cmd := exec.Command("sh", "-c", editor+" "+fmt.Sprintf("%q", path))
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}
		cmd := exec.Command(editor, path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	// macOS fallback: use `open -t` and persist this choice for next time
	if _, err := exec.LookPath("open"); err == nil {
		editor = "open -t"
		if settings, err := cfg.LoadSettings(); err == nil {
			settings.Editor = editor
			_ = cfg.SaveSettings(settings)
		}
		cmd := exec.Command("sh", "-c", editor+" "+fmt.Sprintf("%q", path))
		return cmd.Run()
	}
	return fmt.Errorf("no editor found; set $EDITOR or install 'open'")
}
