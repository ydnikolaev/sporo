package recipe

// The return channel, on disk. A recipe improves in exactly one way — somebody builds it
// somewhere else and says what happened — and the `## Report back` protocol every export
// carries asks them for six things. This file is the author's side of that exchange: filing
// what came back where the next authoring session will find it, under git, with the same
// validation discipline the recipes themselves get.
//
// The channel is FILES, deliberately. A report-back handed over as a file works offline, for
// private teams with no shared tracker, and across every provider; the future site syncs this
// same format rather than replacing it. What the channel is NOT: a merge tool. Which returned
// scar was structural and which was the reader's own weather is the judgment that produces
// the recipe's next version, and that belongs to the authoring skill, not to a verb.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"sporo.dev/sporo/pkg/recipekit"
)

// ValidateReport holds a report-back to the protocol's six markers. It forwards to recipekit
// — the validation is pure, and the future site shares this exact check.
func ValidateReport(src []byte) []Finding {
	return recipekit.ValidateReport(src)
}

// reVersionCited finds the version the reader says they built. The protocol does not force a
// grammar on them, so this is a scan, not a parse — and finding nothing is a warning the CLI
// surfaces, never an error: a scar with no version is still a scar.
var reVersionCited = regexp.MustCompile(`(?i)version\W{0,10}(\d+\.\d+\.\d+)|\bv(\d+\.\d+\.\d+)\b`)

// AddFeedback files one report-back for one of THIS project's recipes. It returns the path
// it filed to and, when the report cites no version, a warning for the CLI to print.
//
// Filing is idempotent on CONTENT: the same bytes filed twice land once, because the loop's
// natural failure is a reader pasting the same report into two sessions and the author
// merging one scar twice. The filename carries a sequence (so `ls` reads in arrival order)
// and a content-hash fragment (so two different reports can never collide).
func AddFeedback(root string, cfg Config, slug string, src []byte) (path, warning string, err error) {
	if f := ValidateReport(src); len(f) > 0 {
		msgs := make([]string, len(f))
		for i, x := range f {
			msgs[i] = x.Msg
		}
		return "", "", fmt.Errorf("this is not a report-back yet — the protocol's sections are its whole value:\n  ✗ %s", strings.Join(msgs, "\n  ✗ "))
	}
	if _, err := os.Stat(filepath.Join(root, cfg.Home, slug+".md")); err != nil {
		return "", "", fmt.Errorf("no recipe %q in this project's home (%s) — feedback files against the recipe it answers, and this project does not author that one", slug, cfg.Home)
	}

	dir := filepath.Join(root, ".sporo", "feedback", slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	hash8 := strings.TrimPrefix(ContentHash(src), "sha256:")[:8]

	ents, _ := os.ReadDir(dir)
	next := 1
	for _, e := range ents {
		name := e.Name()
		if strings.Contains(name, hash8) {
			if prior, err := os.ReadFile(filepath.Join(dir, name)); err == nil && string(prior) == string(src) {
				return filepath.Join(dir, name), citedWarning(src), nil // already filed — idempotent
			}
		}
		var n int
		if _, err := fmt.Sscanf(name, "%03d-", &n); err == nil && n >= next {
			next = n + 1
		}
	}

	path = filepath.Join(dir, fmt.Sprintf("%03d-%s.md", next, hash8))
	if err := os.WriteFile(path, src, 0o644); err != nil {
		return "", "", err
	}
	return path, citedWarning(src), nil
}

func citedWarning(src []byte) string {
	if reVersionCited.Match(src) {
		return ""
	}
	return "the report cites no version — ask the reader which version they built (the export's frontmatter carries it); a report that cannot say is ambiguous the day the recipe changes"
}

// ListFeedback enumerates the filed reports, slug → paths, in arrival order. It reads what
// exists rather than what the registry believes: a report is a file somebody handed over,
// and files are the source of truth for files.
func ListFeedback(root string, cfg Config) (map[string][]string, error) {
	base := filepath.Join(root, ".sporo", "feedback")
	slugs, err := os.ReadDir(base)
	if err != nil {
		return map[string][]string{}, nil //nolint:nilerr // no feedback yet is a state, not an error
	}
	out := map[string][]string{}
	for _, s := range slugs {
		if !s.IsDir() {
			continue
		}
		ents, err := os.ReadDir(filepath.Join(base, s.Name()))
		if err != nil {
			continue
		}
		var paths []string
		for _, e := range ents {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				paths = append(paths, filepath.Join(base, s.Name(), e.Name()))
			}
		}
		sort.Strings(paths)
		if len(paths) > 0 {
			out[s.Name()] = paths
		}
	}
	return out, nil
}
