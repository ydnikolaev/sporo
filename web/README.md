# web/

Two layers, deliberately separate:

- `content/` — the copy, as markdown. The source of truth for the words.
- `site/` — the temporary landing: hand-written static HTML, zero JS, zero external
  requests, inline CSS. No toolchain on purpose — this site is replaced wholesale by the
  marketplace app later, and a build system it never needed would be debt it never repays.

Preview locally:

```
python3 -m http.server 4173 --directory web/site
```

Deploy target (when it happens): any static host behind sporo.dev — the pages assume
nothing but a web server.
