package recipe

import (
	"strings"
	"testing"
	"testing/fstest"
)

// The exported file is the ONLY thing the reader ever sees. A recipe is a payload with no
// consumption path — it says what ground the capability needs and never says what to do when
// the reader's ground is missing half of it — so the delivery step appends the one protocol
// that closes that gap. These teeth exist because the failure is silent: an export missing
// the protocol looks like a complete document, and the reader who most needs it (the one
// whose repository does not match the author's, which is every reader) never learns it existed.

func TestTheExportedRecipeCarriesTheAdoptionProtocol(t *testing.T) {
	corpus := fstest.MapFS{
		"recipes/a-thing.md":   {Data: []byte("<!-- SSOT SOURCE -->\n\n---\nname: a-thing\n---\n# A thing\n")},
		"recipes/_adoption.md": {Data: []byte("<!-- SSOT SOURCE -->\n<!-- house business -->\n\n## Adopt it here\nProbe, do not assume.\n\n## Report back\nNew scars.\n")},
	}
	body, err := Export(corpus, "", "a-thing")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# A thing", "## Adopt it here", "## Report back"} {
		if !strings.Contains(body, want) {
			t.Fatalf("the exported file must be the recipe AND the protocol; %q is missing:\n%s", want, body)
		}
	}
	if strings.Contains(body, "house business") {
		t.Fatalf("the protocol's note to its own authors is not for the reader:\n%s", body)
	}
	if strings.Index(body, "# A thing") > strings.Index(body, "## Adopt it here") {
		t.Fatal("the protocol comes after the recipe: you cannot decide how to adopt a capability you have not read")
	}
}

// Fail CLOSED. A binary whose corpus lost the protocol would otherwise hand a stranger a
// document that reads as complete and quietly omits the only section addressed to them.
func TestAnExportWithNoProtocolIsAnError(t *testing.T) {
	corpus := fstest.MapFS{
		"recipes/a-thing.md": {Data: []byte("<!-- SSOT SOURCE -->\n\n# A thing\n")},
	}
	if _, err := Export(corpus, "", "a-thing"); err == nil {
		t.Fatal("a recipe with no adoption protocol has no consumption path — the export must refuse, not degrade silently")
	}
}

// The adoption protocol's neutrality is checked in pkg/recipekit (that is where the
// neutrality scan lives); its section presence is asserted there too, against the same
// corpus file this test used to read.
