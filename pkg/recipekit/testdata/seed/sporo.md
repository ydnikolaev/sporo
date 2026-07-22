<!-- SSOT SOURCE (demo fixture). -->
---
id: 01JQ8ZK5T9WXYZ0123456789AB
name: sporo-seed
version: 1.0.0
title: Bring in the sporo recipe tool and stand it up
target: sporo@v1.0.0
source: https://sporo.dev
stack: { language: go, runtime: any, why: "the verifying install ran on a Go toolchain" }
verified: { project: demo, release: v1.0.0, date: 2026-07-21 }
effort: moderate
---

# sporo — seed

## Summary

This seed brings in sporo and stands it up in a repository that has never had it: it detects whether sporo is already present, installs it from the origin the frontmatter vouches for, and proves it runs before the reader relies on it. When it is done the reader has a working sporo on their machine and a note left for the next agent.

## What it is

sporo is a single self-contained binary that exports recipes an agent can rebuild from. It lands as one executable the reader runs directly, built from sporo/cmd/main.go, with no long-lived service, so the model the reader needs is simply a command on their own executable path. Its own configuration lives under `.sporo/config.yaml`, sporo's fixed convention in every repository that runs it.

## Install

### Detect whether sporo is already here
**Detect:** ask sporo for its version and note whether it answers, and at which release.
**Done when:** you know whether sporo is already present, and at what version.

### Acquire it from the vouched origin
Fetch and run the official installer from the origin the frontmatter declares:

```
curl https://sporo.dev/install.sh | sh
```

**Done when:** sporo answers its version query in a fresh shell.

## Verify

Run sporo's own version query and read the output — a real command the agent observes, not a claim:

```
sporo --version
```

## Use

Point sporo at a recipe and let it export the first artifact, so the reader sees a real result in their own repository rather than a promise.

## Harness

sporo ships its own agent-facing guidance, so this seed points the reader at that guidance rather than authoring a redundant project rule on top of it — a second source of truth would drift the first time sporo updates.

## Report

| row | what it answers |
|---|---|
| **what it is** | sporo, brought in and stood up |
| **how it works** | a single binary that exports recipes an agent can rebuild from |
| **what was done** | detected, installed from the vouched origin, and verified on this machine |
| **how to use it** | run sporo against a recipe to export the first artifact |
| **suggest next** | wire sporo into the repository's own agent harness |
