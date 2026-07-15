package recipe

// A bundle is how "one document" survives scale. The genre's teeth — one acceptance, a
// `**Done when:**` per step, a ladder per precondition — only hold at the scale of ONE
// capability, so a whole harness or a whole repository is never one recipe. It is several,
// standing on shared ground — and the promise "the reader gets a single document" is kept by
// the DELIVERY step, which composes the members in build order under one adoption protocol,
// not by inflating the authoring genre until its acceptance checks stop meaning anything.
//
// The manifest lives beside the recipes it composes (`<name>.bundle.yaml` in the recipes
// home) and is deliberately small: a title, an optional shared-ground preamble, and the
// members in the order they must be built. Order is not decoration — a member that stands on
// another's output belongs after it, and the manifest is where that knowledge lives once,
// instead of in the memory of whoever exported last.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Bundle is one composition manifest.
type Bundle struct {
	// Bundle is the name — the slug the export verb takes.
	Bundle string `yaml:"bundle"`
	Title  string `yaml:"title"`
	// Ground is the optional shared preamble: what ALL members stand on, stated once, so
	// the composed document does not repeat one ladder per member.
	Ground string `yaml:"ground"`
	// Members are recipe slugs, in build order.
	Members []string `yaml:"members"`
}

const bundleSuffix = ".bundle.yaml"

// LoadBundle reads one manifest from the project's recipes home.
func LoadBundle(home, name string) (Bundle, error) {
	var b Bundle
	data, err := os.ReadFile(filepath.Join(home, name+bundleSuffix))
	if err != nil {
		return b, fmt.Errorf("no bundle %q in this project's home — a bundle is a manifest (%s%s) beside the recipes it composes: %w", name, name, bundleSuffix, err)
	}
	if err := yaml.Unmarshal(data, &b); err != nil {
		return b, fmt.Errorf("bundle %q is malformed — fix the YAML: %w", name, err)
	}
	return b, nil
}

// LintBundle holds a manifest to the only rules it has: a title, at least one member, no
// duplicates, and every member resolvable — in the project home or the official corpus. A
// bundle naming a recipe that does not exist would compose fine today and hand a stranger a
// build order with a hole in it the day the missing member was meant to arrive.
func LintBundle(corpus fs.FS, home, name string, b Bundle) []Finding {
	var out []Finding
	file := name + bundleSuffix
	if strings.TrimSpace(b.Title) == "" {
		out = append(out, Finding{file, 0, "bundle has no title — the composed document opens with it, and an untitled handover reads as a fragment"})
	}
	if len(b.Members) == 0 {
		out = append(out, Finding{file, 0, "bundle has no members — a composition of nothing is a manifest for a document that cannot exist"})
	}
	seen := map[string]bool{}
	for _, m := range b.Members {
		if seen[m] {
			out = append(out, Finding{file, 0, fmt.Sprintf("member %q is listed twice — the build order must say each thing once", m)})
			continue
		}
		seen[m] = true
		if !memberExists(corpus, home, m) {
			out = append(out, Finding{file, 0, fmt.Sprintf("member %q resolves to no recipe here or in the official corpus — a build order with a hole in it transfers the hole", m)})
		}
	}
	return out
}

func memberExists(corpus fs.FS, home, slug string) bool {
	if home != "" {
		if _, err := os.Stat(filepath.Join(home, slug+".md")); err == nil {
			return true
		}
	}
	if _, err := fs.Stat(corpus, "recipes/"+slug+".md"); err == nil {
		return true
	}
	return false
}

// ExportBundle composes the single handed-over document: the bundle's title and shared
// ground, then every member in build order (banner stripped, exactly as authored), then ONE
// adoption protocol at the end. The members' own would-be protocols are not duplicated —
// the protocol is addressed to the reader, and the reader is one agent doing one adoption,
// however many capabilities it spans.
func ExportBundle(corpus fs.FS, home, name string) (string, error) {
	b, err := LoadBundle(home, name)
	if err != nil {
		return "", err
	}
	if f := LintBundle(corpus, home, name, b); len(f) > 0 {
		msgs := make([]string, len(f))
		for i, x := range f {
			msgs[i] = x.Msg
		}
		return "", fmt.Errorf("bundle %q does not compose: %s", name, strings.Join(msgs, "; "))
	}
	protocol, err := adoption(corpus)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("# " + strings.TrimSpace(b.Title) + "\n\n")
	sb.WriteString(fmt.Sprintf("*This document composes %d recipes into one build, in the order they must be "+
		"built. Each part carries the full genre — its own acceptance, contracts, scars. The adoption "+
		"protocol at the very end applies to the whole.*\n", len(b.Members)))
	if g := strings.TrimSpace(b.Ground); g != "" {
		sb.WriteString("\n## The shared ground\n\n" + g + "\n")
	}
	for i, m := range b.Members {
		body, err := memberBody(corpus, home, m)
		if err != nil {
			return "", err
		}
		sb.WriteString(fmt.Sprintf("\n---\n\n<!-- part %d of %d -->\n\n", i+1, len(b.Members)))
		sb.WriteString(strings.TrimRight(body, "\n") + "\n")
	}
	sb.WriteString(protocol)
	return sb.String(), nil
}

// memberBody reads one member the way Export does — project home first, official corpus
// second, banner stripped — but WITHOUT the protocol, which the bundle appends once.
func memberBody(corpus fs.FS, home, slug string) (string, error) {
	if home != "" {
		if b, err := os.ReadFile(filepath.Join(home, slug+".md")); err == nil {
			return strip(string(b)), nil
		}
	}
	b, err := fs.ReadFile(corpus, "recipes/"+slug+".md")
	if err != nil {
		return "", fmt.Errorf("bundle member %q vanished between lint and read: %w", slug, err)
	}
	return strip(string(b)), nil
}

// Bundles enumerates the manifests in a home, for lint's corpus walk.
func Bundles(home string) []string {
	var out []string
	if home == "" {
		return out
	}
	ents, err := os.ReadDir(home)
	if err != nil {
		return out
	}
	for _, e := range ents {
		if !e.IsDir() && strings.HasSuffix(e.Name(), bundleSuffix) {
			out = append(out, strings.TrimSuffix(e.Name(), bundleSuffix))
		}
	}
	return out
}
