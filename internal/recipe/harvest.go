// Package recipe carries the transferable-build genre: the harvest that gathers a
// recipe's raw material from the record, and the export that hands one file to an
// agent in a repository that has never heard of this harness.
//
// The division of labour is the same one the reporting doctrine draws, for the same
// reason: THE MACHINE GATHERS, THE AUTHOR JUDGES. A harvest never decides what a scar
// is — deciding which failure was structural and which was incidental is the entire
// value the author adds. It proposes CANDIDATES, says by what signal it proposed each
// one, and leaves the judgment where judgment belongs.
package recipe

import (
	"fmt"
	"os/exec"
	"strings"
)

// Harvest is the raw material of one recipe, gathered from a revision range of the
// project's own record. Every field is evidence; none of it is a conclusion.
type Harvest struct {
	Range    Range     `json:"range"`
	Releases []Release `json:"releases"`
	// Sources says WHERE the harvest read — the project's own record, resolved through the
	// config seam. Printed with the result, because a reader who cannot see which records
	// were open cannot tell a thin harvest from a thin project.
	Sources   Sources     `json:"sources"`
	Scars     []Candidate `json:"scar_candidates"`
	Design    []Candidate `json:"design_candidates"`
	Gates     []string    `json:"gates_added"`
	Doctrines []string    `json:"doctrines_touched"`
	// Decisions and Knowledge are the WHY the commit log is worst at carrying: the decision
	// log records what was chosen against what, and the knowledge base records the ground the
	// build stands on. A recipe that skips them reconstructs the rationale from commit
	// subjects, which is a guess wearing the record's clothes.
	Decisions []string `json:"decisions_touched"`
	Knowledge []string `json:"knowledge_touched"`
	// Absent names the records this project does not have. A STATED absence, never a silent
	// zero: "no decision log" and "a decision log the harvest failed to read" look identical
	// in an empty array, and only one of them is the author's problem.
	Absent     []string `json:"absent_sources"`
	Signals    []string `json:"signals"`
	Note       string   `json:"note"`
	Unsignaled int      `json:"unsignaled_commits"`
}

// Range is the revision window the harvest covers. Revisions, not timestamps: a recipe
// is distilled from releases, and a release is a tag.
type Range struct {
	Since string `json:"since"`
	Until string `json:"until"`
}

// Release pairs a tag with the record its author wrote at the time — the closest thing
// to a design rationale that survives, and the reason a recipe is harvested rather than
// recalled.
type Release struct {
	Tag  string `json:"tag"`
	Body string `json:"body"`
}

// Candidate is a commit the harvest PROPOSES as raw material, with the signal that made
// it propose. The signal is a heuristic and is named as one: an author who trusts it
// blindly gets a recipe full of typo fixes, and an author who ignores it writes the last
// week from memory. Both failures are the author's to avoid; the machine's job is to put
// the evidence in front of them.
type Candidate struct {
	Hash    string   `json:"hash"`
	Subject string   `json:"subject"`
	Body    string   `json:"body,omitempty"`
	Signals []string `json:"signals"`
}

// scarSignals are the phrases a defect's own record uses when the defect was structural
// enough to be worth teaching. They are DECLARED here, in one table, rather than spread
// through the scanner — a heuristic that hides in code is a heuristic nobody can argue
// with. Extend the table when a real scar was missed; never to make a number look better.
var scarSignals = map[string]string{
	"sabotage":          "the defect was found only by deliberately breaking the thing",
	"found only by":     "the defect survived every green gate",
	"cry-wolf":          "a gate that reds on valid state — the failure mode of gates",
	"silently":          "a failure that produced no error, which is the dangerous kind",
	"exit 0":            "a broken thing that reported success",
	"the scar":          "the author named it a scar in the record",
	"would have":        "a counterfactual — the author reconstructed what the bug would have shipped",
	"reds every":        "a gate that punishes the innocent",
	"never fired":       "a gate that could not fire",
	"turns out":         "a belief that did not survive contact",
	"my own":            "the author's own rule, broken by the author",
	"the opposite of":   "a result that inverted its own intent",
	"a lie":             "a number or claim that was not what it appeared",
	"blind":             "a check that could not see the thing it guarded",
	"double-count":      "an arithmetic trap",
	"not a measurement": "a proxy wearing a measurement's name",
}

// Gather runs the harvest over a revision range in the given repository root, reading the
// records that project actually has. Nothing about which records those are is knowable from
// here — a service repository has no gate registry where a harness keeps one, and a project
// with no decision log is not a broken project — so every source arrives through the config
// seam and every missing one is reported rather than assumed empty.
func Gather(root, since, until string, cfg Config) (*Harvest, error) {
	if since == "" {
		return nil, fmt.Errorf("harvest needs a starting revision (--since): a recipe is distilled " +
			"from a bounded stretch of the record, and an unbounded one is the whole project's history")
	}
	if until == "" {
		until = "HEAD"
	}
	raw, err := gitLog(root, since, until)
	if err != nil {
		return nil, err
	}
	h, err := Parse(raw)
	if err != nil {
		return nil, err
	}
	h.Range = Range{Since: since, Until: until}
	h.Sources = cfg.Sources

	if cfg.Sources.Gates != "" {
		g, err := gatesAdded(root, since, until, cfg.Sources.Gates)
		if err != nil {
			return nil, err
		}
		// Assign only what was FOUND. A direct assignment lets a nil result overwrite the empty
		// slice Parse deliberately created, and the JSON then says `null` where it means `[]` —
		// "this was not looked at" in place of "this was looked at and there was nothing". That
		// is the silent-absence lie this whole module exists to refuse, and the harvest of the
		// reporting releases printed it at me the first time I ran the skill on a real range.
		if g != nil {
			h.Gates = g
		}
	} else {
		h.Absent = append(h.Absent, "gate registry — the recipe's `## Verification` section has no machine source; name the gates by hand or declare `recipe.sources.gates`")
	}
	for _, s := range []struct {
		dir, key string
		into     *[]string
		why      string
	}{
		{cfg.Sources.Doctrine, "doctrine", &h.Doctrines, "principles corpus — `## The principles` has no machine source"},
		{cfg.Sources.Decisions, "decisions", &h.Decisions, "decision log — the WHY behind the build is not in the record, only in commit subjects"},
		{cfg.Sources.Knowledge, "knowledge", &h.Knowledge, "knowledge base — `## The ground it needs` has no machine source"},
	} {
		if s.dir == "" {
			h.Absent = append(h.Absent, s.why)
			continue
		}
		t, err := touched(root, since, until, s.dir)
		if err != nil {
			return nil, err
		}
		if t != nil {
			*s.into = t
		}
	}
	if h.Releases, err = releases(root, since, until); err != nil {
		return nil, err
	}
	for phrase, why := range scarSignals {
		h.Signals = append(h.Signals, phrase+" — "+why)
	}
	h.Note = "candidates, not conclusions: the signals below PROPOSE raw material by matching " +
		"phrases in the record. Which failure was structural and which was incidental is the " +
		"author's judgment, and it is the whole value the author adds. Read the unsignaled " +
		"commits too — a scar nobody wrote down is still a scar, and it is exactly the one the " +
		"machine cannot hand you."
	return h, nil
}

// Parse is pure: the git grammar in, candidates out. Held to a fixture, so the awkward
// shapes (a body with blank lines, a subject that ignores the convention) are exercised
// without a repository.
func Parse(raw string) (*Harvest, error) {
	// Empty, not nil: a harvest that found nothing must SAY nothing-was-found (`[]`), and a
	// JSON `null` in its place reads as "this was not looked at" — the same lie a silent zero
	// tells in a report.
	h := &Harvest{
		Scars: []Candidate{}, Design: []Candidate{},
		Gates: []string{}, Doctrines: []string{}, Decisions: []string{}, Knowledge: []string{},
		Absent: []string{},
	}
	for _, chunk := range strings.Split(raw, "\x01") {
		if strings.TrimSpace(chunk) == "" {
			continue
		}
		head, _, ok := strings.Cut(chunk, "\x02")
		if !ok {
			return nil, fmt.Errorf("malformed git record (no header terminator): %.60q", chunk)
		}
		parts := strings.SplitN(head, "\x1f", 3)
		if len(parts) < 3 {
			return nil, fmt.Errorf("malformed git header (want 3 fields, got %d): %.60q", len(parts), head)
		}
		c := Candidate{Hash: parts[0], Subject: parts[1], Body: strings.TrimSpace(parts[2])}
		hay := strings.ToLower(parts[1] + "\n" + parts[2])
		for phrase := range scarSignals {
			if strings.Contains(hay, phrase) {
				c.Signals = append(c.Signals, phrase)
			}
		}
		typ, _, _ := strings.Cut(parts[1], "(")
		typ, _, _ = strings.Cut(typ, ":")
		typ = strings.TrimSpace(typ)
		if typ == "fix" {
			c.Signals = append(c.Signals, "type:fix")
		}
		// The two piles OVERLAP on purpose. A feature commit whose body confesses what broke
		// on the way is both the design decision AND the scar, and forcing it into one bucket
		// loses whichever question the author asks second. The first live harvest proved it:
		// every release commit in the range carried scar language, so an exclusive switch
		// reported ZERO design commits for three releases that were nothing but design.
		signaled := false
		if len(c.Signals) > 0 {
			h.Scars = append(h.Scars, c)
			signaled = true
		}
		if typ == "feat" {
			d := c
			d.Signals = append([]string{"type:feat"}, c.Signals...)
			h.Design = append(h.Design, d)
			signaled = true
		}
		if !signaled {
			h.Unsignaled++
		}
	}
	return h, nil
}

func gitLog(root, since, until string) (string, error) {
	cmd := exec.Command("git", "log", since+".."+until,
		"--no-merges", "--format=\x01%h\x1f%s\x1f%B\x02")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log %s..%s in %s: %w (is the starting revision a tag this "+
			"repository knows?)", since, until, root, err)
	}
	return string(out), nil
}

// gatesAdded reads the registry's diff rather than the registry's current state: a recipe
// must name the gates that shipped WITH the capability, and the ones that were already
// there are somebody else's story.
func gatesAdded(root, since, until, registry string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--unified=0", since+".."+until, "--", registry)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff of the gate registry: %w", err)
	}
	return ParseGatesDiff(string(out)), nil
}

// ParseGatesDiff is pure, and it is pure BECAUSE of how the first version failed: it looked
// for `- name:` rows, a grammar the registry does not use, so it returned an empty list on
// every input — forever, silently, and indistinguishably from "this release added no gates".
// A parser whose result can never be non-empty is a check that cannot fire, and the only
// thing that catches it is holding it to a fixture of the real grammar.
func ParseGatesDiff(diff string) []string {
	var added []string
	for _, l := range strings.Split(diff, "\n") {
		if !strings.HasPrefix(l, "+") || strings.HasPrefix(l, "+++") {
			continue
		}
		body := strings.TrimPrefix(l, "+")
		// A gate is a top-level key under `gates:` — two spaces of indent, then `name:` with
		// nothing after it. Its attributes (teeth:, where:, reason:) sit deeper and are not
		// gates.
		if !strings.HasPrefix(body, "  ") || strings.HasPrefix(body, "   ") {
			continue
		}
		k, rest, ok := strings.Cut(strings.TrimSpace(body), ":")
		if !ok || strings.TrimSpace(rest) != "" || k == "" || strings.HasPrefix(k, "#") {
			continue
		}
		added = append(added, k)
	}
	return added
}

func touched(root, since, until, dir string) ([]string, error) {
	cmd := exec.Command("git", "log", since+".."+until, "--name-only", "--format=", "--", dir)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log --name-only over %s: %w", dir, err)
	}
	seen := map[string]bool{}
	var files []string
	for _, l := range strings.Split(string(out), "\n") {
		l = strings.TrimSpace(l)
		if l == "" || seen[l] {
			continue
		}
		seen[l] = true
		files = append(files, l)
	}
	return files, nil
}

// releases lists the tags in the range and lifts each one's own section out of the
// changelog. The changelog is where a release states WHY, and why is the one thing a
// commit log is worst at carrying.
func releases(root, since, until string) ([]Release, error) {
	tags, err := tagsInRange(root, since, until)
	if err != nil || len(tags) == 0 {
		return nil, err
	}
	log, err := show(root, until, "CHANGELOG.md")
	if err != nil {
		// A project with no changelog is not a broken project. Absence is a state.
		return nil, nil
	}
	var out []Release
	for _, t := range tags {
		if body := section(log, t); body != "" {
			out = append(out, Release{Tag: t, Body: body})
		}
	}
	return out, nil
}

func tagsInRange(root, since, until string) ([]string, error) {
	before, err := tags(root, since)
	if err != nil {
		return nil, err
	}
	after, err := tags(root, until)
	if err != nil {
		return nil, err
	}
	had := map[string]bool{}
	for _, t := range before {
		had[t] = true
	}
	var out []string
	for _, t := range after {
		if !had[t] {
			out = append(out, t)
		}
	}
	return out, nil
}

func tags(root, rev string) ([]string, error) {
	cmd := exec.Command("git", "tag", "--merged", rev)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git tag --merged %s: %w", rev, err)
	}
	var ts []string
	for _, l := range strings.Split(string(out), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			ts = append(ts, l)
		}
	}
	return ts, nil
}

func show(root, rev, path string) (string, error) {
	cmd := exec.Command("git", "show", rev+":"+path)
	cmd.Dir = root
	out, err := cmd.Output()
	return string(out), err
}

// section lifts one release's block out of a keep-a-changelog body: from its own
// `## [tag]` heading to the next `## ` heading.
func section(changelog, tag string) string {
	lines := strings.Split(changelog, "\n")
	start := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "## ["+tag+"]") {
			start = i
			break
		}
	}
	if start < 0 {
		return ""
	}
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			return strings.TrimSpace(strings.Join(lines[start:i], "\n"))
		}
	}
	return strings.TrimSpace(strings.Join(lines[start:], "\n"))
}
