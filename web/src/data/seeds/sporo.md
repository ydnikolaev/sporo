> **Runner protocol:** v1.1.0

*This preamble is addressed to the agent about to execute the seed printed below it. The seed is
the instruction; this only frames how you run it. You are acting on a repository whose layout you
have never seen, so hold to one discipline throughout: every move accounts for itself, and you
run nothing you cannot trace to an origin the seed vouches for.*

## Run it in order

**Read the anchors first.** Before you touch the tree, read the frontmatter: the target it stands
up — the named tool at its pinned version — and the source it declares. Those two are the promise
you are keeping and the only origin from which you may run code.

**Work the sections in the order they are written** — `## Install`, then `## Verify`, then
`## Use`, then `## Harness`. Each earns the next, so do not skip ahead: a tool you have not
installed is one you cannot verify. `## Install`'s first step probes whether the tool is already
present, and at what version; honor that probe before you install anything, so running the seed a
second time never clobbers a working copy.

**Prove each step before you build on it.** Every `## Install` step names the observable condition
that proves it took — check it, and stop rather than continue past one that did not hold. Then run
the check in `## Verify` and read its output: an installer reporting success is not the same as the
tool running here.

**Run nothing that does not trace to the declared source.** If a step fetches and executes remote
code, its origin must be the source the frontmatter vouches for. Code from anywhere else is an
unaudited privilege you would be taking on the human's behalf — refuse it.

**Wire the tool into this repository's harness — that is part of standing it up, not optional.**
`## Harness` says where the tool's surface belongs; plant it there per *this* repository's own
conventions (where it keeps agent skills, which instruction file its agents read), in roles, and
**ask the human before you change a file they own** — that permission may be the only thing you need
from them. The *verdict* on authoring a new project rule is advice: prefer the tool's own guidance
wherever it ships some, and do not add a second rule where the tool already has one.

## Account to the human

Close by completing the seed's own `## Report` — the fixed audit written for the person who has to
accept what you just did. Fill each row from what actually happened on this tree, above all the row
recording what you changed. Its forward rows — *how to use it* and *suggest next* — are the human's
**usage orientation**: name the surface they reach for (the skill), point them onward, and do not
send them back to run the install you just ran. Leave the table's shape exactly as the seed sets it.
That report, not this preamble, is where the run ends.

---
id: 01KY3NRFFZTXNE3NC88F1T1GSE
name: sporo
version: 1.1.0
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

You work sporo through your agent, not by memorising its verbs. The everyday surface is a **skill**
the agent now has, `sporo-recipe`:

- Tell the agent *"use sporo-recipe to author a recipe from a build already finished here"* — it
  mines this repository's own record, drafts, lints against the genre, seals, and exports **one
  self-contained file** for a person to review. *"…author a seed"* packages a named tool the same way.
- The raw verbs are there when a person wants them directly — `sporo new`, `sporo lint`,
  `sporo seal`, `sporo export`, and the `sporo seed` namespace for seeds — but the skill is the
  surface a human and agent share, and `sporo --help` is the full reference.

This is an orientation, not the whole map: an agent working in this repository can point to more —
the ways sporo fits how this repository already builds and ships.

## Harness

sporo ships its own agent-facing surface, so wiring it in is not "write a rule about sporo" — it is
"let sporo plant that surface where this repository keeps such things." The agent does this as part
of standing the tool up: `sporo init` reads this repository's own conventions and installs the
`sporo-recipe` skill into the skills home it finds, plus an agent-instructions block into the file
this repo's agents already read — no path is named by hand; it detects the layout and adapts.
Because that writes into files the repository owns, the agent asks before it does — the human's one
move may be nothing but granting it. Do **not** hand-author a separate rule restating what sporo is:
that is a second source of truth, and it drifts the first time the tool updates (`sporo update`
re-syncs the managed surface without clobbering local edits). Let the tool own its harness, and keep
the repository's own rules for what sporo does not cover.

## Report

| row | what happened |
|---|---|
| **what it is** | sporo — a single self-installing CLI that turns a build already finished into a transferable recipe (or a seed like this one) an agent in another repository can rebuild from. |
| **how it works** | one static binary on the executable path, a per-repo home it owns, and a sealed corpus it authors and exports; a person drives it through a skill in their agent, not by memorising verbs. |
| **what was done** | detected and installed the binary from the declared source (or built it from source) and verified it runs; if the harness step was taken, it also planted the `sporo-recipe` skill and an agent-instructions block into this repository — with permission, and nothing else changed. |
| **how to use it** | nothing to install by hand — it is set up. To use it, ask the agent to run the `sporo-recipe` skill: *"author a recipe"* (or *"author a seed"*), and it walks the whole build to one sealed file a person reviews. The agent can show more, fitted to this repository. |
| **suggest next** | package a build worth keeping as a recipe and share the exported file with someone on a different stack — the agent does the authoring; a person only points it at the build. |
