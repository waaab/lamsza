#!/usr/bin/env bash
# PROD 502 triage — evidence first, then fix.
#
# OWNER ONLY. Agents must not run this file. It runs ON the droplet
# (146.190.204.232) and it restarts services. See docs/AGENT_ENVIRONMENT_POLICY.md.
#
# How to use:
#   scp scripts/owner-prod-502-triage.sh attila@146.190.204.232:~/
#   ssh attila@146.190.204.232 'bash ~/owner-prod-502-triage.sh'
#
# It prints a labelled report, writes it to ~/prod-502-triage-<timestamp>.log,
# and stops before restarting anything. It only restarts if you pass --restart.
#
# This is Step 0 of docs/GO_LIVE_VERIFICATION_RUNBOOK.md in one paste.

set -u

RESTART=no
[ "${1:-}" = "--restart" ] && RESTART=yes

LOG="$HOME/prod-502-triage-$(date -u '+%Y%m%dT%H%M%SZ').log"
exec > >(tee "$LOG") 2>&1

section() { printf '\n========== %s ==========\n' "$1"; }

printf 'PROD 502 triage — %s on %s\n' "$(date -u '+%Y-%m-%d %H:%M UTC')" "$(hostname)"
printf 'Report file: %s\n' "$LOG"

section '1. What does systemd think?'
systemctl status szotar jatszoter --no-pager || true

section '2. Is anything listening on 8081 / 8082 / 8083?'
ss -ltnp | grep -E ':(8081|8082|8083)' || echo 'NOTHING LISTENING on 8081, 8082 or 8083.'

section '3. Is Postgres up?'
systemctl is-active postgresql || true
ss -ltn | grep -E ':5432' || echo 'Nothing listening on 5432.'

section '4. Why did szotar stop? (last 80 lines)'
journalctl -u szotar -n 80 --no-pager || true

section '5. Why did jatszoter stop? (last 80 lines)'
journalctl -u jatszoter -n 80 --no-pager || true

section '6. Does the PORT in .env match what nginx proxies to?'
for app in szotar jatszoter; do
  printf -- '--- %s ---\n' "$app"
  sudo grep -H '^PORT' "/var/www/$app/.env" 2>/dev/null || echo "no PORT line in /var/www/$app/.env"
  grep -rhoE 'proxy_pass +http://127\.0\.0\.1:[0-9]+' "/etc/nginx/sites-enabled/$app"* 2>/dev/null \
    | sort -u || echo "no proxy_pass found for $app in /etc/nginx/sites-enabled"
done

section '7. Did the deploy land? (binary present and when)'
ls -l /var/www/szotar/szotar /var/www/jatszoter/jatszoter 2>&1 || true

section '8. Were the services enabled at boot? (go-live item 10)'
systemctl is-enabled szotar jatszoter 2>&1 || true

section '9. Last boot — did a reboot start this?'
uptime -p
who -b || true

if [ "$RESTART" = no ]; then
  cat <<EOF

========== STOP: read the evidence above before fixing ==========

Nothing was changed. Match what you see to the failure table in
docs/GO_LIVE_VERIFICATION_RUNBOOK.md Step 0:

  inactive (dead), no recent logs, is-enabled says "disabled"
      -> never started after a reboot.
         Fix: sudo systemctl enable --now szotar jatszoter   (also closes item 10)

  activating (auto-restart), log shows connection refused to 127.0.0.1:5432
      -> Postgres is down. Fix Postgres first, then the apps.

  log shows FATAL: password authentication failed
      -> DATABASE_URL in /var/www/<app>/.env is wrong. Fix it, keep mode 0600.

  section 6 shows PORT and proxy_pass disagreeing
      -> set PORT to match nginx, then restart.

  section 7 shows no binary
      -> the deploy never landed. Redeploy per
         docs/PRODUCTION_ENVIRONMENT_NOTES.md section 6.

When you know the cause, re-run with:  bash $0 --restart

Then paste this file into issue BOG-11:  $LOG
EOF
  exit 0
fi

section '10. Restarting'
sudo systemctl restart szotar jatszoter
sleep 3
systemctl status szotar jatszoter --no-pager || true

section '11. Do the backends answer locally now?'
for port in 8081 8082; do
  code=$(curl -s -o /dev/null -m 10 -w '%{http_code}' "http://127.0.0.1:$port/api/health")
  printf '127.0.0.1:%s/api/health -> %s  (want 200)\n' "$port" "$code"
done

section '12. Do they answer from outside now?'
for url in https://szotar.lamsza.com/api/health https://jatszoter.lamsza.com/api/health; do
  code=$(curl -s -o /dev/null -m 15 -w '%{http_code}' "$url")
  printf '%s -> %s  (want 200)\n' "$url" "$code"
done

cat <<EOF

========== Done ==========
Report: $LOG

Do not close BOG-11 on the restart alone. Paste the report, including the
root-cause log line from section 4 or 5, into the issue. If the cause was a
reboot, go-live item 10 (services enabled at boot) also failed and needs
"sudo systemctl enable szotar jatszoter".
EOF
