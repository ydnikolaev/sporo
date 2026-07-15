package recipe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// `sporo new` — the answer to the genre's real supply constraint, which is not reader
// trust but AUTHORING COST. Eleven gated sections written from a blank page is expensive
// enough that the recipe nobody writes is the common case; the gate can only reject, it
// cannot help. The scaffold helps: every section arrives with a coach comment saying what
// belongs in it (and, for the two sections authors reliably get wrong, what GOOD looks
// like), the frontmatter arrives pre-stamped, and a harvest can pre-seed the scars — the
// highest-value, highest-effort section — as candidates the author judges instead of
// recalls.
//
// Two properties are load-bearing, and both are tested:
//
//   - The scaffold is born a DRAFT (`draft: true`), so the gate does not red on the state
//     the tool itself wrote, and seal/export refuse it until the author says it is done.
//   - The scaffold, minus the draft mark, is GENRE-GREEN. A template that fails its own
//     gate teaches the author that the gate is noise — so the template is the first
//     conformant document the author reads, and filling it in cannot make it less
//     conformant by shape, only by content.
func Scaffold(root string, cfg Config, slug, title string, h *Harvest) (string, error) {
	if title == "" {
		title = "TODO — one line naming the capability, not the technology"
	}
	var b strings.Builder
	project := "this project"
	if p := projectName(root, cfg); p != "" {
		project = p
	}

	fmt.Fprintf(&b, "<!-- SSOT SOURCE (%s). The export strips this banner; edit here, hand over ONLY what `sporo export` prints. -->\n\n", project)
	b.WriteString("---\n")
	fmt.Fprintf(&b, "name: %s\n", slug)
	b.WriteString("version: 0.1.0\n")
	b.WriteString("draft: true\n")
	fmt.Fprintf(&b, "title: %s\n", title)
	b.WriteString("problem: TODO — one sentence, the reader's pain, not your solution\n")
	b.WriteString("prerequisites: [read-files, edit-files] # capabilities, never tool names\n")
	b.WriteString("derived_from: [the build record of one live implementation]\n")
	b.WriteString(`stack: { language: TODO, runtime: TODO, why: "TODO — what property of the build made this the right stack" }` + "\n")
	fmt.Fprintf(&b, "verified: { project: %s, release: TODO, date: %s } # the build that PROVES this — stamp it when the thing works, not before\n", project, time.Now().Format("2006-01-02"))
	b.WriteString("effort: TODO — honest, and name which half eats the budget\n")
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "# %s\n\n", strings.TrimSuffix(title, " — one line naming the capability, not the technology"))
	b.WriteString("<!-- coach: every comment like this one is scaffolding. Replace the TODOs, DELETE the\n" +
		"     coach comments, remove `draft: true` — then `sporo lint`, `sporo seal`, `sporo export`.\n" +
		"     The one rule that cannot bend: the body names ROLES, never your paths, filenames or\n" +
		"     product names. Read `sporo genre` before filling anything. -->\n\n")

	b.WriteString("## The problem\n\n" +
		"<!-- coach: what the reader does NOT have, what the output IS, and the acceptance — how\n" +
		"     they will know they have it — at the top, in the reader's terms. -->\n\nTODO\n\n")

	b.WriteString("## Why the obvious approach fails\n\n" +
		"<!-- coach: the design the next agent WILL reach for first, and the concrete way it\n" +
		"     breaks. This section earns the recipe's existence: without it the reader just\n" +
		"     follows their first instinct. -->\n\nTODO\n\n")

	b.WriteString("## The principles\n\n" +
		"<!-- coach: the payload — portable, load-bearing claims that survive on a stack you have\n" +
		"     never seen. Everything below instantiates something here. Cite a principle that\n" +
		"     already has a home in one line; never re-derive it. -->\n\n")
	if h != nil && len(h.Design) > 0 {
		b.WriteString("<!-- coach: the harvest proposes these design commits as raw material (judge, then delete this list):\n")
		for _, c := range capped(h.Design, 10) {
			fmt.Fprintf(&b, "     - %.12s %s\n", c.Hash, c.Subject)
		}
		b.WriteString("-->\n\n")
	}
	b.WriteString("TODO\n\n")

	b.WriteString("## The ground it needs\n\n" +
		"<!-- coach: what must be STANDING before step one, and why each is load-bearing. Write\n" +
		"     every precondition as a LADDER: probe for it → build the smallest one → degrade,\n" +
		"     and label the degradation where the capability's own output shows it. There is no\n" +
		"     fourth rung. -->\n\n")
	if h != nil && len(h.Absent) > 0 {
		b.WriteString("<!-- coach: this project's own record LACKS the following — the matching sections below\n" +
			"     cannot be harvested and must be sourced by hand:\n")
		for _, a := range h.Absent {
			fmt.Fprintf(&b, "     - %s\n", a)
		}
		b.WriteString("-->\n\n")
	}
	b.WriteString("TODO\n\n")

	b.WriteString("## The contracts\n\n" +
		"<!-- coach: SHOW every shape the capability consumes or emits — the reader copies it\n" +
		"     instead of re-inventing it, incompatibly. Field names in the reader's language,\n" +
		"     placeholders where a value would be a coordinate, a note on any field whose meaning\n" +
		"     is a trap. The gate requires at least one fenced block here. -->\n\n" +
		"```json\n" +
		"{ \"schema\": 1, \"TODO\": \"the record this capability persists\" }\n" +
		"```\n\n")

	// The coach text says "Done-when" WITHOUT the literal bold marker: the gate counts the
	// marker against the step headings, and a comment that utters it would make the tool's
	// own scaffold fail the tool's own arithmetic.
	b.WriteString("## The build sequence\n\n" +
		"<!-- coach: one `###` heading per step, written against the contracts above — an\n" +
		"     interface, not your implementation. EVERY step ends with a bold Done-when line\n" +
		"     naming an OBSERVATION: \"the check exits non-zero on a seeded defect\" is done;\n" +
		"     \"it works\" is a wish. -->\n\n" +
		"### 1. TODO\n\nTODO\n\n**Done when:** TODO — name the observation, not the intention.\n\n")

	b.WriteString("## The seams\n\n" +
		"<!-- coach: what MUST stay configurable so the next project does not inherit YOUR\n" +
		"     values — name the seam and what varies across it, never the value. -->\n\nTODO\n\n")

	b.WriteString("## The scars\n\n" +
		"<!-- coach: the highest-value section, and the one a clean-room rebuild cannot produce.\n" +
		"     Only EARNED scars; keep them concrete, coordinates stripped: \"summing parallel\n" +
		"     sessions produced 25.3 hours inside a 24-hour day\" teaches; \"be careful with\n" +
		"     concurrency\" does not. -->\n\n")
	if h == nil || len(h.Scars) == 0 {
		b.WriteString("### TODO — name the failure\n\n**Symptom:** TODO — what you observed.\n**Root cause:** TODO — why it happened.\n**Fix:** TODO — what durably prevents it.\n\n")
	} else {
		for _, c := range capped(h.Scars, 20) {
			fmt.Fprintf(&b, "### Candidate from the record: %s\n\n", c.Subject)
			fmt.Fprintf(&b, "<!-- coach: proposed by %.12s (signals: %s). Keep it ONLY if the failure was\n"+
				"     structural — delete it if incidental — and strip any coordinate out of the\n"+
				"     subject line above before you remove `draft: true`. -->\n", c.Hash, strings.Join(c.Signals, ", "))
			b.WriteString("**Symptom:** TODO — what broke, observably.\n**Root cause:** TODO — the record knows; read the commit, do not recall.\n**Fix:** TODO — what durably prevents it.\n\n")
		}
		if len(h.Scars) > 20 {
			fmt.Fprintf(&b, "<!-- coach: %d more scar candidates in the harvest file — the cap keeps this draft readable, not the record small. -->\n\n", len(h.Scars)-20)
		}
	}

	b.WriteString("## Verification\n\n" +
		"<!-- coach: the gates that ship WITH the capability (unguarded invariants rot), plus the\n" +
		"     one live check that says it really works — the thing no gate can perform. -->\n\nTODO\n\n")

	b.WriteString("## The trade-offs\n\n" +
		"<!-- coach: what this DESIGN costs, what it refuses, and when NOT to build it at all. A\n" +
		"     recipe that only advocates is marketing. (The stack's cost goes in the next\n" +
		"     section, not here.) -->\n\nTODO — including when not to build it.\n\n")

	b.WriteString("## For the human\n\n" +
		"<!-- coach: the ONLY section for a person. Plain language: what gets built, what you have\n" +
		"     at the end. Name the original stack and split the reasons — ESSENTIAL (the design\n" +
		"     depends on it) vs INCIDENTAL (the author's house; swap freely) — then the stack's\n" +
		"     trade-offs, as trade-offs. -->\n\nTODO\n")

	dir := filepath.Join(root, cfg.Home)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, slug+".md")
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("recipe %q already exists at %s — the scaffold never overwrites; pick another slug or edit the existing file", slug, path)
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// projectName resolves what `verified.project` and the banner should call this repository:
// the config's declared product name when there is one, the directory's own name otherwise.
func projectName(root string, cfg Config) string {
	if len(cfg.Products) > 0 && strings.TrimSpace(cfg.Products[0]) != "" {
		return cfg.Products[0]
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return ""
	}
	return filepath.Base(abs)
}

func capped(c []Candidate, n int) []Candidate {
	if len(c) > n {
		return c[:n]
	}
	return c
}
