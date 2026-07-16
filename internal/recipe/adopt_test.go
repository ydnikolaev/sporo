package recipe

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The handover-side teeth. The two behaviors everything rests on: an adopted text never
// silently mutates (same posture as the seal, other end of the handover), and pull is
// read-only until told otherwise — discovering staleness must never BE the rebuild.

// exportedFixture is what a reader actually receives: frontmatter intact, no banner,
// protocol appended. Version and one exact-bound fence are parameterized so tests can move
// the source forward.
func exportedFixture(version, fenceField string) string {
	return `---
name: nightly-digest
version: ` + version + `
title: A nightly digest that checks itself
problem: The record of a day's work is invisible.
prerequisites: [read-files]
derived_from: [one live build]
stack: { language: go, runtime: any, why: "one static binary" }
verified: { project: elsewhere, release: v1.0.0, date: 2026-07-15 }
effort: an evening
---

# Nightly digest

## The contracts

The feed the fleet's aggregator parses — **Binding: exact**:

` + "```json\n" + `{ "schema": 1, "` + fenceField + `": 12 }` + "\n```" + `

## Adopt it here
probe first
## Report back
send scars
`
}

func adoptWorld(t *testing.T) (root string) {
	t.Helper()
	return t.TempDir()
}

func TestAdoptRecordsTheHandoverVerbatim(t *testing.T) {
	root := adoptWorld(t)
	src := []byte(exportedFixture("1.0.0", "counted"))
	slug, entry, err := Adopt(root, src, "/somewhere/handoff.md")
	if err != nil {
		t.Fatal(err)
	}
	if slug != "nightly-digest" || entry.Version != "1.0.0" || entry.Source != "/somewhere/handoff.md" {
		t.Fatalf("the entry must anchor slug, version and source; got %q %+v", slug, entry)
	}
	if entry.ExactContracts == "" {
		t.Fatal("the fixture carries an exact contract — its digest is what pull compares against")
	}
	stored, err := os.ReadFile(filepath.Join(root, adoptedHome, "nightly-digest.md"))
	if err != nil || string(stored) != string(src) {
		t.Fatalf("the stored copy must be verbatim — it is the only honest anchor: %v", err)
	}
}

func TestReAdoptingTheSameBytesIsIdempotent(t *testing.T) {
	root := adoptWorld(t)
	src := []byte(exportedFixture("1.0.0", "counted"))
	if _, _, err := Adopt(root, src, "a"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Adopt(root, src, "a"); err != nil {
		t.Fatalf("same bytes twice must be a no-op: %v", err)
	}
}

func TestADifferentTextUnderANonIncreasedVersionIsRefused(t *testing.T) {
	root := adoptWorld(t)
	if _, _, err := Adopt(root, []byte(exportedFixture("1.2.0", "counted")), "a"); err != nil {
		t.Fatal(err)
	}
	mutated := strings.Replace(exportedFixture("1.2.0", "counted"), `"schema": 1`, `"schema": 2`, 1)
	if _, _, err := Adopt(root, []byte(mutated), "a"); err == nil || !strings.Contains(err.Error(), "silently mutates") {
		t.Fatalf("a changed text under the same version is the handover-side silent mutation: %v", err)
	}
	if _, _, err := Adopt(root, []byte(exportedFixture("1.1.0", "counted")), "a"); err == nil {
		t.Fatal("...and a REGRESSION is refused too")
	}
	if _, _, err := Adopt(root, []byte(exportedFixture("2.0.0", "counted")), "a"); err != nil {
		t.Fatalf("a genuinely newer version adopts forward: %v", err)
	}
}

func TestAFragmentWithoutFrontmatterIsRefused(t *testing.T) {
	root := adoptWorld(t)
	if _, _, err := Adopt(root, []byte("# just some markdown\n"), ""); err == nil || !strings.Contains(err.Error(), "frontmatter") {
		t.Fatalf("adopt takes the exported file, not a fragment: %v", err)
	}
}

func TestPullReportsANewerSourceAndFlagsExactChanges(t *testing.T) {
	root := adoptWorld(t)
	source := filepath.Join(t.TempDir(), "handoff.md")
	if err := os.WriteFile(source, []byte(exportedFixture("1.0.0", "counted")), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Adopt(root, []byte(exportedFixture("1.0.0", "counted")), source); err != nil {
		t.Fatal(err)
	}

	// Prose-only move: version bumps, the exact fence is untouched.
	if err := os.WriteFile(source, []byte(exportedFixture("1.1.0", "counted")), 0o644); err != nil {
		t.Fatal(err)
	}
	reports, err := Pull(root, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 || reports[0].Status != "update" || reports[0].Latest != "1.1.0" || reports[0].ExactChanged {
		t.Fatalf("a prose-only bump is an update with the exact digest intact: %+v", reports)
	}

	// The loud case: the exact-bound fence changed.
	if err := os.WriteFile(source, []byte(exportedFixture("2.0.0", "renamed")), 0o644); err != nil {
		t.Fatal(err)
	}
	reports, err = Pull(root, "nightly-digest", false)
	if err != nil {
		t.Fatal(err)
	}
	if !reports[0].ExactChanged {
		t.Fatalf("a moved exact contract is the one update a reader must not miss: %+v", reports)
	}
}

func TestPullIsReadOnlyUntilApply(t *testing.T) {
	root := adoptWorld(t)
	source := filepath.Join(t.TempDir(), "handoff.md")
	if err := os.WriteFile(source, []byte(exportedFixture("1.0.0", "counted")), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Adopt(root, []byte(exportedFixture("1.0.0", "counted")), source); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte(exportedFixture("1.1.0", "counted")), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Pull(root, "", false); err != nil {
		t.Fatal(err)
	}
	adopted, err := AdoptedList(root)
	if err != nil {
		t.Fatal(err)
	}
	if adopted["nightly-digest"].Version != "1.0.0" {
		t.Fatal("pull without --apply changed the record — discovering staleness must never BE the rebuild")
	}
	if _, err := Pull(root, "", true); err != nil {
		t.Fatal(err)
	}
	adopted, _ = AdoptedList(root)
	if adopted["nightly-digest"].Version != "1.1.0" {
		t.Fatal("--apply is the explicit second step, and it must actually take it")
	}
	stored, _ := os.ReadFile(filepath.Join(root, adoptedHome, "nightly-digest.md"))
	if !strings.Contains(string(stored), "version: 1.1.0") {
		t.Fatal("--apply must refresh the stored copy too, or the record lies about the bytes")
	}
}

func TestAnUnreachableSourceIsAReportedSkipNeverACrash(t *testing.T) {
	root := adoptWorld(t)
	if _, _, err := Adopt(root, []byte(exportedFixture("1.0.0", "counted")), "/no/such/path/anywhere.md"); err != nil {
		t.Fatal(err)
	}
	reports, err := Pull(root, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if reports[0].Status != "skipped" || !strings.Contains(reports[0].Note, "unreachable") {
		t.Fatalf("an unreachable source is a normal Tuesday, reported and skipped: %+v", reports)
	}
}

func TestARegressedSourceIsRefusedEvenWithApply(t *testing.T) {
	root := adoptWorld(t)
	source := filepath.Join(t.TempDir(), "handoff.md")
	if err := os.WriteFile(source, []byte(exportedFixture("2.0.0", "counted")), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Adopt(root, []byte(exportedFixture("2.0.0", "counted")), source); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte(exportedFixture("1.0.0", "counted")), 0o644); err != nil {
		t.Fatal(err)
	}
	reports, err := Pull(root, "", true)
	if err != nil {
		t.Fatal(err)
	}
	if reports[0].Status != "skipped" || !strings.Contains(reports[0].Note, "OLDER") {
		t.Fatalf("a source that moved BACKWARD is reported, never applied: %+v", reports)
	}
	adopted, _ := AdoptedList(root)
	if adopted["nightly-digest"].Version != "2.0.0" {
		t.Fatal("--apply on a regression must change nothing")
	}
}

func TestPullSpeaksHTTP(t *testing.T) {
	root := adoptWorld(t)
	body := exportedFixture("1.1.0", "counted")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	if _, _, err := Adopt(root, []byte(exportedFixture("1.0.0", "counted")), srv.URL+"/handoff.md"); err != nil {
		t.Fatal(err)
	}
	reports, err := Pull(root, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if reports[0].Status != "update" || reports[0].Latest != "1.1.0" {
		t.Fatalf("an http source is a first-class source: %+v", reports)
	}
}
