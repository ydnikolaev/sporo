package recipe

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// rec builds one git record in the grammar the harvest reads.
func rec(hash, subject, body string) string {
	return "\x01" + hash + "\x1f" + subject + "\x1f" + body + "\x02"
}

// The signal table is a HEURISTIC that proposes; the danger is that it silently proposes
// nothing. A defect whose own message admits it was found only by sabotage is the single
// most valuable thing a recipe can carry, and it must never fall through to the unsignaled
// pile just because its subject line was typed as a `feat`.
func TestASabotageFoundDefectIsProposedEvenWhenItIsNotAFix(t *testing.T) {
	raw := rec("aaa1111", "feat(report): the runtime telemetry surface",
		"The exit code was swallowed by a bare if. Found only by sabotage: removing the venv\n"+
			"made the collector report every epic unreadable and exit 0.")
	h, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(h.Scars) != 1 {
		t.Fatalf("a sabotage-found defect must be proposed as a scar candidate; got %d candidates, %d design",
			len(h.Scars), len(h.Design))
	}
	if !strings.Contains(strings.Join(h.Scars[0].Signals, ","), "sabotage") {
		t.Fatalf("the candidate must say WHY it was proposed; signals=%v", h.Scars[0].Signals)
	}
}

// The complement, and the reason the count is reported: a commit the signals do NOT match
// is not silently dropped. A scar nobody wrote down is still a scar, and the author is told
// how many commits the machine had nothing to say about — an unsignaled pile of zero would
// be a lie of omission dressed as thoroughness.
func TestUnsignaledCommitsAreCountedNotSwallowed(t *testing.T) {
	raw := rec("bbb2222", "chore(deps): bump the yaml parser", "") +
		rec("ccc3333", "docs(status): close the row", "") +
		rec("ddd4444", "fix(sync): the stamp broke the shebang", "")
	h, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if h.Unsignaled != 2 {
		t.Fatalf("want 2 unsignaled commits reported, got %d", h.Unsignaled)
	}
	if len(h.Scars) != 1 || h.Scars[0].Hash != "ddd4444" {
		t.Fatalf("the fix must be proposed; got %+v", h.Scars)
	}
}

// A body with blank lines and a subject that ignores the convention are the shapes a real
// history is full of. The parser holds, or the harvest dies on the first honest repository.
func TestTheParserSurvivesAnUnconventionalHistory(t *testing.T) {
	raw := rec("eee5555", "Merge branch fixes", "\n\nno convention here\n\n") +
		rec("fff6666", "feat(cli): a verb", "body\n\nwith a blank line\n")
	h, err := Parse(raw)
	if err != nil {
		t.Fatalf("the parser must survive a history that ignores the convention: %v", err)
	}
	if len(h.Design) != 1 || h.Unsignaled != 1 {
		t.Fatalf("want 1 design candidate and 1 unsignaled; got %d design, %d unsignaled", len(h.Design), h.Unsignaled)
	}
}

// A malformed record is an error, not a silent skip: a harvest that quietly drops what it
// cannot read hands the author a recipe with a hole in it and no way to know.
func TestAMalformedRecordIsAnErrorNotASilentSkip(t *testing.T) {
	if _, err := Parse("\x01no-terminator-here"); err == nil {
		t.Fatal("a record with no terminator must be an error")
	}
	if _, err := Parse(rec("aaa", "subject", "body")[:12] + "\x02"); err == nil {
		t.Fatal("a header with missing fields must be an error")
	}
}

// Export's delivery contract: one file, no provenance banner. The banner is a message to
// the fleet about a repository the reader does not have.
func TestExportStripsTheBannerAndRefusesTheGenreSpec(t *testing.T) {
	corpus := fstest.MapFS{
		"recipes/_authoring.md": {Data: []byte("<!-- SSOT SOURCE (mate repo). -->\n\n# how to write one\n")},
		"recipes/_adoption.md":  {Data: []byte("<!-- SSOT SOURCE -->\n<!-- a note to us -->\n\n## Adopt it here\nprobe first.\n")},
		"recipes/a-thing.md":    {Data: []byte("<!-- SSOT SOURCE (mate repo). -->\n\n---\nname: a-thing\n---\n# A thing\n")},
	}
	body, err := Export(corpus, "", "a-thing")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(body, "SSOT SOURCE") {
		t.Fatalf("the exported file must not open by talking about its origin repository:\n%s", body)
	}
	if !strings.HasPrefix(body, "---\nname: a-thing") {
		t.Fatalf("the export must begin at the frontmatter:\n%.40q", body)
	}
	if _, err := Export(corpus, "", "_authoring"); err == nil {
		t.Fatal("the genre's shape spec is not a recipe and must not export as one")
	}
	got, err := List(corpus, "")
	if err != nil || len(got) != 1 || got[0].Slug != "a-thing" || got[0].Origin != Official {
		t.Fatalf("List must skip the meta-documents; got %v (%v)", got, err)
	}
}

// The project's OWN recipe is what a consumer authors, and it must be both listed and
// exportable — otherwise "write a recipe about this repo" produces a file nothing can hand
// out. It also WINS over a fleet recipe of the same slug: a build you verified here beats
// the same slug someone else shipped.
func TestTheProjectsOwnRecipesAreListedAndExportedFirst(t *testing.T) {
	corpus := fstest.MapFS{
		"recipes/a-thing.md":    {Data: []byte("<!-- SSOT SOURCE -->\n\n# the fleet's version\n")},
		"recipes/_authoring.md": {Data: []byte("<!-- SSOT SOURCE -->\n")},
		"recipes/_adoption.md":  {Data: []byte("<!-- SSOT SOURCE -->\n\n## Adopt it here\nprobe first.\n")},
	}
	home := t.TempDir()
	write(t, home, "a-thing.md", "<!-- SSOT SOURCE -->\n\n# this repo's own version\n")
	write(t, home, "local-only.md", "<!-- SSOT SOURCE -->\n\n# only here\n")
	write(t, home, "README.md", "not a recipe, but it is a .md in the home")

	got, err := List(corpus, home)
	if err != nil {
		t.Fatal(err)
	}
	origins := map[string]Origin{}
	for _, e := range got {
		origins[e.Slug] = e.Origin
	}
	if origins["local-only"] != Project {
		t.Fatalf("the project's own recipe must be listed as its own; got %v", got)
	}
	if origins["a-thing"] != Project {
		t.Fatalf("on a slug collision the local build wins and SAYS so; got %v", got)
	}
	body, err := Export(corpus, home, "a-thing")
	if err != nil || !strings.Contains(body, "this repo's own version") {
		t.Fatalf("export must prefer the project's own recipe; got %q (%v)", body, err)
	}
}

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Absence must arrive as `[]`, never as `null`. The distinction is the module's own first
// principle — a silent zero looks exactly like a real measurement — and the harvest violated
// it the first time it ran over a range whose gate registry had not changed.
func TestAnEmptyFindingIsAnEmptyListNotNull(t *testing.T) {
	h, err := Parse("")
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(h)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"gates_added", "doctrines_touched", "decisions_touched", "knowledge_touched", "absent_sources", "scar_candidates", "design_candidates"} {
		if strings.Contains(string(b), `"`+key+`":null`) {
			t.Fatalf("%s serialized as null — that reads as \"not looked at\", not \"nothing found\": %s", key, b)
		}
	}
}

// The nested `sources` object must speak the same snake_case dialect as every sibling field —
// it is one JSON contract, not two. A `Sources` field carrying only a yaml tag serializes with
// Go's PascalCase name (`Gates`), a single inconsistent island a consumer's parser trips on.
// This is the ADD-direction teeth: it reddens the moment a new Sources field lands untagged.
func TestSourcesSerializeSnakeCaseInTheJSONContract(t *testing.T) {
	b, err := json.Marshal(Harvest{})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"gates", "doctrine", "decisions", "knowledge"} {
		if !strings.Contains(string(b), `"`+key+`"`) {
			t.Fatalf("sources key %q missing from the JSON contract — a Sources field lost its json tag: %s", key, b)
		}
	}
	for _, leaked := range []string{`"Gates"`, `"Doctrine"`, `"Decisions"`, `"Knowledge"`} {
		if strings.Contains(string(b), leaked) {
			t.Fatalf("PascalCase %s leaked into the JSON contract — a Sources field is missing its json tag: %s", leaked, b)
		}
	}
}

// The teeth of the gate-registry parser. Its first version searched for a row grammar the
// registry does not use, so it returned "no gates added" on every input — a check that
// cannot fire, and one that reads exactly like the truth. The fixture is the registry's real
// shape, taken from the diff of a release that did add a gate.
func TestTheGateParserReadsTheRegistrysRealGrammar(t *testing.T) {
	diff := `--- a/harness/gates.yaml
+++ b/harness/gates.yaml
@@ -40,0 +41,4 @@ gates:
+  recipe-lint:
+    teeth: script
+    where: scripts/recipe-lint-test.sh
+    reason: "the genre is only a genre if its shape is checked"
@@ -60,0 +65 @@
+  report-lint:
`
	got := ParseGatesDiff(diff)
	if len(got) != 2 || got[0] != "recipe-lint" || got[1] != "report-lint" {
		t.Fatalf("want the two added gates, got %v", got)
	}
	if len(ParseGatesDiff("")) != 0 {
		t.Fatal("an empty diff adds no gates")
	}
	// The ADD direction the first version failed: an attribute is not a gate.
	if g := ParseGatesDiff("+    teeth: script\n"); len(g) != 0 {
		t.Fatalf("a gate's attribute must not be read as a gate: %v", g)
	}
}
