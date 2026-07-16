// This file carries the repository's `go generate` directives — nothing else.
//
// The command surface (every verb, its purpose, its flags) is extracted from the live cobra
// tree and committed as web/src/data/surface.json, which the site's docs page renders from. A
// CI step re-runs this generate and fails on any diff, so the page can never drift from the
// binary: change a `Short:` or add a verb without regenerating, and the build reds. The
// snapshot is generated from a `go run` (an unstamped "dev" build), so its version field is
// always "dev" — deterministic for the gate; the site sources the real release version at
// deploy time, not from this file.
//
//go:generate go run ./cmd/sporo docs --out web/src/data/surface.json

package sporo
