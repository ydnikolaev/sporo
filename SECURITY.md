# Security Policy

## Supported versions

Only the **latest release** of `sporo` receives security fixes. Upgrade with
`sporo upgrade` (or reinstall from the latest GitHub release) before reporting an
issue, and confirm it still reproduces there.

## Reporting a vulnerability

Please report security issues **privately** — do not open a public issue.

Use GitHub's private advisory flow: go to the repository's **Security** tab and
click **[Report a vulnerability](https://github.com/ydnikolaev/sporo/security/advisories/new)**.
This opens a private channel visible only to the maintainers.

Include, where you can:

- the `sporo --version` you reproduced on, and your OS/arch;
- the exact steps or a minimal recipe/command that triggers the issue;
- what an attacker gains, and any known workaround.

## Response expectations

- **Acknowledgement:** within 3 business days.
- **Assessment and next steps:** within 7 business days.
- **Fix:** shipped in the next release once a fix is ready; you'll be credited in
  the advisory unless you ask otherwise.

## Verifying release binaries

Every release is signed and carries build provenance, so you can confirm a binary
was built by this repository's release workflow before trusting it:

```bash
# SLSA build provenance (needs only the GitHub CLI):
gh attestation verify sporo_<version>_<os>_<arch>.tar.gz -R ydnikolaev/sporo

# cosign keyless signature over the checksums file:
cosign verify-blob \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp '^https://github.com/ydnikolaev/sporo/.github/workflows/release.yml@refs/tags/v.*$' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  checksums.txt
```
