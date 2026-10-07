#!/usr/bin/env bash
#
# Tests the three "done" gates in scripts/paperclip-issue-update.sh:
# committed (BOG-33), merged into the base, and pushed to origin (BOG-38).
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

# Commit one file on a new branch off main and stay on that branch. The branch
# is deliberately left unmerged - this is the BOG-17 / BOG-28 shape.
branch_with_commit() {
  local dir="$1" branch="$2"
  git -C "$dir" checkout -q -b "$branch"
  printf 'the fix\n' > "$dir/fix.txt"
  git -C "$dir" add -A
  git -C "$dir" commit -qm "the fix"
}

# Give the repo an origin whose main matches local main, so the pushed gate
# is satisfied and only the gate under test can fire.
add_pushed_origin() {
  local dir="$1" remote
  remote="$(mktemp -d)"
  git -C "$remote" init -q --bare -b main
  git -C "$dir" remote add origin "$remote"
  git -C "$dir" push -q origin main
  printf '%s' "$remote"
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
    printf '  FAIL %s - wanted exit %s, got %s\n' "$name" "$want" "$code"
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

# 5. An untracked file is dirty too - this is the BOG-3..BOG-7 shape.
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

# --- Gate 2: merged into the base ------------------------------------------

# 9. Clean tree, committed, on an unmerged branch + done -> refused.
#    This is the BOG-17 / BOG-28 shape: gate 1 passes, nothing shipped.
R="$(make_repo)"; REM="$(add_pushed_origin "$R")"
branch_with_commit "$R" bog-99-unmerged
run 1 "committed on an unmerged branch + done" "$R" done
case "$LAST_OUT" in
  *'not on `main`'*) PASS=$((PASS + 1)); printf '  ok   refusal names the base branch\n' ;;
  *) FAIL=$((FAIL + 1)); printf '  FAIL refusal does not name the base branch\n' ;;
esac
case "$LAST_OUT" in
  *'the fix'*) PASS=$((PASS + 1)); printf '  ok   refusal lists the branch-only commit\n' ;;
  *) FAIL=$((FAIL + 1)); printf '  FAIL refusal does not list the branch-only commit\n' ;;
esac
rm -rf "$R" "$REM"

# 10. Same branch + --allow-unmerged -> accepted.
R="$(make_repo)"; REM="$(add_pushed_origin "$R")"
branch_with_commit "$R" bog-99-unmerged
run 0 "unmerged branch + done --allow-unmerged" "$R" done --allow-unmerged
rm -rf "$R" "$REM"

# 11. Merge the branch into main, push, then done -> accepted.
R="$(make_repo)"; REM="$(add_pushed_origin "$R")"
branch_with_commit "$R" bog-99-merged
git -C "$R" checkout -q main
git -C "$R" merge -q --no-ff -m "merge the fix" bog-99-merged
git -C "$R" push -q origin main
run 0 "merged into main and pushed + done" "$R" done
rm -rf "$R" "$REM"

# 12. The merged branch is still checked out after the merge -> accepted,
#     because HEAD is now an ancestor of main.
R="$(make_repo)"; REM="$(add_pushed_origin "$R")"
branch_with_commit "$R" bog-99-merged
git -C "$R" checkout -q main
git -C "$R" merge -q --no-ff -m "merge the fix" bog-99-merged
git -C "$R" push -q origin main
git -C "$R" checkout -q bog-99-merged
run 0 "merged branch still checked out + done" "$R" done
rm -rf "$R" "$REM"

# 13. An unmerged branch with any other status -> accepted, gate 2 is done-only.
for s in in_progress in_review blocked; do
  R="$(make_repo)"; REM="$(add_pushed_origin "$R")"
  branch_with_commit "$R" bog-99-unmerged
  run 0 "unmerged branch + $s" "$R" "$s"
  rm -rf "$R" "$REM"
done

# 14. --base names a different base branch. The branch lands on `release` and
#     never on `main`, so the verdict has to follow --base.
R="$(make_repo)"; REM="$(add_pushed_origin "$R")"
git -C "$R" branch release
git -C "$R" checkout -q release
branch_with_commit "$R" bog-99-released
git -C "$R" checkout -q release
git -C "$R" merge -q --no-ff -m "merge the fix" bog-99-released
git -C "$R" checkout -q bog-99-released
run 0 "merged into --base release" "$R" done --base release
run 1 "same branch, default base main" "$R" done
rm -rf "$R" "$REM"

# 15. A bad --base is a usage error, not a refusal.
R="$(make_repo)"
run 2 "--base names a ref that does not exist" "$R" done --base nope
rm -rf "$R"

# --- Gate 3: pushed to origin ----------------------------------------------

# 16. main is ahead of origin/main + done -> refused.
R="$(make_repo)"; REM="$(add_pushed_origin "$R")"
printf 'more\n' >> "$R/tracked.txt"
git -C "$R" commit -qam "unpushed work"
run 1 "main ahead of origin/main + done" "$R" done
case "$LAST_OUT" in
  *'is not pushed'*) PASS=$((PASS + 1)); printf '  ok   refusal says the base is not pushed\n' ;;
  *) FAIL=$((FAIL + 1)); printf '  FAIL refusal does not say the base is not pushed\n' ;;
esac
rm -rf "$R" "$REM"

# 17. Same state + --allow-unpushed -> accepted.
R="$(make_repo)"; REM="$(add_pushed_origin "$R")"
printf 'more\n' >> "$R/tracked.txt"
git -C "$R" commit -qam "unpushed work"
run 0 "main ahead of origin/main + done --allow-unpushed" "$R" done --allow-unpushed
rm -rf "$R" "$REM"

# 18. No origin at all -> accepted. A local-only repo is not a failure.
R="$(make_repo)"
run 0 "no origin remote + done" "$R" done
rm -rf "$R"

# 19. Gate 1 fires before gate 2: a dirty unmerged branch reports the dirt.
R="$(make_repo)"; REM="$(add_pushed_origin "$R")"
branch_with_commit "$R" bog-99-unmerged
printf 'half done\n' > "$R/wip.txt"
run 1 "dirty and unmerged + done" "$R" done
case "$LAST_OUT" in
  *uncommitted*) PASS=$((PASS + 1)); printf '  ok   the uncommitted gate reports first\n' ;;
  *) FAIL=$((FAIL + 1)); printf '  FAIL the uncommitted gate did not report first\n' ;;
esac
rm -rf "$R" "$REM"

# 20. --help still prints the option list after the usage block moved.
R="$(make_repo)"
run 0 "--help" "$R" --help
case "$LAST_OUT" in
  *--allow-unmerged*--allow-unpushed*) PASS=$((PASS + 1)); printf '  ok   --help lists the new flags\n' ;;
  *) FAIL=$((FAIL + 1)); printf '  FAIL --help does not list the new flags\n' ;;
esac
rm -rf "$R"

printf '\n%s passed, %s failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
