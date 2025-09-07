
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
