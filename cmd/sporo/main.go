// Command sporo is the recipe CLI: it harvests a build's record, checks a recipe against the
// genre, and exports one as a single self-contained file for an agent in a repository that has
// never heard of this tool.
//
// The verbs are top-level (`sporo lint`, not `sporo recipe lint`) because sporo IS the recipe
// tool — there is no other namespace to disambiguate against. `init`/`update`/`upgrade` and
// the site verbs (`push`/`pull`) arrive in later releases; this entrypoint carries the four
// that make the tool useful standalone today.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	sporo "sporo.dev/sporo"
	"sporo.dev/sporo/internal/recipe"
)

func main() {
	if err := root().Execute(); err != nil {
		os.Exit(1)
	}
}

func root() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "sporo",
		Short:         "Author, check and export transferable build recipes — in any repository",
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	cmd.AddCommand(harvestCmd(), lintCmd(), exportCmd(), listCmd())
	return cmd
}

// harvest mines the record a build already left behind — each release's rationale, the commits
// that fixed something, the gates that shipped, the doctrine and decisions it moved, and the
// defects whose own message says they were found only by deliberately breaking the thing. It
// emits CANDIDATES with the signal that proposed each, and NAMES the records this project does
// not have, because a silent zero and an unread source are the same empty array. It never
// writes a recipe: which failure was structural and which incidental is the judgment that IS
// the recipe, and a machine that guessed it would be confident and wrong where nobody checks.
func harvestCmd() *cobra.Command {
	var root, since, until, out string
	cmd := &cobra.Command{
		Use:   "harvest",
		Short: "Gather a recipe's raw material from a revision range of the project's own record",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := recipe.LoadConfig(root)
			if err != nil {
				return err
			}
			h, err := recipe.Gather(root, since, until, cfg)
			if err != nil {
				return err
			}
			b, err := json.MarshalIndent(h, "", "  ")
			if err != nil {
				return err
			}
			b = append(b, '\n')
			if out == "" {
				_, err = cmd.OutOrStdout().Write(b)
				return err
			}
			if err := os.WriteFile(out, b, 0o644); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"sporo: %s — %d scar candidate(s), %d design commit(s), %d gate(s), %d unsignaled\n",
				out, len(h.Scars), len(h.Design), len(h.Gates), h.Unsignaled)
			for _, a := range h.Absent {
				fmt.Fprintf(cmd.ErrOrStderr(), "sporo: not in this project's record — %s\n", a)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "repository to harvest")
	cmd.Flags().StringVar(&since, "since", "", "starting revision, exclusive (a tag: the release before the work)")
	cmd.Flags().StringVar(&until, "until", "HEAD", "ending revision, inclusive")
	cmd.Flags().StringVar(&out, "out", "", "write the harvest here (default: stdout)")
	return cmd
}

// lint checks a recipe against the genre: the shape (the eleven sections, the frontmatter, a
// `**Done when:**` per build step, symptom/root-cause/fix per scar, a shown shape under The
// contracts) and the one constraint the genre exists for — NEUTRALITY. The body may name
// technologies and show CONTRACTS (shapes a reader copies and adapts); it may not name
// COORDINATES (a path, a filename, a product), which execute in one repository and transfer to
// none. With no argument, the project's own recipes home is checked.
func lintCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:   "lint [dir]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Check a recipe corpus against the genre — shape, acceptance, scars, and neutrality",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := recipe.LoadConfig(root)
			if err != nil {
				return err
			}
			dir := filepath.Join(root, cfg.Home)
			if len(args) == 1 {
				dir = args[0]
			}
			ents, err := os.ReadDir(dir)
			if err != nil {
				return fmt.Errorf("sporo lint: no recipe corpus at %s — author one there, or point the "+
					"linter at the corpus you mean (`sporo lint <dir>`)", dir)
			}
			var findings []recipe.Finding
			n := 0
			for _, e := range ents {
				if !recipe.IsRecipe(e.Name(), e.IsDir()) && !isMeta(e.Name()) {
					continue
				}
				src, err := os.ReadFile(filepath.Join(dir, e.Name()))
				if err != nil {
					return err
				}
				n++
				findings = append(findings, recipe.Lint(e.Name(), src, cfg.Products)...)
			}
			for _, f := range findings {
				fmt.Fprintf(cmd.ErrOrStderr(), "  ✗ %s\n", f)
			}
			if len(findings) > 0 {
				return fmt.Errorf("sporo lint: the corpus has drifted from the genre — a recipe that " +
					"names its origin is a manual (see the genre spec, `_authoring`)")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "sporo lint: %d recipe(s) conformant and neutral ✓\n", n)
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "project root (its config names the recipes home and the forbidden product vocabulary)")
	return cmd
}

// export prints one recipe as a single self-contained file — banner stripped, adoption protocol
// appended — for an agent in a repository that has never heard of this tool. The official corpus
// is compiled into the binary for exactly that reason; the project's own recipes are searched
// first.
func exportCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:   "export <slug>",
		Args:  cobra.ExactArgs(1),
		Short: "Print one recipe as a single self-contained file",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := homeOrNone(root)
			if err != nil {
				return err
			}
			body, err := recipe.Export(sporo.Recipes, home, args[0])
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), body)
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "project root (searched for this repo's own recipes before the official corpus)")
	return cmd
}

func listCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the recipes available here — this project's own, and the official corpus",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := homeOrNone(root)
			if err != nil {
				return err
			}
			entries, err := recipe.List(sporo.Recipes, home)
			if err != nil {
				return err
			}
			for _, e := range entries {
				fmt.Fprintf(cmd.OutOrStdout(), "%-8s %s\n", e.Origin, e.Slug)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "project root (its own recipes are listed alongside the official corpus)")
	return cmd
}

// homeOrNone resolves this project's recipes home. A root that is not a project (no config, no
// home on disk) is not an error for a READ verb: exporting the official corpus from an arbitrary
// directory is exactly what a stranger does.
func homeOrNone(root string) (string, error) {
	cfg, err := recipe.LoadConfig(root)
	if err != nil {
		return "", err
	}
	home := filepath.Join(root, cfg.Home)
	if _, err := os.Stat(home); err != nil {
		return "", nil
	}
	return home, nil
}

// isMeta reports the genre's own `_`-prefixed documents. They are held to the banner alone by
// the linter, but the CLI must still hand them to it (the meta-document IS checked for its
// banner); IsRecipe excludes them, so the corpus loop needs this second predicate.
func isMeta(name string) bool { return len(name) > 0 && name[0] == '_' }
