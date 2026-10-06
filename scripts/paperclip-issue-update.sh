#!/usr/bin/env bash
#
# paperclip-issue-update.sh — the one way an agent writes a Paperclip issue status.
#
# Why this exists: BOG-32 found four tasks (BOG-3, BOG-4, BOG-5, BOG-7) marked
# "done" while every line of their code was still uncommitted on a feature
# branch. Nothing had shipped, and the work was later discarded. This script is
# the choke point that makes that mistake loud instead of silent.
#
# Usage:
#   scripts/paperclip-issue-update.sh <status> [options]
#
#   <status>   backlog | todo | in_progress | in_review | done | blocked | cancelled
#
# Options:
#   --issue <id>      Issue to write. Default: $PAPERCLIP_TASK_ID
#   --comment <text>  Comment to post with the status change.
#   --allow-dirty     Skip the uncommitted-work gate. Use only when the task
#                     produced no code (a brief, a decision, a review).
#   --repo <path>     Repo to check. Default: the repo containing the cwd.
#   --no-shas         Do not append the task's commit SHAs to the comment.
#   --dry-run         Run every check, print the request, write nothing.
#   -h | --help       This text.
#
# Exit codes:
#   0  status written
#   1  refused: uncommitted work in the repo (see --allow-dirty)
#   2  bad usage or missing environment
#   3  the API rejected the write

set -euo pipefail

STATUSES="backlog todo in_progress in_review done blocked cancelled"

die() { printf 'paperclip-issue-update: %s\n' "$*" >&2; exit "${2:-2}"; }

usage() { sed -n '10,29p' "$0" | sed 's|^# \{0,1\}||'; }

STATUS=""
ISSUE="${PAPERCLIP_TASK_ID:-}"
COMMENT=""
ALLOW_DIRTY=0
REPO=""
WANT_SHAS=1
DRY_RUN=0

while [ $# -gt 0 ]; do
  case "$1" in
    -h|--help) usage; exit 0 ;;
    --allow-dirty) ALLOW_DIRTY=1; shift ;;
    --no-shas) WANT_SHAS=0; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    --issue) ISSUE="${2:-}"; [ -n "$ISSUE" ] || die "--issue needs a value"; shift 2 ;;
    --comment) COMMENT="${2:-}"; shift 2 ;;
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
# The gate. Only "done" is gated: a task is allowed to sit dirty while it is
# in progress, in review, blocked or cancelled.
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
# Record what shipped: the commits on this branch that are not on the base.
# A later reviewer asking "what shipped for this task?" gets a direct answer
# instead of having to reconstruct it from the branch.
# ---------------------------------------------------------------------------
SHA_NOTE=""
if [ "$STATUS" = "done" ] && [ "$WANT_SHAS" -eq 1 ]; then
  BRANCH="$(git -C "$REPO" rev-parse --abbrev-ref HEAD 2>/dev/null || echo '')"
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
