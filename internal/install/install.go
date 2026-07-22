// Package install owns `sporo init` and `sporo update` — the two verbs that put sporo's
// authoring surface INTO a repository, and keep it current without ever eating a user's edit.
//
// The model (and why it is not "just copy files"):
//
//   - The agent runtime loads skills only from the filesystem — provider homes like
//     `.claude/` or an `AGENTS.md` at the root — never from inside a binary. So the canonical
//     skill is embedded here (SSOT, version-locked to the binary) and `init` WRITES IT OUT.
//   - What sporo wrote, sporo may rewrite; what the user wrote, sporo may only guard. The
//     registry records every managed file with the hash AS WRITTEN. `update` rewrites a file
//     only while its on-disk content still matches that hash; an edited file is REPORTED and
//     skipped, every time, with no flag to force it. (A fix belongs upstream — in the skill
//     the binary ships — not in a fork inside one repo that the next update would eat.)
//   - Provider placement is an adapter, and a provider fact is verified, then written down:
//     Claude Code loads `.claude/skills/<name>/SKILL.md` (verified against Claude Code docs,
//     2026-07); Codex and most other agents read the repo-root `AGENTS.md` (the cross-agent
//     convention, agents.md, verified 2026-07). The claude adapter activates only when the
//     repo already has a `.claude/` home — creating one in a repo that does not use the
//     provider is litter; the AGENTS.md block is universal, so it is always managed.
//
// `init` = the one-time seeds (a config, a recipes-home README — written if absent, then
// YOURS, never touched again) + the same sync `update` runs. Running `init` twice is safe by
// construction; that property is tested, not promised.
package install

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sporo.dev/sporo/internal/recipe"
)

//go:embed assets/SKILL.md assets/SKILL-seed.md assets/agents-block.md
var assets embed.FS

// skills is the authoring surface `init`/`update` write into a `.claude/` home — one per genre.
// Both are managed the same way (the no-clobber state machine below); a genre is added by adding a
// row here and its embedded asset, not by touching the sync logic.
var skills = []struct {
	rel   string // where it lands in the consumer repo
	asset string // the embedded source in this binary
}{
	{".claude/skills/sporo-recipe/SKILL.md", "assets/SKILL.md"},
	{".claude/skills/sporo-seed/SKILL.md", "assets/SKILL-seed.md"},
}

// Action is one file's outcome, reported to the user. Update prints these; nothing here is a
// silent side effect.
type Action struct {
	Path   string
	Status string // wrote | updated | kept | seeded | skipped
	Note   string
}

const (
	agentsFile = "AGENTS.md"
	blockBegin = "<!-- sporo:begin — managed by `sporo update`; edits inside this block are reported, never clobbered -->"
	blockEnd   = "<!-- sporo:end -->"
)

// Init seeds the project (config, recipes home) and installs the authoring surface. The
// seeds are write-if-absent and never managed afterwards: the config is the project's own
// voice, and a tool that rewrites it on update is a tool that argues with its user.
func Init(root, version string) ([]Action, error) {
	var out []Action

	name := "this-project"
	if abs, err := filepath.Abs(root); err == nil {
		name = filepath.Base(abs)
	}
	out = append(out, seed(root, ".sporo/config.yaml", configSeed(name)))
	out = append(out, seed(root, ".sporo/recipes/README.md", readmeSeed))

	synced, err := sync(root, version)
	if err != nil {
		return out, err
	}
	return append(out, synced...), nil
}

// Update re-syncs the managed surface from the (possibly newer) binary's embedded copy.
// On a repository that was never initialized it refuses with directions rather than
// half-installing: `init` states the intent, `update` maintains it.
func Update(root, version string) ([]Action, error) {
	if _, err := os.Stat(filepath.Join(root, ".sporo")); err != nil {
		return nil, fmt.Errorf("nothing to update here — this repository was never initialized (`sporo init` installs the authoring surface first)")
	}
	return sync(root, version)
}

// sync makes the managed files match the binary, under the no-clobber rule. It is the whole
// update mechanism, and init calls the same code so the two can never disagree about what
// "managed" means.
func sync(root, version string) ([]Action, error) {
	reg, err := recipe.LoadRegistry(root)
	if err != nil {
		return nil, err
	}
	managed := map[string]recipe.Managed{}
	for _, m := range reg.Managed {
		managed[m.Path] = m
	}

	var out []Action

	// The claude adapter: only where the provider home already exists. Every genre's skill is
	// written the same way, in listed order.
	if _, err := os.Stat(filepath.Join(root, ".claude")); err == nil {
		for _, sk := range skills {
			a, err := syncFile(root, sk.rel, skillContent(sk.asset, version), version, managed)
			if err != nil {
				return out, err
			}
			out = append(out, a)
		}
	}

	// The AGENTS.md block: universal, so always managed. The unit is the BLOCK, not the
	// file — the file is the user's, and sporo owns only the region between its markers.
	a, err := syncBlock(root, version, managed)
	if err != nil {
		return out, err
	}
	out = append(out, a)

	reg.Managed = reg.Managed[:0]
	for _, m := range managed {
		reg.Managed = append(reg.Managed, m)
	}
	sortManaged(reg.Managed)
	if err := reg.Save(root); err != nil {
		return out, err
	}
	return out, nil
}

// syncFile is the no-clobber state machine for one whole managed file.
func syncFile(root, rel, desired, version string, managed map[string]recipe.Managed) (Action, error) {
	abs := filepath.Join(root, rel)
	disk, readErr := os.ReadFile(abs)
	entry, tracked := managed[rel]

	write := func(status string) (Action, error) {
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return Action{}, err
		}
		if err := os.WriteFile(abs, []byte(desired), 0o644); err != nil {
			return Action{}, err
		}
		managed[rel] = recipe.Managed{Path: rel, Hash: recipe.ContentHash([]byte(desired)), Binary: version}
		return Action{rel, status, ""}, nil
	}

	switch {
	case readErr != nil:
		return write("wrote")
	case tracked && recipe.ContentHash(disk) == entry.Hash:
		if string(disk) == desired {
			return Action{rel, "kept", "already current"}, nil
		}
		return write("updated")
	case tracked:
		return Action{rel, "skipped", "edited since sporo wrote it — your copy is preserved; the managed original moved on with the binary"}, nil
	case string(disk) == desired:
		// On disk, identical, but not in the registry — a repo restored from a copy, most
		// likely. Adopting it is safe precisely because the content is byte-identical.
		managed[rel] = recipe.Managed{Path: rel, Hash: recipe.ContentHash(disk), Binary: version}
		return Action{rel, "kept", "adopted (identical, untracked)"}, nil
	default:
		return Action{rel, "skipped", "exists but was not written by sporo — refusing to overwrite a file it cannot vouch for"}, nil
	}
}

// syncBlock manages the sporo region inside AGENTS.md. Same rules as a whole file, scoped to
// the block: an edit INSIDE the markers is reported and preserved; the user's text outside
// the markers is never touched at all.
func syncBlock(root, version string, managed map[string]recipe.Managed) (Action, error) {
	desired := agentsBlockContent(version)
	abs := filepath.Join(root, agentsFile)
	disk, readErr := os.ReadFile(abs)
	entry, tracked := managed[agentsFile]

	record := func(block string) {
		managed[agentsFile] = recipe.Managed{Path: agentsFile, Hash: recipe.ContentHash([]byte(block)), Binary: version}
	}

	if readErr != nil {
		if err := os.WriteFile(abs, []byte(desired+"\n"), 0o644); err != nil {
			return Action{}, err
		}
		record(desired)
		return Action{agentsFile, "wrote", ""}, nil
	}

	current, before, after, found := extractBlock(string(disk))
	switch {
	case !found:
		body := strings.TrimRight(string(disk), "\n") + "\n\n" + desired + "\n"
		if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
			return Action{}, err
		}
		record(desired)
		return Action{agentsFile, "updated", "sporo block appended below your content"}, nil
	case tracked && recipe.ContentHash([]byte(current)) == entry.Hash:
		if current == desired {
			return Action{agentsFile, "kept", "block already current"}, nil
		}
		if err := os.WriteFile(abs, []byte(before+desired+after), 0o644); err != nil {
			return Action{}, err
		}
		record(desired)
		return Action{agentsFile, "updated", "sporo block refreshed; your surrounding content untouched"}, nil
	case tracked:
		return Action{agentsFile, "skipped", "the sporo block was edited — your copy is preserved; the managed original moved on with the binary"}, nil
	case current == desired:
		record(current)
		return Action{agentsFile, "kept", "block adopted (identical, untracked)"}, nil
	default:
		return Action{agentsFile, "skipped", "a sporo block exists that sporo did not write — refusing to overwrite it"}, nil
	}
}

// extractBlock splits a file around the sporo block. `before` ends where the block begins and
// `after` starts where it ends, so `before + block + after` reassembles the file exactly.
func extractBlock(s string) (block, before, after string, found bool) {
	i := strings.Index(s, blockBegin)
	if i < 0 {
		return "", "", "", false
	}
	j := strings.Index(s[i:], blockEnd)
	if j < 0 {
		return "", "", "", false
	}
	end := i + j + len(blockEnd)
	return s[i:end], s[:i], s[end:], true
}

func seed(root, rel, content string) Action {
	abs := filepath.Join(root, rel)
	if _, err := os.Stat(abs); err == nil {
		return Action{rel, "kept", "already exists — seeds are written once and never touched again"}
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return Action{rel, "skipped", err.Error()}
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return Action{rel, "skipped", err.Error()}
	}
	return Action{rel, "seeded", ""}
}

// skillContent is one embedded skill with the provenance stamp spliced in AFTER the
// frontmatter — a comment above it would break the provider's frontmatter parsing, which
// expects the document to open with `---`.
func skillContent(asset, version string) string {
	b, err := assets.ReadFile(asset)
	if err != nil {
		panic("the embedded skill is missing from the binary: " + err.Error()) // a build defect, not a runtime state
	}
	stamp := fmt.Sprintf("<!-- SYNCED FROM sporo@%s — managed by `sporo update`: if you edit this copy, update will report it and leave it alone; the maintained original ships with the binary. -->", version)
	s := string(b)
	if i := strings.Index(s[3:], "\n---\n"); strings.HasPrefix(s, "---\n") && i >= 0 {
		at := 3 + i + len("\n---\n")
		return s[:at] + "\n" + stamp + "\n" + s[at:]
	}
	return stamp + "\n\n" + s
}

func agentsBlockContent(version string) string {
	b, err := assets.ReadFile("assets/agents-block.md")
	if err != nil {
		panic("the embedded agents block is missing from the binary: " + err.Error())
	}
	return blockBegin + "\n<!-- sporo@" + version + " -->\n\n" + strings.TrimRight(string(b), "\n") + "\n\n" + blockEnd
}

func configSeed(project string) string {
	return fmt.Sprintf(`# .sporo/config.yaml — this project's view of the recipe tool.
# Seeded once by `+"`sporo init`"+`; from here on it is YOURS — sporo never rewrites it.

# The one name most likely to leak into a recipe written here, banned from every body line.
# A fleet or a monorepo lists its siblings under `+"`products:`"+` instead.
project: %s

# Where this project's own recipes are authored (the default shown).
# home: .sporo/recipes/

# The records `+"`sporo harvest`"+` reads besides git. Declared beats probed; an absent
# source is a stated absence in the harvest, never a silent zero.
# sources:
#   gates:
#   doctrine:
#   decisions:
#   knowledge:
`, project)
}

const readmeSeed = `# Recipes

This project's own recipes live here — one self-contained document per capability, written
so an agent in a repository that has never seen this one can rebuild it. Authored via the
sporo-recipe skill; gated by ` + "`sporo lint`" + `; sealed (` + "`sporo seal`" + `) so a
finished recipe never silently mutates; handed over ONLY as ` + "`sporo export <slug>`" + `
prints it — the export carries the adoption protocol, the source file does not.
`

func sortManaged(m []recipe.Managed) {
	for i := 1; i < len(m); i++ {
		for j := i; j > 0 && m[j].Path < m[j-1].Path; j-- {
			m[j], m[j-1] = m[j-1], m[j]
		}
	}
}
