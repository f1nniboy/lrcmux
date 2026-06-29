// this package exposes the built SvelteKit assets as an embedded
// filesystem so we can bundle the site in the Go binary
package frontend

import "embed"

//go:embed all:build
var BuildFS embed.FS
