// Package version exposes build metadata set via -ldflags at build time.
package version

// These variables are set at build time via -ldflags by GoReleaser.
// Defaults allow local builds without GoReleaser.
var (
	Commit  = ""
	Date    = ""
	BuiltBy = ""
)
