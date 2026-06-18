package console

import "embed"

// Dist contains the built Admin Console SPA.
// Run `npm run build` in the console directory before building the Go binary.
//
//go:embed dist
var Dist embed.FS
