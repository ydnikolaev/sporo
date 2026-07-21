package recipekit

import (
	"regexp"
	"strings"
)

// The seed genre's own format vocabulary — declared HERE, never in lint.go, so the recipe file
// stays untouched (INV-1). reSemver/reULID/reDone/reFence/reHeading/reHTMLComment and
// summaryMinRunes are reused from the recipe genre: a seed's honesty stamps, per-step
// acceptance, and summary floor are the same shapes, and a second copy would let them drift.
var (
	// target is the seed's subject, pinned as `<tool>@<version>`. It is provenance, named once
	// (authoring §1a), not the reader's coordinate — a seed proven against one release is not a
	// promise about the next, and the pin is how the reader knows which promise they hold.
	reTarget = regexp.MustCompile(`^target:\s*"?[^\s@]+@[^\s@]+"?\s*$`)
	// The two seed markers. `**Detect:**` opens the first Install step so the seed is idempotent;
	// a fetch piped into a shell is the most dangerous move a seed makes, so it must cite the
	// declared source origin (authoring §2). Their spelling is pinned in seeds/_authoring.md.
	reDetect    = regexp.MustCompile(`\*\*Detect:\*\*`)
	reBlindPipe = regexp.MustCompile(`(?:curl|wget|fetch)\b.*\|\s*(?:sudo\s+)?(?:sh|bash|zsh|dash)\b`)
)

// seedReportRows are the five fixed Report rows, in order (authoring §4). The human reviews many
// seed runs and a report whose shape shifts is one they must re-read from scratch, so an extra
// row, a missing row, and a reorder each red. Pinned here and in seeds/_authoring.md.
var seedReportRows = []string{
	"what it is",
	"how it works",
	"what was done",
	"how to use it",
	"suggest next",
}

// seedSections is the seven-section body, in order (authoring §1). The sequence is the argument:
// you cannot use a tool you have not verified, verify one you have not installed, install one you
// do not understand, or describe one before orienting the reader to why they want it.
var seedSections = []string{
	"## Summary",
	"## What it is",
	"## Install",
	"## Verify",
	"## Use",
	"## Harness",
	"## Report",
}

// seedKeys is the nine frontmatter keys (authoring §1a). Missing-key findings come from the
// engine's required-key scan; seedFrontmatterChecks is only the format layer over present keys.
var seedKeys = []string{"id", "name", "version", "title", "target", "source", "stack", "verified", "effort"}

// SeedShape is the seed genre as data, registered through the S1 seam the same way RecipeShape is
// — a package-var initializer, because gochecknoinits forbids init(). It shares the generic
// engine (LintShape) with the recipe genre and adds only the two seed-specific hooks.
var SeedShape = RegisterShape(Shape{
	Kind:              KindSeed,
	Sections:          seedSections,
	Keys:              seedKeys,
	FrontmatterChecks: seedFrontmatterChecks,
	BodyChecks:        seedBodyChecks,
})

// seedFrontmatterChecks is the seed genre's format layer over the required keys. Each check fires
// only when the key is present — a missing key is the engine's required-key finding, not this
// one's. id/version/verified/stack mirror the recipe stamps; target and source are the seed's
// trust anchors: which tool, pinned, and the origin every Install step must trace back to.
func seedFrontmatterChecks(fm []string, fail FailFunc) {
	if v := KeyLine(fm, "id"); v != "" && !reULID.MatchString(v) {
		fail(0, "`id:` must be a ULID (26 Crockford-base32 chars) — the seed's permanent identity, minted by the tool and never typed; a report-back and a permalink both hang on it")
	}
	if v := KeyLine(fm, "version"); v != "" && !reSemver.MatchString(v) {
		fail(0, "`version:` must be a semver triple (MAJOR.MINOR.PATCH) — the seed's own version travels in the document, and a report-back binds to it")
	}
	if v := KeyLine(fm, "verified"); v != "" && !strings.Contains(v, "project") {
		fail(0, "`verified:` must name the install that proves this seed (project, release, date) — a seed written from memory hands the reader an untested guess")
	}
	if v := KeyLine(fm, "stack"); v != "" && !strings.Contains(v, "language") {
		fail(0, "`stack:` must name what the verifying install actually ran on (language, runtime) — the reader cannot weigh how far their own ground is from it without the stamp")
	}
	if v := KeyLine(fm, "target"); v != "" && !reTarget.MatchString(v) {
		fail(0, "`target:` must be `<tool>@<version>` — the seed's named, pinned subject; a version that is not pinned cannot say which promise the reader is holding")
	}
	if v := KeyLine(fm, "source"); v != "" && fmValue(fm, "source") == "" {
		fail(0, "`source:` must name a canonical origin the reader can inspect — it is the trust anchor under every Install step, and a blank origin cannot be audited")
	}
}

// seedBodyChecks is the seed genre's body layer, run in a fixed phase order so findings are
// deterministic: the summary floor, the Install trust contract (detect-first + per-step
// acceptance + no uncited pipe), the Verify proof, and the fixed-shape Report. `## Harness` is
// deliberately absent — the engine checks its presence, and the genre never judges its verdict
// (authoring §5).
func seedBodyChecks(name string, lines []string, fail FailFunc) {
	fm := frontmatterOf(lines)

	// 1. Summary floor — orientation before the moves. HTML coach comments do not count: only
	// what the exported reader sees earns the floor.
	summary := strings.Join(SectionBody(lines, "## Summary"), "\n")
	summary = reHTMLComment.ReplaceAllString(summary, "")
	if n := len([]rune(strings.TrimSpace(summary))); n < summaryMinRunes {
		fail(0, "Summary is %d character(s); write a 2–4 sentence orientation of at least %d characters — a label is not an argument", n, summaryMinRunes)
	}

	// 2. Install — detect first (idempotency), and every step carries its acceptance.
	install := SectionBody(lines, "## Install")
	steps := count(install, reHeading)
	if steps == 0 {
		fail(0, "`## Install` has no steps (`### ` headings) — a seed with nothing to acquire is a recipe wearing install commands")
	} else {
		if first := firstStep(install); !hasMatch(first, reDetect) {
			fail(0, "`## Install`'s first `### ` step must open with `**Detect:**` — an install that does not detect first clobbers a working tree; the marker is what makes the seed idempotent")
		}
		if dones := count(install, reDone); steps != dones {
			fail(0, "%d install step(s) but %d `**Done when:**` line(s) — a step with no acceptance is a wish the agent cannot tell is failing until the end", steps, dones)
		}
	}

	// 3. No blind fetch-into-a-shell unless the step cites the declared source origin. When the
	// origin is blank the frontmatter finding already owns it, so this phase stays silent then.
	if srcVal := fmValue(fm, "source"); srcVal != "" {
		for _, step := range splitSteps(install) {
			if hasMatch(step, reBlindPipe) && !strings.Contains(strings.Join(step, "\n"), srcVal) {
				fail(0, "an `## Install` step pipes a fetched script into a shell without citing the `source` origin — remote code run with the reader's privileges must trace to the origin the frontmatter vouches for, or not pipe at all")
			}
		}
	}

	// 4. Verify holds at least one runnable proof — a fenced command block, not prose.
	if count(SectionBody(lines, "## Verify"), reFence) < 2 {
		fail(0, "`## Verify` contains no fenced command block — prose that asserts the tool runs proves nothing; show a command the agent runs and reads")
	}

	// 5. Report is exactly the five fixed rows, in order.
	seedReportCheck(SectionBody(lines, "## Report"), fail)
}

// seedReportCheck holds `## Report` to exactly the five fixed rows, in order (authoring §4): a
// wrong data-row count, a missing row, an extra row, and a reorder each red distinctly, because a
// report that drifts from the five is one the human can no longer read at a glance. The
// cardinality assert is the AUD-001 ride: the set-based missing/unexpected loops both pass a
// duplicated known label, and the reorder check is skipped on a length mismatch, so a Report that
// repeats or pads a known row would otherwise slip through — the count guard is what catches it.
func seedReportCheck(report []string, fail FailFunc) {
	got := reportLabels(report)
	want := seedReportRows

	if len(got) != len(want) {
		fail(0, "`## Report` has %d data row(s), not the %d fixed rows — a duplicated or padded row breaks the five-row shape the human reads the same way every time", len(got), len(want))
	}
	for _, w := range want {
		if !contains(got, w) {
			fail(0, "`## Report` is missing the %q row — the human's audit is a fixed five-row table, and an omission breaks the shape they read the same way every time", w)
		}
	}
	for _, g := range got {
		if !contains(want, g) {
			fail(0, "`## Report` has an unexpected %q row — the audit is exactly five fixed rows; an extra row is one the human did not expect to read", g)
		}
	}
	if sameSet(got, want) && !equalSeq(got, want) {
		fail(0, "`## Report` rows are out of order — they read what it is / how it works / what was done / how to use it / suggest next, the order that walks the human from understanding through the audit to what next")
	}
}

// reportLabels extracts the lowercased label column from a Report table's data rows — those after
// the `|---|` separator, so the header row is never mistaken for data.
func reportLabels(report []string) []string {
	var labels []string
	sawSeparator := false
	for _, l := range report {
		t := strings.TrimSpace(l)
		if !strings.HasPrefix(t, "|") {
			continue
		}
		if isTableSeparator(t) {
			sawSeparator = true
			continue
		}
		if !sawSeparator {
			continue
		}
		cells := strings.Split(strings.Trim(t, "|"), "|")
		labels = append(labels, strings.ToLower(strings.Trim(strings.TrimSpace(cells[0]), "*")))
	}
	return labels
}

// isTableSeparator reports whether a table row is the `|---|` rule (only pipes, dashes, colons,
// whitespace, and at least one dash), which sits between the header and the data rows.
func isTableSeparator(row string) bool {
	dashed := false
	for _, r := range row {
		switch r {
		case '|', '-', ':', ' ', '\t':
		default:
			return false
		}
		if r == '-' {
			dashed = true
		}
	}
	return dashed
}

// frontmatterOf returns the lines between the two `---` fences after the line-1 banner, or nil.
func frontmatterOf(lines []string) []string {
	start, end := -1, -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			if start < 0 {
				start = i
			} else {
				end = i
				break
			}
		}
	}
	if start < 0 || end < 0 {
		return nil
	}
	return lines[start+1 : end]
}

// fmValue returns the trimmed value of a frontmatter `key:` line, or "".
func fmValue(fm []string, key string) string {
	l := KeyLine(fm, key)
	if l == "" {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(l, key+":"))
}

// firstStep returns the first `### ` step block of a section body, or nil.
func firstStep(body []string) []string {
	steps := splitSteps(body)
	if len(steps) == 0 {
		return nil
	}
	return steps[0]
}

// splitSteps breaks a section body into `### `-delimited step blocks, dropping any preamble
// before the first step.
func splitSteps(body []string) [][]string {
	var steps [][]string
	cur := -1
	for _, l := range body {
		if reHeading.MatchString(l) {
			steps = append(steps, []string{l})
			cur = len(steps) - 1
		} else if cur >= 0 {
			steps[cur] = append(steps[cur], l)
		}
	}
	return steps
}

func hasMatch(lines []string, re *regexp.Regexp) bool {
	for _, l := range lines {
		if re.MatchString(l) {
			return true
		}
	}
	return false
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, x := range a {
		if !contains(b, x) {
			return false
		}
	}
	return true
}

func equalSeq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
