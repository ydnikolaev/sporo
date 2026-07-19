# Manifesto

Capability doesn't move the way code does.

Code moves by copying: clone the repo, run the binary, install the package. That works
when the destination looks like the source — same stack, same harness, same company.
It stops working the moment it doesn't. A skill built for one agent's harness doesn't
run in another's. A prompt tuned for one team's conventions produces something else
entirely on a different team's codebase. And some of the time, the code simply cannot
cross the boundary at all — a different company, a different client, a wall neither
side can remove.

What actually survives that crossing is intent: the principles a build rests on, the
shape of what it produces, and — the part everyone skips — the specific ways it went
wrong the first time. A recipe is that intent, written down in a fixed genre, addressed
to an agent that has never seen your repository and has to arrive at your capability
anyway.

This is not a prompt. A prompt is a request; it carries no memory of what breaks. It is
not a skill package; a skill is a promise about *your* harness, silent about everyone
else's. Marketplaces for either optimize for the wrong thing: volume of artifacts, not
the compounding value of a build history. A prompt store gets you more prompts. It
never gets you fewer failures.

A recipe gets you fewer failures, because it has a memory built into its shape. Every
scar section is a debt someone already paid so the next reader doesn't have to. And the
loop doesn't end at export: when a reader builds the recipe somewhere else, they report
back — what they degraded, what broke that the document didn't warn about, what turned
out to be wrong on their ground. That report becomes a new scar. The new scar raises the
version. The recipe you read next month has already survived a failure you would
otherwise hit this month.

That loop is the moat, not the file format. Anyone can write a markdown template with
twelve headings. What compounds is a genre disciplined enough — gated on neutrality,
gated on shown contracts, gated on earned scars — that reports keep landing in a shape
worth reading, version after version, without one author's private conventions leaking
into everyone else's build.

Transferable intent is not a better prompt and not a smaller skill. It's a third thing:
a document that survives being handed to a stranger, checks its own outcome when they
follow it, and gets strictly better every time someone does. Code moves by copying.
Capability moves by teaching. sporo is built for the second kind.
