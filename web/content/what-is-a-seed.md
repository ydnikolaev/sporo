# What is a seed

You've met the recipe — the genre that teaches an agent to *build* a capability
from principles, anywhere, naming nothing concrete. A seed is its sibling, born from
it and inverted: it teaches an agent in a repository that has never had a tool how to
*bring that one tool in* and stand it up. If you're the engineer deciding whether to
let an agent run an install on your tree, this is the page that earns that.

## The one-sentence definition

A seed teaches an agent in a repository it has never seen how to acquire one named
tool and make it real: detect whether the tool is already present, install it from an
origin the seed vouches for, prove the install actually took, put it to use, wire it
into the repository's own agent harness, and then account for every one of those moves
to the human who has to live with them. Its reader is a machine that will act on a
foreign tree — so the whole document is a chain of moves, and exactly one section, at
the end, is written for you.

## Accountability, not neutrality — the constraint inverted

A recipe's hardest constraint is neutrality: name roles, never instances, because its
subject is a transferable pattern. A seed inverts that without abandoning it. Its
subject is *one concrete tool*, so it names that tool freely — the target, its version,
its origin, its own install command all belong in the document, because they are the
document. What a seed must still never name is the reader's tree: their paths, their
filenames, their config locations, the directory where their harness lives.

So the discipline shifts from transferable design to trustworthy action. A seed that
installs without detecting clobbers a working tree. One that installs without verifying
reports a success it never had. One that pipes a stranger's script into a shell with no
cited origin is an unaudited privilege it handed away on the reader's behalf. Every rule
in the genre exists to keep those from happening — the test is the recipe's test pointed
at a different object: could an agent follow this in a repository whose layout you have
never seen?

## Seven sections, in this order

A seed has frontmatter, then seven required body sections. The shape is gated, because a
genre defined only by taste drifts, and a seed that skips a rung leaves a gap an agent
fills by improvising on the reader's machine.

1. **Summary** — a 2–4 sentence orientation before any move: what tool this brings in,
   what standing it up buys the reader, and the state they're in when it's done. It's
   body text, so neutrality applies in full.
2. **What it is** — the tool itself: what it does, the shape of the thing that lands, and
   the model the reader needs in their head before they let it onto their machine.
   Understanding before acquisition — an agent that installs a thing it can't describe
   can't judge whether the install went right.
3. **Install** — the acquisition sequence, one step per heading. This section carries the
   seed's two sharpest teeth: its first step opens with a **Detect** marker (is the tool
   already here, and at what version?), and every step — the detect step included — closes
   with a **Done when** line stating the observable condition that proves it took.
4. **Verify** — the proof the whole install actually works, and it must run at least one
   real command the agent observes, not prose that asserts success. Install can lie: a
   package half-lands, a lookup cache goes stale, a binary arrives without its execute bit.
5. **Use** — how the reader actually puts the now-installed tool to work: the first real
   thing they do with it, made concrete enough to act on but still neutral about the
   reader's own tree.
6. **Harness** — how the tool joins the repository's agent harness so future agents know
   it's here and how to reach for it. This section is advisory: recommend a project-local
   rule only when the tool ships none of its own, or you manufacture a second source of
   truth that drifts the first time the tool updates.
7. **Report** — the closing section, and the only one written for a person: a fixed
   five-row table that accounts for what the seed just did to this repository.

The order is the argument. You cannot use a tool you haven't verified; you cannot verify
one you haven't installed; you cannot install one you don't understand enough to describe.
Reorder the sections and you break the chain — a Verify before an Install verifies nothing.

## The trust contract — the seed's teeth

Four rules, each mechanical, each with a gate:

- **Detect before you install.** The first install step answers one question — is the
  target already present, and at what version? Detect-first is what makes a seed
  idempotent: safe to run twice, safe to run where the tool already exists.
- **Every install step states its acceptance.** A step with no observable **Done when**
  is one the agent can't tell succeeded until the end — which on a foreign tree is exactly
  when failure is most expensive.
- **No blind pipe into a shell.** The genre doesn't ban fetching a remote script and
  piping it into an interpreter — for many real tools it's the official path — but the
  step must cite the origin the frontmatter vouches for, so the human can audit the move
  as "this comes from the origin this seed stands behind," not "from wherever."
- **Verify holds a runnable proof.** A command whose output the agent reads is the
  difference between "I ran the installer" and "the tool actually runs here."

Together they turn an install from a hopeful sequence of commands into an accountable one:
it doesn't clobber, doesn't proceed on faith, doesn't execute the untraceable, and doesn't
claim a success it never observed.

## The Report is for the human — five fixed rows

The Summary orients the agent; the Report is written for the person who has to accept what
the seed just did. Where a recipe describes a design and can do so in prose, a seed took an
action on the reader's tree — so its closing section is a fixed-shape audit, exactly five
rows in order: **what it is**, **how it works**, **what was done** (the actual mutations
this run made — the row that makes the seed accountable), **how to use it**, and
**suggest next**. A human reviews many seed runs; a report whose shape changes every time
is one they must re-read from scratch. The fixed table is a contract with that eye.

## Derived from the tool's current truth, not memory

A seed's raw material is read and run, not recalled. A tool's install command, its verify
check, the origin it ships from — all of it lives in the tool's own current documentation
and in what it does when you actually run it, not in what you remember from some past
release. The **verified** stamp is the proof this happened: it says the seed was run, on a
named machine, at a named release, on a named date — a snapshot of one successful install,
not maintained doctrine. That's what lets a reader tell a fresh proof from a stale claim.
