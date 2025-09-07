package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Project represents the lmux project configuration.
// This is a simplified schema inspired by tmuxinator.
type Project struct {
	Name          string `toml:"name"`
	Root          string `toml:"root"`
	Attach        *bool  `toml:"attach,omitempty"`
	TmuxCommand   string `toml:"tmux_command,omitempty"`
	TmuxOptions   string `toml:"tmux_options,omitempty"`
	StartupWindow string `toml:"startup_window,omitempty"`
	StartupPane   int    `toml:"startup_pane,omitempty"`
	WindowsRaw    []any  `toml:"windows"`

	// Normalized
	Windows []Window `toml:"-"`
}

// Window is the normalized representation after parsing WindowsRaw.
type Window struct {
	Name     string
	Layout   string
	Root     string
	Commands []string
	Panes    []Pane
}

// Pane represents commands inside a window split. Title is optional and not used yet.
type Pane struct {
	Title    string
	Commands []string
}

// EnsureConfigDir returns the lmux config directory path, creating it if needed.
// On macOS we use ~/.config/lmux as requested.
func EnsureConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "lmux")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// ProjectFilePath returns the expected path for a project toml file.
func ProjectFilePath(name string) string {
	dir, _ := EnsureConfigDir()
	return filepath.Join(dir, fmt.Sprintf("%s.toml", name))
}

// LoadProject loads and parses a project by name from the config directory.
func LoadProject(name string) (Project, error) {
	var project Project
	path := ProjectFilePath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		return project, err
	}
	if err := toml.Unmarshal(data, &project); err != nil {
		return project, err
	}
	if err := normalizeProject(&project); err != nil {
		return project, err
	}
	return project, nil
}

// SaveSample writes a sample project file with provided name and workingDir.
func SaveSample(name, workingDir string, force bool) (string, error) {
	path := ProjectFilePath(name)
	if !force {
		if _, err := os.Stat(path); err == nil {
			return "", fmt.Errorf("file exists: %s (use --force to overwrite)", path)
		}
	}
	content := strings.ReplaceAll(SampleTOML, "<%= name %>", name)
	if workingDir == "" {
		workingDir = "~/"
	}
	// Keep the path on a commented line so TOML stays valid
	content = strings.ReplaceAll(content, "# <%= path %>", "# "+workingDir)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func normalizeProject(p *Project) error {
	// Defaults
	if p.Attach == nil {
		t := true
		p.Attach = &t
	}
	if p.TmuxCommand == "" {
		p.TmuxCommand = "tmux"
	}

	windows, err := parseWindows(p.WindowsRaw)
	if err != nil {
		return err
	}
	p.Windows = windows

	if len(p.Windows) == 0 {
		return errors.New("project must have at least one window")
	}
	return nil
}

func parseWindows(raw []any) ([]Window, error) {
	result := make([]Window, 0, len(raw))
	for _, item := range raw {
		// Each item is expected to be a map with a single key: window name -> value
		m, ok := item.(map[string]any)
		if !ok || len(m) == 0 {
			return nil, fmt.Errorf("invalid window entry: %T", item)
		}
		var name string
		var value any
		for k, v := range m {
			name = k
			value = v
			break
		}

		win := Window{Name: name}
		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				win.Commands = []string{v}
			}
		case []any:
			// Interpret as multiple commands to run in the sole pane
			cmds := make([]string, 0, len(v))
			for _, c := range v {
				if s, ok := c.(string); ok && strings.TrimSpace(s) != "" {
					cmds = append(cmds, s)
				}
			}
			win.Commands = cmds
		case map[string]any:
			// Structured window: layout/root/panes
			if layout, ok := v["layout"].(string); ok {
				win.Layout = layout
			}
			if root, ok := v["root"].(string); ok {
				win.Root = root
			}
			if panesRaw, ok := v["panes"]; ok {
				panes, err := parsePanes(panesRaw)
				if err != nil {
					return nil, err
				}
				win.Panes = panes
			}
			// If top-level string command provided (e.g., { server: "rails s" }) that's handled above.
		default:
			return nil, fmt.Errorf("unsupported window value type: %T", value)
		}

		result = append(result, win)
	}
	return result, nil
}

func parsePanes(raw any) ([]Pane, error) {
	arr, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("panes must be an array, got %T", raw)
	}
	panes := make([]Pane, 0, len(arr))
	for _, p := range arr {
		switch v := p.(type) {
		case string:
			panes = append(panes, Pane{Commands: []string{v}})
		case []any:
			cmds := make([]string, 0, len(v))
			for _, c := range v {
				if s, ok := c.(string); ok && strings.TrimSpace(s) != "" {
					cmds = append(cmds, s)
				}
			}
			panes = append(panes, Pane{Commands: cmds})
		case map[string]any:
			// { title: commands }
			var title string
			var commands any
			for k, vv := range v {
				title = k
				commands = vv
				break
			}
			pane := Pane{Title: title}
			switch cv := commands.(type) {
			case string:
				pane.Commands = []string{cv}
			case []any:
				cmds := make([]string, 0, len(cv))
				for _, c := range cv {
					if s, ok := c.(string); ok && strings.TrimSpace(s) != "" {
						cmds = append(cmds, s)
					}
				}
				pane.Commands = cmds
			default:
				return nil, fmt.Errorf("invalid pane commands: %T", commands)
			}
			panes = append(panes, pane)
		default:
			return nil, fmt.Errorf("invalid pane entry: %T", p)
		}
	}
	return panes, nil
}
