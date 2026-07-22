package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runSeed executes the `sporo seed` namespace in-process, returning stdout, stderr, and the run
// error — the seed subtree is a self-contained cobra command, so a fresh seedCmd() per call keeps
// the cases independent (a leaked flag value from one would silently steer the next).
func runSeed(t *testing.T, args ...string) (out, errOut string, err error) {
	t.Helper()
	cmd := seedCmd()
	var o, e bytes.Buffer
	cmd.SetOut(&o)
	cmd.SetErr(&e)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return o.String(), e.String(), err
}

// seedTestRoot lays down a project that declares a seed home. `otherproj` is the neutrality
// vocabulary — a word that appears NOWHERE in the scaffold body or coach comments, so the green
// lint below reds only on a real wiring bug, never on a product name colliding with the fixture
// (the seed scaffold test's own precedent).
func seedTestRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".sporo"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := "home: recipes/\nproject: otherproj\nhomes:\n  seed: seeds/\n"
	if err := os.WriteFile(filepath.Join(root, ".sporo", "config.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// The five verbs, exercised as one loop on a fixture corpus (AC-1): scaffold → finish → lint →
// seal → list → export. Each stage feeds the next, so a break anywhere in the seed namespace's
// wiring to the T3/T4/T5 engine reds this — the CLI's own end-to-end proof.
func TestSeedNamespaceEndToEnd(t *testing.T) {
	root := seedTestRoot(t)

	// new: a draft lands in the seed home.
	out, _, err := runSeed(t, "new", "widget", "--root", root)
	if err != nil {
		t.Fatalf("seed new: %v", err)
	}
	if !strings.Contains(out, "draft at") {
		t.Fatalf("seed new: expected a draft-at line, got %q", out)
	}
	seedPath := filepath.Join(root, "seeds", "widget.md")
	if _, err := os.Stat(seedPath); err != nil {
		t.Fatalf("seed new did not write the draft: %v", err)
	}

	// lint over the draft: skipped, and the corpus is green (a draft is never held to the gate).
	out, _, err = runSeed(t, "lint", "--root", root)
	if err != nil {
		t.Fatalf("seed lint (draft present): %v", err)
	}
	if !strings.Contains(out, "draft(s) not checked") {
		t.Fatalf("seed lint should report the unfinished draft as skipped, got %q", out)
	}

	// finish the draft — the scaffold minus its draft mark is seed-genre-green by construction.
	src, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatal(err)
	}
	finished := bytes.Replace(src, []byte("draft: true\n"), nil, 1)
	if err := os.WriteFile(seedPath, finished, 0o644); err != nil {
		t.Fatal(err)
	}

	// lint the finished seed: conformant and neutral.
	out, errOut, err := runSeed(t, "lint", "--root", root)
	if err != nil {
		t.Fatalf("seed lint (finished): %v\nstderr: %s", err, errOut)
	}
	if !strings.Contains(out, "conformant and neutral") {
		t.Fatalf("seed lint: expected a conformant summary, got %q (stderr %q)", out, errOut)
	}

	// seal: records the seed under kind seed in the registry.
	out, _, err = runSeed(t, "seal", "widget", "--root", root)
	if err != nil {
		t.Fatalf("seed seal: %v", err)
	}
	if !strings.Contains(out, "sporo seed seal: widget") {
		t.Fatalf("seed seal: unexpected output %q", out)
	}
	reg, err := os.ReadFile(filepath.Join(root, ".sporo", "registry.yaml"))
	if err != nil {
		t.Fatalf("seal wrote no registry: %v", err)
	}
	if !strings.Contains(string(reg), "kind: seed") {
		t.Fatalf("seal must record `kind: seed`, registry was:\n%s", reg)
	}

	// list: the project's own seed shows, tagged project.
	out, _, err = runSeed(t, "list", "--root", root)
	if err != nil {
		t.Fatalf("seed list: %v", err)
	}
	if !strings.Contains(out, "project") || !strings.Contains(out, "widget") {
		t.Fatalf("seed list: expected the project widget, got %q", out)
	}

	// export --stdout: the runner preamble first (its version stamp), then the seed body with its
	// fixed Report — self-contained, no file written under --stdout.
	out, _, err = runSeed(t, "export", "widget", "--stdout", "--root", root)
	if err != nil {
		t.Fatalf("seed export: %v", err)
	}
	if !strings.Contains(out, "Runner protocol:") {
		t.Fatalf("seed export must prepend the runner preamble, got %q", out)
	}
	if !strings.Contains(out, "## Report") {
		t.Fatalf("seed export must carry the seed body's Report section, got %q", out)
	}
	if strings.Index(out, "Runner protocol:") > strings.Index(out, "## Report") {
		t.Fatal("seed export inverts the recipe order: the runner preamble must come BEFORE the seed body")
	}
}

// export with no project seed home writes nothing under `.sporo/` — a stranger reading an embedded
// seed from a bare directory must never have a tree created underfoot (the read-verb contract).
func TestSeedExportFromABareDirectoryWritesNothing(t *testing.T) {
	root := t.TempDir()
	// The embedded corpus is meta-doc-only today, so exporting a named seed errors — but the
	// contract under test is the SIDE EFFECT: no `.sporo/` tree, whatever the outcome.
	_, _, _ = runSeed(t, "export", "nonesuch", "--root", root)
	if _, err := os.Stat(filepath.Join(root, ".sporo")); !os.IsNotExist(err) {
		t.Fatal("a read verb in a bare directory must not create a .sporo tree")
	}
}

// A project that declares no seed home gets a clean error from the lint verb, never a crash — a
// repo that authors recipes but not seeds simply has no seed corpus here (REQ-5).
func TestSeedLintWithoutASeedHomeErrorsCleanly(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".sporo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".sporo", "config.yaml"), []byte("home: recipes/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := runSeed(t, "lint", "--root", root)
	if err == nil {
		t.Fatal("seed lint must error when no seed home is declared, not silently pass or crash")
	}
	if !strings.Contains(err.Error(), "seed") {
		t.Fatalf("the no-seed-home error should point at seed tooling, got %v", err)
	}
}

// BL-006: the seed lint summary counts seed INSTANCES and `_`-prefixed genre meta-docs apart, so it
// agrees with `sporo seed list` (which shows only instances). A home with a seed and a meta-doc must
// report "1 seed(s) and 1 meta-doc(s)", never a folded "2 seed(s)".
func TestSeedLintCountsMetaDocsApartFromInstances(t *testing.T) {
	root := seedTestRoot(t)

	// A finished seed instance: scaffold, then drop the draft mark (the scaffold is conformant by
	// construction, so the corpus greens on wiring, not on a fixture's shape).
	if _, _, err := runSeed(t, "new", "widget", "--root", root); err != nil {
		t.Fatalf("seed new: %v", err)
	}
	seedPath := filepath.Join(root, "seeds", "widget.md")
	src, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(seedPath, bytes.Replace(src, []byte("draft: true\n"), nil, 1), 0o644); err != nil {
		t.Fatal(err)
	}
	// A `_`-prefixed meta-doc with a valid provenance banner: held to its banner alone, so it lints
	// clean though it has none of the seed sections — but it counts as a meta-doc, not a seed.
	banner := "<!-- SSOT SOURCE (otherproj). -->\n\n## Anything\n\nprose, not a seed\n"
	if err := os.WriteFile(filepath.Join(root, "seeds", "_notes.md"), []byte(banner), 0o644); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := runSeed(t, "lint", "--root", root)
	if err != nil {
		t.Fatalf("seed lint (seed + meta-doc): %v\nstderr: %s", err, errOut)
	}
	if !strings.Contains(out, "1 seed(s) and 1 meta-doc(s) conformant and neutral") {
		t.Fatalf("seed lint summary must count the meta-doc apart from the seed, got %q", out)
	}
}
