// Command sporo is the recipe CLI: it harvests a build's record, checks a recipe against the
// genre, and exports one as a single self-contained file for an agent in a repository that has
// never heard of this tool.
//
// The verbs are top-level (`sporo lint`, not `sporo recipe lint`) because sporo IS the recipe
// tool — there is no other namespace to disambiguate against. The surface spans both sides of
// a handover: authoring (harvest/new/lint/seal/export, feedback, review), reading (adopt,
// pull, conform), the install surface (init/update, genre, projects) and the binary's own
// freshness (upgrade). Only `push` — publishing into a shared corpus — waits on the site.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	sporo "sporo.dev/sporo"
	"sporo.dev/sporo/internal/install"
	"sporo.dev/sporo/internal/recipe"
	"sporo.dev/sporo/internal/upgrade"
)

// version is stamped by the release pipeline (goreleaser ldflags); "dev" means a local build.
// It is what the provenance stamp on every managed file cites, so a user can tell which
// binary wrote what.
var version = "dev"

func main() {
	if err := root().Execute(); err != nil {
		os.Exit(1)
	}
}

func root() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "sporo",
		Short:         "Author, check and export transferable build recipes — in any repository",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	cmd.AddCommand(harvestCmd(), lintCmd(), exportCmd(), listCmd(), sealCmd(),
		initCmd(), updateCmd(), genreCmd(), feedbackCmd(), reviewCmd(), projectsCmd(), newCmd(),
		conformCmd(), upgradeCmd(), adoptCmd(), pullCmd())

	// The passive freshness hint: one line on stderr when a newer release is known, refreshed
	// through the network at most once a day (the cache answers in between), silent on any
	// failure, and off entirely for dev builds, for CI (nobody is there to read it), for the
	// user who said no (SPORO_NO_UPDATE_CHECK), and after `upgrade` itself.
	cmd.PersistentPostRun = func(c *cobra.Command, args []string) {
		if c.Name() == "upgrade" || os.Getenv("SPORO_NO_UPDATE_CHECK") != "" || os.Getenv("CI") != "" {
			return
		}
		hint := upgrade.Hint(install.GlobalHome(), version, time.Now(), func() (string, error) {
			ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
			defer cancel()
			latest, err := upgrade.Latest(ctx, releaseToken())
			if err != nil {
				return "", err
			}
			return latest.Version(), nil
		})
		if hint != "" {
			fmt.Fprintln(c.ErrOrStderr(), hint)
		}
	}
	return cmd
}

// upgrade replaces THIS binary with the latest release (checksum-validated) — the first half
// of the post-release chain; `sporo update`, per repo, is the second. A dev build refuses:
// its upstream is a checkout, and replacing it with a release would eat the developer's own
// build.
func upgradeCmd() *cobra.Command {
	var check bool
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Update this binary to the latest release — then run `sporo update` in each repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			if check || version == "dev" {
				latest, err := upgrade.Latest(ctx, releaseToken())
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "sporo upgrade: latest is %s, you run %s\n", latest.Version(), version)
				if version == "dev" {
					fmt.Fprintln(cmd.OutOrStdout(), "this is a dev build — it was built from a checkout and does not self-update; `go build` your update, or install a release")
				}
				if notes := strings.TrimSpace(latest.ReleaseNotes); notes != "" && check {
					fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", notes)
				}
				return nil
			}
			latest, err := upgrade.Self(ctx, version, releaseToken())
			if err != nil {
				return err
			}
			if latest.LessOrEqual(version) {
				fmt.Fprintf(cmd.OutOrStdout(), "sporo upgrade: %s is already the latest\n", version)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "sporo upgrade: %s → %s\n", version, latest.Version())
			fmt.Fprintln(cmd.OutOrStdout(), "now push its new skills into your repositories: `sporo update` in each (`sporo projects` lists them)")
			return nil
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "report the latest release and its notes; change nothing")
	return cmd
}

// releaseToken is the optional auth for the private-repository phase. Read here, once, so
// the upgrade package never reaches into the environment behind the command's back.
func releaseToken() string {
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	return os.Getenv("GH_TOKEN")
}

// conform is the reader's half of `Binding: exact`: it checks an output file against a
// recipe's exact-bound contracts, structurally, with a path per violation. It works from
// the EXPORTED file alone (the only thing a reader has), so a whole fleet can run it in CI
// against the one document they all received — which is what turns "the schema must be the
// same everywhere" from an agreement into a check. A recipe with no exact contracts is a
// clean no-op: that absence is the author's declaration (ADR-005), not a failure.
func conformCmd() *cobra.Command {
	var root string
	var contractN int
	cmd := &cobra.Command{
		Use:   "conform <slug|recipe.md> <output-file>",
		Args:  cobra.ExactArgs(2),
		Short: "Check an output file against a recipe's exact-bound contracts",
		RunE: func(cmd *cobra.Command, args []string) error {
			src, err := recipeSource(root, args[0])
			if err != nil {
				return err
			}
			contracts := recipe.ExactContracts(src)
			if len(contracts) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "sporo conform: this recipe declares no exact-bound contracts — every shape is Binding: adapt, and there is nothing to hold an output to (that is the author's declaration, not a failure)")
				return nil
			}
			candidate, err := os.ReadFile(args[1])
			if err != nil {
				return err
			}
			failures := 0
			for _, c := range contracts {
				if contractN > 0 && c.Index != contractN {
					continue
				}
				violations, err := recipe.Conform(c, candidate)
				if err != nil {
					return err
				}
				if len(violations) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "sporo conform: contract #%d ✓\n", c.Index)
					continue
				}
				failures++
				for _, v := range violations {
					fmt.Fprintf(cmd.ErrOrStderr(), "  ✗ contract #%d %s\n", c.Index, v)
				}
			}
			if failures > 0 {
				return fmt.Errorf("sporo conform: the output does not hold the promise its recipe makes — fix the output, or renegotiate the contract with the recipe's author (that is a report-back)")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "project root (searched for the recipe when a slug is given)")
	cmd.Flags().IntVar(&contractN, "contract", 0, "check against one exact contract by its index (default: all)")
	return cmd
}

// recipeSource resolves the document conform reads: a path to a file (the exported handoff
// — the reader's normal case), this project's own home, or the official corpus.
func recipeSource(root, ref string) ([]byte, error) {
	if b, err := os.ReadFile(ref); err == nil {
		return b, nil
	}
	cfg, err := recipe.LoadConfig(root)
	if err != nil {
		return nil, err
	}
	if b, err := os.ReadFile(filepath.Join(root, cfg.Home, ref+".md")); err == nil {
		return b, nil
	}
	if b, err := fs.ReadFile(sporo.Recipes, "recipes/"+ref+".md"); err == nil {
		return b, nil
	}
	return nil, fmt.Errorf("no recipe %q — give a slug from this project or the official corpus, or a path to an exported recipe file", ref)
}

// new scaffolds a draft recipe — every section stubbed with a coach comment saying what
// belongs in it, the frontmatter pre-stamped, and (with --from-harvest) the scars pre-seeded
// as candidates the author judges instead of recalls. The draft cannot be sealed or exported
// until `draft: true` is removed — the scaffold helps the author start, never lets a start
// masquerade as a finish.
func newCmd() *cobra.Command {
	var root, title, harvestFile string
	cmd := &cobra.Command{
		Use:   "new <slug>",
		Args:  cobra.ExactArgs(1),
		Short: "Scaffold a draft recipe — coached section stubs, optionally pre-seeded from a harvest",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := recipe.LoadConfig(root)
			if err != nil {
				return err
			}
			var h *recipe.Harvest
			if harvestFile != "" {
				b, err := os.ReadFile(harvestFile)
				if err != nil {
					return err
				}
				h = &recipe.Harvest{}
				if err := json.Unmarshal(b, h); err != nil {
					return fmt.Errorf("%s is not a harvest file (`sporo harvest --out` writes one): %w", harvestFile, err)
				}
			}
			path, err := recipe.Scaffold(root, cfg, args[0], title, h)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "sporo new: draft at %s\n", path)
			fmt.Fprintln(cmd.OutOrStdout(), "fill the TODOs, delete the coach comments, remove `draft: true` — then `sporo lint`, `sporo seal`, `sporo export`")
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "project root (the draft lands in its recipes home)")
	cmd.Flags().StringVar(&title, "title", "", "the recipe's title (defaults to a TODO)")
	cmd.Flags().StringVar(&harvestFile, "from-harvest", "", "a `sporo harvest --out` file; its scar candidates pre-seed the scars section")
	return cmd
}

// init installs the authoring surface into THIS repository: the skill into the provider
// homes it detects, a managed block into AGENTS.md, and the one-time seeds (config, recipes
// home). Everything written is recorded in the registry with a content hash — that record is
// how `update` later tells its own files from yours.
func initCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Install the recipe-authoring surface into this repository (skill, AGENTS.md block, seeds)",
		RunE: func(cmd *cobra.Command, args []string) error {
			actions, err := install.Init(root, version)
			report(cmd, actions)
			registerBestEffort(cmd, root)
			return err
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "repository to initialize")
	return cmd
}

// update re-syncs the managed surface from this (possibly newer) binary. The chain after a
// release is `sporo upgrade` (new binary — a later verb) then `sporo update` (its new skills
// into this repo). A file the user edited is reported and preserved, never overwritten —
// there is deliberately no --force.
func updateCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Re-sync the managed authoring surface from this binary — never clobbering your edits",
		RunE: func(cmd *cobra.Command, args []string) error {
			actions, err := install.Update(root, version)
			report(cmd, actions)
			if err == nil {
				registerBestEffort(cmd, root)
			}
			return err
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "repository to update")
	return cmd
}

// registerBestEffort notes this repository in the machine-level projects list — the list
// `sporo projects` walks after an upgrade to find stale skills. Best-effort by design: the
// project-local install is the contract, the global list is a courtesy, and a machine with
// an unwritable home still gets a working init.
func registerBestEffort(cmd *cobra.Command, root string) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return
	}
	if err := install.RegisterProject(install.GlobalHome(), abs, version); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "sporo: note — could not record this repo in the global projects list (%v); `sporo projects` will not know about it\n", err)
	}
}

// genre prints the authoring spec from the binary's corpus. In a consumer repository the
// binary is the only place the genre lives, and the skill's first instruction is to read it.
func genreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genre",
		Short: "Print the recipe genre spec — the authoring rules this binary enforces",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := recipe.Genre(sporo.Recipes)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), s)
			return nil
		},
	}
	return cmd
}

func report(cmd *cobra.Command, actions []install.Action) {
	for _, a := range actions {
		if a.Note != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "sporo: %-7s %s — %s\n", a.Status, a.Path, a.Note)
			continue
		}
		fmt.Fprintf(cmd.OutOrStdout(), "sporo: %-7s %s\n", a.Status, a.Path)
	}
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
			n, drafts := 0, 0
			for _, e := range ents {
				if !recipe.IsRecipe(e.Name(), e.IsDir()) && !isMeta(e.Name()) {
					continue
				}
				src, err := os.ReadFile(filepath.Join(dir, e.Name()))
				if err != nil {
					return err
				}
				// A draft is exempt, and says so: reds on the state `sporo new` itself writes
				// would train red-blindness, but a draft silently passing would read as done.
				if recipe.IsDraft(src) {
					drafts++
					continue
				}
				n++
				findings = append(findings, recipe.Lint(e.Name(), src, cfg.Products)...)
			}
			// Bundle manifests ride the same gate: a member that resolves to nothing is a
			// build order with a hole in it, and the hole transfers.
			for _, name := range recipe.Bundles(dir) {
				b, err := recipe.LoadBundle(dir, name)
				if err != nil {
					return err
				}
				findings = append(findings, recipe.LintBundle(sporo.Recipes, dir, name, b)...)
			}
			// The registry's coherence rides the same gate — but only when the corpus being
			// checked is the project's own home. An explicit directory argument may be anyone's
			// corpus, and holding it to THIS project's seals would red on files the registry
			// never sealed.
			if len(args) == 0 {
				regFindings, err := recipe.VerifyRegistry(root, cfg)
				if err != nil {
					return err
				}
				findings = append(findings, regFindings...)
			}
			for _, f := range findings {
				fmt.Fprintf(cmd.ErrOrStderr(), "  ✗ %s\n", f)
			}
			if len(findings) > 0 {
				return fmt.Errorf("sporo lint: the corpus has drifted from the genre — a recipe that " +
					"names its origin is a manual (see the genre spec, `_authoring`)")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "sporo lint: %d recipe(s) conformant and neutral ✓\n", n)
			if drafts > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "sporo lint: %d draft(s) not checked — a draft cannot be sealed or exported; finish it and remove `draft: true`\n", drafts)
			}
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
	var bundle bool
	cmd := &cobra.Command{
		Use:   "export <slug>",
		Args:  cobra.ExactArgs(1),
		Short: "Print one recipe (or, with --bundle, a composed set) as a single self-contained file",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := homeOrNone(root)
			if err != nil {
				return err
			}
			// A bundle composes several recipes into the same delivery contract: one
			// document, one adoption protocol, members in build order (the genre stays
			// one-capability; scale is the export's job).
			var body string
			if bundle {
				body, err = recipe.ExportBundle(sporo.Recipes, home, args[0])
			} else {
				body, err = recipe.Export(sporo.Recipes, home, args[0])
			}
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), body)
			return nil
		},
	}
	cmd.Flags().BoolVar(&bundle, "bundle", false, "treat <slug> as a bundle manifest and compose its members into one document")
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
			// Adopted recipes are listed too — they are capabilities this repo BUILT, and a
			// list that hides them reads as "nothing was ever handed to us".
			adopted, err := recipe.AdoptedList(root)
			if err != nil {
				return err
			}
			slugs := make([]string, 0, len(adopted))
			for s := range adopted {
				slugs = append(slugs, s)
			}
			sort.Strings(slugs)
			for _, s := range slugs {
				fmt.Fprintf(cmd.OutOrStdout(), "%-8s %s (%s)\n", "adopted", s, adopted[s].Version)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "project root (its own recipes are listed alongside the official corpus)")
	return cmd
}

// adopt records a handed-over recipe: the exported file, verbatim, plus (version, hash,
// exact-contract digest, source) in the registry. It is the reader-side twin of the seal —
// the moment "somebody gave me this text" becomes something the repository remembers,
// instead of something one agent session knew.
func adoptCmd() *cobra.Command {
	var root, source string
	cmd := &cobra.Command{
		Use:   "adopt <exported-recipe.md>",
		Args:  cobra.ExactArgs(1),
		Short: "Record a handed-over recipe this repository builds from — the reader-side seal",
		RunE: func(cmd *cobra.Command, args []string) error {
			src, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			if source == "" {
				source = args[0]
			}
			slug, entry, err := recipe.Adopt(root, src, source)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "sporo adopt: %s %s (source: %s)\n", slug, entry.Version, entry.Source)
			if n := len(recipe.ExactContracts(src)); n > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "%d exact contract(s) — wire `sporo conform %s <your output>` into this repo's CI\n", n, args[0])
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "repository that adopts the recipe")
	cmd.Flags().StringVar(&source, "source", "", "where pull re-fetches this recipe from later — a path or an http(s) URL (default: the file argument)")
	return cmd
}

// pull re-checks every adopted recipe against its source. READ-ONLY by default: discovering
// that the source moved on is cheap; acting on it is a rebuild, and that is agent work
// judged against this repository — `--apply` is the explicit second step that refreshes the
// stored copy and the record.
func pullCmd() *cobra.Command {
	var root string
	var apply bool
	cmd := &cobra.Command{
		Use:   "pull [slug]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Check adopted recipes against their sources — loud when an exact contract moved",
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := ""
			if len(args) == 1 {
				slug = args[0]
			}
			reports, err := recipe.Pull(root, slug, apply)
			if err != nil {
				return err
			}
			for _, r := range reports {
				switch r.Status {
				case "up to date":
					fmt.Fprintf(cmd.OutOrStdout(), "sporo pull: %s %s — up to date\n", r.Slug, r.Have)
				case "skipped":
					fmt.Fprintf(cmd.OutOrStdout(), "sporo pull: %s — skipped: %s\n", r.Slug, r.Note)
				case "update":
					fmt.Fprintf(cmd.OutOrStdout(), "sporo pull: %s %s → %s\n", r.Slug, r.Have, r.Latest)
					if r.ExactChanged {
						fmt.Fprintf(cmd.OutOrStdout(), "  an EXACT contract changed (that is a major) — your consumer-facing output must be re-verified; rebuild, then re-run `sporo conform`\n")
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "  safe update: scars and prose moved, the exact contracts did not\n")
					}
					if apply {
						fmt.Fprintf(cmd.OutOrStdout(), "  applied — the stored copy and the record now say %s\n", r.Latest)
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "  (report only — `sporo pull --apply` refreshes the record once you rebuild)\n")
					}
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "repository whose adopted recipes are checked")
	cmd.Flags().BoolVar(&apply, "apply", false, "refresh the stored copy and record for each update (default: report only)")
	return cmd
}

// seal records a recipe's (version, content hash, provenance) in `.sporo/registry.yaml` —
// the moment a draft becomes a promise. From then on the pair is guarded by `sporo lint`:
// content that changes under a sealed version is a finding, and the fix is a version bump
// plus a re-seal, never a quiet edit. Report-backs bind to the version the reader built;
// the seal is what makes that citation mean something.
func sealCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:   "seal <slug>",
		Args:  cobra.ExactArgs(1),
		Short: "Record a recipe's version and content hash in the registry — a sealed recipe never silently mutates",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := recipe.LoadConfig(root)
			if err != nil {
				return err
			}
			entry, err := recipe.Seal(root, cfg, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "sporo seal: %s %s (%s, %s)\n", args[0], entry.Version, entry.Provenance, entry.Hash[:14])
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "project root (its registry records the seal)")
	return cmd
}

// feedback is the return channel's author side. A recipe improves in exactly one way —
// somebody builds it elsewhere and reports back — and this verb files what came back where
// git and the next authoring session will find it. Merging the scars into the recipe's next
// version stays judgment: the skill's job, deliberately not a flag here.
func feedbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feedback",
		Short: "File and list report-backs — the channel a recipe's next version comes from",
	}

	var addRoot string
	add := &cobra.Command{
		Use:   "add <slug> <report.md>",
		Args:  cobra.ExactArgs(2),
		Short: "Validate a reader's report-back against the protocol and file it (`-` reads stdin)",
		RunE: func(cmd *cobra.Command, args []string) error {
			src, err := readArg(cmd, args[1])
			if err != nil {
				return err
			}
			cfg, err := recipe.LoadConfig(addRoot)
			if err != nil {
				return err
			}
			path, warning, err := recipe.AddFeedback(addRoot, cfg, args[0], src)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "sporo feedback: filed %s\n", path)
			if warning != "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "sporo feedback: warning — %s\n", warning)
			}
			return nil
		},
	}
	add.Flags().StringVar(&addRoot, "root", ".", "project root (the recipe this report answers must be authored here)")

	var listRoot string
	list := &cobra.Command{
		Use:   "list [slug]",
		Args:  cobra.MaximumNArgs(1),
		Short: "List the filed report-backs, per recipe",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := recipe.LoadConfig(listRoot)
			if err != nil {
				return err
			}
			all, err := recipe.ListFeedback(listRoot, cfg)
			if err != nil {
				return err
			}
			for slug, paths := range all {
				if len(args) == 1 && args[0] != slug {
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s (%d report(s))\n", slug, len(paths))
				for _, p := range paths {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", p)
				}
			}
			return nil
		},
	}
	list.Flags().StringVar(&listRoot, "root", ".", "project root")

	cmd.AddCommand(add, list)
	return cmd
}

// review is the semantic half of the gate, provider-agnostic by construction: it never calls
// an agent. `review <slug>` composes one self-contained prompt (rubric + verdict schema +
// the exported recipe); the user runs it through whatever agent they have — one or several —
// and `review verify` validates the returned JSON and records the tally beside the seal.
func reviewCmd() *cobra.Command {
	var packRoot string
	cmd := &cobra.Command{
		Use:   "review <slug>",
		Args:  cobra.ExactArgs(1),
		Short: "Build a self-contained review pack for any agent, and verify the verdicts it returns",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := recipe.LoadConfig(packRoot)
			if err != nil {
				return err
			}
			path, err := recipe.BuildReviewPack(sporo.Recipes, packRoot, cfg, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "sporo review: pack at %s\n", path)
			fmt.Fprintf(cmd.OutOrStdout(), "run it through any agent, e.g.  claude -p \"$(cat %s)\" > verdict.json\nthen:  sporo review verify %s verdict.json\n", path, args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&packRoot, "root", ".", "project root")

	var verifyRoot string
	verify := &cobra.Command{
		Use:   "verify <slug> <verdict.json>...",
		Args:  cobra.MinimumNArgs(2),
		Short: "Validate returned verdicts and record the tally beside the recipe's seal",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := recipe.LoadConfig(verifyRoot)
			if err != nil {
				return err
			}
			sum, findings, err := recipe.VerifyVerdicts(verifyRoot, cfg, args[0], args[1:])
			if err != nil {
				return err
			}
			for _, f := range findings {
				fmt.Fprintf(cmd.ErrOrStderr(), "  ✗ %s\n", f)
			}
			if len(findings) > 0 {
				return fmt.Errorf("sporo review: %d verdict problem(s) — nothing was recorded", len(findings))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "sporo review: %s %s — %d verdict(s), mean %.1f/10, %s\n",
				args[0], sum.Version, sum.Verdicts, sum.Mean, sum.Verdict)
			return nil
		},
	}
	verify.Flags().StringVar(&verifyRoot, "root", ".", "project root")

	cmd.AddCommand(verify)
	return cmd
}

// projects lists the repositories this machine installed sporo into — the walk list after an
// upgrade: any repo whose recorded binary is older than this one has stale skills, and
// `sporo update` there is the fix.
func projectsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "projects",
		Short: "List the repositories on this machine that sporo was installed into",
		RunE: func(cmd *cobra.Command, args []string) error {
			ps, err := install.Projects(install.GlobalHome())
			if err != nil {
				return err
			}
			if len(ps) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "sporo: no projects registered yet — `sporo init` in a repository records it here")
				return nil
			}
			for _, p := range ps {
				hint := ""
				if p.Binary != version {
					hint = fmt.Sprintf("  ← installed by %s, this binary is %s: run `sporo update` there", p.Binary, version)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-10s %s  (%s)%s\n", p.Binary, p.Root, p.Updated, hint)
			}
			return nil
		},
	}
}

// readArg reads a file argument, honoring `-` as stdin — a report-back often arrives as a
// paste, and forcing a temp file on the paster is friction the loop cannot afford.
func readArg(cmd *cobra.Command, arg string) ([]byte, error) {
	if arg == "-" {
		return io.ReadAll(cmd.InOrStdin())
	}
	return os.ReadFile(arg)
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
