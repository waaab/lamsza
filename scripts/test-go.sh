#!/usr/bin/env bash
# Run the Go backend suites against a fresh scratch database, never the dev one.
#
# The root `backend` package inserts and deletes rows (users, entries, tags,
# site_settings). Run against the dev database, it left dozens of
# *@test.lamsza users behind. So the suites read only TEST_DATABASE_URL, and a
# local test database name must end in _test (internal/db/testguard.go).
#
#   scripts/test-go.sh                       # every package
#   scripts/test-go.sh ./internal/...        # just these packages
#   scripts/test-go.sh -run TestGetEntries . # extra go test arguments
#
# Each run drops and recreates the scratch database, then applies
# backend/schema/ with scripts/db-bootstrap.sh, the same path CI takes.
# Override the target with TEST_DATABASE_URL (its name must still end in _test).

set -euo pipefail

here=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
repo=$(cd -- "$here/.." && pwd)
# shellcheck source=lib/pg-url.sh
source "$here/lib/pg-url.sh"

url=${TEST_DATABASE_URL:-"postgres://lamsza_user:lamsza_password@localhost:5433/lamsza_test?sslmode=disable"}

pg_url_parse "$url"
target=$PGDATABASE
if [[ $target != *_test ]]; then
	echo "refusing: test database name must end in _test, got '$target'" >&2
	exit 2
fi

runner=$(pg_pick_runner "" psql)
PGDATABASE=postgres
pg_run "$runner" psql -qc "DROP DATABASE IF EXISTS \"$target\" WITH (FORCE)"
PGDATABASE=$target

"$here/db-bootstrap.sh" --create --url "$url"

cd "$repo/backend"
if [[ $# -eq 0 ]]; then
	set -- ./...
fi
TEST_DATABASE_URL=$url go test -count=1 "$@"
