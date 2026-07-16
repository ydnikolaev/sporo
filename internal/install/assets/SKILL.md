---
name: sporo-recipe
description: Distil a capability THIS repository built into a recipe — one self-contained file that teaches an agent in a repository that has never heard of this one how to build the same thing, on the same principles, possibly on a different stack, without repeating the same mistakes. Use after shipping something worth having twice.
argument-hint: "[the capability, and the release or revision range it was built in]"
---

# sporo-recipe

> **Thesis.** The knowledge that makes a build repeatable does not survive on its own. It
> lives in the release notes nobody re-reads, in the decision log nobody revisits, and — the
> part that matters most — in the failures, which are precisely the part an author writing
> from memory rounds off. A recipe is that knowledge, extracted while it is still true and
> written so it can be **followed somewhere else**, by an **agent**, possibly on a stack you
> have never used. Your job is the extraction and the abstraction; the machine's job is to
> make sure you are working from the record and not from your recollection of it.

## Where it runs

**In whatever repository you are standing in.** A recipe is written *about* a build. The
harvest reads this project's record, the gate runs on this project's corpus, and the recipe
lands in this project's own recipes home (default `.sporo/recipes/`). The corpus the binary
ships is somebody *else's* build, delivered read-only: never write there.

## When to run

- A capability shipped, it took real design decisions, and it is worth having twice — in
  another project, another company, or by an agent with no access to this repository.
- NOT for: a feature with no transferable design; a capability that is one library call; or
  anything not actually finished. A recipe written from an intention teaches untested guesses
  as if they were earned.

Read the genre spec before you write a line — `sporo genre` prints it. The eleven sections,
the neutrality constraint, and the distinction the whole genre rests on: **doctrine says what
must be true, a manual says how to operate this system, and a recipe says how to get from
nothing to having it, anywhere.** If a section of your draft is a normative claim with no
build step under it, it is doctrine: cite the principle and move on.

## Procedure

### 1. Harvest the record — before you remember anything

```
sporo harvest --since <the release before the work> --out harvest.json
```

Read it. It gives you each release's stated rationale, the commits that fixed something, the
gates that shipped with the capability, the doctrine, the decisions and the knowledge base it
moved, and — the highest-value pile — the commits whose own message admits the defect was
found only by deliberately breaking the thing.

Three disciplines about this file:

- **The signals propose; you decide.** A candidate is not a scar. Which failure was
  structural and which was incidental is the judgment that IS the recipe.
- **Read the unsignaled commits too.** The harvest reports how many it had nothing to say
  about. A scar nobody wrote down is still a scar, and it is exactly the one the machine
  cannot hand you.
- **Read what it says is ABSENT.** It names the records this project does not have (no
  decision log, no gate registry). Those are the sections you will have to source by hand —
  and an empty section nobody noticed is how a recipe quietly loses its verification half.

Then go **read the sources it points at**, not just the commit subjects: the decision log says
what was chosen *against what*, and the knowledge base says what ground the build stands on.
Both are things a commit message is structurally bad at carrying, and both are what the ground
section and the closing human section of the recipe are made of.

### 2. Find the fork — the section that earns the recipe's existence

Before the sequence, write **why the obvious approach fails**. What will the next agent reach
for first, and how, concretely, does it break? A recipe whose reader would have got there
alone is a recipe that did not need writing. This section is usually the hardest and it is
always the most valuable.

### 3. Scaffold, then write the eleven sections in the genre's shape

Do not start from a blank page:

```
sporo new <slug> --from-harvest harvest.json
```

The scaffold arrives as a **draft** (`draft: true`): every section stubbed with a coach
comment saying what belongs in it, the frontmatter pre-stamped, and the harvest's scar
candidates pre-seeded for you to judge — keep the structural, delete the incidental. A draft
is exempt from the gate and refused by seal and export, so nothing half-written can ship;
removing `draft: true` is the act of saying "this document now stands on its own".

The shape you are filling: frontmatter (with `version` and the `stack` and `verified`
stamps) → the problem and its acceptance → why the obvious approach fails → **the
principles** (the payload) → **the ground it needs** → **the contracts** → the build
sequence, every step ending in `**Done when:**` → the seams → the scars → the verification
→ the trade-offs → **for the human**.

You do **not** write the adoption protocol or the return channel. `## Adopt it here` and
`## Report back` are the same in every recipe, so they live once in the corpus and the export
appends them. Writing your own is how a corpus ends up with eleven drifting copies of one text.

Four of these carry the weight, and they are the four an author under time pressure skimps:

- **The principles** are what the agent actually implements from. Everything below is an
  instance of something here.
- **The contracts** are what the reader cannot re-derive. Every shape the capability consumes
  or emits, **shown in a fenced block** — not described in prose. A shape described in prose is
  a shape each reader re-invents, incompatibly, which destroys the one thing a contract is for.
  The first outside implementer of this corpus's first recipe rated it 3/10 on copy-paste
  artifacts and 10/10 on everything conceptual: the document told him precisely what to build
  and handed him nothing to build it against. **The gate requires a fence here** — and note
  that it never forbade one: a fenced schema was always green, because the coordinate patterns
  match instances and a shape is not one. Products stay banned inside a fence, like everywhere.

  **Then ask the one question the shape cannot answer for itself, and mark the answer:** does
  anything OUTSIDE the emitting repository parse this shape — a fleet's shared aggregator,
  another team's tool? Ask your human if you do not know. Yes → `**Binding: exact**` above
  the fence: name that consumer as a role, the reader copies the shape byte-for-byte, and
  changing it later is a MAJOR version (the seal enforces this). No → `**Binding: adapt**`,
  the default: shown so the reader does not re-invent it, local conventions win.

  An exact shape must parse as data, and should ship its **fixtures** — a `**Fixture:
  valid**` fence (a real instance from the verifying build) and `**Fixture: invalid** —
  <why>` fences (the mutations a consumer must reject). `sporo lint` runs them against the
  shape; every reader's CI checks its own output with `sporo conform <recipe> <output>` from
  the exported file alone. That pair is what makes "the schema is the same across the whole
  team" a check instead of an agreement.
- **The ground it needs** is what must be standing *before* step one — the machine-readable
  source of truth this capability derives from, the structure it writes into, the gates and
  always-on rules that keep it honest after you leave — and **why each is load-bearing**, in
  the reader's language. An agent that does not understand why it needs an SSOT will build
  your capability on top of prose and lose it in a month. This is a precondition section; the
  seams section is the *variable* one, and they bleed into each other if you let them.

  Write it as a **ladder**, never as a statement: *probe* for it → *build the smallest one*
  if it is absent → *degrade, and label the degradation in the capability's own output* if the
  team will not keep even that. Most readers will not have what you are asking for — that is
  the premise of the whole genre — and a precondition with no rung under it does not make them
  stop. It makes them silently point the capability at the nearest weaker source, which is
  where most scars in most recipes began.
- **For the human** is the only section written for a person. It says, in plain language,
  what gets built and what you have at the end; it names **the stack it was originally built
  on and why that is the recommendation**, splitting the reasons into **essential** (the
  design depends on it — replace it and the capability degrades) and **incidental** (it came
  from the author's existing house; swap it without a second thought); and it states the
  trade-offs of that stack as trade-offs — what it bought, what it cost. A stack section that
  reads as advocacy with no cost attached is one an agent will follow into a wall.

**Every scar is symptom → root cause → fix**, and every one was earned. A scar invented to
look thorough teaches a defence against a failure that does not exist, and costs the reader
real work.

### 4. Strip the coordinates, keep the substance

This is the pass where a manual becomes a recipe. Go through the body and remove every token
that only means something inside this repository: the file names, the paths, the product
names, the commands of one specific tool. Replace them with **roles** — "the facts file",
"the collector", "the optional target the project implements".

**The line, and it is the one people get wrong in both directions:**

- A **technology** named as a choice with a reason ("a single statically-linked binary,
  because the reader can run it with no checkout of the source") is *portable* — an agent on
  another stack can weigh it and accept or reject it. Naming your stack is **required**, not
  tolerated.
- A **coordinate** — a path, a filename, a product — executes in one repository and transfers
  to none. It is banned everywhere in the body, including in the stack sections and inside a
  fenced example. There is no zone where it becomes acceptable.
- A **contract** — a schema, a normalized shape, a declared surface — is *portable*, and it is
  the distinction this genre got wrong for a release. A coordinate **executes** in one place; a
  contract is a shape the reader **copies and adapts**. Strip the coordinates *out of* your
  contracts; do not strip the contracts.
- **Do not over-correct into fog.** "Be careful with concurrency" teaches nothing; "summing
  every parallel agent session produced 25.3 hours inside a 24-hour day" teaches in one
  sentence and names no file. A recipe with no specific failures in it is not neutral — it is
  empty. If an instance is genuinely illuminating, it belongs in the appendix, the one section
  allowed to be concrete.

### 4b. Ask the genre's hazards, not just your build's

Some failures belong to the *kind* of thing you built, not to your build — and they are the
ones you fixed in an afternoon and forgot were decisions. The spec's hazard list is the floor:
does it render text it did not write (escape it — this corpus's first recipe scored **0/10**
here and the reader hit it live)? does it join two sources that answer different questions?
does it have a time boundary nobody defined? does it report a total where it means a
measurement? does it fail loudly when its own toolchain breaks? Whatever applies goes into the
principles, the sequence and the scars.

### 5. Gate it, then read it as a stranger

```
sporo lint
```

The gate runs on this project's own recipes home and knows this project's forbidden names. It
checks the shape, the per-step acceptance, the scar markers, the neutrality of every line —
and the registry: a sealed recipe whose content drifted without a version bump is a finding.

Then re-read the body **as an agent in a repository on a different stack that has never heard
of this one**. Every step you could not follow from there is a step still written for someone
who already has the thing. The gate catches the coordinates; only you can catch the
assumption.

### 6. Seal it, then ship the exported file — never the source

```
sporo seal <slug>
sporo export <slug>
```

The seal records (version, content hash, provenance) in the project's registry: from this
moment the recipe never silently mutates — every later edit must bump `version:` and re-seal,
and the gate enforces it. One class of edit is stricter: changing an **exact-bound contract**
demands a MAJOR bump — that shape is somebody else's parser, and the seal refuses to let a
break ship under a compatible-looking version. Then the export composes the deliverable: the provenance banner is
stripped, and the adoption protocol — how to map the recipe's roles onto a repository the
author has never seen, and what to send back afterwards — is appended. Hand over *that* file.
A raw source file handed to a reader arrives without the only section addressed to them, and
the reader needs no harness, no binary and no account to use what you gave them.

Read the export once, as the stranger. Then commit with exact-path staging.

### 7. Merge what comes back

A recipe improves in exactly one way: somebody builds it somewhere else and tells you what
happened. The appended protocol asks them for the stack, what they had to degrade, what they
had to invent because you described a shape instead of showing it, and — the payload — the
**new scars**, as symptom → root cause → fix. When a report-back arrives, file it:

```
sporo feedback add <slug> <report file>
```

It is not feedback to archive: it is the next version of the recipe. Merge the scars, add the
contract they had to design, re-stamp `verified`, note in `derived_from` that the recipe is
now derived from more than one build — then bump `version:` and `sporo seal` again. Three of
the scars in this corpus's first recipe were paid for by its first outside reader.

The loop has a reading side too: when THIS repository builds from somebody else's export,
record the handover with `sporo adopt <file>` and check it later with `sporo pull` — which is
loud when an exact-bound contract moved, because that is the one update a consumer-feeding
build must not sleep through.

## The rules you cannot bend

1. **Harvest before you recall.** An agent writing from memory writes the last week of the
   build confidently and loses the failure from three weeks ago that the reader most needs.
2. **A recipe that names its origin is a manual.** It looks fine to the person who wrote it
   and is useless to the only reader who matters.
3. **Name the stack, and say which half of it is incidental.** The reader is on a different
   one, and a recommendation whose reasons are hidden is one they cannot weigh.
4. **No scars, no recipe.** A build with nothing to warn about did not need a document.
5. **Every step says how you know it is done.** Otherwise the reader finds out at the end,
   which is when it is expensive.
6. **Show every shape you tell them to consume.** A contract in prose is a contract each
   reader re-invents, incompatibly — which is the one thing a contract cannot survive.
7. **The ground is a ladder.** A precondition the reader does not have, stated as a fact,
   does not stop them; it makes them substitute something worse, quietly.
8. **The stamp is honest.** A recipe is a snapshot of one successful build, not maintained
   doctrine, and it says whose build and when.

## Output

One file in **this project's** recipes home, green under `sporo lint`, sealed in the registry,
plus a one-line report: what was harvested (candidates, unsignaled commits, absent records),
what was kept, and what you deliberately left out.
