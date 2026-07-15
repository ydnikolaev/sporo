# sporo

A single self-installing CLI that turns a build you already did into a **recipe**: one
self-contained file that teaches an agent in a repository that has never heard of yours how to
build the same capability — on the same principles, possibly on a different stack, without
repeating your scars.

A recipe names roles, not coordinates; it *shows* the shapes it consumes rather than describing
them; it states its preconditions as a ladder (probe → build the smallest → degrade with a
label); and it carries an adoption protocol so a reader with no tooling at all can follow it.

> Private, pre-release. The name, domain (`sporo.dev`), and per-repo home (`.sporo/`) are fixed;
> the marketplace (a site for open-sourcing and team-sharing recipes) is a later spec.

## The verbs today

```
sporo init      # install the authoring surface into this repo (skill, AGENTS.md block, seeds)
sporo update    # re-sync the managed surface from this binary — never clobbering your edits
sporo genre     # print the authoring spec this binary enforces
sporo harvest   # mine the project's own record for a recipe's raw material
sporo lint      # check a recipe corpus against the genre — shape, scars, neutrality, seals
sporo seal      # record version + content hash in the registry: a sealed recipe never silently mutates
sporo export    # print one recipe as a single self-contained file (protocol appended)
sporo list      # the recipes available here — this project's own, and the official corpus
sporo feedback  # file and list report-backs — the channel a recipe's next version comes from
sporo review    # build a self-contained review pack for ANY agent; verify the verdicts it returns
sporo projects  # the repositories on this machine sporo was installed into
```

Coming in later releases: `upgrade` (self-update the binary; then `sporo update` per repo
pushes its newer skills), bundles (compose several recipes into one handed-over document),
and the site verbs (`push` / `pull`).

## Build

```
go build ./cmd/sporo
go test ./...
```

sporo is developed on the [mate](https://github.com/ydnikolaev/mate) harness — it is a consumer
of mate's doctrines, rules, and skills, and depends on none of its code.
