package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	sporo "sporo.dev/sporo"
	"sporo.dev/sporo/internal/recipe"
)

// webMirrorCmd regenerates the committed export forms the site serves as each recipe's `.md`
// mirror (web/src/data/exports/<slug>.md), byte-for-byte the `sporo export <slug>` output. It is
// Hidden because it is a build tool, not a user verb — the user-facing composer is `sporo export`;
// this one exists so `go generate` can refresh the committed mirror and a git-diff gate can prove
// the site never drifts from the binary's handover file. The composition itself lives once, in
// recipe.Export; this command only fans it across the corpus and writes the results.
func webMirrorCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:    "web-mirror",
		Short:  "Regenerate the site's committed recipe export mirror (used by `go generate`)",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := recipe.WriteMirror(sporo.Recipes, "", out)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "sporo web-mirror: wrote %d export form(s) to %s\n", n, out)
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "directory for the <slug>.md export forms (used by `go generate`)")
	_ = cmd.MarkFlagRequired("out")
	return cmd
}

// seedMirrorCmd is the seed twin of web-mirror: it regenerates the committed export forms the site
// serves as each seed's `.md` mirror (web/src/data/seeds/<slug>.md), byte-for-byte the
// `sporo seed export <slug>` output. Hidden, because it is a build tool run only by `go generate`
// — the user-facing composer is `sporo seed export`; this one exists so a git-diff gate can prove
// the site never drifts from the binary's seed handover. The composition lives once, in
// recipe.SeedExport; this command only fans it across the corpus and writes the results.
//
// One deliberate difference from web-mirror: with ZERO seeds it writes NOTHING and creates NO
// directory. The embedded seed corpus is underscore-only until a real seed is sealed, and git
// stores no empty directory — an unconditional MkdirAll would commit an empty web/src/data/seeds
// that a reader's fresh checkout would lack. So the mirror directory is ABSENT, not empty, until a
// seed exists, and the drift gate stays clean because nothing is written.
func seedMirrorCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:    "seed-mirror",
		Short:  "Regenerate the site's committed seed export mirror (used by `go generate`)",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := writeSeedMirror(sporo.Seeds, "", out)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "sporo seed-mirror: wrote %d export form(s) to %s\n", n, out)
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "directory for the <slug>.md seed export forms (used by `go generate`)")
	_ = cmd.MarkFlagRequired("out")
	return cmd
}

// writeSeedMirror regenerates outDir to hold one `<slug>.md` per seed, each the byte-for-byte
// recipe.SeedExport composition (runner preamble first, banner-stripped body). It is the seed twin
// of recipe.WriteMirror, kept in the CLI (not internal/recipe) because the seed engine is
// INV-1-frozen for this run and this is its only mirror consumer — so it composes from the public
// SeedExport, exactly as seedList keeps the seed-corpus walk here rather than in the engine.
//
// It creates NO directory when there are no seeds to mirror (absent-not-empty): the corpus is
// underscore-only until a real seed lands, git stores no empty dir, and S4 must tolerate the
// directory's ABSENCE. Zero seeds ⇒ zero files ⇒ no dir — the early return is BEFORE any MkdirAll.
// When seeds exist it mirrors WriteMirror: prune stale forms first (a removed seed must not leave
// its export behind), then write each seed's composed form.
func writeSeedMirror(seeds fs.FS, home, outDir string) (int, error) {
	slugs := seedMirrorSlugs(seeds, home)
	if len(slugs) == 0 {
		return 0, nil
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return 0, err
	}
	want := make(map[string]bool, len(slugs))
	for _, s := range slugs {
		want[s+".md"] = true
	}
	existing, err := filepath.Glob(filepath.Join(outDir, "*.md"))
	if err != nil {
		return 0, err
	}
	for _, f := range existing {
		if !want[filepath.Base(f)] {
			if err := os.Remove(f); err != nil {
				return 0, err
			}
		}
	}
	for _, s := range slugs {
		body, err := recipe.SeedExport(seeds, home, s)
		if err != nil {
			return 0, err
		}
		if err := os.WriteFile(filepath.Join(outDir, s+".md"), []byte(body), 0o644); err != nil {
			return 0, err
		}
	}
	return len(slugs), nil
}

// seedMirrorSlugs enumerates the seed slugs to mirror — this project's own seed home first, then the
// embedded corpus (a local build wins over the embedded copy of the same slug). It filters with
// seedFileSlug, the same `.md`-not-`_`-not-README predicate seedList uses, so the underscore genre
// docs are never mirrored. The corpus is an fs.FS parameter (not sporo.Seeds directly) so a test
// can inject a fixture corpus — the embedded one is underscore-only until a seed is sealed.
func seedMirrorSlugs(seeds fs.FS, home string) []string {
	seen := map[string]bool{}
	var out []string
	if home != "" {
		if ents, err := os.ReadDir(home); err == nil {
			for _, e := range ents {
				if slug, ok := seedFileSlug(e.Name(), e.IsDir()); ok {
					seen[slug] = true
					out = append(out, slug)
				}
			}
		}
	}
	if ents, err := fs.ReadDir(seeds, "seeds"); err == nil {
		for _, e := range ents {
			if slug, ok := seedFileSlug(e.Name(), e.IsDir()); ok && !seen[slug] {
				out = append(out, slug)
			}
		}
	}
	sort.Strings(out)
	return out
}
