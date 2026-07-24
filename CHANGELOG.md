# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project adheres to Semantic Versioning.

## [Unreleased]

## [1.1.0]

### Added

- `--root` flag override on `start`.
- Kill-all commands for tmux sessions; show remaining active projects after kill.

### Fixed

- Killing a session/project no longer kills all sessions.
- Attach flag respects project config instead of defaulting to attach.
- Config path traversal via malformed project names.
- SemVer prerelease comparison in `version --check`.
- Custom tmux command support and safer editor invocation.

## [1.0.0]

### Added

- Homebrew tap and Scoop bucket packaging via GoReleaser.
- Linux packages (`.deb`, `.rpm`) via GoReleaser.
- `lmux version --check` to query latest GitHub release.
- Core commands: `init`, `edit`, `editor`, `list`, `start`, `doctor`.

## [0.8.0]

### Added

- Initial public release of `lmux`.
- Cross-platform binaries for macOS, Linux, and Windows.
- TOML-based session configs and tmux integration.
