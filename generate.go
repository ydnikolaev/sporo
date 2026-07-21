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
// The second directive commits the export handover form of every recipe to
// web/src/data/exports/<slug>.md, which the site serves as the recipe `.md` mirror (its copy and
// download actions). Same discipline: the composition lives once in recipe.Export, `go generate`
// refreshes the committed forms, and `make check` reds if the mirror drifts from a fresh run — so
// the downloaded file can never diverge from what `sporo export` produces.
//
// The third directive is the seed twin of the second: it commits each seed's export form to
// web/src/data/seeds/<slug>.md, the byte-for-byte `sporo seed export` handover. It writes nothing
// today — the embedded seed corpus is underscore-only until a seed is sealed — and because git
// stores no empty directory, web/src/data/seeds stays ABSENT (not empty) until then; the drift
// gate is trivially clean for it. When a seed lands, the same discipline holds it to the binary.
//
//go:generate go run ./cmd/sporo docs --out web/src/data/surface.json
//go:generate go run ./cmd/sporo web-mirror --out web/src/data/exports
//go:generate go run ./cmd/sporo seed-mirror --out web/src/data/seeds

package sporo
