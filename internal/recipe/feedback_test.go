package recipe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The return channel's teeth. The property that matters most is the refusal with a NAME: a
// reader who sent four of six sections must learn which two are missing and why they cost
// something — "invalid report" teaches nothing and the loop dies of friction.

const validReport = `Report back for the baseline recipe, version 1.0.0.

**Stack:** Node 22, a single script; kept the "one static artifact" choice, replaced the language.
**Degraded:** the decision log was absent — built the smallest one, labelled "reconstructed" in the output.
**New scars:** one. **Symptom:** double-counted rows. **Root cause:** two sources answering different questions. **Fix:** named the question each answers, stopped adding them.
**Wrong:** nothing contradicted the build.
**Arithmetic:** ran the live check; 3 of 3 matched.
**Missing:** the collector's error shape — had to invent it.
`

func feedbackFixture(t *testing.T) (string, Config) {
	t.Helper()
	root, cfg := sealFixture(t) // a home with baseline.md in it is exactly what feedback needs
	return root, cfg
}

func TestAValidReportIsFiledAndPersisted(t *testing.T) {
	root, cfg := feedbackFixture(t)
	path, warning, err := AddFeedback(root, cfg, "baseline", []byte(validReport))
	if err != nil {
		t.Fatal(err)
	}
	if warning != "" {
		t.Fatalf("a report that cites its version needs no warning; got: %s", warning)
	}
	b, err := os.ReadFile(path)
	if err != nil || string(b) != validReport {
		t.Fatalf("the filed report must be the reader's bytes, exactly: %v", err)
	}
	if !strings.HasPrefix(filepath.Base(path), "001-") {
		t.Fatalf("the first report is 001- so arrival order survives an ls: %s", path)
	}
}

func TestAReportMissingAMarkerIsRefusedByName(t *testing.T) {
	root, cfg := feedbackFixture(t)
	broken := strings.Replace(validReport, "**New scars:**", "**Findings:**", 1)
	_, _, err := AddFeedback(root, cfg, "baseline", []byte(broken))
	if err == nil || !strings.Contains(err.Error(), "New scars") {
		t.Fatalf("the refusal must name the missing section, or the reader cannot fix it; got: %v", err)
	}
}

func TestFilingTheSameReportTwiceIsIdempotent(t *testing.T) {
	root, cfg := feedbackFixture(t)
	first, _, err := AddFeedback(root, cfg, "baseline", []byte(validReport))
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := AddFeedback(root, cfg, "baseline", []byte(validReport))
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("the same bytes must land once — a duplicate is a scar merged twice: %s vs %s", first, second)
	}
	all, err := ListFeedback(root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(all["baseline"]) != 1 {
		t.Fatalf("one report on disk, not two: %v", all["baseline"])
	}
}

func TestTwoDifferentReportsBothLand(t *testing.T) {
	root, cfg := feedbackFixture(t)
	if _, _, err := AddFeedback(root, cfg, "baseline", []byte(validReport)); err != nil {
		t.Fatal(err)
	}
	other := strings.Replace(validReport, "Node 22", "Python 3.12", 1)
	if _, _, err := AddFeedback(root, cfg, "baseline", []byte(other)); err != nil {
		t.Fatal(err)
	}
	all, _ := ListFeedback(root, cfg)
	if len(all["baseline"]) != 2 {
		t.Fatalf("two distinct reports are two files: %v", all["baseline"])
	}
	if !strings.HasPrefix(filepath.Base(all["baseline"][1]), "002-") {
		t.Fatalf("the second report takes the next sequence: %v", all["baseline"])
	}
}

func TestFeedbackForARecipeThisProjectDoesNotAuthorIsRefused(t *testing.T) {
	root, cfg := feedbackFixture(t)
	_, _, err := AddFeedback(root, cfg, "somebody-elses", []byte(validReport))
	if err == nil || !strings.Contains(err.Error(), "does not author") {
		t.Fatalf("feedback files against the recipe it answers; got: %v", err)
	}
}

func TestAVersionlessReportIsAcceptedWithAWarning(t *testing.T) {
	root, cfg := feedbackFixture(t)
	versionless := strings.Replace(validReport, "Report back for the baseline recipe, version 1.0.0.", "Report back.", 1)
	_, warning, err := AddFeedback(root, cfg, "baseline", []byte(versionless))
	if err != nil {
		t.Fatalf("a scar with no version is still a scar — the report must be accepted: %v", err)
	}
	if !strings.Contains(warning, "version") {
		t.Fatalf("...but the author must be told to ask which version was built; got: %q", warning)
	}
}
