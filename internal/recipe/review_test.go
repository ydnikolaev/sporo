package recipe

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// The review's teeth. The pack must survive a copy-paste into an agent with no filesystem
// (self-contained is the whole cross-provider design), and the verify step must refuse every
// malformed verdict BY NAME — a reviewer who gets "invalid" back will not review twice.

func reviewFixture(t *testing.T) (root string, cfg Config, corpus fstest.MapFS) {
	t.Helper()
	root, cfg = sealFixture(t)
	corpus = fstest.MapFS{
		"recipes/_adoption.md": &fstest.MapFile{Data: []byte("<!-- SSOT SOURCE -->\nhouse note\n## Adopt it here\nprobe first\n## Report back\nsend scars\n")},
	}
	return root, cfg, corpus
}

func validVerdict(t *testing.T, slug, version string) []byte {
	t.Helper()
	axes := map[string]AxisScore{}
	for _, a := range reviewAxes {
		axes[a.key] = AxisScore{Score: 8, Note: "evidence: the section exists and is specific"}
	}
	b, err := json.Marshal(Verdict{Schema: 1, Recipe: slug, Version: version, Axes: axes, Call: "adopt", TopGaps: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestThePackIsSelfContained(t *testing.T) {
	root, cfg, corpus := reviewFixture(t)
	promptPath, err := BuildReviewPack(corpus, root, cfg, "baseline")
	if err != nil {
		t.Fatal(err)
	}
	prompt, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatal(err)
	}
	s := string(prompt)
	for _, want := range []string{
		recipeBegin, recipeEnd, // the recipe is inside, delimited
		"## The problem",   // ...and it is the actual recipe text
		"Adopt it here",    // the EXPORTED form — protocol appended, what a reader would score
		"minLength",        // the schema is inline
		"version `1.0.0`",  // the verdict is told which version it reviews
	} {
		if !strings.Contains(s, want) {
			t.Errorf("the pack must contain %q — an agent with no filesystem gets ONLY this file", want)
		}
	}
	for _, a := range reviewAxes {
		if !strings.Contains(s, a.key) {
			t.Errorf("axis %q missing from the rubric", a.key)
		}
	}
	if strings.Contains(s, "<!-- SSOT SOURCE") {
		t.Error("the banner is house business and must not reach a reviewer")
	}
	if _, err := os.Stat(filepath.Join(root, ".sporo", "review", "baseline", "verdict.schema.json")); err != nil {
		t.Error("the schema must also land beside the pack for tooling that wants it bare")
	}
}

func TestAPackForAMissingRecipeErrors(t *testing.T) {
	root, cfg, corpus := reviewFixture(t)
	if _, err := BuildReviewPack(corpus, root, cfg, "no-such"); err == nil {
		t.Fatal("a pack for a recipe that does not exist reviews nothing")
	}
}

func TestAValidVerdictLandsInTheRegistry(t *testing.T) {
	root, cfg, _ := reviewFixture(t)
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "verdict.json")
	if err := os.WriteFile(file, validVerdict(t, "baseline", "1.0.0"), 0o644); err != nil {
		t.Fatal(err)
	}
	sum, findings, err := VerifyVerdicts(root, cfg, "baseline", []string{file})
	if err != nil || len(findings) != 0 {
		t.Fatalf("a valid verdict must verify clean; findings=%v err=%v", findings, err)
	}
	if sum.Mean != 8.0 || sum.Verdict != "adopt" || sum.Verdicts != 1 || sum.Version != "1.0.0" {
		t.Fatalf("the summary must carry the tally: %+v", sum)
	}
	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.Reviews["baseline"]) != 1 {
		t.Fatalf("the tally must be recorded beside the seal: %+v", reg.Reviews)
	}
}

func TestAnUnsealedRecipeCannotBeVerified(t *testing.T) {
	root, cfg, _ := reviewFixture(t)
	file := filepath.Join(t.TempDir(), "verdict.json")
	if err := os.WriteFile(file, validVerdict(t, "baseline", "1.0.0"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := VerifyVerdicts(root, cfg, "baseline", []string{file}); err == nil || !strings.Contains(err.Error(), "seal") {
		t.Fatalf("a review binds a version and a content hash — an unsealed draft has neither; got: %v", err)
	}
}

func TestOneReviserOutweighsTwoAdopters(t *testing.T) {
	root, cfg, _ := reviewFixture(t)
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	files := make([]string, 3)
	for i := range files {
		v := validVerdict(t, "baseline", "1.0.0")
		if i == 2 {
			v = []byte(strings.Replace(string(v), `"verdict":"adopt"`, `"verdict":"revise"`, 1))
		}
		files[i] = filepath.Join(dir, fmt.Sprintf("v%d.json", i))
		if err := os.WriteFile(files[i], v, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	sum, findings, err := VerifyVerdicts(root, cfg, "baseline", files)
	if err != nil || len(findings) != 0 {
		t.Fatalf("three valid verdicts must verify; findings=%v err=%v", findings, err)
	}
	if sum.Verdict != "revise" || sum.Verdicts != 3 {
		t.Fatalf("a named blocking gap must not average away: %+v", sum)
	}
}

func TestEveryMalformationIsRefusedByName(t *testing.T) {
	root, cfg, _ := reviewFixture(t)
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	base := string(validVerdict(t, "baseline", "1.0.0"))
	cases := []struct{ name, verdict, want string }{
		{"wrong recipe", strings.Replace(base, `"recipe":"baseline"`, `"recipe":"other"`, 1), "wrong document"},
		{"version mismatch", strings.Replace(base, `"version":"1.0.0"`, `"version":"0.9.0"`, 1), "bind to the text it read"},
		{"score out of range", strings.Replace(base, `"score":8`, `"score":11`, 1), "0–10"},
		{"empty note", strings.Replace(base, `"note":"evidence: the section exists and is specific"`, `"note":""`, 1), "no note"},
		{"missing axis", strings.Replace(base, `"intent_clarity"`, `"renamed_axis"`, 1), "missing"},
		{"unknown axis", strings.Replace(base, `"intent_clarity"`, `"renamed_axis"`, 1), "unknown axis"},
		{"hedged verdict", strings.Replace(base, `"verdict":"adopt"`, `"verdict":"maybe"`, 1), "hedge"},
		{"wrong schema", strings.Replace(base, `"schema":1`, `"schema":2`, 1), "want 1"},
	}
	for _, c := range cases {
		file := filepath.Join(t.TempDir(), "v.json")
		if err := os.WriteFile(file, []byte(c.verdict), 0o644); err != nil {
			t.Fatal(err)
		}
		_, findings, err := VerifyVerdicts(root, cfg, "baseline", []string{file})
		if err != nil {
			t.Fatalf("%s: a malformed verdict is findings, not an error: %v", c.name, err)
		}
		found := false
		for _, f := range findings {
			if strings.Contains(f.Msg, c.want) {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: expected a finding mentioning %q; got %v", c.name, c.want, findings)
		}
	}
}

func TestFindingsBlockTheTallyFromTheRegistry(t *testing.T) {
	root, cfg, _ := reviewFixture(t)
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "v.json")
	bad := strings.Replace(string(validVerdict(t, "baseline", "1.0.0")), `"score":8`, `"score":11`, 1)
	if err := os.WriteFile(file, []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, findings, err := VerifyVerdicts(root, cfg, "baseline", []string{file}); err != nil || len(findings) == 0 {
		t.Fatalf("expected findings, no error; got findings=%v err=%v", findings, err)
	}
	reg, err := LoadRegistry(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.Reviews["baseline"]) != 0 {
		t.Fatal("a tally with a malformed verdict inside must never reach the registry")
	}
}
