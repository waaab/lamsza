#!/usr/bin/env bash
# Outside-only production probe. Public HTTPS requests, nothing else.
#
# Agents MAY run this. It does not touch the server: no SSH, no config, no deploy.
# See docs/AGENT_ENVIRONMENT_POLICY.md.
#
# Usage: scripts/probe-prod.sh
# Exit code 0 = every endpoint returned its expected status. 1 = something is wrong.

set -u

# url|expected_status|label
CHECKS=(
  "https://szotar.lamsza.com/|200|szotar static"
  "https://szotar.lamsza.com/api/health|200|szotar API health"
  "https://szotar.lamsza.com/api/words|200|szotar API words"
  "https://jatszoter.lamsza.com/|200|jatszoter static"
  "https://jatszoter.lamsza.com/api/health|200|jatszoter API health"
  "https://jatszoter.lamsza.com/api/config/public|200|jatszoter API config"
  "https://jatszoter.lamsza.com/api/daily|200|jatszoter API daily"
)

printf 'Production probe — %s\n\n' "$(date -u '+%Y-%m-%d %H:%M UTC')"
printf '%-26s %-8s %-8s %-9s %s\n' LABEL GOT WANT TTFB RESULT

failures=0
for check in "${CHECKS[@]}"; do
  IFS='|' read -r url want label <<<"$check"
  read -r code ttfb < <(curl -s -o /dev/null -m 20 -w '%{http_code} %{time_starttransfer}' "$url")
  if [ "$code" = "$want" ]; then
    result=ok
  else
    result=FAIL
    failures=$((failures + 1))
  fi
  printf '%-26s %-8s %-8s %-9s %s\n' "$label" "$code" "$want" "${ttfb}s" "$result"
done

echo
if [ "$failures" -eq 0 ]; then
  echo "All $((${#CHECKS[@]})) checks pass."
  exit 0
fi

echo "$failures check(s) failed."
cat <<'EOF'

How to read a failure:
  502 with TTFB under ~0.5s  -> nginx got "connection refused". Nothing is
                                listening on the proxy_pass port. The service is
                                stopped, failed, or bound to a different port.
  502 or 504 with TTFB over
  ~5s                        -> the backend accepted the connection and then
                                hung. The process is up but stuck (database,
                                deadlock, upstream call).
  404 on /api/*              -> the route does not exist in that build. A
                                deploy/route problem, not a down service.
  000                        -> DNS or TLS failure before any HTTP happened.

Next step for a 502: the owner runs Step 0 of docs/GO_LIVE_VERIFICATION_RUNBOOK.md.
Agents must not run it.
EOF
exit 1
