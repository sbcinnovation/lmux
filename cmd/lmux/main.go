package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	cfg "github.com/sbcinnovation/lmux/internal/config"
	"github.com/sbcinnovation/lmux/internal/tmux"
	"github.com/sbcinnovation/lmux/internal/util"
	buildinfo "github.com/sbcinnovation/lmux/internal/version"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "lmux",
		Short: "lmux: simple tmux project runner",
		Long:  "lmux is a lightweight tmux project runner inspired by tmuxinator.",
	}
	rootCmd.SetHelpTemplate(helpTemplate)

	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newDoctorCmd())
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newEditCmd())
	rootCmd.AddCommand(newEditorCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newStartCmd())
	rootCmd.AddCommand(newDetachCmd())
	rootCmd.AddCommand(newKillCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

const helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}Usage:
  {{.UseLine}}

{{if .HasAvailableSubCommands}}Available Commands:
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{- $name := .Name -}}
  {{- if .Aliases -}}
    {{- $name = printf "%s (%s)" .Name (join .Aliases ", ") -}}
  {{- end -}}
  {{rpad $name .NamePadding}} {{.Short}}
{{end}}{{end}}

{{end}}{{if .HasAvailableLocalFlags}}Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

{{end}}{{if .HasAvailableInheritedFlags}}Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}

{{end}}{{if .HasExample}}Examples:
{{.Example}}

{{end}}{{if .HasAvailableSubCommands}}
Use "{{.CommandPath}} [command] --help" for more information about a command.
{{end}}`

func newVersionCmd() *cobra.Command {
	var check bool
	var verbose bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print lmux version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				extra := []string{}
				if strings.TrimSpace(buildinfo.Commit) != "" {
					extra = append(extra, fmt.Sprintf("commit=%s", buildinfo.Commit))
				}
				if strings.TrimSpace(buildinfo.Date) != "" {
					extra = append(extra, fmt.Sprintf("date=%s", buildinfo.Date))
				}
				if strings.TrimSpace(buildinfo.BuiltBy) != "" {
					extra = append(extra, fmt.Sprintf("builtBy=%s", buildinfo.BuiltBy))
				}
				if len(extra) > 0 {
					fmt.Printf("%s (%s)\n", version, strings.Join(extra, ", "))
				} else {
					fmt.Println(version)
				}
			} else {
				fmt.Println(version)
			}

			if check {
				latest, update, url, err := util.CheckForUpdate("sbcinnovation/lmux", version)
				if err != nil {
					return fmt.Errorf("update check failed: %w", err)
				}
				if update {
					if strings.TrimSpace(url) == "" {
						fmt.Printf("update available: %s -> %s\n", version, latest)
					} else {
						fmt.Printf("update available: %s -> %s\n%s\n", version, latest, url)
					}
				} else {
					fmt.Println("you're up to date")
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "check for newer release on GitHub")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show extended build info")
	return cmd
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

func newDetachCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "detach",
		Aliases: []string{"d"},
		Short:   "Detach the current tmux client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tmux.DetachClient()
		},
	}
}

func newKillCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "kill-server",
		Aliases: []string{"k"},
		Short:   "Kill the tmux server (all sessions)",
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)
			fmt.Fprint(os.Stdout, "Kill tmux server and all sessions? [y/N]: ")
			input, err := reader.ReadString('\n')
			if err != nil && !errors.Is(err, io.EOF) {
				return fmt.Errorf("confirmation failed: %w", err)
			}
			resp := strings.ToLower(strings.TrimSpace(input))
			if resp != "y" && resp != "yes" {
				fmt.Fprintln(os.Stdout, "aborted")
				return nil
			}
			return tmux.KillServer()
		},
	}
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
