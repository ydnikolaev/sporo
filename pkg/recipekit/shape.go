package recipekit

import (
	"fmt"
	"strings"
)

// FailFunc records one finding at a line (0 = whole-file). It is the writer half of the
// engine handed to a shape's genre-specific hooks, so a hook appends findings without
// owning the accumulator or the file name.
type FailFunc func(line int, format string, a ...any)

// Shape is an entity genre as data: the ordered required sections and frontmatter keys the
// engine checks for every kind, plus the genre-specific hooks it cannot know generically.
// FrontmatterChecks runs inside the frontmatter-exists branch, after the required-key scan;
// BodyChecks runs after the section presence+order scan. Splitting the extras into a
// frontmatter hook and a body hook is load-bearing: a single lumped hook would move a genre's
// frontmatter-format findings from before the section scan to after it.
//
// Neutrality is the genre's per-instance neutrality policy, also as data: given the parsed
// frontmatter it returns a per-line eraser the engine applies to each scanned line before the
// coordinate scan (or nil for no policy). It mirrors FrontmatterChecks/BodyChecks — the engine
// consults the field uniformly and never branches on Kind. A genre that leaves it nil (the
// recipe) gets today's neutrality exactly: the eraser is the identity, no token is exempt.
type Shape struct {
	Kind              string
	Sections          []string
	Keys              []string
	FrontmatterChecks func(fm []string, fail FailFunc)
	BodyChecks        func(name string, lines []string, fail FailFunc)
	Neutrality        func(fm []string) func(string) string
}

// The closed kind vocabulary — the single SSOT for what an entity can be. It lives in this
// pure package so internal/recipe (which imports recipekit, never the reverse) shares it.
// `seed` is forward-declared here as a registry-ledger member; its lintable Shape arrives in
// S2.
const (
	KindRecipe = "recipe"
	KindSeed   = "seed"
)

var Kinds = []string{KindRecipe, KindSeed}

// ValidKind reports whether k is a member of the closed vocabulary.
func ValidKind(k string) bool {
	for _, v := range Kinds {
		if v == k {
			return true
		}
	}
	return false
}

// shapes is the genre registry — the seam a new kind is added through without touching the
// engine. RecipeShape registers via its package-var initializer; S2's seed shape registers
// the same way. The map has no initializer dependencies, so it is ready before any package
// var that calls RegisterShape.
var shapes = map[string]Shape{}

// RegisterShape adds a shape to the registry and returns it, so a package var can register in
// its own initializer (`var RecipeShape = RegisterShape(Shape{…})`) without an init function.
// An unknown or duplicate kind is a programming error at package load, not a runtime finding,
// so it panics.
func RegisterShape(s Shape) Shape {
	if !ValidKind(s.Kind) {
		panic("recipekit: RegisterShape unknown kind " + s.Kind)
	}
	if _, ok := shapes[s.Kind]; ok {
		panic("recipekit: RegisterShape duplicate kind " + s.Kind)
	}
	shapes[s.Kind] = s
	return s
}

// ShapeFor returns the registered shape for a kind.
func ShapeFor(kind string) (Shape, bool) {
	s, ok := shapes[kind]
	return s, ok
}

// LintShape is the generic genre engine: it checks one document against a shape and returns
// the findings, never printing (the caller owns output, the tests own the assertions). The
// phases run in a fixed order — banner, then the `_`-prefix early return, then frontmatter,
// then the shape's frontmatter hook, then sections, then the shape's body hook, then the
// injected extras, then neutrality — because the sequence IS the finding order every consumer
// of the corpus already depends on.
//
// `extra` is the injection seam for the conform layer's lint half (fixtureFindings), which
// lives in internal/recipe because it parses JSON/YAML — machinery this pure package does not
// import. It runs at the position below (after the body checks, before neutrality, and after
// the `_`-prefix early return), so finding order and the meta-document exemption are preserved
// exactly. Do not "simplify" this into an append after LintShape returns.
//
// A `_`-prefixed name is a corpus document, not an instance: one TEACHES the shape, the other
// is appended to every export. Neither instantiates the genre, so both are held to the banner
// alone — demanding the body sections of a document that teaches them would be the gate
// testing its own habit.
func LintShape(shape Shape, name string, src []byte, products []string, extra ...func(name string, src []byte) []Finding) []Finding {
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
	// The neutrality eraser is computed once, here, from the frontmatter the engine already
	// parses — no second parse — and threaded into neutrality below. It stays nil when the shape
	// declares no policy or there is no frontmatter to derive from, and a nil eraser is the
	// identity (see neutrality), so the recipe path is a structural no-op.
	var erase func(string) string
	if fmStart < 0 || fmEnd < 0 {
		fail(0, "missing frontmatter (--- … ---) with %s", strings.Join(shape.Keys, "/"))
	} else {
		fm := lines[fmStart+1 : fmEnd]
		for _, key := range shape.Keys {
			if !hasKey(fm, key) {
				fail(0, "frontmatter is missing `%s:`", key)
			}
		}
		if shape.FrontmatterChecks != nil {
			shape.FrontmatterChecks(fm, fail)
		}
		if shape.Neutrality != nil {
			erase = shape.Neutrality(fm)
		}
	}

	// The shape — present AND in order. Order is checked, not merely claimed: the sequence is
	// the argument (why you need it → what it rests on → how to build it → what it cost), and
	// a document that opens with its build steps has quietly become a checklist. Checking only
	// presence would let it say anything in any order while the gate reported the genre as
	// enforced.
	prev, prevAt := "", -1
	for _, want := range shape.Sections {
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

	if shape.BodyChecks != nil {
		shape.BodyChecks(name, lines, fail)
	}

	// The conform layer's lint half — injected via `extra`, at this exact position (see the
	// doc comment above). It fires ONLY on recipes that declared an exact contract (ADR-005):
	// the shape must parse, valid fixtures must conform, invalid fixtures must not.
	for _, ex := range extra {
		out = append(out, ex(name, src)...)
	}

	out = append(out, neutrality(name, lines, fmEnd, products, erase)...)
	return out
}
