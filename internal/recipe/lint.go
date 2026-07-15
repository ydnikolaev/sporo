package recipe

// The genre gate, in the binary — deliberately, not as a script in the harness repo.
//
// A recipe is written ABOUT a repository, most often the one the author is standing in, and
// almost never this one. A gate that only the harness can run is therefore a gate nobody
// runs at the moment it matters: the consumer writes a recipe about its own service, names
// its own paths on every second line, and nothing says a word — the document reads perfectly
// to the author, who knows what those paths mean, and is unfollowable for its actual reader,
// who is an agent in a repository that has never seen them. That asymmetry (the reader who
// cannot check the document is not in the room) is the same one the reporting doctrine
// answers with a gate rather than a review note. So the check ships where the recipe is
// written: any consumer with the CLI can run it against its own corpus.
//
// One consequence worth naming: the forbidden PRODUCT vocabulary cannot be a constant here.
// A binary that bans this fleet's names and nothing else is blind to the one name most
// likely to leak — the reader's own project. The product list is a project VALUE and comes
// from the config seam; the principle ("a name that only means something inside one
// repository may not appear in the body") is what lives in the code.

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
var requiredKeys = []string{"name", "title", "problem", "prerequisites", "derived_from", "stack", "verified", "effort"}

// The coordinate vocabulary — the shapes, never a snapshot of today's names, or the scan
// goes blind the day someone adds a directory.
var (
	reFilename = regexp.MustCompile("`[^`]*\\.(json|yaml|yml|md|go|sh|py|ts|tsx|js|css|html|toml|rs|rb|java|sql)`")
	rePathSeg  = regexp.MustCompile("`[^`]*/[^`]*`")
	reHeading  = regexp.MustCompile(`^### `)
	reFence    = regexp.MustCompile("^\\s*```")
	reDone     = regexp.MustCompile(`\*\*Done when:\*\*`)
	reAllow    = "<!-- recipe-lint: allow" // line-precise opt-out; must carry a reason
)

// scarMarkers: a scar missing one of the three teaches nothing — it decays into a paragraph
// of regret that an agent cannot lift out whole.
var scarMarkers = []string{"Symptom", "Root cause", "Fix"}

// Lint checks one recipe against the genre. `products` is the forbidden product vocabulary —
// the names that mean something only inside one repository (the project's own name, and any
// sibling it can see). Findings are returned, never printed: the caller owns the output, and
// the tests own the assertions.
//
// A `_`-prefixed name is a corpus document, not a recipe: one TEACHES the shape, the other
// is appended to every export. Neither instantiates the genre, so both are held to the banner
// alone — demanding the eleven sections of the spec would be the gate testing its own habit.
func Lint(name string, src []byte, products []string) []Finding {
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
		if v := keyLine(fm, "verified"); v != "" && !strings.Contains(v, "project") {
			fail(0, "`verified:` must name the build that proves this recipe (project, release, date) — a recipe written from an intention teaches untested guesses as if they were earned")
		}
		if v := keyLine(fm, "stack"); v != "" && !strings.Contains(v, "language") {
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
	seq := sectionBody(lines, "## The build sequence")
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
	con := sectionBody(lines, "## The contracts")
	if count(con, reFence) == 0 {
		fail(0, "the contracts section shows no shape (no fenced block) — a schema described in prose "+
			"is one every reader has to re-invent, incompatibly; show the shape, and the reader copies it")
	}

	// Every scar is symptom → root cause → fix.
	scarBody := sectionBody(lines, "## The scars")
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
		n := i + 1
		switch {
		case reProducts != nil && reProducts.MatchString(line):
			out = append(out, Finding{name, n, "names a product — a reader in another repository cannot follow it: " + clip(line)})
		case reFilename.MatchString(line):
			out = append(out, Finding{name, n, "names a FILE — recipes name roles (\"the facts file\"), not instances: " + clip(line)})
		case rePathSeg.MatchString(line):
			out = append(out, Finding{name, n, "names a PATH — it means nothing outside the repository it was written in: " + clip(line)})
		}
	}
	return out
}

func hasKey(fm []string, key string) bool { return keyLine(fm, key) != "" }

// sectionAt returns the 0-based line of a `## ` heading, or -1.
func sectionAt(lines []string, want string) int {
	for i, l := range lines {
		if strings.HasPrefix(l, want) {
			return i
		}
	}
	return -1
}

func keyLine(fm []string, key string) string {
	for _, l := range fm {
		if strings.HasPrefix(l, key+":") {
			return l
		}
	}
	return ""
}

// sectionBody returns the lines under a `## ` heading, up to the next `## ` heading.
func sectionBody(lines []string, heading string) []string {
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
