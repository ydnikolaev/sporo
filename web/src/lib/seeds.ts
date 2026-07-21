// Build-time access to the embedded seed corpus that lives at the repo root (../seeds). The sibling
// of corpus.ts: the site reads these files DIRECTLY at build — no snapshot, no gate — so the seed
// cards, the detail pages, and llms-full.txt can never drift from the seeds themselves; they are the
// seeds, rendered. (process.cwd() during `astro build` is web/, so ../seeds resolves to the repo-root
// corpus, a sibling of ../recipes. The Pages workflow's paths filter includes seeds/** so a seed edit
// rebuilds the site.) This is the second instance of the reader pattern — it mirrors corpus.ts idiom
// on purpose; the two readers stay siblings, each leaning only on the shared seal seam.
import fs from 'node:fs';
import path from 'node:path';
import { marked } from 'marked';
import {
  sealVerified as sealVerifiedSeam,
  sealHash,
  sealedAt,
} from './seal.ts';

const seedsDir = path.resolve(process.cwd(), '../seeds');

// The seal-currency check lives in the shared seam (lib/seal.ts); the seed path is the STRICT case —
// only an entry with an explicit `kind === 'seed'` is a seed (a missing or `'recipe'` kind is never
// one), which stops a colliding recipe slug from cross-verifying against a seed's bytes. Routes are
// namespaced under /seeds/ precisely because slugs can collide across kinds.
export function sealVerified(slug: string): boolean {
  return sealVerifiedSeam(seedsDir, slug, 'seed');
}
// sealHash/sealedAt are registry-only (directory-independent), re-exported unchanged from the seam.
export { sealHash, sealedAt };

// The source seeds carry a leading HTML-comment provenance banner (the SSOT marker). It is not
// content; strip it before parsing or rendering.
function stripBanner(t: string): string {
  return t.replace(/^<!--[\s\S]*?-->\s*/, '');
}

function frontmatter(raw: string): { meta: Record<string, string>; body: string } {
  const m = raw.match(/^---\n([\s\S]*?)\n---\n([\s\S]*)$/);
  if (!m) return { meta: {}, body: raw };
  const meta: Record<string, string> = {};
  for (const line of m[1].split('\n')) {
    // Only the flat scalar fields (id, name, version, title, target, source, effort). The nested
    // stack/verified objects are skipped here and mined below where a page needs them. `target` and
    // `source` — the seed's two trust anchors — are flat scalars, so they land straight in `meta`.
    const mm = line.match(/^([a-z_]+):\s+(.+)$/);
    if (mm && !mm[2].startsWith('{') && !mm[2].startsWith('[')) {
      meta[mm[1]] = mm[2].replace(/^["']|["']$/g, '');
    }
  }
  // The seed's honesty stamp lives inside the nested `verified: { project, release, date }` object —
  // whose machine proved this install, at which release, on which date. Mine the three out for the
  // card/detail provenance line (the recipe reader mines only date+project; a seed pins release too).
  const vd = m[1].match(/^verified:\s*\{[^}]*\bdate:\s*"?(\d{4}-\d{2}-\d{2})"?/m);
  if (vd) meta.date = vd[1];
  const vp = m[1].match(/^verified:\s*\{[^}]*?\bproject:\s*([^,}]+)/m);
  if (vp) meta.project = vp[1].replace(/^["']|["']$/g, '').trim();
  const vr = m[1].match(/^verified:\s*\{[^}]*?\brelease:\s*([^,}]+)/m);
  if (vr) meta.release = vr[1].replace(/^["']|["']$/g, '').trim();
  return { meta, body: m[2] };
}

// seedSummary returns the `## Summary` section as plain markdown text (banner/frontmatter already
// stripped by the caller). It is the card's orienting blurb — a seed has no `problem` frontmatter
// scalar the way a recipe does, so the blurb is mined from the body's required Summary section (the
// genre holds it to ≥80 visible characters). Raw markdown, trimmed; the card renders it as text.
function seedSummary(body: string): string {
  const part = body
    .split(/\n(?=## )/)
    .find((p) => /^## Summary\s*$/m.test(p.split('\n')[0] ?? '')) ?? '';
  return part.replace(/^## Summary\s*\n/, '').trim();
}

export interface SeedSection {
  title: string;
  html: string;
  open: boolean;
}
export interface Seed {
  meta: Record<string, string>;
  introHtml: string;
  summaryHtml: string;
  sections: SeedSection[];
  sealed: boolean;
  raw: string;
}

export function readSeed(slug: string): Seed {
  const raw = stripBanner(fs.readFileSync(path.join(seedsDir, `${slug}.md`), 'utf-8'));
  const { meta, body } = frontmatter(raw);
  // Split on level-2 headings; everything before the first `## ` is the intro (# title + thesis
  // blockquote). The Summary section is pulled OUT and surfaced as summaryHtml (like the recipe
  // reader), leaving the remaining gated sections — What it is / Install / Verify / Use / Harness /
  // Report — as the section list.
  const parts = body.split(/\n(?=## )/);
  const bodySections = parts.slice(1);
  const summaryPart = bodySections.find((p) => /^## Summary\s*$/m.test(p.split('\n')[0] ?? '')) ?? '';
  const sections: SeedSection[] = bodySections.filter((p) => p !== summaryPart).map((p) => {
    const title = (p.match(/^## (.+)/)?.[1] ?? '').trim();
    const md = p.replace(/^## .+\n/, '');
    // A seed has no scars section to open by default (that is a recipe payload); every section
    // starts collapsed, and the detail page decides its own emphasis.
    return { title, html: marked.parse(md) as string, open: false };
  });
  // Drop the seed's own `# Title` H1 from the intro — the page already shows the title as its H1.
  // The body starts with a blank line after the frontmatter fence, so the H1 is NOT at index 0;
  // allow leading whitespace or it survives the strip.
  const intro = parts[0].replace(/^\s*#\s+.+\n+/, '');
  const summaryMd = summaryPart.replace(/^## Summary\s*\n/, '');
  return {
    meta,
    introHtml: marked.parse(intro) as string,
    summaryHtml: marked.parse(summaryMd) as string,
    sections,
    sealed: sealVerified(slug),
    raw,
  };
}

// The raw markdown of one seed, banner stripped but frontmatter INTACT — for the detail `.md` twin
// and the llms-full.txt corpus dump (the twin serves the seed as a reader receives it, provenance
// and all).
export function rawSeed(slug: string): string {
  return stripBanner(fs.readFileSync(path.join(seedsDir, `${slug}.md`), 'utf-8'));
}

// The seed genre spec (seeds/_authoring.md), banner stripped — the constitutional document the
// what-is-a-seed page renders and llms-full folds in. There is no separate adoption protocol: a
// seed's Report/Harness sections are intrinsic to its body, so the genre carries one version only.
export function seedGenreSpec(): string {
  return stripBanner(fs.readFileSync(path.join(seedsDir, '_authoring.md'), 'utf-8'));
}

export function seedGenreVersion(): string {
  return frontmatter(seedGenreSpec()).meta.version ?? '';
}

// The site shows the version history from the spec itself, not a second hand-maintained copy. A
// heading renumber can move it without changing this parser; its semantic title is the key.
export function seedGenreChangelogHtml(): string {
  return marked.parse(seedGenreChangelogMarkdown()) as string;
}

export function seedGenreChangelogMarkdown(): string {
  const { body } = frontmatter(seedGenreSpec());
  const start = body.search(/^## \d+\. Version history\s*$/m);
  if (start < 0) return '';
  const section = body.slice(start).replace(/^## .+\n/, '');
  const next = section.search(/^## /m);
  return (next >= 0 ? section.slice(0, next) : section).trim();
}

export interface SeedEntry {
  slug: string;
  meta: Record<string, string>;
  summary: string;
  sealed: boolean;
}

// listSeeds enumerates the official seeds (the `_`-prefixed files — `_authoring.md`, `_runner.md` —
// are the genre spec and its kin, not seeds). It carries the slug — the filename minus `.md` —
// because that is the seed's route key (`/seeds/<slug>.html`) that `getStaticPaths` builds every
// detail page and `.md` twin from, plus the Summary blurb the card needs (a seed has no `problem`
// scalar). The readdir is UNGUARDED, mirroring corpus.ts: seeds/ holds the tracked `_authoring.md`,
// so it never ENOENTs; on today's seed-free corpus this returns [].
export function listSeeds(): SeedEntry[] {
  return fs
    .readdirSync(seedsDir)
    .filter((f) => f.endsWith('.md') && !f.startsWith('_'))
    .map((f) => {
      const slug = f.replace(/\.md$/, '');
      const { meta, body } = frontmatter(stripBanner(fs.readFileSync(path.join(seedsDir, f), 'utf-8')));
      return {
        slug,
        meta,
        summary: seedSummary(body),
        sealed: sealVerified(slug),
      };
    });
}
