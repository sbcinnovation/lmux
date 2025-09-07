// Package config provides project and user-level configuration handling for lmux.
package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Settings holds user-level configuration for lmux.
type Settings struct {
	Editor string `toml:"editor,omitempty"`
}

// settingsFilePath returns the path to the settings TOML file.
func settingsFilePath() (string, error) {
	dir, err := EnsureConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.toml"), nil
}

// LoadSettings loads settings from ~/.config/lmux/settings.toml. If the file
// doesn't exist, returns default (zero-value) settings and no error.
func LoadSettings() (Settings, error) {
	var s Settings
	path, err := settingsFilePath()
	if err != nil {
		return s, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s, nil
		}
		return s, err
	}
	if len(data) == 0 {
		return s, nil
	}
	if err := toml.Unmarshal(data, &s); err != nil {
		return s, err
	}
	return s, nil
}

// SaveSettings writes settings to ~/.config/lmux/settings.toml.
func SaveSettings(s Settings) error {
	path, err := settingsFilePath()
	if err != nil {
		return err
	}
	data, err := toml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
