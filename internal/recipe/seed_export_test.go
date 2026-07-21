package recipe

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// update regenerates the committed golden. A runner-preamble edit legitimately changes every seed
// export, so regenerating is the deliberate step that parallels the seed ledger's pin test reding
// — `go test ./internal/recipe -run Golden -update`.
var update = flag.Bool("update", false, "rewrite the seed-export golden file")

const goldenPath = "testdata/seed_export.golden.md"

// fixtureSeed is a genre-shaped install seed used to exercise the composer. It deliberately carries
// `project: sporo` in its `verified:` frontmatter — provenance, not an internal reference — so the
// self-containment assertion below proves it targets internal PATHS/coordinates, never a naive
// "sporo" substring that would false-red on a legitimate provenance stamp (a real seed, S5's
// a2ahub, carries exactly this).
const fixtureSeed = `<!-- SSOT SOURCE (sporo repo). Consumers receive a provenance-stamped copy — edit HERE, never the synced copies. -->

---
id: 01J9ZC3Q0K2X7V8B4N6M5T1A2W
name: widget
version: 1.0.0
title: Install the widget CLI
target: widget@1.4.0
source: https://example.com/widget
stack: { language: go, runtime: any }
verified: { project: sporo, release: v0.12.0, date: 2026-07-21 }
effort: quick
---

## Summary

The widget CLI turns a repository's build graph into a single reproducible command. This seed
stands it up at a pinned version and proves it runs on the reader's own tree.

## What it is

A small Go binary that reads a build manifest and executes its targets in dependency order.

## Install

### Detect, then install

**Detect:** run ` + "`widget --version`" + ` and read the output — if it already prints 1.4.0, skip to Verify.
Otherwise fetch the pinned release from the declared source and put it on PATH.

**Done when:** ` + "`widget --version`" + ` prints 1.4.0.

## Verify

` + "```" + `
widget --version
` + "```" + `

## Use

Run ` + "`widget build`" + ` from the repository root; it reads the manifest and runs each target.

## Harness

If this repository wires tools into an agent harness, add ` + "`widget build`" + ` as its build gate.

## Report

| field | value |
|---|---|
| what it is | the widget build runner |
| how it works | reads a manifest, runs targets in order |
| what was done | installed 1.4.0, proved it runs |
| how to use it | ` + "`widget build`" + ` from the root |
| suggest next | wire it into the gate |
`

// realCorpus is the embedded runner preamble (read from source, the SSOT) beside the fixture seed.
// Composing the REAL _runner.md — not a fixture one — is load-bearing: this golden is the ONLY
// guard that the shipped preamble stays self-contained (REQ-3). The seed ledger's pin test proves
// byte-stability per version, not neutrality, and `sporo lint` holds a `_`-doc banner-only, so a
// version-bumped runner that leaked an internal path would slip every other check.
func realCorpus(t *testing.T) fstest.MapFS {
	t.Helper()
	runner, err := os.ReadFile("../../seeds/_runner.md")
	if err != nil {
		t.Fatal(err)
	}
	return fstest.MapFS{
		"seeds/_runner.md": {Data: runner},
		"seeds/widget.md":  {Data: []byte(fixtureSeed)},
	}
}

func TestSeedExportGoldenIsByteStable(t *testing.T) {
	got, err := SeedExport(realCorpus(t), "", "widget")
	if err != nil {
		t.Fatal(err)
	}
	if *update {
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("seed export drifted from the golden (%s) — if this change is intended, regenerate with `-update`:\n--- got ---\n%s", goldenPath, got)
	}
}

// The self-containment contract (REQ-3, VAL-1): the handover file names nothing a stranger's
// repository lacks. The check targets internal PATHS and meta-doc coordinates — never a naive
// "sporo" substring, which would false-red on the legitimate `project: sporo` provenance stamp the
// positive assertion below proves survives.
func TestExportedSeedIsSelfContained(t *testing.T) {
	got, err := SeedExport(realCorpus(t), "", "widget")
	if err != nil {
		t.Fatal(err)
	}
	for _, ref := range []string{".sporo/", "_runner.md", "_authoring.md", "internal/recipe", "web/src"} {
		if strings.Contains(got, ref) {
			t.Errorf("exported seed leaks an internal reference %q — the handover file must stand alone in a repo that has never heard of this harness:\n%s", ref, got)
		}
	}
	// Provenance is NOT an internal reference. `project: sporo` in the retained frontmatter is the
	// stamp saying which install proved the seed; it must survive, and a naive substring check that
	// stripped it would break every real seed. This positive assert reds if someone regresses to one.
	if !strings.Contains(got, "project: sporo") {
		t.Errorf("the seed's `verified:` provenance stamp must survive export (agents read it) — got:\n%s", got)
	}
}

func TestSeedExportComposesPreambleFirstBannerFree(t *testing.T) {
	got, err := SeedExport(realCorpus(t), "", "widget")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "SSOT SOURCE") {
		t.Fatalf("the provenance banner is house business about a repo the reader does not have — it must be stripped:\n%s", got)
	}
	if !strings.Contains(got, "> **Runner protocol:** v1.0.0") {
		t.Fatalf("the export must carry the version-stamped runner preamble (INV-2):\n%s", got)
	}
	// REQ-3: the preamble tells the agent to close by filling the seed's own Report.
	if !strings.Contains(got, "`## Report`") {
		t.Fatalf("the runner preamble must instruct the agent to complete the seed's `## Report`:\n%s", got)
	}
	// The frame arrives BEFORE the instruction: the discipline for running a seed safely must
	// precede its first Install step, or the agent reads it too late to have mattered.
	if strings.Index(got, "Runner protocol") > strings.Index(got, "## Install") {
		t.Fatalf("the runner preamble must come before the seed body — a frame read after the fact is no frame:\n%s", got)
	}
	// The seed's own frontmatter is retained (agents read target/source); the runner's meta-doc
	// frontmatter is dropped.
	if !strings.Contains(got, "target: widget@1.4.0") {
		t.Fatalf("the seed body's frontmatter must be retained — the agent reads `target`/`source`:\n%s", got)
	}
	if strings.Contains(got, "name: _runner") {
		t.Fatalf("the runner meta-doc's own frontmatter is house business, not for the reader:\n%s", got)
	}
}

func TestSeedExportRefusesTheGenreMetaDocument(t *testing.T) {
	if _, err := SeedExport(realCorpus(t), "", "_authoring"); err == nil {
		t.Fatal("exporting the seed genre's meta-document to someone who wants the tool helps nobody — it must refuse")
	}
}

func TestSeedExportRefusesADraft(t *testing.T) {
	home := t.TempDir()
	draft := "<!-- SSOT SOURCE -->\n---\nname: widget\ndraft: true\nversion: 0.1.0\n---\n## Summary\nwip\n"
	if err := os.WriteFile(filepath.Join(home, "widget.md"), []byte(draft), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := SeedExport(realCorpus(t), home, "widget"); err == nil || !strings.Contains(err.Error(), "draft") {
		t.Fatalf("a draft has no earned content to hand over — export must refuse with the fix, got: %v", err)
	}
}

// The project's OWN seed home wins over the embedded corpus — the local build is the one its
// author verified. A distinct marker in the local file proves the home copy, not the corpus, was
// exported.
func TestSeedExportPrefersTheProjectHome(t *testing.T) {
	home := t.TempDir()
	local := strings.Replace(fixtureSeed, "the widget build runner", "the LOCAL widget build runner", 1)
	if err := os.WriteFile(filepath.Join(home, "widget.md"), []byte(local), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := SeedExport(realCorpus(t), home, "widget")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "the LOCAL widget build runner") {
		t.Fatalf("the project's own seed home must be searched before the embedded corpus:\n%s", got)
	}
}

// Fail CLOSED. A corpus that lost the runner preamble would otherwise hand a stranger a seed with
// no frame for running it safely — refuse, name the fix.
func TestSeedExportFailsClosedWithoutTheRunnerPreamble(t *testing.T) {
	corpus := fstest.MapFS{"seeds/widget.md": {Data: []byte(fixtureSeed)}}
	if _, err := SeedExport(corpus, "", "widget"); err == nil || !strings.Contains(err.Error(), RunnerDoc) {
		t.Fatalf("a seed export with no runner preamble has no safe consumption frame — it must refuse, got: %v", err)
	}
}

func TestSeedExportOnAnUnknownSlugIsAnError(t *testing.T) {
	if _, err := SeedExport(realCorpus(t), "", "no-such-seed"); err == nil {
		t.Fatal("exporting a seed that does not exist must be an error, not an empty compose")
	}
}

func TestRunnerVersionComesFromItsFrontmatter(t *testing.T) {
	if got, err := RunnerVersion(realCorpus(t)); err != nil || got != "1.0.0" {
		t.Fatalf("runner version: got %q, err %v", got, err)
	}
	if _, err := RunnerVersion(fstest.MapFS{}); err == nil {
		t.Fatal("a corpus with no runner preamble must error, not return an empty version")
	}
}
