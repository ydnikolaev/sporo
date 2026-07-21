<!-- SSOT SOURCE (sporo repo). Consumers receive a provenance-stamped copy via `sporo update` — edit HERE, never the synced copies. -->

---
name: _runner
version: 1.0.0
title: The runner preamble prepended to every seed export
problem: An agent is handed one sealed install seed for a repository it has never seen, and must stand the named tool up without clobbering a working tree or claiming a success it never proved.
prerequisites: [read-files, edit-files, run-commands]
derived_from: [the seed genre]
stack: { language: markdown, runtime: any, why: "the preamble is prose that frames how a seed is executed" }
verified: { project: sporo, release: v0.12.0, date: 2026-07-21 }
effort: reference
---

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

**Take `## Harness` as advice, not a command.** Wire the tool into this repository's agent harness
only as far as that section suggests, and prefer the tool's own guidance wherever it ships some.

## Account to the human

Close by completing the seed's own `## Report` — the fixed audit written for the person who has to
accept what you just did. Fill each row from what actually happened on this tree, above all the row
recording what you changed, and leave the table's shape exactly as the seed sets it. That report,
not this preamble, is where the run ends.
