// Package version exposes build metadata set via -ldflags at build time.
package version

// These variables are set at build time via -ldflags by GoReleaser.
// Defaults allow local builds without GoReleaser.
var (
	Version = "0.1.0"
	Commit  = ""
	Date    = ""
	BuiltBy = ""
)
