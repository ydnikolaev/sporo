package recipe

import (
	"os"
	"strings"
	"testing"
)

// Two suites in one file, deliberately adjacent: the conform teeth (the layer working) and
// the ADR-005 invariance teeth (the layer KNOWING ITS PLACE — an adapt-only recipe never
// meets any of it). The second suite is what keeps the genre universal while the first
// grows: if a future change makes strictness mandatory, it reds here before the genre
// ossifies around one class of recipe.

// exactWithFixtures swaps the baseline's adapt shape for an exact one carrying one valid
// and one invalid fixture.
func exactWithFixtures(valid, invalid string) string {
	contracts := "The feed the fleet's aggregator parses — **Binding: exact**:\n\n" +
		"```json\n" +
		`{ "schema": 1, "counted": 12, "source": "<a role name>" }` + "\n" +
		"```\n\n" +
		"**Fixture: valid** — a real record from the verifying build:\n\n" +
		"```json\n" + valid + "\n```\n\n" +
		"**Fixture: invalid** — the field rename that broke the first consumer:\n\n" +
		"```json\n" + invalid + "\n```\n"
	from := "The shape the collector emits — **Binding: adapt** (rename the fields into your own language):\n\n" +
		"```json" + "\n" +
		`{ "schema": 1, "counted": 12, "absent": { "reachable": false, "reason": "no such source here" } }` + "\n" +
		"```"
	return strings.Replace(conformant, from, contracts, 1)
}

const goodFeed = `{ "schema": 2, "counted": 40, "source": "the nightly collector" }`
const renamedFeed = `{ "schema": 2, "tallied": 40, "source": "the nightly collector" }`

func TestFixturesAreRunAgainstTheirContract(t *testing.T) {
	body := exactWithFixtures(goodFeed, renamedFeed)
	if f := lintFixture(t, body); len(f) != 0 {
		t.Fatalf("a valid fixture that conforms and an invalid one that violates is the layer working; got: %v", f)
	}
}

func TestAValidFixtureThatFailsItsContractReds(t *testing.T) {
	body := exactWithFixtures(renamedFeed, renamedFeed)
	assertRed(t, lintFixture(t, body), "stamped VALID")
}

func TestAnInvalidFixtureThatConformsReds(t *testing.T) {
	body := exactWithFixtures(goodFeed, goodFeed)
	assertRed(t, lintFixture(t, body), "stamped INVALID")
}

func TestAnExactShapeThatDoesNotParseReds(t *testing.T) {
	body := strings.Replace(conformant, "**Binding: adapt** (rename the fields into your own language)",
		"**Binding: exact**", 1)
	body = strings.Replace(body, `{ "schema": 1, "counted": 12, "absent": { "reachable": false, "reason": "no such source here" } }`,
		"a shape in prose, which no machine can hold anyone to", 1)
	assertRed(t, lintFixture(t, body), "does not parse")
}

func TestConformNamesEveryViolationWithItsPath(t *testing.T) {
	body := exactWithFixtures(goodFeed, renamedFeed)
	contracts := ExactContracts([]byte(body))
	if len(contracts) != 1 || len(contracts[0].Fixtures) != 2 {
		t.Fatalf("expected one exact contract with two fixtures, got %+v", contracts)
	}
	violations, err := Conform(contracts[0], []byte(`{ "schema": "two", "tallied": 40, "source": 7, "extra": true }`))
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(violations, "\n")
	for _, want := range []string{"$.schema", "$.counted: missing", "$.tallied: not in the contract", "$.extra: not in the contract"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected a violation mentioning %q; got:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, "$.source") {
		t.Errorf("a placeholder admits any scalar — 7 satisfies it; got:\n%s", joined)
	}
}

func TestConformSpeaksYAMLToo(t *testing.T) {
	body := exactWithFixtures(goodFeed, renamedFeed)
	c := ExactContracts([]byte(body))[0]
	violations, err := Conform(c, []byte("schema: 3\ncounted: 9\nsource: the collector\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Fatalf("a YAML candidate with the same shape conforms; got: %v", violations)
	}
}

// --- The ADR-005 invariance gate -------------------------------------------------------

func TestAnAdaptOnlyRecipeNeverMeetsTheConformLayer(t *testing.T) {
	if got := ExactContracts([]byte(conformant)); len(got) != 0 {
		t.Fatalf("the baseline declares nothing exact and must own no layer-2 machinery; got %+v", got)
	}
	if f := fixtureFindings("baseline.md", []byte(conformant)); len(f) != 0 {
		t.Fatalf("an adapt-only recipe pays nothing for the conform layer (ADR-005); got: %v", f)
	}
	if d := exactContractsDigest([]byte(conformant)); d != "" {
		t.Fatalf("no exact contracts → empty digest → the seal's major rule stays asleep; got %q", d)
	}
}

func TestTheScaffoldActivatesNoStrictness(t *testing.T) {
	root, cfg := scaffoldWorld(t)
	path, err := Scaffold(root, cfg, "flat-by-default", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(ExactContracts(src)) != 0 || exactContractsDigest(src) != "" {
		t.Fatal("the scaffold's default output must not activate layer 2 — strictness is the author's declaration, never the tool's (ADR-005)")
	}
}
