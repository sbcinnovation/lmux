# 🚀 lmux: A Fast Simple Cross Platform tmux Session Manager 🖥️

An open source session manager for tmux, which allows users to manage tmux sessions through simple TOML configuration files, written in Go.

**Why lmux?** 🤔

- ⚡️ Blazingly fast.
  - [Go](https://github.com/golang/go) delivers fast performance and safety.
- 🌎 Cross Platform.
  - Run on Mac, Linux and Windows.
- 🎯 Dead simple.
  - Uses [TOML](https://github.com/toml-lang/toml) for simple configuration.
- 📄 One config = one session.

## Install

Build from source (requires Go 1.22+):

```bash
git clone https://github.com/sky/sbc-lmux
cd sbc-lmux
make install
OR
go build -o lmux ./cmd/lmux
# Place binary in PATH
```

## Quickstart.

1. Set your favourite editor, ex: `lmux editor zed`.
2. Init a config `lmux init session-name`.
3. Start a config `lmux start session-name`.

## Configuration

On macOS, lmux uses `~/.config/lmux` as the config directory.

- Create a project: `lmux init myproj` (opens in `$EDITOR`)
- Edit a project: `lmux edit myproj`
- List projects: `lmux list`
- Start a project: `lmux start myproj`
- Check environment: `lmux doctor`
- Set or show editor: `lmux editor [value]`

### Editor

- To set the editor globally:
  - `lmux editor nvim`
  - `lmux editor "code -w"`
- To show the current editor: `lmux editor`
- You can also set it on first edit: `lmux edit myproj --editor "nvim"`
- Resolution order when opening files:
  1. saved editor in `~/.config/lmux/settings.toml`
  2. `$EDITOR` (auto-saved on first use)
  3. macOS fallback `open -t` (auto-saved)

### TOML schema (simplified)

```toml
name = "myproj"
root = "~/dev/myproj"
attach = true # default true
tmux_command = "tmux" # optional
tmux_options = "-f ~/.tmux.conf" # parsed but not yet applied
startup_window = "1" # optional, index or name
startup_pane = 1 # optional, pane index

[[windows]]
editor.layout = "main-vertical"
editor.panes = ["vim", "bash"]

[[windows]]
server = "echo \"run your server here\""

[[windows]]
logs = "tail -f /var/log/system.log"

```

Notes:

- Window entries can be:
  - `name = "command"` inside an object in the `windows` array
  - `name = ["cmd1", "cmd2"]` (array of commands) inside an object
  - `name = { layout = L, root = PATH, panes = [...] }`
- Panes accept string (single command), array (multiple commands), or `{ title = commands }`.

## Differences from tmuxinator (for now)

- No ERB processing in TOML.
- Project/window hooks (on_project_start, pre_window, etc.) not implemented yet.
- `tmux_options` are parsed but not passed to tmux invocation yet.
- Layout handling is best-effort; panes default to tiled after splits.
- Wemux and socket options are not supported yet.
- Append-to-existing-session is not supported yet.

## Roadmap / TODO

- Apply `tmux_options` and socket options at tmux invocation time.
- Implement project/window hooks (on_project_start/stop/exit; pre_window).
- Add ERB-like variable interpolation or Go templating (optional).
- Improve layout support, synchronize panes before/after.
- Support selecting startup window/pane by name robustly.
- Implement stop/kill and restart commands.
- Support wemux/byobu.

## License

MIT
