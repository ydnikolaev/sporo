<!-- SSOT SOURCE (ag-web@axon). The export strips this banner; edit here, hand over ONLY what `sporo export` prints. -->

---
id: 01KXS83W67BK5QB23TXSMZ073K
name: russian-personal-data-legal-and-consent-banner
version: 0.1.1
title: A compliant personal-data legal set + cookie-consent banner for a Russian-jurisdiction site
problem: A public site that processes any personal data of people in Russia needs a legal-document set and a cookie banner that satisfy 152-ФЗ — not a translated GDPR template that blocks the wrong things and is silent on what the law actually demands.
prerequisites: [read-files, edit-files, publish-web-pages, run-a-browser]
derived_from: [the build record of one live Russian-jurisdiction publication]
stack: { language: "human-readable legal prose (documents) + a client-side consent gate over browser storage (banner)", runtime: "any renderer for the documents; a web client that can hold non-essential scripts back for the banner", why: "the legal half is jurisdiction prose portable to any publishing stack; the banner half needs only a client tier that can gate scripts before they run" }
verified: { project: ag-web@axon, release: "jurisdiction-resolved consent contour, 2026-07", date: 2026-07-18 }
effort: heavy — but the budget is lopsided: the legal half (document content, requisites, the RKN filing, lawyer review) is the work; the banner is an afternoon once the category model is decided.
---

# A compliant personal-data legal set + cookie-consent banner for a Russian-jurisdiction site

## The problem

You run a public site — a magazine, a shop, a service — and it processes personal
data of people located in Russia. It almost certainly does: an analytics tag reads an
IP and a device fingerprint; a contact form takes an email; comments and accounts take
names; a newsletter takes an address. The moment any of that is true, Russian law
(152-ФЗ «О персональных данных», 149-ФЗ, and for marketing 38-ФЗ «О рекламе») binds
you as an **operator of personal data**, with duties that a Western privacy template
does not cover and partly contradicts.

You do not have: the right **set** of legal documents (which documents, how many, why),
the **content** each one must carry to be valid, the **operator registration** the law
requires before you process anything, and a cookie banner that behaves the way the law —
not a marketing plugin — requires.

**The output** is: a published set of separate legal documents carrying mandatory
content and the operator's identifying details; an operator notification filed with the
supervisory authority (Roskomnadzor / РКН); and a cookie-consent banner that holds every
non-essential tracker back until the visitor affirmatively opts in, and can prove the
choice later.

**You have it when:**

- Every document the law obliges for *your* feature set exists, each is a **separate**
  document (consents are not folded into the terms of use), each carries the operator's
  full identifying details, and each cites the specific article it rests on.
- The operator notification is filed **before** processing begins, or you can name the
  narrow statutory exemption that releases you (there are only a few, and most sites meet
  none of them).
- Loading the site in a clean browser fires **zero** network requests to any
  non-essential third party and writes **no** persistent tracking identifier, until the
  visitor clicks accept. Rejecting leaves that same clean state. Strictly-necessary and
  functional storage works regardless.
- You can produce, for any consent, a record of *who* agreed to *what*, *when*, and *in
  which revision of the document*.

## Why the obvious approach fails

The next agent reaches for a **cookie-consent plugin and one "Privacy Policy" page** —
the GDPR shape everyone has seen — translates it to Russian, and ships. It fails on five
concrete points, each of which is a live legal exposure, not a stylistic quibble:

1. **One policy is not the document set.** 152-ФЗ art. 9 requires the subject's
   **consent** to be a specific, informed, freely-given act — and in practice a
   **separate document**, not a clause buried in a terms-of-use page. Distributing
   personal data to an open audience (public author profiles, comments) requires *its own*
   consent under art. 10.1. A newsletter requires *its own* consent (152-ФЗ + 38-ФЗ art.
   18). Fold these together and none of them is valid.

2. **"Accept cookies" is not consent to process personal data, and vice-versa.** They are
   two different legal acts under two different bases. A single banner that conflates them
   satisfies neither.

3. **The GDPR banner blocks the wrong categories.** GDPR-shaped tooling tends to gate
   *functional* preferences (theme, layout) behind consent and can be lax about
   *analytics*. Under the Russian model, strictly-necessary and functional storage needs
   **no** consent (blocking it is a self-inflicted UX wound), while analytics/marketing
   trackers must not fire **before** an opt-in.

4. **It ignores duties that have no GDPR analogue.** Operator **notification** to the
   supervisory authority *before* processing (152-ФЗ art. 22); **data localization** —
   the primary databases holding Russian citizens' data must be physically in Russia
   (art. 18); a **separate notification before any cross-border transfer** (art. 12). A
   plugin does none of this.

5. **It uses the wrong lawful bases and misses the requisites.** Every document must name
   the operator's full identifying details and cite the specific article each purpose
   rests on. A translated template names a foreign entity and the wrong statute.

The reader's instinct produces a site that *looks* compliant and is exposed on all five.
This recipe exists to replace the instinct.

## The principles

1. **Consent is a separate, specific, informed, and free act — never bundled, never a
   condition of access.** (152-ФЗ art. 9.) One consent document per purpose family. It may
   not be pre-ticked, and refusing it may not deny access to content the consent is not
   actually needed for. A consent that is the price of reading the site is not "free" and
   is void.

2. **The document *count* follows your features, not a fixed number.** A minimal read-only
   site needs fewer; each capability that touches personal data pulls in its obligatory
   document (see the feature→document map in the build sequence). Adding a feature is a
   legal event, not just a code change.

3. **The operator's identifying details appear in *every* document.** (152-ФЗ art. 18.1.)
   Legal name / sole-proprietor name, tax and registration numbers, registered address,
   and a contact email. A document without them cannot be relied on.

4. **Categorize cookies by legal necessity, not by vendor.** Strictly-necessary and
   functional storage need **no** consent and are always on; analytics and marketing need
   **prior** consent. The vendor is irrelevant to which bucket a key falls in — its
   *purpose* decides.

5. **The consent record itself needs no consent.** Storing the visitor's own accept/reject
   choice is the operator discharging a legal obligation (proving the choice), not a
   tracking act — it is always-on, and the cookie policy must say so.

6. **Prior-blocking is the load-bearing banner behavior.** No non-essential third-party
   request and no persistent identifier may occur before an affirmative decision — and
   this includes the paths that survive with JavaScript disabled (a no-JS tracking pixel)
   and identifiers created at page-init. "The script waits for consent" is not enough if a
   pixel or an ID beat it.

7. **Consent is versioned and auditable.** Store *who / what / when / which revision /
   from where*. When the meaning of a document or the banner's category set changes, its
   revision changes, and stored consents to the old revision must be re-collected. A
   cosmetic edit is not a new revision; a change in what you told people is.

8. **Withdrawal is as easy as granting.** A visible control re-opens the banner and lets
   the subject revoke. The subject's right to withdraw (art. 9) and to complain to the
   supervisory authority must be stated in the documents.

9. **The regime is keyed by the visitor's *jurisdiction*, never by a market they picked.**
   If the site serves more than one legal regime, a display preference (which market's
   content to show) must not select the *consent* model — that is a legal fact about where
   the person is, and keying it off a dropdown puts a breach one click away. Resolve
   conflicts strictest-wins. (This is a portable instance of the same principle content
   localization needs: never key a legal fact off an assumption.)

10. **Fail secure on every unknown.** An unresolved model, a missing category, a malformed
    law-layer entry — each must resolve to *more* blocking, not less. The harmless error is
    a banner someone didn't legally need; the unrecoverable one is a tracker that fired
    without consent.

## The ground it needs

Four things must be standing before the sequence starts. Each is a ladder: probe for it,
build the smallest version, or degrade with a label the reader of your output can see.

**A definite operator identity (the requisites).** *Why load-bearing:* the requisites are
mandatory content in every document (art. 18.1) and the operator notification cannot be
filed without them. *Probe:* is there a registered legal entity or sole proprietor behind
the site, with tax/registration numbers and a registered address? *Build the smallest:* if
the venture is not yet registered, that registration is itself a prerequisite — there is no
lawful "operator" without it. *Degrade:* none. This rung has no fallback; a site processing
personal data with no registered operator is unlawful, not merely under-documented.

**A complete inventory of what personal data you collect, why, and where it flows.** *Why
load-bearing:* the privacy policy must list every category of data, every purpose with its
legal basis, every recipient, and every cookie/storage key with its retention — and it must
be *complete and current*, because the enforcement gate is "code collects X that no document
discloses." *Probe:* can you enumerate, from the running system, every field a form or an
account takes, every event an analytics tag sends, every storage key the client writes, and
every third-party host the browser contacts? *Build the smallest:* a hand-written data map —
a table of (data point → purpose → legal basis → retention) and a list of (cookie/key →
category → lifetime) and (third-party host → what it receives → is it in Russia?). Twenty
rows, not a system. *Degrade:* if the inventory is partial, publish only what you can
actually stand behind and treat every undisclosed collection as a **defect to close before
launch**, not a rounding error — an undisclosed cookie or host is exactly the finding that
draws a fine.

**A place to record consents (the audit trail).** *Why load-bearing:* principle 7 —
consent you cannot prove is consent you did not obtain. *Probe:* is there an append-only
store you can write a consent row to, keyed by subject and revision? *Build the smallest:* a
single append-only table (or the anonymous visitor's own browser store for the cookie
choice) recording subject, document slug, revision, timestamp, choice. *Degrade:* at
minimum the cookie banner must persist the visitor's choice locally with the revision it was
made under; without even that, you cannot re-prompt correctly and cannot prove refusal.

**Legal review — this is YMYL.** *Why load-bearing:* the documents are binding legal
instruments; wrong content is worse than none. *Probe:* do you have access to a lawyer
competent in 152-ФЗ? *Build the smallest:* use this recipe's structure and the harvested
skeletons as a **draft**, then have it reviewed before you present it as binding. *Degrade:*
if you must publish before review, mark the documents as a draft pending legal review where
users can see it — never present unreviewed prose as a finished legal instrument.

## The contracts

Four shapes. None is read by a consumer outside your own system, so all are
**Binding: adapt** — rename fields into your own language, keep what they *mean*.

**The operator requisites block** — appears verbatim in every document. Placeholders where a
real value would identify a real person; fill them with your own.

**Binding: adapt**

```yaml
operator:
  kind: "legal-entity | sole-proprietor"
  name: "<full registered name>"          # legal name, or sole-proprietor's full name
  tax_id: "<taxpayer number>"             # ИНН
  registration_id: "<registration number>" # ОГРН / ОГРНИП
  address: "<registered address>"
  email: "<contact email for data-subject requests>"
```

**The cookie-category model** — drives what the banner shows and what it blocks. The
*needs_consent* and *blocking* flags are the legal core; getting `essential` or `functional`
wrong in either direction is a scar below.

**Binding: adapt**

```yaml
categories:
  - key: essential      # auth, security, the consent record itself
    needs_consent: false
    default_on: true
    blocking: false     # never held back — the site cannot run without it
  - key: functional     # theme, layout, "already subscribed" — no PII, no identification
    needs_consent: false
    default_on: true
    blocking: false
  - key: analytics      # traffic statistics, visitor identifiers
    needs_consent: true
    default_on: false   # MUST default OFF — opt-in, never pre-ticked
    blocking: true      # held back until an affirmative decision
  - key: marketing      # ad/retargeting tags, if any
    needs_consent: true
    default_on: false
    blocking: true
```

**The banner's stored decision** — persisted in the anonymous visitor's own browser. The
*revision* field is the trap: it binds the choice to the document version it was made under,
so a revision bump can invalidate it (principle 7).

**Binding: adapt**

```json
{
  "revision": "<banner revision this choice was made under>",
  "decidedAt": "<ISO-8601 timestamp>",
  "choice": { "essential": true, "functional": true, "analytics": false, "marketing": false }
}
```

**The consent audit record** — what you must be able to reproduce for any consent, cookie or
document. `subject` is either an account id or, for pre-account acts (newsletter, the cookie
choice), the email or the anonymous visitor id.

**Binding: adapt**

```json
{
  "subject": "<account id | email | anonymous visitor id>",
  "doc_slug": "<which document or 'cookie-categories'>",
  "revision": "<the document/banner revision agreed to>",
  "decided_at": "<ISO-8601 timestamp>",
  "source": "<where the act happened: registration | subscribe | cookie-banner>",
  "choice": "<accept | reject, or the per-category map for cookies>"
}
```

## The build sequence

### 1. Inventory the personal data, cookies, and third-party hosts

Enumerate, from the running system: every field a form or account collects; every property an
analytics event sends; every cookie / local-storage / session-storage key the client writes;
every external host the browser is told to contact. For each host, record what it receives
and whether it is inside Russia.

**Done when:** you have the three tables from the ground section filled, and nothing the code
does is missing from them.

### 2. Assemble the operator requisites

Fill the requisites contract from the registered operator's real details. If there is no
registered operator, stop — that registration is a hard prerequisite.

**Done when:** the requisites block is complete and matches the operator's registration
documents exactly.

### 3. Decide the document set from your features

Map features to obligatory documents. The set is at minimum a **privacy policy** and a
**terms of use**; each capability below pulls in one more:

| A feature that… | obliges the document | resting on |
|---|---|---|
| processes any personal data at all | Privacy policy (Политика обработки ПД) | 152-ФЗ art. 18.1 |
| lets people register / hold an account | Consent to processing (Согласие на обработку ПД) | 152-ФЗ art. 9 |
| publishes user data to an open audience (public profiles, comments) | Consent to *distribution* (Согласие на обработку ПД, разрешённых для распространения) | 152-ФЗ art. 10.1 |
| sends a newsletter / marketing email | Consent to mailings (Согласие на получение рассылок) | 152-ФЗ art. 9 + 38-ФЗ art. 18 |
| sets any cookie / browser storage | Cookie policy (Политика использования файлов cookie) | 149-ФЗ + cookie practice |
| has any usage rules at all | Terms of use (Пользовательское соглашение) | civil law |

Each consent is a **separate** document; the terms of use references the consents but never
contains them.

**Done when:** you can name, for each of your live features, the document it obliges — and
each consent is a standalone document, not a clause.

### 4. Write each document with its mandatory content

Every document opens with the requisites block and a revision date. Each consent document
carries, at minimum (152-ФЗ art. 9 ч. 4): the operator's identity; the subject's identity
and how they are identified; the exhaustive **list of personal data**; the **purposes**; the
**list of actions and the method** of processing; the **term/validity** and the **withdrawal
method**. The privacy policy additionally lists data-subject **rights** (152-ФЗ art. 14),
names **third-party recipients**, states **data localization** (primary databases in
Russia, art. 18), and — if any recipient is outside Russia — a cross-border-transfer block
(art. 12). State that non-special, non-biometric data only is processed, if that is true.

**Done when:** each document carries the requisites, a revision date, and every field its
governing article requires; a reviewer can check each field against the article and find
nothing missing.

### 5. File the operator notification with the supervisory authority

Before processing begins, file the operator notification (уведомление об обработке
персональных данных) with Roskomnadzor, unless you meet one of the narrow statutory
exemptions (152-ФЗ art. 22 ч. 2 — essentially: paper-only processing, a handful of
security/transport carve-outs). A public site with analytics, accounts, or a newsletter
meets none of them.

**Done when:** the notification is filed (or the specific exemption clause that releases you
is written down), *before* the site starts processing.

### 6. Classify every cookie/storage key into the category model

Put each key from step 1 into `essential`, `functional`, `analytics`, or `marketing` by its
**purpose**. Auth tokens, security, and the consent record itself are `essential`. Theme,
layout, "already subscribed" flags with no PII are `functional`. Anything that identifies or
counts a visitor is `analytics` or `marketing`. The cookie policy documents each key, its
purpose, and its retention.

**Done when:** every key from step 1 is in exactly one category, the essential/functional
ones carry no identifying data, and the cookie policy lists them all.

### 7. Build the banner with prior-blocking

The banner shows the categories from the contract, with non-essential ones defaulting **off**
and reject as prominent as accept. Until an affirmative decision is stored: every `blocking`
category's scripts are held back, no persistent identifier is written, and — critically —
the no-JS fallback paths (a tracking pixel in a `<noscript>`) are also suppressed. On accept,
the blocked scripts initialize; on reject, the clean state persists. A visible control
re-opens the banner for withdrawal.

**Done when:** in a clean browser, the network panel shows no request to any non-essential
third-party host and storage holds no tracking identifier until accept is clicked; reject
preserves that; essential/functional keys are present throughout.

### 8. Wire consent recording, versioning, and re-prompting

Persist the stored-decision shape with the banner revision it was made under; write account
and newsletter consents to the audit record with their document revision. When a document or
the banner category set changes meaning, bump its revision so stored consents to the old
revision are re-collected. Asymmetric shelf life is good practice: keep an *acceptance* far
longer than a *refusal* before re-asking, and never destroy a stored consent merely because
the current revision is momentarily unknown (a lagging config load must not delete a
legally-recorded choice).

**Done when:** a consent produces an auditable record bound to a revision; a revision bump
re-prompts affected visitors; and a transient failure to read the current revision keeps
existing consents rather than wiping them.

## The seams

- **The operator requisites** — every value in the requisites block. The recipe fixes the
  *fields*; the operator supplies the values.
- **The document set** — which documents exist is feature-driven; a project with no
  newsletter has no newsletter consent. Never hardcode a fixed list.
- **Retention periods** — how long each cookie/consent lives, and the accept-vs-reject
  asymmetry, are policy values, not constants baked into the banner.
- **The banner revision string** — the token that re-prompts on change. Its format is
  yours; its *meaning* (a change here re-collects consent) is fixed.
- **The category list** — the four categories here are the common set; a site with no
  marketing tags omits that category. What must not vary is the *needs_consent/blocking*
  logic per category.
- **The jurisdiction regime** — for a multi-market site, which legal model applies is
  resolved per visitor-jurisdiction; the *resolution seam* (and strictest-wins default)
  stays configurable, the keying-off-jurisdiction rule does not.

## The scars

### The no-JS tracking pixel fired before consent

**Symptom:** a clean, undecided visitor still produced a request to the analytics vendor's
host — visible in the network panel before any banner interaction.
**Root cause:** the analytics *script* was correctly gated on consent, but the vendor's no-JS
fallback — a tracking pixel inside a `<noscript>` — rendered unconditionally in the served
HTML and fired the moment the page loaded, regardless of the consent state the script
respected.
**Fix:** gate the no-JS pixel on the same consent signal as the script. Prior-blocking is not
"the script waits"; it is "no non-essential request occurs," and the pixel is a request.

### A persistent visitor identifier was minted before any decision

**Symptom:** a first-time visitor who had touched nothing already had a stable tracking id in
browser storage.
**Root cause:** the identifier was created at composable/page init — eagerly, before the
consent state was even read — so the "create if absent" ran on first paint.
**Fix:** create the identifier only after an affirmative analytics decision. An identifier
written before consent is the same violation as a tracker firing before consent.

### The category gate failed *open* on an unknown category

**Symptom:** a category not present in the resolved model was treated as non-blocking, so its
scripts were allowed to run.
**Root cause:** the blocking check defaulted to "allow" when it could not find the category —
the permissive default.
**Fix:** default the blocking check to **blocking** on any category it cannot resolve. Under
principle 10, the unknown must resolve to more blocking, not less.

### Functional storage was wrongly held behind consent

**Symptom:** theme and layout preferences stopped working for visitors who hadn't yet decided
— the site felt broken before the banner was answered.
**Root cause:** the banner treated *all* non-essential storage as consent-gated, sweeping in
functional keys that carry no personal data and legally need no consent.
**Fix:** separate `functional` (always-on, no consent) from `analytics` and `marketing` (gated).
Over-blocking is a real failure, not a safe default — it degrades the product for a
requirement the law does not impose.

### A cosmetic edit re-prompted every live visitor

**Symptom:** a trivial wording change to a policy silently invalidated every stored cookie
consent, and every visitor saw the banner again.
**Root cause:** the banner revision was coupled to the document's revision date, and any edit
bumped the date — so a typo fix counted as a new legal revision.
**Fix:** bump the revision only when the *meaning* changes (the category set, a purpose, a
recipient) — not for cosmetic edits. The revision is a legal event; spend it deliberately.

### The consent model was keyed off the selected market

**Symptom:** a visitor in a strict-consent jurisdiction who switched the site to a
looser-jurisdiction market's content received the looser consent model — and non-essential
trackers fired without the prior consent their real location required. One dropdown click,
silent breach.
**Root cause:** the consent regime was resolved from the *market* the visitor picked (a
display preference) instead of their *jurisdiction* (a legal fact).
**Fix:** key the regime on jurisdiction, never on the selected market; resolve conflicts
strictest-wins. The reverse error — showing a stricter banner than needed — is harmless; this
one is not.

### A malformed law-layer entry resolved to a weaker model

**Symptom:** an invalid entry in the jurisdiction law layer was accepted and silently
resolved to a less-blocking consent model.
**Root cause:** the resolver tolerated a schema-invalid law-layer record instead of rejecting
it, and the tolerance degraded toward *allow*.
**Fix:** enforce the law-layer schema and fail secure — a malformed entry must block the
build or resolve to the strictest model, never to a weaker one.

## Verification

Ship these gates *with* the capability — the invariants rot the day after launch without
them:

- **Cookie-in-code vs disclosed-in-policy:** every cookie/storage key the code sets must be
  described in the cookie policy. Fails the build on an undisclosed key.
- **External-host vs declared-recipient:** every third-party host the browser is told to
  contact must be a declared recipient in the privacy policy; a host outside Russia raises a
  cross-border-transfer warning (art. 12).
- **Banner-revision ⇔ cookie-policy lockstep:** the banner's current revision date and the
  cookie policy's revision date must move together (a category change is a policy change).
- **Feature ⇒ required-document:** a feature that is enabled but whose obligatory document is
  missing fails the build.

**The one live check no gate performs:** open the published site in a fresh browser with no
prior consent, watch the network panel and storage. Zero requests to any non-essential
third-party host; no persistent tracking identifier written. Click reject — the state stays
clean. Reload — still clean. Click accept — now, and only now, the analytics requests appear
and the identifier is written. Essential and functional keys are present the whole time.

## The trade-offs

This design is deliberately heavyweight. It buys defensibility — a documented, auditable,
prior-blocking posture that survives a supervisory-authority inquiry — and it costs:

- **Several documents, not one**, each with mandatory content and each a maintenance
  surface. The separation is legally required, not optional simplification.
- **Analytics coverage.** Prior-blocking means you lose the visitors who decline or never
  answer. That is the price of opt-in; do not "recover" it by firing before consent.
- **Offline bureaucracy.** The operator notification and (if applicable) cross-border filing
  are paperwork with the authority, outside any codebase, with real lead time.
- **A revision discipline.** Every meaning-change re-collects consent; careless edits punish
  your own conversion.

**When not to build the full apparatus:** a purely static site with no accounts, no forms,
no newsletter, no analytics, and no third-party embeds processes no personal data and needs
at most a short cookie disclosure — not the consent machinery. But the moment you add an
analytics tag, a contact form, comments, or accounts, you cross the line into 152-ФЗ and the
full set applies. The failure mode is assuming you are on the static side of that line when
one analytics snippet already put you on the other.

## For the human

You get two things. First, a **set of separate legal documents** — a privacy policy, the
consents your features oblige (account, distribution, newsletter), a cookie policy, and terms
of use — each carrying the operator's identifying details, each citing the article it rests
on, and an operator notification filed with the supervisory authority before you process
anything. Second, a **cookie banner** that holds every non-essential tracker back until the
visitor affirmatively opts in, records the choice with the revision it was made under, and
lets them withdraw — while strictly-necessary and functional storage just works.

**The stack it was built on, and why.** The documents are **human-readable prose** and the
banner is a **client-side gate over the browser's own storage**.

- **Essential** (the design depends on these): the documents must be a *durable, versioned,
  human-readable record* of exactly what you told people — because that record is what a
  regulator or a court reads. And the prior-blocking gate must run *in the client, before any
  non-essential script or pixel executes* — because "block before it runs" cannot be done
  after the fact on the server alone.
- **Incidental** (swap freely): that the documents happen to be Markdown, that the banner
  happens to be one particular component framework, that the store happens to be one
  particular table. An agent on another stack should replace all three without a second
  thought — the recipe holds.

**The stack's cost:** prose documents are easy to read and diff but drift from the code
unless a gate ties them together (hence the cookie-in-code and host-vs-recipient checks); a
client-side gate is the only place prior-blocking *can* live, but it means the correctness of
your compliance rides on client code that must fail secure.

**Optional — you can generate all of this from one source.** Instead of hand-writing and
hand-syncing the documents, you can drive the entire set from a single machine-readable
**registry**: a data map (each data point → purpose → legal basis → retention), the cookie
catalog, the recipient list, and the operator requisites. A generator renders the documents
and the banner's category list from that registry, and the verification gates compare *code*
against the *registry* — so a newly collected field or a new cookie cannot ship without
updating the disclosure, because the build fails otherwise. A working example of this exists.
It is real machinery with real cost, and it is **not required** to be compliant — the
hand-authored path in this recipe is complete on its own. The registry is what turns
"remember to update the policy" from a discipline into a gate; adopt it when the manual sync
becomes the thing you keep getting wrong.
