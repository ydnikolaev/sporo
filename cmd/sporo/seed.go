package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	sporo "sporo.dev/sporo"
	"sporo.dev/sporo/internal/recipe"
	"sporo.dev/sporo/pkg/recipekit"
)

// seed is the recipe tool's second corpus: an install SEED is a self-contained prompt that
// brings a tool INTO a repository (detect → install → verify → use → harness), where a recipe
// exports a capability the repo already BUILT. The two share a genre-generic seam — the same
// registry, the same seal/version discipline, the same neutrality gate — so the seed namespace
// is the recipe verbs' mirror, one kind over: `sporo seed new/lint/seal/export/list` do for a
// seed exactly what their flat counterparts do for a recipe, and read the same idiom.
//
// It is a NAMESPACE (not five more top-level verbs) for one reason: the flat verbs already own
// the bare names because sporo IS the recipe tool; a seed is the second citizen, so it takes a
// prefix rather than fighting `lint`/`seal`/`export` for the root. Every subcommand is
// flags-only, exactly like the flat verbs — no interactive surface, so (framework-first-go) no
// prompt loop and no forms dependency; an agent invocation never blocks on input.
func seedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seed",
		Short: "Author, check, seal and export install seeds — the recipe verbs, one kind over",
	}
	cmd.AddCommand(seedNewCmd(), seedLintCmd(), seedSealCmd(), seedExportCmd(), seedListCmd())
	return cmd
}

// seed new scaffolds a draft seed — the seven gated sections stubbed with coach comments, the
// nine-key frontmatter pre-stamped, born `draft: true` so the gate never reds on the state the
// tool itself wrote. It mirrors flat `new`, minus `--from-harvest`: a seed installs a tool from
// the outside, so there is no local build record to pre-seed it from.
func seedNewCmd() *cobra.Command {
	var root, title string
	cmd := &cobra.Command{
		Use:   "new <slug>",
		Args:  cobra.ExactArgs(1),
		Short: "Scaffold a draft seed — coached section stubs, born a draft",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := recipe.LoadConfig(root)
			if err != nil {
				return err
			}
			path, err := recipe.SeedScaffold(root, cfg, args[0], title)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "sporo seed new: draft at %s\n", path)
			fmt.Fprintln(cmd.OutOrStdout(), "fill the TODOs, delete the coach comments, remove `draft: true` — then `sporo seed lint`, `sporo seed seal`, `sporo seed export`")
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "project root (the draft lands in its seed home)")
	cmd.Flags().StringVar(&title, "title", "", "the seed's title (defaults to a TODO)")
	return cmd
}

// seed lint checks the project's seed corpus against the seed genre and rides the seed-scoped
// seal-coherence sweep — the seed counterpart of flat `lint` with no argument. It has no `[dir]`
// positional: the engine (`LintSeedHome`) walks the DECLARED seed home, because a seed is sealed
// against this project's registry and holding an arbitrary directory to those seals would red on
// files the registry never sealed. A project that declares no seed home gets a clean error, not a
// crash — a repo that authors recipes but not seeds simply has no seed corpus here.
func seedLintCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:   "lint",
		Args:  cobra.NoArgs,
		Short: "Check this project's seed corpus against the genre — shape, acceptance, neutrality, seals",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := recipe.LoadConfig(root)
			if err != nil {
				return err
			}
			findings, n, drafts, err := recipe.LintSeedHome(root, cfg)
			if err != nil {
				return err
			}
			for _, f := range findings {
				fmt.Fprintf(cmd.ErrOrStderr(), "  ✗ %s\n", f)
			}
			if len(findings) > 0 {
				return fmt.Errorf("sporo seed lint: the corpus has drifted from the seed genre — a seed that " +
					"names the reader's own tree is a manual, not a transferable install (see the seed authoring spec, `_authoring`)")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "sporo seed lint: %d seed(s) conformant and neutral ✓\n", n)
			if drafts > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "sporo seed lint: %d draft(s) not checked — a draft cannot be sealed or exported; finish it and remove `draft: true`\n", drafts)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "project root (its config names the seed home and the forbidden product vocabulary)")
	return cmd
}

// seed seal records a seed's (version, content hash, provenance, kind) in the registry — flat
// `seal`, one kind over. It writes `kind: seed`, so the same ledger holds both corpora and the
// seal-coherence gate can scope each to its own home; a slug already sealed as a recipe refuses a
// seed seal (the ledger is slug-keyed — one slug is one kind for its life).
func seedSealCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:   "seal <slug>",
		Args:  cobra.ExactArgs(1),
		Short: "Record a seed's version and content hash in the registry — a sealed seed never silently mutates",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := recipe.LoadConfig(root)
			if err != nil {
				return err
			}
			entry, err := recipe.SealKind(root, cfg, recipekit.KindSeed, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "sporo seed seal: %s %s (%s, %s)\n", args[0], entry.Version, entry.Provenance, entry.Hash[:14])
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "project root (its registry records the seal)")
	return cmd
}

// seed export composes one seed into a single self-contained install file for an agent in a
// repository that has never heard of this tool — flat `export`'s mirror, with the composition
// INVERTED: the runner preamble is prepended (a seed IS the prompt, so the discipline for running
// it arrives before the first Install step), where a recipe export appends its adoption protocol.
// The project's own seed home is searched first, then the embedded corpus. Delivery is the whole
// job, so by default the file is WRITTEN to `.sporo/exports/<slug>.md` and the path printed;
// `--stdout` forces the body out for piping, and a run with NO project seed home falls back to
// stdout on its own — a stranger reading an embedded seed from a bare directory must never have a
// `.sporo/` tree created underfoot.
func seedExportCmd() *cobra.Command {
	var root string
	var toStdout bool
	cmd := &cobra.Command{
		Use:   "export <slug>",
		Args:  cobra.ExactArgs(1),
		Short: "Write one seed as a single self-contained install file — runner preamble first, then the seed",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := seedHomeOrNone(root)
			if err != nil {
				return err
			}
			body, err := recipe.SeedExport(sporo.Seeds, home, args[0])
			if err != nil {
				return err
			}
			if toStdout || home == "" {
				fmt.Fprint(cmd.OutOrStdout(), body)
				return nil
			}
			dir := filepath.Join(root, ".sporo", "exports")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			out := filepath.Join(dir, args[0]+".md")
			if err := os.WriteFile(out, []byte(body), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "sporo seed export: wrote %s — hand this file over, not the source\n", out)
			return nil
		},
	}
	cmd.Flags().BoolVar(&toStdout, "stdout", false, "print the composed document to stdout instead of writing a file (for piping)")
	cmd.Flags().StringVar(&root, "root", ".", "project root (searched for this repo's own seeds before the embedded corpus)")
	return cmd
}

// seed list names the seeds available here — this project's own, and the embedded corpus — flat
// `list`'s mirror over the seed home. It has no adopted ledger: adoption is the reader-side seal
// of a handed-over recipe, and a seed is consumed by RUNNING it, not by recording it.
func seedListCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Short: "List the seeds available here — this project's own, and the embedded corpus",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := seedHomeOrNone(root)
			if err != nil {
				return err
			}
			entries, err := seedList(home)
			if err != nil {
				return err
			}
			for _, e := range entries {
				fmt.Fprintf(cmd.OutOrStdout(), "%-8s %s\n", e.Origin, e.Slug)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "project root (its own seeds are listed alongside the embedded corpus)")
	return cmd
}

// seedHomeOrNone resolves this project's seed home, mirroring homeOrNone for the recipe side. A
// root that declares no seed home, or whose seed home is not on disk yet, is "none" — not an error
// for a READ verb (export/list): exporting an embedded seed from an arbitrary directory is exactly
// what a stranger does.
func seedHomeOrNone(root string) (string, error) {
	cfg, err := recipe.LoadConfig(root)
	if err != nil {
		return "", err
	}
	home, ok := cfg.HomeFor(recipekit.KindSeed)
	if !ok {
		return "", nil
	}
	full := filepath.Join(root, home)
	if _, err := os.Stat(full); err != nil {
		return "", nil //nolint:nilerr // no seed home on disk is "none", not an error (see func name)
	}
	return full, nil
}

// seedList enumerates BOTH seed corpora — the embedded one and this project's — as recipe.Entry
// rows so the CLI shares the recipe list's Origin vocabulary. It lives in the CLI, not the engine,
// because the seed engine is INV-1-frozen for this run and this is the only seed-list consumer;
// three predicate clauses beat a premature genre-parameterized walk with one caller.
func seedList(home string) ([]recipe.Entry, error) {
	seen := map[string]bool{}
	var out []recipe.Entry
	if home != "" {
		if ents, err := os.ReadDir(home); err == nil {
			for _, e := range ents {
				if slug, ok := seedFileSlug(e.Name(), e.IsDir()); ok {
					seen[slug] = true
					out = append(out, recipe.Entry{Slug: slug, Origin: recipe.Project})
				}
			}
		}
	}
	ents, err := fs.ReadDir(sporo.Seeds, "seeds")
	if err != nil {
		return nil, fmt.Errorf("read the embedded seed corpus: %w", err)
	}
	for _, e := range ents {
		if slug, ok := seedFileSlug(e.Name(), e.IsDir()); ok && !seen[slug] {
			out = append(out, recipe.Entry{Slug: slug, Origin: recipe.Official})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Slug < out[j].Slug })
	return out, nil
}

// seedFileSlug reports whether a directory entry is a seed instance (and its slug): a `.md` file
// that is neither a `_`-prefixed genre meta-document nor the home's own README. It mirrors the
// engine's isSeed predicate, which is unexported and inside the frozen seed package.
func seedFileSlug(name string, isDir bool) (string, bool) {
	if isDir || !strings.HasSuffix(name, ".md") || strings.HasPrefix(name, "_") || name == "README.md" {
		return "", false
	}
	return strings.TrimSuffix(name, ".md"), true
}
