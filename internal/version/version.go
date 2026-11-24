// Package version carries build metadata injected with -ldflags at release time.
package version

var (
	// Version is the semantic version without the leading "v".
	Version = "dev"
	// Commit is the short git revision.
	Commit = "unknown"
	// Date is the UTC build timestamp.
	Date = "unknown"
)
