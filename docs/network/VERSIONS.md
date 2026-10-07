# Runtime versions

The one place that records which Ubuntu, Node and Go the Lámsza network runs on,
in production, in CI and on the dev machine. Workflows and docs point here; do
not copy these numbers anywhere else.

Last checked: 2026-10-07.

| | Production droplet | CI (GitHub Actions) | Dev machine | Target |
|---|---|---|---|---|
| Ubuntu | 24.04 LTS | `ubuntu-24.04` (pinned in every workflow) | n/a | 24.04 LTS |
| Node | **20.20.2 (end of life)** | from `.nvmrc`: **24** | 24.21 | 24 LTS |
| Go | 1.25.7 | from `backend/go.mod`: **1.25.7** | 1.27.1 | 1.25.x |

Production is the owner's; agents never change it. Upgrading production Node to
24 LTS is an open item (`OPEN_ITEMS.md`).

## Where each version comes from

- **Ubuntu:** `runs-on: ubuntu-24.04` in each repo's workflow. Never
  `ubuntu-latest`: that moves to a new Ubuntu on GitHub's schedule, and CI would
  stop matching the droplet without anybody deciding it.
- **Node:** `.nvmrc` at each repo root; `actions/setup-node` reads it with
  `node-version-file: .nvmrc`.
- **Go:** the `go` line of `backend/go.mod`; `actions/setup-go` reads it with
  `go-version-file: backend/go.mod`. A newer local Go (1.27 here) builds a
  1.25.7 module fine; `GOTOOLCHAIN=go1.25.7` runs the exact production version.

No workflow names a Node or Go version itself.

## Verified 2026-10-07

| App | Node 20.20.2 | Node 24.21 | Go 1.25.7 |
|---|---|---|---|
| lamsza | build ok; tests pass (198/198) when given the file list | build ok, `npm test` 198/198 | vet, build and tests ok |
| admin | build ok; tests 97/97 with the file list | build ok, `npm test` 97/97 | vet, build and tests ok |
| szotar | build ok; tests 25/25 with the file list | build ok, `npm test` 25/25 | vet, build and tests ok |
| jatszoter | build ok; tests 11/11 with the file list | build ok, `npm test` 11/11 | vet, build and tests ok |

On Node 20 `npm test` itself fails in every repo: the scripts pass a quoted glob
(`node --test "tests/*.test.js"`), and Node 20's test runner does not expand
globs (Node 22 and later do). The code runs on 20; the test command needs 22+.
That is one more reason the target is Node 24 everywhere.

## When the droplet is upgraded

When the production droplet gets a new Ubuntu, Node or Go, update all of these
together, in the same change, in all four repos:

1. `runs-on:` in every workflow job,
2. `.nvmrc`,
3. the `go` line of `backend/go.mod`,
4. the table above.

Then run each repo's checks (`docs/LOCAL_DEV_CHECKS.md`) on the new versions
before pushing.
