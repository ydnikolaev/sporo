package recipekit

import (
	"strings"
	"testing"
)

// A synthetic second genre, built inline and NEVER registered: linting through it is exactly
// what proves the seam — a new kind runs on the same engine with no engine change. Its shape
// is deliberately unlike the recipe's (two sections, one key) so a green here cannot be the
// recipe rules leaking in.
func syntheticShape() Shape {
	return Shape{
		Kind:     KindSeed,
		Sections: []string{"## Alpha", "## Beta"},
		Keys:     []string{"handle"},
		FrontmatterChecks: func(fm []string, fail FailFunc) {
			if v := KeyLine(fm, "handle"); v != "" && !strings.Contains(v, "seed-") {
				fail(0, "`handle:` must name a seed (seed-…)")
			}
		},
		BodyChecks: func(name string, lines []string, fail FailFunc) {
			if len(SectionBody(lines, "## Alpha")) == 0 {
				fail(0, "Alpha section is empty")
			}
		},
	}
}

const syntheticValid = `<!-- SSOT SOURCE -->
---
handle: seed-widget
---
## Alpha
Body text under alpha.
## Beta
Body text under beta.
`

func TestASyntheticShapeLintsAValidFixtureGreen(t *testing.T) {
	if f := LintShape(syntheticShape(), "widget.md", []byte(syntheticValid), nil); len(f) != 0 {
		t.Fatalf("a valid fixture must lint green through the same engine; got: %v", f)
	}
}

func TestASyntheticShapeRedsAMissingSection(t *testing.T) {
	body := strings.Replace(syntheticValid, "## Beta\nBody text under beta.\n", "", 1)
	f := LintShape(syntheticShape(), "widget.md", []byte(body), nil)
	if !hasFinding(f, "missing section: Beta") {
		t.Fatalf("a missing required section must red the synthetic shape; got: %v", f)
	}
}

// The teeth INV-1 actually rests on. The governed suite checks finding messages, never their
// order — every recipe fixture carries exactly one violation — so a phase firing out of
// sequence would still pass green there. Pin the sequence directly: a document that breaks a
// frontmatter check, a section check, and neutrality must return them in phase order
// (frontmatter → section → neutrality), which is the order every consumer of the corpus reads.
func TestFindingsComeBackInPhaseOrder(t *testing.T) {
	const body = `<!-- SSOT SOURCE -->
---
handle: nope
---
## Beta
names glarch here.
## Alpha
Body under alpha.
`
	f := LintShape(syntheticShape(), "widget.md", []byte(body), []string{"glarch"})
	if len(f) < 3 {
		t.Fatalf("expected findings from three phases; got %d: %v", len(f), f)
	}
	if !strings.Contains(f[0].Msg, "handle") {
		t.Fatalf("frontmatter finding must come first; got: %v", f)
	}
	if !strings.Contains(f[1].Msg, "out of order") {
		t.Fatalf("section finding must come second; got: %v", f)
	}
	if !strings.Contains(f[2].Msg, "product") {
		t.Fatalf("neutrality finding must come last; got: %v", f)
	}
}

func TestShapeForRoundTrip(t *testing.T) {
	RegisterShape(Shape{Kind: KindSeed, Sections: []string{"## X"}, Keys: []string{"h"}})
	got, ok := ShapeFor(KindSeed)
	if !ok || got.Kind != KindSeed {
		t.Fatalf("a registered shape must round-trip through ShapeFor; got %+v ok=%v", got, ok)
	}
	if _, ok := ShapeFor(KindRecipe); !ok {
		t.Fatal("the recipe shape must be registered by its initializer")
	}
	if _, ok := ShapeFor("no-such-kind"); ok {
		t.Fatal("ShapeFor must report an unregistered kind as absent")
	}
}

func TestRegisterShapePanicsOnUnknownKind(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("registering a kind outside the vocabulary must panic")
		}
	}()
	RegisterShape(Shape{Kind: "not-in-vocab"})
}

func TestRegisterShapePanicsOnDuplicateKind(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("registering an already-registered kind must panic")
		}
	}()
	RegisterShape(RecipeShape) // recipe is already registered by RecipeShape's initializer
}

func TestValidKind(t *testing.T) {
	for _, k := range []string{KindRecipe, KindSeed} {
		if !ValidKind(k) {
			t.Fatalf("%q is a vocabulary member and must be valid", k)
		}
	}
	for _, k := range []string{"", "seeds", "Recipe", "unknown"} {
		if ValidKind(k) {
			t.Fatalf("%q is not a vocabulary member and must be invalid", k)
		}
	}
}

func hasFinding(findings []Finding, want string) bool {
	for _, f := range findings {
		if strings.Contains(f.Msg, want) {
			return true
		}
	}
	return false
}
