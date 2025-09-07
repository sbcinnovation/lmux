package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	cfg "github.com/sky/sbc-lmux/lmux/internal/config"
	"github.com/sky/sbc-lmux/lmux/internal/tmux"
	"github.com/sky/sbc-lmux/lmux/internal/util"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "lmux",
		Short: "lmux: simple tmux project runner",
		Long:  "lmux is a lightweight tmux project runner inspired by tmuxinator.",
	}

	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newDoctorCmd())
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newEditCmd())
	rootCmd.AddCommand(newEditorCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newStartCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print lmux version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check environment for lmux usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check config dir
			dir, err := cfg.EnsureConfigDir()
			if err != nil {
				return err
			}
			fmt.Printf("config dir: %s\n", dir)

			// Check tmux presence and version
			if err := tmux.CheckTmuxInstalled(); err != nil {
				return err
			}
			ver, _ := tmux.TmuxVersion()
			if ver != "" {
				fmt.Printf("tmux version: %s\n", ver)
			}
			fmt.Println("doctor: OK")
			return nil
		},
	}
}

func newInitCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Create a new project TOML in ~/.config/lmux",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := sanitizeName(args[0])
			if name == "" {
				return errors.New("invalid project name")
			}
			// Fill template hints
			wd, _ := os.Getwd()
			path, err := cfg.SaveSample(name, wd, force)
			if err != nil {
				return err
			}
			fmt.Printf("created %s\n", path)

			// Open in editor if possible
			if err := util.OpenInEditor(path); err != nil {
				fmt.Printf("warning: could not open editor: %v\n", err)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite if file exists")
	return cmd
}

func newEditCmd() *cobra.Command {
	var editorFlag string
	cmd := &cobra.Command{
		Use:   "edit [name]",
		Short: "Open an existing project TOML in editor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := sanitizeName(args[0])
			path := cfg.ProjectFilePath(name)
			if _, err := os.Stat(path); err != nil {
				return fmt.Errorf("project not found: %s", path)
			}

			// If --editor is provided, persist it immediately
			if strings.TrimSpace(editorFlag) != "" {
				settings, _ := cfg.LoadSettings()
				settings.Editor = strings.TrimSpace(editorFlag)
				if err := cfg.SaveSettings(settings); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to save editor setting: %v\n", err)
				}
			}
			return util.OpenInEditor(path)
		},
	}
	cmd.Flags().StringVar(&editorFlag, "editor", "", "set and persist the editor to use (e.g. 'nvim', 'code -w')")
	return cmd
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List projects in ~/.config/lmux",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cfg.EnsureConfigDir()
			if err != nil {
				return err
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				return err
			}
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				name := e.Name()
				if strings.HasSuffix(name, ".toml") {
					fmt.Println(strings.TrimSuffix(name, ".toml"))
				}
			}
			return nil
		},
	}
}

func newStartCmd() *cobra.Command {
	var attach bool
	cmd := &cobra.Command{
		Use:   "start [name]",
		Short: "Start a tmux session for the project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := sanitizeName(args[0])
			project, err := cfg.LoadProject(name)
			if err != nil {
				return err
			}
			if project.Name == "" {
				project.Name = name
			}

			// Default to attach if not specified via flag
			if !cmd.Flags().Changed("attach") {
				attach = true
			}

			return tmux.StartProject(project, attach)
		},
	}
	cmd.Flags().BoolVar(&attach, "attach", true, "attach to the session after starting")
	return cmd
}

func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	name = strings.ReplaceAll(name, " ", "-")
	return name
}

// newEditorCmd provides `lmux editor` to get or set the editor.
// Usage:
//
//	lmux editor           # prints current editor
//	lmux editor nvim      # sets editor to 'nvim'
//	lmux editor "code -w"  # sets editor with args
func newEditorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "editor [command]",
		Short: "Get or set the editor used by lmux",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				settings, _ := cfg.LoadSettings()
				current := strings.TrimSpace(settings.Editor)
				if current == "" {
					current = strings.TrimSpace(os.Getenv("EDITOR"))
				}
				if current == "" {
					fmt.Println("(no editor configured)")
					return nil
				}
				fmt.Println(current)
				return nil
			}

			// Set editor
			ed := strings.TrimSpace(args[0])
			settings, _ := cfg.LoadSettings()
			settings.Editor = ed
			if err := cfg.SaveSettings(settings); err != nil {
				return err
			}
			fmt.Printf("editor set to: %s\n", ed)
			return nil
		},
	}
}
