<!-- SSOT SOURCE (sporo repo). Consumers receive a provenance-stamped copy via `sporo update` — edit HERE, never the synced copies. -->

---
name: daily-progress-report
version: 1.2.0
title: The daily progress report — a document nobody can check, that checks itself
problem: The people funding the work cannot see it happening, and the only thing you can show them is a repository they cannot read.
prerequisites: [read-files, edit-files, run-shell, version-control]
derived_from: [the reporting doctrine, the build record of one live implementation, one independent rebuild on another stack]
stack: { language: go, runtime: "a single static CLI binary + the project's own build target", record: git, why: "the collector must run anywhere the repo is checked out, with no services up" }
verified: { project: mate, release: v0.84.0, date: 2026-07-15 }
effort: "the facts from the version-control record are an evening; the runtime telemetry is the rest of the budget and every scar below"
---

# The daily progress report

> **Thesis.** A progress report is the one artifact in the entire pipeline that **nobody
> checks**. Every other thing a team produces is verified by the machine that consumes it —
> the compiler rejects the code, the test rejects the regression, the deploy rejects the
> bad config. A report is read by a partner, an investor, a client who will never open the
> repository, and it is verified by *nobody*. That asymmetry is the whole design problem.
> The report must therefore **check itself**: every number derived by a script from the
> record, every judgment written and labelled as judgment, and every derivation that is a
> proxy saying so, in the document, where the reader can see it.

## The problem

The work is real and it is invisible. The people who need to see it — partners, investors,
a client — cannot read a repository, will not read a commit log, and are being asked to
keep believing on the strength of a verbal update. Meanwhile the person doing the work is
the least reliable narrator available: not dishonest, just *inside* it, and inside is where
"we made great progress today" comes from.

**You have it when:** a document exists, every period, that a non-engineer can read in
thirty seconds; every quantity in it was computed by a machine from the record and can be
traced back to a persisted file; every estimate is a range with a stated basis; and the
document says, in plain words, what it does **not** know. And when the numbers are
unflattering, the document still ships — that is the day it earns the trust the rest of the
days spend.

## Why the obvious approach fails

The obvious approach is to ask the agent, at the end of the day, to write the report. It
reads the diff, it summarizes, it produces something plausible: *"roughly 30 commits, about
5,000 lines, we're around 60% done."* Every one of those numbers is a fabrication with a
decimal point, and the reader — the one person structurally incapable of catching it — is
the one it is pointed at.

The second-most-obvious approach is worse, because it *looks* rigorous: count the ticked
checkboxes in the plans. In the live build this recipe comes from, one epic had **zero of
forty-one** acceptance boxes ticked while **seven of its seventeen tracked phases were
done**. Nobody had ever maintained the boxes; the phases were the thing the team actually
kept true. A checkbox-counting report would have told the reader **0%** — not imprecise, but
the exact opposite of true, in the one direction that destroys trust permanently.

Both failures share a root: **the report was authored where it should have been derived.**

## The principles

- **The numbers are derived, never authored.** A script computes every quantity from the
  record; the prose cites the persisted facts and never restates them from memory. There is
  no third source.
- **Progress comes from the work-item state, never from prose.** Read completion from the
  machine-readable state the team actually maintains. Where none exists, the report shows no
  progress number **and says why** — a stated absence is information; an invented number is
  a lie with a progress bar around it.
- **Every proxy names itself.** A quantity that is not the thing it stands for carries its
  method into the document, in plain words, not in a comment nobody reads.
- **A measurement retires a proxy only when it says what it measured.** A better source is
  where the most dangerous numbers come from: precise, sourced, and answering a different
  question than the one on the page.
- **Estimates are ranges with a basis.** "Two to three days: two dependency waves remain, of
  which the first is four parallel phases" is an estimate. "2.5 days" is a guess wearing a
  suit.
- **A published period is immutable; the cumulative view is generated.** A reader who saw
  Monday's report must find the same report on Friday.
- **Absence is a state, not a zero.** Every collector must be able to say "not measured, and
  here is why". A silent zero is the one failure mode that looks exactly like success.
- **The report renders strings it did not write.** Commit subjects, work-item titles, branch
  names: all of it comes from the record, and the record is not trusted input. Escape every
  one of them on the way into a shareable artifact. A document whose entire purpose is trust
  cannot be one that executes text somebody else typed.
- **Two collectors that answer different questions must not be added together.** The version
  control system attributes work to the repository the change landed in; a runtime's session
  log attributes it to the directory the session was working in. They are both true, they
  disagree constantly, and the sum of them is a number that answers nothing.

## The ground it needs

Three things must be standing **before** the sequence starts. None of them is this
capability, and all three of them are why it works — a report pipeline built on top of a
project that has none of them derives confident numbers from rot.

Most readers will not have all three. That is the normal case, not a disqualification, and
each one below therefore comes as a **ladder**: probe for it, build the smallest possible
version if it is absent, and — only if the team will not keep even that — degrade to a stated
proxy that is **labelled as one in the output**. What is not on the ladder is the fourth rung
everybody actually takes: quietly pointing the collector at a weaker source that happens to be
lying around. Every scar in this document started there.

**A single source of truth for the work, machine-readable.** Not a plan document, not a board
someone updates when they remember: one artifact, in the repository, that the team *actually
maintains* because something else already breaks when it is wrong. The reason is mechanical —
a report derives from a source, and if that source is prose, the derivation is a parse of a
paragraph and the number it produces is a guess with a decimal point. The obvious-failure
section above is what happens when a project has two candidate sources and the report picks
the one nobody maintains.

- **Probe:** is there any machine-readable state of the work here that something else already
  depends on? Look for what a check or a build reads — not for what looks tidy.
- **If not, build the smallest one that works.** One file, one entry per work item, the shape
  in the contracts below. Twenty lines of it beats a board nobody opens, and the report is not
  the only thing that will end up reading it.
- **If the team will not keep one, degrade — and say so.** Derive work *streams* from the
  scopes the commits already carry, show movement per stream, and label them in the document
  as **activity, not completion**. Never a percentage: a percentage implies a denominator, and
  a proxy has none. A stream that is honest about being a proxy is worth having; a percentage
  invented from commit counts is the lie this recipe exists to prevent.

**A structure the machinery can write into without asking.** A declared home for the
generated artifacts, one directory per period, with a stated rule that a published period is
immutable. This sounds like a filing convention and is actually a correctness constraint: the
cumulative view is a pure function of the persisted periods, so the moment periods are
scattered or rewritten, the aggregate is no longer derivable and someone starts hand-editing
it. A generated artifact with no fixed home becomes a hand-maintained one within a month.

- **Probe:** does this project declare where generated artifacts live? If it has any such
  convention, use it — do not import the author's.
- **If not, declare one now**, in whatever the project uses to hold its own values, before the
  first period is written. A home chosen by the first run is a home nothing can move later.

**Always-on rules and gates that keep the derivation honest after you leave.** The pipeline's
invariants — a proxy names itself, absence is not a zero, the published period is not
rewritten — are not properties of the code, they are properties of *behaviour*, and behaviour
regresses. Put them where the machine enforces them (a gate that reds) and where the agent
reads them before it writes (a rule that is always loaded). An invariant that lives only in a
reviewer's head is a promise, and this document exists because promises are exactly what the
reader cannot check.

- **Probe:** is there a single command that runs this project's checks, and does anything
  refuse a change when it fails?
- **If not, the gate is still worth writing** — as a check the build runs and a human can run.
  A gate nothing invokes is documentation; but the ladder's bottom rung here is not "skip the
  gate", it is "run it by hand every period and write down that you did".

## The contracts

Three shapes. They are the interfaces this capability is built against, and they are the part
that transfers: an implementation is a stack's business, a contract is not. Copy them; where a
shape is **Binding: adapt**, rename the fields into your own language and keep what each one
*means* — where it is **Binding: exact**, the consumer on the other side owns the field names,
and changing them is a new MAJOR version of this recipe, not a preference.

**The facts record — the period's whole durable output.** **Binding: adapt** — unless your
fleet shares one aggregator that ingests these records across projects; then agree the shape
once and treat it as exact from that day. Everything the report says is in
here: the numbers, the method that produced them, and — after the author writes it — the
judgment itself. The cumulative view then reads this file and nothing else. Note three things
the shape is doing on purpose: **the method is disclosed inside the facts**, so a document
cannot be rendered without the disclosure being available to state; **fresh and re-served
tokens are separate fields**, never a total; and **a missing source is an object that says it
is missing**, never a zero.

```json
{
  "date": "2026-07-14",
  "schema": 1,
  "vcs": {
    "reachable": true,
    "totals": { "commits_total": 16, "commits_product": 12, "commits_tooling": 4,
                "lines_added": 4019, "lines_removed": 100, "files_touched": 45 },
    "repositories": [
      { "name": "<a repository this window touched>", "reachable": true,
        "commits_total": 12, "lines_net": 2141,
        "subjects": [ { "kind": "product", "subject": "<the commit subject, escaped on render>" } ] }
    ],
    "method": {
      "sessions": "a gap over 45 minutes between commits opens a new one — a proxy from timestamps, not a measurement",
      "lines": "generated and vendored trees excluded; the excluded set is a project value"
    }
  },
  "telemetry": {
    "reachable": false,
    "absent_reason": "no runtime session logs on this machine — the hours below are the commit-timestamp proxy only",
    "sessions": null,
    "effort_min": null,
    "elapsed_min": null,
    "tokens": { "fresh_input": null, "fresh_output": null, "cache_read": null, "cache_write": null },
    "method": {
      "effort": "summed across concurrent sessions — machine busy-time, NOT the length of the day",
      "elapsed": "the union of the session intervals — wall-clock, and the number a reader means by 'hours'",
      "tokens": "fresh is new work; cache_read is context re-served and is not work"
    }
  },
  "work_items": { "reachable": false, "absent_reason": "no machine-readable work-item state in this project", "items": [] },
  "judgment": {
    "headline": "<one or two sentences: what actually went into the world today>",
    "shipped": [ { "title": "<what shipped>", "why": "<why the reader should care>" } ],
    "streams": [ { "name": "<a work stream>", "status": "<moving | blocked | not started>", "summary": "<one line>" } ],
    "next": "<a range with a basis: 'two to three days — two dependency waves remain, the first is four parallel phases'>",
    "tooling_note": "<the infrastructure's own work, stated and put last>"
  }
}
```

**The work-item contract — the one thing the project implements, and the shared side never
learns.** **Binding: exact** — one shared collector parses what every project emits, and two
dialects of this shape make the collector lie about one of them. Every team's tracker has its own schema and no shared tool should ever parse it.
Declare the normalized shape you consume; let the project print it from one fixed-name target
in whatever it already uses as a build interface. A project with no such target gets a report
with no completion numbers and one honest line saying why — never an invented percentage.

```json
[
  { "id": "<a stable slug>",
    "title": "<a human name a non-engineer can read>",
    "phases_done": 7,
    "phases_total": 17,
    "blocked_by": ["<what it is waiting on>"],
    "actionable_now": ["<the phases that could start today>"] }
]
```

**The runtime surface declaration — a fact you verify, write down, and branch on.**
**Binding: adapt** — it is read only by the collector you build beside it. If your
agent runtimes keep session logs, they hold the truth the commit timestamps only approximate.
Do not encode a runtime's log format in code and do not branch on the runtime's *name*: declare
the surface in data, stamp it with the build you verified it against, and let the code ask the
declaration. A runtime that later writes an already-declared shape then costs nothing, and a
format that rots does so *loudly*, against the stamp.

```yaml
runtimes:
  <the runtime, as you name it>:
    verified_build: "2026-07-14"   # the freshness gate reds when the installed build moves past this
    log_glob: "<where its logs live, as a pattern>"
    project_slug_rule: "<how it derives the log directory's name from a working directory>"
    timestamp_field: timestamp     # one log file spans MANY calendar days — filter events by date, never files
    cwd_field: cwd                 # the working directory recorded per event: the only honest attribution key
    usage_path: <where in the event the token counts live>
    tokens:
      fresh_input: <field>         # verify the convention: does it already exclude cache reads, or not?
      fresh_output: <field>
      cache_read: <field>          # context re-served — NOT new work
      cache_write: <field>
    cache_convention: "<state it explicitly — the two runtimes we checked sit on opposite sides of this line>"
```

## The build sequence

### 1. Name the reader and the acceptance

Write down who reads this and what they must be able to do after reading it. Everything
below is a consequence: the language, the length, which numbers lead, what "done" looks like
in a sentence a non-engineer can repeat to someone else.

**Done when:** you can state, in one line, the decision the reader makes with this document.
If you cannot, you are building a dashboard for yourself.

### 2. Derive the facts into a persisted record

One command, run over a window of the record, that emits **the facts record shown in the
contracts** — the commits split into product work and tooling upkeep, lines added and removed
with generated and vendored trees excluded, distinct files, and work sessions inferred from
commit timestamps. Persist it beside the report. The prose is then written *against this file*,
and a number that is not in it does not go in the report.

Two properties of that shape are load-bearing and easy to drop while implementing it. The
**method lives inside the record**, so the rendered document cannot be produced without the
disclosure being available to state. And **every source is an object that can say it is
missing** — never a bare number that a zero can impersonate.

**Done when:** running it twice over the same window produces the same numbers; every figure
you want in the report is in the file; and a source you deliberately make unreachable comes
back as a stated absence rather than a zero.

### 3. Read progress through one optional contract, not one shared schema

Every team's work-item tracker has its own schema, and no shared tool should ever learn it.
Consume **the work-item contract shown in the contracts section**, and let each project
implement one fixed-name target in whatever it already uses as a build interface, which prints
that shape. The project keeps its schema; the shared tool keeps its neutrality, and never grows
a branch per project.

A project that does not implement the target gets a report with **no completion numbers and one
honest line saying so** — or, if you took the ground section's bottom rung, work *streams*
derived from commit scopes and labelled as activity rather than completion. Never an invented
percentage.

**Done when:** a project with no tracker at all still produces a valid report, and the
report says what it cannot see.

### 4. Measure what the machine actually did — and only claim that

**This step is the whole budget.** The version-control facts are an evening's work; the
session telemetry is where the rest of the time goes and where every scar below was earned.
Plan accordingly, and if you have to cut something, cut this and ship the proxy honestly
labelled — that is a strictly better report than a measured number you did not verify.

Fill in the runtime surface declaration from the contracts section, **against the runtime's
own current documentation and its actual log output — not from memory, and not from this
document**: these formats change, and this one is a snapshot. Then read the events, and hold
to three things the shape is trying to tell you.

**Attribute by the working directory recorded in the event, not by anything else.** And
understand what that means: it is a *different question* from the one the commits answer. The
version-control record says which repository the change landed in; the session log says where
the agent was standing when it worked. An agent working on one repository from a session
opened in another — routine, and the way most agents are actually driven — makes them
disagree. Report them as two answers, never as one.

**Fix the day boundary once, in one place.** Commit timestamps come back in local time; log
events are typically stamped in UTC. Choose one — the reader's local day is the honest choice,
since the report claims to describe *a day* — convert everything to it at the point of
reading, and say which one the document means.

**Split the two numbers that are not the same question.** Effort is *summed* across sessions;
elapsed time is the *union* of the intervals. And keep the old proxy beside the measurement
rather than deleting it — two independent derivations landing in the same range is the closest
thing to corroboration you will get.

**Done when:** the measured figure and the proxy are both in the facts file, each with its own
label; neither is called "hours" without a qualifier; and you have confirmed that events
outside the reported day are excluded by their **timestamp**, not by which file they sit in —
one log file spans many days.

### 5. Write the judgment against the facts, and only the judgment

The machine cannot say what the work *means*, why it matters to this reader, or what is
honestly left. That is the entire job of the author — human or agent — and it is written
strictly against the persisted facts. Ask the reader's owner what the machine cannot see: a
demo happened, a call unblocked something, a dependency arrived. None of that is in the
record and all of it belongs in the report.

**Done when:** every number in your prose can be pointed at, by hand, in the facts file.

### 6. Render twice: once for people, once for machines

Two documents, one truth. The human page is short, visual, in the reader's language, and
**self-contained** — one file that opens offline, with no external requests, because the
reader will forward it to someone who is not on your network. The machine record is the
canonical one, in one fixed language, written explicitly for a future agent reading this
project's history.

Order the human page so it is readable in thirty seconds: what happened, in plain words →
the headline numbers → every work stream with its state, *including the ones that have not
started* and why → what shipped and why it matters → and the tooling's own work **last**
(hiding it would be dishonest; leading with it would be self-absorbed).

**Lead the headline numbers with elapsed time, not effort.** A reader who sees "hours" reads
wall-clock — that is what the word means to everyone outside this pipeline — and effort is the
larger, more flattering number, which is exactly why leading with it is a lie the reader
cannot catch. Elapsed first, effort beside it with its label.

**Escape everything that came from the record.** Commit subjects and work-item titles land in
this page verbatim, and they are text somebody else typed: a subject containing a closing
script tag breaks the page, and one containing an image tag with an error handler does worse
than break it. This is not a hypothetical for a document whose whole point is that it gets
**forwarded**.

**Done when:** the page renders offline with zero external requests; someone who does not work
with you can tell you what happened; and a commit whose subject is a raw markup fragment
renders as *text* on the page — try it, do not reason about it.

### 7. Generate the cumulative view from the facts — never by parsing the prose

The period reports are frozen; the cumulative view is regenerated on every build, from the
facts files alone. A generated artifact that parses generated prose is a house of cards. To
make this possible, the author's judgment must be **written back into the facts** — the facts
file is then the period's whole durable record, numbers and frozen prose together, and the
generator reads nothing else.

Here you will hit a contradiction, and it is better to resolve it now than in a red gate:
**the cumulative view must be regenerable byte-for-byte, and it also carries presentation you
do not control.** The page is one self-contained file (that is what makes it forwardable), so
its styling is *inside* it — and the day the shared styling changes, a gate comparing rendered
bytes reds in every project, for a difference nobody made. The resolution is to split them:
persist the aggregate's **data** — the thing derived from the facts — and gate *that*; let the
rendering refresh freely on the next build. Gate the substance, never the skin.

The generated view carries **no clock**: a "generated on this date" line makes its own drift
gate red on every run, and a gate that cries wolf is worse than no gate.

**Done when:** deleting and regenerating the cumulative view reproduces its persisted data
exactly, and changing only the presentation does **not** turn the gate red.

### 8. Gate what you just verified by hand

Everything you checked once by eye is an invariant with no guard. Ship the gates *with* the
capability: the generated view matches the facts; a published period is not silently
rewritten; a period directory is complete. Each gate gets a fixture that makes it red — a
gate nobody has seen fail is a gate nobody has tested.

**Done when:** you have deliberately broken each invariant and watched the gate catch it.

## The seams

What every project swaps, and what must therefore never be baked into a shared body:

- **Where the reports live** — a declared path, not a convention.
- **The work-item schema** — the project's own; the shared side declares only the normalized
  contract it consumes.
- **The method values** — the session threshold, the excluded paths, which commit types count
  as product work, the idle cutoff for a session log.
- **The audience and the language** of the human page — and, separately, the language of the
  machine record, which stays fixed.
- **The runtime log surfaces** — declared per runtime, in data, stamped with the build they
  were verified against.
- **The number formatting**, and it is less trivial than it looks: a locale's thousands
  separator may be a non-breaking space, and code that splits or parses formatted numbers on
  the assumption of a comma will produce something wrong and plausible. Format at the edge,
  compute on the raw values, and never round-trip a number through its presentation.
- **The day boundary** — which timezone the document means by "a day". A value, stated in the
  document, and applied by every collector.

## The scars

### The checkboxes that said zero

**Symptom:** the progress number would have read 0% on an epic that was 41% done.

**Root cause:** completion was being counted from ticked boxes inside human-written
documents. Nobody maintains those boxes; the team maintained the tracked phases instead.

**Fix:** read completion only from the work-item state the team actually keeps true — and if
you cannot identify such a state, report no percentage and say so.

### Twenty-five hours inside a twenty-four-hour day

**Symptom:** the first live run of the measured hours reported 25.3 hours of work in a single
calendar day. Arithmetically impossible, and it went straight into a draft aimed at partners.

**Root cause:** every agent session's active time was being **summed**. Sessions run
concurrently — subagents, two terminals, parallel worktrees — so the sum is machine effort,
not elapsed time. Both are real; they are different questions.

**Fix:** report effort as the sum and elapsed time as the **union** of the merged intervals,
each under its own label. Neither number may ever wear the other's name.

### A total is not a measurement

**Symptom:** the same run "processed" 1.85 billion tokens. As a headline it was spectacular
and it meant nothing.

**Root cause:** 98% of it was cache reads — the same context re-served turn after turn. The
total was a true number answering no question anybody had.

**Fix:** print what is *fresh* — new input, generated output. Name the rest separately as what
it is (a re-read, not new work), or leave it out. A number that impresses and cannot be acted
on is decoration, and decoration in a trust document is a liability.

### The two runtimes that counted cached tokens on opposite sides of the line

**Symptom:** a single "input tokens" column across two runtimes would have been a figure
nobody could defend.

**Root cause:** one runtime reports cached tokens *inside* its input count; the other reports
them *beside* it. Adding the two columns silently double-counts one of them.

**Fix:** declare each source's convention in data, normalize on read, and record which
convention each source used. The units of a measured quantity are themselves a fact to
verify, not to assume.

### The day's work that was attributed to the wrong repository entirely

*Contributed by the first independent rebuild of this recipe, on another stack. It is the
deepest of the attribution scars and the recipe did not warn him.*

**Symptom:** twelve commits landed in one repository over a day, and that repository showed
**zero** minutes of engagement. The hours were all attributed to a different repository — the
one the sessions had been *opened* in.

**Root cause:** the two collectors answer different questions and the report was treating them
as one. Version control attributes work to the repository the commit **landed in**. The
runtime's session log attributes it to the directory the session was **standing in**. An agent
driven with absolute paths — or told to work on a sibling checkout, which is how agents are
actually used — makes those diverge for an entire day, silently, with both collectors working
exactly as written.

**Fix:** stop trying to reconcile them. Present the session telemetry as **the day's
engagement** and the version-control facts as **what was delivered, per repository**, and say
in the document that they are two different measurements. Any attempt to divide one by the
other, or to attribute machine-hours to a repository through the commits, produces a number
with no referent. (Note the shape of this failure: it is not a bug. Every component was
correct; the *join* was invented.)

### The day that was two different days

*Also contributed by the rebuild — a latent one, caught before it shipped.*

**Symptom:** none yet, which is the point. The commit window was being selected in local time
and the session events were stamped in UTC.

**Root cause:** two record sources with two clocks, and a "day" that was never defined
anywhere.

**Fix:** define the day once, in the document and in the collectors, and convert at the point
of reading. Also: one session log file spans *many* calendar days — filter events by their
timestamp, never by which file they live in, or a day's report inherits a week.

### The commit subject that could have run in the reader's browser

*Also contributed by the rebuild — the recipe was silent on this, and it is the one gap with a
security bit.*

**Symptom:** none, because he escaped it and then went looking for the recipe's warning and
found none.

**Root cause:** the report renders strings from the record — commit subjects, work-item titles
— into a page that is **forwarded to people outside the team**. A commit subject containing a
closing script tag breaks that page; one containing an image tag with an error handler makes
it hostile. The record is not trusted input, and nothing about "it is only our own commit log"
survives contact with a repository that takes contributions.

**Fix:** escape everything that flows from the record into the rendered artifact, at the
boundary, without exception — and test it with a commit whose subject is a raw markup
fragment rather than reasoning about whether it could happen.

### The session that worked in a subdirectory

**Symptom:** a session's work vanished from the report entirely.

**Root cause:** sessions were being attributed to a project by the directory slug the runtime
names its log directory after — and a session started in a *subdirectory* of the repository
gets its own slug. The work was there; the attribution dropped it, silently.

**Fix:** read the working directory **recorded inside the log**, not the one implied by where
the log sits. And state the boundary in the document: a session that moved between two
repositories counts toward both.

### The broken toolchain that reported success

**Symptom:** with the parser's interpreter missing, the collector reported *every* work item
as unreadable — and exited **0**. The report would have shown a project with no epics, and
nothing anywhere would have said the tool was broken rather than the project empty.

**Root cause:** two of them, and the second is the one worth carrying: the failing command's
exit code was captured after a bare conditional, which in the shell reports **success** when
its condition is false. So the loud failure that had been carefully written could not fire.

**Fix:** distinguish the failure classes by exit code — one code for "the data is bad",
another for "the toolchain is broken" — so a caller can never blame the data for a broken
tool. Then **prove it by sabotage**: break the toolchain on purpose and watch the loud path
run. It was only under deliberate sabotage that the swallowed exit code appeared at all.

### The gate that reddened every project for a change none of them made

**Symptom:** the drift gate on the generated cumulative view went red in every consumer the
day the shared design changed — for a difference no project made and none could re-derive.

**Root cause:** the generated page carries its presentation inline (that is what keeps it one
shareable file), and the gate compared bytes. Presentation is released centrally; substance
is not.

**Fix:** gate the **substance**, not the skin: compare what the facts say, and let the
presentation refresh on the next build. A gate that cries wolf is worse than no gate, because
it trains the team to ignore the ones that matter.

### The author who broke his own immutability rule within the hour

**Symptom:** a published period was regenerated "to improve it", and its commit count changed
— because the collection window had slid.

**Root cause:** me. The rule was written, the gate was written, and the override flag was one
keystroke away in a moment of tidying.

**Fix:** the machinery already refused; what failed was the human. Make the period's frozen
state visible at the point of the override, and treat "improve a report someone has already
read" as the thing it is: a rewrite of history that the reader cannot see and would not
forgive.

## Verification

**The gates that must ship with the capability** — each with a fixture that makes it red,
because a gate nobody has watched fail is a gate nobody has tested:

- the generated cumulative view matches the persisted facts (substance, not presentation);
- a published period is complete (its facts, its human page, its machine record);
- a published period is not silently rewritten;
- the collectors report **absence as absence** — a source that is missing produces a stated
  reason, never a zero;
- a string that came from the record and contains raw markup **renders as text**. This one is
  a fixture, not a review note: the failure only appears on the day someone writes an unusual
  commit subject, which is the day the page is already in a partner's inbox.

**The one live check no gate can do for you.** Read the first real output with your own eyes
and *check its arithmetic against the calendar*. That is how the twenty-five-hour day was
caught: no type system, no test, and no gate would have flagged it, because every component
was working exactly as written. The number was impossible, and only a human looking at it
knew that. Do the same for the attribution: if a repository shows delivery and no engagement,
or engagement and no delivery, the join is wrong — not the day.

## The trade-offs

**What this costs.** A machinery build whose budget is wildly uneven, and it is worth knowing
which half you are buying: the version-control facts are an evening, and the runtime telemetry
is everything else — every scar above lives in it. And a real per-period cost that never goes
to zero: the judgment is not automatable, and the moment it is fully automated the document
stops being worth reading. Expect a build step, a gate suite, and a human (or an agent under
supervision) in the loop for the prose.

**What it deliberately refuses.** It will not show anything the record does not carry — no
"time spent thinking", no offline conversations, no work done in another tool. It refuses to
present a point estimate. It refuses to hide the tooling's own commits. And it refuses to
print a zero where it means "not measured", which will occasionally make the report look
*less* impressive than a dishonest one would.

**When not to build it.** If nobody outside the work reads it, this is a dashboard for one,
and a dashboard for one is a hobby. If the report is a one-off, write it by hand. And if the
team's work-item state is not actually maintained, **fix that first** — every number this
machinery derives from it will inherit its rot, and a rigorous pipeline over rotten inputs
produces confident garbage, which is strictly worse than an honest guess.

**The residual risk, stated.** The measured hours are the *machine's* engagement, not a
person's working day, and no amount of precision changes that. The document says so. If your
reader wants a timesheet, this is not one, and pretending otherwise would break the exact
trust the whole apparatus exists to build.

## For the human

*This is the one section written for a person rather than for the agent that will build it.*

**What this is.** A report that goes out every day, to people who are paying for the work and
cannot see it: partners, investors, a client. Its numbers are not written by anybody — they
are computed from the record the work already leaves behind (the commit history, the tracked
work items, the runtime's own session logs), and they are written into a durable file that
the prose then has to cite. A person still writes the *judgment*: what shipped, why it
matters, what is left, how long it is likely to take. But nobody gets to write a number.

**What you get.** Every period: one page a non-engineer reads in thirty seconds — what
shipped, how much work went in, where each stream stands, what is next — plus a machine copy
for the agent, plus a persisted facts file that both were derived from and that can be
re-checked forever. And a cumulative view that is regenerated from those files, so the
history cannot quietly drift from what was published on the day.

**On what principles.** Three, and they are the entire product: *derived, never authored* —
if a number can be computed it is never typed; *every proxy names itself* — an hour figure
inferred from commit timestamps says so, in the document, in words the reader understands;
*absence is a state, not a zero* — when a source is missing the report says it is missing,
because a silent zero looks exactly like a real measurement and is the one failure that
destroys trust permanently.

**The stack it was originally built on, and why that is the recommendation.** A compiled CLI
shipped as one self-contained binary, with the project's own build tool as the only interface
between it and the repository, and version control as the record it reads.

*Essential* — the parts the design actually depends on, and which we recommend on any stack:

- **The record is the version-control history and the work-item state, not a separate log.**
  A pipeline that needs its own timesheet has already lost: people stop filling it in, and the
  numbers rot silently. Deriving from what the work leaves behind anyway is what makes the
  report free at the margin.
- **The collector runs offline, from a checkout, with nothing else up.** No database, no
  service, no network. This is what lets the report be regenerated a year later and still come
  out the same — and a report you cannot re-derive is a report nobody can check.
- **The facts are persisted before the prose is written.** The document is a *citation* of a
  file, not a memory of a conversation. Everything else in the design follows from this.
- **A fixed, narrow contract for anything project-specific.** One optional target the project
  implements; the pipeline declares the shape and never learns the project's schema. Without
  it the tooling grows one branch per project and dies.

*Incidental* — the parts that came from the author's own house, and which an agent on another
stack should swap without a second thought: the language and the fact that it compiles; the
choice of the build tool as the entry point; the templating of the human page. A scripting
runtime, a task runner, a different renderer — all fine. None of them touches a principle.

**What that stack bought, and what it cost.** It bought portability of the worst kind of
dependency: the collector is one file, it runs in any checkout of any project, and a consumer
project needs nothing installed to use it. It bought a real type boundary around the facts
schema, which is the artifact everything else derives from. It cost a build step before every
check, which makes the fast edit-run loop slower than a script would be; it cost a release
cycle (a fix to the collector is not live until the binary ships); and it means the one place
project-specific logic *is* allowed — the optional target — sits in a different language than
the pipeline, which is a seam the author must document or the next reader will not find it.

*(What it costs to build, and when not to build it at all, is above, in the trade-offs — that
is the same question asked of the design rather than of the stack, and it is not repeated
here.)*

## Appendix — how one harness did it

Concrete instances, for illustration only — everything above stands without this section.

In the mate harness the machine half is two verbs: `mate report facts` derives the window
into `docs/reports/<date>/facts.json`, and `mate report build` regenerates the aggregate
`docs/reports/index.html` from every day's facts. The epic seam is an optional `make
report-epics` target — mate declares the normalized JSON contract, and the consumer project
(axon) implements it in `.agents/scripts/report-epics.sh` by wrapping its own tracker parser.
The runtime surfaces are declared per provider in `adapters/<provider>/telemetry.yaml` and
stamped with the build they were verified against; `scripts/surface-freshness.sh` reds when
the installed build moves past the stamp. The judgment half is the `create-report` skill. The
report gate rides inside `mate status`, which every consumer's `make mate-check` already runs
— a gate registered where the surface is not is a gate that passes forever.
