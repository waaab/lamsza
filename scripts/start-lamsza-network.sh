#!/usr/bin/env bash
# Start / stop / restart the local Lámsza network
# (admin, lamsza, szotar, jatszoter).
# Usage:
#   start-lamsza-network.sh [start|stop|restart|status]
#   lamsza-network […]   # if linked from ~/.local/bin
set -euo pipefail

# Do not use $HOME to locate anything. An agent runs this script with HOME
# pointing at a sandbox directory, so $HOME-derived paths send the logs and
# PID files somewhere the operator cannot find them (BOG-31).
#
# The script now lives in the lamsza repo, at scripts/start-lamsza-network.sh
# (BOG-50), and is reached through two symlinks into it:
#   ~/projects/start-lamsza-network.sh  and  ~/.local/bin/lamsza-network
# readlink -f resolves both to the repo copy, so the script's own directory is
# the repo's scripts/ folder, NOT the projects root. Walk up from there until we
# find the directory that holds all four app repos. That also works from a git
# worktree under lamsza/.worktrees/<branch>/scripts.
SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]}")"

find_projects_root() {
	local dir="$1" app ok
	while :; do
		ok=1
		for app in lamsza lamsza-admin lamsza-szotar lamsza-jatszoter; do
			[ -d "$dir/$app" ] || ok=0
		done
		[ "$ok" = 1 ] && { echo "$dir"; return 0; }
		[ "$dir" = "/" ] && return 1
		dir="$(dirname "$dir")"
	done
}

PROJECTS_ROOT="${LAMSZA_PROJECTS_ROOT:-$(find_projects_root "$(dirname "$SCRIPT_PATH")" || true)}"
if [ -z "$PROJECTS_ROOT" ]; then
	echo "$(basename "$0"): cannot find the projects root above $SCRIPT_PATH." >&2
	echo "Expected a directory containing lamsza, lamsza-admin, lamsza-szotar and lamsza-jatszoter." >&2
	echo "Set LAMSZA_PROJECTS_ROOT to point at it." >&2
	exit 1
fi

# The state belongs to the person who owns the checkout, not to whoever runs
# the script. Fall back to $HOME only if the owner lookup fails.
owner_home() {
	local owner home
	owner="$(stat -c '%U' "$SCRIPT_PATH" 2>/dev/null || true)"
	[ -n "$owner" ] || return 1
	home="$(getent passwd "$owner" 2>/dev/null | cut -d: -f6)"
	[ -n "$home" ] || return 1
	echo "$home"
}
STATE_DIR="${LAMSZA_STATE_DIR:-$(owner_home || echo "$HOME")/.cache/lamsza-network}"
mkdir -p "$STATE_DIR"

# name|repo_relpath|frontend_relpath|backend_port|frontend_port
#
# lamsza's frontend_relpath is "." on purpose: that repo keeps its SvelteKit
# frontend at the repo root and has no frontend/ folder. Accepted on BOG-23 —
# see lamsza/docs/network/WAYS_OF_WORKING.md §5. Do not "normalise" it.
APPS=(
	"admin|lamsza-admin|frontend|3000|5173"
	"lamsza|lamsza|.|3001|5174"
	"szotar|lamsza-szotar|frontend|3002|5175"
	"jatszoter|lamsza-jatszoter|frontend|3003|5176"
)

cmd="${1:-start}"

# True when something is listening on the given TCP port.
# The ss output is captured into a variable instead of piped: with pipefail
# set, grep -q exits on the first match and ss then dies of SIGPIPE, which
# would turn a found port into a failed command.
port_open() {
	local port="$1" listing
	listing="$(ss -ltn 2>/dev/null || true)"
	grep -qE ":${port}[[:space:]]" <<<"$listing"
}

kill_port() {
	local port="$1"
	if command -v fuser >/dev/null 2>&1; then
		fuser -k "${port}/tcp" 2>/dev/null || true
	else
		local pids
		pids=$(ss -ltnp 2>/dev/null | awk -v p=":$port" '$4 ~ p"$" {print}' | grep -oE 'pid=[0-9]+' | cut -d= -f2 | sort -u || true)
		for pid in $pids; do
			kill "$pid" 2>/dev/null || true
		done
	fi
}

app_root() {
	local rel="$1"
	echo "$PROJECTS_ROOT/$rel"
}

stop_all() {
	echo "Stopping Lámsza network frontends and backends…"
	local row name rel fe bport fport i ok
	for row in "${APPS[@]}"; do
		IFS='|' read -r name rel fe bport fport <<<"$row"
		kill_port "$fport"
		kill_port "$bport"
		rm -f "$STATE_DIR/${name}-backend.pid" "$STATE_DIR/${name}-frontend.pid"
	done
	pkill -f "$PROJECTS_ROOT/lamsza-admin/backend" 2>/dev/null || true
	pkill -f "$PROJECTS_ROOT/lamsza/backend" 2>/dev/null || true
	pkill -f "$PROJECTS_ROOT/lamsza-szotar/backend" 2>/dev/null || true
	pkill -f "$PROJECTS_ROOT/lamsza-jatszoter/backend" 2>/dev/null || true

	# Wait for the ports to actually close, not just for the signals to be
	# sent. start_apps now skips anything still listening (BOG-56), so a
	# shutdown that outlives this function would make "restart" quietly skip
	# the very app it just killed.
	for i in $(seq 1 "${LAMSZA_STOP_TIMEOUT:-10}"); do
		ok=1
		for row in "${APPS[@]}"; do
			IFS='|' read -r name rel fe bport fport <<<"$row"
			if port_open "$bport" || port_open "$fport"; then
				ok=0
			fi
		done
		[ "$ok" = 1 ] && break
		sleep 1
	done
	echo "Stopped app processes (databases left running)."
}

start_dbs() {
	echo "Starting databases…"
	local row name rel fe bport fport root
	for row in "${APPS[@]}"; do
		IFS='|' read -r name rel fe bport fport <<<"$row"
		# Admin shares main Lámsza Postgres — no own compose
		[ "$name" = "admin" ] && continue
		root="$(app_root "$rel")"
		if [ -f "$root/docker-compose.yml" ] || [ -f "$root/compose.yml" ]; then
			(cd "$root" && docker compose up -d)
		fi
	done
	sleep 2
}

# Never start a second copy of something that is already listening (BOG-56).
# The two halves fail differently, and the quiet one is the dangerous one:
# a duplicate backend dies on its own with "address already in use", but a
# duplicate Vite dev server does not — with strictPort unset it moves to the
# next free port and keeps running, and stop_all only kills the four fixed
# frontend ports, so the stray survives "stop" and has to be hunted by hand.
# Skipping also leaves the existing PID files alone instead of overwriting
# them with the PIDs of processes that are about to die.
start_apps() {
	echo "Starting backends and frontends…"
	local row name rel fe bport fport root blog flog bstate fstate
	for row in "${APPS[@]}"; do
		IFS='|' read -r name rel fe bport fport <<<"$row"
		root="$(app_root "$rel")"
		blog="$STATE_DIR/${name}-backend.log"
		flog="$STATE_DIR/${name}-frontend.log"

		if port_open "$bport"; then
			bstate="already running"
		else
			(
				cd "$root/backend"
				nohup go run . >"$blog" 2>&1 &
				echo $! >"$STATE_DIR/${name}-backend.pid"
			)
			bstate="started"
		fi

		if port_open "$fport"; then
			fstate="already running"
		else
			(
				cd "$root/$fe"
				nohup npm run dev >"$flog" 2>&1 &
				echo $! >"$STATE_DIR/${name}-frontend.pid"
			)
			fstate="started"
		fi

		printf '  %-10s backend :%s %-15s frontend :%s %s\n' \
			"$name" "$bport" "$bstate" "$fport" "$fstate"
	done
}

wait_ready() {
	echo "Waiting for ports…"
	local row name rel fe bport fport i ok
	# LAMSZA_READY_TIMEOUT exists so the test suite can drive "start" to
	# completion in a second instead of waiting out a real boot.
	for i in $(seq 1 "${LAMSZA_READY_TIMEOUT:-60}"); do
		ok=1
		for row in "${APPS[@]}"; do
			IFS='|' read -r name rel fe bport fport <<<"$row"
			port_open "$bport" || ok=0
			port_open "$fport" || ok=0
		done
		[ "$ok" = 1 ] && break
		sleep 1
	done
}

print_status() {
	local row name rel fe bport fport bstate fstate
	printf '%-10s  %-8s  %-8s  %s\n' "APP" "BACKEND" "FRONTEND" "URLS"
	for row in "${APPS[@]}"; do
		IFS='|' read -r name rel fe bport fport <<<"$row"
		bstate="down"
		fstate="down"
		port_open "$bport" && bstate=":$bport"
		port_open "$fport" && fstate=":$fport"
		case "$name" in
			admin) printf '%-10s  %-8s  %-8s  http://localhost:%s\n' "$name" "$bstate" "$fstate" "$fport" ;;
			lamsza) printf '%-10s  %-8s  %-8s  https://lamsza.test  http://localhost:%s\n' "$name" "$bstate" "$fstate" "$fport" ;;
			szotar) printf '%-10s  %-8s  %-8s  https://szotar.lamsza.test  http://localhost:%s\n' "$name" "$bstate" "$fstate" "$fport" ;;
			jatszoter) printf '%-10s  %-8s  %-8s  https://jatszoter.lamsza.test  http://localhost:%s\n' "$name" "$bstate" "$fstate" "$fport" ;;
		esac
	done
	echo "Apps: $PROJECTS_ROOT"
	echo "Logs: $STATE_DIR/"
}

case "$cmd" in
	start)
		start_dbs
		start_apps
		wait_ready
		print_status
		;;
	stop)
		stop_all
		print_status
		;;
	restart)
		stop_all
		start_dbs
		start_apps
		wait_ready
		print_status
		;;
	status)
		print_status
		;;
	*)
		echo "Usage: $0 [start|stop|restart|status]" >&2
		exit 2
		;;
esac
