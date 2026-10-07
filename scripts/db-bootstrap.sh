#!/usr/bin/env bash
# Bring an empty Postgres up to the current lamsza schema in one command.
#
# The 41 files in backend/migrations/ are a historical record of hand-run
# changes, not a runnable sequence - there is no recorded order and several of
# them only ever applied to a schema that has since moved on. So there was no
# way to create a working lamsza database from this repo at all, which is why
# the root `backend` test package could not run in CI (BOG-53).
#
# backend/schema/ is that missing bootstrap: the whole current schema plus the
# reference rows the app needs to function, both generated from the live dev
# database by scripts/db-dump-schema.sh and committed. This script applies them.
#
#   scripts/db-bootstrap.sh                      # the local dev database
#   scripts/db-bootstrap.sh --create             # create it first if missing
#   scripts/db-bootstrap.sh --url "$DATABASE_URL"
#
# It is safe to re-run: the schema is CREATE-only, so a second pass on a
# populated database fails loudly rather than touching data. Point it at an
# empty database.

set -euo pipefail

here=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
repo=$(cd -- "$here/.." && pwd)
# shellcheck source=lib/pg-url.sh
source "$here/lib/pg-url.sh"

DEFAULT_URL="postgres://lamsza_user:lamsza_password@localhost:5433/lamsza?sslmode=disable"

url=${DATABASE_URL:-$DEFAULT_URL}
container=""
create=0

while [[ $# -gt 0 ]]; do
	case $1 in
	--url)
		url=$2
		shift 2
		;;
	--container)
		container=$2
		shift 2
		;;
	--create)
		create=1
		shift
		;;
	-h | --help)
		sed -n '2,22p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
		exit 0
		;;
	*)
		echo "unknown argument: $1" >&2
		exit 2
		;;
	esac
done

pg_url_parse "$url"
runner=$(pg_pick_runner "$container" psql)
target=$PGDATABASE

echo "bootstrap   $PGUSER@$PGHOST:$PGPORT/$target  (runner: $runner)"

if [[ $create -eq 1 ]]; then
	# Connect to the always-present `postgres` database to ask about, and
	# create, the target.
	PGDATABASE=postgres
	exists=$(pg_run "$runner" psql -qtAc \
		"SELECT 1 FROM pg_database WHERE datname = '$target'")
	if [[ $exists == 1 ]]; then
		echo "            database already exists, not creating"
	else
		pg_run "$runner" psql -qc "CREATE DATABASE \"$target\""
		echo "            created database $target"
	fi
	PGDATABASE=$target
fi

for file in "$repo"/backend/schema/[0-9]*.sql; do
	echo "            applying ${file#"$repo"/}"
	# -o /dev/null drops the result rows pg_dump's own set_config/setval calls
	# print. Errors go to stderr, so nothing that matters is hidden.
	pg_run "$runner" psql -v ON_ERROR_STOP=1 -q -o /dev/null <"$file"
done

tables=$(pg_run "$runner" psql -qtAc \
	"SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public'")
settlements=$(pg_run "$runner" psql -qtAc "SELECT count(*) FROM settlements")

echo "done        $tables relations in public, $settlements settlements"

# The suites that need this database skip, rather than fail, when the reference
# rows are missing - so an empty settlements table would quietly turn CI green
# on tests that never ran. Refuse that.
if [[ ${settlements:-0} -eq 0 ]]; then
	echo "settlements is empty: backend/schema/002_reference.sql did not load" >&2
	exit 1
fi
