<!-- SSOT SOURCE (sporo repo). Consumers receive a provenance-stamped copy via `sporo update` — edit HERE, never the synced copies. -->

---
name: _authoring
version: 1.0.1
title: Authoring a seed — the canonical shape of a transferable install
problem: A tool we depend on cannot be brought into a repository that has never heard of it — safely, idempotently, and in a way the human can audit.
prerequisites: [read-files, edit-files, run-commands]
derived_from: [the recipe genre]
stack: { language: markdown, runtime: any, why: "the genre spec is prose about prose" }
verified: { project: sporo, release: v0.12.0, date: 2026-07-21 }
effort: reference
---

# Seed — authoring

> **Thesis.** A seed teaches an **agent in a repository that has never heard of a tool** how
> to bring that tool in and stand it up: detect whether it is already here, install it from an
> origin the seed vouches for, prove the install actually took, put it to use, wire it into the
> repository's own agent harness, and then account for every one of those moves to the human who
> has to live with them. Its reader is a machine that will **act on a foreign tree** — it will
> run installers, touch a PATH, edit a harness — so the whole document is a chain of moves, and
> the hard constraint is not neutrality of design (the recipe's constraint) but **accountability
> of action**: every step says how it knows it worked, no step runs remote code the human cannot
> trace to a declared origin, and the closing section is a fixed audit the human reads the same
> way every time. A seed that installs without detecting clobbers a working tree; one that
> installs without verifying reports success it never had; one that pipes a stranger's script
> into a shell with no cited origin is an unaudited privilege it handed away on the reader's
> behalf. Every rule below exists to keep those from happening.

## 0. When to write a seed (and when not)

Write a seed when **a specific tool, at a specific version, is worth bringing into a repository
that does not have it** — and bringing it in takes real moves: a detect, an install from a
trustworthy origin, a proof that it runs, and a decision about how the repository's future agents
will reach for it. The trigger is a tool whose adoption an agent would otherwise improvise —
installing the wrong build, from an unverified origin, with no check that it landed, and no note
left behind for the next agent.

Do **not** write a seed for: a capability with **no external tool** behind it — that is a recipe,
which teaches the reader to *build* the thing from principles, not to *acquire* one that already
exists; a tool that is already ambient on every plausible reader's machine (the seed would be an
empty detect); or a tool you have not actually installed and run yourself — a seed written from
the tool's marketing page is a guess with an install command in it, and it will hand the next
agent your untested assumptions as if you had earned them.

**A seed is not a recipe and not a manual.** The three genres answer different questions, and
mistaking one for another is the failure this section exists to prevent:

| genre | answers | the reader | names the subject? |
|---|---|---|---|
| **recipe** | how to **build** a capability from nothing, anywhere, from principles | an agent who will implement, possibly on another stack | never — the body names roles, never instances |
| **manual / handbook** | how to **operate** this system day to day | the operator of that already-installed system | yes — that is its whole job |
| **seed** | how to **bring in** one named tool and stand it up, once, in a repo that lacks it | an agent who will run the install on a foreign tree | **yes, the target tool — that is the subject; but never the reader's tree** |

A recipe survives being handed to a stranger *because* it names nothing concrete; a seed survives
because it names its **one tool** precisely — pinned to a version, traced to an origin — while
staying silent about the stranger's own paths and files. The recipe's discipline is transferable
design; the seed's is trustworthy action. If your document has no external tool as its subject, it
is a recipe wearing install commands — write it as a recipe. If it only tells an already-installed
tool's operator which flags to pass, it is a manual — that is not this genre either.

## 1. The shape — seven sections, in this order

Every seed has frontmatter, then seven required body sections in this order. The shape is
**gated**, because a genre defined only by taste drifts into whatever the last author felt like,
and a seed that skips a rung leaves a gap an agent fills by improvising on the reader's machine.

1. **Frontmatter** (`## 1a` below) — the provenance and the trust anchors.
2. **`## Summary`** — a 2–4 sentence orientation for humans and agents before any move: what tool
   this brings in, what standing it up buys the reader, and the state they are in when it is done.
   It is body text, so the neutrality constraint applies in full, and the gate rejects a label or
   stub shorter than 80 visible characters — a Summary too short to be a paragraph is a title
   pretending to be an argument.
3. **`## What it is`** — the tool itself: what it does, the shape of the thing that lands (a single
   binary, a language package, a service the reader runs), and the model the reader needs in their
   head before they let it onto their machine. This is understanding *before* acquisition — an
   agent that installs a thing it cannot describe cannot judge whether the install went right.
4. **`## Install`** — the acquisition sequence, one `###` step per heading. This section carries
   the seed's two sharpest teeth: its **first** step opens with the `**Detect:**` marker (is the
   tool already here, and at what version?), and **every** step — the detect step included — closes
   with a `**Done when:**` line stating the observable condition that proves it took. An install
   sequence with no per-step acceptance is a wish list the agent cannot tell it is failing until
   the end, which on a foreign tree is exactly when failure is most expensive. See §2.
5. **`## Verify`** — the proof the whole install actually works, and it must contain at least one
   **fenced command block**: something the agent *runs* and observes, not prose that asserts
   success. Install can lie — a package half-lands, a PATH does not refresh, a binary arrives
   without the execute bit. Verify is the seed's acceptance at the whole-tool level; a Verify with
   no runnable proof in it verifies nothing.
6. **`## Use`** — how the reader actually puts the now-installed tool to work: the first real thing
   they do with it, in their own repository. This is the payoff the Summary promised, made
   concrete enough to act on but still neutral about the reader's own tree (§3).
7. **`## Harness`** — how the tool joins the repository's **agent harness** so that future agents
   in this repo know it is here and how to reach for it. This section is **advisory**, not a
   command to author a rule — see §5.
8. **`## Report`** — the closing section, and **the only one written for a person**. It is a fixed
   five-row table (§4) that accounts for what the seed just did to this repository. A seed is an
   action taken on the human's behalf; the Report is where the human audits it.

**Why the sequence is the argument.** The order is not a table of contents, it is a chain in which
each section earns the right to the next. You cannot **use** a tool you have not **verified**; you
cannot verify one you have not **installed**; you cannot install one you do not **understand**
enough to describe; you should not describe one before you have **oriented** the reader to why they
want it. Then, having stood the tool up, you **integrate** it into the harness so the effort is not
lost the moment this agent's session ends, and finally you **account** for the whole run to the
human. Reorder the sections and you break the argument — a Verify before an Install verifies
nothing, a Report before a Use audits a job half done. The gate holds the seven in order for
exactly this reason.

### 1a. Frontmatter — what each key vouches for

Nine keys, and unlike a recipe's the load-bearing three are all about **trust in a concrete
thing**: what tool, from where, proven by whom.

- **`id`** — the seed's **permanent identity**, a ULID minted once by the tool and never
  hand-edited. The slug can be renamed; the id cannot, because it is the key a marketplace, a
  permalink, and a report-back thread all hang on. An adopted seed keeps its origin's id.
- **`name`** — the slug: a renamable, human-friendly handle. Identity lives in the id, not here.
- **`version`** (semver) — the **seed's own** version, and the loop's anchor. A seed evolves as the
  target's install story changes; the reader holds only the exported file, so the version travels
  **in** the document, a report-back binds to it, and the scars a reader hits produce the next one.
- **`title`** — one line a human can read to know what this brings in.
- **`target`** (`<tool>@<version>`) — **the subject of the whole seed**, named and pinned. The
  version is not decoration: a seed proven against one release is not a promise about the next, and
  pinning is how the reader knows which promise they are holding. With `source`, this is one of the
  two places a product may be named — it is the seed's subject, not a coordinate.
- **`source`** — the **canonical origin the seed vouches for**: where the tool genuinely comes
  from, an origin the reader (agent and human) can inspect. This is the trust anchor under the
  entire `## Install` section — every acquisition step, and especially any pipe into a shell, must
  trace back to this declared origin (§2). A seed whose `source` is vague is one whose installs
  cannot be audited.
- **`stack`** — an honesty stamp: what the **verifying install actually ran on** (the operating
  system, the architecture, the runtime), stated so the reader can weigh how far their own ground
  is from it. It carries a `language` (or the equivalent runtime axis); the gate holds the stamp to
  that shape so it cannot decay into a mood.
- **`verified`** — the other honesty stamp: this seed is a **snapshot of one successful install**
  — `{project, release, date}` — not maintained doctrine. It says *whose* machine proved it, at
  *which* release, and *when*, so the reader can tell a fresh proof from a stale claim.
- **`effort`** — an honest signal of what standing this up costs, so the reader can budget before
  they start rather than discover it three steps in.

The frontmatter is the one place provenance lives; the body is where the moves happen. Do not put
the reader's own coordinates in either.

## 2. The trust contract — the seed's teeth

A recipe's teeth guard its **neutrality**; a seed's teeth guard the **safety and honesty of an
action taken on a foreign machine**. Four rules, each mechanical, each with a gate.

**Detect before you install — the first step opens with `**Detect:**`.** An agent dropped into a
repository does not know whether the tool is already there. If it installs blindly it clobbers a
working copy, stacks a second one beside the first, or fails on a conflict it created — and it does
all of that as a *mutation* the human never asked for. The first `###` step under `## Install`
therefore opens with the `**Detect:**` marker and answers one question: **is the target already
present, and at what version?** Detect-first is what makes a seed **idempotent** — safe to run
twice, safe to run in a repository that already has the tool — which is the difference between a
seed and a script that assumes a blank machine.

**Every install step closes with `**Done when:**`.** Like a recipe's build sequence, an install
step with no acceptance is a step the agent cannot tell succeeded. Installs fail in the middle in
quiet ways — a package manager reports success while the shell's lookup cache still points at
nothing, a binary lands without its execute bit, a version pin is silently ignored. So every `###`
step under `## Install` — the detect step included — ends with a literal `**Done when:**` line
stating the observable condition that proves it took. The marker is literal and the gate counts it
against the steps, because "each step has an acceptance" is otherwise a promise nobody checks.

**No blind pipe into a shell — unless the step cites `source`.** The most dangerous move a seed can
make is to fetch a remote script and pipe it straight into an interpreter: it executes arbitrary
code, from the network, with the reader's privileges, where neither the agent nor the human can see
what runs before it has run. The genre does **not** ban the pattern outright — for many real tools
it is the official install path — but it requires the step to **cite the `source` origin the
frontmatter declares**, so the human can audit the move as *"this script comes from the origin this
seed vouches for"* rather than *"this script comes from wherever."* A pipe into a shell with no
cited origin is an unaccountable remote execution, and the gate reds it. Citing the declared
`source` is the whole escape; there is no other.

**Verify holds at least one runnable proof.** `## Verify` must contain a **fenced command block** —
a check the agent runs and observes. This is the acceptance for the tool as a whole, the way a
recipe ships one live check that says it really works. Prose that asserts the tool now runs proves
nothing; a command whose output the agent can read is the difference between *"I ran the
installer"* and *"the tool actually runs here."*

Together these four turn an install from a hopeful sequence of commands into an **accountable** one:
it does not clobber (detect), it does not proceed on faith (done-when), it does not execute the
untraceable (cited source), and it does not claim success it never observed (verify).

## 3. Neutrality — the tool is named, the reader's tree is not

A seed inverts the recipe's neutrality without abandoning it. The recipe names **nothing** concrete
in its body, because its subject is a transferable *pattern*. A seed's subject is **one concrete
tool**, so it names that tool freely — the target, its version, its origin, its own install command
all belong in the document, because they *are* the document. What a seed must still not name is the
**reader's tree**: their paths, their filenames, their config locations, the directory where their
harness lives. The test is the recipe's test, pointed at a different object: *could an agent follow
this in a repository whose layout you have never seen?*

- The **tool's install command and its `source` origin** are the subject; they live in fenced
  command blocks and are meant to be concrete. The gate erases URLs before it scans, and fenced
  command lines carry no backticked coordinates, so the install line and the origin pass.
- The **reader's own coordinates** — where the binary should sit on their machine, where their
  agent rules live, what their config file is called — are what varies from repository to
  repository, and naming them assumes a tree you cannot see. Say them as **roles**: *the reader's
  executable path*, *wherever this repository keeps its agent rules*, *the project's own config* —
  never a literal path or filename lifted from your own machine.
- A **backticked path or filename in the prose body still reds** — with one carve-out the
  engine now honors: a coordinate whose path segment or filename **stem is the target tool's own
  name** (segment-stem equality against the declared `target`) is the seed's own subject and
  passes; **every other coordinate — the reader's tree first of all — still reds**, exactly as it
  does for a recipe. The engine exempts the one tool it was told to name and nothing more: it
  erases the target's own artifacts from the line before the coordinate scan, so the tool's own
  directory and files pass while an unrelated path, or a lookalike that merely starts with the
  tool's name, still reds. So the discipline is: concrete about the **tool**, role-only about the
  **tree**.

This is why the target and source live in **frontmatter** — provenance, named once — while the
body speaks in roles about everything that belongs to the reader.

## 4. The Report is for the human — five fixed rows

`## Summary` orients the agent; `## Report` is written for the **person** who has to accept what the
seed just did. It is the seed's analogue of a recipe's closing human section, but where a recipe
*describes a design* and can do so in prose, a seed *took an action on the reader's tree* — so its
closing section is a **fixed-shape audit**, not free text. The reason is the reader's eye: a human
reviews many seed runs, and a report whose shape changes every time is one they must re-read from
scratch every time. A fixed table is a contract with that eye — the human always knows exactly
where the audit of *what changed* sits, and exactly where the *what next* sits.

The `## Report` section is a Markdown table whose data rows are **exactly these five, in this
order**:

| row | what it answers |
|---|---|
| **what it is** | the tool, in one line the human can act on |
| **how it works** | the mechanism, enough that the human can reason about it |
| **what was done** | the actual mutations this run made to **this** repository — the audit |
| **how to use it** | the human's own next move with the now-installed tool |
| **suggest next** | where to go from here — the forward pointer |

The five walk the human from **understanding** (what it is, how it works) through the **audit** (what
was done — the row that makes the seed accountable) to **action and future** (how to use it, suggest
next). The gate holds the shape hard: an **extra** row, a **missing** row, and a **reordered**
column all red, because a report that drifts from the five is one the human can no longer read at a
glance — which is the one thing the fixed shape exists to guarantee. The format is not decoration
around the seed; for the human, the format **is** the product.

## 5. The Harness section is advisory

`## Harness` is where the seed considers how the installed tool joins the repository's **agent
harness** — the rules, skills, and config that tell this repo's future agents what exists and how
to reach for it. The section is **required to be present** (so the author cannot silently skip the
question of integration), but its content is **advisory**, and the gate deliberately checks only
that it exists — never what verdict it reaches.

The rule the section follows: **recommend a project-local rule or skill only when the tool ships
none of its own.** Many tools already carry their own agent-facing guidance — their own skill,
their own docs an agent reads, their own harness surface. Bolting a redundant project rule on top
of that is the custom-code-is-debt anti-pattern: a **second source of truth** that drifts from the
tool's own the first time the tool updates. So the section branches on one question — *does the
target ship harness guidance?* If it does, the seed points the reader at it and stops. Only if it
ships **none** does the seed recommend the reader author a minimal project-local rule, and even then
it says to keep it thin. The gate never checks which branch you took, because the right answer
depends entirely on the tool — and forcing a rule where the tool already has one would manufacture
exactly the debt this section exists to prevent.

## 6. Derived from the tool's current truth, not memory

**The raw material of a seed is read and run, not recalled.** A tool's install command, its verify
check, the origin it ships from, the flags it takes — all of it lives in the tool's **own current
documentation** and in what the tool **actually does when you run it**, not in what you remember it
doing at some past release. An agent writing a seed from memory writes the install it saw last, and
silently hands the reader a step the tool changed two releases ago — the failure the `verified`
stamp exists to catch.

So the division of labour mirrors a recipe's, pointed at a live tool: the agent **reads the tool's
current source of truth and runs the install itself**, and does the one thing that cannot be
automated — **judgment**. Which install path is the essential one and which is an accident of the
author's operating system; which verify check actually proves the tool works versus merely proves a
file exists; whether the tool's own harness guidance is enough or a thin project rule is warranted.
The `verified` stamp is the proof this happened: it says the seed was *run*, on a named machine, at
a named release, on a named date — not that it was assembled from recollection.

## 7. Anti-patterns

| ❌ | ✅ |
|---|---|
| **The blind installer** — the sequence installs without detecting first | the first step is `**Detect:**`; the seed is idempotent and never clobbers a working tree |
| **The hopeful step** — an install step with no acceptance | every step ends with `**Done when:**`, an observable condition the agent can check |
| **The untraceable pipe** — a remote script piped into a shell with no cited origin | the step cites the `source` the frontmatter vouches for, or it does not pipe at all |
| **The unverified install** — `## Verify` asserts success in prose | `## Verify` runs at least one fenced command and reads its output |
| **The manual in disguise** — the body names the reader's paths, files, and config | the tool is named; the reader's tree is spoken of only in roles |
| **The recipe in disguise** — a "seed" for a capability with no external tool | if there is no tool to acquire, it is a recipe — write the recipe |
| **The drifting report** — `## Report` reshaped, rows added or reordered | exactly the five rows, in order: what it is / how it works / what was done / how to use it / suggest next |
| **The redundant rule** — a project rule authored for a tool that ships its own | point at the tool's own guidance; author a rule only when it ships none |
| **The seed from memory** — install steps recalled, not run | read the tool's current docs, run the install, stamp `verified` with the machine that proved it |
| **The eternal seed** — presented as maintained truth | it is a snapshot; the stamp says whose machine, which release, and when |

## 8. Authoring checklist

- [ ] Frontmatter carries `id`, `name`, `version`, `title`, `target`, `source`, `stack`,
      `verified`, `effort` — `id` is the ULID the tool minted (never typed), `target` is
      `<tool>@<version>` pinned, and `source` is an origin the reader can inspect.
- [ ] All seven body sections are present, in order: `## Summary`, `## What it is`, `## Install`,
      `## Verify`, `## Use`, `## Harness`, `## Report`.
- [ ] `## Summary` is a 2–4 sentence orientation of at least 80 visible characters — an argument,
      not a title.
- [ ] `## Install`'s first `###` step opens with `**Detect:**` and answers *is it already here, at
      what version?*
- [ ] **Every** `## Install` step ends with a `**Done when:**` line stating an observable
      acceptance.
- [ ] No fetch-and-pipe into a shell anywhere in `## Install` unless the step cites the `source`
      origin the frontmatter declares.
- [ ] `## Verify` contains at least one fenced command block — a proof the agent runs, not prose.
- [ ] The body names the **target** concretely and the **reader's tree** only in roles — no
      backticked paths, filenames, or config locations lifted from your own machine.
- [ ] `## Harness` is present and follows the advisory rule: recommend a project rule or skill only
      when the tool ships none of its own.
- [ ] `## Report` is a table of **exactly** the five rows, in order — no extra, no omission, no
      reorder.
- [ ] The install was **run**, not recalled, and `verified` is stamped with the machine, release,
      and date that proved it.

## 9. Cross-links

The seed genre is the recipe genre's sibling, born from it and inverted. Where the recipe teaches an
agent to **build** a capability from portable principles and names nothing concrete, the seed
teaches an agent to **acquire** one named tool and stays concrete about exactly that tool while
neutral about the reader's tree. The two share the honesty machinery — a `verified` snapshot stamp,
a `stack` ground stamp, a version that travels in the document — and the same discipline against
memory: a seed's install story is read from the tool's current truth and proven by running it, the
way a recipe's scars are harvested from the record rather than recalled.

The seed genre and its shape are enforced by the same binary that ships the corpus, so any consumer
can hold a seed about its **own** repository to this contract. The genre spec travels wherever the
binary goes.

## 10. Version history

### 1.0.1 — the neutrality carve-out for the target tool

Narrowed §3 bullet 3: the neutrality engine now honors a per-seed carve-out, exempting a coordinate
whose path segment or filename stem equals the declared `target` tool's own name (segment-stem
equality) while every other coordinate — the reader's tree first of all — still reds. The genre's
seven-section shape and nine frontmatter keys are unchanged; only the constitutional prose about how
the engine treats the target's own artifacts is clarified to match the engine.

### 1.0.0 — the original sealed genre

Established the fixed seven-section body (`## Summary`, `## What it is`, `## Install`, `## Verify`,
`## Use`, `## Harness`, `## Report`), the nine frontmatter keys, and the trust contract: an
`## Install` that opens with a `**Detect:**` step and carries a `**Done when:**` on every step, a
ban on piping a remote script into a shell without citing the declared `source`, a `## Verify` that
holds at least one runnable proof, a `## Report` fixed at exactly five rows written for the human,
and an advisory `## Harness` section whose presence is required but whose verdict is the author's.
