package recipekit

import (
	"os"
	"strings"
	"testing"
)

// conformantSeed is the one green fixture: it exercises every seed check on its passing side —
// nine valid frontmatter keys, seven sections in order, an Install that detects first and carries
// a `**Done when:**` on each step, a blind pipe that cites the declared source, a Verify with a
// runnable proof, and the five fixed Report rows in order. It respects the S2 authoring
// constraint: no coordinate syntax in the body, so neutrality does not red it.
func conformantSeed(t *testing.T) string {
	t.Helper()
	src, err := os.ReadFile("testdata/seed/conformant.md")
	if err != nil {
		t.Fatal(err)
	}
	return string(src)
}

// lintSeed runs a seed document through the real registered SeedShape with an empty product list
// (a seed instance names its own tool; the product ban is not what these fixtures probe).
func lintSeed(name, src string) []Finding {
	return LintShape(SeedShape, name, []byte(src), nil)
}

func TestSeedConformantFixtureIsGreen(t *testing.T) {
	if f := lintSeed("acme.md", conformantSeed(t)); len(f) != 0 {
		t.Fatalf("the conformant seed fixture must lint green; got: %v", f)
	}
}

func TestSeedRedsMissingSection(t *testing.T) {
	src := strings.Replace(conformantSeed(t),
		"## Use\n\nPoint the tool at a build description and let it produce the first artifact, so the reader sees a real result in their own repository rather than a promise.\n\n",
		"", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, "missing section: Use") {
		t.Fatalf("a removed section must red; got: %v", f)
	}
}

func TestSeedRedsSectionOutOfOrder(t *testing.T) {
	// Swap the `## Use` and `## Harness` header lines, leaving their bodies in place — Harness
	// now appears where Use should, and the engine's order check reds it.
	s := strings.Replace(conformantSeed(t), "## Use", "@@SWAP@@", 1)
	s = strings.Replace(s, "## Harness", "## Use", 1)
	s = strings.Replace(s, "@@SWAP@@", "## Harness", 1)
	if f := lintSeed("acme.md", s); !hasFinding(f, "out of order") {
		t.Fatalf("a section out of order must red; got: %v", f)
	}
}

func TestSeedRedsMissingKey(t *testing.T) {
	src := strings.Replace(conformantSeed(t), "title: Bring in the acme build tool and stand it up\n", "", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, "missing `title:`") {
		t.Fatalf("a missing required key must red; got: %v", f)
	}
}

func TestSeedRedsBadID(t *testing.T) {
	src := strings.Replace(conformantSeed(t), "id: 01JQ8ZK5T9WXYZ0123456789AB", "id: not-a-ulid", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, "must be a ULID") {
		t.Fatalf("a non-ULID id must red; got: %v", f)
	}
}

func TestSeedRedsBadVersion(t *testing.T) {
	src := strings.Replace(conformantSeed(t), "version: 1.0.0", "version: 1.0", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, "must be a semver triple") {
		t.Fatalf("a non-semver version must red; got: %v", f)
	}
}

func TestSeedRedsBadVerified(t *testing.T) {
	src := strings.Replace(conformantSeed(t),
		"verified: { project: demo, release: v1.0.0, date: 2026-07-21 }",
		"verified: { release: v1.0.0, date: 2026-07-21 }", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, "must name the install that proves") {
		t.Fatalf("a verified stamp with no project must red; got: %v", f)
	}
}

func TestSeedRedsBadStack(t *testing.T) {
	src := strings.Replace(conformantSeed(t),
		`stack: { language: go, runtime: any, why: "the verifying install ran on a Go toolchain" }`,
		"stack: { runtime: any }", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, "must name what the verifying install actually ran on") {
		t.Fatalf("a stack stamp with no language must red; got: %v", f)
	}
}

func TestSeedRedsBadTarget(t *testing.T) {
	src := strings.Replace(conformantSeed(t), "target: acme@v2.3.0", "target: acme", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, "must be `<tool>@<version>`") {
		t.Fatalf("a target with no @version must red; got: %v", f)
	}
}

func TestSeedRedsEmptySource(t *testing.T) {
	src := strings.Replace(conformantSeed(t), "source: https://get.acme.example", "source:", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, "must name a canonical origin") {
		t.Fatalf("a blank source must red; got: %v", f)
	}
}

func TestSeedRedsInstallNotDetectFirst(t *testing.T) {
	src := strings.Replace(conformantSeed(t),
		"**Detect:** ask the tool for its version and note whether it answers, and at which release.\n",
		"", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, "must open with `**Detect:**`") {
		t.Fatalf("an Install whose first step does not detect must red; got: %v", f)
	}
}

func TestSeedRedsMissingDoneWhen(t *testing.T) {
	src := strings.Replace(conformantSeed(t),
		"**Done when:** the tool answers its version query in a fresh shell.\n",
		"", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, "`**Done when:**` line(s)") {
		t.Fatalf("an Install step with no acceptance must red; got: %v", f)
	}
}

func TestSeedRedsInstallWithNoSteps(t *testing.T) {
	s := strings.Replace(conformantSeed(t),
		"### Detect whether the tool is already here\n", "Detect whether the tool is already here\n", 1)
	s = strings.Replace(s,
		"### Acquire it from the vouched origin\n", "Acquire it from the vouched origin\n", 1)
	if f := lintSeed("acme.md", s); !hasFinding(f, "has no steps") {
		t.Fatalf("an Install with no `### ` steps must red; got: %v", f)
	}
}

func TestSeedRedsBlindPipeWithoutSourceCitation(t *testing.T) {
	src := strings.Replace(conformantSeed(t),
		"curl https://get.acme.example/install.sh | sh",
		"curl https://mirror.example/install.sh | sh", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, "without citing the `source` origin") {
		t.Fatalf("a blind pipe from an uncited origin must red; got: %v", f)
	}
}

func TestSeedRedsBlindPipeThroughSudo(t *testing.T) {
	// `curl … | sudo sh` is a common real install line; the interposed `sudo` must not let an
	// uncited remote-into-shell pipe slip past the tooth.
	src := strings.Replace(conformantSeed(t),
		"curl https://get.acme.example/install.sh | sh",
		"curl https://mirror.example/install.sh | sudo sh", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, "without citing the `source` origin") {
		t.Fatalf("a blind pipe through sudo from an uncited origin must red; got: %v", f)
	}
}

func TestSeedBlindPipeCitingSourceIsGreen(t *testing.T) {
	src := conformantSeed(t)
	if !reBlindPipe.MatchString(src) {
		t.Fatal("the conformant fixture must contain a fetch-into-shell pipe for this to prove anything")
	}
	for _, f := range lintSeed("acme.md", src) {
		if strings.Contains(f.Msg, "source") {
			t.Fatalf("a blind pipe that cites the declared source must stay green; got: %v", f)
		}
	}
}

func TestSeedRedsEmptyVerify(t *testing.T) {
	src := strings.Replace(conformantSeed(t),
		"Run the tool's own version query and read the output — a real command the agent observes, not a claim:\n\n```\nacme --version\n```\n",
		"Run the tool's own version query and read the output.\n", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, "no fenced command block") {
		t.Fatalf("a Verify with no runnable proof must red; got: %v", f)
	}
}

func TestSeedRedsShortSummary(t *testing.T) {
	src := strings.Replace(conformantSeed(t),
		"This seed brings in the acme build tool and stands it up in a repository that has never had it: it detects whether the tool is already present, installs it from the origin the frontmatter vouches for, and proves it runs before the reader relies on it. When it is done the reader has a working tool on their machine and a note left for the next agent.",
		"Too short.", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, "Summary is") {
		t.Fatalf("a Summary below the floor must red; got: %v", f)
	}
}

func TestSeedRedsReportExtraRow(t *testing.T) {
	src := strings.Replace(conformantSeed(t),
		"| **suggest next** | wire the tool into the repository's own agent harness |\n",
		"| **suggest next** | wire the tool into the repository's own agent harness |\n| **extra thing** | an unexpected addition to the audit |\n", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, "unexpected") {
		t.Fatalf("a Report with an extra row must red; got: %v", f)
	}
}

func TestSeedRedsReportMissingRow(t *testing.T) {
	src := strings.Replace(conformantSeed(t),
		"| **how to use it** | run the tool against a build description to produce the first artifact |\n",
		"", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, `is missing the "how to use it" row`) {
		t.Fatalf("a Report with a missing row must red; got: %v", f)
	}
}

func TestSeedRedsReportDuplicateRow(t *testing.T) {
	// A duplicated known label with all five rows still present: the set-based missing/unexpected
	// loops both see every label at least once, and the reorder check is skipped on the length
	// mismatch — only the cardinality assert (AUD-001) catches the six-row Report.
	src := strings.Replace(conformantSeed(t),
		"| **suggest next** | wire the tool into the repository's own agent harness |\n",
		"| **suggest next** | wire the tool into the repository's own agent harness |\n| **suggest next** | a duplicated known label the set-based check would let pass |\n", 1)
	if f := lintSeed("acme.md", src); !hasFinding(f, "duplicated or padded row") {
		t.Fatalf("a Report with a duplicated known-label row must red on cardinality; got: %v", f)
	}
}

func TestSeedRedsReportReorderedRows(t *testing.T) {
	s := strings.Replace(conformantSeed(t),
		"| **what it is** | the acme build tool, brought in and stood up |", "@@ROW@@", 1)
	s = strings.Replace(s,
		"| **how it works** | a single binary that renders a build description into an artifact |",
		"| **what it is** | the acme build tool, brought in and stood up |", 1)
	s = strings.Replace(s, "@@ROW@@",
		"| **how it works** | a single binary that renders a build description into an artifact |", 1)
	if f := lintSeed("acme.md", s); !hasFinding(f, "rows are out of order") {
		t.Fatalf("a Report with reordered rows must red; got: %v", f)
	}
}

// seedFixture reads a named seed fixture from testdata.
func seedFixture(t *testing.T, file string) string {
	t.Helper()
	src, err := os.ReadFile("testdata/seed/" + file)
	if err != nil {
		t.Fatal(err)
	}
	return string(src)
}

// VAL-2: a sporo-targeted seed that names sporo (the product) and its own artifacts — the bare
// word, the `.sporo/` convention, a bare `sporo/cmd/main.go` path — greens on BOTH axes, even with
// the repo product list naming sporo, because the neutrality policy erases the instance's own
// declared tool before the scan.
func TestSeedTargetExemptionGreensOwnArtifacts(t *testing.T) {
	src := seedFixture(t, "sporo.md")
	if f := LintShape(SeedShape, "sporo.md", []byte(src), []string{"sporo"}); len(f) != 0 {
		t.Fatalf("a sporo-targeted seed naming sporo and its own artifacts must lint green; got: %v", f)
	}
}

// VAL-3: an unrelated backticked coordinate in the same document still reds — the exemption removes
// only the target's own tokens, never the reader's tree.
func TestSeedNonTargetCoordinateReds(t *testing.T) {
	src := strings.Replace(seedFixture(t, "sporo.md"),
		"Point sporo at a recipe and let it export the first artifact, so the reader sees a real result in their own repository rather than a promise.",
		"Point sporo at a recipe, but never edit `internal/foo.go` by hand.", 1)
	if f := LintShape(SeedShape, "sporo.md", []byte(src), []string{"sporo"}); !hasFinding(f, "names a FILE") {
		t.Fatalf("an unrelated coordinate must still red; got: %v", f)
	}
}

// REQ-4 hyphen boundary: a lookalike that merely STARTS with the tool name (`sporo-helper/…`, the
// reader's own file) still reds — segment-stem equality, not containment.
func TestSeedHyphenLookalikeStillReds(t *testing.T) {
	src := strings.Replace(seedFixture(t, "sporo.md"),
		"Point sporo at a recipe and let it export the first artifact, so the reader sees a real result in their own repository rather than a promise.",
		"Never edit `sporo-helper/main.go`, the reader's own file.", 1)
	if f := LintShape(SeedShape, "sporo.md", []byte(src), []string{"sporo"}); !hasFinding(f, "names a FILE") {
		t.Fatalf("a lookalike that merely starts with the tool name must still red; got: %v", f)
	}
}

// An unrelated BARE (un-backticked) path — the reBarePath axis — still reds, exercising the
// eraser's not-owned branch (the path is kept, not erased).
func TestSeedUnrelatedBarePathReds(t *testing.T) {
	src := strings.Replace(seedFixture(t, "sporo.md"),
		"Point sporo at a recipe and let it export the first artifact, so the reader sees a real result in their own repository rather than a promise.",
		"Point sporo at a recipe, but leave vendor/lib/thing.go alone.", 1)
	if f := LintShape(SeedShape, "sporo.md", []byte(src), []string{"sporo"}); !hasFinding(f, "names a PATH") {
		t.Fatalf("an unrelated bare path must still red; got: %v", f)
	}
}

// The both-axes trap on ONE line: an exempt `.sporo/…` span and an unrelated coordinate share a
// line; the span is erased but the line still reds on the unrelated one — erasure beats first-match
// — and the finding quotes the ORIGINAL line (the erased span is visible in the message), proving
// clip(line), not clip(scanned).
func TestSeedMixedExemptAndUnrelatedOnSameLineStillReds(t *testing.T) {
	src := strings.Replace(seedFixture(t, "sporo.md"),
		"Point sporo at a recipe and let it export the first artifact, so the reader sees a real result in their own repository rather than a promise.",
		"Keep `.sporo/config.yaml` but never touch `internal/foo.go` on one line.", 1)
	f := LintShape(SeedShape, "sporo.md", []byte(src), []string{"sporo"})
	if !hasFinding(f, "names a FILE") {
		t.Fatalf("an unrelated coordinate sharing a line with an exempt one must still red; got: %v", f)
	}
	if !hasFinding(f, "Keep `.sporo/config.yaml`") {
		t.Fatalf("the finding must quote the ORIGINAL line including the erased span; got: %v", f)
	}
}

// N-3 surviving-span non-flip: the same bytes linted under the seed policy and under a nil policy
// must return the unrelated coordinate's finding BYTE-IDENTICAL (Line and Msg) — erasure suppresses
// only the exempt token and never relocates or rewords a co-located survivor. The nil run, which
// exempts nothing, additionally reds the sporo/.sporo tokens the seed policy erases, so it carries
// strictly more findings: proof the exemption is real and one-directional.
func TestSeedExemptionPreservesSurvivorFindingExactly(t *testing.T) {
	src := strings.Replace(seedFixture(t, "sporo.md"),
		"Point sporo at a recipe and let it export the first artifact, so the reader sees a real result in their own repository rather than a promise.",
		"Point sporo at a recipe to export the first artifact.\nThe tool writes to `.sporo/config.yaml` in your tree.\nNever edit `internal/foo.go` by hand.", 1)
	bytes := []byte(src)

	noPolicy := SeedShape
	noPolicy.Neutrality = nil

	seedF := LintShape(SeedShape, "sporo.md", bytes, []string{"sporo"})
	nilF := LintShape(noPolicy, "sporo.md", bytes, []string{"sporo"})

	if len(seedF) != 1 {
		t.Fatalf("the seed policy must leave exactly the one unrelated coordinate reding; got: %v", seedF)
	}
	survivor := seedF[0]
	if !strings.Contains(survivor.Msg, "names a FILE") || !strings.Contains(survivor.Msg, "internal/foo.go") {
		t.Fatalf("the survivor must be the internal/foo.go FILE finding quoting the ORIGINAL line; got: %v", survivor)
	}
	found := false
	for _, f := range nilF {
		if f == survivor {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("the survivor must be byte-identical under a nil policy; survivor=%v nil-run=%v", survivor, nilF)
	}
	if len(nilF) <= len(seedF) {
		t.Fatalf("a nil policy must red the exempt tokens the seed policy erases; nil=%v seed=%v", nilF, seedF)
	}
}

// N-2: a metacharacter-bearing target (`g++`) is regex-escaped, so the eraser compiles and exempts
// the tool's own `g++/driver.go` artifact without a panic.
func TestSeedMetacharTargetExemptsWithoutPanic(t *testing.T) {
	src := seedFixture(t, "metachar.md")
	if f := LintShape(SeedShape, "metachar.md", []byte(src), nil); len(f) != 0 {
		t.Fatalf("a metacharacter-bearing target must exempt its own artifact without panicking; got: %v", f)
	}
}

// A malformed (`@`-less) target yields no exemption: its own frontmatter check reds it, and the
// sporo tokens that a valid target would exempt now red like any document's — no crash.
func TestSeedMalformedTargetGivesNoExemption(t *testing.T) {
	src := strings.Replace(seedFixture(t, "sporo.md"), "target: sporo@v1.0.0", "target: sporo", 1)
	f := LintShape(SeedShape, "sporo.md", []byte(src), []string{"sporo"})
	if !hasFinding(f, "must be `<tool>@<version>`") {
		t.Fatalf("a malformed target must red the frontmatter check; got: %v", f)
	}
	if !hasFinding(f, "names a product") {
		t.Fatalf("with no exemption, the sporo tokens must red; got: %v", f)
	}
}

// The two nil-eraser branches directly: no target key, no `@`, and an empty tool left of the `@`
// all yield a nil (no-exemption) eraser.
func TestSeedNeutralityNilForMalformedTargets(t *testing.T) {
	for _, fm := range [][]string{
		{"name: x"},
		{"target: sporo"},
		{"target: @v1.0.0"},
		{`target: "@1.0"`},
	} {
		if seedNeutrality(fm) != nil {
			t.Fatalf("malformed target %v must yield a nil (no-exemption) eraser", fm)
		}
	}
}

// A `_`-prefixed name is the genre's own meta-document: it teaches the shape, it does not
// instantiate it, so the engine holds it to the line-1 banner alone. A body that would red under
// an instance name must lint green under a `_` name.
func TestSeedUnderscorePrefixIsExemptFromTheGenre(t *testing.T) {
	broken := strings.Replace(conformantSeed(t),
		"## Use\n\nPoint the tool at a build description and let it produce the first artifact, so the reader sees a real result in their own repository rather than a promise.\n\n",
		"", 1)
	if f := lintSeed("_seed.md", broken); len(f) != 0 {
		t.Fatalf("a `_`-prefixed meta-document is held to the banner alone; got: %v", f)
	}
}
