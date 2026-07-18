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
//
// The genre logic itself is PURE (bytes → findings) and lives in pkg/recipekit, so a future
// registry server enforces exactly what the CLI does. This file keeps the package's API
// stable by forwarding to it; the one non-pure piece — the conform layer's fixture check,
// which parses JSON/YAML — is injected into Lint here rather than pulled up into recipekit.

import "sporo.dev/sporo/pkg/recipekit"

// Finding is one violation, located. It aliases recipekit.Finding, so `[]Finding` and
// `recipe.Finding` remain the exact types every caller already uses.
type Finding = recipekit.Finding

// Lint checks one recipe against the genre. It forwards to recipekit.Lint and injects
// fixtureFindings — the conform layer's lint half, which parses candidate shapes and so
// cannot live in the pure package — at the position recipekit.Lint reserves for it.
func Lint(name string, src []byte, products []string) []Finding {
	return recipekit.Lint(name, src, products, fixtureFindings)
}

// IsDraft reports a recipe that declares `draft: true` in its frontmatter — the state
// `sporo new` scaffolds in, exempt from the genre gate and barred from seal and export.
func IsDraft(src []byte) bool {
	return recipekit.IsDraft(src)
}

// sectionBody returns the lines under a `## ` heading, up to the next `## ` heading. Kept as
// an unexported forwarder because conform.go reads contract bodies with it.
func sectionBody(lines []string, heading string) []string {
	return recipekit.SectionBody(lines, heading)
}
