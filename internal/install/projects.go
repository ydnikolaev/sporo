package install

// The global projects registry — `~/.sporo/projects.yaml` — answers the one question a
// per-repo registry cannot: "which repositories on this machine did I install sporo into?"
// After `sporo upgrade` brings a new binary, that list is where the skills are stale; the
// `projects` verb reads it so the user (or their agent) can walk the fleet and `sporo update`
// each one, instead of rediscovering repos by memory.
//
// It is a CONVENIENCE, and the code treats it as one: registration is best-effort, and a
// machine where the home directory is unwritable still gets a fully working `init` — the
// project-local install is the contract, the global list is a courtesy. That is also why the
// file lives outside every repo: it is a fact about this MACHINE, not about any project, and
// committing it anywhere would sync one user's filesystem paths to another user's checkout.

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ProjectEntry is one repository sporo was installed into, and which binary last touched it.
type ProjectEntry struct {
	Root    string `yaml:"root"`
	Binary  string `yaml:"binary"`
	Updated string `yaml:"updated"`
}

type projectsFile struct {
	Schema   int            `yaml:"schema"`
	Projects []ProjectEntry `yaml:"projects,omitempty"`
}

// GlobalHome resolves where machine-level sporo state lives: `SPORO_HOME` when set (the
// seam tests and unusual setups use), else `~/.sporo`. An unresolvable home returns "" and
// the callers treat that as "no global registry today" — never as a reason to fail.
func GlobalHome() string {
	if h := os.Getenv("SPORO_HOME"); h != "" {
		return h
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".sporo")
}

// RegisterProject upserts one repository into the global list, keyed by its absolute root —
// the same repo initialized twice is one entry with a fresher stamp, not two entries.
func RegisterProject(globalHome, absRoot, binaryVersion string) error {
	if globalHome == "" {
		return fmt.Errorf("no global home to register in (no home directory and no SPORO_HOME)")
	}
	pf, err := readProjects(globalHome)
	if err != nil {
		return err
	}
	entry := ProjectEntry{Root: absRoot, Binary: binaryVersion, Updated: time.Now().Format("2006-01-02")}
	replaced := false
	for i, p := range pf.Projects {
		if p.Root == absRoot {
			pf.Projects[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		pf.Projects = append(pf.Projects, entry)
	}
	pf.Schema = 1
	b, err := yaml.Marshal(pf)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(globalHome, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(globalHome, "projects.yaml"), b, 0o644)
}

// Projects lists the registered repositories. A machine with no registry yet is an empty
// list, not an error — the verb's job is to answer, not to scold.
func Projects(globalHome string) ([]ProjectEntry, error) {
	pf, err := readProjects(globalHome)
	if err != nil {
		return nil, err
	}
	return pf.Projects, nil
}

func readProjects(globalHome string) (projectsFile, error) {
	var pf projectsFile
	if globalHome == "" {
		return pf, nil
	}
	data, err := os.ReadFile(filepath.Join(globalHome, "projects.yaml"))
	if err != nil {
		return pf, nil
	}
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return pf, fmt.Errorf("the global projects registry is malformed — fix the YAML (degrading to an empty list would silently forget every repo this machine installed into): %w", err)
	}
	return pf, nil
}
