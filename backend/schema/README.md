# `backend/schema/` — the runnable lamsza schema

Two generated, committed files that bring an empty Postgres to a working lamsza
database in one command:

| File | What it is |
|---|---|
| `001_schema.sql` | The whole current schema, structure only: 42 tables, the `locations` view, the `pg_slugify`/`sync_*` functions and their triggers, indexes, constraints, and the `pg_trgm` and `unaccent` extensions. |
| `002_reference.sql` | The reference rows the app cannot work without: `counties`, `geo_locations`, `settlements`, `settlement_location_types`, `historical_seats`, `county_historical_seats`, `venue_types`, `weather_desc_translations`. |

```bash
scripts/db-bootstrap.sh                     # the local dev database
scripts/db-bootstrap.sh --create            # create it first if it is missing
scripts/db-bootstrap.sh --url "$DATABASE_URL"
```

## Why this exists

`backend/migrations/` holds 41 files that were applied by hand over the life of
the project. It is a historical record, not a runnable sequence: there is no
recorded order, and several files only ever applied to a schema that has since
moved on. Before BOG-53 there was no way to create a working lamsza database
from this repo at all — which is why the root `backend` test package, a database
integration suite, could not run in CI. It got excluded by name instead.

`backend/migrations/` stays where it is. Nothing here replaces it, and old
migrations are not to be deleted; they are the only explanation of how some
columns got there.

## Generated, and committed

Generated, because nobody keeps 2700 lines of DDL correct by hand, and a
bootstrap that is subtly wrong is worse than none at all. Committed, because CI
has to build the database out of the repo, and because a reviewer should see the
schema move in a diff.

After any schema change, regenerate and commit the result with the change:

```bash
scripts/db-dump-schema.sh
```

It reads a live database — the local dev one by default — so what lands here is
what the running app actually has. There is no CI-only copy of the schema: CI
runs the same `scripts/db-bootstrap.sh` over these same two files.

## The allowlist in `002_reference.sql`

`scripts/db-dump-schema.sh` names the reference tables explicitly rather than
excluding the ones to skip. `users`, `sessions`, `admin_sessions`,
`site_settings`, `entries` and everything else user- or admin-generated must
never land in a committed file: that is personal data, and `site_settings` can
hold configuration values that should not be public.

**Adding a table to that list is a decision to publish its contents.** Read the
rows before you do it.

## Drift

Nothing checks that these files match the live dev database byte for byte; that
would need a database, and CI only has an empty one.

What is checked is the drift that broke things: **boot DDL that never reached
the dump.** The backend creates tables and adds columns on startup (`CREATE TABLE
IF NOT EXISTS`, `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`). The dev database
gets them, a bootstrapped one does not. The `backend` test suite cannot see the
gap, because its own setup runs the same boot DDL on the scratch database first:
`admin_audit_log` was missing from this dump for a day and every suite stayed
green while lamsza-admin, which runs no DDL, had no audit table on any fresh
database.

`schema_test.go` in this folder closes that gap without a database: it reads the
non-test Go sources, collects every table and added column the boot DDL names,
and fails if `001_schema.sql` lacks one. It runs in plain `go test ./...` and in
CI. When it goes red, boot the new code against the dev database, run
`scripts/db-dump-schema.sh`, and commit the result with the change.
