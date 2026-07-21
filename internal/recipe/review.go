package recipe

// The semantic half of the gate. `lint` proves a recipe's SHAPE; nothing mechanical can
// prove its VALUE — whether an agent in a foreign repository could actually build from it.
// That judgment needs a second reader, and the design constraint is that sporo must not
// care which one: not everyone has every provider installed, paid for, or permitted.
//
// So the review is a PACK, not a call. `review` composes one self-contained prompt file —
// the rubric, the verdict schema, and the exported recipe, all inline — and any agent the
// user has renders the verdict: pipe the file to one provider's CLI, paste it into another's
// chat, run three and compare. sporo never invokes a provider, which is why it cannot break
// when one changes its flags, its pricing, or its mind. `review verify` closes the loop: it
// validates the returned JSON against the same fixed shape and records the tally beside the
// recipe's seal.
//
// The seven axes are not invented: they are the axes the genre's first outside implementer
// actually scored (and the two lowest scores — copy-paste artifacts, security hazards —
// each produced a section of the genre spec). The rubric institutionalizes that first review.

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"sporo.dev/sporo/pkg/recipekit"
)

var reviewAxes = []struct{ key, def string }{
	{"intent_clarity", "Does the document teach WHY this capability exists and what problem it solves, well enough to re-decide trade-offs the author did not anticipate?"},
	{"scars_value", "Do the scars describe real, specific failures (symptom → root cause → fix) that would genuinely save a builder from repeating them?"},
	{"task_setting", "Is the reader's task well-posed — clear acceptance at the top, a Done-when on every step, no step whose completion is a matter of opinion?"},
	{"build_readiness", "Could an agent start building from this document alone — are the sequence, the ground ladder, and the preconditions actionable rather than aspirational?"},
	{"copy_paste_artifacts", "Are the shapes the capability consumes and emits SHOWN (schemas, contracts, examples a reader copies and adapts) rather than described in prose?"},
	{"stack_neutrality", "Is the body free of coordinates (paths, filenames, products) while still naming technologies with reasons — could this execute anywhere?"},
	{"security_hazards", "Does it address the hazards its genre inherits (escaping rendered text, time boundaries, joined sources, loud toolchain failure) where they apply?"},
}

const (
	recipeBegin = "===== RECIPE UNDER REVIEW (begin) ====="
	recipeEnd   = "===== RECIPE UNDER REVIEW (end) ====="
)

// BuildReviewPack writes `.sporo/review/<slug>/prompt.md` (self-contained: rubric + schema +
// the exported recipe) and `verdict.schema.json` beside it, and returns the prompt's path.
//
// The recipe is embedded between literal marker lines, not inside a fence — a recipe
// contains fences of its own, and nesting them corrupts the first one that closes.
func BuildReviewPack(corpus fs.FS, root string, cfg Config, slug string) (string, error) {
	home := filepath.Join(root, cfg.Home)
	if _, err := os.Stat(home); err != nil {
		home = ""
	}
	body, err := Export(corpus, home, slug)
	if err != nil {
		return "", err
	}
	version := recipeVersion(corpus, root, cfg, slug)

	dir := filepath.Join(root, ".sporo", "review", slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	schema := verdictSchema()
	if err := os.WriteFile(filepath.Join(dir, "verdict.schema.json"), []byte(schema), 0o644); err != nil {
		return "", err
	}
	prompt := reviewPrompt(slug, version, schema, body)
	promptPath := filepath.Join(dir, "prompt.md")
	if err := os.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
		return "", err
	}
	return promptPath, nil
}

// reviewPrompt is the whole pack, in one file, because the reviewer may be an agent with no
// filesystem: everything it needs must survive a copy-paste.
func reviewPrompt(slug, version, schema, body string) string {
	var b strings.Builder
	b.WriteString("# Recipe review — rubric and verdict\n\n")
	b.WriteString("You are an experienced engineer reviewing a **recipe**: a transferable build document\n")
	b.WriteString("that teaches an agent in a repository it has never seen how to build a capability —\n")
	b.WriteString("possibly on a different stack. You are NOT reviewing the capability itself; you are\n")
	b.WriteString("judging whether a stranger could build it FROM THIS DOCUMENT ALONE. Assume no access\n")
	b.WriteString("to the author, the author's repository, or any prior context.\n\n")
	b.WriteString("Score each axis 0–10 with a one-sentence note that names the strongest evidence for\n")
	b.WriteString("your score (a section, a gap, a specific sentence). Harsh and specific beats kind and\n")
	b.WriteString("vague: a low score with a named gap is what improves the next version.\n\n")
	b.WriteString("## The seven axes\n\n")
	for _, a := range reviewAxes {
		fmt.Fprintf(&b, "- **%s** — %s\n", a.key, a.def)
	}
	b.WriteString("\n## The verdict\n\n")
	fmt.Fprintf(&b, "Return ONLY a JSON object — no prose around it — for recipe `%s`, version `%s`,\n", slug, version)
	b.WriteString("matching this schema exactly. The object is your ENTIRE response: print it, or write\n")
	b.WriteString("it to the file whoever handed you this pack asked for — the channel is theirs, the\n")
	b.WriteString("bytes are yours. `verdict` is `adopt` if an agent could build from the document today;\n")
	b.WriteString("`revise` if any axis scores 4 or below, or if any note names a gap that would stop a\n")
	b.WriteString("build regardless of its score — when in doubt, revise. `top_gaps` is the ordered\n")
	b.WriteString("shortlist of what to fix first, empty only if nothing needs fixing:\n\n")
	b.WriteString("```json\n" + schema + "```\n\n")
	b.WriteString(recipeBegin + "\n\n")
	b.WriteString(strings.TrimRight(body, "\n") + "\n\n")
	b.WriteString(recipeEnd + "\n")
	return b.String()
}

// verdictSchema is generated from the axes list rather than written beside it, so the rubric
// the reviewer reads, the schema they fill, and the validator that judges the result cannot
// drift apart — one list, three surfaces.
func verdictSchema() string {
	var axes strings.Builder
	keys := make([]string, len(reviewAxes))
	for i, a := range reviewAxes {
		keys[i] = fmt.Sprintf("%q", a.key)
		fmt.Fprintf(&axes, "      %q: { \"type\": \"object\", \"required\": [\"score\", \"note\"], \"additionalProperties\": false,\n        \"properties\": { \"score\": { \"type\": \"integer\", \"minimum\": 0, \"maximum\": 10 }, \"note\": { \"type\": \"string\", \"minLength\": 1 } } }", a.key)
		if i < len(reviewAxes)-1 {
			axes.WriteString(",")
		}
		axes.WriteString("\n")
	}
	return fmt.Sprintf(`{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "sporo review verdict",
  "type": "object",
  "required": ["schema", "recipe", "version", "axes", "verdict", "top_gaps"],
  "additionalProperties": false,
  "properties": {
    "schema": { "const": 1 },
    "recipe": { "type": "string" },
    "version": { "type": "string", "pattern": "^\\d+\\.\\d+\\.\\d+$" },
    "axes": {
      "type": "object",
      "required": [%s],
      "additionalProperties": false,
      "properties": {
%s      }
    },
    "verdict": { "enum": ["adopt", "revise"] },
    "top_gaps": { "type": "array", "items": { "type": "string" } }
  }
}
`, strings.Join(keys, ", "), axes.String())
}

// AxisScore is one axis of one verdict.
type AxisScore struct {
	Score int    `json:"score"`
	Note  string `json:"note"`
}

// Verdict is what a reviewer returns. The shape is OURS and fixed, so it is validated with
// the standard decoder and explicit checks — a schema library would be a dependency spent
// re-checking a struct Go already checks.
type Verdict struct {
	Schema  int                  `json:"schema"`
	Recipe  string               `json:"recipe"`
	Version string               `json:"version"`
	Axes    map[string]AxisScore `json:"axes"`
	Call    string               `json:"verdict"`
	TopGaps []string             `json:"top_gaps"`
}

// VerifyVerdicts validates returned verdict files and, when every one is valid, records the
// tally beside the recipe's seal. An UNSEALED recipe is refused: a review binds a version and
// a content hash, and a draft has neither — seal first, then review what you sealed.
func VerifyVerdicts(root string, cfg Config, slug string, files []string) (ReviewSummary, []Finding, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return ReviewSummary{}, nil, err
	}
	entry, sealed := reg.Recipes[slug]
	if !sealed {
		return ReviewSummary{}, nil, fmt.Errorf("recipe %q is not sealed — a review binds a version and a content hash, and a draft has neither; `sporo seal %s`, then review what you sealed", slug, slug)
	}
	// The registry is one slug-keyed map across kinds; `review verify` is the RECIPE review path,
	// so a slug sealed as another genre is refused pointing at that genre's tooling — the last of
	// the four read-side kind guards (INV-1).
	if entry.Kind != recipekit.KindRecipe {
		return ReviewSummary{}, nil, fmt.Errorf("%q is sealed as a %s, not a recipe — `review verify` is the recipe review path; a %s is reviewed through the seed tooling, not here", slug, entry.Kind, entry.Kind)
	}

	var findings []Finding
	var verdicts []Verdict
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			return ReviewSummary{}, nil, err
		}
		v, errs := parseVerdict(b, slug, entry.Version)
		for _, msg := range errs {
			findings = append(findings, Finding{File: file, Msg: msg})
		}
		if len(errs) == 0 {
			verdicts = append(verdicts, v)
		}
	}
	if len(findings) > 0 {
		return ReviewSummary{}, findings, nil
	}
	if len(verdicts) == 0 {
		return ReviewSummary{}, nil, fmt.Errorf("no verdicts to verify — the pack is `.sporo/review/%s/prompt.md`; run it through any agent and hand the JSON back here", slug)
	}

	sum := ReviewSummary{Date: time.Now().Format("2006-01-02"), Version: entry.Version, Verdicts: len(verdicts), Verdict: "adopt"}
	total := 0.0
	for _, v := range verdicts {
		s := 0
		for _, a := range v.Axes {
			s += a.Score
		}
		total += float64(s) / float64(len(reviewAxes))
		if v.Call != "adopt" {
			// One reviewer saying "revise" outweighs two saying "adopt": the aggregate is a
			// gate, and a gate that averages away a named blocking gap is not one.
			sum.Verdict = "revise"
		}
	}
	sum.Mean = math.Round(total/float64(len(verdicts))*10) / 10

	if reg.Reviews == nil {
		reg.Reviews = map[string][]ReviewSummary{}
	}
	reg.Reviews[slug] = append(reg.Reviews[slug], sum)
	if err := reg.Save(root); err != nil {
		return ReviewSummary{}, nil, err
	}
	return sum, nil, nil
}

// parseVerdict validates one verdict against the same rules the schema states. Every
// violation is reported, not just the first — a reviewer fixing their output should fix it
// once.
func parseVerdict(b []byte, slug, version string) (Verdict, []string) {
	var v Verdict
	dec := json.NewDecoder(strings.NewReader(string(b)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return v, []string{"not a verdict: " + err.Error()}
	}
	var errs []string
	if v.Schema != 1 {
		errs = append(errs, fmt.Sprintf("schema %d, want 1", v.Schema))
	}
	if v.Recipe != slug {
		errs = append(errs, fmt.Sprintf("verdict is about %q, not %q — the wrong document was reviewed, or the wrong verdict handed in", v.Recipe, slug))
	}
	if v.Version != version {
		errs = append(errs, fmt.Sprintf("verdict reviews version %s but the seal says %s — the review must bind to the text it read; rebuild the pack and re-review", v.Version, version))
	}
	if v.Call != "adopt" && v.Call != "revise" {
		errs = append(errs, fmt.Sprintf("verdict %q — adopt or revise, nothing in between (a hedge is a revise)", v.Call))
	}
	for _, a := range reviewAxes {
		ax, ok := v.Axes[a.key]
		switch {
		case !ok:
			errs = append(errs, fmt.Sprintf("axis %q is missing — a skipped axis is an unexamined failure mode", a.key))
		case ax.Score < 0 || ax.Score > 10:
			errs = append(errs, fmt.Sprintf("axis %q scored %d — the scale is 0–10", a.key, ax.Score))
		case strings.TrimSpace(ax.Note) == "":
			errs = append(errs, fmt.Sprintf("axis %q has no note — a score with no evidence cannot be argued with, which makes it useless", a.key))
		}
	}
	for key := range v.Axes {
		known := false
		for _, a := range reviewAxes {
			if a.key == key {
				known = true
				break
			}
		}
		if !known {
			errs = append(errs, fmt.Sprintf("unknown axis %q — the rubric has exactly seven", key))
		}
	}
	return v, errs
}

// recipeVersion reads the version the pack is reviewing — from the project's own copy when
// it has one, else from the shipped corpus. Missing entirely degrades to "unversioned" in
// the prompt rather than failing the pack: an old recipe is still reviewable, and the
// verify step will hold the verdict to the seal anyway.
func recipeVersion(corpus fs.FS, root string, cfg Config, slug string) string {
	if b, err := os.ReadFile(filepath.Join(root, cfg.Home, slug+".md")); err == nil {
		if v := fmValue(b, "version"); v != "" {
			return v
		}
	}
	if b, err := fs.ReadFile(corpus, path.Join("recipes", slug+".md")); err == nil {
		if v := fmValue(b, "version"); v != "" {
			return v
		}
	}
	return "unversioned"
}
