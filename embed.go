// Package sporo is the module-root package, and it exists for one reason: to EMBED the
// transferable-build corpus into the binary. A recipe's whole purpose is to be handed to an
// agent in a repository that has never heard of this tool, so `sporo export <slug>` must work
// from the binary alone — on a machine with no `.sporo/`, no checkout, and no network.
//
// `all:` is load-bearing, not decoration. A bare `//go:embed recipes` silently skips every
// file whose name begins with `_` — which is the genre's own shape spec (`_authoring.md`) and
// the adoption protocol (`_adoption.md`) that every export appends. Silently: the build
// succeeds and the corpus in the binary is simply missing the two files the whole genre rests
// on. `all:` keeps them in.
package sporo

import "embed"

// Recipes is the official corpus, compiled into the binary: the genre assets (`_authoring.md`,
// `_adoption.md`) and every recipe shipped with the tool.
//
//go:embed all:recipes
var Recipes embed.FS
