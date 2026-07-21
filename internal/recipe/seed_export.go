package recipe

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// RunnerDoc is the seed genre's execution preamble — the short constitutional meta-doc that
// frames how ANY seed is run. Unlike the recipe's adoption protocol (which is APPENDED, because
// a recipe is a payload that needs a consumption path afterwards), the runner preamble is
// PREPENDED, because a seed IS the prompt and the preamble is the frame the agent reads first.
const RunnerDoc = "_runner.md"

// SeedExport hands ONE self-contained file to an agent in a repository that has never heard of
// this harness — the seed's counterpart to Export, and its deliberate MIRROR-INVERSION.
//
// A recipe export is `body + adoption protocol`: the payload, then how to consume it. A seed
// export is `runner preamble + body`: the frame, then the instruction. The order is the whole
// difference. A recipe describes a capability and the reader decides how to adopt it afterwards;
// a seed IS an executable install the agent runs top-to-bottom, so the discipline for running it
// safely — read the anchors, work the sections in order, run nothing that does not trace to the
// declared source, close by filling the Report — must arrive BEFORE the first Install step, not
// after the last one. Hence: preamble first.
//
// The provenance banner is stripped from both the preamble and the body (house business about a
// repository the reader does not have). The seed body's FRONTMATTER is retained: the agent reads
// `target` and `source` to know which promise it is keeping and the only origin it may run code
// from. The runner's own meta-doc frontmatter is dropped — its keys are the ledger's business,
// and the reader wants the prose.
//
// The project's OWN seed home is searched first, then the embedded corpus — a seed about the
// repository you are standing in is the one you most want to hand out, and if a slug exists in
// both, the local build is the one its author verified. It fails CLOSED when the runner preamble
// is missing: a seed shipped without it would tell the agent what to build and never how to run
// it safely.
func SeedExport(seeds fs.FS, home, slug string) (string, error) {
	if strings.HasPrefix(slug, "_") {
		return "", fmt.Errorf("%q is the seed genre's own meta-document, not a seed: it teaches how to "+
			"WRITE one, and exporting it to someone who wants to INSTALL the tool helps nobody", slug)
	}
	preamble, err := runnerPreamble(seeds)
	if err != nil {
		return "", err
	}
	if home != "" {
		if b, err := os.ReadFile(filepath.Join(home, slug+".md")); err == nil {
			if IsDraft(b) {
				return "", fmt.Errorf("%q is still a draft — exporting it would hand a stranger TODOs as if they were earned; finish it, remove `draft: true`, get `sporo seed lint` green, then export", slug)
			}
			return composeSeed(preamble, string(b)), nil
		}
	}
	b, err := fs.ReadFile(seeds, path.Join("seeds", slug+".md"))
	if err != nil {
		return "", fmt.Errorf("no seed %q in the seed corpus — export composes a seed that exists: %w", slug, err)
	}
	return composeSeed(preamble, string(b)), nil
}

// composeSeed joins the version-stamped runner preamble to the banner-stripped seed body. The two
// are separated by a blank line, NOT the recipe path's `---` thematic break: the seed body opens
// with its own `---` frontmatter fence, and a `---` immediately before it would read as an empty
// second frontmatter block. The seed's frontmatter fence is the boundary the reader sees.
func composeSeed(preamble, seedSrc string) string {
	return strings.TrimRight(preamble, "\n") + "\n\n" + strip(seedSrc)
}

// runnerPreamble is the head of every seed export: the runner meta-doc with its provenance banner
// and its own frontmatter stripped (the reader wants the prose, not the ledger keys), a
// `> **Runner protocol:** v<version>` stamp prepended so the handed-over file names the
// constitution it shipped with (INV-2). The intro paragraph before the first `## ` heading is
// KEPT — unlike the recipe adoption protocol (whose pre-heading lines are a note to its own
// authors), the runner's opening is addressed to the executing agent, so dropping-above-first-`##`
// would discard exactly the framing the preamble exists to deliver.
func runnerPreamble(seeds fs.FS) (string, error) {
	b, err := fs.ReadFile(seeds, path.Join("seeds", RunnerDoc))
	if err != nil {
		return "", fmt.Errorf("the runner preamble is missing from the corpus (%s): every exported seed "+
			"carries it as its first instruction — a seed handed over without it tells the agent what to "+
			"BUILD and never how to RUN it safely; the binary is broken, not the seed: %w", RunnerDoc, err)
	}
	version, err := specVersion(b, RunnerDoc)
	if err != nil {
		return "", err
	}
	prose := stripFrontmatter(strip(string(b)))
	return "> **Runner protocol:** v" + version + "\n\n" + prose, nil
}

// RunnerVersion exposes the version of the runner preamble embedded in the binary — the sibling of
// AdoptionVersion and GenreVersion. The preamble file remains the SSOT: the version is read from
// its frontmatter, never a second constant that can drift from the prose it claims to identify.
func RunnerVersion(seeds fs.FS) (string, error) {
	b, err := fs.ReadFile(seeds, path.Join("seeds", RunnerDoc))
	if err != nil {
		return "", fmt.Errorf("the runner preamble is missing from the corpus (%s): %w", RunnerDoc, err)
	}
	return specVersion(b, RunnerDoc)
}

// stripFrontmatter drops a leading `---`-fenced frontmatter block (and the blank lines below it),
// returning the prose. It assumes the banner is already gone, so the block is at the head. A
// document with no frontmatter is returned unchanged.
func stripFrontmatter(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return s
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			rest := lines[i+1:]
			for len(rest) > 0 && strings.TrimSpace(rest[0]) == "" {
				rest = rest[1:]
			}
			return strings.Join(rest, "\n")
		}
	}
	return s
}
