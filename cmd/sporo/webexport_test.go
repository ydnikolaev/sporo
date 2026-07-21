package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"sporo.dev/sporo/internal/recipe"
)

// The seed mirror is the seed twin of web-mirror: per seed it commits the byte-for-byte
// `sporo seed export` form the site serves. Its load-bearing difference is absent-not-empty — with
// no seeds it must create NO directory, because git stores no empty dir and a committed-empty
// mirror is a directory a reader's fresh checkout would lack (which S4 must tolerate).

func mustRunner(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("../../seeds/_runner.md")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// Today's committed state: the embedded seed corpus is underscore-only (the runner preamble and the
// genre spec), so the mirror writes zero files and — the invariant that matters — creates no
// directory. This is what keeps `go generate` from committing an empty web/src/data/seeds.
func TestSeedMirrorCreatesNoDirectoryWithZeroSeeds(t *testing.T) {
	corpus := fstest.MapFS{"seeds/_runner.md": {Data: mustRunner(t)}}
	out := filepath.Join(t.TempDir(), "seeds")
	n, err := writeSeedMirror(corpus, "", out)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("wrote %d forms, want 0 — an underscore-only corpus has no seeds to mirror", n)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("the seed mirror dir must stay ABSENT (not empty) until a seed exists — git stores no empty dir: %v", err)
	}
}

// A seed in the corpus is mirrored to one `<slug>.md`, byte-for-byte its SeedExport form — runner
// preamble first, then the banner-stripped body — and the underscore runner is never itself a form.
func TestSeedMirrorWritesTheExportFormPerSeed(t *testing.T) {
	corpus := fstest.MapFS{
		"seeds/_runner.md": {Data: mustRunner(t)},
		"seeds/widget.md":  {Data: []byte(seedMirrorFixture)},
	}
	out := filepath.Join(t.TempDir(), "seeds")
	n, err := writeSeedMirror(corpus, "", out)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("wrote %d forms, want 1 (widget — never the `_`-prefixed runner)", n)
	}
	want, err := recipe.SeedExport(corpus, "", "widget")
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(out, "widget.md"))
	if err != nil {
		t.Fatalf("mirror is missing widget.md: %v", err)
	}
	if string(got) != want {
		t.Fatalf("widget.md is not byte-for-byte the SeedExport form — the whole point of committing it is that it is:\n got: %q\nwant: %q", got, want)
	}
	// Preamble-first: the runner frame precedes the seed body's first section.
	if strings.Index(string(got), "Runner protocol") > strings.Index(string(got), "## Install") {
		t.Fatalf("the composed mirror must be preamble-first — the runner frame before the seed body:\n%s", got)
	}
}

const seedMirrorFixture = `<!-- SSOT SOURCE -->

---
name: widget
version: 1.0.0
target: widget@1.4.0
source: https://example.com/widget
---

## Install

Install the widget CLI at the pinned version.

**Done when:** ` + "`widget --version`" + ` prints 1.4.0.
`
