package recipekit

import (
	"os"
	"sort"
	"strings"
	"testing"
)

// genreShapes is the compatibility ledger for the shape enforced by requiredSections. A
// section can be added, removed or reordered only by introducing a new spec version key;
// changing the snapshot under an existing key would erase what that released version meant.
var genreShapes = map[string]string{
	"1.0.0": strings.Join([]string{
		"## The problem",
		"## Why the obvious approach fails",
		"## The principles",
		"## The ground it needs",
		"## The contracts",
		"## The build sequence",
		"## The seams",
		"## The scars",
		"## Verification",
		"## The trade-offs",
		"## For the human",
		"-- frontmatter --",
		strings.Join(requiredKeys, "\n"),
	}, "\n"),
	// The current version's key list is INLINED, not a live `strings.Join(requiredKeys, …)`: a pin
	// built from the same variable it guards moves with a key-only edit and never reds without a
	// version bump (BL-002). The literal must equal requiredKeys byte-for-byte until a new version
	// key is added — a key change under 2.0.0 is exactly what this row must catch.
	"2.0.0": strings.Join([]string{
		"## Summary",
		"## The problem",
		"## Why the obvious approach fails",
		"## The principles",
		"## The ground it needs",
		"## The contracts",
		"## The build sequence",
		"## The seams",
		"## The scars",
		"## Verification",
		"## The trade-offs",
		"## For the human",
		"-- frontmatter --",
		"id\nname\nversion\ntitle\nproblem\nprerequisites\nderived_from\nstack\nverified\neffort",
	}, "\n"),
}

func genreShapeMatches(version string, sections, keys []string) bool {
	want, ok := genreShapes[version]
	got := strings.Join(append(append([]string(nil), sections...), append([]string{"-- frontmatter --"}, strings.Join(keys, "\n"))...), "\n")
	return ok && got == want
}

func TestRequiredShapeIsBoundToTheGenreVersion(t *testing.T) {
	src, err := os.ReadFile("../../recipes/_authoring.md")
	if err != nil {
		t.Fatal(err)
	}
	version := FrontmatterValue(src, "version")
	if !genreShapeMatches(version, requiredSections, requiredKeys) {
		t.Fatalf("requiredSections changed without a genre-version bump: add the new version and exact shape to genreShapes, then record the compatibility change in the spec changelog (current version %q)", version)
	}
}

// The ADD-direction teeth: today's real shape being green proves nothing if the gate ignores
// the next required section. Seed one extra member and require the same version to go red.
func TestGenreShapeVersionGateHasAddDirectionTeeth(t *testing.T) {
	mutated := append(append([]string(nil), requiredSections...), "## A newly required section")
	if genreShapeMatches("2.0.0", mutated, requiredKeys) {
		t.Fatal("adding a required section under the same genre version must red the gate")
	}
	mutatedKeys := append(append([]string(nil), requiredKeys...), "new_required_key")
	if genreShapeMatches("2.0.0", requiredSections, mutatedKeys) {
		t.Fatal("adding a required frontmatter key under the same genre version must red the gate")
	}
}

func TestAnIncompatibleGenreShapeChangeRequiresANewMajor(t *testing.T) {
	versions := make([]string, 0, len(genreShapes))
	for version := range genreShapes {
		versions = append(versions, version)
	}
	sort.Slice(versions, func(i, j int) bool { return SemverNewer(versions[j], versions[i]) })
	for i := 1; i < len(versions); i++ {
		prev, next := versions[i-1], versions[i]
		if genreShapes[prev] != genreShapes[next] && SemverMajor(next) <= SemverMajor(prev) {
			t.Fatalf("genre shape changed from %s to %s without a new MAJOR", prev, next)
		}
	}
}
