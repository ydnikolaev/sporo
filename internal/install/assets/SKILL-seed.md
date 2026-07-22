---
name: sporo-seed
description: Author a seed — one self-contained install file that teaches an agent in a repository that has never heard of a tool how to bring that tool in and stand it up: detect whether it is already there, install it from an origin the seed vouches for, prove it runs, put it to use, wire it into the repo's harness, and account for every move to the human. Use when a specific tool, at a specific version, is worth bringing into repositories that lack it.
argument-hint: "[the tool and the version to bring in]"
---

# sporo-seed

> **Thesis.** A seed teaches an **agent in a repository that has never heard of a tool** how to
> bring that tool in and stand it up — and it does so as a chain of *moves on a foreign tree*:
> it will run installers, touch a PATH, edit a harness. So the discipline is not the recipe's
> **neutrality of design** but **accountability of action**: every step says how it knows it
> worked, no step runs code the human cannot trace to a declared origin, and the closing section
> is a fixed audit the human reads the same way every time. Your job is to read the tool's
> *current* truth, **run the install yourself**, and exercise judgment — which install path is
> essential and which is an accident of your OS, which check actually proves the tool runs. The
> machine's job is to hold the shape and refuse the unaccountable move.

## When to run

- **A specific tool, at a specific version, is worth bringing into a repository that lacks it**,
  and bringing it in takes real moves: a detect, an install from a trustworthy origin, a proof it
  runs, and a decision about how this repo's future agents reach for it.
- **NOT for:** a capability with **no external tool** behind it — that is a *recipe* (`sporo-recipe`),
  which teaches an agent to BUILD a thing from principles, not acquire one; a tool already ambient
  on every plausible machine (the seed would be an empty detect); or a tool you have **not installed
  and run yourself** — a seed from the tool's marketing page is a guess with an install command in it.

A seed is not a recipe and not a manual. A **recipe** survives being handed to a stranger *because*
it names nothing concrete. A **seed** survives because it names its **one tool** precisely — pinned
to a version, traced to an origin — while staying silent about the stranger's own tree. If your
document has no external tool as its subject, it is a recipe wearing install commands; write the
recipe.

## Where it runs

**In whatever repository you are standing in.** A seed is written *about* a tool; it lands in this
project's own seed home (`seeds/`). The corpus the binary ships is read-only — never write there.

Read the genre spec before you write a line:

```
sporo genre --seed
```

It is the SSOT for the shape: the seven sections, the trust contract, and the neutrality inversion.
Everything below is its working summary — the spec wins where they differ.

## Procedure

### 1. Install and run the tool yourself — the raw material is read and run, not recalled

Before you write anything, **read the tool's current documentation and run the real install on your
own machine.** A seed written from memory writes the install you saw last release and silently hands
the reader a step the tool changed. Observe: which command actually acquired it, from which origin;
what proves it landed; the first real thing you did with it. That observation *is* the seed — and the
`verified` stamp is your promise that it happened.

Do the one thing that cannot be automated — **judgment**: which install path is the essential one and
which is an accident of your OS; which verify check proves the tool *works* versus merely proves a
file exists; whether the tool ships its own agent guidance or a thin project rule is warranted.

### 2. Scaffold — do not start from a blank page

```
sporo seed new <slug>
```

The scaffold arrives as a **draft** (`draft: true`): the seven sections stubbed with coach comments,
the frontmatter pre-stamped. A draft is exempt from the gate and refused by seal and export, so
nothing half-written can ship; removing `draft: true` is the act of saying "this seed stands on its
own." Fill the nine frontmatter keys — the load-bearing three are all about **trust in a concrete
thing**: `target` (`<tool>@<version>`, pinned — the subject), `source` (the canonical origin the seed
vouches for, an origin the reader can inspect), and `verified` (`{project, release, date}` — whose
machine proved it). `id` is a ULID the tool minted; never type it.

### 3. Write the seven sections — and earn the seed's teeth

The shape, in this order (the sequence IS the argument — you cannot **use** a tool you have not
**verified**, verify one you have not **installed**, install one you cannot **describe**):

`## Summary` → `## What it is` → `## Install` → `## Verify` → `## Use` → `## Harness` → `## Report`.

Four teeth are mechanical, each gated — they turn a hopeful sequence of commands into an
**accountable** one:

- **Detect before you install.** The **first** `###` step under `## Install` opens with the
  `**Detect:**` marker and answers one question: *is the target already here, and at what version?*
  Detect-first is what makes a seed **idempotent** — safe to run twice, safe in a repo that already
  has the tool — the difference between a seed and a script that assumes a blank machine.
- **Every install step closes with `**Done when:**`.** A literal line stating the *observable*
  condition that proves the step took — the detect step included. Installs fail in quiet ways (a
  package manager reports success while the shell's lookup still points at nothing; a binary lands
  without its execute bit). A step with no acceptance is one the agent cannot tell is failing until
  the end — which, on a foreign tree, is when failure is most expensive.
- **No blind pipe into a shell.** The most dangerous move a seed makes is fetching a remote script
  and piping it into an interpreter — arbitrary code, from the network, with the reader's
  privileges. The genre does **not** ban it (for many tools it is the official path) but requires the
  step to **cite the `source` origin the frontmatter declares**, so the human audits it as "this
  comes from the origin this seed vouches for." A pipe with no cited origin reds. Citing `source` is
  the whole escape; there is no other.
- **`## Verify` holds at least one fenced command block** — a check the agent *runs and observes*,
  not prose asserting success. It is the acceptance for the tool as a whole. Prose that says "the
  tool now runs" proves nothing; a command whose output the agent can read is the difference between
  "I ran the installer" and "the tool actually runs here."

`## Use` is the **usage orientation** for the human (read by both): it **must name the skills** a
human and agent share (that surface is the whole point of standing the tool up) and shows the first
real move through it; it **may carry a command** only when the move is genuinely the human's; it is an
orientation the reading agent **extends in place**, and it never re-lists the install the agent
already ran.

### 4. Name the tool, not the reader's tree — the neutrality inversion

A recipe names **nothing** concrete; a seed names its **one tool** freely — the target, its version,
its origin, its install command all belong in the document, because they *are* the document. What a
seed must still not name is the **reader's tree**: their paths, their filenames, their config
locations, where their harness lives. Say those as **roles** — *the reader's executable path*,
*wherever this repository keeps its agent rules*, *the project's own config* — never a literal lifted
from your machine. The test is the recipe's, pointed at a different object: *could an agent follow
this in a repository whose layout you have never seen?* (The gate exempts the target tool's own name
from the coordinate scan, and erases URLs — so the install line and the cited origin pass, while the
reader's tree still reds.)

### 5. The Report is for the human — five fixed rows

`## Summary` orients the agent; `## Report` is the **only section written for the person** who has to
accept what the seed just did. It is a Markdown table of **exactly these five rows, in this order**:

| row | what it answers |
|---|---|
| **what it is** | the tool, in one line the human can act on |
| **how it works** | the mechanism, enough that the human can reason about it |
| **what was done** | the actual mutations this run made to **this** repository — the audit |
| **how to use it** | the working surface **named** (the skill a human and agent share); a command only if it is the human's own — never the install already run |
| **suggest next** | where to go from here |

An **extra** row, a **missing** row, or a **reordered** column all red: a human reviews many seed
runs, and a report whose shape changes every time is one they must re-read from scratch. For the
human, the fixed shape **is** the product.

### 6. `## Harness` — how the agent plants the tool, advisorily

`## Harness` says how the installed tool joins **this repository's** agent harness so future agents
know it is here. Planting it is the **agent's** move, part of standing the tool up: read *this* repo's
own conventions — where it keeps agent skills, which instruction file its agents read — and wire the
tool's surface **there**, in roles, **asking before you mutate a file the human owns**. The section is
required to be present, but its **verdict is advisory** and the gate checks only that it exists. The
rule it follows: **recommend a project-local rule or skill only when the tool ships none of its own** —
bolting a redundant rule on a tool that already carries its own guidance manufactures a second source
of truth that drifts the first time the tool updates.

### 7. Gate it, then read it as a stranger

```
sporo seed lint
```

The gate checks the seven-section shape and order, the detect-first opener, the per-step
`**Done when:**`, the cited-source pipe rule, the runnable Verify fence, the five-row Report, and the
neutrality of every line (it knows this project's forbidden names). Then re-read the body **as an
agent dropped into a repository whose layout you have never seen**: every step you could not follow
from there is one still written for someone who already has the tool.

### 8. Seal it, then ship the exported file — never the source

```
sporo seed seal <slug>
sporo seed export <slug>
```

The seal records (id, version, content hash, provenance): from this moment the seed never silently
mutates — every later edit must bump `version:` and re-seal, and the gate enforces it. The export
composes the deliverable: the provenance banner is stripped and the **runner protocol** — the
preamble that frames, for the agent about to execute it, *how* to run a seed (read the anchors first,
work the sections in order, prove each step, account for every move) — is prepended. Hand over *that*
file. A raw source handed to a reader arrives without the frame the whole genre rests on.

## The rules you cannot bend

1. **Run it, don't recall it.** The install command, the verify check, the origin — all live in the
   tool's *current* truth and in what it *does when you run it*, not in memory.
2. **Detect before you install.** A seed that installs blind clobbers a working tree; detect-first is
   what makes it idempotent.
3. **Every step says how it knows it took.** A `**Done when:**` on each, or the agent finds out it
   failed at the end — on a foreign tree, the most expensive time.
4. **No untraceable pipe.** A remote script into a shell cites the declared `source`, or it does not
   pipe at all.
5. **Verify runs something.** At least one fenced command the agent observes — not prose that asserts
   success.
6. **Name the tool, role the tree.** Concrete about the one tool; the reader's paths, files, and
   config are spoken of only in roles.
7. **The Report is exactly five rows, in order.** For the human, the fixed shape is the product.
8. **The stamp is honest.** A seed is a snapshot of one successful install — whose machine, which
   release, when — not maintained doctrine.

## Output

One seed in **this project's** seed home (`seeds/`), green under `sporo seed lint`, sealed in the
registry, plus a one-line report: what tool it brings in, what the install actually did on your
machine to prove it, and the judgment calls you made (which path was essential, whether a project
rule was warranted).
