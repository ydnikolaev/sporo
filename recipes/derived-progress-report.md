<!-- SSOT SOURCE (sporo repo). Consumers receive a provenance-stamped copy via `sporo update` — edit HERE, never the synced copies. -->

---
id: 01KXS4J44AWVNJMYZSWKYD5CC9
name: derived-progress-report
version: 1.0.1
title: The daily progress report that checks itself — derived numbers, a page for people, a record the knowledge base ingests
problem: A recurring report is read by people who cannot verify a single figure in it, and later by machines that cannot parse the prose it is written in.
prerequisites: [read-files, edit-files, run-shell, read-version-control-history, http-client, scheduled-jobs]
derived_from: [the build record of one live implementation, its release history, its fix commits, and one downstream consumer's misdiagnosis]
stack: { language: go, runtime: "a single statically-linked binary + a coding agent", why: "the numbers must be derived by something with no imagination, and the reader of the recipe must be able to run the collector with no checkout of its source" }
verified: { project: internal-harness, release: v0.88.1, date: 2026-07-16 }
effort: "large — the collector and the wire client are a few days of ordinary code; the budget goes to the two halves that look free and are not: deciding what each number is allowed to be called, and the gates that keep the answer true after you leave"
---

# The daily progress report that checks itself

## The problem

You ship work every day and the people who fund it, depend on it, or plan around it cannot
see any of it. So somebody writes a summary. That summary is the one artifact your team
produces that **nobody checks**: every other output is consumed by a machine that rejects it
when it is wrong, and a report is consumed by a reader who has no repository to open. Now add
a second reader who arrives later — a shared knowledge base that must answer *"what shipped on
that project last week, and who worked on it"* — and prose alone stops being enough.

What you build: for each period, one **durable record of derived facts**, one **narrative
document** carrying a machine-derived header, one **self-contained page** for people, one
**regenerated aggregate** across every period ever reported, and one **idempotent transmission**
of the period into a knowledge base that lives outside your repository and parses what you send.

**The acceptance, in one sentence:** every quantity on the page can be traced to the facts
record a script derived (there is no third source), the period a reader saw yesterday is
byte-for-byte the same document today, and re-sending a period changes nothing on the other
side — while a period nobody could measure says so, on the page, instead of showing a zero.

## Why the obvious approach fails

The next agent will do the obvious thing, and it is worth naming precisely because it produces
something that **looks right**: ask a capable model to read the period's changes and write the
report — counting the work items by the ticked boxes in their documents, rendering it to a
page, and posting that page to the knowledge base.

Every part of that breaks, and none of the breaks is visible to the reader:

- **A recalled number is plausible, and plausible is indistinguishable from true.** Asked how
  many changes landed, a model produces a number of the right magnitude. It is wrong, the
  reader cannot check it, and nothing in the pipeline ever disagrees with it.
- **Counting ticked boxes reports the opposite of the truth, not an approximation of it.** In
  the verifying build, one live work item carried **0 of 41** ticked acceptance boxes in its
  documents while **7 of its 17** tracked phases were done. The box-counting report says
  **0%** — to a stakeholder, on a real project, about work that was more than a third
  finished. Not imprecise: inverted, in the direction that destroys trust fastest.
- **Rendering first makes the page the source of truth, and the page is prose a model wrote.**
  Everything downstream — the cumulative view, the header the knowledge base reads — then has
  to parse generated prose to recover numbers that existed, structured, thirty seconds
  earlier. A generated artifact that parses generated prose is a house of cards.
- **The consumer on the other side does not want your page.** It wants a shape it can key,
  dedup, and index. Posting a document and hoping means every project on the fleet invents its
  own dialect, and the divergence is found by the consumer, live.

The fork: **derive the facts into a durable record first, write the judgment against that
record, and render both the page and the wire shape from it.** Everything below is that
inversion.

## The principles

1. **The numbers are derived; the judgment is written; there is no third source.** A script
   computes every quantity from the repository into a persisted record, and the prose cites
   that record. A number not in the record does not go on the page. (The genre's own division
   of labour: the machine gathers, the author abstracts.)
2. **Completion comes from the work items' own machine-readable state — never from a mark
   inside a human-written document.** Boxes are not maintained by anyone; the state a check or
   a board already reads is.
3. **Every proxy carries its method into the document.** A quantity that is not the thing it
   stands for ships with the sentence that says so, and the sentence lives **inside the facts
   record** — so the page cannot be rendered without the disclosure being available to state.
4. **Absence is a state, never a zero.** A source that is missing, unreadable, or not kept on
   this machine is reported as an absence with its reason. A silent zero is a lie told with a
   number, and it is worse than the proxy it replaced.
5. **A total is not a measurement, and two questions never share one name.** Concurrent effort
   summed is not elapsed time; volume handled is not volume produced. Where both are worth
   having, both are printed, each under its own name.
6. **A published period is immutable; the aggregate is regenerated.** The period contains
   prose (irreproducible → frozen); the aggregate contains only what the facts say
   (reproducible → rebuilt, and gated like any other generated artifact). The aggregate is
   built from the facts and never by parsing the prose.
7. **A derived number is safe to embed; a derived string is not.** Change subjects, titles and
   model-written headlines are untrusted the moment they reach markup, a structured header, or
   a wire body — regardless of how trustworthy their source is. Escaping is achieved
   **by construction** (a typed value handed to a serializer), never by remembering to escape.
8. **The wire shape belongs to its consumer, not to you.** Outside the repository, field names
   are somebody else's parser. Copy the shape exactly, key it so that re-sending replaces
   rather than duplicates, and treat a change to it as a breaking change.
9. **A period is named explicitly, never inferred from a window's edge.** A window and the day
   it describes are two different facts; deriving one from the other is how a record lands
   under the wrong key.
10. **A credential is read at use time and rides in the transport header only** — never in the
    body, never in a log, never on disk. This is what lets the sender print its entire payload
    for inspection without leaking anything.
11. **Every artifact names which half of it was machine-derived, and which agent produced
    which half.** Who *sent* a record and who *built* it are different facts; a reader who
    cannot tell them apart will diagnose the wrong component (see the scars).
12. **A gate that cries wolf is worse than no gate.** A generated artifact carries no clock,
    and a drift check compares the substance the project can re-derive, never the shared skin
    it merely inlines.

## The ground it needs

Five things must be standing before step one. Each is a **ladder**: probe for it, build the
smallest one if it is missing, or degrade — and label the degradation where the report's own
reader can see it. There is no fourth rung.

### 1. A machine-readable history of changes, with timestamps and authors

**Why it is load-bearing:** it is the only source in the building that cannot be flattered.
Every delivery number and the effort proxy come from here.

- **Probe:** anything that can answer "which changes landed between these two instants, by
  whom, touching what" without a human transcribing it.
- **Build the smallest:** if changes are not recorded with timestamps, stop — this capability
  has no floor beneath it, and you are building a different thing first.
- **Degrade:** there is no honest degradation. A report whose delivery numbers are typed is
  the failure this recipe exists to prevent.

### 2. A machine-readable state for the work items

**Why it is load-bearing:** progress is the number a stakeholder reads first, and principle 2
says it may not come from prose. Without this, the report has no progress at all — which is a
supportable outcome, and inventing one is not.

- **Probe:** look for state that something already *acts on* — a check that fails when it is
  wrong, a board that renders from it. Not "a file that looks like a tracker".
- **Build the smallest:** one optional target the project implements, which prints the
  normalized rows in the contracts section below. It is twenty lines. **The capability must
  never parse the project's own schema** — that would carry one project's tracker into every
  future project forever. The shared side declares the contract; the project prints it.
- **Degrade:** no target → the report shows **no progress bars and states why**, carrying the
  reason as a first-class value in the facts record (a stated absence, not an empty list). The
  reason reaches the page.

### 3. A durable per-period home, and a check the project already runs

**Why it is load-bearing:** the record must survive as the unit of history (principle 6), and
the gates that keep the report honest must ride inside a command the project **already** runs.
A new standalone gate needs every project to wire it up, and the project that forgets is
exactly the one whose published history quietly stops matching its facts.

- **Probe:** one command that the team believes means "green". A per-period directory
  convention anywhere in the docs tree.
- **Build the smallest:** a directory per period holding the facts record and the documents;
  the existing check gains two lines.
- **Degrade:** no check runner → the gates become a scheduled job that reports loudly. If they
  are not run at all, say so in the report's own honesty note: this report is ungated.

### 4. External access: a place to send the record, a credential, and a clock

**Why it is load-bearing:** the second reader is a machine outside your repository. Without a
sink this capability is still worth building — but its ingestion half does not exist, and the
recipe's exact contract is the part you skip.

- **Probe:** does the fleet have a knowledge base that ingests documents, and does it publish
  a wire shape? If it does, that shape wins over the one below.
- **Build the smallest:** an endpoint that accepts the envelope in the contracts section,
  upserts on (project, period), and requires a scoped write credential. The unattended run
  needs a scheduler on a machine that has the credential.
- **Degrade:** no sink → the reports stay local and the aggregate is the whole deliverable;
  say in the human note that this project's history is not queryable by the fleet, so nobody
  assumes it is.

### 5. A presentation layer for the page

**Why it is load-bearing:** the page is forwarded — attached, mailed, opened on a phone with
no access to your network. It must be **one self-contained file with no external fetches and
no scripting**, or it will render as a broken skeleton in exactly the situation it was made
for.

- **Probe:** the project's own design tokens or shared stylesheet.
- **Build the smallest:** one stylesheet, inlined at build time.
- **Degrade:** no styling layer → an unstyled but self-contained page. Never a page that
  fetches its style at open time.

## The contracts

### The wire envelope

The one shape a machine **outside** the emitting repository parses: the fleet's shared
knowledge base keys, dedups and indexes it. Copy it byte-for-byte.

**Binding: exact**

```json
{
  "kind": "progress-report",
  "schema": 1,
  "project": "<project-slug>",
  "date": "<YYYY-MM-DD>",
  "title": "<human title of the period's document>",
  "source": "<sender-name>/<sender-version>",
  "reviewed": false,
  "content_format": "markdown",
  "content": "<the entire narrative document, its header included>",
  "metadata": null,
  "facts": null
}
```

Field notes — three of these carry traps:

| field | meaning, and the trap |
|---|---|
| `kind` | routes future record kinds; the consumer branches on it rather than sniffing the body |
| `schema` | the **wire** version. Version it on its own axis, never coupled to your local record's version — otherwise a local shape change silently misreports the wire contract |
| `project` + `date` | **the dedup key.** Re-sending the same pair must replace, never duplicate. `date` is the period the report is *about*, not when it was sent (principle 9) |
| `title` | untrusted free text (it is derived from the document) — it is serialized, never concatenated |
| `source` | **names the SENDER, not the generator.** Both facts matter and the field carries one; carry the generating agent's version separately or a downstream reader will diagnose the wrong component (see the scars) |
| `reviewed` | `false` = an unattended run produced it, nobody confirmed the scope; `true` = a human ran it and read the output. The consumer should surface this on any answer it derives — an unreviewed record is a weaker source |
| `content` | the **entire** narrative, header included. This is the primary ingestible document |
| `metadata` | declared **nullable on purpose**: the consumer requires the key and treats its interior as a loose index, because projects legitimately carry different keys. The shape it *should* hold is the adapt-bound block below |
| `facts` | the structured record, verbatim, when the sink opts in. Also nullable: a consumer that reads a missing key as starvation and an explicit `null` as "not sent" gives the reader one unambiguous answer — **emit the key always**, `null` when facts are withheld or dropped to fit the size cap |

**Fixture: valid**

```json
{
  "kind": "progress-report",
  "schema": 1,
  "project": "acme-web",
  "date": "2026-07-16",
  "title": "acme-web — progress report, 2026-07-16",
  "source": "acme-reporter/1.4.2",
  "reviewed": false,
  "content_format": "markdown",
  "content": "---\nkind: progress-report\nproject: acme-web\n---\n\n# acme-web — progress report\n\nCheckout redesign reached its third phase.\n",
  "metadata": { "kind": "progress-report", "project": "acme-web", "date": "2026-07-16" },
  "facts": { "schema": 1, "project": "acme-web" }
}
```

**Fixture: invalid** — the sender renamed a key into its own language; the consumer's dedup starves

```json
{
  "kind": "progress-report",
  "schema": 1,
  "slug": "acme-web",
  "date": "2026-07-16",
  "title": "acme-web — progress report, 2026-07-16",
  "source": "acme-reporter/1.4.2",
  "reviewed": false,
  "content_format": "markdown",
  "content": "# acme-web\n",
  "metadata": null,
  "facts": null
}
```

**Fixture: invalid** — the confirmation flag arrived as a string, so every record reads as truthy and "unreviewed" stops being distinguishable

```json
{
  "kind": "progress-report",
  "schema": 1,
  "project": "acme-web",
  "date": "2026-07-16",
  "title": "acme-web — progress report, 2026-07-16",
  "source": "acme-reporter/1.4.2",
  "reviewed": "false",
  "content_format": "markdown",
  "content": "# acme-web\n",
  "metadata": null,
  "facts": null
}
```

**Fixture: invalid** — an extra top-level field is a private dialect the consumer never agreed to; it belongs in the loose index block

```json
{
  "kind": "progress-report",
  "schema": 1,
  "project": "acme-web",
  "date": "2026-07-16",
  "title": "acme-web — progress report, 2026-07-16",
  "source": "acme-reporter/1.4.2",
  "reviewed": false,
  "content_format": "markdown",
  "content": "# acme-web\n",
  "metadata": null,
  "facts": null,
  "internal_run_id": "8f21"
}
```

### The index block — the header the narrative carries and the envelope repeats

Derived wholly from the facts record and stamped atop the narrative, so a machine can classify
the period without parsing the prose beneath it. It is the **same block** the envelope's
`metadata` carries, pre-parsed, so the consumer need not read the document's header format.

**Binding: adapt** — the consumer requires the key, not the interior; different projects
legitimately carry different measurements, and a record rejected over a schema quibble is a
record that silently stops arriving.

```yaml
kind: progress-report
schema: 1
project: <project-slug>
date: <YYYY-MM-DD>
window: { since: <RFC3339>, until: <RFC3339>, label: <human window name> }
author: <the distinct authors of the period's changes, derived>
reviewed: false
lang: <the language of the PROSE; the numbers have none>
repos: [<each repository this period covers>]
metrics:
  commits: 0
  code_commits: 0
  insertions: 0
  deletions: 0
  files: 0
  hours_derived: 0.0        # the PROXY (principle 3) — always present
  # the four below are emitted ONLY when a runtime was actually read.
  # absent key = not measured; a zero here means "measured, and it was zero".
  agent_hours: 0.0          # effort: concurrent spans SUMMED
  wall_clock_hours: 0.0     # elapsed: the same spans, overlaps MERGED
  tokens_input: 0           # fresh input only
  tokens_output: 0
epics: [{ slug: <work-item-slug>, percent: 0, phases: <done>/<total>, status: <lifecycle state> }]
tags: [<project>, progress-report, <work-item-slug>]
headline: <the frozen one-line judgment>
```

The trap is the four optional metrics: make them **nullable/absent-capable in your types**, or
"not measured" and "measured as zero" collapse into the same output — which is principle 4
broken in the one half where nobody can see it.

### The facts record — the durable local unit

**Binding: adapt** — nothing outside the repository parses it; the aggregate and the header
are both derived from it, so its shape is yours to name as long as it carries these roles.

```json
{
  "schema": 1,
  "project": "<project-slug>",
  "lang": "<language of the prose>",
  "date": "<the canonical period label — passed IN, never inferred from the window>",
  "window": { "since": "<RFC3339>", "until": "<RFC3339>", "label": "<human window name>" },
  "git": { "commits": 0, "code_commits": 0, "insertions": 0, "deletions": 0, "files": 0, "hours": 0.0, "authors": ["<name <address>>"] },
  "repos": [{ "name": "<repository role name>", "commits": 0, "insertions": 0, "deletions": 0 }],
  "epics": [{ "slug": "<work-item-slug>", "title": "<title>", "percent": 0, "phases_done": 0, "phases_total": 0, "status": "<lifecycle state>" }],
  "epics_source": "<how the rows were obtained, or 'none: <why>' — never silently empty>",
  "epics_warnings": ["<what the seam could not read; this reaches the page>"],
  "runtime": {
    "agent_hours": 0.0,
    "wall_clock_hours": 0.0,
    "sessions": 0,
    "tokens": { "input": 0, "output": 0, "cache_read": 0, "cache_write": 0, "processed": 0 },
    "providers": [{ "provider": "<runtime role>", "verified_build": "<the build its log shape was read against>", "sessions": 0, "agent_hours": 0.0 }],
    "absences": ["<a runtime with no readable log here, and why>"],
    "note": "<what these numbers do and do not measure>"
  },
  "method": {
    "session_gap_minutes": 0,
    "excluded_paths": ["<generated and vendored trees>"],
    "code_types": ["<which change types count as product work>"],
    "hours_note": "<the sentence the page must print about the proxy>",
    "runtime_note": "<the sentence the page must print about the measurement>"
  },
  "summary": {
    "headline": "<the frozen judgment>",
    "body": "<the frozen narrative summary>",
    "epics": [{ "slug": "<work-item-slug>", "line": "<what this item means, in the reader's language>", "estimate": "<a RANGE with its basis>" }]
  }
}
```

Two roles here are load-bearing and easy to mistake for bookkeeping. **`method`** is what makes
principle 3 mechanical: the disclosure sentences travel *with* the numbers, so the page cannot
be produced without them. **`summary`** is the frozen judgment written *back* into the record —
because the aggregate is built from this file and never from the rendered page, a judgment that
does not land here does not exist in the history.

### The normalized work-item rows the project prints

**Binding: adapt** — this is the seam between the shared capability and the project's own
tracker schema. The shared side declares it; the project's optional target emits it.

```json
[
  { "slug": "<work-item-slug>", "title": "<title>", "phases_done": 0, "phases_total": 0, "status": "<lifecycle state>", "blocked_on": "<what it waits on, in plain language, or null>" }
]
```

`blocked_on` is not decoration: an unstarted item stays on the page **with its reason**. A row
omitted because 0% "looks bad" is the one omission a reader cannot detect.

### The reply the sender must classify

**Binding: adapt** — the sender logs the identifier and depends on nothing in it.

```json
{ "id": "<stored document id>", "status": "stored" }
```

Classify by status class, not by parsing prose: **2xx** = stored (both "created" and "replaced"
are success — the sender must not care which); **credential/malformed/too-large** = fail
loudly, no retry, because retrying a misconfiguration just repeats it; **rate-limited or
server-side** = retry with backoff, then leave the period on disk and exit non-zero. A failed
send never mutates local state — the period is on disk either way, and can be re-sent.

## The build sequence

### 1. Build the collector — every number, derived, into the record

Take a window and the **period's canonical label as separate inputs** (principle 9), walk the
version-control history, and write the facts record from the contracts section. Exclude
generated and vendored trees from the line counts; derive effort by clustering change
timestamps into sessions with a configurable gap; write the `method` sentences alongside. The
collector is **pure with respect to the repository**: same window, same history, same bytes
out.

**Done when:** running it twice over a closed window produces byte-identical records, and the
record contains no field that a human typed.

### 2. Declare the work-item seam — and let it be absent

Declare the normalized rows as a contract and call an **optional** target the project
implements. Never parse the project's own tracker. When the target is missing, unreadable, or
returns nothing, record the reason as a value (`none: <why>`) and carry anything the seam could
not read into a warnings list that reaches the page.

**Done when:** a project with no such target produces a report with no progress bars and a
printed reason; a project with one produces rows; and a tracker that fails to parse produces a
warning on the page rather than a silently shorter list.

### 3. Add the measured half — beside the proxy, never over it

If the agent runtimes keep session logs, read them through a **declared surface** — does this
runtime persist a log, where, in what shape, which fields carry the working directory, the
timestamp, the model, the token counts — stamped with the build you verified the shape
against. Ask the surface, never the runtime's name. Normalize what the runtimes disagree
about. Emit effort (summed) and elapsed time (merged) under **separate names**, and never
overwrite the timestamp-derived proxy: the two sit side by side.

**Done when:** effort ≥ elapsed on a day with parallel sessions, elapsed can never exceed the
window's length, and a machine with no logs produces a stated absence with its reason instead
of a zero.

### 4. Write the judgment against the record, and freeze it back

The agent reads the facts and writes what a machine cannot: what the work **means** in the
reader's language, what is left, and an estimate **as a range with its basis** ("two to three
days: two dependency waves remain, the first of which is four parallel phases"). It types no
quantity. Then the judgment is written back into the facts record.

**Done when:** every number in the prose appears in the record; the aggregate (step 7) can show
the period's headline without ever opening the rendered page.

### 5. Render the page for people — self-contained, no scripting, escaped by construction

Inline the styling; no external fetches; no scripting. Order it so it reads in thirty seconds:
a plain-words overview → the headline numbers, each under its own name → the work-item bars,
**including the ones at zero, each with its reason** → what shipped, each item with why it
matters → the rest, in small print → the tooling's own work, last (hiding it is dishonest;
leading with it is self-absorbed). Print the `method` sentences as they are. Every untrusted
string reaches the markup through a serializer or an escaper, never through concatenation.

**Done when:** the page opens correctly from a mail attachment on a machine with no network,
and a change subject containing markup renders as visible text rather than executing or
truncating the page.

### 6. Stamp the machine-readable header, and gate it

Emit the index block from the facts record — **never** from the history directly, so it can
only say what the facts say — by marshalling a **typed value**, so an untrusted string
round-trips as a quoted scalar instead of splitting the document. Add a check mode that
re-derives the block and fails on drift. Tolerate a period with no block at all (a record
written before the header existed is not drift), so the gate lands inert on history.

**Done when:** hand-editing one number in a stamped header makes the check fail, re-deriving
makes it pass, and a period predating the header stays green.

### 7. Generate the aggregate — from the facts, never from the prose

One self-contained page across every period ever reported, built **only** from the facts
records. It carries **no clock** — a "generated on" line makes its own drift gate fail on every
run. Gate it for drift, comparing **what the facts say**, not the shared skin it inlines.

**Done when:** two consecutive builds with no new period produce identical substance; a
centrally released style change does not turn any project's check red; and the aggregate's
content demonstrably comes from the records (delete the rendered pages, rebuild, nothing is
lost).

### 8. Build the sender — pure body, impure shell

Split it: a **pure** function that turns a persisted period into the exact envelope (facts +
narrative in, bytes out — no clock, no network, so the shape can be held to a fixture), and a
thin shell that moves the bytes. Assemble from a typed value and serialize. Read the credential
at send time and put it in the transport header only. Send the structured record verbatim when
the sink opts in — the same bytes on both sides. Classify the reply by status class per the
contract. Keep a machine-local ledger of what was already sent so a re-run of an unchanged
period is a no-op — as an **optimisation on top of** the consumer's own idempotency: a lost or
corrupt ledger degrades to "send anyway", never to an error.

**Done when:** a dry run prints the entire payload with no credential anywhere in it; the
payload passes a conform check against the exact contract; sending the same unchanged period
twice produces one stored record and the second call is a no-op; and a forced re-send of a
changed period replaces rather than duplicates.

### 9. Add the unattended run — one atomic verb for the scheduler

Collect → refuse to overwrite an immutable published period → **hold back a period a human
already confirmed** (an unattended draft must never overwrite it, because latest-wins on the
consumer's side would silently downgrade a confirmed record) → stamp it unconfirmed → rebuild
the aggregate → send. It ships a draft; a human-run confirmation later supersedes it under the
same key.

**Done when:** the scheduled run stamps the **period it was asked for** (not the window's
edge), a confirmed period survives an unattended run untouched, and running it twice in one
night stores one record.

### 10. Wire both gates into the command the project already runs

The drift gates ride inside the existing check, not in a gate of their own.

**Done when:** the project's own green command turns red on a seeded drift in either a stamped
header or the aggregate — with no extra wiring in any project that already runs the check.

## The seams

What must stay configurable, or the next project inherits values it never chose:

- **Where periods live** — the reports home, through the project's structure seam.
- **The work-item schema** — the project's alone. The shared side declares the normalized rows
  and consumes only those; a shared component that parses one project's tracker carries that
  project's schema to every project forever.
- **The method values** — the session gap, the excluded paths, which change types count as
  product work. These are the levers people argue about; a number someone disagrees with is
  fixed here, never in the prose.
- **The audience and the language of the human document** — while the machine-readable record
  stays in one fixed language, because its reader is a system reading the project's history
  later.
- **The sink** — the endpoint's location, the **name** of the environment variable holding the
  credential (never the credential), whether the structured record rides along, and whether
  this project sends at all. Per project: each project is its own document stream on the
  consumer's side.
- **The period's coverage** — the repositories a period folds in, standing plus a per-period
  opt-in.
- **What "reviewed" means operationally** — who is allowed to stamp confirmation.

## The scars

### Ticked boxes reported the exact opposite of the truth

**Symptom:** a report generated from ticked acceptance boxes showed **0%** on a work item that
was more than a third complete: **0 of 41** boxes ticked across its documents, while **7 of its
17** tracked phases were `done`.
**Root cause:** the boxes had never been the artifact anyone maintained. Nothing consumed them,
so nothing kept them true — and a report is the first consumer prose has ever had.
**Fix:** progress is read from the work items' machine-readable state or it is not read at all;
where no such state exists, the report shows no progress and prints why (principle 2, ground 2).

### Summing parallel sessions produced 25.3 hours inside a 24-hour day

**Symptom:** the effort figure exceeded the length of the day it described.
**Root cause:** agent sessions run concurrently — subagents, worktrees, two terminals — so the
sum of their active spans is machine **effort**, which is true, and impossible as **elapsed
time**, which is what a reader assumes a figure in hours means.
**Fix:** two numbers, each under its own name: effort is summed, elapsed time is the union of
the intervals and can never exceed the window. The gap between them *is* the parallelism, which
is information, not an error to smooth away.

### The headline number was 1.85 billion, and 98% of it was the same context re-read

**Symptom:** a period "processed" 1.85 billion tokens — a spectacular figure, and meaningless.
**Root cause:** cache re-reads dominate the total: the same context re-served turn after turn.
A total handled is not a measurement of work done. Worse, two runtimes counted their cached
tokens on **opposite sides** of the input line (one inside it, one beside it), so a single
naive input column across both would have been a number nobody could defend.
**Fix:** print fresh input and output only; name everything else separately; normalize the
providers' conventions on a declared flag and record which convention each source used.

### The period was stamped a day late, under the wrong dedup key

**Symptom:** the unattended run asked for a given day and produced a record labelled the day
after; asked for yesterday, it stamped today. Found on the first end-to-end send against the
live endpoint — every gate was green.
**Root cause:** the run's windows are half-open — `[day 00:00, next-day 00:00)` — so the upper
bound is the **next** day's midnight, and the label was derived from it. A window alone cannot
name its period when its upper bound is exclusive. Because the consumer dedups on
(project, period), the record would have landed under the wrong key — the one failure that
cannot be corrected by re-sending.
**Fix:** the canonical period label is an **explicit input** to the collector (the caller
already computes it for the directory), falling back to the window's edge only in the
convention where that edge sits inside the target period. Teeth: an end-to-end test asserts the
stamped label equals the day requested.

### A change subject with a colon split the machine-readable header

**Symptom:** a header block, hand-assembled from derived strings, broke its own parser; the
same class of string reaching the page could inject markup.
**Root cause:** a derived string is untrusted regardless of how trustworthy its source is — a
subject line carrying a colon, a leading dash, or markup is free text the moment it enters a
structured document. Escaping-by-remembering fails at exactly one call site, eventually.
**Fix:** the block, the page and the wire body are all built from **typed values handed to a
serializer** — never concatenated. The escape then cannot be forgotten, because nobody performs
it. This one hazard belongs to the genre: anything that renders text it did not write, into an
artifact that gets forwarded, inherits it.

### The collector blamed the data for a broken toolchain

**Symptom:** on the first live run against a real project, **every** work item came back
"unreadable" — in a report whose entire purpose is telling the truth about the data.
**Root cause:** the task runner and the operator's shell resolved **different interpreters**,
only one of which had the parsing library installed. The script worked by hand and died under
the runner; the caller read the failure as *"this tracker is invalid"*. A broken toolchain wore
a data failure's clothes.
**Fix:** the toolchain a target runs on is **declared, never ambient** — resolved in a stated
order, and when it cannot be, the failure is loud, carries the fix, and is distinguishable from
a data failure. A collector that reports "no data" when it means "I am broken" produces a report
that merely looks quiet.

### A downstream team diagnosed a regression that did not exist

**Symptom:** the knowledge-base team read a record stamped with a recent sender version, saw a
near-empty index block, and correctly reasoned their way to *"the index serialization regressed
in that version"*. It had not. The record had simply been **built** by an older agent, before
the header existed, and **sent** by a new one.
**Root cause:** the field named the *sender* while the reader took it for the *generator*, and
nothing anywhere recorded the generator at all. A real consumer misdiagnosed a real defect out
of one field's ambiguity — which is the strongest possible evidence that a name is wrong.
**Fix:** carry both facts, each named for what it is. Any field that could mean "who made this"
or "who moved this" must say which, and the other must exist somewhere.

### Attribution by directory name silently dropped whole sessions

**Symptom:** work done in a subdirectory of the project vanished from the measured half.
**Root cause:** the runtime files a session under a slug of the directory it started in — a
session started one level down gets its **own** slug. Matching the project by that slug is
matching on a lossy key.
**Fix:** both runtimes record the working directory **inside** the log; attribution reads that
and asks whether the path is inside the project. Where a period spans repositories, the
sessions are deduplicated: a session must count once even when it sits under two roots.

### A byte-for-byte drift gate reddened projects that had changed nothing

**Symptom:** the aggregate's drift check would fail in every project the day a **central design
release** touched a token — for a difference the project neither made nor could re-derive.
**Root cause:** the aggregate inlines the shared skin (that is what keeps it one shareable
file), so a byte comparison compares two things at once: the substance the project owns and the
presentation somebody else releases.
**Fix:** gate the substance, not the skin — the comparison covers what the facts say, and the
presentation refreshes on the next build. The cost is **stated, not discovered**: an aggregate
wears the previous release's skin until the next period is generated.

## Verification

The gates that must ship **with** the capability — an unguarded invariant rots back to nothing,
and every one of these guards a hand-check somebody would otherwise do once:

- **Header drift** — re-derive the index block from the facts and fail on any difference.
  Teeth: hand-edit one number, the check must fail. It must also stay green on a period that
  predates the block.
- **Aggregate drift** — rebuild and compare the substance. Teeth: hand-edit the aggregate's
  content, the check must fail; release a style change centrally, the check must stay green.
- **Immutability** — the collector refuses to overwrite a published period. Teeth: re-run over
  a published period and observe the refusal, not a rewrite.
- **The wire body against the exact contract** — the pure payload builder is held to a fixture,
  and the project's own check conforms a real payload against the shape shipped in this
  document. Teeth: rename one field, the check must fail. This is what turns "everyone sends
  the same shape" from an agreement into a gate.
- **Absence is not zero** — a fixture with no readable runtime log must emit a stated absence.
  Teeth: a test asserting that "measured, and it was zero" and "not measured" produce different
  output.
- **Surface freshness** — every declared runtime-log surface carries the build it was verified
  against, and a check fails when the installed build has moved past it. A log format rots
  exactly as silently as a hook does: a provider renames one field, the report keeps printing a
  number, and nothing says a word. (In the verifying build this gate caught a stale stamp within
  an hour of being widened to cover every stamped surface.)
- **Both drift gates ride inside the command the project already runs.** A gate registered where
  the project does not look is a gate that passes forever.

**The one live check no gate performs:** generate a real period, open the page from a mail
attachment on a machine with no network, and send it — twice. Then read it back from the
consumer. The day-late stamp, the empty index block, and the "unreadable" trackers were all
found this way, by looking at the first real output, after everything was green.

## The trade-offs

- **You now own a schema and a collector.** The facts record is a durable shape with a version
  and history behind it; every change to it is a migration question. This is the price of
  principle 1, and it is not small.
- **Freezing a period makes mistakes permanent.** A published record someone read cannot be
  quietly improved — that is the point, and it will hurt the first time a typo ships.
- **The judgment is not free.** A machine derives the numbers; a person or an agent still has
  to say what they mean, in the reader's language, every period. If nobody is willing to spend
  that, you will produce a page of numbers nobody reads.
- **The seam costs the project something.** Progress requires the project to publish its work
  items' state. Projects that will not do it get an honest report with no progress in it — a
  supportable outcome, and a disappointing one.
- **An exact wire contract binds the fleet.** Changing a field is a breaking change for
  everything that parses it; the flexibility lives in the loose index block, on purpose.
- **The aggregate wears last release's skin** until the next period is generated — the stated
  cost of gating substance over bytes.

**When not to build this at all:** when nothing recurs (a one-off answer to "how's it going?"
is a sentence, not a pipeline); when the reader **can** check the claim themselves (a changelog
is the record, and the changes are readable); when the artifact is an internal orientation note
for people who can open the repository; and when nobody outside the work is asking — the
machinery earns its place only when a report recurs, because recurring is what makes someone
start trusting it.

## For the human

You get a pipeline that makes a daily report **impossible to fake by accident**. A collector
reads your history and writes every number into one durable record per day, with the method
behind each number sitting beside it. An agent then writes the only part a machine cannot — what
the work means and what is honestly left — against that record, and cannot type a number without
it being visibly absent from the facts. Out of the same record come three things: a page a
partner can open on a phone with no network, a Markdown document a machine can read, and one
idempotent transmission into the knowledge base your fleet queries later. Re-sending a day
changes nothing; a day someone read on Monday is the same document on Friday; and a day nothing
could measure says so, out loud, instead of showing a confident zero.

**The stack it was built on:** a single statically-linked binary (Go) doing every derivation,
driven by a coding agent that writes only the judgment; the version-control log as the source of
truth; JSON for the durable record and the wire body, YAML for the header a machine reads off
the top of the document; the page as one self-contained HTML file with zero scripting.

**Essential** — replace these and the capability degrades into the thing it was built to
replace:

- **A compiled, dependency-free collector.** The numbers must come from something with no
  imagination, and it must run identically on a laptop, in a scheduled job, and inside a check.
  The moment deriving a number requires a service to be up, somebody will type the number.
- **A durable machine-readable record as the single axis.** The page, the header, the aggregate
  and the wire body are all *projections* of it. This is the choice everything else rests on.
- **Serialization from typed values.** Escaping is achieved by construction, never remembered.
- **A self-contained, zero-scripting page.** The artifact gets forwarded; anything it must fetch
  at open time will be missing exactly when it matters.
- **Checks that run on one command with no services up** — otherwise the gates get skipped, and
  ungated invariants rot.

**Incidental** — the author's house, swap without a second thought: Go itself (any language that
compiles to one runnable artifact does this); YAML for the header (JSON reads fine, it is just
denser on top of a document people also read); the particular directory convention per period;
the house stylesheet; the specific verb names.

**What that stack bought, and what it cost.** It bought distribution: one artifact, no runtime to
install, so the same collector runs in a scheduled job and in a check, and the "it works on my
machine" class of report failure never appeared. It bought a compiler that makes the "measured
zero vs not measured" distinction a type, not a convention. It cost **ceremony**: a shape change
is a code change, a release, and a version bump — no quick reshaping of the record over lunch,
which is a real tax when the shape is still moving. It cost a **second language in the loop**:
the agent that writes the judgment does not write the collector, so the split has to be
maintained deliberately rather than emerging. And it cost the **rebuild-before-you-benefit** step
for every consumer of the binary: an improvement to a number's method reaches a project only when
that project reinstalls, which is a real coordination cost the day the fix is urgent.
