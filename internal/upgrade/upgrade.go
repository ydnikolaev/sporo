// Package upgrade owns the binary's own freshness: `sporo upgrade` (replace this executable
// with the latest release) and the passive version hint (tell the user a newer one exists,
// without ever getting in their way).
//
// The two-surface model, restated because it is the thing users conflate: the BINARY updates
// itself here; the SKILLS a binary wrote into a repository update via `sporo update`, per
// repo. The chain after a release is `sporo upgrade` → `sporo update` in each repo (`sporo
// projects` lists them), and every message this package prints ends by saying so — a fleet
// whose binaries are new and whose skills are stale believes it upgraded.
//
// The passive hint's discipline (it runs after EVERY command, so its budget is strict):
// at most one network probe per TTL, answered from a cache file the rest of the time; a
// hard timeout; silent on ANY failure (an offline machine is a normal machine, not an
// error path); stderr, one line, never blocking; and two opt-outs — the user's
// (SPORO_NO_UPDATE_CHECK) and the machine's (CI), because a hint nobody is there to read
// is noise in a log.
package upgrade

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/creativeprojects/go-selfupdate"
	"gopkg.in/yaml.v3"
)

// Repo is where releases live. A coordinate, deliberately: this is the one package that is
// ABOUT this product's own distribution, not about anyone's repository.
const Repo = "ydnikolaev/sporo"

// ChecksumAsset is goreleaser's checksum file, stable-named so the validator can find it
// on any release.
const ChecksumAsset = "checksums.txt"

// Latest asks the release host for the newest version. The token is optional — required
// only while the repository is private — and comes from the caller so this package never
// reads the environment behind anyone's back.
//
// Deliberately NO validator here: go-selfupdate resolves the validation asset at detection
// time, so a detector with a validator cannot even SEE a release that lacks the checksum
// file — and asking "what is the latest?" must work against any release. Validation
// belongs to the one path that writes to disk (Self), which is exactly where it stays.
func Latest(ctx context.Context, token string) (*selfupdate.Release, error) {
	updater, err := newUpdater(token, false)
	if err != nil {
		return nil, err
	}
	latest, found, err := updater.DetectLatest(ctx, selfupdate.ParseSlug(Repo))
	if err != nil {
		return nil, fmt.Errorf("could not reach the releases (while the repository is private, set GITHUB_TOKEN — `gh auth token` prints one): %w", err)
	}
	if !found {
		return nil, fmt.Errorf("no release found for this platform")
	}
	return latest, nil
}

// Self replaces the running executable with the latest release, checksum-validated.
func Self(ctx context.Context, current, token string) (*selfupdate.Release, error) {
	updater, err := newUpdater(token, true)
	if err != nil {
		return nil, err
	}
	return updater.UpdateSelf(ctx, current, selfupdate.ParseSlug(Repo))
}

func newUpdater(token string, validated bool) (*selfupdate.Updater, error) {
	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{APIToken: token})
	if err != nil {
		return nil, err
	}
	cfg := selfupdate.Config{Source: source}
	if validated {
		cfg.Validator = &selfupdate.ChecksumValidator{UniqueFilename: ChecksumAsset}
	}
	return selfupdate.NewUpdater(cfg)
}

// checkState is the hint's memory, one small file in the machine home.
type checkState struct {
	CheckedAt time.Time `yaml:"checked_at"`
	Latest    string    `yaml:"latest"`
}

// TTL is how often the hint is allowed one network probe. A day: release cadence is days,
// and a CLI that phones home more often than it is released is surveilling, not helping.
const TTL = 24 * time.Hour

// Hint returns the one-line nudge when a newer release is known, refreshing its knowledge
// through `fetch` at most once per TTL. Everything about it fails SILENT and CHEAP: a dev
// build never hints (it was built from a checkout — the release train is not its upstream),
// a failed fetch records the attempt and stays quiet (offline is a normal state), and the
// comparison is answered from the cache file every run in between.
func Hint(home, current string, now time.Time, fetch func() (string, error)) string {
	if current == "dev" || home == "" {
		return ""
	}
	path := filepath.Join(home, "version-check.yaml")
	var st checkState
	if b, err := os.ReadFile(path); err == nil {
		_ = yaml.Unmarshal(b, &st)
	}
	if now.Sub(st.CheckedAt) >= TTL {
		st.CheckedAt = now
		if latest, err := fetch(); err == nil {
			st.Latest = latest
		}
		if err := os.MkdirAll(home, 0o755); err == nil {
			if b, err := yaml.Marshal(st); err == nil {
				_ = os.WriteFile(path, b, 0o644)
			}
		}
	}
	if st.Latest == "" || !newer(st.Latest, current) {
		return ""
	}
	return fmt.Sprintf("sporo: %s is available (you run %s) — `sporo upgrade`, then `sporo update` in each repo (`sporo projects` lists them)", st.Latest, current)
}

// newer compares two semver triples, tolerant of a leading v.
func newer(candidate, current string) bool {
	a, okA := triple(candidate)
	b, okB := triple(current)
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

func triple(v string) ([3]int, bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return [3]int{}, false
	}
	var out [3]int
	for i, p := range parts {
		if i == 2 {
			// A prerelease or build suffix (2.0.0-rc1) still orders by its patch digits.
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
