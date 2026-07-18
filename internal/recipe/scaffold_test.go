package recipe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// The scaffold's two load-bearing properties, asserted as hard as any gate: it is born a
// draft (so the tool never reds on its own output, and seal/export refuse it), and — minus
// the draft mark — it is GENRE-GREEN (a template that fails its own gate teaches the author
// the gate is noise).

func scaffoldWorld(t *testing.T) (root string, cfg Config) {
	t.Helper()
	root = t.TempDir()
	return root, Config{Home: ".sporo/recipes/", Products: []string{"otherproj"}}
}

func TestTheScaffoldIsBornADraftAndRefusedByExportAndSeal(t *testing.T) {
	root, cfg := scaffoldWorld(t)
	path, err := Scaffold(root, cfg, "my-capability", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !IsDraft(src) {
		t.Fatal("the scaffold must declare itself a draft — otherwise the gate reds on the state the tool itself wrote")
	}
	if _, err := Seal(root, cfg, "my-capability"); err == nil || !strings.Contains(err.Error(), "draft") {
		t.Fatalf("seal must refuse a draft: %v", err)
	}
	corpus := fstest.MapFS{
		"recipes/_adoption.md": {Data: []byte("<!-- SSOT SOURCE -->\n## Adopt it here\nprobe\n## Report back\nscars\n")},
	}
	if _, err := Export(corpus, filepath.Join(root, cfg.Home), "my-capability"); err == nil || !strings.Contains(err.Error(), "draft") {
		t.Fatalf("export must refuse a draft — a stranger must never receive TODOs as if they were earned: %v", err)
	}
}

func TestTheScaffoldMinusTheDraftMarkIsGenreGreen(t *testing.T) {
	root, cfg := scaffoldWorld(t)
	path, err := Scaffold(root, cfg, "my-capability", "A capability worth having twice", nil)
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	finished := strings.Replace(string(src), "draft: true\n", "", 1)
	if f := Lint("my-capability.md", []byte(finished), cfg.Products); len(f) != 0 {
		t.Fatalf("the scaffold is the first conformant document the author reads — it may not fail its own gate:\n%v", f)
	}
	// The id is minted, not typed — so it must be present and a real ULID the moment the draft
	// is born, or the author hits a gate on a field they were never meant to touch. Lint above
	// already enforces the ULID grammar (a bad id would red it); here we assert it is minted
	// and of ULID length.
	if id := fmValue(src, "id"); len(id) != 26 {
		t.Fatalf("the scaffold must mint a 26-char ULID id, got %q", id)
	}
}

func TestTheScaffoldNeverOverwrites(t *testing.T) {
	root, cfg := scaffoldWorld(t)
	if _, err := Scaffold(root, cfg, "my-capability", "", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := Scaffold(root, cfg, "my-capability", "", nil); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("a second scaffold over a real draft would eat the author's work: %v", err)
	}
}

func TestAHarvestSeedsTheScarsAsJudgeableCandidates(t *testing.T) {
	root, cfg := scaffoldWorld(t)
	h := &Harvest{
		Scars: []Candidate{
			{Hash: "abc123def456", Subject: "the check that could not fire", Signals: []string{"fix"}},
			{Hash: "fed654cba321", Subject: "a day that was actually two days", Signals: []string{"sabotage-found"}},
		},
		Design: []Candidate{{Hash: "aaa111bbb222", Subject: "collector emits one normalized record"}},
		Absent: []string{"decision log — rationale must be sourced by hand"},
	}
	path, err := Scaffold(root, cfg, "seeded", "", h)
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(src)
	for _, want := range []string{
		"the check that could not fire",
		"a day that was actually two days",
		"collector emits one normalized record",
		"decision log — rationale must be sourced by hand",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("the harvest's %q must land in the draft where the author will judge it", want)
		}
	}
	scars := sectionBody(strings.Split(s, "\n"), "## The scars")
	headings := 0
	for _, l := range scars {
		if strings.HasPrefix(l, "### ") {
			headings++
		}
	}
	if headings != 2 {
		t.Fatalf("each scar candidate gets its own heading (got %d)", headings)
	}
	for _, m := range []string{"Symptom", "Root cause", "Fix"} {
		if n := strings.Count(strings.Join(scars, "\n"), "**"+m+":**"); n != 2 {
			t.Fatalf("every candidate carries the three markers as stubs (%s: %d)", m, n)
		}
	}
}
