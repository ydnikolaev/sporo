# sporo — security & provenance

You run a binary you downloaded from the internet, so you should be able to prove it is
really ours before you trust it. Every sporo release is signed, attested, and inventoried,
and the source it is built from is continuously scanned. None of this needs an account, a
key, or our servers.

## Verify the binary

Every release is signed with **cosign** (keyless — there is no signing key you have to
trust), carries a **SLSA build-provenance attestation** binding the binary to this
repository's release workflow, and ships a full **SBOM**. Confirm what you downloaded is
really ours, offline, in two commands:

```
# build provenance — who built it, and from which commit
gh attestation verify sporo_…_linux_amd64.tar.gz -R ydnikolaev/sporo
# → built by ydnikolaev/sporo · release.yml

# signature — the keyless cosign bundle over the checksums
cosign verify-blob --bundle checksums.txt.sigstore.json checksums.txt
# → Verified OK
```

## The source is continuously scanned

- **CodeQL** — static security analysis on every push and weekly.
- **govulncheck** — the official Go vulnerability scanner, on every PR and daily.
- **gitleaks** — secret scanning across the working tree and full history.
- **Dependabot** — automated dependency + GitHub-Actions updates, security updates immediate.
- **coverage-gated** and **SHA-pinned CI** — a coverage floor enforced from one source of
  truth, and every GitHub Action pinned to a full commit SHA (no moving tags to hijack).

## We show the findings, not just the badge

`govulncheck` runs on every PR and daily. A vulnerability we cannot yet fix — an unmaintained
transitive dependency with no patched version — is **not hidden**: it is recorded in a public,
commented allowlist ([.govulncheck-allow.txt](https://github.com/ydnikolaev/sporo/blob/main/.govulncheck-allow.txt))
with the reason, so the gate stays green on the known one and loud on any new one. Same
discipline as the recipe seals: a claim is only made when it is earned.

## Report a vulnerability

Please do not open a public issue for a security problem. Use GitHub's private advisory flow —
[Report a vulnerability](https://github.com/ydnikolaev/sporo/security/advisories/new) on the
repository's Security tab, a channel visible only to the maintainers.

- **Acknowledgement** within 3 business days; assessment and next steps within 7.
- **Supported version** — only the latest release receives fixes; upgrade with `sporo upgrade`
  and confirm it still reproduces.
- **Credit** in the published advisory, unless you would rather stay anonymous.

Full policy: [SECURITY.md](https://github.com/ydnikolaev/sporo/blob/main/SECURITY.md).
