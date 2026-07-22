import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';

// The `attested` provenance mark, verified at BUILD, fail-closed. `isAttested` asks GitHub whether a
// Sigstore/OIDC attestation exists for the EXACT export-mirror bytes this site serves — the same
// `web/src/data/exports/<slug>.md` the corpus workflow (attest-corpus.yml) checksummed and signed. It
// mirrors the releases fetch: a plain `fetch` against api.github.com, degrade-gracefully, no runtime
// request (resolved once at build, baked into the static page).
//
// It does NOT re-verify the Sigstore signature here — the site is not the root of trust. The whole
// point of the mark is that the ADOPTER verifies offline themselves, from the file alone:
// `gh attestation verify <file> -R ydnikolaev/sporo`. This check only earns the right to SHOW the
// mark: any missing token, API error, 404, or empty result returns false, so the mark never renders
// for bytes that are not genuinely attested. Same posture as `sealVerified` — a claim the page makes
// only when it holds, never by accident.
const REPO = 'ydnikolaev/sporo';
const exportsDir = path.resolve(process.cwd(), 'src/data/exports');

export async function isAttested(slug: string): Promise<boolean> {
  let digest: string;
  try {
    // Buffer (no encoding): the exact bytes `sha256sum` hashed in the workflow, byte-for-byte.
    const raw = fs.readFileSync(path.join(exportsDir, `${slug}.md`));
    digest = 'sha256:' + crypto.createHash('sha256').update(raw).digest('hex');
  } catch {
    return false; // no export mirror for this slug — nothing could have been attested
  }

  const token = process.env.GITHUB_TOKEN || process.env.GH_TOKEN;
  const headers: Record<string, string> = {
    Accept: 'application/vnd.github+json',
    'User-Agent': 'sporo-site-build',
    'X-GitHub-Api-Version': '2022-11-28',
  };
  if (token) headers.Authorization = `Bearer ${token}`;

  try {
    // The attestations API is keyed by subject digest and scoped to the repo — only this repo's
    // workflow (which alone holds `attestations: write` here) can have created one for these bytes.
    const res = await fetch(`https://api.github.com/repos/${REPO}/attestations/${digest}`, { headers });
    if (res.status === 404) return false; // no attestation for these bytes yet (e.g. before the first attest run)
    if (!res.ok) throw new Error(`GitHub API ${res.status}`);
    const body = (await res.json()) as { attestations?: unknown[] };
    return Array.isArray(body.attestations) && body.attestations.length > 0;
  } catch (e) {
    // Fail-closed: an unreachable API withholds the mark, it never asserts one. Logged so a build
    // that silently stops showing the mark is diagnosable (mirrors the releases fetch).
    console.warn(`[attest] ${slug} attestation check failed, mark withheld: ${(e as Error).message}`);
    return false;
  }
}
