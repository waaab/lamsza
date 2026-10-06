#!/usr/bin/env bash
#
# Applies the branch protection decided on BOG-54 to `main` in all four lamsza
# network repos. Nobody can run this from the dev machine today: there is no
# GitHub token there and no `gh` CLI. It is committed so that the moment a token
# exists the decision is one command away instead of sixteen clicks across four
# repo settings pages, and so the exact intended settings are reviewable in git
# rather than living only in a repo's web UI.
#
#   GITHUB_TOKEN=ghp_... bash docs/network/apply-branch-protection.sh --apply
#
# The token needs `repo` scope (classic) or Administration: write (fine-grained)
# on all four repos. Without --apply this prints the payload and changes nothing.
#
# See R7 in docs/network/WAYS_OF_WORKING.md for the decision and its cost.
set -euo pipefail

REPOS=(waaab/lamsza waaab/lamsza-admin waaab/lamsza-szotar waaab/lamsza-jatszoter)

# Exactly the two test/build jobs, in every repo. Deliberately NOT `ci-status`:
# that job is gated on `github.ref == refs/heads/main`, so it never runs on a
# pull-request branch, and a required check that never reports leaves every pull
# request blocked forever.
read -r -d '' PAYLOAD <<'JSON' || true
{
  "required_status_checks": {
    "strict": false,
    "contexts": ["frontend", "backend"]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": null,
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "required_conversation_resolution": false
}
JSON

# Why each non-obvious field is set the way it is:
#
# enforce_admins: true — this one is load-bearing, and the naive choice (false,
#   to leave the owner an escape hatch) would make the whole thing decorative.
#   Agents push with the owner's SSH key, so to GitHub every agent push *is* an
#   admin push. With enforce_admins false, admins bypass required checks, which
#   is precisely the actor this is meant to constrain. The cost is real and the
#   owner should know it: it also stops the owner pushing straight to `main`.
#   Lift it from Settings > Branches when a hand-fix is needed.
#
# required_pull_request_reviews: null — no required reviewer. There is one
#   engineering agent and one owner; requiring a review would mean nothing could
#   ever merge. The gate here is "green", not "seen".
#
# strict: false — do not force a branch to be rebased onto the latest `main`
#   before merging. WAYS_OF_WORKING §3 already runs one task per repo at a time,
#   so `main` very rarely moves under an open branch, and strict:true would add
#   a rebase-and-wait loop to every merge for that narrow window. Flip this to
#   true if parallel work per repo ever becomes normal.

if [ "${1:-}" != "--apply" ]; then
  echo "Dry run. Would PUT this to each repo's /branches/main/protection:"
  echo
  printf '%s\n' "$PAYLOAD"
  echo
  printf 'Repos: %s\n' "${REPOS[*]}"
  echo "Re-run with --apply (and GITHUB_TOKEN set) to make the change."
  exit 0
fi

: "${GITHUB_TOKEN:?GITHUB_TOKEN is required to apply branch protection}"

fail=0
for repo in "${REPOS[@]}"; do
  echo "=== $repo ==="
  code=$(curl -sS -o /tmp/bp-body.$$ -w '%{http_code}' \
    -X PUT \
    -H "Authorization: Bearer $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github+json" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    -d "$PAYLOAD" \
    "https://api.github.com/repos/$repo/branches/main/protection")
  if [ "$code" != "200" ]; then
    echo "  FAILED (HTTP $code)"
    sed 's/^/    /' /tmp/bp-body.$$
    fail=1
    continue
  fi
  # Read it back rather than trusting the write: the response to a successful PUT
  # is the stored protection, so this is what GitHub will actually enforce.
  jq -r '
    "  required checks : " + ((.required_status_checks.contexts // []) | join(", ")) +
    "\n  strict          : " + (.required_status_checks.strict | tostring) +
    "\n  include admins  : " + (.enforce_admins.enabled | tostring) +
    "\n  force pushes    : " + (.allow_force_pushes.enabled | tostring)
  ' /tmp/bp-body.$$
done
rm -f /tmp/bp-body.$$

if [ "$fail" -ne 0 ]; then
  echo
  echo "At least one repo was not protected. Nothing here is partially applied"
  echo "within a repo, but the set of repos is: re-run after fixing the token."
  exit 1
fi
echo
echo "All four repos protected. Agents now push a branch and merge once green;"
echo "update R7 in docs/network/WAYS_OF_WORKING.md if that changes again."
