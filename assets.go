// Package tablon exposes embedded assets used by the server binary.
package tablon

import _ "embed"

// OpenAPI is the machine-readable API contract served at /openapi.yaml.
//
//go:embed docs/api/openapi.yaml
var OpenAPI []byte
