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
| push branches to GitHub | change DNS, domains or registrar settings |
| open pull requests | commit or merge to `main` |
| write runbooks for the owner to execute | deploy, restart services, or reboot the droplet |

`main` on each repo is the production branch. **Only Attila commits and merges to `main`,
and only Attila deploys.** An agent that wants a production change describes it as an exact
command list for the owner and stops there.

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
