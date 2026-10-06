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
| commit and merge locally | change DNS, domains or registrar settings |
| open pull requests; push to `main`, or merge a pull request into it, **only after Attila confirms** | deploy, restart services, or reboot the droplet |

All four apps are **under development**. Until the code is finished and working on
localhost, production is out of scope — agents do not probe it, document it, or open tasks
about it. Attila owns go-live and will say when that changes.

## Git: commit and merge locally; push only on confirmation (2026-10-07)

Agents may **commit and merge locally**, including into local `main`.

**Pushing to `main` always needs Attila's explicit confirmation**, every time. Show what will
be pushed (the commits, `git log origin/main..main`), ask, and push only after a yes. A yes
covers that push, not later ones. Merging a pull request into `main` on GitHub counts as a
push and needs the same confirmation.

The grant covers **git only**. It does not grant anything on the server.

History: on 2026-10-06 Attila granted agents commit, merge and push to GitHub, `main`
included, because `main` is not connected to production. Attila narrowed it to the rule above on
2026-10-07. When each repo's `main` is wired to production, a merge to `main` *becomes* a
deploy, and Attila will set the rule for that day.

## What this means in practice

1. **"Verified" means verified on localhost.** With the output recorded, not summarised.
2. **No agent debugs a live site.** If an agent notices one is down, that is Attila's to
   handle, and it is not a reason to open a task.
3. **Local checks come first.** `docs/LOCAL_DEV_CHECKS.md` is the standing check: bring the
   network up, smoke every API, run every suite.
