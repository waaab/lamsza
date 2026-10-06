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

printf '\n%s passed, %s failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
