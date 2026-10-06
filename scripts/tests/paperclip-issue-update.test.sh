#!/usr/bin/env bash
#
# Tests the uncommitted-work gate in scripts/paperclip-issue-update.sh.
#
# Every case runs with --dry-run, so nothing is ever written to Paperclip.
# Each case builds a throwaway git repo in a temp dir, so the real workspace
# is never touched.
#
#   scripts/tests/paperclip-issue-update.test.sh

set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT="$HERE/../paperclip-issue-update.sh"
[ -f "$SCRIPT" ] || { echo "missing: $SCRIPT" >&2; exit 2; }

PASS=0
FAIL=0

# A temp repo with one commit, one .gitignore entry, and a clean tree.
make_repo() {
  local dir
  dir="$(mktemp -d)"
  git -C "$dir" init -q -b main
  git -C "$dir" config user.email test@example.com
  git -C "$dir" config user.name "Test"
  printf 'ignored.log\n' > "$dir/.gitignore"
  printf 'hello\n' > "$dir/tracked.txt"
  git -C "$dir" add -A
  git -C "$dir" commit -qm "initial"
  printf '%s' "$dir"
}

# run <expected-exit> <name> <repo> <args...>
run() {
  local want="$1" name="$2" repo="$3"; shift 3
  local out code
  out="$(cd "$repo" && PAPERCLIP_TASK_ID=TEST-1 "$SCRIPT" "$@" --dry-run 2>&1)"
  code=$?
  if [ "$code" -eq "$want" ]; then
    PASS=$((PASS + 1))
    printf '  ok   %s (exit %s)\n' "$name" "$code"
  else
    FAIL=$((FAIL + 1))
    printf '  FAIL %s — wanted exit %s, got %s\n' "$name" "$want" "$code"
    printf '%s\n' "$out" | sed 's/^/       | /'
  fi
  LAST_OUT="$out"
}

echo "paperclip-issue-update gate"

# 1. Clean tree + done -> accepted.
R="$(make_repo)"
run 0 "clean tree + done" "$R" done
rm -rf "$R"

# 2. Dirty tree + done -> refused, and the dirty path is named.
R="$(make_repo)"
printf 'changed\n' >> "$R/tracked.txt"
run 1 "dirty tree + done" "$R" done
case "$LAST_OUT" in
  *tracked.txt*) PASS=$((PASS + 1)); printf '  ok   refusal names the dirty path\n' ;;
  *) FAIL=$((FAIL + 1)); printf '  FAIL refusal does not name the dirty path\n' ;;
esac
rm -rf "$R"

# 3. Dirty tree + done --allow-dirty -> accepted.
R="$(make_repo)"
printf 'changed\n' >> "$R/tracked.txt"
run 0 "dirty tree + done --allow-dirty" "$R" done --allow-dirty
rm -rf "$R"

# 4. Dirty tree + any other status -> accepted, the gate does not fire.
for s in in_progress in_review blocked cancelled todo backlog; do
  R="$(make_repo)"
  printf 'changed\n' >> "$R/tracked.txt"
  run 0 "dirty tree + $s" "$R" "$s"
  rm -rf "$R"
done

# 5. An untracked file is dirty too — this is the BOG-3..BOG-7 shape.
R="$(make_repo)"
printf 'new\n' > "$R/newfile.txt"
run 1 "untracked file + done" "$R" done
rm -rf "$R"

# 6. Only .gitignore'd files dirty -> accepted.
R="$(make_repo)"
printf 'noise\n' > "$R/ignored.log"
run 0 "only ignored files dirty + done" "$R" done
rm -rf "$R"

# 7. A bad status is rejected before anything else happens.
R="$(make_repo)"
run 2 "unknown status" "$R" finished
rm -rf "$R"

# 8. No status at all is rejected.
R="$(make_repo)"
run 2 "no status" "$R"
rm -rf "$R"

printf '\n%s passed, %s failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
