# Agent Environment Policy — local only

**Set by:** Attila (owner), 2026-10-06
**Applies to:** every agent working on `lamsza`, `szotar`, `jatszoter`, `lamsza-admin`

---

## The rule

**Agents work on the local machine only. Production is the owner's, alone.**

| Agents do | Agents never do |
|---|---|
| develop and test on `localhost` | touch the production server |
| run the local DBs, backends and frontends | hold or ask for SSH keys or server credentials |
| run the unit tests and local smoke checks | change production `.env`, nginx, systemd or TLS |
| commit, merge and push to GitHub | change DNS, domains or registrar settings |
| open and merge pull requests | deploy, restart services, or reboot the droplet |
| write runbooks for the owner to execute | log in to the droplet by any route |

## Git: granted (2026-10-06)

Attila granted agents **commit, merge and push to GitHub**, including `main`. The first rule
drafted here said otherwise; this line replaces it.

The grant covers **git only**. It does not grant anything on the server. An agent still never
deploys, never restarts a service, and never changes a live config.

**The reason Attila gave, and the limit that comes with it:** `main` is not connected to
production, and will not be for now. A push is therefore only a push. His words: *"until then
you are allowed to commit, merge."*

⚠️ **So the grant expires on its own condition.** The plan is to wire each repo's `main` to
production. On the day that happens, a merge to `main` *becomes* a deploy — the reason behind
the grant is gone, and the grant goes with it. Agents stop merging to `main` at that point and
wait for Attila to set the new rule. The usual shape: agents merge to a staging branch, Attila
promotes to `main`.

## What this means in practice

1. **No agent fixes a production incident.** An agent can detect one from outside (a public
   `curl`, a 502) and must report it with evidence. The fix is the owner's.
2. **"Verified" means verified locally.** An agent may only tick a checklist item it proved
   on `localhost`. A go-live checklist item about a live domain stays open until the owner
   checks it on the live domain.
3. **Read-only outside probing is allowed.** Public, unauthenticated `GET`s against live
   URLs are fine for reporting. Nothing that writes, authenticates as an admin, or loads a
   live system.
4. **The agent's job is that the deploy is boring.** Everything that *can* be proved locally
   is proved locally first, so the owner's production pass is a confirmation and not a debug
   session.

## The two runbooks

| Doc | Owner | Scope |
|---|---|---|
| `docs/LOCAL_VERIFICATION_RUNBOOK.md` | **agents** | localhost equivalents of go-live items 6–11 and 13 |
| `docs/GO_LIVE_VERIFICATION_RUNBOOK.md` | **Attila only** | the live droplet and the live domains |

They cover the same behaviours on purpose. The local one runs on every change; the
production one runs once per cutover, by the owner.
