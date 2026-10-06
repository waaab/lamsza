#!/usr/bin/env bash
#
# Tests the projects-root resolution in scripts/start-lamsza-network.sh.
#
# The script moved into this repo on BOG-50, so its own directory is no longer
# the projects root: it is <root>/lamsza/scripts. It now walks up looking for
# the directory that holds all four app repos. These cases prove that walk,
# including through the two symlinks the operator actually uses.
#
# Every case builds a throwaway tree of empty directories in a temp dir and
# only ever runs "status", which reads TCP state and prints — it never starts
# a backend, a frontend or a database.
#
#   scripts/tests/start-lamsza-network.test.sh

set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT="$HERE/../start-lamsza-network.sh"
[ -f "$SCRIPT" ] || { echo "missing: $SCRIPT" >&2; exit 2; }

PASS=0
FAIL=0
LAST_OUT=""

# A fake projects root: the four app folders plus a lamsza/scripts copy of the
# script, which is where the real one lives. Pass extra arguments to leave some
# of the app folders out.
make_tree() {
	local dir keep
	dir="$(mktemp -d)"
	for keep in "$@"; do
		mkdir -p "$dir/$keep"
	done
	mkdir -p "$dir/lamsza/scripts"
	cp "$SCRIPT" "$dir/lamsza/scripts/start-lamsza-network.sh"
	printf '%s' "$dir"
}

complete_tree() {
	make_tree lamsza lamsza-admin szotar jatszoter
}

# Run the script at $1 with the remaining arguments, in an environment with no
# LAMSZA_PROJECTS_ROOT and a throwaway state dir, and record the output.
run() {
	local expected="$1" label="$2" path="$3"
	shift 3
	local out rc state
	state="$(mktemp -d)"
	out="$(env -u LAMSZA_PROJECTS_ROOT LAMSZA_STATE_DIR="$state" "$path" "$@" 2>&1)"
	rc=$?
	LAST_OUT="$out"
	rm -rf "$state"
	if [ "$rc" = "$expected" ]; then
		PASS=$((PASS + 1))
		printf '  ok   %s (exit %s)\n' "$label" "$rc"
	else
		FAIL=$((FAIL + 1))
		printf '  FAIL %s (exit %s, wanted %s)\n%s\n' "$label" "$rc" "$expected" "$out"
	fi
}

# Assert the reported "Apps:" root. This is the whole point: a resolution that
# exits 0 but points at lamsza/scripts would start nothing.
expect_root() {
	local label="$1" want="$2"
	case "$LAST_OUT" in
		*"Apps: $want"$'\n'*|*"Apps: $want")
			PASS=$((PASS + 1)); printf '  ok   %s\n' "$label" ;;
		*)
			FAIL=$((FAIL + 1)); printf '  FAIL %s (wanted "Apps: %s")\n%s\n' "$label" "$want" "$LAST_OUT" ;;
	esac
}

# 1. Run directly from the repo copy: the root is two levels up, not scripts/.
T="$(complete_tree)"
run 0 "direct call from lamsza/scripts" "$T/lamsza/scripts/start-lamsza-network.sh" status
expect_root "resolves to the projects root, not lamsza/scripts" "$T"
rm -rf "$T"

# 2. Through ~/projects/start-lamsza-network.sh, the operator's path.
T="$(complete_tree)"
ln -s "$T/lamsza/scripts/start-lamsza-network.sh" "$T/start-lamsza-network.sh"
run 0 "through the projects-root symlink" "$T/start-lamsza-network.sh" status
expect_root "symlink resolves to the projects root" "$T"
rm -rf "$T"

# 3. Through a second hop, the ~/.local/bin/lamsza-network shape.
T="$(complete_tree)"
ln -s "$T/lamsza/scripts/start-lamsza-network.sh" "$T/start-lamsza-network.sh"
mkdir -p "$T/bin"
ln -s "$T/start-lamsza-network.sh" "$T/bin/lamsza-network"
run 0 "through a symlink to a symlink" "$T/bin/lamsza-network" status
expect_root "two symlink hops resolve to the projects root" "$T"
rm -rf "$T"

# 4. From a git worktree under lamsza/.worktrees/<branch>/scripts.
T="$(complete_tree)"
mkdir -p "$T/lamsza/.worktrees/some-branch/scripts"
cp "$SCRIPT" "$T/lamsza/.worktrees/some-branch/scripts/start-lamsza-network.sh"
run 0 "from a worktree copy" "$T/lamsza/.worktrees/some-branch/scripts/start-lamsza-network.sh" status
expect_root "a worktree copy still finds the projects root" "$T"
rm -rf "$T"

# 5. An incomplete tree is refused loudly instead of guessing a root.
T="$(make_tree lamsza szotar)"
run 1 "tree missing lamsza-admin and jatszoter" "$T/lamsza/scripts/start-lamsza-network.sh" status
case "$LAST_OUT" in
	*LAMSZA_PROJECTS_ROOT*) PASS=$((PASS + 1)); printf '  ok   the error names LAMSZA_PROJECTS_ROOT\n' ;;
	*) FAIL=$((FAIL + 1)); printf '  FAIL the error does not name LAMSZA_PROJECTS_ROOT\n%s\n' "$LAST_OUT" ;;
esac
rm -rf "$T"

# 6. LAMSZA_PROJECTS_ROOT still wins, and skips the walk entirely.
T="$(make_tree lamsza szotar)"
STATE="$(mktemp -d)"
LAST_OUT="$(LAMSZA_PROJECTS_ROOT=/somewhere/else LAMSZA_STATE_DIR="$STATE" \
	"$T/lamsza/scripts/start-lamsza-network.sh" status 2>&1)"
if [ $? = 0 ]; then
	PASS=$((PASS + 1)); printf '  ok   LAMSZA_PROJECTS_ROOT overrides an incomplete tree (exit 0)\n'
else
	FAIL=$((FAIL + 1)); printf '  FAIL LAMSZA_PROJECTS_ROOT did not override\n%s\n' "$LAST_OUT"
fi
expect_root "the override is the reported root" "/somewhere/else"
rm -rf "$T" "$STATE"

# 7. An unknown command is still a usage error, not a resolution error.
T="$(complete_tree)"
run 2 "unknown command" "$T/lamsza/scripts/start-lamsza-network.sh" wobble
rm -rf "$T"

# ---------------------------------------------------------------------------
# "start" is idempotent (BOG-56).
#
# The cases below really do run "start" and "restart", which the cases above
# deliberately avoid. They are safe because they run against a PATH shim:
#
#   ss     reports a port as listening for every file in $LAMSZA_TEST_PORTS
#   fuser  "frees" a port by deleting its file, optionally after
#          $LAMSZA_TEST_KILL_DELAY seconds to model a slow shutdown — one
#          file per port, so the eight concurrent kills cannot race
#   go/npm record where they were invoked and exit, starting nothing
#   pkill  does nothing
#
# So no real port is read and no real process is signalled: these cases are
# safe to run while the operator's own network is up on 3000-3003/5173-5176.
# ---------------------------------------------------------------------------

make_stubs() {
	local bin="$1"
	mkdir -p "$bin"

cat >"$bin/ss" <<'STUB'
#!/usr/bin/env bash
for f in "$LAMSZA_TEST_PORTS"/*; do
	[ -e "$f" ] || continue
	printf 'LISTEN 0 4096 127.0.0.1:%s 0.0.0.0:*\n' "$(basename "$f")"
done
exit 0
STUB

cat >"$bin/fuser" <<'STUB'
#!/usr/bin/env bash
delay="${LAMSZA_TEST_KILL_DELAY:-0}"
for arg in "$@"; do
	case "$arg" in
		*/tcp)
			port="${arg%/tcp}"
			if [ "$delay" = 0 ]; then
				rm -f "$LAMSZA_TEST_PORTS/$port"
			else
				( sleep "$delay"; rm -f "$LAMSZA_TEST_PORTS/$port" ) >/dev/null 2>&1 &
			fi
			;;
	esac
done
exit 0
STUB

cat >"$bin/go" <<'STUB'
#!/usr/bin/env bash
echo "go $* <- $PWD" >>"$LAMSZA_TEST_LAUNCHES"
STUB

cat >"$bin/npm" <<'STUB'
#!/usr/bin/env bash
echo "npm $* <- $PWD" >>"$LAMSZA_TEST_LAUNCHES"
STUB

cat >"$bin/pkill" <<'STUB'
#!/usr/bin/env bash
exit 0
STUB

	chmod +x "$bin"/*
}

# A complete tree plus the backend/ and frontend/ folders start_apps cds into,
# the stub PATH, and the two scratch files the stubs read and write.
live_tree() {
	LIVE_ROOT="$(complete_tree)"
	mkdir -p "$LIVE_ROOT"/lamsza-admin/{backend,frontend} \
		"$LIVE_ROOT"/lamsza/backend \
		"$LIVE_ROOT"/szotar/{backend,frontend} \
		"$LIVE_ROOT"/jatszoter/{backend,frontend}
	LIVE_BIN="$LIVE_ROOT/stub-bin"
	make_stubs "$LIVE_BIN"
	LIVE_STATE="$(mktemp -d)"
	LIVE_PORTS="$LIVE_ROOT/ports"
	LIVE_LAUNCHES="$LIVE_ROOT/launches"
	mkdir -p "$LIVE_PORTS"
	: >"$LIVE_LAUNCHES"
}

live_cleanup() {
	rm -rf "$LIVE_ROOT" "$LIVE_STATE"
}

# Declare every port of a fully-running network open, in ss -ltn's shape.
all_ports_open() {
	local port
	for port in 3000 3001 3002 3003 5173 5174 5175 5176; do
		: >"$LIVE_PORTS/$port"
	done
}

run_live() {
	local expected="$1" label="$2"
	shift 2
	local out rc
	out="$(env -u LAMSZA_PROJECTS_ROOT \
		LAMSZA_STATE_DIR="$LIVE_STATE" \
		LAMSZA_READY_TIMEOUT=1 \
		LAMSZA_TEST_PORTS="$LIVE_PORTS" \
		LAMSZA_TEST_LAUNCHES="$LIVE_LAUNCHES" \
		PATH="$LIVE_BIN:$PATH" \
		"$LIVE_ROOT/lamsza/scripts/start-lamsza-network.sh" "$@" 2>&1)"
	rc=$?
	LAST_OUT="$out"
	if [ "$rc" = "$expected" ]; then
		PASS=$((PASS + 1)); printf '  ok   %s (exit %s)\n' "$label" "$rc"
	else
		FAIL=$((FAIL + 1)); printf '  FAIL %s (exit %s, wanted %s)\n%s\n' "$label" "$rc" "$expected" "$out"
	fi
}

expect_out() {
	local label="$1" want="$2"
	case "$LAST_OUT" in
		*"$want"*) PASS=$((PASS + 1)); printf '  ok   %s\n' "$label" ;;
		*) FAIL=$((FAIL + 1)); printf '  FAIL %s (wanted "%s" in the output)\n%s\n' "$label" "$want" "$LAST_OUT" ;;
	esac
}

# How many go/npm processes start_apps actually launched.
expect_launches() {
	local label="$1" want="$2" got
	got="$(grep -c . "$LIVE_LAUNCHES" || true)"
	if [ "$got" = "$want" ]; then
		PASS=$((PASS + 1)); printf '  ok   %s (%s launched)\n' "$label" "$got"
	else
		FAIL=$((FAIL + 1)); printf '  FAIL %s (%s launched, wanted %s)\n%s\n' "$label" "$got" "$want" "$(cat "$LIVE_LAUNCHES")"
	fi
}

# 8. start against a fully-running network starts nothing and says so.
live_tree
all_ports_open
echo "sentinel" >"$LIVE_STATE/lamsza-frontend.pid"
run_live 0 "start against a live network" start
expect_launches "start launches nothing when every port is taken" 0
expect_out "it reports the skipped backend" "backend :3001 already running"
expect_out "it reports the skipped frontend" "frontend :5174 already running"
# Nothing was launched, so nothing could have slid onto 5177+, and the PID
# file still points at the process that is actually running.
if [ "$(cat "$LIVE_STATE/lamsza-frontend.pid")" = "sentinel" ]; then
	PASS=$((PASS + 1)); printf '  ok   the existing PID file is left alone\n'
else
	FAIL=$((FAIL + 1)); printf '  FAIL the PID file was overwritten: %s\n' "$(cat "$LIVE_STATE/lamsza-frontend.pid")"
fi
live_cleanup

# 9. The guard is a skip, not a refusal: a cold machine still starts all eight.
live_tree
run_live 0 "start on a cold machine" start
expect_launches "start launches all eight processes" 8
expect_out "it reports what it started" "backend :3001 started"
live_cleanup

# 10. restart is unaffected — stop_all frees the ports before start_apps looks.
live_tree
all_ports_open
run_live 0 "restart against a live network" restart
expect_launches "restart still launches all eight processes" 8
live_cleanup

# 11. A shutdown slower than stop_all's old fixed sleep still restarts: the
#     skip guard must not be fooled by a port that has not closed *yet*.
live_tree
all_ports_open
export LAMSZA_TEST_KILL_DELAY=3
run_live 0 "restart with a slow shutdown" restart
unset LAMSZA_TEST_KILL_DELAY
expect_launches "restart waits for the ports to close, then starts all eight" 8
live_cleanup

printf '\n%s passed, %s failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
