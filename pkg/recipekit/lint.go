// Package recipekit is the pure, dependency-free core of sporo's recipe logic — the
// genre gate, the content-hash and semver currency, the frontmatter reader, the
// report-back validator — extracted so a future registry server can import exactly the
// same rules the CLI enforces. Everything here is a pure function of bytes: no filesystem,
// no network, no config. `internal/recipe` keeps its full API and forwards to this package,
// so nothing outside changes.
//
// recipekit MUST NOT import internal/recipe — every function is bytes → result, which is
// what keeps that true and the seam clean.
package recipekit

import (
	"fmt"
	"regexp"
	"strings"
)

// Finding is one violation, located. Line is 1-based; 0 means the finding is about the file
// as a whole (a missing section has no line to point at).
type Finding struct {
	File string
	Line int
	Msg  string
}

func (f Finding) String() string {
	if f.Line > 0 {
		return fmt.Sprintf("%s:%d: %s", f.File, f.Line, f.Msg)
	}
	return fmt.Sprintf("%s: %s", f.File, f.Msg)
}

// requiredSections is the genre's shape (the authoring spec, §1), in order. A genre defined
// only by taste drifts into whatever the last author felt like writing, so the shape is
// checked rather than trusted. `## Appendix` is deliberately absent: it is optional, and it
// is the one section where instances are allowed.
var requiredSections = []string{
	"## The problem",
	"## Why the obvious approach fails",
	"## The principles",
	"## The ground it needs",
	"## The contracts",
	"## The build sequence",
	"## The seams",
	"## The scars",
	"## Verification",
	"## The trade-offs",
	"## For the human",
}

// requiredKeys is the frontmatter. `stack` and `verified` are honesty stamps, not metadata:
// one says what the build ran on, the other says it is a snapshot of a build that actually
// happened. Both are the reader's only defence against a recipe written from an intention.
// `version` is the loop's anchor: a report-back that cannot say WHICH text its author built
// is unusable the day the recipe changes, and the exported file is the only thing the reader
// has — so the version travels in the document, not in a registry the reader never sees.
var requiredKeys = []string{"id", "name", "version", "title", "problem", "prerequisites", "derived_from", "stack", "verified", "effort"}

// The coordinate vocabulary — the shapes, never a snapshot of today's names, or the scan
// goes blind the day someone adds a directory.
var (
	reFilename = regexp.MustCompile("`[^`]*\\.(json|yaml|yml|md|go|sh|py|ts|tsx|js|css|html|toml|rs|rb|java|sql)`")
	rePathSeg  = regexp.MustCompile("`[^`]*/[^`]*`")
	// A path does not need backticks to be a coordinate. The bare pattern is deliberately
	// conservative — two or more directory segments AND a file extension, with a dot-free
	// first segment — so that prose fractions ("24/7"), alternations ("and/or"), initialisms
	// ("TCP/IP") and bare domains ("docs.example.com/…", dotted first segment) stay green.
	// URLs are erased from the line before any scanning (a link is a reference, not a
	// coordinate in the reader's tree). The cost of the conservatism is a known gap — a
	// one-segment bare path ("src/main.c") passes — and the semantic review's neutrality
	// axis is the declared counterweight, not a wider regex that reds on prose.
	reBarePath = regexp.MustCompile(`(?:^|[^\w./])((?:[\w-]+/){2,}[\w-]+\.[A-Za-z]{1,5})\b`)
	reURL      = regexp.MustCompile(`https?://\S+`)
	// Binding is the strictness of one shown shape, and it is two-valued on purpose:
	// `exact` — a consumer OUTSIDE the emitting repository reads this shape (a fleet's
	// collector, another team's tool); it is copied byte-for-byte, and changing it later is
	// a MAJOR version. `adapt` — shown so the reader does not re-invent it; local
	// conventions win. A shape with no binding leaves the reader guessing which kind they
	// are holding, which is how one team's "example" becomes another team's broken feed.
	//
	// reFence and reBinding are ALSO declared in internal/recipe (conform.go still reads the
	// same markers) — the two copies must stay byte-identical, since seal and lint must never
	// disagree about which fence is which.
	reBinding    = regexp.MustCompile(`\*\*Binding: (exact|adapt)\*\*`)
	reBindingAny = regexp.MustCompile(`\*\*Binding:`)
	reHeading    = regexp.MustCompile(`^### `)
	reFence      = regexp.MustCompile("^\\s*```")
	reFixture    = regexp.MustCompile(`\*\*Fixture: (valid|invalid)\*\*`)
	reDone       = regexp.MustCompile(`\*\*Done when:\*\*`)
	reSemver     = regexp.MustCompile(`^version:\s*"?\d+\.\d+\.\d+"?\s*$`)
	// A ULID: 26 chars of Crockford base32 (no I/L/O/U). The first char is [0-7] because a
	// ULID is 130 bits (26×5) and the 48-bit timestamp caps the leading char at 7 — anything
	// higher is an over-long value that only LOOKS like a ULID. Minted by NewID(), never typed.
	reULID  = regexp.MustCompile(`^id:\s*"?[0-7][0-9A-HJKMNP-TV-Z]{25}"?\s*$`)
	reAllow = "<!-- recipe-lint: allow" // line-precise opt-out; must carry a reason
)

// scarMarkers: a scar missing one of the three teaches nothing — it decays into a paragraph
// of regret that an agent cannot lift out whole.
var scarMarkers = []string{"Symptom", "Root cause", "Fix"}

// Lint checks one recipe against the genre. `products` is the forbidden product vocabulary —
// the names that mean something only inside one repository (the project's own name, and any
// sibling it can see). Findings are returned, never printed: the caller owns the output, and
// the tests own the assertions.
//
// `extra` is the injection seam for the conform layer's lint half (fixtureFindings), which
// lives in internal/recipe because it parses JSON/YAML — machinery this pure package does not
// import. It runs at the ORIGINAL call site (after the contract checks, before neutrality, and
// after the `_`-prefix early return below), so finding order and the meta-document exemption
// are preserved exactly. Do not "simplify" this into an append after Lint returns.
//
// A `_`-prefixed name is a corpus document, not a recipe: one TEACHES the shape, the other
// is appended to every export. Neither instantiates the genre, so both are held to the banner
// alone — demanding the eleven sections of the spec would be the gate testing its own habit.
func Lint(name string, src []byte, products []string, extra ...func(name string, src []byte) []Finding) []Finding {
	var out []Finding
	fail := func(line int, format string, a ...any) {
		out = append(out, Finding{File: name, Line: line, Msg: fmt.Sprintf(format, a...)})
	}
	lines := strings.Split(string(src), "\n")

	if len(lines) == 0 || !strings.Contains(lines[0], "<!-- SSOT SOURCE") {
		fail(1, "missing provenance banner on line 1 (`<!-- SSOT SOURCE… -->`) — the marker `recipe export` strips on the way out")
	}
	if strings.HasPrefix(name, "_") {
		return out
	}

	// Frontmatter. Line 1 is the banner, so the fences are the first two `---` after it.
	fmStart, fmEnd := -1, -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			if fmStart < 0 {
				fmStart = i
			} else {
				fmEnd = i
				break
			}
		}
	}
	if fmStart < 0 || fmEnd < 0 {
		fail(0, "missing frontmatter (--- … ---) with %s", strings.Join(requiredKeys, "/"))
	} else {
		fm := lines[fmStart+1 : fmEnd]
		for _, key := range requiredKeys {
			if !hasKey(fm, key) {
				fail(0, "frontmatter is missing `%s:`", key)
			}
		}
		if v := KeyLine(fm, "id"); v != "" && !reULID.MatchString(v) {
			fail(0, "`id:` must be a ULID (26 Crockford-base32 chars) — it is the recipe's permanent identity, minted by `sporo new`, never typed or edited; a hand-written id is how two recipes end up claiming the same permalink")
		}
		if v := KeyLine(fm, "version"); v != "" && !reSemver.MatchString(v) {
			fail(0, "`version:` must be a semver triple (MAJOR.MINOR.PATCH) — the report-back channel binds to it, and a version that cannot be ordered cannot say which text superseded which")
		}
		if v := KeyLine(fm, "verified"); v != "" && !strings.Contains(v, "project") {
			fail(0, "`verified:` must name the build that proves this recipe (project, release, date) — a recipe written from an intention teaches untested guesses as if they were earned")
		}
		if v := KeyLine(fm, "stack"); v != "" && !strings.Contains(v, "language") {
			fail(0, "`stack:` must name what the build actually ran on (language, runtime, why) — the reader cannot weigh the recommendation without it")
		}
	}

	// The shape — present AND in order. Order is checked, not merely claimed: the sequence is
	// the argument (why you need it → what it rests on → how to build it → what it cost), and
	// a recipe that opens with its build steps has quietly become a checklist. Checking only
	// presence would let the document say anything in any order while the gate reported the
	// genre as enforced.
	prev, prevAt := "", -1
	for _, want := range requiredSections {
		at := sectionAt(lines, want)
		if at < 0 {
			fail(0, "missing section: %s", strings.TrimPrefix(want, "## "))
			continue
		}
		if prevAt >= 0 && at < prevAt {
			fail(at+1, "section out of order: %q must come after %q — the sequence IS the argument",
				strings.TrimPrefix(want, "## "), strings.TrimPrefix(prev, "## "))
		}
		prev, prevAt = want, at
	}

	// Every build step carries its acceptance. Counted, not assumed: "each step says how you
	// know it is done" is otherwise a promise nobody checks.
	seq := SectionBody(lines, "## The build sequence")
	steps, dones := count(seq, reHeading), count(seq, reDone)
	switch {
	case steps == 0:
		fail(0, "the build sequence has no steps (`### ` headings)")
	case steps != dones:
		fail(0, "%d build step(s) but %d `**Done when:**` line(s) — a step with no acceptance is a wish, and the reader finds out at the end", steps, dones)
	}

	// The contracts must be SHOWN, not described. A shape rendered as prose ("a slug, a title,
	// phases done out of total") reads as complete to its author and forces every reader to
	// re-derive it — differently, in an incompatible form, and the interoperation the contract
	// existed for is gone. The first implementer of this genre's own first recipe scored it 3/10
	// on copy-paste artifacts for exactly this: not one example of any shape it told him to
	// consume.
	//
	// A CONTRACT is not a COORDINATE, and this is the distinction that took a live build to see.
	// A path, a filename, a product is a coordinate: it EXECUTES in one repository and transfers
	// to none. A schema, a normalized shape, a declared surface is a contract: it is a shape the
	// reader COPIES AND ADAPTS, and it transfers exactly as well as the principle it serves.
	// Banning both is how a genre defends its neutrality into uselessness.
	//
	// Note what is deliberately NOT here: a fenced-block exemption from the coordinate scan. It
	// would be a hole (a real path, hidden in a fence, passes everywhere) — and it is not needed:
	// the coordinate patterns are backtick-anchored, and a fenced block has no backticks, so a
	// schema example already passes. Verified before this rule was written, which is why the
	// rule is purely additive. Products stay banned inside fences like everywhere else: a
	// coordinate that leaks into an example is still a coordinate.
	con := SectionBody(lines, "## The contracts")
	if count(con, reFence) == 0 {
		fail(0, "the contracts section shows no shape (no fenced block) — a schema described in prose "+
			"is one every reader has to re-invent, incompatibly; show the shape, and the reader copies it")
	}
	// Every shown shape declares its binding BEFORE the fence opens. The check is
	// positional, not a count: two markers stacked on one shape and none on the next would
	// pass a count while the unlabelled shape ships.
	pending, inFence := 0, false
	for _, l := range con {
		switch {
		case reFence.MatchString(l):
			if !inFence {
				if pending == 0 {
					fail(0, "a contract fence carries no `**Binding:**` marker — the reader cannot tell an interoperability boundary from a sketch; write `**Binding: exact**` (a consumer outside the repository reads this shape — changing it is a MAJOR version) or `**Binding: adapt**` (shown so the reader does not re-invent it) above the fence")
				}
				pending, inFence = 0, true
			} else {
				inFence = false
			}
		case !inFence && reBindingAny.MatchString(l):
			if !reBinding.MatchString(l) {
				fail(0, "`**Binding:**` must say `exact` or `adapt` — nothing in between: a shape either has an outside consumer or it does not, and a third word would let an author avoid deciding")
			} else {
				pending++
			}
		case !inFence && reFixture.MatchString(l):
			// A fixture fence is labelled too — by its own marker. It belongs to the exact
			// shape above it, so demanding a second Binding on it would be the gate
			// misreading the layer it guards.
			pending++
		}
	}

	// Every scar is symptom → root cause → fix.
	scarBody := SectionBody(lines, "## The scars")
	scars := count(scarBody, reHeading)
	if scars == 0 {
		fail(0, "no scars (`### ` headings under The scars) — a build with nothing to warn about did not need a recipe")
	}
	for _, marker := range scarMarkers {
		re := regexp.MustCompile(`\*\*` + regexp.QuoteMeta(marker) + `:\*\*`)
		if n := count(scarBody, re); n != scars {
			fail(0, "%d scar(s) but %d `**%s:**` marker(s) — a scar missing that half teaches nothing", scars, n, marker)
		}
	}

	// The conform layer's lint half — injected via `extra`, at this exact position (see the
	// doc comment above). It fires ONLY on recipes that declared an exact contract (ADR-005):
	// the shape must parse, valid fixtures must conform, invalid fixtures must not.
	for _, ex := range extra {
		out = append(out, ex(name, src)...)
	}

	out = append(out, neutrality(name, lines, fmEnd, products)...)
	return out
}

// neutrality is the constraint the genre exists for. The ban is on COORDINATES — a path, a
// filename, a product — never on concreteness: a technology named as a choice with a reason
// ("a single statically-linked binary, because the reader runs it with no checkout of the
// source") is portable, and an agent on another stack can weigh it. A coordinate executes in
// one repository and transfers to none.
//
// Scanned over the BODY only. The frontmatter names the project on purpose (that is
// provenance, not instruction) and the appendix is the one section where instances are
// allowed and are explicitly marked as illustration. There is no other exempt zone — in
// particular the stack sections get no licence, because a technology was never a coordinate.
func neutrality(name string, lines []string, fmEnd int, products []string) []Finding {
	var out []Finding
	var reProducts *regexp.Regexp
	if len(products) > 0 {
		quoted := make([]string, 0, len(products))
		for _, p := range products {
			if p = strings.TrimSpace(p); p != "" {
				quoted = append(quoted, regexp.QuoteMeta(p))
			}
		}
		if len(quoted) > 0 {
			reProducts = regexp.MustCompile(`(?i)\b(` + strings.Join(quoted, "|") + `)\b`)
		}
	}

	// The scan window runs from the end of the frontmatter to the appendix — or, with no
	// appendix, PAST the last line. The bound is exclusive, and an inclusive one here is the
	// classic off-by-one: a violation on the final line falls outside the window, which is
	// exactly where a hurried author appends one.
	start := fmEnd + 1
	end := len(lines)
	for i, l := range lines {
		if strings.HasPrefix(l, "## Appendix") {
			end = i
			break
		}
	}
	for i := start; i < end; i++ {
		line := lines[i]
		if strings.Contains(line, reAllow) {
			continue
		}
		scanned := reURL.ReplaceAllString(line, "")
		n := i + 1
		switch {
		case reProducts != nil && reProducts.MatchString(scanned):
			out = append(out, Finding{name, n, "names a product — a reader in another repository cannot follow it: " + clip(line)})
		case reFilename.MatchString(scanned):
			out = append(out, Finding{name, n, "names a FILE — recipes name roles (\"the facts file\"), not instances: " + clip(line)})
		case rePathSeg.MatchString(scanned):
			out = append(out, Finding{name, n, "names a PATH — it means nothing outside the repository it was written in: " + clip(line)})
		case reBarePath.MatchString(scanned):
			out = append(out, Finding{name, n, "names a PATH (backticks or not) — it means nothing outside the repository it was written in: " + clip(line)})
		}
	}
	return out
}

// IsDraft reports a recipe that declares `draft: true` in its frontmatter — the state
// `sporo new` scaffolds in. A draft is exempt from the genre gate (a red gate on the state
// the tool itself writes trains red-blindness — the cry-wolf failure) and, for exactly the
// same reason, a draft can neither be sealed nor exported: the exemption and the export ban
// are one rule seen from two sides. What never leaves the house does not have to be finished;
// what leaves the house must be.
func IsDraft(src []byte) bool {
	return FrontmatterValue(src, "draft") == "true"
}

func hasKey(fm []string, key string) bool { return KeyLine(fm, key) != "" }

// sectionAt returns the 0-based line of a `## ` heading, or -1.
func sectionAt(lines []string, want string) int {
	for i, l := range lines {
		if strings.HasPrefix(l, want) {
			return i
		}
	}
	return -1
}

// KeyLine returns the frontmatter line that declares `key:`, or "".
func KeyLine(fm []string, key string) string {
	for _, l := range fm {
		if strings.HasPrefix(l, key+":") {
			return l
		}
	}
	return ""
}

// SectionBody returns the lines under a `## ` heading, up to the next `## ` heading.
func SectionBody(lines []string, heading string) []string {
	var out []string
	in := false
	for _, l := range lines {
		if strings.HasPrefix(l, heading) {
			in = true
			continue
		}
		if in && strings.HasPrefix(l, "## ") {
			break
		}
		if in {
			out = append(out, l)
		}
	}
	return out
}

func count(lines []string, re *regexp.Regexp) int {
	n := 0
	for _, l := range lines {
		if re.MatchString(l) {
			n++
		}
	}
	return n
}

func clip(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 60 {
		return s[:60]
	}
	return s
}
