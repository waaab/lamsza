#!/usr/bin/env bash
# Shared plumbing for the two schema scripts: parse a postgres:// URL into PG*
# environment variables, and run a client binary either directly or inside the
# Postgres container.
#
# The container path exists because this machine has no postgresql-client
# installed - `pg_dump` and `psql` only exist inside `lamsza-db`. CI is the
# other way round: the GitHub runner has the client and the database is a
# service container reachable over TCP. One script has to work in both.

# pg_url_parse <url> — exports PGUSER, PGPASSWORD, PGHOST, PGPORT, PGDATABASE.
pg_url_parse() {
	local url=$1 rest creds hostpart

	case $url in
	postgres://* | postgresql://*) ;;
	*)
		echo "not a postgres URL: $url" >&2
		return 2
		;;
	esac

	rest=${url#*://}
	rest=${rest%%\?*} # drop ?sslmode=... and friends

	if [[ $rest == *@* ]]; then
		creds=${rest%%@*}
		hostpart=${rest#*@}
		export PGUSER=${creds%%:*}
		if [[ $creds == *:* ]]; then
			export PGPASSWORD=${creds#*:}
		fi
	else
		hostpart=$rest
	fi

	export PGDATABASE=${hostpart#*/}
	hostpart=${hostpart%%/*}
	export PGHOST=${hostpart%%:*}
	if [[ $hostpart == *:* ]]; then
		export PGPORT=${hostpart#*:}
	else
		export PGPORT=5432
	fi

	if [[ -z ${PGUSER:-} || -z ${PGDATABASE:-} ]]; then
		echo "could not read user and database out of: $url" >&2
		return 2
	fi
}

# pg_pick_runner <container> <needed-binary>... — echoes "direct" or
# "docker:<container>", and fails if neither is usable. An explicit container
# wins; otherwise the binaries this caller needs being on PATH wins; otherwise
# fall back to the default container if it is running.
#
# The needed binaries are per-caller on purpose: db-bootstrap.sh only wants
# psql, and the GitHub runner has the client package, so it must not be pushed
# onto the docker path just because some other script wants pg_dump.
pg_pick_runner() {
	local forced=${1:-}
	shift || true
	local bin have=1

	if [[ -n $forced ]]; then
		echo "docker:$forced"
		return 0
	fi
	for bin in "$@"; do
		command -v "$bin" >/dev/null 2>&1 || have=0
	done
	if [[ $# -gt 0 && $have -eq 1 ]]; then
		echo direct
		return 0
	fi
	if command -v docker >/dev/null 2>&1 &&
		docker ps --format '{{.Names}}' 2>/dev/null | grep -qx "${PG_DEFAULT_CONTAINER:-lamsza-db}"; then
		echo "docker:${PG_DEFAULT_CONTAINER:-lamsza-db}"
		return 0
	fi

	echo "no way to reach Postgres: $* ${*:+is/are} not on PATH and the" >&2
	echo "${PG_DEFAULT_CONTAINER:-lamsza-db} container is not running." >&2
	echo "Install postgresql-client, start the container, or pass --container." >&2
	return 2
}

# pg_run <runner> <binary> [args...] — stdin and stdout are passed through.
#
# In docker mode the connection is rewritten to the in-container socket:
# PGHOST/PGPORT describe how the *host* reaches the database, and inside the
# container that same database is localhost:5432.
pg_run() {
	local runner=$1 bin=$2
	shift 2

	if [[ $runner == direct ]]; then
		"$bin" "$@"
		return
	fi

	docker exec -i \
		-e PGUSER="$PGUSER" \
		-e PGPASSWORD="${PGPASSWORD:-}" \
		-e PGDATABASE="$PGDATABASE" \
		-e PGHOST=localhost \
		-e PGPORT=5432 \
		"${runner#docker:}" "$bin" "$@"
}
