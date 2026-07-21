<!-- SSOT SOURCE (demo fixture). -->
---
id: 01JQ8ZK5T9WXYZ0123456789AB
name: acme-seed
version: 1.0.0
title: Bring in the acme build tool and stand it up
target: acme@v2.3.0
source: https://get.acme.example
stack: { language: go, runtime: any, why: "the verifying install ran on a Go toolchain" }
verified: { project: demo, release: v1.0.0, date: 2026-07-21 }
effort: moderate
---

# acme — seed

## Summary

This seed brings in the acme build tool and stands it up in a repository that has never had it: it detects whether the tool is already present, installs it from the origin the frontmatter vouches for, and proves it runs before the reader relies on it. When it is done the reader has a working tool on their machine and a note left for the next agent.

## What it is

The acme tool is a single self-contained binary that turns a declarative build description into an artifact. It lands as one executable the reader runs directly, with no long-lived service and no background daemon, so the model the reader needs is simply a command on their own executable path.

## Install

### Detect whether the tool is already here
**Detect:** ask the tool for its version and note whether it answers, and at which release.
**Done when:** you know whether the tool is already present, and at what version.

### Acquire it from the vouched origin
Fetch and run the official installer from the origin the frontmatter declares:

```
curl https://get.acme.example/install.sh | sh
```

**Done when:** the tool answers its version query in a fresh shell.

## Verify

Run the tool's own version query and read the output — a real command the agent observes, not a claim:

```
acme --version
```

## Use

Point the tool at a build description and let it produce the first artifact, so the reader sees a real result in their own repository rather than a promise.

## Harness

The acme tool ships its own agent-facing guidance, so this seed points the reader at that guidance rather than authoring a redundant project rule on top of it — a second source of truth would drift the first time the tool updates.

## Report

| row | what it answers |
|---|---|
| **what it is** | the acme build tool, brought in and stood up |
| **how it works** | a single binary that renders a build description into an artifact |
| **what was done** | detected, installed from the vouched origin, and verified on this machine |
| **how to use it** | run the tool against a build description to produce the first artifact |
| **suggest next** | wire the tool into the repository's own agent harness |
