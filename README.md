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
sporo harvest   # mine the project's own record for a recipe's raw material
sporo lint      # check a recipe corpus against the genre — shape, scars, neutrality
sporo export    # print one recipe as a single self-contained file (protocol appended)
sporo list      # the recipes available here — this project's own, and the official corpus
```

Coming in later releases: `init` / `update` (install and drift-sync the authoring skill into a
repo's provider homes), `upgrade` (self-update the binary), and the site verbs (`push` / `pull`).

## Build

```
go build ./cmd/sporo
go test ./...
```

sporo is developed on the [mate](https://github.com/ydnikolaev/mate) harness — it is a consumer
of mate's doctrines, rules, and skills, and depends on none of its code.
