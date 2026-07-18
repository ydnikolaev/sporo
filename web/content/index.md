# A recipe rebuilds the capability. A skill just runs it.

sporo turns a build you already did into a **recipe**: one self-contained file that
teaches an AI agent in a repository that has never seen yours how to build the same
capability — on its own stack, in its own harness, without repeating your scars.

A skill runs in your harness. A recipe rebuilds the capability in any harness.

```
curl -fsSL sporo.dev/install.sh | sh
```

Also coming when the repository goes public: `brew install ydnikolaev/tap/sporo` and
`go install sporo.dev/sporo/cmd/sporo@latest`.

---

## What is a recipe

A recipe is not a prompt. Not a template. Not a packaged skill. It is a transferable,
verifiable build intent — prose an agent reads once and follows to arrive at a
capability you already have, on ground it has never stood on before.

The name comes from spore: the minimal portable unit that carries enough to grow the
organism in a foreign environment. A recipe carries the same minimum — principles,
contracts, scars — and leaves behind everything that only makes sense in your repo.

**The boundary that matters:**

| | a skill | a recipe |
|---|---|---|
| runs where | inside the harness that defines it | anywhere an agent can read and act |
| assumes | your tools, your paths, your conventions | nothing — it probes and maps |
| transfers | by copying the package | by being rebuilt, once, from scratch |
| fails silently when | the harness changes underneath it | failure is designed to be loud — every step carries its own acceptance, every degradation must be labelled |

A skill is fast because it doesn't re-derive anything. A recipe is portable because it
re-derives everything, deliberately, on the reader's own ground.

## How it works

**Author side.** You already built the capability once. sporo mines that fact.

1. `sporo harvest` — mine your repo's own record (decisions, gates, fix commits) for a
   recipe's raw material. The scars come from history, not memory.
2. Author — turn the harvested material into the recipe's fixed shape: the problem, the
   principles, the ground it needs, the contracts it shows, the build sequence, the
   scars.
3. `sporo lint` — gate the genre: shape, scars, neutrality. A recipe that names your
   files instead of roles fails here before it ships to anyone.
4. `sporo seal` — version and content-hash the recipe in the registry, so it never
   silently mutates under a reader who trusted a specific version.
5. `sporo export <slug>` — write one self-contained file (to `.sporo/exports/`), the
   adoption protocol appended; `--stdout` pipes it instead. This is what you hand over —
   never the source.

**Reader side.** The exported file is the only thing a stranger's agent needs.

- **Probe, don't assume.** For every precondition the recipe states, the reader's agent
  checks its own repository before building anything.
- **Map roles, never substitute.** The recipe names *the facts file*, *the collector* —
  roles, not paths. The reader's agent decides where each role lives here, writes the
  mapping down, and never quietly reaches for a weaker source lying around instead.
- **Report back.** When the build is done, the reader sends back what only they could
  know: what they degraded, what broke that the recipe didn't warn about, what turned
  out to be wrong.

**The loop.** Report-backs are new scars. New scars raise the recipe's version. The
recipe you read today is worth more than the one written a month ago, because someone
else already paid down a failure you would otherwise hit yourself. The moat isn't the
file format — it's this loop. A recipe with no report-backs is just a well-written
guess.

## Why prose, not packages

A skill package is opaque until it runs: you trust it or you audit its code. A recipe is
prose. A human can read it, understand exactly what it will do, and decide whether to
let an agent do it — before a single line executes.

That trust isn't a promise about the outcome. It's structural: every build step in a
recipe carries a **Done when:** line, and the recipe ships a **Verification** section
with gates that must exist alongside the capability. sporo doesn't tell you the result
will be correct. It makes the result check itself.

`sporo review` extends the same idea outward: it emits a self-contained review pack that
any AI agent — Claude, Codex, anything — can score against a fixed rubric, provider-
agnostic by construction. The verification doesn't depend on which agent is reading it.

## When not to use it

Recipes are not the right tool everywhere, and pretending otherwise would be the kind of
overclaim this project is built to avoid.

- **Same stack, same harness, same company** — hand over the code, or package a skill.
  A recipe re-derives what you could have just copied; that's waste, not portability.
- **Determinism matters more than adaptability** — if the destination environment is
  identical to yours, zero re-derivation cost beats a document an agent has to
  interpret.
- **The capability is one library call** — the recipe is the install command. Writing
  eleven sections around that teaches nothing.
- **Nothing went wrong building it** — a recipe's highest-value section is its scars. A
  build with no scars didn't need one.

Recipes earn their cost where the destination is genuinely heterogeneous: a different
stack, a different agent, a different company, and code that literally cannot be handed
over. That's the wedge — private teams transferring capabilities across a boundary code
can't cross.

## Get sporo

```
curl -fsSL sporo.dev/install.sh | sh
```

A single static Go binary. macOS, Linux, Windows. No dependencies, no runtime to
install first.

Private pre-release. [What is a recipe →](./what-is-a-recipe.md) · [Read the manifesto →](./manifesto.md)
