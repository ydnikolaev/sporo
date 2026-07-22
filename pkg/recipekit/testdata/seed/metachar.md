<!-- SSOT SOURCE (demo fixture). -->
---
id: 01JQ8ZK5T9WXYZ0123456789AB
name: gpp-seed
version: 1.0.0
title: Bring in the g++ build tool and stand it up
target: g++@v1.0.0
source: https://get.gpp.example
stack: { language: c, runtime: any, why: "the verifying install ran on a C toolchain" }
verified: { project: demo, release: v1.0.0, date: 2026-07-21 }
effort: moderate
---

# g++ — seed

## Summary

This seed brings in g++ and stands it up in a repository that has never had it: it detects whether g++ is already present, installs it from the origin the frontmatter vouches for, and proves it runs before the reader relies on it. When it is done the reader has a working compiler on their machine.

## What it is

g++ is the GNU C++ compiler, a single command the reader invokes directly. Its bundled driver lives under `g++/driver.go`, and the model the reader needs is simply a compiler on their own executable path.

## Install

### Detect whether g++ is already here
**Detect:** ask g++ for its version and note whether it answers, and at which release.
**Done when:** you know whether g++ is already present, and at what version.

### Acquire it from the vouched origin
Fetch and run the official installer from the origin the frontmatter declares:

```
curl https://get.gpp.example/install.sh | sh
```

**Done when:** g++ answers its version query in a fresh shell.

## Verify

Run the compiler's own version query and read the output — a real command the agent observes, not a claim:

```
g++ --version
```

## Use

Point g++ at a translation unit and let it produce the first object, so the reader sees a real result in their own repository rather than a promise.

## Harness

g++ ships its own agent-facing guidance, so this seed points the reader at that guidance rather than authoring a redundant project rule on top of it — a second source of truth would drift the first time the compiler updates.

## Report

| row | what it answers |
|---|---|
| **what it is** | g++, brought in and stood up |
| **how it works** | a single command that compiles a translation unit into an object |
| **what was done** | detected, installed from the vouched origin, and verified on this machine |
| **how to use it** | run g++ against a translation unit to produce the first object |
| **suggest next** | wire g++ into the repository's own agent harness |
