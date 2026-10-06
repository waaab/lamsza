#!/usr/bin/env bash
#
# paperclip-issue-update.sh — the one way an agent writes a Paperclip issue status.
#
# Why this exists: BOG-32 found four tasks (BOG-3, BOG-4, BOG-5, BOG-7) marked
# "done" while every line of their code was still uncommitted on a feature
# branch. Nothing had shipped, and the work was later discarded. This script is
# the choke point that makes that mistake loud instead of silent.
#
# BOG-38 widened it. BOG-17 and BOG-28 passed the first gate — their code was
# committed — but it sat on a branch nobody merged, and the merge in turn sat on
# a local main nobody pushed. "Done" has to mean the owner has the fix, so there
# are three gates now: committed, merged into the base, and pushed to origin.
#
# Usage:
#   scripts/paperclip-issue-update.sh <status> [options]
#
#   <status>   backlog | todo | in_progress | in_review | done | blocked | cancelled
#
# Options:
#   --issue <id>       Issue to write. Default: $PAPERCLIP_TASK_ID
#   --comment <text>   Comment to post with the status change.
#   --allow-dirty      Skip the uncommitted-work gate. Use only when the task
#                      produced no code (a brief, a decision, a review).
#   --allow-unmerged   Skip the merged-into-base gate. Use only when the task
#                      is meant to stay on a branch (a snapshot, a spike) and
#                      say why in the comment.
#   --allow-unpushed   Skip the pushed-to-origin gate. Use only when the repo
#                      has no remote on purpose.
#   --base <ref>       Branch the work has to land on. Default: the first of
#                      main, master that exists.
#   --repo <path>      Repo to check. Default: the repo containing the cwd.
#   --no-shas          Do not append the task's commit SHAs to the comment.
#   --dry-run          Run every check, print the request, write nothing.
#   -h | --help        This text.
#
# Exit codes:
#   0  status written
#   1  refused by a gate: uncommitted, unmerged, or unpushed
#   2  bad usage or missing environment
#   3  the API rejected the write

set -euo pipefail

STATUSES="backlog todo in_progress in_review done blocked cancelled"

die() { printf 'paperclip-issue-update: %s\n' "$*" >&2; exit "${2:-2}"; }

usage() { sed -n '/^# Usage:/,/^# *3  the API/p' "$0" | sed 's|^# \{0,1\}||'; }

STATUS=""
ISSUE="${PAPERCLIP_TASK_ID:-}"
COMMENT=""
ALLOW_DIRTY=0
ALLOW_UNMERGED=0
ALLOW_UNPUSHED=0
BASE_REF=""
REPO=""
WANT_SHAS=1
DRY_RUN=0

while [ $# -gt 0 ]; do
  case "$1" in
    -h|--help) usage; exit 0 ;;
    --allow-dirty) ALLOW_DIRTY=1; shift ;;
    --allow-unmerged) ALLOW_UNMERGED=1; shift ;;
    --allow-unpushed) ALLOW_UNPUSHED=1; shift ;;
    --no-shas) WANT_SHAS=0; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    --issue) ISSUE="${2:-}"; [ -n "$ISSUE" ] || die "--issue needs a value"; shift 2 ;;
    --comment) COMMENT="${2:-}"; shift 2 ;;
    --base) BASE_REF="${2:-}"; [ -n "$BASE_REF" ] || die "--base needs a value"; shift 2 ;;
    --repo) REPO="${2:-}"; [ -n "$REPO" ] || die "--repo needs a value"; shift 2 ;;
    -*) die "unknown option: $1" ;;
    *)
      [ -z "$STATUS" ] || die "give exactly one status (got '$STATUS' and '$1')"
      STATUS="$1"; shift ;;
  esac
done

[ -n "$STATUS" ] || { usage >&2; die "no status given"; }
case " $STATUSES " in
  *" $STATUS "*) ;;
  *) die "'$STATUS' is not a status. Pick one of: $STATUSES" ;;
esac

# ---------------------------------------------------------------------------
# Locate the repo.
# ---------------------------------------------------------------------------
if [ -z "$REPO" ]; then
  REPO="$(git rev-parse --show-toplevel 2>/dev/null || true)"
  if [ -z "$REPO" ] && [ -n "${PAPERCLIP_WORKSPACE_CWD:-}" ]; then
    REPO="$(git -C "$PAPERCLIP_WORKSPACE_CWD" rev-parse --show-toplevel 2>/dev/null || true)"
  fi
fi
[ -n "$REPO" ] || die "not inside a git repo, and no --repo given"
git -C "$REPO" rev-parse --git-dir >/dev/null 2>&1 || die "not a git repo: $REPO"

# ---------------------------------------------------------------------------
# Resolve the base branch the work has to land on, and the branch HEAD is on.
# Local refs only: "is it merged" and "is it pushed" are two separate
# questions, and mixing origin/main into the first one answers neither.
# ---------------------------------------------------------------------------
BRANCH="$(git -C "$REPO" symbolic-ref --quiet --short HEAD 2>/dev/null || echo '')"

if [ -z "$BASE_REF" ]; then
  for candidate in main master; do
    if git -C "$REPO" rev-parse --verify --quiet "refs/heads/$candidate" >/dev/null 2>&1; then
      BASE_REF="$candidate"; break
    fi
  done
else
  git -C "$REPO" rev-parse --verify --quiet "$BASE_REF" >/dev/null 2>&1 \
    || die "--base '$BASE_REF' is not a ref in $REPO"
fi

# ---------------------------------------------------------------------------
# Gate 1 — committed. Only "done" is gated: a task is allowed to sit dirty
# while it is in progress, in review, blocked or cancelled.
#
# `git status --porcelain` already honours .gitignore, so ignored paths
# (node_modules, dist, logs, .claude/settings.local.json, ...) never reach here.
# ---------------------------------------------------------------------------
if [ "$STATUS" = "done" ] && [ "$ALLOW_DIRTY" -eq 0 ]; then
  DIRTY="$(git -C "$REPO" status --porcelain)"
  if [ -n "$DIRTY" ]; then
    {
      printf '\nREFUSED: will not mark %s done. The repo still has uncommitted work.\n\n' "${ISSUE:-this issue}"
      printf '  repo: %s\n' "$REPO"
      printf '  branch: %s\n\n' "$(git -C "$REPO" rev-parse --abbrev-ref HEAD 2>/dev/null || echo '(detached)')"
      printf '%s\n' "$DIRTY" | sed 's/^/  /'
      printf '\nCommit the work, or move the task back. Do not mark it done.\n'
      printf 'If none of this belongs to your task — a shared workspace, another\n'
      printf 'run, or a task whose output is not code — pass --allow-dirty.\n\n'
    } >&2
    exit 1
  fi
fi

# ---------------------------------------------------------------------------
# Gate 2 — merged. BOG-17 and BOG-28 both passed gate 1 and still shipped
# nothing: the commits were real, on a branch nobody merged. HEAD has to be an
# ancestor of the base branch before the task can call itself done.
# ---------------------------------------------------------------------------
if [ "$STATUS" = "done" ] && [ "$ALLOW_UNMERGED" -eq 0 ] && [ -n "$BASE_REF" ]; then
  if ! git -C "$REPO" merge-base --is-ancestor HEAD "$BASE_REF" 2>/dev/null; then
    AHEAD="$(git -C "$REPO" log --no-merges --format='%h %s' "$BASE_REF..HEAD" 2>/dev/null | head -50 || true)"
    {
      printf '\nREFUSED: will not mark %s done. The work is not on `%s`.\n\n' "${ISSUE:-this issue}" "$BASE_REF"
      printf '  repo: %s\n' "$REPO"
      printf '  branch: %s\n' "${BRANCH:-(detached HEAD)}"
      printf '  base: %s\n\n' "$BASE_REF"
      if [ -n "$AHEAD" ]; then
        printf 'These commits exist only on the branch:\n\n'
        printf '%s\n' "$AHEAD" | sed 's/^/  /'
      fi
      printf '\nMerge the branch into `%s` (or land a rebased copy), then mark it done.\n' "$BASE_REF"
      printf 'If the work is meant to stay on a branch — a snapshot, a spike, work\n'
      printf 'another task owns — pass --allow-unmerged and say why in the comment.\n\n'
    } >&2
    exit 1
  fi
fi

# ---------------------------------------------------------------------------
# Gate 3 — pushed. A merge into a local `main` nobody pushed is still invisible
# to the owner. Only checked when the repo actually has the matching remote
# branch; a local-only repo is not a failure.
# ---------------------------------------------------------------------------
if [ "$STATUS" = "done" ] && [ "$ALLOW_UNPUSHED" -eq 0 ] && [ -n "$BASE_REF" ]; then
  REMOTE_BASE="origin/$BASE_REF"
  if git -C "$REPO" rev-parse --verify --quiet "refs/remotes/$REMOTE_BASE" >/dev/null 2>&1; then
    if ! git -C "$REPO" merge-base --is-ancestor "$BASE_REF" "$REMOTE_BASE" 2>/dev/null; then
      UNPUSHED="$(git -C "$REPO" log --format='%h %s' "$REMOTE_BASE..$BASE_REF" 2>/dev/null | head -50 || true)"
      {
        printf '\nREFUSED: will not mark %s done. `%s` is not pushed.\n\n' "${ISSUE:-this issue}" "$BASE_REF"
        printf '  repo: %s\n' "$REPO"
        printf '  base: %s\n' "$BASE_REF"
        printf '  remote: %s\n\n' "$REMOTE_BASE"
        if [ -n "$UNPUSHED" ]; then
          printf 'These commits are on the local base only:\n\n'
          printf '%s\n' "$UNPUSHED" | sed 's/^/  /'
        fi
        printf '\nRun: git -C %s push origin %s\n' "$REPO" "$BASE_REF"
        printf 'If this repo has no remote on purpose, pass --allow-unpushed.\n\n'
      } >&2
      exit 1
    fi
  fi
fi

# ---------------------------------------------------------------------------
# Record what shipped: the commits on this branch that are not on the base.
# A later reviewer asking "what shipped for this task?" gets a direct answer
# instead of having to reconstruct it from the branch.
# ---------------------------------------------------------------------------
SHA_NOTE=""
if [ "$STATUS" = "done" ] && [ "$WANT_SHAS" -eq 1 ]; then
  BASE=""
  for candidate in origin/main main origin/master master; do
    if git -C "$REPO" rev-parse --verify --quiet "$candidate" >/dev/null 2>&1; then
      BASE="$candidate"; break
    fi
  done
  LOG=""
  if [ -n "$BASE" ] && [ "$BRANCH" != "${BASE#origin/}" ]; then
    LOG="$(git -C "$REPO" log --no-merges --format='%h %s' "$BASE..HEAD" 2>/dev/null | head -50 || true)"
  fi
  if [ -n "$LOG" ]; then
    SHA_NOTE="$(printf 'Commits on `%s` not on `%s`:\n\n```\n%s\n```' "$BRANCH" "$BASE" "$LOG")"
  else
    SHA_NOTE="$(printf 'HEAD: `%s` on `%s` (no branch-only commits found).' \
      "$(git -C "$REPO" rev-parse --short HEAD 2>/dev/null || echo 'unknown')" "${BRANCH:-unknown}")"
  fi
fi

BODY="$COMMENT"
if [ -n "$SHA_NOTE" ]; then
  if [ -n "$BODY" ]; then
    BODY="$(printf '%s\n\n---\n\n%s' "$BODY" "$SHA_NOTE")"
  else
    BODY="$SHA_NOTE"
  fi
fi

# ---------------------------------------------------------------------------
# Write it.
# ---------------------------------------------------------------------------
[ -n "$ISSUE" ] || die "no issue id: pass --issue or set PAPERCLIP_TASK_ID"

PAYLOAD="$(STATUS="$STATUS" BODY="$BODY" node -e '
const out = { status: process.env.STATUS };
if (process.env.BODY) out.comment = process.env.BODY;
process.stdout.write(JSON.stringify(out));
')"

if [ "$DRY_RUN" -eq 1 ]; then
  printf 'dry run — would PATCH issue %s with:\n%s\n' "$ISSUE" "$PAYLOAD"
  exit 0
fi

[ -n "${PAPERCLIP_API_URL:-}" ] || die "PAPERCLIP_API_URL is not set"
[ -n "${PAPERCLIP_API_KEY:-}" ] || die "PAPERCLIP_API_KEY is not set"

BASE="${PAPERCLIP_API_URL%/}"; BASE="${BASE%/api}"

CURL_ARGS=(
  -sS -X PATCH "$BASE/api/issues/$ISSUE"
  -H "Authorization: Bearer $PAPERCLIP_API_KEY"
  -H "Content-Type: application/json"
)
if [ -n "${PAPERCLIP_RUN_ID:-}" ]; then
  CURL_ARGS+=(-H "X-Paperclip-Run-Id: $PAPERCLIP_RUN_ID")
fi

RESPONSE_FILE="$(mktemp "${PAPERCLIP_RUN_SCRATCH_DIR:-${TMPDIR:-/tmp}}/issue-update.XXXXXX")"
trap 'rm -f "$RESPONSE_FILE"' EXIT

HTTP_CODE="$(curl "${CURL_ARGS[@]}" -o "$RESPONSE_FILE" -w '%{http_code}' -d "$PAYLOAD")"

case "$HTTP_CODE" in
  2*) printf 'issue %s -> %s\n' "$ISSUE" "$STATUS" ;;
  *)
    cat "$RESPONSE_FILE" >&2
    printf '\n' >&2
    die "the API refused the write (HTTP $HTTP_CODE)" 3
    ;;
esac
