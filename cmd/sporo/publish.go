package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"sporo.dev/sporo/internal/recipe"
)

// `sporo publish <slug>` is the front door to the shared corpus. It does two things and nothing in
// between: it VERIFIES the recipe from its bytes (recipe.PublishPreflight — sealed, gate-passed,
// neutral, refusing each failure by name), then HANDS the branch the author already committed off to
// whichever PR-open mechanism the machine has. It never creates branches, commits, or stages files —
// there is nothing for it to mutate behind the author's back; the review point is the PR itself.
//
// The handoff degrades but never blocks (the operator chose "gh + git-fallback"):
//   - gh installed, authenticated, and interactive → `gh pr create` (the canonical primitive: it
//     pushes the branch and, lacking write access to the corpus, offers to fork).
//   - otherwise → the exact `git push` + compare-URL steps, which need only the git every recipe
//     author already has. "No gh" costs two paste steps; a token/API path is parked with the
//     registry plan, not built, because at this scale it is setup burden for no gain.
//
// The merge on the corpus side is what attests the exported bytes — publish opens the door, the
// corpus workflow (docs/design/attested-provenance.md, S2) signs what comes through it.

// corpusBase is the branch a publish PR targets. The corpus is the origin repository itself (this
// repo dogfoods its own corpus), and its trunk is main.
const corpusBase = "main"

func publishCmd() *cobra.Command {
	var root string
	var check bool
	cmd := &cobra.Command{
		Use:   "publish <slug>",
		Args:  cobra.ExactArgs(1),
		Short: "Verify a recipe is sealed and gate-passed, then open a corpus PR",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := recipe.PublishPreflight(root, args[0])
			if err != nil {
				return err
			}
			if check {
				// The CI/server semantic: one line, exit 0, no PR — this is what the corpus workflow
				// runs on the pushed bytes before it attests them.
				fmt.Fprintf(cmd.OutOrStdout(), "sporo publish --check: %s %s is sealed and gate-passed (%s)\n", res.Slug, res.Entry.Version, res.Hash[:14])
				return nil
			}
			return openCorpusPR(root, res, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "project root")
	cmd.Flags().BoolVar(&check, "check", false, "verify only (the CI/server semantic): print the one-line result and exit")
	return cmd
}

// openCorpusPR runs the handoff after a passed pre-flight. It refuses to publish from the base
// branch — a PR opens *from* a branch — and otherwise takes the gh path when it is viable, the
// printed-steps path when it is not. res.Version and res.Hash go into the PR so the reviewer starts
// from the seal, not a re-derivation.
func openCorpusPR(root string, res *recipe.PublishResult, stdout io.Writer) error {
	branch, err := gitBranch(root)
	if err != nil {
		return err
	}
	if branch == corpusBase || branch == "HEAD" {
		return fmt.Errorf("you are on %q — publish opens a PR *from* a branch; commit your sealed recipe on one first (`git switch -c publish-%s`), then re-run", branch, res.Slug)
	}

	title := fmt.Sprintf("publish(%s): %s %s", res.Entry.Kind, res.Slug, res.Entry.Version)
	body := fmt.Sprintf("Sealed and gate-passed locally (`sporo publish --check`): %s %s (%s).\n\n"+
		"Merging attests the exported bytes to the official sporo pipeline.", res.Slug, res.Entry.Version, res.Hash[:14])

	if ghReady() {
		fmt.Fprintf(stdout, "sporo publish: %s %s verified (%s) — opening a PR with gh…\n", res.Slug, res.Entry.Version, res.Hash[:14])
		c := exec.Command("gh", "pr", "create", "--base", corpusBase, "--head", branch, "--title", title, "--body", body)
		c.Dir = root
		// Stream the terminal through: gh's own push/confirm prompts ARE the human's review point.
		c.Stdin, c.Stdout, c.Stderr = os.Stdin, stdout, os.Stderr
		return c.Run()
	}

	// Fallback: print the exact steps. Everything the author needs to open the PR by hand, and
	// nothing done behind their back. The header stays neutral about WHY (gh may be installed but the
	// run is piped/agent-driven — claiming "gh is not available" then would be a lie); the install
	// tip is printed only when gh is genuinely off PATH.
	fmt.Fprintf(stdout, "sporo publish: %s %s — sealed and gate-passed (%s). Open the PR by hand:\n", res.Slug, res.Entry.Version, res.Hash[:14])
	fmt.Fprintf(stdout, "\n  git push -u origin %s\n", branch)
	if owner, repoName, perr := originSlug(root); perr == nil {
		fmt.Fprintf(stdout, "  open https://github.com/%s/%s/compare/%s...%s?expand=1\n", owner, repoName, corpusBase, branch)
		fmt.Fprintf(stdout, "\n(No write access to %s/%s? Fork it, push the branch to your fork, and open the PR from there.)\n", owner, repoName)
	}
	if !ghInstalled() {
		fmt.Fprintln(stdout, "\ntip: install GitHub CLI (https://cli.github.com) and run `gh auth login` to let `sporo publish` open the PR for you next time.")
	}
	return nil
}

// ghInstalled reports only whether gh is on PATH — the one fact that distinguishes "install gh" from
// "gh is here, this run just isn't interactive".
func ghInstalled() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

// ghReady reports whether the gh path is viable: gh installed, authenticated, AND an interactive
// terminal. The TTY gate is the point — `gh pr create` prompts (where to push, confirm the PR), and
// a prompt that blocks on absent stdin is exactly what an agent-invoked or piped run must avoid
// (framework-first-go). A non-interactive run takes the printed-steps path, which never blocks.
func ghReady() bool {
	if !ghInstalled() {
		return false
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}
	// `gh auth status` exits non-zero when no account is logged in.
	return exec.Command("gh", "auth", "status").Run() == nil
}

func gitBranch(root string) (string, error) {
	c := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	c.Dir = root
	out, err := c.Output()
	if err != nil {
		return "", fmt.Errorf("could not read the current git branch (is %s a git repository?): %w", root, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// originSlug parses owner/repo out of the origin remote, handling both the SSH
// (git@github.com:owner/repo.git) and HTTPS (https://github.com/owner/repo.git) forms. It is
// best-effort: the compare URL is a convenience atop the git push, so a parse miss degrades to
// "push and open a PR", never an error.
func originSlug(root string) (owner, repo string, err error) {
	c := exec.Command("git", "remote", "get-url", "origin")
	c.Dir = root
	out, err := c.Output()
	if err != nil {
		return "", "", err
	}
	return parseGithubSlug(strings.TrimSpace(string(out)))
}

// parseGithubSlug pulls owner/repo out of a github.com remote URL in either the SSH
// (git@github.com:owner/repo.git) or HTTPS (https://github.com/owner/repo.git) form. Split out from
// the git shell-out so the fiddly parsing is unit-tested without a repository.
func parseGithubSlug(remote string) (owner, repo string, err error) {
	url := strings.TrimSuffix(remote, ".git")
	if i := strings.Index(url, "github.com"); i >= 0 {
		rest := strings.TrimLeft(url[i+len("github.com"):], ":/")
		if parts := strings.Split(rest, "/"); len(parts) >= 2 && parts[0] != "" && parts[1] != "" {
			return parts[0], parts[1], nil
		}
	}
	return "", "", fmt.Errorf("origin %q is not a github.com remote", remote)
}
