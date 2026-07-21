// The shared seal-currency seam: the trust-critical logic that decides whether a corpus file is
// shown "sealed". Both the recipe reader (corpus.ts) and the seed reader (seeds.ts) need the exact
// same check — a recipe is sealed only when its file STILL hashes to the value the CLI committed —
// so it is single-sourced here rather than duplicated, where a divergence would let a badge outrun
// the bytes it claims. corpus.ts keeps its own public surface as thin wrappers over this module.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import { parse as parseYaml } from 'yaml';

// The seal registry the CLI writes at `.sporo/registry.yaml` (repo root, one level up from the
// corpus). It records, per slug, the sha256 of the source bytes at seal time. The site reads it so
// the "sealed" badge is not decorative: a file is shown sealed only when it STILL hashes to the
// sealed value — the same coherence `sporo lint` enforces in CI. Reading the committed registry and
// recomputing the hash is a verification anyone can reproduce from the repo. The map is FLAT, keyed
// by slug, schema 2; each entry carries an optional `kind` (`recipe`|`seed`) that an absent value
// defaults to `recipe` (the committed registry has no `kind:` lines on its recipe entries).
const registryPath = path.resolve(process.cwd(), '../.sporo/registry.yaml');

export interface SealEntry {
  hash?: string;
  sealed_at?: string;
  kind?: string;
}
export function loadRegistry(): Record<string, SealEntry> {
  try {
    const doc = parseYaml(fs.readFileSync(registryPath, 'utf-8')) as {
      recipes?: Record<string, SealEntry>;
    };
    return doc?.recipes ?? {};
  } catch {
    // No registry, or unreadable — nothing is shown sealed; the badge degrades to gate-passed.
    return {};
  }
}
const registry = loadRegistry();

// sealVerified recomputes the seal currency the CLI uses — sha256 over the file's RAW source bytes,
// provenance banner INCLUDED (the seal hashes the file as sealed, not the banner-stripped display
// form) — and returns true only when it matches the committed seal. A drifted or unsealed file
// returns false, so the badge can never outrun the bytes it claims.
//
// `dir` is the corpus directory the file lives in (recipes/ or seeds/); `kind` gates which consumer
// this entry may serve, and the two paths treat `kind` asymmetrically ON PURPOSE:
//   - kind === 'recipe' — default-then-check: an entry whose `kind` is absent OR `'recipe'` is a
//     recipe. Requiring a strict `'recipe'` would make every recipe badge vanish today (the
//     committed registry has no `kind:` fields) and break recipe-page byte-identity.
//   - kind === 'seed' — strict: only an explicit `kind === 'seed'` entry is a seed. This stops a
//     colliding recipe slug from cross-verifying against a seed's bytes (routes are namespaced
//     precisely because slugs can collide across kinds).
export function sealVerified(dir: string, slug: string, kind: 'recipe' | 'seed'): boolean {
  const entry = registry[slug];
  if (!entry?.hash) return false;
  const entryKind = entry.kind ?? 'recipe';
  if (kind === 'seed' ? entryKind !== 'seed' : entryKind !== 'recipe') return false;
  const raw = fs.readFileSync(path.join(dir, `${slug}.md`)); // Buffer: the exact bytes Go sealed
  const got = 'sha256:' + crypto.createHash('sha256').update(raw).digest('hex');
  return got === entry.hash;
}

// sealHash returns the sha256 the registry committed for a slug (or null). It is the fingerprint the
// trust panel SHOWS; the panel's actual verification recomputes independently against GitHub, so
// this value is display, not the root of trust. Registry-only — no directory or kind needed.
export function sealHash(slug: string): string | null {
  return registry[slug]?.hash ?? null;
}

// sealedAt returns the machine-recorded seal timestamp (RFC3339) the registry committed, or null.
// Unlike the frontmatter's author-typed `verified.date`, this is stamped by `sporo seal` itself.
export function sealedAt(slug: string): string | null {
  return registry[slug]?.sealed_at ?? null;
}
