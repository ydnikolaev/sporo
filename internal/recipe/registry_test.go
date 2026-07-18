package recipe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The seal's teeth. The one behavior that matters most is the REFUSAL: changed content under
// an unchanged version is the silent mutation the registry exists to catch, and if the seal
// ever lets it through, every downstream integrity claim (report-backs bind to versions, a
// marketplace recipe never mutates) is marketing.

func sealFixture(t *testing.T) (root string, cfg Config) {
	t.Helper()
	root = t.TempDir()
	cfg = Config{Home: ".sporo/recipes/"}
	home := filepath.Join(root, cfg.Home)
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "baseline.md"), []byte(conformant), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, cfg
}

func TestSealRecordsVersionHashAndLocalProvenance(t *testing.T) {
	root, cfg := sealFixture(t)
	entry, err := Seal(root, cfg, "baseline")
	if err != nil {
		t.Fatal(err)
	}
	if entry.Version != "1.0.0" || entry.Provenance != "local" || !strings.HasPrefix(entry.Hash, "sha256:") {
		t.Fatalf("a first seal records the frontmatter version, a content hash, and local provenance; got %+v", entry)
	}
	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatal(err)
	}
	if reg.Recipes["baseline"] != entry {
		t.Fatalf("the seal must be persisted, not returned and forgotten; registry has %+v", reg.Recipes["baseline"])
	}
}

func TestSealRecordsTheID(t *testing.T) {
	root, cfg := sealFixture(t)
	entry, err := Seal(root, cfg, "baseline")
	if err != nil {
		t.Fatal(err)
	}
	if entry.ID != "01ARZ3NDEKTSV4RRFFQ69G5FAV" {
		t.Fatalf("the seal must record the frontmatter id so the ledger witnesses the identity, not just the file; got %q", entry.ID)
	}
}

func TestSealRefusesAChangedID(t *testing.T) {
	root, cfg := sealFixture(t)
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	// The id is permanent. A bumped version does NOT license rewriting it — a new id under an
	// old slug is a different recipe stealing a permalink.
	rewritten := strings.Replace(conformant, "id: 01ARZ3NDEKTSV4RRFFQ69G5FAV", "id: 01BX5ZZKBKACTAV9WEVGEMMVRZ", 1)
	rewritten = strings.Replace(rewritten, "version: 1.0.0", "version: 2.0.0", 1)
	if err := os.WriteFile(filepath.Join(root, cfg.Home, "baseline.md"), []byte(rewritten), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Seal(root, cfg, "baseline"); err == nil || !strings.Contains(err.Error(), "permanent identity") {
		t.Fatalf("a changed id must be refused even under a version bump; got: %v", err)
	}
}

func TestSealBackfillsAnIDOntoAPreIDSeal(t *testing.T) {
	root, cfg := sealFixture(t)
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	// Simulate a registry sealed before ids existed: strip the id from the stored entry.
	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatal(err)
	}
	e := reg.Recipes["baseline"]
	e.ID = ""
	reg.Recipes["baseline"] = e
	if err := reg.Save(root); err != nil {
		t.Fatal(err)
	}
	// Re-sealing the same bytes must backfill the id WITHOUT demanding a version bump — the
	// ledger is catching up to a field the file already carries, not recording a mutation.
	entry, err := Seal(root, cfg, "baseline")
	if err != nil {
		t.Fatalf("backfilling an id onto an old seal must not be treated as a silent mutation: %v", err)
	}
	if entry.ID != "01ARZ3NDEKTSV4RRFFQ69G5FAV" {
		t.Fatalf("the id must be backfilled from the frontmatter; got %q", entry.ID)
	}
}

func TestResealingUnchangedContentIsIdempotent(t *testing.T) {
	root, cfg := sealFixture(t)
	first, err := Seal(root, cfg, "baseline")
	if err != nil {
		t.Fatal(err)
	}
	second, err := Seal(root, cfg, "baseline")
	if err != nil {
		t.Fatalf("re-sealing the same content must be a no-op, or no script can ever call seal safely: %v", err)
	}
	if first != second {
		t.Fatalf("idempotent means the SAME entry: %+v vs %+v", first, second)
	}
}

func TestSealRefusesChangedContentUnderAnUnchangedVersion(t *testing.T) {
	root, cfg := sealFixture(t)
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(conformant, "Derive, never restate.", "Derive, never restate. And a new claim.", 1)
	if err := os.WriteFile(filepath.Join(root, cfg.Home, "baseline.md"), []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Seal(root, cfg, "baseline"); err == nil || !strings.Contains(err.Error(), "bump") {
		t.Fatalf("changed content under the same version is the silent mutation the seal exists to catch; got: %v", err)
	}
}

func TestSealAcceptsChangedContentUnderABumpedVersion(t *testing.T) {
	root, cfg := sealFixture(t)
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	bumped := strings.Replace(conformant, "version: 1.0.0", "version: 1.1.0", 1)
	bumped = strings.Replace(bumped, "Derive, never restate.", "Derive, never restate. A reader's scar landed here.", 1)
	if err := os.WriteFile(filepath.Join(root, cfg.Home, "baseline.md"), []byte(bumped), 0o644); err != nil {
		t.Fatal(err)
	}
	entry, err := Seal(root, cfg, "baseline")
	if err != nil {
		t.Fatal(err)
	}
	if entry.Version != "1.1.0" {
		t.Fatalf("a bumped version seals the new pair; got %+v", entry)
	}
}

func TestResealPreservesProvenance(t *testing.T) {
	root, cfg := sealFixture(t)
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatal(err)
	}
	e := reg.Recipes["baseline"]
	e.Provenance = "team:acme"
	reg.Recipes["baseline"] = e
	if err := reg.Save(root); err != nil {
		t.Fatal(err)
	}
	bumped := strings.Replace(conformant, "version: 1.0.0", "version: 2.0.0", 1)
	if err := os.WriteFile(filepath.Join(root, cfg.Home, "baseline.md"), []byte(bumped), 0o644); err != nil {
		t.Fatal(err)
	}
	entry, err := Seal(root, cfg, "baseline")
	if err != nil {
		t.Fatal(err)
	}
	if entry.Provenance != "team:acme" {
		t.Fatalf("sealing a new version of a team recipe must not quietly claim it as local; got %+v", entry)
	}
}

// The fleet rule: an exact-bound contract is somebody else's parser, and changing it under
// a minor bump ships a break wearing a compatible version number.

func exactFixture() string {
	return strings.Replace(conformant, "**Binding: adapt** (rename the fields into your own language)",
		"**Binding: exact** (the fleet's collector parses this shape)", 1)
}

func TestChangingAnExactContractUnderAMinorBumpIsRefused(t *testing.T) {
	root, cfg := sealFixture(t)
	if err := os.WriteFile(filepath.Join(root, cfg.Home, "baseline.md"), []byte(exactFixture()), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(exactFixture(), `"counted": 12`, `"tallied": 12`, 1)
	changed = strings.Replace(changed, "version: 1.0.0", "version: 1.1.0", 1)
	if err := os.WriteFile(filepath.Join(root, cfg.Home, "baseline.md"), []byte(changed), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Seal(root, cfg, "baseline"); err == nil || !strings.Contains(err.Error(), "MAJOR") {
		t.Fatalf("a renamed field in an exact shape under a minor bump must be refused as a fleet break; got: %v", err)
	}
	major := strings.Replace(changed, "version: 1.1.0", "version: 2.0.0", 1)
	if err := os.WriteFile(filepath.Join(root, cfg.Home, "baseline.md"), []byte(major), 0o644); err != nil {
		t.Fatal(err)
	}
	entry, err := Seal(root, cfg, "baseline")
	if err != nil {
		t.Fatalf("the same change under a major bump is the rule being followed: %v", err)
	}
	if entry.Version != "2.0.0" {
		t.Fatalf("expected the major seal, got %+v", entry)
	}
}

func TestAProseEditAroundAnExactContractStaysMinor(t *testing.T) {
	root, cfg := sealFixture(t)
	if err := os.WriteFile(filepath.Join(root, cfg.Home, "baseline.md"), []byte(exactFixture()), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	reworded := strings.Replace(exactFixture(), "Derive, never restate.", "Derive, never restate — a reader's scar landed here.", 1)
	reworded = strings.Replace(reworded, "version: 1.0.0", "version: 1.1.0", 1)
	if err := os.WriteFile(filepath.Join(root, cfg.Home, "baseline.md"), []byte(reworded), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatalf("only the fence CONTENTS are the promise — prose rewording is a minor change: %v", err)
	}
}

func TestVerifyIsGreenAfterASeal(t *testing.T) {
	root, cfg := sealFixture(t)
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	f, err := VerifyRegistry(root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(f) != 0 {
		t.Fatalf("a freshly sealed corpus must verify green, or the gate cries wolf on the state the tool itself wrote: %v", f)
	}
}

func TestVerifyFindsContentDriftedFromItsSeal(t *testing.T) {
	root, cfg := sealFixture(t)
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(conformant, "Derive, never restate.", "Quietly different.", 1)
	if err := os.WriteFile(filepath.Join(root, cfg.Home, "baseline.md"), []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	assertFinding(t, root, cfg, "drifted")
}

func TestVerifyFindsAChangedID(t *testing.T) {
	root, cfg := sealFixture(t)
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	rewritten := strings.Replace(conformant, "id: 01ARZ3NDEKTSV4RRFFQ69G5FAV", "id: 01BX5ZZKBKACTAV9WEVGEMMVRZ", 1)
	if err := os.WriteFile(filepath.Join(root, cfg.Home, "baseline.md"), []byte(rewritten), 0o644); err != nil {
		t.Fatal(err)
	}
	assertFinding(t, root, cfg, "permanent identity")
}

func TestVerifyFindsAVersionBumpedButNotResealed(t *testing.T) {
	root, cfg := sealFixture(t)
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	bumped := strings.Replace(conformant, "version: 1.0.0", "version: 1.2.0", 1)
	if err := os.WriteFile(filepath.Join(root, cfg.Home, "baseline.md"), []byte(bumped), 0o644); err != nil {
		t.Fatal(err)
	}
	assertFinding(t, root, cfg, "re-seal")
}

func TestVerifyFindsASealedRecipeThatVanished(t *testing.T) {
	root, cfg := sealFixture(t)
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, cfg.Home, "baseline.md")); err != nil {
		t.Fatal(err)
	}
	assertFinding(t, root, cfg, "missing from the recipes home")
}

func TestAnUnsealedDraftHasNoObligations(t *testing.T) {
	root := t.TempDir()
	cfg := Config{Home: ".sporo/recipes/"}
	home := filepath.Join(root, cfg.Home)
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	// A draft is not sealed and owes the registry nothing — the all-sealed sweep exempts it by
	// design, because a draft has no version to promise yet.
	draft := strings.Replace(conformant, "effort: reference", "effort: reference\ndraft: true", 1)
	if err := os.WriteFile(filepath.Join(home, "baseline.md"), []byte(draft), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := VerifyRegistry(root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(f) != 0 {
		t.Fatalf("a draft is not sealed and owes the registry nothing; got: %v", f)
	}
}

func TestSealStampsSealedAtOnceAndKeepsIt(t *testing.T) {
	root, cfg := sealFixture(t)
	orig := sealNow
	t.Cleanup(func() { sealNow = orig })

	sealNow = func() string { return "2026-01-01T00:00:00Z" }
	e1, err := Seal(root, cfg, "baseline")
	if err != nil {
		t.Fatal(err)
	}
	if e1.SealedAt != "2026-01-01T00:00:00Z" {
		t.Fatalf("the first seal must stamp sealed_at; got %q", e1.SealedAt)
	}

	// An idempotent re-seal (same bytes, same version) is not a new seal event — sealed_at must
	// not move, even though the clock has.
	sealNow = func() string { return "2099-12-31T00:00:00Z" }
	e2, err := Seal(root, cfg, "baseline")
	if err != nil {
		t.Fatal(err)
	}
	if e2.SealedAt != "2026-01-01T00:00:00Z" {
		t.Fatalf("an idempotent re-seal must preserve sealed_at; got %q", e2.SealedAt)
	}
}

func TestAFinishedRecipeMustBeSealed(t *testing.T) {
	// A conformant (non-draft) recipe that was never sealed is published in intent but unwitnessed
	// by the registry — the own-home gate must flag it, so "all recipes are sealed" stays true.
	root, cfg := sealFixture(t)
	f, err := VerifyRegistry(root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(f) != 1 || !strings.Contains(f[0].String(), "not sealed") {
		t.Fatalf("a finished, unsealed recipe must be flagged by the own-home gate; got: %v", f)
	}
}

func TestAMalformedRegistryIsAHardError(t *testing.T) {
	root, _ := sealFixture(t)
	if err := os.MkdirAll(filepath.Join(root, ".sporo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".sporo", "registry.yaml"), []byte("recipes: [not: a: map"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRegistry(root); err == nil {
		t.Fatal("a malformed registry must be a hard error — degrading to an empty one silently unseals every recipe in the project")
	}
}

func assertFinding(t *testing.T, root string, cfg Config, want string) {
	t.Helper()
	f, err := VerifyRegistry(root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range f {
		if strings.Contains(x.Msg, want) {
			return
		}
	}
	t.Fatalf("expected a finding mentioning %q; got: %v", want, f)
}
