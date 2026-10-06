# Agent Environment Policy — local development only

**Set by:** Attila (owner), 2026-10-06
**Applies to:** every agent working on `lamsza`, `szotar`, `jatszoter`, `lamsza-admin`

---

## The rule

**Agents develop and test on this machine. Production is not agent work at all.**

| Agents do | Agents never do |
|---|---|
| develop and test on `localhost` | touch the production server |
| run the local DBs, backends and frontends | hold or ask for SSH keys or server credentials |
| run the unit tests and local smoke checks | change production `.env`, nginx, systemd or TLS |
| commit, merge and push to GitHub | change DNS, domains or registrar settings |
| open and merge pull requests | deploy, restart services, or reboot the droplet |

All four apps are **under development**. Until the code is finished and working on
localhost, production is out of scope — agents do not probe it, document it, or open tasks
about it. Attila owns go-live and will say when that changes.

## Git: granted (2026-10-06)

Attila granted agents **commit, merge and push to GitHub**, including `main`.

The grant covers **git only**. It does not grant anything on the server.

**The reason, and the limit that comes with it:** `main` is not connected to production, so
a push is only a push. The plan is to wire each repo's `main` to production later. On that
day a merge to `main` *becomes* a deploy, the reason behind the grant is gone, and agents
stop merging to `main` until Attila sets the new rule.

## What this means in practice

1. **"Verified" means verified on localhost.** With the output recorded, not summarised.
2. **No agent debugs a live site.** If an agent notices one is down, that is Attila's to
   handle, and it is not a reason to open a task.
3. **Local checks come first.** `docs/LOCAL_DEV_CHECKS.md` is the standing check: bring the
   network up, smoke every API, run every suite.
