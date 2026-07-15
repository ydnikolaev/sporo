<!-- SSOT SOURCE (sporo repo). Consumers receive a provenance-stamped copy via `sporo update` — edit HERE, never the synced copies. -->

---
name: _authoring
title: Authoring a recipe — the canonical shape of a transferable build
problem: A capability we built cannot be rebuilt by anyone who did not watch us build it.
prerequisites: [read-files, edit-files]
derived_from: [the recipe genre itself]
stack: { language: markdown, runtime: any, why: "the genre spec is prose about prose" }
verified: { project: mate, release: v0.84.0, date: 2026-07-15 }
effort: reference
---

# Recipe — authoring

> **Thesis.** A recipe teaches an **agent in a repository you have never seen** how to
> arrive at a capability you already have: the principles it rests on, the ground it needs
> underneath, the obvious road that dead-ends, the scars, and the price. Its reader is a
> machine that will implement — possibly on another stack, under other constraints — so the
> document is written *for* that machine, and only its closing section is written for a
> person. Transferability is not a style here, it is a **hard constraint with teeth**: the
> moment a recipe names your files, your commands or your product, it stops being
> reproducible anywhere else and quietly becomes a manual for the one project that already
> has the thing. Every rule below exists to keep that from happening.

## 0. When to write one (and when not)

Write a recipe when **a capability you built is worth having twice** — in another project,
in another company, by an agent with no access to your repository. The trigger is a build
that (a) took real design decisions, not just typing, and (b) produced **scars**: things
that went wrong in ways the next builder would repeat.

Do **not** write one for: a feature with no transferable design (it was obvious, and the
recipe would say so at length); a capability that is one library call (the recipe is the
install command); or a thing you have not actually finished — a recipe written from an
intention is a plan with the word "recipe" on it, and it will teach the next agent your
untested guesses as if they were earned.

**One capability, one recipe.** The genre's teeth — one acceptance at the top, a `**Done
when:**` on every step, a ladder under every precondition — only hold at the scale of one
capability. A whole harness or a whole repository is not one capability; it is several,
standing on shared ground. Write each as its own recipe and let the **delivery step compose**
a set into the single handed-over document (shared ground first, members in build order, one
adoption protocol at the end). Inflating one recipe to repository scale produces a document
whose steps cannot be individually accepted — which is a wish list with a table of contents.

**A recipe is not a doctrine and not a manual.** The three genres do not overlap, and
mistaking one for another is the failure this section exists to prevent:

| genre | answers | audience | names your files? |
|---|---|---|---|
| **doctrine** | what must be true **always** | an agent working *under* the doctrine | never — it is principle |
| **manual / handbook** | how to operate **this** system | the operator of that system | **yes, that is its job** |
| **recipe** | how to get from **nothing to having it**, once, anywhere | an agent in a foreign repository | **never — that is the constraint** |

A doctrine has no build sequence: it is normative, and it holds after the thing exists. A
manual is worthless outside the system it documents. A recipe is the only one of the three
that has to survive being handed to a stranger. If a section of your recipe is a normative
claim with no build step under it, it is doctrine — **cite the principle and move on**; do
not restate a corpus inside a recipe.

## 1. The shape — eleven sections, in this order

Every recipe is these parts, in this order. The shape is **gated**, because a genre defined
only by taste drifts into whatever the last author felt like writing.

1. **Frontmatter** — `name`, `version` (semver), `title`, `problem` (one sentence),
   `prerequisites` (capabilities, never tool names), `derived_from`, `stack`, `verified:
   {project, release, date}`, `effort`. Two of these are honesty stamps rather than metadata:
   `verified` says a recipe is a **snapshot of one successful build** (whose, which release,
   when) and not maintained doctrine; `stack` records what that build actually ran on, so the
   reader can weigh how far their own ground is from it. `version` is the loop's anchor: the
   exported file is the only thing the reader has, so the version travels **in** the document,
   a report-back binds to it, and new scars from readers are what produce the next one. The
   frontmatter is the **one place a product may be named** — it is provenance, not
   instruction.
2. **`## The problem`** — what you do not have, **what the output is**, and **how you will
   know you have it**. The acceptance goes here, at the top, in the reader's terms. A recipe
   whose success condition is vague produces a build nobody can call finished.
3. **`## Why the obvious approach fails`** — the fork in the road. State the design the next
   agent *will* reach for first, and the concrete way it breaks. This section is what makes
   a recipe worth more than the reader's own first instinct; without it, they will simply
   follow that instinct.
4. **`## The principles`** — **the payload.** The load-bearing claims, portable, each one
   survivable on a stack you have never seen. Everything below is an instance of something
   here; everything above is why you need it. Where a claim already lives in a doctrine,
   name the principle in one line and cite it rather than re-deriving it.
5. **`## The ground it needs`** — what must be **standing underneath** before the sequence
   can start, and **why each one is load-bearing**: the single source of truth this
   capability derives from, the structure it expects to write into, the gates and always-on
   rules that keep it from rotting the day after it ships. Say the *why* in the reader's
   language, not as an appeal to your own conventions — an agent that does not understand
   why it needs an SSOT will build the capability on top of prose and lose it in a month.
   This is a **precondition** section, not a variable one (the seams section below is the
   variable one — the two bleed into each other if you let them). Write it as a **ladder**, not
   a statement (§5): most readers will not have what it asks for, and a precondition with no
   rung under it is one they will quietly substitute rather than build.
6. **`## The contracts`** — the shapes this capability consumes and emits, **shown**, in fenced
   blocks (§4). Not described: *shown*. This is the section a recipe cannot fake, and the gate
   requires at least one fence in it.
7. **`## The build sequence`** — the steps, one `###` heading each, and **every step ends
   with a `**Done when:**` line**. A sequence without per-step acceptance is a wish list: the
   reader cannot tell whether they are on track until the end, which is exactly when it is
   expensive to find out. The marker is literal and the gate counts it against the steps,
   because "each step has an acceptance" is otherwise a promise nobody checks.
8. **`## The seams`** — what MUST stay configurable, and why. This is where the recipe
   protects the next project from inheriting *your* values: the paths, the thresholds, the
   vocabulary. Name the seam and what varies across it, never the value. **The ground is what
   the reader stands up underneath the capability; the seams are what they are free to swap
   inside it.**
9. **`## The scars`** — what actually went wrong. One `###` heading per scar, each carrying
   the three literal markers **`**Symptom:**`**, **`**Root cause:**`** and **`**Fix:**`** —
   the shape is fixed so the section cannot decay into a paragraph of regret, and so an agent
   can lift a scar out whole. Each scar must be **earned**: one invented to look thorough
   teaches a defence against a failure that does not exist and costs the reader real work.
   This is the highest-value section in the document, and the one a clean-room
   reimplementation cannot produce.
10. **`## Verification`** — the gates that must ship *with* the capability (a capability whose
    invariants are unguarded rots back to nothing), and the one live check that says it really
    works.
11. **`## The trade-offs`** — what this **design** costs, what it deliberately refuses, and
    **when not to build it at all**. A recipe that only advocates is marketing. (The cost of
    the *stack* is a different axis and belongs in the closing section.)
12. **`## For the human`** — the closing section, and **the only one written for a person**.
    See §3 below: it is not a summary of the document, it is the author's declaration of what
    was built, what it runs on, and what that choice bought and cost. It does **not** restate
    the trade-offs — the same question asked of the stack rather than of the design.
13. **`## Appendix`** — optional, and the **only** section where concrete names, paths and
    commands are allowed. It is explicitly an illustration, and everything above it must
    stand without it.

**Two sections you do not write, and must not.** `## Adopt it here` and `## Report back` — the
protocol for getting the capability into a repository the author has never seen, and the
channel that sends back what only the reader can know — are **identical for every recipe**.
They are authored once for the whole corpus and **appended by the delivery step**, so one
improvement reaches every recipe at once and no author maintains a private copy that drifts.
Write neither. The recipe-specific half of adoption already lives in the ground ladder: what
each precondition needs, and how to degrade honestly when it is missing.

**What the reader is handed is the EXPORTED file, not the source.** The export is composed —
banner stripped, protocol appended — so a recipe handed over as a raw source file arrives
without the only section addressed to its reader. Export, then hand over.

## 2. The neutrality constraint — technologies are free, coordinates never are

**The body names roles, never instances.** "The facts file", "the collector", "the optional
target the project implements" — not a path, not a filename, not a command of one specific
tool, not a product's name. The test is mechanical and a gate applies it: could an agent
follow this in a repository that has never heard of yours?

The line that is easy to get wrong: **the ban is on coordinates, not on concreteness.**

- A **technology named as a choice with a rationale** — "a single statically-linked binary,
  because the reader can run it with no checkout of the source" — is *portable*: an agent on
  another stack can weigh it and accept or reject it. Naming it is the job of the ground
  ladder and of the closing For-the-human section.
- A **coordinate** — a path, a filename, a product, a command that only exists in your tree
  — executes in one repository and transfers to none. It is banned **everywhere in the
  body**, including in the sections that talk about the stack. There is no zone where a
  coordinate becomes acceptable; if the sentence needs one, the sentence is about your repo.
- A **contract** — a schema, a normalized shape, a declared surface — is *portable*, and this
  is the distinction the genre got wrong for a whole release. A coordinate **executes** in one
  repository; a contract is a shape the reader **copies and adapts**, and it transfers exactly
  as well as the principle it serves. Banning both is how a genre defends its neutrality into
  uselessness — see §4.
- The **scars** stay concrete. "Be careful with concurrency" teaches nothing. "Summing every
  parallel agent session produced 25.3 hours inside a 24-hour day" teaches in one sentence,
  and it names no file. **A recipe with no specific failures in it is not neutral — it is
  empty.**

## 3. The closing section is for the human — and it is about the stack

Everything above the closing section is written for the agent that will implement. **The
closing section is written for the person who has to decide whether to let it.** (Numbered
cross-references rotted once already — the contracts section shifted every number after it —
so this document points by name.) It is deliberately last: the machine payload comes first,
and a human preamble in front of it would only push the principles down the page. It
carries, declaratively and without a build step in sight:

- **What this is** — in plain language: what gets built, on what principles, what you have
  at the end. Someone who reads only this section should be able to say what they would get.
- **The stack it was originally built on, and why that is the recommendation.** Name the
  language, the runtime, the shape of the artifacts. Then separate the two kinds of reason,
  because they are not the same and the reader must be able to tell them apart:
  - **essential** — the choice the design actually depends on (a durable machine-readable
    record; a check that runs on one command with no services up). Replace it and the
    capability degrades.
  - **incidental** — the choice that came from the author's existing house (this language,
    this test runner, this directory habit). An agent on another stack should swap it
    without a second thought, and saying so out loud is what keeps the recipe honest.
- **The trade-offs of that stack, stated as trade-offs** — what it bought, what it cost, what
  it made hard. Not advocacy: a reader who is on a different stack needs to know which of
  your pains they inherit and which they escape.

A recipe whose stack section reads as a recommendation with no cost attached is one an agent
will follow into a wall.

## 4. Show the shape — the contracts section

**A shape described in prose is a shape every reader re-invents, incompatibly.** "The
collector emits a slug, a title, phases done out of total, and what blocks it" reads as
complete to the author, who has the file open in another window. To the reader it is a design
task, and two readers who do it will produce two shapes that cannot talk to each other — which
destroys the one thing a contract exists for.

So the contracts section **shows** them, in fenced blocks: the record the capability persists,
the shape it consumes from the project, the surface it declares about anything external. Field
names in the reader's language, placeholders where a value would be a coordinate, and a comment
on any field whose *meaning* is a trap. Then the build sequence is written against those
shapes rather than against an implementation, and the reader can start from something.

This is the section a recipe cannot fake, and the gate requires at least one fence in it. It
exists because a real implementer scored the corpus's first recipe **3/10 on copy-paste
artifacts** while every other axis scored high: the document taught him what to build and
handed him not one shape to build it against.

Two things it is not:

- **Not the appendix.** A contract is the shape; the appendix is one filled instance of it,
  from the author's own tree, marked as illustration. Show the shape above; put the real,
  populated one below if it illuminates.
- **Not a licence.** A fenced block is not a sanctuary: a product name, a real path, a real
  filename inside an example is still a coordinate, and it is *likelier* there than anywhere
  else, because an example gets copied out of a working tree. The gate scans fences like any
  other line, and it does not need an exemption to let a schema through — the coordinate
  patterns match instances, and a shape is not one.

## 5. The ground is a ladder, not a statement

Most readers will not have what the ground section asks for. That is the normal case — the
whole premise of the genre is that the reader's repository is not the author's — and a
precondition stated as a fact leaves them exactly one honest option, which is to stop. So they
will not stop. They will point the capability at whatever weaker source is lying around, and
the capability will produce confident output over rot.

Write each precondition as three rungs:

1. **Probe** — how to find out whether this repository already has one, in terms of what it
   would *do*, not what it would be called. ("Look for a machine-readable state that a check
   or a build already reads" — not "look for a tracker file".)
2. **Build the smallest one** — if it is absent, the minimum version that works, whose shape
   is in the contracts section. A precondition worth a ladder is usually twenty lines.
3. **Degrade, and label it in the output** — if the team will not keep even that, the honest
   fallback, *and the label it must carry where the reader of the capability's output can see
   it*. A degraded source that is not labelled downstream is indistinguishable from the real
   one, which is the failure the recipe was written to prevent.

There is no fourth rung, and the recipe should say so.

## 6. Derived from the record, not from memory

**The raw material of a recipe is harvested, not recalled.** The principles are already in
the doctrine; the ground is already in the structure standard and the gate registry; the
decisions are already in the decision log and the knowledge base; the sequence and the *why*
are already in the release record; the scars are already in the fix commits — including the
ones whose message says the defect was found only by deliberately breaking the thing. An
agent writing a recipe from memory writes the last week of it, confidently, and silently
loses the failure from three weeks ago that the reader most needs.

So: the machine gathers the record, and the author does the one thing a machine cannot —
**abstraction**. Deciding which failure was incidental and which was structural, which step
was essential and which was an accident of the stack: that is judgment, and it is the whole
value the author adds. It is the same division of labour a good progress report makes (facts
derived, judgment written), for the same reason.

## 7. Anti-patterns

| ❌ | ✅ |
|---|---|
| **The manual in disguise** — the body names files, commands, or the product | roles only in the body; instances live in the appendix, marked as illustration |
| **The contract in prose** — "it emits a slug, a title, and the counts" | show the shape in a fenced block; the reader copies it instead of re-inventing it |
| **The ground stated as a fact** — "you need a machine-readable work-item state" | a ladder: probe → build the smallest one → degrade with a label; there is no fourth rung |
| **The stack section that lists technologies** | it names them *and* says which are essential, which incidental, and what each cost |
| **The scar-free recipe** — smooth, confident, nothing ever went wrong | the failures, with symptom → root cause → fix; a build with no scars did not need a recipe |
| **The doctrine restated** — normative claims with no build step | cite the principle, keep the sequence |
| **The abstraction fog** — "handle errors carefully", "consider performance" | concrete failure, coordinates stripped: *what* broke and *what it produced* |
| **The recipe from an intention** — written before the thing worked | write it after it works, and stamp `verified` with the build that proves it |
| **The unverifiable step** — a sequence step with no acceptance | every step says how the reader knows it is done |
| **The ground assumed** — the sequence starts as if an SSOT and a gate suite already exist | §5 names what must be standing first, and why it is load-bearing |
| **The eternal recipe** — presented as maintained truth | it is a snapshot; the stamp says whose and when |

## 8. The hazards a capability inherits from its genre

Some failures do not belong to *your* build — they belong to any build of that **kind**, and
they are the ones an author reliably fails to mention, because they were solved in an
afternoon and never felt like design. Ask these of the capability you are writing up, and put
what applies into the principles, the sequence and the scars:

- **Does it render text it did not write?** Anything that comes from a record — a commit
  subject, a user's title, a filename — and lands in an artifact that gets **forwarded**, must
  be escaped at the boundary. The first implementer of this corpus's first recipe scored it
  **0/10** on this: the document never said the word.
- **Does it join two sources that answer different questions?** Two collectors with two
  attribution keys will disagree, constantly, while both are working correctly. Say which
  question each answers; never add them.
- **Does it have a time boundary?** Two records with two clocks (one local, one UTC) and a
  "day" nobody defined is a silent mis-attribution, not an error.
- **Does it report a total where it means a measurement?** A sum of concurrent things is not
  an elapsed thing; a byte count is not a work count. Name the method, or drop the number.
- **Does it fail loudly when its own toolchain is broken?** A collector that reports "no data"
  when it means "I am broken" produces a report that looks merely quiet.

The list is a **floor**. What it is really asking: *which failure of this kind did I fix so
early that I forgot it was a decision?*

## 9. Porting checklist

- [ ] Frontmatter carries `name`, `version`, `title`, `problem`, `prerequisites`,
      `derived_from`, `stack`, `verified`, `effort` — and `effort` is honest about which half
      of the build eats the budget.
- [ ] All eleven sections are present, in order (the appendix is optional). `## Adopt it here`
      and `## Report back` are **not** among them — delivery appends those.
- [ ] The body contains **no** repository path, filename, or product name — in **any**
      section, the stack sections and the fenced examples included. Technologies are free,
      contracts are free; coordinates are not.
- [ ] `## The ground it needs` states the SSOT, the structure, and the gates the capability
      rests on — **why** each is load-bearing, and a **ladder** under each: probe → build the
      smallest → degrade with a label.
- [ ] `## The contracts` **shows** every shape the capability consumes or emits, in fenced
      blocks, with a note on any field whose meaning is a trap.
- [ ] The build sequence is written against those contracts — an **interface**, not an
      implementation. A reader on another stack must be able to follow it.
- [ ] Every build-sequence step ends with a `**Done when:**` line.
- [ ] Every scar is a real one and carries `**Symptom:**`, `**Root cause:**`, `**Fix:**`.
- [ ] The genre hazards in §8 have been asked, and the ones that apply are in the document.
- [ ] The trade-offs section says when NOT to build this.
- [ ] `## For the human` says what gets built in plain language, names the original stack,
      splits **essential** from **incidental**, and states what that stack cost — without
      restating the trade-offs.
- [ ] The verification section names the gates that ship with the capability.
- [ ] The lint gate (`sporo lint`) is green on the corpus.
- [ ] The recipe is **sealed** (`sporo seal`) — version and content hash recorded, so the
      text a reader cites can never silently mutate under them.
- [ ] You have **exported** it and read what the stranger actually receives.

## 10. Cross-links

The genre grew out of a harness's doctrine corpus, and three of its debts are worth naming
without sending the reader to documents they do not have. The rule that knowledge which must
travel may not name its origin comes from that corpus's documentation principles — this spec
is its executable form. The harvest-then-judge division (the machine gathers the record, the
author does the abstraction) is borrowed from the same corpus's reporting principles, for
the same reason: facts derived, judgment written and labelled. And the ground section asks
the reader to stand up exactly the structure a well-kept project already has — a
machine-readable source of truth, a decision log, gates.

The adoption protocol appended to every export is authored ONCE for the whole corpus (in the
corpus's own `_adoption` document) — edit it there, never in a recipe. Recipes are authored
through the sporo-recipe skill and checked by `sporo lint`, which the binary carries so any
consumer can run it on a recipe about its **own** repository; `sporo genre` prints this spec
wherever the binary goes.
