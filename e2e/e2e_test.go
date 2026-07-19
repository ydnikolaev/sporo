// Package e2e drives the BUILT BINARY through the whole product loop in a repository that
// did not exist a second ago — because that is the only test that answers the question the
// unit suite cannot: does the tool a stranger downloads actually work in a repo that has
// never heard of it? Everything is isolated: a fresh temp repo, a fresh SPORO_HOME, the
// binary built with GOWORK=off (a green build under a workspace proves nothing about a
// fresh checkout).
//
// The steps run IN ORDER inside one test, sharing one repo — deliberately. The loop is the
// product (init → author → lint → seal → export → review → feedback → update), and testing
// each verb against a synthetic midpoint would never catch a verb that breaks the state its
// successor needs.
package e2e

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var bin string

func TestMain(m *testing.M) {
	// Wrapped so the temp-dir cleanup defer actually runs — os.Exit skips defers.
	os.Exit(run(m))
}

func run(m *testing.M) int {
	dir, err := os.MkdirTemp("", "sporo-e2e-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	bin = filepath.Join(dir, "sporo")
	build := exec.Command("go", "build", "-o", bin, "sporo.dev/sporo/cmd/sporo")
	build.Env = append(os.Environ(), "GOWORK=off")
	if out, err := build.CombinedOutput(); err != nil {
		panic("GOWORK=off build failed — a fresh checkout would not build either:\n" + string(out))
	}
	return m.Run()
}

// world is the isolated universe one end-to-end run lives in.
type world struct {
	repo string // the consumer repository
	home string // SPORO_HOME — the machine-level state, never the developer's real one
}

func (w world) run(t *testing.T, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = w.repo
	// SPORO_NO_UPDATE_CHECK: the passive version hint must never let a test suite reach the
	// network — a hermetic run is the point of this package.
	cmd.Env = append(os.Environ(), "SPORO_HOME="+w.home, "SPORO_NO_UPDATE_CHECK=1")
	var out, errb strings.Builder
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	code = 0
	ee := &exec.ExitError{}
	if errors.As(err, &ee) {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("sporo %v did not run at all: %v", args, err)
	}
	return out.String(), errb.String(), code
}

func (w world) mustRun(t *testing.T, args ...string) string {
	t.Helper()
	out, errb, code := w.run(t, args...)
	if code != 0 {
		t.Fatalf("sporo %v exited %d\nstdout: %s\nstderr: %s", args, code, out, errb)
	}
	return out
}

func (w world) write(t *testing.T, rel, content string) {
	t.Helper()
	abs := filepath.Join(w.repo, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (w world) read(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(w.repo, rel))
	if err != nil {
		t.Fatalf("expected %s to exist: %v", rel, err)
	}
	return string(b)
}

// A conformant recipe, as a stranger would author one. Duplicated from the unit fixtures on
// purpose: this package tests the binary, not the library, and importing the library's
// fixture would quietly couple the two suites.
const recipeV1 = `<!-- SSOT SOURCE (this project). -->

---
id: 01ARZ3NDEKTSV4RRFFQ69G5FAV
name: nightly-digest
version: 1.0.0
title: A nightly digest that checks itself
problem: The record of a day's work is invisible to the people funding it.
prerequisites: [read-files, run-shell]
derived_from: [one live build]
stack: { language: go, runtime: any, why: "one static binary" }
verified: { project: e2e-consumer, release: v0.1.0, date: 2026-07-15 }
effort: an evening
---

# Nightly digest

## Summary
This capability builds a small checked record and a readable nightly digest, preserving the
facts an agent needs to reproduce the result without inheriting the originating repository.

## The problem
You do not have the thing. You have it when the check passes.

## Why the obvious approach fails
The obvious approach hardcodes the origin and stops working anywhere else.

## The principles
Derive, never restate.

## The ground it needs
A machine-readable source of truth, because prose cannot be gated.

## The contracts

The record it persists — **Binding: adapt** (rename fields into your language):

` + "```json" + `
{ "schema": 1, "counted": 12 }
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
Built on a compiled language with a single binary — essential: it runs with no checkout.
Incidental: the test runner. It cost a longer build.
`

const reportBack = `Recipe: nightly-digest, version 1.0.0.

**Stack:** rebuilt on a scripting runtime; kept the single-artifact property.
**Degraded:** no decision log here — labelled the rationale section "reconstructed".
**New scars:** none beyond the recipe's own.
**Wrong:** nothing contradicted.
**Arithmetic:** ran the live check; it matched.
**Missing:** the shape of the summary record — had to design it.
`

func TestEndToEnd(t *testing.T) {
	w := world{repo: t.TempDir(), home: t.TempDir()}
	if err := os.MkdirAll(filepath.Join(w.repo, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Run("init installs the surface and is idempotent", func(t *testing.T) {
		w.mustRun(t, "init")
		for _, rel := range []string{".claude/skills/sporo-recipe/SKILL.md", "AGENTS.md", ".sporo/config.yaml", ".sporo/recipes/README.md", ".sporo/registry.yaml"} {
			w.read(t, rel)
		}
		before := w.read(t, ".sporo/registry.yaml")
		out := w.mustRun(t, "init")
		if strings.Contains(out, "wrote") || strings.Contains(out, "seeded") || strings.Contains(out, "updated") {
			t.Fatalf("second init claims to have changed something:\n%s", out)
		}
		if w.read(t, ".sporo/registry.yaml") != before {
			t.Fatal("second init changed the registry")
		}
	})

	t.Run("a scaffold is a draft: coached, gate-exempt, and unable to ship", func(t *testing.T) {
		out := w.mustRun(t, "new", "half-baked", "--title", "A capability still cooking")
		if !strings.Contains(out, "draft") {
			t.Fatalf("new must say it wrote a draft: %s", out)
		}
		if lint := w.mustRun(t, "lint"); !strings.Contains(lint, "draft(s) not checked") {
			t.Fatalf("lint must report the draft it skipped, or a half-written file reads as done: %s", lint)
		}
		if _, stderr, code := w.run(t, "seal", "half-baked"); code == 0 || !strings.Contains(stderr, "draft") {
			t.Fatalf("seal must refuse a draft (code %d): %s", code, stderr)
		}
		if _, stderr, code := w.run(t, "export", "half-baked"); code == 0 || !strings.Contains(stderr, "draft") {
			t.Fatalf("export must refuse a draft (code %d): %s", code, stderr)
		}
		// Finishing = the scaffold's own instruction: remove the draft mark. The scaffold is
		// genre-green by construction, so that one removal is all it takes here.
		src := w.read(t, ".sporo/recipes/half-baked.md")
		w.write(t, ".sporo/recipes/half-baked.md", strings.Replace(src, "draft: true\n", "", 1))
		// The authoring gate is `lint <home>` — genre only. No-arg `lint` is the SHIPPED gate and
		// would now red on a finished-but-unsealed recipe (the all-sealed sweep); a recipe must be
		// genre-clean BEFORE it is sealed, so the pre-seal check cannot be the one that demands a seal.
		w.mustRun(t, "lint", ".sporo/recipes/")
		w.mustRun(t, "seal", "half-baked")
	})

	t.Run("author, lint, seal", func(t *testing.T) {
		w.write(t, ".sporo/recipes/nightly-digest.md", recipeV1)
		// Genre check before sealing (see the note above): `lint <home>`, not the no-arg shipped gate.
		out := w.mustRun(t, "lint", ".sporo/recipes/")
		if !strings.Contains(out, "conformant") {
			t.Fatalf("lint did not report a green corpus: %s", out)
		}
		if out := w.mustRun(t, "seal", "nightly-digest"); !strings.Contains(out, "1.0.0") {
			t.Fatalf("seal did not report the sealed version: %s", out)
		}
		w.mustRun(t, "seal", "nightly-digest") // idempotent
	})

	t.Run("a silent edit under a sealed version is caught by lint and refused by seal", func(t *testing.T) {
		w.write(t, ".sporo/recipes/nightly-digest.md", strings.Replace(recipeV1, "Derive, never restate.", "Derive, always.", 1))
		if _, stderr, code := w.run(t, "lint"); code == 0 || !strings.Contains(stderr, "drifted") {
			t.Fatalf("lint must red on content drifted from its seal (code %d): %s", code, stderr)
		}
		if _, stderr, code := w.run(t, "seal", "nightly-digest"); code == 0 || !strings.Contains(stderr, "bump") {
			t.Fatalf("seal must refuse a changed text under an unchanged version (code %d): %s", code, stderr)
		}
		bumped := strings.Replace(recipeV1, "version: 1.0.0", "version: 1.1.0", 1)
		bumped = strings.Replace(bumped, "Derive, never restate.", "Derive, always.", 1)
		w.write(t, ".sporo/recipes/nightly-digest.md", bumped)
		w.mustRun(t, "seal", "nightly-digest")
		w.mustRun(t, "lint")
	})

	t.Run("an exact contract change demands a major version", func(t *testing.T) {
		exact := strings.Replace(recipeV1, "name: nightly-digest", "name: crm-feed", 1)
		exact = strings.Replace(exact, "**Binding: adapt** (rename fields into your language)",
			"**Binding: exact** (the fleet's aggregator parses this shape)", 1)
		w.write(t, ".sporo/recipes/crm-feed.md", exact)
		w.mustRun(t, "seal", "crm-feed")

		broken := strings.Replace(exact, `"counted": 12`, `"tallied": 12`, 1)
		broken = strings.Replace(broken, "version: 1.0.0", "version: 1.1.0", 1)
		w.write(t, ".sporo/recipes/crm-feed.md", broken)
		if _, stderr, code := w.run(t, "seal", "crm-feed"); code == 0 || !strings.Contains(stderr, "MAJOR") {
			t.Fatalf("a renamed field in an exact shape under a minor bump is a fleet break and must be refused (code %d): %s", code, stderr)
		}
		major := strings.Replace(broken, "version: 1.1.0", "version: 2.0.0", 1)
		w.write(t, ".sporo/recipes/crm-feed.md", major)
		w.mustRun(t, "seal", "crm-feed")
	})

	t.Run("conform holds every project's output to the exact shape — from the export alone", func(t *testing.T) {
		handoff := w.mustRun(t, "export", "crm-feed", "--stdout")
		w.write(t, "handoff.md", handoff)
		w.write(t, "feed.json", `{ "schema": 3, "tallied": 40 }`)
		if out := w.mustRun(t, "conform", "handoff.md", "feed.json"); !strings.Contains(out, "✓") {
			t.Fatalf("a conforming feed must pass against the exported file — the only document a reader has: %s", out)
		}
		w.write(t, "bad-feed.json", `{ "schema": 3, "counted": 40 }`)
		if _, stderr, code := w.run(t, "conform", "handoff.md", "bad-feed.json"); code == 0 || !strings.Contains(stderr, "tallied") {
			t.Fatalf("the renamed field must fail with a path naming what the consumer will miss (code %d): %s", code, stderr)
		}
		// The ADR-005 posture, end to end: an adapt-only recipe has nothing to conform to.
		if out := w.mustRun(t, "conform", "nightly-digest", "feed.json"); !strings.Contains(out, "declares no exact-bound contracts") {
			t.Fatalf("an adapt-only recipe is a clean no-op, not a failure: %s", out)
		}
	})

	t.Run("export composes the deliverable", func(t *testing.T) {
		out := w.mustRun(t, "export", "nightly-digest", "--stdout")
		if !strings.Contains(out, "## Adopt it here") || !strings.Contains(out, "## Report back") {
			t.Fatal("the export must carry the adoption protocol")
		}
		if strings.Contains(out, "<!-- SSOT SOURCE") {
			t.Fatal("the banner is house business and must be stripped")
		}
		// The default (no --stdout) writes the composed file instead of printing it — the
		// delivery contract, since export exists to HAND a recipe over.
		msg := w.mustRun(t, "export", "nightly-digest")
		if !strings.Contains(msg, ".sporo/exports/nightly-digest.md") {
			t.Fatalf("export must report the file it wrote: %s", msg)
		}
		if got := w.read(t, ".sporo/exports/nightly-digest.md"); !strings.Contains(got, "## Adopt it here") {
			t.Fatal("the written export must carry the adoption protocol")
		}
	})

	t.Run("review pack is self-contained and a verdict lands in the registry", func(t *testing.T) {
		w.mustRun(t, "review", "nightly-digest")
		prompt := w.read(t, ".sporo/review/nightly-digest/prompt.md")
		for _, want := range []string{"intent_clarity", "security_hazards", "RECIPE UNDER REVIEW", "## Adopt it here"} {
			if !strings.Contains(prompt, want) {
				t.Fatalf("the review prompt must be self-contained; missing %q", want)
			}
		}
		axes := []string{"intent_clarity", "scars_value", "task_setting", "build_readiness", "copy_paste_artifacts", "stack_neutrality", "security_hazards"}
		verdict := map[string]any{"schema": 1, "recipe": "nightly-digest", "version": "1.1.0", "verdict": "adopt", "top_gaps": []string{"none"}}
		am := map[string]any{}
		for _, a := range axes {
			am[a] = map[string]any{"score": 8, "note": "credible"}
		}
		verdict["axes"] = am
		vb, _ := json.Marshal(verdict)
		w.write(t, "verdict.json", string(vb))
		out := w.mustRun(t, "review", "verify", "nightly-digest", "verdict.json")
		if !strings.Contains(out, "adopt") {
			t.Fatalf("verify must report the verdict: %s", out)
		}
		if !strings.Contains(w.read(t, ".sporo/registry.yaml"), "reviews") {
			t.Fatal("a verified review must be recorded in the registry")
		}
	})

	t.Run("feedback files a report-back, idempotently", func(t *testing.T) {
		w.write(t, "report.md", reportBack)
		w.mustRun(t, "feedback", "add", "nightly-digest", "report.md")
		w.mustRun(t, "feedback", "add", "nightly-digest", "report.md")
		list := w.mustRun(t, "feedback", "list")
		if !strings.Contains(list, "nightly-digest") {
			t.Fatalf("the filed report must be listed: %s", list)
		}
		ents, err := os.ReadDir(filepath.Join(w.repo, ".sporo", "feedback", "nightly-digest"))
		if err != nil || len(ents) != 1 {
			t.Fatalf("a byte-identical report filed twice must exist once, got %d (%v)", len(ents), err)
		}
	})

	t.Run("a bundle composes into one document with one protocol", func(t *testing.T) {
		second := strings.Replace(recipeV1, "name: nightly-digest", "name: weekly-rollup", 1)
		w.write(t, ".sporo/recipes/weekly-rollup.md", second)
		w.write(t, ".sporo/recipes/reporting.bundle.yaml", "bundle: reporting\ntitle: The reporting stack\nmembers: [nightly-digest, weekly-rollup]\n")
		out := w.mustRun(t, "export", "--bundle", "reporting", "--stdout")
		if n := strings.Count(out, "## Adopt it here"); n != 1 {
			t.Fatalf("one composition, one protocol; got %d", n)
		}
		if strings.Index(out, "name: nightly-digest") >= strings.Index(out, "name: weekly-rollup") {
			t.Fatal("members must compose in build order")
		}
		w.mustRun(t, "seal", "weekly-rollup") // seal the new member so the shipped corpus is whole
		w.mustRun(t, "lint")                  // the manifest rides the same gate, and a valid one stays green
	})

	t.Run("adopt records the handover, pull is loud when an exact contract moves", func(t *testing.T) {
		// handoff.md (the crm-feed 2.0.0 export) exists from the conform subtest; adopt it.
		out := w.mustRun(t, "adopt", "handoff.md")
		if !strings.Contains(out, "crm-feed 2.0.0") || !strings.Contains(out, "conform") {
			t.Fatalf("adopt must anchor slug+version and point an exact-carrying recipe at conform: %s", out)
		}
		// The source moves forward with a changed exact fence — the case a consumer-feeding
		// build must not sleep through.
		bumped := strings.Replace(w.read(t, "handoff.md"), "version: 2.0.0", "version: 3.0.0", 1)
		bumped = strings.Replace(bumped, `"tallied": 12`, `"relabeled": 12`, 1)
		w.write(t, "handoff.md", bumped)
		out = w.mustRun(t, "pull")
		if !strings.Contains(out, "2.0.0 → 3.0.0") || !strings.Contains(out, "EXACT contract changed") {
			t.Fatalf("pull must report the delta and shout about the moved exact contract: %s", out)
		}
		w.mustRun(t, "pull", "--apply")
		if list := w.mustRun(t, "list"); !strings.Contains(list, "adopted") || !strings.Contains(list, "crm-feed (3.0.0)") {
			t.Fatalf("list must show the adopted recipe at its applied version: %s", list)
		}
	})

	t.Run("update never clobbers and projects knows this repo", func(t *testing.T) {
		skill := ".claude/skills/sporo-recipe/SKILL.md"
		edited := w.read(t, skill) + "\nlocal amendment\n"
		w.write(t, skill, edited)
		out := w.mustRun(t, "update")
		if !strings.Contains(out, "skipped") {
			t.Fatalf("an edited managed file must be reported: %s", out)
		}
		if w.read(t, skill) != edited {
			t.Fatal("update overwrote a user's edit — the one forbidden move")
		}
		abs, _ := filepath.EvalSymlinks(w.repo)
		projects := w.mustRun(t, "projects")
		if !strings.Contains(projects, abs) && !strings.Contains(projects, w.repo) {
			t.Fatalf("the global projects list must know this repo: %s", projects)
		}
	})

	t.Run("genre prints the spec anywhere", func(t *testing.T) {
		out := w.mustRun(t, "genre")
		if !strings.Contains(out, "Recipe — authoring") {
			t.Fatal("genre must print the authoring spec")
		}
	})
}
