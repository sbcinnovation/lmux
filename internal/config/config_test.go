package config

import "testing"

func TestParseWindowsRejectsMultipleWindowNames(t *testing.T) {
	_, err := parseWindows([]any{map[string]any{"editor": "nvim", "server": "go run ."}})
	if err == nil {
		t.Fatal("parseWindows accepted a window entry with multiple names")
	}
}

func TestParsePanesRejectsMultipleTitles(t *testing.T) {
	_, err := parsePanes([]any{map[string]any{"editor": "nvim", "server": "go run ."}})
	if err == nil {
		t.Fatal("parsePanes accepted a pane entry with multiple titles")
	}
}
