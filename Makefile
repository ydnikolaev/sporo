# Makefile — the make-ABI home (doctrine: .mate/doctrine/code/interface.md §1). Each
# target is the canonical verb a synced skill trusts (`make check`, `make gen`, …); the
# project's gates live here and are called from `check`, and CI runs `make check` so the
# local ceiling and CI are the same gate.
#
# PUBLIC (contributors get `make test/lint/check`). The one private bit — the classify
# boundary gate — degrades gracefully in a public checkout (its script is private harness);
# the public CI backstop covers CI regardless.
#
# GOWORK=off is a gate everywhere: the module must build without a workspace, or a stray
# go.work leaks replacements a consumer's `go install` would never see.
export GOWORK := off

.PHONY: check test lint fmt fmt-check tidy-check vulncheck coverage gen build clean \
        workflow-lint mate-check classify classify-teeth harness-sync

## check — the COMPLETE quality gate; the ceiling CI and skills trust before "done".
## Not a middle rung: for a faster loop add check-fast BELOW this, never a fuller gate above.
check: fmt-check tidy-check lint workflow-lint
	go build ./...
	$(MAKE) coverage
	go run ./cmd/sporo lint
	go generate ./...
	git diff --exit-code -- web/src/data/surface.json
	$(MAKE) vulncheck
	$(MAKE) classify
	$(MAKE) classify-teeth

## test — tests only (no lint/coverage gate).
test:
	go test ./...

## coverage — run tests with a profile, then enforce the coveragepolicy SSOT
## (per-package + global floor). The checker lives in internal/coveragepolicy so CI can run it.
coverage:
	go test -covermode=set -coverprofile=coverage.out ./...
	go run internal/coveragepolicy/covercheck.go coverage.out

## lint — golangci-lint (the full linter suite; subsumes go vet and gosec).
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint missing: https://golangci-lint.run/welcome/install/"; exit 1; }
	golangci-lint run ./...

## fmt — format in place with gofumpt (idempotent).
fmt:
	@command -v gofumpt >/dev/null 2>&1 || { echo "gofumpt missing: go install mvdan.cc/gofumpt@latest"; exit 1; }
	gofumpt -w .

## fmt-check — fail if any Go file is not gofumpt-clean (CI-safe, no writes).
fmt-check:
	@command -v gofumpt >/dev/null 2>&1 || { echo "gofumpt missing: go install mvdan.cc/gofumpt@latest"; exit 1; }
	@out=$$(gofumpt -l .); if [ -n "$$out" ]; then echo "not gofumpt-clean:"; echo "$$out"; echo "run: make fmt"; exit 1; fi
	@echo "fmt: clean"

## tidy-check — go.mod / go.sum are tidy (no missing or superfluous requires).
tidy-check:
	go mod tidy
	@git diff --exit-code go.mod go.sum || { echo "go.mod/go.sum changed — run 'go mod tidy' and commit"; exit 1; }

## vulncheck — official Go vulnerability scanner, gated by .govulncheck-allow.txt: fails on
## any CALLED vuln not on the accepted list, so a NEW vuln is loud but an unfixable one
## doesn't hold the gate red forever. Part of `check` (needs the network), and also runs on
## a daily CI schedule + on PR, since disclosures appear over time, not only per-commit.
vulncheck:
	@command -v govulncheck >/dev/null 2>&1 || { echo "govulncheck missing: go install golang.org/x/vuln/cmd/govulncheck@latest"; exit 1; }
	@out=$$(govulncheck ./... 2>&1) || true; \
	found=$$(printf '%s\n' "$$out" | grep -oE 'GO-[0-9]{4}-[0-9]+' | sort -u); \
	new=""; for id in $$found; do grep -qxF "$$id" .govulncheck-allow.txt 2>/dev/null || new="$$new $$id"; done; \
	if [ -n "$$new" ]; then printf '%s\n' "$$out"; echo; echo "❌ NEW vulnerabilities (not in .govulncheck-allow.txt):$$new"; exit 1; fi; \
	if [ -n "$$found" ]; then echo "vulncheck: OK — only accepted vulns present:$$(printf '%s' "$$found" | tr '\n' ' ' | sed 's/^/ /')"; else echo "vulncheck: OK — no called vulnerabilities"; fi

## workflow-lint — every GitHub Action must be SHA-pinned (defeats tag-hijack; Dependabot
## still updates the pins) and the workflows must be syntactically valid. Runs in `check`.
workflow-lint:
	@bad=$$(grep -rnE 'uses: +[^ ]+@' .github/workflows 2>/dev/null | grep -vE '@[0-9a-f]{40}([ "#]|$$)' | grep -v 'uses: \./' || true); \
	if [ -n "$$bad" ]; then echo "❌ unpinned action(s) — pin to a full commit SHA (# vX.Y.Z):"; echo "$$bad"; exit 1; fi; \
	echo "workflow-lint: all actions SHA-pinned"
	@command -v actionlint >/dev/null 2>&1 && actionlint || echo "workflow-lint: actionlint not installed locally (CI runs it) — go install github.com/rhysd/actionlint/cmd/actionlint@latest"

## gen — regenerate all derived artifacts (the surface snapshot).
gen:
	go generate ./...

## build — production build.
build:
	go build -o bin/sporo ./cmd/sporo

## clean — sweep controllable ephemera.
clean:
	rm -rf bin dist coverage.out

## classify — the publish-boundary gate: no private path is tracked in the public repo, and
## every top-level entry is classified. The gate script is private harness (absent from a
## public checkout), so this degrades gracefully — the CI backstop covers CI regardless.
classify:
	@if [ -f scripts/classify-guard.sh ]; then bash scripts/classify-guard.sh; \
	else echo "classify: private gate absent (public checkout) — CI backstop covers it"; fi

## classify-teeth — proves the classify gate actually goes red (suite health).
classify-teeth:
	@if [ -f scripts/classify-guard.teeth.sh ]; then bash scripts/classify-guard.teeth.sh; \
	else echo "classify-teeth: private (public checkout) — skipped"; fi

## harness-sync — mirror project-authored strays into the private docs repo
## (docs/_harness/) for history + backup. Run before committing inside docs/.
harness-sync:
	bash scripts/harness-sync.sh

## mate-check — the synced-harness lane: managed-file drift. mate owns this lane; delegate to
## the binary if present, else state the absence honestly. Suite health (teeth) rides here.
mate-check: classify-teeth
	@if command -v mate >/dev/null 2>&1; then mate status; \
	else echo "mate-check: drift lane is mate's — 'mate' not on PATH, not wired here yet"; fi
