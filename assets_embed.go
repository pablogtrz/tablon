//go:build embed

package tablon

import "embed"

// Frontend contains the built Svelte SPA. Build it with
// `pnpm --dir frontend build` and compile with `-tags embed`.
//
//go:embed all:frontend/dist
var Frontend embed.FS
