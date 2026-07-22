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
id: 01J9ZC3Q0K2X7V8B4N6M5T1A2W
name: widget
version: 1.0.0
title: Install the widget CLI
target: widget@1.4.0
source: https://example.com/widget
stack: { language: go, runtime: any }
verified: { project: sporo, release: v0.12.0, date: 2026-07-21 }
effort: quick
---

## Summary

The widget CLI turns a repository's build graph into a single reproducible command. This seed
stands it up at a pinned version and proves it runs on the reader's own tree.

## What it is

A small Go binary that reads a build manifest and executes its targets in dependency order.

## Install

### Detect, then install

**Detect:** run `widget --version` and read the output — if it already prints 1.4.0, skip to Verify.
Otherwise fetch the pinned release from the declared source and put it on PATH.

**Done when:** `widget --version` prints 1.4.0.

## Verify

```
widget --version
```

## Use

Run `widget build` from the repository root; it reads the manifest and runs each target.

## Harness

If this repository wires tools into an agent harness, add `widget build` as its build gate.

## Report

| field | value |
|---|---|
| what it is | the widget build runner |
| how it works | reads a manifest, runs targets in order |
| what was done | installed 1.4.0, proved it runs |
| how to use it | `widget build` from the root |
| suggest next | wire it into the gate |
