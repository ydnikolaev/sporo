package recipekit

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// seedGenreShapes is the compatibility ledger for the shape enforced by SeedShape. A section can
// be added, removed or reordered only by introducing a new spec version key; changing the snapshot
// under an existing key would erase what that released version meant. It is the seed-local sibling
// of the recipe genre's genreShapes — a parallel ledger, because the recipe snapshot tests are
// off-limits (INV-1) and a shared helper would have to reach into them.
//
// The key lists are INLINED, not a live `strings.Join(seedKeys, …)`: a pin built from the same
// variable it guards moves with a key-only edit and never reds without a version bump (BL-002).
// Both rows carry the identical literal because 1.0.1 is a byte-equal PATCH of 1.0.0; a key change
// under either version is exactly what these rows must catch.
var seedGenreShapes = map[string]string{
	"1.0.0": strings.Join([]string{
		"## Summary",
		"## What it is",
		"## Install",
		"## Verify",
		"## Use",
		"## Harness",
		"## Report",
		"-- frontmatter --",
		"id\nname\nversion\ntitle\ntarget\nsource\nstack\nverified\neffort",
	}, "\n"),
	// 1.0.1 is a PATCH clarification of §3's neutrality prose — the seven-section shape and nine
	// keys are unchanged, so the snapshot is byte-identical to 1.0.0 (no MAJOR required).
	"1.0.1": strings.Join([]string{
		"## Summary",
		"## What it is",
		"## Install",
		"## Verify",
		"## Use",
		"## Harness",
		"## Report",
		"-- frontmatter --",
		"id\nname\nversion\ntitle\ntarget\nsource\nstack\nverified\neffort",
	}, "\n"),
}

func seedGenreShapeMatches(version string, sections, keys []string) bool {
	want, ok := seedGenreShapes[version]
	got := strings.Join(append(append([]string(nil), sections...), append([]string{"-- frontmatter --"}, strings.Join(keys, "\n"))...), "\n")
	return ok && got == want
}

func TestSeedRequiredShapeIsBoundToTheGenreVersion(t *testing.T) {
	src, err := os.ReadFile("../../seeds/_authoring.md")
	if err != nil {
		t.Fatal(err)
	}
	version := FrontmatterValue(src, "version")
	if !seedGenreShapeMatches(version, seedSections, seedKeys) {
		t.Fatalf("seedSections changed without a genre-version bump: add the new version and exact shape to seedGenreShapes, then record the compatibility change in the spec changelog (current version %q)", version)
	}
}

// The ADD-direction teeth: today's real shape being green proves nothing if the gate ignores the
// next required section. Seed one extra member and require the same version to go red.
func TestSeedGenreShapeVersionGateHasAddDirectionTeeth(t *testing.T) {
	mutated := append(append([]string(nil), seedSections...), "## A newly required section")
	if seedGenreShapeMatches("1.0.0", mutated, seedKeys) {
		t.Fatal("adding a required section under the same genre version must red the gate")
	}
	mutatedKeys := append(append([]string(nil), seedKeys...), "new_required_key")
	if seedGenreShapeMatches("1.0.0", seedSections, mutatedKeys) {
		t.Fatal("adding a required frontmatter key under the same genre version must red the gate")
	}
}

func TestAnIncompatibleSeedGenreShapeChangeRequiresANewMajor(t *testing.T) {
	versions := make([]string, 0, len(seedGenreShapes))
	for version := range seedGenreShapes {
		versions = append(versions, version)
	}
	sort.Slice(versions, func(i, j int) bool { return SemverNewer(versions[j], versions[i]) })
	for i := 1; i < len(versions); i++ {
		prev, next := versions[i-1], versions[i]
		if seedGenreShapes[prev] != seedGenreShapes[next] && SemverMajor(next) <= SemverMajor(prev) {
			t.Fatalf("seed genre shape changed from %s to %s without a new MAJOR", prev, next)
		}
	}
}

var reSeedSpecVersion = regexp.MustCompile(`(?m)^version:\s*.+$`)

// seedSpecHashes is the released-byte ledger for the seed genre's constitutional text. The version
// line is normalized to avoid a self-referential hash; every other byte must move under a new key.
// The normalize-and-hash logic mirrors the recipe spec-hash ledger's specPayloadHash rather than
// reusing it: that helper is an unexported symbol inside spec_version_test.go, which INV-1 forbids
// touching — a self-contained seed ledger must not couple to a private of the file it must not edit.
var seedSpecHashes = map[string]string{
	"1.0.0": "sha256:7538f88e3416fe22c8f7766fcc7abd3e962493e9059d8eaae664e923e07c4884",
	"1.0.1": "sha256:9b3a64565e31f8fc2e32963ca78cdd0861defdd7f6afeee354278d46f35b4b4e",
}

func seedSpecPayloadHash(src []byte) string {
	stripped := StripBanner(string(src))
	normalized := reSeedSpecVersion.ReplaceAllString(stripped, "version: <normalized>")
	return ContentHash([]byte(normalized))
}

func seedSpecHashMatches(version string, src []byte) bool {
	want, ok := seedSpecHashes[version]
	return ok && seedSpecPayloadHash(src) == want
}

func TestReleasedSeedSpecBytesAreBoundToTheirVersion(t *testing.T) {
	src := mustReadSeedSpec(t)
	version := FrontmatterValue(src, "version")
	if !seedSpecHashMatches(version, src) {
		t.Fatalf("seeds/_authoring.md changed without a version bump: add the new version and payload hash to seedSpecHashes (version %q, hash %q)", version, seedSpecPayloadHash(src))
	}
}

func TestSeedSpecVersionGateHasMutationTeeth(t *testing.T) {
	src := mustReadSeedSpec(t)
	version := FrontmatterValue(src, "version")
	mutated := append(append([]byte(nil), src...), []byte("\nA silently changed released rule.\n")...)
	if seedSpecHashMatches(version, mutated) {
		t.Fatalf("seeds/_authoring.md content mutation under version %s must red the gate", version)
	}
}

func mustReadSeedSpec(t *testing.T) []byte {
	t.Helper()
	src, err := os.ReadFile("../../seeds/_authoring.md")
	if err != nil {
		t.Fatal(err)
	}
	return src
}

// runnerSpecHashes is the released-byte ledger for the seed runner preamble — the short
// constitutional meta-doc (seeds/_runner.md) that seed export prepends to every seal. It is a
// sibling of seedSpecHashes, not an entry in it: the two docs version independently (the genre
// spec and the execution preamble are different constitutions), so a bump to one must not force a
// bump to the other. The normalize-and-hash path is shared — seedSpecPayloadHash is a pure
// function of bytes, and the version line is normalized the same way — so no second helper is
// warranted; only the ledger and its reader are runner-specific.
var runnerSpecHashes = map[string]string{
	"1.0.0": "sha256:0015f0c664367ec75c3f4a15609de85aca8ec9f0efbd69b6fc63adc0e730cd24",
}

func runnerSpecHashMatches(version string, src []byte) bool {
	want, ok := runnerSpecHashes[version]
	return ok && seedSpecPayloadHash(src) == want
}

func TestReleasedRunnerPreambleBytesAreBoundToTheirVersion(t *testing.T) {
	src := mustReadRunnerPreamble(t)
	version := FrontmatterValue(src, "version")
	if !runnerSpecHashMatches(version, src) {
		t.Fatalf("seeds/_runner.md changed without a version bump: add the new version and payload hash to runnerSpecHashes (version %q, hash %q)", version, runnerSpecPayloadHash(src))
	}
}

func TestRunnerPreambleVersionGateHasMutationTeeth(t *testing.T) {
	src := mustReadRunnerPreamble(t)
	version := FrontmatterValue(src, "version")
	mutated := append(append([]byte(nil), src...), []byte("\nA silently changed runner rule.\n")...)
	if runnerSpecHashMatches(version, mutated) {
		t.Fatalf("seeds/_runner.md content mutation under version %s must red the gate", version)
	}
}

func runnerSpecPayloadHash(src []byte) string { return seedSpecPayloadHash(src) }

func mustReadRunnerPreamble(t *testing.T) []byte {
	t.Helper()
	src, err := os.ReadFile("../../seeds/_runner.md")
	if err != nil {
		t.Fatal(err)
	}
	return src
}
