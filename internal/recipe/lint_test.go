package recipe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The genre logic itself — the teeth of recipe-lint — moved to pkg/recipekit along with its
// fixture suite; those isolated-violation tests now live in pkg/recipekit/lint_test.go. What
// stays here is what needs THIS package's staying surface: the shared `conformant` fixture
// (used by conform/registry/fleet/bundle/scaffold tests), the `lintFixture`/`assertRed`
// helpers the conform tests reuse, and the two tests that exercise the staying `IsRecipe`.

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
This capability builds a small, mechanically checked record so the shared fixture can prove the
staying package accepts one complete transferable capability before exercising its callers.

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

func lintFixture(t *testing.T, body string) []Finding {
	t.Helper()
	return Lint("fixture.md", []byte(body), []string{"mate", "axon"})
}

// The shipped corpus is checked by the same code the gate runs — not by a copy of it. A
// genre spec whose own instances violate it is a spec nobody believes. This also exercises
// the Lint forwarder (recipe.Lint → recipekit.Lint + the injected fixtureFindings) against
// real recipes.
func TestTheShippedCorpusIsConformant(t *testing.T) {
	ents, err := os.ReadDir("../../recipes")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range ents {
		if !IsRecipe(e.Name(), e.IsDir()) {
			continue
		}
		src, err := os.ReadFile(filepath.Join("../../recipes", e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		n++
		if f := Lint(e.Name(), src, []string{"sporo"}); len(f) != 0 {
			t.Errorf("%s is not conformant:\n  %v", e.Name(), f)
		}
	}
	if n == 0 {
		t.Fatal("no corpus on disk — this test would pass vacuously, which is the failure it exists to prevent")
	}
}

// The cry-wolf teeth. mate SEEDS a README into every consumer's recipes home; if the gate
// treats it as a recipe, a project that has only ever run `pull` has a red gate for a change
// it did not make — and a gate that reds on the state it ships trains red-blindness, which is
// worse than no gate at all. Caught on the first live pull into a fresh consumer, not by any
// unit test: the fixtures never contained the file the harness itself writes.
func TestTheSeededReadmeIsNotARecipe(t *testing.T) {
	if IsRecipe("README.md", false) {
		t.Fatal("the recipes home's own README is not a recipe — treating it as one reds every freshly-pulled consumer")
	}
	if !IsRecipe("daily-progress-report.md", false) {
		t.Fatal("...and a real recipe still is one")
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
