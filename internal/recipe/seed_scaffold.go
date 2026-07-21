package recipe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sporo.dev/sporo/pkg/recipekit"
)

// SeedScaffold — `sporo seed new`, the seed genre's answer to the same supply constraint the
// recipe scaffold answers: a nine-key frontmatter and seven gated sections written from a blank
// page is expensive enough that the seed nobody writes is the common case. The scaffold hands the
// author the first conformant document — pre-stamped, coach-commented — so filling it in cannot
// make it LESS conformant by shape, only by content.
//
// Two properties are load-bearing and both are tested, mirroring the recipe scaffold:
//
//   - The seed is born a DRAFT (`draft: true`), so the gate does not red on the state the tool
//     itself wrote, and seal/export refuse it until the author says it is done.
//   - The seed, minus the draft mark, is SEED-GENRE-GREEN — the seven sections in order, the nine
//     frontmatter keys, a detect-first `## Install` step that carries its acceptance, a fenced
//     `## Verify` proof, and the fixed five-row `## Report`. A template that fails its own gate
//     teaches the author the gate is noise; this one is the first document the gate passes.
//
// The pre-stamped placeholders are honest, not lies: a TODO acceptance is a blank to fill, never a
// claim that something was proven. The trust contract (detect + per-step done-when + a runnable
// Verify + no uncited pipe) is satisfied by SHAPE here; its content is left for the author to earn.
func SeedScaffold(root string, cfg Config, slug, title string) (string, error) {
	home, ok := cfg.HomeFor(recipekit.KindSeed)
	if !ok {
		return "", fmt.Errorf("this project declares no seed corpus to scaffold into — a seed is authored under a `homes: {seed: …}` home; declare one in the project config, then `sporo seed new`")
	}
	if title == "" {
		title = "TODO — one line naming the tool this brings in"
	}
	var b strings.Builder
	project := "this project"
	if p := projectName(root, cfg); p != "" {
		project = p
	}

	fmt.Fprintf(&b, "<!-- SSOT SOURCE (%s). The export strips this banner; edit here, hand over ONLY what `sporo seed export` prints. -->\n\n", project)
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\n", NewID())
	fmt.Fprintf(&b, "name: %s\n", slug)
	b.WriteString("version: 0.1.0\n")
	b.WriteString("draft: true\n")
	fmt.Fprintf(&b, "title: %s\n", title)
	// The target is the seed's pinned subject; the line is held to `<tool>@<version>` by the gate,
	// so the placeholder is valid-by-shape (no trailing comment — the format check reads the whole
	// line) and the guidance lives in the coach comment below the frontmatter instead.
	b.WriteString("target: TODO-name-the-tool@0.0.0\n")
	b.WriteString("source: TODO — the canonical origin the reader can inspect; every Install step traces back here\n")
	b.WriteString(`stack: { language: TODO, runtime: TODO, why: "TODO — what the verifying install actually ran on" }` + "\n")
	fmt.Fprintf(&b, "verified: { project: %s, release: TODO, date: %s } # stamp this when the install PROVES out, not before\n", project, time.Now().Format("2006-01-02"))
	b.WriteString("effort: TODO — honest, so the reader can budget before they start rather than three steps in\n")
	b.WriteString("---\n\n")

	fmt.Fprintf(&b, "# %s\n\n", strings.TrimSuffix(title, " — one line naming the tool this brings in"))
	b.WriteString("<!-- coach: every comment like this one is scaffolding. Replace the TODOs, DELETE the\n" +
		"     coach comments, remove `draft: true` — then `sporo seed lint`, `sporo seed seal`,\n" +
		"     `sporo seed export`. Leave `id:` ALONE — it is this seed's permanent identity, minted\n" +
		"     once; editing it breaks the permalink a marketplace and every report-back hang on.\n" +
		"     The one rule that cannot bend: name the TARGET tool concretely (pinned, traced to its\n" +
		"     source), but speak of the reader's own tree only in ROLES — never their paths,\n" +
		"     filenames, or config locations. Read the seed authoring spec before filling anything. -->\n\n")

	b.WriteString("## Summary\n\n" +
		"<!-- coach: 2–4 sentences orienting a human or agent before any move — what tool this brings\n" +
		"     in, what standing it up buys the reader, and the state they are in when it is done. Body\n" +
		"     text, so the neutrality rule holds: name the tool, not the reader's tree. -->\n\n" +
		"TODO — Write a short orientation naming the tool this seed installs, what standing it up buys the reader, and the state they are left in when it is done.\n\n")

	b.WriteString("## What it is\n\n" +
		"<!-- coach: the tool itself — what it does, the shape of the thing that lands (a single\n" +
		"     binary, a language package, a service the reader runs), and the model the reader needs\n" +
		"     in their head before letting it onto their machine. Understanding BEFORE acquisition. -->\n\nTODO\n\n")

	// The coach text describes the two Install markers WITHOUT uttering them literally: the body
	// checks count the detect and done-when markers across this whole section and match the step
	// headings by their leading `###`, so a comment that spelled any of the three would inflate the
	// tool's own arithmetic and red its own scaffold — the same trap the recipe scaffold avoids.
	b.WriteString("## Install\n\n" +
		"<!-- coach: the acquisition sequence, one step per H3 heading. The FIRST step opens with the\n" +
		"     detect marker — is the tool already here, and at what version? — which makes the seed\n" +
		"     idempotent and keeps a second run from clobbering a working tree. EVERY step, the detect\n" +
		"     one included, closes with a bold done-when acceptance line naming an OBSERVATION the\n" +
		"     agent can check. If a step fetches and runs remote code, cite the source origin the\n" +
		"     frontmatter vouches for — never pipe from wherever. -->\n\n" +
		"### 1. TODO — detect whether the tool is already present\n\n" +
		"**Detect:** TODO — probe for the target and read its version; say what a present install looks like.\n\n" +
		"TODO — the acquisition move, run from the declared source.\n\n" +
		"**Done when:** TODO — name the observable condition that proves this step took.\n\n")

	b.WriteString("## Verify\n\n" +
		"<!-- coach: the proof the whole install works — at least one FENCED command block the agent\n" +
		"     runs and reads, not prose. Install can lie: a package half-lands, a PATH does not\n" +
		"     refresh, a binary arrives without its execute bit. Show a command whose output settles\n" +
		"     it. -->\n\n" +
		"```\n" +
		"TODO --version   # a command whose output proves the tool actually runs here\n" +
		"```\n\n" +
		"TODO — say what output proves the install took.\n\n")

	b.WriteString("## Use\n\n" +
		"<!-- coach: the first real thing the reader does with the now-installed tool, in their own\n" +
		"     repository — the payoff the Summary promised, concrete enough to act on but still\n" +
		"     neutral about the reader's tree. -->\n\nTODO\n\n")

	b.WriteString("## Harness\n\n" +
		"<!-- coach: how the tool joins THIS repository's agent harness so future agents know it is\n" +
		"     here. Advisory: recommend a thin project-local rule only when the tool ships none of its\n" +
		"     own — point at the tool's own guidance wherever it has some, or you author a second\n" +
		"     source of truth that drifts the first time the tool updates. -->\n\nTODO\n\n")

	b.WriteString("## Report\n\n" +
		"<!-- coach: the ONLY section for a person — a fixed five-row audit of what this run did to\n" +
		"     the repository. Fill each cell; do not add, drop, or reorder a row — the human reads the\n" +
		"     same five every time, and the fixed shape IS the product. -->\n\n" +
		"| row | what happened |\n" +
		"|---|---|\n" +
		"| **what it is** | TODO — the tool, in one line the human can act on |\n" +
		"| **how it works** | TODO — the mechanism, enough that the human can reason about it |\n" +
		"| **what was done** | TODO — the actual mutations this run made to this repository |\n" +
		"| **how to use it** | TODO — the human's own next move with the now-installed tool |\n" +
		"| **suggest next** | TODO — where to go from here, the forward pointer |\n")

	dir := filepath.Join(root, home)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, slug+".md")
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("seed %q already exists at %s — the scaffold never overwrites; pick another slug or edit the existing file", slug, path)
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}
