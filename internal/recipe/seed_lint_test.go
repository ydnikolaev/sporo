package recipe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sporo.dev/sporo/pkg/recipekit"
)

// finishedSeedInHome drops a conformant, non-draft seed into a project's seed home — the scaffold's
// own minus-draft output, so the corpus the walk greens is exactly the document the scaffold-green
// property already proves conformant. This keeps the DEC-001 walk tests self-consistent with the
// scaffold and independent of the export fixture's (separately proven) greenness.
func finishedSeedInHome(t *testing.T, root string, cfg Config, slug string) string {
	t.Helper()
	path, err := SeedScaffold(root, cfg, slug, "Install the thing")
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	finished := strings.Replace(string(src), "draft: true\n", "", 1)
	if err := os.WriteFile(path, []byte(finished), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// DEC-001 at the single-document level: a conformant seed lints silent, a malformed one reds.
func TestLintSeedGreensAConformantSeedAndRedsAMalformedOne(t *testing.T) {
	root, cfg := seedScaffoldWorld(t)
	green, err := os.ReadFile(finishedSeedInHome(t, root, cfg, "my-tool"))
	if err != nil {
		t.Fatal(err)
	}
	if f := LintSeed("my-tool.md", green, cfg.Products); len(f) != 0 {
		t.Fatalf("a conformant seed must lint clean:\n%v", f)
	}
	malformed := []byte("<!-- SSOT SOURCE -->\n---\nname: broken\n---\n## Summary\ntoo short\n")
	if f := LintSeed("broken.md", malformed, cfg.Products); len(f) == 0 {
		t.Fatal("a seed missing keys, sections, and the trust contract must red — DEC-001")
	}
}

// DEC-001 through the walk: the seed corpus greens today.
func TestLintSeedHomeGreensAConformantCorpus(t *testing.T) {
	root, cfg := seedScaffoldWorld(t)
	finishedSeedInHome(t, root, cfg, "my-tool")
	findings, n, metas, drafts, err := LintSeedHome(root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("a conformant seed corpus must walk clean:\n%v", findings)
	}
	if n != 1 || metas != 0 || drafts != 0 {
		t.Fatalf("expected 1 seed checked, 0 meta-docs, 0 drafts; got n=%d metas=%d drafts=%d", n, metas, drafts)
	}
}

// DEC-001 through the walk: a malformed member surfaces its findings.
func TestLintSeedHomeRedsAMalformedSeedInTheCorpus(t *testing.T) {
	root, cfg := seedScaffoldWorld(t)
	home, _ := cfg.HomeFor(recipekit.KindSeed)
	dir := filepath.Join(root, home)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	malformed := "<!-- SSOT SOURCE -->\n---\nname: broken\n---\n## Summary\nnope\n"
	if err := os.WriteFile(filepath.Join(dir, "broken.md"), []byte(malformed), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, _, _, _, err := LintSeedHome(root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Fatal("the walk must surface a malformed seed's findings — DEC-001")
	}
}

// A born-draft scaffold is exempt: the walk must not red on the state `sporo seed new` writes.
func TestLintSeedHomeSkipsDrafts(t *testing.T) {
	root, cfg := seedScaffoldWorld(t)
	if _, err := SeedScaffold(root, cfg, "wip-tool", ""); err != nil {
		t.Fatal(err)
	}
	findings, n, metas, drafts, err := LintSeedHome(root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("a draft is exempt from the gate — the walk must not red on it:\n%v", findings)
	}
	if n != 0 || metas != 0 || drafts != 1 {
		t.Fatalf("expected 0 checked, 0 meta-docs, 1 draft; got n=%d metas=%d drafts=%d", n, metas, drafts)
	}
}

// The genre's own `_`-prefixed meta-documents are handed to the linter but held to their banner
// alone — a meta-doc carrying a valid banner passes though it has none of the seven seed sections,
// while one missing its banner still reds. This is the seed mirror of the recipe walk's isMeta path.
func TestLintSeedHomeHoldsMetaDocumentsToTheBanner(t *testing.T) {
	root, cfg := seedScaffoldWorld(t)
	home, _ := cfg.HomeFor(recipekit.KindSeed)
	dir := filepath.Join(root, home)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// A banner-bearing meta-doc with no seed sections: the `_`-prefix early return exempts it.
	banner := "<!-- SSOT SOURCE (otherproj). -->\n\n## Anything\n\nprose, not a seed\n"
	if err := os.WriteFile(filepath.Join(dir, "_notes.md"), []byte(banner), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, n, metas, _, err := LintSeedHome(root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("a `_`-prefixed meta-document is held to its banner alone — a valid banner must not red:\n%v", findings)
	}
	// The meta-doc is still handed to the linter (for its banner), so it counts as checked — but as
	// a meta-doc, not a seed instance, so it does not inflate the seed count `sporo seed list` shows.
	if n != 0 || metas != 1 {
		t.Fatalf("the meta-document counts as a checked meta-doc, not a seed; got n=%d metas=%d", n, metas)
	}
	// Now break the banner: the meta-doc is still linted, so the missing-banner finding surfaces.
	if err := os.WriteFile(filepath.Join(dir, "_notes.md"), []byte("## Anything\n\nno banner here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, _, _, _, err = LintSeedHome(root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Fatal("a meta-document missing its provenance banner must still red — the walk hands it to the linter")
	}
}

// A project that declares no seed home has no seed corpus — a stated absence, a clean error, never
// a crash (REQ-5).
func TestLintSeedHomeReportsAProjectWithNoSeedCorpus(t *testing.T) {
	root := t.TempDir()
	cfg := Config{Home: ".sporo/recipes/", Homes: map[string]string{recipekit.KindRecipe: ".sporo/recipes/"}}
	if _, _, _, _, err := LintSeedHome(root, cfg); err == nil {
		t.Fatal("a project with no declared seed home must return a clean error, never crash")
	}
}

// The seed-scoped seal-coherence sweep fires: a sealed seed whose content drifted without a version
// bump reds through the SHARED helper (T3), proving the sweep is wired into the walk. The tamper
// keeps the seed genre-green, so the ONLY finding is the coherence signal — asserted by message, so
// a lint red can never masquerade as the sweep firing.
func TestLintSeedHomeCoherenceSweepFiresOnATamperedSeal(t *testing.T) {
	root, cfg := seedScaffoldWorld(t)
	path := finishedSeedInHome(t, root, cfg, "my-tool")
	if _, err := SealKind(root, cfg, recipekit.KindSeed, "my-tool"); err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(string(src), "Write a short orientation", "Draft a short orientation", 1)
	if tampered == string(src) {
		t.Fatal("the tamper substitution changed nothing — the scaffold text this test edits has moved")
	}
	if err := os.WriteFile(path, []byte(tampered), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, _, _, _, err := LintSeedHome(root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	var drifted bool
	for _, f := range findings {
		if strings.Contains(f.Msg, "drifted") {
			drifted = true
		}
	}
	if !drifted {
		t.Fatalf("the seed-scoped coherence sweep must red a sealed seed whose content drifted without a version bump:\n%v", findings)
	}
}
