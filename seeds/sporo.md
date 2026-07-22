<!-- SSOT SOURCE (sporo). The export strips this banner; edit here, hand over ONLY what `sporo seed export` prints. -->

---
id: 01KY3NRFFZTXNE3NC88F1T1GSE
name: sporo
version: 1.0.0
title: sporo — the recipe & seed CLI, self-installed into this repository
target: sporo@0.13.0
source: https://sporo.dev
stack: { language: go, runtime: darwin/arm64, why: "the verifying install ran a darwin/arm64 release archive; sporo ships static binaries for darwin and linux on amd64 and arm64" }
verified: { project: sporo, release: v0.13.0, date: 2026-07-22 }
effort: reference
---

# sporo — the recipe & seed CLI, self-installed into this repository

## Summary

sporo is a single self-installing CLI that turns a build you already finished into a transferable
**recipe** — one self-contained file that teaches an agent in a repository which has never heard of
your work how to rebuild the capability from principles. Standing it up gives this repository the
authoring surface: the verbs to scaffold, lint, seal, and export both recipes and seeds. When this
seed is done the tool answers on the command line, its embedded corpus lists, and the repository can
author its own transferable builds.

## What it is

sporo is a single static binary — no runtime to install, no dependency tree to resolve — that puts
itself on the command line and then works inside any repository. It keeps its per-repo state in a
home directory it owns, carries an embedded corpus of recipes and seeds, and maintains a registry
that seals what it authors: a sealed artifact records its version and content hash, so it can never
silently mutate underneath a reader.

Two shapes travel through it. A **recipe** teaches an agent to *build* a capability from roles and
principles, naming nothing concrete, so it survives being handed to a stranger. A **seed** — the
shape you are reading now — teaches an agent to *acquire* one named tool and stand it up. Hold sporo
in your head as an authoring tool that produces self-contained, auditable files, not a service you
leave running.

## Install

### 1. Detect whether sporo is already present

**Detect:** probe the command line for the tool and read what it reports:

```
sporo --version
```

A present install prints a version line — a release build prints its pinned version, a build from
source reports a development version. Any version line means the tool is already here, so skip
straight to Verify; only a "command not found" means it is absent and you should continue below.

**Done when:** the probe either prints a version line (present — go to Verify) or reports the
command is missing (absent — continue to step 2).

### 2. Install from the canonical origin (script route)

The one-liner the project publishes fetches the latest release, verifies its checksum against the
release's own checksums, and installs the binary to the reader's executable path. It traces to the
origin the `source` frontmatter declares, and refuses to install anything whose checksum does not
match — so the pipe runs only code that comes from the declared source:

```
curl -fsSL https://sporo.dev/install.sh | sh
```

If the chosen executable path is not on the shell's lookup path, the script prints the one line to
add. On macOS it also clears the quarantine flag so the fresh binary runs.

**Done when:** the version probe prints a released version line on the command line.

### 3. Install with the language toolchain (source route)

With a Go 1.26+ toolchain present, the tool can be built and installed from source instead of
fetching a release archive:

```
go install sporo.dev/sporo/cmd/sporo@latest
```

This builds the same source, but self-reports a development version rather than a pinned one,
because the version stamp is applied only when a release archive is assembled. Make sure the
toolchain's install directory is on the shell's lookup path.

**Done when:** the tool resolves on the command line and its help lists the verbs.

## Verify

Prove the tool actually runs here — not that an installer reported success. Run the version probe
together with a functional verb that exercises the embedded corpus:

```
sporo --version
sporo list
```

A version line plus a listed corpus is the proof: the binary landed and its embedded corpus reads.
A missing binary, or a version line above an empty list, means the install did not fully take —
return to Install.

## Use

The first move is to install the authoring surface into this repository, then scaffold and ship a
build with it:

- `sporo init` — install the authoring surface (a skill, an agent-instructions block, and the
  embedded seeds) into this repository.
- `sporo new` scaffolds a coached recipe draft; `sporo seed new` scaffolds a seed draft. A draft
  cannot ship until it lints clean.
- `sporo lint` — check a draft against the genre: shape, per-step acceptance, neutrality, and seals.
- `sporo seal` — record a build's version and content hash in the registry, so a sealed build never
  silently mutates.
- `sporo export` — print one recipe (or, under `sporo seed`, one seed) as a single self-contained
  file, ready to hand to an agent in another repository.

Point at `sporo --help` and the project's own docs for the full reference; this seed teaches only
the setup surface, not every verb.

## Harness

sporo ships its own agent-facing guidance, so the right move is to point this repository's future
agents at it rather than author a parallel description by hand. Running `sporo init` installs that
surface: a skill the repo's agents can invoke and an agent-instructions block that records the tool
is here and how to reach for it. Do **not** write a separate project rule restating what sporo is —
that is a second source of truth, and it drifts from the tool's own the first time the tool updates
(`sporo update` re-syncs the managed surface without clobbering your edits). Let the tool own its
harness, and keep the repository's hand-written rules for what the tool does not cover.

## Report

| row | what happened |
|---|---|
| **what it is** | sporo — a single self-installing CLI that turns a finished build into a transferable recipe or seed. |
| **how it works** | one static binary on the executable path, a per-repo home it owns, and a sealed corpus it authors and exports. |
| **what was done** | installed the binary to the reader's executable path from the declared source (or built it from source) and verified it runs; nothing else in the tree changed unless `sporo init` was run. |
| **how to use it** | run `sporo init` to install the authoring surface, then `sporo new` to scaffold a first recipe. |
| **suggest next** | author and seal a first recipe, or hand this tool to another repository by exporting this seed with `sporo seed export`. |
