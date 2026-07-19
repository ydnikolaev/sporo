package recipekit

import (
	"os"
	"strings"
	"testing"
)

// The teeth of recipe-lint. Every fixture is ISOLATED — it carries exactly ONE violation —
// because a fixture with two defects can be red for the wrong reason, and a gate that is
// green only by accident is worse than a missing one. The FIRST fixture is the conformant
// baseline: without it, every "the gate reds" assertion below passes for a gate that reds on
// everything, and the shell version of this suite proved that is not hypothetical (a broken
// regex made three neutrality fixtures red for a reason that had nothing to do with
// neutrality, and they all looked like passes).

const conformant = `<!-- SSOT SOURCE (mate repo). -->

---
id: 01ARZ3NDEKTSV4RRFFQ69G5FAV
name: baseline
version: 1.0.0
title: A conformant recipe
problem: The gate must have something it does NOT red on.
prerequisites: [read-files]
derived_from: [the fixture itself]
stack: { language: go, runtime: any, why: "one static binary" }
verified: { project: fixture, release: v0.0.0, date: 2026-07-14 }
effort: reference
---

# Baseline

## Summary
This capability builds a small, mechanically checked record so the rest of the fixture can prove
the genre gate accepts one complete transferable capability before each isolated failure.

## The problem
You do not have the thing. You have it when the check passes.

## Why the obvious approach fails
The obvious approach hardcodes the origin and stops working anywhere else.

## The principles
Derive, never restate.

## The ground it needs
A machine-readable source of truth, because prose cannot be gated.

## The contracts

The shape the collector emits — **Binding: adapt** (rename the fields into your own language):

` + "```json" + `
{ "schema": 1, "counted": 12, "absent": { "reachable": false, "reason": "no such source here" } }
` + "```" + `

## The build sequence

### 1. Stand up the record
Write the source of truth first.

**Done when:** the record parses.

## The seams
The vocabulary the project owns.

## The scars

### The check that could not fire
**Symptom:** always green.
**Root cause:** it searched for a grammar nothing emits.
**Fix:** hold the parser to a fixture of the real grammar.

## Verification
A gate, with teeth.

## The trade-offs
It costs a build step. Do not build it for a one-off.

## For the human
It was built on a compiled language with a single binary — essential: it runs with no
checkout. Incidental: the test runner. It cost a longer build.
`

// lintFixture runs the pure genre lint (no injected conform-layer hook — the conformant
// baseline declares an adapt contract with no fixtures, so the two agree).
func lintFixture(t *testing.T, body string) []Finding {
	t.Helper()
	return Lint("fixture.md", []byte(body), []string{"mate", "axon"})
}

func TestTheConformantBaselineIsGreen(t *testing.T) {
	if f := lintFixture(t, conformant); len(f) != 0 {
		t.Fatalf("the baseline must be green, or every red below is meaningless; got: %v", f)
	}
}

func TestAMissingSectionReds(t *testing.T) {
	// The section that carries the whole point for the human reader — and the newest, so the
	// one most likely to be skipped by an author porting an older recipe forward.
	body := strings.Replace(conformant, "## For the human", "## Postscript", 1)
	assertRed(t, lintFixture(t, body), "For the human")
}

func TestASummaryThatIsOnlyALabelReds(t *testing.T) {
	body := strings.Replace(conformant,
		"This capability builds a small, mechanically checked record so the rest of the fixture can prove\n"+
			"the genre gate accepts one complete transferable capability before each isolated failure.",
		"A checked record.", 1)
	assertRed(t, lintFixture(t, body), "at least 80")
}

// The sequence IS the argument: principles before the build, cost after it. A recipe that
// opens with its steps is a checklist wearing the genre's headings.
func TestSectionsOutOfOrderRed(t *testing.T) {
	lines := strings.Split(conformant, "\n")
	var moved, rest []string
	in := false
	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "## The build sequence"):
			in = true
		case in && strings.HasPrefix(l, "## "):
			in = false
		}
		if in {
			moved = append(moved, l)
		} else {
			rest = append(rest, l)
		}
	}
	// Splice the build sequence in ahead of the principles — every section still present.
	var out []string
	for _, l := range rest {
		if strings.HasPrefix(l, "## The principles") {
			out = append(out, moved...)
		}
		out = append(out, l)
	}
	assertRed(t, lintFixture(t, strings.Join(out, "\n")), "out of order")
}

func TestAStepWithoutItsAcceptanceReds(t *testing.T) {
	body := strings.Replace(conformant, "**Done when:** the record parses.", "", 1)
	assertRed(t, lintFixture(t, body), "Done when")
}

func TestAScarMissingItsRootCauseReds(t *testing.T) {
	body := strings.Replace(conformant, "**Root cause:** it searched for a grammar nothing emits.", "it just broke.", 1)
	assertRed(t, lintFixture(t, body), "Root cause")
}

// The loop's anchor. A report-back binds to the version its author actually built; a recipe
// with no version — or one that cannot be ordered — makes every report ambiguous the day the
// text changes.
func TestAMissingIDReds(t *testing.T) {
	// The id is minted by `sporo new`, so a recipe that lacks one was hand-assembled or ported
	// from before ids existed — either way it has no permalink, and the gate must say so.
	body := strings.Replace(conformant, "id: 01ARZ3NDEKTSV4RRFFQ69G5FAV\n", "", 1)
	assertRed(t, lintFixture(t, body), "id")
}

func TestAnIDThatIsNotAULIDReds(t *testing.T) {
	// A hand-typed id is exactly what mints two recipes onto the same permalink. `l` and `o`
	// are outside Crockford base32, and the value is not 26 chars — either alone must red.
	body := strings.Replace(conformant, "id: 01ARZ3NDEKTSV4RRFFQ69G5FAV", "id: not-a-real-ulid", 1)
	assertRed(t, lintFixture(t, body), "ULID")
}

func TestAMissingVersionReds(t *testing.T) {
	body := strings.Replace(conformant, "version: 1.0.0\n", "", 1)
	assertRed(t, lintFixture(t, body), "version")
}

func TestAVersionThatIsNotSemverReds(t *testing.T) {
	body := strings.Replace(conformant, "version: 1.0.0", "version: latest", 1)
	assertRed(t, lintFixture(t, body), "semver")
}

func TestAMissingStackStampReds(t *testing.T) {
	// A recipe with no stack tells the reader to trust a recommendation whose ground they
	// cannot see. The gate is what makes the requirement real.
	body := strings.Replace(conformant, "stack: { language: go, runtime: any, why: \"one static binary\" }\n", "", 1)
	assertRed(t, lintFixture(t, body), "stack")
}

// The three neutrality reds. Each names a COORDINATE — the thing that executes in one
// repository and transfers to none.

func TestAProductNameInTheBodyReds(t *testing.T) {
	body := strings.Replace(conformant, "Derive, never restate.", "Derive it the way mate does.", 1)
	assertRed(t, lintFixture(t, body), "product")
}

func TestAFilenameInTheBodyReds(t *testing.T) {
	body := strings.Replace(conformant, "Derive, never restate.", "Derive it into `facts.json`.", 1)
	assertRed(t, lintFixture(t, body), "FILE")
}

func TestAPathInTheBodyReds(t *testing.T) {
	body := strings.Replace(conformant, "Derive, never restate.", "The collector lives in `internal/report/`.", 1)
	assertRed(t, lintFixture(t, body), "PATH")
}

// The bare-path teeth: a coordinate does not need backticks to execute in exactly one
// repository. And the greens are asserted as hard as the red — the conservative pattern
// exists precisely so prose arithmetic, initialisms and links never pay for it.

func TestABarePathWithoutBackticksReds(t *testing.T) {
	body := strings.Replace(conformant, "Derive, never restate.", "The collector lives in internal/report/facts.go beside the parser.", 1)
	assertRed(t, lintFixture(t, body), "PATH")
}

func TestProseSlashesAndLinksStayGreen(t *testing.T) {
	body := strings.Replace(conformant, "Derive, never restate.",
		"Available 24/7, and/or over TCP/IP; the schema is published at "+
			"https://example.com/static/v2/schema.json and mirrored on docs.example.com/latest/spec.json.", 1)
	if f := lintFixture(t, body); len(f) != 0 {
		t.Fatalf("fractions, alternations, initialisms and links are prose, not coordinates; got: %v", f)
	}
}

// The contracts teeth. A shape described in prose is a shape every reader re-invents,
// incompatibly — so the section must SHOW one.

func TestAContractsSectionWithNoShapeReds(t *testing.T) {
	// Strip the two fence markers and nothing else: the section still TALKS about the shape,
	// which is exactly the failure — it reads as complete to its author and hands the reader
	// nothing to copy.
	body := strings.Replace(conformant, "```json", "", 1)
	body = strings.Replace(body, "```", "", 1)
	assertRed(t, lintFixture(t, body), "fenced block")
}

// The binding teeth: a shown shape must say which kind it is. `exact` is somebody else's
// parser; `adapt` is anti-re-invention. A shape that says neither is how one team's
// "example" becomes another team's broken feed.

func TestAContractFenceWithoutABindingReds(t *testing.T) {
	body := strings.Replace(conformant,
		"The shape the collector emits — **Binding: adapt** (rename the fields into your own language):",
		"The shape the collector emits:", 1)
	assertRed(t, lintFixture(t, body), "Binding")
}

func TestABindingWithAThirdWordReds(t *testing.T) {
	body := strings.Replace(conformant, "**Binding: adapt**", "**Binding: strict-ish**", 1)
	assertRed(t, lintFixture(t, body), "exact")
}

func TestAnExactBindingIsGreen(t *testing.T) {
	body := strings.Replace(conformant, "**Binding: adapt** (rename the fields into your own language)",
		"**Binding: exact** (the fleet's collector parses this shape)", 1)
	if f := lintFixture(t, body); len(f) != 0 {
		t.Fatalf("an exact binding is the marker working, not a violation; got: %v", f)
	}
}

// The regression guard for a hole this gate never had — and nearly acquired. The first
// implementer of the corpus's own first recipe scored it 3/10 on copy-paste artifacts, and
// the story that formed was "our neutrality gate makes a schema example un-writable, so we
// must exempt fenced blocks from the coordinate scan". Running the gate on a real fenced
// example took two minutes and showed the story was false: the coordinate patterns are
// backtick-anchored, a fence carries no backticks, and a schema example was always green.
// The recipe had no examples because nobody wrote any.
//
// This test is what keeps that true. If someone tightens the coordinate patterns into a
// bare-token scan, it goes red HERE — before anyone reaches for an exemption that would
// let a real path hide inside a fence.
func TestASchemaExampleInAFenceIsGreenWithoutAnyExemption(t *testing.T) {
	body := strings.Replace(conformant, "Derive, never restate.",
		"The collector emits this shape:\n\n"+
			"```json\n"+
			`{ "date": "2026-07-14", "sources": { "vcs": { "reachable": true, "commits": 16 } },`+"\n"+
			`  "method": { "sessions": "a gap over 45 minutes opens a new one — a proxy, not a measurement" } }`+"\n"+
			"```\n\n"+
			"and the surface each runtime declares:\n\n"+
			"```yaml\n"+
			"runtimes:\n"+
			"  your-runtime:\n"+
			"    verified_build: \"2026-07-14\"\n"+
			"    usage_path: message.usage\n"+
			"```", 1)
	if f := lintFixture(t, body); len(f) != 0 {
		t.Fatalf("a fenced schema is a CONTRACT — a shape the reader copies and adapts — and it "+
			"transfers as well as the principle it serves; it must be green with no exemption. got: %v", f)
	}
}

// ...and the other half of the same line: a fence is not a sanctuary. A product name inside
// an example is still a name that means nothing in the reader's repository — an example is
// the likeliest place for one to leak, because it is copied out of a working tree.
func TestAProductNameInsideAFenceStillReds(t *testing.T) {
	body := strings.Replace(conformant, "Derive, never restate.",
		"```json\n{ \"project\": \"mate\", \"commits\": 16 }\n```", 1)
	assertRed(t, lintFixture(t, body), "product")
}

// A TECHNOLOGY is not a coordinate. This is the line the genre turns on, and getting it
// wrong in either direction destroys the document: ban technologies and the stack section
// becomes unwritable (the reader can no longer weigh the recommendation); allow coordinates
// and the recipe is a manual again. So the green is asserted as hard as the reds.
func TestNamingTheStackIsGreen(t *testing.T) {
	body := strings.Replace(conformant, "Derive, never restate.",
		"Build it as a single statically-linked Go binary with an embedded YAML registry, "+
			"because the reader can then run it with no checkout of the source. On a JS stack, "+
			"a bundled CLI buys the same property.", 1)
	if f := lintFixture(t, body); len(f) != 0 {
		t.Fatalf("technologies named as a choice with a reason are PORTABLE and must stay green; got: %v", f)
	}
}

// The appendix is the one section where instances are allowed — it is explicitly an
// illustration, and everything above it stands without it.
func TestTheAppendixMayNameInstances(t *testing.T) {
	body := conformant + "\n## Appendix — how one harness did it\nThe collector is `internal/report/facts.go` in mate.\n"
	if f := lintFixture(t, body); len(f) != 0 {
		t.Fatalf("the appendix may name instances; got: %v", f)
	}
}

// The off-by-one that the shell version's teeth actually caught: with no appendix, the scan
// window's end bound must be PAST the last line, or a violation on the final line — exactly
// where a hurried author appends one — falls outside it and the gate is blind.
func TestAViolationOnTheVeryLastLineReds(t *testing.T) {
	body := strings.TrimRight(conformant, "\n") + "\nSee `internal/report/facts.go`."
	assertRed(t, lintFixture(t, body), "FILE")
}

func TestTheOptOutIsLinePreciseAndMustCarryAReason(t *testing.T) {
	body := strings.Replace(conformant, "Derive, never restate.",
		"Derive it into `facts.json`. <!-- recipe-lint: allow the schema's own name is the subject here -->", 1)
	if f := lintFixture(t, body); len(f) != 0 {
		t.Fatalf("the line-precise opt-out must be honored; got: %v", f)
	}
}

// The genre spec teaches the shape rather than instantiating it. Holding it to the sections
// would be the gate testing its own habit — and the real one on disk is the fixture, so this
// also proves the corpus we ship is green.
func TestTheGenreSpecIsHeldToTheBannerOnly(t *testing.T) {
	f := Lint("_authoring.md", []byte("<!-- SSOT SOURCE -->\nno sections here at all\n"), []string{"mate"})
	if len(f) != 0 {
		t.Fatalf("the meta-document is held to the banner alone; got: %v", f)
	}
	f = Lint("_authoring.md", []byte("# no banner\n"), []string{"mate"})
	if len(f) == 0 {
		t.Fatal("...but it still needs the provenance banner")
	}
}

// IsDraft rides the frontmatter reader — a draft is the one state exempt from the gate.
func TestIsDraft(t *testing.T) {
	draft := strings.Replace(conformant, "effort: reference", "effort: reference\ndraft: true", 1)
	if !IsDraft([]byte(draft)) {
		t.Fatal("a recipe declaring `draft: true` must read as a draft")
	}
	if IsDraft([]byte(conformant)) {
		t.Fatal("the conformant baseline is not a draft")
	}
}

// The protocol ships to strangers BY DEFINITION — it is the one text in the corpus that every
// exported recipe carries — so it is held to the same neutrality as any recipe body. A
// coordinate here would leak into every recipe at once, which is the whole reason it is worth
// a test of its own: the blast radius is the corpus. (Moved here from internal/recipe with
// the neutrality scan it exercises; the corpus lives at the same relative depth.)
func TestTheAdoptionProtocolIsNeutral(t *testing.T) {
	const adoptionDoc = "_adoption.md"
	src, err := os.ReadFile("../../recipes/" + adoptionDoc)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(src), "\n")
	if f := neutrality(adoptionDoc, lines, 0, []string{"sporo"}); len(f) != 0 {
		t.Fatalf("the adoption protocol names a coordinate — and it rides on EVERY exported recipe:\n  %v", f)
	}
	for _, want := range []string{"## Adopt it here", "## Report back"} {
		if !strings.Contains(string(src), want) {
			t.Fatalf("the protocol is missing %q", want)
		}
	}
}

func assertRed(t *testing.T, findings []Finding, want string) {
	t.Helper()
	for _, f := range findings {
		if strings.Contains(f.Msg, want) {
			return
		}
	}
	t.Fatalf("expected a finding mentioning %q; got: %v", want, findings)
}
