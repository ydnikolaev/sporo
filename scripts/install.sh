#!/bin/sh
# Installs the latest `sporo` release from GitHub. POSIX sh — no bashisms — so it
# runs under whatever /bin/sh the user's box ships (dash, ash, busybox, ...).
#
# The repo (ydnikolaev/sporo) is PUBLIC, so no token is needed: the release-asset
# API (`/releases/assets/{id}`) serves a public repo's assets to an unauthenticated
# request. A GITHUB_TOKEN is honoured if present — it only lifts GitHub's low
# unauthenticated API rate limit, nothing more — but it is entirely optional.

set -eu

OWNER="ydnikolaev"
REPO="sporo"
BINARY="sporo"

log() { printf 'sporo-install: %s\n' "$*" >&2; }
die() { log "$*"; exit 1; }

# --- 1. detect OS/arch, map to goreleaser's archive naming ------------------

os_raw="$(uname -s)"
arch_raw="$(uname -m)"

case "$os_raw" in
  Darwin) os="darwin" ;;
  Linux) os="linux" ;;
  MINGW*|MSYS*|CYGWIN*) os="windows" ;;
  *) die "unsupported OS: $os_raw" ;;
esac

case "$arch_raw" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) die "unsupported architecture: $arch_raw (sporo ships amd64/arm64 only)" ;;
esac

ext="tar.gz"
[ "$os" = "windows" ] && ext="zip"

asset_name="${BINARY}_LATEST_${os}_${arch}.${ext}"
checksums_name="checksums.txt"

# --- 2. auth headers (optional; public repo needs none) ---------------------

auth_header=""
if [ -n "${GITHUB_TOKEN:-}" ]; then
  auth_header="Authorization: Bearer ${GITHUB_TOKEN}"
fi

api="https://api.github.com/repos/${OWNER}/${REPO}"

curl_json() {
  if [ -n "$auth_header" ]; then
    curl -fsSL -H "$auth_header" -H "Accept: application/vnd.github+json" "$1"
  else
    curl -fsSL -H "Accept: application/vnd.github+json" "$1"
  fi
}

# --- 3. resolve the latest release + matching asset ids ---------------------

release_json="$(curl_json "${api}/releases/latest")" \
  || die "couldn't reach GitHub's releases API for ${OWNER}/${REPO} (private repo? check GITHUB_TOKEN)"

tag="$(printf '%s' "$release_json" | grep -m1 '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
[ -n "$tag" ] || die "couldn't parse a tag_name out of the releases/latest response"
version="${tag#v}"

asset_name="${BINARY}_${version}_${os}_${arch}.${ext}"
checksums_name="checksums.txt"

# Pull "id" + "name" pairs out of the assets array and grep for the ones we want.
# (Not using jq: it isn't guaranteed present, and this script has exactly two
# fields to extract.)
#
# GitHub returns the release JSON pretty-printed — one field per line, each ending
# in a comma. `tr ',' '\n'` therefore turns every `,\n` into a BLANK line, so an
# asset object splits to `"id":…`, «», `"node_id":…`, «», `"name":…`, and the id
# ends up FOUR lines above the name, out of a `-B2` window's reach. Dropping the
# blank lines first restores id-is-two-lines-above-name, which holds for both the
# pretty and the compact JSON shape. (This is the exact bug that hid until the
# install path was driven end-to-end against a real release.)
find_asset_id() {
  name="$1"
  printf '%s' "$release_json" \
    | tr ',' '\n' \
    | sed '/^[[:space:]]*$/d' \
    | grep -B2 "\"name\": *\"${name}\"" \
    | grep '"id"' \
    | head -1 \
    | sed -E 's/[^0-9]*([0-9]+).*/\1/'
}

asset_id="$(find_asset_id "$asset_name")"
checksums_id="$(find_asset_id "$checksums_name")"
[ -n "$asset_id" ] || die "release ${tag} has no asset named ${asset_name} — check .goreleaser.yaml's name_template still matches"

# --- 4. download via the asset API (works for private repos with a token) --

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

download_asset() {
  id="$1"
  out="$2"
  if [ -n "$auth_header" ]; then
    curl -fsSL -H "$auth_header" -H "Accept: application/octet-stream" \
      "${api}/releases/assets/${id}" -o "$out"
  else
    # Public phase: same endpoint still works unauthenticated once the repo is public.
    curl -fsSL -H "Accept: application/octet-stream" \
      "${api}/releases/assets/${id}" -o "$out"
  fi
}

log "downloading ${asset_name} (release ${tag})"
download_asset "$asset_id" "${workdir}/${asset_name}"

if [ -n "$checksums_id" ]; then
  download_asset "$checksums_id" "${workdir}/${checksums_name}"
else
  log "warning: no checksums asset found — skipping verification"
fi

# --- 5. verify checksum ------------------------------------------------------

if [ -f "${workdir}/${checksums_name}" ]; then
  expected="$(grep "  ${asset_name}\$" "${workdir}/${checksums_name}" | awk '{print $1}')"
  [ -n "$expected" ] || die "asset ${asset_name} not listed in ${checksums_name}"

  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "${workdir}/${asset_name}" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "${workdir}/${asset_name}" | awk '{print $1}')"
  else
    die "no sha256sum or shasum on PATH — can't verify the download, refusing to install unverified binary"
  fi

  [ "$expected" = "$actual" ] || die "checksum mismatch for ${asset_name}: expected ${expected}, got ${actual}"
  log "checksum verified"
fi

# --- 6. unpack ---------------------------------------------------------------

case "$ext" in
  tar.gz) tar -xzf "${workdir}/${asset_name}" -C "$workdir" ;;
  zip) unzip -q "${workdir}/${asset_name}" -d "$workdir" ;;
esac

bin_path="${workdir}/${BINARY}"
[ "$os" = "windows" ] && bin_path="${bin_path}.exe"
[ -f "$bin_path" ] || die "unpacked archive but didn't find ${BINARY} inside — archive layout changed?"
chmod +x "$bin_path"

# --- 7. install: prefer /usr/local/bin, fall back to ~/.local/bin ----------

install_dir="/usr/local/bin"
if [ ! -w "$install_dir" ] 2>/dev/null; then
  install_dir="${HOME}/.local/bin"
  mkdir -p "$install_dir"
fi

cp "$bin_path" "${install_dir}/${BINARY}"
log "installed ${BINARY} ${version} -> ${install_dir}/${BINARY}"

# --- 8. macOS Gatekeeper note ------------------------------------------------

if [ "$os" = "darwin" ]; then
  xattr -d com.apple.quarantine "${install_dir}/${BINARY}" 2>/dev/null || true
  log "macOS: cleared the com.apple.quarantine flag on the binary."
  log "  (if Gatekeeper still balks, run: xattr -d com.apple.quarantine ${install_dir}/${BINARY})"
fi

# --- 9. PATH hint -------------------------------------------------------------

case ":$PATH:" in
  *":${install_dir}:"*) ;;
  *) log "note: ${install_dir} isn't on your PATH — add: export PATH=\"${install_dir}:\$PATH\"" ;;
esac

log "done — run '${BINARY} --help' to get started."

# --- other channels -----------------------------------------------------------
# `curl -fsSL sporo.dev/install.sh | sh` is the one-liner people run (this file is
# mirrored to that URL by the Pages deploy). A `brew install ydnikolaev/tap/sporo`
# path is wired and ready in .goreleaser.yaml — it turns on the moment the
# HOMEBREW_TAP_TOKEN secret exists (see docs/distribution.md).
