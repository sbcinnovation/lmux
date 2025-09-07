package config

// SampleTOML is a minimal template inspired by tmuxinator's sample (TOML format).
// We do not support ERB; placeholders are replaced in code.
const SampleTOML = `# <%= path %>

name = "<%= name %>"
root = "~/"

# Optional tmux socket (not yet supported)
# socket_name = "foo"

# Project hooks (not yet supported in lmux)
# on_project_start = "command"
# on_project_first_start = "command"
# on_project_restart = "command"
# on_project_exit = "command"
# on_project_stop = "command"

# pre_window = "echo 'setup env'"   # not yet supported
# tmux_options = "-f ~/.tmux.conf"  # supported (parsed only)
# tmux_command = "tmux"              # supported
# startup_window = "1"               # supported (by index or name)
# startup_pane = 1                    # supported (by index)
# attach = true                       # supported

[[windows]]
editor.layout = "main-vertical"
editor.panes = ["vim", "bash"]

[[windows]]
server = "echo \"run your server here\""

[[windows]]
logs = "tail -f /var/log/system.log"
`
