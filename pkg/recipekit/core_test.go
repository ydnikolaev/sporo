package recipekit

import (
	"strings"
	"testing"
)

func TestContentHash(t *testing.T) {
	h := ContentHash([]byte("hello"))
	// sha256("hello"), the registry's fixed currency — a stable, prefixed hex string.
	const want = "sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if h != want {
		t.Fatalf("ContentHash mismatch:\n got %q\nwant %q", h, want)
	}
	if ContentHash([]byte("hello")) == ContentHash([]byte("world")) {
		t.Fatal("different bytes must hash differently")
	}
}

func TestFindingString(t *testing.T) {
	if got := (Finding{File: "a.md", Line: 12, Msg: "boom"}).String(); got != "a.md:12: boom" {
		t.Fatalf("located finding: got %q", got)
	}
	if got := (Finding{File: "a.md", Line: 0, Msg: "whole file"}).String(); got != "a.md: whole file" {
		t.Fatalf("whole-file finding: got %q", got)
	}
}

func TestSemverMajor(t *testing.T) {
	cases := map[string]int{"1.2.3": 1, "12.0.0": 12, "0.9.9": 0, "latest": 0, "": 0, "3": 3}
	for v, want := range cases {
		if got := SemverMajor(v); got != want {
			t.Errorf("SemverMajor(%q) = %d, want %d", v, got, want)
		}
	}
}

func TestSemverTriple(t *testing.T) {
	if tr, ok := SemverTriple("v1.2.3-rc1"); !ok || tr != [3]int{1, 2, 3} {
		t.Fatalf("v-prefix + prerelease: got %v ok=%v", tr, ok)
	}
	if tr, ok := SemverTriple("2.10.0+build.7"); !ok || tr != [3]int{2, 10, 0} {
		t.Fatalf("build metadata: got %v ok=%v", tr, ok)
	}
	for _, bad := range []string{"1.2", "1.2.x", "latest", "1.2.3.4"} {
		if _, ok := SemverTriple(bad); ok {
			t.Errorf("SemverTriple(%q) should not parse", bad)
		}
	}
}

func TestSemverNewer(t *testing.T) {
	if !SemverNewer("1.1.0", "1.0.9") {
		t.Fatal("1.1.0 is newer than 1.0.9")
	}
	if SemverNewer("1.0.0", "1.0.0") {
		t.Fatal("equal versions are not newer")
	}
	if SemverNewer("1.0.0", "2.0.0") {
		t.Fatal("older is not newer")
	}
	if SemverNewer("garbage", "1.0.0") {
		t.Fatal("an unparseable candidate is never newer")
	}
}

func TestFrontmatterValue(t *testing.T) {
	// A source shape (banner on line 0) and an export shape (frontmatter IS line 0) must both read.
	source := "<!-- SSOT SOURCE -->\n---\nname: baseline\nversion: \"1.2.0\"\n---\nbody"
	export := "---\nname: baseline\nversion: 1.2.0\n---\nbody"
	for _, src := range []string{source, export} {
		if got := FrontmatterValue([]byte(src), "name"); got != "baseline" {
			t.Errorf("name: got %q", got)
		}
		if got := FrontmatterValue([]byte(src), "version"); got != "1.2.0" {
			t.Errorf("version (quotes trimmed): got %q", got)
		}
	}
	if got := FrontmatterValue([]byte("no frontmatter here"), "name"); got != "" {
		t.Errorf("no frontmatter must read empty, got %q", got)
	}
}

func TestKeyLineAndSectionBody(t *testing.T) {
	fm := []string{"id: X", "name: baseline", "version: 1.0.0"}
	if got := KeyLine(fm, "name"); got != "name: baseline" {
		t.Fatalf("KeyLine: got %q", got)
	}
	if got := KeyLine(fm, "absent"); got != "" {
		t.Fatalf("KeyLine absent: got %q", got)
	}
	lines := strings.Split("## A\nalpha\nbeta\n## B\ngamma", "\n")
	body := SectionBody(lines, "## A")
	if strings.Join(body, ",") != "alpha,beta" {
		t.Fatalf("SectionBody: got %v", body)
	}
}

func TestExactContractsDigest(t *testing.T) {
	// An exact-bound fence contributes; an adapt one does not — so an all-adapt recipe digests
	// to empty, and flipping the binding to exact makes the digest non-empty.
	if d := ExactContractsDigest([]byte(conformant)); d != "" {
		t.Fatalf("the baseline's contract is adapt-bound — its exact digest must be empty, got %q", d)
	}
	exact := strings.Replace(conformant, "**Binding: adapt** (rename the fields into your own language)",
		"**Binding: exact**", 1)
	if d := ExactContractsDigest([]byte(exact)); d == "" {
		t.Fatal("an exact-bound fence must produce a non-empty digest")
	}
}

func TestStripBanner(t *testing.T) {
	got := StripBanner("<!-- SSOT SOURCE (mate) -->\n\n# Title\nbody")
	if got != "# Title\nbody" {
		t.Fatalf("StripBanner must drop the banner and leading blanks, got %q", got)
	}
	if got := StripBanner("# No banner\nbody"); got != "# No banner\nbody" {
		t.Fatalf("a bannerless document is untouched, got %q", got)
	}
}

const validReport = `Report back for the baseline recipe, version 1.0.0.

**Stack:** Node 22, a single script; kept the "one static artifact" choice, replaced the language.
**Degraded:** the decision log was absent — built the smallest one, labelled "reconstructed" in the output.
**New scars:** one. **Symptom:** double-counted rows. **Root cause:** two sources answering different questions. **Fix:** named the question each answers, stopped adding them.
**Wrong:** nothing contradicted the build.
**Arithmetic:** ran the live check; 3 of 3 matched.
**Missing:** the collector's error shape — had to invent it.
`

func TestValidateReport(t *testing.T) {
	if f := ValidateReport([]byte(validReport)); len(f) != 0 {
		t.Fatalf("a complete report-back must validate clean; got: %v", f)
	}
	broken := strings.Replace(validReport, "**New scars:**", "**Findings:**", 1)
	f := ValidateReport([]byte(broken))
	if len(f) == 0 {
		t.Fatal("a report missing a marker must be reported")
	}
	found := false
	for _, x := range f {
		if strings.Contains(x.Msg, "New scars") {
			found = true
		}
	}
	if !found {
		t.Fatalf("the missing marker must be named; got: %v", f)
	}
}
