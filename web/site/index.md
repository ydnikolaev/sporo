# sporo — a recipe rebuilds the capability. A skill just runs it.

sporo turns a build you already did into a **recipe**: one self-contained file that teaches
an AI agent in a repository that has never seen yours how to build the same capability — on
its own stack, in its own harness, without repeating your scars.

> A skill runs in your harness. A recipe rebuilds the capability in any harness.

Install: `curl -fsSL sporo.dev/install.sh | sh` — one static binary, macOS / Linux / Windows.

## What is a recipe

Transferable, verifiable build intent. The name comes from *spore* — the minimal portable
unit that grows the organism in a foreign environment. A recipe carries that minimum —
principles, contracts, scars — and leaves behind everything that only makes sense in your repo.

| | a skill | a recipe |
|---|---|---|
| runs where | inside the harness that defines it | anywhere an agent can read and act |
| assumes | your tools, your paths, your conventions | nothing — it probes and maps |
| transfers | by copying the package | by being rebuilt, once, from scratch |
| on failure | breaks silently when the harness drifts | loud by design — every step carries its own acceptance |

## The genre

Eleven gated sections — the problem (acceptance first), why the obvious approach fails, the
principles, the ground it needs (every precondition a ladder: probe → build the smallest →
degrade with a label), the contracts (shapes shown, each with a binding: `adapt` or `exact`),
the build sequence (a `Done when:` per step), the seams, the scars (symptom → root cause →
fix, earned), verification, the trade-offs, and one closing section for the human.
Neutrality is mechanical: the body names roles, never paths, filenames or product names.

## How it works

**Author:** `sporo harvest` (mine the repo's record) → `sporo new` (coached draft, scars
pre-seeded) → `sporo lint` (the genre gate) → `sporo seal` (version + content hash — sealed
text never silently mutates) → `sporo export` (the one file you hand over).

**Reader:** probe the repository, map roles onto local conventions, agree the outcome with a
human, never silently substitute a weaker source — the protocol is appended to every export.

**The loop:** report-backs are new scars; new scars raise the version. The moat isn't the
file format — it's this loop.

## For teams

A shape one shared consumer parses is marked `Binding: exact`: the seal refuses changes to
it under anything less than a major version, fixtures travel with the shape, and every
adopter checks their own output in CI with `sporo conform` against the handed-over file
alone.

## When not to use it

Same stack, same harness, same company — copy the code. Determinism over adaptability —
ship the binary. One library call — that's the install command. No scars — no recipe needed.
Recipes earn their cost where the destination is heterogeneous and code can't cross.

## More

- [What is a recipe](https://sporo.dev/what-is-a-recipe.md)
- [Manifesto](https://sporo.dev/manifesto.md)
- [Source on GitHub](https://github.com/ydnikolaev/sporo)
