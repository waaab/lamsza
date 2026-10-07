# `docs/superpowers/` — archived design records

Every file under `plans/` and `specs/` is **dated and archived**. Each one records the
design and the step-by-step build of one feature at the time it was built. They are not
current documentation and they are deliberately not rewritten when the code moves — that is
the convention already stated in `project-status-audit.md`.

Read them as history. For what is true now, use the live docs:

| For | Read |
|---|---|
| How the four apps are organised | `docs/network/WAYS_OF_WORKING.md` |
| What "verified" means, and the port map | `docs/LOCAL_DEV_CHECKS.md` |
| Current product state | `CHANGELOG.md` (the old `project-status-audit.md` is in `docs/history/`) |
| Locations / settlements data contract | `docs/SETTLEMENTS_AND_LOCATIONS.md` |

## The one trap to know about

These plans were written while the admin UI and the `/api/admin/*` routes lived in **this**
repo. They do not any more.

- **This app has no `/admin` route** and `backend/main.go` registers no `/api/admin/*`
  route. The admin UI is the independent
  [`waaab/lamsza-admin`](https://github.com/waaab/lamsza-admin) app —
  `http://localhost:5173` locally (API `:3000`), `admin.lamsza.com` in production. See
  `docs/ADMIN_EXTRACTION.md`.
- Wherever a plan here says `src/routes/admin/+page.svelte`, the live equivalent is
  `lamsza-admin`'s `frontend/src/routes/+page.svelte`. That path **does not exist in this
  repo** — do not create it.
- Wherever a plan here says an admin endpoint such as `GET /api/admin/listing-queue` or
  `PUT /api/admin/county_seat`, that endpoint is served by `lamsza-admin` on `:3000`, not by
  this app on `:3001`.

The Go admin handlers that used to live in this repo (`backend/internal/handlers/admin_*.go`)
were deleted on BOG-42; they exist only in `lamsza-admin`.
