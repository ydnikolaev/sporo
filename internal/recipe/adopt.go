package recipe

// The reader's half of version distribution — the piece the loop was missing. An author
// seals, exports and hands over; a reader builds; and then the recipe moves on (a new scar,
// a new MAJOR on an exact contract) while the reader's copy sits still, silently stale. The
// author cannot fix this: they do not know who adopted. Only the reader's own repository can
// remember what it built from and where that text came from — so that is exactly what
// `adopt` records, and what `pull` re-checks.
//
// Two disciplines, both deliberate:
//
//   - `adopt` stores the EXPORTED file verbatim, not a normalized form. That file is the
//     thing the reader actually built from; its hash is the only honest anchor for "did the
//     source change". Same bytes re-adopt as a no-op; different bytes under a version that
//     did not increase are refused with the same posture as the seal — a handed-over text
//     never silently mutates on either end of the handover.
//   - `pull` is READ-ONLY unless told otherwise. Discovering that the source moved on is
//     cheap and safe; acting on it is a rebuild — agent work, judged against this repo —
//     and a tool that overwrote the reader's record as a side effect of *checking* would be
//     making that judgment for them. `--apply` is the explicit second step.
//
// Staleness deliberately does NOT ride `sporo lint`: the gate must run offline (a red gate
// that needs the network is a gate that reds on airplanes), and a pull needs the network by
// definition. The verb is the check.

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AdoptedEntry is one handed-over recipe this repository built from.
type AdoptedEntry struct {
	// Version is the frontmatter version of the text the reader actually received —
	// what every report-back and every staleness comparison anchors to.
	Version string `yaml:"version"`
	// Hash is ContentHash of the stored file, so a changed source is detected on bytes,
	// never on trust.
	Hash string `yaml:"hash"`
	// ExactContracts mirrors the seal's digest of the exact-bound fences. When a pull sees
	// THIS change, the update is not prose — somebody's parser moved, and the reader's
	// consumer-facing output must be re-verified.
	ExactContracts string `yaml:"exact_contracts,omitempty"`
	Date           string `yaml:"date"`
	// Source is where `pull` re-fetches from: a filesystem path or an http(s) URL. Recorded
	// as given — the reader's own convention, not normalized behind their back.
	Source string `yaml:"source,omitempty"`
}

// adoptedHome is where the verbatim copies live, beside (not inside) the authoring home:
// an adopted recipe is somebody else's text and must never be swept into this project's
// own lint/seal/export surface.
const adoptedHome = ".sporo/adopted"

// Adopt records a handed-over exported recipe: verbatim copy + registry entry.
func Adopt(root string, src []byte, source string) (string, AdoptedEntry, error) {
	slug := fmValue(src, "name")
	version := fmValue(src, "version")
	if slug == "" || version == "" {
		return "", AdoptedEntry{}, fmt.Errorf("this file carries no `name:`/`version:` frontmatter — every real export does; adopt the file `sporo export` printed, not a fragment of it")
	}

	dir := filepath.Join(root, adoptedHome)
	path := filepath.Join(dir, slug+".md")
	entry := AdoptedEntry{
		Version:        version,
		Hash:           ContentHash(src),
		ExactContracts: exactContractsDigest(src),
		Date:           time.Now().Format("2006-01-02"),
		Source:         source,
	}

	reg, err := LoadRegistry(root)
	if err != nil {
		return "", AdoptedEntry{}, err
	}
	if existing, err := os.ReadFile(path); err == nil && string(existing) != string(src) {
		// A different text at the same slug is legal only as a forward move. Anything else
		// is the handover-side version of the silent mutation the seal refuses.
		prev := fmValue(existing, "version")
		if !semverNewer(version, prev) {
			return "", AdoptedEntry{}, fmt.Errorf("a different %q is already adopted at %s, and the new file says %s — a handed-over text never silently mutates; adopt only a newer version (that is what `sporo pull` reports)", slug, prev, version)
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", AdoptedEntry{}, err
	}
	if err := os.WriteFile(path, src, 0o644); err != nil {
		return "", AdoptedEntry{}, err
	}
	if reg.Adopted == nil {
		reg.Adopted = map[string]AdoptedEntry{}
	}
	reg.Adopted[slug] = entry
	if err := reg.Save(root); err != nil {
		return "", AdoptedEntry{}, err
	}
	return slug, entry, nil
}

// PullReport is one adopted recipe's staleness verdict.
type PullReport struct {
	Slug   string
	Status string // "up to date" | "update" | "skipped"
	Have   string
	Latest string
	// ExactChanged marks the loud case: the exact-contract digest moved, so the reader's
	// consumer-facing output must be re-verified, not just the prose re-read.
	ExactChanged bool
	Note         string
	fetched      []byte
}

// maxPullBody caps a fetched source. A recipe is prose; anything past this is not one.
const maxPullBody = 2 << 20

// Pull re-checks adopted recipes against their sources. `slug` narrows to one; empty means
// all, in stable order. With `apply`, an update also refreshes the stored copy and the
// registry entry — never on a regression, never on a skip.
func Pull(root, slug string, apply bool) ([]PullReport, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, err
	}
	var slugs []string
	for s := range reg.Adopted {
		if slug == "" || s == slug {
			slugs = append(slugs, s)
		}
	}
	if len(slugs) == 0 {
		if slug != "" {
			return nil, fmt.Errorf("nothing adopted as %q — `sporo adopt <exported file>` records a handover first", slug)
		}
		return nil, fmt.Errorf("nothing adopted in this repository — `sporo adopt <exported file>` records a handover first")
	}
	sort.Strings(slugs)

	var out []PullReport
	changed := false
	for _, s := range slugs {
		entry := reg.Adopted[s]
		r := PullReport{Slug: s, Have: entry.Version}
		body, err := fetchSource(entry.Source)
		if err != nil {
			r.Status, r.Note = "skipped", err.Error()
			out = append(out, r)
			continue
		}
		latest := fmValue(body, "version")
		switch {
		case latest == "":
			r.Status, r.Note = "skipped", "the source carries no `version:` frontmatter — it is not an exported recipe anymore"
		case latest == entry.Version:
			r.Status, r.Latest = "up to date", latest
		case semverNewer(latest, entry.Version):
			r.Status, r.Latest = "update", latest
			r.ExactChanged = exactContractsDigest(body) != entry.ExactContracts
			r.fetched = body
			if apply {
				if err := os.WriteFile(filepath.Join(root, adoptedHome, s+".md"), body, 0o644); err != nil {
					return out, err
				}
				entry.Version, entry.Hash = latest, ContentHash(body)
				entry.ExactContracts = exactContractsDigest(body)
				entry.Date = time.Now().Format("2006-01-02")
				reg.Adopted[s] = entry
				changed = true
			}
		default:
			r.Status, r.Latest = "skipped", latest
			r.Note = fmt.Sprintf("the source says %s, OLDER than your %s — refusing to regress a handover", latest, entry.Version)
		}
		out = append(out, r)
	}
	if changed {
		if err := reg.Save(root); err != nil {
			return out, err
		}
	}
	return out, nil
}

// AdoptedList returns the adopted slugs with their versions, for `sporo list`.
func AdoptedList(root string) (map[string]AdoptedEntry, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, err
	}
	return reg.Adopted, nil
}

// fetchSource reads a recorded source: http(s), or a filesystem path as given (the reader
// recorded it in their own convention; resolving it behind their back would move it).
// Every failure is a value, not a crash — an unreachable source is a normal Tuesday.
func fetchSource(source string) ([]byte, error) {
	if source == "" {
		return nil, fmt.Errorf("no source recorded at adopt time — re-adopt with --source to make pull possible")
	}
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(source)
		if err != nil {
			return nil, fmt.Errorf("source unreachable: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("source answered %d", resp.StatusCode)
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, maxPullBody+1))
		if err != nil {
			return nil, fmt.Errorf("source read failed: %v", err)
		}
		if len(body) > maxPullBody {
			return nil, fmt.Errorf("source exceeds the %dMB cap — a recipe is prose, and this is not one", maxPullBody>>20)
		}
		return body, nil
	}
	b, err := os.ReadFile(source)
	if err != nil {
		return nil, fmt.Errorf("source unreachable: %v", err)
	}
	return b, nil
}

// semverNewer: a deliberate small duplicate of the compare in the upgrade package. The
// recipe engine must not import the binary's own distribution machinery for three lines of
// integer comparison — the seam between "the tool updates itself" and "a recipe has
// versions" is worth more than the deduplication.
func semverNewer(candidate, current string) bool {
	a, okA := semverTriple(candidate)
	b, okB := semverTriple(current)
	if !okA || !okB {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

func semverTriple(v string) ([3]int, bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return [3]int{}, false
	}
	var out [3]int
	for i, p := range parts {
		if i == 2 {
			if j := strings.IndexAny(p, "-+"); j >= 0 {
				p = p[:j]
			}
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return [3]int{}, false
		}
		out[i] = n
	}
	return out, true
}
