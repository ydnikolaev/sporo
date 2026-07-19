<!-- SSOT SOURCE (sporo repo). Consumers receive a provenance-stamped copy via `sporo update` — edit HERE, never the synced copies. -->

---
name: _adoption
version: 1.0.0
title: The adoption and report-back protocol appended to every recipe export
---

<!--
The two sections below are the same in EVERY recipe, so they are not authored in any of
them. They are appended to whatever the delivery step hands over — which means one edit
here improves the protocol for the entire corpus at once, and no author copies forty lines
of boilerplate into a document that then drifts from every other copy of it.

They are also the only sections addressed to the READER rather than written by the author:
everything above them says how the capability works; these say how to get it into a
repository the author has never seen, and how to send back what only the reader can know.

Neutrality applies here twice over — this text ships to strangers by definition.
-->

## Adopt it here

*This section is addressed to the agent that is reading this recipe inside its own
repository. The recipe above knows the capability; only you know the ground. Do this before
the first build step.*

**Probe, do not assume.** For each thing the ground section says must be standing, go and
look for it here. Do not reason about what a project like this "usually has" — read the
repository. Then say, in one line each, what you found and what you did not.

**Agree the outcome before you build.** The recipe says what the capability *can* produce;
only your operator knows what it *must* produce here. Ask them, before the first build step:
which of the outputs are needed in this project, in what form, and which outside consumers
will read them. Do not assume every output transfers — a capability that emits three
artifacts at home may owe this project only one. And treat the bindings as law: a shape
marked `**Binding: adapt**` is yours to rename into this project's language; a shape marked
`**Binding: exact**` is not yours at all — something on the other side parses it, so confirm
with your operator that the consumer applies here, then copy it byte-for-byte — and check
your output against it mechanically, in this project's own CI, from this document alone (the
recipe tool's conform check does exactly that; if you cannot run it, validate against the
shape by hand and say so in your report).

**Map the roles onto this repository's homes.** The recipe names roles — *the facts file*,
*the collector*, *the optional target the project implements* — because a path is the one
thing that cannot travel. You own the other half: decide where each role lives here, in
this project's own conventions, and **write the mapping down where the next agent will find
it** (the project's configuration, not your context window). A mapping that exists only in
the session that made it is a mapping that will be re-invented, differently, next week.

**What is missing, propose — do not silently substitute.** When a precondition is genuinely
absent, you have exactly three honest moves, in this order:

1. **Create it**, at the smallest size that works. The recipe's ground section says what
   shape it needs and why it is load-bearing; carry that reason across in the language of
   *this* project, not as an appeal to someone else's conventions. A precondition worth a
   ladder in the recipe is usually worth twenty lines here.
2. **Degrade, and label it in the output.** If the recipe states a fallback, take it — and
   carry its label all the way into what the capability produces. A degraded source that is
   not labelled where the reader can see it is indistinguishable from the real thing, which
   is the failure the recipe was written to prevent.
3. **Stop and ask.** If neither is available, say so before you build. A capability standing
   on a precondition that is not there produces confident output over rot.

There is no fourth move, and in particular: **do not quietly reach for a weaker source that
happens to be lying around.** That substitution is how nearly every scar in every recipe
began.

**Do not import the author's structure.** What transfers is the principles, the contracts,
and the scars. The directory layout, the file names, the command names and the tool are the
author's house, and copying them buys you nothing and costs you a shape that does not fit.

**State the plan before you build.** The mapping, what you will create, what you will
degrade and how it will be labelled. That statement is the reader's only chance to correct
you while it is still cheap.

## Report back

*A recipe improves in exactly one way: someone builds it somewhere else and says what
happened. Every scar in the document above is a debt paid by an earlier reader. When you
finish, send this back to whoever handed you the recipe. Open with the recipe's name and the
`version:` from its frontmatter — a report that cannot say which text it built floats free
of every fix it triggers.*

- **Stack** — what you actually built it on, and which of the author's essential choices you
  kept, replaced, or could not honour.
- **Degraded** — every place you took a weaker path than the recipe asked for, what forced
  it, and the label it now carries in the output.
- **New scars** — the payload, and the reason this section exists. Anything that broke, that
  the recipe did not warn you about, as **symptom → root cause → fix**. A scar you hit is a
  scar the next reader does not have to.
- **Wrong** — anything the recipe asserts that turned out not to be true on your ground.
  Being contradicted by a build is the only way a recipe finds out it is stale.
- **Arithmetic** — the recipe names one live check that no gate can perform. Say whether you
  ran it, and what it said.
- **Missing** — the thing you had to invent because the recipe described it in prose and did
  not hand you the shape. If you had to design a contract, send it: it belongs above.

Keep it short. Five honest lines beat a report nobody writes.
