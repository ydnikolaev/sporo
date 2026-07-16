# What is a recipe

You've read the pitch. Here's the actual genre — the fixed shape every recipe has to
take, and why each part of it exists. If you're the senior engineer who's going to
decide whether to let an agent act on one of these, this is the page that earns that.

## The one-sentence definition

A recipe teaches an agent in a repository it has never seen how to arrive at a
capability you already built — the principles it rests on, the ground it needs
underneath, the road that looks obvious and dead-ends, the scars, and the price. It is
written for the machine that will implement it. Exactly one section, at the end, is
written for you.

## Why it has to be neutral, and what that actually bans

The hardest constraint in the genre isn't length or tone. It's this: the moment a
recipe names your files, your commands, or your product, it stops being reproducible
anywhere else and quietly turns into a manual for the one project that already has the
thing.

So the rule is: **name roles, never instances.** "The facts file," "the collector," "the
optional target the project implements" — never a path, a filename, or a product name.
The test is mechanical: could an agent follow this sentence in a repository that has
never heard of yours? If the sentence needs a coordinate to make sense, the sentence is
about your repo, not about the capability.

This is easy to over-apply, and the genre draws the line carefully:

- **A technology named as a choice with a rationale is fine.** "A single statically-
  linked binary, because the reader can run it with no checkout of the source" is
  portable — another agent can weigh that reasoning on its own stack and take it or
  leave it.
- **A coordinate is never fine.** A path, a filename, a command that only exists in your
  tree executes in exactly one repository. There's no section where it becomes
  acceptable — not the stack notes, not the examples.
- **A contract is not a coordinate.** A schema or a declared shape is something the
  reader copies and adapts; it travels as well as the principle it serves. Banning
  contracts alongside coordinates is a mistake the genre made once and fixed — a recipe
  that won't show you a shape teaches you what to build and hands you nothing to build
  it against.
- **Scars stay concrete, on purpose.** "Be careful with concurrency" teaches nothing.
  "Summing every parallel session produced 25.3 hours inside a 24-hour day" teaches in
  one sentence — and it doesn't name a single file. A recipe with no specific failure in
  it isn't neutral. It's empty.

## The eleven sections, and what each one is actually for

A recipe isn't free-form prose. It's a fixed shape, checked by a gate, in this order:

1. **Frontmatter** — name, title, the problem in one sentence, prerequisites (stated as
   capabilities, never tool names), what it's derived from, the stack it was verified
   on, and when. This is the one place a product may be named — it's provenance, not
   instruction. `verified: {project, release, date}` exists to say out loud: this is a
   snapshot of one successful build, not maintained doctrine.
2. **The problem** — what you don't have, what the output actually is, and how you'll
   know you have it. The acceptance criteria live here, at the top, in the reader's
   terms — not buried at the end where a vague success condition becomes obvious only
   after the build is already expensive to redo.
3. **Why the obvious approach fails** — the fork in the road, named honestly. This is
   the section that makes a recipe worth more than the reader's own first instinct;
   without it, they just follow that instinct into the same dead end you already found.
4. **The principles** — the payload. The load-bearing, portable claims that survive on a
   stack nobody's tested them on yet. Everything below is an instance of something here.
5. **The ground it needs** — the preconditions, and this is where most recipes would
   fail if they stated them as facts instead of a ladder (more below).
6. **The contracts** — the shapes this capability consumes and emits, shown in fenced
   blocks, not described in prose. The gate requires at least one fence here, because a
   real implementer once scored a recipe 3/10 on this exact axis while everything else
   scored high: it taught him what to build and handed him nothing to build against.
7. **The build sequence** — one heading per step, and every step ends with a literal
   `**Done when:**` line. A sequence without per-step acceptance is a wish list — you
   can't tell you're off track until the end, which is the most expensive place to find
   out.
8. **The seams** — what must stay configurable, and why: the paths, the thresholds, the
   vocabulary that belong to the reader's project, not yours.
9. **The scars** — the highest-value section in the document, and the one a clean-room
   reimplementation cannot produce on its own. Each one carries three literal markers:
   **Symptom**, **Root cause**, **Fix**. The shape is fixed so it can't decay into a
   paragraph of vague regret, and each scar has to be earned — an invented one teaches a
   defense against a failure that doesn't exist, and costs the reader real work chasing
   it.
10. **Verification** — the gates that ship with the capability, plus the one live check
    that says it actually works. Not a promise. A test.
11. **The trade-offs** — what the design costs, what it refuses, and when not to build
    it at all. A recipe that only advocates is marketing wearing a build sequence.

Two more things happen around the eleven sections, and neither is authored per-recipe:
the closing section written for a human (below), and an appendix — the only place
concrete names and paths are allowed, explicitly marked as illustration, and everything
above it has to stand without it.

## Why the ground is a ladder, not a statement

Most readers will not have what a precondition asks for. That's not an edge case — it's
the entire premise of the genre, because the reader's repository is, by definition, not
the author's. State a precondition as a bare fact ("you need a machine-readable work-
item state") and you've left the reader exactly one honest option: stop. They won't
stop. They'll point the capability at whatever weaker thing is lying around, and it will
produce confident output over rot.

So every precondition is written as three rungs, and there is no fourth:

1. **Probe** — how to check whether this repository already has it, described by what
   it would *do*, not what it would be called.
2. **Build the smallest one** — if it's absent, the minimum version that works, whose
   shape is already in the contracts section.
3. **Degrade, and label it** — if the team won't keep even that, the honest fallback,
   with a label that follows the output all the way to wherever the reader of *that*
   sees it. An unlabeled degraded source is indistinguishable from the real thing — which
   is the exact failure the ladder exists to prevent.

## The two sections you never write

Every recipe ends with the same two sections, appended at export time, never authored
per-recipe: **Adopt it here** (probe your repo, map roles onto your own conventions,
never silently substitute a weaker source, state the plan before building) and
**Report back** (what you built on, what you degraded, the new scars you hit, anything
the recipe got wrong). They're identical across the corpus on purpose — one improvement
to the protocol reaches every recipe at once, and no author maintains a private copy
that drifts from everyone else's.

This is also the mechanism that makes recipes get better over time instead of going
stale the way documentation usually does. A report-back is a new scar. A new scar bumps
the recipe's version. The recipe a stranger reads next month has already absorbed a
failure you hit this month.

## What you're actually being handed

The exported file — never the source. The export strips the authoring banner and
appends the adoption protocol, so what a stranger receives is complete: the machine
payload first, the human closing section last, and instructions for how to get it into
a repository that has never heard of yours.
