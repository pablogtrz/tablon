//go:build !embed

package tablon

import "embed"

// Frontend is empty in normal development and test builds. Build with
// `pnpm --dir frontend build` and `-tags embed` to serve the SPA from the binary.
var Frontend embed.FS
