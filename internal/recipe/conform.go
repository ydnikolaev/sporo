package recipe

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// The conform layer — what turns `Binding: exact` from a sentence into a check.
//
// The gap it closes is the one a team hits in production: a fleet shares a recipe whose
// output feeds one consumer, every project "copies the shape byte-for-byte", and the first
// divergence is found by the consumer, live. The recipe carried the shape; nobody could RUN
// it. So: an exact-bound fence must be machine-readable, it may carry FIXTURES (a valid
// instance, and the invalid ones a consumer must reject), and `sporo conform` checks any
// candidate output against it — in the reader's CI, against the exported file they were
// handed, no checkout of the author's repository required.
//
// ADR-005 discipline, load-bearing: everything in this file activates only behind an
// author's `Binding: exact` declaration. An adapt-only recipe never meets any of it — no
// parse requirement, no fixtures, `conform` is an explicit no-op. The genre's flat core is
// guarded by tests that hold exactly that.
//
// What "conforms" means here — STRUCTURAL conformance to a shown example, not JSON Schema.
// The genre shows examples because a reader learns a shape from an instance, and forcing
// authors to write formal schemas would trade the document's readability for machinery.
// The example IS the contract, under these rules, recursively:
//
//   - object: the candidate carries EXACTLY the example's keys — a missing key starves the
//     consumer, an extra key is a dialect it never agreed to;
//   - array: every candidate element conforms to the example's first element (an empty
//     candidate array is legal — no items is not a shape violation);
//   - string `<like this>`: a placeholder — any scalar satisfies it;
//   - other string / number / bool: the candidate matches the TYPE (values are free);
//   - null: the field is declared nullable — anything, including null, satisfies it.
type Contract struct {
	// Index is the ordinal among the recipe's exact-bound shapes (1-based) — how `conform
	// --contract` addresses one when a recipe carries several.
	Index int
	Lang  string
	Body  string
	// Fixtures are the mini-TCK: instances the author stamps valid or invalid. Lint runs
	// them against the shape — a "valid" fixture that fails, or an "invalid" one that
	// passes, is a contract that cannot mean what its author thinks it means.
	Fixtures []Fixture
}

type Fixture struct {
	Valid bool
	Note  string
	Body  string
}

var (
	reFenceOpen = regexp.MustCompile("^\\s*```(\\w*)\\s*$")
	reFixture   = regexp.MustCompile(`\*\*Fixture: (valid|invalid)\*\*`)
)

// ExactContracts parses the exact-bound shapes (and their fixtures) out of a recipe's
// contracts section. It reads the same markers lint enforces, so the two cannot disagree
// about which fence is which.
func ExactContracts(src []byte) []Contract {
	con := sectionBody(strings.Split(string(src), "\n"), "## The contracts")
	var out []Contract
	var fenceBody []string
	binding := ""            // the marker most recently seen outside a fence
	fixture := (*Fixture)(nil) // pending fixture marker, if any
	role := ""               // what the OPEN fence is: "exact" | "fixture" | ""
	lang := ""
	inFence := false
	for _, l := range con {
		if m := reFenceOpen.FindStringSubmatch(l); m != nil && !inFence {
			inFence, lang, fenceBody = true, m[1], nil
			switch {
			case fixture != nil && len(out) > 0:
				role = "fixture"
			case binding == "exact":
				role = "exact"
			default:
				role = ""
			}
			continue
		}
		if inFence && reFence.MatchString(l) {
			switch role {
			case "exact":
				out = append(out, Contract{Index: len(out) + 1, Lang: lang, Body: strings.Join(fenceBody, "\n")})
			case "fixture":
				f := *fixture
				f.Body = strings.Join(fenceBody, "\n")
				last := &out[len(out)-1]
				last.Fixtures = append(last.Fixtures, f)
			}
			inFence, role, binding, fixture = false, "", "", nil
			continue
		}
		if inFence {
			fenceBody = append(fenceBody, l)
			continue
		}
		if reBinding.MatchString(l) {
			if strings.Contains(l, "**Binding: exact**") {
				binding = "exact"
			} else {
				binding = "adapt"
			}
			fixture = nil
			continue
		}
		if m := reFixture.FindStringSubmatch(l); m != nil {
			note := ""
			if i := strings.Index(l, "—"); i >= 0 {
				note = strings.TrimSpace(l[i+len("—"):])
			}
			fixture = &Fixture{Valid: m[1] == "valid", Note: note}
		}
	}
	return out
}

// Conform checks one candidate document against one exact contract. The violations name
// the path, because "does not conform" without a path is a verdict the reader has to
// re-derive by eye — the exact failure this layer exists to remove.
func Conform(c Contract, candidate []byte) ([]string, error) {
	example, err := parseShape(c.Lang, []byte(c.Body))
	if err != nil {
		return nil, fmt.Errorf("the contract's own example does not parse as %s — the recipe is broken, not your output: %w", langName(c.Lang), err)
	}
	cand, err := parseAny(candidate)
	if err != nil {
		return nil, fmt.Errorf("the candidate does not parse as JSON or YAML: %w", err)
	}
	var out []string
	shapeConform(example, cand, "$", &out)
	return out, nil
}

func shapeConform(example, cand any, path string, out *[]string) {
	switch ex := example.(type) {
	case nil:
		return // declared nullable — anything satisfies it
	case map[string]any:
		cm, ok := cand.(map[string]any)
		if !ok {
			*out = append(*out, fmt.Sprintf("%s: expected an object, got %s", path, typeName(cand)))
			return
		}
		var keys []string
		for k := range ex {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			cv, present := cm[k]
			if !present {
				*out = append(*out, fmt.Sprintf("%s.%s: missing — the consumer parses this field and will starve without it", path, k))
				continue
			}
			shapeConform(ex[k], cv, path+"."+k, out)
		}
		var extra []string
		for k := range cm {
			if _, known := ex[k]; !known {
				extra = append(extra, k)
			}
		}
		sort.Strings(extra)
		for _, k := range extra {
			*out = append(*out, fmt.Sprintf("%s.%s: not in the contract — an extra field is a dialect the consumer never agreed to", path, k))
		}
	case []any:
		ca, ok := cand.([]any)
		if !ok {
			*out = append(*out, fmt.Sprintf("%s: expected an array, got %s", path, typeName(cand)))
			return
		}
		if len(ex) == 0 {
			return
		}
		for i, item := range ca {
			shapeConform(ex[0], item, fmt.Sprintf("%s[%d]", path, i), out)
		}
	case string:
		if strings.HasPrefix(ex, "<") && strings.HasSuffix(ex, ">") {
			switch cand.(type) {
			case string, float64, int, bool:
				return
			default:
				*out = append(*out, fmt.Sprintf("%s: the placeholder admits any scalar, got %s", path, typeName(cand)))
			}
			return
		}
		if _, ok := cand.(string); !ok {
			*out = append(*out, fmt.Sprintf("%s: expected a string, got %s", path, typeName(cand)))
		}
	case bool:
		if _, ok := cand.(bool); !ok {
			*out = append(*out, fmt.Sprintf("%s: expected a boolean, got %s", path, typeName(cand)))
		}
	case float64, int:
		switch cand.(type) {
		case float64, int:
		default:
			*out = append(*out, fmt.Sprintf("%s: expected a number, got %s", path, typeName(cand)))
		}
	}
}

// fixtureFindings is the lint half of the layer: it fires ONLY on recipes that declared an
// exact contract (ADR-005), and it holds the declaration to its own consequences — the
// shape must parse, a valid fixture must conform, an invalid one must not.
func fixtureFindings(name string, src []byte) []Finding {
	var out []Finding
	for _, c := range ExactContracts(src) {
		if _, err := parseShape(c.Lang, []byte(c.Body)); err != nil {
			out = append(out, Finding{name, 0, fmt.Sprintf("exact contract #%d does not parse as %s — `exact` means a machine on the other side, and a shape no machine can read is a promise no machine can hold: %v", c.Index, langName(c.Lang), err)})
			continue
		}
		for i, f := range c.Fixtures {
			violations, err := Conform(c, []byte(f.Body))
			if err != nil {
				out = append(out, Finding{name, 0, fmt.Sprintf("fixture %d of exact contract #%d does not parse: %v", i+1, c.Index, err)})
				continue
			}
			switch {
			case f.Valid && len(violations) > 0:
				out = append(out, Finding{name, 0, fmt.Sprintf("a fixture stamped VALID fails its own contract (#%d) — the contract cannot mean what its author thinks it means: %s", c.Index, violations[0])})
			case !f.Valid && len(violations) == 0:
				out = append(out, Finding{name, 0, fmt.Sprintf("a fixture stamped INVALID conforms to contract #%d — it defends against nothing; make it violate the shape, or say why it should not exist", c.Index)})
			}
		}
	}
	return out
}

func parseShape(lang string, b []byte) (any, error) {
	if lang == "yaml" || lang == "yml" {
		var v any
		err := yaml.Unmarshal(b, &v)
		return normalizeYAML(v), err
	}
	var v any
	err := json.Unmarshal(b, &v)
	return v, err
}

// parseAny reads a candidate as JSON first (the overwhelmingly common feed format), then
// YAML — and says so when both fail, instead of guessing at the author's intent.
func parseAny(b []byte) (any, error) {
	var v any
	if err := json.Unmarshal(b, &v); err == nil {
		return v, nil
	}
	if err := yaml.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return normalizeYAML(v), nil
}

// normalizeYAML lifts yaml's map[any]any / int into the same vocabulary json produces, so
// shapeConform speaks one type language.
func normalizeYAML(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			t[k] = normalizeYAML(val)
		}
		return t
	case map[any]any:
		m := map[string]any{}
		for k, val := range t {
			m[fmt.Sprint(k)] = normalizeYAML(val)
		}
		return m
	case []any:
		for i, val := range t {
			t[i] = normalizeYAML(val)
		}
		return t
	case int:
		return float64(t)
	default:
		return v
	}
}

func typeName(v any) string {
	switch v.(type) {
	case nil:
		return "null"
	case map[string]any:
		return "an object"
	case []any:
		return "an array"
	case string:
		return "a string"
	case bool:
		return "a boolean"
	case float64, int:
		return "a number"
	default:
		return fmt.Sprintf("%T", v)
	}
}

func langName(l string) string {
	if l == "" {
		return "JSON (the fence names no language)"
	}
	return l
}
